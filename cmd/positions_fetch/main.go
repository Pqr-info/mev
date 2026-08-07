package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// We run a Puppeteer script via gemma-cobrowser to scrape live positions
	script := `
        return await page.evaluate(() => {
            if (!window.location.href.includes('Accounts/Positions')) {
                return { error: 'Not on positions page' };
            }
            const positions = [];
            // Scaffolded DOM selectors for Schwab
            const rows = document.querySelectorAll('tr[data-event-name="PositionRow"]');
            rows.forEach(row => {
                const symbol = row.querySelector('.symbol-cell')?.innerText?.trim();
                const qty = row.querySelector('.qty-cell')?.innerText?.trim();
                if (symbol && qty) {
                    positions.push({symbol, qty: parseFloat(qty.replace(/,/g, ''))});
                }
            });
            return { positions };
        });
    `

	reqBody, _ := json.Marshal(map[string]string{"code": script})
	resp, err := http.Post("http://localhost:3456/api/run-script", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error connecting to gemma-cobrowser:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		os.Exit(1)
	}

	outPath := filepath.Join("output", "positions.json")
	if err := os.MkdirAll("output", 0755); err != nil {
		fmt.Println("Error creating output dir:", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, body, 0644); err != nil {
		fmt.Println("Error writing positions.json:", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully fetched positions and wrote to %s\n", outPath)
}
