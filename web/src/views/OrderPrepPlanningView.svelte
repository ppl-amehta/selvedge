<script>
  import { onMount } from 'svelte';
  import { api } from '../api.js';
  import Combobox from '../components/Combobox.svelte';

  let { masters = {} } = $props();

  // Orders List & Selected Order
  let ordersList = $state([]);
  let selectedOrderId = $state('');
  let prepData = $state(null);
  let isLoading = $state(false);
  let activeTab = $state('prep'); // 'prep' (Pharma) | 'bom' (BOM & SOH) | 'plan' (Routing & Gantt)
  let selectedItemIndex = $state(0);

  // Station 2: Pharma State
  let pharmaType = $state('Body');
  let pharmaNo = $state('');
  let changesToPharma = $state('');
  let measurementRows = $state([
    { point: 'Chest (Half)', tolerance: '± 0.5 cm', xs: 46, s: 48, m: 50, l: 52, xl: 54, xxl: 56, xxxl: 58 },
    { point: 'Body Length', tolerance: '± 0.7 cm', xs: 66, s: 68, m: 70, l: 72, xl: 74, xxl: 76, xxxl: 78 },
    { point: 'Across Shoulder', tolerance: '± 0.5 cm', xs: 40, s: 42, m: 44, l: 46, xl: 48, xxl: 50, xxxl: 52 },
    { point: 'Sleeve Length', tolerance: '± 0.5 cm', xs: 20, s: 21, m: 22, l: 23, xl: 24, xxl: 25, xxxl: 26 },
    { point: 'Neck Opening', tolerance: '± 0.3 cm', xs: 17, s: 17.5, m: 18, l: 18.5, xl: 19, xxl: 19.5, xxxl: 20 }
  ]);
  let isSavingPharma = $state(false);

  // Station 2: BOM State
  let bomLineType = $state('FABRIC'); // 'FABRIC' | 'ACCESSORY' | 'PACKAGING'
  let bomItemName = $state('');
  let bomSubType = $state('Main Body');
  let bomConsumptionAvg = $state(0.35);
  let bomWastagePct = $state(3.0);
  let bomUom = $state('Kgs');
  let bomRatePerUnit = $state(380.0);
  let bomComments = $state('');
  let isSavingBom = $state(false);

  // Procurement Requisition Modal
  let isPoModalOpen = $state(false);
  let poTargetBomLine = $state(null);
  let poBufferQty = $state(5.0);
  let poVendor = $state('');
  let poEstRate = $state(0.0);
  let poRequiredBy = $state('');
  let isRaisingPo = $state(false);

  // Station 3: Routing & Scheduling Steps
  let routingSteps = $state([]);
  let isSavingSteps = $state(false);

  // State Transition Modal
  let isAdvanceModalOpen = $state(false);
  let targetState = $state('');
  let advanceReason = $state('');
  let isAdvancing = $state(false);

  // Master Dropdown Options
  const pharmaTypes = ['Body', 'Collar', 'Sleeve', 'Neck Collar', 'Placket', 'Pocket', 'Cuff', 'Rib', 'Hood', 'Complete Set'];
  const fabricMasters = [
    'Super Combed Pique 220', 'Single Jersey 180', 'Fleece 320', 'Rib 1x1 / 2x2',
    'Interlock 240', 'Collar Rib 400', 'Cuff Rib 400', 'Woven Twill 220',
    'Terry 280', 'Looperknit 260', 'Bonded Knit 340'
  ];
  const trimMasters = [
    'Vislon Zip 24" Auto-Lock', 'Coil Zip 18" CE', 'Metal Zipper #5 Brass',
    'Button 18L Chalk 4-Hole', 'Button 24L Horn', 'Drawcord Flat 15mm Cotton',
    'Aglet Matte Black 4mm', 'Eyelet Gunmetal 8mm', 'Elastic 38mm Woven',
    'Micro-Fusing 25 GSM', 'Taffeta Care Label', 'Woven Damask Neck Label'
  ];
  const packagingMasters = [
    'Self-Adhesive Polybag 12x15', 'Master Carton 7-Ply Heavy Duty',
    'Brown Paper Wrap Roll', 'Branded Hangtag Polybag Set'
  ];
  const processMasters = ['Cutting', 'Branding', 'Machine', 'Finishing', 'Packing'];
  const vendorMasters = [
    'Cutting Room Line A', 'Stitching Floor Line 1', 'Stitching Floor Line 2',
    'Rajesh Embroidery Works', 'Apex Screen Printing Works', 'YKK India Pvt Ltd',
    'Arvind Mills Ltd', 'Vardhman Textiles Ltd', 'Finishing & Steam Iron Section',
    'Packing Section Floor 3'
  ];

  // Derived current active item
  let currentItem = $derived(
    prepData?.items?.[selectedItemIndex] || null
  );

  onMount(async () => {
    await refreshOrderList();
  });

  async function refreshOrderList() {
    try {
      ordersList = await api.listOrdersForPrep();
      if (ordersList.length > 0 && !selectedOrderId) {
        selectedOrderId = ordersList[0].id;
        await loadPrepDetails(selectedOrderId);
      }
    } catch (err) {
      console.error('Failed to load orders for prep:', err);
    }
  }

  async function handleOrderSelect(orderId) {
    selectedOrderId = orderId;
    selectedItemIndex = 0;
    await loadPrepDetails(orderId);
  }

  async function loadPrepDetails(orderId) {
    if (!orderId) return;
    try {
      isLoading = true;
      prepData = await api.getOrderPrepAndPlan(orderId);
      
      // Seed default routing steps if none exist yet
      if (prepData.routingSteps && prepData.routingSteps.length > 0) {
        routingSteps = prepData.routingSteps.map(s => ({ ...s }));
      } else {
        seedDefaultRouting();
      }

      // Sync form fields from selected item if available
      if (prepData.items && prepData.items.length > 0) {
        const item = prepData.items[selectedItemIndex];
        if (item) {
          bomItemName = item.fabricName || 'Super Combed Pique 220';
          pharmaNo = `PH-${item.subCategory || 'POLO'}-2026-01`;
        }
      }
    } catch (err) {
      if (window.showToast) window.showToast(err.message, 'error');
    } finally {
      isLoading = false;
    }
  }

  function seedDefaultRouting() {
    const today = new Date().toISOString().slice(0, 10);
    const addDays = (d, n) => {
      const dt = new Date(d);
      dt.setDate(dt.getDate() + n);
      return dt.toISOString().slice(0, 10);
    };

    routingSteps = [
      { sequenceOrder: 1, processCode: 'Cutting', stepKind: 'INTERNAL', vendorName: 'Cutting Room Line A', durationDays: 2, plannedStartDate: today, plannedEndDate: addDays(today, 2), plannedQty: prepData?.order?.totalQty || 100, comments: 'Precision lay cutting' },
      { sequenceOrder: 2, processCode: 'Branding', stepKind: 'VENDOR', vendorName: 'Rajesh Embroidery Works', durationDays: 3, plannedStartDate: addDays(today, 3), plannedEndDate: addDays(today, 5), plannedQty: prepData?.order?.totalQty || 100, comments: 'Embellishments / embroidery' },
      { sequenceOrder: 3, processCode: 'Machine', stepKind: 'INTERNAL', vendorName: 'Stitching Floor Line 2', durationDays: 4, plannedStartDate: addDays(today, 6), plannedEndDate: addDays(today, 9), plannedQty: prepData?.order?.totalQty || 100, comments: 'Overlock + Flatlock assembly' },
      { sequenceOrder: 4, processCode: 'Finishing', stepKind: 'INTERNAL', vendorName: 'Finishing Section', durationDays: 2, plannedStartDate: addDays(today, 10), plannedEndDate: addDays(today, 11), plannedQty: prepData?.order?.totalQty || 100, comments: 'Thread trim, buttoning, steam iron' },
      { sequenceOrder: 5, processCode: 'Packing', stepKind: 'INTERNAL', vendorName: 'Packing Section', durationDays: 1, plannedStartDate: addDays(today, 12), plannedEndDate: addDays(today, 12), plannedQty: prepData?.order?.totalQty || 100, comments: 'Polybag & 7-ply master carton' }
    ];
  }

  function addMeasurementPoint() {
    measurementRows = [
      ...measurementRows,
      { point: 'New Spec Point', tolerance: '± 0.5 cm', xs: 0, s: 0, m: 0, l: 0, xl: 0, xxl: 0, xxxl: 0 }
    ];
  }

  function removeMeasurementPoint(idx) {
    measurementRows = measurementRows.filter((_, i) => i !== idx);
  }

  async function handleSavePharma() {
    if (!currentItem) return;
    try {
      isSavingPharma = true;
      await api.savePharma({
        orderId: selectedOrderId,
        orderItemId: currentItem.id,
        pharmaType,
        pharmaNo,
        changesToPharma,
        measurementChart: measurementRows,
        status: 'APPROVED'
      });
      if (window.showToast) window.showToast(`✓ Pharma ${pharmaNo} saved for ${currentItem.productName}`, 'success');
      await loadPrepDetails(selectedOrderId);
    } catch (err) {
      if (window.showToast) window.showToast(err.message, 'error');
    } finally {
      isSavingPharma = false;
    }
  }

  async function handleSaveBomLine() {
    if (!currentItem) return;
    try {
      isSavingBom = true;
      await api.saveBomLine({
        orderId: selectedOrderId,
        orderItemId: currentItem.id,
        lineType: bomLineType,
        itemName: bomItemName,
        subType: bomSubType,
        consumptionAvg: parseFloat(bomConsumptionAvg) || 0,
        wastagePct: parseFloat(bomWastagePct) || 0,
        uom: bomUom,
        unitQty: currentItem.totalQty,
        ratePerUnit: parseFloat(bomRatePerUnit) || 0,
        comments: bomComments
      });
      if (window.showToast) window.showToast(`✓ BOM requirement calculated for ${bomItemName}`, 'success');
      bomComments = '';
      await loadPrepDetails(selectedOrderId);
    } catch (err) {
      if (window.showToast) window.showToast(err.message, 'error');
    } finally {
      isSavingBom = false;
    }
  }

  async function handleDeleteBomLine(lineId) {
    if (!confirm('Are you sure you want to remove this BOM requirement line?')) return;
    try {
      await api.deleteBomLine(lineId);
      if (window.showToast) window.showToast('BOM line deleted', 'info');
      await loadPrepDetails(selectedOrderId);
    } catch (err) {
      if (window.showToast) window.showToast(err.message, 'error');
    }
  }

  function openPoModal(line) {
    poTargetBomLine = line;
    poBufferQty = line.shortfallQty > 50 ? 10.0 : 5.0;
    poVendor = line.lineType === 'FABRIC' ? 'Arvind Mills Ltd' : 'Prashant Trims & Accessories';
    poEstRate = line.ratePerUnit || (line.lineType === 'FABRIC' ? 380.0 : 15.0);
    poRequiredBy = routingSteps[0]?.plannedStartDate || prepData?.order?.clientEtd || '';
    isPoModalOpen = true;
  }

  async function handleRaisePo() {
    if (!poTargetBomLine) return;
    try {
      isRaisingPo = true;
      const res = await api.raiseProcurementPO({
        orderId: selectedOrderId,
        bomLineId: poTargetBomLine.id,
        materialName: poTargetBomLine.itemName,
        materialType: poTargetBomLine.lineType,
        uom: poTargetBomLine.uom,
        requiredQty: poTargetBomLine.shortfallQty,
        orderBufferQty: parseFloat(poBufferQty) || 0,
        vendorName: poVendor,
        estimatedRate: parseFloat(poEstRate) || 0,
        requiredByDate: poRequiredBy
      });
      if (window.showToast) window.showToast(`✓ Purchase requisition ${res.poNo} generated!`, 'success');
      isPoModalOpen = false;
      await loadPrepDetails(selectedOrderId);
    } catch (err) {
      if (window.showToast) window.showToast(err.message, 'error');
    } finally {
      isRaisingPo = false;
    }
  }

  function addRoutingStep() {
    const last = routingSteps[routingSteps.length - 1];
    const today = new Date().toISOString().slice(0, 10);
    const start = last ? last.plannedEndDate : today;
    routingSteps = [
      ...routingSteps,
      {
        sequenceOrder: routingSteps.length + 1,
        processCode: 'Finishing',
        stepKind: 'INTERNAL',
        vendorName: 'Finishing Floor',
        durationDays: 2,
        plannedStartDate: start,
        plannedEndDate: start,
        plannedQty: prepData?.order?.totalQty || 100,
        comments: ''
      }
    ];
  }

  function removeRoutingStep(idx) {
    routingSteps = routingSteps.filter((_, i) => i !== idx);
    // Re-index sequences
    routingSteps = routingSteps.map((s, i) => ({ ...s, sequenceOrder: i + 1 }));
  }

  async function handleSaveRouting() {
    try {
      isSavingSteps = true;
      await api.saveRoutingSteps(selectedOrderId, routingSteps);
      if (window.showToast) window.showToast('✓ Routing plan saved with dependency DAG', 'success');
      await loadPrepDetails(selectedOrderId);
    } catch (err) {
      if (window.showToast) window.showToast(err.message, 'error');
    } finally {
      isSavingSteps = false;
    }
  }

  function openAdvanceModal(state) {
    targetState = state;
    advanceReason = '';
    isAdvanceModalOpen = true;
  }

  async function handleAdvanceState() {
    if (!targetState) return;
    try {
      isAdvancing = true;
      const res = await api.advanceOrderState(selectedOrderId, targetState, advanceReason);
      if (window.showToast) window.showToast(`✓ ${res.message}`, 'success');
      isAdvanceModalOpen = false;
      await refreshOrderList();
      await loadPrepDetails(selectedOrderId);
    } catch (err) {
      if (window.showToast) window.showToast(err.message, 'error');
    } finally {
      isAdvancing = false;
    }
  }
</script>

<div class="prep-container">
  <!-- Top Order Selector & Status Banner -->
  <div class="card prep-header-card">
    <div class="prep-header-top">
      <div>
        <div class="prep-eyebrow">STATION 2 & 3: PRODUCTION PREP & SCHEDULING</div>
        <h2 class="prep-title">Order Preparation & Gantt Planning OS</h2>
      </div>

      <div class="order-picker-box">
        <label for="order-select-dropdown" class="picker-label">Select Order for Prep:</label>
        <select id="order-select-dropdown" class="order-select" bind:value={selectedOrderId} onchange={(e) => handleOrderSelect(e.target.value)}>
          {#each ordersList as o}
            <option value={o.id}>
              {o.productionOrderNo} ({o.clientName} - {o.totalQty} pcs) [{o.status}]
            </option>
          {/each}
        </select>
        <button class="btn btn-secondary btn-sm" onclick={() => loadPrepDetails(selectedOrderId)}>
          ↻ Refresh
        </button>
      </div>
    </div>

    {#if prepData?.order}
      <div class="order-summary-grid">
        <div class="summary-pill">
          <span class="pill-label">Production Order:</span>
          <span class="pill-value font-mono">{prepData.order.productionOrderNo}</span>
        </div>
        <div class="summary-pill">
          <span class="pill-label">Client Name:</span>
          <span class="pill-value">{prepData.order.clientName}</span>
        </div>
        <div class="summary-pill">
          <span class="pill-label">Total Quantity:</span>
          <span class="pill-value font-mono">{prepData.order.totalQty} pcs</span>
        </div>
        <div class="summary-pill">
          <span class="pill-label">Client ETD:</span>
          <span class="pill-value font-mono">{prepData.order.clientEtd || 'Not Specified'}</span>
        </div>
        <div class="summary-pill">
          <span class="pill-label">Workflow Gate:</span>
          <span class="status-badge status-{prepData.order.status.toLowerCase()}">
            {prepData.order.status}
          </span>
        </div>
      </div>

      <!-- Action Gate Bar -->
      <div class="gate-actions-bar">
        <span class="gate-label">Available Workflow Transitions:</span>
        <div class="gate-btn-group">
          {#if prepData.order.status === 'BOOKED'}
            <button class="btn btn-primary btn-sm" onclick={() => openAdvanceModal('IN_PREP')}>
              ▶ Start Prep (Move to IN_PREP)
            </button>
          {:else if prepData.order.status === 'IN_PREP'}
            <button class="btn btn-success btn-sm" onclick={() => openAdvanceModal('PREP_COMPLETE')}>
              ✓ Mark PREP_COMPLETE (Freeze BOM & Pharma)
            </button>
          {:else if prepData.order.status === 'PREP_COMPLETE'}
            <button class="btn btn-indigo btn-sm" onclick={() => openAdvanceModal('PLANNED')}>
              📅 Approve & Freeze Routing (Move to PLANNED)
            </button>
          {:else if prepData.order.status === 'PLANNED'}
            <button class="btn btn-emerald btn-sm" onclick={() => openAdvanceModal('READY_FOR_EXECUTION')}>
              🚀 Release to Shop Floor (READY_FOR_EXECUTION)
            </button>
          {/if}
        </div>
      </div>
    {/if}
  </div>

  <!-- Main Multi-Tab Workspace -->
  <div class="prep-main-layout">
    <!-- Left / Tabs Navigation -->
    <div class="tab-rail">
      <button 
        class="tab-btn" 
        class:active={activeTab === 'prep'} 
        onclick={() => activeTab = 'prep'}
      >
        <span class="tab-icon">📐</span>
        <div class="tab-text">
          <div class="tab-title">1. Pharma & Patterns</div>
          <div class="tab-sub">Tech pack cut specs & measurement charts</div>
        </div>
      </button>

      <button 
        class="tab-btn" 
        class:active={activeTab === 'bom'} 
        onclick={() => activeTab = 'bom'}
      >
        <span class="tab-icon">🧵</span>
        <div class="tab-text">
          <div class="tab-title">2. BOM & SOH Deficit</div>
          <div class="tab-sub">Requirement math, warehouse stock & POs</div>
        </div>
      </button>

      <button 
        class="tab-btn" 
        class:active={activeTab === 'plan'} 
        onclick={() => activeTab = 'plan'}
      >
        <span class="tab-icon">📅</span>
        <div class="tab-text">
          <div class="tab-title">3. Scheduling & Gantt</div>
          <div class="tab-sub">Routing steps, line vendor & ETD slack</div>
        </div>
      </button>
    </div>

    <!-- Active Tab Content Area -->
    <div class="tab-body card">
      {#if isLoading}
        <div class="loading-state">
          <div class="spinner"></div>
          <p>Loading Order Prep & Planning data from Unified Graph...</p>
        </div>
      {:else if !prepData}
        <div class="empty-state">
          <p>Please select an order above to begin preparation.</p>
        </div>
      {:else}

        <!-- Product Items Selector (Multi-Product Orders) -->
        {#if prepData.items && prepData.items.length > 1}
          <div class="product-item-selector">
            <span class="selector-label">Product Item:</span>
            <div class="item-pills">
              {#each prepData.items as it, idx}
                <button 
                  class="item-pill" 
                  class:active={selectedItemIndex === idx}
                  onclick={() => selectedItemIndex = idx}
                >
                  <span class="pill-seq">#{it.itemSequence}</span>
                  <span class="pill-name">{it.productName} ({it.colorName})</span>
                  <span class="pill-qty">{it.totalQty} pcs</span>
                  {#if it.shortfallCount > 0}
                    <span class="badge-shortfall-dot" title="{it.shortfallCount} material shortage"></span>
                  {/if}
                </button>
              {/each}
            </div>
          </div>
        {/if}

        <!-- Product Context Card -->
        {#if currentItem}
          <div class="product-context-banner">
            {#if currentItem.productImageUrl}
              <img src={currentItem.productImageUrl} alt="Product" class="product-thumb" />
            {/if}
            <div class="context-details">
              <div class="context-title">
                #{currentItem.itemSequence} {currentItem.productName} 
                <span class="color-chip">{currentItem.colorName}</span>
              </div>
              <div class="context-sub">
                Category: <strong>{currentItem.category} ({currentItem.subCategory})</strong> • 
                Order Qty: <strong>{currentItem.totalQty} pcs</strong> • 
                Fabric Master: <strong>{currentItem.fabricName || 'Standard'}</strong> • 
                Packaging: <strong>{currentItem.custPackaging || 'Polybag'}</strong>
              </div>
            </div>
          </div>
        {/if}

        <!-- ==================================================== -->
        <!-- TAB 1: PHARMA & MEASUREMENT CHARTS -->
        <!-- ==================================================== -->
        {#if activeTab === 'prep'}
          <div class="tab-section">
            <div class="section-head">
              <div>
                <h3 class="section-title">Pharma Pattern & Tech Pack Specifications</h3>
                <p class="section-desc">CAD pattern cut codes, revisions and grading measurement chart</p>
              </div>
            </div>

            <div class="form-grid">
              <div class="form-group">
                <span class="form-label-title">Pharma Spec Type <span class="required">*</span></span>
                <Combobox 
                  options={pharmaTypes} 
                  bind:value={pharmaType} 
                  placeholder="e.g. Body, Collar, Sleeve..." 
                  allowCustom={true} 
                />
              </div>

              <div class="form-group">
                <label for="pharma-code-input" class="form-label-title">Pharma / Pattern Code <span class="required">*</span></label>
                <input 
                  id="pharma-code-input"
                  type="text" 
                  class="form-control" 
                  bind:value={pharmaNo} 
                  placeholder="e.g. PH-POLO-2026-01" 
                />
              </div>

              <div class="form-group span-2">
                <label for="pharma-changes-input" class="form-label-title">Changes to Pharma / Pattern Alterations</label>
                <textarea 
                  id="pharma-changes-input"
                  class="form-control" 
                  bind:value={changesToPharma} 
                  rows="2" 
                  placeholder="Record any pattern modifications, shrinkage compensation, or grading deviations..."
                ></textarea>
              </div>
            </div>

            <!-- Measurement Chart Table -->
            <div class="measurement-chart-box">
              <div class="table-head-bar">
                <h4 class="table-title">Grading Measurement Chart (Specs in cm / inches)</h4>
                <button class="btn btn-secondary btn-sm" onclick={addMeasurementPoint}>
                  + Add Spec Point
                </button>
              </div>

              <div class="table-container">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th style="min-width: 140px;">Measurement Point</th>
                      <th>Tolerance</th>
                      <th>XS</th>
                      <th>S</th>
                      <th>M</th>
                      <th>L</th>
                      <th>XL</th>
                      <th>XXL</th>
                      <th>XXXL</th>
                      <th></th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each measurementRows as row, idx}
                      <tr>
                        <td><input type="text" class="table-input" bind:value={row.point} /></td>
                        <td><input type="text" class="table-input" style="width: 80px;" bind:value={row.tolerance} /></td>
                        <td><input type="number" step="0.1" class="table-input num-col" bind:value={row.xs} /></td>
                        <td><input type="number" step="0.1" class="table-input num-col" bind:value={row.s} /></td>
                        <td><input type="number" step="0.1" class="table-input num-col" bind:value={row.m} /></td>
                        <td><input type="number" step="0.1" class="table-input num-col" bind:value={row.l} /></td>
                        <td><input type="number" step="0.1" class="table-input num-col" bind:value={row.xl} /></td>
                        <td><input type="number" step="0.1" class="table-input num-col" bind:value={row.xxl} /></td>
                        <td><input type="number" step="0.1" class="table-input num-col" bind:value={row.xxxl} /></td>
                        <td>
                          <button class="btn-icon danger" onclick={() => removeMeasurementPoint(idx)} title="Remove Point">
                            ✕
                          </button>
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>

              <div style="margin-top: 14px; display: flex; justify-content: flex-end;">
                <button class="btn btn-primary" onclick={handleSavePharma} disabled={isSavingPharma}>
                  {isSavingPharma ? 'Saving...' : '💾 Save Pharma Spec'}
                </button>
              </div>
            </div>

            <!-- Existing Pharma Specs for this Item -->
            {#if currentItem?.pharmaPatterns && currentItem.pharmaPatterns.length > 0}
              <div class="existing-specs-box">
                <h4 class="subhead">Registered Pharma Specs for this Item ({currentItem.pharmaPatterns.length}):</h4>
                <div class="specs-grid">
                  {#each currentItem.pharmaPatterns as pp}
                    <div class="spec-card">
                      <div class="spec-card-head">
                        <span class="spec-type-pill">{pp.pharmaType}</span>
                        <span class="spec-no">{pp.pharmaNo}</span>
                        <span class="badge badge-success">{pp.status}</span>
                      </div>
                      {#if pp.changesToPharma}
                        <div class="spec-changes"><strong>Revisions:</strong> {pp.changesToPharma}</div>
                      {/if}
                      <div class="spec-meta">{pp.measurementChart?.length || 0} measurement grading point(s) recorded</div>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}
          </div>

        <!-- ==================================================== -->
        <!-- TAB 2: BOM & LIVE SOH DEFICIT ENGINE -->
        <!-- ==================================================== -->
        {:else if activeTab === 'bom'}
          <div class="tab-section">
            <div class="section-head">
              <div>
                <h3 class="section-title">Bill of Materials (BOM) & Inventory Shortfall</h3>
                <p class="section-desc">Consumption math, automatic requirement calculation & warehouse stock matching</p>
              </div>
            </div>

            <!-- New BOM Line Entry Card -->
            <div class="bom-entry-card">
              <h4 class="subhead">+ Add Material Requirement Line</h4>
              <div class="form-grid">
                <div class="form-group">
                  <span class="form-label-title">Material Type</span>
                  <select class="form-control" bind:value={bomLineType}>
                    <option value="FABRIC">Fabric (Kg / Mtr)</option>
                    <option value="ACCESSORY">Trim / Accessory (Zips, Buttons, Cords)</option>
                    <option value="PACKAGING">Packaging (Polybags, Cartons)</option>
                    <option value="LABEL">Labels (Main, Care, Size)</option>
                  </select>
                </div>

                <div class="form-group">
                  <span class="form-label-title">Material Item Name <span class="required">*</span></span>
                  <Combobox 
                    options={bomLineType === 'FABRIC' ? fabricMasters : (bomLineType === 'PACKAGING' ? packagingMasters : trimMasters)} 
                    bind:value={bomItemName} 
                    placeholder="Search material master..." 
                    allowCustom={true} 
                  />
                </div>

                <div class="form-group">
                  <label for="bom-subtype-input" class="form-label-title">Placement / Sub-Type</label>
                  <input id="bom-subtype-input" type="text" class="form-control" bind:value={bomSubType} placeholder="e.g. Body, Collar Rib, Placket..." />
                </div>

                <div class="form-group">
                  <label for="bom-avg-input" class="form-label-title">Consumption Avg (per piece)</label>
                  <input id="bom-avg-input" type="number" step="0.001" class="form-control" bind:value={bomConsumptionAvg} placeholder="e.g. 0.35 kg or 1 pc" />
                </div>

                <div class="form-group">
                  <label for="bom-wastage-input" class="form-label-title">Wastage Buffer (%)</label>
                  <input id="bom-wastage-input" type="number" step="0.1" class="form-control" bind:value={bomWastagePct} placeholder="e.g. 3.0%" />
                </div>

                <div class="form-group">
                  <span class="form-label-title">UOM</span>
                  <Combobox 
                    options={['Kgs', 'Meters', 'Pieces', 'Gross', 'Sets', 'Cartons']} 
                    bind:value={bomUom} 
                    placeholder="UOM" 
                    allowCustom={true} 
                  />
                </div>

                <div class="form-group">
                  <label for="bom-rate-input" class="form-label-title">Est. Unit Rate (₹)</label>
                  <input id="bom-rate-input" type="number" step="0.1" class="form-control" bind:value={bomRatePerUnit} placeholder="Rate ₹ per UOM" />
                </div>

                <div class="form-group span-2">
                  <label for="bom-comments-input" class="form-label-title">Comments / Technical Notes</label>
                  <input id="bom-comments-input" type="text" class="form-control" bind:value={bomComments} placeholder="Optional notes for cutting room or supplier..." />
                </div>
              </div>

              <!-- Computed Preview -->
              <div class="computed-preview-box">
                <span>Total Units: <strong>{currentItem?.totalQty || 100} pcs</strong></span>
                <span>× Avg: <strong>{bomConsumptionAvg}</strong></span>
                <span>+ Wastage: <strong>{bomWastagePct}%</strong></span>
                <span>= Required: <strong class="text-accent">{Math.round((currentItem?.totalQty || 100) * bomConsumptionAvg * (1 + bomWastagePct/100) * 100) / 100} {bomUom}</strong></span>
                <button class="btn btn-primary btn-sm" onclick={handleSaveBomLine} disabled={isSavingBom}>
                  {isSavingBom ? 'Calculating...' : '+ Add to BOM'}
                </button>
              </div>
            </div>

            <!-- Existing BOM Lines Table -->
            <div class="bom-table-box">
              <h4 class="subhead">BOM Requirements & Live Inventory Matching</h4>
              {#if !currentItem?.bomLines || currentItem.bomLines.length === 0}
                <div class="empty-hint">No BOM requirement lines added yet. Add fabrics and accessories above.</div>
              {:else}
                <div class="table-container">
                  <table class="data-table">
                    <thead>
                      <tr>
                        <th>Type</th>
                        <th>Item Material</th>
                        <th>Sub-Type</th>
                        <th>Avg / Pc</th>
                        <th>Required Qty</th>
                        <th>SOH in Warehouse</th>
                        <th>Deficit Shortfall</th>
                        <th>Status</th>
                        <th>Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {#each currentItem.bomLines as line}
                        <tr>
                          <td><span class="code-pill">{line.lineType}</span></td>
                          <td><strong>{line.itemName}</strong></td>
                          <td>{line.subType || '—'}</td>
                          <td>{line.consumptionAvg} {line.uom}</td>
                          <td><strong class="font-mono">{line.requiredQty} {line.uom}</strong></td>
                          <td>
                            <span class="font-mono" class:text-success={line.sohQty >= line.requiredQty}>
                              {line.sohQty} {line.uom}
                            </span>
                          </td>
                          <td>
                            {#if line.shortfallQty > 0}
                              <span class="shortfall-badge">
                                ⚠️ Deficit: -{line.shortfallQty} {line.uom}
                              </span>
                            {:else}
                              <span class="stock-badge">✓ In Stock</span>
                            {/if}
                          </td>
                          <td><span class="badge badge-info">{line.status}</span></td>
                          <td>
                            <div style="display:flex;gap:6px;">
                              {#if line.shortfallQty > 0}
                                <button class="btn btn-amber btn-xs" onclick={() => openPoModal(line)}>
                                  Raise PO
                                </button>
                              {/if}
                              <button class="btn-icon danger" onclick={() => handleDeleteBomLine(line.id)} title="Delete line">
                                ✕
                              </button>
                            </div>
                          </td>
                        </tr>
                      {/each}
                    </tbody>
                  </table>
                </div>
              {/if}
            </div>

            <!-- Procurement Requisitions List -->
            {#if prepData?.requisitions && prepData.requisitions.length > 0}
              <div class="requisitions-box">
                <h4 class="subhead">Raised Purchase Requisitions & POs ({prepData.requisitions.length}):</h4>
                <div class="table-container">
                  <table class="data-table">
                    <thead>
                      <tr>
                        <th>PO No</th>
                        <th>Material</th>
                        <th>Type</th>
                        <th>Required Qty</th>
                        <th>Buffer Qty</th>
                        <th>Total PO Qty</th>
                        <th>Vendor</th>
                        <th>Required By</th>
                        <th>Status</th>
                      </tr>
                    </thead>
                    <tbody>
                      {#each prepData.requisitions as r}
                        <tr>
                          <td><span class="font-mono bold">{r.poNo}</span></td>
                          <td>{r.materialName}</td>
                          <td><span class="code-pill">{r.materialType}</span></td>
                          <td>{r.requiredQty} {r.uom}</td>
                          <td>+{r.orderBufferQty} {r.uom}</td>
                          <td><strong class="font-mono">{r.totalPoQty} {r.uom}</strong></td>
                          <td>{r.vendorName}</td>
                          <td>{r.requiredByDate || '—'}</td>
                          <td><span class="badge badge-warning">{r.status}</span></td>
                        </tr>
                      {/each}
                    </tbody>
                  </table>
                </div>
              </div>
            {/if}
          </div>

        <!-- ==================================================== -->
        <!-- TAB 3: ORDER PLANNING & GANTT TIMELINE -->
        <!-- ==================================================== -->
        {:else if activeTab === 'plan'}
          <div class="tab-section">
            <div class="section-head">
              <div>
                <h3 class="section-title">Order Planning & Capacity Scheduling (Station 3)</h3>
                <p class="section-desc">Routing stages, vendor line allocations, and delivery ETD delta tracking</p>
              </div>

              <div class="head-actions">
                <button class="btn btn-secondary btn-sm" onclick={addRoutingStep}>
                  + Add Stage
                </button>
                <button class="btn btn-primary btn-sm" onclick={handleSaveRouting} disabled={isSavingSteps}>
                  {isSavingSteps ? 'Saving...' : '💾 Save Routing Plan'}
                </button>
              </div>
            </div>

            <!-- Gantt ETD Slack Indicator Card -->
            {#if prepData?.ganttSummary}
              <div class="gantt-kpi-card" class:overdue-risk={prepData.ganttSummary.isOverdueRisk}>
                <div class="kpi-group">
                  <span class="kpi-label">Earliest Start:</span>
                  <span class="kpi-value font-mono">{prepData.ganttSummary.earliestStart || '—'}</span>
                </div>
                <div class="kpi-group">
                  <span class="kpi-label">Scheduled Finish:</span>
                  <span class="kpi-value font-mono">{prepData.ganttSummary.latestFinish || '—'}</span>
                </div>
                <div class="kpi-group">
                  <span class="kpi-label">Client ETD:</span>
                  <span class="kpi-value font-mono">{prepData.ganttSummary.clientEtd || 'Not Set'}</span>
                </div>
                <div class="kpi-group">
                  <span class="kpi-label">ETD Slack Buffer:</span>
                  <span class="kpi-value font-mono" class:text-danger={prepData.ganttSummary.slackDays < 0}>
                    {prepData.ganttSummary.slackDays} days
                  </span>
                </div>
                <div class="kpi-alert-msg">
                  {prepData.ganttSummary.statusText}
                </div>
              </div>
            {/if}

            <!-- Visual Gantt Timeline Bar Diagram -->
            <div class="visual-gantt-box">
              <h4 class="subhead">Visual Sequence Timeline (Finish-to-Start DAG)</h4>
              <div class="gantt-chart-wrapper">
                {#each routingSteps as st, idx}
                  <div class="gantt-row">
                    <div class="gantt-label">
                      <span class="step-num">#{st.sequenceOrder}</span>
                      <strong>{st.processCode}</strong>
                      <span class="vendor-tag">({st.vendorName || 'In-House'})</span>
                    </div>

                    <div class="gantt-bar-area">
                      <div 
                        class="gantt-bar bar-{st.processCode.toLowerCase()}" 
                        style="margin-left: {idx * 16}%; width: {Math.max(18, st.durationDays * 12)}%;"
                      >
                        <span class="bar-text">{st.durationDays}d • {st.plannedStartDate} → {st.plannedEndDate}</span>
                      </div>
                    </div>
                  </div>
                {/each}
              </div>
            </div>

            <!-- Routing Steps Table Builder -->
            <div class="routing-table-box">
              <h4 class="subhead">Routing Step Configuration</h4>
              <div class="table-container">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th style="width: 50px;">Seq</th>
                      <th>Process Stage</th>
                      <th>Kind</th>
                      <th>Vendor / Production Line</th>
                      <th style="width: 90px;">Days</th>
                      <th>Start Date</th>
                      <th>End Date</th>
                      <th>Planned Qty</th>
                      <th>Comments</th>
                      <th></th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each routingSteps as step, idx}
                      <tr>
                        <td><strong>#{step.sequenceOrder}</strong></td>
                        <td>
                          <select class="table-select" bind:value={step.processCode}>
                            {#each processMasters as p}
                              <option value={p}>{p}</option>
                            {/each}
                          </select>
                        </td>
                        <td>
                          <select class="table-select" bind:value={step.stepKind}>
                            <option value="INTERNAL">Internal</option>
                            <option value="VENDOR">Vendor (Job-Work)</option>
                          </select>
                        </td>
                        <td>
                          <Combobox 
                            options={vendorMasters} 
                            bind:value={step.vendorName} 
                            placeholder="Line / Vendor..." 
                            allowCustom={true} 
                          />
                        </td>
                        <td>
                          <input type="number" class="table-input num-col" bind:value={step.durationDays} />
                        </td>
                        <td>
                          <input type="date" class="table-input" bind:value={step.plannedStartDate} />
                        </td>
                        <td>
                          <input type="date" class="table-input" bind:value={step.plannedEndDate} />
                        </td>
                        <td>
                          <input type="number" class="table-input num-col" bind:value={step.plannedQty} />
                        </td>
                        <td>
                          <input type="text" class="table-input" bind:value={step.comments} placeholder="Notes..." />
                        </td>
                        <td>
                          <button class="btn-icon danger" onclick={() => removeRoutingStep(idx)} title="Remove step">
                            ✕
                          </button>
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        {/if}

      {/if}
    </div>
  </div>
</div>

<!-- Raise PO Modal -->
{#if isPoModalOpen && poTargetBomLine}
  <div class="modal-backdrop" onclick={() => isPoModalOpen = false} role="presentation">
    <div class="modal-card" onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1">
      <div class="modal-head">
        <h3 class="modal-title">📦 Raise Procurement Requisition (PO)</h3>
        <button class="btn-icon" onclick={() => isPoModalOpen = false}>✕</button>
      </div>

      <div class="modal-body">
        <p class="modal-intro">
          Generate an authoritative procurement PO for shortfall material:
          <strong>{poTargetBomLine.itemName}</strong>
        </p>

        <div class="form-grid">
          <div class="form-group">
            <span class="form-label-title">Deficit Shortfall Qty</span>
            <input type="text" class="form-control" readonly value="{poTargetBomLine.shortfallQty} {poTargetBomLine.uom}" />
          </div>

          <div class="form-group">
            <label for="po-buffer-input" class="form-label-title">Buffer / Safety Stock Qty</label>
            <input id="po-buffer-input" type="number" step="0.1" class="form-control" bind:value={poBufferQty} />
          </div>

          <div class="form-group">
            <span class="form-label-title">Suggested Supplier / Vendor</span>
            <Combobox 
              options={vendorMasters} 
              bind:value={poVendor} 
              placeholder="Select supplier..." 
              allowCustom={true} 
            />
          </div>

          <div class="form-group">
            <label for="po-rate-input" class="form-label-title">Estimated Unit Rate (₹)</label>
            <input id="po-rate-input" type="number" step="0.1" class="form-control" bind:value={poEstRate} />
          </div>

          <div class="form-group span-2">
            <label for="po-req-date-input" class="form-label-title">Required By Date (Must precede Cutting)</label>
            <input id="po-req-date-input" type="date" class="form-control" bind:value={poRequiredBy} />
          </div>
        </div>

        <div class="po-summary-banner">
          Total Order PO Qty: <strong>{Math.round((poTargetBomLine.shortfallQty + (parseFloat(poBufferQty)||0))*100)/100} {poTargetBomLine.uom}</strong> •
          Est. Valuation: <strong>₹{Math.round((poTargetBomLine.shortfallQty + (parseFloat(poBufferQty)||0)) * poEstRate)}</strong>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => isPoModalOpen = false}>Cancel</button>
        <button class="btn btn-primary" onclick={handleRaisePo} disabled={isRaisingPo}>
          {isRaisingPo ? 'Generating...' : '✓ Authorize & Raise PO'}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Advance State Modal -->
{#if isAdvanceModalOpen}
  <div class="modal-backdrop" onclick={() => isAdvanceModalOpen = false} role="presentation">
    <div class="modal-card" onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1">
      <div class="modal-head">
        <h3 class="modal-title">⚡ Advance Order Workflow Gate</h3>
        <button class="btn-icon" onclick={() => isAdvanceModalOpen = false}>✕</button>
      </div>

      <div class="modal-body">
        <p>Transition order <strong>{prepData?.order?.productionOrderNo}</strong> to:</p>
        <div class="target-state-pill">{targetState}</div>

        <div class="form-group" style="margin-top: 16px;">
          <label for="advance-reason-input" class="form-label-title">Transition Approval Reason / Sign-off Notes</label>
          <textarea 
            id="advance-reason-input"
            class="form-control" 
            bind:value={advanceReason} 
            rows="3" 
            placeholder="Record review notes or gate sign-off authorization..."
          ></textarea>
        </div>
      </div>

      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => isAdvanceModalOpen = false}>Cancel</button>
        <button class="btn btn-primary" onclick={handleAdvanceState} disabled={isAdvancing}>
          {isAdvancing ? 'Validating Gate...' : 'Confirm Gate Advance'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .prep-container {
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .prep-header-card {
    padding: 20px 24px;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
  }

  .prep-header-top {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 20px;
    flex-wrap: wrap;
  }

  .prep-eyebrow {
    font-size: 0.72rem;
    font-weight: 700;
    color: var(--accent-primary);
    letter-spacing: 0.08em;
  }

  .prep-title {
    font-size: 1.4rem;
    font-weight: 800;
    margin: 4px 0 0 0;
    color: var(--text-primary);
  }

  .order-picker-box {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }

  .picker-label {
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--text-secondary);
  }

  .order-select {
    padding: 8px 12px;
    background: var(--bg-input);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    color: var(--text-primary);
    font-weight: 600;
    font-size: 0.88rem;
  }

  .order-summary-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid var(--border-color);
  }

  .summary-pill {
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    padding: 6px 12px;
    border-radius: var(--radius-sm);
    font-size: 0.85rem;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .pill-label {
    color: var(--text-muted);
  }

  .pill-value {
    color: var(--text-primary);
    font-weight: 700;
  }

  .status-badge {
    padding: 3px 8px;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
  }

  .status-booked { background: #3b82f620; color: #3b82f6; border: 1px solid #3b82f640; }
  .status-in_prep { background: #f59e0b20; color: #f59e0b; border: 1px solid #f59e0b40; }
  .status-prep_complete { background: #10b98120; color: #10b981; border: 1px solid #10b98140; }
  .status-planned { background: #6366f120; color: #6366f1; border: 1px solid #6366f140; }
  .status-ready_for_execution { background: #05966920; color: #059669; border: 1px solid #05966940; }

  .gate-actions-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 14px;
    padding: 10px 14px;
    background: var(--bg-elevated);
    border-radius: var(--radius-sm);
  }

  .gate-label {
    font-size: 0.82rem;
    font-weight: 600;
    color: var(--text-secondary);
  }

  .prep-main-layout {
    display: grid;
    grid-template-columns: 260px 1fr;
    gap: 20px;
  }

  @media (max-width: 900px) {
    .prep-main-layout {
      grid-template-columns: 1fr;
    }
  }

  .tab-rail {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .tab-btn {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 14px;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    text-align: left;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .tab-btn:hover {
    background: var(--bg-elevated);
    border-color: var(--accent-primary);
  }

  .tab-btn.active {
    background: var(--bg-elevated);
    border-color: var(--accent-primary);
    box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.2);
  }

  .tab-icon {
    font-size: 1.3rem;
  }

  .tab-title {
    font-size: 0.92rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .tab-sub {
    font-size: 0.75rem;
    color: var(--text-muted);
    margin-top: 2px;
  }

  .tab-body {
    padding: 24px;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    min-height: 550px;
  }

  .product-item-selector {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;
    padding-bottom: 14px;
    border-bottom: 1px solid var(--border-color);
  }

  .selector-label {
    font-size: 0.85rem;
    font-weight: 700;
    color: var(--text-secondary);
  }

  .item-pills {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .item-pill {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    border-radius: var(--radius-sm);
    background: var(--bg-elevated);
    border: 1px solid var(--border-color);
    cursor: pointer;
    font-size: 0.82rem;
    color: var(--text-primary);
  }

  .item-pill.active {
    background: var(--accent-primary);
    color: #fff;
    border-color: var(--accent-primary);
  }

  .badge-shortfall-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #ef4444;
  }

  .product-context-banner {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 12px 16px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    margin-bottom: 20px;
  }

  .product-thumb {
    width: 48px;
    height: 48px;
    border-radius: 6px;
    object-fit: cover;
    border: 1px solid var(--border-color);
  }

  .context-title {
    font-size: 1.05rem;
    font-weight: 800;
    color: var(--text-primary);
  }

  .context-sub {
    font-size: 0.82rem;
    color: var(--text-muted);
    margin-top: 2px;
  }

  .color-chip {
    display: inline-block;
    padding: 2px 8px;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    font-size: 0.8rem;
    margin-left: 6px;
  }

  .section-head {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 18px;
  }

  .section-title {
    font-size: 1.15rem;
    font-weight: 800;
    margin: 0;
    color: var(--text-primary);
  }

  .section-desc {
    font-size: 0.82rem;
    color: var(--text-muted);
    margin: 3px 0 0 0;
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 16px;
    margin-bottom: 16px;
  }

  .form-group.span-2 {
    grid-column: span 2;
  }

  .measurement-chart-box, .bom-table-box, .routing-table-box, .visual-gantt-box {
    margin-top: 24px;
    padding-top: 20px;
    border-top: 1px solid var(--border-color);
  }

  .table-head-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .table-title, .subhead {
    font-size: 0.95rem;
    font-weight: 700;
    color: var(--text-primary);
    margin: 0 0 10px 0;
  }

  .table-container {
    overflow-x: auto;
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
  }

  .data-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.84rem;
  }

  .data-table th {
    background: var(--bg-elevated);
    padding: 10px 12px;
    text-align: left;
    font-weight: 700;
    color: var(--text-muted);
    border-bottom: 1px solid var(--border-color);
  }

  .data-table td {
    padding: 8px 12px;
    border-bottom: 1px solid var(--border-subtle);
    color: var(--text-primary);
  }

  .table-input, .table-select {
    width: 100%;
    padding: 6px 8px;
    background: var(--bg-input);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    color: var(--text-primary);
    font-size: 0.82rem;
  }

  .table-input.num-col {
    text-align: right;
  }

  .btn-icon.danger {
    color: #ef4444;
    background: none;
    border: none;
    cursor: pointer;
    font-weight: 700;
    padding: 4px 8px;
  }

  .bom-entry-card {
    background: var(--bg-elevated);
    border: 1px solid var(--border-color);
    padding: 16px;
    border-radius: var(--radius-sm);
    margin-bottom: 20px;
  }

  .computed-preview-box {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 10px 14px;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    font-size: 0.85rem;
    flex-wrap: wrap;
    justify-content: space-between;
  }

  .text-accent { color: var(--accent-primary); font-weight: 800; font-size: 1rem; }
  .text-success { color: #10b981; font-weight: 700; }
  .text-danger { color: #ef4444; font-weight: 700; }

  .shortfall-badge {
    background: rgba(239, 68, 68, 0.12);
    color: #ef4444;
    border: 1px solid rgba(239, 68, 68, 0.3);
    padding: 3px 8px;
    border-radius: 4px;
    font-weight: 700;
    font-size: 0.78rem;
  }

  .stock-badge {
    background: rgba(16, 185, 129, 0.12);
    color: #10b981;
    border: 1px solid rgba(16, 185, 129, 0.3);
    padding: 3px 8px;
    border-radius: 4px;
    font-weight: 700;
    font-size: 0.78rem;
  }

  .gantt-kpi-card {
    display: flex;
    align-items: center;
    gap: 20px;
    padding: 14px 18px;
    background: rgba(16, 185, 129, 0.08);
    border: 1px solid rgba(16, 185, 129, 0.3);
    border-radius: var(--radius-sm);
    margin-bottom: 20px;
    flex-wrap: wrap;
  }

  .gantt-kpi-card.overdue-risk {
    background: rgba(239, 68, 68, 0.08);
    border-color: rgba(239, 68, 68, 0.3);
  }

  .kpi-group {
    display: flex;
    flex-direction: column;
  }

  .kpi-label { font-size: 0.72rem; color: var(--text-muted); font-weight: 600; text-transform: uppercase; }
  .kpi-value { font-size: 1.05rem; font-weight: 800; color: var(--text-primary); }
  .kpi-alert-msg { font-weight: 700; margin-left: auto; font-size: 0.88rem; }

  .gantt-chart-wrapper {
    display: flex;
    flex-direction: column;
    gap: 10px;
    background: var(--bg-elevated);
    padding: 16px;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border-color);
  }

  .gantt-row {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .gantt-label {
    width: 220px;
    font-size: 0.82rem;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .step-num {
    background: var(--accent-primary);
    color: #fff;
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 700;
  }

  .vendor-tag {
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .gantt-bar-area {
    flex: 1;
    position: relative;
    background: var(--bg-surface);
    height: 32px;
    border-radius: 6px;
    border: 1px solid var(--border-subtle);
    display: flex;
    align-items: center;
  }

  .gantt-bar {
    height: 24px;
    border-radius: 4px;
    display: flex;
    align-items: center;
    padding: 0 10px;
    color: #ffffff;
    font-size: 0.75rem;
    font-weight: 700;
    box-shadow: 0 2px 6px rgba(0,0,0,0.15);
  }

  .bar-cutting { background: linear-gradient(135deg, #3b82f6, #1d4ed8); }
  .bar-branding { background: linear-gradient(135deg, #8b5cf6, #6d28d9); }
  .bar-machine { background: linear-gradient(135deg, #059669, #047857); }
  .bar-finishing { background: linear-gradient(135deg, #f59e0b, #d97706); }
  .bar-packing { background: linear-gradient(135deg, #ec4899, #be185d); }

  .bar-text {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .target-state-pill {
    display: inline-block;
    padding: 6px 14px;
    border-radius: 6px;
    background: var(--accent-primary);
    color: #fff;
    font-weight: 800;
    font-size: 1rem;
    margin-top: 6px;
  }

  .po-summary-banner {
    padding: 12px;
    background: var(--bg-elevated);
    border-radius: 6px;
    margin-top: 14px;
    font-size: 0.88rem;
  }

  .loading-state, .empty-state, .empty-hint {
    padding: 40px;
    text-align: center;
    color: var(--text-muted);
  }
</style>
