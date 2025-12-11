<#
.SYNOPSIS
    Software Management Script
.DESCRIPTION
    Search, download, and install software using winget.
.VERSION
    3.0
#>

# Set console to UTF-8
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

# ============================================
# CONFIGURATION
# ============================================
$ErrorActionPreference = 'SilentlyContinue'
$ProgressPreference = 'SilentlyContinue'
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$DownloadDir = Join-Path -Path $ScriptDir -ChildPath "Downloads"
$ModulesPath = Join-Path -Path $ScriptDir -ChildPath "modules"

# Import modules
. (Join-Path $ModulesPath "common.ps1")
. (Join-Path $ModulesPath "search.ps1")
. (Join-Path $ModulesPath "download.ps1")
. (Join-Path $ModulesPath "install.ps1")
. (Join-Path $ModulesPath "update.ps1")

# Create Download directory if it doesn't exist
if (-not (Test-Path -Path $DownloadDir)) {
    New-Item -ItemType Directory -Path $DownloadDir | Out-Null
}

# ============================================
# UI FUNCTIONS
# ============================================
function Show-Menu {
    Clear-Host
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "   SOFTWARE MANAGER v3.0 (winget)      " -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "  1. Search and Install Software"
    Write-Host "  2. Check for Updates"
    Write-Host "  3. Exit"
    Write-Host ""
}

# ============================================
# WORKFLOWS
# ============================================
function Search-And-Install-Workflow {
    Write-Status "Enter search term (leave blank to list popular apps):" -Type "Header"
    $searchTerm = Read-Host
    
    Write-Status "Searching..." -Type "Info"
    $results = @(Search-Software -searchTerm $searchTerm)
    
    if ($null -eq $results -or $results.Count -eq 0) {
        Write-Status "No software found." -Type "Warning"
        Confirm-Action
        return
    }

    Write-Status "Search Results:" -Type "Header"
    Write-Host ("{0,3} {1,-30} {2,-40} {3,-15}" -f "#", "Name", "Id", "Version")
    Write-Host ("-"*90)
    for ($i = 0; $i -lt $results.Count; $i++) {
        $result = $results[$i]
        Write-Host ("{0,3}. {1,-30} {2,-40} {3,-15}" -f ($i + 1), $result.Name, $result.Id, $result.Version)
    }

    Write-Host ""
    $choice = Read-Host "Enter the number of the software to proceed with (or press Enter to cancel)"
    
    if ([string]::IsNullOrWhiteSpace($choice)) {
        return
    }
    
    $index = 0
    if (($choice -as [int]) -and ($choice -ge 1) -and ($choice -le $results.Count)) {
        $index = [int]$choice - 1
        $selectedSoftware = $results[$index]
        
        Write-Host ""
        Write-Status "You selected: $($selectedSoftware.Name)" -Type "Header"
        Write-Host "  1. Install (Recommended)"
        Write-Host "  2. Download Only"
        $action = Read-Host "Choose an action"

        switch ($action) {
            "1" { Install-SoftwareById -PackageId $selectedSoftware.Id }
            "2" { Download-Software -PackageId $selectedSoftware.Id -DownloadDir $DownloadDir }
            default { Write-Status "Invalid action." -Type "Warning" }
        }
    } else {
        Write-Status "Invalid selection." -Type "Warning"
    }
    
    Confirm-Action
}

# ============================================
# MAIN EXECUTION
# ============================================
function Main {
    # Check if winget is available
    if (-not (Get-Command winget -ErrorAction SilentlyContinue)) {
        Write-Status "Winget is not installed or not in PATH. Please install it to use this script." -Type "Error"
        Confirm-Action
        return
    }

    while ($true) {
        Show-Menu
        $choice = Read-Host "Enter your choice"
        
        switch ($choice) {
            "1" { Search-And-Install-Workflow }
            "2" { Check-For-Updates }
            "3" { break }
            default {
                Write-Status "Invalid choice. Please try again." -Type "Warning"
                Start-Sleep -Seconds 2
            }
        }
    }
}

Main
