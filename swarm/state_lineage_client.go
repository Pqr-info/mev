package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type FiveDAddress struct {
	Packed [16]byte `json:"packed"`
	Base27 string   `json:"base27"`
	X      uint32   `json:"X"`
	Y      uint32   `json:"Y"`
	Z      uint32   `json:"Z"`
	Phi    uint16   `json:"Phi"`
	Lambda uint16   `json:"Lambda"`
}

type StateSnapshot struct {
	Addr    FiveDAddress `json:"Addr"`
	Payload []byte       `json:"Payload"`
}

// LogStateTransition fires an async request to the canonical 5D-ASP Lineage service.
func LogStateTransition(addr FiveDAddress, payload []byte) {
	go func() {
		snap := StateSnapshot{
			Addr:    addr,
			Payload: payload,
		}
		data, err := json.Marshal(snap)
		if err != nil {
			fmt.Printf("[STATE-CLIENT] Failed to marshal snapshot: %v\n", err)
			return
		}

		req, err := http.NewRequest("POST", "http://localhost:9085/api/v1/state/advance", bytes.NewBuffer(data))
		if err != nil {
			fmt.Printf("[STATE-CLIENT] Failed to create request: %v\n", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("[STATE-CLIENT] Failed to submit state transition: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			fmt.Printf("[STATE-CLIENT] State Lineage service returned status: %d\n", resp.StatusCode)
		} else {
			fmt.Printf("[STATE-CLIENT] Successfully logged state transition to canonical 5D graph.\n")
		}
	}()
}
