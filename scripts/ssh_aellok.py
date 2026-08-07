import paramiko

host = '192.168.12.234'
port = 22
user = 'aellok'
password = 'm3sh'

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())

try:
    print(f"Connecting to {host}:{port} as {user}...")
    client.connect(hostname=host, port=port, username=user, password=password, timeout=10)
    print("Connected! Testing Windows PowerShell...")
    stdin, stdout, stderr = client.exec_command("powershell.exe -Command \"Write-Output 'Hello from MAX'\"")
    
    print("STDOUT:", stdout.read().decode().strip())
    print("STDERR:", stderr.read().decode().strip())
except Exception as e:
    print("Error:", e)
finally:
    client.close()
