// FPCC Selvedge - Station 1: Inward & QR Bundle Breakdown View
import { api } from '../api.js';
import { createCombobox } from '../components/combobox.js';
import { generateQRTagSVG } from '../components/qr.js';

export function initInwardView({ container, masters, onInwardCompleted }) {
  let currentMasters = masters;
  let inwardType = 'Fabric';
  let parentCode = '';
  let bundles = [];

  // Combobox references
  let vendorCombobox;
  let uomCombobox;
  let fpccFabCombobox;
  let fpccColCombobox;

  render();

  async function render() {
    container.innerHTML = `
      <div class="view-header">
        <div class="view-title-group">
          <h2>Station 1: Raw Material Inward & Multi-Bundle QR</h2>
          <p>Dock receipt, contextual material specs, physical rolls breakdown, and QR barcode labeling</p>
        </div>
        <div class="header-badges">
          <button class="btn btn-secondary btn-sm" id="btn-open-remnant-modal">🔄 Return Remnant / Bit / Reject</button>
          <button class="btn btn-secondary btn-sm" id="btn-open-discrepancies-modal">⚠️ View Discrepancy Log</button>
          <button class="btn btn-secondary btn-sm" id="btn-refresh-parent-code">🔄 Refresh Code</button>
        </div>
      </div>

      <form id="inward-form">
        <!-- Step A: Dock Inwarding -->
        <div class="card-panel">
          <div class="panel-header">
            <div class="panel-title">Step A: Gate Receipt & Commercial Inward</div>
            <div id="badge-parent-code" class="system-badge" style="font-size:14px;font-weight:700;">FETCHING...</div>
          </div>

          <div class="grid-4">
            <div class="form-group">
              <label class="form-label">Parent Lot Code <span class="required">*</span></label>
              <input type="text" id="inp-parent-code" class="input-control input-mono" readonly />
            </div>
            <div class="form-group">
              <label class="form-label">Inward Date <span class="required">*</span></label>
              <input type="date" id="inp-inward-date" class="input-control" />
            </div>
            <div class="form-group">
              <label class="form-label">Vendor Invoice No <span class="required">*</span></label>
              <input type="text" id="inp-invoice-no" class="input-control input-mono" placeholder="INV-2026-XXXX" />
            </div>
            <div class="form-group">
              <label class="form-label">Unit of Measure (UOM) <span class="required">*</span></label>
              <div id="cb-uom"></div>
            </div>
          </div>

          <div class="grid-3" style="margin-top:8px;">
            <div class="form-group">
              <label class="form-label">Vendor Name <span class="required">*</span> <span class="helper">Auto-registers</span></label>
              <div id="cb-vendor"></div>
            </div>
            <div class="form-group">
              <label class="form-label">Vendor GSTIN</label>
              <input type="text" id="inp-vendor-gstin" class="input-control input-mono" readonly placeholder="Auto-populated" />
            </div>
            <div class="form-group">
              <label class="form-label">Vendor Location / City</label>
              <input type="text" id="inp-vendor-address" class="input-control" readonly placeholder="Auto-populated" />
            </div>
          </div>

          <div class="form-group" style="margin-top:8px;">
            <label class="form-label">Inward Item Type (Strict Enum) <span class="required">*</span></label>
            <div class="badge-select-wrapper" id="group-inward-type">
              <span class="badge-option selected" data-val="Fabric">Fabric (Knitted / Woven)</span>
              <span class="badge-option" data-val="Zips">Zips & Fasteners</span>
              <span class="badge-option" data-val="Packaging">Packaging Materials</span>
              <span class="badge-option" data-val="Trims">Trims & Elastics</span>
              <span class="badge-option" data-val="Labels">Labels & Tags</span>
              <span class="badge-option" data-val="Finished Goods">Finished Garments</span>
            </div>
          </div>

          <!-- Financials & Weight Balancing -->
          <hr style="border:none;border-top:1px solid var(--border-subtle);margin:16px 0;" />
          <div class="grid-3">
            <div class="form-group">
              <label class="form-label">Billed Invoice Qty <span class="required">*</span></label>
              <input type="number" id="inp-invoice-qty" class="input-control numeric" step="0.01" min="0" placeholder="0.00" />
            </div>
            <div class="form-group">
              <label class="form-label">Total Invoice Value (₹)</label>
              <input type="number" id="inp-invoice-amount" class="input-control numeric" step="0.5" min="0" placeholder="0.00" />
            </div>
            <div class="form-group">
              <label class="form-label">Physical Scale Received Qty <span class="required">*</span></label>
              <input type="number" id="inp-physical-qty" class="input-control numeric" step="0.01" min="0" placeholder="0.00" />
            </div>
          </div>

          <!-- Dynamic Discrepancy Alert Banner -->
          <div id="discrepancy-banner" style="display:none;margin-top:12px;"></div>

          <!-- Contextual Specs for Fabric -->
          <div id="fabric-context-panel" style="margin-top:16px;">
            <div class="panel-header" style="margin-bottom:12px;padding-bottom:8px;">
              <div class="panel-title" style="font-size:13px;">Fabric Technical Specifications</div>
            </div>
            <div class="grid-3">
              <div class="form-group">
                <label class="form-label">Vendor Fabric Name</label>
                <input type="text" id="inp-vendor-fab" class="input-control" placeholder="Mill trade name" />
              </div>
              <div class="form-group">
                <label class="form-label">FPCC Standard Fabric Name <span class="required">*</span></label>
                <div id="cb-fpcc-fab"></div>
              </div>
              <div class="form-group">
                <label class="form-label">Invoice GSM</label>
                <input type="number" id="inp-inv-gsm" class="input-control numeric" placeholder="e.g. 220" />
              </div>
            </div>
            <div class="grid-3">
              <div class="form-group">
                <label class="form-label">Vendor Color Reference</label>
                <input type="text" id="inp-vendor-col" class="input-control" placeholder="e.g. Shade 42B" />
              </div>
              <div class="form-group">
                <label class="form-label">FPCC Standard Color <span class="required">*</span></label>
                <div id="cb-fpcc-col"></div>
              </div>
              <div class="form-group">
                <label class="form-label">Dia / Width (Inches)</label>
                <input type="number" id="inp-dia" class="input-control numeric" placeholder="e.g. 60" />
              </div>
            </div>
          </div>

          <div class="form-group" style="margin-top:8px;">
            <label class="form-label">Dock Receiver Remarks / Lorry Receipt (LR) No</label>
            <input type="text" id="inp-dock-remarks" class="input-control" placeholder="e.g. LR# 94821 via TCI Express, roll wrappers intact" />
          </div>
        </div>

        <!-- Step B: Bundle Breakdown & QR Generator -->
        <div class="card-panel">
          <div class="panel-header">
            <div class="panel-title">Step B: Physical Bundles Breakdown & QR Code Tagging</div>
            <div id="bundle-tally" style="font-family:var(--font-mono);font-size:13px;color:var(--text-muted);">
              Tally: 0.00 / 0.00 Kgs (0 Bundles)
            </div>
          </div>

          <div style="display:flex;align-items:center;gap:16px;margin-bottom:16px;">
            <div style="width:200px;">
              <label class="form-label">Number of Rolls / Bundles</label>
              <input type="number" id="inp-bundle-count" class="input-control numeric" min="1" max="50" value="4" />
            </div>
            <button type="button" class="btn btn-secondary" id="btn-generate-bundles" style="margin-top:22px;">
              ⚡ Auto-Generate Bundle Breakdown
            </button>
            <button type="button" class="btn btn-secondary" id="btn-add-bundle-row" style="margin-top:22px;">
              + Add 1 Bundle Row
            </button>
          </div>

          <div class="data-table-container">
            <table class="data-table" id="bundle-table">
              <thead>
                <tr>
                  <th style="width:160px;">Bundle QR Code</th>
                  <th class="numeric" style="width:140px;">Weight / Qty</th>
                  <th class="numeric" style="width:120px;">Measured GSM</th>
                  <th>Storage Location / Bin</th>
                  <th style="width:130px;">QC Status</th>
                  <th>QC Remarks</th>
                  <th style="width:120px;text-align:right;">QR Tag</th>
                  <th style="width:40px;"></th>
                </tr>
              </thead>
              <tbody id="bundle-tbody">
                <!-- Dynamically populated bundle rows -->
              </tbody>
            </table>
          </div>
        </div>

        <!-- Submission Bar -->
        <div style="display:flex;justify-content:flex-end;gap:16px;margin-bottom:40px;">
          <button type="reset" class="btn btn-secondary">Reset Dock Form</button>
          <button type="submit" class="btn btn-success" id="btn-submit-inward" style="padding:12px 28px;font-size:15px;font-weight:700;">
            📦 SAVE INWARD & COMMIT BUNDLES TO GRAPH
          </button>
        </div>
      </form>

      <!-- Modal 1: Printable Factory QR Tag Modal -->
      <div class="modal-overlay" id="modal-qr-preview">
        <div class="modal-box" style="max-width:400px;">
          <div class="modal-header">
            <h3>Factory QR Label Tag</h3>
            <button class="btn btn-secondary btn-sm modal-close">&times;</button>
          </div>
          <div class="modal-body" id="qr-preview-body" style="text-align:center;padding:24px;">
            <!-- QR SVG tag inserted here -->
          </div>
          <div class="modal-footer" style="justify-content:space-between;">
            <button class="btn btn-secondary modal-close">Close</button>
            <button class="btn btn-primary" id="btn-print-tag">🖨️ Print Zebra Tag</button>
          </div>
        </div>
      </div>

      <!-- Modal 2: Return Remnant / Bit / Reject Modal -->
      <div class="modal-overlay" id="modal-remnant-return">
        <div class="modal-box">
          <div class="modal-header">
            <h3>Return Remnant / Bit / Reject Roll</h3>
            <button class="btn btn-secondary btn-sm modal-close">&times;</button>
          </div>
          <div class="modal-body">
            <p style="font-size:12px;color:var(--text-muted);margin-bottom:12px;">
              Returns leftover roll remnants or defect rolls from cutting. Automatically updates bundle code prefix (<strong>Bit</strong> or <strong>Rej</strong>) and emits smart workflow alerts:
            </p>
            <form id="remnant-form">
              <div class="grid-2">
                <div class="form-group">
                  <label class="form-label">Original Bundle Code <span class="required">*</span></label>
                  <input type="text" id="inp-ret-bundle-code" class="input-control input-mono" placeholder="e.g. 261019-A-1/4" required />
                </div>
                <div class="form-group">
                  <label class="form-label">Return Classification <span class="required">*</span></label>
                  <select class="select-control" id="sel-ret-type">
                    <option value="bits">Bits / Usable Remnant (Prefix: Bit)</option>
                    <option value="reject">Reject / Defect Roll (Prefix: Rej)</option>
                    <option value="remnant">Standard Remnant</option>
                  </select>
                </div>
              </div>
              <div class="grid-2">
                <div class="form-group">
                  <label class="form-label">Returned Net Weight (Kg) <span class="required">*</span></label>
                  <input type="number" id="inp-ret-weight" class="input-control numeric" step="0.01" min="0.01" placeholder="e.g. 4.85" required />
                </div>
                <div class="form-group">
                  <label class="form-label">Return Bin Location</label>
                  <input type="text" id="inp-ret-loc" class="input-control input-mono" placeholder="e.g. BIT-RACK-01" value="A/R01/S01" />
                </div>
              </div>
              <div class="form-group">
                <label class="form-label">Defect / Remnant Reason</label>
                <input type="text" id="inp-ret-remarks" class="input-control" placeholder="e.g. Lycra broken filament elastane drop at 18m" />
              </div>
              <div style="display:flex;justify-content:flex-end;margin-top:16px;">
                <button type="submit" class="btn btn-primary" id="btn-submit-remnant">Confirm Return & Generate New QR</button>
              </div>
            </form>
          </div>
        </div>
      </div>

      <!-- Modal 3: Discrepancy Log Modal -->
      <div class="modal-overlay" id="modal-discrepancies">
        <div class="modal-box" style="max-width:850px;">
          <div class="modal-header">
            <h3>Dock Inward Discrepancy & Shortage Register</h3>
            <button class="btn btn-secondary btn-sm modal-close">&times;</button>
          </div>
          <div class="modal-body">
            <div class="data-table-container">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>Report No</th>
                    <th>Parent Lot</th>
                    <th>Vendor</th>
                    <th>Invoice No</th>
                    <th class="numeric">Inv Qty</th>
                    <th class="numeric">Actual</th>
                    <th class="numeric">Variance</th>
                    <th class="numeric">Financial Loss</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody id="discrepancies-tbody">
                  <!-- Dynamically populated -->
                </tbody>
              </table>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary modal-close">Close</button>
          </div>
        </div>
      </div>
    `;

    setupControls();
    setupEventListeners();
    await refreshParentCode();
  }

  function setupControls() {
    const today = new Date().toISOString().split('T')[0];
    document.getElementById('inp-inward-date').value = today;

    // Comboboxes
    vendorCombobox = createCombobox({
      container: document.getElementById('cb-vendor'),
      category: 'vendor',
      placeholder: 'Search 74+ vendors or type new...',
      items: currentMasters.vendors || [],
      displayKey: 'name',
      valueKey: 'name',
      onSelect: (vendor) => {
        if (vendor.gstin) document.getElementById('inp-vendor-gstin').value = vendor.gstin;
        if (vendor.address) document.getElementById('inp-vendor-address').value = vendor.address;
      }
    });

    uomCombobox = createCombobox({
      container: document.getElementById('cb-uom'),
      category: 'uom',
      placeholder: 'Select UOM...',
      items: currentMasters.uoms || [],
      displayKey: 'display_name',
      valueKey: 'display_name',
      initialValue: 'Kgs'
    });

    fpccFabCombobox = createCombobox({
      container: document.getElementById('cb-fpcc-fab'),
      category: 'fabric',
      placeholder: 'Search 225+ fabrics...',
      items: currentMasters.fabrics || [],
      displayKey: 'display_name',
      valueKey: 'display_name',
      initialValue: '100% Cotton Single Jersey 180 GSM'
    });

    fpccColCombobox = createCombobox({
      container: document.getElementById('cb-fpcc-col'),
      category: 'color',
      placeholder: 'Search 301+ colors...',
      items: currentMasters.colors || [],
      displayKey: 'display_name',
      valueKey: 'display_name',
      initialValue: 'Jet Black'
    });
  }

  function setupEventListeners() {
    // Inward Type Badges
    const groupInward = document.getElementById('group-inward-type');
    groupInward.querySelectorAll('.badge-option').forEach(badge => {
      badge.addEventListener('click', () => {
        groupInward.querySelectorAll('.badge-option').forEach(b => b.classList.remove('selected'));
        badge.classList.add('selected');
        inwardType = badge.dataset.val;

        const fabPanel = document.getElementById('fabric-context-panel');
        if (inwardType === 'Fabric') {
          fabPanel.style.display = 'block';
        } else {
          fabPanel.style.display = 'none';
        }
      });
    });

    // Refresh Parent Code
    document.getElementById('btn-refresh-parent-code').addEventListener('click', refreshParentCode);

    // Quantity Discrepancy Monitoring
    document.getElementById('inp-invoice-qty').addEventListener('input', recalculateDiscrepancy);
    document.getElementById('inp-physical-qty').addEventListener('input', () => {
      recalculateDiscrepancy();
      updateBundleTally();
    });
    document.getElementById('inp-invoice-amount').addEventListener('input', recalculateDiscrepancy);

    // Bundle Generation
    document.getElementById('btn-generate-bundles').addEventListener('click', autoGenerateBundles);
    document.getElementById('btn-add-bundle-row').addEventListener('click', () => {
      addBundleRow(bundles.length + 1, bundles.length + 1, 0, 'A/R01/S01');
    });

    // Remnant Return Modal
    const modalRemnant = document.getElementById('modal-remnant-return');
    document.getElementById('btn-open-remnant-modal').addEventListener('click', () => {
      document.getElementById('remnant-form').reset();
      modalRemnant.classList.add('open');
    });

    document.getElementById('remnant-form').addEventListener('submit', handleRemnantSubmit);

    // Discrepancy Modal
    const modalDiscrepancies = document.getElementById('modal-discrepancies');
    document.getElementById('btn-open-discrepancies-modal').addEventListener('click', async () => {
      await loadDiscrepanciesModal();
      modalDiscrepancies.classList.add('open');
    });

    // Print Tag in QR preview
    document.getElementById('btn-print-tag').addEventListener('click', () => {
      window.print();
    });

    // Close Modals
    document.querySelectorAll('.modal-close').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.target.closest('.modal-overlay').classList.remove('open');
      });
    });

    // Form Submit
    document.getElementById('inward-form').addEventListener('submit', handleInwardSubmit);
  }

  async function refreshParentCode() {
    try {
      parentCode = await api.getNextParentCode();
      document.getElementById('inp-parent-code').value = parentCode;
      document.getElementById('badge-parent-code').textContent = parentCode;
    } catch (err) {
      console.error(err);
    }
  }

  function recalculateDiscrepancy() {
    const invQty = parseFloat(document.getElementById('inp-invoice-qty').value) || 0.0;
    const phyQty = parseFloat(document.getElementById('inp-physical-qty').value) || 0.0;
    const invAmt = parseFloat(document.getElementById('inp-invoice-amount').value) || 0.0;
    const banner = document.getElementById('discrepancy-banner');

    if (invQty <= 0 || phyQty <= 0) {
      banner.style.display = 'none';
      return;
    }

    const variance = phyQty - invQty;
    const variancePct = ((variance / invQty) * 100).toFixed(2);
    const unitRate = invAmt > 0 ? (invAmt / invQty) : 0.0;
    const financialImpact = Math.abs(variance * unitRate);

    banner.style.display = 'block';

    if (variance < -0.05) {
      // Shortage
      banner.className = 'banner-alert danger';
      banner.innerHTML = `
        <div>
          <div class="banner-title">🚨 Dock Shortfall Discrepancy Detected: ${Math.abs(variance).toFixed(2)} Kgs (${Math.abs(variancePct)}% Shortage)</div>
          <div class="banner-desc">
            Physical scale weight is less than supplier invoice. Financial exposure: <strong>₹${financialImpact.toLocaleString('en-IN', { minimumFractionDigits: 2 })}</strong>.<br/>
            ⚡ <em>Rippling Automation Armed: System will auto-generate Debit Note draft & freeze supplier payout on commit.</em>
          </div>
        </div>
      `;
    } else if (variance > 0.05) {
      // Excess
      banner.className = 'banner-alert warning';
      banner.innerHTML = `
        <div>
          <div class="banner-title">⚠️ Dock Excess Received: +${variance.toFixed(2)} Kgs (+${variancePct}%)</div>
          <div class="banner-desc">
            Physical received exceeds supplier invoice billed quantity. Review supplier reconciliation terms.
          </div>
        </div>
      `;
    } else {
      // Perfect Match
      banner.className = 'banner-alert success';
      banner.innerHTML = `
        <div>
          <div class="banner-title">✓ Perfect Balance: Physical weight matches invoice billed quantity exactly</div>
        </div>
      `;
    }
  }

  function autoGenerateBundles() {
    const count = parseInt(document.getElementById('inp-bundle-count').value, 10) || 1;
    const phyQty = parseFloat(document.getElementById('inp-physical-qty').value) || 0.0;
    const defaultWeight = count > 0 && phyQty > 0 ? parseFloat((phyQty / count).toFixed(2)) : 0.0;

    const tbody = document.getElementById('bundle-tbody');
    tbody.innerHTML = '';
    bundles = [];

    const pCode = document.getElementById('inp-parent-code').value.trim() || '261019-A';

    for (let i = 1; i <= count; i++) {
      // Distribute any remainder to the last bundle
      let weight = defaultWeight;
      if (i === count && phyQty > 0) {
        const sumPrevious = defaultWeight * (count - 1);
        weight = parseFloat((phyQty - sumPrevious).toFixed(2));
      }
      addBundleRow(i, count, weight, `A/R01/S0${((i - 1) % 6) + 1}`, pCode);
    }

    updateBundleTally();
  }

  function addBundleRow(seq, total, weight = 0, loc = 'A/R01/S01', pCode = '') {
    const parent = pCode || document.getElementById('inp-parent-code').value.trim() || '261019-A';
    const bundleCode = `${parent}-${seq}/${total}`;
    const tbody = document.getElementById('bundle-tbody');

    const tr = document.createElement('tr');
    tr.dataset.bundleCode = bundleCode;
    tr.innerHTML = `
      <td class="input-mono"><strong>${bundleCode}</strong></td>
      <td><input type="number" class="input-control numeric bundle-weight" step="0.01" min="0" value="${weight}" /></td>
      <td><input type="number" class="input-control numeric bundle-gsm" placeholder="GSM" value="180" /></td>
      <td><input type="text" class="input-control input-mono bundle-loc" value="${loc}" placeholder="Bin Code" /></td>
      <td>
        <select class="select-control bundle-qc">
          <option value="Passed" selected>Passed</option>
          <option value="Rejected">Rejected</option>
          <option value="Pending">Pending</option>
        </select>
      </td>
      <td><input type="text" class="input-control bundle-qc-remarks" placeholder="Notes (rolls condition)" /></td>
      <td style="text-align:right;">
        <button type="button" class="btn btn-secondary btn-sm btn-preview-qr">Preview QR</button>
      </td>
      <td><button type="button" class="btn btn-danger btn-sm btn-del-bundle">&times;</button></td>
    `;

    tr.querySelector('.btn-del-bundle').addEventListener('click', () => {
      tr.remove();
      updateBundleTally();
    });

    tr.querySelector('.bundle-weight').addEventListener('input', updateBundleTally);

    tr.querySelector('.btn-preview-qr').addEventListener('click', () => {
      showQRModal(bundleCode, tr);
    });

    tbody.appendChild(tr);
    updateBundleTally();
  }

  function updateBundleTally() {
    let sumWeight = 0;
    let count = 0;
    document.querySelectorAll('.bundle-weight').forEach(inp => {
      sumWeight += parseFloat(inp.value) || 0.0;
      count++;
    });

    const phyQty = parseFloat(document.getElementById('inp-physical-qty').value) || 0.0;
    const tallyEl = document.getElementById('bundle-tally');
    if (!tallyEl) return;

    const diff = (sumWeight - phyQty).toFixed(2);
    if (Math.abs(diff) < 0.01 && phyQty > 0) {
      tallyEl.innerHTML = `<span style="color:var(--accent-emerald);">✓ Perfect Breakdown: ${sumWeight.toFixed(2)} / ${phyQty.toFixed(2)} Kgs (${count} Bundles)</span>`;
    } else if (diff > 0) {
      tallyEl.innerHTML = `<span style="color:var(--accent-rose);">⚠ Over Breakdown: ${sumWeight.toFixed(2)} / ${phyQty.toFixed(2)} Kgs (+${diff} Kgs)</span>`;
    } else {
      tallyEl.innerHTML = `<span style="color:var(--accent-amber);">Sum: ${sumWeight.toFixed(2)} / ${phyQty.toFixed(2)} Kgs (Remaining: ${Math.abs(diff)} Kgs)</span>`;
    }
  }

  function showQRModal(bundleCode, tr) {
    const weight = parseFloat(tr.querySelector('.bundle-weight').value) || 0.0;
    const loc = tr.querySelector('.bundle-loc').value || 'A/R01/S01';
    const qc = tr.querySelector('.bundle-qc').value || 'Passed';
    const item = inwardType === 'Fabric' ? fpccFabCombobox.getValue() : inwardType;
    const col = inwardType === 'Fabric' ? fpccColCombobox.getValue() : '';

    const svgHtml = generateQRTagSVG({
      code: bundleCode,
      item,
      color: col,
      qty: weight,
      uom: uomCombobox.getValue() || 'Kgs',
      location: loc,
      qcStatus: qc
    });

    document.getElementById('qr-preview-body').innerHTML = svgHtml;
    document.getElementById('modal-qr-preview').classList.add('open');
  }

  async function handleInwardSubmit(e) {
    e.preventDefault();

    const pCode = document.getElementById('inp-parent-code').value.trim();
    const vendorName = vendorCombobox.getValue();
    const invoiceNo = document.getElementById('inp-invoice-no').value.trim();
    const invQty = parseFloat(document.getElementById('inp-invoice-qty').value) || 0.0;
    const phyQty = parseFloat(document.getElementById('inp-physical-qty').value) || 0.0;

    if (!pCode || !vendorName || !invoiceNo || invQty <= 0 || phyQty <= 0) {
      if (window.showToast) {
        window.showToast('Please fill all required dock inward fields (Vendor, Invoice, Quantities)', 'error');
      }
      return;
    }

    // Collect bundles
    const bundleRows = document.querySelectorAll('#bundle-tbody tr');
    if (bundleRows.length === 0) {
      if (window.showToast) {
        window.showToast('Please break down this lot into at least 1 bundle roll in Step B', 'error');
      }
      return;
    }

    const bundlePayloads = [];
    bundleRows.forEach((tr, idx) => {
      const bCode = tr.dataset.bundleCode;
      const w = parseFloat(tr.querySelector('.bundle-weight').value) || 0.0;
      const gsm = parseFloat(tr.querySelector('.bundle-gsm').value) || 0.0;
      const loc = tr.querySelector('.bundle-loc').value.trim() || 'A/R01/S01';
      const qc = tr.querySelector('.bundle-qc').value;
      const remarks = tr.querySelector('.bundle-qc-remarks').value.trim();

      bundlePayloads.push({
        bundleSequence: idx + 1,
        totalBundles: bundleRows.length,
        itemType: inwardType,
        itemName: inwardType === 'Fabric' ? fpccFabCombobox.getValue() : inwardType,
        colorName: inwardType === 'Fabric' ? fpccColCombobox.getValue() : '',
        inwardWeightQty: w,
        measuredGsm: gsm,
        locationBin: loc,
        qcStatus: qc,
        qcRemarks: remarks,
      });
    });

    const payload = {
      parentCode: pCode,
      inwardDate: document.getElementById('inp-inward-date').value,
      vendorId: vendorName,
      vendorName,
      invoiceNo,
      inwardType,
      uom: uomCombobox.getValue() || 'Kgs',
      invoiceQty: invQty,
      invoiceAmount: parseFloat(document.getElementById('inp-invoice-amount').value) || 0.0,
      physicalReceivedQty: phyQty,
      fabricType: inwardType === 'Fabric' ? 'Knitted' : '',
      vendorFabName: document.getElementById('inp-vendor-fab').value.trim(),
      fpccFabName: fpccFabCombobox.getValue(),
      vendorColor: document.getElementById('inp-vendor-col').value.trim(),
      fpccColor: fpccColCombobox.getValue(),
      invoiceGsm: parseFloat(document.getElementById('inp-inv-gsm').value) || 0.0,
      dia: parseFloat(document.getElementById('inp-dia').value) || 0.0,
      remarks: document.getElementById('inp-dock-remarks').value.trim(),
      bundles: bundlePayloads,
    };

    const submitBtn = document.getElementById('btn-submit-inward');
    submitBtn.disabled = true;
    submitBtn.textContent = '⏳ INWARDING & DISPATCHING WORKFLOWS...';

    try {
      const res = await api.inwardParentLot(payload);
      if (window.showToast) {
        window.showToast(`✓ Inward Lot ${pCode} & ${bundlePayloads.length} bundles recorded in graph!`, 'success');
      }

      // Reset Form & Get Next Code
      document.getElementById('inward-form').reset();
      document.getElementById('bundle-tbody').innerHTML = '';
      document.getElementById('discrepancy-banner').style.display = 'none';
      await refreshParentCode();

      if (onInwardCompleted) onInwardCompleted(res);
    } catch (err) {
      if (window.showToast) {
        window.showToast(`Inward failed: ${err.message}`, 'error');
      }
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = '📦 SAVE INWARD & COMMIT BUNDLES TO GRAPH';
    }
  }

  async function handleRemnantSubmit(e) {
    e.preventDefault();
    const bundleCode = document.getElementById('inp-ret-bundle-code').value.trim();
    const returnType = document.getElementById('sel-ret-type').value;
    const weight = parseFloat(document.getElementById('inp-ret-weight').value) || 0.0;
    const loc = document.getElementById('inp-ret-loc').value.trim();
    const remarks = document.getElementById('inp-ret-remarks').value.trim();

    try {
      const res = await api.returnRemnant({
        bundleCode,
        returnType,
        returnedWeightKg: weight,
        returnLocation: loc,
        remarks
      });

      if (window.showToast) {
        window.showToast(`✓ Return processed! New Bundle Code: ${res.data.updatedBundleCode}`, 'success');
      }
      document.getElementById('modal-remnant-return').classList.remove('open');
    } catch (err) {
      if (window.showToast) {
        window.showToast(`Return failed: ${err.message}`, 'error');
      }
    }
  }

  async function loadDiscrepanciesModal() {
    const tbody = document.getElementById('discrepancies-tbody');
    tbody.innerHTML = `<tr><td colspan="9" style="text-align:center;padding:16px;">Loading discrepancy reports...</td></tr>`;

    try {
      const reports = await api.listDiscrepancies();
      if (!reports || reports.length === 0) {
        tbody.innerHTML = `<tr><td colspan="9" style="text-align:center;padding:16px;color:var(--text-subtle);">No discrepancy records found</td></tr>`;
        return;
      }

      tbody.innerHTML = reports.map(r => `
        <tr>
          <td class="input-mono"><strong>${r.report_no}</strong></td>
          <td class="input-mono">${r.parent_code}</td>
          <td>${r.vendor_name}</td>
          <td class="input-mono">${r.invoice_no}</td>
          <td class="numeric">${r.invoice_qty}</td>
          <td class="numeric">${r.actual_qty}</td>
          <td class="numeric" style="color:${r.variance_qty < 0 ? 'var(--accent-rose)' : 'var(--accent-emerald)'};font-weight:700;">
            ${r.variance_qty > 0 ? '+' : ''}${r.variance_qty}
          </td>
          <td class="numeric" style="font-weight:700;">₹${r.financial_impact.toFixed(2)}</td>
          <td>
            <span class="badge-option selected" style="font-size:10px;padding:3px 6px;">${r.status}</span>
          </td>
        </tr>
      `).join('');
    } catch (err) {
      tbody.innerHTML = `<tr><td colspan="9" style="text-align:center;color:var(--accent-rose);">Error: ${err.message}</td></tr>`;
    }
  }
}
