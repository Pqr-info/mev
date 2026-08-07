import winrm
import sys

host = '192.168.12.234'
user = 'LocalAdmin'
password = 'm3sh'

ps_script = """
net localgroup administrators aellok /add
New-NetFirewallRule -Name "SSH_22" -DisplayName "SSH 22" -Direction Inbound -Action Allow -Protocol TCP -LocalPort 22 -ErrorAction SilentlyContinue
New-NetFirewallRule -Name "SSH_2222" -DisplayName "SSH 2222" -Direction Inbound -Action Allow -Protocol TCP -LocalPort 2222 -ErrorAction SilentlyContinue
Write-Host "aellok added to admins, and firewall rules created."
"""

try:
    print(f"Executing on {host} as {user}...")
    session = winrm.Session(host, auth=(user, password), transport='ntlm', server_cert_validation='ignore')
    r = session.run_ps(ps_script)
    print("STDOUT:", r.std_out.decode('utf-8', errors='ignore'))
    print("STDERR:", r.std_err.decode('utf-8', errors='ignore'))
except Exception as e:
    print(e)
