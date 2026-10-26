<script>
  import { onMount } from 'svelte';
  import { api } from '../api.js';
  import Combobox from '../components/Combobox.svelte';
  import QRTagModal from '../components/QRTagModal.svelte';
  import RemnantModal from '../components/RemnantModal.svelte';
  import DiscrepancyModal from '../components/DiscrepancyModal.svelte';
  import GraphInspectorModal from '../components/GraphInspectorModal.svelte';

  let { masters = {} } = $props();

  // Active Station: '1A' (Gate Entry), '1B' (Floor Count), '1C' (QC Inspection)
  let activeStation = $state('1A');

  // Loading & Modal States
  let isLoading = $state(false);
  let isSubmitting = $state(false);
  let isQRModalOpen = $state(false);
  let qrModalData = $state({ tagList: [] });
  let isRemnantModalOpen = $state(false);
  let isDiscrepancyModalOpen = $state(false);
  let isGraphModalOpen = $state(false);
  let graphInspectNodeId = $state('');

  // Queue State
  let awaitingCountLots = $state([]);
  let awaitingQcLots = $state([]);
  let allLots = $state([]);

  // =========================================================
  // STATION 1A: GATE ENTRY STATE
  // =========================================================
  let gateParentCode = $state('');
  let gateInwardDate = $state(new Date().toISOString().split('T')[0]);
  let gateInwardType = $state('Fabric');
  let gateVendorName = $state('');
  let gateInvoiceNo = $state('');
  let gateInvoiceQty = $state('');
  let gateInvoiceAmount = $state('');
  let gateUom = $state('Kgs');
  let gateVendorProduct = $state('');
  let gateFpccProduct = $state('');
  let gateVendorColor = $state('');
  let gateFpccColor = $state('');
  let gateNoOfBundles = $state(4);
  let gateComment = $state('');

  // Conditional specs
  let gateFabricType = $state('Knitted');
  let gateInvoiceGsm = $state(180);
  let gateDia = $state(60);
  let gateComposition = $state('100% Super Combed Cotton');
  let gateFabricRate = $state('');

  let gateVendorZipName = $state('');
  let gateZipType = $state('Metal');
  let gateTapeColor = $state('');
  let gateTeethColor = $state('Antique Brass');
  let gateSlider = $state('Auto-lock');
  let gateClosure = $state('Open-end');

  let gatePackagingType = $state('Polybag');
  let gateVendorPackagingName = $state('');

  let gateTrimType = $state('Button');
  let gateTrimSize = $state('18L');
  let gateTrimRate = $state('');
  let gateVendorTrimName = $state('');

  let gateFgType = $state('Polo');
  let gateVendorFgName = $state('');

  let gateLabelType = $state('Woven');
  let gateLabelGenre = $state('Main Label');

  // Bundle stubs for Station 1A
  let gateBundles = $state([]);

  // =========================================================
  // STATION 1B: FLOOR COUNT STATE
  // =========================================================
  let floorSelectedLotCode = $state('');
  let floorLotDetail = $state(null);
  let floorInvoicedQty = $state('');
  let floorCountedQty = $state('');
  let floorLocationBin = $state('A/R01/S01');
  let floorCountedBy = $state('Storekeeper Ramesh');
  let floorComments = $state('');
  let floorHoldOnShort = $state(true);
  let floorBundles = $state([]);
  let floorSizeLines = $state([]);

  // =========================================================
  // STATION 1C: QC INSPECTION STATE
  // =========================================================
  let qcSelectedLotCode = $state('');
  let qcLotDetail = $state(null);
  let qcInspector = $state('QC Lead Priya');
  let qcBundles = $state([]);

  // =========================================================
  // MASTER DROPDOWN OPTIONS
  // =========================================================
  let vendorOptions = $derived((masters.vendors || []).map(v => ({ value: v.name, label: v.name })));
  let colorOptions = $derived((masters.colors || []).map(c => ({ value: c.display_name, label: c.display_name })));
  let locationOptions = $derived((masters.locations || []).map(l => ({ value: l.code, label: `${l.code} - ${l.display_name}` })));
  let uomOptions = $derived((masters.uoms || []).map(u => ({ value: u.code, label: u.display_name })));

  let fabricOptions = $derived((masters.fabrics || []).map(f => ({ value: f.display_name, label: f.display_name })));
  let zipOptions = $derived((masters.zips || []).map(z => ({ value: z.display_name, label: z.display_name })));
  let trimOptions = $derived((masters.trims || []).map(t => ({ value: t.display_name, label: t.display_name })));
  let packagingOptions = $derived((masters.packagings || []).map(p => ({ value: p.display_name, label: p.display_name })));
  let labelOptions = $derived((masters.labels || []).map(l => ({ value: l.display_name, label: l.display_name })));
  let productOptions = $derived((masters.products || []).map(p => ({ value: p.display_name, label: p.display_name })));

  // Type-specific product combobox options
  let fpccProductOptions = $derived.by(() => {
    switch (gateInwardType) {
      case 'Fabric': return fabricOptions;
      case 'Zips': return zipOptions;
      case 'Packaging': return packagingOptions;
      case 'Trims': return trimOptions;
      case 'Labels': return labelOptions;
      case 'Finished Goods': return productOptions;
      default: return [];
    }
  });

  // Default UOM per inward type
  $effect(() => {
    if (gateInwardType === 'Fabric') gateUom = 'Kgs';
    else if (gateInwardType === 'Zips') gateUom = 'Pcs';
    else if (gateInwardType === 'Packaging') gateUom = 'Pcs';
    else if (gateInwardType === 'Trims') gateUom = 'Gross';
    else if (gateInwardType === 'Labels') gateUom = 'Pcs';
    else if (gateInwardType === 'Finished Goods') gateUom = 'Pcs';
  });

  // Re-generate Station 1A bundle stubs when count or parentCode changes
  $effect(() => {
    const count = parseInt(gateNoOfBundles, 10) || 1;
    const pCode = gateParentCode || 'YYMMDD-A';
    const rateVal = parseFloat(gateFabricRate) || parseFloat(gateTrimRate) || 0;

    const list = [];
    for (let i = 1; i <= count; i++) {
      const existing = gateBundles[i - 1] || {};
      list.push({
        sequence: i,
        total: count,
        bundleCode: `${pCode}-${i}/${count}`,
        sizeSpec: existing.sizeSpec || (gateInwardType === 'Fabric' ? `${gateDia}" / ${gateInvoiceGsm} GSM` : ''),
        rate: existing.rate !== undefined && existing.rate !== '' ? existing.rate : rateVal,
        status: 'Pending'
      });
    }
    gateBundles = list;
  });

  onMount(async () => {
    await refreshParentCode();
    await refreshQueues();
  });

  async function refreshParentCode() {
    try {
      gateParentCode = await api.getNextParentCode();
    } catch (err) {
      console.error(err);
    }
  }

  async function refreshQueues() {
    try {
      isLoading = true;
      const lots = await api.listLots('ALL');
      allLots = lots;
      awaitingCountLots = lots.filter(l => l.stage === 'AWAITING_COUNT' || l.stage === 'DISCREPANCY_HOLD');
      awaitingQcLots = lots.filter(l => l.stage === 'AWAITING_QC' || l.stage === 'QC_COMPLETE');
    } catch (err) {
      console.error('Failed to load queues:', err);
    } finally {
      isLoading = false;
    }
  }

  // ---------------------------------------------------------
  // STATION 1A ACTIONS
  // ---------------------------------------------------------
  async function handleGateSubmit(e) {
    e.preventDefault();
    if (!gateParentCode || !gateVendorName.trim() || !gateInvoiceNo.trim() || !gateInvoiceQty) {
      if (window.showToast) window.showToast('Please enter Parent Code, Vendor, Invoice No, and positive Invoice Qty', 'error');
      return;
    }

    isSubmitting = true;
    try {
      const attrs = {};
      if (gateInwardType === 'Fabric') {
        attrs.fabricType = gateFabricType;
        attrs.invoiceGsm = parseFloat(gateInvoiceGsm) || 0;
        attrs.dia = parseFloat(gateDia) || 0;
        attrs.composition = gateComposition;
        attrs.rate = parseFloat(gateFabricRate) || 0;
        attrs.vendorFabName = gateVendorProduct;
        attrs.fpccFabName = gateFpccProduct;
      } else if (gateInwardType === 'Zips') {
        attrs.vendorZipName = gateVendorZipName || gateVendorProduct;
        attrs.zipType = gateZipType;
        attrs.tapeColor = gateTapeColor;
        attrs.teethColor = gateTeethColor;
        attrs.slider = gateSlider;
        attrs.closure = gateClosure;
      } else if (gateInwardType === 'Packaging') {
        attrs.packagingType = gatePackagingType;
        attrs.fpccPackagingName = gateFpccProduct;
        attrs.vendorPackagingName = gateVendorPackagingName || gateVendorProduct;
      } else if (gateInwardType === 'Trims') {
        attrs.trimType = gateTrimType;
        attrs.trimSize = gateTrimSize;
        attrs.trimRate = parseFloat(gateTrimRate) || 0;
        attrs.fpccTrimName = gateFpccProduct;
        attrs.vendorTrimName = gateVendorTrimName || gateVendorProduct;
      } else if (gateInwardType === 'Finished Goods') {
        attrs.finishedGoodType = gateFgType;
        attrs.vendorFgName = gateVendorFgName || gateVendorProduct;
        attrs.fpccFgName = gateFpccProduct;
      } else if (gateInwardType === 'Labels') {
        attrs.labelType = gateLabelType;
        attrs.labelGenre = gateLabelGenre;
      }

      const stubsPayload = gateBundles.map(b => ({
        sequence: b.sequence,
        sizeSpec: b.sizeSpec,
        rate: parseFloat(b.rate) || 0
      }));

      const payload = {
        parentCode: gateParentCode,
        inwardDate: gateInwardDate,
        inwardType: gateInwardType,
        vendorName: gateVendorName,
        invoiceNo: gateInvoiceNo,
        invoiceQty: parseFloat(gateInvoiceQty) || 0,
        invoiceAmount: parseFloat(gateInvoiceAmount) || 0,
        uom: gateUom,
        vendorProduct: gateVendorProduct,
        fpccProduct: gateFpccProduct,
        vendorColor: gateVendorColor,
        fpccColor: gateFpccColor,
        noOfBundles: parseInt(gateNoOfBundles, 10) || 1,
        comment: gateComment,
        attrs,
        bundles: stubsPayload
      };

      const res = await api.gateInward(payload);
      if (window.showToast) {
        window.showToast(`✓ Gate Inward ${res.parentCode} booked! (${res.bundles.length} roll stubs generated)`, 'success');
      }

      // Refresh parent code and queues
      const bookedCode = gateParentCode;
      await refreshParentCode();
      await refreshQueues();

      // Offer immediate transition to Station 1B
      if (confirm(`Lot ${bookedCode} successfully staged at dock!\n\nWould you like to switch to Station 1B (Floor Count) to weigh & reconcile this lot now?`)) {
        activeStation = '1B';
        await selectLotForFloorCount(bookedCode);
      }
    } catch (err) {
      if (window.showToast) window.showToast('Gate inward failed: ' + err.message, 'error');
    } finally {
      isSubmitting = false;
    }
  }

  // ---------------------------------------------------------
  // STATION 1B ACTIONS
  // ---------------------------------------------------------
  async function selectLotForFloorCount(pCode) {
    floorSelectedLotCode = pCode;
    if (!pCode) {
      floorLotDetail = null;
      return;
    }

    try {
      isLoading = true;
      const detail = await api.getLot(pCode);
      floorLotDetail = detail;
      floorInvoicedQty = detail.lot.invoiceQty;
      floorCountedQty = detail.lot.countedQty > 0 ? detail.lot.countedQty : detail.lot.invoiceQty;
      floorLocationBin = detail.bundles[0]?.locationBin || 'A/R01/S01';
      floorComments = detail.lot.remarks || '';

      // Initialize bundles with default split if 0
      const bList = detail.bundles.map(b => ({
        bundleCode: b.bundleCode,
        sequence: b.bundleSequence,
        countedQty: b.inwardQty > 0 ? b.inwardQty : (detail.lot.invoiceQty / (detail.bundles.length || 1)),
        locationBin: b.locationBin || 'A/R01/S01',
        measuredGsm: b.measuredGsm || detail.lot.attrs?.invoiceGsm || 180,
        measuredDia: b.measuredDia || detail.lot.attrs?.dia || 60,
        sizeSpec: b.sizeSpec || '',
        rate: b.rate || detail.lot.unit_rate || 0,
        comment: b.bundleComment || ''
      }));
      floorBundles = bList;

      // Default size lines for sized items
      if (detail.lot.inwardType === 'Finished Goods' || detail.lot.inwardType === 'Zips' || detail.lot.inwardType === 'Trims') {
        floorSizeLines = [
          { size: 'S', invoiced: 0, counted: 0, comment: '' },
          { size: 'M', invoiced: 0, counted: 0, comment: '' },
          { size: 'L', invoiced: 0, counted: 0, comment: '' },
          { size: 'XL', invoiced: 0, counted: 0, comment: '' }
        ];
      } else {
        floorSizeLines = [];
      }
    } catch (err) {
      if (window.showToast) window.showToast('Failed to load lot: ' + err.message, 'error');
    } finally {
      isLoading = false;
    }
  }

  // Reactive Floor Count Discrepancy Sensor
  let floorDiscrepancy = $derived.by(() => {
    const invQ = parseFloat(floorInvoicedQty) || 0;
    const cntQ = parseFloat(floorCountedQty) || 0;
    const unitRate = floorLotDetail?.lot?.invoiceAmount && invQ > 0 ? (floorLotDetail.lot.invoiceAmount / invQ) : 0;

    if (invQ <= 0 || cntQ <= 0) return null;

    const variance = cntQ - invQ;
    const variancePct = ((variance / invQ) * 100).toFixed(2);
    const financialExposure = Math.abs(variance * unitRate);

    if (variance < -0.05) {
      return {
        status: 'SHORTAGE_FLAGGED',
        type: 'danger',
        variance: variance.toFixed(2),
        variancePct: Math.abs(variancePct),
        exposure: financialExposure.toFixed(2),
        title: `🚨 Dock Shortfall Discrepancy: ${Math.abs(variance).toFixed(2)} ${floorLotDetail?.lot?.uom || 'Kgs'} (${Math.abs(variancePct)}% Shortage)`,
        desc: `Physical counted weight is less than billed supplier invoice. Financial debit note exposure: ₹${financialExposure.toLocaleString('en-IN', { minimumFractionDigits: 2 })}. Placing on DISCREPANCY_HOLD will freeze supplier payout.`
      };
    } else if (variance > 0.05) {
      return {
        status: 'EXCESS_FLAGGED',
        type: 'warning',
        variance: variance.toFixed(2),
        variancePct: Math.abs(variancePct),
        exposure: financialExposure.toFixed(2),
        title: `⚠️ Dock Excess Received: +${variance.toFixed(2)} ${floorLotDetail?.lot?.uom || 'Kgs'} (+${variancePct}%)`,
        desc: `Physical received exceeds billed supplier invoice. Review delivery challan and reconcile invoice.`
      };
    } else {
      return {
        status: 'MATCHED',
        type: 'success',
        variance: '0.00',
        variancePct: '0.00',
        exposure: '0.00',
        title: `✓ Perfect Balance: Physical weight matches billed supplier invoice exactly`,
        desc: `Ready to advance lot to Station 1C for QC Inspection.`
      };
    }
  });

  // Distribute total evenly across bundles
  function distributeFloorCountEvenly() {
    const total = parseFloat(floorCountedQty) || 0;
    const count = floorBundles.length;
    if (count === 0 || total <= 0) return;

    const base = parseFloat((total / count).toFixed(2));
    const list = floorBundles.map((b, i) => {
      let w = base;
      if (i === count - 1) {
        const sumPrev = base * (count - 1);
        w = parseFloat((total - sumPrev).toFixed(2));
      }
      return { ...b, countedQty: w };
    });
    floorBundles = list;
  }

  // Auto-sum bundle weights into floorCountedQty
  function autoSumFloorBundles() {
    let sum = 0;
    for (const b of floorBundles) {
      sum += parseFloat(b.countedQty) || 0;
    }
    floorCountedQty = parseFloat(sum.toFixed(2));
  }

  async function handleFloorSubmit(e) {
    e.preventDefault();
    if (!floorSelectedLotCode || !floorCountedQty) {
      if (window.showToast) window.showToast('Please select a lot and enter counted quantity', 'error');
      return;
    }

    isSubmitting = true;
    try {
      const payload = {
        parentCode: floorSelectedLotCode,
        invoicedQty: parseFloat(floorInvoicedQty) || 0,
        countedQty: parseFloat(floorCountedQty) || 0,
        locationBin: floorLocationBin,
        countedBy: floorCountedBy,
        comments: floorComments,
        holdOnShort: floorHoldOnShort,
        sizeLines: floorSizeLines,
        bundles: floorBundles.map(b => ({
          bundleCode: b.bundleCode,
          sequence: b.sequence,
          countedQty: parseFloat(b.countedQty) || 0,
          locationBin: b.locationBin || floorLocationBin,
          measuredGsm: parseFloat(b.measuredGsm) || 0,
          measuredDia: parseFloat(b.measuredDia) || 0,
          sizeSpec: b.sizeSpec,
          rate: parseFloat(b.rate) || 0,
          comment: b.comment
        }))
      };

      const res = await api.recordFloorCount(payload);
      if (window.showToast) {
        window.showToast(`✓ Floor count recorded for ${res.parentCode}! Status: ${res.status}`, 'success');
      }

      await refreshQueues();
      const reconciledCode = floorSelectedLotCode;

      if (res.stage === 'DISCREPANCY_HOLD') {
        if (window.showToast) window.showToast(`⚠️ Lot placed on DISCREPANCY_HOLD. Shortage: ${Math.abs(res.varianceQty)} ${floorLotDetail?.lot?.uom}`, 'warning');
      } else {
        if (confirm(`Lot ${reconciledCode} advanced to AWAITING_QC!\n\nWould you like to switch to Station 1C (QC Inspection) now?`)) {
          activeStation = '1C';
          await selectLotForQC(reconciledCode);
        }
      }
    } catch (err) {
      if (window.showToast) window.showToast('Floor count failed: ' + err.message, 'error');
    } finally {
      isSubmitting = false;
    }
  }

  async function handleResolveShortage(action) {
    if (!floorSelectedLotCode) return;
    const reason = prompt(`Enter resolution remarks for ${action}:`, action === 'APPROVE_SHORTAGE' ? 'Credit note approved by vendor' : 'Warehouse supervisor recount');
    if (reason === null) return;

    try {
      isSubmitting = true;
      const res = await api.resolveDiscrepancy(floorSelectedLotCode, action, reason);
      if (window.showToast) window.showToast(`✓ Discrepancy resolved: ${res.action} -> Stage: ${res.newStage}`, 'success');
      await refreshQueues();
      await selectLotForFloorCount(floorSelectedLotCode);
    } catch (err) {
      if (window.showToast) window.showToast('Resolution failed: ' + err.message, 'error');
    } finally {
      isSubmitting = false;
    }
  }

  // ---------------------------------------------------------
  // STATION 1C ACTIONS
  // ---------------------------------------------------------
  async function selectLotForQC(pCode) {
    qcSelectedLotCode = pCode;
    if (!pCode) {
      qcLotDetail = null;
      return;
    }

    try {
      isLoading = true;
      const detail = await api.getLot(pCode);
      qcLotDetail = detail;

      qcBundles = detail.bundles.map(b => ({
        id: b.id,
        bundleCode: b.bundleCode,
        sequence: b.bundleSequence,
        itemName: b.itemName,
        colorName: b.colorName,
        inwardQty: b.inwardQty,
        locationBin: b.locationBin,
        sizeSpec: b.sizeSpec,
        qcStatus: b.qcStatus === 'Pending' ? 'Passed' : b.qcStatus, // Default preview to Passed
        qcRemarks: b.qcRemarks || '',
        qcImageUrl: b.qcImageUrl || '',
        isUploading: false
      }));
    } catch (err) {
      if (window.showToast) window.showToast('Failed to load QC lot: ' + err.message, 'error');
    } finally {
      isLoading = false;
    }
  }

  function markAllQCPass() {
    qcBundles = qcBundles.map(b => ({ ...b, qcStatus: 'Passed', qcRemarks: b.qcRemarks || 'Passed standard inspection' }));
  }

  async function handleQCImageUpload(bundleIdx, file) {
    if (!file) return;
    try {
      qcBundles[bundleIdx].isUploading = true;
      const res = await api.uploadFile(file);
      qcBundles[bundleIdx].qcImageUrl = res.url;
      if (window.showToast) window.showToast('✓ Inspection photo attached!', 'success');
    } catch (err) {
      if (window.showToast) window.showToast('Photo upload failed: ' + err.message, 'error');
    } finally {
      qcBundles[bundleIdx].isUploading = false;
    }
  }

  async function handleQCSubmit(e) {
    e.preventDefault();
    if (!qcSelectedLotCode || qcBundles.length === 0) {
      if (window.showToast) window.showToast('Please select a lot to inspect', 'error');
      return;
    }

    isSubmitting = true;
    try {
      const payload = {
        parentCode: qcSelectedLotCode,
        inspector: qcInspector,
        inspections: qcBundles.map(b => ({
          bundleCode: b.bundleCode,
          result: b.qcStatus === 'Passed' ? 'PASS' : 'FAIL',
          remarks: b.qcRemarks,
          imageUrl: b.qcImageUrl,
          inspectedBy: qcInspector
        }))
      };

      const res = await api.recordBundleQC(payload);
      if (window.showToast) {
        window.showToast(`✓ QC Inspection committed! (${res.passCount} Passed, ${res.failCount} Rejected)`, 'success');
      }

      await refreshQueues();
      await selectLotForQC(qcSelectedLotCode);

      // If lot completed, offer to open Zebra thermal tag modal
      if (res.lotCompleted) {
        openAllZebraTags();
      }
    } catch (err) {
      if (window.showToast) window.showToast('QC Inspection failed: ' + err.message, 'error');
    } finally {
      isSubmitting = false;
    }
  }

  function openAllZebraTags() {
    if (!qcLotDetail || !qcBundles.length) return;
    const tags = qcBundles.map(b => ({
      bundleCode: b.bundleCode,
      item: qcLotDetail.lot.fpccProduct || qcLotDetail.lot.vendorProduct || qcLotDetail.lot.inwardType,
      color: qcLotDetail.lot.fpccColor || qcLotDetail.lot.vendorColor || '',
      qty: b.inwardQty,
      uom: qcLotDetail.lot.uom,
      sizeSpec: b.sizeSpec,
      location: b.locationBin,
      qcStatus: b.qcStatus
    }));

    qrModalData = { tagList: tags };
    isQRModalOpen = true;
  }

  function openSingleZebraTag(b) {
    qrModalData = {
      tagList: [{
        bundleCode: b.bundleCode,
        item: qcLotDetail?.lot?.fpccProduct || qcLotDetail?.lot?.vendorProduct || b.itemName,
        color: qcLotDetail?.lot?.fpccColor || qcLotDetail?.lot?.vendorColor || b.colorName,
        qty: b.inwardQty,
        uom: qcLotDetail?.lot?.uom || 'Kgs',
        sizeSpec: b.sizeSpec,
        location: b.locationBin,
        qcStatus: b.qcStatus
      }]
    };
    isQRModalOpen = true;
  }
</script>

<div class="inward-view">
  <!-- Top Navigation & Station Switcher -->
  <header class="station-header card">
    <div class="header-main">
      <div>
        <h2>Inward Intake Pipeline & QC Verification</h2>
        <p class="subtitle">Station 1A (Gate Entry) ➔ Station 1B (Floor Count Reconciliation) ➔ Station 1C (QC Inspection & Zebra Tagging)</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-secondary btn-sm" onclick={refreshQueues}>
          🔄 Refresh Queues
        </button>
        <button class="btn btn-secondary btn-sm" onclick={() => isDiscrepancyModalOpen = true}>
          ⚖️ Discrepancies
        </button>
        <button class="btn btn-secondary btn-sm" onclick={() => isRemnantModalOpen = true}>
          ✂️ Return Bits
        </button>
      </div>
    </div>

    <!-- Segmented Station Tabs -->
    <div class="station-tabs">
      <button 
        class="station-tab-btn {activeStation === '1A' ? 'active' : ''}" 
        onclick={() => activeStation = '1A'}
      >
        <span class="tab-icon">🚚</span>
        <div class="tab-text">
          <div class="tab-title">Station 1A: Gate Entry</div>
          <div class="tab-desc">Log supplier delivery & book roll stubs</div>
        </div>
      </button>

      <button 
        class="station-tab-btn {activeStation === '1B' ? 'active' : ''}" 
        onclick={() => activeStation = '1B'}
      >
        <span class="tab-icon">⚖️</span>
        <div class="tab-text">
          <div class="tab-title">
            Station 1B: Floor Count
            {#if awaitingCountLots.length > 0}
              <span class="badge badge-warning">{awaitingCountLots.length} pending</span>
            {/if}
          </div>
          <div class="tab-desc">Weigh rolls & flag dock shortages</div>
        </div>
      </button>

      <button 
        class="station-tab-btn {activeStation === '1C' ? 'active' : ''}" 
        onclick={() => activeStation = '1C'}
      >
        <span class="tab-icon">🔍</span>
        <div class="tab-text">
          <div class="tab-title">
            Station 1C: QC Inspection
            {#if awaitingQcLots.length > 0}
              <span class="badge badge-info">{awaitingQcLots.length} lots</span>
            {/if}
          </div>
          <div class="tab-desc">Binary pass/fail & Zebra QR print</div>
        </div>
      </button>
    </div>
  </header>

  <!-- ========================================================= -->
  <!-- STATION 1A: GATE ENTRY & INWARD BOOKING                   -->
  <!-- ========================================================= -->
  {#if activeStation === '1A'}
    <form onsubmit={handleGateSubmit} class="station-form">
      <!-- Universal Inward Section -->
      <section class="card section-card">
        <div class="section-header">
          <h3>1. Gate Intake & Commercial Invoice Header</h3>
          <span class="badge badge-primary">Station 1A Staging</span>
        </div>

        <div class="form-grid">
          <div class="form-group">
            <label for="gate-pcode">Parent Code (Lot Sequence)</label>
            <div class="input-with-action">
              <input id="gate-pcode" class="form-control" bind:value={gateParentCode} placeholder="YYMMDD-A" required />
              <button type="button" class="btn btn-secondary btn-sm" onclick={refreshParentCode} title="Suggest Next Parent Code">🔄</button>
            </div>
            <span class="help-text">Format: YYMMDD-A (e.g. {gateParentCode || '261026-A'})</span>
          </div>

          <div class="form-group">
            <label for="gate-date">Inward Date</label>
            <input id="gate-date" type="date" class="form-control" bind:value={gateInwardDate} required />
          </div>

          <div class="form-group">
            <label for="gate-type">Inward Type (Material Genre)</label>
            <select id="gate-type" class="form-control" bind:value={gateInwardType} required>
              <option value="Fabric">Fabric (Knitted / Woven / Rib)</option>
              <option value="Zips">Zips (Nylon / Metal / Vislon)</option>
              <option value="Packaging">Packaging (Polybag / Carton / Hanger)</option>
              <option value="Trims">Trims (Buttons / Elastic / Drawcord)</option>
              <option value="Finished Goods">Finished Goods (Polo / Hoodie / Tees)</option>
              <option value="Labels">Labels (Woven / Care / Size)</option>
            </select>
          </div>

          <div class="form-group">
            <label for="gate-vendor">Supplier / Vendor Entity</label>
            <Combobox 
              options={vendorOptions} 
              bind:value={gateVendorName} 
              placeholder="Search or enter supplier name..." 
              allowCustom={true} 
            />
          </div>

          <div class="form-group">
            <label for="gate-inv-no">Supplier Invoice No.</label>
            <input id="gate-inv-no" class="form-control" bind:value={gateInvoiceNo} placeholder="e.g. INV-9902" required />
          </div>

          <div class="form-group">
            <label for="gate-inv-qty">Billed Invoice Quantity</label>
            <div class="input-with-unit">
              <input id="gate-inv-qty" type="number" step="0.01" class="form-control" bind:value={gateInvoiceQty} placeholder="0.00" required />
              <select class="unit-select" bind:value={gateUom}>
                {#each uomOptions as u}
                  <option value={u.value}>{u.value}</option>
                {/each}
              </select>
            </div>
          </div>

          <div class="form-group">
            <label for="gate-inv-amt">Invoice Total Amount (₹)</label>
            <input id="gate-inv-amt" type="number" step="0.01" class="form-control" bind:value={gateInvoiceAmount} placeholder="₹ 0.00" />
            {#if gateInvoiceAmount && gateInvoiceQty > 0}
              <span class="help-text">Calculated Unit Rate: ₹{(parseFloat(gateInvoiceAmount)/parseFloat(gateInvoiceQty)).toFixed(2)} / {gateUom}</span>
            {/if}
          </div>

          <div class="form-group">
            <label for="gate-bundles">Number of Rolls / Bundles to Generate</label>
            <input id="gate-bundles" type="number" min="1" max="100" class="form-control" bind:value={gateNoOfBundles} required />
            <span class="help-text">Generates bundle stubs <code>{gateParentCode || 'YYMMDD-A'}-1/{gateNoOfBundles}</code> .. <code>{gateNoOfBundles}/{gateNoOfBundles}</code></span>
          </div>
        </div>

        <div class="form-grid" style="margin-top: 12px;">
          <div class="form-group">
            <label for="gate-vp">Vendor Product Name</label>
            <input id="gate-vp" class="form-control" bind:value={gateVendorProduct} placeholder="What the vendor calls this item..." />
          </div>

          <div class="form-group">
            <label for="gate-fpccp">FPCC Internal Master Product</label>
            <Combobox 
              options={fpccProductOptions} 
              bind:value={gateFpccProduct} 
              placeholder="Search internal item master..." 
              allowCustom={true} 
            />
          </div>

          <div class="form-group">
            <label for="gate-vc">Vendor Color</label>
            <input id="gate-vc" class="form-control" bind:value={gateVendorColor} placeholder="Supplier shade name..." />
          </div>

          <div class="form-group">
            <label for="gate-fpccc">FPCC Internal Color</label>
            <Combobox 
              options={colorOptions} 
              bind:value={gateFpccColor} 
              placeholder="Standard master shade..." 
              allowCustom={true} 
            />
          </div>
        </div>

        <div class="form-group" style="margin-top: 12px;">
          <label for="gate-comm">Gate Inward Remarks / Delivery Notes</label>
          <input id="gate-comm" class="form-control" bind:value={gateComment} placeholder="e.g. Delivery challan attached, truck MH-12-8802..." />
        </div>
      </section>

      <!-- Conditional Specs Section depending on Inward Type -->
      <section class="card section-card">
        <div class="section-header">
          <h3>2. Conditional Type Specifications ({gateInwardType})</h3>
          <span class="badge badge-info">Contextual Master Specs</span>
        </div>

        {#if gateInwardType === 'Fabric'}
          <div class="form-grid">
            <div class="form-group">
              <label for="gate-fab-type">Fabric Construction</label>
              <select id="gate-fab-type" class="form-control" bind:value={gateFabricType}>
                <option value="Knitted">Knitted</option>
                <option value="Woven">Woven</option>
                <option value="Single Jersey">Single Jersey</option>
                <option value="Pique">Pique</option>
                <option value="Fleece">Fleece</option>
                <option value="Rib">Rib 1x1 / 2x2</option>
                <option value="Interlock">Interlock</option>
              </select>
            </div>
            <div class="form-group">
              <label for="gate-gsm">Billed Invoice GSM</label>
              <input id="gate-gsm" type="number" class="form-control" bind:value={gateInvoiceGsm} placeholder="e.g. 180" />
            </div>
            <div class="form-group">
              <label for="gate-dia">Invoice Dia (Inches)</label>
              <input id="gate-dia" type="number" class="form-control" bind:value={gateDia} placeholder="e.g. 60" />
            </div>
            <div class="form-group">
              <label for="gate-comp">Fiber Composition</label>
              <input id="gate-comp" class="form-control" bind:value={gateComposition} placeholder="e.g. 100% Super Combed Cotton" />
            </div>
            <div class="form-group">
              <label for="gate-fab-rate">Fabric Unit Rate (₹/Kg)</label>
              <input id="gate-fab-rate" type="number" step="0.01" class="form-control" bind:value={gateFabricRate} placeholder="₹ per Kg" />
            </div>
          </div>
        {:else if gateInwardType === 'Zips'}
          <div class="form-grid">
            <div class="form-group">
              <label for="gate-zip-vname">Vendor Zip Name</label>
              <input id="gate-zip-vname" class="form-control" bind:value={gateVendorZipName} placeholder="e.g. No 5 Antique Brass" />
            </div>
            <div class="form-group">
              <label for="gate-zip-type">Zip Type</label>
              <select id="gate-zip-type" class="form-control" bind:value={gateZipType}>
                <option value="Metal">Metal</option>
                <option value="Nylon">Nylon Coil</option>
                <option value="Vislon">Vislon Molded Plastic</option>
                <option value="Invisible">Invisible Concealed</option>
                <option value="Waterproof">Waterproof PU Coated</option>
              </select>
            </div>
            <div class="form-group">
              <label for="gate-zip-slider">Slider Spec</label>
              <select id="gate-zip-slider" class="form-control" bind:value={gateSlider}>
                <option value="Auto-lock">Auto-lock</option>
                <option value="Non-lock">Non-lock</option>
                <option value="Pin-lock">Pin-lock</option>
                <option value="Reversible">Reversible</option>
              </select>
            </div>
            <div class="form-group">
              <label for="gate-zip-closure">Closure Mechanism</label>
              <select id="gate-zip-closure" class="form-control" bind:value={gateClosure}>
                <option value="Open-end">Open-end (Jacket)</option>
                <option value="Close-end">Close-end (Trousers/Pockets)</option>
                <option value="Two-way open-end">Two-way Open End</option>
              </select>
            </div>
            <div class="form-group">
              <label for="gate-zip-teeth">Teeth Finish / Color</label>
              <input id="gate-zip-teeth" class="form-control" bind:value={gateTeethColor} placeholder="e.g. Antique Brass, Gunmetal" />
            </div>
            <div class="form-group">
              <label for="gate-zip-tape">Tape Color</label>
              <input id="gate-zip-tape" class="form-control" bind:value={gateTapeColor} placeholder="Tape shade..." />
            </div>
          </div>
        {:else if gateInwardType === 'Packaging'}
          <div class="form-grid">
            <div class="form-group">
              <label for="gate-pkg-type">Packaging Type</label>
              <select id="gate-pkg-type" class="form-control" bind:value={gatePackagingType}>
                <option value="Polybag">Polybag (Self-Adhesive)</option>
                <option value="Corrugated Box">Corrugated Master Carton</option>
                <option value="Hanger">Garment Hanger</option>
                <option value="Silica Gel">Silica Gel Pouch</option>
                <option value="Barcode Sticker">Barcode / Price Tag Sticker</option>
              </select>
            </div>
            <div class="form-group">
              <label for="gate-pkg-vname">Vendor Packaging Name</label>
              <input id="gate-pkg-vname" class="form-control" bind:value={gateVendorPackagingName} placeholder="Supplier packaging name..." />
            </div>
          </div>
        {:else if gateInwardType === 'Trims'}
          <div class="form-grid">
            <div class="form-group">
              <label for="gate-trm-type">Trim Type</label>
              <select id="gate-trm-type" class="form-control" bind:value={gateTrimType}>
                <option value="Button">Button (Horn / Poly / Metal)</option>
                <option value="Rivet">Rivet</option>
                <option value="Eyelet">Eyelet</option>
                <option value="Drawcord">Drawcord / Drawstring</option>
                <option value="Elastic">Knitted / Woven Elastic</option>
                <option value="Twill Tape">Twill Tape</option>
                <option value="Badge">Silicone / Metal Badge</option>
              </select>
            </div>
            <div class="form-group">
              <label for="gate-trm-size">Size Spec</label>
              <input id="gate-trm-size" class="form-control" bind:value={gateTrimSize} placeholder="e.g. 18L, 24L, 20mm" />
            </div>
            <div class="form-group">
              <label for="gate-trm-rate">Unit Rate (₹)</label>
              <input id="gate-trm-rate" type="number" step="0.01" class="form-control" bind:value={gateTrimRate} placeholder="Rate per unit" />
            </div>
          </div>
        {:else if gateInwardType === 'Finished Goods'}
          <div class="form-grid">
            <div class="form-group">
              <label for="gate-fg-type">Garment Style Type</label>
              <select id="gate-fg-type" class="form-control" bind:value={gateFgType}>
                <option value="Polo">Classic Polo Shirt</option>
                <option value="T-Shirt">Crewneck T-Shirt</option>
                <option value="Hoodie">Pullover Hoodie</option>
                <option value="Sweatshirt">Sweatshirt</option>
                <option value="Joggers">Joggers / Sweatpants</option>
              </select>
            </div>
            <div class="form-group">
              <label for="gate-fg-vname">Vendor Style Code / Name</label>
              <input id="gate-fg-vname" class="form-control" bind:value={gateVendorFgName} placeholder="Supplier garment code..." />
            </div>
          </div>
        {:else if gateInwardType === 'Labels'}
          <div class="form-grid">
            <div class="form-group">
              <label for="gate-lbl-type">Label Substrate</label>
              <select id="gate-lbl-type" class="form-control" bind:value={gateLabelType}>
                <option value="Woven">Damask Woven</option>
                <option value="Printed">Printed Satin</option>
                <option value="Heat Transfer">Heat Transfer Tagless</option>
                <option value="Leather Patch">Leather / Jacron Patch</option>
              </select>
            </div>
            <div class="form-group">
              <label for="gate-lbl-genre">Label Function / Genre</label>
              <select id="gate-lbl-genre" class="form-control" bind:value={gateLabelGenre}>
                <option value="Main Label">Main Brand Label</option>
                <option value="Wash Care">Wash Care & Composition</option>
                <option value="Size Tab">Size Indicator Tab</option>
                <option value="Flag Label">Side Seam Flag Label</option>
                <option value="Hem Tag">Bottom Hem Woven Tag</option>
              </select>
            </div>
          </div>
        {/if}
      </section>

      <!-- Bundle Stubs Preview (Per bundle size and rate) -->
      <section class="card section-card">
        <div class="section-header">
          <h3>3. Bundle Stubs Generation Preview (N = {gateBundles.length})</h3>
          <span class="badge badge-secondary">Size & Rate per Bundle</span>
        </div>

        <p class="section-desc">
          The system will auto-generate {gateBundles.length} individual bundle stubs with format 
          <code>{gateParentCode || 'YYMMDD-A'}-1/{gateBundles.length}</code>. You can customize per-bundle size and rates below.
        </p>

        <div class="table-container">
          <table class="data-table">
            <thead>
              <tr>
                <th style="width: 80px;">Roll #</th>
                <th>Generated Bundle Code</th>
                <th>Size / Dimensions Spec</th>
                <th>Unit Rate (₹)</th>
                <th>Initial Status</th>
              </tr>
            </thead>
            <tbody>
              {#each gateBundles as b, idx}
                <tr>
                  <td><strong>#{b.sequence}</strong></td>
                  <td><span class="code-pill">{b.bundleCode}</span></td>
                  <td>
                    <input 
                      class="form-control form-control-sm" 
                      bind:value={b.sizeSpec} 
                      placeholder="Size spec e.g. 60 cm, 24L, XL..." 
                    />
                  </td>
                  <td>
                    <input 
                      type="number" 
                      step="0.01" 
                      class="form-control form-control-sm" 
                      bind:value={b.rate} 
                      placeholder="Rate ₹" 
                    />
                  </td>
                  <td>
                    <span class="badge badge-secondary">Pending (Dock Staged)</span>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>

        <div class="submit-bar">
          <button type="submit" class="btn btn-primary btn-lg" disabled={isSubmitting}>
            {#if isSubmitting}
              ⏳ Committing Gate Inward...
            {:else}
              🚚 Commit Gate Entry & Generate {gateBundles.length} Roll Stubs
            {/if}
          </button>
        </div>
      </section>
    </form>
  {/if}

  <!-- ========================================================= -->
  <!-- STATION 1B: FLOOR COUNT & RECONCILIATION                  -->
  <!-- ========================================================= -->
  {#if activeStation === '1B'}
    <div class="station-container">
      <!-- Awaiting Floor Count Queue Selector -->
      <section class="card section-card">
        <div class="section-header">
          <h3>Select Lot for Physical Floor Count</h3>
          <span class="badge badge-warning">{awaitingCountLots.length} Awaiting Count</span>
        </div>

        {#if awaitingCountLots.length === 0}
          <div class="empty-state">
            <p>🎉 No lots currently waiting for floor count! All gate arrivals have been reconciled.</p>
          </div>
        {:else}
          <div class="lot-chips-grid">
            {#each awaitingCountLots as lot}
              <button 
                type="button" 
                class="lot-chip {floorSelectedLotCode === lot.parentCode ? 'selected' : ''}"
                onclick={() => selectLotForFloorCount(lot.parentCode)}
              >
                <div class="chip-header">
                  <span class="chip-code">{lot.parentCode}</span>
                  <span class="badge {lot.stage === 'DISCREPANCY_HOLD' ? 'badge-danger' : 'badge-warning'}">
                    {lot.stage}
                  </span>
                </div>
                <div class="chip-meta">
                  <strong>{lot.vendorName}</strong> • {lot.inwardType}<br/>
                  Invoice: <strong>{lot.invoiceQty} {lot.uom}</strong> (Inv: {lot.invoiceNo})
                </div>
              </button>
            {/each}
          </div>
        {/if}
      </section>

      {#if floorLotDetail}
        <form onsubmit={handleFloorSubmit} class="station-form">
          <!-- Lot Details & Discrepancy Sensor -->
          <section class="card section-card">
            <div class="section-header">
              <div>
                <h3>Reconciliation: {floorLotDetail.lot.parentCode}</h3>
                <span class="subtitle">Supplier: {floorLotDetail.lot.vendorName} | Invoice: {floorLotDetail.lot.invoiceNo}</span>
              </div>
              <div class="chip-badges">
                <span class="badge badge-primary">{floorLotDetail.lot.inwardType}</span>
                {#if floorLotDetail.lot.attrs?.invoiceGsm}
                  <span class="badge badge-secondary">{floorLotDetail.lot.attrs.invoiceGsm} GSM</span>
                {/if}
                {#if floorLotDetail.lot.attrs?.dia}
                  <span class="badge badge-secondary">{floorLotDetail.lot.attrs.dia}" Dia</span>
                {/if}
                <button 
                  type="button" 
                  class="btn btn-secondary btn-sm" 
                  onclick={() => { graphInspectNodeId = floorLotDetail.lot.id; isGraphModalOpen = true; }}
                >
                  🕸️ Trace Graph
                </button>
              </div>
            </div>

            <!-- Invoiced vs Actual Counted Row -->
            <div class="form-grid" style="margin-top: 16px;">
              <div class="form-group">
                <label for="floor-inv-q">Official Invoiced Quantity ({floorLotDetail.lot.uom})</label>
                <input id="floor-inv-q" type="number" step="0.01" class="form-control" bind:value={floorInvoicedQty} required />
                <span class="help-text">From supplier invoice #{floorLotDetail.lot.invoiceNo}</span>
              </div>

              <div class="form-group">
                <label for="floor-act-q">Actual Physical Counted Quantity ({floorLotDetail.lot.uom})</label>
                <div class="input-with-action">
                  <input id="floor-act-q" type="number" step="0.01" class="form-control" bind:value={floorCountedQty} required />
                  <button type="button" class="btn btn-secondary btn-sm" onclick={distributeFloorCountEvenly} title="Split evenly across all rolls">
                    ➗ Split
                  </button>
                  <button type="button" class="btn btn-secondary btn-sm" onclick={autoSumFloorBundles} title="Sum from bundle table">
                    ∑ Sum
                  </button>
                </div>
                <span class="help-text">Weighed scale weight or physical tally</span>
              </div>

              <div class="form-group">
                <label for="floor-bin">Warehouse Staging Bin Location</label>
                <Combobox 
                  options={locationOptions} 
                  bind:value={floorLocationBin} 
                  placeholder="Select or enter warehouse bin e.g. A/R01/S01..." 
                  allowCustom={true} 
                />
              </div>

              <div class="form-group">
                <label for="floor-by">Storekeeper / Receiving Agent</label>
                <input id="floor-by" class="form-control" bind:value={floorCountedBy} required />
              </div>
            </div>

            <!-- Live Reactive Discrepancy Banner -->
            {#if floorDiscrepancy}
              <div class="discrepancy-banner {floorDiscrepancy.type}" style="margin-top: 16px;">
                <div class="discrepancy-title">{floorDiscrepancy.title}</div>
                <div class="discrepancy-desc">{floorDiscrepancy.desc}</div>
                {#if floorDiscrepancy.status === 'SHORTAGE_FLAGGED'}
                  <div class="discrepancy-action-row">
                    <label class="checkbox-label">
                      <input type="checkbox" bind:checked={floorHoldOnShort} />
                      <strong>Put Lot on DISCREPANCY_HOLD</strong> (Requires vendor credit note approval before proceeding to QC)
                    </label>
                  </div>
                {/if}
              </div>
            {/if}

            <!-- Shortage Hold Actions (If Lot is already in DISCREPANCY_HOLD) -->
            {#if floorLotDetail.lot.stage === 'DISCREPANCY_HOLD'}
              <div class="hold-actions-box">
                <h4>⚠️ This Lot is on DISCREPANCY_HOLD</h4>
                <p>A dock shortage discrepancy was detected. Choose an authorized resolution:</p>
                <div class="btn-group">
                  <button type="button" class="btn btn-success" onclick={() => handleResolveShortage('APPROVE_SHORTAGE')}>
                    ✓ Approve Shortage (Advance to QC)
                  </button>
                  <button type="button" class="btn btn-warning" onclick={() => handleResolveShortage('FORCE_RECOUNT')}>
                    🔄 Reset & Force Physical Recount
                  </button>
                </div>
              </div>
            {/if}
          </section>

          <!-- Per-Bundle Physical Breakdown Table -->
          <section class="card section-card">
            <div class="section-header">
              <h3>Roll Breakdown & Warehouse Bin Allocation ({floorBundles.length} Bundles)</h3>
              <div class="btn-group">
                <button type="button" class="btn btn-secondary btn-sm" onclick={distributeFloorCountEvenly}>
                  ➗ Distribute Evenly
                </button>
                <button type="button" class="btn btn-secondary btn-sm" onclick={autoSumFloorBundles}>
                  ∑ Auto-Calculate Total
                </button>
              </div>
            </div>

            <div class="table-container">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>Bundle Code</th>
                    <th>Counted Qty ({floorLotDetail.lot.uom})</th>
                    <th>Warehouse Bin Location</th>
                    {#if floorLotDetail.lot.inwardType === 'Fabric'}
                      <th>Measured GSM</th>
                      <th>Measured Dia</th>
                    {:else}
                      <th>Size Spec</th>
                      <th>Rate (₹)</th>
                    {/if}
                    <th>Roll Remarks</th>
                  </tr>
                </thead>
                <tbody>
                  {#each floorBundles as b}
                    <tr>
                      <td><span class="code-pill">{b.bundleCode}</span></td>
                      <td>
                        <input 
                          type="number" 
                          step="0.01" 
                          class="form-control form-control-sm" 
                          bind:value={b.countedQty} 
                          oninput={autoSumFloorBundles}
                          required 
                        />
                      </td>
                      <td>
                        <input 
                          class="form-control form-control-sm" 
                          bind:value={b.locationBin} 
                          placeholder="e.g. A/R01/S01" 
                          required 
                        />
                      </td>
                      {#if floorLotDetail.lot.inwardType === 'Fabric'}
                        <td>
                          <input 
                            type="number" 
                            class="form-control form-control-sm" 
                            bind:value={b.measuredGsm} 
                            placeholder="GSM" 
                          />
                        </td>
                        <td>
                          <input 
                            type="number" 
                            class="form-control form-control-sm" 
                            bind:value={b.measuredDia} 
                            placeholder="Dia" 
                          />
                        </td>
                      {:else}
                        <td>
                          <input 
                            class="form-control form-control-sm" 
                            bind:value={b.sizeSpec} 
                            placeholder="Size" 
                          />
                        </td>
                        <td>
                          <input 
                            type="number" 
                            step="0.01" 
                            class="form-control form-control-sm" 
                            bind:value={b.rate} 
                            placeholder="₹ Rate" 
                          />
                        </td>
                      {/if}
                      <td>
                        <input 
                          class="form-control form-control-sm" 
                          bind:value={b.comment} 
                          placeholder="Roll notes..." 
                        />
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>

            <div class="form-group" style="margin-top: 14px;">
              <label for="floor-comm">Floor Count Storekeeper Remarks</label>
              <textarea id="floor-comm" class="form-control" rows="2" bind:value={floorComments} placeholder="Dock count observations, scale tare weight notes..."></textarea>
            </div>

            <div class="submit-bar">
              <button type="submit" class="btn btn-primary btn-lg" disabled={isSubmitting}>
                {#if isSubmitting}
                  ⏳ Committing Floor Count...
                {:else}
                  ⚖️ Commit Floor Count & Advance to QC
                {/if}
              </button>
            </div>
          </section>
        </form>
      {/if}
    </div>
  {/if}

  <!-- ========================================================= -->
  <!-- STATION 1C: QC INSPECTION & ZEBRA TAGGING                 -->
  <!-- ========================================================= -->
  {#if activeStation === '1C'}
    <div class="station-container">
      <!-- Awaiting QC Queue Selector -->
      <section class="card section-card">
        <div class="section-header">
          <h3>Select Lot for QC Inspection</h3>
          <span class="badge badge-info">{awaitingQcLots.length} Lots Available</span>
        </div>

        {#if awaitingQcLots.length === 0}
          <div class="empty-state">
            <p>🔍 No lots currently awaiting QC inspection! Reconcile incoming lots in Station 1B first.</p>
          </div>
        {:else}
          <div class="lot-chips-grid">
            {#each awaitingQcLots as lot}
              <button 
                type="button" 
                class="lot-chip {qcSelectedLotCode === lot.parentCode ? 'selected' : ''}"
                onclick={() => selectLotForQC(lot.parentCode)}
              >
                <div class="chip-header">
                  <span class="chip-code">{lot.parentCode}</span>
                  <span class="badge {lot.stage === 'QC_COMPLETE' ? 'badge-success' : 'badge-primary'}">
                    {lot.stage}
                  </span>
                </div>
                <div class="chip-meta">
                  <strong>{lot.vendorName}</strong> • {lot.inwardType}<br/>
                  Total: <strong>{lot.countedQty || lot.invoiceQty} {lot.uom}</strong> ({lot.noOfBundles} Rolls)
                </div>
              </button>
            {/each}
          </div>
        {/if}
      </section>

      {#if qcLotDetail}
        <form onsubmit={handleQCSubmit} class="station-form">
          <!-- QC Header Dashboard -->
          <section class="card section-card">
            <div class="section-header">
              <div>
                <h3>Station 1C Quality Control: {qcLotDetail.lot.parentCode}</h3>
                <span class="subtitle">Material: {qcLotDetail.lot.fpccProduct || qcLotDetail.lot.vendorProduct} | Supplier: {qcLotDetail.lot.vendorName}</span>
              </div>
              <div class="btn-group">
                <button type="button" class="btn btn-secondary btn-sm" onclick={markAllQCPass}>
                  ✓ Quick Mark All PASS
                </button>
                <button type="button" class="btn btn-primary btn-sm" onclick={openAllZebraTags}>
                  🖨️ Print All Zebra QR Tags
                </button>
              </div>
            </div>

            <div class="form-grid" style="margin-top: 14px;">
              <div class="form-group">
                <label for="qc-insp">QC Inspector Name</label>
                <input id="qc-insp" class="form-control" bind:value={qcInspector} required />
              </div>
              <div class="form-group">
                <span style="font-size: 12px; font-weight: 600; color: var(--text-secondary); text-transform: uppercase;">Lot Status</span>
                <div style="margin-top: 8px;">
                  <span class="badge {qcLotDetail.lot.stage === 'QC_COMPLETE' ? 'badge-success' : 'badge-primary'}">
                    {qcLotDetail.lot.stage}
                  </span>
                  <span style="margin-left: 10px; font-size: 13px; color: var(--text-secondary);">
                    {qcBundles.filter(b => b.qcStatus === 'Passed').length} Passed • {qcBundles.filter(b => b.qcStatus === 'Rejected').length} Rejected
                  </span>
                </div>
              </div>
            </div>
          </section>

          <!-- Per-Roll Binary QC Cards -->
          <section class="card section-card">
            <div class="section-header">
              <h3>Per-Unit Inspection Checklist ({qcBundles.length} Bundles/Rolls)</h3>
              <span class="badge badge-secondary">Binary Pass/Fail + Photo Evidence</span>
            </div>

            <div class="qc-cards-list">
              {#each qcBundles as b, idx}
                <div class="qc-bundle-card {b.qcStatus === 'Rejected' ? 'rejected' : 'passed'}">
                  <div class="qc-bundle-header">
                    <div class="bundle-identity">
                      <span class="code-pill">{b.bundleCode}</span>
                      <span class="bundle-specs-tag">
                        <strong>{b.inwardQty} {qcLotDetail.lot.uom}</strong> • Bin: <strong>{b.locationBin}</strong>
                        {#if b.sizeSpec} • Size: {b.sizeSpec}{/if}
                      </span>
                    </div>

                    <!-- Binary Pass / Reject Switcher -->
                    <div class="binary-qc-toggle">
                      <button 
                        type="button" 
                        class="qc-toggle-btn pass {b.qcStatus === 'Passed' ? 'active' : ''}"
                        onclick={() => b.qcStatus = 'Passed'}
                      >
                        ✓ PASS
                      </button>
                      <button 
                        type="button" 
                        class="qc-toggle-btn fail {b.qcStatus === 'Rejected' ? 'active' : ''}"
                        onclick={() => b.qcStatus = 'Rejected'}
                      >
                        ✕ REJECT
                      </button>
                    </div>
                  </div>

                  <div class="qc-bundle-body">
                    <div class="qc-photo-section">
                      {#if b.qcImageUrl}
                        <div class="qc-image-preview">
                          <img src={b.qcImageUrl} alt="QC Evidence" />
                          <button type="button" class="btn-remove-photo" onclick={() => b.qcImageUrl = ''}>&times;</button>
                        </div>
                      {:else}
                        <label class="photo-upload-dropzone">
                          <input 
                            type="file" 
                            accept="image/*" 
                            style="display:none;" 
                            onchange={(e) => handleQCImageUpload(idx, e.target.files[0])} 
                          />
                          {#if b.isUploading}
                            <span>⏳ Uploading...</span>
                          {:else}
                            <span>📷 Attach QC Photo</span>
                          {/if}
                        </label>
                      {/if}
                    </div>

                    <div class="qc-remarks-section">
                      <input 
                        class="form-control" 
                        bind:value={b.qcRemarks} 
                        placeholder="Defect remarks or inspection findings..." 
                      />
                      <div class="defect-chips">
                        <button type="button" class="chip-btn" onclick={() => b.qcRemarks = 'Flawless fabric / Passed'}>Clean / OK</button>
                        <button type="button" class="chip-btn" onclick={() => b.qcRemarks = 'Minor shade variation'}>Shade variation</button>
                        <button type="button" class="chip-btn" onclick={() => b.qcRemarks = 'Knitting snags / holes'}>Knitting snags</button>
                        <button type="button" class="chip-btn" onclick={() => b.qcRemarks = 'Oil stain on selvedge'}>Oil stain</button>
                        <button type="button" class="chip-btn" onclick={() => b.qcRemarks = 'Dia measurement deviation'}>Dia deviation</button>
                      </div>
                    </div>

                    <div class="qc-tag-action">
                      <button type="button" class="btn btn-secondary btn-sm" onclick={() => openSingleZebraTag(b)}>
                        🏷️ Print Tag
                      </button>
                    </div>
                  </div>
                </div>
              {/each}
            </div>

            <div class="submit-bar">
              <button type="submit" class="btn btn-primary btn-lg" disabled={isSubmitting}>
                {#if isSubmitting}
                  ⏳ Committing QC Inspection...
                {:else}
                  🔍 Commit QC Inspection & Finalize Lot
                {/if}
              </button>
            </div>
          </section>
        </form>
      {/if}
    </div>
  {/if}
</div>

<!-- Modal Dialogs -->
<QRTagModal 
  bind:isOpen={isQRModalOpen} 
  tagList={qrModalData.tagList}
/>

<RemnantModal 
  bind:isOpen={isRemnantModalOpen} 
/>

<DiscrepancyModal 
  bind:isOpen={isDiscrepancyModalOpen} 
/>

<GraphInspectorModal 
  bind:isOpen={isGraphModalOpen} 
  nodeId={graphInspectNodeId} 
/>

<style>
  .inward-view {
    display: flex;
    flex-direction: column;
    gap: 20px;
    padding-bottom: 60px;
  }

  .station-header {
    padding: 20px;
  }

  .header-main {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    flex-wrap: wrap;
    gap: 12px;
    margin-bottom: 20px;
  }

  .subtitle {
    color: var(--text-secondary);
    font-size: 13px;
    margin-top: 4px;
  }

  .header-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .station-tabs {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 12px;
    border-top: 1px solid var(--border-color);
    padding-top: 16px;
  }

  .station-tab-btn {
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 14px 18px;
    background: var(--bg-surface);
    border: 2px solid var(--border-color);
    border-radius: var(--radius-md);
    cursor: pointer;
    text-align: left;
    transition: all 0.2s ease;
  }

  .station-tab-btn:hover {
    border-color: var(--primary-color);
    background: rgba(99, 102, 241, 0.04);
  }

  .station-tab-btn.active {
    border-color: var(--primary-color);
    background: rgba(99, 102, 241, 0.08);
    box-shadow: 0 0 0 1px var(--primary-color);
  }

  .tab-icon {
    font-size: 24px;
  }

  .tab-title {
    font-size: 14px;
    font-weight: 700;
    color: var(--text-primary);
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .tab-desc {
    font-size: 12px;
    color: var(--text-secondary);
    margin-top: 2px;
  }

  .section-card {
    padding: 20px;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    border-bottom: 1px solid var(--border-color);
    padding-bottom: 12px;
  }

  .section-header h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 700;
  }

  .section-desc {
    font-size: 13px;
    color: var(--text-secondary);
    margin-bottom: 14px;
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 14px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .form-group label {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }

  .input-with-action {
    display: flex;
    gap: 6px;
  }

  .input-with-unit {
    display: flex;
  }

  .input-with-unit input {
    border-top-right-radius: 0;
    border-bottom-right-radius: 0;
  }

  .unit-select {
    width: 90px;
    background: var(--bg-subtle);
    border: 1px solid var(--border-color);
    border-left: none;
    border-top-right-radius: var(--radius-sm);
    border-bottom-right-radius: var(--radius-sm);
    padding: 0 8px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .help-text {
    font-size: 11px;
    color: var(--text-tertiary);
  }

  .code-pill {
    font-family: var(--font-mono);
    font-size: 12px;
    font-weight: 700;
    background: var(--bg-subtle);
    padding: 4px 8px;
    border-radius: 4px;
    border: 1px solid var(--border-color);
  }

  .table-container {
    overflow-x: auto;
    margin-top: 10px;
  }

  .data-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
  }

  .data-table th, .data-table td {
    padding: 10px 12px;
    border-bottom: 1px solid var(--border-color);
    text-align: left;
  }

  .data-table th {
    background: var(--bg-subtle);
    font-weight: 600;
    color: var(--text-secondary);
    text-transform: uppercase;
    font-size: 11px;
  }

  .submit-bar {
    margin-top: 24px;
    display: flex;
    justify-content: flex-end;
  }

  /* Discrepancy Sensor */
  .discrepancy-banner {
    border-radius: var(--radius-md);
    padding: 16px 20px;
    border-left: 6px solid;
  }

  .discrepancy-banner.danger {
    background: rgba(239, 68, 68, 0.08);
    border-color: #ef4444;
    color: #991b1b;
  }

  .discrepancy-banner.warning {
    background: rgba(245, 158, 11, 0.08);
    border-color: #f59e0b;
    color: #92400e;
  }

  .discrepancy-banner.success {
    background: rgba(16, 185, 129, 0.08);
    border-color: #10b981;
    color: #065f46;
  }

  .discrepancy-title {
    font-size: 14px;
    font-weight: 800;
  }

  .discrepancy-desc {
    font-size: 12.5px;
    margin-top: 4px;
    line-height: 1.4;
  }

  .discrepancy-action-row {
    margin-top: 10px;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    cursor: pointer;
  }

  .hold-actions-box {
    margin-top: 16px;
    background: rgba(239, 68, 68, 0.06);
    border: 1px solid rgba(239, 68, 68, 0.3);
    border-radius: var(--radius-md);
    padding: 16px;
  }

  .hold-actions-box h4 {
    color: #b91c1c;
    margin: 0 0 6px 0;
  }

  .hold-actions-box p {
    font-size: 13px;
    color: var(--text-secondary);
    margin: 0 0 12px 0;
  }

  /* Lot Selector Grid */
  .lot-chips-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 12px;
  }

  .lot-chip {
    background: var(--bg-surface);
    border: 2px solid var(--border-color);
    border-radius: var(--radius-md);
    padding: 12px 14px;
    text-align: left;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .lot-chip:hover {
    border-color: var(--primary-color);
  }

  .lot-chip.selected {
    border-color: var(--primary-color);
    background: rgba(99, 102, 241, 0.06);
    box-shadow: 0 0 0 1px var(--primary-color);
  }

  .chip-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 6px;
  }

  .chip-code {
    font-family: var(--font-mono);
    font-weight: 800;
    font-size: 14px;
  }

  .chip-meta {
    font-size: 12px;
    color: var(--text-secondary);
    line-height: 1.4;
  }

  /* Station 1C QC Cards */
  .qc-cards-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .qc-bundle-card {
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    padding: 14px 16px;
    transition: all 0.2s ease;
  }

  .qc-bundle-card.rejected {
    border-left: 5px solid #ef4444;
    background: rgba(239, 68, 68, 0.02);
  }

  .qc-bundle-card.passed {
    border-left: 5px solid #10b981;
  }

  .qc-bundle-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 10px;
    margin-bottom: 12px;
  }

  .bundle-identity {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .bundle-specs-tag {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .binary-qc-toggle {
    display: flex;
    border: 1px solid var(--border-color);
    border-radius: var(--radius-sm);
    overflow: hidden;
  }

  .qc-toggle-btn {
    border: none;
    padding: 6px 16px;
    font-size: 12px;
    font-weight: 800;
    cursor: pointer;
    background: var(--bg-subtle);
    color: var(--text-secondary);
    transition: all 0.2s ease;
  }

  .qc-toggle-btn.pass.active {
    background: #10b981;
    color: #fff;
  }

  .qc-toggle-btn.fail.active {
    background: #ef4444;
    color: #fff;
  }

  .qc-bundle-body {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 16px;
    align-items: center;
  }

  .photo-upload-dropzone {
    width: 140px;
    height: 48px;
    border: 1px dashed var(--border-color);
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 11px;
    font-weight: 600;
    color: var(--text-secondary);
    cursor: pointer;
    background: var(--bg-subtle);
    transition: border-color 0.2s;
  }

  .photo-upload-dropzone:hover {
    border-color: var(--primary-color);
  }

  .qc-image-preview {
    position: relative;
    width: 54px;
    height: 54px;
  }

  .qc-image-preview img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: 4px;
    border: 1px solid var(--border-color);
  }

  .btn-remove-photo {
    position: absolute;
    top: -6px;
    right: -6px;
    background: #ef4444;
    color: #fff;
    border: none;
    border-radius: 50%;
    width: 18px;
    height: 18px;
    font-size: 12px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .defect-chips {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
    margin-top: 6px;
  }

  .chip-btn {
    border: 1px solid var(--border-color);
    background: var(--bg-subtle);
    border-radius: 12px;
    padding: 2px 8px;
    font-size: 10.5px;
    cursor: pointer;
    color: var(--text-secondary);
  }

  .chip-btn:hover {
    background: var(--primary-color);
    color: #fff;
    border-color: var(--primary-color);
  }

  .empty-state {
    text-align: center;
    padding: 30px;
    color: var(--text-secondary);
    font-size: 14px;
  }

  @media (max-width: 768px) {
    .qc-bundle-body {
      grid-template-columns: 1fr;
    }
  }
</style>
