import winrm
import sys
import subprocess
import time
import urllib.request
import os

host = '192.168.12.234'
user = 'aellok'
password = 'm3sh'
local_ip = '192.168.12.110'
zip_file = 'pqr.info.zip'
zip_path = f'D:\\{zip_file}'

if not os.path.exists(zip_path):
    print(f"Error: {zip_path} does not exist.")
    sys.exit(1)

print("Starting local HTTP server...")
httpd = subprocess.Popen(['python', '-m', 'http.server', '8888'], cwd=r'D:\\')
time.sleep(2)

try:
    urllib.request.urlopen(f"http://{local_ip}:8888/{zip_file}", timeout=3)
    print("HTTP server is accessible locally!")
except Exception as e:
    print(f"Failed to access local HTTP server: {e}")
    httpd.kill()
    sys.exit(1)

ps_script = f"""
$ErrorActionPreference = 'Stop'
$targetDir = 'C:\\temp\\pqr.info'
$zipPath = 'C:\\temp\\pqr.info.zip'

if (Test-Path $targetDir) {{
    Write-Host "Removing existing directory $targetDir..."
    Remove-Item -Recurse -Force $targetDir
}}
New-Item -ItemType Directory -Path $targetDir -Force | Out-Null

Write-Host "Downloading {zip_file} from host..."
Invoke-WebRequest -Uri "http://{local_ip}:8888/{zip_file}" -OutFile $zipPath

Write-Host "Extracting {zip_file} to $targetDir using .NET ZipFile..."
Add-Type -AssemblyName System.IO.Compression.FileSystem
[System.IO.Compression.ZipFile]::ExtractToDirectory($zipPath, $targetDir)

Write-Host "Cleaning up zip file..."
Remove-Item $zipPath -Force

Write-Host "Monorepo successfully deployed to $targetDir!"
"""

print(f"Deploying and executing on {host} via WinRM...")
try:
    session = winrm.Session(host, auth=(user, password), transport='ntlm', server_cert_validation='ignore')
    r = session.run_ps(ps_script)
    print("=== STDOUT ===")
    print(r.std_out.decode('utf-8', errors='ignore'))
    if r.status_code != 0:
        print("=== STDERR ===")
        print(r.std_err.decode('utf-8', errors='ignore'))
        sys.exit(r.status_code)
    print("SUCCESS!")
except Exception as e:
    print(f"Exception: {e}")
    sys.exit(1)
finally:
    httpd.kill()
