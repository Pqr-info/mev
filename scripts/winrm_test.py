import winrm
import sys

host = '192.168.12.234'
users = ['LocalAdmin', 'aellok']
password = 'm3sh'

for user in users:
    try:
        print(f"Testing WinRM to {host} as {user}...")
        session = winrm.Session(host, auth=(user, password), transport='ntlm', server_cert_validation='ignore')
        r = session.run_cmd('ipconfig', ['/all'])
        if r.status_code == 0:
            print(f"SUCCESS with {user}!")
            print(r.std_out.decode('utf-8', errors='ignore'))
            sys.exit(0)
        else:
            print(f"Failed with {user}, status code {r.status_code}")
            print(r.std_err.decode('utf-8', errors='ignore'))
    except Exception as e:
        print(f"Exception with {user}: {e}")
