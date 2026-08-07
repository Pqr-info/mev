$pass = ConvertTo-SecureString "m3sh" -AsPlainText -Force
$cred = New-Object System.Management.Automation.PSCredential ("LocalAdmin", $pass)
Start-Process powershell -Credential $cred -ArgumentList "-NoProfile -ExecutionPolicy Bypass -Command `"Set-Item WSMan:\localhost\Client\TrustedHosts -Value '192.168.12.234' -Force`"" -WindowStyle Hidden
