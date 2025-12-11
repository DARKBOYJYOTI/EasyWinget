# modules\install.ps1
. "$PSScriptRoot\common.ps1"

function Install-SoftwareById {
    param(
        [Parameter(Mandatory=$true)]
        [string]$PackageId
    )

    Write-Status "Installing $PackageId..." -Type "Info"
    
    winget install --id $PackageId -h --accept-source-agreements --accept-package-agreements
    
    if ($LASTEXITCODE -eq 0) {
        Write-Status "$PackageId installed successfully." -Type "Success"
    } else {
        Write-Status "Installation failed for $PackageId." -Type "Error"
    }
}
