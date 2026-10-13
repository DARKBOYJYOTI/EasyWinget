<script>
  import { 
    Sparkles, 
    Download, 
    Clock, 
    X, 
    ExternalLink, 
    ArrowRight, 
    ShieldCheck,
    FileText
  } from '@lucide/svelte';

  export let updateInfo = {
    hasUpdate: false,
    currentVersion: '5.0.0',
    latestVersion: '5.0.0',
    releaseName: '',
    releaseNotes: '',
    releaseUrl: '',
    downloadUrl: '',
    publishedAt: ''
  };

  export let onUpdateNow = () => {};
  export let onRemindLater = () => {};
  export let onClose = () => {};

  function formatDate(isoStr) {
    if (!isoStr) return '';
    try {
      const d = new Date(isoStr);
      return d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
    } catch (e) {
      return '';
    }
  }

  function handleBackdropClick(e) {
    if (e.target === e.currentTarget) {
      onRemindLater();
    }
  }

  function handleKeydown(e) {
    if (e.key === 'Escape') {
      onRemindLater();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<!-- svelte-ignore a11y-click-events-have-key-events -->
<div class="update-modal-backdrop" on:click={handleBackdropClick} role="presentation">
  <div class="update-modal-dialog" role="dialog" aria-modal="true" aria-labelledby="update-modal-title">
    
    <!-- Header with glowing badge and close button -->
    <div class="modal-header">
      <div class="header-left">
        <div class="glow-icon-box">
          <Sparkles size={22} color="var(--accent-light, #60cdff)" />
        </div>
        <div class="header-text">
          <div class="modal-title-row">
            <h2 id="update-modal-title" class="modal-title">EasyWinGet Update Available</h2>
            <span class="badge-new">New Release</span>
          </div>
          <p class="modal-subtitle">A newer, faster version of EasyWinGet is ready to install</p>
        </div>
      </div>
      <button class="btn-close" on:click={onClose} title="Dismiss (Remind Later)" aria-label="Close">
        <X size={18} />
      </button>
    </div>

    <!-- Version Comparison Card -->
    <div class="version-banner">
      <div class="version-pill current">
        <span class="version-label">Current Version</span>
        <span class="version-value">v{updateInfo?.currentVersion || '5.0.0'}</span>
      </div>

      <div class="version-arrow">
        <ArrowRight size={20} color="var(--accent-light, #60cdff)" />
      </div>

      <div class="version-pill latest">
        <span class="version-label">Latest Version</span>
        <span class="version-value highlight">v{updateInfo?.latestVersion || 'Latest'}</span>
      </div>
    </div>

    <!-- Release Info & Changelog Preview -->
    <div class="release-info-section">
      <div class="release-meta-row">
        <div class="release-title-group">
          <FileText size={15} color="var(--text-muted)" />
          <span class="release-name">{updateInfo?.releaseName || ('EasyWinGet v' + (updateInfo?.latestVersion || ''))}</span>
        </div>
        {#if updateInfo?.publishedAt}
          <span class="release-date">{formatDate(updateInfo.publishedAt)}</span>
        {/if}
      </div>

      <div class="notes-scroll-container">
        {#if updateInfo?.releaseNotes && updateInfo.releaseNotes.trim().length > 0}
          <div class="notes-content">{updateInfo.releaseNotes}</div>
        {:else}
          <div class="notes-placeholder">
            <ShieldCheck size={18} color="var(--color-success-light, #6ccb5f)" />
            <span>This update includes the latest performance improvements, security enhancements, and stability fixes.</span>
          </div>
        {/if}
      </div>
    </div>

    <!-- Modal Action Buttons -->
    <div class="modal-footer">
      <button class="btn-remind" on:click={onRemindLater} title="Remind me again in 24 hours">
        <Clock size={16} />
        <span>Remind Me Later</span>
      </button>

      <button class="btn-update-now" on:click={onUpdateNow} title="Download and install update">
        <Download size={16} />
        <span>Update Now</span>
      </button>
    </div>

  </div>
</div>

<style>
  .update-modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.65);
    backdrop-filter: blur(14px);
    -webkit-backdrop-filter: blur(14px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 99999;
    padding: 18px;
    animation: fadeIn 0.18s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .update-modal-dialog {
    background: var(--bg-surface, #262626);
    border: 1px solid var(--border-card-hover, rgba(255, 255, 255, 0.16));
    box-shadow: 0 24px 60px rgba(0, 0, 0, 0.6), 0 4px 16px rgba(0, 0, 0, 0.4);
    border-radius: var(--radius-md, 12px);
    width: 100%;
    max-width: 520px;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    position: relative;
    animation: modalSlideUp 0.22s cubic-bezier(0.16, 1, 0.3, 1);
  }

  @keyframes modalSlideUp {
    from {
      opacity: 0;
      transform: translateY(14px) scale(0.97);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  .modal-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    padding: 22px 24px 16px 24px;
    gap: 16px;
    border-bottom: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.06));
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 14px;
    flex: 1;
  }

  .glow-icon-box {
    width: 46px;
    height: 46px;
    border-radius: 12px;
    background: linear-gradient(135deg, rgba(0, 120, 212, 0.25), rgba(96, 205, 255, 0.15));
    border: 1px solid var(--accent-badge-border, rgba(96, 205, 255, 0.3));
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    box-shadow: 0 0 20px rgba(96, 205, 255, 0.18);
  }

  .header-text {
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  .modal-title-row {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .modal-title {
    font-size: 1.15rem;
    font-weight: 600;
    color: var(--text-primary, #ffffff);
    margin: 0;
    letter-spacing: -0.01em;
  }

  .badge-new {
    background: var(--accent-badge-bg, rgba(96, 205, 255, 0.16));
    border: 1px solid var(--accent-badge-border, rgba(96, 205, 255, 0.35));
    color: var(--accent-light, #60cdff);
    font-size: 0.72rem;
    font-weight: 600;
    padding: 2px 7px;
    border-radius: var(--radius-pill, 999px);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .modal-subtitle {
    font-size: 0.84rem;
    color: var(--text-muted, #b4b4b4);
    margin: 0;
  }

  .btn-close {
    background: transparent;
    border: none;
    color: var(--text-muted, #b4b4b4);
    cursor: pointer;
    padding: 6px;
    border-radius: var(--radius-xs, 6px);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s ease;
  }

  .btn-close:hover {
    color: var(--text-primary, #ffffff);
    background: rgba(255, 255, 255, 0.1);
  }

  .version-banner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin: 18px 24px 14px 24px;
    padding: 12px 18px;
    background: var(--bg-card, #2b2b2b);
    border: 1px solid var(--border-card, rgba(255, 255, 255, 0.08));
    border-radius: var(--radius-sm, 8px);
  }

  .version-pill {
    display: flex;
    flex-direction: column;
    gap: 3px;
  }

  .version-label {
    font-size: 0.72rem;
    font-weight: 500;
    color: var(--text-muted, #b4b4b4);
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }

  .version-value {
    font-size: 0.98rem;
    font-weight: 600;
    font-family: var(--font-mono, monospace);
    color: var(--text-secondary, #e0e0e0);
  }

  .version-value.highlight {
    color: var(--accent-light, #60cdff);
  }

  .version-arrow {
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0.9;
  }

  .release-info-section {
    padding: 0 24px 18px 24px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .release-meta-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 0.8rem;
    color: var(--text-muted, #b4b4b4);
  }

  .release-title-group {
    display: flex;
    align-items: center;
    gap: 6px;
    font-weight: 500;
    color: var(--text-secondary, #e0e0e0);
  }

  .release-date {
    font-size: 0.76rem;
    opacity: 0.8;
  }

  .notes-scroll-container {
    background: var(--bg-input, #202020);
    border: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.06));
    border-radius: var(--radius-sm, 8px);
    max-height: 160px;
    min-height: 80px;
    overflow-y: auto;
    padding: 12px 14px;
    font-size: 0.82rem;
    line-height: 1.5;
  }

  .notes-content {
    white-space: pre-wrap;
    word-break: break-word;
    color: var(--text-secondary, #e0e0e0);
    font-family: var(--font-main, sans-serif);
  }

  .notes-placeholder {
    display: flex;
    align-items: center;
    gap: 10px;
    color: var(--text-muted, #b4b4b4);
    height: 100%;
    padding: 8px 0;
  }

  .modal-footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 12px;
    padding: 16px 24px;
    background: var(--bg-card, #202020);
    border-top: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.06));
  }

  .btn-remind {
    background: transparent;
    border: 1px solid var(--border-card-hover, rgba(255, 255, 255, 0.16));
    color: var(--text-secondary, #e0e0e0);
    padding: 8px 16px;
    border-radius: var(--radius-sm, 8px);
    font-size: 0.86rem;
    font-weight: 500;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    transition: all 0.15s ease;
  }

  .btn-remind:hover {
    background: rgba(255, 255, 255, 0.07);
    color: var(--text-primary, #ffffff);
    border-color: rgba(255, 255, 255, 0.28);
  }

  .btn-update-now {
    background: var(--accent-primary, #0078d4);
    border: 1px solid var(--accent-hover, #1084d8);
    color: var(--accent-contrast-text, #ffffff);
    padding: 8px 18px;
    border-radius: var(--radius-sm, 8px);
    font-size: 0.86rem;
    font-weight: 600;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    box-shadow: 0 4px 14px rgba(0, 120, 212, 0.35);
    transition: all 0.18s ease;
  }

  .btn-update-now:hover {
    background: var(--accent-hover, #1084d8);
    box-shadow: 0 6px 18px rgba(0, 120, 212, 0.48);
    transform: translateY(-1px);
  }

  .btn-update-now:active {
    transform: translateY(0);
    background: var(--accent-active, #005a9e);
  }
</style>
