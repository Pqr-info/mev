$users = @("LocalAdmin", ".\LocalAdmin", "Administrator", ".\Administrator", "$env:COMPUTERNAME\LocalAdmin", "theal", ".\theal")
$passText = "m3sh"
$pass = ConvertTo-SecureString $passText -AsPlainText -Force

foreach ($u in $users) {
    try {
        $cred = New-Object System.Management.Automation.PSCredential ($u, $pass)
        Start-Process powershell -Credential $cred -ArgumentList "-NoProfile -Command exit" -WindowStyle Hidden -ErrorAction Stop
        Write-Host "SUCCESS with user: $u"
        exit
    } catch {
        Write-Host "Failed with user: $u"
    }
}
