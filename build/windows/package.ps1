# Package the host-generated desktop assets on the Windows release VM.
$ErrorActionPreference = "Stop"
$env:CGO_ENABLED = "1"
$env:GOTOOLCHAIN = "auto"
# The release dependencies are public. An interactive Git helper can wait forever over SSH.
$env:GIT_TERMINAL_PROMPT = "0"
$env:GCM_INTERACTIVE = "Never"
$env:GIT_CONFIG_PARAMETERS = "'credential.helper='"

function Invoke-Tool {
    $tool, $arguments = $args
    & $tool @arguments
    if ($LASTEXITCODE -ne 0) { throw "$tool failed with exit code $LASTEXITCODE" }
}

Set-Location "$PSScriptRoot\..\.."
if (-not (Test-Path backend/frontend/public/index.html -PathType Leaf)) {
    throw "Missing desktop frontend. Run just build before syncing the Windows VM."
}

Write-Output "Windows: generating resources"
Push-Location build
Invoke-Tool go tool wails3 generate syso -arch amd64 -icon windows/icon.ico -manifest windows/wails.exe.manifest -info windows/info.json -out ../backend/wails_windows_amd64.syso
Pop-Location

Write-Output "Windows: compiling the application"
try {
    Invoke-Tool go build -tags production -trimpath -buildvcs=false "-ldflags=-w -s -H windowsgui" -o bin/smeggtuner.exe ./backend
} finally {
    Remove-Item backend/wails_windows_amd64.syso
}

Write-Output "Windows: packaging the installer"
$binary = (Resolve-Path bin/smeggtuner.exe).Path
Invoke-Tool go tool wails3 generate webview2bootstrapper -dir "$PSScriptRoot\nsis"
Set-Location "$PSScriptRoot\nsis"
Invoke-Tool makensis "-DARG_WAILS_AMD64_BINARY=$binary" project.nsi
