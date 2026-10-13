<script>
  import { onMount, onDestroy } from 'svelte';
  import { 
    HardDrive, 
    RefreshCw, 
    Download, 
    Trash2, 
    Search, 
    FileSpreadsheet,
    X,
    Loader2
  } from '@lucide/svelte';
  import AppIcon from './AppIcon.svelte';
  import SkeletonLoader from './SkeletonLoader.svelte';
  import { 
    GetInstalled, 
    UninstallPackage, 
    SelectExportPath, 
    ExportApps 
  } from '../../wailsjs/go/main/App.js';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js';

  export let onStartJob = (jobId, title, options = {}) => {};
  export let onShowToast = (type, title, msg) => {};
  export let onUpdateCount = (count) => {};
  export let onConfirmRequest = async (opts) => true;
  export let runningJobs = {};

  let installedApps = [];
  let isLoading = false;
  let filterText = '';
  let searchInputRef;

  $: cleanFilter = (filterText || '').trim().toLowerCase();

  $: filteredApps = !cleanFilter
    ? (installedApps || [])
    : (installedApps || []).filter(app => {
        const name = (app?.name || '').toLowerCase();
        const id = (app?.id || '').toLowerCase();
        return name.includes(cleanFilter) || id.includes(cleanFilter);
      });

  function clearFilter() {
    filterText = '';
    if (searchInputRef) {
      searchInputRef.value = '';
      searchInputRef.focus();
    }
  }

  function handleFilterKeydown(e) {
    if (e.key === 'Escape') {
      e.preventDefault();
      e.stopPropagation();
      clearFilter();
    }
  }

  function handleGlobalKeydown(e) {
    if (e.key === 'Escape' && filterText) {
      e.preventDefault();
      clearFilter();
    }
  }

  onMount(() => {
    loadApps(false);

    // Listen for realtime background updates from SQLite
    EventsOn('installed:updated', (apps) => {
      if (apps && Array.isArray(apps)) {
        installedApps = apps;
        onUpdateCount(installedApps.length);
      }
    });

    // Listen for job completion to refresh installed apps automatically
    EventsOn('job:done', async (data) => {
      if (data && data.success) {
        for (const [pkgId, action] of Object.entries(runningJobs)) {
          if (action === 'uninstall') {
            installedApps = installedApps.filter(a => 
              (a.id || '').toLowerCase() !== pkgId.toLowerCase() && 
              (a.name || '').toLowerCase() !== pkgId.toLowerCase()
            );
            onUpdateCount(installedApps.length);
          }
        }
        try {
          const apps = await GetInstalled(false);
          if (apps && Array.isArray(apps)) {
            installedApps = apps;
            onUpdateCount(installedApps.length);
          }
        } catch (e) {}
      }
    });

    return () => {
      EventsOff('installed:updated');
      EventsOff('job:done');
    };
  });

  async function loadApps(forceRefresh = false) {
    isLoading = true;
    try {
      const apps = await GetInstalled(forceRefresh);
      if (apps && Array.isArray(apps)) {
        installedApps = apps;
        onUpdateCount(installedApps.length);
      }
      if (forceRefresh) {
        onShowToast('success', 'Refreshed', `Loaded ${installedApps.length} installed apps`);
      }
    } catch (err) {
      onShowToast('error', 'Failed to load installed apps', String(err));
    } finally {
      isLoading = false;
    }
  }

  async function handleUninstall(app) {
    const ok = await onConfirmRequest({
      title: `Uninstall ${app.name}?`,
      message: `Are you sure you want to uninstall "${app.name}"? This will launch the uninstaller to remove it from your computer.`,
      confirmText: 'Uninstall',
      cancelText: 'Cancel',
      isDanger: true,
      iconType: 'danger'
    });
    if (!ok) return;

    try {
      const jobId = await UninstallPackage(app.id, app.name, false);
      onStartJob(jobId, `Uninstalling ${app.name}`, { packageId: app.id, action: 'uninstall' });
      onShowToast('info', 'Uninstall Started', `Uninstalling ${app.name}...`);
    } catch (err) {
      onShowToast('error', 'Uninstall Failed', String(err));
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

<svelte:window on:keydown={handleGlobalKeydown} />

<div class="installed-page view-animated">
  <!-- Pinned / Sticky Top Zone: Header + Search Filter Bar -->
  <div class="sticky-top-zone">
    <div class="page-header">
      <div class="header-left">
        <h2>Installed Applications</h2>
        <p>Manage, remove, and export all applications installed on your Windows system</p>
      </div>

      <div class="header-actions">
        <button class="btn-secondary" on:click={handleExport} disabled={isLoading || installedApps.length === 0}>
          <FileSpreadsheet size={15} /> Export List
        </button>
        <button class="btn-primary" on:click={() => loadApps(true)} disabled={isLoading}>
          <RefreshCw size={15} class={isLoading ? 'animate-spin' : ''} /> Refresh
        </button>
      </div>
    </div>

    <div class="filter-bar">
      <div class="search-input-wrap">
        <Search size={16} class="search-icon" />
        <input 
          bind:this={searchInputRef}
          type="text" 
          placeholder="Filter installed apps by name or ID... (Esc to clear)"
          bind:value={filterText}
          on:keydown={handleFilterKeydown}
        />
        {#if filterText}
          <button class="clear-btn" on:click={clearFilter} title="Clear filter (Esc)" aria-label="Clear filter">
            <X size={15} />
          </button>
        {/if}
      </div>
      <div class="count-badge">
        <span>{filteredApps.length} of {installedApps.length} apps</span>
      </div>
    </div>
  </div>

  <div class="installed-body">
    {#if isLoading && installedApps.length === 0}
      <SkeletonLoader type="row" count={9} />
    {:else if filteredApps.length === 0 && !isLoading}
      <div class="state-container">
        <p class="state-title">No matching applications</p>
        {#if cleanFilter}
          <p class="state-sub">No installed packages found matching "{filterText}"</p>
          <button class="btn-secondary" on:click={clearFilter} style="margin-top: 10px;">
            <X size={14} /> Clear Filter & Show All ({installedApps.length})
          </button>
        {:else}
          <p class="state-sub">No installed applications found on system</p>
          <button class="btn-primary" on:click={() => loadApps(true)} style="margin-top: 10px;">
            <RefreshCw size={14} /> Scan Again
          </button>
        {/if}
      </div>
    {:else}
      <div class="apps-list">
        {#each filteredApps as app, index (app.id ? `${app.id}__${app.version || ''}__${index}` : `app__${index}`)}
          {@const appKey = (app.id || '').toLowerCase()}
          <div class="app-row">
            <AppIcon packageId={app.id} name={app.name} size={44} />
            
            <div class="app-details">
              <div class="app-title-line">
                <span class="app-name" title={app.name}>{app.name}</span>
                {#if app.version}
                  <span class="badge badge-blue">{app.version}</span>
                {/if}
              </div>
              <div class="app-id" title={app.id}>{app.id}</div>
            </div>

            <div class="app-actions">
              {#if runningJobs[appKey] === 'uninstall'}
                <button 
                  class="btn-danger btn-busy" 
                  title="Uninstalling {app.name}..."
                  disabled
                >
                  <Loader2 size={15} class="animate-spin" /> Uninstalling...
                </button>
              {:else}
                <button 
                  class="btn-danger" 
                  title="Uninstall {app.name}"
                  on:click={() => handleUninstall(app)}
                >
                  <Trash2 size={15} /> Uninstall
                </button>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .installed-page {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
    padding: 0;
  }

  .sticky-top-zone {
    background: var(--bg-app);
    border-bottom: 1px solid var(--border-subtle);
    padding: 24px 36px 14px 36px;
    flex-shrink: 0;
    z-index: 10;
    backdrop-filter: blur(10px);
  }

  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 18px;
  }

  .header-left h2 {
    font-size: 1.6rem;
    font-weight: 600;
    letter-spacing: -0.01em;
    color: var(--text-primary);
  }

  .header-left p {
    color: var(--text-secondary);
    font-size: 0.9rem;
    margin-top: 4px;
  }

  .header-actions {
    display: flex;
    gap: 8px;
  }

  .filter-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }

  .search-input-wrap {
    display: flex;
    align-items: center;
    gap: 10px;
    background: var(--bg-card);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-pill);
    padding: 4px 16px;
    flex: 1;
    max-width: 480px;
    height: 40px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
    transition: background-color 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease;
  }
  .search-input-wrap:hover {
    border-color: var(--border-card-hover);
  }
  .search-input-wrap:focus-within {
    background: var(--bg-surface);
    border-color: var(--accent-light);
    box-shadow: 0 0 0 2px var(--accent-badge-border), 0 0 16px var(--accent-glow);
  }

  .search-input-wrap input {
    flex: 1;
    background: transparent;
    border: none;
    box-shadow: none;
    padding: 6px 0;
    font-size: 0.92rem;
    color: #ffffff;
  }

  .search-input-wrap input::placeholder {
    color: rgba(255, 255, 255, 0.65);
    opacity: 1;
  }

  :global([data-theme="light"]) .search-input-wrap input {
    color: #111827;
  }

  :global([data-theme="light"]) .search-input-wrap input::placeholder {
    color: rgba(0, 0, 0, 0.55);
  }

  .clear-btn {
    background: transparent;
    color: var(--text-muted);
    border: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 4px;
    border-radius: 50%;
    transition: color 0.15s ease, background-color 0.15s ease;
  }
  .clear-btn:hover {
    color: var(--text-primary);
    background: var(--bg-card-hover);
  }

  .count-badge {
    font-size: 0.84rem;
    color: var(--text-muted);
  }

  .installed-body {
    flex: 1;
    overflow-y: auto;
    padding: 16px 36px 36px 36px;
  }

  .state-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 70px 20px;
    gap: 14px;
    color: var(--text-secondary);
  }

  .apps-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .app-row {
    display: flex;
    align-items: center;
    gap: 14px;
    background: var(--bg-card);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-sm);
    padding: 10px 16px;
    transition: background-color 0.18s cubic-bezier(0.1, 0.9, 0.2, 1), 
                border-color 0.18s cubic-bezier(0.1, 0.9, 0.2, 1), 
                box-shadow 0.18s cubic-bezier(0.1, 0.9, 0.2, 1);
  }

  .app-row:hover {
    background: var(--bg-card-hover);
    border-color: var(--accent-badge-border);
    box-shadow: 0 0 0 1px var(--accent-badge-border), 0 2px 12px var(--accent-glow);
  }

  .app-details {
    flex: 1;
    min-width: 0;
  }

  .app-title-line {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .app-name {
    font-weight: 600;
    font-size: 0.95rem;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .app-id {
    font-size: 0.78rem;
    color: var(--text-muted);
    font-family: var(--font-mono);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    margin-top: 2px;
  }

  .app-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }
</style>
