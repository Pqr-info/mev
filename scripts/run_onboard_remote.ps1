$ErrorActionPreference = "Stop"

$hostIp = "192.168.12.234"
$user = "LocalAdmin"
$passText = "m3sh"

Write-Host "Creating credentials for $user ..."
$pass = ConvertTo-SecureString $passText -AsPlainText -Force
$cred = New-Object System.Management.Automation.PSCredential ($user, $pass)

Write-Host "Establishing PSSession to $hostIp ..."
$session = New-PSSession -ComputerName $hostIp -Credential $cred

try {
    Write-Host "Checking if C:\Users\theal\.gemini exists on MAX..."
    $brainExists = Invoke-Command -Session $session -ScriptBlock { Test-Path "C:\Users\theal\.gemini" }
    
    if (-not $brainExists) {
        Write-Host "Creating C:\Users\theal\.gemini on MAX..."
        Invoke-Command -Session $session -ScriptBlock { New-Item -ItemType Directory -Force -Path "C:\Users\theal\.gemini" | Out-Null }
    }

    $remoteDir = "C:\temp_onboard"
    Write-Host "Creating remote directory $remoteDir ..."
    Invoke-Command -Session $session -ScriptBlock { New-Item -ItemType Directory -Force -Path $remoteDir | Out-Null }

    Write-Host "Copying scripts to MAX..."
    Copy-Item -Path "d:\pqr.info\mev\scripts\onboard_max.ps1" -Destination $remoteDir -ToSession $session
    Copy-Item -Path "d:\pqr.info\mev\scripts\max_inference_bootstrap.sh" -Destination $remoteDir -ToSession $session

    Write-Host "Executing onboard_max.ps1 on MAX..."
    Invoke-Command -Session $session -ScriptBlock {
        Set-Location C:\temp_onboard
        .\onboard_max.ps1
    }
} finally {
    Write-Host "Cleaning up PSSession..."
    Remove-PSSession -Session $session
}
Write-Host "Remote execution finished."
