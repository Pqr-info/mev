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

	"pqr.info/mev/news"
)

type ReplaySegment struct {
	Epoch     string
	Events    []TemporalEvent
	StartTime time.Time
	EndTime   time.Time
}

type TimeMachineReplay struct {
	tme          *TemporalMemoryEngine
	router       TeleporterRouter
	newsProvider news.NewsProvider
}

func NewTimeMachineReplay(t *TemporalMemoryEngine, router TeleporterRouter) *TimeMachineReplay {
	return &TimeMachineReplay{
		tme:    t,
		router: router,
	}
}

func (tm *TimeMachineReplay) SetNewsProvider(provider news.NewsProvider) {
	tm.newsProvider = provider
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
	// Fallback to mock if requested
	if os.Getenv("USE_MOCK_BRAIN") == "true" {
		return "[MOCK GEMINI] Decision: SAFE. Rationale: Replay segment metrics fall within standard MEV risk bounds.", nil
	}

	model := os.Getenv("HEAVY_BRAIN_MODEL")
	if model == "" {
		model = "qwen2.5-coder:32b"
	}
	host := os.Getenv("HEAVY_BRAIN_HOST")
	if host == "" {
		host = "http://192.168.12.234:11434"
	}

	// Assuming Ollama or LM Studio OpenAI-compatible endpoint
	url := fmt.Sprintf("%s/v1/chat/completions", host)

	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type Payload struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
	}

	body := Payload{
		Model: model,
		Messages: []Message{
			{Role: "user", Content: query},
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

	// Authorization header if needed (e.g. for Gemini OpenAI compatibility API, or generic LM Studio auth)
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 30 * time.Second} // Larger timeout for local models
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("local brain returned status %s: %s", resp.Status, string(respBytes))
	}

	var result struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("empty response from predictive brain")
}

// ConsultFastBrain sends queries to the Gemma instance hosted on LM Studio (FAST_BRAIN_HOST).
func (tm *TimeMachineReplay) ConsultFastBrain(ctx context.Context, query string) (string, error) {
	if os.Getenv("USE_MOCK_BRAIN") == "true" {
		return "[MOCK GEMMA] Decision: SAFE. Rationale: Replay segment looks fine.", nil
	}

	model := os.Getenv("FAST_BRAIN_MODEL")
	if model == "" {
		model = "gemma2"
	}
	host := os.Getenv("FAST_BRAIN_HOST")
	if host == "" {
		host = "http://192.168.12.234:1234"
	}

	url := fmt.Sprintf("%s/v1/chat/completions", host)

	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type Payload struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
	}

	body := Payload{
		Model: model,
		Messages: []Message{
			{Role: "user", Content: query},
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

	client := &http.Client{Timeout: 30 * time.Second} 
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("fast brain returned status %s: %s", resp.Status, string(respBytes))
	}

	var result struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("empty response from fast brain")
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
	
	return tm.ConsultFastBrain(ctx, prompt)
}
