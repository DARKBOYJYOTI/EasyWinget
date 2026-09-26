@echo off
setlocal
title EasyWinGet - Go + Svelte Desktop

echo ========================================================
echo   EasyWinGet - Windows Package Manager (Go + Svelte)
echo ========================================================
echo.

if exist "build\bin\EasyWinGet.exe" (
    echo Launching EasyWinGet Native...
    start "" "build\bin\EasyWinGet.exe"
    exit /b 0
)

echo Built executable not found. Running with Wails in development mode...
wails dev
pause
