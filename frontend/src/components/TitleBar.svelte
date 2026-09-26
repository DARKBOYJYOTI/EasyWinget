<script>
  import { onMount, onDestroy } from 'svelte';
  import { 
    WindowMinimise, 
    WindowToggleMaximise, 
    WindowHide,
    WindowIsMaximised,
    WindowIsFullscreen,
    Quit
  } from '../../wailsjs/runtime/runtime.js';
  import { GetSetting } from '../../wailsjs/go/main/App.js';

  let isMax = false;
  let minimizeToTray = true;

  async function checkMaximised() {
    try {
      const maximized = await WindowIsMaximised();
      let fullscreen = false;
      try {
        fullscreen = await WindowIsFullscreen();
      } catch (e) {}

      const isMaxOrFull = Boolean(maximized || fullscreen);
      isMax = isMaxOrFull;

      // When maximized or fullscreen, disable Wails synthetic edge-resize calculations
      // so the resize arrow cursor does not appear at window borders
      if (typeof window !== 'undefined' && window.wails && window.wails.flags) {
        if (isMaxOrFull) {
          window.wails.flags.enableResize = false;
          if (document.documentElement.style.cursor && document.documentElement.style.cursor.includes('resize')) {
            document.documentElement.style.cursor = '';
          }
        } else {
          window.wails.flags.enableResize = true;
        }
      }

      if (typeof document !== 'undefined') {
        document.documentElement.classList.toggle('is-window-maximized', isMaxOrFull);
      }
    } catch (e) {}
  }

  function handleMinimise() {
    WindowMinimise();
  }

  async function handleToggleMaximise() {
    WindowToggleMaximise();
    setTimeout(checkMaximised, 60);
    setTimeout(checkMaximised, 160);
    setTimeout(checkMaximised, 320);
  }

  function handleClose() {
    if (minimizeToTray) {
      WindowHide();
    } else {
      Quit();
    }
  }

  onMount(async () => {
    checkMaximised();
    window.addEventListener('resize', checkMaximised);

    try {
      const val = await GetSetting('easywinget_minimize_to_tray', 'true');
      minimizeToTray = val !== 'false';
    } catch (e) {
      minimizeToTray = true;
    }

    const onTrayChange = (e) => {
      minimizeToTray = Boolean(e.detail);
    };
    window.addEventListener('easywinget:minimize-tray-changed', onTrayChange);

    const interval = setInterval(checkMaximised, 600);

    const handleMouseMove = () => {
      if (isMax) {
        if (window.wails && window.wails.flags && window.wails.flags.enableResize) {
          window.wails.flags.enableResize = false;
        }
        if (document.documentElement.style.cursor && document.documentElement.style.cursor.includes('resize')) {
          document.documentElement.style.cursor = '';
        }
      }
    };
    window.addEventListener('mousemove', handleMouseMove, { passive: true });

    return () => {
      window.removeEventListener('resize', checkMaximised);
      window.removeEventListener('mousemove', handleMouseMove);
      clearInterval(interval);
    };
  });
</script>

<div class="titlebar-drag-region" style="--wails-draggable: drag">
  <div class="window-controls" style="--wails-draggable: none">
    <button class="win-btn win-minimize" title="Minimize" on:click={handleMinimise}>
      <svg width="10" height="1" viewBox="0 0 10 1">
        <rect width="10" height="1" fill="currentColor" />
      </svg>
    </button>
    <button class="win-btn win-maximize" title={isMax ? "Restore" : "Maximize"} on:click={handleToggleMaximise}>
      {#if isMax}
        <svg width="10" height="10" viewBox="0 0 10 10">
          <path fill="none" stroke="currentColor" stroke-width="1" d="M2.5 1.5h6v6m-7-4h6v6h-6z" />
        </svg>
      {:else}
        <svg width="10" height="10" viewBox="0 0 10 10">
          <rect width="9" height="9" x="0.5" y="0.5" fill="none" stroke="currentColor" stroke-width="1" />
        </svg>
      {/if}
    </button>
    <button class="win-btn win-close" title="Close" on:click={handleClose}>
      <svg width="10" height="10" viewBox="0 0 10 10">
        <path fill="currentColor" d="M0.7 0L0 0.7 4.3 5 0 9.3l0.7 0.7L5 5.7 9.3 10l0.7-0.7L5.7 5 10 0.7 9.3 0 5 4.3z" />
      </svg>
    </button>
  </div>
</div>

<style>
  .titlebar-drag-region {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 38px;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    z-index: 999;
    user-select: none;
  }

  .window-controls {
    display: flex;
    height: 100%;
    align-items: center;
  }

  .win-btn {
    width: 46px;
    height: 100%;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: background-color 0.15s ease, color 0.15s ease;
    padding: 0;
  }

  .win-btn:hover {
    background: rgba(255, 255, 255, 0.08);
    color: var(--text-primary);
  }

  :global([data-theme="light"]) .win-btn:hover {
    background: rgba(0, 0, 0, 0.06);
    color: #000000;
  }

  .win-close:hover {
    background: #c42b1c !important;
    color: #ffffff !important;
  }

  .win-close:active {
    background: #b22718 !important;
  }
</style>
