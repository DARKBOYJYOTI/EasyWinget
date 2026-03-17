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
npm start          # or: node server.js
```

Windows helpers:

- `run.bat`: elevate + install deps if needed + start server.
- `install.bat`: install to `C:\EasyWinGet` + create shortcuts.

Lint (ESLint 9 flat config + eslint-config-prettier):

```sh
npm run lint           # report only
npm run lint:fix       # auto-fix
```

Format (Prettier):

```sh
npm run format         # auto-format all files
npm run format:check   # check only (CI-friendly)
```

Quick syntax check (works on any OS):

```sh
node --check server.js
node --check utils/winget.js
```

Tests: none currently (`npm test` intentionally exits 1).
If adding tests, prefer Node's built-in runner:

```sh
node --test                                                    # all tests
node --test test/winget.test.js                                # single file
node --test --test-name-pattern="parseUpdates" test/winget.test.js  # single test by name
```

## Cursor / Copilot Rules

- No `.cursorrules` or `.cursor/rules/` found.
- No `.github/copilot-instructions.md` found.

## Code Style

### Language & Modules

- CommonJS (`"type": "commonjs"` in `package.json`); use `require(...)` and `module.exports`.
- Target Node 18+ (web globals like `fetch`, `AbortSignal`, `TextDecoder` are available).

### Formatting (enforced by Prettier)

- 4-space indentation, no tabs.
- Semicolons required.
- Single quotes in JS files.
- Trailing commas: `es5` (arrays, objects — not function params).
- Print width: 100 columns.
- Arrow parens: always (`(x) => ...`).
- Line endings: LF.
- Bracket spacing enabled: `{ foo }`.

### Imports / Requires

- Group in order: Node built-ins, external deps, local modules.
- Prefer `const`; destructure only when it improves clarity.

### Naming

- `camelCase` for variables and functions.
- `PascalCase` for singleton-ish objects (e.g., `State`, `DOM` in frontend).
- `UPPER_SNAKE_CASE` for true constants (e.g., `PORT`, `DOWNLOAD_DIR`).

### ESLint Rules

- `no-unused-vars`: warn only; `args: 'none'`, `caughtErrors: 'none'`, vars prefixed `_` are ignored.
- `no-control-regex`: off (intentional ANSI/control-char handling).
- `no-empty`: error, but empty `catch` blocks are allowed (`allowEmptyCatch: true`).
- `no-useless-escape`: off (CLI output cleanup patterns).
- Backend files (`server.js`, `utils/**`) use `sourceType: 'commonjs'` with Node globals.
- Frontend files (`gui/**`) use `sourceType: 'script'` with browser globals.

### Error Handling

- Log unexpected backend errors with a clear prefix (e.g., `[Manifest Error]`, `[Details Error]`).
- Avoid empty `catch (e) {}` unless best-effort and tightly scoped; do not silently swallow primary failures.
- If you must swallow, do it only around non-critical cleanup (cache delete, reader cancel, optional HEAD checks).

### Backend (Express) Conventions

- Response shape: `res.json({ success: true, ... })` or `res.status(code).json({ success: false, error/message: ... })`.
- Always `return` after writing a response to avoid double-sends.
- Validate/normalize `req.query`/`req.body` (IDs, filenames, paths); block path traversal (`..`).
- Prefer existing routing style in `server.js` (single file) unless doing a larger refactor.
- When spawning processes, prefer `utils/jobs.js` or `execFile`-style APIs; avoid interpolating user input into shell strings.

### Jobs (Long-Running Commands)

- Use `utils/jobs.js` for install/upgrade/download/uninstall so output is captured and pollable.
- Pass args as an array; avoid assembling shell strings.
- Cancel via `jobs.cancelJob(jobId)`. Job IDs are `job-<uuid>`; logs go to `jobs/<jobId>.log`.

### Winget Wrapper

- `utils/winget.js` parses `winget` output using column positions; keep parsing tolerant to spacing/header variations.
- Preserve `chcp 65001` usage for UTF-8 output.
- Winget can return non-zero exit codes for non-fatal states; treat `stdout` as the primary signal.

### Cache

- `utils/cache.js` writes JSON to `data/`: `installed.json`, `updates.json`, `ignored.json`, `downloads.json`.
- Cache invalidation is sometimes done by deleting a file (see uninstall flow in `server.js`).

### Filesystem Safety

- Sanitize user-derived path segments (remove `<>:"/\|?*`, trim); reject `..` and absolute paths.
- Keep file operations scoped under the intended base directory (e.g., `Downloads/`, `data/`).
- Use `path.join(base, segment)` and validate `segment` before joining.

### Frontend (Vanilla JS)

- Keep global state in `State` and DOM references in `DOM` (`gui/script.js`).
- Use `apiCall(endpoint)` for fetches (consistent errors + cache-busting) and `showToast(...)` for user-visible failures.
- UI expects `success` flags and sometimes coerces `apps/updates/files` into arrays; keep responses consistent.

## Development Workflow

1. Make a small change in `server.js` or `utils/*`.
2. Run `node --check <file>` to catch syntax errors fast.
3. Run `npm run lint` and `npm run format:check` before committing.
4. On Windows: run `node server.js` and exercise the relevant UI flow.
5. If you touched WinGet parsing: validate against real `winget` output samples (spacing and headers vary).

## Adding Tests

- No `test/` directory exists yet; create `test/` with `*.test.js` files as needed.
- Use Node's built-in test runner (`node:test` + `node:assert/strict`) — no Jest/Vitest dependency.
- Keep tests deterministic: unit-test parsers (`parseApps`, `parseUpdates`) and helpers.
- If you want parser unit tests, export the pure parse functions from `utils/winget.js`.

## Dependencies & Offline Packaging

- Runtime deps are vendored as `.tgz` under `offline-packages/`.
- If adding a runtime dep, consider offline distribution (update `offline-packages/README.md`, optionally vendor a `.tgz`).
