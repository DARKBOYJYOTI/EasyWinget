# modules\search.ps1
. "$PSScriptRoot\common.ps1"

function Search-Software {
    param(
        [string]$searchTerm
    )
    
    if ([string]::IsNullOrWhiteSpace($searchTerm)) {
        $wingetOutput = winget search "" --accept-source-agreements *>&1
    } else {
        $wingetOutput = winget search $searchTerm --accept-source-agreements *>&1
    }

    $lines = $wingetOutput | Out-String | ForEach-Object { $_.Split([Environment]::NewLine) }

    # Find header and separator lines
    $headerLine = $lines | Where-Object { $_.Trim().StartsWith("Name") } | Select-Object -First 1
    $separatorLine = $lines | Where-Object { $_.Trim().StartsWith("---") } | Select-Object -First 1

    if ([string]::IsNullOrWhiteSpace($headerLine) -or [string]::IsNullOrWhiteSpace($separatorLine)) {
        # This can happen if no results are found, winget outputs a different message.
        return @()
    }

    # Find column indices from the header
    $idColIndex = $headerLine.IndexOf("Id")
    $versionColIndex = $headerLine.IndexOf("Version")

    if ($idColIndex -lt 0 -or $versionColIndex -lt 0) {
        Write-Status "Could not parse winget search output columns." -Type "Error"
        return @()
    }

    # Find where the data starts
    $dataStartIndex = [array]::IndexOf($lines, $separatorLine) + 1

    $results = @()
    for ($i = $dataStartIndex; $i -lt $lines.Length; $i++) {
        $line = $lines[$i]
        if ([string]::IsNullOrWhiteSpace($line)) { continue }

        # Extract based on column indices
        $name = $line.Substring(0, $idColIndex).Trim()
        $id = $line.Substring($idColIndex, $versionColIndex - $idColIndex).Trim()
        $versionAndMatch = $line.Substring($versionColIndex).Trim()
        
        $versionParts = $versionAndMatch.Split(' ', [System.StringSplitOptions]::RemoveEmptyEntries)
        $version = if ($versionParts.Length -gt 0) { $versionParts[0] } else { "Unknown" }

        if (-not ([string]::IsNullOrWhiteSpace($name))) {
            # Sometimes the ID parsing grabs the version too if columns are misaligned
            $idParts = $id.Split(' ', [System.StringSplitOptions]::RemoveEmptyEntries)
            if ($idParts.Length -gt 1) {
                $id = $idParts[0]
                if ($version -eq "Unknown" -and $idParts.Length -gt 1) {
                    $version = $idParts[1]
                }
            }

            $results += [pscustomobject]@{
                Name = $name
                Id = $id
                Version = $version
            }
        }
    }
    
    return $results
}
