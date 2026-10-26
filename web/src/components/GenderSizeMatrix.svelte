<script>
  let {
    category = 'Apparel',
    totalQty = $bindable(0),
    matrix = $bindable({}),
    unitPrice = 0,
    isLocked = $bindable(false)
  } = $props();

  const apparelSizes = ['XS', 'S', 'M', 'L', 'XL', '2XL', '3XL', '4XL', '5XL', 'Free'];
  const nonApparelSizes = ['Standard / Free', 'Small', 'Medium', 'Large'];
  const allGenders = ['Men', 'Women', 'Unisex', 'Kids'];

  let activeGenders = $state(['Men', 'Women']);

  // Ensure matrix structure is initialized
  $effect(() => {
    if (category === 'Apparel') {
      for (const g of activeGenders) {
        if (!matrix[g]) {
          matrix[g] = {};
        }
      }
    } else {
      if (!matrix['General']) {
        matrix['General'] = {};
      }
    }
  });

  // Calculate grand total from matrix cells
  let calculatedMatrixSum = $derived.by(() => {
    let sum = 0;
    if (category === 'Apparel') {
      for (const g of activeGenders) {
        const row = matrix[g] || {};
        for (const s of apparelSizes) {
          sum += parseInt(row[s], 10) || 0;
        }
      }
    } else {
      const row = matrix['General'] || {};
      for (const s of nonApparelSizes) {
        sum += parseInt(row[s], 10) || 0;
      }
    }
    return sum;
  });

  // Math Checker Lock Condition
  let hasQtyEntered = $derived(totalQty > 0);
  let isBalanced = $derived(calculatedMatrixSum === totalQty && hasQtyEntered);
  let mathDiff = $derived(totalQty - calculatedMatrixSum);

  // Synchronize lock state to parent
  $effect(() => {
    if (!hasQtyEntered) {
      isLocked = true; // Lock if 0
    } else if (calculatedMatrixSum !== totalQty) {
      isLocked = true; // Hard Lock if matrix sum doesn't match totalQty
    } else {
      isLocked = false;
    }
  });

  function toggleGender(g) {
    if (activeGenders.includes(g)) {
      if (activeGenders.length > 1) {
        activeGenders = activeGenders.filter(item => item !== g);
        delete matrix[g];
      }
    } else {
      activeGenders = [...activeGenders, g];
      if (!matrix[g]) matrix[g] = {};
    }
  }

  function handleSetTotalFromMatrix() {
    totalQty = calculatedMatrixSum;
  }
</script>

<div class="matrix-box {isLocked ? 'locked-border' : 'balanced-border'}">
  <div class="matrix-header">
    <div class="title-col">
      <span class="label">Gender-Size-Quantity Production Matrix</span>
      <span class="sub-label">
        {#if category === 'Apparel'}
          Multi-Gender Breakdown • Authoritative Math Checked
        {:else}
          Non-Apparel Unit Allocation Matrix
        {/if}
      </span>
    </div>

    <!-- Math Checker Status Pill -->
    <div class="status-pill-group">
      {#if !hasQtyEntered}
        <div class="alert-badge warning">
          <span>⚠ Total Quantity is 0</span>
        </div>
      {:else if isBalanced}
        <div class="alert-badge success">
          <span class="badge-icon">✓</span>
          <span>Matrix Balances: <strong>{calculatedMatrixSum} Pcs</strong></span>
        </div>
      {:else if mathDiff > 0}
        <div class="alert-badge error">
          <span class="badge-icon">🔒</span>
          <span>Matrix Sum ({calculatedMatrixSum}) &lt; Total ({totalQty}) • <strong>Short by {mathDiff} Pcs</strong></span>
          <button type="button" class="quick-fix-btn" onclick={handleSetTotalFromMatrix} title="Sync target total to matrix sum">
            Set Total = {calculatedMatrixSum}
          </button>
        </div>
      {:else}
        <div class="alert-badge error">
          <span class="badge-icon">🔒</span>
          <span>Matrix Sum ({calculatedMatrixSum}) &gt; Total ({totalQty}) • <strong>Excess by {Math.abs(mathDiff)} Pcs</strong></span>
          <button type="button" class="quick-fix-btn" onclick={handleSetTotalFromMatrix} title="Sync target total to matrix sum">
            Set Total = {calculatedMatrixSum}
          </button>
        </div>
      {/if}
    </div>
  </div>

  {#if category === 'Apparel'}
    <!-- Gender Toggle Filter Chips -->
    <div class="gender-bar">
      <span class="gender-label">Active Genders:</span>
      {#each allGenders as g}
        <button
          type="button"
          class="gender-chip {activeGenders.includes(g) ? 'active' : ''}"
          onclick={() => toggleGender(g)}
        >
          {activeGenders.includes(g) ? '✓ ' : '+ '}{g}
        </button>
      {/each}
    </div>

    <!-- 2D Matrix Grid -->
    <div class="table-scroll-container">
      <table class="matrix-table">
        <thead>
          <tr>
            <th class="gender-col">Gender</th>
            {#each apparelSizes as s}
              <th class="size-col">{s}</th>
            {/each}
            <th class="total-col">Subtotal</th>
          </tr>
        </thead>
        <tbody>
          {#each activeGenders as g}
            {@const gRow = matrix[g] || {}}
            {@const rowSum = apparelSizes.reduce((sum, s) => sum + (parseInt(gRow[s], 10) || 0), 0)}
            <tr>
              <td class="gender-cell"><strong>{g}</strong></td>
              {#each apparelSizes as s}
                <td class="size-cell">
                  <input
                    type="number"
                    min="0"
                    placeholder="0"
                    class="cell-input {parseInt(gRow[s], 10) > 0 ? 'filled' : ''}"
                    bind:value={gRow[s]}
                  />
                </td>
              {/each}
              <td class="row-sum-cell">{rowSum} Pcs</td>
            </tr>
          {/each}
        </tbody>
        <tfoot>
          <tr class="summary-row">
            <td class="gender-cell"><strong>Column Totals</strong></td>
            {#each apparelSizes as s}
              {@const colSum = activeGenders.reduce((sum, g) => sum + (parseInt((matrix[g] || {})[s], 10) || 0), 0)}
              <td class="col-sum-cell">{colSum || '-'}</td>
            {/each}
            <td class="grand-total-cell">
              <strong>{calculatedMatrixSum} Pcs</strong>
            </td>
          </tr>
        </tfoot>
      </table>
    </div>
  {:else}
    <!-- Non-Apparel Simplified Size Matrix -->
    <div class="non-apparel-grid">
      {#each nonApparelSizes as s}
        <div class="non-apparel-item">
          <label for="non-apparel-{s}">{s}</label>
          <input
            id="non-apparel-{s}"
            type="number"
            min="0"
            placeholder="0"
            class="input-control numeric"
            value={matrix['General']?.[s] ?? ''}
            oninput={(e) => {
              if (!matrix['General']) matrix['General'] = {};
              matrix['General'][s] = parseInt(e.target.value, 10) || 0;
            }}
          />
        </div>
      {/each}
      <div class="non-apparel-total">
        <span class="nat-label">Allocated Units</span>
        <span class="nat-value {isBalanced ? 'success' : 'warn'}">{calculatedMatrixSum} Pcs</span>
      </div>
    </div>
  {/if}

  {#if isLocked}
    <div class="lock-warning-bar">
      <span>🔒 <strong>ACTION LOCKED:</strong> Production orders cannot be booked until the Gender-Size Matrix sum exactly equals the line item Total Quantity ({totalQty} Pcs).</span>
    </div>
  {/if}
</div>

<style>
  .matrix-box {
    background-color: var(--bg-input);
    border: 1.5px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 16px;
    transition: border-color 0.2s ease;
  }

  .matrix-box.locked-border {
    border-color: rgba(244, 63, 94, 0.45);
    background: rgba(244, 63, 94, 0.02);
  }

  .matrix-box.balanced-border {
    border-color: rgba(16, 185, 129, 0.35);
  }

  .matrix-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
    flex-wrap: wrap;
    gap: 8px;
  }

  .title-col {
    display: flex;
    flex-direction: column;
  }

  .label {
    font-size: 11px;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-main);
  }

  .sub-label {
    font-size: 10.5px;
    color: var(--text-muted);
  }

  .status-pill-group {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .alert-badge {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 5px 12px;
    border-radius: 9999px;
    font-size: 12px;
    font-family: var(--font-mono);
    font-weight: 600;
  }

  .alert-badge.success {
    background: rgba(16, 185, 129, 0.12);
    border: 1px solid rgba(16, 185, 129, 0.3);
    color: var(--accent-emerald);
  }

  .alert-badge.error {
    background: rgba(244, 63, 94, 0.15);
    border: 1px solid rgba(244, 63, 94, 0.4);
    color: var(--accent-rose);
  }

  .alert-badge.warning {
    background: rgba(245, 158, 11, 0.12);
    border: 1px solid rgba(245, 158, 11, 0.3);
    color: var(--accent-amber);
  }

  .quick-fix-btn {
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.2);
    color: #fff;
    font-size: 10.5px;
    padding: 2px 8px;
    border-radius: 4px;
    cursor: pointer;
  }

  .quick-fix-btn:hover {
    background: rgba(255, 255, 255, 0.2);
  }

  .gender-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
  }

  .gender-label {
    font-size: 11px;
    color: var(--text-muted);
    font-weight: 600;
  }

  .gender-chip {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    color: var(--text-muted);
    font-size: 11px;
    font-weight: 600;
    padding: 4px 10px;
    border-radius: var(--radius-sm);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .gender-chip.active {
    background: rgba(37, 99, 235, 0.15);
    border-color: var(--accent-primary);
    color: #93c5fd;
  }

  .table-scroll-container {
    overflow-x: auto;
  }

  .matrix-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 12px;
  }

  .matrix-table th, .matrix-table td {
    padding: 6px 8px;
    text-align: center;
    border: 1px solid var(--border-subtle);
  }

  .matrix-table th {
    background: var(--bg-surface);
    color: var(--text-muted);
    font-size: 10.5px;
    font-weight: 700;
  }

  .gender-col, .gender-cell {
    text-align: left !important;
    width: 80px;
    background: var(--bg-surface);
  }

  .size-cell {
    padding: 3px !important;
  }

  .cell-input {
    width: 100%;
    min-width: 48px;
    height: 32px;
    background: transparent;
    border: 1px solid transparent;
    color: var(--text-main);
    text-align: center;
    font-family: var(--font-mono);
    font-size: 12px;
    border-radius: 4px;
  }

  .cell-input:focus {
    background: var(--bg-elevated);
    border-color: var(--accent-primary);
    outline: none;
  }

  .cell-input.filled {
    background: rgba(37, 99, 235, 0.08);
    font-weight: 700;
    color: #60a5fa;
  }

  .row-sum-cell, .col-sum-cell {
    font-family: var(--font-mono);
    font-weight: 600;
    color: var(--text-muted);
    background: rgba(255, 255, 255, 0.02);
  }

  .grand-total-cell {
    font-family: var(--font-mono);
    font-weight: 800;
    color: var(--accent-emerald);
    background: rgba(16, 185, 129, 0.08);
  }

  .non-apparel-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(130px, 1fr));
    gap: 12px;
    align-items: flex-end;
  }

  .non-apparel-item label {
    display: block;
    font-size: 11px;
    color: var(--text-muted);
    margin-bottom: 4px;
    font-weight: 600;
  }

  .non-apparel-total {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    padding: 8px 12px;
    border-radius: var(--radius-sm);
    display: flex;
    flex-direction: column;
    justify-content: center;
  }

  .nat-label {
    font-size: 10px;
    color: var(--text-subtle);
    text-transform: uppercase;
  }

  .nat-value {
    font-family: var(--font-mono);
    font-weight: 800;
    font-size: 14px;
  }

  .nat-value.success { color: var(--accent-emerald); }
  .nat-value.warn { color: var(--accent-rose); }

  .lock-warning-bar {
    margin-top: 12px;
    padding: 8px 12px;
    background: rgba(244, 63, 94, 0.1);
    border-left: 3px solid var(--accent-rose);
    border-radius: 4px;
    color: #fca5a5;
    font-size: 11.5px;
  }
</style>
