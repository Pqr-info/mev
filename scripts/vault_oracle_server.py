#!/usr/bin/env python3
"""
🔐 Vault Oracle Server Manager
==============================
CLI interface to store and retrieve the Oracle Linux Server SSH Key (`id_sovereign`)
in HashiCorp Vault and execute seamless authenticated SSH logins to 46.224.219.174.

Usage:
  python vault_oracle_server.py save --key-file C:/Users/theal/.ssh/id_sovereign
  python vault_oracle_server.py fetch
  python vault_oracle_server.py ssh --user opc
"""

import sys
import os
import json
import subprocess
import urllib.request
import urllib.error

VAULT_ADDR = os.environ.get("VAULT_ADDR", "http://localhost:8200").rstrip("/")
VAULT_TOKEN = os.environ.get("VAULT_TOKEN", "root")
SECRET_PATH = "/v1/secret/data/oracle_server/id_sovereign"
LOCAL_KEY_PATH = os.path.expanduser("~/.ssh/id_sovereign")
ORACLE_HOST = os.environ.get("ORACLE_HOST", "46.224.219.174")
DEFAULT_USER = os.environ.get("ORACLE_USER", "opc")

def vault_request(endpoint: str, method: str = "GET", data: dict = None):
    url = f"{VAULT_ADDR}{endpoint}"
    headers = {
        "X-Vault-Token": VAULT_TOKEN,
        "Content-Type": "application/json"
    }
    payload = json.dumps(data).encode("utf-8") if data else None
    req = urllib.request.Request(url, data=payload, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=10) as response:
            return json.loads(response.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        body = e.read().decode("utf-8")
        raise Exception(f"Vault HTTP {e.code}: {body}")
    except Exception as e:
        raise Exception(f"Vault Connection Error ({url}): {e}")

def save_key_to_vault(key_content: str):
    print(f"[*] Storing id_sovereign key into HashiCorp Vault at {SECRET_PATH}...")
    payload = {
        "data": {
            "id_sovereign": key_content,
            "host": ORACLE_HOST,
            "default_user": DEFAULT_USER,
            "updated_at": subprocess.check_output(["date", "-u"] if os.name != 'nt' else ["powershell", "Get-Date -Format o"]).decode().strip()
        }
    }
    res = vault_request(SECRET_PATH, method="POST", data=payload)
    print("[+] Key successfully saved to HashiCorp Vault!")
    return res

def fetch_key_from_vault() -> str:
    print(f"[*] Retrieving id_sovereign key from HashiCorp Vault ({SECRET_PATH})...")
    res = vault_request(SECRET_PATH, method="GET")
    key_content = res.get("data", {}).get("data", {}).get("id_sovereign")
    if not key_content:
        raise Exception("id_sovereign key not found in Vault payload.")
    
    # Ensure local directory exists
    os.makedirs(os.path.dirname(LOCAL_KEY_PATH), exist_ok=True)
    with open(LOCAL_KEY_PATH, "w", encoding="utf-8") as f:
        f.write(key_content)
    
    if os.name != 'nt':
        os.chmod(LOCAL_KEY_PATH, 0o600)
    print(f"[+] Key saved locally to {LOCAL_KEY_PATH} (permissions set to 0600)")
    return LOCAL_KEY_PATH

def connect_ssh(user: str = DEFAULT_USER):
    if not os.path.exists(LOCAL_KEY_PATH):
        fetch_key_from_vault()
    
    cmd = ["ssh", "-i", LOCAL_KEY_PATH, "-o", "StrictHostKeyChecking=no", f"{user}@{ORACLE_HOST}"]
    print(f"[*] Executing SSH login to {user}@{ORACLE_HOST}...")
    subprocess.run(cmd)

def main():
    if len(sys.argv) < 2:
        print("Usage:")
        print("  python vault_oracle_server.py save <key_text_or_filepath>")
        print("  python vault_oracle_server.py fetch")
        print("  python vault_oracle_server.py ssh [user]")
        sys.exit(1)

    action = sys.argv[1].lower()
    if action == "save":
        if len(sys.argv) < 3:
            print("Error: Specify key text or filepath.")
            sys.exit(1)
        arg = sys.argv[2]
        if os.path.exists(arg):
            with open(arg, "r", encoding="utf-8") as f:
                content = f.read()
        else:
            content = arg
        save_key_to_vault(content)

    elif action == "fetch":
        fetch_key_from_vault()

    elif action == "ssh":
        user = sys.argv[2] if len(sys.argv) > 2 else DEFAULT_USER
        connect_ssh(user)
    else:
        print(f"Unknown action: {action}")
        sys.exit(1)

if __name__ == "__main__":
    main()
