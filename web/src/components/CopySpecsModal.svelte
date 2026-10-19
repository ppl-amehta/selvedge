<script>
  import { api } from '../api.js';

  let {
    isOpen = $bindable(false),
    onCopySpecs = () => {}
  } = $props();

  let ordersList = $state([]);
  let isLoading = $state(false);
  let errorMsg = $state('');

  $effect(() => {
    if (isOpen) {
      loadOrders();
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

  async function handleSelect(orderNo) {
    try {
      const res = await api.getOrder(orderNo);
      onCopySpecs(res.data);
      isOpen = false;
      if (window.showToast) {
        window.showToast(`Cloned garment specs from ${orderNo}`, 'success');
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
    <div class="modal-box" style="max-width:760px;">
      <div class="modal-header">
        <h3>Clone Specifications from Previous Order</h3>
        <button class="btn btn-secondary btn-sm" onclick={() => isOpen = false}>&times;</button>
      </div>
      <div class="modal-body">
        <p style="font-size:12px;color:var(--text-muted);margin-bottom:12px;">
          Select any prior order to populate product, color, fabric, and trim configurations:
        </p>

        {#if isLoading}
          <div style="text-align:center;padding:24px;color:var(--text-muted);">Loading prior orders...</div>
        {:else if errorMsg}
          <div style="color:var(--accent-rose);padding:12px;">Error: {errorMsg}</div>
        {:else if ordersList.length === 0}
          <div style="text-align:center;padding:24px;color:var(--text-subtle);">No prior production orders found</div>
        {:else}
          <div class="data-table-container">
            <table class="data-table">
              <thead>
                <tr>
                  <th>Order No</th>
                  <th>Client</th>
                  <th>Product</th>
                  <th>Color</th>
                  <th class="numeric">Total Units</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {#each ordersList as ord}
                  <tr>
                    <td class="input-mono"><strong>{ord.production_order_no}</strong></td>
                    <td>{ord.client_name}</td>
                    <td>{ord.product_name}</td>
                    <td>{ord.color_name}</td>
                    <td class="numeric">{ord.total_qty} Pcs</td>
                    <td style="text-align:right;">
                      <button
                        type="button"
                        class="btn btn-primary btn-sm"
                        onclick={() => handleSelect(ord.production_order_no)}
                      >
                        Clone Specs
                      </button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => isOpen = false}>Close</button>
      </div>
    </div>
  </div>
{/if}
