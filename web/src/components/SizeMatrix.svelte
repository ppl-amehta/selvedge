<script>
  let {
    sizeBreakdown = $bindable({}),
    unitPrice = 0,
    taxRate = 5,
    taxBilling = 'Extra',
    totalQty = $bindable(0),
    totalValue = $bindable(0)
  } = $props();

  const sizes = ['XS', 'S', 'M', 'L', 'XL', '2XL', '3XL', '4XL', '5XL', 'Free'];

  // Reactive Total Units via Svelte 5 $derived
  let calculatedTotalQty = $derived.by(() => {
    let sum = 0;
    for (const s of sizes) {
      sum += parseInt(sizeBreakdown[s], 10) || 0;
    }
    return sum;
  });

  // Reactive Commercial Valuation via Svelte 5 $derived
  let calculatedTotalValue = $derived.by(() => {
    const base = calculatedTotalQty * (parseFloat(unitPrice) || 0.0);
    const rate = parseFloat(taxRate) || 0.0;
    if (taxBilling === 'Extra') {
      const tax = base * (rate / 100.0);
      return base + tax;
    }
    return base;
  });

  $effect(() => {
    totalQty = calculatedTotalQty;
    totalValue = calculatedTotalValue;
  });
</script>

<div class="size-matrix-container">
  <div class="matrix-header">
    <span class="label">Production Size Matrix</span>
    <div class="totals-pill">
      <span>Total Units: <strong>{calculatedTotalQty} Pcs</strong></span>
      <span class="divider">•</span>
      <span>Total Valuation: <strong>₹{calculatedTotalValue.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</strong></span>
      <span class="tax-tag">({taxBilling} GST {taxRate}%)</span>
    </div>
  </div>

  <div class="size-grid">
    {#each sizes as size}
      <div class="size-col">
        <label for="size-{size}">{size}</label>
        <input
          id="size-{size}"
          type="number"
          class="input-control numeric size-input"
          min="0"
          placeholder="0"
          bind:value={sizeBreakdown[size]}
        />
      </div>
    {/each}
  </div>
</div>

<style>
  .size-matrix-container {
    background-color: var(--bg-input);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 16px;
  }

  .matrix-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .label {
    font-size: 11px;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-muted);
  }

  .totals-pill {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12.5px;
    font-family: var(--font-mono);
    color: var(--accent-emerald);
    background: rgba(16, 185, 129, 0.08);
    border: 1px solid rgba(16, 185, 129, 0.2);
    padding: 4px 12px;
    border-radius: 9999px;
  }

  .divider {
    color: var(--text-subtle);
  }

  .tax-tag {
    font-size: 11px;
    color: var(--text-subtle);
  }

  .size-grid {
    display: grid;
    grid-template-columns: repeat(10, 1fr);
    gap: 8px;
  }

  .size-col {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
  }

  .size-col label {
    font-size: 11px;
    font-weight: 700;
    color: var(--text-muted);
  }

  .size-input {
    text-align: center;
    font-family: var(--font-mono);
    font-weight: 700;
    padding: 8px 4px;
  }
</style>
