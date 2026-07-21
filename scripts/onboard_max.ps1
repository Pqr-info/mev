Param(
    [string]$GeminiBrainDir = "C:\Users\theal\.gemini",
    [string]$ShareName = "GeminiBrain",
    [string]$WslDistroName = "JetWebTimeMachineOS",
    [string]$MeshEndpointPort = "8000"   # vLLM primary
)

Write-Host "=== MAX Onboarding (Host) ==="

# 1. Create SMB share for .gemini
Write-Host "Creating SMB share for $GeminiBrainDir ..."
if (-Not (Test-Path $GeminiBrainDir)) {
    Write-Host "ERROR: Brain directory not found: $GeminiBrainDir" -ForegroundColor Red
    exit 1
}

# Remove existing share if present
$existingShare = Get-SmbShare -Name $ShareName -ErrorAction SilentlyContinue
if ($existingShare) {
    Write-Host "Existing share $ShareName found, removing..."
    Revoke-SmbShareAccess -Name $ShareName -AccountName "Everyone" -Force -ErrorAction SilentlyContinue
    Remove-SmbShare -Name $ShareName -Force
}

New-SmbShare -Name $ShareName -Path $GeminiBrainDir -FullAccess "Everyone" | Out-Null
Write-Host "SMB share $ShareName created for $GeminiBrainDir"

# 2. Set environment variables (system-level)
Write-Host "Setting environment variables..."
[Environment]::SetEnvironmentVariable("GEMINI_BRAIN_DIR", $GeminiBrainDir, "Machine")
[Environment]::SetEnvironmentVariable("JETWEB_ORGAN_ID", "MAX", "Machine")
[Environment]::SetEnvironmentVariable("JETWEB_TIME_MACHINE", "ENABLED", "Machine")
[Environment]::SetEnvironmentVariable("MAX_VLLM_PORT", $MeshEndpointPort, "Machine")
Write-Host "Environment variables set."

# 3. Verify WSL distro exists
Write-Host "Checking WSL distro $WslDistroName ..."
$wslList = wsl.exe --list --verbose
if ($wslList -notmatch $WslDistroName) {
    Write-Host "ERROR: WSL distro $WslDistroName not found." -ForegroundColor Red
    exit 1
}
Write-Host "WSL distro $WslDistroName found."

# 4. Copy bootstrap script to WSL and trigger it
Write-Host "Copying bootstrap script into WSL and fixing line endings..."
$bootstrapSource = Join-Path $PSScriptRoot "max_inference_bootstrap.sh"
# Use wslpath to get the WSL-compatible path for the source script
$wslSource = wsl.exe -d $WslDistroName -- wslpath -a $bootstrapSource
$bootstrapTarget = "/opt/max/max_inference_bootstrap.sh"

wsl.exe -d $WslDistroName -- bash -lc "mkdir -p /opt/max && tr -d '\r' < `"$wslSource`" > $bootstrapTarget && chmod +x $bootstrapTarget"

Write-Host "Starting WSL bootstrap for vLLM + Qwen3-Coder-30B + Gemma-4-e4b ..."
wsl.exe -d $WslDistroName -- bash -lc "$bootstrapTarget"

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: WSL bootstrap failed." -ForegroundColor Red
    exit 1
}

Write-Host "WSL bootstrap completed."

# 5. Verification: can MAX see transcript.jsonl via SMB/Mount
Write-Host "Verifying transcript.jsonl visibility..."
$transcriptPath = Join-Path $GeminiBrainDir "transcript.jsonl"
if (-Not (Test-Path $transcriptPath)) {
    Write-Host "WARNING: transcript.jsonl not found at $transcriptPath" -ForegroundColor Yellow
} else {
    Write-Host "transcript.jsonl found at $transcriptPath"
}

Write-Host "=== MAX Onboarding (Host) complete ==="
