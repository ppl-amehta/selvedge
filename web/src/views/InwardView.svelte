<script>
  import { onMount } from 'svelte';
  import { api } from '../api.js';
  import Combobox from '../components/Combobox.svelte';
  import QRTagModal from '../components/QRTagModal.svelte';
  import RemnantModal from '../components/RemnantModal.svelte';
  import DiscrepancyModal from '../components/DiscrepancyModal.svelte';

  let { masters = {} } = $props();

  let parentCode = $state('');
  let inwardDate = $state(new Date().toISOString().split('T')[0]);
  let vendorName = $state('');
  let vendorGstin = $state('');
  let vendorAddress = $state('');
  let invoiceNo = $state('');
  let inwardType = $state('Fabric');
  let uom = $state('Kgs');

  let invoiceQty = $state('');
  let invoiceAmount = $state('');
  let physicalReceivedQty = $state('');

  // Contextual fabric specs
  let vendorFabName = $state('');
  let fpccFabName = $state('100% Cotton Single Jersey 180 GSM');
  let vendorColor = $state('');
  let fpccColor = $state('Jet Black');
  let invoiceGsm = $state(180);
  let dia = $state(60);
  let dockRemarks = $state('');

  // Step B Bundles
  let bundleCountInput = $state(4);
  let bundles = $state([]);

  // Modals
  let isQRModalOpen = $state(false);
  let qrModalData = $state({});
  let isRemnantModalOpen = $state(false);
  let isDiscrepancyModalOpen = $state(false);
  let isSubmitting = $state(false);

  onMount(async () => {
    await refreshParentCode();
  });

  async function refreshParentCode() {
    try {
      parentCode = await api.getNextParentCode();
    } catch (err) {
      console.error(err);
    }
  }

  function handleVendorSelect(v) {
    if (v.gstin) vendorGstin = v.gstin;
    if (v.address) vendorAddress = v.address;
  }

  // Reactive Discrepancy Sensor via Svelte 5 $derived
  let discrepancyInfo = $derived.by(() => {
    const invQ = parseFloat(invoiceQty) || 0.0;
    const phyQ = parseFloat(physicalReceivedQty) || 0.0;
    const invAmt = parseFloat(invoiceAmount) || 0.0;

    if (invQ <= 0 || phyQ <= 0) return null;

    const variance = phyQ - invQ;
    const variancePct = ((variance / invQ) * 100).toFixed(2);
    const unitRate = invAmt > 0 ? (invAmt / invQ) : 0.0;
    const financialExposure = Math.abs(variance * unitRate);

    if (variance < -0.05) {
      return {
        type: 'danger',
        title: `🚨 Dock Shortfall Discrepancy: ${Math.abs(variance).toFixed(2)} Kgs (${Math.abs(variancePct)}% Shortage)`,
        desc: `Physical scale weight is less than billed supplier invoice. Financial exposure: ₹${financialExposure.toLocaleString('en-IN', { minimumFractionDigits: 2 })}. ⚡ Rippling Automation: System will draft Debit Note & freeze supplier payout on commit.`
      };
    } else if (variance > 0.05) {
      return {
        type: 'warning',
        title: `⚠️ Dock Excess Received: +${variance.toFixed(2)} Kgs (+${variancePct}%)`,
        desc: `Physical received exceeds billed supplier invoice. Review delivery challan and reconcile invoice.`
      };
    } else {
      return {
        type: 'success',
        title: `✓ Perfect Balance: Physical weight matches invoice billed quantity exactly`,
        desc: `Ready for roll breakdown and QR matrix tagging.`
      };
    }
  });

  // Reactive Bundle Tally via Svelte 5 $derived
  let bundleTally = $derived.by(() => {
    let sum = 0;
    for (const b of bundles) {
      sum += parseFloat(b.weight) || 0.0;
    }
    const phy = parseFloat(physicalReceivedQty) || 0.0;
    const diff = (sum - phy).toFixed(2);
    return { sum, phy, diff, count: bundles.length };
  });

  function autoGenerateBundles() {
    const count = parseInt(bundleCountInput, 10) || 1;
    const phy = parseFloat(physicalReceivedQty) || 0.0;
    const defaultWeight = count > 0 && phy > 0 ? parseFloat((phy / count).toFixed(2)) : 0.0;
    const pCode = parentCode || '261019-A';

    const list = [];
    for (let i = 1; i <= count; i++) {
      let w = defaultWeight;
      if (i === count && phy > 0) {
        const sumPrev = defaultWeight * (count - 1);
        w = parseFloat((phy - sumPrev).toFixed(2));
      }
      list.push({
        sequence: i,
        total: count,
        bundleCode: `${pCode}-${i}/${count}`,
        weight: w,
        gsm: 180,
        location: `A/R01/S0${((i - 1) % 6) + 1}`,
        qcStatus: 'Passed',
        remarks: ''
      });
    }
    bundles = list;
  }

  function addBundleRow() {
    const seq = bundles.length + 1;
    const pCode = parentCode || '261019-A';
    bundles = [
      ...bundles,
      {
        sequence: seq,
        total: seq,
        bundleCode: `${pCode}-${seq}/${seq}`,
        weight: 0,
        gsm: 180,
        location: 'A/R01/S01',
        qcStatus: 'Passed',
        remarks: ''
      }
    ];
  }

  function removeBundle(idx) {
    bundles = bundles.filter((_, i) => i !== idx);
  }

  function openQRPreview(bundle) {
    qrModalData = {
      bundleCode: bundle.bundleCode,
      item: inwardType === 'Fabric' ? fpccFabName : inwardType,
      color: inwardType === 'Fabric' ? fpccColor : '',
      qty: bundle.weight,
      uom,
      location: bundle.location,
      qcStatus: bundle.qcStatus
    };
    isQRModalOpen = true;
  }

  async function handleSubmit(e) {
    e.preventDefault();
    if (!parentCode || !vendorName.trim() || !invoiceNo.trim() || !invoiceQty || !physicalReceivedQty) {
      if (window.showToast) window.showToast('Please fill all mandatory dock inward fields (Vendor, Invoice, Quantities)', 'error');
      return;
    }

    if (bundles.length === 0) {
      if (window.showToast) window.showToast('Please break down lot into at least 1 roll bundle in Step B', 'error');
      return;
    }

    isSubmitting = true;
    const bundlePayloads = bundles.map(b => ({
      bundleSequence: b.sequence,
      totalBundles: bundles.length,
      itemType: inwardType,
      itemName: inwardType === 'Fabric' ? fpccFabName : inwardType,
      colorName: inwardType === 'Fabric' ? fpccColor : '',
      inwardWeightQty: parseFloat(b.weight) || 0.0,
      measuredGsm: parseFloat(b.gsm) || 0.0,
      locationBin: b.location || 'A/R01/S01',
      qcStatus: b.qcStatus,
      qcRemarks: b.remarks
    }));

    const payload = {
      parentCode,
      inwardDate,
      vendorId: vendorName,
      vendorName,
      invoiceNo,
      inwardType,
      uom,
      invoiceQty: parseFloat(invoiceQty) || 0.0,
      invoiceAmount: parseFloat(invoiceAmount) || 0.0,
      physicalReceivedQty: parseFloat(physicalReceivedQty) || 0.0,
      fabricType: inwardType === 'Fabric' ? 'Knitted' : '',
      vendorFabName,
      fpccFabName,
      vendorColor,
      fpccColor,
      invoiceGsm: parseFloat(invoiceGsm) || 0.0,
      dia: parseFloat(dia) || 0.0,
      remarks: dockRemarks,
      bundles: bundlePayloads
    };

    try {
      const res = await api.inwardParentLot(payload);
      if (window.showToast) {
        window.showToast(`✓ Inward Lot ${parentCode} committed into Unified Graph! (${bundles.length} rolls tagged)`, 'success');
      }

      // Reset
      invoiceNo = '';
      invoiceQty = '';
      invoiceAmount = '';
      physicalReceivedQty = '';
      dockRemarks = '';
      bundles = [];
      await refreshParentCode();
    } catch (err) {
      if (window.showToast) {
        window.showToast(`Inward failed: ${err.message}`, 'error');
      }
    } finally {
      isSubmitting = false;
    }
  }
</script>

<div class="view-header" style="display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:24px;">
  <div>
    <h2 style="font-size:22px;font-weight:800;color:#fff;">Station 1: Raw Material Inward & Multi-Bundle QR</h2>
    <p style="font-size:13px;color:var(--text-muted);margin-top:4px;">
      Dock Receiving Scale, Discrepancy Monitoring & Child Roll Barcode Breakdown
    </p>
  </div>
  <div style="display:flex;gap:10px;">
    <button type="button" class="btn btn-secondary btn-sm" onclick={() => isRemnantModalOpen = true}>
      🔄 Return Remnant / Bit / Reject
    </button>
    <button type="button" class="btn btn-secondary btn-sm" onclick={() => isDiscrepancyModalOpen = true}>
      ⚠️ View Discrepancy Log
    </button>
    <button type="button" class="btn btn-secondary btn-sm" onclick={refreshParentCode}>
      🔄 Refresh Code
    </button>
  </div>
</div>

<form onsubmit={handleSubmit}>
  <!-- Step A: Dock Inwarding -->
  <div class="card-panel">
    <div class="panel-header">
      <div class="panel-title">Step A: Gate Receipt & Commercial Dock Inward</div>
      <div class="system-badge" style="font-size:13px;">{parentCode || 'FETCHING...'}</div>
    </div>

    <div style="display:grid;grid-template-columns:repeat(4, 1fr);gap:16px;">
      <div class="form-group">
        <label class="form-label">Parent Lot Code <span class="required">*</span></label>
        <input type="text" class="input-control input-mono" value={parentCode} readonly />
      </div>
      <div class="form-group">
        <label class="form-label">Inward Date <span class="required">*</span></label>
        <input type="date" class="input-control" bind:value={inwardDate} />
      </div>
      <div class="form-group">
        <label class="form-label">Vendor Invoice No <span class="required">*</span></label>
        <input type="text" class="input-control input-mono" placeholder="INV-2026-XXXX" bind:value={invoiceNo} />
      </div>
      <div class="form-group">
        <label class="form-label">Unit of Measure (UOM) <span class="required">*</span></label>
        <Combobox
          category="uom"
          items={masters.uoms || []}
          bind:value={uom}
          placeholder="Select UOM..."
        />
      </div>
    </div>

    <div style="display:grid;grid-template-columns:repeat(3, 1fr);gap:16px;margin-top:8px;">
      <div class="form-group">
        <label class="form-label">Vendor Name <span class="required">*</span> <span class="helper">Auto-registers</span></label>
        <Combobox
          category="vendor"
          items={masters.vendors || []}
          displayKey="name"
          valueKey="name"
          placeholder="Search 74+ vendors or add new..."
          bind:value={vendorName}
          onSelect={handleVendorSelect}
        />
      </div>
      <div class="form-group">
        <label class="form-label">Vendor GSTIN</label>
        <input type="text" class="input-control input-mono" placeholder="Auto-populated" value={vendorGstin} readonly />
      </div>
      <div class="form-group">
        <label class="form-label">Vendor Location</label>
        <input type="text" class="input-control" placeholder="Auto-populated" value={vendorAddress} readonly />
      </div>
    </div>

    <div class="form-group" style="margin-top:8px;">
      <label class="form-label">Inward Item Type (Strict Enum) <span class="required">*</span></label>
      <div class="badge-select-wrapper">
        {#each ['Fabric', 'Zips', 'Packaging', 'Trims', 'Labels', 'Finished Goods'] as type}
          <span
            class="badge-option"
            class:selected={inwardType === type}
            onclick={() => inwardType = type}
          >
            {type}
          </span>
        {/each}
      </div>
    </div>

    <!-- Commercials & Weight Discrepancy Reconciliation -->
    <hr style="border:none;border-top:1px solid var(--border-subtle);margin:16px 0;" />
    <div style="display:grid;grid-template-columns:repeat(3, 1fr);gap:16px;">
      <div class="form-group">
        <label class="form-label">Billed Invoice Qty <span class="required">*</span></label>
        <input
          type="number"
          class="input-control numeric"
          step="0.01"
          min="0"
          placeholder="0.00"
          bind:value={invoiceQty}
        />
      </div>
      <div class="form-group">
        <label class="form-label">Total Invoice Value (₹)</label>
        <input
          type="number"
          class="input-control numeric"
          step="0.5"
          min="0"
          placeholder="0.00"
          bind:value={invoiceAmount}
        />
      </div>
      <div class="form-group">
        <label class="form-label">Physical Scale Weight / Qty <span class="required">*</span></label>
        <input
          type="number"
          class="input-control numeric"
          step="0.01"
          min="0"
          placeholder="0.00"
          bind:value={physicalReceivedQty}
        />
      </div>
    </div>

    <!-- Live Discrepancy Banner -->
    {#if discrepancyInfo}
      <div class="banner-alert {discrepancyInfo.type}" style="margin-top:14px;">
        <div>
          <div class="banner-title" style="font-weight:800;font-size:13.5px;">{discrepancyInfo.title}</div>
          <div class="banner-desc" style="font-size:12px;margin-top:4px;opacity:0.9;">{discrepancyInfo.desc}</div>
        </div>
      </div>
    {/if}

    <!-- Fabric Contextual Specs -->
    {#if inwardType === 'Fabric'}
      <div style="margin-top:16px;background:rgba(255,255,255,0.015);border:1px solid var(--border-subtle);border-radius:var(--radius-md);padding:16px;">
        <div style="font-size:12px;font-weight:800;text-transform:uppercase;color:var(--text-muted);margin-bottom:12px;">
          Fabric Technical Parameters
        </div>
        <div style="display:grid;grid-template-columns:repeat(3, 1fr);gap:16px;">
          <div class="form-group">
            <label class="form-label">Vendor Fabric Name</label>
            <input type="text" class="input-control" placeholder="Mill trade name" bind:value={vendorFabName} />
          </div>
          <div class="form-group">
            <label class="form-label">FPCC Standard Fabric <span class="required">*</span></label>
            <Combobox
              category="fabric"
              items={masters.fabrics || []}
              bind:value={fpccFabName}
              placeholder="Search 176+ fabrics..."
            />
          </div>
          <div class="form-group">
            <label class="form-label">Invoice GSM</label>
            <input type="number" class="input-control numeric" placeholder="e.g. 180" bind:value={invoiceGsm} />
          </div>
        </div>
        <div style="display:grid;grid-template-columns:repeat(3, 1fr);gap:16px;margin-top:8px;">
          <div class="form-group">
            <label class="form-label">Vendor Color Reference</label>
            <input type="text" class="input-control" placeholder="e.g. Shade 42B" bind:value={vendorColor} />
          </div>
          <div class="form-group">
            <label class="form-label">FPCC Standard Color <span class="required">*</span></label>
            <Combobox
              category="color"
              items={masters.colors || []}
              bind:value={fpccColor}
              placeholder="Search 289+ colors..."
            />
          </div>
          <div class="form-group">
            <label class="form-label">Dia / Width (Inches)</label>
            <input type="number" class="input-control numeric" placeholder="e.g. 60" bind:value={dia} />
          </div>
        </div>
      </div>
    {/if}

    <div class="form-group" style="margin-top:12px;">
      <label class="form-label">Dock Receiver Remarks / Lorry Receipt (LR) No</label>
      <input type="text" class="input-control" placeholder="e.g. LR# 94821 via TCI Express, rolls shrink-wrapped" bind:value={dockRemarks} />
    </div>
  </div>

  <!-- Step B: Bundle Breakdown & QR Generator -->
  <div class="card-panel">
    <div class="panel-header">
      <div class="panel-title">Step B: Physical Bundles Breakdown & Zebra QR Tagging</div>
      <div style="font-family:var(--font-mono);font-size:12.5px;">
        {#if bundleTally.count > 0}
          {#if Math.abs(parseFloat(bundleTally.diff)) < 0.01 && bundleTally.phy > 0}
            <span style="color:var(--accent-emerald);">✓ Perfect Breakdown: {bundleTally.sum.toFixed(2)} / {bundleTally.phy.toFixed(2)} Kgs ({bundleTally.count} Rolls)</span>
          {:else if parseFloat(bundleTally.diff) > 0}
            <span style="color:var(--accent-rose);">⚠ Over Breakdown: {bundleTally.sum.toFixed(2)} / {bundleTally.phy.toFixed(2)} Kgs (+{bundleTally.diff} Kgs)</span>
          {:else}
            <span style="color:var(--accent-amber);">Sum: {bundleTally.sum.toFixed(2)} / {bundleTally.phy.toFixed(2)} Kgs (Remaining: {Math.abs(parseFloat(bundleTally.diff)).toFixed(2)} Kgs)</span>
          {/if}
        {/if}
      </div>
    </div>

    <div style="display:flex;align-items:center;gap:16px;margin-bottom:16px;">
      <div style="width:200px;">
        <label class="form-label">Number of Rolls</label>
        <input type="number" class="input-control numeric" min="1" max="50" bind:value={bundleCountInput} />
      </div>
      <button type="button" class="btn btn-secondary" onclick={autoGenerateBundles} style="margin-top:22px;">
        ⚡ Auto-Generate Rolls Breakdown
      </button>
      <button type="button" class="btn btn-secondary" onclick={addBundleRow} style="margin-top:22px;">
        + Add 1 Roll Row
      </button>
    </div>

    <div class="data-table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th style="width:160px;">Bundle QR Code</th>
            <th class="numeric" style="width:130px;">Weight ({uom})</th>
            <th class="numeric" style="width:120px;">Measured GSM</th>
            <th>Warehouse Bin Location</th>
            <th style="width:130px;">QC Status</th>
            <th>QC Notes</th>
            <th style="width:120px;text-align:right;">QR Tag</th>
            <th style="width:40px;"></th>
          </tr>
        </thead>
        <tbody>
          {#each bundles as b, idx}
            <tr>
              <td class="input-mono"><strong>{b.bundleCode}</strong></td>
              <td>
                <input
                  type="number"
                  class="input-control numeric"
                  step="0.01"
                  min="0"
                  bind:value={b.weight}
                />
              </td>
              <td>
                <input
                  type="number"
                  class="input-control numeric"
                  placeholder="GSM"
                  bind:value={b.gsm}
                />
              </td>
              <td>
                <input
                  type="text"
                  class="input-control input-mono"
                  placeholder="Bin Code"
                  bind:value={b.location}
                />
              </td>
              <td>
                <select class="select-control" bind:value={b.qcStatus}>
                  <option value="Passed">Passed</option>
                  <option value="Rejected">Rejected</option>
                  <option value="Pending">Pending</option>
                </select>
              </td>
              <td>
                <input
                  type="text"
                  class="input-control"
                  placeholder="Roll condition"
                  bind:value={b.remarks}
                />
              </td>
              <td style="text-align:right;">
                <button
                  type="button"
                  class="btn btn-secondary btn-sm"
                  onclick={() => openQRPreview(b)}
                >
                  Preview QR
                </button>
              </td>
              <td>
                <button
                  type="button"
                  class="btn btn-danger btn-sm"
                  onclick={() => removeBundle(idx)}
                >
                  &times;
                </button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </div>

  <!-- Submit Bar -->
  <div style="display:flex;justify-content:flex-end;gap:16px;margin-bottom:48px;">
    <button
      type="submit"
      class="btn btn-success"
      disabled={isSubmitting}
      style="padding:14px 32px;font-size:15px;font-weight:800;"
    >
      {isSubmitting ? '⏳ COMMITTING TO GRAPH...' : '📦 SAVE INWARD & COMMIT BUNDLES TO GRAPH'}
    </button>
  </div>
</form>

<QRTagModal
  bind:isOpen={isQRModalOpen}
  bundleCode={qrModalData.bundleCode}
  item={qrModalData.item}
  color={qrModalData.color}
  qty={qrModalData.qty}
  uom={qrModalData.uom}
  location={qrModalData.location}
  qcStatus={qrModalData.qcStatus}
/>

<RemnantModal
  bind:isOpen={isRemnantModalOpen}
  onReturnProcessed={(data) => {
    console.log('Return processed:', data);
  }}
/>

<DiscrepancyModal
  bind:isOpen={isDiscrepancyModalOpen}
/>
