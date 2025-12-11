# modules\download.ps1
. "$PSScriptRoot\common.ps1"

function Download-Software {
    param(
        [Parameter(Mandatory=$true)]
        [string]$PackageId,

        [Parameter(Mandatory=$true)]
        [string]$DownloadDir
    )

    Write-Status "Downloading $PackageId..." -Type "Info"
    
    winget download --id $PackageId --accept-package-agreements --accept-source-agreements --download-directory $DownloadDir
    
    if ($LASTEXITCODE -eq 0) {
        Write-Status "Download completed successfully to '$DownloadDir'." -Type "Success"
    } else {
        Write-Status "Download failed for $PackageId." -Type "Error"
    }
}
