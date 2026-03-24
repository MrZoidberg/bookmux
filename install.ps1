# Install script for BookMux (Windows, PowerShell)
$repo = "MrZoidberg/bookmux"

$os = "Windows"
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "x86_64" } else { "arm64" }

# Fetch latest release data
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
$version = $release.tag_name
$version_num = $version.TrimStart('v')

# URL format based on GoReleaser template: {{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}.zip
$fileName = "bookmux_${version_num}_${os}_${arch}.zip"
$url = "https://github.com/$repo/releases/download/$version/$fileName"

$destDir = "$HOME\.local\bin"
if (!(Test-Path $destDir)) { New-Item -ItemType Directory -Path $destDir | Out-Null }

Write-Host "--------------------------------------------------------"
Write-Host "Downloading BookMux $version for $os $arch..."
$zipFile = Join-Path $env:TEMP "bookmux.zip"
$extractPath = Join-Path $env:TEMP "bookmux_extract"

Invoke-WebRequest -Uri $url -OutFile $zipFile

if (Test-Path $extractPath) { Remove-Item -Path $extractPath -Recurse -Force }
Expand-Archive -Path $zipFile -DestinationPath $extractPath

# Move binary to local bin
Move-Item -Path "$extractPath\bookmux.exe" -Destination "$destDir\bookmux.exe" -Force

# Cleanup
Remove-Item -Path $zipFile
Remove-Item -Path $extractPath -Recurse

Write-Host "BookMux $version installed successfully!"
Write-Host "Location: $destDir\bookmux.exe"
Write-Host ""
Write-Host "Make sure '$destDir' is in your PATH."
Write-Host "  [System.Environment]::SetEnvironmentVariable('Path', [System.Environment]::GetEnvironmentVariable('Path', 'User') + ';$destDir', 'User')"
Write-Host "--------------------------------------------------------"
