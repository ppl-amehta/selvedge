// FPCC Selvedge - Station 0: Order Taking View
import { api } from '../api.js';
import { createCombobox } from '../components/combobox.js';

export function initOrdersView({ container, masters, onOrderBooked }) {
  let currentMasters = masters;
  let orderType = 'Bulk';
  let category = 'Apparel';
  let gender = 'Men';
  let taxRate = 5.0;
  let taxBilling = 'Extra';
  let freightMode = 'Surface';
  let freightCost = 'Extra';

  // Combobox references
  let clientCombobox;
  let subCatCombobox;
  let productCombobox;
  let colorCombobox;
  let fabricCombobox;
  let mainLabelCombobox;
  let custPkgCombobox;

  render();

  async function render() {
    container.innerHTML = `
      <div class="view-header">
        <div class="view-title-group">
          <h2>Station 0: Production Order Taking</h2>
          <p>Book orders into the Unified Graph System of Record with 34-point specification control</p>
        </div>
        <div class="header-badges">
          <button class="btn btn-secondary btn-sm" id="btn-copy-specs">📋 Copy Specs from Previous Order</button>
          <button class="btn btn-secondary btn-sm" id="btn-refresh-ord-no">🔄 Refresh No</button>
        </div>
      </div>

      <form id="order-form">
        <!-- Panel 1: Order Identification & Client -->
        <div class="card-panel">
          <div class="panel-header">
            <div class="panel-title">1. Identification & Client Entity</div>
            <div id="badge-order-no" class="system-badge" style="font-size:14px;font-weight:700;">FETCHING...</div>
          </div>
          <div class="grid-4">
            <div class="form-group">
              <label class="form-label">Production Order No <span class="required">*</span></label>
              <input type="text" id="inp-order-no" class="input-control input-mono" readonly />
            </div>
            <div class="form-group">
              <label class="form-label">Sales Order / PO Ref No</label>
              <input type="text" id="inp-so-no" class="input-control" placeholder="e.g. SO-2026-99" />
            </div>
            <div class="form-group">
              <label class="form-label">Booking Date <span class="required">*</span></label>
              <input type="date" id="inp-booking-date" class="input-control" />
            </div>
            <div class="form-group">
              <label class="form-label">Target ETD <span class="required">*</span></label>
              <input type="date" id="inp-target-etd" class="input-control" />
            </div>
          </div>

          <div class="form-group" style="margin-top:8px;">
            <label class="form-label">Order Type (Strict Enum) <span class="required">*</span></label>
            <div class="badge-select-wrapper" id="group-order-type">
              <span class="badge-option selected" data-val="Bulk">Bulk Production</span>
              <span class="badge-option" data-val="Sample">Sample</span>
              <span class="badge-option" data-val="FOC">FOC (Free of Cost)</span>
              <span class="badge-option" data-val="Repeat">Repeat Run</span>
              <span class="badge-option" data-val="Development">Development</span>
            </div>
          </div>

          <hr style="border:none;border-top:1px solid var(--border-subtle);margin:16px 0;" />

          <div class="grid-3">
            <div class="form-group">
              <label class="form-label">Client Name <span class="required">*</span> <span class="helper">Auto-registers</span></label>
              <div id="cb-client"></div>
            </div>
            <div class="form-group">
              <label class="form-label">Contact Person</label>
              <input type="text" id="inp-client-contact" class="input-control" placeholder="Procurement manager" />
            </div>
            <div class="form-group">
              <label class="form-label">Phone / Mobile</label>
              <input type="text" id="inp-client-phone" class="input-control" placeholder="+91 98xxx" />
            </div>
          </div>

          <div class="grid-3">
            <div class="form-group">
              <label class="form-label">Client Email</label>
              <input type="email" id="inp-client-email" class="input-control" placeholder="client@corp.com" />
            </div>
            <div class="form-group">
              <label class="form-label">Client GSTIN</label>
              <input type="text" id="inp-client-gstin" class="input-control input-mono" placeholder="27XXXXX1234F1Z1" />
            </div>
            <div class="form-group">
              <label class="form-label">Billing Address</label>
              <input type="text" id="inp-client-billing" class="input-control" placeholder="HQ Address" />
            </div>
          </div>
        </div>

        <!-- Panel 2: Garment Specification -->
        <div class="card-panel">
          <div class="panel-header">
            <div class="panel-title">2. Garment Specification & Master Attributes</div>
          </div>

          <div class="grid-2">
            <div class="form-group">
              <label class="form-label">Category <span class="required">*</span></label>
              <div class="badge-select-wrapper" id="group-category">
                <span class="badge-option selected" data-val="Apparel">Apparel</span>
                <span class="badge-option" data-val="Non-Apparel">Non-Apparel</span>
              </div>
            </div>
            <div class="form-group">
              <label class="form-label">Gender <span class="required">*</span></label>
              <div class="badge-select-wrapper" id="group-gender">
                <span class="badge-option selected" data-val="Men">Men</span>
                <span class="badge-option" data-val="Women">Women</span>
                <span class="badge-option" data-val="Unisex">Unisex</span>
                <span class="badge-option" data-val="Boys">Boys</span>
                <span class="badge-option" data-val="Girls">Girls</span>
              </div>
            </div>
          </div>

          <div class="grid-4" style="margin-top:10px;">
            <div class="form-group">
              <label class="form-label">Sub-Category <span class="required">*</span></label>
              <div id="cb-subcategory"></div>
            </div>
            <div class="form-group">
              <label class="form-label">Product Name <span class="required">*</span></label>
              <div id="cb-product"></div>
            </div>
            <div class="form-group">
              <label class="form-label">Color Name <span class="required">*</span></label>
              <div id="cb-color"></div>
            </div>
            <div class="form-group">
              <label class="form-label">Fabric Name <span class="required">*</span></label>
              <div id="cb-fabric"></div>
            </div>
          </div>
        </div>

        <!-- Panel 3: Commercials & Pricing -->
        <div class="card-panel">
          <div class="panel-header">
            <div class="panel-title">3. Commercial Terms & Taxes</div>
          </div>

          <div class="grid-4">
            <div class="form-group">
              <label class="form-label">Unit Price (₹) <span class="required">*</span></label>
              <input type="number" id="inp-unit-price" class="input-control numeric" step="0.5" min="0" placeholder="0.00" value="450.00" />
            </div>
            <div class="form-group">
              <label class="form-label">Tax Rate (Strict Enum) <span class="required">*</span></label>
              <div class="badge-select-wrapper" id="group-tax-rate">
                <span class="badge-option selected" data-val="5">5%</span>
                <span class="badge-option" data-val="12">12%</span>
                <span class="badge-option" data-val="18">18%</span>
              </div>
            </div>
            <div class="form-group">
              <label class="form-label">Tax Billing Mode</label>
              <div class="badge-select-wrapper" id="group-tax-billing">
                <span class="badge-option selected" data-val="Extra">Extra</span>
                <span class="badge-option" data-val="Included">Included</span>
              </div>
            </div>
            <div class="form-group">
              <label class="form-label">Freight Mode & Cost</label>
              <div class="badge-select-wrapper" id="group-freight-mode">
                <span class="badge-option selected" data-val="Surface">Surface</span>
                <span class="badge-option" data-val="Air">Air</span>
                <span class="badge-option" data-val="Courier">Courier</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Panel 4: Size Matrix & Real-time Live Math -->
        <div class="card-panel">
          <div class="panel-header">
            <div class="panel-title">4. Size Breakdown Matrix & Instant Valuations</div>
            <div style="font-family:var(--font-mono);font-size:15px;color:var(--accent-emerald);" id="display-order-totals">
              Total Qty: <strong>0 Pcs</strong> | Total Value: <strong>₹0.00</strong>
            </div>
          </div>

          <div class="size-matrix-grid">
            <div class="size-input-col"><label>XS</label><input type="number" class="input-control size-box" data-size="XS" min="0" placeholder="0" /></div>
            <div class="size-input-col"><label>S</label><input type="number" class="input-control size-box" data-size="S" min="0" placeholder="0" /></div>
            <div class="size-input-col"><label>M</label><input type="number" class="input-control size-box" data-size="M" min="0" placeholder="0" /></div>
            <div class="size-input-col"><label>L</label><input type="number" class="input-control size-box" data-size="L" min="0" placeholder="0" /></div>
            <div class="size-input-col"><label>XL</label><input type="number" class="input-control size-box" data-size="XL" min="0" placeholder="0" /></div>
            <div class="size-input-col"><label>2XL</label><input type="number" class="input-control size-box" data-size="2XL" min="0" placeholder="0" /></div>
            <div class="size-input-col"><label>3XL</label><input type="number" class="input-control size-box" data-size="3XL" min="0" placeholder="0" /></div>
            <div class="size-input-col"><label>4XL</label><input type="number" class="input-control size-box" data-size="4XL" min="0" placeholder="0" /></div>
            <div class="size-input-col"><label>5XL</label><input type="number" class="input-control size-box" data-size="5XL" min="0" placeholder="0" /></div>
            <div class="size-input-col"><label>Free</label><input type="number" class="input-control size-box" data-size="Free" min="0" placeholder="0" /></div>
          </div>
        </div>

        <!-- Panel 5: Multi-Shipping Destinations -->
        <div class="card-panel">
          <div class="panel-header">
            <div class="panel-title">
              <label style="display:flex;align-items:center;gap:8px;cursor:pointer;">
                <input type="checkbox" id="chk-multi-shipping" style="width:16px;height:16px;" />
                <span>5. Multi-Destination Shipping Split</span>
              </label>
            </div>
            <div id="multi-shipping-status" style="font-size:12px;color:var(--text-subtle);">Direct Single Destination</div>
          </div>

          <div id="multi-shipping-panel" style="display:none;">
            <p style="font-size:12px;color:var(--text-muted);margin-bottom:12px;">Split this order across branch offices, retail stores, or stadium hubs:</p>
            <div class="data-table-container">
              <table class="data-table" id="shipping-dest-table">
                <thead>
                  <tr>
                    <th>Destination / Hub Address</th>
                    <th class="numeric" style="width:120px;">Qty (Pcs)</th>
                    <th>Carrier / Transporter</th>
                    <th>Docket / AWB No</th>
                    <th>Target Dispatch</th>
                    <th style="width:50px;"></th>
                  </tr>
                </thead>
                <tbody id="shipping-dest-tbody">
                  <!-- Dynamically populated rows -->
                </tbody>
              </table>
            </div>
            <div style="margin-top:12px;display:flex;justify-content:space-between;align-items:center;">
              <button type="button" class="btn btn-secondary btn-sm" id="btn-add-shipping-row">+ Add Destination</button>
              <div id="shipping-tally" style="font-family:var(--font-mono);font-size:12px;">Allocated: 0 / 0 Pcs</div>
            </div>
          </div>
        </div>

        <!-- Panel 6: Player / Staff Personalization Roster -->
        <div class="card-panel">
          <div class="panel-header">
            <div class="panel-title">6. Personalization Roster (Jersey / Staff Customization)</div>
            <button type="button" class="btn btn-secondary btn-sm" id="btn-excel-paste-modal">📋 Bulk Paste from Excel</button>
          </div>

          <div class="data-table-container">
            <table class="data-table" id="roster-table">
              <thead>
                <tr>
                  <th style="width:100px;">Size</th>
                  <th>Player / Employee Name</th>
                  <th style="width:100px;">Jersey No</th>
                  <th>Custom Remarks</th>
                  <th style="width:50px;"></th>
                </tr>
              </thead>
              <tbody id="roster-tbody">
                <!-- Roster rows -->
              </tbody>
            </table>
          </div>
          <button type="button" class="btn btn-secondary btn-sm" id="btn-add-roster-row" style="margin-top:12px;">+ Add Personalization Row</button>
        </div>

        <!-- Panel 7: Trims, Branding & Packaging -->
        <div class="card-panel">
          <div class="panel-header">
            <div class="panel-title">7. Branding, Labels & Packaging Specs</div>
          </div>

          <div class="grid-3">
            <div class="form-group">
              <label class="form-label">Branding Type</label>
              <input type="text" id="inp-branding-type" class="input-control" placeholder="e.g. Chest Embroidery + Back Screen Print" />
            </div>
            <div class="form-group">
              <label class="form-label">Main Label Spec</label>
              <div id="cb-main-label"></div>
            </div>
            <div class="form-group">
              <label class="form-label">Packaging Requirement</label>
              <div id="cb-cust-pkg"></div>
            </div>
          </div>

          <div class="form-group" style="margin-top:8px;">
            <label class="form-label">Special Production Instructions / Client Comments</label>
            <textarea id="inp-client-comments" class="textarea-control" rows="2" placeholder="e.g. Bio-wash finish mandatory. Double stitch armhole seams. Strict needle check."></textarea>
          </div>
        </div>

        <!-- Submission Bar -->
        <div style="display:flex;justify-content:flex-end;gap:16px;margin-bottom:40px;">
          <button type="reset" class="btn btn-secondary">Reset Form</button>
          <button type="submit" class="btn btn-primary" id="btn-submit-order" style="padding:12px 28px;font-size:15px;font-weight:700;">
            🚀 BOOK PRODUCTION ORDER INTO GRAPH
          </button>
        </div>
      </form>

      <!-- Modal 1: Copy Specs from Previous Order -->
      <div class="modal-overlay" id="modal-copy-specs">
        <div class="modal-box">
          <div class="modal-header">
            <h3>Copy Specs from Previous Production Order</h3>
            <button class="btn btn-secondary btn-sm modal-close">&times;</button>
          </div>
          <div class="modal-body">
            <p style="font-size:12px;color:var(--text-muted);margin-bottom:12px;">Select an existing order to pre-populate product, garment specifications, color, fabric, and label configurations:</p>
            <div class="data-table-container">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>Order No</th>
                    <th>Client</th>
                    <th>Product</th>
                    <th>Color</th>
                    <th>Qty</th>
                    <th></th>
                  </tr>
                </thead>
                <tbody id="prev-orders-tbody">
                  <!-- Populated dynamically -->
                </tbody>
              </table>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary modal-close">Cancel</button>
          </div>
        </div>
      </div>

      <!-- Modal 2: Bulk Excel Paste Parser for Personalization -->
      <div class="modal-overlay" id="modal-excel-paste">
        <div class="modal-box">
          <div class="modal-header">
            <h3>Bulk Import Personalization Roster</h3>
            <button class="btn btn-secondary btn-sm modal-close">&times;</button>
          </div>
          <div class="modal-body">
            <p style="font-size:12px;color:var(--text-muted);margin-bottom:8px;">
              Copy columns from your Excel or Google Sheet: <strong>[Size] [Tab] [Name] [Tab] [Number] [Tab] [Remarks]</strong> and paste below:
            </p>
            <textarea id="excel-paste-area" class="textarea-control" rows="10" placeholder="M	Virat Kohli	18	Captain&#10;L	Rohit Sharma	45	Opener&#10;XL	Jasprit Bumrah	93	Fast Bowler" style="font-family:var(--font-mono);font-size:12px;"></textarea>
          </div>
          <div class="modal-footer">
            <button class="btn btn-secondary modal-close">Cancel</button>
            <button class="btn btn-primary" id="btn-parse-excel-roster">Import Rows</button>
          </div>
        </div>
      </div>
    `;

    setupControls();
    setupEventListeners();
    await refreshOrderNo();
  }

  function setupControls() {
    // Default Dates
    const today = new Date().toISOString().split('T')[0];
    const etd = new Date(Date.now() + 14 * 86400000).toISOString().split('T')[0];
    document.getElementById('inp-booking-date').value = today;
    document.getElementById('inp-target-etd').value = etd;

    // Comboboxes
    clientCombobox = createCombobox({
      container: document.getElementById('cb-client'),
      category: 'client',
      placeholder: 'Search client or type new...',
      items: currentMasters.clients || [],
      displayKey: 'name',
      valueKey: 'name',
      onSelect: (client) => {
        if (client.contact_person) document.getElementById('inp-client-contact').value = client.contact_person;
        if (client.phone) document.getElementById('inp-client-phone').value = client.phone;
        if (client.email) document.getElementById('inp-client-email').value = client.email;
        if (client.gstin) document.getElementById('inp-client-gstin').value = client.gstin;
        if (client.address) document.getElementById('inp-client-billing').value = client.address;
      }
    });

    const subCategories = [
      { display_name: 'Polo Shirt' },
      { display_name: 'Round Neck T-Shirt' },
      { display_name: 'V-Neck T-Shirt' },
      { display_name: 'Hoodie' },
      { display_name: 'Sweatshirt' },
      { display_name: 'Joggers / Trackpants' },
      { display_name: 'Shorts' },
      { display_name: 'Jacket' },
      { display_name: 'Jersey' }
    ];

    subCatCombobox = createCombobox({
      container: document.getElementById('cb-subcategory'),
      category: 'subcategory',
      placeholder: 'Select sub-category...',
      items: subCategories,
      displayKey: 'display_name',
      valueKey: 'display_name',
      initialValue: 'Polo Shirt'
    });

    productCombobox = createCombobox({
      container: document.getElementById('cb-product'),
      category: 'product',
      placeholder: 'Search product...',
      items: currentMasters.products || [],
      displayKey: 'display_name',
      valueKey: 'display_name',
      initialValue: 'Classic Polo'
    });

    colorCombobox = createCombobox({
      container: document.getElementById('cb-color'),
      category: 'color',
      placeholder: 'Search color (301 colors)...',
      items: currentMasters.colors || [],
      displayKey: 'display_name',
      valueKey: 'display_name',
      initialValue: 'Navy Blue'
    });

    fabricCombobox = createCombobox({
      container: document.getElementById('cb-fabric'),
      category: 'fabric',
      placeholder: 'Search fabric (225 fabrics)...',
      items: currentMasters.fabrics || [],
      displayKey: 'display_name',
      valueKey: 'display_name',
      initialValue: '100% Cotton Pique 220 GSM'
    });

    mainLabelCombobox = createCombobox({
      container: document.getElementById('cb-main-label'),
      category: 'label',
      placeholder: 'Select label...',
      items: currentMasters.labels || [],
      displayKey: 'display_name',
      valueKey: 'display_name',
      initialValue: 'Woven Satin Main Label'
    });

    custPkgCombobox = createCombobox({
      container: document.getElementById('cb-cust-pkg'),
      category: 'packaging',
      placeholder: 'Select packaging...',
      items: currentMasters.packaging || [],
      displayKey: 'display_name',
      valueKey: 'display_name',
      initialValue: 'Self Adhesive Polybag'
    });

    // Initialize 1 default empty destination & 1 default roster row
    addShippingRow();
    addRosterRow();
  }

  function setupEventListeners() {
    // Strict Enum Badges handlers
    setupBadgeGroup('group-order-type', (val) => { orderType = val; });
    setupBadgeGroup('group-category', (val) => { category = val; });
    setupBadgeGroup('group-gender', (val) => { gender = val; });
    setupBadgeGroup('group-tax-rate', (val) => { taxRate = parseFloat(val); recalculateMath(); });
    setupBadgeGroup('group-tax-billing', (val) => { taxBilling = val; recalculateMath(); });
    setupBadgeGroup('group-freight-mode', (val) => { freightMode = val; });

    // Refresh Order No
    document.getElementById('btn-refresh-ord-no').addEventListener('click', refreshOrderNo);

    // Size Box & Price Math
    document.querySelectorAll('.size-box').forEach(input => {
      input.addEventListener('input', recalculateMath);
    });
    document.getElementById('inp-unit-price').addEventListener('input', recalculateMath);

    // Multi-Shipping Toggle
    const chkMulti = document.getElementById('chk-multi-shipping');
    chkMulti.addEventListener('change', () => {
      const panel = document.getElementById('multi-shipping-panel');
      const status = document.getElementById('multi-shipping-status');
      if (chkMulti.checked) {
        panel.style.display = 'block';
        status.textContent = 'Active Multi-Destination Split';
        status.style.color = 'var(--accent-primary)';
      } else {
        panel.style.display = 'none';
        status.textContent = 'Direct Single Destination';
        status.style.color = 'var(--text-subtle)';
      }
      updateShippingTally();
    });

    document.getElementById('btn-add-shipping-row').addEventListener('click', addShippingRow);
    document.getElementById('btn-add-roster-row').addEventListener('click', () => addRosterRow());

    // Copy Specs Modal
    const modalCopy = document.getElementById('modal-copy-specs');
    document.getElementById('btn-copy-specs').addEventListener('click', async () => {
      await loadPreviousOrdersModal();
      modalCopy.classList.add('open');
    });

    // Excel Paste Modal
    const modalExcel = document.getElementById('modal-excel-paste');
    document.getElementById('btn-excel-paste-modal').addEventListener('click', () => {
      document.getElementById('excel-paste-area').value = '';
      modalExcel.classList.add('open');
    });

    document.getElementById('btn-parse-excel-roster').addEventListener('click', () => {
      parseExcelRoster();
      modalExcel.classList.remove('open');
    });

    // Close Modals
    document.querySelectorAll('.modal-close').forEach(btn => {
      btn.addEventListener('click', (e) => {
        e.target.closest('.modal-overlay').classList.remove('open');
      });
    });

    // Form Submit
    document.getElementById('order-form').addEventListener('submit', handleOrderSubmit);
  }

  function setupBadgeGroup(groupId, onChange) {
    const group = document.getElementById(groupId);
    if (!group) return;
    group.querySelectorAll('.badge-option').forEach(badge => {
      badge.addEventListener('click', () => {
        group.querySelectorAll('.badge-option').forEach(b => b.classList.remove('selected'));
        badge.classList.add('selected');
        onChange(badge.dataset.val);
      });
    });
  }

  async function refreshOrderNo() {
    try {
      const no = await api.getNextOrderNo();
      document.getElementById('inp-order-no').value = no;
      document.getElementById('badge-order-no').textContent = no;
    } catch (err) {
      console.error(err);
    }
  }

  function recalculateMath() {
    let totalQty = 0;
    document.querySelectorAll('.size-box').forEach(inp => {
      totalQty += parseInt(inp.value, 10) || 0;
    });

    const unitPrice = parseFloat(document.getElementById('inp-unit-price').value) || 0.0;
    const baseValue = totalQty * unitPrice;
    let taxAmount = 0.0;

    if (taxBilling === 'Extra') {
      taxAmount = baseValue * (taxRate / 100.0);
    } else {
      taxAmount = baseValue - (baseValue / (1.0 + taxRate / 100.0));
    }

    const grandTotal = taxBilling === 'Extra' ? baseValue + taxAmount : baseValue;

    document.getElementById('display-order-totals').innerHTML = `
      Total Qty: <strong>${totalQty} Pcs</strong> | Total Value: <strong>₹${grandTotal.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</strong>
      <span style="font-size:11px;color:var(--text-subtle);">(${taxBilling} GST ${taxRate}%)</span>
    `;

    updateShippingTally();
  }

  function addShippingRow(dest = '', qty = '', carrier = 'Bluedart Surface', docket = '', etd = '') {
    const tbody = document.getElementById('shipping-dest-tbody');
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td><input type="text" class="input-control ship-dest" placeholder="City or Store Address" value="${dest}" /></td>
      <td><input type="number" class="input-control ship-qty numeric" placeholder="0" min="0" value="${qty}" /></td>
      <td><input type="text" class="input-control ship-carrier" placeholder="Carrier" value="${carrier}" /></td>
      <td><input type="text" class="input-control ship-docket input-mono" placeholder="AWB / Docket" value="${docket}" /></td>
      <td><input type="date" class="input-control ship-etd" value="${etd}" /></td>
      <td><button type="button" class="btn btn-danger btn-sm btn-del-row">&times;</button></td>
    `;
    tr.querySelector('.btn-del-row').addEventListener('click', () => {
      tr.remove();
      updateShippingTally();
    });
    tr.querySelector('.ship-qty').addEventListener('input', updateShippingTally);
    tbody.appendChild(tr);
    updateShippingTally();
  }

  function updateShippingTally() {
    let totalAllocated = 0;
    document.querySelectorAll('.ship-qty').forEach(inp => {
      totalAllocated += parseInt(inp.value, 10) || 0;
    });

    let orderQty = 0;
    document.querySelectorAll('.size-box').forEach(inp => {
      orderQty += parseInt(inp.value, 10) || 0;
    });

    const tallyEl = document.getElementById('shipping-tally');
    if (!tallyEl) return;

    if (totalAllocated === orderQty && orderQty > 0) {
      tallyEl.innerHTML = `<span style="color:var(--accent-emerald);">✓ Perfect Balance: ${totalAllocated} / ${orderQty} Pcs</span>`;
    } else if (totalAllocated > orderQty) {
      tallyEl.innerHTML = `<span style="color:var(--accent-rose);">⚠ Over-allocated: ${totalAllocated} / ${orderQty} Pcs (+${totalAllocated - orderQty})</span>`;
    } else {
      tallyEl.innerHTML = `<span style="color:var(--accent-amber);">Allocated: ${totalAllocated} / ${orderQty} Pcs (Remaining: ${orderQty - totalAllocated})</span>`;
    }
  }

  function addRosterRow(size = 'M', name = '', number = '', remarks = '') {
    const tbody = document.getElementById('roster-tbody');
    const tr = document.createElement('tr');
    tr.innerHTML = `
      <td>
        <select class="select-control roster-size">
          ${['XS', 'S', 'M', 'L', 'XL', '2XL', '3XL', '4XL', '5XL', 'Free'].map(s => `<option value="${s}" ${s === size ? 'selected' : ''}>${s}</option>`).join('')}
        </select>
      </td>
      <td><input type="text" class="input-control roster-name" placeholder="Name on jersey" value="${name}" /></td>
      <td><input type="text" class="input-control roster-num input-mono" placeholder="No." value="${number}" /></td>
      <td><input type="text" class="input-control roster-remarks" placeholder="Notes (e.g. Captain / Sponsor logo)" value="${remarks}" /></td>
      <td><button type="button" class="btn btn-danger btn-sm btn-del-row">&times;</button></td>
    `;
    tr.querySelector('.btn-del-row').addEventListener('click', () => tr.remove());
    tbody.appendChild(tr);
  }

  function parseExcelRoster() {
    const raw = document.getElementById('excel-paste-area').value.trim();
    if (!raw) return;

    const lines = raw.split(/\r?\n/);
    const tbody = document.getElementById('roster-tbody');
    tbody.innerHTML = '';

    let parsedCount = 0;
    lines.forEach(line => {
      if (!line.trim()) return;
      const parts = line.split('\t').map(p => p.trim());
      const size = parts[0] || 'M';
      const name = parts[1] || '';
      const num = parts[2] || '';
      const rem = parts[3] || '';
      addRosterRow(size, name, num, rem);
      parsedCount++;
    });

    if (window.showToast) {
      window.showToast(`Imported ${parsedCount} personalization records`, 'success');
    }
  }

  async function loadPreviousOrdersModal() {
    const tbody = document.getElementById('prev-orders-tbody');
    tbody.innerHTML = `<tr><td colspan="6" style="text-align:center;padding:16px;">Loading recent orders...</td></tr>`;

    try {
      const orders = await api.listOrders();
      if (!orders || orders.length === 0) {
        tbody.innerHTML = `<tr><td colspan="6" style="text-align:center;padding:16px;color:var(--text-subtle);">No previous orders recorded in system</td></tr>`;
        return;
      }

      tbody.innerHTML = orders.map(ord => `
        <tr>
          <td class="input-mono"><strong>${ord.production_order_no}</strong></td>
          <td>${ord.client_name}</td>
          <td>${ord.product_name}</td>
          <td>${ord.color_name}</td>
          <td class="numeric">${ord.total_qty} Pcs</td>
          <td style="text-align:right;">
            <button type="button" class="btn btn-primary btn-sm btn-apply-specs" data-ord="${ord.production_order_no}">Copy Specs</button>
          </td>
        </tr>
      `).join('');

      tbody.querySelectorAll('.btn-apply-specs').forEach(btn => {
        btn.addEventListener('click', async () => {
          const ordNo = btn.dataset.ord;
          await copySpecsFromOrder(ordNo);
          document.getElementById('modal-copy-specs').classList.remove('open');
        });
      });
    } catch (err) {
      tbody.innerHTML = `<tr><td colspan="6" style="text-align:center;color:var(--accent-rose);">Error: ${err.message}</td></tr>`;
    }
  }

  async function copySpecsFromOrder(orderNo) {
    try {
      const res = await api.getOrder(orderNo);
      const ord = res.data;
      if (!ord) return;

      clientCombobox.setValue(ord.client_name);
      document.getElementById('inp-client-contact').value = ord.contact_person || '';
      document.getElementById('inp-client-phone').value = ord.phone || '';
      document.getElementById('inp-client-email').value = ord.email || '';
      document.getElementById('inp-client-gstin').value = ord.gstin || '';
      document.getElementById('inp-client-billing').value = ord.billing_address || '';

      productCombobox.setValue(ord.product_name || '');
      colorCombobox.setValue(ord.color_name || '');
      fabricCombobox.setValue(ord.fabric_name || '');
      subCatCombobox.setValue(ord.sub_category || '');

      document.getElementById('inp-unit-price').value = ord.unit_price || 0;
      document.getElementById('inp-branding-type').value = ord.branding_type || '';
      mainLabelCombobox.setValue(ord.main_label || '');
      custPkgCombobox.setValue(ord.cust_packaging || '');
      document.getElementById('inp-client-comments').value = ord.client_comment || '';

      recalculateMath();

      if (window.showToast) {
        window.showToast(`Copied specs from ${orderNo}`, 'success');
      }
    } catch (err) {
      if (window.showToast) {
        window.showToast(`Failed to copy specs: ${err.message}`, 'error');
      }
    }
  }

  async function handleOrderSubmit(e) {
    e.preventDefault();

    const orderNo = document.getElementById('inp-order-no').value.trim();
    const clientName = clientCombobox.getValue();
    const productName = productCombobox.getValue();
    const colorName = colorCombobox.getValue();
    const fabricName = fabricCombobox.getValue();
    const subCategory = subCatCombobox.getValue();

    if (!orderNo || !clientName || !productName || !colorName) {
      if (window.showToast) {
        window.showToast('Please fill all mandatory fields (Client, Product, Color)', 'error');
      }
      return;
    }

    // Size Breakdown
    const sizeMatrix = {};
    let totalQty = 0;
    document.querySelectorAll('.size-box').forEach(inp => {
      const s = inp.dataset.size;
      const q = parseInt(inp.value, 10) || 0;
      if (q > 0) {
        sizeMatrix[s] = q;
        totalQty += q;
      }
    });

    if (totalQty === 0) {
      if (window.showToast) {
        window.showToast('Total quantity must be greater than 0', 'error');
      }
      return;
    }

    // Shipping Destinations
    const isMultiShipping = document.getElementById('chk-multi-shipping').checked;
    const shippingLocations = [];
    if (isMultiShipping) {
      document.querySelectorAll('#shipping-dest-tbody tr').forEach(tr => {
        const dest = tr.querySelector('.ship-dest').value.trim();
        const qty = parseInt(tr.querySelector('.ship-qty').value, 10) || 0;
        const carrier = tr.querySelector('.ship-carrier').value.trim();
        const docketNo = tr.querySelector('.ship-docket').value.trim();
        const etd = tr.querySelector('.ship-etd').value.trim();
        if (dest && qty > 0) {
          shippingLocations.push({ destination: dest, quantity: qty, carrier, docketNo, etd });
        }
      });
    }

    // Personalization Roster
    const personalization = [];
    document.querySelectorAll('#roster-tbody tr').forEach(tr => {
      const size = tr.querySelector('.roster-size').value;
      const name = tr.querySelector('.roster-name').value.trim();
      const number = tr.querySelector('.roster-num').value.trim();
      const remarks = tr.querySelector('.roster-remarks').value.trim();
      if (name || number) {
        personalization.push({ size, name, number, remarks });
      }
    });

    const payload = {
      productionOrderNo: orderNo,
      salesOrderNo: document.getElementById('inp-so-no').value.trim(),
      orderType,
      bookingDate: document.getElementById('inp-booking-date').value,
      clientId: clientCombobox.getValue(),
      clientName,
      contactPerson: document.getElementById('inp-client-contact').value.trim(),
      phone: document.getElementById('inp-client-phone').value.trim(),
      email: document.getElementById('inp-client-email').value.trim(),
      billingAddress: document.getElementById('inp-client-billing').value.trim(),
      shippingAddress: document.getElementById('inp-client-billing').value.trim(),
      gstin: document.getElementById('inp-client-gstin').value.trim(),
      category,
      subCategory,
      productName,
      colorName,
      fabricName,
      gender,
      unitPrice: parseFloat(document.getElementById('inp-unit-price').value) || 0.0,
      taxRate,
      taxBilling,
      freightMode,
      freightCost,
      totalQty,
      sizeBreakdown: sizeMatrix,
      multipleShipping: isMultiShipping,
      shippingLocations,
      personalization,
      brandingType: document.getElementById('inp-branding-type').value.trim(),
      mainLabel: mainLabelCombobox.getValue(),
      custPackaging: custPkgCombobox.getValue(),
      clientComment: document.getElementById('inp-client-comments').value.trim(),
      targetEtd: document.getElementById('inp-target-etd').value,
    };

    const submitBtn = document.getElementById('btn-submit-order');
    submitBtn.disabled = true;
    submitBtn.textContent = '⏳ RECORDING INTO UNIFIED GRAPH...';

    try {
      const result = await api.createOrder(payload);
      if (window.showToast) {
        window.showToast(`✓ Production Order ${orderNo} successfully booked! Node: ${result.graphNodeId}`, 'success');
      }

      // Reset Form & Fetch Next Order No
      document.getElementById('order-form').reset();
      document.querySelectorAll('.size-box').forEach(i => i.value = '');
      document.getElementById('shipping-dest-tbody').innerHTML = '';
      document.getElementById('roster-tbody').innerHTML = '';
      addShippingRow();
      addRosterRow();
      recalculateMath();
      await refreshOrderNo();

      if (onOrderBooked) onOrderBooked(result);
    } catch (err) {
      if (window.showToast) {
        window.showToast(`Failed to book order: ${err.message}`, 'error');
      }
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = '🚀 BOOK PRODUCTION ORDER INTO GRAPH';
    }
  }
}
