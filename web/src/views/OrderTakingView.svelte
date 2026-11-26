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
  let items = $state(/** @type {any[]} */ ([createInitialProductItem()]));

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

  // Authoritative Master Genres from attached specifications
  const apparelGenres = [
    'Polo', 'Bottoms', 'Jackets', 'Hoodies and Pullovers', 'Skorts', 'Dresses', 'Onsies', 'Caps',
    'Traded apparel', 'Quarter zip', 'Shirt', 'Shorts', 'Sweat Pants', 'Sweat shirt', 'Sweater',
    'Sweatshirt', 'Tee', 'Tights', 'Vest'
  ];

  const nonApparelGenres = [
    'Drinkwear', 'Headgear', 'Bags', 'Stationary', 'Electronics', 'Others'
  ];

  let resolvedApparelGenres = $derived((masters.apparel_genres && masters.apparel_genres.length > 0)
    ? masters.apparel_genres
    : apparelGenres);

  let resolvedNonApparelGenres = $derived((masters.non_apparel_genres && masters.non_apparel_genres.length > 0)
    ? masters.non_apparel_genres
    : nonApparelGenres);

  const brandLabelList = [
    'ETTU', 'Jeep', 'OMG', 'Illa', 'Waterdrop', 'American Athlete', 'TUC', 'Chocotopia', 'Citreon'
  ];

  let resolvedBrandLabels = $derived.by(() => {
    if (masters.label_brands && masters.label_brands.length > 0) {
      const dbBrands = masters.label_brands.map(b => b.display_name || b.code || b);
      return Array.from(new Set([...dbBrands, ...brandLabelList]));
    }
    return brandLabelList;
  });

  const additionalLabelTypes = [
    'Size label', 'Main+Size Label', 'washcare Label', 'COO Label', 'side Label',
    'Tag Label', 'Heatseal', 'Paper_sticker', 'DTF Branding', 'Flag Label', 'Patch Label'
  ];

  const labelGenres = ['Woven', 'Printed', 'DTF', 'Tareaway'];

  const custPackagingOptions = [
    'Poly-Carton', 'Poly+name', 'Ettu box with sticker', 'Ettu box with names', 'bulk packing', 'Golf Packing'
  ];

  const nonApparelSizesMap = {
    'Drinkwear': ['100ml', '150ml', '200ml', '250ml', '300ml', '500ml', '750ml', '1Litre', '1.5 Litre', '2 Litre'],
    'Headgear': ['Std', 'Kids', 'Large', 'Small'],
    'Caps': ['Std', 'Kids', 'Large', 'Small'],
    'Stationary': ['A4', 'A3', 'A5', 'A6', 'B5'],
    'Stationery': ['A4', 'A3', 'A5', 'A6', 'B5'],
    'Bags': ['15 Inch', '15.6 Inch', '10.5 Inch', 'Std'],
    'Electronics': ['Std'],
    'Others': ['Std']
  };

  const defaultLabelDescriptions = [
    'CSCrew +size Navy/grey', 'CS Crew Navy/Grey', 'ETTU +Size Navy/aqua', 'ETTU +Size Grey/Navy',
    'made in india white/black', 'CAT white/back', 'OMG navy/orange', 'size black/white',
    'jeep patch black/white', 'Jeep Sleeve label black/white', 'jeep side label black/white',
    'jeep 100% cotton', 'jeep 50%poly 50% cotton', 'jeep 95%poly 5% spandex', 'jeep 100% polyester',
    'jeep 50% poly 45%cotton 5% spendex', 'jeep 88%poly 12% spendex', 'jeep 92% poly 8% spandex',
    'size white/black', 'Citreon black/white', 'Citreon side label black/white', 'Citreon Cap Label Black/white',
    'nvidia tag label black/grn', 'waterdrop size white/black', 'size label black/red', 'american athlete round label',
    'amercian athlete tag', 'size black/beige', 'TUC Blazer', 'TUC Main', 'Cocotopia main', 'Chocotopia side label',
    'size beige + Brown', 'size white/navy', 'Illa main label', 'Ettu Jr +Size white',
    'ETTU onsie+composition+size white', 'ETTU onsie+composition+size cool grey', 'Nvidia +composition+size cool grey',
    'Nvidia +composition+size white', 'nividia onsie+composition+size white', 'nividia onsie+composition+size cool grey',
    'Ettu +size cool grey', 'Ettu +size white', 'Jeep Gender+size cool grey', 'Jeep Gender+size white',
    'Cat composition +size white', 'Cat composition +size black', 'astryea white', 'cloudbolt white', 'cloud bolt black', 'sample'
  ];

  let labelDescriptionsList = $derived.by(() => {
    if (masters.label_descriptions && masters.label_descriptions.length > 0) {
      const dbNames = masters.label_descriptions.map(l => l.display_name || l.code || l);
      return Array.from(new Set([...dbNames, ...defaultLabelDescriptions]));
    }
    return defaultLabelDescriptions;
  });

  function getNonApparelSizes(genre) {
    const g = (genre || '').trim();
    for (const [key, list] of Object.entries(nonApparelSizesMap)) {
      if (g.toLowerCase().includes(key.toLowerCase())) return list;
    }
    return ['Std', 'Small', 'Medium', 'Large'];
  }

  function getFilteredProductsFor(item) {
    const list = masters.products || [];
    if (item.subCategory) {
      const sc = item.subCategory.toLowerCase().trim();
      const matched = list.filter(p => {
        const pSc = (p.meta?.subcategory || p.meta?.subCategory || '').toLowerCase().trim();
        return pSc === sc;
      });
      if (matched.length > 0) return matched;
    }
    const isNA = item.category === 'Non Apparel' || item.category === 'Non-Apparel';
    const targetCat = isNA ? 'non apparel' : 'apparel';
    return list.filter(p => {
      const pCat = (p.meta?.category || '').toLowerCase().replace('-', ' ').trim();
      return pCat === targetCat;
    });
  }

  let brandingTypes = $derived((masters.branding_types && masters.branding_types.length > 0)
    ? masters.branding_types.map(b => ({ value: b.display_name, label: b.display_name }))
    : [
        { value: 'Print', label: 'Screen Print' },
        { value: 'DTF', label: 'Direct to Film (DTF)' },
        { value: 'Plastisol Print', label: 'Plastisol Print' },
        { value: 'HD print', label: 'High Density (HD) Print' },
        { value: 'Waterbased Print', label: 'Waterbased Print' },
        { value: 'Embroidery', label: 'Embroidery' },
        { value: 'Unik', label: 'Unik' }
      ]);

  function createInitialProductItem() {
    return {
      id: 'it_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6),
      category: 'Apparel',
      subCategory: 'Polo',
      productName: '3 Way polo',
      colorName: 'Classic Navy',
      fabricName: 'Venus',
      fabricComposition: '100%Poly',
      size: '',
      unitPrice: 450.0,
      totalQty: 100,
      clientEtd: new Date(Date.now() + 14 * 86400000).toISOString().split('T')[0],
      clientRemarks: '',
      productImageUrl: '',
      brandLabel: 'ETTU',
      brandLabelGenre: 'Woven',
      brandLabelDesc: 'ETTU +Size Navy/aqua',
      additionalLabels: [
        { labelType: 'washcare Label', description: 'made in india white/black', genre: 'Woven' }
      ],
      mainLabel: 'ETTU (ETTU +Size Navy/aqua)',
      otherLabel: 'washcare Label: made in india white/black',
      custPackaging: 'Poly-Carton',
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
    if (items.length > 0) {
      const prev = items[items.length - 1];
      const clonedBrandings = (prev.brandings || []).map(b => ({
        ...b,
        id: 'bnd_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6),
        brandingFileUrl: b.brandingFileUrl || ''
      }));
      const clonedAdditionalLabels = (prev.additionalLabels || []).map(al => ({
        ...al
      }));

      const isNonApp = prev.category === 'Non Apparel' || prev.category === 'Non-Apparel';

      items = [...items, {
        id: 'it_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6),
        category: prev.category || 'Apparel',
        subCategory: prev.subCategory || (isNonApp ? 'Drinkwear' : 'Polo'),
        productName: prev.productName || '',
        colorName: prev.colorName || '',
        fabricName: prev.fabricName || '',
        fabricComposition: prev.fabricComposition || '',
        size: prev.size || (isNonApp ? '500ml' : ''),
        unitPrice: prev.unitPrice || 0,
        totalQty: 0, // NEVER COPY QUANTITIES: strictly 0
        clientEtd: prev.clientEtd || new Date(Date.now() + 14 * 86400000).toISOString().split('T')[0],
        clientRemarks: prev.clientRemarks || '',
        productImageUrl: prev.productImageUrl || '',
        brandLabel: prev.brandLabel || 'ETTU',
        brandLabelGenre: prev.brandLabelGenre || 'Woven',
        brandLabelDesc: prev.brandLabelDesc || '',
        additionalLabels: clonedAdditionalLabels,
        mainLabel: prev.mainLabel || '',
        otherLabel: prev.otherLabel || '',
        custPackaging: prev.custPackaging || 'Poly-Carton',
        matrix: isNonApp ? { General: {} } : { Men: {}, Women: {} },
        brandings: clonedBrandings,
        personalizations: [],
        isLocked: true, // locked until quantity > 0
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

    const clonedAdditionalLabels = (src.additionalLabels || []).map(al => ({ ...al }));

    items[destIdx] = {
      ...dest,
      category: src.category,
      subCategory: src.subCategory,
      productName: src.productName,
      colorName: src.colorName,
      fabricName: src.fabricName,
      fabricComposition: src.fabricComposition,
      size: src.size || dest.size,
      unitPrice: src.unitPrice,
      brandLabel: src.brandLabel,
      brandLabelGenre: src.brandLabelGenre || 'Woven',
      brandLabelDesc: src.brandLabelDesc,
      additionalLabels: clonedAdditionalLabels,
      mainLabel: src.mainLabel,
      otherLabel: src.otherLabel,
      custPackaging: src.custPackaging,
      clientRemarks: src.clientRemarks,
      brandings: clonedBrandings,
      // Preserve existing destination quantities
      totalQty: dest.totalQty || 0
    };

    if (window.showToast) {
      window.showToast(`Copied specs from Item #${srcIdx + 1} to Item #${destIdx + 1} (Quantities preserved)`, 'success');
    }
  }

  function onProductNameSelect(idx, val) {
    const name = typeof val === 'object' ? (val.display_name || val.code || '') : String(val);
    items[idx].productName = name;
    if (!name) return;

    // Connect and auto-fill related dependent properties across the Order Taking page
    const prod = (masters.products || []).find(p =>
      p.display_name?.toLowerCase() === name.toLowerCase() || p.code?.toLowerCase() === name.toLowerCase()
    );

    if (prod && prod.meta) {
      const pCat = (prod.meta.category || '').replace('-', ' ').trim();
      if (pCat) {
        items[idx].category = pCat.toLowerCase() === 'non apparel' ? 'Non Apparel' : 'Apparel';
      }
      const pSub = prod.meta.subcategory || prod.meta.subCategory;
      if (pSub) {
        items[idx].subCategory = pSub;
      }
      if (prod.meta.basePrice && !items[idx].unitPrice) {
        items[idx].unitPrice = prod.meta.basePrice;
      }
    }

    // Auto-detect size for Non-Apparel
    if (items[idx].category === 'Non Apparel') {
      const lower = name.toLowerCase();
      if (lower.includes('1000 ml') || lower.includes('1000ml')) items[idx].size = '1000 ml';
      else if (lower.includes('1200 ml') || lower.includes('1200ml')) items[idx].size = '1.5 Litre';
      else if (lower.includes('750 ml') || lower.includes('750ml')) items[idx].size = '750ml';
      else if (lower.includes('650 ml') || lower.includes('650ml')) items[idx].size = '650ml';
      else if (lower.includes('500 ml') || lower.includes('500ml')) items[idx].size = '500ml';
      else if (lower.includes('450 ml') || lower.includes('450ml')) items[idx].size = '500ml';
      else if (lower.includes('350 ml') || lower.includes('350ml')) items[idx].size = '300ml';
      else if (lower.includes('15.6')) items[idx].size = '15.6 Inch';
      else if (lower.includes('15 inch') || lower.includes('15-inch')) items[idx].size = '15 Inch';
      else if (lower.includes('10.5')) items[idx].size = '10.5 Inch';
      else if (!items[idx].size) items[idx].size = getNonApparelSizes(items[idx].subCategory)[0] || 'Std';

      items[idx].isLocked = (Number(items[idx].totalQty) || 0) <= 0;
    } else {
      // Auto-suggest fabric for Apparel based on product name / genre
      if (!items[idx].fabricName) {
        const lower = name.toLowerCase();
        if (lower.includes('polo')) items[idx].fabricName = 'Pique';
        else if (lower.includes('hoodie') || lower.includes('jacket')) items[idx].fabricName = 'Classic Fleece';
        else if (lower.includes('tee') || lower.includes('t-shirt')) items[idx].fabricName = 'Sinker';
        else if (lower.includes('pant') || lower.includes('jogger')) items[idx].fabricName = 'Looperknit';
      }
    }
  }

  function onCategoryChange(idx, newCat) {
    items[idx].category = newCat;
    if (newCat === 'Non Apparel') {
      items[idx].subCategory = 'Drinkwear';
      items[idx].size = '500ml';
      items[idx].fabricName = '';
      items[idx].isLocked = (Number(items[idx].totalQty) || 0) <= 0;
    } else {
      items[idx].subCategory = 'Polo';
      items[idx].size = '';
      items[idx].fabricName = items[idx].fabricName || 'Pique';
      items[idx].isLocked = true;
    }
  }

  function onSubCategoryChange(idx, newSubCategory) {
    items[idx].subCategory = newSubCategory;
    if (items[idx].category === 'Non Apparel') {
      const sizes = getNonApparelSizes(newSubCategory);
      if (!sizes.includes(items[idx].size)) {
        items[idx].size = sizes[0] || 'Std';
      }
    }

    // Check if any previous item matches this subCategory to auto-copy common properties
    const matchingPrev = items.find((it, i) => i !== idx && it.subCategory === newSubCategory);
    if (matchingPrev) {
      if (!items[idx].fabricName && matchingPrev.fabricName) items[idx].fabricName = matchingPrev.fabricName;
      if (!items[idx].unitPrice && matchingPrev.unitPrice) items[idx].unitPrice = matchingPrev.unitPrice;
      if (!items[idx].brandLabel && matchingPrev.brandLabel) items[idx].brandLabel = matchingPrev.brandLabel;
      if (!items[idx].brandLabelDesc && matchingPrev.brandLabelDesc) items[idx].brandLabelDesc = matchingPrev.brandLabelDesc;
      if (!items[idx].custPackaging && matchingPrev.custPackaging) items[idx].custPackaging = matchingPrev.custPackaging;
      if ((!items[idx].additionalLabels || items[idx].additionalLabels.length === 0) && matchingPrev.additionalLabels?.length > 0) {
        items[idx].additionalLabels = matchingPrev.additionalLabels.map(al => ({ ...al }));
      }
      if (items[idx].brandings.length === 0 && matchingPrev.brandings.length > 0) {
        items[idx].brandings = matchingPrev.brandings.map(b => ({
          ...b,
          id: 'bnd_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6)
        }));
      }
    }
  }

  function toggleAdditionalLabel(idx, labelType) {
    if (!items[idx].additionalLabels) items[idx].additionalLabels = [];
    const existingIdx = items[idx].additionalLabels.findIndex(al => al.labelType === labelType);
    if (existingIdx >= 0) {
      items[idx].additionalLabels = items[idx].additionalLabels.filter((_, i) => i !== existingIdx);
    } else {
      items[idx].additionalLabels = [
        ...items[idx].additionalLabels,
        { labelType, description: '', genre: 'Woven' }
      ];
    }
  }

  function removeAdditionalLabel(idx, alIdx) {
    items[idx].additionalLabels = items[idx].additionalLabels.filter((_, i) => i !== alIdx);
  }

  function removeProductItem(idx) {
    if (items.length <= 1) {
      if (window.showToast) window.showToast('Order must contain at least one product item', 'warning');
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
    if (client.name) clientName = client.name;
    if (client.contact_person) contactPerson = client.contact_person;
    if (client.phone) phone = client.phone;
    if (client.email) email = client.email;
    if (client.gstin) gstin = client.gstin;
    if (client.billing_address) {
      billingAddress = client.billing_address;
    } else if (client.address) {
      billingAddress = client.address;
    }
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
    if (!ord) return;
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

    if (Array.isArray(ord.items) && ord.items.length > 0) {
      items = ord.items.map(it => {
        const isNonApp = it.category === 'Non Apparel' || it.category === 'Non-Apparel';
        const clonedBrandings = Array.isArray(it.brandings)
          ? it.brandings.map(b => ({
              ...b,
              id: 'bnd_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6),
              brandingFileUrl: b.brandingFileUrl || ''
            }))
          : [];
        const clonedPersonalizations = Array.isArray(it.personalizations)
          ? it.personalizations.map(p => ({
              ...p,
              id: 'prs_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6),
              qty: 0
            }))
          : [];
        const clonedAddLabels = Array.isArray(it.additionalLabels)
          ? it.additionalLabels.map(al => ({ ...al }))
          : [];

        return {
          ...it,
          id: 'it_cloned_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6),
          category: isNonApp ? 'Non Apparel' : 'Apparel',
          subCategory: it.subCategory || (isNonApp ? 'Drinkwear' : 'Polo'),
          fabricName: it.fabricName || '',
          fabricComposition: it.fabricComposition || '',
          brandLabel: it.brandLabel || it.mainLabel || '',
          brandLabelGenre: it.brandLabelGenre || 'Woven',
          brandLabelDesc: it.brandLabelDesc || '',
          additionalLabels: clonedAddLabels,
          // STRICT REQUIREMENT: DO NOT COPY QUANTITIES
          totalQty: 0,
          size: isNonApp ? (it.size || '') : '',
          matrix: isNonApp ? { General: {} } : { Men: {}, Women: {} },
          brandings: clonedBrandings,
          personalizations: clonedPersonalizations,
          isLocked: false,
          isCollapsed: false
        };
      });
    }
  }

  // Overall Aggregates & Math Lock Verification
  let grandTotalUnits = $derived.by(() => {
    return items.reduce((sum, item) => sum + (Number(item.totalQty) || 0), 0);
  });

  let grandTaxableSubtotal = $derived.by(() => {
    return items.reduce((sum, item) => {
      const q = Number(item.totalQty) || 0;
      const p = Number(item.unitPrice) || 0.0;
      return sum + (q * p);
    }, 0);
  });

  let grandOrderValuation = $derived.by(() => {
    const sub = grandTaxableSubtotal;
    const rate = Number(taxRate) || 0.0;
    if (taxBilling === 'Extra') {
      return sub + (sub * (rate / 100.0));
    }
    return sub;
  });

  // Authoritative Hard Lock Check: Any unbalanced item locks the entire order!
  // Note: Non-Apparel items are excluded from Matrix Math Checker
  let lockedItemInfo = $derived.by(() => {
    for (let i = 0; i < items.length; i++) {
      const it = items[i];
      const qty = Number(it.totalQty) || 0;
      if (qty <= 0) {
        return { isLocked: true, index: i + 1, productName: it.productName || 'Item #' + (i + 1), reason: 'Quantity must be greater than 0' };
      }
      if (!it.productName || !it.productName.trim()) {
        return { isLocked: true, index: i + 1, productName: 'Item #' + (i + 1), reason: 'Product Name is required' };
      }
      const isNonApp = it.category === 'Non Apparel' || it.category === 'Non-Apparel';
      if (!isNonApp && it.isLocked) {
        return { isLocked: true, index: i + 1, productName: it.productName, reason: 'Gender-Size Matrix sum does not match target total quantity' };
      }
    }
    return { isLocked: false, index: -1, productName: '', reason: '' };
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
        window.showToast(`LOCK VIOLATION: Item #${lockedItemInfo.index} (${lockedItemInfo.productName}) - ${lockedItemInfo.reason}.`, 'error');
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
        taxRate: Number(taxRate) || 0.0,
        taxBilling,
        freightMode,
        freightCost,
        clientComment: overallClientComment,
        targetEtd: items[0]?.clientEtd || '',
        items: items.map((it, idx) => {
          const isNonApp = it.category === 'Non Apparel' || it.category === 'Non-Apparel';
          const mainLbl = it.brandLabel
            ? (it.brandLabelDesc ? `${it.brandLabel} (${it.brandLabelDesc})` : it.brandLabel)
            : (it.mainLabel || '');
          const otherLbl = (it.additionalLabels && it.additionalLabels.length > 0)
            ? it.additionalLabels.map(al => al.description ? `${al.labelType}: ${al.description}` : al.labelType).join('; ')
            : (it.otherLabel || '');

          const computedMatrix = isNonApp
            ? { General: { [it.size || 'Std']: Number(it.totalQty) || 0 } }
            : it.matrix;

          return {
            id: it.id,
            itemSequence: idx + 1,
            category: isNonApp ? 'Non Apparel' : 'Apparel',
            subCategory: it.subCategory,
            productName: it.productName,
            colorName: it.colorName,
            fabricName: isNonApp ? '' : it.fabricName,
            fabricComposition: isNonApp ? '' : (it.fabricComposition || ''),
            size: it.size || '',
            unitPrice: Number(it.unitPrice) || 0.0,
            totalQty: Number(it.totalQty) || 0,
            clientEtd: it.clientEtd,
            clientRemarks: it.clientRemarks,
            productImageUrl: it.productImageUrl,
            brandLabel: it.brandLabel || '',
            brandLabelGenre: it.brandLabelGenre || 'Woven',
            brandLabelDesc: it.brandLabelDesc || '',
            additionalLabels: it.additionalLabels || [],
            mainLabel: mainLbl,
            otherLabel: otherLbl,
            custPackaging: it.custPackaging || 'Poly-Carton',
            matrix: computedMatrix,
            brandings: it.brandings || [],
            personalizations: it.personalizations || []
          };
        }),
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
            displayKey="name"
            valueKey="name"
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
        {#each items as item, idx (item.id)}
          <div class="product-card {item.isLocked ? 'item-locked' : 'item-valid'}">
            <!-- Product Card Header -->
            <div class="product-card-head">
              <div class="head-left">
                <span class="product-seq-badge">Product #{idx + 1}</span>
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
                    onclick={() => removeProductItem(idx)}
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
                        if (e.currentTarget.value !== '') {
                          copyPropertiesFromItem(idx, parseInt(e.currentTarget.value, 10));
                          e.currentTarget.value = '';
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
                      value={item.category}
                      onchange={(e) => onCategoryChange(idx, e.currentTarget.value)}
                    >
                      <option value="Apparel">Apparel</option>
                      <option value="Non Apparel">Non Apparel</option>
                    </select>
                  </div>

                  <!-- Sub-Category / Genre -->
                  <div class="form-group">
                    <label for="item-subcat-{item.id}">Product Genre / Sub-Category</label>
                    <Combobox
                      id="item-subcat-{item.id}"
                      items={item.category === 'Non Apparel' ? resolvedNonApparelGenres : resolvedApparelGenres}
                      bind:value={item.subCategory}
                      placeholder="Select genre (Polo, Jacket, Drinkwear...)"
                      onSelect={(val) => onSubCategoryChange(idx, typeof val === 'object' ? (val.display_name || val.value) : val)}
                    />
                  </div>

                  <!-- Product Name (Connected to Genre & Category) -->
                  <div class="form-group">
                    <label for="item-pname-{item.id}">Product Name</label>
                    <Combobox
                      id="item-pname-{item.id}"
                      items={getFilteredProductsFor(item)}
                      displayKey="display_name"
                      valueKey="display_name"
                      bind:value={item.productName}
                      placeholder="Search or enter product..."
                      onSelect={(val) => onProductNameSelect(idx, val)}
                    />
                  </div>

                  <!-- Color Combobox (477+ Master Colors from Excel) -->
                  <div class="form-group">
                    <label for="item-col-{item.id}">Color Name (477+ Master Colors)</label>
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

                  <!-- Fabric Name & Fabric Composition (Only for Apparel) -->
                  {#if item.category === 'Apparel'}
                    <div class="form-group">
                      <label for="item-fab-{item.id}">Fabric Name (FPCC Master Fabrics)</label>
                      <Combobox
                        id="item-fab-{item.id}"
                        items={masters.fabrics || []}
                        category="fabric"
                        displayKey="display_name"
                        valueKey="display_name"
                        bind:value={item.fabricName}
                        placeholder="Select FPCC fabric (Venus, Classic Fleece...)"
                      />
                    </div>

                    <div class="form-group">
                      <label for="item-fabcomp-{item.id}">Fabric Composition (Inward Composition Master)</label>
                      <Combobox
                        id="item-fabcomp-{item.id}"
                        items={masters.compositions || []}
                        category="composition"
                        displayKey="display_name"
                        valueKey="display_name"
                        bind:value={item.fabricComposition}
                        placeholder="Select fiber blend (95%Poly 5%SP, 100%Cot...)"
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
                      Total Quantity (Pcs) {#if item.category === 'Apparel'}<span class="math-req">*Authoritative Math Checked</span>{:else}<span class="math-req">*Line Quantity</span>{/if}
                    </label>
                    <input
                      id="item-target-qty-{item.id}"
                      type="number"
                      min="1"
                      class="input-control numeric font-mono font-bold"
                      bind:value={item.totalQty}
                      oninput={() => {
                        if (item.category === 'Non Apparel') {
                          item.isLocked = (Number(item.totalQty) || 0) <= 0;
                        }
                      }}
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

                <!-- APPAREL VS NON-APPAREL SIZING CONFIGURATION -->
                {#if item.category === 'Apparel'}
                  <!-- GENDER-SIZE-QTY MATRIX WITH REAL-TIME MATH CHECKER (APPAREL ONLY) -->
                  <div class="matrix-section-wrap">
                    <GenderSizeMatrix
                      category={item.category}
                      subCategory={item.subCategory}
                      bind:totalQty={item.totalQty}
                      bind:matrix={item.matrix}
                      unitPrice={item.unitPrice}
                      bind:isLocked={item.isLocked}
                    />
                  </div>
                {:else}
                  <!-- NON-APPAREL SINGLE SIZE & QUANTITY (NO GENDER, NO MATRIX MATH CHECKER) -->
                  <div class="non-apparel-size-wrap">
                    <div class="non-apparel-header">
                      <div class="na-title-group">
                        <span class="na-badge">📦 Non-Apparel Single Size Config</span>
                        <span class="na-tip">No gender configuration. Each different size/spec in this category is added as a separate product line item.</span>
                      </div>
                    </div>

                    <div class="non-apparel-controls-grid">
                      <div class="form-group">
                        <label for="item-na-size-{item.id}">Size / Capacity Specification <span class="req">*</span></label>
                        <Combobox
                          id="item-na-size-{item.id}"
                          items={getNonApparelSizes(item.subCategory)}
                          bind:value={item.size}
                          placeholder="Select or enter size (e.g. 500ml, Std, 15 Inch)..."
                        />
                      </div>

                      <div class="form-group">
                        <label for="item-na-qty-{item.id}">Item Quantity (Pcs) <span class="req">*</span></label>
                        <input
                          id="item-na-qty-{item.id}"
                          type="number"
                          min="1"
                          class="input-control numeric font-mono font-bold"
                          placeholder="Enter piece count..."
                          bind:value={item.totalQty}
                          oninput={() => {
                            item.isLocked = (Number(item.totalQty) || 0) <= 0;
                          }}
                        />
                      </div>
                    </div>
                  </div>
                {/if}

                <!-- BRANDING & EMBELLISHMENTS (MULTIPLE PER PRODUCT) -->
                <div class="brandings-section-wrap">
                  <EmbellishmentList bind:brandings={item.brandings} {brandingTypes} />
                </div>

                <!-- LABELS & PACKAGING -->
                <div class="labels-packaging-card">
                  <div class="labels-header">
                    <span class="labels-title">🏷️ Product Labels & Embellished Trims</span>
                    <span class="labels-sub">Brand Label with Description & Multi-Select Additional Labels</span>
                  </div>

                  <div class="brand-label-grid">
                    <div class="form-group">
                      <label for="item-brand-label-{item.id}">Brand Label <span class="req">*</span></label>
                      <Combobox
                        id="item-brand-label-{item.id}"
                        items={resolvedBrandLabels}
                        bind:value={item.brandLabel}
                        placeholder="Select Brand Label (ETTU, Jeep, OMG, Illa...)"
                      />
                    </div>

                    <div class="form-group">
                      <label for="item-brand-genre-{item.id}">Brand Label Genre</label>
                      <select
                        id="item-brand-genre-{item.id}"
                        class="select-control"
                        bind:value={item.brandLabelGenre}
                      >
                        {#each labelGenres as lg}
                          <option value={lg}>{lg}</option>
                        {/each}
                      </select>
                    </div>

                    <div class="form-group">
                      <label for="item-brand-desc-{item.id}">Brand Label Description / Tech Specs</label>
                      <Combobox
                        id="item-brand-desc-{item.id}"
                        items={labelDescriptionsList}
                        bind:value={item.brandLabelDesc}
                        placeholder="Select description or enter custom..."
                      />
                    </div>
                  </div>

                  <!-- Additional Labels Multi-Select Chips -->
                  <div class="additional-labels-wrap">
                    <span class="additional-labels-label">Additional Labels (Multi-Select):</span>
                    <div class="additional-labels-chips">
                      {#each additionalLabelTypes as aType}
                        {@const isSelected = (item.additionalLabels || []).some(al => al.labelType === aType)}
                        <button
                          type="button"
                          class="label-chip {isSelected ? 'selected' : ''}"
                          onclick={() => toggleAdditionalLabel(idx, aType)}
                        >
                          {isSelected ? '✓ ' : '+ '}{aType}
                        </button>
                      {/each}
                    </div>
                  </div>

                  <!-- Additional Labels Configured Rows with Descriptions -->
                  {#if item.additionalLabels && item.additionalLabels.length > 0}
                    <div class="additional-labels-table-wrap">
                      {#each item.additionalLabels as al, alIdx}
                        <div class="additional-label-entry">
                          <span class="al-type-tag">🏷️ {al.labelType}</span>

                          <div class="al-genre-select">
                            <select class="select-control compact" bind:value={al.genre}>
                              {#each labelGenres as lg}
                                <option value={lg}>{lg}</option>
                              {/each}
                            </select>
                          </div>

                          <div class="al-desc-input">
                            <Combobox
                              items={labelDescriptionsList}
                              bind:value={al.description}
                              placeholder={`Description for ${al.labelType}...`}
                            />
                          </div>

                          <button
                            type="button"
                            class="al-del-btn"
                            onclick={() => removeAdditionalLabel(idx, alIdx)}
                            title="Remove label"
                          >
                            ✕
                          </button>
                        </div>
                      {/each}
                    </div>
                  {/if}

                  <!-- Customer Packaging Specs -->
                  <div class="form-group" style="margin-top:16px;">
                    <label for="item-pkg-{item.id}">Customer Packaging Specs <span class="req">*</span></label>
                    <Combobox
                      id="item-pkg-{item.id}"
                      items={custPackagingOptions}
                      placeholder="Select customer packaging (Poly-Carton, Ettu box with sticker...)"
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

  /* Non-Apparel Single Size Styling */
  .non-apparel-size-wrap {
    background: var(--bg-tertiary);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .non-apparel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
  }

  .na-title-group {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }

  .na-badge {
    font-size: 11px;
    font-weight: 800;
    text-transform: uppercase;
    background: rgba(59, 130, 246, 0.15);
    color: #60a5fa;
    border: 1px solid rgba(59, 130, 246, 0.3);
    padding: 3px 8px;
    border-radius: 4px;
  }

  .na-tip {
    font-size: 11.5px;
    color: var(--text-muted);
  }

  .non-apparel-controls-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 14px;
  }

  @media (max-width: 640px) {
    .non-apparel-controls-grid {
      grid-template-columns: 1fr;
    }
  }

  /* Labels & Packaging Styling */
  .labels-packaging-card {
    background: var(--bg-input);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .labels-header {
    display: flex;
    flex-direction: column;
    gap: 2px;
    border-bottom: 1px solid var(--border-subtle);
    padding-bottom: 8px;
  }

  .labels-title {
    font-size: 13px;
    font-weight: 700;
    color: var(--text-main);
  }

  .labels-sub {
    font-size: 11px;
    color: var(--text-subtle);
  }

  .brand-label-grid {
    display: grid;
    grid-template-columns: 1fr 0.8fr 1.6fr;
    gap: 14px;
  }

  @media (max-width: 640px) {
    .brand-label-grid {
      grid-template-columns: 1fr;
    }
  }

  .additional-labels-wrap {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .additional-labels-label {
    font-size: 11.5px;
    font-weight: 700;
    color: var(--text-secondary);
  }

  .additional-labels-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .label-chip {
    font-size: 11px;
    font-weight: 600;
    padding: 4px 10px;
    border-radius: 9999px;
    border: 1px solid var(--border-subtle);
    background: var(--bg-surface);
    color: var(--text-muted);
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .label-chip:hover {
    border-color: var(--accent-primary);
    color: var(--text-main);
  }

  .label-chip.selected {
    background: var(--accent-primary);
    border-color: var(--accent-primary);
    color: #fff;
    font-weight: 700;
  }

  .additional-labels-table-wrap {
    display: flex;
    flex-direction: column;
    gap: 8px;
    background: rgba(0, 0, 0, 0.15);
    padding: 12px;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border-subtle);
  }

  .additional-label-entry {
    display: grid;
    grid-template-columns: 140px 110px 1fr 32px;
    gap: 10px;
    align-items: center;
  }

  @media (max-width: 768px) {
    .additional-label-entry {
      grid-template-columns: 1fr;
      gap: 6px;
    }
  }

  .al-type-tag {
    font-size: 11.5px;
    font-weight: 700;
    color: var(--text-main);
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    padding: 4px 8px;
    border-radius: 4px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .al-genre-select {
    width: 100%;
  }

  .al-desc-input {
    width: 100%;
  }

  .al-del-btn {
    background: transparent;
    border: 1px solid rgba(244, 63, 94, 0.3);
    color: var(--accent-rose);
    border-radius: 4px;
    height: 32px;
    width: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
  }

  .al-del-btn:hover {
    background: rgba(244, 63, 94, 0.15);
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
