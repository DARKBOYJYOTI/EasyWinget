const express = require('express');
const cors = require('cors');
const path = require('path');
const fs = require('fs');
const { open } = require('fs/promises'); // For async file checks if needed, though mostly using sync for simple logic or callbacks

const winget = require('./utils/winget');
const jobs = require('./utils/jobs');
const cache = require('./utils/cache');

const app = express();
const PORT = 8080;

app.use(cors());
app.use(express.json());

// --- HEARTBEAT & AUTO-SHUTDOWN ---
let lastHeartbeat = Date.now();
const HEARTBEAT_TIMEOUT = 30000; // 30 seconds (much more stable)
const GRACE_PERIOD = 10000;    // 10 seconds startup grace

// Allow client to send keepalive signal
app.post('/api/keepalive', (req, res) => {
    lastHeartbeat = Date.now();
    res.json({ success: true });
});

// INSTANT SHUTDOWN signal
app.post('/api/shutdown', (req, res) => {
    console.log('Shutdown signal received. Exiting immediately...');
    process.exit(0);
});

// Check for inactivity
setTimeout(() => {
    console.log('Heartbeat monitor started...');
    setInterval(() => {
        if (Date.now() - lastHeartbeat > HEARTBEAT_TIMEOUT) {
            console.log('Client disconnected (timeout). Shutting down...');
            process.exit(0);
        }
    }, 2000); // Check every 2s
}, GRACE_PERIOD);

// Static Files
app.use(express.static(path.join(__dirname, 'gui')));

// Download Directory
const DOWNLOAD_DIR = path.join(__dirname, 'Downloads');
if (!fs.existsSync(DOWNLOAD_DIR)) {
    fs.mkdirSync(DOWNLOAD_DIR, { recursive: true });
}

// --- VERSION ---
app.get('/version.json', (req, res) => {
    res.sendFile(path.join(__dirname, 'version.json'));
});

// --- INSTALLED ---
app.get('/api/installed', async (req, res) => {
    const cached = cache.load('installed');
    if (cached) {
        return res.json({ success: true, apps: cached });
    }
    // Fallback to refresh
    try {
        const apps = await winget.listInstalled();
        cache.save('installed', apps);
        res.json({ success: true, apps });
    } catch (e) {
        res.status(500).json({ success: false, error: e.message });
    }
});

app.get('/api/refresh-installed', async (req, res) => {
    try {
        const apps = await winget.listInstalled();
        cache.save('installed', apps);
        res.json({ success: true, apps });
    } catch (e) {
        res.status(500).json({ success: false, error: e.message });
    }
});

// --- UPDATES ---
app.get('/api/updates', async (req, res) => {
    const cached = cache.load('updates');
    const ignored = cache.load('ignored') || [];
    const ignoredIds = (Array.isArray(ignored) ? ignored : []).map(i => i.id);

    if (cached) {
        const filtered = cached.filter(u => !ignoredIds.includes(u.id));
        return res.json({ success: true, updates: filtered });
    }

    try {
        const updates = await winget.listUpdates();
        cache.save('updates', updates);
        const filtered = updates.filter(u => !ignoredIds.includes(u.id));
        res.json({ success: true, updates: filtered });
    } catch (e) {
        res.status(500).json({ success: false, error: e.message });
    }
});

app.get('/api/refresh-updates', async (req, res) => {
    try {
        const updates = await winget.listUpdates();
        cache.save('updates', updates);

        const ignored = cache.load('ignored') || [];
        const ignoredIds = (Array.isArray(ignored) ? ignored : []).map(i => i.id);
        const filtered = updates.filter(u => !ignoredIds.includes(u.id));

        res.json({ success: true, updates: filtered });
    } catch (e) {
        res.status(500).json({ success: false, error: e.message });
    }
});

// --- IGNORED ---
app.get('/api/ignored', (req, res) => {
    const ignored = cache.load('ignored') || [];
    res.json({ success: true, apps: ignored });
});

app.get('/api/ignore', (req, res) => {
    const { id, name } = req.query;
    if (!id) return res.status(400).json({ success: false, message: 'No ID provided' });

    let ignored = cache.load('ignored') || [];
    if (!ignored.find(i => i.id === id)) {
        ignored.push({ id, name });
        cache.save('ignored', ignored);
    }
    res.json({ success: true, message: `Ignored ${id}` });
});

app.get('/api/unignore', (req, res) => {
    const { id } = req.query;
    if (!id) return res.status(400).json({ success: false, message: 'No ID provided' });

    let ignored = cache.load('ignored') || [];
    ignored = ignored.filter(i => i.id !== id);
    cache.save('ignored', ignored);
    res.json({ success: true, message: `Unignored ${id}` });
});

// --- SEARCH ---
app.get('/api/search', async (req, res) => {
    const { q } = req.query;
    if (!q) return res.json({ success: true, results: [] });

    try {
        const results = await winget.search(q);
        res.json({ success: true, results });
    } catch (e) {
        res.status(500).json({ success: false, error: e.message });
    }
});

// --- JOBS (INSTALL/UNINSTALL/ETC) ---
app.get('/api/install', (req, res) => {
    const { id } = req.query;
    if (!id) return res.status(400).json({ success: false, message: 'No ID provided' });

    // args: install --id <id> ...
    const jobId = jobs.startJob('winget', ['install', '--id', id, '--accept-source-agreements', '--accept-package-agreements', '--verbose']);
    res.json({ success: true, jobId });
});

app.get('/api/uninstall', (req, res) => {
    const { id, name } = req.query;
    if (!id) return res.status(400).json({ success: false, message: 'No ID provided' });

    // Use native Windows uninstall via registry (more reliable than winget for some apps)
    const scriptPath = path.join(__dirname, 'utils', 'native-uninstall.ps1');

    // Use the app name directly - it's more reliable than parsing IDs
    // The name comes from the GUI and is the actual display name
    const searchName = name || id.split('.').pop();

    const jobId = jobs.startJob('powershell', ['-ExecutionPolicy', 'Bypass', '-File', scriptPath, '-AppName', searchName, '-PackageId', id]);

    // Invalidate installed cache
    try {
        fs.unlinkSync(cache.FILES.installed);
    } catch (e) { }

    res.json({ success: true, jobId });
});

app.get('/api/update', (req, res) => {
    const { id } = req.query;
    if (!id) return res.status(400).json({ success: false, message: 'No ID provided' });

    const jobId = jobs.startJob('winget', ['upgrade', '--id', id, '--accept-source-agreements', '--accept-package-agreements', '--verbose']);
    res.json({ success: true, jobId });
});

app.get('/api/download', (req, res) => {
    const { id, name } = req.query;
    if (!id) return res.status(400).json({ success: false, message: 'No ID provided' });

    // Determine download directory
    let targetDir = DOWNLOAD_DIR;
    if (name) {
        // Sanitize name for folder
        const safeName = name.replace(/[<>:"/\\|?*]/g, '').trim();
        if (safeName) {
            targetDir = path.join(DOWNLOAD_DIR, safeName);
        }
    }

    if (!fs.existsSync(targetDir)) {
        fs.mkdirSync(targetDir, { recursive: true });
    }

    const jobId = jobs.startJob('winget', ['download', '--id', id, '--download-directory', targetDir, '--accept-source-agreements', '--verbose']);
    res.json({ success: true, jobId });
});

app.get('/api/status', (req, res) => {
    const { id } = req.query;
    if (!id) return res.status(400).json({ success: false, error: 'No Job ID' });

    const job = jobs.getJob(id);
    if (!job) return res.status(404).json({ success: false, error: 'Job not found' });

    const output = jobs.getJobOutput(id) || "";

    res.json({
        success: job.done && job.exitCode === 0,
        done: job.done,
        output: output
    });
});

// --- DOWNLOADED FILES ---
function getFilesRecursive(dir, baseDir = dir) {
    let results = [];
    try {
        const list = fs.readdirSync(dir);
        list.forEach(file => {
            // Skip "Dependencies" folder
            if (file.toLowerCase() === 'dependencies') return;

            const fullPath = path.join(dir, file);
            try {
                const stat = fs.statSync(fullPath);
                if (stat && stat.isDirectory()) {
                    results = results.concat(getFilesRecursive(fullPath, baseDir));
                } else {
                    if (/\.(exe|msi|zip|7z|rar|iso|msix|appx)$/i.test(file)) {
                        results.push({
                            Name: path.relative(baseDir, fullPath),
                            Length: stat.size,
                            // Ensure valid ISO string
                            LastWriteTime: stat.mtime ? stat.mtime.toISOString() : new Date().toISOString()
                        });
                    }
                }
            } catch (e) { }
        });
    } catch (e) { }
    return results;
}

app.get('/api/downloaded', (req, res) => {
    try {
        const files = getFilesRecursive(DOWNLOAD_DIR);
        cache.save('downloads', files);
        res.json({ success: true, files });
    } catch (e) {
        res.status(500).json({ success: false, error: e.message });
    }
});

app.get('/api/downloaded/delete', (req, res) => {
    const { file } = req.query;
    if (!file) return res.status(400).json({ success: false, message: 'No file' });

    // Check if file is in a subfolder (e.g. "AppName/file.exe")
    const parts = file.split(path.sep);
    const isSubfolder = parts.length > 1;

    try {
        if (isSubfolder) {
            // Delete the parent folder (e.g. "Downloads/AppName")
            // Security check: ensure no '..' to escape downloads
            if (file.includes('..')) throw new Error('Invalid path');

            const folderName = parts[0];
            const folderPath = path.join(DOWNLOAD_DIR, folderName);

            if (fs.existsSync(folderPath)) {
                fs.rmSync(folderPath, { recursive: true, force: true });
            }
        } else {
            // Just delete the file
            const filePath = path.join(DOWNLOAD_DIR, file);
            if (fs.existsSync(filePath)) {
                fs.unlinkSync(filePath);
            }
        }

        // Refresh cache
        const files = getFilesRecursive(DOWNLOAD_DIR);
        cache.save('downloads', files);

        res.json({ success: true, message: `Deleted ${file}` });
    } catch (e) {
        res.status(500).json({ success: false, message: e.message });
    }
});

app.get('/api/downloaded/run', (req, res) => {
    const { file } = req.query;
    if (!file) return res.status(400).json({ success: false, message: 'No file' });

    const filePath = path.join(DOWNLOAD_DIR, file);
    if (fs.existsSync(filePath)) {
        // Use PowerShell Start-Process to handle MSIX, UAC, etc. nicely
        // and wrap it in a job to provide feedback
        const jobId = jobs.startJob('powershell', ['-Command', 'Start-Process', '-FilePath', `'${filePath}'`, '-PassThru']);
        res.json({ success: true, jobId, message: `Launching ${file}` });
    } else {
        res.status(404).json({ success: false, message: 'File not found' });
    }
});

// Start Server
app.listen(PORT, () => {
    console.log(`EasyWinGet Node Server running at http://localhost:${PORT}`);
    require('child_process').exec(`start http://localhost:${PORT}`);
});
