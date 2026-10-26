<script>
  import { api } from '../api.js';

  let {
    isOpen = $bindable(false),
    onApplyOrder = () => {}
  } = $props();

  let salesOrderNo = $state('SO-GOOGLE-001');
  let isFetching = $state(false);
  let errorMsg = $state('');
  let fetchedData = $state(null);

  const sampleOrders = [
    { code: 'SO-GOOGLE-001', label: 'Google India (Polo & Sipper)' },
    { code: 'SO-CRED-002', label: 'CRED (Premium Round Neck Tee)' },
    { code: 'SO-ZERODHA-003', label: 'Zerodha (Corporate Apparel & Swag)' }
  ];

  async function handleFetch() {
    const soNo = salesOrderNo.trim();
    if (!soNo) {
      errorMsg = 'Please enter a Zoho Sales Order No';
      return;
    }

    try {
      isFetching = true;
      errorMsg = '';
      fetchedData = null;
      const res = await api.fetchZohoOrder(soNo);
      fetchedData = res;
    } catch (err) {
      errorMsg = err.message || 'Failed to pull order from Zoho';
    } finally {
      isFetching = false;
    }
  }

  function handleApply() {
    if (!fetchedData) return;
    onApplyOrder(fetchedData);
    isOpen = false;
    if (window.showToast) {
      window.showToast(`Applied ${fetchedData.items?.length || 0} product lines from Zoho Sales Order ${fetchedData.salesOrderNo}`, 'success');
    }
  }
</script>

{#if isOpen}
  <div class="modal-backdrop" onclick={() => isOpen = false} role="presentation">
    <div class="modal-card" onclick={(e) => e.stopPropagation()} role="dialog">
      <div class="modal-head">
        <div class="title-group">
          <span class="badge">Zoho Books / Inventory API</span>
          <h3>Programmatic Order Pull from Zoho</h3>
        </div>
        <button type="button" class="close-btn" onclick={() => isOpen = false}>✕</button>
      </div>

      <div class="modal-body">
        <div class="search-row">
          <div class="input-wrap">
            <label for="zoho-so-input">Sales Order Number</label>
            <input
              id="zoho-so-input"
              type="text"
              class="input-control input-mono"
              placeholder="e.g. SO-2026-7890"
              bind:value={salesOrderNo}
              onkeydown={(e) => e.key === 'Enter' && handleFetch()}
            />
          </div>
          <button
            type="button"
            class="btn btn-primary fetch-btn"
            onclick={handleFetch}
            disabled={isFetching}
          >
            {isFetching ? 'Fetching...' : '📥 Pull from Zoho'}
          </button>
        </div>

        <!-- Sample Quick Picks -->
        <div class="sample-chips">
          <span class="chip-label">Quick Test Orders:</span>
          {#each sampleOrders as s}
            <button
              type="button"
              class="chip-btn"
              onclick={() => { salesOrderNo = s.code; handleFetch(); }}
            >
              {s.code} ({s.label.split(' ')[0]})
            </button>
          {/each}
        </div>

        {#if errorMsg}
          <div class="error-banner">
            <span>⚠ {errorMsg}</span>
          </div>
        {/if}

        <!-- Retrieved Data Preview -->
        {#if fetchedData}
          <div class="data-preview-box">
            <div class="preview-banner">
              <span class="preview-status">✓ Sales Order Located in Zoho</span>
              <span class="preview-so">{fetchedData.salesOrderNo}</span>
            </div>

            <!-- Client Info -->
            <div class="info-grid">
              <div class="info-item">
                <span class="info-label">Client Name</span>
                <span class="info-val highlight">{fetchedData.clientName}</span>
              </div>
              <div class="info-item">
                <span class="info-label">Contact Person</span>
                <span class="info-val">{fetchedData.contactPerson} ({fetchedData.phone})</span>
              </div>
              <div class="info-item">
                <span class="info-label">GSTIN</span>
                <span class="info-val">{fetchedData.gstin}</span>
              </div>
              <div class="info-item">
                <span class="info-label">Tax Rate & Billing</span>
                <span class="info-val">{fetchedData.taxRate}% GST ({fetchedData.taxBilling})</span>
              </div>
              <div class="info-item full">
                <span class="info-label">Billing & Shipping Address</span>
                <span class="info-val">{fetchedData.billingAddress}</span>
              </div>
            </div>

            <!-- Items Table Preview -->
            <div class="items-preview-title">Product Line Items ({fetchedData.items?.length || 0})</div>
            <div class="preview-table-wrap">
              <table class="preview-table">
                <thead>
                  <tr>
                    <th>Category</th>
                    <th>Sub-Category</th>
                    <th>Product Name</th>
                    <th>Color</th>
                    <th class="numeric">Qty (Pcs)</th>
                    <th class="numeric">Unit Price</th>
                  </tr>
                </thead>
                <tbody>
                  {#each fetchedData.items || [] as it}
                    <tr>
                      <td><span class="cat-pill">{it.category}</span></td>
                      <td>{it.subCategory}</td>
                      <td><strong>{it.productName}</strong></td>
                      <td>{it.colorName}</td>
                      <td class="numeric font-mono">{it.totalQty}</td>
                      <td class="numeric font-mono">₹{it.unitPrice}</td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          </div>
        {/if}
      </div>

      <div class="modal-foot">
        <button type="button" class="btn btn-secondary" onclick={() => isOpen = false}>
          Cancel
        </button>
        <button
          type="button"
          class="btn btn-primary"
          onclick={handleApply}
          disabled={!fetchedData}
        >
          ✓ Populate Order Form with Zoho Data
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 20px;
  }

  .modal-card {
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg);
    width: 100%;
    max-width: 680px;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.6);
  }

  .modal-head {
    padding: 16px 20px;
    border-bottom: 1px solid var(--border-subtle);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .badge {
    display: inline-block;
    font-size: 10px;
    background: rgba(37, 99, 235, 0.15);
    border: 1px solid rgba(37, 99, 235, 0.3);
    color: #93c5fd;
    padding: 2px 8px;
    border-radius: 4px;
    font-weight: 700;
    text-transform: uppercase;
    margin-bottom: 4px;
  }

  .title-group h3 {
    font-size: 16px;
    font-weight: 700;
    color: var(--text-main);
  }

  .close-btn {
    background: transparent;
    border: none;
    color: var(--text-muted);
    font-size: 16px;
    cursor: pointer;
  }

  .modal-body {
    padding: 16px 20px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .search-row {
    display: flex;
    gap: 10px;
    align-items: flex-end;
  }

  .input-wrap {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .input-wrap label {
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    color: var(--text-muted);
  }

  .fetch-btn {
    height: 38px;
    white-space: nowrap;
  }

  .sample-chips {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
  }

  .chip-label {
    font-size: 11px;
    color: var(--text-subtle);
  }

  .chip-btn {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    color: var(--text-muted);
    font-size: 11px;
    padding: 3px 8px;
    border-radius: var(--radius-sm);
    cursor: pointer;
    font-family: var(--font-mono);
  }

  .chip-btn:hover {
    border-color: var(--accent-primary);
    color: #fff;
  }

  .error-banner {
    padding: 10px 12px;
    background: rgba(244, 63, 94, 0.1);
    border: 1px solid rgba(244, 63, 94, 0.3);
    border-radius: var(--radius-sm);
    color: var(--accent-rose);
    font-size: 12px;
  }

  .data-preview-box {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 14px;
    margin-top: 4px;
  }

  .preview-banner {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--border-subtle);
  }

  .preview-status {
    font-size: 12px;
    font-weight: 700;
    color: var(--accent-emerald);
  }

  .preview-so {
    font-family: var(--font-mono);
    font-size: 12px;
    font-weight: 700;
    color: var(--accent-primary);
  }

  .info-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
    font-size: 12px;
    margin-bottom: 14px;
  }

  .info-item {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .info-item.full {
    grid-column: 1 / -1;
  }

  .info-label {
    font-size: 10px;
    text-transform: uppercase;
    color: var(--text-subtle);
    font-weight: 700;
  }

  .info-val {
    color: var(--text-main);
  }

  .info-val.highlight {
    font-weight: 700;
    color: #93c5fd;
  }

  .items-preview-title {
    font-size: 11px;
    font-weight: 800;
    text-transform: uppercase;
    color: var(--text-muted);
    margin-bottom: 6px;
  }

  .preview-table-wrap {
    overflow-x: auto;
  }

  .preview-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 11.5px;
  }

  .preview-table th {
    background: var(--bg-input);
    color: var(--text-muted);
    font-size: 10px;
    font-weight: 700;
    padding: 6px 8px;
    text-align: left;
  }

  .preview-table td {
    padding: 6px 8px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.03);
  }

  .cat-pill {
    font-size: 10px;
    background: rgba(255, 255, 255, 0.08);
    padding: 2px 6px;
    border-radius: 4px;
  }

  .numeric { text-align: right; }
  .font-mono { font-family: var(--font-mono); }

  .modal-foot {
    padding: 14px 20px;
    border-top: 1px solid var(--border-subtle);
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }
</style>
