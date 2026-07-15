package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRTTicketingAPI(t *testing.T) {
	adapter := NewAdapter(Config{})
	mux := http.NewServeMux()
	mux.HandleFunc("/rt/REST/2.0/tickets", adapter.handleRTTickets)
	mux.HandleFunc("/rt/REST/2.0/tickets/", adapter.handleRTTicket)

	createReq := httptest.NewRequest(http.MethodPost, "/rt/REST/2.0/tickets", strings.NewReader(`{"Queue":"SOS","Subject":"deploy failed","Text":"Mesh adapter recovery loop"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	mux.ServeHTTP(createRR, createReq)

	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d: %s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var created map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	id, ok := created["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected created ticket id, got %#v", created["id"])
	}

	readReq := httptest.NewRequest(http.MethodGet, "/rt/REST/2.0/tickets/"+id, nil)
	readRR := httptest.NewRecorder()
	mux.ServeHTTP(readRR, readReq)

	if readRR.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, readRR.Code, readRR.Body.String())
	}

	if !strings.Contains(readRR.Body.String(), "deploy failed") {
		t.Fatalf("expected ticket body to contain subject, got %s", readRR.Body.String())
	}
}
