package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SubstrateClient handles RPC interactions with the Substrate node.
// In production, this wraps github.com/centrifuge/go-substrate-rpc-client/v4
type SubstrateClient struct {
	RPCUrl string
	mu     sync.RWMutex
}

func NewSubstrateClient(rpcUrl string) *SubstrateClient {
	if rpcUrl == "" {
		rpcUrl = "ws://127.0.0.1:9944"
	}
	return &SubstrateClient{
		RPCUrl: rpcUrl,
	}
}

// MintNFT simulates broadcasting an extrinsic to `pallet_nfts` or `pallet_uniques`.
func (sc *SubstrateClient) MintNFT(ctx context.Context, collectionID uint32, metadataURI string, owner string) (string, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Mocking the Substrate RPC broadcast
	fmt.Printf("[SUBSTRATE] Broadcasting Mint extrinsic to %s. Collection: %d, URI: %s, Owner: %s\n", sc.RPCUrl, collectionID, metadataURI, owner)
	time.Sleep(200 * time.Millisecond) // Simulate network delay

	txHash := fmt.Sprintf("0x%x", time.Now().UnixNano())
	return txHash, nil
}

// ListNFT simulates calling the marketplace pallet to list an asset for sale.
func (sc *SubstrateClient) ListNFT(ctx context.Context, collectionID uint32, itemID uint32, price uint64) (string, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	fmt.Printf("[SUBSTRATE] Broadcasting List extrinsic. Item %d/%d for %d units.\n", collectionID, itemID, price)
	txHash := fmt.Sprintf("0x%x", time.Now().UnixNano())
	return txHash, nil
}

// BuyNFT simulates submitting a transaction to purchase a listed NFT.
func (sc *SubstrateClient) BuyNFT(ctx context.Context, collectionID uint32, itemID uint32, buyer string) (string, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	fmt.Printf("[SUBSTRATE] Broadcasting Buy extrinsic from buyer %s for Item %d/%d.\n", buyer, collectionID, itemID)
	txHash := fmt.Sprintf("0x%x", time.Now().UnixNano())
	return txHash, nil
}

// SubmitTransaction submits a raw MEV payload to the Substrate peer node.
func (sc *SubstrateClient) SubmitTransaction(ctx context.Context, payload []byte) (string, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	fmt.Printf("[SUBSTRATE] Routing MEV transaction payload to peer node at %s\n", sc.RPCUrl)
	time.Sleep(50 * time.Millisecond) // Lower latency insertion

	txHash := fmt.Sprintf("0x%x", time.Now().UnixNano())
	return txHash, nil
}
