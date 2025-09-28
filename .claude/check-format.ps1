# PowerShell hook for Windows
$json = [Console]::In.ReadToEnd() | ConvertFrom-Json

if ($json.tool_name -match "Write|Edit|MultiEdit") {
    $file = $json.tool_input.file_path

    if ($file -and $file -match "\.go$") {
        Write-Host "✅ [Hook] Checking Go file: $file"

        # Check format
        $formatIssues = gofmt -d $file 2>$null
        if ($formatIssues) {
            Write-Host "⚠️ FORMAT ERROR DETECTED in $file:"
            Write-Host $formatIssues
        } else {
            Write-Host "✔️ Format OK"
        }
    }
}