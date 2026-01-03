const { exec } = require('child_process');
const fs = require('fs');
const path = require('path');

// Regex patterns ported from parser.ps1
// Note: JS regex is slightly different, but the logic remains the same.

// Updates: Name  Id  Version  Available
const RGX_UPDATE = /^(.+?)\s{2,}([\w\.\-]+)\s{2,}(\S+)\s{2,}(\S+)/;
// Installed/Search: Name  Id  Version
const RGX_APP = /^(.+?)\s{2,}([a-zA-Z0-9\.\-\{\}_]+)\s{2,}([vV]?\d[^\s]*)/;

/**
 * Strips ANSI codes and splits into lines
 */
/**
 * Strips ANSI codes and splits into lines
 */
function cleanOutput(raw) {
    // Regex to strip ANSI (CSI & OSC)
    let clean = raw.replace(/[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[a-zA-Z]/g, '');
    clean = clean.replace(/[\u001b\u009b]][^\u0007\u001b]*[\u0007\u001b\\]/g, '');

    // Aggressively remove spinner artifacts (from jobs.js findings)
    clean = clean.replace(/(\r[\-\\\|\/]+)|([\-\\\|\/]+\r)/g, '');
    clean = clean.replace(/^[\-\\\|\/]\r/gm, '');

    // Split lines
    return clean.split(/\r?\n/).map(l => l.replace(/^\s+/, '').trimEnd()).filter(l => l.length > 0);
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
        let id = "";
        let version = "";

        if (colVersion > -1 && line.length > colVersion) {
            id = line.substring(colId, colVersion).trim();
            // Version is from colVersion to end (or next column like Match/Source)
            // But we can usually take the rest or split by space for the first token?
            // "Version" column often has just the version.
            // But search has "Match" and "Source" after.
            // Let's take the version as the next block of non-space text after colVersion?
            // Or better: find the next gap?
            // Actually, for search, Version is followed by Match/Source.
            // Let's assume Version is space-delimited if we don't have colMatch.
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
        if (line.toLowerCase().includes('no applicable') || line.toLowerCase().includes('no package')) continue;

        if (line.length < colId) continue;

        const name = line.substring(0, colId).trim();
        let id = "";
        let current = "";
        let available = "";

        if (colVersion > -1 && line.length > colVersion) {
            id = line.substring(colId, colVersion).trim();

            if (colAvailable > -1 && line.length > colAvailable) {
                current = line.substring(colVersion, colAvailable).trim();
                // Available is remainder, maybe followed by Source? Usually empty for upgrade list
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
 * Executes a winget command (synchronous-like, returns Promise)
 */
function invoke(args) {
    return new Promise((resolve, reject) => {
        // Ensure UTF-8 execution
        const cmd = `chcp 65001 > nul && winget ${args}`;
        exec(cmd, { encoding: 'utf8', maxBuffer: 10 * 1024 * 1024 }, (err, stdout, stderr) => {
            // Winget often returns non-zero codes even for partial successes (like "no updates found")
            // So we generally resolve stdout unless it's a catastrophic failure
            if (err && !stdout) {
                return reject(err);
            }
            resolve(stdout || "");
        });
    });
}

module.exports = {
    search: async (query) => {
        const out = await invoke(`search "${query}" -s winget --accept-source-agreements`);
        return parseApps(out);
    },
    listInstalled: async () => {
        const out = await invoke(`list --accept-source-agreements`);
        return parseApps(out);
    },
    listUpdates: async () => {
        const out = await invoke(`upgrade --include-unknown --accept-source-agreements`);
        return parseUpdates(out);
    },
    // Raw invoke for debug/custom
    invoke
};
