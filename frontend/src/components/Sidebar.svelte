<script>
  import { 
    Search, 
    HardDrive, 
    ArrowUpCircle, 
    Download, 
    Terminal, 
    Settings, 
    Package,
    Info
  } from '@lucide/svelte';


  export let activeTab = 'search';
  export let installedCount = 0;
  export let updatesCount = 0;
  export let downloadsCount = 0;
  export let searchCount = null;
  export let isJobRunning = false;
  export let version = '5.0.0';
</script>

<aside class="sidebar">
  <div class="sidebar-header">
    <div class="brand">
      <div class="logo-box">
        <Package size={31} class="brand-icon-svg" />
      </div>
      <div class="brand-text">
        <div class="brand-title">EasyWinGet</div>
        <div class="brand-sub">Next-Gen Package Hub</div>
      </div>
    </div>
  </div>

  <nav class="sidebar-nav">
    <button 
      class="nav-item" 
      class:active={activeTab === 'search'} 
      on:click={() => activeTab = 'search'}
    >
      <div class="nav-indicator"></div>
      <Search size={19} class="nav-icon" />
      <span class="nav-label">Discover</span>
      {#if searchCount !== null && searchCount !== undefined}
        <span class="count-pill search-pill" class:zero-pill={searchCount === 0}>{searchCount}</span>
      {/if}
    </button>

    <button 
      class="nav-item" 
      class:active={activeTab === 'installed'} 
      on:click={() => activeTab = 'installed'}
    >
      <div class="nav-indicator"></div>
      <HardDrive size={19} class="nav-icon" />
      <span class="nav-label">Installed</span>
      {#if installedCount > 0}
        <span class="count-pill">{installedCount}</span>
      {/if}
    </button>

    <button 
      class="nav-item" 
      class:active={activeTab === 'updates'} 
      on:click={() => activeTab = 'updates'}
    >
      <div class="nav-indicator"></div>
      <ArrowUpCircle size={19} class="nav-icon" />
      <span class="nav-label">Updates</span>
      {#if updatesCount > 0}
        <span class="count-pill updates-pill">{updatesCount}</span>
      {/if}
    </button>

    <button 
      class="nav-item" 
      class:active={activeTab === 'downloads'} 
      on:click={() => activeTab = 'downloads'}
    >
      <div class="nav-indicator"></div>
      <Download size={19} class="nav-icon" />
      <span class="nav-label">Downloads</span>
      {#if downloadsCount > 0}
        <span class="count-pill">{downloadsCount}</span>
      {/if}
    </button>

    <div class="nav-divider"></div>

    <button 
      class="nav-item" 
      class:active={activeTab === 'terminal'} 
      on:click={() => activeTab = 'terminal'}
    >
      <div class="nav-indicator"></div>
      <Terminal size={19} class="nav-icon" />
      <span class="nav-label">Console</span>
      {#if isJobRunning}
        <span class="pulse-indicator"></span>
      {/if}
    </button>

    <button 
      class="nav-item" 
      class:active={activeTab === 'settings'} 
      on:click={() => activeTab = 'settings'}
    >
      <div class="nav-indicator"></div>
      <Settings size={19} class="nav-icon" />
      <span class="nav-label">Settings</span>
    </button>

    <button 
      class="nav-item" 
      class:active={activeTab === 'about'} 
      on:click={() => activeTab = 'about'}
    >
      <div class="nav-indicator"></div>
      <Info size={19} class="nav-icon" />
      <span class="nav-label">About</span>
    </button>
  </nav>


  <!-- Sidebar Footer: Centered Version Only -->
  <div class="sidebar-footer">
    <span class="version-centered">v{version}</span>
  </div>
</aside>

<style>
  .sidebar {
    width: 250px;
    height: 100vh;
    background: var(--bg-sidebar);
    border-right: 1px solid var(--border-subtle);
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
    user-select: none;
  }

  .sidebar-header {
    padding: 28px 18px 24px 18px;
    margin-bottom: 10px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  }

  :global([data-theme="light"]) .sidebar-header {
    border-bottom: 1px solid rgba(0, 0, 0, 0.05);
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 15px;
  }

  /* Logo Box with Dynamic Sheen Animation */
  .logo-box {
    width: 52px;
    height: 52px;
    border-radius: var(--radius-md);
    background: var(--bg-card);
    border: 1px solid var(--border-card);
    display: flex;
    align-items: center;
    justify-content: center;
    position: relative;
    overflow: hidden;
    box-shadow: var(--shadow-card);
    flex-shrink: 0;
  }

  .logo-box::after {
    content: '';
    position: absolute;
    top: -50%;
    left: -50%;
    width: 200%;
    height: 200%;
    background: linear-gradient(
      60deg,
      transparent 35%,
      rgba(255, 255, 255, 0.45) 50%,
      transparent 65%
    );
    transform: translateX(-150%) rotate(25deg);
    animation: iconSheen 8s ease-in-out infinite;
    pointer-events: none;
  }

  @keyframes iconSheen {
    0%, 82% {
      transform: translateX(-150%) rotate(25deg);
      opacity: 0;
    }
    85% {
      opacity: 0.85;
    }
    100% {
      transform: translateX(180%) rotate(25deg);
      opacity: 0;
    }
  }

  :global(.brand-icon-svg) {
    color: #ffffff; /* White in Dark Mode (overridden in style.css for light) */
    transition: color var(--transition-fast);
    z-index: 1;
  }

  .brand-title {
    font-size: 1.36rem;
    font-weight: 700;
    letter-spacing: -0.018em;
    color: var(--text-primary);
    line-height: 1.2;
  }

  .brand-sub {
    font-size: 0.86rem;
    color: var(--accent-light);
    font-weight: 500;
    margin-top: 3px;
  }

  .sidebar-nav {
    flex: 1;
    padding: 10px 8px;
    display: flex;
    flex-direction: column;
    gap: 2px;
    overflow-y: auto;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 13px;
    width: 100%;
    padding: 10px 14px;
    border-radius: var(--radius-sm);
    background: transparent;
    color: var(--text-secondary);
    font-size: 0.94rem;
    font-weight: 400;
    position: relative;
    border: 1px solid transparent;
    transition: background-color 0.18s cubic-bezier(0.1, 0.9, 0.2, 1), 
                color 0.18s ease, 
                border-color 0.18s ease;
  }

  .nav-item:hover {
    color: var(--text-primary);
    background: rgba(255, 255, 255, 0.05);
  }

  .nav-item.active {
    color: var(--text-primary);
    background: var(--accent-badge-bg, rgba(255, 255, 255, 0.08));
    border: 1px solid var(--accent-badge-border, rgba(255, 255, 255, 0.06));
    font-weight: 600;
  }

  .nav-indicator {
    position: absolute;
    left: 3px;
    top: 50%;
    width: 3.5px;
    height: 18px;
    border-radius: var(--radius-pill);
    background: var(--accent-light);
    transform: translateY(-50%) scaleY(0.2);
    opacity: 0;
    transition: transform 0.22s cubic-bezier(0.1, 0.9, 0.2, 1), opacity 0.18s ease;
  }

  .nav-item.active .nav-indicator {
    transform: translateY(-50%) scaleY(1);
    opacity: 1;
  }

  :global(.nav-icon) {
    color: var(--text-secondary);
    transition: color 0.18s ease;
  }
  .nav-item:hover :global(.nav-icon) {
    color: var(--text-primary);
  }
  .nav-item.active :global(.nav-icon) {
    color: var(--accent-light);
  }

  .nav-label {
    flex: 1;
    text-align: left;
  }

  .count-pill {
    font-size: 0.76rem;
    padding: 2px 7px;
    border-radius: var(--radius-pill);
    background: rgba(255, 255, 255, 0.08);
    color: var(--text-secondary);
    font-weight: 600;
  }

  .updates-pill {
    background: var(--color-warning-bg);
    color: var(--color-warning-text);
    border: 1px solid rgba(252, 225, 0, 0.3);
  }

  .search-pill {
    background: var(--accent-badge-bg, rgba(96, 205, 255, 0.14));
    color: var(--accent-light, #60cdff);
    border: 1px solid var(--accent-badge-border, rgba(96, 205, 255, 0.28));
    animation: searchPillPop 0.22s cubic-bezier(0.16, 1, 0.3, 1) both;
  }

  .search-pill.zero-pill {
    background: rgba(255, 255, 255, 0.05);
    color: var(--text-muted, #888888);
    border: 1px solid var(--border-subtle, rgba(255, 255, 255, 0.08));
  }

  :global([data-theme="light"]) .search-pill {
    background: var(--accent-badge-bg, rgba(0, 120, 212, 0.1));
    color: var(--accent-primary, #0078d4);
    border-color: var(--accent-badge-border, rgba(0, 120, 212, 0.25));
  }

  :global([data-theme="light"]) .search-pill.zero-pill {
    background: rgba(0, 0, 0, 0.05);
    color: rgba(0, 0, 0, 0.45);
    border-color: rgba(0, 0, 0, 0.08);
  }

  @keyframes searchPillPop {
    0% {
      transform: scale(0.6);
      opacity: 0;
    }
    100% {
      transform: scale(1);
      opacity: 1;
    }
  }

  .pulse-indicator {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--color-success-light);
    box-shadow: 0 0 6px var(--color-success-light);
    animation: pulseGlow 1s infinite alternate;
  }

  .nav-divider {
    height: 1px;
    background: var(--border-subtle);
    margin: 8px 6px;
  }

  /* Centered Version Only in Sidebar Footer */
  .sidebar-footer {
    padding: 14px 16px;
    border-top: 1px solid var(--border-subtle);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .version-centered {
    font-size: 0.78rem;
    color: var(--text-muted);
    font-family: var(--font-mono);
    letter-spacing: 0.05em;
    opacity: 0.85;
  }
</style>
