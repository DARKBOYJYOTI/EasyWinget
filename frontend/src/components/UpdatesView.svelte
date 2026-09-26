<script>
  import { onMount, onDestroy } from 'svelte';
  import { 
    ArrowUpCircle, 
    RefreshCw, 
    CheckCheck, 
    EyeOff, 
    Eye, 
    ArrowRight,
    CheckCircle2,
    ChevronDown,
    ChevronUp
  } from '@lucide/svelte';
  import AppIcon from './AppIcon.svelte';
  import SkeletonLoader from './SkeletonLoader.svelte';
  import { 
    GetUpdates, 
    UpgradePackage, 
    UpgradeAll, 
    GetIgnoredApps, 
    IgnoreApp, 
    UnignoreApp 
  } from '../../wailsjs/go/main/App.js';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js';

  export let onStartJob = (jobId, title) => {};
  export let onShowToast = (type, title, msg) => {};
  export let onUpdateCount = (count) => {};

  let updates = [];
  let ignoredApps = [];
  let isLoading = false;
  let showIgnored = false;

  onMount(() => {
    loadData(false);

    // Listen for realtime background updates from SQLite
    EventsOn('updates:updated', (upList) => {
      if (upList && Array.isArray(upList)) {
        updates = upList;
        onUpdateCount(updates.length);
      }
    });

    return () => {
      EventsOff('updates:updated');
    };
  });

  async function loadData(forceRefresh = false) {
    isLoading = true;
    try {
      const [upList, igList] = await Promise.all([
        GetUpdates(forceRefresh),
        GetIgnoredApps()
      ]);
      updates = upList || [];
      ignoredApps = igList || [];
      onUpdateCount(updates.length);

      if (forceRefresh) {
        onShowToast('success', 'Updates Refreshed', `Found ${updates.length} available updates`);
      }
    } catch (err) {
      onShowToast('error', 'Error Checking Updates', String(err));
    } finally {
      isLoading = false;
    }
  }

  async function handleUpgrade(app) {
    try {
      const jobId = await UpgradePackage(app.id);
      onStartJob(jobId, `Upgrading ${app.name}`);
      onShowToast('info', 'Update Started', `Updating ${app.name}...`);
    } catch (err) {
      onShowToast('error', 'Update Failed', String(err));
    }
  }

  async function handleUpgradeAll() {
    if (updates.length === 0) return;
    try {
      const jobId = await UpgradeAll();
      onStartJob(jobId, 'Upgrading All Outdated Packages');
      onShowToast('info', 'Batch Update Started', 'Updating all available applications...');
    } catch (err) {
      onShowToast('error', 'Batch Update Failed', String(err));
    }
  }

  async function handleIgnore(app) {
    try {
      await IgnoreApp(app.id, app.name);
      onShowToast('info', 'Ignored Update', `Ignored future updates for ${app.name}`);
      await loadData(false);
    } catch (err) {
      onShowToast('error', 'Action Failed', String(err));
    }
  }

  async function handleUnignore(app) {
    try {
      await UnignoreApp(app.id);
      onShowToast('success', 'Restored', `Unignored ${app.name}`);
      await loadData(false);
    } catch (err) {
      onShowToast('error', 'Action Failed', String(err));
    }
  }
</script>

<div class="updates-page view-animated">
  <!-- Pinned Sticky Header Zone -->
  <div class="sticky-top-zone">
    <div class="page-header">
      <div class="header-left">
        <h2>Available Updates</h2>
        <p>Keep your system and applications secure with the latest official releases</p>
      </div>

      <div class="header-actions">
        <button 
          class="btn-primary" 
          on:click={handleUpgradeAll} 
          disabled={isLoading || updates.length === 0}
        >
          <CheckCheck size={15} /> Update All ({updates.length})
        </button>
        <button 
          class="btn-secondary" 
          on:click={() => loadData(true)} 
          disabled={isLoading}
        >
          <RefreshCw size={15} class={isLoading ? 'animate-spin' : ''} /> Check Updates
        </button>
      </div>
    </div>
  </div>

  <div class="updates-body">
    {#if isLoading && updates.length === 0}
      <SkeletonLoader type="row" count={6} />
    {:else if updates.length === 0}
      <div class="state-container all-good">
        <CheckCircle2 size={48} color="var(--color-success-light)" />
        <p class="state-title">Everything is up to date!</p>
        <p class="state-sub">All installed packages match the latest repository versions</p>
      </div>
    {:else}
      <div class="updates-list">
        {#each updates as app (app.id)}
          <div class="update-card">
            <AppIcon packageId={app.id} name={app.name} size={46} />

            <div class="update-info">
              <div class="app-name" title={app.name}>{app.name}</div>
              <div class="app-id" title={app.id}>{app.id}</div>
              
              <div class="version-diff">
                <span class="ver-old">{app.version || 'installed'}</span>
                <ArrowRight size={14} class="arrow-icon" />
                <span class="ver-new">{app.available || 'latest'}</span>
              </div>
            </div>

            <div class="update-actions">
              <button class="btn-primary" on:click={() => handleUpgrade(app)}>
                <ArrowUpCircle size={16} /> Update
              </button>
              <button 
                class="btn-ghost" 
                title="Ignore this update" 
                on:click={() => handleIgnore(app)}
              >
                <EyeOff size={15} />
              </button>
            </div>
          </div>
        {/each}
      </div>
    {/if}

    <!-- Modern Redesigned Ignored Updates Section -->
    {#if ignoredApps.length > 0}
      <div class="ignored-card">
        <button class="ignored-header" on:click={() => showIgnored = !showIgnored}>
          <div class="ignored-header-left">
            <div class="ignored-icon-wrap">
              <EyeOff size={15} />
            </div>
            <div class="ignored-title-group">
              <span class="ignored-title">Ignored Updates</span>
              <span class="ignored-pill">{ignoredApps.length} {ignoredApps.length === 1 ? 'app' : 'apps'} hidden</span>
            </div>
          </div>
          <div class="ignored-chevron">
            {#if showIgnored}
              <ChevronUp size={16} />
            {:else}
              <ChevronDown size={16} />
            {/if}
          </div>
        </button>

        {#if showIgnored}
          <div class="ignored-list-container">
            <div class="ignored-list">
              {#each ignoredApps as app (app.id)}
                <div class="ignored-app-row">
                  <div class="ignored-left">
                    <AppIcon packageId={app.id} name={app.name} size={36} />
                    <div class="ignored-info">
                      <div class="ignored-name">{app.name || app.id}</div>
                      <div class="ignored-id-badge">{app.id}</div>
                    </div>
                  </div>
                  <button 
                    class="btn-unignore" 
                    on:click={() => handleUnignore(app)}
                    title="Restore {app.name || app.id} to available updates"
                  >
                    <Eye size={13} /> Unignore
                  </button>
                </div>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>

<style>
  .updates-page {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
    padding: 0;
  }

  .sticky-top-zone {
    background: var(--bg-app);
    border-bottom: 1px solid var(--border-subtle);
    padding: 24px 36px 16px 36px;
    flex-shrink: 0;
    z-index: 10;
    backdrop-filter: blur(10px);
  }

  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
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

  .updates-body {
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 20px;
    padding: 20px 36px 36px 36px;
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

  .all-good {
    padding: 80px 20px;
  }

  .state-title {
    font-size: 1.1rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .state-sub {
    font-size: 0.85rem;
    color: var(--text-muted);
  }

  .updates-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .update-card {
    display: flex;
    align-items: center;
    gap: 14px;
    background: var(--bg-card);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-sm);
    padding: 12px 16px;
    transition: background-color 0.18s cubic-bezier(0.1, 0.9, 0.2, 1), 
                border-color 0.18s cubic-bezier(0.1, 0.9, 0.2, 1), 
                box-shadow 0.18s cubic-bezier(0.1, 0.9, 0.2, 1);
  }

  .update-card:hover {
    background: var(--bg-card-hover);
    border-color: var(--accent-badge-border);
    box-shadow: 0 0 0 1px var(--accent-badge-border), 0 2px 12px var(--accent-glow);
  }

  .update-info {
    flex: 1;
    min-width: 0;
  }

  .app-name {
    font-weight: 600;
    font-size: 0.96rem;
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

  .version-diff {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    margin-top: 4px;
    font-size: 0.78rem;
    font-family: var(--font-mono);
  }

  .ver-old {
    color: var(--text-muted);
    text-decoration: line-through;
  }

  :global(.arrow-icon) {
    color: var(--accent-light);
  }

  .ver-new {
    color: var(--color-success-light);
    font-weight: 600;
    background: var(--color-success-bg);
    padding: 1px 6px;
    border-radius: var(--radius-xs);
  }

  .update-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  /* Modern Redesigned Ignored Updates Section */
  .ignored-card {
    background: var(--bg-card);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-md);
    overflow: hidden;
    margin-top: 14px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    transition: border-color 0.2s ease, box-shadow 0.2s ease;
  }
  .ignored-card:hover {
    border-color: var(--border-card-hover);
  }

  .ignored-header {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 20px;
    background: transparent;
    border: none;
    cursor: pointer;
    transition: background-color 0.18s ease;
  }
  .ignored-header:hover {
    background: var(--bg-card-hover);
  }

  .ignored-header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .ignored-icon-wrap {
    width: 32px;
    height: 32px;
    border-radius: var(--radius-sm);
    background: rgba(255, 255, 255, 0.06);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-secondary);
  }

  :global([data-theme="light"]) .ignored-icon-wrap {
    background: rgba(0, 0, 0, 0.05);
  }

  .ignored-title-group {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .ignored-title {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .ignored-pill {
    font-size: 0.75rem;
    font-weight: 500;
    padding: 2px 8px;
    border-radius: var(--radius-pill);
    background: var(--border-card);
    color: var(--text-muted);
  }

  .ignored-chevron {
    color: var(--text-muted);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: transform 0.2s ease, color 0.2s ease;
  }
  .ignored-header:hover .ignored-chevron {
    color: var(--text-primary);
  }

  .ignored-list-container {
    border-top: 1px solid var(--border-subtle);
    background: rgba(0, 0, 0, 0.12);
    padding: 12px 16px;
  }

  :global([data-theme="light"]) .ignored-list-container {
    background: rgba(0, 0, 0, 0.02);
  }

  .ignored-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .ignored-app-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 14px;
    padding: 10px 16px;
    background: var(--bg-card);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-sm);
    transition: background-color 0.18s ease, border-color 0.18s ease;
  }
  .ignored-app-row:hover {
    background: var(--bg-card-hover);
    border-color: var(--border-card-hover);
  }

  .ignored-left {
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;
  }

  .ignored-info {
    min-width: 0;
  }

  .ignored-name {
    font-size: 0.92rem;
    font-weight: 600;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .ignored-id-badge {
    font-size: 0.76rem;
    color: var(--text-muted);
    font-family: var(--font-mono);
    margin-top: 2px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .btn-unignore {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 14px;
    font-size: 0.82rem;
    font-weight: 500;
    border-radius: var(--radius-sm);
    background: var(--bg-surface);
    border: 1px solid var(--border-card-hover);
    color: var(--text-primary);
    cursor: pointer;
    transition: all 0.18s ease;
    flex-shrink: 0;
  }
  .btn-unignore:hover {
    background: var(--accent-primary);
    border-color: var(--accent-light);
    color: #ffffff;
    box-shadow: 0 0 12px var(--accent-glow);
  }
</style>
