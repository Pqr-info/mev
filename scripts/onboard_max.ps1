<#
.SYNOPSIS
    Onboards "MAX", the high-powered mesh inference node (AI Max 395 CPU, 96GB GPU).
.DESCRIPTION
    This script configures the MAX node to act as an inference server and links its
    agent brain to the host's .gemini brain folder.
#>

param (
    [Parameter(Mandatory=$true)]
    [string]$HostIpOrName,

    [Parameter(Mandatory=$false)]
    [string]$ShareName = "gemini_brain"
)

Write-Host "Starting MAX onboarding sequence..." -ForegroundColor Cyan

# 1. Mount the Shared Brain
$RemoteShare = "\\$HostIpOrName\$ShareName"
$LocalBrainPath = "$env:USERPROFILE\.gemini"

Write-Host "Configuring shared brain from $RemoteShare..."
if (Test-Path $RemoteShare) {
    if (Test-Path $LocalBrainPath) {
        Write-Host "Backing up existing local brain..."
        Rename-Item -Path $LocalBrainPath -NewName ".gemini_backup_$(Get-Date -Format 'yyyyMMdd_HHmmss')"
    }
    
    Write-Host "Creating symlink to shared brain..."
    cmd /c mklink /D "$LocalBrainPath" "$RemoteShare" | Out-Null
    Write-Host "Brain synchronization complete." -ForegroundColor Green
} else {
    Write-Warning "Could not access remote share $RemoteShare. Make sure the folder is shared on the host."
    Write-Host "To share on host, run: New-SmbShare -Name 'gemini_brain' -Path 'C:\Users\theal\.gemini' -FullAccess Everyone" -ForegroundColor Yellow
}

# 2. Install Inference Engine (Ollama)
Write-Host "Checking for Ollama (GPU Inference Engine)..."
if (-not (Get-Command ollama -ErrorAction SilentlyContinue)) {
    Write-Host "Installing Ollama..."
    Invoke-WebRequest -Uri "https://ollama.com/download/OllamaSetup.exe" -OutFile "OllamaSetup.exe"
    Start-Process -FilePath ".\OllamaSetup.exe" -ArgumentList "/silent" -Wait
    Remove-Item ".\OllamaSetup.exe"
    Write-Host "Ollama installed." -ForegroundColor Green
} else {
    Write-Host "Ollama is already installed." -ForegroundColor Green
}

# 3. Pull default models for 96GB GPU
Write-Host "Pulling high-performance models for 96GB GPU..."
# Llama 3.1 70B requires ~40-50GB, easily fits in 96GB.
Write-Host "Pulling llama3.1:70b..."
ollama pull llama3.1:70b

Write-Host "MAX node successfully onboarded!" -ForegroundColor Cyan
Write-Host "Ensure that Ollama is running and accessible over the network if the host needs to query it."
Write-Host "To expose Ollama to the network, set OLLAMA_HOST=0.0.0.0 on this machine." -ForegroundColor Yellow
