import winrm
import sys

host = '192.168.12.234'
user = 'aellok'
password = 'm3sh'

ps_script = """
if (-not (Test-Path C:\\temp)) { New-Item -ItemType Directory -Path C:\\temp }
Invoke-WebRequest -Uri "https://dl-cdn.alpinelinux.org/alpine/v3.18/releases/x86_64/alpine-minirootfs-3.18.0-x86_64.tar.gz" -OutFile "C:\\temp\\alpine.tar.gz"
Write-Host "Downloaded Alpine rootfs to C:\\temp\\alpine.tar.gz"
if (-not (Test-Path C:\\wsl\\JetWebTimeMachineOS)) { New-Item -ItemType Directory -Path C:\\wsl\\JetWebTimeMachineOS -Force }
wsl --import JetWebTimeMachineOS C:\\wsl\\JetWebTimeMachineOS C:\\temp\\alpine.tar.gz
Write-Host "Imported WSL distro JetWebTimeMachineOS"
wsl -l -v
"""

try:
    session = winrm.Session(host, auth=(user, password), transport='ntlm', server_cert_validation='ignore')
    r = session.run_ps(ps_script)
    print("STDOUT:", r.std_out.decode('utf-8', errors='ignore'))
    print("STDERR:", r.std_err.decode('utf-8', errors='ignore'))
except Exception as e:
    print(e)
