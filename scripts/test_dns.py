import winrm
import sys

host = '192.168.12.234'
user = 'aellok'
password = 'm3sh'

session = winrm.Session(host, auth=(user, password), transport='ntlm', server_cert_validation='ignore')
r = session.run_cmd('wsl', ['-d', 'JetWebTimeMachineOS', '--', '/bin/ash', '-lc', 'apk update'])
with open('out.txt', 'wb') as f:
    f.write(r.std_out)
with open('err.txt', 'wb') as f:
    f.write(r.std_err)
