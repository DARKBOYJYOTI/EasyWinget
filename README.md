# 🚀 EasyWinGet

**The Professional Web Interface for Windows Package Manager (Winget).**
*Modern. Fast. Offline-Capable.*

![EasyWinGet Interface](assets/preview.png)

EasyWinGet is a powerful, lightweight web-based GUI for Winget. It bridges the gap between the command line and a modern user experience, offering advanced features like bulk installation, task management, and offline support.

---

## ✨ Features

- **🌐 Modern Web UI**: A sleek, responsive interface built with Vanilla JS and CSS (No heavy frameworks).
- **📦 Bulk Operations**: Install, update, or uninstall multiple apps simultaneously.
- **⚡ Fast Search**: Real-time search with caching and debouncing.
- **🛑 Task Management**: Cancel running downloads or installations instantly.
- **🔌 Offline Capable**:
    - Bundled Node.js installer (no internet required to set up).
    - Offline icon caching system.
    - Graceful degradation when the internet is lost.
- **🎨 Visual Feedback**: Real-time terminal logs, progress animations, and connection status indicators.

---

## 🛠️ Prerequisites

- **OS**: Windows 10 (Version 1809+) or Windows 11.
- **Winget**: Pre-installed on modern Windows.
    - *Missing Winget?* [Download App Installer from Microsoft](https://aka.ms/getwinget).

---

## 📥 Installation

> **Note:** This release includes a bundled Node.js runtime and dependencies, making it fully portable and offline-ready.

### Option 1: Full Installation (Recommended)
Installs EasyWinGet to `C:\EasyWinGet` and creates Desktop/Start Menu shortcuts.

1.  **Download** the latest release `.zip` from GitHub.
2.  Extract the folder.
3.  Right-click `install.bat` and select **Run as Administrator**.
4.  The script will:
    - Install Node.js (if missing).
    - Copy files to `C:\EasyWinGet`.
    - Setup dependencies.
    - Create shortcuts on your Desktop and Start Menu.
5.  Launch **EasyWinGet** from your Desktop!

### Option 2: Portable Mode
Run directly from the folder without installing.

1.  **Download** & Extract the release.
2.  Double-click `run.bat`.
3.  The script will locally check/install dependencies and launch the server.

---

## 📂 Project Structure

```bash
EasyWinGet/
├── gui/                  # Frontend Source
│   ├── index.html        # Main Entry
│   ├── script.js         # Core Logic (UI, API calls)
│   └── style.css         # Styling & Animations
├── server.js             # Backend (Express + Node-PTY)
├── utils/                # Helper Scripts
│   ├── winget.js         # Winget Wrapper
│   └── jobs.js           # Process Manager
├── data/                 # Local Data & Cache
├── offline-packages/     # Bundled Dependencies (Node.js, NPM modules)
├── install.bat           # Automated Installer Script
└── run.bat               # Portable Launcher Script
```

---

## 🛡️ License

This project is open-source. Feel free to modify and distribute.
*Powered by [Winget](https://learn.microsoft.com/en-us/windows/package-manager/winget/).*
