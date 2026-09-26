<script>
  export let toasts = [];
  export let removeToast = (id) => {};

  let lastToastId = null;

  function playChime(type) {
    try {
      const AudioCtx = window.AudioContext || window.webkitAudioContext;
      if (!AudioCtx) return;
      const ctx = new AudioCtx();
      const now = ctx.currentTime;
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.type = 'sine';

      if (type === 'error') {
        osc.frequency.setValueAtTime(440, now);
        osc.frequency.exponentialRampToValueAtTime(311.13, now + 0.15);
      } else {
        osc.frequency.setValueAtTime(587.33, now);
        osc.frequency.setValueAtTime(880, now + 0.08);
      }

      gain.gain.setValueAtTime(0.001, now);
      gain.gain.exponentialRampToValueAtTime(0.14, now + 0.02);
      gain.gain.exponentialRampToValueAtTime(0.001, now + 0.3);

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.start(now);
      osc.stop(now + 0.32);
    } catch (e) {}
  }

  $: if (toasts && toasts.length > 0) {
    const current = toasts[0];
    if (current && current.id !== lastToastId) {
      playChime(current.type);
      lastToastId = current.id;
    }
  }
</script>

{#if toasts && toasts.length > 0}
  <div class="toast-center-wrap">
    {#each toasts as t (t.id)}
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div class="android-toast" on:click={() => removeToast(t.id)} title="Click to dismiss">
        <span class="toast-text">{t.message || t.title}</span>
      </div>
    {/each}
  </div>
{/if}

<style>
  .toast-center-wrap {
    position: fixed;
    bottom: 28px;
    left: calc(250px + (100vw - 250px) / 2);
    transform: translateX(-50%);
    display: flex;
    flex-direction: column;
    align-items: center;
    z-index: 99999;
    pointer-events: none;
    max-width: min(540px, calc(100vw - 290px));
  }

  /* Android Style Pill Toast without Icon */
  .android-toast {
    pointer-events: auto;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 9px 22px;
    border-radius: var(--radius-pill);
    background: rgba(34, 34, 36, 0.94);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    color: #ffffff;
    border: 1px solid rgba(255, 255, 255, 0.12);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.38), 0 2px 8px rgba(0, 0, 0, 0.2);
    animation: androidToastIn 0.24s cubic-bezier(0.16, 1, 0.3, 1) forwards;
    cursor: pointer;
    user-select: none;
    max-width: 100%;
    word-break: break-word;
    transition: transform 0.1s ease, opacity 0.15s ease;
  }

  .android-toast:active {
    transform: scale(0.96);
  }

  :global([data-theme="light"]) .android-toast {
    background: rgba(245, 245, 247, 0.96);
    color: #1c1c1e;
    border-color: rgba(0, 0, 0, 0.12);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.14), 0 2px 6px rgba(0, 0, 0, 0.08);
  }

  :global([data-theme="light"]) .toast-text {
    color: #1c1c1e;
  }

  .toast-text {
    font-size: 0.88rem;
    font-weight: 500;
    line-height: 1.35;
    letter-spacing: 0.01em;
    color: #ffffff;
  }

  @keyframes androidToastIn {
    0% {
      opacity: 0;
      transform: translateY(14px) scale(0.94);
    }
    100% {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
</style>
