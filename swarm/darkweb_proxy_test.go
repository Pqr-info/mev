package main

import (
	"context"
	"strings"
	"testing"
)

func TestDarkWebProxyDialer(t *testing.T) {
	proxy, err := NewDarkWebProxy("127.0.0.1:9050")
	if err != nil {
		t.Fatalf("Failed to init proxy: %v", err)
	}

	_, err = proxy.HandleOnionRequest(context.Background(), "http://expyuz5drlpjerasi5adg56radqkeuvalnvzt32252a5c7w734d2c7ad.onion")
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "No connection could be made") {
			t.Logf("Expected connection refused because local Tor is not running: %v", err)
		} else {
			t.Fatalf("Unexpected error: %v", err)
		}
	}
}
