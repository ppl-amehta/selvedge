<script>
  import { onMount } from 'svelte';
  import { api } from '../api.js';

  let { masters = {} } = $props();

  // State
  let orders = $state([]);
  let selectedOrderId = $state('');
  let activeJob = $state(null);
  let isLoading = $state(false);
  let activeTab = $state('lay'); // 'lay' | 'sizes' | 'bundles' | 'remnants'

  // Lay Form State
  let rollCode = $state('');
  let markerName = $state('Marker-Std-1');
  let plies = $state(20);
  let layWeightKg = $state('');
  let operatorName = $state('Floor Operator 1');
  let ratio = $state({}); // { S: 1, M: 2, L: 2, XL: 1 }

  // Bundle Print Modal State
  let selectedBundle = $state(null);
  let isSlipModalOpen = $state(false);

  // Remnant Form State
  let remnantRollCode = $state('');
  let remnantType = $state('Bit'); // 'Bit' | 'Rej'
  let remnantWeightKg = $state('');
  let remnantLocation = $state('RACK-CUT-01');
  let remnantReason = $state('');

  // Derived calculations
  let totalPiecesInLay = $derived(() => {
    let sumRatio = 0;
    for (const sz in ratio) {
      sumRatio += (Number(ratio[sz]) || 0);
    }
    return sumRatio * (Number(plies) || 0);
  });

  let efficiencyEstimate = $derived(() => {
    const pcs = totalPiecesInLay();
    const wt = parseFloat(layWeightKg) || 0;
    if (pcs > 0 && wt > 0 && activeJob?.plannedBomAvgKg > 0) {
      const actualAvg = wt / pcs;
      const planned = activeJob.plannedBomAvgKg;
      const eff = (planned / actualAvg) * 100;
      return eff.toFixed(1);
    }
    return '100.0';
  });

  async function loadOrders() {
    isLoading = true;
    try {
      orders = await api.listCuttingOrders();
      if (orders.length > 0 && !selectedOrderId) {
        selectOrder(orders[0].id);
      }
    } catch (err) {
      if (window.showToast) window.showToast('Failed to load cutting orders: ' + err.message, 'error');
    } finally {
      isLoading = false;
    }
  }

  async function selectOrder(id) {
    selectedOrderId = id;
    isLoading = true;
    try {
      const raw = await api.getCuttingJob(id);
      const job = raw?.job || {};
      const hdr = raw?.orderHeader || {};
      activeJob = {
        ...raw,
        id: job.id || id,
        clientName: hdr.clientName || raw?.clientName || '',
        productionOrderNo: job.productionOrderNo || raw?.productionOrderNo || '',
        productName: job.productName || raw?.productName || '',
        colorName: job.colorName || raw?.colorName || '',
        fabricName: job.fabricName || raw?.fabricName || '',
        targetQty: raw?.totalTargetQty ?? job.targetQty ?? 0,
        pharmaNo: job.pharmaNo || raw?.pharmaNo || '',
        sizesProgress: raw?.sizeProgressList || raw?.sizesProgress || [],
        cutLays: raw?.lays || raw?.cutLays || [],
        cutBundles: raw?.bundles || raw?.cutBundles || []
      };
      // Pre-fill ratio keys from active job sizes
      if (activeJob?.sizesProgress) {
        const initialRatio = {};
        activeJob.sizesProgress.forEach(s => {
          initialRatio[s.size] = ratio[s.size] || 1;
        });
        ratio = initialRatio;
      }
      if (activeJob?.issuedRolls?.length > 0 && (!rollCode || !activeJob.issuedRolls.some(r => r.rollCode === rollCode))) {
        rollCode = activeJob.issuedRolls[0].rollCode;
      } else if (!activeJob?.issuedRolls?.length) {
        rollCode = '';
      }
    } catch (err) {
      activeJob = null;
      if (window.showToast) window.showToast('Failed to load cutting job: ' + err.message, 'error');
    } finally {
      isLoading = false;
    }
  }

  async function handleRecordLay(e) {
    if (e) e.preventDefault();
    if (!selectedOrderId) return;
    if (!rollCode) {
      if (window.showToast) window.showToast('Please select or scan an issued fabric roll', 'error');
      return;
    }
    const pcs = totalPiecesInLay();
    if (pcs <= 0) {
      if (window.showToast) window.showToast('Total pieces in lay must be greater than zero', 'error');
      return;
    }

    try {
      const selectedRoll = (activeJob?.issuedRolls || []).find(r => r.rollCode === rollCode.trim());
      const payload = {
        cuttingJobId: activeJob?.id || selectedOrderId,
        rollCode: rollCode.trim(),
        fabricName: selectedRoll?.fabricName || activeJob?.fabricName || '',
        rollWeightKg: Number(selectedRoll?.rollWeightKg ?? selectedRoll?.issuedWeightKg) || 0,
        markerName: markerName.trim(),
        ratio: ratio,
        plies: Number(plies) || 1,
        layWeightKg: parseFloat(layWeightKg) || 0,
        operatorName: operatorName.trim()
      };
      const res = await api.recordCutLay(payload);
      if (window.showToast) {
        window.showToast(`✓ Cut Lay #${res?.layNumber || ''} recorded (${res?.totalPiecesInLay || pcs} pcs)!`, 'success');
      }
      // Reload job data & orders summary counters
      await Promise.all([selectOrder(selectedOrderId), loadOrders()]);
      activeTab = 'bundles';
    } catch (err) {
      if (window.showToast) window.showToast('Failed to record cut lay: ' + err.message, 'error');
    }
  }

  async function handleRecordRemnant(e) {
    if (e) e.preventDefault();
    if (!selectedOrderId || !remnantRollCode || !remnantWeightKg) return;

    try {
      const payload = {
        cuttingJobId: activeJob?.id || selectedOrderId,
        sourceRollCode: remnantRollCode.trim(),
        prefix: remnantType,
        weightKg: parseFloat(remnantWeightKg),
        binLocation: remnantLocation.trim(),
        defectReason: remnantReason.trim(),
        operatorName: operatorName.trim()
      };
      const res = await api.recordCuttingRemnant(payload);
      if (window.showToast) {
        window.showToast(`✓ Remnant recorded! New Code: ${res?.id || 'Stored'}`, 'success');
      }
      remnantWeightKg = '';
      remnantReason = '';
      await selectOrder(selectedOrderId);
      activeTab = 'remnants';
    } catch (err) {
      if (window.showToast) window.showToast('Failed to record remnant: ' + err.message, 'error');
    }
  }

  async function openSlip(bundle) {
    try {
      selectedBundle = await api.getBundleSlip(bundle.id);
      isSlipModalOpen = true;
    } catch (err) {
      if (window.showToast) window.showToast('Failed to load bundle slip: ' + err.message, 'error');
    }
  }

  onMount(() => {
    loadOrders();
  });
</script>

<div class="cutting-container">
  <!-- Top Operator Bar: Order Queue -->
  <header class="cutting-header">
    <div class="header-left">
      <span class="station-icon">✂️</span>
      <div>
        <h2>Station 4: Floor Cutting Station</h2>
        <p class="subtitle">Industrial Lay Planning, Roll Issue, Bundle QR Generation & Remnant Control</p>
      </div>
    </div>

    <div class="header-actions">
      <button class="btn btn-secondary btn-sm" onclick={loadOrders} disabled={isLoading}>
        🔄 Refresh Queue
      </button>
    </div>
  </header>

  <!-- Horizontal Job Selector Strip -->
  <div class="job-strip">
    {#each orders as ord (ord.id)}
      <button
        class="job-card {selectedOrderId === ord.id ? 'active' : ''}"
        onclick={() => selectOrder(ord.id)}
      >
        <div class="job-card-top">
          <span class="order-no">{ord.productionOrderNo}</span>
          <span class="status-pill status-{ord.cuttingJobStatus.toLowerCase()}">{ord.cuttingJobStatus}</span>
        </div>
        <div class="client-name">{ord.clientName}</div>
        <div class="product-info">{ord.productName} ({ord.colorName})</div>
        <div class="qty-progress">
          <div class="qty-labels">
            <span>Cut: <strong>{ord.totalCutQty}</strong> / {ord.targetQty}</span>
            <span class="balance">Bal: {ord.balanceQty}</span>
          </div>
          <div class="progress-bar-bg">
            <div
              class="progress-bar-fill"
              style="width: {Math.min(100, Math.round((ord.totalCutQty / (ord.targetQty || 1)) * 100))}%"
            ></div>
          </div>
        </div>
      </button>
    {/each}
  </div>

  {#if activeJob}
    <!-- Active Job Details Banner -->
    <div class="active-banner">
      <div class="banner-col">
        <span class="label">Client & Order:</span>
        <span class="value">{activeJob.clientName} • #{activeJob.productionOrderNo}</span>
      </div>
      <div class="banner-col">
        <span class="label">Product / Fabric:</span>
        <span class="value">{activeJob.productName} ({activeJob.colorName}) • {activeJob.fabricName}</span>
      </div>
      <div class="banner-col">
        <span class="label">Pharma Tech Pack:</span>
        <span class="value badge-pharma">{activeJob.pharmaNo || 'Standard SOP'}</span>
      </div>
      <div class="banner-col">
        <span class="label">Overall Progress:</span>
        <span class="value text-success">{activeJob.totalCutQty} / {activeJob.targetQty} pcs ({Math.round((activeJob.totalCutQty / (activeJob.targetQty || 1)) * 100)}%)</span>
      </div>
    </div>

    <!-- Navigation Tabs -->
    <div class="cutting-tabs">
      <button class="tab-btn {activeTab === 'lay' ? 'active' : ''}" onclick={() => activeTab = 'lay'}>
        ⚡ Record Cut Lay
      </button>
      <button class="tab-btn {activeTab === 'sizes' ? 'active' : ''}" onclick={() => activeTab = 'sizes'}>
        📊 Size Progress ({activeJob?.sizesProgress?.length || 0})
      </button>
      <button class="tab-btn {activeTab === 'bundles' ? 'active' : ''}" onclick={() => activeTab = 'bundles'}>
        📦 Cut Bundles ({activeJob?.cutBundles?.length || 0})
      </button>
      <button class="tab-btn {activeTab === 'remnants' ? 'active' : ''}" onclick={() => activeTab = 'remnants'}>
        ♻️ Remnants & Bits ({activeJob?.remnants?.length || 0})
      </button>
    </div>

    <!-- TAB 1: RECORD CUT LAY -->
    {#if activeTab === 'lay'}
      <div class="lay-grid">
        <!-- Form Left Column -->
        <div class="lay-form-card">
          <h3>Record New Cutting Lay</h3>
          <p class="section-desc">Enter marker ratio and plies. The system will automatically compute bundle slips and update inventory consumption.</p>

          <form onsubmit={handleRecordLay}>
            <div class="form-row">
              <div class="form-group flex-1">
                <label for="cut-roll-code">Fabric Roll Code <span class="req">*</span></label>
                <div class="input-with-button">
                  <input
                    id="cut-roll-code"
                    type="text"
                    class="form-control"
                    placeholder="Scan or enter Roll QR..."
                    bind:value={rollCode}
                    required
                  />
                  {#if activeJob?.issuedRolls?.length > 0}
                    <select
                      class="form-control select-compact"
                      aria-label="Pick Issued Roll"
                      onchange={(e) => rollCode = e.currentTarget.value}
                    >
                      <option value="">-- Pick Issued Roll --</option>
                      {#each activeJob.issuedRolls as r}
                        <option value={r.rollCode}>{r.rollCode} ({r.rollWeightKg ?? r.issuedWeightKg ?? 0} kg)</option>
                      {/each}
                    </select>
                  {/if}
                </div>
              </div>

              <div class="form-group flex-1">
                <label for="cut-marker-id">Marker Identifier</label>
                <input
                  id="cut-marker-id"
                  type="text"
                  class="form-control"
                  placeholder="e.g. Marker-M-L-2"
                  bind:value={markerName}
                />
              </div>
            </div>

            <!-- Plies Counter with Big Touch Increments -->
            <div class="form-group">
              <label for="cut-plies">Number of Plies (Layers) <span class="req">*</span></label>
              <div class="plies-control">
                <button type="button" class="btn btn-touch" onclick={() => plies = Math.max(1, plies - 5)}>-5</button>
                <button type="button" class="btn btn-touch" onclick={() => plies = Math.max(1, plies - 1)}>-1</button>
                <input id="cut-plies" type="number" class="form-control plies-input" min="1" bind:value={plies} required />
                <button type="button" class="btn btn-touch" onclick={() => plies = plies + 1}>+1</button>
                <button type="button" class="btn btn-touch" onclick={() => plies = plies + 5}>+5</button>
                <button type="button" class="btn btn-touch" onclick={() => plies = plies + 10}>+10</button>
              </div>
            </div>

            <!-- Size Ratio Grid -->
            <div class="form-group">
              <span class="field-label">Marker Size Ratio (Pieces per Ply) <span class="req">*</span></span>
              <div class="ratio-matrix">
                {#if activeJob?.sizesProgress}
                  {#each activeJob.sizesProgress as sp}
                    <div class="ratio-cell">
                      <span class="ratio-size-title">{sp.size}</span>
                      <input
                        type="number"
                        min="0"
                        aria-label="Ratio for size {sp.size}"
                        class="form-control ratio-input"
                        bind:value={ratio[sp.size]}
                        placeholder="0"
                      />
                      <span class="ratio-sub">Bal: {sp.balance}</span>
                    </div>
                  {/each}
                {/if}
              </div>
            </div>

            <div class="form-row">
              <div class="form-group flex-1">
                <label for="cut-lay-weight">Lay Fabric Weight (Kg)</label>
                <input
                  id="cut-lay-weight"
                  type="number"
                  step="0.01"
                  class="form-control"
                  placeholder="e.g. 18.45"
                  bind:value={layWeightKg}
                />
              </div>
              <div class="form-group flex-1">
                <label for="cut-operator">Floor Operator Name</label>
                <input
                  id="cut-operator"
                  type="text"
                  class="form-control"
                  bind:value={operatorName}
                />
              </div>
            </div>

            <!-- Lay Summary & Submit -->
            <div class="lay-summary-box">
              <div class="stat-pill">
                <span class="stat-lbl">Total Pieces:</span>
                <span class="stat-val highlight">{totalPiecesInLay()} pcs</span>
              </div>
              <div class="stat-pill">
                <span class="stat-lbl">Est. Efficiency:</span>
                <span class="stat-val {parseFloat(efficiencyEstimate()) >= 95 ? 'text-success' : 'text-warning'}">
                  {efficiencyEstimate()}%
                </span>
              </div>
            </div>

            <button type="submit" class="btn btn-primary btn-block btn-lg">
              ⚡ CUT LAY & GENERATE PIECE BUNDLES
            </button>
          </form>
        </div>

        <!-- Right Column: Recent Lays -->
        <div class="lay-history-card">
          <h3>Recent Cut Lays ({activeJob?.cutLays?.length || 0})</h3>
          {#if activeJob?.cutLays?.length > 0}
            <div class="lays-scroll">
              {#each activeJob.cutLays as lay}
                <div class="lay-history-item">
                  <div class="lay-item-head">
                    <span class="lay-badge">Lay #{lay.layNumber}</span>
                    <span class="lay-time">{new Date(lay.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                  </div>
                  <div class="lay-item-body">
                    <div>Roll: <strong>{lay.rollCode}</strong></div>
                    <div>Plies: <strong>{lay.plies}</strong> • Pcs: <strong>{lay.totalPiecesInLay}</strong></div>
                    <div>Efficiency: <span class="text-success">{lay.efficiencyPct?.toFixed(1) || 100}%</span></div>
                  </div>
                </div>
              {/each}
            </div>
          {:else}
            <div class="empty-state">No lays recorded yet for this order.</div>
          {/if}
        </div>
      </div>
    {/if}

    <!-- TAB 2: SIZES PROGRESS -->
    {#if activeTab === 'sizes'}
      <div class="table-card">
        <h3>Size Balance Matrix</h3>
        <table class="data-table">
          <thead>
            <tr>
              <th>Size</th>
              <th>Target Qty</th>
              <th>Previously Cut</th>
              <th>Total Cut</th>
              <th>Balance</th>
              <th>Completion</th>
            </tr>
          </thead>
          <tbody>
            {#if activeJob?.sizesProgress}
              {#each activeJob.sizesProgress as sz}
                <tr>
                  <td><strong class="size-tag">{sz.size}</strong></td>
                  <td>{sz.targetQty}</td>
                  <td>{sz.previouslyCut}</td>
                  <td><strong class="text-success">{sz.totalCut}</strong></td>
                  <td><strong class="{sz.balance > 0 ? 'text-warning' : 'text-muted'}">{sz.balance}</strong></td>
                  <td>
                    <div class="progress-container">
                      <div class="progress-bar-bg">
                        <div class="progress-bar-fill" style="width: {sz.completionPct}%"></div>
                      </div>
                      <span class="pct-text">{sz.completionPct}%</span>
                    </div>
                  </td>
                </tr>
              {/each}
            {/if}
          </tbody>
        </table>
      </div>
    {/if}

    <!-- TAB 3: CUT BUNDLES -->
    {#if activeTab === 'bundles'}
      <div class="table-card">
        <div class="table-card-head">
          <h3>Generated Cut Bundles</h3>
          <span class="sub-count">{activeJob?.cutBundles?.length || 0} bundle(s) active</span>
        </div>

        {#if activeJob?.cutBundles?.length > 0}
          <div class="bundles-grid">
            {#each activeJob.cutBundles as b}
              <div class="bundle-box">
                <div class="bundle-top">
                  <span class="bundle-seq">#{b.bundleSequence}</span>
                  <span class="size-pill">{b.size}</span>
                </div>
                <div class="bundle-id">{b.id}</div>
                <div class="bundle-pieces"><strong>{b.pieces}</strong> Pieces</div>
                <div class="bundle-next">Next: {b.nextProcess || 'Sewing Room'}</div>
                <button class="btn btn-secondary btn-sm btn-block" onclick={() => openSlip(b)}>
                  🖨️ Print Slip ({b.slipPrintedCount || 0})
                </button>
              </div>
            {/each}
          </div>
        {:else}
          <div class="empty-state">No bundles created yet. Record a cut lay to generate bundles.</div>
        {/if}
      </div>
    {/if}

    <!-- TAB 4: REMNANTS & BITS -->
    {#if activeTab === 'remnants'}
      <div class="remnant-grid">
        <div class="remnant-form-card">
          <h3>Return Roll Remnant or Bit</h3>
          <p class="section-desc">Store remaining fabric or flag damaged roll ends with Bit/Rej QR codes.</p>

          <form onsubmit={handleRecordRemnant}>
            <div class="form-group">
              <label for="cut-rem-roll-code">Original Roll Code <span class="req">*</span></label>
              <input
                id="cut-rem-roll-code"
                type="text"
                class="form-control"
                placeholder="e.g. 261019-A-1/4"
                bind:value={remnantRollCode}
                required
              />
            </div>

            <div class="form-row">
              <div class="form-group flex-1">
                <label for="cut-rem-type">Classification</label>
                <select id="cut-rem-type" class="form-control" bind:value={remnantType}>
                  <option value="Bit">Bit / Usable Remnant (Prefix: Bit)</option>
                  <option value="Rej">Reject / Defect Roll (Prefix: Rej)</option>
                </select>
              </div>

              <div class="form-group flex-1">
                <label for="cut-rem-weight">Returned Weight (Kg) <span class="req">*</span></label>
                <input
                  id="cut-rem-weight"
                  type="number"
                  step="0.01"
                  class="form-control"
                  placeholder="e.g. 3.45"
                  bind:value={remnantWeightKg}
                  required
                />
              </div>
            </div>

            <div class="form-group">
              <label for="cut-rem-location">Bin / Rack Location</label>
              <input
                id="cut-rem-location"
                type="text"
                class="form-control"
                placeholder="e.g. RACK-CUT-01"
                bind:value={remnantLocation}
              />
            </div>

            <div class="form-group">
              <label for="cut-rem-reason">Remarks or Defect Reason</label>
              <input
                id="cut-rem-reason"
                type="text"
                class="form-control"
                placeholder="Optional explanation..."
                bind:value={remnantReason}
              />
            </div>

            <button type="submit" class="btn btn-primary btn-block">
              ♻️ Record Remnant & Print Return QR
            </button>
          </form>
        </div>

        <div class="remnant-list-card">
          <h3>Stored Remnants ({activeJob?.remnants?.length || 0})</h3>
          {#if activeJob?.remnants?.length > 0}
            <div class="remnants-scroll">
              {#each activeJob.remnants as rem}
                <div class="remnant-item">
                  <span class="remnant-prefix prefix-{rem.prefix?.toLowerCase()}">{rem.prefix}</span>
                  <div class="remnant-details">
                    <div>Original: <strong>{rem.sourceRollCode || rem.originalRollCode}</strong></div>
                    <div>Weight: <strong>{rem.weightKg} kg</strong> • Location: <strong>{rem.binLocation || rem.rackLocation}</strong></div>
                  </div>
                </div>
              {/each}
            </div>
          {:else}
            <div class="empty-state">No remnants recorded for this cutting job.</div>
          {/if}
        </div>
      </div>
    {/if}
  {:else if !isLoading}
    <div class="empty-state large">
      <span>✂️</span>
      <h3>No Cutting Job Selected</h3>
      <p>Select an order from the queue above to start cutting.</p>
    </div>
  {/if}
</div>

<!-- Bundle Slip Print Modal -->
{#if isSlipModalOpen && selectedBundle}
  <div class="modal-overlay">
    <div class="slip-modal">
      <div class="slip-header">
        <h4>Bundle Route Slip</h4>
        <button class="btn btn-secondary btn-sm" onclick={() => isSlipModalOpen = false}>✕</button>
      </div>
      <div class="slip-body">
        <div class="slip-ticket">
          <div class="slip-ticket-head">
            <span class="slip-logo">⚡ FPCC SELVEDGE</span>
            <span class="slip-no">#{selectedBundle.id}</span>
          </div>
          <div class="slip-barcode-mock">|||||||||||||||||||||||||||||||||</div>
          <div class="slip-meta-grid">
            <div>Order: <strong>{selectedBundle.productionOrderNo || activeJob?.productionOrderNo}</strong></div>
            <div>Client: <strong>{selectedBundle.clientName || activeJob?.clientName}</strong></div>
            <div>Product: <strong>{activeJob?.productName}</strong></div>
            <div>Color: <strong>{activeJob?.colorName}</strong></div>
            <div class="slip-big-size">Size: <span>{selectedBundle.size}</span></div>
            <div class="slip-big-pcs">Qty: <span>{selectedBundle.pieces} pcs</span></div>
          </div>
          <div class="slip-footer">
            <span>Next: {selectedBundle.nextProcess || 'Sewing Assembly'}</span>
            <span>Printed: {new Date().toLocaleTimeString()}</span>
          </div>
        </div>
      </div>
      <div class="slip-actions">
        <button class="btn btn-secondary" onclick={() => window.print()}>🖨️ Print Direct</button>
        <button class="btn btn-primary" onclick={() => isSlipModalOpen = false}>Done</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .cutting-container {
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 16px;
    max-width: 1400px;
    margin: 0 auto;
  }

  .cutting-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: var(--bg-card);
    padding: 16px 20px;
    border-radius: 10px;
    border: 1px solid var(--border-color);
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .station-icon {
    font-size: 32px;
  }

  .cutting-header h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 700;
  }

  .subtitle {
    margin: 2px 0 0;
    font-size: 13px;
    color: var(--text-muted);
  }

  /* Job selector strip */
  .job-strip {
    display: flex;
    gap: 12px;
    overflow-x: auto;
    padding-bottom: 6px;
  }

  .job-card {
    flex: 0 0 260px;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    padding: 12px;
    text-align: left;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .job-card:hover {
    border-color: var(--accent-primary);
  }

  .job-card.active {
    border-color: var(--accent-primary);
    background: rgba(59, 130, 246, 0.08);
    box-shadow: 0 0 0 2px var(--accent-primary);
  }

  .job-card-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 4px;
  }

  .order-no {
    font-weight: 700;
    font-size: 14px;
  }

  .status-pill {
    font-size: 10px;
    font-weight: 600;
    padding: 2px 6px;
    border-radius: 4px;
    text-transform: uppercase;
  }

  .status-open {
    background: rgba(16, 185, 129, 0.15);
    color: #10b981;
  }

  .client-name {
    font-size: 12px;
    color: var(--text-muted);
  }

  .product-info {
    font-size: 12px;
    font-weight: 500;
    margin: 4px 0 8px;
  }

  .qty-labels {
    display: flex;
    justify-content: space-between;
    font-size: 11px;
    margin-bottom: 4px;
  }

  .balance {
    color: #f59e0b;
    font-weight: 600;
  }

  .progress-bar-bg {
    height: 6px;
    background: var(--bg-input, #e2e8f0);
    border-radius: 3px;
    overflow: hidden;
  }

  .progress-bar-fill {
    height: 100%;
    background: #10b981;
    transition: width 0.3s ease;
  }

  /* Active Job Details Banner */
  .active-banner {
    display: flex;
    flex-wrap: wrap;
    gap: 20px;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    padding: 12px 20px;
    border-radius: 8px;
  }

  .banner-col {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .banner-col .label {
    font-size: 11px;
    color: var(--text-muted);
  }

  .banner-col .value {
    font-size: 13px;
    font-weight: 600;
  }

  .badge-pharma {
    display: inline-block;
    background: rgba(139, 92, 246, 0.15);
    color: #8b5cf6;
    padding: 1px 6px;
    border-radius: 4px;
    font-size: 11px;
  }

  /* Tabs */
  .cutting-tabs {
    display: flex;
    gap: 8px;
    border-bottom: 1px solid var(--border-color);
    padding-bottom: 8px;
  }

  .tab-btn {
    background: transparent;
    border: 1px solid transparent;
    padding: 8px 16px;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 600;
    color: var(--text-muted);
    cursor: pointer;
  }

  .tab-btn:hover {
    color: var(--text-color);
  }

  .tab-btn.active {
    background: var(--bg-card);
    border-color: var(--border-color);
    color: var(--accent-primary);
  }

  /* Lay Grid */
  .lay-grid {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: 16px;
  }

  .lay-form-card, .lay-history-card, .table-card, .remnant-form-card, .remnant-list-card {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    padding: 18px;
  }

  .lay-form-card h3, .lay-history-card h3, .table-card h3, .remnant-form-card h3, .remnant-list-card h3 {
    margin: 0 0 4px;
    font-size: 16px;
  }

  .section-desc {
    font-size: 12px;
    color: var(--text-muted);
    margin: 0 0 16px;
  }

  .form-row {
    display: flex;
    gap: 12px;
  }

  .flex-1 {
    flex: 1;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-bottom: 14px;
  }

  .form-group label, .form-group .field-label {
    font-size: 12px;
    font-weight: 600;
  }

  .req {
    color: #ef4444;
  }

  .form-control {
    background: var(--bg-input, #ffffff);
    border: 1px solid var(--border-color);
    color: var(--text-color);
    padding: 8px 12px;
    border-radius: 6px;
    font-size: 13px;
  }

  .input-with-button {
    display: flex;
    gap: 6px;
  }

  .select-compact {
    max-width: 180px;
    font-size: 12px;
  }

  /* Plies Control */
  .plies-control {
    display: flex;
    gap: 6px;
    align-items: center;
  }

  .plies-input {
    width: 80px;
    text-align: center;
    font-size: 16px;
    font-weight: 700;
  }

  .btn-touch {
    background: var(--bg-input, #e2e8f0);
    border: 1px solid var(--border-color);
    color: var(--text-color);
    padding: 8px 12px;
    border-radius: 6px;
    font-weight: 600;
    cursor: pointer;
  }

  .btn-touch:hover {
    background: var(--border-color);
  }

  /* Ratio Matrix */
  .ratio-matrix {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
    background: rgba(0, 0, 0, 0.05);
    padding: 12px;
    border-radius: 6px;
    border: 1px solid var(--border-color);
  }

  .ratio-cell {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    width: 65px;
  }

  .ratio-size-title {
    font-size: 12px;
    font-weight: 700;
  }

  .ratio-input {
    width: 100%;
    text-align: center;
    font-size: 14px;
    font-weight: 600;
    padding: 6px 4px;
  }

  .ratio-sub {
    font-size: 10px;
    color: var(--text-muted);
  }

  /* Summary Box */
  .lay-summary-box {
    display: flex;
    gap: 14px;
    margin: 16px 0;
    padding: 12px;
    background: rgba(59, 130, 246, 0.08);
    border: 1px solid rgba(59, 130, 246, 0.2);
    border-radius: 6px;
  }

  .stat-pill {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .stat-lbl {
    font-size: 12px;
    color: var(--text-muted);
  }

  .stat-val {
    font-size: 14px;
    font-weight: 700;
  }

  .stat-val.highlight {
    font-size: 16px;
    color: var(--accent-primary);
  }

  /* Buttons */
  .btn {
    padding: 8px 14px;
    border-radius: 6px;
    font-weight: 600;
    font-size: 13px;
    cursor: pointer;
    border: none;
    transition: all 0.15s ease;
  }

  .btn-primary {
    background: var(--accent-primary, #3b82f6);
    color: #ffffff;
  }

  .btn-primary:hover {
    filter: brightness(1.1);
  }

  .btn-secondary {
    background: var(--bg-input, #e2e8f0);
    border: 1px solid var(--border-color);
    color: var(--text-color);
  }

  .btn-block {
    width: 100%;
  }

  .btn-lg {
    padding: 14px 20px;
    font-size: 15px;
    letter-spacing: 0.5px;
  }

  .btn-sm {
    padding: 4px 10px;
    font-size: 12px;
  }

  /* History scroll */
  .lays-scroll, .remnants-scroll {
    display: flex;
    flex-direction: column;
    gap: 10px;
    max-height: 480px;
    overflow-y: auto;
  }

  .lay-history-item {
    border: 1px solid var(--border-color);
    border-radius: 6px;
    padding: 10px;
  }

  .lay-item-head {
    display: flex;
    justify-content: space-between;
    margin-bottom: 4px;
  }

  .lay-badge {
    font-weight: 700;
    font-size: 12px;
  }

  .lay-time {
    font-size: 11px;
    color: var(--text-muted);
  }

  .lay-item-body {
    font-size: 12px;
    color: var(--text-muted);
  }

  /* Table */
  .data-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
  }

  .data-table th, .data-table td {
    padding: 10px 12px;
    text-align: left;
    border-bottom: 1px solid var(--border-color);
  }

  .size-tag {
    display: inline-block;
    padding: 2px 8px;
    background: rgba(0, 0, 0, 0.05);
    border-radius: 4px;
  }

  .progress-container {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .pct-text {
    font-size: 11px;
    color: var(--text-muted);
    min-width: 35px;
  }

  /* Bundles Grid */
  .bundles-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
    gap: 12px;
    margin-top: 12px;
  }

  .bundle-box {
    border: 1px solid var(--border-color);
    border-radius: 6px;
    padding: 12px;
    background: rgba(0, 0, 0, 0.02);
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .bundle-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .bundle-seq {
    font-weight: 700;
  }

  .size-pill {
    background: #3b82f6;
    color: white;
    padding: 1px 6px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 700;
  }

  .bundle-id {
    font-family: monospace;
    font-size: 10px;
    color: var(--text-muted);
  }

  .bundle-pieces {
    font-size: 13px;
  }

  .bundle-next {
    font-size: 11px;
    color: var(--text-muted);
    margin-bottom: 4px;
  }

  /* Remnants */
  .remnant-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
  }

  .remnant-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px;
    border: 1px solid var(--border-color);
    border-radius: 6px;
  }

  .remnant-prefix {
    font-weight: 700;
    font-size: 11px;
    padding: 3px 6px;
    border-radius: 4px;
    text-transform: uppercase;
  }

  .prefix-bit {
    background: rgba(16, 185, 129, 0.15);
    color: #10b981;
  }

  .prefix-rej {
    background: rgba(239, 68, 68, 0.15);
    color: #ef4444;
  }

  .remnant-details {
    font-size: 12px;
  }

  .empty-state {
    padding: 30px;
    text-align: center;
    color: var(--text-muted);
    font-size: 13px;
  }

  .empty-state.large {
    padding: 60px 20px;
  }

  .empty-state.large span {
    font-size: 48px;
  }

  /* Modal */
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
  }

  .slip-modal {
    background: var(--bg-card);
    border-radius: 8px;
    width: 380px;
    border: 1px solid var(--border-color);
    box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
    overflow: hidden;
  }

  .slip-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-color);
  }

  .slip-body {
    padding: 16px;
  }

  .slip-ticket {
    border: 2px dashed #94a3b8;
    padding: 14px;
    border-radius: 6px;
    background: #ffffff;
    color: #0f172a;
    font-family: monospace;
  }

  .slip-ticket-head {
    display: flex;
    justify-content: space-between;
    font-size: 12px;
    font-weight: 700;
    border-bottom: 1px solid #cbd5e1;
    padding-bottom: 6px;
    margin-bottom: 8px;
  }

  .slip-barcode-mock {
    text-align: center;
    letter-spacing: 2px;
    font-weight: 700;
    margin: 8px 0;
  }

  .slip-meta-grid {
    font-size: 11px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .slip-big-size, .slip-big-pcs {
    font-size: 14px;
    margin-top: 4px;
  }

  .slip-big-size span, .slip-big-pcs span {
    font-size: 18px;
    font-weight: 800;
  }

  .slip-footer {
    display: flex;
    justify-content: space-between;
    font-size: 10px;
    border-top: 1px solid #cbd5e1;
    margin-top: 8px;
    padding-top: 4px;
    color: #64748b;
  }

  .slip-actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    padding: 12px 16px;
    border-top: 1px solid var(--border-color);
  }

  .text-success { color: #10b981; }
  .text-warning { color: #f59e0b; }
  .text-muted { color: var(--text-muted); }
</style>
