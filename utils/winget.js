const { execFile } = require('child_process');

/**
 * Strips ANSI codes and splits into lines
 */
function cleanOutput(raw) {
    // Regex to strip ANSI (CSI & OSC)
    let clean = raw.replace(/[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[a-zA-Z]/g, '');
    clean = clean.replace(/[\u001b\u009b]][^\u0007\u001b]*[\u0007\u001b\\]/g, '');

    // Aggressively remove spinner artifacts (from jobs.js findings)
    clean = clean.replace(/(\r[-\\|\/]+)|([-\\|\/]+\r)/g, '');
    clean = clean.replace(/^[-\\|\/]\r/gm, '');

    // Split lines
    return clean
        .split(/\r?\n/)
        .map((l) => l.replace(/^\s+/, '').trimEnd())
        .filter((l) => l.length > 0);
}

/**
 * Parses 'winget search' or 'winget list' output using column positions
 */
function parseApps(output) {
    const lines = cleanOutput(output);
    const items = [];

    let colId = -1;
    let colVersion = -1;

    for (const line of lines) {
        // Detect Header
        if (colId === -1 && line.includes('Name') && line.includes('Id')) {
            const nameIdx = line.indexOf('Name');
            colId = line.indexOf('Id') - nameIdx;
            colVersion = line.indexOf('Version') - nameIdx;
            continue;
        }

        // Skip separators or invalid lines
        if (colId === -1 || line.startsWith('---') || line.startsWith('Name')) continue;

        // Strict column parsing
        // If line is shorter than colId, it's invalid
        if (line.length < colId) continue;

        const name = line.substring(0, colId).trim();
        let id = '';
        let version = '';

        if (colVersion > -1 && line.length > colVersion) {
            id = line.substring(colId, colVersion).trim();
            const rest = line.substring(colVersion).trim();
            version = rest.split(/\s+/)[0];
        } else {
            // Fallback if Version column not found or line short
            id = line.substring(colId).trim();
        }

        if (name && id) {
            items.push({ name, id, version });
        }
    }
    return items;
}

/**
 * Parses 'winget upgrade' output using column positions
 */
function parseUpdates(output) {
    const lines = cleanOutput(output);
    const items = [];

    let colId = -1;
    let colVersion = -1; // "Version" (Current)
    let colAvailable = -1; // "Available"

    for (const line of lines) {
        // Detect Header
        if (colId === -1 && line.includes('Name') && line.includes('Id')) {
            const nameIdx = line.indexOf('Name');
            colId = line.indexOf('Id') - nameIdx;
            colVersion = line.indexOf('Version') - nameIdx;
            colAvailable = line.indexOf('Available') - nameIdx;
            continue;
        }

        if (colId === -1 || line.startsWith('---') || line.startsWith('Name')) continue;

        // Skip footer/summary lines (e.g., "1 upgrades available.")
        if (line.toLowerCase().includes('upgrade') && line.includes('available')) continue;
        if (
            line.toLowerCase().includes('no applicable') ||
            line.toLowerCase().includes('no package')
        )
            continue;

        if (line.length < colId) continue;

        const name = line.substring(0, colId).trim();
        let id = '';
        let current = '';
        let available = '';

        if (colVersion > -1 && line.length > colVersion) {
            id = line.substring(colId, colVersion).trim();

            if (colAvailable > -1 && line.length > colAvailable) {
                current = line.substring(colVersion, colAvailable).trim();
                available = line.substring(colAvailable).split(/\s+/)[0];
            } else {
                current = line.substring(colVersion).trim();
            }
        } else {
            id = line.substring(colId).trim();
        }

        // Skip if ID is empty, too short, or looks like a sentence fragment
        if (!id || id.length < 3 || id.includes(' ')) continue;

        if (name && id) {
            items.push({ name, id, current, version: available });
        }
    }
    return items;
}

/**
 * Executes a winget command safely using execFile with array args.
 * Returns a Promise resolving to stdout.
 */
function invoke(args) {
    return new Promise((resolve, reject) => {
        // Use cmd.exe to run chcp first for UTF-8, then winget
        // We pass args as array to avoid shell injection
        const cmdArgs = ['/c', 'chcp', '65001', '>', 'nul', '&&', 'winget', ...args];
        execFile(
            'cmd.exe',
            cmdArgs,
            { encoding: 'utf8', maxBuffer: 10 * 1024 * 1024 },
            (err, stdout) => {
                // Winget often returns non-zero codes even for partial successes
                if (err && !stdout) {
                    return reject(err);
                }
                resolve(stdout || '');
            }
        );
    });
}

module.exports = {
    search: async (query) => {
        const out = await invoke(['search', query, '-s', 'winget', '--accept-source-agreements']);
        return parseApps(out);
    },
    listInstalled: async () => {
        const out = await invoke(['list', '--accept-source-agreements']);
        return parseApps(out);
    },
    listUpdates: async () => {
        const out = await invoke(['upgrade', '--include-unknown', '--accept-source-agreements']);
        return parseUpdates(out);
    },
    // Export parsers for testing
    parseApps,
    parseUpdates,
    // Raw invoke for debug/custom
    invoke,
};
