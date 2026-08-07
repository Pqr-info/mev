#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
🤖 Monica CLI - Command Line Interface for Monica AI API
======================================================
Retrieves the master API key from CockroachDB 'critical_data' and opens
an interactive shell to communicate with the Monica platform.

Usage:
  python monica_cli.py "What is the state of the mesh?"
  python monica_cli.py --interactive
"""

import sys
import os
import json
import argparse
import urllib.request
import urllib.error
import psycopg2

DB_URL = "postgresql://root@46.224.219.174:5196/antigravity?sslmode=disable"
DEFAULT_MODEL = "gpt-4o"  # Default model for Monica API completions

def get_monica_key():
    """
    Fetches the Monica API key directly from the critical_data table in CockroachDB.
    """
    try:
        conn = psycopg2.connect(DB_URL)
        cursor = conn.cursor()
        cursor.execute("""
            SELECT content FROM critical_data 
            WHERE title = '#germany' 
               OR content LIKE '%Platform.monica.im%'
            LIMIT 1;
        """)
        row = cursor.fetchone()
        cursor.close()
        conn.close()
        
        if not row:
            return None
            
        # Parse the key from the note body
        content = row[0]
        for line in content.split("\n"):
            if "API Key" in line or "sk-" in line:
                parts = line.split("Key")[-1].replace(":", "").strip()
                if "sk-" in parts:
                    return parts
                # Fallback splitting by spaces
                for token in line.split():
                    if token.startswith("sk-"):
                        return token
        return None
    except Exception as e:
        print(f"[WARN] Database connection failed: {e}. Checking environment variable...")
        return os.environ.get("MONICA_API_KEY")

def send_chat_completion(api_key, messages, model=DEFAULT_MODEL):
    """
    Sends chat completions request to Monica OpenAI-compatible endpoint.
    """
    url = "https://openapi.monica.im/v1/chat/completions"
    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json"
    }
    data = {
        "model": model,
        "messages": messages,
        "temperature": 0.7
    }
    
    req_body = json.dumps(data).encode("utf-8")
    req = urllib.request.Request(url, data=req_body, headers=headers, method="POST")
    
    try:
        with urllib.request.urlopen(req, timeout=30) as response:
            res_data = json.loads(response.read().decode("utf-8"))
            return res_data["choices"][0]["message"]["content"]
    except urllib.error.HTTPError as e:
        error_body = e.read().decode("utf-8")
        try:
            err_json = json.loads(error_body)
            msg = err_json.get("error", {}).get("message", error_body)
        except Exception:
            msg = error_body
        raise Exception(f"HTTP {e.code}: {msg}")
    except Exception as e:
        raise Exception(f"Network error: {e}")

def start_interactive_session(api_key, model):
    print("=================================================================")
    print("💬 Monica AI CLI - Interactive Session Started")
    print(f"Model: {model} | Type 'exit' or 'quit' to terminate")
    print("=================================================================")
    
    messages = [
        {"role": "system", "content": "You are a helpful, concise AI coding and systems assistant."}
    ]
    
    while True:
        try:
            user_input = input("\nYou > ").strip()
            if not user_input:
                continue
            if user_input.lower() in ["exit", "quit"]:
                print("Ending session. Goodbye!")
                break
                
            messages.append({"role": "user", "content": user_input})
            
            print("Monica > Thinking...", end="\r", flush=True)
            response = send_chat_completion(api_key, messages, model)
            
            # Clear "Thinking..." and print the response
            print(" " * 20, end="\r", flush=True)
            print(f"Monica > {response}")
            
            messages.append({"role": "assistant", "content": response})
        except KeyboardInterrupt:
            print("\nEnding session. Goodbye!")
            break
        except Exception as e:
            print(f"\n[ERROR] {e}")

def main():
    parser = argparse.ArgumentParser(description="Monica AI Platform Command Line Interface.")
    parser.add_argument("prompt", nargs="?", type=str, help="Single prompt request to send to Monica AI.")
    parser.add_argument("-i", "--interactive", action="store_true", help="Start an interactive chat session.")
    parser.add_argument("-m", "--model", type=str, default=DEFAULT_MODEL, help="Monica completion model to target.")
    
    args = parser.parse_args()
    
    api_key = get_monica_key()
    if not api_key:
        print("[CRITICAL] Could not locate Monica API Key in Database or environment.")
        sys.exit(1)
        
    if args.interactive or not args.prompt:
        start_interactive_session(api_key, args.model)
    else:
        messages = [{"role": "user", "content": args.prompt}]
        try:
            print("Monica > Thinking...", end="\r", flush=True)
            response = send_chat_completion(api_key, messages, args.model)
            print(" " * 20, end="\r", flush=True)
            print(response)
        except Exception as e:
            print(f"\n[ERROR] {e}")
            sys.exit(1)

if __name__ == "__main__":
    main()
