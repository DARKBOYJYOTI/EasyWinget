# Offline Packages for EasyWinGet

This folder contains pre-downloaded packages for **complete offline installation**.

## 📦 Included Packages:
- `express-5.2.1.tgz` - Web server framework
- `cors-2.8.5.tgz` - Cross-Origin Resource Sharing middleware  
- `node-pty-1.1.0.tgz` - Terminal emulation for running WinGet commands

## 🔧 Optional: Node.js Offline Installer

To enable **completely offline** installation (no internet needed):

1. Download Node.js installer:
   ```
   https://nodejs.org/dist/v22.13.1/node-v22.13.1-x64.msi
   ```

2. Save it as: `node-installer.msi` in this folder

3. When `run.bat` is executed, it will automatically install Node.js from this file if Node.js is not found!

## 🚀 How It Works:

When you run `run.bat`, it will:
1. **Check for Node.js** - Install from `node-installer.msi` if missing (offline)
2. **Try offline npm packages** - Install from .tgz files (no internet needed)
3. **Fallback to online** - Use winget/npm if offline files missing

## 📥 For Distribution:

Include this `offline-packages` folder with:
- ✅ npm packages (.tgz files) - Already included
- ⚠️ `node-installer.msi` - Download and add manually (~35 MB)

## 🔄 Updating Packages:

**Refresh npm packages:**
```batch
npm pack express cors node-pty --pack-destination offline-packages
```

**Update Node.js installer:**
- Download latest from: https://nodejs.org/en/download
- Rename to `node-installer.msi`

---
**Note:** With all offline files, EasyWinGet can be installed on systems with **zero internet access**!
