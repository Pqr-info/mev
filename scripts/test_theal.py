import winrm
import sys

host = '192.168.12.234'
user = 'LocalAdmin'
password = 'm3sh'

cmd = "schtasks /create /tn TestWSL_theal /tr \"cmd.exe /c wsl -l -v > C:\\temp\\wsl_out.txt\" /ru theal /rp m3sh /sc once /st 00:00 /f"
print("Running command:", cmd)

try:
    session = winrm.Session(host, auth=(user, password), transport='ntlm', server_cert_validation='ignore')
    r = session.run_cmd('cmd', ['/c', cmd])
    print("STDOUT:", r.std_out.decode('utf-8', errors='ignore'))
    print("STDERR:", r.std_err.decode('utf-8', errors='ignore'))
    
    r2 = session.run_cmd('schtasks', ['/run', '/tn', 'TestWSL_theal'])
    print("Run STDOUT:", r2.std_out.decode('utf-8', errors='ignore'))
except Exception as e:
    print(e)
