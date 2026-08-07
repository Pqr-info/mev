import winrm
import sys
import subprocess
import time
import urllib.request

host = '192.168.12.234'
user = 'aellok'
password = 'm3sh'
local_ip = '192.168.12.110'

print("Starting local HTTP server...")
httpd = subprocess.Popen(['python', '-m', 'http.server', '8888'], cwd=r'd:\pqr.info\mev')
time.sleep(2)

try:
    urllib.request.urlopen(f"http://{local_ip}:8888/scripts/onboard_max.ps1", timeout=3)
    print("HTTP server is accessible locally!")
except Exception as e:
    print(f"Failed to access local HTTP server: {e}")
    httpd.kill()
    sys.exit(1)

ps_script = f"""
$ErrorActionPreference = 'Stop'
$targetDir = 'C:\\temp\\mev\\scripts'
if (-not (Test-Path $targetDir)) {{ New-Item -ItemType Directory -Path $targetDir -Force | Out-Null }}

Write-Host "Downloading scripts from host..."
Invoke-WebRequest -Uri "http://{local_ip}:8888/scripts/onboard_max.ps1" -OutFile "$targetDir\\onboard_max.ps1"
Invoke-WebRequest -Uri "http://{local_ip}:8888/scripts/max_inference_bootstrap.sh" -OutFile "$targetDir\\max_inference_bootstrap.sh"

Write-Host "Files deployed to $targetDir. Executing onboard_max.ps1..."
Set-Location "C:\\temp\\mev"
powershell.exe -ExecutionPolicy Bypass -File "$targetDir\\onboard_max.ps1"
"""

print(f"Deploying and executing on {host} via WinRM...")
try:
    session = winrm.Session(host, auth=(user, password), transport='ntlm', server_cert_validation='ignore')
    r = session.run_ps(ps_script)
    print("=== STDOUT (Saving to out.txt) ===")
    with open('out.txt', 'wb') as f:
        f.write(r.std_out)
    if r.status_code != 0:
        print("=== STDERR (Saving to err.txt) ===")
        with open('err.txt', 'wb') as f:
            f.write(r.std_err)
        sys.exit(r.status_code)
    print("SUCCESS!")
except Exception as e:
    print(f"Exception: {e}")
    sys.exit(1)
finally:
    httpd.kill()
