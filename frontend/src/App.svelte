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
  import AppUpdateModal from './components/AppUpdateModal.svelte';
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
    GetWindowsTransparencyEnabled,
    GetInitialTab,
    RefreshAll,
    OpenURLInDefaultBrowser,
    InstallPackage,
    DownloadPackage,
    UninstallPackage,
    CheckAppSelfUpdate
  } from '../wailsjs/go/main/App.js';
  import { EventsOn } from '../wailsjs/runtime/runtime.js';
  import { Terminal, X, Clock } from '@lucide/svelte';
  import { applyAccentToDocument } from './utils/theme.js';

  let activeTab = 'search';
  let previousTab = 'search';
  let lastNonTerminalTab = 'search';

  $: if (activeTab !== 'terminal') {
    lastNonTerminalTab = activeTab;
  }

  let installedCount = 0;
  let updatesCount = 0;
  let downloadsCount = 0;
  let searchCount = null;
  function getSystemTheme() {
    if (typeof window !== 'undefined' && window.matchMedia) {
      return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }
    return 'dark';
  }

  function getInitialTheme() {
    let pref = 'auto';
    if (typeof localStorage !== 'undefined') {
      pref = localStorage.getItem('easywinget_theme') || 'auto';
    }
    const current = pref === 'auto' ? getSystemTheme() : pref;
    return { pref, current };
  }

  const initialTheme = getInitialTheme();
  let themePreference = initialTheme.pref; // 'auto' | 'dark' | 'light'
  let currentTheme = initialTheme.current; // 'light' if OS is light, 'dark' if dark
  let accentColor = '#0078d4';
  let accentSource = 'default'; // 'default' or 'windows'
  if (typeof localStorage !== 'undefined') {
    accentColor = localStorage.getItem('easywinget_accent_color') || '#0078d4';
    accentSource = localStorage.getItem('easywinget_accent_source') || 'default';
  }

  let installedIds = new Set();
  let appVersion = '4.0.0';

  let activeJob = null; // { id, title, packageId, action }
  let isJobRunning = false;
  let runningJobs = {}; // { [packageId: string]: 'install' | 'upgrade' | 'uninstall' | 'download' | 'queued' }
  let jobQueue = []; // [{ id, name, action, interactive }]
  let selectedPackageId = null;
  let toasts = [];

  // EasyWinGet App Self-Update State
  let showAppUpdateModal = false;
  let appUpdateInfo = null;

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

    // 2. Load saved theme from localStorage first (0ms), then verify with SQLite
    try {
      let savedTheme = '';
      if (typeof localStorage !== 'undefined') {
        savedTheme = localStorage.getItem('easywinget_theme') || '';
      }
      if (!savedTheme) {
        try {
          savedTheme = await GetSetting('easywinget_theme', '');
        } catch (e) {}
      }
      applyTheme(savedTheme || 'auto');
    } catch (e) {
      applyTheme('auto');
    }

    // 3. Load saved accent source & color from SQLite database
    try {
      let savedSource = '';
      let savedColor = '';
      if (typeof localStorage !== 'undefined') {
        savedSource = localStorage.getItem('easywinget_accent_source') || '';
        savedColor = localStorage.getItem('easywinget_accent_color') || '';
      }
      if (!savedSource) {
        try {
          savedSource = await GetSetting('easywinget_accent_source', '');
          savedColor = await GetSetting('easywinget_accent_color', '');
        } catch (e) {}
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
      const v = await GetAppVersion();
      if (v) appVersion = v;
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

    // Check if app was launched with a specific tab request (e.g. protocol activation easywinget://updates)
    try {
      const initTab = await GetInitialTab();
      if (initTab) {
        activeTab = initTab;
      }
    } catch (e) {}

    // Global Window Hidden & Zero CPU Suspension Sync
    const updateWindowHiddenState = () => {
      const isHidden = Boolean(document.hidden || document.visibilityState === 'hidden');
      if (typeof document !== 'undefined') {
        document.documentElement.dataset.windowHidden = isHidden ? 'true' : 'false';
      }
    };
    updateWindowHiddenState();
    if (typeof document !== 'undefined') {
      document.addEventListener('visibilitychange', updateWindowHiddenState);
      window.addEventListener('focus', updateWindowHiddenState);
      window.addEventListener('blur', () => { setTimeout(updateWindowHiddenState, 80); });
      window.addEventListener('easywinget:window-hidden', updateWindowHiddenState);
      window.addEventListener('easywinget:window-shown', updateWindowHiddenState);
    }

    // Global link interceptor: open any external http/https link in default Windows browser (via explorer.exe / shell)
    const handleGlobalLinkClick = (e) => {
      const link = e.target && e.target.closest ? e.target.closest('a') : null;
      if (link && link.href) {
        const href = link.href;
        if (href.startsWith('http://') || href.startsWith('https://') || href.startsWith('mailto:')) {
          e.preventDefault();
          e.stopPropagation();
          OpenURLInDefaultBrowser(href);
        }
      }
    };
    if (typeof document !== 'undefined') {
      document.addEventListener('click', handleGlobalLinkClick, true);
    }

    // 6. Global Wails Events
    try {
      // Direct navigation to tab event (e.g. from protocol activation or Windows Toast click)
      EventsOn('navigate:tab', (tab) => {
        if (tab && typeof tab === 'string') {
          activeTab = tab;
        }
      });

      // Backend toast notification (e.g. from background update check)
      EventsOn('toast:notify', (data) => {
        if (data) {
          showToast(data.type || 'info', data.title, data.message, data.actionTab);
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

      // Job completion tracker with instant optimistic UI update, background sync, and queue runner
      EventsOn('job:done', async (data) => {
        const finishedJob = activeJob;
        isJobRunning = false;

        if (data && data.success && finishedJob) {
          const pkgId = finishedJob.packageId ? finishedJob.packageId.toLowerCase() : null;
          const action = finishedJob.action;

          // 1. Instant optimistic update for installed set and badge counts
          if (action === 'install' && pkgId) {
            installedIds = new Set([...installedIds, pkgId]);
            installedCount += 1;
          } else if (action === 'uninstall' && pkgId) {
            const next = new Set(installedIds);
            next.delete(pkgId);
            installedIds = next;
            if (installedCount > 0) installedCount -= 1;
          } else if (action === 'upgrade') {
            if (pkgId === 'all') {
              updatesCount = 0;
            } else if (updatesCount > 0) {
              updatesCount -= 1;
            }
          }

          // 2. Trigger automatic background rescan
          try {
            await RefreshAll();
          } catch (e) {
            console.warn('Auto-refresh failed:', e);
          }
        }

        // Clean up completed package from runningJobs
        if (finishedJob && finishedJob.packageId) {
          const finishedKey = finishedJob.packageId.toLowerCase();
          const nextRunning = { ...runningJobs };
          delete nextRunning[finishedKey];
          runningJobs = nextRunning;
        }

        // 3. Process next item from installation queue if any!
        if (jobQueue.length > 0) {
          const nextJob = jobQueue[0];
          jobQueue = jobQueue.slice(1);

          showToast('info', 'Next Install Starting', `Starting installation of ${nextJob.name}...`);

          // Safe 600ms delay to ensure Windows Installer (msiexec.exe) releases its mutex
          setTimeout(async () => {
            try {
              let nextJobId = '';
              if (nextJob.action === 'install') {
                nextJobId = await InstallPackage(nextJob.id, nextJob.interactive || false, false, '');
              } else if (nextJob.action === 'download') {
                nextJobId = await DownloadPackage(nextJob.id, nextJob.name);
              } else if (nextJob.action === 'uninstall') {
                nextJobId = await UninstallPackage(nextJob.id, nextJob.name, false);
              }

              if (nextJobId) {
                handleStartJob(nextJobId, `${nextJob.action === 'download' ? 'Downloading' : 'Installing'} ${nextJob.name}`, {
                  packageId: nextJob.id,
                  action: nextJob.action
                });
              }
            } catch (err) {
              showToast('error', 'Queued Task Failed', `Failed to start ${nextJob.name}: ${err}`);
              const k = (nextJob.id || '').toLowerCase();
              const nextRunning = { ...runningJobs };
              delete nextRunning[k];
              runningJobs = nextRunning;
            }
          }, 600);
        } else {
          runningJobs = {};
        }
      });
      // 7. Check EasyWinGet Self-Update with 24-Hour Remind-Later Throttling
      setTimeout(checkStartupAppUpdate, 1500);
    } catch (e) {
      console.warn('EventsOn error:', e);
    }
  });

  async function checkStartupAppUpdate() {
    try {
      // 24-Hour Throttling: If user clicked "Remind me later", skip prompt for 24h
      const remindedAtStr = await GetSetting('easywinget_app_update_reminded_at', '0');
      const remindedAt = Number(remindedAtStr) || 0;
      const now = Date.now();
      const oneDayMs = 24 * 60 * 60 * 1000; // 24 hours

      if (remindedAt > 0 && (now - remindedAt) < oneDayMs) {
        return;
      }

      const info = await CheckAppSelfUpdate();
      if (info && info.hasUpdate) {
        appUpdateInfo = info;
        showAppUpdateModal = true;
      }
    } catch (err) {
      console.warn('Startup app update check failed:', err);
    }
  }

  async function handleAppUpdateNow() {
    if (appUpdateInfo) {
      const targetUrl = appUpdateInfo.downloadUrl || appUpdateInfo.releaseUrl || 'https://github.com/DARKBOYJYOTI/EasyWinget/releases/latest';
      await OpenURLInDefaultBrowser(targetUrl);
    }
    showAppUpdateModal = false;
  }

  async function handleAppRemindLater() {
    try {
      await SetSetting('easywinget_app_update_reminded_at', Date.now().toString());
    } catch (e) {}
    showAppUpdateModal = false;
    showToast('info', 'Update Postponed', 'We will remind you again tomorrow.');
  }

  function handleCloseAppUpdate() {
    handleAppRemindLater();
  }

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

  function showToast(type, title, message = '', actionTab = null) {
    if (toastTimer) {
      clearTimeout(toastTimer);
      toastTimer = null;
    }
    const id = Date.now() + Math.random();
    // Only one toast at a time: override any currently active toast
    toasts = [{ id, type, title, message, actionTab }];

    toastTimer = setTimeout(() => {
      toasts = [];
      toastTimer = null;
    }, 4200);
  }

  function removeToast(id) {
    if (toastTimer) {
      clearTimeout(toastTimer);
      toastTimer = null;
    }
    toasts = toasts.filter(t => t.id !== id);
  }

  function handleStartJob(jobId, title, options = {}) {
    previousTab = lastNonTerminalTab;
    const pkgId = options.packageId ? options.packageId.toLowerCase() : null;
    let action = options.action;
    if (!action) {
      const lower = title.toLowerCase();
      if (lower.includes('upgrad') || lower.includes('update')) action = 'upgrade';
      else if (lower.includes('uninstal')) action = 'uninstall';
      else if (lower.includes('download')) action = 'download';
      else action = 'install';
    }

    activeJob = { id: jobId, title, packageId: pkgId, action };
    isJobRunning = true;
    if (pkgId) {
      runningJobs = { ...runningJobs, [pkgId]: action };
    }
    activeTab = 'terminal';
  }

  function enqueueJob(item) {
    const pkgKey = (item.id || '').toLowerCase();
    if (runningJobs[pkgKey]) {
      showToast('warning', 'Already in Progress', `${item.name} is already ${runningJobs[pkgKey]}ing.`);
      return;
    }
    runningJobs = { ...runningJobs, [pkgKey]: 'queued' };
    jobQueue = [...jobQueue, item];
    showToast('info', 'Added to Queue', `${item.name} added to installation queue (#${jobQueue.length})`);
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
    {searchCount}
    {isJobRunning}
    version={appVersion}
  />

  <main class="main-content">
    <div class="view-panel" class:hidden={activeTab !== 'search'}>
      <SearchView 
        {installedIds}
        {runningJobs}
        {isJobRunning}
        onEnqueueJob={enqueueJob}
        onOpenDetails={openDetails}
        onStartJob={handleStartJob}
        onShowToast={showToast}
        onConfirmRequest={requestConfirm}
        onUpdateCount={(c) => searchCount = c}
      />
    </div>

    <div class="view-panel" class:hidden={activeTab !== 'installed'}>
      <InstalledView 
        {runningJobs}
        onStartJob={handleStartJob}
        onShowToast={showToast}
        onUpdateCount={(c) => installedCount = c}
        onConfirmRequest={requestConfirm}
      />
    </div>

    <div class="view-panel" class:hidden={activeTab !== 'updates'}>
      <UpdatesView 
        {runningJobs}
        onStartJob={handleStartJob}
        onShowToast={showToast}
        onUpdateCount={(c) => updatesCount = c}
      />
    </div>

    <div class="view-panel" class:hidden={activeTab !== 'downloads'}>
      <DownloadsView 
        onStartJob={handleStartJob}
        onShowToast={showToast}
        onUpdateCount={(c) => downloadsCount = c}
      />
    </div>

    {#if activeTab === 'terminal'}
      <div class="terminal-tab-wrapper view-animated">
        <TaskTerminal 
          {activeJob}
          onClose={() => { activeTab = previousTab || lastNonTerminalTab || 'search'; }}
          onShowToast={showToast}
          onJobDone={() => {
            isJobRunning = false;
            runningJobs = {};
          }}
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
      <div class="docked-taskbar" on:click={() => { previousTab = activeTab; activeTab = 'terminal'; }}>
        <div class="docked-pulse"></div>
        <Terminal size={15} color="var(--accent-light)" />
        <span class="docked-title">{activeJob?.title || 'Background Task'}</span>
        {#if jobQueue.length > 0}
          <span class="docked-queue-badge">+{jobQueue.length} in queue</span>
        {/if}
        <button class="btn-primary docked-btn">View Console</button>
        <button 
          class="docked-close" 
          title="Dismiss status bar" 
          on:click={(e) => { e.stopPropagation(); isJobRunning = false; runningJobs = {}; }}
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
      {runningJobs}
      {isJobRunning}
      onEnqueueJob={enqueueJob}
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

  <!-- EasyWinGet App Self-Update Modal -->
  {#if showAppUpdateModal && appUpdateInfo}
    <AppUpdateModal 
      updateInfo={appUpdateInfo}
      onUpdateNow={handleAppUpdateNow}
      onRemindLater={handleAppRemindLater}
      onClose={handleCloseAppUpdate}
    />
  {/if}

  <!-- Toast Container with Sound Synthesizer -->
  <Toast 
    {toasts} 
    {removeToast} 
    onToastClick={(t) => {
      if (t && t.actionTab) {
        activeTab = t.actionTab;
      }
    }}
  />
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

  .view-panel {
    height: 100%;
    width: 100%;
  }

  .view-panel.hidden {
    display: none !important;
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

  .docked-queue-badge {
    background: rgba(245, 158, 11, 0.18);
    border: 1px solid rgba(245, 158, 11, 0.45);
    color: #f59e0b;
    font-size: 0.74rem;
    font-weight: 600;
    padding: 2px 8px;
    border-radius: var(--radius-pill);
    white-space: nowrap;
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
