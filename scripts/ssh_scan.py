import paramiko
import sys

host = '192.168.12.234'
ports = [22, 2222, 22222, 911, 5985, 5986, 11111]
user = 'aellok'
password = 'm3sh'

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())

for port in ports:
    try:
        print(f"Trying {host}:{port}...")
        client.connect(hostname=host, port=port, username=user, password=password, timeout=3)
        print(f"SUCCESS on port {port}!")
        stdin, stdout, stderr = client.exec_command("echo 'Hello from MAX'")
        print("STDOUT:", stdout.read().decode().strip())
        client.close()
        sys.exit(0)
    except Exception as e:
        print(f"Failed on port {port}: {e}")

