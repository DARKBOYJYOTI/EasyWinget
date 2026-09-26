<script>
  import { onMount } from 'svelte';
  import { 
    Download, 
    RefreshCw, 
    Play, 
    Folder, 
    Trash2, 
    FileText, 
    Loader2 
  } from '@lucide/svelte';
  import { 
    GetDownloadedFiles, 
    RunDownloadedInstaller, 
    OpenDownloadedFolder, 
    OpenDownloadsDirectory,
    DeleteDownloadedFile 
  } from '../../wailsjs/go/main/App.js';

  export let onStartJob = (jobId, title) => {};
  export let onShowToast = (type, title, msg) => {};
  export let onUpdateCount = (count) => {};

  let files = [];
  let isLoading = false;

  onMount(() => {
    loadFiles();
  });

  function formatBytes(bytes) {
    if (!bytes || bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  function formatDate(isoStr) {
    if (!isoStr) return '';
    try {
      const d = new Date(isoStr);
      return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch (e) {
      return isoStr;
    }
  }

  async function loadFiles() {
    isLoading = true;
    try {
      const f = await GetDownloadedFiles();
      files = f || [];
      onUpdateCount(files.length);
    } catch (err) {
      onShowToast('error', 'Error Reading Downloads', String(err));
    } finally {
      isLoading = false;
    }
  }

  async function handleRun(file) {
    try {
      const jobId = await RunDownloadedInstaller(file.name);
      onStartJob(jobId, `Launching Installer: ${file.name}`);
      onShowToast('info', 'Launching', `Starting ${file.name}...`);
    } catch (err) {
      onShowToast('error', 'Launch Error', String(err));
    }
  }

  async function handleOpen(file) {
    try {
      await OpenDownloadedFolder(file.name);
    } catch (err) {
      onShowToast('error', 'Open Folder Error', String(err));
    }
  }

  async function handleOpenFolder() {
    try {
      await OpenDownloadsDirectory();
    } catch (err) {
      onShowToast('error', 'Open Folder Error', String(err));
    }
  }

  async function handleDelete(file) {
    try {
      await DeleteDownloadedFile(file.name);
      onShowToast('info', 'Deleted File', `Removed ${file.name}`);
      await loadFiles();
    } catch (err) {
      onShowToast('error', 'Delete Error', String(err));
    }
  }
</script>

<div class="downloads-page view-animated">
  <!-- Pinned Sticky Top Header Zone -->
  <div class="sticky-top-zone">
    <div class="page-header">
      <div class="header-left">
        <h2>Downloaded Installers</h2>
        <p>Offline packages downloaded for backup, re-installation, or portable deployment</p>
      </div>

      <div class="header-actions">
        <button class="btn-secondary" on:click={handleOpenFolder} title="Open EasyWinGet Downloads Folder in Explorer">
          <Folder size={15} /> Open Folder
        </button>
        <button class="btn-secondary" on:click={loadFiles} disabled={isLoading}>
          <RefreshCw size={15} class={isLoading ? 'animate-spin' : ''} /> Refresh
        </button>
      </div>
    </div>
  </div>

  <div class="downloads-body">
    {#if isLoading && files.length === 0}
      <div class="state-container">
        <Loader2 size={40} class="animate-spin text-accent" />
        <p>Scanning Downloads directory...</p>
      </div>
    {:else if files.length === 0}
      <div class="state-container">
        <Download size={48} class="text-muted" />
        <p class="state-title">No offline installers downloaded</p>
        <p class="state-sub">Use the "Download" button on any package in Discover to save it offline</p>
      </div>
    {:else}
      <div class="files-list">
        {#each files as f (f.name)}
          <div class="file-card">
            <div class="file-icon">
              <FileText size={22} color="#38bdf8" />
            </div>

            <div class="file-info">
              <div class="file-name" title={f.name}>{f.name}</div>
              <div class="file-meta">
                <span class="file-size">{formatBytes(f.length)}</span>
                <span class="meta-dot">•</span>
                <span class="file-date">{formatDate(f.lastWriteTime)}</span>
              </div>
            </div>

            <div class="file-actions">
              <button class="btn-primary" on:click={() => handleRun(f)}>
                <Play size={15} /> Install
              </button>
              <button class="btn-secondary" title="Reveal in Windows Explorer" on:click={() => handleOpen(f)}>
                <Folder size={15} /> Folder
              </button>
              <button class="btn-danger" title="Delete Installer" on:click={() => handleDelete(f)}>
                <Trash2 size={15} />
              </button>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .downloads-page {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
    padding: 0;
  }

  .sticky-top-zone {
    background: var(--bg-app);
    border-bottom: 1px solid var(--border-subtle);
    padding: 24px 36px 18px 36px;
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

  .downloads-body {
    flex: 1;
    overflow-y: auto;
    padding: 24px 36px 36px 36px;
  }

  .state-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 80px 20px;
    gap: 16px;
    color: var(--text-secondary);
  }

  .state-title {
    font-size: 1.2rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .state-sub {
    font-size: 0.9rem;
    color: var(--text-muted);
  }

  .files-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .file-card {
    display: flex;
    align-items: center;
    gap: 16px;
    background: var(--bg-card);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-md);
    padding: 16px 20px;
    transition: background-color 0.18s cubic-bezier(0.1, 0.9, 0.2, 1), 
                border-color 0.18s cubic-bezier(0.1, 0.9, 0.2, 1), 
                box-shadow 0.18s cubic-bezier(0.1, 0.9, 0.2, 1);
  }

  .file-card:hover {
    background: var(--bg-card-hover);
    border-color: var(--accent-badge-border);
    box-shadow: 0 0 0 1px var(--accent-badge-border), 0 2px 14px var(--accent-glow);
  }

  .file-icon {
    width: 38px;
    height: 38px;
    border-radius: var(--radius-sm);
    background: rgba(96, 205, 255, 0.12);
    border: 1px solid rgba(96, 205, 255, 0.2);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .file-info {
    flex: 1;
    min-width: 0;
  }

  .file-name {
    font-weight: 600;
    font-size: 0.98rem;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .file-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.8rem;
    color: var(--text-muted);
    margin-top: 3px;
  }

  .file-size {
    color: var(--accent-primary);
    font-weight: 500;
  }

  .meta-dot {
    opacity: 0.5;
  }

  .file-actions {
    display: flex;
    align-items: center;
    gap: 10px;
  }
</style>
