import winrm
import sys

host = '192.168.12.234'
user = 'aellok'
password = 'm3sh'

session = winrm.Session(host, auth=(user, password), transport='ntlm', server_cert_validation='ignore')
r = session.run_cmd('wsl', ['-l', '-v'])
with open('wsl_list.txt', 'wb') as f:
    f.write(r.std_out)
