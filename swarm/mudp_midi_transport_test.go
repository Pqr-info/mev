package main

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestMUDPMIDITransport(t *testing.T) {
	tme := NewTemporalMemoryEngine()
	transport := NewMUDPMIDITransport(9092, tme)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := transport.Start(ctx)
	if err != nil {
		t.Fatalf("failed to start MUDP-MIDI transport: %v", err)
	}
	defer transport.Stop()

	// Wait for listener socket boot
	time.Sleep(50 * time.Millisecond)

	// Create sender UDP socket
	rAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9092")
	if err != nil {
		t.Fatalf("failed to resolve UDP destination: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, rAddr)
	if err != nil {
		t.Fatalf("failed to dial UDP listener: %v", err)
	}
	defer conn.Close()

	// Send raw 3-byte MIDI payload: Status=0x90 (Note On/BUY), Data1=0x02 (AVGO), Data2=0x28 (Qty=40)
	payload := []byte{0x90, 0x02, 0x28}
	_, err = conn.Write(payload)
	if err != nil {
		t.Fatalf("failed to write UDP payload: %v", err)
	}

	// Wait for processing loop to digest packet
	time.Sleep(100 * time.Millisecond)

	events := tme.GetRecentEvents()
	if len(events) == 0 {
		t.Fatalf("expected 1 event appended to engine, got 0")
	}

	expectedType := "ORDER_BUY_AVGO"
	if events[0].PayloadType != expectedType {
		t.Errorf("expected payload type %s, got %s", expectedType, events[0].PayloadType)
	}
}
