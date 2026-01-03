# EasyWinGet 📦

> **A Modern GUI for Windows Package Manager (WinGet)**

EasyWinGet is a lightweight, modern web interface for managing Windows applications. It wraps the powerful `winget` command-line tool in a beautiful, easy-to-use GUI that works seamlessly.

![EasyWinGet Preview](https://via.placeholder.com/800x400?text=App+Preview)

## ✨ Features

- **🔎 Search & Discover**: Instantly search thousands of apps available on WinGet.
- **📥 One-Click Install**: Install apps with a single click, no terminal needed.
- **🔄 Updates Manager**: Auto-detects outdated apps and allows bulk updating.
- **🎨 Modern UI**: Clean, dark-themed responsive interface.
- **📵 Offline Capable**: Smart installer works even without internet (if packages are bundled).
- **⚡ Smart Server**:
  - Auto-starts server and browser.
  - **Instant Shutdown**: Closes automatically when you close the browser tab.
  - **Auto-Minimize**: Hides console to keep your workspace clean.
- **🛠️ Task Manager**: Real-time progress tracking for installs/updates.

## 🚀 Quick Start

### Standard (Online)
1. Download or clone this repository.
2. Double-click **`run.bat`**.
3. The app will open in your browser automatically!

### Offline Mode 📵
Great for isolated environments or USB distribution!
1. Include the `offline-packages/` folder.
2. (Optional) Add `node-installer.msi` to `offline-packages/` for zero-internet setup.
3. Run **`run.bat`**.
   - It will auto-install Node.js and dependencies from the local folder.

## 📂 Project Structure

```bash
EasyWinGet/
├── run.bat              # 🚀 Universal Launcher (Auto-setup & Run)
├── server.js            # 🧠 Main Backend Server (Express + Node-PTY)
├── gui/                 # 🎨 Frontend UI
│   ├── index.html       #    - Main Interface
│   ├── style.css        #    - Application Styles (Dark Theme)
│   └── script.js        #    - Heartbeat & UI Logic
├── utils/               # ⚙️ Backend Logic
│   ├── jobs.js          #    - PTY/Terminal Job Manager
│   ├── winget.js        #    - WinGet Command Wrapper
│   └── cache.js         #    - JSON File Caching
├── offline-packages/    # 📦 Bundled npm packages (.tgz) & installers
└── Downloads/           # 📂 Downloaded App Installers
```

## 🛠️ Technology Stack

- **Backend**: Node.js, Express.js
- **Terminal**: `node-pty` (runs PowerShell/WinGet in background)
- **Frontend**: Vanilla HTML5, CSS3 (Modern Flexbox/Grid), JavaScript (ES6+)
- **Integration**: WinGet CLI, PowerShell

## 🔧 Requirements

- Windows 10 (1809+) or Windows 11.
- [App Installer](https://www.microsoft.com/p/app-installer/9nblggh4nns1) (Winget).

### Installing WinGet (if missing)
If you don't have WinGet installed (try running `winget` in cmd), install it via:

1. **Microsoft Store**:
   - Search for **"App Installer"** and update/install it.
   - [Direct Link](https://www.microsoft.com/p/app-installer/9nblggh4nns1)

2. **Manual Download**:
   - Download the latest `.msixbundle` from the [official GitHub releases](https://github.com/microsoft/winget-cli/releases).
   - Double-click to install.

## 🤝 Contributing

1. Fork the repo
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

Distributed under the MIT License. See `LICENSE` for more information.

---
**Author:** Jyoti Karmakar
**Version:** 4.5.0
