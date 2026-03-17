const fs = require('fs');
const pty = require('node-pty');
const path = require('path');
const crypto = require('crypto'); // Native node module

const JOBS_DIR = path.join(__dirname, '../jobs');
const activeJobs = new Map();

// Auto-cleanup interval (clean completed jobs after 10 minutes)
const JOB_RETENTION_MS = 10 * 60 * 1000;

// Ensure jobs dir exists
if (!fs.existsSync(JOBS_DIR)) {
    fs.mkdirSync(JOBS_DIR, { recursive: true });
}

// Regex to strip ANSI (CSI & OSC)
const stripAnsi = (str) => {
    let clean = str.replace(/[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[a-zA-Z]/g, '');
    clean = clean.replace(/[\u001b\u009b]][^\u0007\u001b]*[\u0007\u001b\\]/g, '');

    // Handle Backspace (\x08) simulation
    while (clean.includes('\x08')) {
        clean = clean.replace(/[^\x08]\x08/, ''); // Remove char + backspace
        clean = clean.replace(/^\x08+/, ''); // Remove leading backspaces (safeguard)
    }
    return clean;
};

function createJobId() {
    return 'job-' + crypto.randomUUID();
}

/**
 * Auto-cleanup completed jobs after retention period.
 * Runs every 2 minutes.
 */
function scheduleCleanup() {
    setInterval(
        () => {
            const now = Date.now();
            for (const [jobId, job] of activeJobs.entries()) {
                if (job.done && job.completedAt && now - job.completedAt > JOB_RETENTION_MS) {
                    console.log(`[Job Manager] Auto-cleaning job ${jobId}`);
                    // Delete log file
                    try {
                        if (fs.existsSync(job.logFile)) {
                            fs.unlinkSync(job.logFile);
                        }
                    } catch (e) {}
                    activeJobs.delete(jobId);
                }
            }
        },
        2 * 60 * 1000
    );
}

// Start the cleanup scheduler
scheduleCleanup();

module.exports = {
    startJob: (command, args) => {
        const jobId = createJobId();
        const logFile = path.join(JOBS_DIR, `${jobId}.log`);

        console.log(`[Job Manager] Starting Job ${jobId}: ${command} ${args.join(' ')}`);

        // Detect if this is a PowerShell command
        const isPowerShell = command.toLowerCase().includes('powershell');

        // Always use PowerShell to ensure consistent argument handling
        const psPath = 'C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe';

        let ptyProcess;

        if (isPowerShell) {
            // Native PowerShell script/command
            ptyProcess = pty.spawn(psPath, args, {
                name: 'xterm-color',
                cols: 120,
                rows: 30,
                cwd: process.cwd(),
                env: process.env,
            });
        } else {
            // General commands (winget, etc.) wrapped in PowerShell
            // Escape arguments for PowerShell (single quotes for literals)
            const psArgs = args.map((a) => {
                if (a.includes(' ') || a.includes('(') || a.includes(')')) return `'${a}'`;
                return a;
            });

            const fullCmd = `${command} ${psArgs.join(' ')}`;

            // Using -Command allows normal execution
            ptyProcess = pty.spawn(psPath, ['-NoProfile', '-Command', fullCmd], {
                name: 'xterm-color',
                cols: 120,
                rows: 30,
                cwd: process.cwd(),
                env: process.env,
            });
        }

        const logStream = fs.createWriteStream(logFile, { flags: 'a', encoding: 'utf8' });

        ptyProcess.onData((data) => {
            let clean = stripAnsi(data);

            // Aggressively remove spinner artifacts
            clean = clean.replace(/(\r[-\\|\/]+)|([-\\|\/]+\r)/g, '');
            clean = clean.replace(/^[-\\|\/]\r/gm, '');
            clean = clean.replace(/^\s*\\\s*$/gm, '');
            clean = clean.replace(/^\s*[-\\|\/]\s*$/gm, '');

            if (clean.length > 0) {
                logStream.write(clean);
            }
        });

        const jobData = {
            id: jobId,
            process: ptyProcess,
            logFile: logFile, // absolute path
            startTime: new Date(),
            done: false,
            exitCode: null,
            completedAt: null,
        };

        activeJobs.set(jobId, jobData);

        ptyProcess.onExit((res) => {
            console.log(`[Job Manager] Job ${jobId} exited with code ${res.exitCode}`);
            jobData.done = true;
            jobData.exitCode = res.exitCode;
            jobData.completedAt = Date.now();
            logStream.end();
        });

        return jobId;
    },

    getJob: (jobId) => {
        return activeJobs.get(jobId);
    },

    getJobOutput: (jobId) => {
        const job = activeJobs.get(jobId);
        if (!job) return null;

        // Read file content
        try {
            if (fs.existsSync(job.logFile)) {
                return fs.readFileSync(job.logFile, 'utf8');
            }
        } catch (e) {
            console.error(`Error reading log for ${jobId}:`, e);
        }
        return '';
    },

    cleanupJob: (jobId) => {
        const job = activeJobs.get(jobId);
        if (job) {
            try {
                if (fs.existsSync(job.logFile)) {
                    fs.unlinkSync(job.logFile);
                }
            } catch (e) {}
        }
        activeJobs.delete(jobId);
    },

    cancelJob: (jobId) => {
        const job = activeJobs.get(jobId);
        if (job && job.process) {
            console.log(`[Job Manager] Killing job ${jobId}`);
            try {
                job.process.kill();
                job.done = true;
                job.completedAt = Date.now();
                return true;
            } catch (e) {
                console.error(`[Job Manager] Error killing job ${jobId}:`, e);
                return false;
            }
        }
        return false;
    },
};
