<script>
  let {
    isMultiShipping = $bindable(false),
    shippingLocations = $bindable([]),
    orderTotalQty = 0
  } = $props();

  let allocatedQty = $derived.by(() => {
    return shippingLocations.reduce((sum, row) => sum + (parseInt(row.quantity, 10) || 0), 0);
  });

  function addRow() {
    shippingLocations = [
      ...shippingLocations,
      { destination: '', quantity: '', carrier: 'Bluedart Surface', docketNo: '', etd: '' }
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
  <div class="header">
    <label class="toggle-label">
      <input type="checkbox" bind:checked={isMultiShipping} class="checkbox-input" />
      <span class="title">Split Across Multiple Shipping Destinations</span>
    </label>

    {#if isMultiShipping}
      <div class="tally-badge">
        {#if allocatedQty === orderTotalQty && orderTotalQty > 0}
          <span style="color:var(--accent-emerald);">✓ Perfect Allocation: {allocatedQty} / {orderTotalQty} Pcs</span>
        {:else if allocatedQty > orderTotalQty}
          <span style="color:var(--accent-rose);">⚠ Over-Allocated: {allocatedQty} / {orderTotalQty} Pcs (+{allocatedQty - orderTotalQty})</span>
        {:else}
          <span style="color:var(--accent-amber);">Allocated: {allocatedQty} / {orderTotalQty} Pcs (Remaining: {orderTotalQty - allocatedQty})</span>
        {/if}
      </div>
    {/if}
  </div>

  {#if isMultiShipping}
    <div class="data-table-container" style="margin-top:12px;">
      <table class="data-table">
        <thead>
          <tr>
            <th>Destination / Store Hub Address</th>
            <th class="numeric" style="width:130px;">Units (Pcs)</th>
            <th>Carrier / Freight</th>
            <th>Docket / AWB No</th>
            <th>Dispatch Target</th>
            <th style="width:50px;"></th>
          </tr>
        </thead>
        <tbody>
          {#each shippingLocations as loc, idx}
            <tr>
              <td>
                <input
                  type="text"
                  class="input-control"
                  placeholder="e.g. Bangalore Store #4"
                  bind:value={loc.destination}
                />
              </td>
              <td>
                <input
                  type="number"
                  class="input-control numeric"
                  min="0"
                  placeholder="0"
                  bind:value={loc.quantity}
                />
              </td>
              <td>
                <input
                  type="text"
                  class="input-control"
                  placeholder="Transporter"
                  bind:value={loc.carrier}
                />
              </td>
              <td>
                <input
                  type="text"
                  class="input-control input-mono"
                  placeholder="AWB / Docket"
                  bind:value={loc.docketNo}
                />
              </td>
              <td>
                <input
                  type="date"
                  class="input-control"
                  bind:value={loc.etd}
                />
              </td>
              <td>
                <button
                  type="button"
                  class="btn btn-danger btn-sm"
                  onclick={() => removeRow(idx)}
                >
                  &times;
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <button type="button" class="btn btn-secondary btn-sm" onclick={addRow} style="margin-top:12px;">
      + Add Delivery Destination
    </button>
  {/if}
</div>

<style>
  .multi-shipping-box {
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    background-color: var(--bg-surface);
    padding: 16px;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .toggle-label {
    display: flex;
    align-items: center;
    gap: 10px;
    cursor: pointer;
  }

  .checkbox-input {
    width: 17px;
    height: 17px;
    accent-color: var(--accent-primary);
  }

  .title {
    font-size: 13px;
    font-weight: 700;
    color: var(--text-main);
  }

  .tally-badge {
    font-size: 12px;
    font-family: var(--font-mono);
    font-weight: 600;
  }
</style>
