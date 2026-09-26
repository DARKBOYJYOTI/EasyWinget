<script>
  import { onMount } from 'svelte';
  import { 
    X, 
    ExternalLink, 
    Download, 
    FileText, 
    Shield, 
    Tag, 
    Loader2,
    Trash2
  } from '@lucide/svelte';
  import AppIcon from './AppIcon.svelte';
  import { 
    GetAppDetails, 
    InstallPackage, 
    DownloadPackage,
    UninstallPackage
  } from '../../wailsjs/go/main/App.js';

  export let packageId = '';
  export let installedIds = new Set();
  export let onClose = () => {};
  export let onStartJob = (jobId, title) => {};
  export let onShowToast = (type, title, msg) => {};
  export let onConfirmRequest = async (opts) => true;

  let details = null;
  let isLoading = true;

  $: isInstalled = (() => {
    if (!installedIds || !(installedIds instanceof Set)) return false;
    const pid = (packageId || '').toLowerCase();
    const dname = (details?.name || '').toLowerCase();
    const did = (details?.id || '').toLowerCase();
    return (pid && installedIds.has(pid)) || 
           (did && installedIds.has(did)) || 
           (dname && installedIds.has(dname));
  })();

  onMount(async () => {
    try {
      details = await GetAppDetails(packageId);
    } catch (err) {
      onShowToast('error', 'Details Error', String(err));
    } finally {
      isLoading = false;
    }
  });

  async function handleInstall(interactive = false) {
    if (!details) return;
    try {
      const mode = interactive ? 'Interactive' : 'Silent';
      const jobId = await InstallPackage(details.id, interactive, false, '');
      onStartJob(jobId, `Installing ${details.name || details.id} (${mode})`);
      onClose();
    } catch (err) {
      onShowToast('error', 'Install Error', String(err));
    }
  }

  async function handleUninstall() {
    if (!details) return;
    const appName = details.name || packageId;
    const ok = await onConfirmRequest({
      title: `Uninstall ${appName}?`,
      message: `Are you sure you want to uninstall "${appName}" from your computer?`,
      confirmText: 'Uninstall',
      cancelText: 'Cancel',
      isDanger: true,
      iconType: 'danger'
    });
    if (!ok) return;

    try {
      const jobId = await UninstallPackage(details.id || packageId, appName, false);
      onStartJob(jobId, `Uninstalling ${appName}`);
      onShowToast('info', 'Uninstall Started', `Uninstalling ${appName}...`);
      onClose();
    } catch (err) {
      onShowToast('error', 'Uninstall Failed', String(err));
    }
  }

  async function handleDownload() {
    if (!details) return;
    try {
      const jobId = await DownloadPackage(details.id, details.name);
      onStartJob(jobId, `Downloading ${details.name || details.id}`);
      onClose();
    } catch (err) {
      onShowToast('error', 'Download Error', String(err));
    }
  }
</script>

<div class="modal-backdrop" on:click={onClose}>
  <div class="modal-card" on:click={(e) => e.stopPropagation()}>
    <div class="modal-header">
      <div class="header-main">
        <AppIcon packageId={packageId} name={details?.name || packageId} size={42} homepage={details?.homepage} />
        <div>
          <h3>{details?.name || packageId}</h3>
          <span class="header-id">{packageId}</span>
        </div>
      </div>
      <button class="close-btn" on:click={onClose}><X size={16} /></button>
    </div>

    <div class="modal-body">
      {#if isLoading}
        <div class="loading-state">
          <Loader2 size={36} class="animate-spin text-accent" />
          <p>Fetching WinGet manifest details...</p>
        </div>
      {:else if details}
        {#if details.description}
          <div class="info-group">
            <label>Description</label>
            <p class="description-text">{details.description}</p>
          </div>
        {/if}

        <div class="grid-props">
          {#if details.version}
            <div class="prop-item">
              <span class="prop-label">Version</span>
              <span class="prop-val mono-val">{details.version}</span>
            </div>
          {/if}

          {#if details.publisher}
            <div class="prop-item">
              <span class="prop-label">Publisher</span>
              <span class="prop-val">{details.publisher}</span>
            </div>
          {/if}

          {#if details.author}
            <div class="prop-item">
              <span class="prop-label">Author</span>
              <span class="prop-val">{details.author}</span>
            </div>
          {/if}

          {#if details.license}
            <div class="prop-item">
              <span class="prop-label">License</span>
              <span class="prop-val">{details.license}</span>
            </div>
          {/if}

          {#if details.installerType}
            <div class="prop-item">
              <span class="prop-label">Installer Type</span>
              <span class="prop-val uppercase">{details.installerType}</span>
            </div>
          {/if}

          {#if details.installerSha256}
            <div class="prop-item full-width">
              <span class="prop-label">SHA256 Hash</span>
              <span class="prop-val hash-val">{details.installerSha256}</span>
            </div>
          {/if}
        </div>

        <!-- Links -->
        {#if details.homepage || details.releaseNotesUrl}
          <div class="links-row">
            {#if details.homepage}
              <a href={details.homepage} target="_blank" rel="noreferrer" class="ext-link">
                <ExternalLink size={14} class="link-icon" /> Official Website
              </a>
            {/if}
            {#if details.releaseNotesUrl}
              <a href={details.releaseNotesUrl} target="_blank" rel="noreferrer" class="ext-link">
                <FileText size={14} class="link-icon" /> Release Notes
              </a>
            {/if}
          </div>
        {/if}

        <!-- Tags -->
        {#if details.tags && details.tags.length > 0}
          <div class="tags-wrap">
            <span class="tag-title"><Tag size={14} /> Tags:</span>
            {#each details.tags as tag}
              <span class="chip-tag">{tag}</span>
            {/each}
          </div>
        {/if}
      {:else}
        <p class="error-msg">Could not load manifest for this package.</p>
      {/if}
    </div>

    <div class="modal-footer">
      <button class="btn-secondary" on:click={handleDownload} disabled={isLoading || !details}>
        <Download size={15} /> Download
      </button>
      {#if isInstalled}
        <button class="btn-danger" on:click={handleUninstall} disabled={isLoading || !details}>
          <Trash2 size={15} /> Uninstall
        </button>
      {:else}
        <button class="btn-primary" on:click={() => handleInstall(false)} disabled={isLoading || !details}>
          Install
        </button>
      {/if}
    </div>
  </div>
</div>

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.72);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 24px;
    animation: modalBackdropFade 0.18s ease-out;
  }


  .modal-card {
    background: var(--bg-sidebar);
    border: 1px solid var(--border-card-hover);
    border-radius: var(--radius-lg);
    width: 100%;
    max-width: 640px;
    max-height: 85vh;
    display: flex;
    flex-direction: column;
    box-shadow: var(--shadow-flyout);
    overflow: hidden;
    animation: modalScaleIn 0.22s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .modal-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    padding: 20px 24px;
    border-bottom: 1px solid var(--border-subtle);
  }

  .header-main h3 {
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .header-id {
    font-size: 0.8rem;
    color: var(--text-muted);
    font-family: var(--font-mono);
    margin-top: 2px;
  }

  .close-btn {
    background: transparent;
    color: var(--text-muted);
    padding: 6px;
    border-radius: var(--radius-sm);
    transition: all var(--transition-fast);
  }
  .close-btn:hover {
    color: var(--text-primary);
    background: rgba(255, 255, 255, 0.08);
  }

  .modal-body {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 20px;
    scroll-behavior: smooth;
  }

  .loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 60px 0;
    gap: 14px;
    color: var(--text-secondary);
  }

  .info-group label {
    display: block;
    font-size: 0.76rem;
    font-weight: 600;
    text-transform: uppercase;
    color: var(--text-muted);
    margin-bottom: 6px;
    letter-spacing: 0.04em;
  }

  .description-text {
    font-size: 0.92rem;
    line-height: 1.6;
    color: var(--text-secondary);
  }

  .grid-props {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 14px;
    background: rgba(0, 0, 0, 0.25);
    border: 1px solid var(--border-subtle);
    padding: 16px;
    border-radius: var(--radius-md);
  }

  .prop-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .prop-item.full-width {
    grid-column: span 2;
  }

  .prop-label {
    font-size: 0.76rem;
    color: var(--text-muted);
  }

  .prop-val {
    font-size: 0.9rem;
    color: var(--text-primary);
    font-weight: 500;
  }

  .mono-val {
    font-family: var(--font-mono);
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--accent-light);
    letter-spacing: 0.02em;
  }

  .hash-val {
    font-family: var(--font-mono);
    font-size: 0.76rem;
    word-break: break-all;
    color: var(--text-secondary);
  }

  .uppercase {
    text-transform: uppercase;
  }

  .links-row {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-top: 4px;
  }

  .ext-link {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    color: var(--accent-light);
    background: var(--accent-badge-bg);
    border: 1px solid var(--accent-badge-border);
    padding: 7px 14px;
    border-radius: var(--radius-sm);
    text-decoration: none;
    font-size: 0.88rem;
    font-weight: 600;
    transition: all 0.2s cubic-bezier(0.1, 0.9, 0.2, 1);
  }

  .ext-link:hover {
    background: var(--accent-glow);
    border-color: var(--accent-light);
    color: #ffffff;
    box-shadow: 0 0 14px var(--accent-glow);
    transform: none;
  }

  :global(.link-icon) {
    color: var(--accent-light);
    transition: transform 0.2s ease;
  }

  .ext-link:hover :global(.link-icon) {
    transform: scale(1.1);
  }

  .tags-wrap {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
  }

  .tag-title {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 0.82rem;
    color: var(--text-muted);
  }

  .chip-tag {
    font-size: 0.78rem;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-subtle);
    padding: 4px 10px;
    border-radius: var(--radius-full);
    color: var(--text-secondary);
  }

  .modal-footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 12px;
    padding: 16px 24px;
    border-top: 1px solid var(--border-subtle);
    background: rgba(0, 0, 0, 0.2);
  }

  @keyframes modalBackdropFade {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes modalScaleIn {
    from {
      opacity: 0;
      transform: scale(0.96) translateY(6px);
    }
    to {
      opacity: 1;
      transform: scale(1) translateY(0);
    }
  }
</style>
