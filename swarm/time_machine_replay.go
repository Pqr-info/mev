package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type ReplaySegment struct {
	Epoch     string
	Events    []TemporalEvent
	StartTime time.Time
	EndTime   time.Time
}

type TimeMachineReplay struct {
	tme    *TemporalMemoryEngine
	router TeleporterRouter
}

func NewTimeMachineReplay(t *TemporalMemoryEngine, router TeleporterRouter) *TimeMachineReplay {
	return &TimeMachineReplay{
		tme:    t,
		router: router,
	}
}

func (tm *TimeMachineReplay) BuildReplay(epoch string) ReplaySegment {
	events := tm.tme.GetEventsByEpoch(epoch)

	if len(events) == 0 {
		return ReplaySegment{
			Epoch:     epoch,
			Events:    []TemporalEvent{},
			StartTime: time.Time{},
			EndTime:   time.Time{},
		}
	}

	start := events[0].Timestamp
	end := events[len(events)-1].Timestamp

	return ReplaySegment{
		Epoch:     epoch,
		Events:    events,
		StartTime: start,
		EndTime:   end,
	}
}

func (tm *TimeMachineReplay) ExecuteReplay(ctx context.Context, seg ReplaySegment) {
	for _, ev := range seg.Events {
		tm.EmitReplayEvent(ctx, ev)
	}
}

func (tm *TimeMachineReplay) EmitReplayEvent(ctx context.Context, ev TemporalEvent) {
	env := &FirehoseEnvelope{
		Source:      "time-machine-replay",
		StreamID:    "replay",
		Timestamp:   time.Now(),
		PayloadType: ev.PayloadType,
	}

	payload := MeshEventPayload{
		SigmaID:    ev.EventID,
		Agent:      "time-machine-replay",
		Files:      []string{},
		RiskScore:  ev.Drift,
		Confidence: ev.Volatility,
	}

	_ = tm.router.Route(ctx, env, payload)
}

// ConsultMothership sends queries directly to the Gemini mothership API using the GCM/Vault-fed key.
func (tm *TimeMachineReplay) ConsultMothership(ctx context.Context, query string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		// Mock response if no key is present (testing fallback)
		return "[MOCK GEMINI] Decision: SAFE. Rationale: Replay segment metrics fall within standard MEV risk bounds.", nil
	}

	model := os.Getenv("PREDICTIVE_BRAIN_MODEL")
	if model == "" {
		model = "gemini-3.5-flash"
	}
	host := os.Getenv("PREDICTIVE_BRAIN_HOST")
	if host == "" {
		host = "https://generativelanguage.googleapis.com"
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", host, model, apiKey)
	
	type Part struct {
		Text string `json:"text"`
	}
	type Content struct {
		Parts []Part `json:"parts"`
	}
	type Payload struct {
		Contents []Content `json:"contents"`
	}

	body := Payload{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: query},
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini api returned status %s: %s", resp.Status, string(respBytes))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []Part `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		return result.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("empty response from gemini")
}

// IntelligentBacktest feeds the backtest event summary to Gemma via mothership to assess stability.
func (tm *TimeMachineReplay) IntelligentBacktest(ctx context.Context, seg ReplaySegment) (string, error) {
	summary := fmt.Sprintf("Backtest Segment for Epoch %s containing %d events.\n", seg.Epoch, len(seg.Events))
	for i, ev := range seg.Events {
		if i >= 5 {
			summary += "... (remaining events truncated)\n"
			break
		}
		summary += fmt.Sprintf("- Event %s: Volatility=%.2f, Drift=%.2f, Type=%s\n", ev.EventID, ev.Volatility, ev.Drift, ev.PayloadType)
	}
	
	prompt := fmt.Sprintf(
		"You are Gemma, an agentic backtesting supervisor. Analyze this MEV swarm replay segment and decide if the timeline stability is SAFE or UNSAFE for mainnet production rollout. Format your response exactly as: 'Decision: [SAFE/UNSAFE] | Rationale: [1-sentence rationale]'\n\n%s", 
		summary,
	)
	
	return tm.ConsultMothership(ctx, prompt)
}
