package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/net/proxy"
)

// DarkWebProxy defines a Tor SOCKS5 dialer for accessing .onion services.
type DarkWebProxy struct {
	Client *http.Client
}

// NewDarkWebProxy initializes a new DarkWebProxy routed through the specified SOCKS5 port.
// For example, "127.0.0.1:9050"
func NewDarkWebProxy(socks5Endpoint string) (*DarkWebProxy, error) {
	// Create a SOCKS5 dialer
	dialer, err := proxy.SOCKS5("tcp", socks5Endpoint, nil, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
	}

	// Create an HTTP transport that uses the SOCKS5 dialer
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (netConn net.Conn, err error) {
			// proxy.Dialer doesn't natively support DialContext in older x/net versions,
			// but Dial does work since we only use TCP.
			return dialer.Dial(network, addr)
		},
		ResponseHeaderTimeout: 30 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}

	return &DarkWebProxy{
		Client: client,
	}, nil
}

// QueueOnionScrape dials the target .onion URL, encrypts the response with the client's 5D address, and saves it.
// Returns a job ID representing the file.
func (dwp *DarkWebProxy) QueueOnionScrape(ctx context.Context, targetURL string, client5D string) (string, error) {
	// Generate unique Job ID based on timestamp and hash
	jobID := fmt.Sprintf("onion_%d", time.Now().UnixNano())

	// Run scraping and uploading asynchronously
	go func(id, url, addr string) {
		fmt.Printf("[DARKWEB-PROXY] Starting async scrape for %s (JobID: %s)\n", url, id)
		
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("[DARKWEB-PROXY] Failed to create request: %v\n", err)
			return
		}

		resp, err := dwp.Client.Do(req)
		if err != nil {
			fmt.Printf("[DARKWEB-PROXY] Failed to reach onion service %s: %v\n", url, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("[DARKWEB-PROXY] Onion service returned non-200 status: %d\n", resp.StatusCode)
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("[DARKWEB-PROXY] Failed to read response body: %v\n", err)
			return
		}

		// Zero-Knowledge Encryption
		// Derive AES-256 key from SHA-256 hash of the 5D address
		keyHash := sha256.Sum256([]byte(addr))
		block, err := aes.NewCipher(keyHash[:])
		if err != nil {
			fmt.Printf("[DARKWEB-PROXY] Failed to init AES cipher: %v\n", err)
			return
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			fmt.Printf("[DARKWEB-PROXY] Failed to init GCM: %v\n", err)
			return
		}

		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			fmt.Printf("[DARKWEB-PROXY] Failed to generate nonce: %v\n", err)
			return
		}

		ciphertext := gcm.Seal(nonce, nonce, body, nil)

		// Write to universal Hetzner SSH drive mount.
		var cacheDir string
		if runtime.GOOS == "windows" {
			cacheDir = `Z:\darkweb_cache`
		} else {
			cacheDir = `/mnt/zdrive/darkweb_cache`
		}
		os.MkdirAll(cacheDir, 0755)

		destFile := filepath.Join(cacheDir, id+".enc")
		if err := os.WriteFile(destFile, ciphertext, 0644); err != nil {
			fmt.Printf("[DARKWEB-PROXY] Failed to write encrypted file to %s: %v\n", destFile, err)
			return
		}

		fmt.Printf("[DARKWEB-PROXY] Successfully routed and encrypted payload %s to universal mount.\n", id)
		
		// Log to 5D state lineage
		eventPayload, _ := json.Marshal(map[string]string{
			"job_id": id,
			"target": url,
			"client": addr,
		})
		LogStateTransition(FiveDAddress{}, eventPayload)
	}(jobID, targetURL, client5D)

	return jobID, nil
}

// HandleOnionRequest handles a direct onion request and returns the body bytes.
func (dwp *DarkWebProxy) HandleOnionRequest(ctx context.Context, targetURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := dwp.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}


