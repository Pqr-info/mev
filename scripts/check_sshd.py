import winrm
import sys

host = '192.168.12.234'
user = 'aellok'
password = 'm3sh'

session = winrm.Session(host, auth=(user, password), transport='ntlm', server_cert_validation='ignore')
r1 = session.run_ps('Get-Service sshd')
r2 = session.run_ps('Get-NetFirewallRule -DisplayName "OpenSSH Server (sshd)"')
with open('sshd_check.txt', 'w', encoding='utf-8') as f:
    f.write(r1.std_out.decode('utf-8', errors='ignore'))
    f.write(r2.std_out.decode('utf-8', errors='ignore'))
