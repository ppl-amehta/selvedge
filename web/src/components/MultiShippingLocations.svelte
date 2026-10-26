<script>
  let {
    isMultiShipping = $bindable(false),
    shippingLocations = $bindable([]),
    items = [],
    orderTotalQty = 0,
    defaultFreightMode = 'Surface'
  } = $props();

  const freightModes = ['Surface', 'Air', 'Courier', 'Self-Pickup'];

  let allocatedQty = $derived.by(() => {
    return shippingLocations.reduce((sum, row) => sum + (parseInt(row.quantity, 10) || 0), 0);
  });

  function addRow() {
    shippingLocations = [
      ...shippingLocations,
      {
        id: 'dest_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6),
        orderItemId: '',
        destinationAddress: '',
        quantity: 0,
        freightMode: defaultFreightMode,
        carrier: 'Bluedart Express',
        docketNo: '',
        targetEtd: ''
      }
    ];
  }

  function removeRow(idx) {
    shippingLocations = shippingLocations.filter((_, i) => i !== idx);
  }

  $effect(() => {
    if (isMultiShipping && shippingLocations.length === 0) {
      addRow();
    }
  });
</script>

<div class="multi-shipping-box">
  <div class="shipping-header">
    <label class="toggle-container">
      <input type="checkbox" bind:checked={isMultiShipping} class="checkbox-input" />
      <span class="title">Split Across Multiple Shipping Destinations</span>
    </label>

    {#if isMultiShipping}
      <div class="allocation-badge">
        {#if allocatedQty === orderTotalQty && orderTotalQty > 0}
          <span class="status-pill success">✓ Perfect Allocation: {allocatedQty} / {orderTotalQty} Pcs</span>
        {:else if allocatedQty > orderTotalQty}
          <span class="status-pill error">⚠ Over-Allocated: {allocatedQty} / {orderTotalQty} Pcs (+{allocatedQty - orderTotalQty})</span>
        {:else}
          <span class="status-pill warn">Allocated: {allocatedQty} / {orderTotalQty} Pcs (Remaining: {orderTotalQty - allocatedQty})</span>
        {/if}
      </div>
    {/if}
  </div>

  {#if isMultiShipping}
    <div class="table-scroll-container" style="margin-top: 12px;">
      <table class="data-table">
        <thead>
          <tr>
            <th style="width: 170px;">Linked Product</th>
            <th>Delivery Address / Store Hub</th>
            <th style="width: 110px;" class="numeric">Units (Pcs)</th>
            <th style="width: 130px;">Freight Mode</th>
            <th style="width: 140px;">Carrier / Transporter</th>
            <th style="width: 130px;">Docket / AWB No</th>
            <th style="width: 130px;">Target ETD</th>
            <th style="width: 44px;"></th>
          </tr>
        </thead>
        <tbody>
          {#each shippingLocations as loc, idx}
            <tr>
              <td>
                <select class="select-control compact" bind:value={loc.orderItemId}>
                  <option value="">All Products (Consolidated)</option>
                  {#each items as it, itIdx}
                    <option value={it.id || 'item_' + (itIdx + 1)}>
                      #{itIdx + 1} {it.productName || 'Product'} ({it.totalQty || 0} Pcs)
                    </option>
                  {/each}
                </select>
              </td>
              <td>
                <input
                  type="text"
                  class="input-control compact"
                  placeholder="e.g. Bangalore Distribution Hub #4"
                  bind:value={loc.destinationAddress}
                />
              </td>
              <td>
                <input
                  type="number"
                  class="input-control compact numeric"
                  min="0"
                  placeholder="0"
                  bind:value={loc.quantity}
                />
              </td>
              <td>
                <select class="select-control compact" bind:value={loc.freightMode}>
                  {#each freightModes as fm}
                    <option value={fm}>{fm}</option>
                  {/each}
                </select>
              </td>
              <td>
                <input
                  type="text"
                  class="input-control compact"
                  placeholder="e.g. Bluedart"
                  bind:value={loc.carrier}
                />
              </td>
              <td>
                <input
                  type="text"
                  class="input-control compact input-mono"
                  placeholder="AWB No."
                  bind:value={loc.docketNo}
                />
              </td>
              <td>
                <input
                  type="date"
                  class="input-control compact"
                  bind:value={loc.targetEtd}
                />
              </td>
              <td>
                <button
                  type="button"
                  class="del-row-btn"
                  onclick={() => removeRow(idx)}
                  title="Remove location"
                >
                  ✕
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <div class="footer-actions">
      <button type="button" class="btn btn-secondary btn-sm" onclick={addRow}>
        + Add Another Shipping Destination
      </button>
    </div>
  {/if}
</div>

<style>
  .multi-shipping-box {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 16px;
    margin-top: 14px;
  }

  .shipping-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
  }

  .toggle-container {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
  }

  .checkbox-input {
    width: 16px;
    height: 16px;
    cursor: pointer;
    accent-color: var(--accent-primary);
  }

  .title {
    font-size: 12px;
    font-weight: 700;
    color: var(--text-main);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .status-pill {
    padding: 3px 10px;
    border-radius: 9999px;
    font-size: 11.5px;
    font-family: var(--font-mono);
    font-weight: 600;
  }

  .status-pill.success {
    background: rgba(16, 185, 129, 0.12);
    border: 1px solid rgba(16, 185, 129, 0.3);
    color: var(--accent-emerald);
  }

  .status-pill.error {
    background: rgba(244, 63, 94, 0.12);
    border: 1px solid rgba(244, 63, 94, 0.3);
    color: var(--accent-rose);
  }

  .status-pill.warn {
    background: rgba(245, 158, 11, 0.12);
    border: 1px solid rgba(245, 158, 11, 0.3);
    color: var(--accent-amber);
  }

  .table-scroll-container {
    overflow-x: auto;
  }

  .data-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 12px;
  }

  .data-table th {
    background: var(--bg-input);
    color: var(--text-muted);
    font-size: 10.5px;
    font-weight: 700;
    text-align: left;
    padding: 8px 10px;
    border-bottom: 1px solid var(--border-subtle);
  }

  .data-table td {
    padding: 6px 8px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.03);
  }

  .compact {
    height: 32px !important;
    font-size: 12px !important;
    padding: 4px 8px !important;
  }

  .del-row-btn {
    background: transparent;
    border: none;
    color: var(--text-subtle);
    font-size: 12px;
    cursor: pointer;
    padding: 4px 8px;
    border-radius: 4px;
  }

  .del-row-btn:hover {
    color: var(--accent-rose);
    background: rgba(244, 63, 94, 0.1);
  }

  .footer-actions {
    margin-top: 10px;
  }
</style>
