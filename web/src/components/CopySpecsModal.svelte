<script>
  import { api } from '../api.js';

  let {
    isOpen = $bindable(false),
    onCopySpecs = () => {}
  } = $props();

  let dialogEl = $state(null);
  let ordersList = $state([]);
  let searchQuery = $state('');
  let isLoading = $state(false);
  let isApplying = $state(false);
  let errorMsg = $state('');

  $effect(() => {
    if (!dialogEl) return;
    if (isOpen) {
      if (!dialogEl.open) {
        dialogEl.showModal();
      }
      loadOrders();
      searchQuery = '';
      errorMsg = '';
    } else if (dialogEl.open) {
      dialogEl.close();
    }
  });

  function closeModal() {
    isOpen = false;
    if (dialogEl && dialogEl.open) {
      dialogEl.close();
    }
  }

  function handleDialogClick(e) {
    if (e.target === dialogEl) {
      closeModal();
    }
  }

  async function loadOrders() {
    isLoading = true;
    errorMsg = '';
    try {
      ordersList = await api.listOrders();
    } catch (err) {
      errorMsg = err.message;
    } finally {
      isLoading = false;
    }
  }

  let filteredOrders = $derived.by(() => {
    const q = searchQuery.trim().toLowerCase();
    if (!q) return ordersList;
    return ordersList.filter(o => {
      const orderNo = (o.production_order_no || '').toLowerCase();
      const client = (o.client_name || '').toLowerCase();
      const product = (o.product_name || '').toLowerCase();
      const color = (o.color_name || '').toLowerCase();
      const type = (o.order_type || '').toLowerCase();
      return orderNo.includes(q) || client.includes(q) || product.includes(q) || color.includes(q) || type.includes(q);
    });
  });

  async function handleSelect(orderNo) {
    if (isApplying) return;
    isApplying = true;
    errorMsg = '';
    try {
      const res = await api.getOrder(orderNo);
      if (res?.success === false) {
        throw new Error(res.error || 'Failed to load order specifications');
      }
      const orderData = res?.data ?? res;
      closeModal();
      onCopySpecs(orderData);
      if (window.showToast) {
        window.showToast(`Cloned specs from ${orderNo} (Quantities reset to 0)`, 'success');
      }
    } catch (err) {
      errorMsg = err?.message || 'Failed to load specs';
      if (window.showToast) {
        window.showToast(`Failed to load specs: ${errorMsg}`, 'error');
      }
    } finally {
      isApplying = false;
    }
  }
</script>

<dialog
  bind:this={dialogEl}
  class="copy-specs-dialog"
  aria-labelledby="copy-specs-title"
  closedby="any"
  onclose={() => { isOpen = false; }}
  oncancel={() => { isOpen = false; }}
  onclick={handleDialogClick}
>
  <div class="modal-box" style="max-width:820px;width:100%;max-height:90vh;overflow:hidden;display:flex;flex-direction:column;">
    <div class="modal-header">
      <div>
        <h3 id="copy-specs-title">📋 Copy / Clone Specifications from Previous Order</h3>
        <span style="font-size:12px;color:var(--text-muted);">
          Clones product attributes, fabrics, brandings, & trims. <strong>Quantities are strictly reset to 0</strong> for the new booking.
        </span>
      </div>
      <button type="button" class="btn btn-secondary btn-sm" onclick={closeModal} aria-label="Close modal">&times;</button>
    </div>

      <div class="modal-body" style="padding:16px 20px;">
        <!-- Search Bar with Live Filter -->
        <div class="search-box-row" style="margin-bottom:16px;">
          <div style="position:relative;flex:1;">
            <input
              type="text"
              class="form-control"
              placeholder="🔍 Search orders by Order No, Client Name, Garment Product, Color, or Type..."
              bind:value={searchQuery}
            />
            {#if searchQuery}
              <button
                type="button"
                style="position:absolute;right:10px;top:50%;transform:translateY(-50%);background:none;border:none;cursor:pointer;color:var(--text-muted);font-size:14px;"
                onclick={() => searchQuery = ''}
              >
                ✕
              </button>
            {/if}
          </div>
        </div>

        {#if isLoading}
          <div style="text-align:center;padding:32px;color:var(--text-muted);">
            <div class="spinner-sm" style="margin:0 auto 10px;"></div>
            Loading historical production orders...
          </div>
        {:else if errorMsg}
          <div style="color:var(--accent-rose);padding:14px;background:rgba(239,68,68,0.1);border-radius:var(--radius-sm);">
            ⚠️ Error: {errorMsg}
          </div>
        {:else if filteredOrders.length === 0}
          <div style="text-align:center;padding:32px;color:var(--text-muted);">
            {#if searchQuery}
              No orders found matching "<strong>{searchQuery}</strong>". Try a different search term.
            {:else}
              No prior production orders found in the system.
            {/if}
          </div>
        {:else}
          <div class="data-table-container" style="max-height:420px;overflow-y:auto;">
            <table class="data-table">
              <thead>
                <tr>
                  <th>Order No</th>
                  <th>Client</th>
                  <th>Type</th>
                  <th>Garment / Product</th>
                  <th>Color</th>
                  <th class="numeric">Past Total</th>
                  <th style="text-align:right;">Action</th>
                </tr>
              </thead>
              <tbody>
                {#each filteredOrders as ord}
                  <tr>
                    <td class="input-mono">
                      <strong>{ord.production_order_no}</strong>
                    </td>
                    <td>
                      <strong>{ord.client_name}</strong>
                    </td>
                    <td>
                      <span class="badge badge-secondary">{ord.order_type || 'Bulk'}</span>
                    </td>
                    <td>{ord.product_name || 'Standard'}</td>
                    <td>{ord.color_name || '—'}</td>
                    <td class="numeric" style="color:var(--text-muted);">
                      {ord.total_qty || 0} Pcs
                    </td>
                    <td style="text-align:right;">
                      <button
                        type="button"
                        class="btn btn-primary btn-sm"
                        onclick={() => handleSelect(ord.production_order_no)}
                      >
                        📋 Copy Specs
                      </button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </div>

      <div class="modal-footer" style="justify-content:space-between;">
        <span style="font-size:12px;color:var(--text-subtle);">
          Showing {filteredOrders.length} of {ordersList.length} orders
        </span>
        <button type="button" class="btn btn-secondary" onclick={closeModal}>Cancel</button>
      </div>
    </div>
</dialog>

<style>
  dialog.copy-specs-dialog {
    padding: 0;
    border: none;
    background: transparent;
    color: inherit;
    max-width: min(820px, calc(100vw - 24px));
    width: 100%;
    margin: auto;
    overscroll-behavior: contain;
  }

  dialog.copy-specs-dialog::backdrop {
    background: rgba(10, 14, 23, 0.72);
    backdrop-filter: blur(4px);
    overflow: hidden;
    overscroll-behavior: contain;
  }
</style>
