<script>
  import { onMount } from 'svelte';
  import { api } from '../api.js';
  import Combobox from '../components/Combobox.svelte';
  import SizeMatrix from '../components/SizeMatrix.svelte';
  import MultiShipping from '../components/MultiShipping.svelte';
  import PersonalizationRoster from '../components/PersonalizationRoster.svelte';
  import CopySpecsModal from '../components/CopySpecsModal.svelte';

  let { masters = {} } = $props();

  // Order Model State (Svelte 5 Runes)
  let orderNo = $state('');
  let salesOrderNo = $state('');
  let orderType = $state('Bulk');
  let bookingDate = $state(new Date().toISOString().split('T')[0]);
  let targetEtd = $state(new Date(Date.now() + 14 * 86400000).toISOString().split('T')[0]);

  let clientName = $state('');
  let contactPerson = $state('');
  let phone = $state('');
  let email = $state('');
  let gstin = $state('');
  let billingAddress = $state('');

  let category = $state('Apparel');
  let gender = $state('Men');
  let subCategory = $state('Polo Shirt');
  let productName = $state('Classic Polo');
  let colorName = $state('Navy Blue');
  let fabricName = $state('100% Cotton Pique 220 GSM');

  let unitPrice = $state(450.0);
  let taxRate = $state(5.0);
  let taxBilling = $state('Extra');
  let freightMode = $state('Surface');
  let freightCost = $state('Extra');

  let sizeBreakdown = $state({});
  let totalQty = $state(0);
  let totalValue = $state(0);

  let isMultiShipping = $state(false);
  let shippingLocations = $state([]);
  let personalization = $state([]);

  let brandingType = $state('Chest Logo Embroidery');
  let mainLabel = $state('Woven Satin Main Label');
  let custPackaging = $state('Self Adhesive Polybag');
  let clientComment = $state('');

  let isCopySpecsOpen = $state(false);
  let isSubmitting = $state(false);

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

  function handleApplyClonedSpecs(ord) {
    clientName = ord.client_name || '';
    contactPerson = ord.contact_person || '';
    phone = ord.phone || '';
    email = ord.email || '';
    gstin = ord.gstin || '';
    billingAddress = ord.billing_address || '';

    productName = ord.product_name || '';
    colorName = ord.color_name || '';
    fabricName = ord.fabric_name || '';
    subCategory = ord.sub_category || '';
    unitPrice = ord.unit_price || 0;
    brandingType = ord.branding_type || '';
    mainLabel = ord.main_label || '';
    custPackaging = ord.cust_packaging || '';
    clientComment = ord.client_comment || '';
  }

  async function handleSubmit(e) {
    e.preventDefault();
    if (!orderNo || !clientName.trim() || !productName.trim() || !colorName.trim()) {
      if (window.showToast) window.showToast('Please fill all mandatory fields (Client, Product, Color)', 'error');
      return;
    }

    if (totalQty <= 0) {
      if (window.showToast) window.showToast('Total quantity must be greater than 0', 'error');
      return;
    }

    isSubmitting = true;
    const payload = {
      productionOrderNo: orderNo,
      salesOrderNo,
      orderType,
      bookingDate,
      clientId: clientName,
      clientName,
      contactPerson,
      phone,
      email,
      billingAddress,
      shippingAddress: billingAddress,
      gstin,
      category,
      subCategory,
      productName,
      colorName,
      fabricName,
      gender,
      unitPrice: parseFloat(unitPrice) || 0.0,
      taxRate: parseFloat(taxRate) || 0.0,
      taxBilling,
      freightMode,
      freightCost,
      totalQty,
      sizeBreakdown,
      multipleShipping: isMultiShipping,
      shippingLocations: isMultiShipping ? shippingLocations : [],
      personalization: personalization.filter(p => p.name || p.number),
      brandingType,
      mainLabel,
      custPackaging,
      clientComment,
      targetEtd
    };

    try {
      const res = await api.createOrder(payload);
      if (window.showToast) {
        window.showToast(`✓ Production Order ${orderNo} booked into Unified Graph!`, 'success');
      }

      // Reset Form state
      salesOrderNo = '';
      sizeBreakdown = {};
      personalization = [];
      shippingLocations = [];
      isMultiShipping = false;
      clientComment = '';
      await refreshOrderNo();
    } catch (err) {
      if (window.showToast) {
        window.showToast(`Order submission failed: ${err.message}`, 'error');
      }
    } finally {
      isSubmitting = false;
    }
  }
</script>

<div class="view-header" style="display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:24px;">
  <div>
    <h2 style="font-size:22px;font-weight:800;color:#fff;">Station 0: Production Order Taking</h2>
    <p style="font-size:13px;color:var(--text-muted);margin-top:4px;">
      34-Point Specification Control & Bi-Temporal System of Record
    </p>
  </div>
  <div style="display:flex;gap:10px;">
    <button type="button" class="btn btn-secondary btn-sm" onclick={() => isCopySpecsOpen = true}>
      📋 Copy Specs from Prior Order
    </button>
    <button type="button" class="btn btn-secondary btn-sm" onclick={refreshOrderNo}>
      🔄 Refresh No
    </button>
  </div>
</div>

<form onsubmit={handleSubmit}>
  <!-- Panel 1: Order Identification & Client -->
  <div class="card-panel">
    <div class="panel-header">
      <div class="panel-title">1. Identification & Client Entity</div>
      <div class="system-badge" style="font-size:13px;">{orderNo || 'FETCHING...'}</div>
    </div>

    <div style="display:grid;grid-template-columns:repeat(4, 1fr);gap:16px;">
      <div class="form-group">
        <label class="form-label">Production Order No <span class="required">*</span></label>
        <input type="text" class="input-control input-mono" value={orderNo} readonly />
      </div>
      <div class="form-group">
        <label class="form-label">Sales Order / PO Ref</label>
        <input type="text" class="input-control" placeholder="e.g. SO-2026-99" bind:value={salesOrderNo} />
      </div>
      <div class="form-group">
        <label class="form-label">Booking Date <span class="required">*</span></label>
        <input type="date" class="input-control" bind:value={bookingDate} />
      </div>
      <div class="form-group">
        <label class="form-label">Target ETD <span class="required">*</span></label>
        <input type="date" class="input-control" bind:value={targetEtd} />
      </div>
    </div>

    <div class="form-group" style="margin-top:8px;">
      <label class="form-label">Order Type (Strict Enum) <span class="required">*</span></label>
      <div class="badge-select-wrapper">
        {#each ['Bulk', 'Sample', 'FOC', 'Repeat', 'Development'] as type}
          <span
            class="badge-option"
            class:selected={orderType === type}
            onclick={() => orderType = type}
          >
            {type}
          </span>
        {/each}
      </div>
    </div>

    <hr style="border:none;border-top:1px solid var(--border-subtle);margin:16px 0;" />

    <div style="display:grid;grid-template-columns:repeat(3, 1fr);gap:16px;">
      <div class="form-group">
        <label class="form-label">Client Name <span class="required">*</span> <span class="helper">Auto-registers</span></label>
        <Combobox
          category="client"
          items={masters.clients || []}
          displayKey="name"
          valueKey="name"
          placeholder="Search 14+ clients or add new..."
          bind:value={clientName}
          onSelect={handleClientSelect}
        />
      </div>
      <div class="form-group">
        <label class="form-label">Contact Person</label>
        <input type="text" class="input-control" placeholder="Procurement Manager" bind:value={contactPerson} />
      </div>
      <div class="form-group">
        <label class="form-label">Phone / WhatsApp</label>
        <input type="text" class="input-control" placeholder="+91 98xxx" bind:value={phone} />
      </div>
    </div>

    <div style="display:grid;grid-template-columns:repeat(3, 1fr);gap:16px;">
      <div class="form-group">
        <label class="form-label">Client Email</label>
        <input type="email" class="input-control" placeholder="client@company.com" bind:value={email} />
      </div>
      <div class="form-group">
        <label class="form-label">Client GSTIN</label>
        <input type="text" class="input-control input-mono" placeholder="27XXXXX1234F1Z1" bind:value={gstin} />
      </div>
      <div class="form-group">
        <label class="form-label">Billing / Shipping Address</label>
        <input type="text" class="input-control" placeholder="Headquarters Address" bind:value={billingAddress} />
      </div>
    </div>
  </div>

  <!-- Panel 2: Garment Specification -->
  <div class="card-panel">
    <div class="panel-header">
      <div class="panel-title">2. Garment Specification & Master Attributes</div>
    </div>

    <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px;margin-bottom:14px;">
      <div class="form-group">
        <label class="form-label">Category <span class="required">*</span></label>
        <div class="badge-select-wrapper">
          {#each ['Apparel', 'Non-Apparel'] as cat}
            <span
              class="badge-option"
              class:selected={category === cat}
              onclick={() => category = cat}
            >
              {cat}
            </span>
          {/each}
        </div>
      </div>
      <div class="form-group">
        <label class="form-label">Target Gender <span class="required">*</span></label>
        <div class="badge-select-wrapper">
          {#each ['Men', 'Women', 'Unisex', 'Boys', 'Girls'] as gen}
            <span
              class="badge-option"
              class:selected={gender === gen}
              onclick={() => gender = gen}
            >
              {gen}
            </span>
          {/each}
        </div>
      </div>
    </div>

    <div style="display:grid;grid-template-columns:repeat(4, 1fr);gap:16px;">
      <div class="form-group">
        <label class="form-label">Sub-Category <span class="required">*</span></label>
        <Combobox
          category="subcategory"
          items={subCategories}
          bind:value={subCategory}
          placeholder="Select sub-category..."
        />
      </div>
      <div class="form-group">
        <label class="form-label">Product Name <span class="required">*</span></label>
        <Combobox
          category="product"
          items={masters.products || []}
          bind:value={productName}
          placeholder="Search products..."
        />
      </div>
      <div class="form-group">
        <label class="form-label">Color Name <span class="required">*</span></label>
        <Combobox
          category="color"
          items={masters.colors || []}
          bind:value={colorName}
          placeholder="Search 289+ colors..."
        />
      </div>
      <div class="form-group">
        <label class="form-label">Fabric Name <span class="required">*</span></label>
        <Combobox
          category="fabric"
          items={masters.fabrics || []}
          bind:value={fabricName}
          placeholder="Search 176+ fabrics..."
        />
      </div>
    </div>
  </div>

  <!-- Panel 3: Commercials & Pricing -->
  <div class="card-panel">
    <div class="panel-header">
      <div class="panel-title">3. Commercial Terms & Taxes</div>
    </div>

    <div style="display:grid;grid-template-columns:repeat(4, 1fr);gap:16px;">
      <div class="form-group">
        <label class="form-label">Unit Price (₹) <span class="required">*</span></label>
        <input
          type="number"
          class="input-control numeric"
          step="0.5"
          min="0"
          placeholder="0.00"
          bind:value={unitPrice}
        />
      </div>
      <div class="form-group">
        <label class="form-label">Tax Rate (Strict Enum) <span class="required">*</span></label>
        <div class="badge-select-wrapper">
          {#each [5, 12, 18] as rate}
            <span
              class="badge-option"
              class:selected={taxRate === rate}
              onclick={() => taxRate = rate}
            >
              {rate}%
            </span>
          {/each}
        </div>
      </div>
      <div class="form-group">
        <label class="form-label">Tax Billing Mode</label>
        <div class="badge-select-wrapper">
          {#each ['Extra', 'Included'] as mode}
            <span
              class="badge-option"
              class:selected={taxBilling === mode}
              onclick={() => taxBilling = mode}
            >
              {mode}
            </span>
          {/each}
        </div>
      </div>
      <div class="form-group">
        <label class="form-label">Freight Logistics Mode</label>
        <div class="badge-select-wrapper">
          {#each ['Surface', 'Air', 'Courier'] as fMode}
            <span
              class="badge-option"
              class:selected={freightMode === fMode}
              onclick={() => freightMode = fMode}
            >
              {fMode}
            </span>
          {/each}
        </div>
      </div>
    </div>
  </div>

  <!-- Panel 4: Size Breakdown Matrix -->
  <div class="card-panel">
    <div class="panel-header">
      <div class="panel-title">4. Size Breakdown Matrix & Real-Time Math</div>
    </div>
    <SizeMatrix
      bind:sizeBreakdown={sizeBreakdown}
      bind:totalQty={totalQty}
      bind:totalValue={totalValue}
      {unitPrice}
      {taxRate}
      {taxBilling}
    />
  </div>

  <!-- Panel 5: Multi-Shipping Logistics -->
  <div class="card-panel">
    <div class="panel-header">
      <div class="panel-title">5. Multi-Shipping Logistics Allocation</div>
    </div>
    <MultiShipping
      bind:isMultiShipping={isMultiShipping}
      bind:shippingLocations={shippingLocations}
      orderTotalQty={totalQty}
    />
  </div>

  <!-- Panel 6: Player / Staff Personalization Roster -->
  <div class="card-panel">
    <div class="panel-header">
      <div class="panel-title">6. Personalization Roster (Jersey / Corporate Uniforms)</div>
    </div>
    <PersonalizationRoster
      bind:personalization={personalization}
    />
  </div>

  <!-- Panel 7: Trims, Branding & Packaging -->
  <div class="card-panel">
    <div class="panel-header">
      <div class="panel-title">7. Branding, Labels & Packaging Configurations</div>
    </div>

    <div style="display:grid;grid-template-columns:repeat(3, 1fr);gap:16px;">
      <div class="form-group">
        <label class="form-label">Branding Type / Placement</label>
        <input
          type="text"
          class="input-control"
          placeholder="e.g. Chest Embroidery + Back DTF"
          bind:value={brandingType}
        />
      </div>
      <div class="form-group">
        <label class="form-label">Main Label Spec</label>
        <Combobox
          category="label"
          items={masters.labels || []}
          bind:value={mainLabel}
          placeholder="Select label..."
        />
      </div>
      <div class="form-group">
        <label class="form-label">Packaging Requirement</label>
        <Combobox
          category="packaging"
          items={masters.packaging || []}
          bind:value={custPackaging}
          placeholder="Select packaging..."
        />
      </div>
    </div>

    <div class="form-group" style="margin-top:8px;">
      <label class="form-label">Special Factory Notes / Client Comments</label>
      <textarea
        class="textarea-control"
        rows="2"
        placeholder="e.g. Bio-wash finish mandatory. Double stitch armhole seams. Strict needle check."
        bind:value={clientComment}
      ></textarea>
    </div>
  </div>

  <!-- Submit Bar -->
  <div style="display:flex;justify-content:flex-end;gap:16px;margin-bottom:48px;">
    <button
      type="submit"
      class="btn btn-primary"
      disabled={isSubmitting}
      style="padding:14px 32px;font-size:15px;font-weight:800;"
    >
      {isSubmitting ? '⏳ RECORDING INTO UNIFIED GRAPH...' : '🚀 BOOK PRODUCTION ORDER INTO GRAPH'}
    </button>
  </div>
</form>

<CopySpecsModal
  bind:isOpen={isCopySpecsOpen}
  onCopySpecs={handleApplyClonedSpecs}
/>
