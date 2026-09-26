<script>
  import { onMount, onDestroy } from 'svelte';
  import { Search, Download, Info, Plus, ChevronDown, ChevronUp, Sparkles, Globe, Check, X, Trash2 } from '@lucide/svelte';
  import AppIcon from './AppIcon.svelte';
  import SkeletonLoader from './SkeletonLoader.svelte';
  import { SearchApps, InstallPackage, DownloadPackage, UninstallPackage } from '../../wailsjs/go/main/App.js';

  export let onOpenDetails = (id) => {};
  export let onStartJob = (jobId, title) => {};
  export let onShowToast = (type, title, msg) => {};
  export let onConfirmRequest = async (opts) => true;
  export let onUpdateCount = (count) => {};
  export let installedIds = new Set();

  function isInstalled(pkg) {
    if (!pkg || !installedIds) return false;
    const idKey = (pkg.id || '').toLowerCase();
    const nameKey = (pkg.name || '').toLowerCase();
    return installedIds.has(idKey) || installedIds.has(nameKey);
  }

  let searchQuery = '';
  let searchResults = [];
  let isSearching = false;
  let hasSearched = false;
  let debounceTimer;
  let showMoreCategories = false;
  let activePopularTab = 'all';
  let autoCycleInterval;

  const primaryCategories = [
    { label: 'Development', term: 'development' },
    { label: 'Browsers', term: 'browser' },
    { label: 'Media', term: 'media' },
    { label: 'Utilities', term: 'utilities' },
    { label: 'Gaming', term: 'gaming' },
    { label: 'Communication', term: 'communication' }
  ];

  const extraCategories = [
    { label: 'Productivity', term: 'productivity' },
    { label: 'Security', term: 'security' },
    { label: 'Graphics & Design', term: 'design' },
    { label: 'System Tools', term: 'tools' },
    { label: 'Cloud & DevOps', term: 'cloud' },
    { label: 'Education', term: 'education' }
  ];

  // Comprehensive Catalog of World's Most Popular Windows Software
  const allPopularPackages = [
    // World Top Essentials
    { name: 'Google Chrome', id: 'Google.Chrome', category: 'Browsers', tab: 'browsers', badge: 'Top #1' },
    { name: 'Visual Studio Code', id: 'Microsoft.VisualStudioCode', category: 'Dev', tab: 'dev', badge: 'Popular' },
    { name: 'Discord', id: 'Discord.Discord', category: 'Social', tab: 'social', badge: 'Popular' },
    { name: 'Spotify', id: 'Spotify.Spotify', category: 'Media', tab: 'media', badge: 'Popular' },
    { name: 'VLC Media Player', id: 'VideoLAN.VLC', category: 'Media', tab: 'media', badge: 'Essential' },
    { name: '7-Zip', id: '7zip.7zip', category: 'Utilities', tab: 'utilities', badge: 'Essential' },
    { name: 'Git', id: 'Git.Git', category: 'Dev', tab: 'dev', badge: 'Essential' },
    { name: 'Mozilla Firefox', id: 'Mozilla.Firefox', category: 'Browsers', tab: 'browsers', badge: 'Top #2' },
    
    // Developer & Tools
    { name: 'Node.js LTS', id: 'OpenJS.NodeJS.LTS', category: 'Dev', tab: 'dev', badge: 'Runtime' },
    { name: 'Python 3', id: 'Python.Python.3.12', category: 'Dev', tab: 'dev', badge: 'Runtime' },
    { name: 'Docker Desktop', id: 'Docker.DockerDesktop', category: 'Dev', tab: 'dev', badge: 'DevOps' },
    { name: 'Postman', id: 'Postman.Postman', category: 'Dev', tab: 'dev', badge: 'API' },
    
    // Utilities & System
    { name: 'Microsoft PowerToys', id: 'Microsoft.PowerToys', category: 'Utilities', tab: 'utilities', badge: 'System' },
    { name: 'Notepad++', id: 'Notepad++.Notepad++', category: 'Utilities', tab: 'utilities', badge: 'Editor' },
    { name: 'OBS Studio', id: 'OBSProject.OBSStudio', category: 'Media', tab: 'media', badge: 'Creator' },
    { name: 'Steam', id: 'Valve.Steam', category: 'Gaming', tab: 'gaming', badge: 'Gaming' },
    
    // Communication & Productivity
    { name: 'WhatsApp', id: 'WhatsApp.WhatsApp', category: 'Social', tab: 'social', badge: 'Chat' },
    { name: 'Telegram Desktop', id: 'Telegram.TelegramDesktop', category: 'Social', tab: 'social', badge: 'Chat' },
    { name: 'Brave Browser', id: 'Brave.Brave', category: 'Browsers', tab: 'browsers', badge: 'Privacy' },
    { name: 'Bitwarden', id: 'Bitwarden.Bitwarden', category: 'Security', tab: 'utilities', badge: 'Security' }
  ];

  const popularTabs = [
    { id: 'all', label: 'All World Popular' },
    { id: 'browsers', label: 'Browsers' },
    { id: 'dev', label: 'Development' },
    { id: 'media', label: 'Media & Audio' },
    { id: 'social', label: 'Social & Chat' },
    { id: 'utilities', label: 'Utilities' }
  ];

  $: displayedPopular = allPopularPackages.filter(p => {
    if (activePopularTab === 'all') return true;
    return p.tab === activePopularTab;
  });

  let isHoveringFeatured = false;
  let isWindowVisible = true;

  function resetAutoCycle() {
    if (autoCycleInterval) clearInterval(autoCycleInterval);
    autoCycleInterval = setInterval(() => {
      // Pause completely if window is hidden, minimized to tray, or blurred in background
      if (document.hidden || !isWindowVisible || searchQuery.trim() !== '' || isSearching || isHoveringFeatured) {
        return;
      }
      const currentIndex = popularTabs.findIndex(t => t.id === activePopularTab);
      const nextIndex = (currentIndex + 1) % popularTabs.length;
      activePopularTab = popularTabs[nextIndex].id;
    }, 9500);
  }

  function handleSelectPopularTab(tabId) {
    activePopularTab = tabId;
    resetAutoCycle();
  }

  function onVisibilityChange() {
    isWindowVisible = !document.hidden && document.visibilityState === 'visible';
  }

  onMount(() => {
    resetAutoCycle();
    if (typeof document !== 'undefined') {
      document.addEventListener('visibilitychange', onVisibilityChange);
      window.addEventListener('focus', () => { isWindowVisible = true; });
      window.addEventListener('blur', () => {
        if (document.hidden) isWindowVisible = false;
      });
    }
  });

  onDestroy(() => {
    if (autoCycleInterval) clearInterval(autoCycleInterval);
    if (debounceTimer) clearTimeout(debounceTimer);
    if (typeof document !== 'undefined') {
      document.removeEventListener('visibilitychange', onVisibilityChange);
    }
  });


  function handleInput() {
    clearTimeout(debounceTimer);
    if (!searchQuery.trim()) {
      searchResults = [];
      hasSearched = false;
      isSearching = false;
      onUpdateCount(null);
      return;
    }
    debounceTimer = setTimeout(() => {
      executeSearch();
    }, 350);
  }

  function handleClearSearch() {
    searchQuery = '';
    searchResults = [];
    hasSearched = false;
    isSearching = false;
    onUpdateCount(null);
  }

  async function executeSearch() {
    const q = searchQuery.trim();
    if (q.length < 2) {
      if (q.length === 0) onUpdateCount(null);
      return;
    }

    isSearching = true;
    hasSearched = true;

    try {
      const results = await SearchApps(q);
      searchResults = results || [];
      onUpdateCount(searchResults.length);
    } catch (err) {
      onShowToast('error', 'Search Failed', String(err));
      searchResults = [];
      onUpdateCount(0);
    } finally {
      isSearching = false;
    }
  }

  function handleCategoryClick(term) {
    searchQuery = term;
    executeSearch();
  }

  async function handleInstall(pkg, interactive = false) {
    try {
      const jobId = await InstallPackage(pkg.id, interactive, false, '');
      onStartJob(jobId, `Installing ${pkg.name}`);
      onShowToast('info', 'Installation Started', `Installing ${pkg.name}...`);
    } catch (err) {
      onShowToast('error', 'Install Error', String(err));
    }
  }

  async function handleUninstall(pkg) {
    const ok = await onConfirmRequest({
      title: `Uninstall ${pkg.name}?`,
      message: `Are you sure you want to uninstall "${pkg.name}" from your computer?`,
      confirmText: 'Uninstall',
      cancelText: 'Cancel',
      isDanger: true,
      iconType: 'danger'
    });
    if (!ok) return;

    try {
      const jobId = await UninstallPackage(pkg.id, pkg.name, false);
      onStartJob(jobId, `Uninstalling ${pkg.name}`);
      onShowToast('info', 'Uninstall Started', `Uninstalling ${pkg.name}...`);
    } catch (err) {
      onShowToast('error', 'Uninstall Failed', String(err));
    }
  }

  async function handleDownload(pkg) {
    try {
      const jobId = await DownloadPackage(pkg.id, pkg.name);
      onStartJob(jobId, `Downloading ${pkg.name}`);
      onShowToast('info', 'Download Started', `Downloading installer for ${pkg.name}...`);
    } catch (err) {
      onShowToast('error', 'Download Error', String(err));
    }
  }
</script>

<div class="search-page view-animated">
  <!-- Sticky Top Zone: Title + Spaced Search Box + Category Chips -->

  <div class="search-header-zone">
    <div class="search-header">
      <div class="header-text">
        <h2>Discover & Install</h2>
        <p>Search thousands of applications from the official Windows Package Manager repository</p>
      </div>

      <div class="search-box-wrap">
        <div class="search-box">
          <Search size={18} class="search-icon" />
          <input 
            type="text" 
            placeholder="Search apps by name, ID or command (e.g. Chrome, Git, VSCode)..."
            bind:value={searchQuery}
            on:input={handleInput}
            on:keydown={(e) => e.key === 'Enter' && executeSearch()}
          />
          {#if searchQuery}
            <button class="clear-btn" on:click={handleClearSearch} title="Clear search" aria-label="Clear search">
              <X size={15} />
            </button>
          {/if}
        </div>

        <div class="category-chips">
          {#each primaryCategories as cat}
            <button class="chip" on:click={() => handleCategoryClick(cat.term)}>{cat.label}</button>
          {/each}

          {#if showMoreCategories}
            {#each extraCategories as cat}
              <button class="chip chip-extra" on:click={() => handleCategoryClick(cat.term)}>{cat.label}</button>
            {/each}
          {/if}

          <button 
            class="chip chip-more" 
            on:click={() => showMoreCategories = !showMoreCategories}
            title="Show more categories"
          >
            {#if showMoreCategories}
              <span>Less</span> <ChevronUp size={12} />
            {:else}
              <span>+ More</span> <ChevronDown size={12} />
            {/if}
          </button>
        </div>
      </div>
    </div>
  </div>

  <div class="search-body">
    {#if isSearching}
      <div class="loading-state">
        <p class="loading-label">Searching WinGet repository...</p>
        <SkeletonLoader type="card" count={8} />
      </div>
    {:else if hasSearched && searchResults.length === 0}
      <div class="state-container">
        <p class="state-title">No packages found for "{searchQuery}"</p>
        <p class="state-sub">Try searching with a shorter keyword or generic software name</p>
      </div>
    {:else if searchResults.length > 0}
      <div class="results-meta">
        <span>Found {searchResults.length} packages for "{searchQuery}"</span>
      </div>
      <div class="results-grid">
        {#each searchResults as pkg (pkg.id)}
          <div class="app-card">
            <div class="app-card-header">
              <AppIcon packageId={pkg.id} name={pkg.name} size={46} />

              <div class="app-info">
                <div class="app-name" title={pkg.name}>{pkg.name}</div>
                <div class="app-id" title={pkg.id}>{pkg.id}</div>
              </div>
              {#if pkg.version}
                <span class="badge badge-blue">{pkg.version}</span>
              {/if}
            </div>

            <div class="app-card-actions">
              {#if isInstalled(pkg)}
                <button class="btn-installed flex-1" title="Already installed - Click to view info" on:click={() => onOpenDetails(pkg.id)}>
                  <Check size={14} class="installed-check" />
                  <span>Installed</span>
                </button>
              {:else}
                <button class="btn-primary btn-install flex-1" on:click={() => handleInstall(pkg)}>
                  <span>Install</span>
                </button>
              {/if}
              <button class="btn-secondary btn-action" title="Download Installer" on:click={() => handleDownload(pkg)}>
                <Download size={14} />
              </button>
              <button class="btn-secondary btn-action" title="View Manifest Details" on:click={() => onOpenDetails(pkg.id)}>
                <Info size={14} />
              </button>
            </div>
          </div>
        {/each}
      </div>
    {:else}
      <!-- Featured / Popular Apps with Category Filter & Auto-Cycle -->
      <div 
        class="featured-section"
        on:mouseenter={() => isHoveringFeatured = true}
        on:mouseleave={() => isHoveringFeatured = false}
        role="region"
        aria-label="Popular Software Categories"
      >
        <div class="section-header">
          <div class="section-title">
            <Globe size={18} color="var(--accent-light)" />
            <span>World's Most Popular Software</span>
          </div>

          <div class="popular-tabs">
            {#each popularTabs as tab}
              <button 
                class="tab-btn {activePopularTab === tab.id ? 'active' : ''}"
                on:click={() => handleSelectPopularTab(tab.id)}
              >
                <span>{tab.label}</span>
              </button>

            {/each}
          </div>
        </div>

        <div class="popular-grid-wrapper">
          {#key activePopularTab}
            <div class="results-grid popular-results-grid">
              {#each displayedPopular as pkg, i (pkg.id)}
                <div 
                  class="app-card popular-card-animated"
                  style="--card-delay: {i * 40}ms;"
                >
                  <div class="app-card-header">
                    <AppIcon packageId={pkg.id} name={pkg.name} size={46} />

                    <div class="app-info">
                      <div class="app-name" title={pkg.name}>{pkg.name}</div>
                      <div class="app-id" title={pkg.id}>{pkg.id}</div>
                    </div>
                    {#if pkg.badge}
                      <span class="badge badge-blue">{pkg.badge}</span>
                    {/if}
                  </div>

                  <div class="app-card-actions">
                    {#if isInstalled(pkg)}
                      <button class="btn-installed flex-1" title="Already installed - Click to view info" on:click={() => onOpenDetails(pkg.id)}>
                        <Check size={14} class="installed-check" />
                        <span>Installed</span>
                      </button>
                    {:else}
                      <button class="btn-primary btn-install flex-1" on:click={() => handleInstall(pkg)}>
                        <span>Install</span>
                      </button>
                    {/if}
                    <button class="btn-secondary btn-action" title="Download Installer" on:click={() => handleDownload(pkg)}>
                      <Download size={14} />
                    </button>
                    <button class="btn-secondary btn-action" title="View Details" on:click={() => onOpenDetails(pkg.id)}>
                      <Info size={14} />
                    </button>
                  </div>
                </div>
              {/each}
            </div>
          {/key}
        </div>
      </div>
    {/if}

  </div>
</div>

<style>
  .search-page {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
    padding: 0;
  }

  .search-header-zone {
    background: var(--bg-app);
    border-bottom: 1px solid var(--border-subtle);


    padding: 26px 36px 20px 36px;
    flex-shrink: 0;
    z-index: 10;
    backdrop-filter: blur(10px);
  }

  .search-header {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    margin-bottom: 0;
  }

  .header-text h2 {
    font-size: 2.45rem;
    font-weight: 800;
    letter-spacing: -0.025em;
    line-height: 1.2;
    color: var(--text-primary);
  }

  .header-text p {
    color: var(--text-secondary);
    font-size: 1.02rem;
    line-height: 1.45;
    margin-top: 6px;
    margin-bottom: 24px;
  }

  .search-box-wrap {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    width: 100%;
    max-width: 680px;
  }

  .search-box {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 12px;
    background: var(--bg-card);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-pill);
    padding: 6px 18px;
    height: 44px;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.12);
    transition: background-color 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease;
  }

  .search-box:hover {
    border-color: var(--border-card-hover);
  }

  .search-box:focus-within {
    background: var(--bg-surface);
    border-color: var(--accent-light);
    box-shadow: 0 0 0 2px var(--accent-badge-border), 0 0 18px var(--accent-glow);
  }

  :global(.search-icon) {
    color: rgba(255, 255, 255, 0.95);
    flex-shrink: 0;
    transition: color 0.15s ease;
  }

  :global([data-theme="light"]) :global(.search-icon) {
    color: #1a1a1a;
  }

  .search-box input {
    flex: 1;
    background: transparent;
    border: none;
    box-shadow: none;
    font-size: 0.94rem;
    padding: 7px 0;
    color: #ffffff;
  }

  .search-box input::placeholder {
    color: rgba(255, 255, 255, 0.65);
    opacity: 1;
  }

  :global([data-theme="light"]) .search-box input {
    color: #111827;
  }

  :global([data-theme="light"]) .search-box input::placeholder {
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

  .category-chips {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    align-items: center;
    gap: 8px;
    margin-top: 2px;
    width: 100%;
  }

  .search-body {
    flex: 1;
    overflow-y: auto;
    padding: 24px 36px 36px 36px;
  }

  .chip {
    background: var(--bg-card);
    border: 1px solid var(--border-subtle);
    color: var(--text-secondary);
    padding: 5px 13px;
    border-radius: var(--radius-pill);
    font-size: 0.8rem;
    font-weight: 500;
    transition: background-color 0.15s ease, color 0.15s ease, border-color 0.15s ease, box-shadow 0.15s ease;
  }
  .chip:hover {
    background: var(--bg-card-hover);
    color: var(--text-primary);
    border-color: var(--accent-badge-border);
    box-shadow: 0 0 8px var(--accent-glow);
  }

  .chip-more {
    background: var(--accent-badge-bg, rgba(255, 255, 255, 0.05));
    border-color: var(--accent-badge-border, var(--border-card-hover));
    color: var(--accent-light);
    font-weight: 600;
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }
  .chip-more:hover {
    background: var(--accent-badge-border);
    color: var(--accent-light);
  }

  .search-body {
    flex: 1;
    overflow-y: auto;
    padding: 24px 36px 48px 36px;
  }

  .loading-state {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .loading-label {
    font-size: 0.85rem;
    color: var(--text-secondary);
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

  .state-title {
    font-size: 1.05rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .state-sub {
    font-size: 0.85rem;
    color: var(--text-muted);
  }

  .results-meta {
    font-size: 0.82rem;
    color: var(--text-muted);
    margin-bottom: 14px;
  }

  .featured-section {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .section-header {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    border-bottom: 1px solid var(--border-subtle);
    padding-bottom: 10px;
  }

  .section-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .popular-tabs {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .tab-btn {
    position: relative;
    background: transparent;
    border: 1px solid transparent;
    color: var(--text-muted);
    font-size: 0.8rem;
    font-weight: 500;
    padding: 5px 12px;
    border-radius: var(--radius-sm, 6px);
    cursor: pointer;
    transition: all 0.22s cubic-bezier(0.16, 1, 0.3, 1);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
  }

  .tab-btn:hover {
    color: var(--text-primary);
    background: rgba(255, 255, 255, 0.05);
  }

  .tab-btn.active {
    color: var(--accent-light, #60a5fa);
    background: var(--accent-badge-bg, rgba(96, 165, 250, 0.12));
    border-color: var(--accent-badge-border, rgba(96, 165, 250, 0.28));
    font-weight: 600;
  }

  .popular-grid-wrapper {
    position: relative;

    width: 100%;
    min-height: 280px;
  }

  .popular-results-grid {
    width: 100%;
  }

  .popular-card-animated {
    animation: popularCardSlideIn 0.42s cubic-bezier(0.16, 1, 0.3, 1) both;
    animation-delay: var(--card-delay, 0ms);
  }

  @keyframes popularCardSlideIn {
    0% {
      opacity: 0;
      transform: translateY(16px) scale(0.975);
      filter: blur(3px);
    }
    60% {
      filter: blur(0);
    }
    100% {
      opacity: 1;
      transform: translateY(0) scale(1);
      filter: blur(0);
    }
  }

  .results-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(310px, 1fr));
    gap: 16px;
  }


  .app-card {
    background: var(--bg-card);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-md);
    padding: 18px;
    display: flex;
    flex-direction: column;
    gap: 14px;
    transition: background-color 0.18s cubic-bezier(0.1, 0.9, 0.2, 1), 
                border-color 0.18s cubic-bezier(0.1, 0.9, 0.2, 1), 
                box-shadow 0.18s cubic-bezier(0.1, 0.9, 0.2, 1);
  }

  .app-card:hover {
    background: var(--bg-card-hover);
    border-color: var(--accent-badge-border);
    box-shadow: 0 0 0 1px var(--accent-badge-border), 0 4px 18px var(--accent-glow);
  }

  .app-card-header {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .app-info {
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

  .app-card-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  /* Pixel-perfect Install Button */
  .btn-install {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    height: 32px;
    padding: 0 14px;
    font-size: 0.85rem;
    font-weight: 500;
  }

  .btn-install span {
    line-height: 1;
  }

  /* Installed State Button - subtle theme / muted disabled style */
  .btn-installed {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    height: 32px;
    padding: 0 12px;
    font-size: 0.84rem;
    font-weight: 500;
    background: rgba(255, 255, 255, 0.04);
    color: var(--text-secondary);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-sm);
    cursor: pointer;
    user-select: none;
    transition: all var(--transition-fast);
  }

  .btn-installed :global(.installed-check) {
    color: var(--accent-light);
    opacity: 0.85;
    flex-shrink: 0;
  }

  .btn-installed:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: var(--accent-badge-border);
    color: var(--text-primary);
  }

  .btn-installed:hover :global(.installed-check) {
    opacity: 1;
  }

  :global([data-theme="light"]) .btn-installed {
    background: rgba(0, 0, 0, 0.04);
    color: #616161;
    border: 1px solid rgba(0, 0, 0, 0.09);
  }

  :global([data-theme="light"]) .btn-installed :global(.installed-check) {
    color: var(--accent-primary);
  }

  :global([data-theme="light"]) .btn-installed:hover {
    background: rgba(0, 0, 0, 0.07);
    border-color: var(--accent-badge-border);
    color: #1a1a1a;
  }

  .btn-action {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    padding: 0;
  }

  .flex-1 {
    flex: 1;
  }
</style>
