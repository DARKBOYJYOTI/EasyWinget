# EasyWinGet Portable Edition Packager
# Builds standalone executable and packages it into a self-contained Portable ZIP with local data persistence.

$ErrorActionPreference = "Stop"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "   EasyWinGet Portable Edition Builder    " -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

$root = Split-Path -Parent $PSScriptRoot
$binDir = "$root\build\bin"
$exePath = "$binDir\EasyWinGet.exe"
$portableFolder = "$binDir\EasyWinGet-portable"
$zipPath = "$binDir\EasyWinGet-portable.zip"

$pfxPath = "$root\build\windows\DARKBOYJYOTI.pfx"
$pfxPwdRaw = if ($env:SIGN_PFX_PASSWORD) { $env:SIGN_PFX_PASSWORD } else { "" }

# 1. Compile Standalone Executable via Wails
Write-Host "`n[1/3] Compiling standalone executable via Wails..." -ForegroundColor Yellow
& wails build

# 2. Sign Standalone Executable (if certificate is present and password provided)
if ((Test-Path $pfxPath) -and (Test-Path $exePath) -and $pfxPwdRaw) {
    Write-Host "`n[2/3] Digitally signing EasyWinGet.exe with Authenticode signature..." -ForegroundColor Yellow
    $pfxPwd = ConvertTo-SecureString -String $pfxPwdRaw -Force -AsPlainText
    $cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2($pfxPath, $pfxPwd)
    $sig = Set-AuthenticodeSignature -Certificate $cert -FilePath $exePath -HashAlgorithm SHA256
    Write-Host "Signed $exePath ($($sig.SignerCertificate.Subject))" -ForegroundColor Green
} else {
    Write-Host "`n[2/3] Skipping digital signature (PFX certificate or SIGN_PFX_PASSWORD not provided)." -ForegroundColor Gray
}

# 3. Assemble Portable Distribution Folder
Write-Host "`n[3/3] Creating Portable package structure..." -ForegroundColor Yellow
if (Test-Path $portableFolder) {
    Remove-Item -Recurse -Force $portableFolder
}
New-Item -ItemType Directory -Force -Path $portableFolder | Out-Null
New-Item -ItemType Directory -Force -Path "$portableFolder\data" | Out-Null
New-Item -ItemType Directory -Force -Path "$portableFolder\downloads" | Out-Null

Copy-Item $exePath -Destination "$portableFolder\EasyWinGet.exe" -Force

# Create portable marker file
New-Item -ItemType File -Force -Path "$portableFolder\portable" | Out-Null

# Create portable README
$readmeContent = @"
========================================================================
                      EasyWinGet Portable Edition
========================================================================

EasyWinGet Portable is completely self-contained:
- Database:   Saved in .\data\easywinget.db
- Icons:      Cached in .\data\icons\
- Downloads:  Stored in .\downloads\
- Job Logs:   Stored in .\jobs\

No installation required. You can run this directly from a USB flash drive
or any directory without touching Windows system files or AppData.

Publisher: DARKBOYJYOTI
Repository: https://github.com/DARKBOYJYOTI/EasyWinget
========================================================================
"@
$readmeContent | Out-File -Encoding utf8 "$portableFolder\README.txt"

# Compress into portable zip
if (Test-Path $zipPath) {
    Remove-Item -Force $zipPath
}
Compress-Archive -Path "$portableFolder\*" -DestinationPath $zipPath -Force

$zipSizeMb = [math]::Round((Get-Item $zipPath).Length / 1MB, 2)
Write-Host "`n========================================================" -ForegroundColor Green
Write-Host " EasyWinGet Portable Edition Successfully Built!       " -ForegroundColor Green
Write-Host "========================================================" -ForegroundColor Green
Write-Host "ZIP Archive: $zipPath ($zipSizeMb MB)" -ForegroundColor Green
Write-Host "Folder:      $portableFolder" -ForegroundColor Cyan
Write-Host "Storage:     Self-contained (data/, downloads/, jobs/ alongside EXE)" -ForegroundColor Cyan
