package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// OuroborosErrorHook intercepts global errors and creates SOS tickets
type OuroborosErrorHook struct{}

func (h OuroborosErrorHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level == zerolog.ErrorLevel || level == zerolog.FatalLevel {
		// Prevent recursive error logging if the ticket API itself fails
		if strings.Contains(msg, "Ouroboros RED FLAG") || strings.Contains(msg, "Ouroboros FAILED") {
			return
		}
		
		go func(errMsg string) {
			payload := map[string]string{
				"Queue":   "SOS",
				"Subject": "Ouroboros Auto-Heal Request",
				"Text":    errMsg,
			}
			data, _ := json.Marshal(payload)
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Post("http://localhost:8100/rt/REST/2.0/tickets", "application/json", bytes.NewReader(data))
			if err == nil {
				resp.Body.Close()
			}
		}(msg)
	}
}

// OuroborosMonitor is responsible for monitoring dynamic port ranges to ensure
// no unauthorized one-off processes are squatting without Substrate registration.
// It also acts as a self-healing agent to sanitize CRLF endings from execution scripts.
type OuroborosMonitor struct {
	stopCh chan struct{}
}

func NewOuroborosMonitor() *OuroborosMonitor {
	return &OuroborosMonitor{
		stopCh: make(chan struct{}),
	}
}

func (o *OuroborosMonitor) InstallLogHook() {
	log.Logger = log.Logger.Hook(OuroborosErrorHook{})
}

func (o *OuroborosMonitor) Start() {
	go o.monitorLoop()
}

func (o *OuroborosMonitor) Stop() {
	close(o.stopCh)
}

func (o *OuroborosMonitor) monitorLoop() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			o.sweepDynamicPorts()
			o.healScripts([]string{".", "C:\\temp_mev_transfer"})
		case <-o.stopCh:
			return
		}
	}
}

func (o *OuroborosMonitor) sweepDynamicPorts() {
	// Sweep 8080-8089
	for port := 8080; port <= 8089; port++ {
		target := fmt.Sprintf("localhost:%d", port)
		conn, err := net.DialTimeout("tcp", target, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			log.Error().
				Int("port", port).
				Msg("Ouroboros RED FLAG: Unauthorized process detected on dynamic port! Must register with Substrate.")
		}
	}
}

func (o *OuroborosMonitor) healScripts(targetDirs []string) {
	for _, dir := range targetDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil 
			}
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".sh") {
				contentBytes, err := os.ReadFile(path)
				if err != nil {
					return nil
				}
				content := string(contentBytes)
				if strings.Contains(content, "\r\n") {
					healedContent := strings.ReplaceAll(content, "\r\n", "\n")
					err = os.WriteFile(path, []byte(healedContent), info.Mode())
					if err == nil {
						log.Info().Str("file", path).Msg("Ouroboros HEALED: Stripped CRLF from script")
					} else {
						log.Warn().Err(err).Str("file", path).Msg("Ouroboros FAILED to heal script")
					}
				}
			}
			return nil
		})
		if err != nil {
			log.Warn().Err(err).Str("dir", dir).Msg("Ouroboros FAILED to walk directory")
		}
	}
}

