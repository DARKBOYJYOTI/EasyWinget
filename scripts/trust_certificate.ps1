# Elevates to Administrator to install DARKBOYJYOTI Code Signing Certificate into Windows Trusted Root & Trusted Publishers
param()

$root = Split-Path -Parent $PSScriptRoot
$cerPath = "$root\build\windows\DARKBOYJYOTI.cer"

if (-not (Test-Path $cerPath)) {
    Write-Error "Certificate file $cerPath not found."
    exit 1
}

$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "Requesting Administrator elevation to trust DARKBOYJYOTI certificate..." -ForegroundColor Cyan
    Start-Process powershell.exe -Verb RunAs -ArgumentList "-ExecutionPolicy Bypass -File `"$PSCommandPath`""
    exit
}

Write-Host "Adding DARKBOYJYOTI to Trusted Root Certification Authorities..." -ForegroundColor Yellow
certutil.exe -addstore -f "Root" $cerPath

Write-Host "Adding DARKBOYJYOTI to Trusted Publishers..." -ForegroundColor Yellow
certutil.exe -addstore -f "TrustedPublisher" $cerPath

Write-Host "`n========================================================" -ForegroundColor Green
Write-Host " [SUCCESS] DARKBOYJYOTI Certificate Installed & Trusted! " -ForegroundColor Green
Write-Host "========================================================" -ForegroundColor Green
Write-Host "UAC prompts will now display: Verified publisher: DARKBOYJYOTI" -ForegroundColor Cyan
Start-Sleep -Seconds 2
