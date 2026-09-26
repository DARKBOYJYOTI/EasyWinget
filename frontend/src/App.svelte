<script>
  import { onMount } from 'svelte';
  import Sidebar from './components/Sidebar.svelte';
  import SearchView from './components/SearchView.svelte';
  import InstalledView from './components/InstalledView.svelte';
  import UpdatesView from './components/UpdatesView.svelte';
  import DownloadsView from './components/DownloadsView.svelte';
  import TaskTerminal from './components/TaskTerminal.svelte';
  import DetailsModal from './components/DetailsModal.svelte';
  import SettingsModal from './components/SettingsModal.svelte';
  import AboutView from './components/AboutView.svelte';
  import ConfirmModal from './components/ConfirmModal.svelte';
  import Toast from './components/Toast.svelte';
  import TitleBar from './components/TitleBar.svelte';


  import { 
    GetAppVersion, 
    GetSystemAccentColor, 
    GetInstalled, 
    GetUpdates, 
    GetDownloadedFiles,
    GetSetting,
    SetSetting,
    GetWindowsTransparencyEnabled
  } from '../wailsjs/go/main/App.js';
  import { EventsOn } from '../wailsjs/runtime/runtime.js';
  import { Terminal, X } from '@lucide/svelte';
  import { applyAccentToDocument } from './utils/theme.js';

  let activeTab = 'search';
  let installedCount = 0;
  let updatesCount = 0;
  let downloadsCount = 0;
  let appVersion = '4.0.0';
  let themePreference = 'dark'; // 'dark' | 'light' | 'auto'
  let currentTheme = 'dark';    // rendered theme 'dark' | 'light'
  let accentColor = '#0078d4';
  let accentSource = 'default'; // 'default' or 'windows'

  let installedIds = new Set();

  let activeJob = null; // { id, title }
  let isJobRunning = false;
  let selectedPackageId = null;
  let toasts = [];

  // Global Fluent Confirm Modal State
  let confirmState = {
    show: false,
    title: 'Confirm Action',
    message: 'Are you sure you want to proceed?',
    confirmText: 'Confirm',
    cancelText: 'Cancel',
    isDanger: false,
    iconType: 'warning',
    onConfirm: () => {},
    onCancel: () => {}
  };

  function requestConfirm(options = {}) {
    return new Promise((resolve) => {
      confirmState = {
        show: true,
        title: options.title || 'Confirm Action',
        message: options.message || 'Are you sure you want to proceed?',
        confirmText: options.confirmText || 'Confirm',
        cancelText: options.cancelText || 'Cancel',
        isDanger: options.isDanger || false,
        iconType: options.iconType || (options.isDanger ? 'danger' : 'warning'),
        onConfirm: () => {
          confirmState.show = false;
          resolve(true);
        },
        onCancel: () => {
          confirmState.show = false;
          resolve(false);
        }
      };
    });
  }

  function getSystemTheme() {
    if (typeof window !== 'undefined' && window.matchMedia) {
      return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }
    return 'dark';
  }

  function updateInstalledSet(apps) {
    if (!apps || !Array.isArray(apps)) return;
    installedCount = apps.length;
    const s = new Set();
    for (const app of apps) {
      if (app.id) s.add(app.id.toLowerCase());
      if (app.name) s.add(app.name.toLowerCase());
    }
    installedIds = s;
  }

  async function syncWindowsTransparency() {
    try {
      const enabled = await GetWindowsTransparencyEnabled();
      if (typeof document !== 'undefined') {
        document.documentElement.dataset.transparency = enabled ? 'true' : 'false';
      }
    } catch (e) {}
  }

  onMount(async () => {
    // 0. Sync Windows Transparency Effect (Mica & Acrylic Material integration)
    syncWindowsTransparency();
    if (typeof window !== 'undefined') {
      window.addEventListener('focus', syncWindowsTransparency);
    }
    const transInterval = setInterval(syncWindowsTransparency, 5000);

    // 1. Listen for OS theme changes when in 'auto' mode
    if (typeof window !== 'undefined' && window.matchMedia) {
      const mq = window.matchMedia('(prefers-color-scheme: dark)');
      mq.addEventListener('change', () => {
        if (themePreference === 'auto') {
          applyTheme('auto');
        }
      });
    }

    // 2. Load saved theme from SQLite database (permanent persistence), fallback to localStorage
    try {
      let savedTheme = '';
      try {
        savedTheme = await GetSetting('easywinget_theme', '');
      } catch (e) {}
      if (!savedTheme && typeof localStorage !== 'undefined') {
        savedTheme = localStorage.getItem('easywinget_theme') || '';
      }
      applyTheme(savedTheme || 'dark');
    } catch (e) {
      applyTheme('dark');
    }

    // 3. Load saved accent source & color from SQLite database
    try {
      let savedSource = '';
      let savedColor = '';
      try {
        savedSource = await GetSetting('easywinget_accent_source', '');
        savedColor = await GetSetting('easywinget_accent_color', '');
      } catch (e) {}

      if (!savedSource && typeof localStorage !== 'undefined') {
        savedSource = localStorage.getItem('easywinget_accent_source') || '';
        savedColor = localStorage.getItem('easywinget_accent_color') || '';
      }

      savedSource = savedSource || 'default';
      savedColor = savedColor || '#0078d4';

      if (savedSource === 'windows') {
        const sysColor = await GetSystemAccentColor();
        applyAccent(sysColor || savedColor, 'windows');
      } else {
        applyAccent(savedColor, 'default');
      }
    } catch (e) {
      applyAccent('#0078d4', 'default');
    }

    // 4. Get Version
    try {
      appVersion = await GetAppVersion();
    } catch (e) {}

    // 5. Instant initial counts & installed mapping from SQLite (< 2ms)
    try {
      const [installed, updates, files] = await Promise.all([
        GetInstalled(false),
        GetUpdates(false),
        GetDownloadedFiles()
      ]);
      if (installed) {
        updateInstalledSet(installed);
      }
      if (updates) updatesCount = updates.length;
      if (files) downloadsCount = files.length;
    } catch (e) {
      console.warn('Initial counts fetch:', e);
    }

    // 6. Global Wails Events
    try {
      // Backend toast notification (e.g. from background update check)
      EventsOn('toast:notify', (data) => {
        if (data) {
          showToast(data.type || 'info', data.title, data.message);
        }
      });

      // Background installed update
      EventsOn('installed:updated', (apps) => {
        if (apps && Array.isArray(apps)) {
          updateInstalledSet(apps);
        }
      });

      // Background updates update
      EventsOn('updates:updated', (ups) => {
        if (ups && Array.isArray(ups)) {
          updatesCount = ups.length;
        }
      });

      // Job completion tracker
      EventsOn('job:done', (data) => {
        isJobRunning = false;
      });
    } catch (e) {
      console.warn('EventsOn error:', e);
    }
  });

  function applyTheme(preference) {
    themePreference = preference;
    const effective = preference === 'auto' ? getSystemTheme() : preference;
    currentTheme = effective;
    if (typeof document !== 'undefined') {
      document.documentElement.setAttribute('data-theme', effective);
    }
    try {
      localStorage.setItem('easywinget_theme', preference);
    } catch (e) {}
    try {
      SetSetting('easywinget_theme', preference);
    } catch (e) {}

    // Recalculate accent auto-luminance for effective theme
    applyAccent(accentColor, accentSource);
  }

  function applyAccent(color, source = 'default') {
    accentColor = color;
    accentSource = source;
    const isDark = currentTheme === 'dark';
    applyAccentToDocument(color, isDark);
    try {
      localStorage.setItem('easywinget_accent_color', color);
      localStorage.setItem('easywinget_accent_source', source);
    } catch (e) {}
    try {
      SetSetting('easywinget_accent_color', color);
      SetSetting('easywinget_accent_source', source);
    } catch (e) {}
  }

  let toastTimer = null;

  function showToast(type, title, message) {
    if (toastTimer) {
      clearTimeout(toastTimer);
      toastTimer = null;
    }
    const id = Date.now() + Math.random();
    // Only one toast at a time: override any currently active toast
    toasts = [{ id, type, title, message }];

    toastTimer = setTimeout(() => {
      toasts = [];
      toastTimer = null;
    }, 3200);
  }

  function removeToast(id) {
    if (toastTimer) {
      clearTimeout(toastTimer);
      toastTimer = null;
    }
    toasts = toasts.filter(t => t.id !== id);
  }

  function handleStartJob(jobId, title) {
    activeJob = { id: jobId, title };
    isJobRunning = true;
    activeTab = 'terminal';
  }

  function openDetails(id) {
    selectedPackageId = id;
  }

  function closeDetails() {
    selectedPackageId = null;
  }
</script>

<div class="app-layout">
  <TitleBar />
  <Sidebar 
    bind:activeTab 
    {installedCount} 
    {updatesCount} 
    {downloadsCount} 
    {isJobRunning}
    version={appVersion}
  />

  <main class="main-content">
    {#if activeTab === 'search'}
      <div class="view-animated">
        <SearchView 
          {installedIds}
          onOpenDetails={openDetails}
          onStartJob={handleStartJob}
          onShowToast={showToast}
          onConfirmRequest={requestConfirm}
        />
      </div>
    {:else if activeTab === 'installed'}
      <div class="view-animated">
        <InstalledView 
          onStartJob={handleStartJob}
          onShowToast={showToast}
          onUpdateCount={(c) => installedCount = c}
          onConfirmRequest={requestConfirm}
        />
      </div>
    {:else if activeTab === 'updates'}
      <div class="view-animated">
        <UpdatesView 
          onStartJob={handleStartJob}
          onShowToast={showToast}
          onUpdateCount={(c) => updatesCount = c}
        />
      </div>
    {:else if activeTab === 'downloads'}
      <div class="view-animated">
        <DownloadsView 
          onStartJob={handleStartJob}
          onShowToast={showToast}
          onUpdateCount={(c) => downloadsCount = c}
        />
      </div>
    {:else if activeTab === 'terminal'}
      <div class="terminal-tab-wrapper view-animated">
        <TaskTerminal 
          {activeJob}
          onClose={() => activeTab = 'search'}
          onShowToast={showToast}
          onJobDone={() => isJobRunning = false}
        />
      </div>
    {:else if activeTab === 'settings'}
      <div class="view-animated">
        <SettingsModal 
          version={appVersion}
          {themePreference}
          {currentTheme}
          {accentColor}
          {accentSource}
          onThemeChange={applyTheme}
          onAccentChange={applyAccent}
          onStartJob={handleStartJob}
          onShowToast={showToast}
          onConfirmRequest={requestConfirm}
          onNavigateTab={(tab) => activeTab = tab}
        />
      </div>
    {:else if activeTab === 'about'}
      <div class="view-animated">
        <AboutView 
          version={appVersion}
          onShowToast={showToast}
        />
      </div>
    {/if}


    <!-- Bottom Docked Task Bar when browsing another tab while a job is running -->
    {#if isJobRunning && activeTab !== 'terminal'}
      <div class="docked-taskbar" on:click={() => activeTab = 'terminal'}>
        <div class="docked-pulse"></div>
        <Terminal size={15} color="var(--accent-light)" />
        <span class="docked-title">{activeJob?.title || 'Background Task'}</span>
        <button class="btn-primary docked-btn">View Console</button>
        <button 
          class="docked-close" 
          title="Dismiss status bar" 
          on:click={(e) => { e.stopPropagation(); isJobRunning = false; }}
        >
          <X size={14} />
        </button>
      </div>
    {/if}
  </main>

  <!-- Package Details Modal -->
  {#if selectedPackageId}
    <DetailsModal 
      packageId={selectedPackageId}
      {installedIds}
      onClose={closeDetails}
      onStartJob={handleStartJob}
      onShowToast={showToast}
      onConfirmRequest={requestConfirm}
    />
  {/if}

  <!-- Fluent Confirmation & Alert Modal -->
  <ConfirmModal 
    show={confirmState.show}
    title={confirmState.title}
    message={confirmState.message}
    confirmText={confirmState.confirmText}
    cancelText={confirmState.cancelText}
    isDanger={confirmState.isDanger}
    iconType={confirmState.iconType}
    onConfirm={confirmState.onConfirm}
    onCancel={confirmState.onCancel}
  />

  <!-- Toast Container with Sound Synthesizer -->
  <Toast {toasts} {removeToast} />
</div>

<style>
  .app-layout {
    display: flex;
    width: 100vw;
    height: 100vh;
    background: var(--bg-app);
    overflow: hidden;
  }

  .main-content {
    flex: 1;
    height: 100vh;
    overflow: hidden;
    position: relative;
    background: var(--bg-app);
    padding-top: 38px; /* Clean clearance below custom Windows 11 title bar */
    box-sizing: border-box;
  }

  .terminal-tab-wrapper {
    height: 100%;
    padding: 20px 24px;
    box-sizing: border-box;
  }

  .docked-taskbar {
    position: absolute;
    bottom: 16px;
    left: 28px;
    right: 28px;
    background: var(--bg-surface);
    border: 1px solid var(--border-card-hover);
    border-radius: var(--radius-sm);
    padding: 8px 16px;
    display: flex;
    align-items: center;
    gap: 12px;
    cursor: pointer;
    box-shadow: var(--shadow-flyout);
    z-index: 100;
    transition: background-color var(--transition-fast), border-color var(--transition-fast);
  }

  .docked-taskbar:hover {
    background: var(--bg-card-hover);
    border-color: var(--accent-light);
  }

  .docked-pulse {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--color-success-light);
    box-shadow: 0 0 6px var(--color-success-light);
    animation: pulseGlow 1s infinite alternate;
  }

  .docked-title {
    font-size: 0.85rem;
    font-weight: 500;
    color: var(--text-primary);
    flex: 1;
  }

  .docked-btn {
    padding: 4px 10px;
    font-size: 0.78rem;
  }

  .docked-close {
    background: transparent;
    color: var(--text-muted);
    border: none;
    cursor: pointer;
    padding: 4px;
    border-radius: var(--radius-xs);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    transition: all var(--transition-fast);
  }
  .docked-close:hover {
    color: var(--text-primary);
    background: rgba(255, 255, 255, 0.1);
  }
</style>
