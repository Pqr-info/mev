import winrm
import sys
import subprocess
import time
import urllib.request

host = '192.168.12.234'
user = 'aellok'
password = 'm3sh'

ps_script = """
$ErrorActionPreference = 'Stop'
$url = "https://cloud-images.ubuntu.com/wsl/jammy/current/ubuntu-jammy-wsl-amd64-ubuntu22.04lts.rootfs.tar.gz"
$dest = "C:\\temp\\mev\\ubuntu-jammy-wsl-amd64-ubuntu22.04lts.rootfs.tar.gz"
$distroName = "Ubuntu-22.04"
$installPath = "C:\\temp\\mev\\Ubuntu"

if (-not (Test-Path "C:\\temp\\mev")) { New-Item -ItemType Directory -Path "C:\\temp\\mev" -Force | Out-Null }
if (-not (Test-Path $installPath)) { New-Item -ItemType Directory -Path $installPath -Force | Out-Null }

if (-not (Test-Path $dest)) {
    Write-Host "Downloading Ubuntu rootfs..."
    Invoke-WebRequest -Uri $url -OutFile $dest
}

Write-Host "Unregistering existing $distroName if any..."
wsl --unregister $distroName 2>$null

Write-Host "Importing Ubuntu..."
wsl --import $distroName $installPath $dest

Write-Host "Checking if WSL works..."
wsl -d $distroName -- bash -c "cat /etc/os-release"
Write-Host "Done!"
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
