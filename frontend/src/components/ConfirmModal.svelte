<script>
  import { fade, scale } from 'svelte/transition';
  import { AlertTriangle, Trash2, HelpCircle, Info, X } from '@lucide/svelte';

  export let show = false;
  export let title = 'Confirm Action';
  export let message = 'Are you sure you want to proceed?';
  export let confirmText = 'Confirm';
  export let cancelText = 'Cancel';
  export let isDanger = false;
  export let iconType = 'warning'; // 'warning' | 'danger' | 'info' | 'question'
  export let onConfirm = () => {};
  export let onCancel = () => {};

  function handleKeydown(e) {
    if (!show) return;
    if (e.key === 'Escape') {
      handleCancel();
    } else if (e.key === 'Enter') {
      handleConfirm();
    }
  }

  function handleCancel() {
    show = false;
    onCancel();
  }

  function handleConfirm() {
    show = false;
    onConfirm();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if show}
  <!-- Backdrop -->
  <div 
    class="modal-backdrop"
    transition:fade={{ duration: 150 }}
    on:click={handleCancel}
  >
    <!-- Content Dialog Window -->
    <div 
      class="dialog-box" 
      role="dialog"
      aria-modal="true"
      transition:scale={{ duration: 180, start: 0.94 }}
      on:click|stopPropagation
    >
      <div class="dialog-header">
        <div class="dialog-icon-box" class:danger={isDanger}>
          {#if isDanger || iconType === 'danger'}
            <Trash2 size={20} class="icon-danger" />
          {:else if iconType === 'warning'}
            <AlertTriangle size={20} class="icon-warning" />
          {:else if iconType === 'info'}
            <Info size={20} class="icon-info" />
          {:else}
            <HelpCircle size={20} class="icon-primary" />
          {/if}
        </div>
        <div class="dialog-title-wrap">
          <h3 class="dialog-title">{title}</h3>
        </div>
        <button class="dialog-close" on:click={handleCancel} title="Close">
          <X size={16} />
        </button>
      </div>

      <div class="dialog-body">
        <p class="dialog-message">{message}</p>
      </div>

      <div class="dialog-actions">
        <button class="btn-secondary dialog-btn" on:click={handleCancel}>
          {cancelText}
        </button>
        <button 
          class="dialog-btn" 
          class:btn-danger={isDanger} 
          class:btn-primary={!isDanger}
          on:click={handleConfirm}
        >
          {confirmText}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.72);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 99999;
    padding: 20px;
    animation: modalBackdropFade 0.15s ease-out;
  }


  .dialog-box {
    width: 100%;
    max-width: 440px;
    background: var(--bg-surface);
    border: 1px solid var(--border-card-hover);
    border-radius: var(--radius-lg);
    box-shadow: 0 20px 50px rgba(0, 0, 0, 0.48), 0 0 0 1px rgba(255, 255, 255, 0.08);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .dialog-header {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 20px 22px 14px 22px;
  }

  .dialog-icon-box {
    width: 38px;
    height: 38px;
    border-radius: var(--radius-sm);
    background: var(--accent-badge-bg);
    border: 1px solid var(--accent-badge-border);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .dialog-icon-box.danger {
    background: var(--color-danger-bg);
    border-color: rgba(209, 52, 56, 0.35);
  }

  :global(.icon-danger) {
    color: var(--color-danger-light);
  }
  :global(.icon-warning) {
    color: var(--color-warning-text);
  }
  :global(.icon-info) {
    color: var(--accent-light);
  }
  :global(.icon-primary) {
    color: var(--accent-light);
  }

  .dialog-title-wrap {
    flex: 1;
  }

  .dialog-title {
    font-size: 1.15rem;
    font-weight: 600;
    color: var(--text-primary);
    margin: 0;
    letter-spacing: -0.01em;
  }

  .dialog-close {
    background: transparent;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 6px;
    border-radius: var(--radius-xs);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background-color var(--transition-fast), color var(--transition-fast);
  }
  .dialog-close:hover {
    background: rgba(255, 255, 255, 0.08);
    color: var(--text-primary);
  }

  .dialog-body {
    padding: 0 22px 20px 22px;
  }

  .dialog-message {
    color: var(--text-secondary);
    font-size: 0.93rem;
    line-height: 1.5;
    margin: 0;
  }

  .dialog-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    padding: 14px 22px;
    background: var(--bg-card);
    border-top: 1px solid var(--border-subtle);
  }

  .dialog-btn {
    min-width: 90px;
    padding: 8px 18px;
    font-size: 0.9rem;
    font-weight: 500;
  }

  .btn-danger {
    background: var(--color-danger);
    color: #ffffff;
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: var(--radius-sm);
    cursor: pointer;
    transition: background-color var(--transition-fast), transform var(--transition-fast);
  }
  .btn-danger:hover {
    background: #e81123;
  }
</style>
