package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type TimeMachineServer struct {
	replay *TimeMachineReplay
	port   int
}

func NewTimeMachineServer(replay *TimeMachineReplay, port int) *TimeMachineServer {
	return &TimeMachineServer{
		replay: replay,
		port:   port,
	}
}

func (s *TimeMachineServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/replay", s.handleReplay)
	mux.HandleFunc("/news", s.handleNews)
	mux.HandleFunc("/inference/route", s.handleInferenceRoute)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("TimeMachine Server listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *TimeMachineServer) handleReplay(w http.ResponseWriter, r *http.Request) {
	// e.g., /replay?date=2023-06-01&speed=1.0
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		http.Error(w, "date parameter required", http.StatusBadRequest)
		return
	}

	speedStr := r.URL.Query().Get("speed")
	speed := 1.0
	if speedStr != "" {
		if v, err := strconv.ParseFloat(speedStr, 64); err == nil && v > 0 {
			speed = v
		}
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()
	
	// Build the replay segment for the given date/epoch
	// For simplicity, we pass dateStr as the epoch string to the replay engine
	segment := s.replay.BuildReplay(dateStr)
	
	if len(segment.Events) == 0 {
		fmt.Fprintf(w, "event: end\ndata: {\"message\": \"No market events found for %s\"}\n\n", dateStr)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Calculate base time to simulate speed
	baseSimTime := segment.Events[0].Timestamp
	baseRealTime := time.Now()

	for _, ev := range segment.Events {
		select {
		case <-ctx.Done():
			return
		default:
			// Calculate how long to wait
			simDelta := ev.Timestamp.Sub(baseSimTime)
			expectedRealDelta := time.Duration(float64(simDelta) / speed)
			elapsedReal := time.Since(baseRealTime)

			if expectedRealDelta > elapsedReal {
				time.Sleep(expectedRealDelta - elapsedReal)
			}

			// Also emit to the internal teleporter router
			s.replay.EmitReplayEvent(context.Background(), ev)

			// And write to the SSE stream
			b, _ := json.Marshal(ev)
			fmt.Fprintf(w, "data: %s\n\n", string(b))
			flusher.Flush()
		}
	}

	fmt.Fprintf(w, "event: end\ndata: {\"message\": \"Replay complete\"}\n\n")
	flusher.Flush()
}

func (s *TimeMachineServer) handleNews(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		http.Error(w, "date parameter required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	
	if s.replay.newsProvider == nil {
		fmt.Fprintf(w, "event: error\ndata: {\"message\": \"News provider not configured\"}\n\n")
		flusher.Flush()
		return
	}

	// Fetch news for the date
	layout := "2006-01-02"
	targetDate, err := time.Parse(layout, dateStr)
	if err != nil {
		http.Error(w, "invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	start := targetDate
	end := targetDate.Add(24 * time.Hour)

	newsItems, err := s.replay.newsProvider.GetNewsBetween(start, end)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"error\": \"%s\"}\n\n", err.Error())
		flusher.Flush()
		return
	}

	if len(newsItems) == 0 {
		fmt.Fprintf(w, "event: end\ndata: {\"message\": \"No news found for %s\"}\n\n", dateStr)
		flusher.Flush()
		return
	}

	// Simulate streaming out the news items evenly over a short period for demonstration
	for _, item := range newsItems {
		select {
		case <-ctx.Done():
			return
		default:
			b, _ := json.Marshal(item)
			fmt.Fprintf(w, "data: %s\n\n", string(b))
			flusher.Flush()
			time.Sleep(100 * time.Millisecond) // artificially space out output
		}
	}

	fmt.Fprintf(w, "event: end\ndata: {\"message\": \"News replay complete\"}\n\n")
	flusher.Flush()
}

func (s *TimeMachineServer) handleInferenceRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Query  string             `json:"query"`
		Sender map[string]float64 `json:"sender"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if payload.Query == "" {
		http.Error(w, "missing query", http.StatusBadRequest)
		return
	}

	// Route query to the local mothership engine
	result, err := s.replay.ConsultMothership(r.Context(), payload.Query)
	if err != nil {
		http.Error(w, fmt.Sprintf("inference failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"result": result,
	})
}
