$HostName = "MAX"
$Username = Read-Host "Enter username for $HostName (leave blank if transferring via Windows Network Share / SMB)"

# Skill and plugin paths (the prerequisites)
$GemmaCobrowserSkill = "C:\Users\theal\.gemini\config\skills\gemma-cobrowser"
$ChromeDevToolsPlugin = "C:\Users\theal\.gemini\config\plugins\chrome-devtools-plugin"
$CopilotBridgeReceiver = "C:\Users\theal\.gemini\config\skills\copilot-bridge-receiver"
$CopilotSync = "C:\Users\theal\.gemini\config\skills\copilot-sync"

if ([string]::IsNullOrWhiteSpace($Username)) {
    Write-Host "Transferring via Windows Network Share..."
    $DestPath = "\\$HostName\C$\temp\gemma_skills"
    
    # Create directory on MAX
    New-Item -ItemType Directory -Force -Path $DestPath | Out-Null
    
    # Copy files
    Copy-Item -Path $GemmaCobrowserSkill -Destination $DestPath -Recurse -Force
    Copy-Item -Path $ChromeDevToolsPlugin -Destination $DestPath -Recurse -Force
    Copy-Item -Path $CopilotBridgeReceiver -Destination $DestPath -Recurse -Force
    Copy-Item -Path $CopilotSync -Destination $DestPath -Recurse -Force
    
    Write-Host "Successfully transferred gemma-cobrowser and prerequisites to $DestPath on $HostName!"
} else {
    Write-Host "Transferring via SCP..."
    scp -r $GemmaCobrowserSkill "$Username@$HostName`:~/"
    scp -r $ChromeDevToolsPlugin "$Username@$HostName`:~/"
    scp -r $CopilotBridgeReceiver "$Username@$HostName`:~/"
    scp -r $CopilotSync "$Username@$HostName`:~/"
    Write-Host "Successfully transferred gemma-cobrowser and prerequisites to ~/"
}
