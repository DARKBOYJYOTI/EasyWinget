function Write-Status {
    param(
        [string]$Message,
        [ValidateSet('Info', 'Success', 'Warning', 'Error', 'Header')]
        [string]$Type = 'Info'
    )
    
    $icons = @{
        Info    = '[i]'
        Success = '[+]'
        Warning = '[!]'
        Error   = '[X]'
        Header  = '[*]'
    }
    
    $colors = @{
        Info    = 'Cyan'
        Success = 'Green'
        Warning = 'Yellow'
        Error   = 'Red'
        Header  = 'Magenta'
    }
    
    Write-Host "$($icons[$Type]) $Message" -ForegroundColor $colors[$Type]
}

function Confirm-Action {
    Write-Host ""
    Write-Status "Press Enter to continue..." -Type "Info"
    Read-Host | Out-Null
}
