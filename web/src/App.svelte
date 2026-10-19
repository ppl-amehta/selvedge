<script>
  import { onMount } from 'svelte';
  import { api } from './api.js';
  import OrderTakingView from './views/OrderTakingView.svelte';
  import InwardView from './views/InwardView.svelte';
  import GraphInspectorModal from './components/GraphInspectorModal.svelte';

  let activeStation = $state('orders'); // 'orders' | 'inward'
  let masters = $state({});
  let isGraphOpen = $state(false);
  let isOnline = $state(false);
  let toasts = $state([]);

  // Global Toast Function
  if (typeof window !== 'undefined') {
    window.showToast = (message, type = 'info') => {
      const id = Date.now() + Math.random();
      toasts = [...toasts, { id, message, type }];
      setTimeout(() => {
        toasts = toasts.filter(t => t.id !== id);
      }, 4000);
    };
  }

  onMount(async () => {
    try {
      const health = await api.health();
      isOnline = health.status === 'ok';

      const data = await api.getMasters();
      masters = data;
    } catch (err) {
      console.error('System boot error:', err);
    }
  });
</script>

<header class="topbar">
  <div style="display:flex;align-items:center;gap:14px;">
    <div class="brand-badge">⚡ SELVEDGE</div>
    <div>
      <div class="brand-title">FPCC Enterprise Operating System</div>
      <div class="brand-sub">Unified Graph System of Record • Bi-Temporal Engine</div>
    </div>
  </div>

  <nav class="station-nav">
    <button
      class="station-btn"
      class:active={activeStation === 'orders'}
      onclick={() => activeStation = 'orders'}
    >
      <span>Station 0: Order Taking</span>
    </button>
    <button
      class="station-btn"
      class:active={activeStation === 'inward'}
      onclick={() => activeStation = 'inward'}
    >
      <span>Station 1: Inward & QR Bundles</span>
    </button>
  </nav>

  <div style="display:flex;align-items:center;gap:14px;">
    <button class="btn btn-secondary btn-sm" onclick={() => isGraphOpen = true}>
      🕸️ Graph Explorer
    </button>
    <div class="system-badge">
      <span class="system-dot"></span>
      <span>{isOnline ? 'ONLINE (Go 1.27 SQLite WAL)' : 'CONNECTING...'}</span>
    </div>
  </div>
</header>

<main class="main-content">
  {#if activeStation === 'orders'}
    <OrderTakingView {masters} />
  {:else if activeStation === 'inward'}
    <InwardView {masters} />
  {/if}
</main>

<GraphInspectorModal bind:isOpen={isGraphOpen} />

<!-- Global Toast Container -->
<div class="toast-container">
  {#each toasts as t (t.id)}
    <div class="toast {t.type}">
      <span>{t.type === 'success' ? '✓' : t.type === 'error' ? '✕' : 'ℹ'}</span>
      <div>{t.message}</div>
    </div>
  {/each}
</div>
