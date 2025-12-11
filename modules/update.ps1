# modules\update.ps1
. "$PSScriptRoot\common.ps1"

function Check-For-Updates {
    Write-Status "Checking for updates..." -Type "Info"
    
    $updateOutput = winget upgrade --include-unknown | Out-String
    
    if ($updateOutput -match "No applicable update found" -or $updateOutput -match "No installed package found matching input criteria") {
        Write-Status "All packages are up to date." -Type "Success"
    } else {
        Write-Host $updateOutput
        $choice = Read-Host "Updates are available. Do you want to upgrade all? (y/n)"
        if ($choice -eq 'y') {
            Write-Status "Upgrading all packages..." -Type "Info"
            winget upgrade --all -h --accept-source-agreements --accept-package-agreements --include-unknown
            if ($LASTEXITCODE -eq 0) {
                Write-Status "All packages upgraded successfully." -Type "Success"
            } else {
                Write-Status "An error occurred during the upgrade." -Type "Error"
            }
        }
    }
    Confirm-Action
}
