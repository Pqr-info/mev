package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Global dummy values to avoid compile-time dead-code elimination
var (
	benchTxHash [32]byte
	benchData   []byte
	benchNonce  uint64
)

// --------------------------------------------------------------------
// 1. TYPEDEFS & STRUCTS FOR TESTING
// --------------------------------------------------------------------

type ExecutionLegState string

const (
	LegPending   ExecutionLegState = "PENDING"
	LegCompleted ExecutionLegState = "COMPLETED"
	LegFailed    ExecutionLegState = "FAILED"
)

type TxHashBytes [32]byte

type TradePosition struct {
	ID           string
	Asset        string
	Size         float64
	CeFiStatus   ExecutionLegState
	DeFiTxHash   TxHashBytes
	DeFiStatus   ExecutionLegState
	SubmittedAt  time.Time
}

type TxReceipt struct {
	TxHash       TxHashBytes
	Success      bool
	GasUsed      uint64
	RevertReason string
}

type WatcherShard struct {
	mu           sync.RWMutex
	activeTrades map[TxHashBytes]*TradePosition
}

type DeltaWatcherRouter struct {
	shards      [256]*WatcherShard
	reHedgeChan chan *TradePosition
	receiptChan chan TxReceipt
	closeChan   chan struct{}
}

func NewDeltaWatcherRouter() *DeltaWatcherRouter {
	dw := &DeltaWatcherRouter{
		reHedgeChan: make(chan *TradePosition, 1024),
		receiptChan: make(chan TxReceipt, 4096),
		closeChan:   make(chan struct{}),
	}
	for i := 0; i < 256; i++ {
		dw.shards[i] = &WatcherShard{
			activeTrades: make(map[TxHashBytes]*TradePosition),
		}
	}
	return dw
}

func (dw *DeltaWatcherRouter) getShard(hash TxHashBytes) *WatcherShard {
	return dw.shards[hash[0]]
}

func (dw *DeltaWatcherRouter) TrackTrade(trade *TradePosition) {
	shard := dw.getShard(trade.DeFiTxHash)
	shard.mu.Lock()
	shard.activeTrades[trade.DeFiTxHash] = trade
	shard.mu.Unlock()
}

func (dw *DeltaWatcherRouter) StartMonitoringLoop(ctx context.Context, hedger *Hedger) {
	// Spin up a pool of dedicated emergency hedge workers
	for i := 0; i < 4; i++ {
		go dw.reHedgeWorker(ctx, i, hedger)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-dw.closeChan:
			return
		case receipt := <-dw.receiptChan:
			dw.processReceipt(receipt)
		}
	}
}

func (dw *DeltaWatcherRouter) IngestReceipt(receipt TxReceipt) {
	select {
	case dw.receiptChan <- receipt:
	default:
		fmt.Printf("[WARN] Receipt channel saturated. Dropping prefix: %x\n", receipt.TxHash[:4])
	}
}

func (dw *DeltaWatcherRouter) processReceipt(receipt TxReceipt) {
	shard := dw.getShard(receipt.TxHash)
	
	shard.mu.RLock()
	trade, exists := shard.activeTrades[receipt.TxHash]
	shard.mu.RUnlock()

	if !exists {
		return
	}

	shard.mu.Lock()
	delete(shard.activeTrades, receipt.TxHash)
	shard.mu.Unlock()

	if receipt.Success {
		trade.DeFiStatus = LegCompleted
		fmt.Printf("[TELEMETRY] Multi-leg trade %s settled cleanly.\n", trade.ID)
	} else {
		trade.DeFiStatus = LegFailed
		fmt.Printf("[CRITICAL] Revert on hash %x! Triggering hedge circuit.\n", receipt.TxHash[:4])
		
		select {
		case dw.reHedgeChan <- trade:
		default:
			fmt.Printf("[FATAL] HEDGING QUEUE SATURATED FOR ASSET %s\n", trade.Asset)
		}
	}
}

func (dw *DeltaWatcherRouter) reHedgeWorker(ctx context.Context, workerID int, hedger *Hedger) {
	for {
		select {
		case <-ctx.Done():
			return
		case trade := <-dw.reHedgeChan:
			fmt.Printf("[HEALER-WORKER-%d] Executing priority emergency liquidation for %s\n", workerID, trade.Asset)
			_ = hedger.ExecuteEmergencyHedgeWithFallback(ctx, trade)
		}
	}
}

// OrganStatus represents engine state
type OrganStatus string
const (
	StatusNominal  OrganStatus = "NOMINAL"
	StatusDegraded OrganStatus = "DEGRADED"
	StatusHealing  OrganStatus = "HEALING"
)

type MeshAdapter struct {
	mu           sync.RWMutex
	status       OrganStatus
	gasThreshold float64
}

func NewMeshAdapter(maxGasPriceGwei float64) *MeshAdapter {
	return &MeshAdapter{
		status:       StatusNominal,
		gasThreshold: maxGasPriceGwei,
	}
}

func (m *MeshAdapter) CheckHealth() OrganStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

func (m *MeshAdapter) HealerHookModeB(ctx context.Context, currentGasGwei float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if currentGasGwei > m.gasThreshold*1.5 {
		m.status = StatusDegraded
		fmt.Printf("[HEALER] Anomaly detected: Gas at %.2f Gwei spikes above boundary!\n", currentGasGwei)
	}
}

type Hedger struct {
	meshAdapter *MeshAdapter
}

func NewHedger(adapter *MeshAdapter) *Hedger {
	return &Hedger{meshAdapter: adapter}
}

func (h *Hedger) ExecuteEmergencyHedgeWithFallback(parentCtx context.Context, trade *TradePosition) error {
	hedgeCtx, cancelHedge := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelHedge()

	fmt.Printf("[HEALER] Tier 1: Attempting hedge on primary venue for %s (Size: %.4f)...\n", trade.Asset, trade.Size)
	err := h.executeCeFiOrder(hedgeCtx, "PRIMARY_CEX", trade)
	if err == nil {
		fmt.Println("[HEALER] Tier 1 hedge succeeded.")
		return nil
	}
	fmt.Printf("[WARN] Tier 1 hedge failed: %v. Escalating to Tier 2.\n", err)

	fmt.Printf("[HEALER] Tier 2: Attempting hedge on backup venue for %s...\n", trade.Asset)
	err = h.executeCeFiOrder(hedgeCtx, "BACKUP_CEX", trade)
	if err == nil {
		fmt.Println("[HEALER] Tier 2 hedge succeeded.")
		return nil
	}
	fmt.Printf("[WARN] Tier 2 hedge failed: %v. Escalating to Tier 3 (On-chain fail-safe).\n", err)

	fmt.Printf("[HEALER] Tier 3: Executing permissionless on-chain swap for %s...\n", trade.Asset)
	err = h.executeOnChainDeFiSwap(hedgeCtx, trade)
	if err == nil {
		fmt.Println("[HEALER] Tier 3 on-chain fallback swap succeeded. Position recovered.")
		return nil
	}

	h.quarantineNode(trade, err)
	return fmt.Errorf("all hedging tiers exhausted: %w", err)
}

func (h *Hedger) executeCeFiOrder(ctx context.Context, venue string, trade *TradePosition) error {
	dialCtx, cancel := context.WithTimeout(ctx, 350*time.Millisecond)
	defer cancel()

	select {
	case <-dialCtx.Done():
		return dialCtx.Err()
	case <-time.After(10 * time.Millisecond):
		return fmt.Errorf("venue %s network timeout (HTTP 503)", venue)
	}
}

func (h *Hedger) executeOnChainDeFiSwap(ctx context.Context, trade *TradePosition) error {
	fmt.Printf("[DeFi] Raw transaction broadcast to mempool via local execution wallet for %s.\n", trade.Asset)
	return nil
}

func (h *Hedger) quarantineNode(trade *TradePosition, lastErr error) {
	h.meshAdapter.mu.Lock()
	h.meshAdapter.status = StatusHealing
	h.meshAdapter.mu.Unlock()
	fmt.Printf("[FATAL ANOMALY] Hard lock engaged. Node quarantined. Unhedged trade ID: %s | Last Error: %v\n", trade.ID, lastErr)
}

// --------------------------------------------------------------------
// 2. RPC SIMULATOR FOR INTEGRATION
// --------------------------------------------------------------------

type RPCResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type MockBlockHeader struct {
	Number  string `json:"number"`
	BaseFee string `json:"baseFeePerGas"`
}

type LocalRPCSimulator struct {
	sockPath    string
	listener    net.Listener
	wg          sync.WaitGroup
	baseFeeGwei uint64
	mu          sync.RWMutex
	running     int32
}

func NewLocalRPCSimulator(sockPath string) *LocalRPCSimulator {
	_ = os.Remove(sockPath)
	return &LocalRPCSimulator{
		sockPath:    sockPath,
		baseFeeGwei: 25,
	}
}

func (s *LocalRPCSimulator) Start(ctx context.Context) error {
	l, err := net.Listen("unix", s.sockPath)
	if err != nil {
		return fmt.Errorf("failed to bind UDS listener: %w", err)
	}
	s.listener = l
	atomic.StoreInt32(&s.running, 1)

	fmt.Printf("[SIMULATOR] Mock RPC Socket online at %s\n", s.sockPath)
	s.wg.Add(1)
	go s.acceptLoop(ctx)
	return nil
}

func (s *LocalRPCSimulator) SetBaseFee(gwei uint64) {
	s.mu.Lock()
	s.baseFeeGwei = gwei
	s.mu.Unlock()
}

func (s *LocalRPCSimulator) acceptLoop(ctx context.Context) {
	defer s.wg.Done()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handleConnection(ctx, conn)
	}
}

func (s *LocalRPCSimulator) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			var req map[string]interface{}
			if err := decoder.Decode(&req); err != nil {
				return
			}

			method, _ := req["method"].(string)
			reqID, _ := req["id"].(float64)

			s.mu.RLock()
			currentFee := s.baseFeeGwei
			s.mu.RUnlock()

			var resp RPCResponse
			resp.Jsonrpc = "2.0"
			resp.ID = int(reqID)

			switch method {
			case "eth_ethclient_HeaderByNumber", "eth_getBlockByNumber":
				resp.Result = MockBlockHeader{
					Number:  "0x10c",
					BaseFee: fmt.Sprintf("0x%x", currentFee*1000000000),
				}
			case "eth_sendRawTransaction":
				if currentFee > 50 {
					fmt.Println("[SIMULATOR] Gas spike threshold reached! Simulating Tier 1/2 API Blackout...")
					return
				}
				resp.Result = "0x4a5c531d0d97034b7f3299719cf5c490a61358ef9c7199ec68cf5b2b2b1fb2b2"
			default:
				resp.Result = "0x0"
			}

			encoder := json.NewEncoder(conn)
			if err := encoder.Encode(resp); err != nil {
				return
			}
		}
	}
}

func (s *LocalRPCSimulator) Close() {
	if s.listener != nil {
		_ = s.listener.Close()
	}
	s.wg.Wait()
	_ = os.Remove(s.sockPath)
	fmt.Println("[SIMULATOR] Clean shutdown executed.")
}

// --------------------------------------------------------------------
// 3. INTEGRATION TEST
// --------------------------------------------------------------------

func TestIntegrationMevSim(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sockPath := "./pqr_mev_mock.sock"
	sim := NewLocalRPCSimulator(sockPath)
	if err := sim.Start(ctx); err != nil {
		t.Fatalf("failed to start simulator: %v", err)
	}
	defer sim.Close()

	adapter := NewMeshAdapter(30.0)
	watcher := NewDeltaWatcherRouter()
	hedger := NewHedger(adapter)

	go watcher.StartMonitoringLoop(ctx, hedger)

	mockHash := TxHashBytes{0x13, 0x37, 0xff}
	activeTrade := &TradePosition{
		ID:         "MEV-ARB-01",
		Asset:      "WETH",
		Size:       2.5,
		DeFiTxHash: mockHash,
		CeFiStatus: LegCompleted,
	}
	watcher.TrackTrade(activeTrade)

	fmt.Println("\n--- SCENARIO A: NOMINAL NETWORK OPERATIONS ---")
	watcher.IngestReceipt(TxReceipt{
		TxHash:  mockHash,
		Success: true,
	})
	time.Sleep(10 * time.Millisecond)

	fmt.Println("\n--- SCENARIO B: GAS SPIKE & API BLACKOUT CASCADES ---")
	failHash := TxHashBytes{0xde, 0xad, 0xbe, 0xef}
	poisonTrade := &TradePosition{
		ID:         "MEV-ARB-02",
		Asset:      "WETH",
		Size:       5.0,
		DeFiTxHash: failHash,
		CeFiStatus: LegCompleted,
	}
	watcher.TrackTrade(poisonTrade)

	sim.SetBaseFee(55)
	adapter.HealerHookModeB(ctx, 55.0)

	watcher.IngestReceipt(TxReceipt{
		TxHash:       failHash,
		Success:      false,
		RevertReason: "Slippage: Unprofitable Route Frame",
	})

	time.Sleep(50 * time.Millisecond)
	fmt.Println("[SYSTEM] Dispatching orphaned trade vector to defensive circuit...")
	_ = hedger.ExecuteEmergencyHedgeWithFallback(ctx, poisonTrade)

	time.Sleep(100 * time.Millisecond)
}

// --------------------------------------------------------------------
// 4. BENCHMARKS
// --------------------------------------------------------------------

func BenchmarkAtomicNonceTracker(b *testing.B) {
	var localNonce uint64 = 42

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		current := atomic.LoadUint64(&localNonce)
		atomic.AddUint64(&localNonce, 1)
		benchNonce = current
	}
}

func BenchmarkZeroAllocationFeeCap(b *testing.B) {
	baseFee := big.NewInt(25000000000)
	gasMultiplier := big.NewInt(130)
	twoBuffer := big.NewInt(2)
	tipCap := big.NewInt(2000000000)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		adjustedBaseFee := new(big.Int).Mul(baseFee, gasMultiplier)
		adjustedBaseFee.Div(adjustedBaseFee, big.NewInt(100))

		feeCap := new(big.Int).Mul(adjustedBaseFee, twoBuffer)
		feeCap.Add(feeCap, tipCap)

		if feeCap.Sign() == 0 {
			b.Fail()
		}
	}
}

func manualPackCalldata(methodSelector [4]byte, amountIn uint256Bytes, path []byte) []byte {
	buffer := make([]byte, 4+32+len(path))
	copy(buffer[0:4], methodSelector[:])
	copy(buffer[4:36], amountIn[:])
	copy(buffer[36:], path)
	return buffer
}

type uint256Bytes [32]byte

func BenchmarkManualBytePacker(b *testing.B) {
	selector := [4]byte{0xaa, 0xbb, 0xcc, 0xdd}
	amount := uint256Bytes{31: 0xff}
	mockPath := make([]byte, 43)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res := manualPackCalldata(selector, amount, mockPath)
		benchData = res
	}
}
