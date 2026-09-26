# 🚀 EasyWinGet Native

**Modern Desktop GUI for Windows Package Manager (WinGet).**
_Ultra-fast, native, single-executable desktop application built with **Go** and **Svelte 5**._

![EasyWinGet Interface](assets/preview.png)

EasyWinGet provides a beautiful, native desktop interface for Windows Package Manager (WinGet). It brings Windows 11 Fluent / Mica aesthetics, real-time interactive terminal logs, offline package caching, batch installs, and updates management directly to your desktop.

---

## ✨ Features

- **⚡ Native Windows App**: Compiles into a single, lightweight `.exe` binary powered by **Go** and **Wails v2** (No Node.js runtime required for end-users!).
- **🎨 Modern Svelte 5 Frontend**: Clean, responsive, glassmorphic Windows 11 Fluent UI with dark mode by default.
- **🔍 Instant Search & Discover**: Fast package search across thousands of WinGet apps with categories, filters, and debouncing.
- **📦 Installed Apps Manager**: View, filter, and uninstall applications (supporting both WinGet and Native Windows uninstallers).
- **🔄 Updates Center**: Detect outdated applications, view version diffs, update individual apps or update all with one click, and ignore unwanted updates.
- **💾 Offline Downloads Manager**: Download installer files (.exe, .msi, .msix) for backup, offline storage, or portable distribution.
- **🖥️ Real-time Console**: Live streaming terminal execution with process cancellation support and background docking.
- **📤 Export & Import**: Backup your installed applications list to JSON and easily restore them on a new PC.

---

## 🛠️ Requirements & Development

- **OS**: Windows 10 (Version 1809+) or Windows 11.
- **WinGet**: Built into modern Windows systems.
- **Go**: 1.24+
- **Wails CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Node.js**: Required only for compiling the Svelte frontend during development.

### Running in Development Mode

```powershell
wails dev
```

### Compiling Production Binary (.exe)

```powershell
wails build
```
The compiled single `.exe` is generated at `build/bin/EasyWinGet.exe`.

---

## 📂 Project Structure

```
EasyWinGet/
├── backend/              # Go Backend Services
│   ├── cache/            # Thread-safe JSON caching
│   ├── downloads/        # Installer downloads manager
│   ├── icons/            # Manifest and local app icon resolution
│   ├── jobs/             # Real-time background job runner & streaming
│   ├── models/           # Shared data structures
│   └── winget/           # WinGet CLI wrapper & parsers
├── frontend/             # Modern Svelte 5 + Vite UI
│   ├── src/
│   │   ├── components/   # Svelte components (Sidebar, Search, Installed, Updates, etc.)
│   │   ├── App.svelte    # Root UI coordinator
│   │   ├── main.js       # Entry point
│   │   └── style.css     # Fluent Mica design system & tokens
│   ├── wailsjs/          # Auto-generated Go-to-Svelte bindings
│   └── package.json
├── build/                # Windows application assets & manifest
│   ├── appicon.png       # App icon
│   └── bin/              # Compiled output (EasyWinGet.exe)
├── utils/                # Helper PowerShell scripts (icons, native uninstall)
├── main.go               # Wails app initialization & native window configuration
├── app.go                # Wails API bridges
├── wails.json            # Wails project config
└── run.bat               # Quick launcher
```

---

## 🛡️ License

This project is open-source under the [MIT License](LICENSE).
Developed by Jyoti Karmakar ([DARKBOYJYOTI](https://github.com/DARKBOYJYOTI)).
