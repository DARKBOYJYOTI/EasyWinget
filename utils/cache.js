const fs = require('fs');
const path = require('path');

const DATA_DIR = path.join(__dirname, '../data');

// Ensure data dir
if (!fs.existsSync(DATA_DIR)) {
    fs.mkdirSync(DATA_DIR, { recursive: true });
}

const FILES = {
    installed: path.join(DATA_DIR, 'installed.json'),
    updates: path.join(DATA_DIR, 'updates.json'),
    ignored: path.join(DATA_DIR, 'ignored.json'),
    downloads: path.join(DATA_DIR, 'downloads.json')
};

module.exports = {
    FILES,
    save: (key, data) => {
        const file = FILES[key];
        if (!file) throw new Error(`Unknown cache key: ${key}`);
        fs.writeFileSync(file, JSON.stringify(data, null, 2), 'utf8');
        console.log(`[Cache] Saved ${key} (${Array.isArray(data) ? data.length : 'obj'} items)`);
    },
    load: (key) => {
        const file = FILES[key];
        if (!file) throw new Error(`Unknown cache key: ${key}`);
        if (!fs.existsSync(file)) return null;
        try {
            const content = fs.readFileSync(file, 'utf8').replace(/^\uFEFF/, '');
            return JSON.parse(content);
        } catch (e) {
            console.error(`[Cache] Error loading ${key}:`, e);
            return null;
        }
    }
};
