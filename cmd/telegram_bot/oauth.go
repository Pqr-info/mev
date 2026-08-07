package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	clientID     = os.Getenv("GRAPH_CLIENT_ID")
	clientSecret = os.Getenv("GRAPH_CLIENT_SECRET")
	tenantID     = os.Getenv("GRAPH_TENANT_ID")
	redirectURL  = os.Getenv("GRAPH_REDIRECT_URL")
	authBase     = "https://login.microsoftonline.com"
	tokenPath    = "/oauth2/v2.0/token"
	authPath     = "/oauth2/v2.0/authorize"
	db           *sql.DB
)

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

type UserToken struct {
	TelegramID   int64
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

func (t *UserToken) IsExpired() bool {
	// Add 1 minute buffer
	return time.Now().Add(1 * time.Minute).After(t.ExpiresAt)
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./tokens.db")
	if err != nil {
		log.Fatalf("Failed to open sqlite db: %v", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS user_tokens (
		telegram_id    INTEGER PRIMARY KEY,
		access_token   TEXT NOT NULL,
		refresh_token  TEXT NOT NULL,
		expires_at     DATETIME NOT NULL
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
}

func handleOAuthLogin(w http.ResponseWriter, r *http.Request) {
	tgID := r.URL.Query().Get("tg_id")
	if tgID == "" {
		http.Error(w, "missing tg_id", http.StatusBadRequest)
		return
	}

	state := generateState(tgID)
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURL)
	q.Set("response_mode", "query")
	q.Set("scope", "offline_access Chat.Read Mail.Read Calendars.Read")
	q.Set("state", state)

	authURL := authBase + "/" + tenantID + authPath + "?" + q.Encode()
	http.Redirect(w, r, authURL, http.StatusFound)
}

func handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		http.Error(w, "missing code/state", http.StatusBadRequest)
		return
	}

	tgIDStr, err := parseState(state)
	if err != nil {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	var tgID int64
	fmt.Sscanf(tgIDStr, "%d", &tgID)

	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURL)

	tokenURL := authBase + "/" + tenantID + tokenPath
	resp, err := http.PostForm(tokenURL, form)
	if err != nil {
		http.Error(w, "token request failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var tok Token
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		http.Error(w, "decode token failed", http.StatusInternalServerError)
		return
	}

	if err := storeTokensForTelegramUser(tgID, tok.AccessToken, tok.RefreshToken, tok.ExpiresIn); err != nil {
		http.Error(w, "store token failed", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Authentication complete. You can now close this page and use /askcopilot in Telegram."))
}

func generateState(tgID string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	nonce := base64.RawURLEncoding.EncodeToString(b)
	return base64.RawURLEncoding.EncodeToString([]byte(tgID + "|" + nonce))
}

func parseState(s string) (string, error) {
	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	parts := string(data)
	for i := 0; i < len(parts); i++ {
		if parts[i] == '|' {
			return parts[:i], nil
		}
	}
	return "", fmt.Errorf("invalid state format")
}

func storeTokensForTelegramUser(tgID int64, accessToken, refreshToken string, expiresIn int) error {
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	query := `
	INSERT INTO user_tokens (telegram_id, access_token, refresh_token, expires_at) 
	VALUES (?, ?, ?, ?)
	ON CONFLICT(telegram_id) DO UPDATE SET 
		access_token = excluded.access_token,
		refresh_token = excluded.refresh_token,
		expires_at = excluded.expires_at;`
		
	_, err := db.Exec(query, tgID, accessToken, refreshToken, expiresAt)
	return err
}

func loadTokensForTelegramUser(tgID int64) (*UserToken, error) {
	var t UserToken
	t.TelegramID = tgID
	query := `SELECT access_token, refresh_token, expires_at FROM user_tokens WHERE telegram_id = ?`
	
	err := db.QueryRow(query, tgID).Scan(&t.AccessToken, &t.RefreshToken, &t.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not an error, just not found
		}
		return nil, err
	}
	return &t, nil
}

func refreshGraphToken(refreshToken string) (*Token, error) {
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("redirect_uri", redirectURL)

	tokenURL := authBase + "/" + tenantID + tokenPath
	resp, err := http.PostForm(tokenURL, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
	    return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var tok Token
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, err
	}
	return &tok, nil
}
