# Spark-OS (SOS) Windows PowerShell Deployment Script
$ErrorActionPreference = "Stop"

$VPS_IP = "46.224.219.174"
$REMOTE_OPT = "/opt/sos"
$REMOTE_ETC = "/etc/sos"
$TEMP_DEPLOY = "$env:TEMP\pqr-deploy"

# Clean local temporary deploy directory
if (Test-Path $TEMP_DEPLOY) {
    Remove-Item -Recurse -Force $TEMP_DEPLOY
}
New-Item -ItemType Directory -Path "$TEMP_DEPLOY\mev" | Out-Null

Write-Host "=== Preparing and copying source tree ==="
# Use robocopy to sync source files quickly excluding build directories
$exclude = @(".git", ".github", "node_modules", ".venv", "__pycache__", "target")
robocopy "d:\pqr.info\mev" "$TEMP_DEPLOY\mev" /MIR /XD $exclude /R:1 /W:1 | Out-Null
Remove-Item -Recurse -Force "$TEMP_DEPLOY\mev\bin" -ErrorAction SilentlyContinue

# Copy jetweb-time-machine if exists
if (Test-Path "d:\pqr.info\jetweb-time-machine") {
    New-Item -ItemType Directory -Path "$TEMP_DEPLOY\jetweb-time-machine" | Out-Null
    robocopy "d:\pqr.info\jetweb-time-machine" "$TEMP_DEPLOY\jetweb-time-machine" /MIR /XD $exclude /R:1 /W:1 | Out-Null
    Remove-Item -Recurse -Force "$TEMP_DEPLOY\jetweb-time-machine\bin" -ErrorAction SilentlyContinue
}

# Copy substrate-node-template if exists
if (Test-Path "d:\pqr.info\substrate-node-template") {
    New-Item -ItemType Directory -Path "$TEMP_DEPLOY\substrate-node-template" | Out-Null
    robocopy "d:\pqr.info\substrate-node-template" "$TEMP_DEPLOY\substrate-node-template" /MIR /XD $exclude /R:1 /W:1 | Out-Null
    Remove-Item -Recurse -Force "$TEMP_DEPLOY\substrate-node-template\target" -ErrorAction SilentlyContinue
}

Write-Host "=== [PQRL0] Preparing Server Directory Topology ==="
$null | ssh -T -o StrictHostKeyChecking=no root@$VPS_IP "mkdir -p /opt/sos/mev /opt/sos/jetweb-time-machine /opt/sos/substrate-node-template /etc/sos /var/sos"

Write-Host "=== [PQRL1] Archiving and Uploading Codebase Tarballs ==="
# Compress using Windows native tar
$prev_dir = Get-Location
Set-Location $TEMP_DEPLOY
tar.exe -czf "$env:TEMP\mev.tar.gz" mev
if (Test-Path "jetweb-time-machine") {
    tar.exe -czf "$env:TEMP\jetweb-time-machine.tar.gz" jetweb-time-machine
}
if (Test-Path "substrate-node-template") {
    tar.exe -czf "$env:TEMP\substrate-node-template.tar.gz" substrate-node-template
}
Set-Location $prev_dir

# Upload via scp
$null | scp -o StrictHostKeyChecking=no "$env:TEMP\mev.tar.gz" "root@${VPS_IP}:/tmp/"
if (Test-Path "$env:TEMP\jetweb-time-machine.tar.gz") {
    $null | scp -o StrictHostKeyChecking=no "$env:TEMP\jetweb-time-machine.tar.gz" "root@${VPS_IP}:/tmp/"
}
if (Test-Path "$env:TEMP\substrate-node-template.tar.gz") {
    $null | scp -o StrictHostKeyChecking=no "$env:TEMP\substrate-node-template.tar.gz" "root@${VPS_IP}:/tmp/"
}
$null | scp -o StrictHostKeyChecking=no "d:\pqr.info\mev\runlevels.toml" "root@${VPS_IP}:${REMOTE_ETC}/runlevels.toml"

Write-Host "=== [PQRL5] Transferring & running remote deployment ==="
$null | scp -o StrictHostKeyChecking=no "d:\pqr.info\mev\deploy_remote.sh" "root@${VPS_IP}:/tmp/deploy_remote.sh"
$null | ssh -T -o StrictHostKeyChecking=no root@$VPS_IP "bash /tmp/deploy_remote.sh"

Write-Host "Deployment completed successfully!"
