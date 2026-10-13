<script>
  import { onMount } from 'svelte';
  import { 
    Settings, 
    Moon, 
    Sun, 
    Laptop,
    Trash2, 
    UploadCloud, 
    FileSpreadsheet, 
    Info, 
    Check, 
    HardDrive, 
    Power,
    Palette,
    Bell
  } from '@lucide/svelte';
  import { 
    ClearCache, 
    SelectImportPath, 
    ImportApps, 
    SelectExportPath, 
    ExportApps,
    GetSystemAccentColor,
    GetAutoStart,
    SetAutoStart,
    GetSetting,
    SetSetting,
    ApplyAutoUpdateSettings
  } from '../../wailsjs/go/main/App.js';

  export let version = '5.0.0';
  export let themePreference = 'dark'; // 'dark' | 'light' | 'auto'
  export let currentTheme = 'dark';
  export let accentColor = '#0078d4';
  export let accentSource = 'default'; // 'default' or 'windows'
  export let onThemeChange = (pref) => {};
  export let onAccentChange = (color, source) => {};
  export let onStartJob = (jobId, title) => {};
  export let onShowToast = (type, title, msg) => {};
  export let onConfirmRequest = async (opts) => true;
  export let onNavigateTab = (tab) => {};

  let isClearing = false;
  let autoStartEnabled = true;
  let minimizeToTrayEnabled = true;
  let autoUpdateCheckEnabled = true;
  let updateCheckInterval = '1440';
  let windowsAccent = '#0078d4';

  onMount(async () => {
    try {
      windowsAccent = await GetSystemAccentColor();

      // 1. Auto-start: default to enabled on first run
      const autostartInit = await GetSetting('easywinget_autostart_init', '');
      if (!autostartInit) {
        try {
          await SetAutoStart(true);
          await SetSetting('easywinget_autostart_init', 'done');
          autoStartEnabled = true;
        } catch (e) {
          autoStartEnabled = await GetAutoStart();
        }
      } else {
        autoStartEnabled = await GetAutoStart();
      }

      // 2. Minimize to tray: default to enabled
      const trayVal = await GetSetting('easywinget_minimize_to_tray', 'true');
      minimizeToTrayEnabled = trayVal !== 'false';

      // 3. Automated background update check & notification interval
      const autoCheckVal = await GetSetting('easywinget_auto_update_check', 'true');
      autoUpdateCheckEnabled = autoCheckVal !== 'false';
      const intervalVal = await GetSetting('easywinget_update_check_interval', '1440');
      updateCheckInterval = intervalVal || '1440';
    } catch (err) {
      console.error('Error loading system settings:', err);
    }
  });


  function selectTheme(pref) {
    themePreference = pref;
    onThemeChange(pref);
    const label = pref === 'auto' ? 'System Auto' : (pref === 'dark' ? 'Dark Mode' : 'Light Mode');
    onShowToast('info', 'Theme Changed', `Switched to ${label}`);
  }

  async function selectAccentSource(source) {
    accentSource = source;
    if (source === 'windows') {
      try {
        windowsAccent = await GetSystemAccentColor();
        onAccentChange(windowsAccent, 'windows');
        onShowToast('info', 'Windows Accent Applied', `Using system accent color`);
      } catch (e) {
        onAccentChange('#0078d4', 'default');
      }
    } else {
      onAccentChange('#0078d4', 'default');
      onShowToast('info', 'Default Accent Applied', 'Switched to Fluent Blue');
    }
  }

  async function handleAutoStartToggle(e) {
    const checked = e.target.checked;
    autoStartEnabled = checked;
    try {
      await SetAutoStart(checked);
      if (checked) {
        onShowToast('success', 'Auto-Start Enabled', 'EasyWinGet will launch silently in background on startup');
      } else {
        onShowToast('info', 'Auto-Start Disabled', 'Removed from Windows startup');
      }
    } catch (err) {
      autoStartEnabled = !checked;
      onShowToast('error', 'Auto-Start Error', String(err));
    }
  }

  async function handleMinimizeToTrayToggle(e) {
    const checked = e.target.checked;
    minimizeToTrayEnabled = checked;
    try {
      await SetSetting('easywinget_minimize_to_tray', checked ? 'true' : 'false');
      if (typeof window !== 'undefined') {
        window.dispatchEvent(new CustomEvent('easywinget:minimize-tray-changed', { detail: checked }));
      }
      if (checked) {
        onShowToast('success', 'Tray Minimization Enabled', 'Closing window keeps EasyWinGet in Action Center / System Tray');
      } else {
        onShowToast('info', 'Tray Minimization Disabled', 'Closing window will terminate the application');
      }
    } catch (err) {
      minimizeToTrayEnabled = !checked;
      onShowToast('error', 'Settings Error', String(err));
    }
  }

  async function handleAutoUpdateToggle(e) {
    const checked = e.target.checked;
    autoUpdateCheckEnabled = checked;
    try {
      await SetSetting('easywinget_auto_update_check', checked ? 'true' : 'false');
      await ApplyAutoUpdateSettings();
      if (checked) {
        onShowToast('success', 'Update Checks Enabled', 'EasyWinGet will periodically scan for software updates in the background');
      } else {
        onShowToast('info', 'Update Checks Disabled', 'Automatic background update checking paused');
      }
    } catch (err) {
      autoUpdateCheckEnabled = !checked;
      onShowToast('error', 'Settings Error', String(err));
    }
  }

  async function handleIntervalChange(e) {
    const val = e.target.value;
    updateCheckInterval = val;
    try {
      await SetSetting('easywinget_update_check_interval', val);
      await ApplyAutoUpdateSettings();
      const labels = {
        '3': '3 Minutes (Testing)',
        '60': '1 Hour',
        '180': '3 Hours',
        '360': '6 Hours',
        '720': '12 Hours',
        '1440': '24 Hours (Daily)'
      };
      onShowToast('success', 'Check Interval Updated', `Updates will be checked every ${labels[val] || val + ' minutes'}`);
    } catch (err) {
      onShowToast('error', 'Settings Error', String(err));
    }
  }


  async function handleClearCache() {
    const ok = await onConfirmRequest({
      title: 'Clear SQLite Cache?',
      message: 'This will reset cached package metadata and icons in the local SQLite database. EasyWinGet will re-sync fresh data as needed.',
      confirmText: 'Clear Cache',
      cancelText: 'Cancel',
      isDanger: true,
      iconType: 'warning'
    });
    if (!ok) return;

    isClearing = true;
    try {
      await ClearCache();
      onShowToast('success', 'Cache Cleared', 'Application SQLite cache has been cleared cleanly');
    } catch (err) {
      onShowToast('error', 'Error Clearing Cache', String(err));
    } finally {
      isClearing = false;
    }
  }

  async function handleImport() {
    try {
      const path = await SelectImportPath();
      if (!path) return;
      const jobId = await ImportApps(path);
      onStartJob(jobId, `Importing Packages from ${path}`);
      onShowToast('info', 'Import Started', 'WinGet is importing packages from file...');
    } catch (err) {
      onShowToast('error', 'Import Failed', String(err));
    }
  }

  async function handleExport() {
    try {
      const path = await SelectExportPath();
      if (!path) return;
      await ExportApps(path);
      onShowToast('success', 'Export Complete', `Saved installed packages to ${path}`);
    } catch (err) {
      onShowToast('error', 'Export Failed', String(err));
    }
  }
</script>

<div class="settings-page view-animated">
  <div class="settings-inner">
    <div class="page-header">
      <h2>Settings</h2>
      <p>Personalize your experience, manage system data, and configure package manager options</p>
    </div>

    <div class="settings-body">
      <!-- 1. Appearance / Theme Section -->
      <div class="settings-card">
        <div class="card-title-row">
          <div class="card-icon">
            <Palette size={18} color="currentColor" />
          </div>
          <div class="card-heading">
            <h3>Appearance & Accent Color</h3>
            <p>Choose theme mode and sync accent color with Windows system settings</p>
          </div>
        </div>

        <div class="theme-selector-grid">
          <!-- Dark Mode -->
          <button 
            class="theme-card" 
            class:selected={themePreference === 'dark'}
            on:click={() => selectTheme('dark')}
          >
            <div class="theme-preview dark-preview">
              <div class="preview-sidebar"></div>
              <div class="preview-content">
                <div class="preview-pill" style="background: {accentColor}"></div>
                <div class="preview-line"></div>
              </div>
            </div>
            <div class="theme-label">
              <Moon size={15} />
              <span>Dark Mode</span>
            </div>
            {#if themePreference === 'dark'}
              <div class="check-badge"><Check size={12} /></div>
            {/if}
          </button>

          <!-- Light Mode -->
          <button 
            class="theme-card" 
            class:selected={themePreference === 'light'}
            on:click={() => selectTheme('light')}
          >
            <div class="theme-preview light-preview">
              <div class="preview-sidebar"></div>
              <div class="preview-content">
                <div class="preview-pill" style="background: {accentColor}"></div>
                <div class="preview-line"></div>
              </div>
            </div>
            <div class="theme-label">
              <Sun size={15} />
              <span>Light Mode</span>
            </div>
            {#if themePreference === 'light'}
              <div class="check-badge"><Check size={12} /></div>
            {/if}
          </button>

          <!-- System Auto -->
          <button 
            class="theme-card" 
            class:selected={themePreference === 'auto'}
            on:click={() => selectTheme('auto')}
          >
            <div class="theme-preview auto-preview">
              <div class="preview-half-dark">
                <div class="mini-pill" style="background: {accentColor}"></div>
              </div>
              <div class="preview-half-light">
                <div class="mini-pill" style="background: {accentColor}"></div>
              </div>
            </div>
            <div class="theme-label">
              <Laptop size={15} />
              <span>System Auto</span>
            </div>
            {#if themePreference === 'auto'}
              <div class="check-badge"><Check size={12} /></div>
            {/if}
          </button>
        </div>

        <!-- Accent Color Setting -->
        <div class="setting-item-row">
          <div class="item-text">
            <div class="item-title">Accent Color Source</div>
            <div class="item-desc">Choose between EasyWinGet Default Theme or Windows 11 System Accent</div>
          </div>

          <div class="accent-toggle-group">
            <button 
              class="accent-pill {accentSource === 'default' ? 'active' : ''}"
              on:click={() => selectAccentSource('default')}
            >
              <span class="color-dot default-dot"></span>
              <span>Default Theme</span>
            </button>
            <button 
              class="accent-pill {accentSource === 'windows' ? 'active' : ''}"
              on:click={() => selectAccentSource('windows')}
            >
              <span class="color-dot" style="background-color: {windowsAccent}"></span>
              <span>Windows Accent</span>
            </button>
          </div>
        </div>
      </div>

      <!-- 2. Windows Startup & System Integration -->
      <div class="settings-card">
        <div class="card-title-row">
          <div class="card-icon">
            <Power size={18} color="currentColor" />
          </div>
          <div class="card-heading">
            <h3>Windows Startup & Background Running</h3>
            <p>Configure background startup and action center minimize behavior</p>
          </div>
        </div>

        <div class="setting-item-row">
          <div class="item-text">
            <div class="item-title">Start Silently on Windows Boot</div>
            <div class="item-desc">Starts EasyWinGet in the background (minimized to system tray) when Windows logs in</div>
          </div>
          <label class="toggle-switch">
            <input 
              type="checkbox" 
              checked={autoStartEnabled} 
              on:change={handleAutoStartToggle}
            />
            <span class="toggle-slider"></span>
          </label>
        </div>

        <div class="setting-item-row">
          <div class="item-text">
            <div class="item-title">Minimize to System Tray on Close</div>
            <div class="item-desc">Closing window keeps EasyWinGet running in Action Center / System Tray</div>
          </div>
          <label class="toggle-switch">
            <input 
              type="checkbox" 
              checked={minimizeToTrayEnabled} 
              on:change={handleMinimizeToTrayToggle}
            />
            <span class="toggle-slider"></span>
          </label>
        </div>
      </div>

      <!-- Automated Update Checks & Interactive Notifications -->
      <div class="settings-card">
        <div class="card-title-row">
          <div class="card-icon">
            <Bell size={18} color="currentColor" />
          </div>
          <div class="card-heading">
            <h3>Automated Updates & Notifications</h3>
            <p>Periodically scan for outdated applications and receive interactive Windows notifications</p>
          </div>
        </div>

        <div class="setting-item-row">
          <div class="item-text">
            <div class="item-title">Periodic Background Update Checks</div>
            <div class="item-desc">When an internet connection is available, scans WinGet for app upgrades and sends an interactive notification</div>
          </div>
          <label class="toggle-switch">
            <input 
              type="checkbox" 
              checked={autoUpdateCheckEnabled} 
              on:change={handleAutoUpdateToggle}
            />
            <span class="toggle-slider"></span>
          </label>
        </div>

        {#if autoUpdateCheckEnabled}
          <div class="setting-item-row">
            <div class="item-text">
              <div class="item-title">Update Check Frequency</div>
              <div class="item-desc">Frequency of automated update scans in the background</div>
            </div>
            <div class="select-wrapper">
              <select 
                class="fluent-select" 
                value={updateCheckInterval} 
                on:change={handleIntervalChange}
              >
                <option value="3">3 Minutes (Fast Testing)</option>
                <option value="60">Every 1 Hour</option>
                <option value="180">Every 3 Hours</option>
                <option value="360">Every 6 Hours</option>
                <option value="720">Every 12 Hours</option>
                <option value="1440">Every 24 Hours (Once Daily - Default)</option>
              </select>
            </div>
          </div>
        {/if}
      </div>

      <!-- 3. SQLite Data & Cache Management -->
      <div class="settings-card">
        <div class="card-title-row">
          <div class="card-icon">
            <HardDrive size={18} color="currentColor" />
          </div>
          <div class="card-heading">
            <h3>High-Performance SQLite Storage</h3>
            <p>Package listings, updates, and settings are cached in pure Go SQLite for sub-millisecond loading</p>
          </div>
        </div>

        <div class="setting-item-row">
          <div class="item-text">
            <div class="item-title">Wipe SQLite Local Cache</div>
            <div class="item-desc">Clears cached tables and triggers a fresh background scan from WinGet repository</div>
          </div>
          <button class="btn-danger" on:click={handleClearCache} disabled={isClearing}>
            <Trash2 size={14} /> Clear Cache
          </button>
        </div>
      </div>

      <!-- 4. Backup & Restore Section -->
      <div class="settings-card">
        <div class="card-title-row">
          <div class="card-icon">
            <UploadCloud size={18} color="currentColor" />
          </div>
          <div class="card-heading">
            <h3>Backup & Migration</h3>
            <p>Export your installed applications or batch restore packages from a JSON manifest</p>
          </div>
        </div>

        <div class="setting-item-row">
          <div class="item-text">
            <div class="item-title">Export Installed Applications</div>
            <div class="item-desc">Creates a portable JSON file containing all installed packages</div>
          </div>
          <button class="btn-secondary" on:click={handleExport}>
            <FileSpreadsheet size={14} /> Export JSON
          </button>
        </div>

        <div class="setting-item-row">
          <div class="item-text">
            <div class="item-title">Import & Install Packages</div>
            <div class="item-desc">Batch-installs packages from an exported WinGet JSON configuration file</div>
          </div>
          <button class="btn-primary" on:click={handleImport}>
            <UploadCloud size={14} /> Import File
          </button>
        </div>
      </div>

      <!-- 5. System & About -->
      <div class="settings-card">
        <div class="card-title-row">
          <div class="card-icon">
            <Info size={18} color="currentColor" />
          </div>
          <div class="card-heading">
            <h3>About & Legal Center</h3>
            <p>Application diagnostics, Indian Law Terms of Service, Privacy Policy & User Guide</p>
          </div>
        </div>

        <div class="setting-item-row">
          <div class="item-text">
            <div class="item-title">Explore Full About & Legal Hub</div>
            <div class="item-desc">Access technical architecture, Indian IT Act 2000 legal terms, DPDP Privacy Policy, and help guide</div>
          </div>
          <button class="btn-primary" on:click={() => onNavigateTab('about')}>
            <Info size={14} /> Open About Hub
          </button>
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  .settings-page {
    display: flex;
    flex-direction: column;
    align-items: center;
    width: 100%;
    height: 100%;
    overflow-y: auto;
    padding: 32px 24px 64px;
  }

  .settings-inner {
    width: 100%;
    max-width: 720px;
    display: flex;
    flex-direction: column;
  }

  .page-header {
    margin-bottom: 24px;
  }

  .page-header h2 {
    font-size: 1.6rem;
    font-weight: 600;
    letter-spacing: -0.01em;
    color: var(--text-primary);
  }

  .page-header p {
    color: var(--text-secondary);
    font-size: 0.9rem;
    margin-top: 4px;
  }

  .settings-body {
    display: flex;
    flex-direction: column;
    gap: 16px;
    width: 100%;
  }

  .settings-card {
    background: var(--bg-card);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-md);
    padding: 20px 22px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    transition: border-color var(--transition-fast);
  }

  .card-title-row {
    display: flex;
    align-items: flex-start;
    gap: 14px;
  }

  /* Optimized Settings Card Icons - Crisp White in Dark Mode, Crisp Dark in Light Mode with Accent Border */
  .card-icon {
    width: 36px;
    height: 36px;
    border-radius: var(--radius-sm);
    background: var(--accent-badge-bg, rgba(255, 255, 255, 0.04));
    border: 1px solid var(--accent-badge-border, rgba(255, 255, 255, 0.12));
    color: #ffffff; /* Crisp white icon in Dark mode */
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
    transition: all var(--transition-fast);
  }

  :global([data-theme="light"]) .card-icon {
    background: var(--accent-badge-bg, rgba(0, 0, 0, 0.03));
    border: 1px solid var(--accent-badge-border, rgba(0, 0, 0, 0.12));
    color: #191919; /* Crisp dark icon in Light mode */
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
  }

  .card-heading h3 {
    font-size: 0.98rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .card-heading p {
    font-size: 0.84rem;
    color: var(--text-secondary);
    margin-top: 2px;
  }

  /* 3-Column Theme Selector Grid */
  .theme-selector-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
  }

  @media (max-width: 620px) {
    .theme-selector-grid {
      grid-template-columns: 1fr;
    }
  }

  .theme-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
    padding: 12px;
    background: var(--bg-surface);
    border: 2px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    position: relative;
    cursor: pointer;
    transition: border-color var(--transition-fast), background-color var(--transition-fast);
  }

  .theme-card:hover {
    border-color: var(--border-card-hover);
  }

  .theme-card.selected {
    border-color: var(--accent-light);
    background: var(--accent-badge-bg, rgba(96, 205, 255, 0.08));
  }

  .theme-preview {
    width: 100%;
    height: 64px;
    border-radius: 6px;
    display: flex;
    overflow: hidden;
    border: 1px solid rgba(128, 128, 128, 0.2);
  }

  .dark-preview {
    background: #18191b;
  }
  .dark-preview .preview-sidebar {
    width: 32%;
    background: #202124;
    border-right: 1px solid rgba(255, 255, 255, 0.08);
  }
  .dark-preview .preview-content {
    flex: 1;
    padding: 8px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .dark-preview .preview-pill {
    width: 60%;
    height: 7px;
    border-radius: 4px;
  }
  .dark-preview .preview-line {
    width: 85%;
    height: 5px;
    border-radius: 3px;
    background: rgba(255, 255, 255, 0.15);
  }

  .light-preview {
    background: #f3f3f3;
  }
  .light-preview .preview-sidebar {
    width: 32%;
    background: #e9e9e9;
    border-right: 1px solid rgba(0, 0, 0, 0.08);
  }
  .light-preview .preview-content {
    flex: 1;
    padding: 8px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .light-preview .preview-pill {
    width: 60%;
    height: 7px;
    border-radius: 4px;
  }
  .light-preview .preview-line {
    width: 85%;
    height: 5px;
    border-radius: 3px;
    background: rgba(0, 0, 0, 0.12);
  }

  /* Auto Preview Split */
  .auto-preview {
    display: flex;
  }
  .preview-half-dark {
    flex: 1;
    background: #18191b;
    border-right: 1px solid rgba(255, 255, 255, 0.1);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .preview-half-light {
    flex: 1;
    background: #f3f3f3;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .mini-pill {
    width: 70%;
    height: 6px;
    border-radius: 3px;
  }

  .theme-label {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 0.85rem;
    font-weight: 500;
    color: var(--text-primary);
  }

  .check-badge {
    position: absolute;
    top: 8px;
    right: 8px;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: var(--accent-light);
    color: #000000;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  :global([data-theme="light"]) .check-badge {
    color: #ffffff;
    background: var(--accent-primary);
  }

  /* Setting Item Row */
  .setting-item-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding-top: 12px;
    border-top: 1px solid var(--border-subtle);
  }

  .item-text {
    flex: 1;
  }

  .item-title {
    font-size: 0.88rem;
    font-weight: 500;
    color: var(--text-primary);
  }

  .item-desc {
    font-size: 0.78rem;
    color: var(--text-muted);
    margin-top: 2px;
  }

  /* Accent Toggle Group */
  .accent-toggle-group {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .accent-pill {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 6px 14px;
    background: var(--bg-card);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-sm);
    color: var(--text-secondary);
    font-size: 0.82rem;
    font-weight: 500;
    transition: all var(--transition-fast);
  }

  .accent-pill:hover {
    background: var(--bg-card-hover);
    color: var(--text-primary);
    border-color: var(--border-card-hover);
  }

  .accent-pill.active {
    background: var(--accent-badge-bg, rgba(96, 205, 255, 0.12));
    border-color: var(--accent-light);
    color: var(--text-primary);
    font-weight: 600;
  }

  .color-dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    border: 1px solid rgba(255, 255, 255, 0.2);
  }

  .default-dot {
    background-color: #0078d4;
  }

  /* Fluent Select Dropdown */
  .select-wrapper {
    position: relative;
    display: inline-flex;
    align-items: center;
  }

  .fluent-select {
    appearance: none;
    -webkit-appearance: none;
    background: var(--bg-card);
    color: var(--text-primary);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-sm);
    padding: 7px 34px 7px 12px;
    font-size: 0.84rem;
    font-family: inherit;
    font-weight: 500;
    cursor: pointer;
    outline: none;
    transition: all var(--transition-fast);
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='14' height='14' viewBox='0 0 24 24' fill='none' stroke='%23888888' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 10px center;
  }

  .fluent-select:hover {
    border-color: var(--accent-light, #60a5fa);
    background-color: var(--bg-card-hover);
  }

  .fluent-select:focus {
    border-color: var(--accent-primary, #0078d4);
    box-shadow: 0 0 0 2px var(--accent-badge-bg, rgba(0, 120, 212, 0.25));
  }

  .fluent-select option {
    background: var(--bg-card);
    color: var(--text-primary);
  }
</style>

