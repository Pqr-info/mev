import winrm

host = '192.168.12.234'
user = 'LocalAdmin'
password = 'm3sh'

ps_script = """
$user = Get-LocalUser -Name theal
$user | Format-List *
if ($user.PrincipalSource -eq "MicrosoftAccount") {
    Write-Host "It is a Microsoft Account!"
}
"""

try:
    session = winrm.Session(host, auth=(user, password), transport='ntlm', server_cert_validation='ignore')
    r = session.run_ps(ps_script)
    print("STDOUT:", r.std_out.decode('utf-8', errors='ignore'))
    print("STDERR:", r.std_err.decode('utf-8', errors='ignore'))
except Exception as e:
    print(e)
