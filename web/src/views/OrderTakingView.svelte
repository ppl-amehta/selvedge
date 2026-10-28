<script>
  import { onMount } from 'svelte';
  import { api } from '../api.js';
  import Combobox from '../components/Combobox.svelte';
  import GenderSizeMatrix from '../components/GenderSizeMatrix.svelte';
  import EmbellishmentList from '../components/EmbellishmentList.svelte';
  import PersonalizationRoster from '../components/PersonalizationRoster.svelte';
  import ProductImageUploader from '../components/ProductImageUploader.svelte';
  import MultiShippingLocations from '../components/MultiShippingLocations.svelte';
  import ZohoPullModal from '../components/ZohoPullModal.svelte';
  import CopySpecsModal from '../components/CopySpecsModal.svelte';

  let { masters = {} } = $props();

  // 1. Uniform Order Header State (Svelte 5 Runes)
  let orderNo = $state('');
  let salesOrderNo = $state('');
  let orderType = $state('Bulk');
  let bookingDate = $state(new Date().toISOString().split('T')[0]);
  let sourceChannel = $state('manual');

  let clientName = $state('');
  let contactPerson = $state('');
  let phone = $state('');
  let email = $state('');
  let gstin = $state('');
  let billingAddress = $state('');

  let taxRate = $state(5.0);
  let taxBilling = $state('Extra');
  let freightMode = $state('Surface');
  let freightCost = $state('Extra');
  let overallClientComment = $state('');

  // 2. Multi-Product Line Items
  let items = $state([createInitialProductItem()]);

  // 3. Multi-Shipping Destinations
  let isMultiShipping = $state(false);
  let shippingLocations = $state([]);

  // Modals & UI State
  let isZohoModalOpen = $state(false);
  let isCopySpecsOpen = $state(false);
  let isSubmitting = $state(false);

  const orderTypes = ['Bulk', 'Sample', 'FOC', 'Repeat', 'Development'];
  const taxRates = [5.0, 12.0, 18.0];
  const taxBillingOptions = ['Extra', 'Included'];
  const freightModes = ['Surface', 'Air', 'Courier', 'Self-Pickup'];
  const freightCosts = ['Inclusive', 'Extra', 'To be paid by client'];

  const apparelSubCategories = [
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

  const nonApparelSubCategories = [
    { display_name: 'Drinkware / Sippers' },
    { display_name: 'Bags & Backpacks' },
    { display_name: 'Caps & Headwear' },
    { display_name: 'Notebooks & Stationery' },
    { display_name: 'Lanyards & Badges' },
    { display_name: 'Tech Accessories' }
  ];

  function createInitialProductItem() {
    return {
      id: 'it_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6),
      category: 'Apparel',
      subCategory: 'Polo Shirt',
      productName: 'Classic Pique Polo',
      colorName: 'Navy Blue',
      fabricName: '100% Cotton Pique 220 GSM',
      unitPrice: 450.0,
      totalQty: 100,
      clientEtd: new Date(Date.now() + 14 * 86400000).toISOString().split('T')[0],
      clientRemarks: '',
      productImageUrl: '',
      mainLabel: 'Woven Satin Main Label',
      otherLabel: 'Heat Transfer Care Label',
      custPackaging: 'Self Adhesive Polybag',
      matrix: {
        Men: { S: 20, M: 40, L: 20 },
        Women: { S: 10, M: 10 }
      },
      brandings: [
        {
          id: 'bnd_' + Date.now(),
          brandingType: 'emb',
          position: 'Left Chest',
          size: '3.5 x 2.0 inches',
          placement: '7cm down from collar seam',
          color: 'Navy Blue & White',
          comment: 'High needle count 3D embroidery',
          brandingFileUrl: ''
        }
      ],
      personalizations: [],
      isLocked: false,
      isCollapsed: false
    };
  }

  function addProductItem() {
    // If there is an existing product item, copy its properties (fabric, color, labels, packaging, brandings, unitPrice, remarks)
    // while strictly resetting quantities to 0!
    if (items.length > 0) {
      const prev = items[items.length - 1];
      const clonedBrandings = (prev.brandings || []).map(b => ({
        ...b,
        id: 'bnd_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6),
        brandingFileUrl: b.brandingFileUrl || ''
      }));

      items = [...items, {
        id: 'it_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6),
        category: prev.category || 'Apparel',
        subCategory: prev.subCategory || 'Polo Shirt',
        productName: prev.productName || '',
        colorName: prev.colorName || '',
        fabricName: prev.fabricName || '',
        unitPrice: prev.unitPrice || 0,
        totalQty: 0, // NEVER COPY QUANTITIES: strictly 0
        clientEtd: prev.clientEtd || new Date(Date.now() + 14 * 86400000).toISOString().split('T')[0],
        clientRemarks: prev.clientRemarks || '',
        productImageUrl: prev.productImageUrl || '',
        mainLabel: prev.mainLabel || '',
        otherLabel: prev.otherLabel || '',
        custPackaging: prev.custPackaging || '',
        matrix: prev.category === 'Apparel' ? { Men: {}, Women: {} } : { General: {} },
        brandings: clonedBrandings,
        personalizations: [],
        isLocked: false,
        isCollapsed: false
      }];
      if (window.showToast) {
        window.showToast(`Added product copying specs from Item #${items.length - 1} (Quantities reset to 0)`, 'info');
      }
    } else {
      items = [createInitialProductItem()];
    }
  }

  function copyPropertiesFromItem(destIdx, srcIdx) {
    if (srcIdx < 0 || srcIdx >= items.length || srcIdx === destIdx) return;
    const src = items[srcIdx];
    const dest = items[destIdx];

    const clonedBrandings = (src.brandings || []).map(b => ({
      ...b,
      id: 'bnd_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6),
      brandingFileUrl: b.brandingFileUrl || ''
    }));

    items[destIdx] = {
      ...dest,
      category: src.category,
      subCategory: src.subCategory,
      productName: src.productName,
      colorName: src.colorName,
      fabricName: src.fabricName,
      unitPrice: src.unitPrice,
      mainLabel: src.mainLabel,
      otherLabel: src.otherLabel,
      custPackaging: src.custPackaging,
      clientRemarks: src.clientRemarks,
      brandings: clonedBrandings,
      // Keep existing destination quantities or reset to 0 if not entered
      totalQty: dest.totalQty || 0
    };

    if (window.showToast) {
      window.showToast(`Copied specs from Item #${srcIdx + 1} to Item #${destIdx + 1} (Quantities preserved)`, 'success');
    }
  }

  function onSubCategoryChange(idx, newSubCategory) {
    items[idx].subCategory = newSubCategory;
    // Check if any previous item matches this subCategory
    const matchingPrev = items.find((it, i) => i !== idx && it.subCategory === newSubCategory);
    if (matchingPrev) {
      // Auto-copy properties from matching item
      if (!items[idx].fabricName && matchingPrev.fabricName) items[idx].fabricName = matchingPrev.fabricName;
      if (!items[idx].unitPrice && matchingPrev.unitPrice) items[idx].unitPrice = matchingPrev.unitPrice;
      if (!items[idx].mainLabel && matchingPrev.mainLabel) items[idx].mainLabel = matchingPrev.mainLabel;
      if (!items[idx].otherLabel && matchingPrev.otherLabel) items[idx].otherLabel = matchingPrev.otherLabel;
      if (!items[idx].custPackaging && matchingPrev.custPackaging) items[idx].custPackaging = matchingPrev.custPackaging;
      if (items[idx].brandings.length === 0 && matchingPrev.brandings.length > 0) {
        items[idx].brandings = matchingPrev.brandings.map(b => ({
          ...b,
          id: 'bnd_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6)
        }));
      }
    }
  }

  function removeProductItem(idx) {
    if (items.length <= 1) {
      if (window.showToast) window.showToast('Order must contain at least one product item', 'warn');
      return;
    }
    items = items.filter((_, i) => i !== idx);
  }

  onMount(async () => {
    await refreshOrderNo();
  });

  async function refreshOrderNo() {
    try {
      orderNo = await api.getNextOrderNo();
    } catch (err) {
      console.error(err);
    }
  }

  function handleClientSelect(client) {
    if (client.contact_person) contactPerson = client.contact_person;
    if (client.phone) phone = client.phone;
    if (client.email) email = client.email;
    if (client.gstin) gstin = client.gstin;
    if (client.address) billingAddress = client.address;
  }

  // Handle Zoho Programmatic Order Application
  function handleApplyZoho(zohoData) {
    sourceChannel = 'zoho';
    if (zohoData.salesOrderNo) salesOrderNo = zohoData.salesOrderNo;
    if (zohoData.clientName) clientName = zohoData.clientName;
    if (zohoData.contactPerson) contactPerson = zohoData.contactPerson;
    if (zohoData.phone) phone = zohoData.phone;
    if (zohoData.email) email = zohoData.email;
    if (zohoData.gstin) gstin = zohoData.gstin;
    if (zohoData.billingAddress) billingAddress = zohoData.billingAddress;
    if (zohoData.taxRate) taxRate = zohoData.taxRate;
    if (zohoData.taxBilling) taxBilling = zohoData.taxBilling;
    if (zohoData.freightMode) freightMode = zohoData.freightMode;
    if (zohoData.freightCost) freightCost = zohoData.freightCost;
    if (zohoData.clientComment) overallClientComment = zohoData.clientComment;

    if (zohoData.items && zohoData.items.length > 0) {
      items = zohoData.items.map((it, idx) => ({
        id: 'it_zoho_' + Date.now() + '_' + idx,
        category: it.category || 'Apparel',
        subCategory: it.subCategory || 'Polo Shirt',
        productName: it.productName || 'Garment',
        colorName: it.colorName || 'Default Color',
        fabricName: '',
        unitPrice: it.unitPrice || 0,
        totalQty: it.totalQty || 0,
        clientEtd: zohoData.targetEtd || '',
        clientRemarks: '',
        productImageUrl: '',
        mainLabel: 'Standard Brand Label',
        otherLabel: 'Care Label',
        custPackaging: 'Standard Polybag',
        matrix: it.category === 'Apparel'
          ? { Men: it.sizeBreakdown || { M: it.totalQty } }
          : { General: it.sizeBreakdown || { 'Standard / Free': it.totalQty } },
        brandings: [],
        personalizations: [],
        isLocked: false,
        isCollapsed: false
      }));
    }
  }

  // Handle Specs Cloning
  function handleApplyClonedSpecs(ord) {
    clientName = ord.client_name || '';
    contactPerson = ord.contact_person || '';
    phone = ord.phone || '';
    email = ord.email || '';
    gstin = ord.gstin || '';
    billingAddress = ord.billing_address || '';
    taxRate = ord.tax_rate || 5.0;
    taxBilling = ord.tax_billing || 'Extra';
    freightMode = ord.freight_mode || 'Surface';
    freightCost = ord.freight_cost || 'Extra';

    if (ord.items && ord.items.length > 0) {
      items = ord.items.map(it => ({
        ...it,
        id: 'it_cloned_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6),
        // STRICT REQUIREMENT: DO NOT COPY QUANTITIES
        totalQty: 0,
        matrix: it.category === 'Apparel' ? { Men: {}, Women: {} } : { General: {} },
        personalizations: (it.personalizations || []).map(p => ({ ...p, qty: 0 })),
        isLocked: false,
        isCollapsed: false
      }));
    }
  }

  // Overall Aggregates & Math Lock Verification
  let grandTotalUnits = $derived.by(() => {
    return items.reduce((sum, item) => sum + (parseInt(item.totalQty, 10) || 0), 0);
  });

  let grandTaxableSubtotal = $derived.by(() => {
    return items.reduce((sum, item) => {
      const q = parseInt(item.totalQty, 10) || 0;
      const p = parseFloat(item.unitPrice) || 0.0;
      return sum + (q * p);
    }, 0);
  });

  let grandOrderValuation = $derived.by(() => {
    const sub = grandTaxableSubtotal;
    const rate = parseFloat(taxRate) || 0.0;
    if (taxBilling === 'Extra') {
      return sub + (sub * (rate / 100.0));
    }
    return sub;
  });

  // Authoritative Hard Lock Check: Any unbalanced item locks the entire order!
  let lockedItemInfo = $derived.by(() => {
    for (let i = 0; i < items.length; i++) {
      const it = items[i];
      if (it.isLocked || (parseInt(it.totalQty, 10) || 0) <= 0) {
        return { isLocked: true, index: i + 1, productName: it.productName || 'Item #' + (i + 1) };
      }
    }
    return { isLocked: false, index: -1, productName: '' };
  });

  let isMathLocked = $derived(lockedItemInfo.isLocked);

  async function handleSubmit(e) {
    e.preventDefault();

    if (!orderNo || !clientName.trim()) {
      if (window.showToast) window.showToast('Please enter Client Name and Production Order No', 'error');
      return;
    }

    if (isMathLocked) {
      if (window.showToast) {
        window.showToast(`MATH LOCK: Item #${lockedItemInfo.index} (${lockedItemInfo.productName}) has an unbalanced matrix. Balance all product quantities to proceed.`, 'error');
      }
      return;
    }

    if (grandTotalUnits <= 0) {
      if (window.showToast) window.showToast('Order total units must be greater than 0', 'error');
      return;
    }

    isSubmitting = true;
    try {
      const payload = {
        productionOrderNo: orderNo,
        salesOrderNo,
        orderType,
        bookingDate,
        sourceChannel,
        clientName,
        contactPerson,
        phone,
        email,
        gstin,
        billingAddress,
        shippingAddress: billingAddress,
        taxRate: parseFloat(taxRate) || 0.0,
        taxBilling,
        freightMode,
        freightCost,
        clientComment: overallClientComment,
        targetEtd: items[0]?.clientEtd || '',
        items: items.map((it, idx) => ({
          id: it.id,
          itemSequence: idx + 1,
          category: it.category,
          subCategory: it.subCategory,
          productName: it.productName,
          colorName: it.colorName,
          fabricName: it.fabricName,
          unitPrice: parseFloat(it.unitPrice) || 0.0,
          totalQty: parseInt(it.totalQty, 10) || 0,
          clientEtd: it.clientEtd,
          clientRemarks: it.clientRemarks,
          productImageUrl: it.productImageUrl,
          mainLabel: it.mainLabel,
          otherLabel: it.otherLabel,
          custPackaging: it.custPackaging,
          matrix: it.matrix,
          brandings: it.brandings || [],
          personalizations: it.personalizations || []
        })),
        multipleShipping: isMultiShipping,
        shippingLocations: isMultiShipping ? shippingLocations : []
      };

      const res = await api.createOrder(payload);
      if (window.showToast) {
        window.showToast(`✓ Production Order ${res.productionOrderNo} booked successfully! Graph vertices & bi-temporal edges created.`, 'success');
      }

      // Reset / reload with new order number
      await refreshOrderNo();
      salesOrderNo = '';
      sourceChannel = 'manual';
      items = [createInitialProductItem()];
      shippingLocations = [];
      isMultiShipping = false;
      overallClientComment = '';
    } catch (err) {
      if (window.showToast) {
        window.showToast('Order Booking Error: ' + err.message, 'error');
      }
    } finally {
      isSubmitting = false;
    }
  }
</script>

<div class="order-taking-page">
  <form onsubmit={handleSubmit}>
    <!-- Top Action Bar -->
    <div class="page-topbar">
      <div class="page-title-group">
        <div class="title-row">
          <span class="badge-station">Station 0</span>
          <h2>Order Taking & Commercial Specs</h2>
          {#if sourceChannel === 'zoho'}
            <span class="badge-channel zoho">✓ Zoho Synced ({salesOrderNo})</span>
          {:else}
            <span class="badge-channel manual">Manual Order Creation</span>
          {/if}
        </div>
        <p class="subtitle">Multi-Product Master-Detail Booking • Math-Checked Matrix • Bi-Temporal Graph SGi Record</p>
      </div>

      <div class="top-actions">
        <button
          type="button"
          class="btn btn-secondary zoho-btn"
          onclick={() => isZohoModalOpen = true}
        >
          📥 Pull from Zoho
        </button>

        <button
          type="button"
          class="btn btn-secondary"
          onclick={() => isCopySpecsOpen = true}
        >
          📋 Copy Past Specs
        </button>

        <button
          type="submit"
          class="btn btn-primary book-order-btn {isMathLocked ? 'locked-btn' : ''}"
          disabled={isMathLocked || isSubmitting}
        >
          {#if isSubmitting}
            <span class="spinner-sm"></span> Booking Order...
          {:else if isMathLocked}
            🔒 Math Locked (Item #{lockedItemInfo.index})
          {:else}
            ⚡ Book Production Order ({grandTotalUnits} Pcs)
          {/if}
        </button>
      </div>
    </div>

    <!-- Math Lock Warning Banner if Locked -->
    {#if isMathLocked}
      <div class="math-lock-banner">
        <div class="banner-icon">🔒</div>
        <div class="banner-text">
          <strong>ACTION LOCKED — MATH CHECKER VIOLATION:</strong>
          Product #{lockedItemInfo.index} ({lockedItemInfo.productName}) Gender-Size Matrix sum does not match the entered target total quantity. You must resolve the discrepancy before this order can be booked.
        </div>
      </div>
    {/if}

    <!-- SECTION 1: UNIFORM ORDER HEADER -->
    <div class="panel header-panel">
      <div class="panel-header">
        <div class="panel-title-wrap">
          <span class="panel-num">1</span>
          <h3>Uniform Order Header</h3>
        </div>
        <div class="header-pills">
          <span class="pill-order-no">{orderNo || 'Generating...'}</span>
          <button type="button" class="refresh-no-btn" onclick={refreshOrderNo} title="Refresh Order No">⟳</button>
        </div>
      </div>

      <div class="header-form-grid">
        <!-- Order Type -->
        <div class="form-group">
          <label for="order-type">Order Type</label>
          <select id="order-type" class="select-control" bind:value={orderType}>
            {#each orderTypes as ot}
              <option value={ot}>{ot}</option>
            {/each}
          </select>
        </div>

        <!-- Sales Order No -->
        <div class="form-group">
          <label for="sales-order-no">Sales Order No (Zoho / Client PO)</label>
          <input
            id="sales-order-no"
            type="text"
            class="input-control input-mono"
            placeholder="e.g. SO-2026-0042"
            bind:value={salesOrderNo}
          />
        </div>

        <!-- Production Order No -->
        <div class="form-group">
          <label for="prod-order-no">Production Order No (FPCC Format)</label>
          <input
            id="prod-order-no"
            type="text"
            class="input-control input-mono font-bold"
            bind:value={orderNo}
          />
        </div>

        <!-- Booking Date -->
        <div class="form-group">
          <label for="booking-date">Booking Date</label>
          <input
            id="booking-date"
            type="date"
            class="input-control"
            bind:value={bookingDate}
          />
        </div>

        <!-- Client Name Combobox -->
        <div class="form-group full-span-mobile">
          <label for="client-name-input">Client Name (Auto-suggest / Register New)</label>
          <Combobox
            id="client-name-input"
            items={masters.clients || []}
            category="client"
            displayKey="display_name"
            valueKey="display_name"
            bind:value={clientName}
            placeholder="Search client or type new..."
            onSelect={handleClientSelect}
          />
        </div>

        <!-- Contact Person -->
        <div class="form-group">
          <label for="contact-person">Contact Person</label>
          <input
            id="contact-person"
            type="text"
            class="input-control"
            placeholder="Authorized merchandiser"
            bind:value={contactPerson}
          />
        </div>

        <!-- Phone -->
        <div class="form-group">
          <label for="phone">Phone / Mobile</label>
          <input
            id="phone"
            type="tel"
            class="input-control"
            placeholder="+91..."
            bind:value={phone}
          />
        </div>

        <!-- Email -->
        <div class="form-group">
          <label for="email">Email</label>
          <input
            id="email"
            type="email"
            class="input-control"
            placeholder="orders@client.com"
            bind:value={email}
          />
        </div>

        <!-- GSTIN -->
        <div class="form-group">
          <label for="gstin">Client GSTIN</label>
          <input
            id="gstin"
            type="text"
            class="input-control input-mono"
            placeholder="29AAAAA0000A1Z5"
            bind:value={gstin}
          />
        </div>

        <!-- Billing Address -->
        <div class="form-group full-width">
          <label for="billing-address">Billing & Registered Address</label>
          <input
            id="billing-address"
            type="text"
            class="input-control"
            placeholder="Corporate HQ / Billing address line"
            bind:value={billingAddress}
          />
        </div>

        <!-- Tax Rate -->
        <div class="form-group">
          <label for="tax-rate">Tax Rate (GST)</label>
          <select id="tax-rate" class="select-control font-mono" bind:value={taxRate}>
            {#each taxRates as tr}
              <option value={tr}>{tr}% GST</option>
            {/each}
          </select>
        </div>

        <!-- Tax Billing -->
        <div class="form-group">
          <label for="tax-billing">Tax Billing</label>
          <select id="tax-billing" class="select-control" bind:value={taxBilling}>
            {#each taxBillingOptions as tb}
              <option value={tb}>{tb}</option>
            {/each}
          </select>
        </div>

        <!-- Freight Mode -->
        <div class="form-group">
          <label for="freight-mode">Freight Mode</label>
          <select id="freight-mode" class="select-control" bind:value={freightMode}>
            {#each freightModes as fm}
              <option value={fm}>{fm}</option>
            {/each}
          </select>
        </div>

        <!-- Freight Cost -->
        <div class="form-group">
          <label for="freight-cost">Freight Cost</label>
          <select id="freight-cost" class="select-control" bind:value={freightCost}>
            {#each freightCosts as fc}
              <option value={fc}>{fc}</option>
            {/each}
          </select>
        </div>

        <!-- Overall Remarks -->
        <div class="form-group full-width">
          <label for="overall-client-comment">Order-Level Merchandiser Notes / Delivery Constraints</label>
          <input
            id="overall-client-comment"
            type="text"
            class="input-control"
            placeholder="Priority client, batch dispatch required, specific gate clearance..."
            bind:value={overallClientComment}
          />
        </div>
      </div>
    </div>

    <!-- SECTION 2: MULTI-PRODUCT LINE ITEMS -->
    <div class="panel items-panel">
      <div class="panel-header">
        <div class="panel-title-wrap">
          <span class="panel-num">2</span>
          <h3>Product Line Items ({items.length} Product{items.length > 1 ? 's' : ''})</h3>
        </div>
        <button type="button" class="btn btn-secondary btn-sm add-item-btn" onclick={addProductItem}>
          + Add Product to Order
        </button>
      </div>

      <div class="products-list">
        {#each items as item, itemIdx}
          <div class="product-card {item.isLocked ? 'item-locked' : 'item-valid'}">
            <!-- Product Card Header -->
            <div class="product-card-head">
              <div class="head-left">
                <span class="product-seq-badge">Product #{itemIdx + 1}</span>
                <span class="product-title-text">{item.productName || 'Unspecified Product'}</span>
                <span class="tag-pill">{item.category}</span>
                <span class="tag-pill color-pill">{item.colorName}</span>
                <span class="units-pill">{item.totalQty} Pcs</span>
              </div>

              <div class="head-right">
                {#if item.isLocked}
                  <span class="lock-indicator">🔒 Math Mismatch</span>
                {:else}
                  <span class="valid-indicator">✓ Balanced</span>
                {/if}

                <button
                  type="button"
                  class="collapse-btn"
                  onclick={() => item.isCollapsed = !item.isCollapsed}
                  title="Toggle Section"
                >
                  {item.isCollapsed ? '▼ Expand' : '▲ Collapse'}
                </button>

                {#if items.length > 1}
                  <button
                    type="button"
                    class="remove-item-btn"
                    onclick={() => removeProductItem(itemIdx)}
                    title="Remove product line"
                  >
                    ✕ Delete Line
                  </button>
                {/if}
              </div>
            </div>

            <!-- Product Card Body (Collapsible) -->
            {#if !item.isCollapsed}
              <div class="product-card-body">
                {#if items.length > 1}
                  <div class="quick-copy-specs-banner" style="margin-bottom:14px;background:var(--bg-tertiary);padding:8px 12px;border-radius:var(--radius-sm);display:flex;align-items:center;gap:10px;font-size:12px;">
                    <span style="font-weight:600;color:var(--text-muted);">📋 Quick Spec Sync:</span>
                    <select
                      class="select-control"
                      style="width:auto;font-size:12px;padding:4px 10px;height:auto;"
                      onchange={(e) => {
                        if (e.target.value !== '') {
                          copyPropertiesFromItem(idx, parseInt(e.target.value, 10));
                          e.target.value = '';
                        }
                      }}
                    >
                      <option value="">Copy properties from another item in this order...</option>
                      {#each items as other, oIdx}
                        {#if oIdx !== idx}
                          <option value={oIdx}>
                            Item #{oIdx + 1}: {other.productName || other.subCategory} ({other.colorName || 'No Color'})
                          </option>
                        {/if}
                      {/each}
                    </select>
                    <span style="color:var(--text-subtle);font-size:11px;">(Quantities will not be overwritten)</span>
                  </div>
                {/if}

                <!-- Basic Product Attributes Grid -->
                <div class="product-fields-grid">
                  <!-- Category -->
                  <div class="form-group">
                    <label for="item-cat-{item.id}">Category</label>
                    <select
                      id="item-cat-{item.id}"
                      class="select-control"
                      bind:value={item.category}
                      onchange={() => {
                        // Switch default subcategory on category toggle
                        if (item.category === 'Non-Apparel') {
                          onSubCategoryChange(idx, 'Drinkware / Sippers');
                        } else {
                          onSubCategoryChange(idx, 'Polo Shirt');
                        }
                      }}
                    >
                      <option value="Apparel">Apparel</option>
                      <option value="Non-Apparel">Non-Apparel</option>
                    </select>
                  </div>

                  <!-- Sub-Category -->
                  <div class="form-group">
                    <label for="item-subcat-{item.id}">Sub-Category</label>
                    <Combobox
                      id="item-subcat-{item.id}"
                      items={item.category === 'Apparel' ? apparelSubCategories : nonApparelSubCategories}
                      category="subcategory"
                      displayKey="display_name"
                      valueKey="display_name"
                      bind:value={item.subCategory}
                      placeholder="Select or add sub-category..."
                      onSelect={(val) => onSubCategoryChange(idx, val?.display_name || val)}
                    />
                  </div>

                  <!-- Product Name -->
                  <div class="form-group">
                    <label for="item-pname-{item.id}">Product Name</label>
                    <Combobox
                      id="item-pname-{item.id}"
                      items={masters.products || []}
                      category="product"
                      displayKey="display_name"
                      valueKey="display_name"
                      bind:value={item.productName}
                      placeholder="Search product master..."
                    />
                  </div>

                  <!-- Color Combobox (289+ Master Colors) -->
                  <div class="form-group">
                    <label for="item-col-{item.id}">Color Name (289+ Master Colors)</label>
                    <Combobox
                      id="item-col-{item.id}"
                      items={masters.colors || []}
                      category="color"
                      displayKey="display_name"
                      valueKey="display_name"
                      bind:value={item.colorName}
                      placeholder="Type color or code..."
                    />
                  </div>

                  <!-- Fabric Name (for Apparel) -->
                  {#if item.category === 'Apparel'}
                    <div class="form-group">
                      <label for="item-fab-{item.id}">Fabric Composition (176+ Master Fabrics)</label>
                      <Combobox
                        id="item-fab-{item.id}"
                        items={masters.fabrics || []}
                        category="fabric"
                        displayKey="display_name"
                        valueKey="display_name"
                        bind:value={item.fabricName}
                        placeholder="Search fabric knit/weave..."
                      />
                    </div>
                  {/if}

                  <!-- Unit Price -->
                  <div class="form-group">
                    <label for="item-price-{item.id}">Unit Price (₹)</label>
                    <input
                      id="item-price-{item.id}"
                      type="number"
                      step="0.01"
                      min="0"
                      class="input-control numeric font-mono"
                      bind:value={item.unitPrice}
                    />
                  </div>

                  <!-- Client Requested ETD -->
                  <div class="form-group">
                    <label for="item-etd-{item.id}">Client Requested ETD</label>
                    <input
                      id="item-etd-{item.id}"
                      type="date"
                      class="input-control"
                      bind:value={item.clientEtd}
                    />
                  </div>

                  <!-- Target Total Quantity -->
                  <div class="form-group">
                    <label for="item-target-qty-{item.id}">
                      Total Quantity (Pcs) <span class="math-req">*Authoritative Math Checked</span>
                    </label>
                    <input
                      id="item-target-qty-{item.id}"
                      type="number"
                      min="1"
                      class="input-control numeric font-mono font-bold"
                      bind:value={item.totalQty}
                    />
                  </div>

                  <!-- Product Image Uploader -->
                  <div class="form-group full-width">
                    <ProductImageUploader
                      bind:imageUrl={item.productImageUrl}
                      label="Product Design / Techpack Reference Image"
                    />
                  </div>

                  <!-- Item Text Remarks -->
                  <div class="form-group full-width">
                    <label for="item-remarks-{item.id}">Client Text Remarks / Line Instructions</label>
                    <input
                      id="item-remarks-{item.id}"
                      type="text"
                      class="input-control"
                      placeholder="Fit comments, wash instructions, silicone softener..."
                      bind:value={item.clientRemarks}
                    />
                  </div>
                </div>

                <!-- GENDER-SIZE-QTY MATRIX WITH REAL-TIME MATH CHECKER -->
                <div class="matrix-section-wrap">
                  <GenderSizeMatrix
                    category={item.category}
                    bind:totalQty={item.totalQty}
                    bind:matrix={item.matrix}
                    unitPrice={item.unitPrice}
                    bind:isLocked={item.isLocked}
                  />
                </div>

                <!-- BRANDING & EMBELLISHMENTS (MULTIPLE PER PRODUCT) -->
                <div class="brandings-section-wrap">
                  <EmbellishmentList bind:brandings={item.brandings} />
                </div>

                <!-- LABELS & PACKAGING -->
                <div class="trims-packaging-grid">
                  <div class="form-group">
                    <label for="item-main-label-{item.id}">Main Label</label>
                    <input
                      id="item-main-label-{item.id}"
                      type="text"
                      class="input-control"
                      placeholder="Woven Satin / Heat Transfer..."
                      bind:value={item.mainLabel}
                    />
                  </div>

                  <div class="form-group">
                    <label for="item-other-label-{item.id}">Other Labels (Care / Size / Flag)</label>
                    <input
                      id="item-other-label-{item.id}"
                      type="text"
                      class="input-control"
                      placeholder="Wash care label, size pip..."
                      bind:value={item.otherLabel}
                    />
                  </div>

                  <div class="form-group">
                    <label for="item-pkg-{item.id}">Customer Packaging Specs</label>
                    <input
                      id="item-pkg-{item.id}"
                      type="text"
                      class="input-control"
                      placeholder="Self adhesive polybag, barcode sticker..."
                      bind:value={item.custPackaging}
                    />
                  </div>
                </div>

                <!-- PER-PIECE PERSONALIZATION ROSTER (STRICT 1 NAME + 1 NUMBER & 4-COLUMN IMPORT) -->
                <div class="personalization-section-wrap">
                  <PersonalizationRoster bind:personalization={item.personalizations} />
                </div>
              </div>
            {/if}
          </div>
        {/each}
      </div>
    </div>

    <!-- SECTION 3: MULTIPLE SHIPPING DESTINATIONS LINKED TO PRODUCTS -->
    <div class="panel shipping-panel">
      <div class="panel-header">
        <div class="panel-title-wrap">
          <span class="panel-num">3</span>
          <h3>Logistics & Multi-Shipping Destinations</h3>
        </div>
      </div>

      <MultiShippingLocations
        bind:isMultiShipping={isMultiShipping}
        bind:shippingLocations={shippingLocations}
        items={items}
        orderTotalQty={grandTotalUnits}
        defaultFreightMode={freightMode}
      />
    </div>

    <!-- SECTION 4: COMMERCIAL SUMMARY & SUBMISSION FOOTER -->
    <div class="commercial-footer-bar">
      <div class="summary-pills">
        <div class="summary-card">
          <span class="sc-label">Total Lines</span>
          <span class="sc-val">{items.length} Products</span>
        </div>
        <div class="summary-card">
          <span class="sc-label">Grand Total Units</span>
          <span class="sc-val">{grandTotalUnits} Pcs</span>
        </div>
        <div class="summary-card">
          <span class="sc-label">Taxable Base</span>
          <span class="sc-val font-mono">₹{grandTaxableSubtotal.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
        </div>
        <div class="summary-card highlight">
          <span class="sc-label">Grand Order Valuation ({taxBilling} GST {taxRate}%)</span>
          <span class="sc-val font-mono">₹{grandOrderValuation.toLocaleString('en-IN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</span>
        </div>
      </div>

      <div class="submit-action-wrap">
        {#if isMathLocked}
          <div class="lock-tip">
            🔒 Locked on Product #{lockedItemInfo.index}
          </div>
        {/if}
        <button
          type="submit"
          class="btn btn-primary submit-final-btn {isMathLocked ? 'locked' : ''}"
          disabled={isMathLocked || isSubmitting}
        >
          {#if isSubmitting}
            <span class="spinner-sm"></span> Booking Order...
          {:else if isMathLocked}
            🔒 Action Locked: Balance Item #{lockedItemInfo.index}
          {:else}
            ⚡ Book Production Order ({grandTotalUnits} Pcs)
          {/if}
        </button>
      </div>
    </div>
  </form>
</div>

<!-- Zoho Pull Modal -->
<ZohoPullModal
  bind:isOpen={isZohoModalOpen}
  onApplyOrder={handleApplyZoho}
/>

<!-- Copy Past Specs Modal -->
<CopySpecsModal
  bind:isOpen={isCopySpecsOpen}
  onCopySpecs={handleApplyClonedSpecs}
/>

<style>
  .order-taking-page {
    padding: 24px;
    max-width: 1400px;
    margin: 0 auto;
  }

  .page-topbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    flex-wrap: wrap;
    gap: 14px;
  }

  .title-row {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .badge-station {
    font-size: 10px;
    font-weight: 800;
    text-transform: uppercase;
    background: linear-gradient(135deg, #2563eb, #6366f1);
    color: #fff;
    padding: 4px 8px;
    border-radius: var(--radius-sm);
  }

  .badge-channel {
    font-size: 11px;
    font-weight: 700;
    padding: 3px 8px;
    border-radius: 9999px;
  }

  .badge-channel.zoho {
    background: rgba(16, 185, 129, 0.15);
    border: 1px solid rgba(16, 185, 129, 0.3);
    color: var(--accent-emerald);
  }

  .badge-channel.manual {
    background: rgba(255, 255, 255, 0.08);
    color: var(--text-muted);
  }

  .page-title-group h2 {
    font-size: 20px;
    font-weight: 800;
    letter-spacing: -0.02em;
    color: var(--text-main);
  }

  .subtitle {
    font-size: 12px;
    color: var(--text-muted);
    margin-top: 2px;
  }

  .top-actions {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .zoho-btn {
    border-color: rgba(37, 99, 235, 0.4);
    color: #93c5fd;
  }

  .zoho-btn:hover {
    background: rgba(37, 99, 235, 0.15);
  }

  .math-lock-banner {
    display: flex;
    align-items: center;
    gap: 12px;
    background: rgba(244, 63, 94, 0.12);
    border: 1.5px solid var(--accent-rose);
    border-radius: var(--radius-md);
    padding: 12px 18px;
    margin-bottom: 20px;
    color: #fca5a5;
    font-size: 13px;
  }

  .banner-icon {
    font-size: 20px;
  }

  .panel {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg);
    padding: 22px;
    margin-bottom: 24px;
    box-shadow: var(--shadow-panel);
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 18px;
    padding-bottom: 12px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .panel-title-wrap {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .panel-num {
    width: 22px;
    height: 22px;
    background: var(--accent-primary);
    color: #fff;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 11px;
    font-weight: 800;
  }

  .panel-header h3 {
    font-size: 15px;
    font-weight: 700;
    color: var(--text-main);
  }

  .header-pills {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .pill-order-no {
    font-family: var(--font-mono);
    font-size: 13px;
    font-weight: 800;
    color: #60a5fa;
    background: rgba(37, 99, 235, 0.12);
    border: 1px solid rgba(37, 99, 235, 0.3);
    padding: 4px 12px;
    border-radius: var(--radius-sm);
  }

  .refresh-no-btn {
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    color: var(--text-muted);
    width: 28px;
    height: 28px;
    border-radius: var(--radius-sm);
    cursor: pointer;
  }

  .header-form-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 14px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 5px;
  }

  .form-group label {
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
  }

  .full-width {
    grid-column: 1 / -1;
  }

  .font-bold { font-weight: 700; }
  .math-req { color: var(--accent-rose); font-size: 10px; text-transform: none; }

  /* Products List */
  .products-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .product-card {
    background: var(--bg-elevated);
    border: 1.5px solid var(--border-subtle);
    border-radius: var(--radius-md);
    overflow: hidden;
    transition: all 0.2s ease;
  }

  .product-card.item-locked {
    border-color: rgba(244, 63, 94, 0.5);
  }

  .product-card.item-valid {
    border-color: rgba(16, 185, 129, 0.4);
  }

  .product-card-head {
    padding: 12px 18px;
    background: rgba(255, 255, 255, 0.02);
    border-bottom: 1px solid var(--border-subtle);
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 10px;
  }

  .head-left {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .product-seq-badge {
    font-size: 11px;
    font-weight: 800;
    text-transform: uppercase;
    background: var(--accent-primary);
    color: #fff;
    padding: 3px 8px;
    border-radius: 4px;
  }

  .product-title-text {
    font-size: 14px;
    font-weight: 700;
    color: var(--text-main);
  }

  .tag-pill {
    font-size: 10.5px;
    background: var(--bg-input);
    border: 1px solid var(--border-subtle);
    padding: 2px 8px;
    border-radius: 4px;
    color: var(--text-muted);
  }

  .color-pill {
    color: #93c5fd;
    border-color: rgba(37, 99, 235, 0.3);
  }

  .units-pill {
    font-family: var(--font-mono);
    font-size: 11.5px;
    font-weight: 700;
    color: var(--accent-emerald);
  }

  .head-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .lock-indicator {
    font-size: 11px;
    font-weight: 700;
    color: var(--accent-rose);
    background: rgba(244, 63, 94, 0.12);
    border: 1px solid rgba(244, 63, 94, 0.3);
    padding: 2px 8px;
    border-radius: 4px;
  }

  .valid-indicator {
    font-size: 11px;
    font-weight: 700;
    color: var(--accent-emerald);
    background: rgba(16, 185, 129, 0.12);
    border: 1px solid rgba(16, 185, 129, 0.3);
    padding: 2px 8px;
    border-radius: 4px;
  }

  .collapse-btn {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    color: var(--text-muted);
    font-size: 11px;
    padding: 4px 10px;
    border-radius: 4px;
    cursor: pointer;
  }

  .remove-item-btn {
    background: transparent;
    border: 1px solid rgba(244, 63, 94, 0.3);
    color: var(--accent-rose);
    font-size: 11px;
    padding: 4px 10px;
    border-radius: 4px;
    cursor: pointer;
  }

  .remove-item-btn:hover {
    background: rgba(244, 63, 94, 0.1);
  }

  .product-card-body {
    padding: 18px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .product-fields-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 12px;
  }

  .trims-packaging-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 12px;
    background: var(--bg-input);
    padding: 14px;
    border-radius: var(--radius-md);
    border: 1px solid var(--border-subtle);
  }

  /* Commercial Footer Bar */
  .commercial-footer-bar {
    position: sticky;
    bottom: 20px;
    background: rgba(14, 20, 34, 0.96);
    backdrop-filter: blur(12px);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-lg);
    padding: 16px 24px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.6);
    z-index: 90;
    flex-wrap: wrap;
    gap: 16px;
  }

  .summary-pills {
    display: flex;
    align-items: center;
    gap: 14px;
    flex-wrap: wrap;
  }

  .summary-card {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .summary-card.highlight {
    background: rgba(16, 185, 129, 0.1);
    border: 1px solid rgba(16, 185, 129, 0.3);
    padding: 6px 14px;
    border-radius: var(--radius-md);
  }

  .sc-label {
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    color: var(--text-subtle);
  }

  .sc-val {
    font-size: 15px;
    font-weight: 800;
    color: var(--text-main);
  }

  .highlight .sc-val {
    color: var(--accent-emerald);
  }

  .submit-action-wrap {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .lock-tip {
    font-size: 11.5px;
    color: var(--accent-rose);
    font-weight: 600;
  }

  .submit-final-btn {
    height: 44px;
    padding: 0 24px;
    font-size: 13.5px;
    font-weight: 800;
    letter-spacing: 0.02em;
  }

  .submit-final-btn.locked {
    background: #475569 !important;
    border-color: #334155 !important;
    cursor: not-allowed;
    opacity: 0.6;
  }

  .spinner-sm {
    display: inline-block;
    width: 14px;
    height: 14px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: #fff;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
    margin-right: 6px;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>
