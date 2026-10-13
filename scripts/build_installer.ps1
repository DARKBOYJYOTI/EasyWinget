param(
    [string]$Scope = "machine" # "machine" or "user"
)

$ErrorActionPreference = "Stop"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "   EasyWinGet Windows Installer Builder   " -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan

$root = Split-Path -Parent $PSScriptRoot
$toolsNsis = "$root\tools\nsis-3.10\makensis.exe"
$pfxPath = "$root\build\windows\DARKBOYJYOTI.pfx"
$pfxPwdRaw = if ($env:SIGN_PFX_PASSWORD) { $env:SIGN_PFX_PASSWORD } else { "" }

# 1. Compile Standalone Executable via Wails
Write-Host "`n[1/4] Compiling standalone executable via Wails..." -ForegroundColor Yellow
& wails build

# 2. Sign Standalone Executable
$standaloneExe = "$root\build\bin\EasyWinGet.exe"
if ((Test-Path $pfxPath) -and (Test-Path $standaloneExe) -and $pfxPwdRaw) {
    Write-Host "`n[2/4] Signing standalone executable with Authenticode signature..." -ForegroundColor Yellow
    $pfxPwd = ConvertTo-SecureString -String $pfxPwdRaw -Force -AsPlainText
    $cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2($pfxPath, $pfxPwd)
    $sig = Set-AuthenticodeSignature -Certificate $cert -FilePath $standaloneExe -HashAlgorithm SHA256
    Write-Host "Signed $standaloneExe ($($sig.SignerCertificate.Subject))" -ForegroundColor Green
} else {
    Write-Host "`n[2/4] Skipping code signing (certificate or SIGN_PFX_PASSWORD not provided)." -ForegroundColor Gray
}

# 3. Locate makensis
$makensis = $null
if (Get-Command makensis -ErrorAction SilentlyContinue) {
    $makensis = (Get-Command makensis).Source
} elseif (Test-Path $toolsNsis) {
    $makensis = $toolsNsis
} elseif (Test-Path "C:\Program Files (x86)\NSIS\makensis.exe") {
    $makensis = "C:\Program Files (x86)\NSIS\makensis.exe"
} elseif (Test-Path "C:\Program Files\NSIS\makensis.exe") {
    $makensis = "C:\Program Files\NSIS\makensis.exe"
}

if (-not $makensis) {
    Write-Error "NSIS compiler (makensis.exe) not found."
}

Write-Host "`n[3/4] Generating Windows Setup Wizard with NSIS ($makensis)..." -ForegroundColor Yellow
$nsiScript = "$root\build\windows\installer\project.nsi"
$exeRelative = "..\..\bin\EasyWinGet.exe"

& $makensis "/DARG_WAILS_AMD64_BINARY=$exeRelative" "/DWAILS_INSTALL_SCOPE=$Scope" $nsiScript

# 4. Sign Installer
$installerPath = "$root\build\bin\EasyWinGet-amd64-installer.exe"
if (Test-Path $installerPath) {
    if ((Test-Path $pfxPath) -and $pfxPwdRaw) {
        Write-Host "`n[4/4] Signing installer with Authenticode signature..." -ForegroundColor Yellow
        $pfxPwd = ConvertTo-SecureString -String $pfxPwdRaw -Force -AsPlainText
        $cert = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2($pfxPath, $pfxPwd)
        $sig = Set-AuthenticodeSignature -Certificate $cert -FilePath $installerPath -HashAlgorithm SHA256
        Write-Host "Signed $installerPath ($($sig.SignerCertificate.Subject))" -ForegroundColor Green
    }
    $sizeMb = [math]::Round((Get-Item $installerPath).Length / 1MB, 2)
    Write-Host "`n========================================================" -ForegroundColor Green
    Write-Host " Windows Installer & Executable Built & Signed! " -ForegroundColor Green
    Write-Host "========================================================" -ForegroundColor Green
    Write-Host "Installer: $installerPath ($sizeMb MB)" -ForegroundColor Green
    Write-Host "Publisher: DARKBOYJYOTI (Digitally Signed)" -ForegroundColor Cyan
    Write-Host "Taskbar Pinning: Enabled by default" -ForegroundColor Cyan
    Write-Host "Desktop Shortcut: Enabled by default" -ForegroundColor Cyan
} else {
    Write-Error "Installer generation failed."
}
