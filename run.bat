@echo off
chcp 65001 >nul 2>&1
:: Check for admin privileges
net session >nul 2>&1
if %errorLevel% == 0 (
    goto :run
) else (
    goto :elevate
)

:elevate
echo Requesting administrator privileges...
powershell -Command "Start-Process '%~f0' -Verb RunAs"
exit /b

:run
title Software Manager
color 0F
echo ====================================
echo   SYSTEM UPDATE CHECKER
echo ====================================
echo.
powershell.exe -ExecutionPolicy Bypass -NoProfile -File "%~dp0script.ps1"
echo.
echo ====================================
echo   Press any key to exit...
echo ====================================
pause >nul