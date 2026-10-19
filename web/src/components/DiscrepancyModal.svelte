<script>
  import { api } from '../api.js';

  let {
    isOpen = $bindable(false)
  } = $props();

  let reports = $state([]);
  let isLoading = $state(false);

  $effect(() => {
    if (isOpen) {
      loadReports();
    }
  });

  async function loadReports() {
    isLoading = true;
    try {
      reports = await api.listDiscrepancies();
    } catch (err) {
      console.error(err);
    } finally {
      isLoading = false;
    }
  }
</script>

{#if isOpen}
  <div class="modal-overlay">
    <div class="modal-box" style="max-width:880px;">
      <div class="modal-header">
        <h3>Dock Shortfall Discrepancy & Variance Register</h3>
        <button class="btn btn-secondary btn-sm" onclick={() => isOpen = false}>&times;</button>
      </div>
      <div class="modal-body">
        {#if isLoading}
          <div style="text-align:center;padding:24px;color:var(--text-muted);">Loading discrepancy ledger...</div>
        {:else if reports.length === 0}
          <div style="text-align:center;padding:24px;color:var(--text-subtle);">No dock discrepancies logged</div>
        {:else}
          <div class="data-table-container">
            <table class="data-table">
              <thead>
                <tr>
                  <th>Report No</th>
                  <th>Parent Code</th>
                  <th>Vendor</th>
                  <th>Invoice No</th>
                  <th class="numeric">Inv Qty</th>
                  <th class="numeric">Actual Scale</th>
                  <th class="numeric">Shortfall</th>
                  <th class="numeric">Financial Exposure</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {#each reports as r}
                  <tr>
                    <td class="input-mono"><strong>{r.report_no}</strong></td>
                    <td class="input-mono">{r.parent_code}</td>
                    <td>{r.vendor_name}</td>
                    <td class="input-mono">{r.invoice_no}</td>
                    <td class="numeric">{r.invoice_qty}</td>
                    <td class="numeric">{r.actual_qty}</td>
                    <td class="numeric" style="color:var(--accent-rose);font-weight:700;">
                      {r.variance_qty < 0 ? r.variance_qty : `-${r.variance_qty}`}
                    </td>
                    <td class="numeric" style="font-weight:700;">
                      ₹{r.financial_impact?.toFixed(2) || '0.00'}
                    </td>
                    <td>
                      <span class="badge-option selected" style="font-size:10.5px;padding:3px 8px;">
                        {r.status}
                      </span>
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
