# EasyWinGet Agent Notes

This repo is a Windows-first Node/Express app with a vanilla HTML/CSS/JS UI.
The backend shells out to `winget` and PowerShell and manages long-running operations via a PTY job manager.
Non-Windows environments can run basic Node syntax checks, but most runtime flows (`winget`, `powershell`, `start`) will fail.

## Repo Map

- `server.js`: Express server, API routes, static hosting for `gui/`.
- `utils/winget.js`: wraps `winget` CLI and parses column output.
- `utils/jobs.js`: long-running job manager (`node-pty`) and log files under `jobs/`.
- `utils/cache.js`: JSON cache persisted under `data/`.
- `gui/`: frontend (`index.html`, `script.js`, `style.css`).
- `offline-packages/`: bundled `.tgz` deps and optional `node-installer.msi`.

## Commands

Install deps (online):

```sh
npm install
```

Install deps (offline, matches `run.bat`/`install.bat`):

```bat
npm install offline-packages\express-5.2.1.tgz offline-packages\cors-2.8.5.tgz offline-packages\node-pty-1.1.0.tgz
```

Run server:

```sh
npm start
# or
node server.js
```

Windows helpers:

- `run.bat`: elevate + install deps if needed + start server.
- `install.bat`: install to `C:\EasyWinGet` + create shortcuts.

Build/lint: none configured.
Quick checks:

```sh
node --check server.js
```

Tests: none currently (`npm test` intentionally exits 1).
If adding tests, prefer Node's built-in runner:

```sh
node --test
node --test test/winget.test.js
node --test --test-name-pattern="parseUpdates" test/winget.test.js
```

## Runtime Notes (Windows)

- Requires Windows 10/11 with WinGet available (`winget` CLI).
- Many actions assume admin/elevation (see `run.bat`, `install.bat`).
- Server listens on `127.0.0.1:8080` and `server.js` tries to open a browser via `start`.
- `utils/jobs.js` hardcodes PowerShell at `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`.
- Long-running operations write logs under `jobs/` and are queried via `/api/status?id=...`.

## Development Workflow (Recommended)

1. Make a small change in `server.js` or `utils/*`.
2. Run `node --check server.js` (and check any edited file) to catch syntax errors fast.
3. Run `node server.js` on Windows and exercise the relevant UI flow.
4. If you touched WinGet parsing: validate against real `winget` output samples (spacing and headers vary).

## Adding Tests (Guide)

- There is no `test/` directory today; feel free to create `test/` with `*.test.js`.
- Use Node's built-in runner (no Jest/Vitest dependency needed).
- Keep tests deterministic: prefer unit tests around parsers (`parseApps`, `parseUpdates`) and helpers.
- Suggested structure:

```js
// test/winget.test.js
const test = require('node:test');
const assert = require('node:assert/strict');
const winget = require('../utils/winget');

test('parseUpdates handles header + rows', async () => {
    const out =
        'Name  Id  Version  Available\n----  --  -------  ---------\nApp  Foo.Bar  1.0  2.0';
    const updates = winget.__parseUpdatesForTest ? winget.__parseUpdatesForTest(out) : [];
    assert.ok(Array.isArray(updates));
});
```

(If you want parser unit tests, consider exporting the pure parse functions explicitly.)

## Cursor / Copilot Rules

- No `.cursorrules` or `.cursor/rules/` found.
- No `.github/copilot-instructions.md` found.

## Conventions

Language/modules:

- CommonJS (`package.json` has `"type": "commonjs"`); use `require(...)` and `module.exports`.

Formatting:

- 4-space indentation; semicolons; K&R braces.
- Prefer single quotes in Node files.

Imports / requires:

- Group in order: Node built-ins, external deps, local modules.
- Prefer `const`; destructure only when it improves clarity.

Naming:

- `camelCase` for vars/functions; `PascalCase` for singleton-ish objects (e.g., `State`, `DOM`).
- `UPPER_SNAKE_CASE` for true constants (e.g., `PORT`, `DOWNLOAD_DIR`).

Backend (Express):

- Response shape: `res.json({ success: true, ... })` or `res.status(code).json({ success: false, error/message: ... })`.
- Always `return` after writing a response (avoid double-sends in callbacks/guards).
- Validate/normalize `req.query`/`req.body` (IDs, filenames, paths); block obvious traversal (`..`).
- Prefer the existing routing style in `server.js` (single file) unless doing a larger refactor.
- When spawning processes, prefer `utils/jobs.js` or `execFile`-style APIs; avoid interpolating user input into shell strings.

Error handling/logging:

- Log unexpected backend errors with a clear prefix (e.g., `[Manifest Error]`).
- Avoid empty `catch (e) {}` unless best-effort and tightly scoped; do not silently swallow primary failures.
- If you must swallow, do it only around non-critical cleanup (cache delete, reader cancel, optional HEAD checks).

Jobs (long-running commands):

- Use `utils/jobs.js` for install/upgrade/download/uninstall so output is captured and pollable.
- Pass args as an array; avoid assembling shell strings (quoting is handled in the job wrapper).
- Cancel via `jobs.cancelJob(jobId)`.
- Job IDs are generated as `job-<uuid>`; logs are appended to `jobs/<jobId>.log`.

Winget wrapper:

- `utils/winget.js` parses `winget` output using column positions; keep parsing tolerant to spacing/headers.
- Preserve `chcp 65001` usage for UTF-8 output.
- Winget can return non-zero exit codes for non-fatal states; treat `stdout` as the primary signal.

Cache:

- `utils/cache.js` writes JSON to `data/`:
    - `installed.json`, `updates.json`, `ignored.json`, `downloads.json`.
- Cache invalidation is sometimes done by deleting a file (see uninstall flow in `server.js`).

Filesystem safety:

- Sanitize user-derived path segments (remove `<>:"/\\|?*`, trim); reject `..` and absolute paths.
- Keep file operations scoped under the intended base directory (e.g., `Downloads/`, `data/`).
- Prefer `path.join(base, segment)` and validate `segment` before joining.

Frontend (vanilla JS):

- Keep global state in `State` and DOM references in `DOM` (`gui/script.js`).
- Use `apiCall(endpoint)` for fetches (consistent errors + cache-busting) and `showToast(...)` for user-visible failures.
- UI expects `success` flags and sometimes coerces `apps/updates/files` into arrays; keep responses consistent.
- Prefer progressive rendering and lazy icon loading (existing `IntersectionObserver` pattern).

Dependencies / offline packaging:

- Runtime deps are vendored as `.tgz` under `offline-packages/`.
- If adding a runtime dep, consider offline distribution (update `offline-packages/README.md`, optionally vendor a `.tgz`).
