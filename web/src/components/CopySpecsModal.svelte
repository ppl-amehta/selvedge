<script>
  import { api } from '../api.js';

  let {
    isOpen = $bindable(false),
    onCopySpecs = () => {}
  } = $props();

  let ordersList = $state([]);
  let searchQuery = $state('');
  let isLoading = $state(false);
  let errorMsg = $state('');

  $effect(() => {
    if (isOpen) {
      loadOrders();
      searchQuery = '';
    }
  });

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
    try {
      const res = await api.getOrder(orderNo);
      onCopySpecs(res.data);
      isOpen = false;
      if (window.showToast) {
        window.showToast(`Cloned specs from ${orderNo} (Quantities reset to 0)`, 'success');
      }
    } catch (err) {
      if (window.showToast) {
        window.showToast(`Failed to load specs: ${err.message}`, 'error');
      }
    }
  }
</script>

{#if isOpen}
  <div class="modal-overlay">
    <div class="modal-box" style="max-width:820px;">
      <div class="modal-header">
        <div>
          <h3>📋 Copy / Clone Specifications from Previous Order</h3>
          <span style="font-size:12px;color:var(--text-muted);">
            Clones product attributes, fabrics, brandings, & trims. <strong>Quantities are strictly reset to 0</strong> for the new booking.
          </span>
        </div>
        <button class="btn btn-secondary btn-sm" onclick={() => isOpen = false}>&times;</button>
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
              autofocus
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
        <button class="btn btn-secondary" onclick={() => isOpen = false}>Cancel</button>
      </div>
    </div>
  </div>
{/if}
