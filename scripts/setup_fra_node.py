#!/usr/bin/env python3
import sys
import os
import subprocess
import paramiko

HOST = "142.248.31.101"
USER = "root"
PASSWORD = "Element5!"
PUB_KEY_PATH = os.path.expanduser("~/.ssh/gemini-masterer-key.pem.pub")

def main():
    with open(PUB_KEY_PATH, "r", encoding="utf-8") as f:
        pub_key = f.read().strip()

    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect(HOST, port=22, username=USER, password=PASSWORD, timeout=10)

    sftp = client.open_sftp()
    try:
        sftp.mkdir(".ssh")
    except Exception:
        pass

    # Read existing authorized_keys
    auth_keys = ""
    try:
        with sftp.open(".ssh/authorized_keys", "r") as f:
            auth_keys = f.read().decode('utf-8')
    except Exception:
        pass

    if pub_key not in auth_keys:
        auth_keys += "\n" + pub_key + "\n"
        with sftp.open(".ssh/authorized_keys", "w") as f:
            f.write(auth_keys)

    sftp.chmod(".ssh", 0o700)
    sftp.chmod(".ssh/authorized_keys", 0o600)
    sftp.close()

    stdin, stdout, stderr = client.exec_command("hostname; uname -a; lscpu | grep 'Model name'")
    res = stdout.read().decode('utf-8', errors='ignore')
    client.close()

    sys.stdout.buffer.write(b"\n=== FRA.pqr.info (142.248.31.101) Integrated! ===\n")
    sys.stdout.buffer.write(res.encode('utf-8'))
    sys.stdout.buffer.write(b"\n[+] SSH Key deployed successfully!\n")

if __name__ == "__main__":
    main()
