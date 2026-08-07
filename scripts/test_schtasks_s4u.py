import winrm
import sys

host = '192.168.12.234'
user = 'LocalAdmin'
password = 'm3sh'

ps_script = """
$ErrorActionPreference = 'Stop'
schtasks /create /tn "TestWSLS4U" /tr "cmd.exe /c wsl -l -v > C:\\temp\\wsl_s4u.txt" /ru "theal" /np /sc once /st 00:00 /f 2>&1
schtasks /run /tn "TestWSLS4U" 2>&1
Start-Sleep -Seconds 3
if (Test-Path C:\\temp\\wsl_s4u.txt) {
    Get-Content C:\\temp\\wsl_s4u.txt
} else {
    Write-Host "File not found"
}
"""

try:
    print(f"Executing on {host} as {user}...")
    session = winrm.Session(host, auth=(user, password), transport='ntlm', server_cert_validation='ignore')
    r = session.run_ps(ps_script)
    print("STDOUT:", r.std_out.decode('utf-16', errors='ignore'))
    print("STDERR:", r.std_err.decode('utf-16', errors='ignore'))
except Exception as e:
    print(e)
