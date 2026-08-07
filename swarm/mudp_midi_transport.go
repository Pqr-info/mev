package main

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// MIDIMessage defines a standard 3-byte MIDI message
type MIDIMessage struct {
	Status byte // 0x90 = BUY, 0x80 = SELL
	Data1  byte // Pitch (maps to Symbol registry index)
	Data2  byte // Velocity (maps to Quantity scalar)
}

// DecodeMIDI parses a 3-byte slice into a structural MIDI message.
func DecodeMIDI(b []byte) (MIDIMessage, error) {
	if len(b) < 3 {
		return MIDIMessage{}, fmt.Errorf("invalid midi frame size: expected 3 bytes, got %d", len(b))
	}
	return MIDIMessage{
		Status: b[0],
		Data1:  b[1],
		Data2:  b[2],
	}, nil
}

// EncodeMIDI serializes a MIDI message into a 3-byte slice.
func (m MIDIMessage) Encode() []byte {
	return []byte{m.Status, m.Data1, m.Data2}
}

// MUDPMIDITransport defines the UDP socket listener for MIDI events.
type MUDPMIDITransport struct {
	port       int
	conn       *net.UDPConn
	tme        *TemporalMemoryEngine
	registry   map[byte]string
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
}

// NewMUDPMIDITransport instantiates the transport layer.
func NewMUDPMIDITransport(port int, tme *TemporalMemoryEngine) *MUDPMIDITransport {
	// Initialize index-to-symbol mapping
	reg := map[byte]string{
		0x01: "TXN",
		0x02: "AVGO",
		0x03: "TSM",
		0x04: "THO",
		0x05: "FF",
		0x06: "OMCL",
	}

	return &MUDPMIDITransport{
		port:     port,
		tme:      tme,
		registry: reg,
	}
}

// Start opens the MUDP socket and spawns the listener loop.
func (t *MUDPMIDITransport) Start(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", t.port))
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	t.conn = conn

	subCtx, cancel := context.WithCancel(ctx)
	t.cancelFunc = cancel

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		t.listenLoop(subCtx)
	}()

	fmt.Printf("[MUDP-MIDI-TRANSPORT] Online. Listening on UDP port %d ...\n", t.port)
	return nil
}

// Stop closes connections and waits for goroutines to clean up.
func (t *MUDPMIDITransport) Stop() {
	if t.cancelFunc != nil {
		t.cancelFunc()
	}
	if t.conn != nil {
		_ = t.conn.Close()
	}
	t.wg.Wait()
}

func (t *MUDPMIDITransport) listenLoop(ctx context.Context) {
	buf := make([]byte, 64)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Read from socket
			_ = t.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			n, _, err := t.conn.ReadFrom(buf)
			if err != nil {
				continue
			}

			// Parse MIDI frames (each event is exactly 3 bytes)
			for i := 0; i+3 <= n; i += 3 {
				msg, err := DecodeMIDI(buf[i : i+3])
				if err != nil {
					continue
				}

				t.handleMessage(msg)
			}
		}
	}
}

func (t *MUDPMIDITransport) handleMessage(msg MIDIMessage) {
	// Status interpretation
	action := "UNKNOWN"
	if msg.Status == 0x90 {
		action = "BUY"
	} else if msg.Status == 0x80 {
		action = "SELL"
	}

	symbol, exists := t.registry[msg.Data1]
	if !exists {
		symbol = "UNKNOWN"
	}

	qty := float64(msg.Data2)

	fmt.Printf("[MUDP-MIDI] Received Low-Latency Event: Action=%s, Symbol=%s (Index: 0x%02X), Qty=%.0f (Raw Val: 0x%02X)\n",
		action, symbol, msg.Data1, qty, msg.Data2,
	)

	// Append event directly to the TemporalMemoryEngine (bypassing full TCP routing)
	if symbol != "UNKNOWN" && action != "UNKNOWN" {
		t.tme.AppendEvent(TemporalEvent{
			EventID:     fmt.Sprintf("midi-%d", time.Now().UnixNano()),
			PayloadType: fmt.Sprintf("ORDER_%s_%s", action, symbol),
			Drift:       0.01,
			Volatility:  0.0,
			Timestamp:   time.Now(),
		})
	}
}
