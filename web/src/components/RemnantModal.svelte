<script>
  import { api } from '../api.js';

  let {
    isOpen = $bindable(false),
    onReturnProcessed = () => {}
  } = $props();

  let bundleCode = $state('');
  let returnType = $state('bits');
  let weight = $state('');
  let location = $state('A/R01/S01');
  let remarks = $state('');
  let isSubmitting = $state(false);

  async function handleSubmit(e) {
    e.preventDefault();
    if (!bundleCode.trim() || !weight) return;

    isSubmitting = true;
    try {
      const res = await api.returnRemnant({
        bundleCode: bundleCode.trim(),
        returnType,
        returnedWeightKg: parseFloat(weight),
        returnLocation: location.trim(),
        remarks: remarks.trim()
      });

      if (window.showToast) {
        window.showToast(`✓ Return processed! New Code: ${res.data.updatedBundleCode}`, 'success');
      }
      onReturnProcessed(res.data);
      isOpen = false;
      bundleCode = '';
      weight = '';
      remarks = '';
    } catch (err) {
      if (window.showToast) {
        window.showToast(`Return failed: ${err.message}`, 'error');
      }
    } finally {
      isSubmitting = false;
    }
  }
</script>

{#if isOpen}
  <div class="modal-overlay">
    <div class="modal-box">
      <div class="modal-header">
        <h3>Return Remnant / Bit / Defect Roll from Cutting</h3>
        <button class="btn btn-secondary btn-sm" onclick={() => isOpen = false}>&times;</button>
      </div>
      <form onsubmit={handleSubmit}>
        <div class="modal-body">
          <p style="font-size:12px;color:var(--text-muted);margin-bottom:14px;">
            Floor operators return usable bits or defect rolls. Updates stock ledger and generates <strong>Bit</strong> or <strong>Rej</strong> prefix QR codes:
          </p>

          <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px;">
            <div class="form-group">
              <label class="form-label">Original Bundle Code <span class="required">*</span></label>
              <input
                type="text"
                class="input-control input-mono"
                placeholder="e.g. 261019-A-1/4"
                bind:value={bundleCode}
                required
              />
            </div>

            <div class="form-group">
              <label class="form-label">Return Classification <span class="required">*</span></label>
              <select class="select-control" bind:value={returnType}>
                <option value="bits">Bits / Usable Remnant (Prefix: Bit)</option>
                <option value="reject">Reject / Defect Roll (Prefix: Rej)</option>
                <option value="remnant">Standard Remnant</option>
              </select>
            </div>
          </div>

          <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px;">
            <div class="form-group">
              <label class="form-label">Returned Weight (Kg) <span class="required">*</span></label>
              <input
                type="number"
                class="input-control numeric"
                step="0.01"
                min="0.01"
                placeholder="e.g. 4.85"
                bind:value={weight}
                required
              />
            </div>

            <div class="form-group">
              <label class="form-label">Return Bin Location</label>
              <input
                type="text"
                class="input-control input-mono"
                bind:value={location}
                placeholder="e.g. A/R01/S01"
              />
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">Defect / Remnant Explanation</label>
            <input
              type="text"
              class="input-control"
              placeholder="e.g. Lycra broken filament elastane drop at 18m"
              bind:value={remarks}
            />
          </div>
        </div>

        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" onclick={() => isOpen = false}>Cancel</button>
          <button type="submit" class="btn btn-primary" disabled={isSubmitting}>
            {isSubmitting ? 'Processing...' : 'Confirm Return & Generate New QR'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
