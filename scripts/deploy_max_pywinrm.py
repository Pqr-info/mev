import winrm
import base64
import sys
import os

host = '192.168.12.234'
user = 'aellok'
password = 'm3sh'

# Read local scripts
with open(r'd:\pqr.info\mev\scripts\onboard_max.ps1', 'rb') as f:
    onboard_b64 = base64.b64encode(f.read()).decode('ascii')
    
with open(r'd:\pqr.info\mev\scripts\max_inference_bootstrap.sh', 'rb') as f:
    bootstrap_b64 = base64.b64encode(f.read()).decode('ascii')

# PowerShell command to decode and run
ps_script = f"""
$ErrorActionPreference = 'Stop'
$targetDir = 'C:\\temp\\mev\\scripts'
if (-not (Test-Path $targetDir)) {{ New-Item -ItemType Directory -Path $targetDir -Force | Out-Null }}

$onboardBytes = [System.Convert]::FromBase64String('{onboard_b64}')
[System.IO.File]::WriteAllBytes("$targetDir\\onboard_max.ps1", $onboardBytes)

$bootstrapBytes = [System.Convert]::FromBase64String('{bootstrap_b64}')
[System.IO.File]::WriteAllBytes("$targetDir\\max_inference_bootstrap.sh", $bootstrapBytes)

Write-Host "Files deployed to $targetDir. Executing onboard_max.ps1..."
Set-Location "C:\\temp\\mev"
& "$targetDir\\onboard_max.ps1"
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
