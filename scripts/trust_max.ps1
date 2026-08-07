$ErrorActionPreference = "Stop"

$user = "LocalAdmin"
$passText = "m3sh"

Write-Host "Creating credentials for local elevation..."
$pass = ConvertTo-SecureString $passText -AsPlainText -Force
$cred = New-Object System.Management.Automation.PSCredential ($user, $pass)

try {
    Write-Host "Invoking command on localhost to set TrustedHosts..."
    Invoke-Command -ComputerName localhost -Credential $cred -ScriptBlock {
        Set-Item WSMan:\localhost\Client\TrustedHosts -Value "192.168.12.234" -Force
        Get-Item WSMan:\localhost\Client\TrustedHosts
    }
    Write-Host "TrustedHosts updated successfully."
} catch {
    Write-Host "Error updating TrustedHosts: $_" -ForegroundColor Red
}
