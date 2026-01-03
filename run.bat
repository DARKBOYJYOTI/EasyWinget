@echo off
setlocal enabledelayedexpansion
cd /d "%~dp0"

:: Check for Admin
NET SESSION >nul 2>&1
IF %ERRORLEVEL% NEQ 0 (
    echo Requesting admin privileges...
    powershell -Command "Start-Process cmd -ArgumentList '/k cd /d \"%CD%\" && \"%~f0\"' -Verb RunAs"
    exit
)

cls
echo ==================================================
echo          EasyWinGet - Quick Start
echo ==================================================
echo.

:: ============================================
:: Check Node.js
:: ============================================
node -v >nul 2>&1
IF %ERRORLEVEL% NEQ 0 (
    echo [!] Node.js not installed
    
    IF EXIST "offline-packages\node-installer.msi" (
        echo [*] Installing Node.js - this may take 2-3 minutes...
        msiexec /i "offline-packages\node-installer.msi" /qn /norestart
        echo [OK] Node.js installed! Please restart this script.
        pause
        exit
    ) ELSE (
        echo [ERROR] Node.js installer not found!
        echo Please download: https://nodejs.org
        pause
        exit /b 1
    )
) ELSE (
    echo [OK] Node.js installed
)

:: ============================================
:: Check Dependencies
:: ============================================
IF NOT EXIST "node_modules\" (
    echo [!] Installing packages...
    
    IF EXIST "offline-packages\express-5.2.1.tgz" (
        echo [*] Installing from offline packages...
        call npm install offline-packages\express-5.2.1.tgz offline-packages\cors-2.8.5.tgz offline-packages\node-pty-1.1.0.tgz 2>nul
    ) ELSE (
        echo [*] Installing from npm...
        call npm install 2>nul
    )
    
    IF EXIST "node_modules\" (
        echo [OK] Packages installed
    ) ELSE (
        echo [ERROR] Installation failed!
        pause
        exit /b 1
    )
) ELSE (
    echo [OK] Packages ready
)

:: ============================================
:: Clean Old Jobs
:: ============================================
IF EXIST "jobs\" (
    echo [*] Cleaning old jobs...
    del /q "jobs\*.*" 2>nul
)

:: ============================================
:: Start Server
:: ============================================
echo.
echo ==================================================
echo          Starting Server...
echo ==================================================
echo.
echo Server: http://localhost:8080
echo Press Ctrl+C to stop
echo.
echo Minimizing to taskbar...
timeout /t 2 /nobreak >nul

:: Minimize the console window using PowerShell
powershell -Command "$w=Add-Type -MemberDefinition '[DllImport(\"user32.dll\")]public static extern bool ShowWindow(IntPtr h,int s);[DllImport(\"kernel32.dll\")]public static extern IntPtr GetConsoleWindow();' -Name W -Namespace N -PassThru;$w::ShowWindow($w::GetConsoleWindow(),2)" >nul 2>&1

node server.js

echo.
:: Server exited
exit
