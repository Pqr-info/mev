package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gopkg.in/yaml.v3"
	
	"pqr.info/mev/safety"
)

type TelegramConfig struct {
	Token            string `yaml:"telegram_bot_token"`
	AuthorizedChatID int64  `yaml:"authorized_chat_id"`
	CopilotToken     string `yaml:"copilot_token"`
}

type PaperConfig struct {
	AlpacaBaseURL   string `yaml:"alpaca_base_url"`
	AlpacaAPIKeyID  string `yaml:"alpaca_api_key_id"`
	AlpacaAPISecret string `yaml:"alpaca_api_secret"`
	Mode            string `yaml:"mode"`
}

type StagedOrder struct {
	Symbol      string  `json:"symbol"`
	Qty         int     `json:"qty"`
	Side        string  `json:"side"`
	Type        string  `json:"type"`
	TimeInForce string  `json:"time_in_force"`
	Reason      string  `json:"reason"`
	Confidence  float64 `json:"confidence"`
}

type StagedOrdersPayload struct {
	GeneratedAt string        `json:"generated_at"`
	Orders      []StagedOrder `json:"orders"`
}

type SystemState struct {
	Paused bool `json:"paused"`
}

var lastNotified string

func main() {
	tgBytes, err := os.ReadFile(filepath.Join("config", "telegram.yaml"))
	if err != nil {
		log.Fatalf("Failed to read telegram config: %v", err)
	}
	var tgCfg TelegramConfig
	yaml.Unmarshal(tgBytes, &tgCfg)

	initDB()
	
	go func() {
		http.HandleFunc("/oauth/login", handleOAuthLogin)
		http.HandleFunc("/oauth/callback", handleOAuthCallback)
		log.Println("OAuth server listening on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("OAuth server failed: %v", err)
		}
	}()

	paperBytes, err := os.ReadFile(filepath.Join("config", "paper.yaml"))
	if err != nil {
		log.Fatalf("Failed to read paper config: %v", err)
	}
	var paperCfg PaperConfig
	yaml.Unmarshal(paperBytes, &paperCfg)

	bot, err := tgbotapi.NewBotAPI(tgCfg.Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	go pollStagedOrders(bot, tgCfg.AuthorizedChatID)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.Chat.ID != tgCfg.AuthorizedChatID {
				continue
			}
			handleCommand(bot, update.Message, tgCfg)
		} else if update.CallbackQuery != nil {
			if update.CallbackQuery.Message.Chat.ID != tgCfg.AuthorizedChatID {
				bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, "Unauthorized"))
				continue
			}
			handleCallback(bot, update.CallbackQuery, tgCfg, paperCfg)
		}
	}
}

func handleCommand(bot *tgbotapi.BotAPI, message *tgbotapi.Message, tgCfg TelegramConfig) {
	cmd := message.Command()
	if cmd == "" && message.Text != "" && message.Text[0] == '/' {
		cmd = message.Text[1:]
		// handle commands with args nicely if they used simple string matching
		parts := strings.SplitN(cmd, " ", 2)
		cmd = parts[0]
	}

	switch cmd {
	case "start", "menu":
		sendMenu(bot, message.Chat.ID)
	case "pnl":
		handlePnL(bot, message.Chat.ID)
	case "positions":
		handlePositions(bot, message.Chat.ID)
	case "status":
		handleStatus(bot, message.Chat.ID)
	case "pause":
		setSystemState(bot, message.Chat.ID, true)
	case "resume":
		safety.SetTemporalMetrics(0.0, 1.0) // Reset drift when resumed
		setSystemState(bot, message.Chat.ID, false)
	case "mock_drift":
		safety.SetTemporalMetrics(0.8, 0.4)
		msg := tgbotapi.NewMessage(message.Chat.ID, "🧪 *Simulated Temporal Drift Injected.* Organism will now perceive reality as fractured.")
		msg.ParseMode = "Markdown"
		bot.Send(msg)
	case "askcopilot":
		handleAskCopilot(bot, message.Chat.ID, message.Text, message.From.ID)
	case "scrape":
		handleScrape(bot, message.Chat.ID, message.Text)
	}
}

func sendMenu(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "🎛️ *MEV Command Center* \nSelect an option below or type a command:")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Positions", "cmd|positions"),
			tgbotapi.NewInlineKeyboardButtonData("📈 PnL", "cmd|pnl"),
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Status", "cmd|status"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛑 Pause", "cmd|pause"),
			tgbotapi.NewInlineKeyboardButtonData("▶️ Resume", "cmd|resume"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🤖 Ask Copilot", "cmd|askcopilot_hint"),
			tgbotapi.NewInlineKeyboardButtonData("🕸️ Scrape URL", "cmd|scrape_hint"),
		),
	)
	bot.Send(msg)
}

func handlePositions(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "⏳ Fetching live positions from Schwab...")
	bot.Send(msg)

	cmd := exec.Command("cmd", "/c", ".\\bin\\positions_fetch.exe")
	cmd.Dir = "."
	err := cmd.Run()

	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Failed to fetch positions (script error)"))
		return
	}

	b, err := os.ReadFile(filepath.Join("output", "positions.json"))
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Could not read positions file."))
		return
	}

	reply := fmt.Sprintf("📊 <b>Live Positions</b>\n<pre>%s</pre>", string(b))
	m := tgbotapi.NewMessage(chatID, reply)
	m.ParseMode = "HTML"
	bot.Send(m)
}

type RiskTelemetry struct {
	Timestamp      string  `json:"timestamp"`
	Equity         float64 `json:"equity"`
	DailyPnL       float64 `json:"daily_pnl"`
	DailyPnLPct    float64 `json:"daily_pnl_pct"`
	BuyingPower    float64 `json:"buying_power"`
	InitialMargin  float64 `json:"initial_margin"`
	Cash           float64 `json:"cash"`
}

func handlePnL(bot *tgbotapi.BotAPI, chatID int64) {
	b, err := os.ReadFile(filepath.Join("output", "risk_telemetry.json"))
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ PnL Telemetry not available yet."))
		return
	}
	
	var rt RiskTelemetry
	json.Unmarshal(b, &rt)
	
	emoji := "🟩"
	if rt.DailyPnL < 0 {
		emoji = "🟥"
	}

	text := fmt.Sprintf("📈 <b>Risk Telemetry</b>\n\n"+
		"Equity: $%.2f\n"+
		"Daily PnL: %s $%.2f (%.2f%%)\n\n"+
		"Buying Power: $%.2f\n"+
		"Cash: $%.2f",
		rt.Equity, emoji, rt.DailyPnL, rt.DailyPnLPct, rt.BuyingPower, rt.Cash)

	m := tgbotapi.NewMessage(chatID, text)
	m.ParseMode = "HTML"
	bot.Send(m)
}

func handleStatus(bot *tgbotapi.BotAPI, chatID int64) {
	state := getSystemState()
	statusText := "🟢 *RUNNING*"
	if state.Paused {
		statusText = "🛑 *PAUSED*"
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⚙️ *System Status*\n\nTrading Engine: %s\nCheck logs for gemma-cobrowser and Time Machine.", statusText))
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func setSystemState(bot *tgbotapi.BotAPI, chatID int64, paused bool) {
	state := SystemState{Paused: paused}
	b, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(filepath.Join("config", "state.json"), b, 0644)

	var text string
	if paused {
		text = "🛑 *Organism Paused*\nNew trades will not be generated."
	} else {
		text = "▶️ *Organism Resumed*\nTrading logic re-enabled."
	}
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func getSystemState() SystemState {
	b, err := os.ReadFile(filepath.Join("config", "state.json"))
	if err != nil {
		return SystemState{Paused: false}
	}
	var state SystemState
	json.Unmarshal(b, &state)
	return state
}

func pollStagedOrders(bot *tgbotapi.BotAPI, chatID int64) {
	stagedPath := filepath.Join("output", "staged_alpaca_orders.json")
	for {
		time.Sleep(5 * time.Second)

		state := getSystemState()
		if state.Paused {
			continue
		}

		// --- TEMPORAL SAFETY GATE ---
		if !safety.IsTemporallySafe() {
			stress, stability := safety.GetTemporalMetrics()
			
			// Auto-pause the organism
			setSystemState(bot, chatID, true)
			
			// Fire emergency alert
			alertText := fmt.Sprintf("🚨 <b>TEMPORAL DRIFT DETECTED</b>\n\nOrganism auto-paused to protect capital.\n\nTemporal Stress: <b>%.2f</b>\nConsensus Stability: <b>%.2f</b>\n\n<i>Resolve node drift and type /resume to re-enable.</i>", stress, stability)
			msg := tgbotapi.NewMessage(chatID, alertText)
			msg.ParseMode = "HTML"
			bot.Send(msg)
			
			// Eject the staged payload so we don't trip on it continuously
			os.Rename(stagedPath, filepath.Join("output", "staged_alpaca_orders.aborted_temporal_drift.json"))
			continue
		}
		// ----------------------------

		b, err := os.ReadFile(stagedPath)
		if err != nil {
			continue 
		}

		var payload StagedOrdersPayload
		if err := json.Unmarshal(b, &payload); err != nil || len(payload.Orders) == 0 {
			continue
		}

		if payload.GeneratedAt == lastNotified {
			continue
		}

		lastNotified = payload.GeneratedAt
		sendApprovalRequest(bot, chatID, payload)
	}
}

func sendApprovalRequest(bot *tgbotapi.BotAPI, chatID int64, payload StagedOrdersPayload) {
	var msgText strings.Builder
	msgText.WriteString(fmt.Sprintf("🤖 <b>New Trade Recommendation</b>\n<i>Generated at: %s</i>\n\n", payload.GeneratedAt))
	
	for i, o := range payload.Orders {
		msgText.WriteString(fmt.Sprintf("%d) <b>%s %d %s</b>\n", i+1, strings.ToUpper(o.Side), o.Qty, o.Symbol))
		msgText.WriteString(fmt.Sprintf("   Reason: %s\n", o.Reason))
		msgText.WriteString(fmt.Sprintf("   Confidence: %.2f\n\n", o.Confidence))
	}

	msg := tgbotapi.NewMessage(chatID, msgText.String())
	msg.ParseMode = "HTML"

	approveData := fmt.Sprintf("approve|%s", payload.GeneratedAt)
	rejectData := fmt.Sprintf("reject|%s", payload.GeneratedAt)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Approve", approveData),
			tgbotapi.NewInlineKeyboardButtonData("❌ Reject", rejectData),
		),
	)
	msg.ReplyMarkup = keyboard

	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send telegram message: %v", err)
	}
}

func handleCallback(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, tgCfg TelegramConfig, paperCfg PaperConfig) {
	parts := strings.Split(query.Data, "|")
	if len(parts) != 2 {
		return
	}
	action := parts[0]
	arg := parts[1]

	bot.Request(tgbotapi.NewCallback(query.ID, "Processing..."))

	if action == "cmd" {
		switch arg {
		case "positions":
			handlePositions(bot, query.Message.Chat.ID)
		case "pnl":
			handlePnL(bot, query.Message.Chat.ID)
		case "status":
			handleStatus(bot, query.Message.Chat.ID)
		case "pause":
			setSystemState(bot, query.Message.Chat.ID, true)
		case "resume":
			setSystemState(bot, query.Message.Chat.ID, false)
		case "askcopilot_hint":
			bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "To use Copilot, send a message like:\n/askcopilot What's the latest news on Microsoft?"))
		case "scrape_hint":
			bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "To scrape a URL, send a message like:\n/scrape https://example.com"))
		}
		return
	}

	stagedPath := filepath.Join("output", "staged_alpaca_orders.json")
	b, err := os.ReadFile(stagedPath)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "File not found or already processed"))
		return
	}

	var payload StagedOrdersPayload
	json.Unmarshal(b, &payload)

	if payload.GeneratedAt != arg {
		bot.Send(tgbotapi.NewMessage(query.Message.Chat.ID, "Mismatch in generation ID"))
		return
	}

	if action == "approve" {
		executeAlpacaOrders(payload, paperCfg)
		os.Rename(stagedPath, filepath.Join("output", "staged_alpaca_orders.executed.json"))
		
		edit := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, query.Message.Text+"\n\n✅ <b>EXECUTED ON ALPACA</b>")
		edit.ParseMode = "HTML"
		bot.Send(edit)
	} else if action == "reject" {
		os.Rename(stagedPath, filepath.Join("output", "staged_alpaca_orders.rejected.json"))

		edit := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, query.Message.Text+"\n\n❌ <b>REJECTED BY USER</b>")
		edit.ParseMode = "HTML"
		bot.Send(edit)
	}
}

func executeAlpacaOrders(payload StagedOrdersPayload, cfg PaperConfig) {
	for _, o := range payload.Orders {
		apiPayload := map[string]interface{}{
			"symbol":        o.Symbol,
			"qty":           o.Qty,
			"side":          o.Side,
			"type":          o.Type,
			"time_in_force": o.TimeInForce,
		}
		body, _ := json.Marshal(apiPayload)
		req, _ := http.NewRequest("POST", cfg.AlpacaBaseURL+"/v2/orders", bytes.NewBuffer(body))
		req.Header.Set("APCA-API-KEY-ID", cfg.AlpacaAPIKeyID)
		req.Header.Set("APCA-API-SECRET-KEY", cfg.AlpacaAPISecret)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Failed to execute %s: %v", o.Symbol, err)
			continue
		}
		resp.Body.Close()
	}
}

func handleAskCopilot(bot *tgbotapi.BotAPI, chatID int64, text string, tgID int64) {
	tokens, err := loadTokensForTelegramUser(tgID)
	if err != nil || tokens == nil {
		loginURL := fmt.Sprintf("https://your-domain.com/oauth/login?tg_id=%d", tgID)
		if redirectURL != "" {
			u, _ := url.Parse(redirectURL)
			loginURL = fmt.Sprintf("%s://%s/oauth/login?tg_id=%d", u.Scheme, u.Host, tgID)
		}
		bot.Send(tgbotapi.NewMessage(chatID, "To use Copilot, please authenticate your Microsoft account here:\n"+loginURL))
		return
	}

	accessToken := tokens.AccessToken
	if tokens.IsExpired() {
		bot.Send(tgbotapi.NewMessage(chatID, "⏳ Refreshing token..."))
		newTok, err := refreshGraphToken(tokens.RefreshToken)
		if err != nil {
			loginURL := fmt.Sprintf("https://your-domain.com/oauth/login?tg_id=%d", tgID)
			if redirectURL != "" {
				u, _ := url.Parse(redirectURL)
				loginURL = fmt.Sprintf("%s://%s/oauth/login?tg_id=%d", u.Scheme, u.Host, tgID)
			}
			bot.Send(tgbotapi.NewMessage(chatID, "Your session expired and couldn't be refreshed. Please re-authenticate:\n"+loginURL))
			return
		}
		accessToken = newTok.AccessToken
		storeTokensForTelegramUser(tgID, newTok.AccessToken, newTok.RefreshToken, newTok.ExpiresIn)
	}
	
	parts := strings.SplitN(text, " ", 2)
	if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Please provide a prompt. Example: /askcopilot What's the latest news?"))
		return
	}
	prompt := strings.TrimSpace(parts[1])
	
	msg := tgbotapi.NewMessage(chatID, "⏳ Asking Microsoft 365 Copilot...")
	processingMsg, _ := bot.Send(msg)
	
	answer, err := askCopilotAPI(prompt, accessToken)
	
	var edit tgbotapi.EditMessageTextConfig
	if err != nil {
		edit = tgbotapi.NewEditMessageText(chatID, processingMsg.MessageID, fmt.Sprintf("❌ Error from Copilot API: %v", err))
	} else {
		// Telegram markdown needs some escaping or HTML might be better.
		edit = tgbotapi.NewEditMessageText(chatID, processingMsg.MessageID, "🤖 <b>Copilot:</b>\n"+answer)
		edit.ParseMode = "HTML"
	}
	bot.Send(edit)
}

func askCopilotAPI(prompt string, token string) (string, error) {
	// 1. Start conversation
	req, _ := http.NewRequest("POST", "https://graph.microsoft.com/beta/copilot/conversations", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("failed to create conversation, status: %d", resp.StatusCode)
	}
	
	var convRes map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&convRes); err != nil {
		return "", fmt.Errorf("failed to decode conversation response: %v", err)
	}
	
	convID, ok := convRes["id"].(string)
	if !ok {
		convID, ok = convRes["conversationId"].(string)
		if !ok {
			return "", fmt.Errorf("could not find conversation ID in response")
		}
	}
	
	// 2. Send message
	chatPayload := map[string]interface{}{
		"message": map[string]interface{}{
			"text": prompt,
		},
	}
	chatBody, _ := json.Marshal(chatPayload)
	req2, _ := http.NewRequest("POST", fmt.Sprintf("https://graph.microsoft.com/beta/copilot/conversations/%s/chat", convID), bytes.NewBuffer(chatBody))
	req2.Header.Set("Authorization", "Bearer "+token)
	req2.Header.Set("Content-Type", "application/json")
	
	resp2, err := client.Do(req2)
	if err != nil {
		return "", err
	}
	defer resp2.Body.Close()
	
	if resp2.StatusCode < 200 || resp2.StatusCode >= 300 {
		return "", fmt.Errorf("failed to send chat, status: %d", resp2.StatusCode)
	}
	
	var chatRes map[string]interface{}
	if err := json.NewDecoder(resp2.Body).Decode(&chatRes); err != nil {
		return "", fmt.Errorf("failed to decode chat response: %v", err)
	}
	
	// Graph API reply structure
	if msg, ok := chatRes["message"].(map[string]interface{}); ok {
		if content, ok := msg["content"].(string); ok {
			return content, nil
		}
		if text, ok := msg["text"].(string); ok {
			return text, nil
		}
	}
	
	dump, _ := json.MarshalIndent(chatRes, "", "  ")
	return string(dump), nil
}

func handleScrape(bot *tgbotapi.BotAPI, chatID int64, text string) {
	parts := strings.SplitN(text, " ", 2)
	if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Please provide a URL. Example: /scrape https://example.com"))
		return
	}
	urlStr := strings.TrimSpace(parts[1])
	
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("⏳ Scraping %s ...", urlStr))
	processingMsg, _ := bot.Send(msg)
	
	content, err := scrapeURL(urlStr)
	
	if err != nil {
		edit := tgbotapi.NewEditMessageText(chatID, processingMsg.MessageID, fmt.Sprintf("❌ Error scraping URL: %v", err))
		bot.Send(edit)
		return
	}
	
	// Truncate if necessary (Telegram limit is 4096)
	if len(content) > 4000 {
		content = content[:4000] + "\n\n...[TRUNCATED]"
	}
	
	edit := tgbotapi.NewEditMessageText(chatID, processingMsg.MessageID, "🕸️ <b>Scraped Content:</b>\n<pre>"+content+"</pre>")
	edit.ParseMode = "HTML"
	bot.Send(edit)
}

func scrapeURL(urlStr string) (string, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	body := string(bodyBytes)
	
	// Basic regex to strip out script/style tags and then html tags
	reScript := regexp.MustCompile(`(?is)<(script|style).*?>.*?</\1>`)
	body = reScript.ReplaceAllString(body, "")
	
	reTags := regexp.MustCompile(`(?is)<.*?>`)
	body = reTags.ReplaceAllString(body, " ")
	
	// Collapse multiple spaces
	reSpaces := regexp.MustCompile(`\s+`)
	body = reSpaces.ReplaceAllString(body, " ")
	
	return strings.TrimSpace(body), nil
}
