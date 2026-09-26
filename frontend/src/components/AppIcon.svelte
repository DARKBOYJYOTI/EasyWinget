<script>
  import { onMount, onDestroy } from 'svelte';
  import { getDomainForPackage, getGoogleIconUrl, getDuckDuckGoIconUrl } from '../utils/icons.js';
  import { GetLocalAppIcon, GetManifestIcon } from '../../wailsjs/go/main/App.js';

  export let packageId = '';
  export let name = '';
  export let size = 44;
  export let homepage = '';

  // Shared in-memory cache across all AppIcon instances to prevent repeat lookups
  const staticCache = typeof window !== 'undefined' ? (window.__easywinget_icon_cache = window.__easywinget_icon_cache || new Map()) : new Map();

  let containerEl;
  let observer;
  let isVisible = false;
  let isLoading = true;
  let isLoaded = false;
  let currentSrc = '';
  let sources = [];
  let currentSourceIndex = 0;
  let attemptedManifest = false;

  $: cacheKey = (packageId || name || '').toLowerCase();

  $: if (isVisible && cacheKey) {
    initOrUpdate();
  }

  onMount(() => {
    // Check if icon is already cached in memory
    if (cacheKey && staticCache.has(cacheKey)) {
      currentSrc = staticCache.get(cacheKey);
      isLoading = false;
      isLoaded = true;
      isVisible = true;
      return;
    }

    if (typeof IntersectionObserver !== 'undefined' && containerEl) {
      observer = new IntersectionObserver((entries) => {
        if (entries[0] && entries[0].isIntersecting) {
          isVisible = true;
          if (observer) {
            observer.disconnect();
            observer = null;
          }
        }
      }, {
        rootMargin: '120px' // Preload slightly before entering viewport
      });
      observer.observe(containerEl);
    } else {
      isVisible = true;
    }
  });

  onDestroy(() => {
    if (observer) {
      observer.disconnect();
      observer = null;
    }
  });

  async function initOrUpdate() {
    if (staticCache.has(cacheKey)) {
      currentSrc = staticCache.get(cacheKey);
      isLoading = false;
      isLoaded = true;
      return;
    }

    isLoading = true;
    attemptedManifest = false;
    const list = [];

    // 1. Try local native registry icon (Go BCUninstaller / AppX / MSIX)
    if (packageId || name) {
      try {
        const local = await GetLocalAppIcon(packageId, name);
        if (local) {
          list.push(local);
        }
      } catch (e) {}
    }

    // 2. Resolve domain for web icon fetching
    const domain = getDomainForPackage(packageId, homepage, name);
    if (domain) {
      const gUrl = getGoogleIconUrl(domain);
      if (gUrl) list.push(gUrl);

      const ddgUrl = getDuckDuckGoIconUrl(domain);
      if (ddgUrl) list.push(ddgUrl);
    }

    sources = list;
    currentSourceIndex = 0;

    if (sources.length > 0) {
      currentSrc = sources[0];
      staticCache.set(cacheKey, currentSrc);
      isLoading = false;
    } else if (packageId && !attemptedManifest) {
      fetchManifestIcon(packageId, homepage);
    } else {
      isLoading = false;
    }
  }

  async function fetchManifestIcon(pkgId, home) {
    attemptedManifest = true;
    try {
      const res = await GetManifestIcon(pkgId, home || '');
      if (res && res.success && res.iconUrl) {
        sources = [res.iconUrl, ...sources];
        currentSrc = res.iconUrl;
        staticCache.set(cacheKey, currentSrc);
      } else if (res && res.domain) {
        const gUrl = getGoogleIconUrl(res.domain);
        if (gUrl) {
          sources = [gUrl, ...sources];
          currentSrc = gUrl;
          staticCache.set(cacheKey, currentSrc);
        }
      }
    } catch (e) {
    } finally {
      isLoading = false;
    }
  }

  function handleImgLoad() {
    isLoading = false;
    isLoaded = true;
    if (currentSrc) {
      staticCache.set(cacheKey, currentSrc);
    }
  }

  function handleImgError() {
    currentSourceIndex++;
    if (currentSourceIndex < sources.length) {
      currentSrc = sources[currentSourceIndex];
    } else if (!attemptedManifest && packageId) {
      currentSrc = '';
      fetchManifestIcon(packageId, homepage);
    } else {
      currentSrc = '';
      isLoading = false;
    }
  }

  function getFallbackColor(str) {
    if (!str) return '#0078d4';
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      hash = str.charCodeAt(i) + ((hash << 5) - hash);
    }
    const colors = [
      '#0078d4', // Windows Blue
      '#107c41', // Excel Green
      '#d83b01', // Office Orange
      '#5c2d91', // Teams Purple
      '#008272', // Teal
      '#e81123', // Red
      '#0063b1', // Cobalt
      '#881798'  // Magenta
    ];
    return colors[Math.abs(hash) % colors.length];
  }

  function getInitial(str) {
    if (!str) return 'W';
    const clean = str.replace(/[^a-zA-Z0-9]/g, ' ').trim();
    return clean ? clean.charAt(0).toUpperCase() : 'W';
  }
</script>

<div 
  bind:this={containerEl}
  class="icon-container" 
  style="width: {size}px; height: {size}px; min-width: {size}px; min-height: {size}px;"
>
  {#if isLoading && !currentSrc}
    <div class="icon-spinner-wrap">
      <div class="fluent-spinner"></div>
    </div>
  {/if}

  {#if currentSrc}
    <img 
      src={currentSrc} 
      alt={name || packageId}
      class="app-img icon-fade-in" 
      loading="lazy"
      on:load={handleImgLoad}
      on:error={handleImgError}
    />
  {:else if !isLoading}
    <div 
      class="fallback-squircle icon-fade-in"
      style="background: linear-gradient(135deg, {getFallbackColor(name || packageId)} 0%, rgba(20, 20, 20, 0.8) 140%);"
    >
      <span class="fallback-letter">{getInitial(name || packageId)}</span>
    </div>
  {/if}
</div>

<style>
  .icon-container {
    position: relative;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 9px;
    overflow: hidden;
    flex-shrink: 0;
    user-select: none;
    background: rgba(255, 255, 255, 0.04);
    box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.08);
  }

  .icon-spinner-wrap {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255, 255, 255, 0.02);
  }

  .fluent-spinner {
    width: 14px;
    height: 14px;
    border: 2px solid rgba(255, 255, 255, 0.15);
    border-top-color: var(--accent-light, #60cdff);
    border-radius: 50%;
    animation: spin 0.75s cubic-bezier(0.4, 0, 0.2, 1) infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .app-img {
    width: 100%;
    height: 100%;
    object-fit: contain;
    padding: 3px;
    border-radius: 8px;
    image-rendering: -webkit-optimize-contrast;
  }

  .icon-fade-in {
    animation: iconApper 0.26s cubic-bezier(0.1, 0.9, 0.2, 1) forwards;
  }

  @keyframes iconApper {
    from {
      opacity: 0;
      transform: scale(0.88);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }

  .fallback-squircle {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    box-shadow: inset 0 1px 1px rgba(255, 255, 255, 0.25);
  }

  .fallback-letter {
    font-size: 1.1rem;
    font-weight: 700;
    color: #ffffff;
    font-family: var(--font-main);
  }
</style>
