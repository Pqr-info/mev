import paramiko
import sys

host = '192.168.12.234'
port = 911
user = 'sos'
password = 'SovereignAdmin2026!'

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())

try:
    print(f"Connecting to {host}:{port} as {user}...")
    client.connect(hostname=host, port=port, username=user, password=password, timeout=10)
    print("Connected! Testing WSL interop by calling powershell.exe...")
    stdin, stdout, stderr = client.exec_command("powershell.exe -Command \"Write-Output 'Hello from Windows'\"")
    
    out = stdout.read().decode().strip()
    err = stderr.read().decode().strip()
    
    print("STDOUT:", out)
    print("STDERR:", err)
    
except Exception as e:
    print("Error:", e)
finally:
    client.close()
