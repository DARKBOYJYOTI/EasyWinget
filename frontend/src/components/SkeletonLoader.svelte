<script>
  export let count = 6;
  export let type = 'card'; // 'card' or 'row'
</script>

{#if type === 'card'}
  <div class="skeleton-grid">
    {#each Array(count) as _}
      <div class="skeleton-card">
        <div class="skeleton-header">
          <div class="skeleton-shimmer sk-icon"></div>
          <div class="sk-info">
            <div class="skeleton-shimmer sk-title"></div>
            <div class="skeleton-shimmer sk-sub"></div>
          </div>
          <div class="skeleton-shimmer sk-badge"></div>
        </div>
        <div class="skeleton-actions">
          <div class="skeleton-shimmer sk-btn-main"></div>
          <div class="skeleton-shimmer sk-btn-icon"></div>
          <div class="skeleton-shimmer sk-btn-icon"></div>
        </div>
      </div>
    {/each}
  </div>
{:else}
  <div class="skeleton-list">
    {#each Array(count) as _}
      <div class="skeleton-row">
        <div class="skeleton-shimmer sk-row-icon"></div>
        <div class="sk-row-info">
          <div class="skeleton-shimmer sk-row-title"></div>
          <div class="skeleton-shimmer sk-row-sub"></div>
        </div>
        <div class="skeleton-row-actions">
          <div class="skeleton-shimmer sk-row-btn"></div>
          <div class="skeleton-shimmer sk-row-btn"></div>
        </div>
      </div>
    {/each}
  </div>
{/if}

<style>
  .skeleton-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(290px, 1fr));
    gap: 14px;
    width: 100%;
  }

  .skeleton-card {
    background: var(--bg-card);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-md);
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .skeleton-header {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .sk-icon {
    width: 42px;
    height: 42px;
    border-radius: var(--radius-sm);
    flex-shrink: 0;
  }

  .sk-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .sk-title {
    width: 65%;
    height: 14px;
  }

  .sk-sub {
    width: 45%;
    height: 10px;
  }

  .sk-badge {
    width: 45px;
    height: 18px;
    border-radius: var(--radius-pill);
  }

  .skeleton-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .sk-btn-main {
    flex: 1;
    height: 32px;
    border-radius: var(--radius-sm);
  }

  .sk-btn-icon {
    width: 32px;
    height: 32px;
    border-radius: var(--radius-sm);
  }

  /* List / Row Style */
  .skeleton-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
    width: 100%;
  }

  .skeleton-row {
    display: flex;
    align-items: center;
    gap: 14px;
    background: var(--bg-card);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-sm);
    padding: 10px 16px;
  }

  .sk-row-icon {
    width: 38px;
    height: 38px;
    border-radius: var(--radius-sm);
    flex-shrink: 0;
  }

  .sk-row-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .sk-row-title {
    width: 200px;
    max-width: 50%;
    height: 14px;
  }

  .sk-row-sub {
    width: 130px;
    max-width: 35%;
    height: 10px;
  }

  .skeleton-row-actions {
    display: flex;
    gap: 8px;
  }

  .sk-row-btn {
    width: 80px;
    height: 30px;
    border-radius: var(--radius-sm);
  }
</style>
