$hostsPath = "C:\Windows\System32\drivers\etc\hosts"
$lines = Get-Content $hostsPath
$filtered = @()
foreach ($line in $lines) {
    $trimmed = $line.Trim()
    if ($trimmed -match '^127\.0\.0\.1\s+\S+\.local$') {
        continue
    }
    $filtered += $line
}
$content = ($filtered -join "`r`n").TrimEnd() + "`r`n"
Set-Content -Path $hostsPath -Value $content -NoNewline -Force
ipconfig /flushdns
Write-Output "HOSTS_CLEANED_SUCCESSFULLY"
