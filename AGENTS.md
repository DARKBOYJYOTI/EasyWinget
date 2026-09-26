# EasyWinGet Agent Notes

This repo is a native Windows desktop application built with **Go (Golang)**, **Wails v2**, and **Svelte 5**.
It manages Windows Package Manager (`winget`) operations with real-time streaming, background jobs, caching, and offline installer tracking.
A complete backup of the previous Node.js/Express version is preserved under `.bak/` (ignored by git).

## Repo Map

- `main.go`: Wails application entry point, window configuration (Mica backdrop, dark theme), asset embedding.
- `app.go`: Bridge between Wails runtime and backend services, exposing typed methods to the frontend.
- `backend/`:
  - `winget/winget.go`: Wraps `winget` CLI, handles UTF-8 execution, column parsing (Search, List, Upgrades, Show/Details).
  - `jobs/jobs.go`: Real-time background job runner with live log streaming via Wails events (`job:log`, `job:done`) and process cancellation.
  - `db/db.go`: High-speed CGO-free pure Go SQLite database for sub-millisecond local caching (WAL mode).
  - `installed/registry.go`: Pure Go native Windows Registry scanner & uninstaller (BCUninstaller style, zero PowerShell dependencies).
  - `icons/icons.go`: Local executable icon extraction and manifest favicon scraping.
  - `downloads/downloads.go`: Installer downloads scanner and manager (%USERPROFILE%\Downloads\EasyWinGet).
  - `system/autostart.go`: Windows Registry auto-start and accent color integration.
  - `models/models.go`: Shared data structs.
- `frontend/`:
  - `src/App.svelte`: Root UI coordinator and layout with 3-way theme switching and OS theme listeners.
  - `src/utils/theme.js`: Fluent 2 dynamic luminance and auto-contrast color system for Windows accent.
  - `src/utils/icons.js`: Multi-stage intelligent icon resolver with Google S2 Favicon API & DuckDuckGo fallback.
  - `src/components/`:
    - `Sidebar.svelte`: Windows 11 Fluent sidebar navigation with high-contrast active icons.
    - `SearchView.svelte`: Discover / search WinGet repository with categories and chips.
    - `InstalledView.svelte`: Installed applications list, filtering, export, and uninstaller.
    - `UpdatesView.svelte`: Outdated apps detection, version comparison, update-all, and ignore update.
    - `DownloadsView.svelte`: Offline installers manager.
    - `TaskTerminal.svelte`: Real-time live log viewer with smart auto-scroll, stdin, and cancelation.
    - `DetailsModal.svelte`: Manifest inspection modal.
    - `SettingsModal.svelte`: Centered settings cards, 3-way theme toggle, and styled card icons.
    - `Toast.svelte`: Web Audio chime and notification toasts.
  - `wailsjs/`: Auto-generated bindings for Go methods and runtime.
- `build/bin/EasyWinGet.exe`: Compiled native standalone Windows desktop executable (zero external runtime required).

## Development Commands

Run in live-reload development mode:

```powershell
wails dev
```

Build standalone Windows desktop executable (`.exe`):

```powershell
wails build
```

Frontend build test:

```powershell
cd frontend
npm run build
```

Go syntax and type check:

```powershell
go vet ./...
```
