@echo off
setlocal enabledelayedexpansion
cd /d "%~dp0"

echo ==================================================
echo EasyWinGet - Installer
echo ==================================================
echo.

:: ============================================
:: Check for Admin
:: ============================================
NET SESSION >nul 2>&1
IF %ERRORLEVEL% NEQ 0 (
    echo Requesting admin privileges...
    powershell -Command "Start-Process '%~f0' -Verb RunAs"
    exit /b
)

echo [OK] Running with Administrator privileges
echo.

:: ============================================
:: Check Node.js Installation
:: ============================================
echo [*] Checking Node.js installation...
node -v >nul 2>&1
IF %ERRORLEVEL% NEQ 0 (
    echo [!] Node.js not found.
    
    IF EXIST "offline-packages\node-installer.msi" (
        echo [*] Installing Node.js from offline package...
        msiexec /i "%~dp0offline-packages\node-installer.msi" /qn /norestart
        IF %ERRORLEVEL% EQU 0 (
            echo [OK] Node.js installed successfully!
            echo [!] IMPORTANT: Please restart your computer, then run install.bat again.
            echo.
            pause
            exit /b 0
        ) ELSE (
            echo [ERROR] Failed to install Node.js! Error code: %ERRORLEVEL%
            pause
            exit /b 1
        )
    ) ELSE (
        echo [ERROR] Node.js installer not found!
        echo Please download Node.js from: https://nodejs.org
        pause
        exit /b 1
    )
) ELSE (
    for /f "tokens=*" %%i in ('node -v') do set NODE_VERSION=%%i
    echo [OK] Node.js !NODE_VERSION! is installed
)

:: ============================================
:: Stop Existing Server Instances
:: ============================================
echo.
echo [*] Stopping existing server instances...

:: Method 1: Using taskkill (more reliable)
taskkill /F /IM node.exe >nul 2>&1

:: Method 2: Using PowerShell for port check (fallback)
powershell -Command "Get-Process -Name node -ErrorAction SilentlyContinue | Stop-Process -Force" >nul 2>&1

:: Wait for processes to terminate
timeout /t 2 /nobreak >nul
echo [OK] Existing processes stopped

:: ============================================
:: Prepare Installation Directory
:: ============================================
echo.
echo [*] Preparing Installation Directory (C:\EasyWinGet)...

:: Create directory if not exists
if not exist "C:\EasyWinGet" (
    mkdir "C:\EasyWinGet"
    echo [OK] Directory created
) ELSE (
    echo [OK] Directory exists
)

:: ============================================
:: Preserve Downloads Folder
:: ============================================
echo [*] Preserving Downloads folder...
set "DOWNLOADS_BACKUP="
if exist "C:\EasyWinGet\Downloads" (
    set "DOWNLOADS_BACKUP=1"
    if not exist "C:\EasyWinGet\downloads_backup_temp" (
        ren "C:\EasyWinGet\Downloads" "downloads_backup_temp" >nul 2>&1
    )
    if not exist "C:\EasyWinGet\downloads_backup_temp" (
        ren "C:\EasyWinGet\downloads" "downloads_backup_temp" >nul 2>&1
    )
)
if exist "C:\EasyWinGet\downloads" (
    set "DOWNLOADS_BACKUP=1"
    if not exist "C:\EasyWinGet\downloads_backup_temp" (
        ren "C:\EasyWinGet\downloads" "downloads_backup_temp" >nul 2>&1
    )
)

:: ============================================
:: Clean Old Files
:: ============================================
echo [*] Cleaning old files...
cd /d "C:\EasyWinGet"
for /f "tokens=*" %%d in ('dir /b /a-d 2^>nul') do (
    del /f /q "%%d" >nul 2>&1
)
for /f "tokens=*" %%d in ('dir /b /ad 2^>nul ^| findstr /v "downloads_backup_temp"') do (
    rd /s /q "%%d" >nul 2>&1
)
echo [OK] Old files cleaned

:: ============================================
:: Copy New Files
:: ============================================
cd /d "%~dp0"
echo [*] Copying new files...

:: Use xcopy for simpler and more reliable copying
xcopy "%~dp0*" "C:\EasyWinGet\" /E /I /Y /H /C /Q
if %ERRORLEVEL% EQU 0 (
    echo [OK] Files copied successfully
) ELSE (
    echo [ERROR] Failed to copy files! Error code: %ERRORLEVEL%
    echo [!] This might be due to locked files. Please close any running instances and try again.
    pause
    exit /b 1
)

:: ============================================
:: Restore Downloads Folder
:: ============================================
echo [*] Restoring Downloads folder...
if "%DOWNLOADS_BACKUP%"=="1" (
    if exist "C:\EasyWinGet\downloads_backup_temp" (
        if exist "C:\EasyWinGet\Downloads" rd /s /q "C:\EasyWinGet\Downloads" >nul 2>&1
        if exist "C:\EasyWinGet\downloads" rd /s /q "C:\EasyWinGet\downloads" >nul 2>&1
        ren "C:\EasyWinGet\downloads_backup_temp" "Downloads" >nul 2>&1
        echo [OK] Downloads folder restored
    )
) ELSE (
    echo [OK] No Downloads folder to restore
)

:: ============================================
:: Verify Required Files
:: ============================================
echo.
echo [*] Verifying installation...

set "MISSING_FILES="

if not exist "C:\EasyWinGet\server.js" (
    echo [ERROR] server.js not found!
    set "MISSING_FILES=1"
)
if not exist "C:\EasyWinGet\utils\start_server_hidden.vbs" (
    echo [ERROR] start_server_hidden.vbs not found!
    set "MISSING_FILES=1"
)
if not exist "C:\EasyWinGet\utils\server_runner.bat" (
    echo [ERROR] server_runner.bat not found!
    set "MISSING_FILES=1"
)
if not exist "C:\EasyWinGet\gui\default-icon.ico" (
    echo [WARNING] default-icon.ico not found, using default icon
)

if "%MISSING_FILES%"=="1" (
    echo [ERROR] Installation incomplete! Missing required files.
    pause
    exit /b 1
)
echo [OK] All required files verified

:: ============================================
:: Install Dependencies
:: ============================================
cd /d "C:\EasyWinGet"
echo.
echo [*] Checking dependencies...

IF NOT EXIST "node_modules\" (
    echo [!] node_modules not found. Installing dependencies...
    
    IF EXIST "offline-packages\express-5.2.1.tgz" (
        echo [*] Installing from offline packages...
        call npm install offline-packages\express-5.2.1.tgz offline-packages\cors-2.8.5.tgz offline-packages\node-pty-1.1.0.tgz --no-audit --no-fund
        IF %ERRORLEVEL% NEQ 0 (
            echo [WARNING] Offline install failed, trying online...
            call npm install --no-audit --no-fund
        )
    ) ELSE (
        echo [*] Installing from npm...
        call npm install --no-audit --no-fund
    )
    
    IF EXIST "node_modules\" (
        echo [OK] Dependencies installed
    ) ELSE (
        echo [ERROR] Failed to install dependencies!
        pause
        exit /b 1
    )
) ELSE (
    echo [OK] Dependencies already installed
)

:: ============================================
:: Create Shortcuts
:: ============================================
echo.
echo [*] Creating shortcuts...

:: Desktop Shortcut
set "VBS_FILE=%TEMP%\CreateShortcut.vbs"
(
    echo Set oWS = WScript.CreateObject^("WScript.Shell"^)
    echo sLinkFile = oWS.ExpandEnvironmentStrings^("%%USERPROFILE%%\Desktop\EasyWinGet.lnk"^)
    echo Set oLink = oWS.CreateShortcut^(sLinkFile^)
    echo oLink.TargetPath = "wscript.exe"
    echo oLink.Arguments = """C:\EasyWinGet\utils\start_server_hidden.vbs"""
    echo oLink.WorkingDirectory = "C:\EasyWinGet"
    echo oLink.Description = "EasyWinGet Package Manager"
    if exist "C:\EasyWinGet\gui\default-icon.ico" (
        echo oLink.IconLocation = "C:\EasyWinGet\gui\default-icon.ico"
    )
    echo oLink.Save
) > "%VBS_FILE%"
cscript //nologo "%VBS_FILE%"
del "%VBS_FILE%" >nul 2>&1
echo [OK] Desktop shortcut created

:: Start Menu Shortcut
set "START_MENU=%APPDATA%\Microsoft\Windows\Start Menu\Programs"
(
    echo Set oWS = WScript.CreateObject^("WScript.Shell"^)
    echo sLinkFile = "%START_MENU%\EasyWinGet.lnk"
    echo Set oLink = oWS.CreateShortcut^(sLinkFile^)
    echo oLink.TargetPath = "wscript.exe"
    echo oLink.Arguments = """C:\EasyWinGet\utils\start_server_hidden.vbs"""
    echo oLink.WorkingDirectory = "C:\EasyWinGet"
    echo oLink.Description = "EasyWinGet Package Manager"
    if exist "C:\EasyWinGet\gui\default-icon.ico" (
        echo oLink.IconLocation = "C:\EasyWinGet\gui\default-icon.ico"
    )
    echo oLink.Save
) > "%VBS_FILE%"
cscript //nologo "%VBS_FILE%"
del "%VBS_FILE%" >nul 2>&1
echo [OK] Start Menu shortcut created

:: ============================================
:: Clean Job Files
:: ============================================
echo.
echo [*] Cleaning old job files...
if exist "C:\EasyWinGet\jobs\" (
    del /q "C:\EasyWinGet\jobs\*.*" >nul 2>&1
    echo [OK] Job files cleaned
)

:: ============================================
:: Installation Complete
:: ============================================
echo.
echo ==================================================
echo [OK] Installation Complete!
echo ==================================================
echo.
echo  Installation Path: C:\EasyWinGet
echo  Desktop Shortcut:  EasyWinGet.lnk
echo  Start Menu:        Programs ^> EasyWinGet
echo  Server URL:        http://localhost:8080
echo.
echo ==================================================
echo.

:: ============================================
:: Launch Server
:: ============================================
echo Launching EasyWinGet...
timeout /t 2 /nobreak >nul

:: Start the hidden server
start "" wscript.exe "C:\EasyWinGet\utils\start_server_hidden.vbs"

:: Wait and open browser
timeout /t 3 /nobreak >nul
start http://localhost:8080

echo.
echo EasyWinGet is now running!
echo You can close this window.
echo.
pause
exit /b 0
