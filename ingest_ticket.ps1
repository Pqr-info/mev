$env:MEMORY_BACKEND="sqlite"
cd d:\pqr.info\mev\memory-graph\ts
$bun = "C:\Users\theal\.bun\bin\bun.exe"

function Store-Memory {
    param($type, $title, $content, $tags)
    $output = & $bun run src/cli.ts store --type $type --title $title --content $content --tags $tags
    $id = ($output | Select-String -Pattern "Memory stored successfully with ID: (.*)" | % { $_.Matches.Groups[1].Value })
    return $id
}

function Link-Memory {
    param($from, $to, $rel, $str)
    & $bun run src/cli.ts link $from $to $rel --strength $str
}

$task = Store-Memory -type "task" -title "TICKET-004: Deployment Friction for remote-windows-admin" -content "Clone jetweb-time-machine windows-remote-admin bundle and activate remote desktop." -tags "task,deployment"
$prob1 = Store-Memory -type "problem" -title "Discovery Friction" -content "Exact repository name not found; required CLI querying." -tags "discovery,github"
$prob2 = Store-Memory -type "problem" -title "Auth Friction" -content "Invalid GITHUB_TOKEN environment variable blocked clone over HTTPS/SSH." -tags "auth,github"
$sol2 = Store-Memory -type "solution" -title "Clear GITHUB_TOKEN" -content "Unset GITHUB_TOKEN to force keyring usage for gh repo clone." -tags "auth,workaround"
$prob3 = Store-Memory -type "problem" -title "UAC Blocker" -content "Headless privilege escalation failed due to EnableLUA=1 blocking UAC prompts." -tags "uac,privilege"
$sol3 = Store-Memory -type "solution" -title "Manual Fallback" -content "Halted automation and instructed human operator to manually execute setup.ps1." -tags "workaround,fallback"
$doc = Store-Memory -type "workflow" -title "Friction Documentation" -content "Created github_issue_report.md and TICKET-004.md tracking lineage." -tags "documentation,ticket"

Link-Memory -from $prob1 -to $task -rel "RELATED_TO" -str 0.8
Link-Memory -from $prob2 -to $task -rel "RELATED_TO" -str 0.8
Link-Memory -from $prob3 -to $task -rel "RELATED_TO" -str 0.8

Link-Memory -from $sol2 -to $prob2 -rel "SOLVES" -str 1.0
Link-Memory -from $sol3 -to $prob3 -rel "SOLVES" -str 0.9

Link-Memory -from $doc -to $task -rel "RELATED_TO" -str 0.7
Link-Memory -from $doc -to $prob1 -rel "RELATED_TO" -str 0.9
Link-Memory -from $doc -to $prob2 -rel "RELATED_TO" -str 0.9
Link-Memory -from $doc -to $prob3 -rel "RELATED_TO" -str 0.9

Write-Host "Ingestion complete!"
