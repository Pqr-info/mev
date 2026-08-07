#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
🏦 Charles Schwab Human-In-The-Loop (HITL) Playwright Scraper
=============================================================
Launches a visible Chrome browser window, navigates to the Schwab Client Portal,
and waits for the user to complete login + 2FA. Once logged in, it extracts 
the Accounts Summary DOM and saves it to the scratch folder to automatically
initialize the paper trading portfolio with real balances and stock positions.
"""

import os
import json
import asyncio
from playwright.async_api import async_playwright

# State paths matching schwab_paper_api.py
CURRENT_BRAIN_DIR = r"C:\Users\theal\.gemini\antigravity\brain\9d35ca65-e535-454e-86d3-b48e41fc4998"
LEGACY_BRAIN_DIR = r"C:\Users\theal\.gemini\antigravity-cli\brain\27cc7b87-864e-4692-aae4-745760bc1eb9"

SCHWAB_SUMMARY_URL = "https://client.schwab.com/app/accounts/summary"

def save_dom_data(dom_content):
    payload = {"dom": dom_content}
    
    # Save to both locations to ensure compatibility
    for directory in [CURRENT_BRAIN_DIR, LEGACY_BRAIN_DIR]:
        if not os.path.exists(directory):
            continue
        scratch_dir = os.path.join(directory, "scratch")
        os.makedirs(scratch_dir, exist_ok=True)
        file_path = os.path.join(scratch_dir, "schwab_summary_dom.json")
        
        with open(file_path, "w", encoding="utf-8") as f:
            json.dump(payload, f, indent=2)
        print(f"[+] Saved Schwab Summary DOM to: {file_path}")

async def run_hitl_scraper():
    async with async_playwright() as p:
        # Use local persistent Chrome profile to maintain cookies/sessions
        user_data_dir = os.path.join(os.environ.get("LOCALAPPDATA", ""), "Google", "Chrome", "User Data", "SchwabScraperSession")
        os.makedirs(user_data_dir, exist_ok=True)
        
        print("[*] Starting Human-In-The-Loop Playwright Session...")
        print("[*] Launching browser window...")
        
        browser = await p.chromium.launch_persistent_context(
            user_data_dir=user_data_dir,
            headless=False,
            channel="chrome",
            args=["--disable-blink-features=AutomationControlled"]
        )
        
        page = await browser.new_page()
        print(f"[*] Navigating to Schwab: {SCHWAB_SUMMARY_URL}")
        await page.goto(SCHWAB_SUMMARY_URL)
        
        print("\n" + "="*80)
        print("  HUMAN-IN-THE-LOOP INTERVENTION REQUIRED:")
        print("  Please log in and perform any 2FA/OTP directly in the browser window.")
        print("  The script will automatically detect when you reach the Accounts Summary.")
        print("="*80 + "\n")
        
        # Poll and check if the user has reached the summary page
        logged_in = False
        for attempt in range(120): # 120 attempts = 4 minutes timeout
            if page.is_closed():
                print("[-] Browser was closed before completing login.")
                break
                
            url = page.url
            if "accounts/summary" in url:
                print("[+] Accounts Summary page reached. Waiting for tables to load...")
                try:
                    # Wait for either position table rows or specific account labels to be visible
                    await page.wait_for_selector("table", timeout=15000)
                    logged_in = True
                    break
                except Exception as e:
                    print(f"[*] Still waiting for full page load: {e}")
                    
            await asyncio.sleep(2)
            
        if logged_in:
            print("[*] Fetching page DOM source...")
            # Capture outer html content
            dom_content = await page.content()
            
            # Save scraped DOM
            save_dom_data(dom_content)
            print("[+] HITL Scraper successfully completed!")
        else:
            print("[-] Scraper failed or timed out before reaching Schwab Accounts Summary.")
            
        print("[*] Closing browser session...")
        await browser.close()

if __name__ == "__main__":
    asyncio.run(run_hitl_scraper())
