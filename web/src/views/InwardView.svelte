<script>
  import { onMount, untrack } from 'svelte';
  import { api } from '../api.js';
  import Combobox from '../components/Combobox.svelte';
  import QRTagModal from '../components/QRTagModal.svelte';
  import RemnantModal from '../components/RemnantModal.svelte';
  import DiscrepancyModal from '../components/DiscrepancyModal.svelte';
  import GraphInspectorModal from '../components/GraphInspectorModal.svelte';
  import {
    cleanDisplayNames,
    normInwardKind,
    normalizeNonApparelGenre,
    defaultUomForInward,
    getInwardTypeOptions,
    getZipTeethSizes,
    getPackagingVariants,
    getPackagingSizes,
    getTrimItems,
    getTrimSizes,
    getTrimColors,
    getGenreOptionsForInward,
    getProductOptionsForInward,
    getAllowedSizesForInward,
    getDefaultSizeSpecForInward,
    reconcileBundleSizeSpec
  } from '../lib/inwardRules.js';

  let { masters = {} } = $props();

  // Active Station: '1A' (Gate Entry - Desktop), '1B' (Floor Count - Tablet), '1C' (QC Inspection - Phone)
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

  // 1. Top-Level Invoice Details (Split Header)
  let invoiceSearchQuery = $state('');
  let existingInvoices = $state([]);
  let isSearchingInvoices = $state(false);
  let selectedInvoice = $state(null);
  let taggedLotsOnInvoice = $state([]);

  let gateInvoiceNo = $state('');
  let gateVendorName = $state('');
  let gateInwardDate = $state(new Date().toISOString().split('T')[0]);
  let gateInvoiceQty = $state('');
  let gateInvoiceAmount = $state('');
  let gateCurrency = $state('INR');
  let gateUom = $state('Kgs');

  // 2. Parent Inward Lot Details
  let gateParentCode = $state('');
  let gateInwardType = $state('Fabric');
  let gateNoOfBundles = $state(4);
  let gateComment = $state('');
  let qrCopiesPerBundle = $state(3); // Default 3 QR copies per bundle

  // Single Product Fields (used for Non-Fabric inwards)
  let gateVendorProduct = $state('');
  let gateFpccProduct = $state('');
  let gateVendorColor = $state('');
  let gateFpccColor = $state('');

  // Conditional specs for single product inwards
  let gateVendorZipName = $state('');
  let gateZipType = $state('Metal');
  let gateZipTeethSize = $state('5.0');
  let gateTapeColor = $state('');
  let gateTeethColor = $state('Antique silver');
  let gateSlider = $state('DA8DR4');
  let gateClosure = $state('open ended');

  let gatePackagingType = $state('Polybag');
  let gatePackagingVariant = $state('Ettu');
  let gatePackagingSize = $state('12x16"');
  let gateVendorPackagingName = $state('');

  let gateTrimType = $state('Button');
  let gateTrimSize = $state('18L');
  let gateTrimRate = $state('');
  let gateVendorTrimName = $state('');
  let gateTrimCordEnd = $state('');
  let gateTrimFusingGsm = $state('');
  let gateTrimThreadLength = $state('');

  let gateFgType = $state('Polo');
  let gateGender = $state('Men');
  let gateVendorFgName = $state('');
  let lastActiveInwardType = $state('Fabric');

  let gateLabelCategory = $state('Brand Label');
  let gateLabelType = $state('ETTU');
  let gateLabelGenre = $state('Woven');
  let gateLabelDescription = $state('');
  let gateLabelBaseColor = $state('');
  let gateLabelTextColor = $state('');

  let customBundleDistribution = $state(false);

  // 3. Multi-Product Fabric Consignment Lines (when inwardType === 'Fabric')
  let fabricProducts = $state([
    createDefaultFabricProduct(1, 'Body Fabric')
  ]);

  function createDefaultFabricProduct(seq = 1, role = 'Body Fabric') {
    return {
      id: 'fp_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6),
      role: role,
      fabricType: seq === 1 ? 'Single Jersey' : (seq === 2 ? 'Rib 1x1 / 2x2' : 'Single Jersey'),
      fpccProduct: seq === 1 ? 'Single Jersey 180 GSM' : (seq === 2 ? '1x1 Collar Rib 220 GSM' : 'Cuff Rib'),
      vendorProduct: seq === 1 ? 'Super Combed Cotton SJ' : (seq === 2 ? 'Collar Rib 1x1' : 'Cuff Rib'),
      color: 'Navy Blue',
      uom: 'Kgs',
      dia: seq === 1 ? '60"' : '34"',
      gsm: seq === 1 ? '180' : '220',
      composition: '100% Super Combed Cotton',
      rate: seq === 1 ? 350.0 : 380.0,
      rollCount: seq === 1 ? 3 : 1
    };
  }

  function addFabricProduct() {
    const nextSeq = fabricProducts.length + 1;
    const roleName = nextSeq === 2 ? 'Collar Rib' : (nextSeq === 3 ? 'Cuff Rib' : `Fabric Part #${nextSeq}`);
    fabricProducts = [...fabricProducts, createDefaultFabricProduct(nextSeq, roleName)];
    if (window.showToast) {
      window.showToast(`Added ${roleName} line to consignment lot`, 'info');
    }
  }

  function removeFabricProduct(idx) {
    if (fabricProducts.length <= 1) {
      if (window.showToast) window.showToast('Lot must have at least one fabric product', 'warning');
      return;
    }
    fabricProducts = fabricProducts.filter((_, i) => i !== idx);
  }

  // Derive total bundles for fabric consignment
  let fabricTotalRolls = $derived.by(() => {
    return fabricProducts.reduce((sum, fp) => sum + (Number(fp.rollCount) || 1), 0);
  });

  // Keep gateNoOfBundles in sync with fabric total rolls
  $effect(() => {
    if (normInwardKind(gateInwardType) === 'fabric') {
      gateNoOfBundles = fabricTotalRolls;
    }
  });

  // Bundle stubs for Station 1A
  let gateBundles = $state([]);

  // =========================================================
  // STATION 1B: FLOOR COUNT STATE (TABLET VIEW)
  // =========================================================
  let floorSelectedLotCode = $state('');
  let floorLotDetail = $state(null);
  let floorInvoicedQty = $state('');
  let floorCountedQty = $state('');
  let floorLocationBin = $state('A/R01/S01');
  let floorCountedBy = $state('Storekeeper Ramesh');
  let floorComments = $state('');
  let floorHoldOnShort = $state(true);
  const FLOOR_ATTR_KEYS = [
    'fabricType', 'composition', 'invoiceGsm', 'dia',
    'zipType', 'teethSize', 'slider', 'closure', 'teethColor', 'tapeColor',
    'trimType', 'trimSize', 'cordEnd', 'fusingGsm', 'threadLength',
    'packagingType', 'packagingVariant', 'packagingSize',
    'labelCategory', 'labelGenre', 'labelType', 'labelDescription', 'baseColor', 'textColor',
    'finishedGoodType', 'gender'
  ];

  function createEmptyFloorAttrs() {
    /** @type {Record<string, any>} */
    const out = {};
    for (const k of FLOOR_ATTR_KEYS) out[k] = '';
    return out;
  }

  let floorBundles = $state([]);
  let floorAttrs = $state(createEmptyFloorAttrs());

  // =========================================================
  // STATION 1C: QC INSPECTION STATE (PHONE VIEW)
  // =========================================================
  let qcSelectedLotCode = $state('');
  let qcLotDetail = $state(null);
  let qcInspector = $state('QC Lead Priya');
  let qcBundles = $state([]);
  let qcActiveBundleIndex = $state(0); // Single bundle focus card on mobile

  // Audio Recording & Speech-to-Text State
  let activeRecordingBundleIdx = $state(-1);
  let isRecordingAudio = $state(false);
  let mediaRecorderInstance = null;
  let audioChunks = [];
  let speechRecognitionInstance = null;
  let recordingTimerSeconds = $state(0);
  let recordingTimerInterval = null;

  // =========================================================
  // MASTER DATA DROPDOWNS & CATEGORY-LINKED OPTIONS
  // =========================================================
  let inwardTypeOptions = $derived(getInwardTypeOptions(masters));

  const defaultFabricTypes = [
    'Fabric', 'Collar', 'Cuff', 'Rib', 'Bonded', 'Canvass', 'Fleece', 'Grindle', 'Hemp',
    'Interlock', 'Jacquard', 'Knitted', 'Looperknit', 'NS Lycra', 'Pique', 'Single Jersey', 'Terry', 'Woven', 'WR-rain'
  ];
  let fabricTypeOptions = $derived((masters.fabric_types && masters.fabric_types.length > 0)
    ? cleanDisplayNames(masters.fabric_types)
    : defaultFabricTypes);

  const standardGsmOptions = ['140', '160', '180', '200', '220', '240', '260', '280', '300', '320', '350', '380', '400'];
  const standardDiaOptions = ['24"', '26"', '28"', '30"', '32"', '34"', '36"', '40"', '48"', '60"'];

  let zipTypeOptions = $derived((masters.zip_types && masters.zip_types.length > 0)
    ? cleanDisplayNames(masters.zip_types)
    : ['Coil', 'Waterproof', 'Vislon Zip', 'Two way Coil', 'Metal', 'Reverse Coil', 'Aluminium', 'Plastic']);

  let zipSliderOptions = $derived((masters.zip_sliders && masters.zip_sliders.length > 0)
    ? cleanDisplayNames(masters.zip_sliders)
    : ['Auto-lock', 'DA8DR4', 'DADR7 EJ PB 12', 'metal', 'rubber', 'DA8DR7', 'DABDR4']);

  let zipClosureOptions = $derived((masters.zip_closures && masters.zip_closures.length > 0)
    ? cleanDisplayNames(masters.zip_closures)
    : ['open ended', 'closed ended', '2 way', 'roll']);

  let teethColorOptions = $derived((masters.zip_finishes && masters.zip_finishes.length > 0)
    ? cleanDisplayNames(masters.zip_finishes)
    : ['Antique Brass', 'Gunmetal', 'Shiny Silver', 'Matte Black', 'Golden', 'Antique silver', 'Gold', 'Matt silver', 'Matt Gold']);

  let packagingTypeOptions = $derived.by(() => {
    const cleaned = cleanDisplayNames(masters.packaging_types || []);
    return cleaned.length > 0
      ? cleaned
      : ['Polybag', 'Carton Tape', 'Hangtag', 'Loop Pins', 'Butter Paper', 'Carton', 'Sticker', 'Shrink wrap', 'Courier Bag', 'Other'];
  });

  let trimTypeOptions = $derived((masters.trim_types && masters.trim_types.length > 0)
    ? cleanDisplayNames(masters.trim_types)
    : ['Elastic', 'Tapes', 'Cord', 'Button', 'Fusing', 'Thread', 'Eyelets', 'Stopper', 'Other Trims']);

  let genderOptions = $derived.by(() => {
    const raw = cleanDisplayNames(masters.genders || []);
    const base = raw.length > 0 ? raw : ['Men', 'Women', 'Unisex', 'Infant', 'Jr', 'Kids'];
    return [...base, 'All Apparel'];
  });

  let labelGenreOptions = $derived((masters.label_genres && masters.label_genres.length > 0)
    ? cleanDisplayNames(masters.label_genres)
    : ['Woven', 'Printed', 'DTF', 'Tareaway']);

  const qcRemarkOptions = [
    'Clean / OK - Flawless inspection pass',
    'Minor shade variation within tolerance',
    'Knitting snags / pinholes detected',
    'Oil stain on selvedge edge',
    'Dia measurement deviation',
    'Weight out of tolerance',
    'Slider sticking / rough action',
    'Color mismatch against approved master swatch'
  ];

  // Common quick-staging warehouse bins for one-tap allocation
  let quickBins = $derived((masters.locations && masters.locations.length > 0)
    ? masters.locations.slice(0, 8).map(l => l.code)
    : ['1.M/R/O', '2.M/R/W', '3.M/R/C', '4.Cutting Room', 'A/R01/S01', 'A/R01/S02', 'B/R02/S01', 'B/R02/S02']);

  // Master Entities & Categorized Lists
  let vendorOptions = $derived((masters.vendors || []).map(v => ({ value: v.name, label: v.name })));
  let colorOptions = $derived((masters.colors || []).map(c => ({ value: c.display_name, label: c.display_name, code: c.code })));
  let locationOptions = $derived((masters.locations || []).map(l => ({ value: l.code, label: `${l.code} - ${l.display_name}`, code: l.code })));
  let uomOptions = $derived((masters.uoms || []).map(u => ({ value: u.code, label: u.display_name, code: u.code })));
  let fabricOptions = $derived((masters.fabrics || []).map(f => ({ value: f.display_name, label: f.display_name })));

  let gateKind = $derived(normInwardKind(gateInwardType));
  let floorKind = $derived(normInwardKind(floorLotDetail?.lot?.inwardType));
  let isFloorFabric = $derived(floorLotDetail ? floorKind === 'fabric' : false);

  let compositionOptions = $derived((masters.compositions && masters.compositions.length > 0)
    ? cleanDisplayNames(masters.compositions)
    : ['95%Poly 5%SP', '100%Poly', '100%Cot', '60%Cot 40%Poly', '95%Cot 5%SP']);

  let zipTeethSizeOptions = $derived(getZipTeethSizes(masters));

  function packagingVariantOptionsFor(pType) {
    return getPackagingVariants(masters, pType);
  }

  function packagingSizeOptionsFor(pType) {
    return getPackagingSizes(masters, pType);
  }

  function trimItemsForType(tType) {
    return getTrimItems(masters, tType);
  }

  function trimSizesForType(tType) {
    return getTrimSizes(masters, tType);
  }

  function trimColorsForType(tType) {
    return getTrimColors(masters, tType, colorOptions);
  }

  let finishedGenreOptions = $derived(getGenreOptionsForInward(masters, gateInwardType));

  let activeGateCtx = $derived({
    genre: gateKind === 'non_apparel' ? normalizeNonApparelGenre(gateInwardType) : gateFgType,
    gender: gateGender,
    trimType: gateTrimType,
    trimSize: gateTrimSize,
    packagingType: gatePackagingType,
    packagingSize: gatePackagingSize,
    labelCategory: gateLabelCategory,
    labelGenre: gateLabelGenre,
    zipType: gateZipType,
    teethSize: gateZipTeethSize,
    closure: gateClosure
  });

  let bundleSizeSuggestions = $derived(getAllowedSizesForInward(masters, gateInwardType, activeGateCtx));

  let fpccProductOptions = $derived(getProductOptionsForInward(masters, gateInwardType, activeGateCtx));

  let floorBundleSizeSuggestions = $derived.by(() => {
    if (!floorLotDetail?.lot) return [];
    const lot = floorLotDetail.lot;
    return getAllowedSizesForInward(masters, lot.inwardType, {
      genre: floorAttrs.finishedGoodType || normalizeNonApparelGenre(lot.inwardType),
      gender: floorAttrs.gender || '',
      trimType: floorAttrs.trimType || '',
      packagingType: floorAttrs.packagingType || ''
    });
  });

  function applyInwardTypeTransition(newType, preserveUom = false) {
    const clean = (newType || '').trim();
    const nextKind = normInwardKind(clean);
    if (nextKind === 'unset') return;
    if (clean === lastActiveInwardType) return;

    lastActiveInwardType = clean;
    customBundleDistribution = false;
    gateFpccProduct = '';
    gateVendorProduct = '';
    gateVendorFgName = '';
    gateVendorZipName = '';
    gateVendorTrimName = '';
    gateVendorPackagingName = '';
    gateFpccColor = '';
    gateVendorColor = '';
    gateTrimRate = '';

    if (nextKind === 'non_apparel') {
      gateFgType = normalizeNonApparelGenre(clean);
    } else if (nextKind === 'other') {
      const naGenres = getGenreOptionsForInward(masters, clean);
      gateFgType = naGenres[0] || 'Others';
    } else if (nextKind === 'apparel') {
      const appGenres = getGenreOptionsForInward(masters, clean);
      if (!appGenres.includes(gateFgType)) {
        gateFgType = appGenres.includes('Polo') ? 'Polo' : (appGenres[0] || 'Polo');
      }
      if (!gateGender) gateGender = 'Men';
    } else if (nextKind === 'trims') {
      gateTrimCordEnd = '';
      gateTrimFusingGsm = '';
      gateTrimThreadLength = '';
      gateTrimSize = getTrimSizes(masters, gateTrimType)[0] || '';
    } else if (nextKind === 'packaging') {
      const variants = getPackagingVariants(masters, gatePackagingType);
      const pSizes = getPackagingSizes(masters, gatePackagingType);
      gatePackagingVariant = variants[0] || '';
      gatePackagingSize = pSizes[0] || '';
    } else if (nextKind === 'zips') {
      const tSizes = getZipTeethSizes(masters);
      if (!tSizes.includes(gateZipTeethSize)) {
        gateZipTeethSize = tSizes[0] || '5.0';
      }
    } else if (nextKind === 'label') {
      const lGenres = cleanDisplayNames(masters?.label_genres || []);
      if (lGenres.length > 0 && !lGenres.includes(gateLabelGenre)) {
        gateLabelGenre = lGenres[0];
      }
      const lOpts = getProductOptionsForInward(masters, clean, { labelCategory: gateLabelCategory });
      gateLabelType = lOpts[0]?.value || '';
    }

    if (!preserveUom) {
      gateUom = defaultUomForInward(clean, gateTrimType);
    }

    const nextCtx = {
      genre: nextKind === 'non_apparel' ? normalizeNonApparelGenre(clean) : gateFgType,
      gender: gateGender,
      trimType: gateTrimType,
      trimSize: gateTrimSize,
      packagingType: gatePackagingType,
      packagingSize: gatePackagingSize,
      labelCategory: gateLabelCategory,
      labelGenre: gateLabelGenre,
      zipType: gateZipType,
      teethSize: gateZipTeethSize,
      closure: gateClosure
    };
    const nextAllowed = getAllowedSizesForInward(masters, clean, nextCtx);
    const nextDefault = getDefaultSizeSpecForInward(masters, clean, nextCtx);
    gateBundles = gateBundles.map(b => ({
      ...b,
      sizeSpec: reconcileBundleSizeSpec('', nextAllowed, nextDefault)
    }));
  }

  function onApparelGenreChange(val) {
    gateFgType = val;
    if ((val || '').toLowerCase() === 'onsies') {
      gateGender = 'Infant';
    }
    const valid = getProductOptionsForInward(masters, gateInwardType, { ...activeGateCtx, genre: val });
    if (gateFpccProduct && !valid.some(p => p.value === gateFpccProduct)) {
      gateFpccProduct = '';
    }
  }

  function onGenderChange(val) {
    gateGender = val;
    const allowed = getAllowedSizesForInward(masters, gateInwardType, { ...activeGateCtx, gender: val });
    const def = getDefaultSizeSpecForInward(masters, gateInwardType, { ...activeGateCtx, gender: val });
    gateBundles = gateBundles.map(b => ({
      ...b,
      sizeSpec: reconcileBundleSizeSpec(b.sizeSpec, allowed, def)
    }));
  }

  function onTrimTypeChange(val) {
    gateTrimType = val;
    gateFpccProduct = '';
    gateVendorTrimName = '';
    gateFpccColor = '';
    gateTrimCordEnd = '';
    gateTrimFusingGsm = '';
    gateTrimThreadLength = '';
    gateTrimSize = getTrimSizes(masters, val)[0] || '';
    if (!selectedInvoice?.uom) {
      gateUom = defaultUomForInward(gateInwardType, val);
    }
  }

  function onPackagingTypeChange(val) {
    gatePackagingType = val;
    gateFpccProduct = '';
    gateVendorPackagingName = '';
    gatePackagingVariant = getPackagingVariants(masters, val)[0] || '';
    gatePackagingSize = getPackagingSizes(masters, val)[0] || '';
  }

  function onLabelCategoryChange() {
    const lOpts = getProductOptionsForInward(masters, gateInwardType, { labelCategory: gateLabelCategory });
    gateLabelType = lOpts[0]?.value || '';
  }

  $effect(() => {
    const cur = (gateInwardType || '').trim();
    if (cur && normInwardKind(cur) !== 'unset' && cur !== lastActiveInwardType) {
      untrack(() => applyInwardTypeTransition(cur, Boolean(selectedInvoice?.inwardType)));
    }
  });

  function redistributeGateBundlesEvenly() {
    customBundleDistribution = false;
    const total = parseFloat(gateInvoiceQty) || 0;
    const count = gateBundles.length || 1;
    const per = total > 0 ? Math.round((total / count) * 1000) / 1000 : 0;
    let rem = total > 0 ? Math.round((total - per * count) * 1000) / 1000 : 0;
    gateBundles = gateBundles.map((b, idx) => ({
      ...b,
      invoicedQty: idx === count - 1 ? Math.round((per + rem) * 1000) / 1000 : per
    }));
  }

  function onGateBundleQtyChange() {
    customBundleDistribution = true;
    const sum = gateBundles.reduce((acc, b) => acc + (parseFloat(b.invoicedQty) || 0), 0);
    gateInvoiceQty = String(Math.round(sum * 1000) / 1000);
  }

  // Re-generate Station 1A bundle stubs when count, products, or parentCode changes
  $effect(() => {
    const pCode = gateParentCode || 'YYMMDD-A';
    const invTotal = parseFloat(gateInvoiceQty) || 0;
    const list = [];

    if (gateKind === 'fabric' && fabricProducts.length > 0) {
      const totalCount = fabricTotalRolls || 1;
      const defaultPer = invTotal > 0 ? Math.round((invTotal / totalCount) * 1000) / 1000 : 0;
      let rem = invTotal > 0 ? Math.round((invTotal - defaultPer * totalCount) * 1000) / 1000 : 0;
      let seq = 0;
      for (const fp of fabricProducts) {
        const rCount = Number(fp.rollCount) || 1;
        for (let r = 0; r < rCount; r++) {
          seq++;
          const existing = (untrack(() => gateBundles))[seq - 1] || {};
          const evenVal = seq === totalCount ? Math.round((defaultPer + rem) * 1000) / 1000 : defaultPer;
          list.push({
            sequence: seq,
            total: totalCount,
            bundleCode: `${pCode}-${seq}/${totalCount}`,
            itemName: fp.fpccProduct || fp.vendorProduct || fp.fabricType,
            colorName: fp.color || 'Navy Blue',
            sizeSpec: `${fp.fabricType} (Dia: ${fp.dia || '60"'}, ${fp.gsm || 180} GSM)`,
            invoicedQty: customBundleDistribution && existing.invoicedQty !== undefined ? existing.invoicedQty : evenVal,
            rate: existing.rate !== undefined && existing.rate !== '' ? existing.rate : fp.rate,
            status: 'Pending'
          });
        }
      }
    } else {
      const count = Number(gateNoOfBundles) || 1;
      const defaultPer = invTotal > 0 ? Math.round((invTotal / count) * 1000) / 1000 : 0;
      let rem = invTotal > 0 ? Math.round((invTotal - defaultPer * count) * 1000) / 1000 : 0;
      const rateVal = parseFloat(gateTrimRate) || 0;
      const defaultSpec = getDefaultSizeSpecForInward(masters, gateInwardType, activeGateCtx);
      const allowedSpecs = bundleSizeSuggestions;
      const defaultZipItem = gateKind === 'zips'
        ? `${gateZipType || 'Metal'} #${gateZipTeethSize || '5.0'} ${gateClosure || 'open ended'}`.trim()
        : '';
      for (let i = 1; i <= count; i++) {
        const existing = (untrack(() => gateBundles))[i - 1] || {};
        const evenVal = i === count ? Math.round((defaultPer + rem) * 1000) / 1000 : defaultPer;
        list.push({
          sequence: i,
          total: count,
          bundleCode: `${pCode}-${i}/${count}`,
          itemName: gateFpccProduct || gateVendorProduct || (gateKind === 'label' ? gateLabelType : (gateKind === 'zips' ? (gateVendorZipName || defaultZipItem) : gateInwardType)),
          colorName: gateKind === 'label' ? (gateLabelBaseColor || gateFpccColor || '') : (gateKind === 'zips' ? (gateTapeColor || gateTeethColor || '') : (gateFpccColor || gateVendorColor || '')),
          sizeSpec: reconcileBundleSizeSpec(existing.sizeSpec, allowedSpecs, defaultSpec),
          invoicedQty: customBundleDistribution && existing.invoicedQty !== undefined ? existing.invoicedQty : evenVal,
          rate: existing.rate !== undefined && existing.rate !== '' ? existing.rate : rateVal,
          status: 'Pending'
        });
      }
    }
    gateBundles = list;
  });

  onMount(async () => {
    // Read URL query parameter ?station=1A/1B/1C for direct independent device launch
    const urlParams = new URLSearchParams(window.location.search);
    const stationParam = urlParams.get('station');
    if (stationParam && ['1A', '1B', '1C'].includes(stationParam)) {
      activeStation = stationParam;
    }

    await refreshParentCode();
    await refreshQueues();
    await handleSearchInvoices('');
  });

  function setStation(stationKey) {
    activeStation = stationKey;
    const url = new URL(window.location.href);
    url.searchParams.set('station', stationKey);
    window.history.replaceState(null, '', url.toString());
  }

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

  // Invoice Search & Autofill
  async function handleSearchInvoices(q) {
    try {
      isSearchingInvoices = true;
      existingInvoices = await api.searchInvoices(q);
    } catch (err) {
      console.error('Failed to search invoices:', err);
    } finally {
      isSearchingInvoices = false;
    }
  }

  function handleSelectExistingInvoice(inv) {
    selectedInvoice = inv;
    gateInvoiceNo = inv.invoiceNo;
    gateVendorName = inv.vendorName;
    gateInwardDate = inv.invoiceDate;
    if (inv.inwardType) {
      gateInwardType = inv.inwardType;
      applyInwardTypeTransition(inv.inwardType, true);
    }
    gateInvoiceAmount = inv.invoiceAmount;
    gateUom = inv.uom || defaultUomForInward(inv.inwardType || gateInwardType, gateTrimType);
    taggedLotsOnInvoice = inv.taggedLots || [];
    if (window.showToast) {
      window.showToast(`Autofilled invoice ${inv.invoiceNo} (Locked to Inward Type: ${inv.inwardType || gateInwardType}, ${taggedLotsOnInvoice.length} lots tagged)`, 'info');
    }
  }

  function unlockSelectedInvoice() {
    selectedInvoice = null;
    taggedLotsOnInvoice = [];
    gateInvoiceNo = '';
  }

  // ---------------------------------------------------------
  // STATION 1A ACTIONS (INVOICE ENTRY - DESKTOP)
  // ---------------------------------------------------------
  async function handleGateSubmit(e) {
    e.preventDefault();
    if (gateKind === 'unset') {
      if (window.showToast) window.showToast('Please select a valid Inward Type first', 'error');
      return;
    }
    const bundleQtySum = gateBundles.reduce((acc, b) => acc + (parseFloat(b.invoicedQty) || 0), 0);
    const effectiveInvoiceQty = (parseFloat(gateInvoiceQty) || 0) > 0 ? parseFloat(gateInvoiceQty) : bundleQtySum;
    if (!gateParentCode || !gateVendorName.trim() || !gateInvoiceNo.trim() || effectiveInvoiceQty <= 0) {
      if (window.showToast) window.showToast('Please enter Parent Code, Vendor, Invoice No, and positive Invoice Qty (overall or per bundle)', 'error');
      return;
    }

    if (gateKind !== 'fabric' && bundleSizeSuggestions.length > 0) {
      const invalidBundle = gateBundles.find(b => (b.sizeSpec || '').trim() && !bundleSizeSuggestions.includes(b.sizeSpec.trim()));
      if (invalidBundle) {
        if (window.showToast) window.showToast(`Invalid size/spec "${invalidBundle.sizeSpec}" for ${gateInwardType}. Select an allowed ${gateInwardType} size.`, 'error');
        return;
      }
    }

    if (fpccProductOptions.length > 0 && gateFpccProduct.trim() && gateKind !== 'fabric' && gateKind !== 'zips') {
      const validProd = fpccProductOptions.some(p => p.value === gateFpccProduct.trim() || p.label === gateFpccProduct.trim());
      if (!validProd) {
        if (window.showToast) window.showToast(`Product "${gateFpccProduct}" does not belong to category ${gateInwardType}. Pick a valid item from the dropdown.`, 'error');
        return;
      }
    }

    isSubmitting = true;
    try {
      const attrs = {};
      let fabricPayload = [];

      if (gateKind === 'fabric') {
        fabricPayload = fabricProducts.map(fp => ({
          id: fp.id,
          fabricType: fp.fabricType,
          fpccProduct: fp.fpccProduct,
          vendorProduct: fp.vendorProduct,
          color: fp.color,
          uom: fp.uom || 'Kgs',
          dia: fp.dia || '60"',
          gsm: Number(fp.gsm) || 180,
          composition: fp.composition,
          rate: Number(fp.rate) || 0,
          rollCount: Number(fp.rollCount) || 1
        }));

        /** @type {Record<string, any>} */
        const primary = fabricProducts[0] || {};
        attrs.fabricType = primary.fabricType;
        attrs.invoiceGsm = Number(primary.gsm) || 0;
        attrs.dia = parseFloat(String(primary.dia || '').replace(/[^0-9.]/g, '')) || 0;
        attrs.composition = primary.composition;
        attrs.rate = Number(primary.rate) || 0;
        attrs.vendorFabName = primary.vendorProduct;
        attrs.fpccFabName = primary.fpccProduct;
      } else if (gateKind === 'zips') {
        attrs.vendorZipName = gateVendorZipName || gateVendorProduct;
        attrs.zipType = gateZipType;
        attrs.teethSize = gateZipTeethSize;
        attrs.tapeColor = gateTapeColor;
        attrs.teethColor = gateTeethColor;
        attrs.slider = gateSlider;
        attrs.closure = gateClosure;
      } else if (gateKind === 'packaging') {
        attrs.packagingType = gatePackagingType;
        attrs.packagingVariant = gatePackagingVariant;
        attrs.packagingSize = gatePackagingSize;
        attrs.fpccPackagingName = gateFpccProduct;
        attrs.vendorPackagingName = gateVendorPackagingName || gateVendorProduct;
      } else if (gateKind === 'trims') {
        attrs.trimType = gateTrimType;
        attrs.trimSize = gateTrimSize;
        attrs.trimRate = parseFloat(gateTrimRate) || 0;
        attrs.fpccTrimName = gateFpccProduct;
        attrs.vendorTrimName = gateVendorTrimName || gateVendorProduct;
        if (gateTrimCordEnd) attrs.cordEnd = gateTrimCordEnd;
        if (gateTrimFusingGsm) attrs.fusingGsm = gateTrimFusingGsm;
        if (gateTrimThreadLength) attrs.threadLength = gateTrimThreadLength;
      } else if (gateKind === 'apparel' || gateKind === 'non_apparel' || gateKind === 'other') {
        const resolvedGenre = gateKind === 'non_apparel' ? normalizeNonApparelGenre(gateInwardType) : gateFgType;
        attrs.finishedGoodType = resolvedGenre;
        if (gateKind === 'apparel') {
          attrs.gender = gateGender;
        }
        attrs.vendorFgName = gateVendorFgName || gateVendorProduct;
        attrs.fpccFgName = gateFpccProduct;
      } else if (gateKind === 'label') {
        attrs.labelCategory = gateLabelCategory;
        attrs.labelType = gateLabelType;
        attrs.labelGenre = gateLabelGenre;
        attrs.labelDescription = gateLabelDescription;
        attrs.baseColor = gateLabelBaseColor;
        attrs.textColor = gateLabelTextColor;
      }

      const stubsPayload = gateBundles.map(b => ({
        sequence: b.sequence,
        invoicedQty: parseFloat(b.invoicedQty) || 0,
        sizeSpec: b.sizeSpec,
        rate: parseFloat(b.rate) || 0
      }));

      const resolvedZipFpcc = gateFpccProduct || `${gateZipType || 'Metal'} #${gateZipTeethSize || '5.0'} ${gateClosure || 'open ended'}`.trim();
      const payload = {
        parentCode: gateParentCode,
        inwardDate: gateInwardDate,
        inwardType: gateInwardType,
        vendorName: gateVendorName,
        invoiceNo: gateInvoiceNo,
        invoiceQty: effectiveInvoiceQty,
        invoiceAmount: parseFloat(gateInvoiceAmount) || 0,
        uom: gateUom,
        vendorProduct: gateKind === 'fabric'
          ? (fabricProducts[0]?.vendorProduct || '')
          : (gateVendorFgName || gateVendorZipName || gateVendorTrimName || gateVendorPackagingName || gateVendorProduct),
        fpccProduct: gateKind === 'fabric'
          ? (fabricProducts[0]?.fpccProduct || '')
          : (gateFpccProduct || (gateKind === 'label' ? gateLabelType : (gateKind === 'zips' ? resolvedZipFpcc : ''))),
        vendorColor: gateKind === 'fabric' ? (fabricProducts[0]?.color || '') : gateVendorColor,
        fpccColor: gateKind === 'fabric' ? (fabricProducts[0]?.color || '') : (gateFpccColor || gateLabelBaseColor),
        noOfBundles: gateKind === 'fabric' ? fabricTotalRolls : (Number(gateNoOfBundles) || 1),
        comment: gateComment,
        attrs,
        bundles: stubsPayload,
        fabricProducts: fabricPayload,
        qrCopies: Number(qrCopiesPerBundle) || 3
      };

      const res = await api.gateInward(payload);
      if (window.showToast) {
        window.showToast(`✓ Invoice Entry ${res.parentCode} booked! (${res.bundles.length} bundle stubs generated)`, 'success');
      }

      const bookedCode = gateParentCode;
      customBundleDistribution = false;
      await refreshParentCode();
      await refreshQueues();
      await handleSearchInvoices('');

      if (confirm(`Lot ${bookedCode} successfully staged at dock!\n\nWould you like to switch to Station 1B (Inward Verification) to measure & reconcile this lot now?`)) {
        setStation('1B');
        await selectLotForFloorCount(bookedCode);
      }
    } catch (err) {
      if (window.showToast) window.showToast('Invoice Entry failed: ' + err.message, 'error');
    } finally {
      isSubmitting = false;
    }
  }

  // ---------------------------------------------------------
  // STATION 1B ACTIONS (INWARD COUNT - TABLET)
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
      const rawBundles = Array.isArray(detail?.bundles) ? detail.bundles : [];
      floorInvoicedQty = detail.lot.invoiceQty;
      floorCountedQty = detail.lot.countedQty > 0 ? detail.lot.countedQty : detail.lot.invoiceQty;
      floorLocationBin = rawBundles[0]?.locationBin || 'A/R01/S01';
      floorComments = detail.lot.remarks || '';
      const loadedAttrs = { ...createEmptyFloorAttrs(), ...(detail.lot.attrs || {}) };
      const lotKind = normInwardKind(detail.lot.inwardType);
      if (lotKind === 'non_apparel') {
        loadedAttrs.finishedGoodType = normalizeNonApparelGenre(detail.lot.inwardType) || loadedAttrs.finishedGoodType || '';
      } else if (lotKind === 'apparel' && !loadedAttrs.gender) {
        loadedAttrs.gender = 'Men';
      } else if (lotKind === 'fabric') {
        if (!loadedAttrs.fabricType && detail.lot.fabricProducts?.[0]?.fabricType) {
          loadedAttrs.fabricType = detail.lot.fabricProducts[0].fabricType;
        }
        if (!loadedAttrs.invoiceGsm && detail.lot.fabricProducts?.[0]?.gsm) {
          loadedAttrs.invoiceGsm = detail.lot.fabricProducts[0].gsm;
        }
        if (!loadedAttrs.dia && detail.lot.fabricProducts?.[0]?.dia) {
          loadedAttrs.dia = detail.lot.fabricProducts[0].dia;
        }
      }
      for (const k of FLOOR_ATTR_KEYS) {
        loadedAttrs[k] = loadedAttrs[k] !== null && loadedAttrs[k] !== undefined ? String(loadedAttrs[k]) : '';
      }
      if (lotKind === 'fabric' && loadedAttrs.dia) {
        const cleanDia = loadedAttrs.dia.replace(/"/g, '').trim();
        loadedAttrs.dia = cleanDia ? `${cleanDia}"` : '';
      }
      floorAttrs = loadedAttrs;
      floorLotDetail = { ...detail, bundles: rawBundles };

      const lotIsFabric = lotKind === 'fabric';
      const defaultRate = detail.lot.invoiceQty > 0 ? Math.round((detail.lot.invoiceAmount / detail.lot.invoiceQty) * 100) / 100 : 0;

      const bList = rawBundles.map(b => ({
        bundleCode: b.bundleCode,
        sequence: b.bundleSequence,
        itemName: b.itemName || detail.lot.fpccProduct || detail.lot.inwardType,
        colorName: b.colorName || detail.lot.fpccColor || 'Standard',
        invoicedQty: b.invoicedQty > 0 ? b.invoicedQty : Math.round((detail.lot.invoiceQty / (rawBundles.length || 1)) * 1000) / 1000,
        countedQty: b.inwardQty > 0 ? b.inwardQty : (b.invoicedQty > 0 ? b.invoicedQty : Math.round((detail.lot.invoiceQty / (rawBundles.length || 1)) * 1000) / 1000),
        locationBin: b.locationBin || 'A/R01/S01',
        measuredGsm: lotIsFabric ? String(b.measuredGsm || loadedAttrs.invoiceGsm || '180') : '',
        measuredDia: lotIsFabric ? String(b.measuredDia || loadedAttrs.dia || '60"') : '',
        sizeSpec: b.sizeSpec || '',
        rate: b.rate || defaultRate,
        comment: b.bundleComment || ''
      }));
      floorBundles = bList;
    } catch (err) {
      if (window.showToast) window.showToast('Failed to load lot: ' + err.message, 'error');
    } finally {
      isLoading = false;
    }
  }

  let floorDiscrepancy = $derived.by(() => {
    const invQ = parseFloat(floorInvoicedQty) || 0;
    const cntQ = parseFloat(floorCountedQty) || 0;
    const unitRate = floorLotDetail?.lot?.invoiceAmount && invQ > 0 ? (floorLotDetail.lot.invoiceAmount / invQ) : 0;

    if (invQ <= 0 || cntQ <= 0) return null;

    const variance = cntQ - invQ;
    const variancePctNum = Math.abs((variance / invQ) * 100);
    const variancePct = ((variance / invQ) * 100).toFixed(2);
    const variancePctAbs = variancePctNum.toFixed(2);
    const financialExposure = Math.abs(variance * unitRate);

    if (variance < -0.05) {
      return {
        status: 'SHORTAGE_FLAGGED',
        type: 'danger',
        variance: variance.toFixed(2),
        variancePct: variancePctNum,
        exposure: financialExposure.toFixed(2),
        title: `🚨 Dock Shortage Discrepancy: ${Math.abs(variance).toFixed(2)} ${floorLotDetail?.lot?.uom || 'Kgs'} (${variancePctAbs}% Shortage)`,
        desc: `Physical counted weight is less than billed supplier invoice. Debit note exposure: ₹${financialExposure.toLocaleString('en-IN', { minimumFractionDigits: 2 })}. Placing on DISCREPANCY_HOLD freezes supplier debit.`
      };
    } else if (variance > 0.05) {
      return {
        status: 'EXCESS_FLAGGED',
        type: 'warning',
        variance: variance.toFixed(2),
        variancePct: variancePctNum,
        exposure: financialExposure.toFixed(2),
        title: `⚠️ Dock Excess Received: +${variance.toFixed(2)} ${floorLotDetail?.lot?.uom || 'Kgs'} (+${variancePct}%)`,
        desc: `Physical received exceeds billed supplier invoice. Reconcile with delivery challan.`
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

  function distributeFloorCountEvenly() {
    const total = parseFloat(floorCountedQty) || 0;
    const count = floorBundles.length;
    if (count === 0 || total <= 0) return;
    const perRoll = Math.round((total / count) * 100) / 100;
    let remainder = Math.round((total - (perRoll * count)) * 100) / 100;

    floorBundles = floorBundles.map((b, idx) => {
      let q = perRoll;
      if (idx === count - 1) q = Math.round((perRoll + remainder) * 100) / 100;
      return { ...b, countedQty: q };
    });
    if (window.showToast) window.showToast(`Divided ${total} evenly across ${count} bundles (${perRoll} each)`, 'info');
  }

  function autoSumFloorBundles() {
    const sum = floorBundles.reduce((acc, b) => acc + (parseFloat(b.countedQty) || 0), 0);
    floorCountedQty = String(Math.round(sum * 100) / 100);
    if (window.showToast) window.showToast(`Aggregated bundle sum: ${floorCountedQty}`, 'info');
  }

  async function handleFloorSubmit(e) {
    e.preventDefault();
    if (!floorSelectedLotCode) return;

    if (!isFloorFabric && floorBundleSizeSuggestions.length > 0) {
      const bad = floorBundles.find(b => (b.sizeSpec || '').trim() && !floorBundleSizeSuggestions.includes(b.sizeSpec.trim()));
      if (bad) {
        if (window.showToast) window.showToast(`Invalid size/spec "${bad.sizeSpec}" for ${floorLotDetail?.lot?.inwardType}. Select a valid option from the size dropdown.`, 'error');
        return;
      }
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
        attrs: { ...floorAttrs },
        sizeLines: floorBundles.map(b => ({
          bundleCode: b.bundleCode,
          size: b.sizeSpec,
          invoiced: parseFloat(b.invoicedQty) || 0,
          counted: parseFloat(b.countedQty) || 0,
          rate: parseFloat(b.rate) || 0
        })),
        bundles: floorBundles.map(b => ({
          bundleCode: b.bundleCode,
          sequence: b.sequence,
          invoicedQty: parseFloat(b.invoicedQty) || 0,
          countedQty: parseFloat(b.countedQty) || 0,
          locationBin: b.locationBin || floorLocationBin,
          measuredGsm: isFloorFabric ? (parseFloat(b.measuredGsm) || 0) : 0,
          measuredDia: isFloorFabric ? (parseFloat(String(b.measuredDia).replace(/[^0-9.]/g, '')) || 0) : 0,
          sizeSpec: b.sizeSpec,
          rate: parseFloat(b.rate) || 0,
          attrs: { ...floorAttrs },
          comment: b.comment
        }))
      };

      const res = await api.recordFloorCount(payload);
      if (window.showToast) {
        window.showToast(`✓ Inward count recorded for ${res.parentCode}! Status: ${res.status}`, 'success');
      }

      await refreshQueues();
      const reconciledCode = floorSelectedLotCode;

      if (res.stage === 'DISCREPANCY_HOLD') {
        if (window.showToast) window.showToast(`⚠️ Lot placed on DISCREPANCY_HOLD. Shortage: ${Math.abs(res.varianceQty)} ${floorLotDetail?.lot?.uom}`, 'warning');
      } else {
        if (confirm(`Lot ${reconciledCode} advanced to AWAITING_QC!\n\nWould you like to switch to Station 1C (QC Inspection) now?`)) {
          setStation('1C');
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
  // STATION 1C ACTIONS (QC INSPECTION - PHONE)
  // ---------------------------------------------------------
  async function selectLotForQC(pCode) {
    qcSelectedLotCode = pCode;
    qcActiveBundleIndex = 0;
    if (!pCode) {
      qcLotDetail = null;
      return;
    }

    try {
      isLoading = true;
      const detail = await api.getLot(pCode);
      const rawBundles = Array.isArray(detail?.bundles) ? detail.bundles : [];
      qcLotDetail = { ...detail, bundles: rawBundles };

      qcBundles = rawBundles.map(b => ({
        id: b.id,
        bundleCode: b.bundleCode,
        sequence: b.bundleSequence,
        itemName: b.itemName || detail.lot.fpccProduct,
        colorName: b.colorName || detail.lot.fpccColor,
        inwardQty: b.inwardQty,
        locationBin: b.locationBin,
        sizeSpec: b.sizeSpec,
        qcStatus: b.qcStatus === 'Pending' ? 'Passed' : b.qcStatus,
        qcRemarks: b.qcRemarks || '',
        qcImageUrl: b.qcImageUrl || '',
        qcAudioUrl: b.qcAudioUrl || '',
        qcTranscript: b.qcTranscript || '',
        qcAudioBlobUrl: '',
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
    if (window.showToast) window.showToast('All rolls marked PASS', 'success');
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

  // Audio Recording & Speech Recognition for QC
  async function startAudioRecording(bundleIdx) {
    activeRecordingBundleIdx = bundleIdx;
    isRecordingAudio = true;
    audioChunks = [];
    recordingTimerSeconds = 0;

    clearInterval(recordingTimerInterval);
    recordingTimerInterval = setInterval(() => {
      recordingTimerSeconds++;
    }, 1000);

    // 1. Initialize Speech-to-Text Transcription via Web Speech API
    const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
    if (SpeechRecognition) {
      try {
        speechRecognitionInstance = new SpeechRecognition();
        speechRecognitionInstance.continuous = true;
        speechRecognitionInstance.interimResults = true;
        speechRecognitionInstance.lang = 'en-IN';

        const initialRemarks = qcBundles[bundleIdx].qcRemarks || '';

        speechRecognitionInstance.onresult = (event) => {
          let text = '';
          for (let i = event.resultIndex; i < event.results.length; i++) {
            text += event.results[i][0].transcript;
          }
          if (text.trim()) {
            const prefix = initialRemarks ? initialRemarks.trim() + ' • ' : '';
            qcBundles[bundleIdx].qcRemarks = prefix + text.trim();
            qcBundles[bundleIdx].qcTranscript = text.trim();
          }
        };

        speechRecognitionInstance.onerror = (e) => {
          console.warn('Speech recognition notice:', e.error);
        };

        speechRecognitionInstance.start();
      } catch (err) {
        console.warn('Speech recognition start failed:', err);
      }
    }

    // 2. Microphone Recording via MediaRecorder
    try {
      if (navigator.mediaDevices && navigator.mediaDevices.getUserMedia) {
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
        let mimeType = 'audio/webm';
        if (typeof MediaRecorder !== 'undefined') {
          if (!MediaRecorder.isTypeSupported('audio/webm')) {
            if (MediaRecorder.isTypeSupported('audio/mp4')) mimeType = 'audio/mp4';
            else if (MediaRecorder.isTypeSupported('audio/ogg')) mimeType = 'audio/ogg';
          }
          mediaRecorderInstance = new MediaRecorder(stream, { mimeType });
          mediaRecorderInstance.ondataavailable = (e) => {
            if (e.data && e.data.size > 0) audioChunks.push(e.data);
          };
          mediaRecorderInstance.start(250);
        }
      }
    } catch (err) {
      console.warn('Microphone error:', err);
      if (window.showToast) window.showToast('Microphone note active (type remarks freely)', 'info');
    }
  }

  async function stopAudioRecording(bundleIdx) {
    isRecordingAudio = false;
    clearInterval(recordingTimerInterval);

    // Stop speech recognition
    if (speechRecognitionInstance) {
      try { speechRecognitionInstance.stop(); } catch (e) {}
      speechRecognitionInstance = null;
    }

    // Stop media recorder and create audio upload
    if (mediaRecorderInstance && mediaRecorderInstance.state !== 'inactive') {
      try {
        mediaRecorderInstance.stop();
        mediaRecorderInstance.stream.getTracks().forEach(t => t.stop());

        const mime = mediaRecorderInstance.mimeType || 'audio/webm';
        const audioBlob = new Blob(audioChunks, { type: mime });
        const ext = mime.includes('mp4') ? 'mp4' : (mime.includes('ogg') ? 'ogg' : 'webm');
        const bCode = (qcBundles[bundleIdx]?.bundleCode || 'bundle').replace(/[^a-zA-Z0-9_-]/g, '_');
        const audioFile = new File([audioBlob], `qc_${bCode}_note.${ext}`, { type: mime });

        qcBundles[bundleIdx].qcAudioBlobUrl = URL.createObjectURL(audioBlob);

        // Upload to server asynchronously
        try {
          const res = await api.uploadFile(audioFile);
          if (res && res.url) {
            qcBundles[bundleIdx].qcAudioUrl = res.url;
            if (window.showToast) window.showToast('✓ Audio note recorded & transcribed!', 'success');
          }
        } catch (uploadErr) {
          console.warn('Audio upload notice:', uploadErr);
        }
      } catch (err) {
        console.warn('MediaRecorder stop error:', err);
      }
    }

    activeRecordingBundleIdx = -1;
  }

  async function handleQCSubmit(e) {
    e.preventDefault();
    if (!qcSelectedLotCode || qcBundles.length === 0) {
      if (window.showToast) window.showToast('Please select a lot with bundles to QC', 'error');
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
          audioUrl: b.qcAudioUrl || '',
          transcript: b.qcTranscript || '',
          inspectedBy: qcInspector
        }))
      };

      const res = await api.recordBundleQC(payload);
      if (window.showToast) {
        window.showToast(`✓ QC Inspection committed! ${res.passCount} Passed, ${res.failCount} Rejected.`, 'success');
      }

      await refreshQueues();
      openAllZebraTags();
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
      item: b.itemName || qcLotDetail.lot.fpccProduct || qcLotDetail.lot.vendorProduct || qcLotDetail.lot.inwardType,
      color: b.colorName || qcLotDetail.lot.fpccColor || qcLotDetail.lot.vendorColor || '',
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
        item: b.itemName || qcLotDetail?.lot?.fpccProduct || qcLotDetail?.lot?.vendorProduct,
        color: b.colorName || qcLotDetail?.lot?.fpccColor || qcLotDetail?.lot?.vendorColor,
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
  <!-- Top Navigation & Station Switcher (Desktop, Tablet, Phone Badges) -->
  <header class="station-header card">
    <div class="header-main">
      <div>
        <h2>Inward Intake Pipeline & QC Verification</h2>
        <p class="subtitle">Unified System of Record: Station 1A (Desktop) ➔ Station 1B (Tablet) ➔ Station 1C (Mobile Phone)</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-secondary btn-sm" onclick={refreshQueues}>
          🔄 Refresh Queues
        </button>
        <button class="btn btn-secondary btn-sm" onclick={() => isDiscrepancyModalOpen = true}>
          ⚖️ Discrepancy Registry
        </button>
        <button class="btn btn-secondary btn-sm" onclick={() => isRemnantModalOpen = true}>
          ✂️ Fabric Remnants
        </button>
      </div>
    </div>

    <!-- Station Navigation Tabs with Device Badges -->
    <div class="station-tabs">
      <button 
        class="station-tab-btn {activeStation === '1A' ? 'active' : ''}" 
        onclick={() => setStation('1A')}
      >
        <span class="tab-icon">🖥️</span>
        <div class="tab-text">
          <div class="tab-title">
            Station 1A: Invoice Entry
            <span class="device-pill desktop">Desktop View</span>
          </div>
          <div class="tab-desc">Step 1: Commercial Invoice &amp; Per-Bundle Booking</div>
        </div>
      </button>

      <button 
        class="station-tab-btn {activeStation === '1B' ? 'active' : ''}" 
        onclick={() => setStation('1B')}
      >
        <span class="tab-icon">📟</span>
        <div class="tab-text">
          <div class="tab-title">
            Station 1B: Inward Verification
            <span class="device-pill tablet">Tablet View</span>
            {#if awaitingCountLots.length > 0}
              <span class="badge badge-warning">{awaitingCountLots.length} pending</span>
            {/if}
          </div>
          <div class="tab-desc">Step 2: Measure Inward Qty, Bins &amp; Discrepancies</div>
        </div>
      </button>

      <button 
        class="station-tab-btn {activeStation === '1C' ? 'active' : ''}" 
        onclick={() => setStation('1C')}
      >
        <span class="tab-icon">📱</span>
        <div class="tab-text">
          <div class="tab-title">
            Station 1C: QC Inspection
            <span class="device-pill phone">Phone View</span>
            {#if awaitingQcLots.length > 0}
              <span class="badge badge-info">{awaitingQcLots.length} lots</span>
            {/if}
          </div>
          <div class="tab-desc">Binary Pass/Fail, Audio Notes & Tags</div>
        </div>
      </button>
    </div>
  </header>

  <!-- ========================================================= -->
  <!-- STATION 1A: GATE ENTRY & INWARD BOOKING (DESKTOP)         -->
  <!-- ========================================================= -->
  {#if activeStation === '1A'}
    <form onsubmit={handleGateSubmit} class="station-form">
      <!-- 1. Dedicated Top-Level Invoice Card (Multiple Parent Lots Can Tag to One Invoice) -->
      <section class="card section-card invoice-header-card">
        <div class="section-header">
          <div>
            <h3>📄 1. Supplier Commercial Invoice Header</h3>
            <span class="subtitle">Top-level commercial invoice details. Multiple parent lots can be inwarded and linked under this single invoice.</span>
          </div>
          <span class="badge badge-primary">Invoice System of Record</span>
        </div>

        <!-- Invoice Search & Autofill Quick Bar -->
        <div class="invoice-search-bar" style="margin-bottom: 16px;">
          <label for="inv-search-input" style="font-size:12px;font-weight:700;color:var(--text-secondary);display:block;margin-bottom:6px;">
            Search Existing Invoices for Autofill / Adding Multiple Inward Lots:
          </label>
          <div style="display:flex;gap:10px;">
            <input 
              id="inv-search-input"
              type="text" 
              class="form-control" 
              placeholder="🔍 Search previously recorded invoices by Invoice No or Vendor..." 
              bind:value={invoiceSearchQuery} 
              oninput={(e) => handleSearchInvoices(e.currentTarget.value)} 
            />
          </div>

          {#if existingInvoices.length > 0 && invoiceSearchQuery.trim()}
            <div class="invoice-suggest-dropdown">
              {#each existingInvoices as inv}
                <button 
                  type="button" 
                  class="invoice-suggest-item"
                  onclick={() => handleSelectExistingInvoice(inv)}
                >
                  <div style="display:flex;justify-content:space-between;align-items:center;gap:8px;">
                    <strong>#{inv.invoiceNo}</strong>
                    <div style="display:flex;gap:6px;align-items:center;">
                      {#if inv.inwardType}
                        <span class="badge badge-primary">{inv.inwardType}</span>
                      {/if}
                      <span class="badge badge-secondary">{inv.taggedLots?.length || 0} Lots Tagged</span>
                    </div>
                  </div>
                  <div style="font-size:12px;color:var(--text-secondary);margin-top:2px;">
                    {inv.vendorName} • {inv.invoiceDate} • <strong>{inv.totalLotsQty || inv.invoiceQty} {inv.uom}</strong> • ₹{inv.invoiceAmount}
                  </div>
                </button>
              {/each}
            </div>
          {/if}
        </div>

        <!-- Active Invoice Form Fields (Overall Invoice Qty Removed; 1 Inward Type per Invoice Enforced) -->
        <div class="form-grid">
          <div class="form-group">
            <label for="gate-inv-no">Supplier Invoice No.</label>
            <input
              id="gate-inv-no"
              class="form-control input-mono font-bold"
              bind:value={gateInvoiceNo}
              oninput={() => { if (selectedInvoice && gateInvoiceNo !== selectedInvoice.invoiceNo) selectedInvoice = null; }}
              placeholder="e.g. INV-9902"
              required
            />
          </div>

          <!-- Supplier / Vendor via Combobox -->
          <div class="form-group">
            <span class="form-label-title">Supplier / Vendor Entity</span>
            <Combobox 
              options={vendorOptions} 
              bind:value={gateVendorName} 
              placeholder="Search or enter supplier name..." 
              category="vendor"
              allowCustom={true} 
            />
          </div>

          <div class="form-group">
            <label for="gate-date">Invoice Date</label>
            <input id="gate-date" type="date" class="form-control" bind:value={gateInwardDate} required />
          </div>

          <!-- Only One Inward Type per Invoice -->
          <div class="form-group">
            <div style="display:flex;justify-content:space-between;align-items:center;">
              <span class="form-label-title">Invoice Inward Type (1 Type / Invoice)</span>
              {#if selectedInvoice?.inwardType}
                <button type="button" class="btn btn-secondary btn-xs" onclick={unlockSelectedInvoice} title="Start fresh invoice">Unlock</button>
              {/if}
            </div>
            {#if selectedInvoice?.inwardType}
              <input class="form-control font-bold" value={gateInwardType} readonly title="Locked to existing invoice inward type" />
              <span class="help-text">🔒 Locked to {selectedInvoice.inwardType} (only 1 inward type per invoice)</span>
            {:else}
              <Combobox 
                options={inwardTypeOptions} 
                bind:value={gateInwardType} 
                placeholder="Select inward type..." 
                allowCustom={false}
                onSelect={(item) => applyInwardTypeTransition(item?.value || item?.display_name || gateInwardType)}
              />
            {/if}
          </div>

          <div class="form-group">
            <label for="gate-inv-amt">Invoice Total Amount (₹)</label>
            <input id="gate-inv-amt" type="number" step="0.01" class="form-control" bind:value={gateInvoiceAmount} placeholder="₹ 0.00" />
            {#if gateInvoiceAmount && Number(gateInvoiceQty) > 0}
              <span class="help-text">Avg Invoice Rate: ₹{(parseFloat(gateInvoiceAmount) / Number(gateInvoiceQty)).toFixed(2)} / {gateUom}</span>
            {/if}
          </div>

          <div class="form-group">
            <span class="form-label-title">Unit of Measure (UOM)</span>
            <Combobox 
              options={uomOptions} 
              bind:value={gateUom} 
              placeholder="e.g. Kgs, Pcs, Meter..." 
              category="uom"
              allowCustom={true} 
            />
          </div>
        </div>

        <!-- Tagged Lots on this Invoice Banner -->
        {#if taggedLotsOnInvoice.length > 0}
          <div class="tagged-lots-banner" style="margin-top: 14px;background:var(--bg-tertiary);padding:10px 14px;border-radius:var(--radius-sm);display:flex;align-items:center;gap:10px;">
            <span style="font-weight:700;font-size:12px;color:var(--text-muted);">🏷️ Inward Lots Linked to Invoice #{gateInvoiceNo}:</span>
            <div style="display:flex;gap:6px;flex-wrap:wrap;">
              {#each taggedLotsOnInvoice as lotCode}
                <span class="code-pill">{lotCode}</span>
              {/each}
            </div>
            <span style="font-size:11px;color:var(--text-subtle);margin-left:auto;">(Booking another {gateInwardType} lot under this invoice)</span>
          </div>
        {/if}
      </section>

      <!-- 2. Consignment Parent Lot Details Card (Invoice Entry) -->
      <section class="card section-card">
        <div class="section-header">
          <div>
            <h3>2. Consignment Parent Lot Invoice Entry (Parent Code: {gateParentCode || 'YYMMDD-A'})</h3>
            <span class="subtitle">Step 1 Invoice parameters • Enter lot invoice qty (auto-distributed evenly across bundles, editable per bundle below).</span>
          </div>
          <span class="badge badge-secondary">Step 1: Invoice Spec</span>
        </div>

        <div class="form-grid">
          <div class="form-group">
            <label for="gate-pcode">Parent Code (Auto-Generated Daily Sequence)</label>
            <div class="input-with-action">
              <input id="gate-pcode" class="form-control font-bold" bind:value={gateParentCode} placeholder="YYMMDD-A" required />
              <button type="button" class="btn btn-secondary btn-sm" onclick={refreshParentCode} title="Suggest Next Parent Code">🔄</button>
            </div>
            <span class="help-text">Unified Format: YYMMDD-A (e.g. {gateParentCode || '261028-A'})</span>
          </div>

          <!-- Lot Invoice Quantity (auto-distributes to per-bundle stubs) -->
          <div class="form-group">
            <label for="gate-lot-inv-qty">Lot Invoice Qty ({gateUom})</label>
            <div class="input-with-action">
              <input
                id="gate-lot-inv-qty"
                type="number"
                step="0.01"
                class="form-control font-bold"
                bind:value={gateInvoiceQty}
                oninput={() => { customBundleDistribution = false; }}
                placeholder="e.g. 200"
                required
              />
              <button type="button" class="btn btn-secondary btn-sm" onclick={redistributeGateBundlesEvenly} title="Distribute lot qty evenly across bundles">
                ➗ Split
              </button>
            </div>
            <span class="help-text">Auto-splits across bundles below ({gateBundles.length} bundles)</span>
          </div>

          <!-- QR Copies Stepper (Default 3 copies per bundle) -->
          <div class="form-group">
            <label for="qr-copies">Thermal QR Copies per Bundle</label>
            <div class="input-with-action" style="align-items:center;">
              <button type="button" class="btn btn-secondary btn-sm" onclick={() => { if (qrCopiesPerBundle > 1) qrCopiesPerBundle--; }}>−</button>
              <input id="qr-copies" type="number" min="1" max="10" class="form-control font-bold" style="text-align:center;width:70px;" bind:value={qrCopiesPerBundle} />
              <button type="button" class="btn btn-secondary btn-sm" onclick={() => { if (qrCopiesPerBundle < 10) qrCopiesPerBundle++; }}>+</button>
              <span class="help-text" style="margin-left:8px;">Default 3 (Outer, Core, Swatch)</span>
            </div>
          </div>
        </div>

        <!-- ==================================================== -->
        <!-- DYNAMIC SPECIFICATION PER INWARD TYPE                -->
        <!-- ==================================================== -->
        {#if gateKind === 'unset'}
          <div style="margin-top:20px;border-top:1px solid var(--border-color);padding:16px;background:var(--bg-tertiary);border-radius:var(--radius-sm);font-size:13px;color:var(--text-muted);">
            Select an <strong>Invoice Inward Type</strong> above to load category-locked item catalogs, attributes, and allowed size specs.
          </div>
        {:else if gateKind === 'fabric'}
          <div class="fabric-multi-product-box" style="margin-top:20px;border-top:1px solid var(--border-color);padding-top:16px;">
            <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px;">
              <div>
                <h4 style="margin:0;font-size:15px;color:var(--text-main);">Fabric Multi-Component Consignment Lines</h4>
                <span style="font-size:12px;color:var(--text-muted);">
                  Inward multi-component families (Body Fabric + Collar + Cuff + Rib) under parent code <strong>{gateParentCode}</strong>.
                </span>
              </div>
              <button type="button" class="btn btn-secondary btn-sm" onclick={addFabricProduct}>
                + Add Fabric Component (e.g. Collar / Cuff)
              </button>
            </div>

            <div class="fabric-lines-list">
              {#each fabricProducts as fp, idx}
                <div class="fabric-line-card" style="background:var(--bg-tertiary);border:1px solid var(--border-color);border-radius:var(--radius-md);padding:14px;margin-bottom:12px;">
                  <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:10px;">
                    <div style="display:flex;align-items:center;gap:10px;">
                      <span class="badge badge-primary">Fabric Component #{idx + 1}</span>
                      <input 
                        type="text" 
                        class="form-control" 
                        style="width:200px;font-weight:700;padding:3px 8px;font-size:12px;" 
                        bind:value={fp.role} 
                        placeholder="e.g. Body Fabric / Collar Rib..." 
                      />
                    </div>
                    {#if fabricProducts.length > 1}
                      <button type="button" class="btn btn-danger btn-sm" onclick={() => removeFabricProduct(idx)} title="Remove product line">
                        ✕ Remove
                      </button>
                    {/if}
                  </div>

                  <div class="form-grid">
                    <div class="form-group">
                      <span class="form-label-title">Fabric Type</span>
                      <Combobox 
                        options={fabricTypeOptions} 
                        bind:value={fp.fabricType} 
                        placeholder="Fabric, Collar, Cuff, Rib..." 
                        allowCustom={true} 
                      />
                    </div>

                    <div class="form-group">
                      <span class="form-label-title">FPCC Master Fabric Name</span>
                      <Combobox 
                        options={fabricOptions} 
                        bind:value={fp.fpccProduct} 
                        placeholder="FPCC fabric item..." 
                        allowCustom={true} 
                      />
                    </div>

                    <div class="form-group">
                      <span class="form-label-title">Vendor Fabric Name</span>
                      <input class="form-control" bind:value={fp.vendorProduct} placeholder="Vendor name for fabric..." />
                    </div>

                    <div class="form-group">
                      <span class="form-label-title">Composition</span>
                      <Combobox 
                        options={compositionOptions} 
                        bind:value={fp.composition} 
                        placeholder="e.g. 95%Poly 5%SP..." 
                        allowCustom={true} 
                      />
                    </div>

                    <div class="form-group">
                      <span class="form-label-title">Color Shade</span>
                      <Combobox 
                        options={colorOptions} 
                        bind:value={fp.color} 
                        placeholder="Color shade..." 
                        category="color"
                        allowCustom={true} 
                      />
                    </div>

                    <div class="form-group">
                      <span class="form-label-title">Invoice Dia (Inches)</span>
                      <Combobox 
                        options={standardDiaOptions} 
                        bind:value={fp.dia} 
                        placeholder="e.g. 60 in..." 
                        allowCustom={true} 
                      />
                    </div>

                    <div class="form-group">
                      <span class="form-label-title">Invoice GSM</span>
                      <Combobox 
                        options={standardGsmOptions} 
                        bind:value={fp.gsm} 
                        placeholder="e.g. 180..." 
                        allowCustom={true} 
                      />
                    </div>

                    <div class="form-group">
                      <span class="form-label-title">Invoice Rate (₹/Kg)</span>
                      <input type="number" step="0.01" class="form-control" bind:value={fp.rate} placeholder="₹ 0.00" />
                    </div>

                    <div class="form-group">
                      <span class="form-label-title">Rolls in Lot</span>
                      <input type="number" min="1" max="100" class="form-control font-bold" bind:value={fp.rollCount} />
                    </div>
                  </div>
                </div>
              {/each}
            </div>

            <div style="background:var(--bg-surface);padding:10px 14px;border-radius:var(--radius-sm);display:flex;justify-content:space-between;align-items:center;border:1px solid var(--border-color);">
              <span style="font-size:13px;font-weight:700;">
                Consignment Total: <span style="color:var(--primary-color);">{fabricTotalRolls} Rolls</span> across {fabricProducts.length} Fabric Components
              </span>
              <span style="font-size:12px;color:var(--text-muted);">
                Generates {gateParentCode}-1/{fabricTotalRolls} .. {fabricTotalRolls}/{fabricTotalRolls}
              </span>
            </div>
          </div>
        {:else}
          <!-- ==================================================== -->
          <!-- NON-FABRIC DYNAMIC SPECIFICATION (PER INWARD TYPE)   -->
          <!-- ==================================================== -->
          <div style="margin-top:20px;border-top:1px solid var(--border-color);padding-top:16px;">
            <div style="margin-bottom:12px;">
              <h4 style="margin:0;font-size:15px;color:var(--text-main);">Invoice Item Specification ({gateInwardType})</h4>
              <span style="font-size:12px;color:var(--text-muted);">
                Dynamic Step 1 Invoice attributes relative to <strong>{gateInwardType}</strong>. No GSM/Dia fields shown.
              </span>
            </div>

            <div class="form-grid">
              {#if gateKind === 'zips'}
                <div class="form-group">
                  <span class="form-label-title">Zip Type</span>
                  <Combobox options={zipTypeOptions} bind:value={gateZipType} allowCustom={false} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Teeth Size</span>
                  <Combobox options={zipTeethSizeOptions} bind:value={gateZipTeethSize} placeholder="5.0, 8.0, 10.0..." allowCustom={false} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Slider Spec</span>
                  <Combobox options={zipSliderOptions} bind:value={gateSlider} allowCustom={true} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Closure Type</span>
                  <Combobox options={zipClosureOptions} bind:value={gateClosure} allowCustom={false} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Teeth Finish / Color</span>
                  <Combobox options={teethColorOptions} bind:value={gateTeethColor} allowCustom={true} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Tape Color</span>
                  <Combobox options={colorOptions} bind:value={gateTapeColor} placeholder="Select tape color..." category="color" allowCustom={true} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">FPCC Zip Item (Zips Only)</span>
                  <Combobox options={fpccProductOptions} bind:value={gateFpccProduct} placeholder="Select Zip spec..." allowCustom={true} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Vendor Zip Name</span>
                  <input class="form-control" bind:value={gateVendorZipName} placeholder="Vendor zip code / model..." />
                </div>
              {:else if gateKind === 'packaging'}
                <div class="form-group">
                  <span class="form-label-title">Packaging Type</span>
                  <Combobox
                    options={packagingTypeOptions}
                    bind:value={gatePackagingType}
                    allowCustom={false}
                    onSelect={(item) => onPackagingTypeChange(item?.value || item?.display_name || gatePackagingType)}
                  />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Packaging Brand / Variant</span>
                  <Combobox options={packagingVariantOptionsFor(gatePackagingType)} bind:value={gatePackagingVariant} placeholder="Jeep, Ettu, Blank, DHL..." allowCustom={true} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Packaging Size / Dimension</span>
                  <Combobox options={packagingSizeOptionsFor(gatePackagingType)} bind:value={gatePackagingSize} placeholder="Select standard size..." allowCustom={false} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">FPCC Packaging Item ({fpccProductOptions.length})</span>
                  <Combobox options={fpccProductOptions} bind:value={gateFpccProduct} placeholder="Select FPCC packaging..." allowCustom={fpccProductOptions.length === 0} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Vendor Packaging Name</span>
                  <input class="form-control" bind:value={gateVendorPackagingName} placeholder="Vendor invoice item name..." />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Color Shade</span>
                  <Combobox options={colorOptions} bind:value={gateFpccColor} placeholder="Color..." category="color" allowCustom={true} />
                </div>
              {:else if gateKind === 'trims'}
                <div class="form-group">
                  <span class="form-label-title">Trim Type</span>
                  <Combobox
                    options={trimTypeOptions}
                    bind:value={gateTrimType}
                    allowCustom={false}
                    onSelect={(item) => onTrimTypeChange(item?.value || item?.display_name || gateTrimType)}
                  />
                </div>
                <div class="form-group">
                  <span class="form-label-title">FPCC {gateTrimType} Item ({fpccProductOptions.length})</span>
                  <Combobox options={fpccProductOptions} bind:value={gateFpccProduct} placeholder={fpccProductOptions.length > 0 ? `Select ${gateTrimType}...` : `Enter ${gateTrimType} spec...`} allowCustom={fpccProductOptions.length === 0} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">{gateTrimType === 'Thread' ? 'Thread Tex Size' : 'Trim Size / Width'}</span>
                  <Combobox options={trimSizesForType(gateTrimType)} bind:value={gateTrimSize} placeholder="Select size..." allowCustom={false} />
                </div>
                {#if gateTrimType === 'Cord'}
                  <div class="form-group">
                    <span class="form-label-title">Cord Ends</span>
                    <Combobox options={cleanDisplayNames(masters.trim_cord_ends || [])} bind:value={gateTrimCordEnd} allowCustom={false} />
                  </div>
                {:else if gateTrimType === 'Fusing'}
                  <div class="form-group">
                    <span class="form-label-title">Fusing GSM</span>
                    <Combobox options={cleanDisplayNames(masters.trim_fusing_gsms || [])} bind:value={gateTrimFusingGsm} allowCustom={false} />
                  </div>
                {:else if gateTrimType === 'Thread'}
                  <div class="form-group">
                    <span class="form-label-title">Thread Length</span>
                    <Combobox options={cleanDisplayNames(masters.trim_thread_lengths || [])} bind:value={gateTrimThreadLength} allowCustom={false} />
                  </div>
                {/if}
                <div class="form-group">
                  <span class="form-label-title">Trim Color</span>
                  <Combobox options={trimColorsForType(gateTrimType)} bind:value={gateFpccColor} placeholder="Color..." category="color" allowCustom={true} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Vendor Trim Name</span>
                  <input class="form-control" bind:value={gateVendorTrimName} placeholder="What vendor bills it as..." />
                </div>
              {:else if gateKind === 'label'}
                <div class="form-group">
                  <span class="form-label-title">Label Category</span>
                  <select class="form-control" bind:value={gateLabelCategory} onchange={onLabelCategoryChange}>
                    <option value="Brand Label">Brand Label</option>
                    <option value="Additional Label">Additional Label</option>
                  </select>
                </div>
                <div class="form-group">
                  <span class="form-label-title">Label Genre (Brand &amp; Additional)</span>
                  <Combobox options={labelGenreOptions} bind:value={gateLabelGenre} allowCustom={false} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">{gateLabelCategory === 'Additional Label' ? 'Additional Label Type' : 'Brand Label'}</span>
                  <Combobox options={fpccProductOptions} bind:value={gateLabelType} allowCustom={false} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Label Description</span>
                  <Combobox options={cleanDisplayNames(masters.label_descriptions || [])} bind:value={gateLabelDescription} allowCustom={true} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Base Color</span>
                  <Combobox options={colorOptions} bind:value={gateLabelBaseColor} category="color" allowCustom={true} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Text / Print Color</span>
                  <Combobox options={colorOptions} bind:value={gateLabelTextColor} category="color" allowCustom={true} />
                </div>
              {:else if gateKind === 'apparel'}
                <div class="form-group">
                  <span class="form-label-title">Apparel Genre</span>
                  <Combobox
                    options={finishedGenreOptions}
                    bind:value={gateFgType}
                    allowCustom={false}
                    onSelect={(item) => onApparelGenreChange(item?.value || item?.display_name || gateFgType)}
                  />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Target Gender (Constrains Sizes)</span>
                  <Combobox
                    options={genderOptions}
                    bind:value={gateGender}
                    allowCustom={false}
                    onSelect={(item) => onGenderChange(item?.value || item?.display_name || gateGender)}
                  />
                </div>
                <div class="form-group">
                  <span class="form-label-title">FPCC Apparel Product ({fpccProductOptions.length} {gateFgType} SKUs)</span>
                  <Combobox
                    options={fpccProductOptions}
                    bind:value={gateFpccProduct}
                    placeholder={fpccProductOptions.length > 0 ? `Select ${gateFgType} product...` : `No seeded ${gateFgType} SKUs — enter SKU...`}
                    allowCustom={fpccProductOptions.length === 0}
                  />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Vendor Item Name</span>
                  <input class="form-control" bind:value={gateVendorFgName} placeholder="Vendor billed item name..." />
                </div>
                <div class="form-group">
                  <span class="form-label-title">FPCC Color</span>
                  <Combobox options={colorOptions} bind:value={gateFpccColor} category="color" allowCustom={true} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Vendor Color</span>
                  <Combobox options={colorOptions} bind:value={gateVendorColor} category="color" allowCustom={true} />
                </div>
              {:else}
                {@const lockedGenre = gateKind === 'non_apparel' ? normalizeNonApparelGenre(gateInwardType) : ''}
                <div class="form-group">
                  <span class="form-label-title">{lockedGenre ? 'Product Category (Locked by Inward Type)' : 'Non-Apparel Genre'}</span>
                  {#if lockedGenre}
                    <input class="form-control font-bold" value={lockedGenre} readonly />
                  {:else}
                    <Combobox
                      options={finishedGenreOptions}
                      bind:value={gateFgType}
                      allowCustom={false}
                      onSelect={(item) => {
                        gateFgType = item?.value || item?.display_name || gateFgType;
                        gateFpccProduct = '';
                      }}
                    />
                  {/if}
                </div>
                <div class="form-group">
                  <span class="form-label-title">FPCC {lockedGenre || gateFgType || gateInwardType} Product ({fpccProductOptions.length} SKUs)</span>
                  <Combobox
                    options={fpccProductOptions}
                    bind:value={gateFpccProduct}
                    placeholder={fpccProductOptions.length > 0 ? `Select ${lockedGenre || gateFgType} product...` : `No seeded ${lockedGenre || gateFgType} SKUs — enter custom name...`}
                    allowCustom={fpccProductOptions.length === 0}
                  />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Vendor Item Name</span>
                  <input class="form-control" bind:value={gateVendorFgName} placeholder="Vendor billed item name..." />
                </div>
                <div class="form-group">
                  <span class="form-label-title">FPCC Color</span>
                  <Combobox options={colorOptions} bind:value={gateFpccColor} category="color" allowCustom={true} />
                </div>
                <div class="form-group">
                  <span class="form-label-title">Vendor Color</span>
                  <Combobox options={colorOptions} bind:value={gateVendorColor} category="color" allowCustom={true} />
                </div>
              {/if}

              <div class="form-group">
                <label for="nonfab-rate">Invoice Unit Rate (₹)</label>
                <input id="nonfab-rate" type="number" step="0.01" class="form-control" bind:value={gateTrimRate} placeholder="₹ / {gateUom}" />
              </div>

              <div class="form-group">
                <label for="nonfab-bundles">Number of Bundles / Cartons</label>
                <input id="nonfab-bundles" type="number" min="1" max="100" class="form-control font-bold" bind:value={gateNoOfBundles} required />
              </div>
            </div>
          </div>
        {/if}

        <div class="form-group" style="margin-top: 16px;">
          <label for="gate-comm">Invoice Gate Remarks / Delivery Vehicle / Staging Notes</label>
          <input id="gate-comm" class="form-control" bind:value={gateComment} placeholder="e.g. Delivery vehicle MH-12-8802, dock bay 3, invoice challan stamped..." />
        </div>
      </section>

      <!-- 3. Pre-Generated Bundle Stubs Preview Table -->
      <section class="card section-card">
        <div class="section-header">
          <div>
            <h3>3. Per-Bundle Invoice Breakdown ({gateBundles.length} Bundles)</h3>
            <span class="subtitle">Sizes &amp; Invoice Quantities live per bundle (auto-distributed evenly from lot qty, editable row-by-row).</span>
          </div>
          <div style="display:flex;gap:8px;align-items:center;">
            <button type="button" class="btn btn-secondary btn-sm" onclick={redistributeGateBundlesEvenly}>➗ Reset Even Split</button>
            <span class="badge badge-secondary">{qrCopiesPerBundle} QR Copies / Bundle</span>
          </div>
        </div>

        <div class="table-container">
          <table class="data-table">
            <thead>
              <tr>
                <th>Bundle Code</th>
                <th>Item / Part</th>
                <th>Color</th>
                <th>{gateKind === 'fabric' ? 'Fabric Type / Dia / GSM' : gateKind === 'zips' ? 'Teeth Size' : gateKind === 'apparel' ? `Size (${gateGender})` : 'Size / Spec'}</th>
                <th>Invoice Qty ({gateUom})</th>
                <th>Invoice Rate (₹)</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {#each gateBundles as b, bIdx}
                <tr>
                  <td class="code-pill">{b.bundleCode}</td>
                  <td><strong>{b.itemName || gateFpccProduct || gateInwardType}</strong></td>
                  <td>{b.colorName || 'Standard'}</td>
                  <td>
                    {#if gateKind === 'fabric'}
                      <span class="code-pill" style="font-size:11px;">{b.sizeSpec}</span>
                    {:else if bundleSizeSuggestions.length > 0}
                      <div style="min-width:150px;">
                        <Combobox options={bundleSizeSuggestions} bind:value={gateBundles[bIdx].sizeSpec} placeholder="Select size..." allowCustom={false} />
                      </div>
                    {:else}
                      <input 
                        class="form-control" 
                        style="font-size:12px;padding:3px 8px;width:170px;" 
                        bind:value={gateBundles[bIdx].sizeSpec} 
                        placeholder="Size spec..." 
                      />
                    {/if}
                  </td>
                  <td>
                    <input 
                      type="number" 
                      step="0.01" 
                      class="form-control font-bold" 
                      style="font-size:12px;padding:3px 8px;width:110px;" 
                      bind:value={gateBundles[bIdx].invoicedQty} 
                      oninput={onGateBundleQtyChange}
                      required
                    />
                  </td>
                  <td>
                    <input 
                      type="number" 
                      step="0.01" 
                      class="form-control" 
                      style="font-size:12px;padding:3px 8px;width:100px;" 
                      bind:value={gateBundles[bIdx].rate} 
                    />
                  </td>
                  <td>
                    <span class="badge badge-secondary">Pending (Invoice Dock)</span>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>

        <div class="submit-bar">
          <button type="submit" class="btn btn-primary btn-lg" disabled={isSubmitting}>
            {#if isSubmitting}
              ⏳ Booking Invoice Entry...
            {:else}
              🚚 Commit Invoice Entry &amp; Generate {gateBundles.length} Bundle Stubs
            {/if}
          </button>
        </div>
      </section>
    </form>
  {/if}

  <!-- ========================================================= -->
  <!-- STATION 1B: FLOOR COUNT & RECONCILIATION (TABLET VIEW)    -->
  <!-- ========================================================= -->
  {#if activeStation === '1B'}
    <div class="station-container tablet-container">
      <section class="card section-card">
        <div class="section-header">
          <div>
            <h3>Select Lot for Physical Inward Count (Step 2)</h3>
            <span class="subtitle">Tablet Optimized • Measure inward qty, assign staging bins &amp; flag discrepancies</span>
          </div>
          <span class="badge badge-warning">{awaitingCountLots.length} Awaiting Inward</span>
        </div>

        {#if awaitingCountLots.length === 0}
          <div class="empty-state">
            <p>🎉 No lots currently waiting for inward count! All invoice arrivals have been reconciled.</p>
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
                  <div style="font-weight:700;color:var(--text-main);margin-bottom:2px;">
                    ✨ {lot.fpccProduct || lot.vendorProduct}
                  </div>
                  <strong>{lot.vendorName}</strong> • {lot.inwardType}<br/>
                  Invoice: <strong>{lot.invoiceQty} {lot.uom}</strong> (Inv: #{lot.invoiceNo})
                </div>
              </button>
            {/each}
          </div>
        {/if}
      </section>

      {#if floorLotDetail}
        <form onsubmit={handleFloorSubmit} class="station-form">
          <section class="card section-card tablet-reconcil-card">
            <div class="section-header">
              <div>
                <!-- Prominently Surfaced FPCC Product Name -->
                <div class="fpcc-hero-badge" style="margin-bottom:6px;background:rgba(37,99,235,0.12);padding:6px 12px;border-radius:var(--radius-sm);display:inline-flex;align-items:center;gap:8px;border:1px solid rgba(37,99,235,0.25);">
                  <span style="font-size:11px;font-weight:800;letter-spacing:0.05em;color:var(--primary-color);">FPCC PRODUCT:</span>
                  <strong style="font-size:16px;color:var(--primary-color);">{floorLotDetail.lot.fpccProduct || floorLotDetail.lot.vendorProduct || 'Standard Inward'}</strong>
                </div>
                <h3>Reconciliation: {floorLotDetail.lot.parentCode}</h3>
                <span class="subtitle">Vendor Item: <strong>{floorLotDetail.lot.vendorProduct || '—'}</strong> | Supplier: <strong>{floorLotDetail.lot.vendorName}</strong> | Invoice: #{floorLotDetail.lot.invoiceNo}</span>
              </div>
              <div class="chip-badges">
                <span class="badge badge-primary">{floorLotDetail.lot.inwardType}</span>
                {#if isFloorFabric && floorLotDetail.lot.attrs?.invoiceGsm}
                  <span class="badge badge-secondary">Invoice GSM: {floorLotDetail.lot.attrs.invoiceGsm}</span>
                {/if}
                {#if isFloorFabric && floorLotDetail.lot.attrs?.dia}
                  <span class="badge badge-secondary">Invoice Dia: {floorLotDetail.lot.attrs.dia}"</span>
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

            <!-- Dynamic Step 2 Inward Attributes Summary / Editable Specs (Always Shown) -->
            <div style="margin-top:12px;padding:12px 14px;background:var(--bg-tertiary);border:1px solid var(--border-color);border-radius:var(--radius-sm);font-size:12px;">
              <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:8px;flex-wrap:wrap;gap:8px;">
                <span style="font-weight:700;color:var(--primary-color);">Step 2 Inward Specs ({floorLotDetail.lot.inwardType})</span>
                <span style="color:var(--text-muted);font-size:11px;">
                  Product: <strong>{floorLotDetail.lot.fpccProduct || floorLotDetail.lot.vendorProduct || '—'}</strong>
                </span>
              </div>

              {#if floorKind === 'fabric'}
                <div style="display:flex;flex-wrap:wrap;gap:14px;margin-bottom:10px;align-items:center;">
                  <span>Invoice GSM: <strong>{floorAttrs.invoiceGsm || '—'}</strong></span>
                  <span>Invoice Dia: <strong>{floorAttrs.dia ? `${String(floorAttrs.dia).replace(/"/g, '')}"` : '—'}</strong></span>
                  <span>Fabric Type: <strong>{floorAttrs.fabricType || '—'}</strong></span>
                  <span>Composition: <strong>{floorAttrs.composition || '—'}</strong></span>
                  {#if floorLotDetail.lot.fabricProducts && floorLotDetail.lot.fabricProducts.length > 1}
                    <span>Components: <strong>{floorLotDetail.lot.fabricProducts.map(f => `${f.fabricType || f.fpccProduct} (${f.gsm} GSM / ${f.dia})`).join(' · ')}</strong></span>
                  {/if}
                </div>
                <div class="form-grid" style="gap:10px;">
                  <div class="form-group" style="margin:0;">
                    <span class="form-label-title">Verified Fabric Type</span>
                    <Combobox options={fabricTypeOptions} bind:value={floorAttrs.fabricType} placeholder="Knitted, Pique, Single Jersey..." allowCustom={true} />
                  </div>
                  <div class="form-group" style="margin:0;">
                    <span class="form-label-title">Composition</span>
                    <Combobox options={compositionOptions} bind:value={floorAttrs.composition} placeholder="e.g. 100%Cot..." allowCustom={true} />
                  </div>
                  <div class="form-group" style="margin:0;">
                    <span class="form-label-title">Invoice / Master GSM</span>
                    <Combobox options={standardGsmOptions} bind:value={floorAttrs.invoiceGsm} placeholder="e.g. 220" allowCustom={true} />
                  </div>
                  <div class="form-group" style="margin:0;">
                    <span class="form-label-title">Invoice / Master Dia</span>
                    <Combobox options={standardDiaOptions} bind:value={floorAttrs.dia} placeholder="e.g. 32&quot;" allowCustom={true} />
                  </div>
                </div>
              {:else if floorKind === 'zips'}
                <div class="form-grid" style="gap:10px;">
                  <div class="form-group" style="margin:0;">
                    <span class="form-label-title">Zip Type</span>
                    <Combobox options={zipTypeOptions} bind:value={floorAttrs.zipType} allowCustom={false} />
                  </div>
                  <div class="form-group" style="margin:0;">
                    <span class="form-label-title">Teeth Size</span>
                    <Combobox options={zipTeethSizeOptions} bind:value={floorAttrs.teethSize} allowCustom={false} />
                  </div>
                  <div class="form-group" style="margin:0;">
                    <span class="form-label-title">Slider</span>
                    <Combobox options={zipSliderOptions} bind:value={floorAttrs.slider} allowCustom={true} />
                  </div>
                  <div class="form-group" style="margin:0;">
                    <span class="form-label-title">Closure</span>
                    <Combobox options={zipClosureOptions} bind:value={floorAttrs.closure} allowCustom={false} />
                  </div>
                  <div class="form-group" style="margin:0;">
                    <span class="form-label-title">Teeth Finish</span>
                    <Combobox options={teethColorOptions} bind:value={floorAttrs.teethColor} allowCustom={true} />
                  </div>
                  <div class="form-group" style="margin:0;">
                    <span class="form-label-title">Tape Color</span>
                    <Combobox options={colorOptions} bind:value={floorAttrs.tapeColor} category="color" allowCustom={true} />
                  </div>
                </div>
              {:else if floorKind === 'trims'}
                <div class="form-grid" style="gap:10px;">
                  <div class="form-group" style="margin:0;">
                    <span class="form-label-title">Trim Type</span>
                    <Combobox options={trimTypeOptions} bind:value={floorAttrs.trimType} allowCustom={false} />
                  </div>
                  <div class="form-group" style="margin:0;">
                    <span class="form-label-title">Trim Size / Width</span>
                    <Combobox options={trimSizesForType(floorAttrs.trimType)} bind:value={floorAttrs.trimSize} allowCustom={false} />
                  </div>
                  {#if floorAttrs.cordEnd}<span>Cord Ends: <strong>{floorAttrs.cordEnd}</strong></span>{/if}
                  {#if floorAttrs.fusingGsm}<span>Fusing GSM: <strong>{floorAttrs.fusingGsm}</strong></span>{/if}
                  {#if floorAttrs.threadLength}<span>Length: <strong>{floorAttrs.threadLength}</strong></span>{/if}
                </div>
              {:else if floorKind === 'packaging'}
                <div class="form-grid" style="gap:10px;">
                  <div class="form-group" style="margin:0;">
                    <span class="form-label-title">Packaging Type</span>
                    <Combobox options={packagingTypeOptions} bind:value={floorAttrs.packagingType} allowCustom={false} />
                  </div>
                  <div class="form-group" style="margin:0;">
                    <span class="form-label-title">Variant</span>
                    <Combobox options={packagingVariantOptionsFor(floorAttrs.packagingType)} bind:value={floorAttrs.packagingVariant} allowCustom={true} />
                  </div>
                  <div class="form-group" style="margin:0;">
                    <span class="form-label-title">Size / Dimension</span>
                    <Combobox options={packagingSizeOptionsFor(floorAttrs.packagingType)} bind:value={floorAttrs.packagingSize} allowCustom={false} />
                  </div>
                </div>
              {:else if floorKind === 'label'}
                <div style="display:flex;flex-wrap:wrap;gap:14px;">
                  <span>Category: <strong>{floorAttrs.labelCategory || 'Brand Label'}</strong></span>
                  <span>Genre: <strong>{floorAttrs.labelGenre || '—'}</strong></span>
                  <span>Description: <strong>{floorAttrs.labelDescription || '—'}</strong></span>
                  <span>Colors: <strong>{floorAttrs.baseColor || '—'} / {floorAttrs.textColor || '—'}</strong></span>
                </div>
              {:else if floorKind === 'apparel'}
                <div class="form-grid" style="gap:10px;">
                  <div class="form-group" style="margin:0;">
                    <span class="form-label-title">Apparel Genre</span>
                    <Combobox options={getGenreOptionsForInward(masters, 'Apparel')} bind:value={floorAttrs.finishedGoodType} allowCustom={false} />
                  </div>
                  <div class="form-group" style="margin:0;">
                    <span class="form-label-title">Target Gender (Bundle Size Filter)</span>
                    <Combobox options={genderOptions} bind:value={floorAttrs.gender} allowCustom={false} />
                  </div>
                </div>
              {:else}
                <div style="display:flex;flex-wrap:wrap;gap:14px;">
                  <span>Product Category: <strong>{normalizeNonApparelGenre(floorLotDetail.lot.inwardType) || floorAttrs.finishedGoodType || floorLotDetail.lot.inwardType}</strong></span>
                  <span>Allowed Sizes: <strong>{floorBundleSizeSuggestions.join(', ') || 'Standard'}</strong></span>
                </div>
              {/if}
            </div>

            <!-- Step 1 Invoice vs Step 2 Inward Touch Inputs -->
            <div class="form-grid" style="margin-top: 16px;">
              <div class="form-group">
                <label for="floor-inv-q">Step 1: Invoice Qty ({floorLotDetail.lot.uom})</label>
                <input id="floor-inv-q" type="number" step="0.01" class="form-control touch-input font-bold" bind:value={floorInvoicedQty} required />
                <span class="help-text">From Invoice #{floorLotDetail.lot.invoiceNo}</span>
              </div>

              <div class="form-group">
                <label for="floor-act-q">Step 2: Inward Qty ({floorLotDetail.lot.uom})</label>
                <div class="input-with-action">
                  <input id="floor-act-q" type="number" step="0.01" class="form-control touch-input font-bold" bind:value={floorCountedQty} required />
                  <button type="button" class="btn btn-secondary btn-sm touch-btn" onclick={distributeFloorCountEvenly} title="Split evenly across bundles">
                    ➗ Split
                  </button>
                  <button type="button" class="btn btn-secondary btn-sm touch-btn" onclick={autoSumFloorBundles} title="Sum from bundle table">
                    ∑ Sum
                  </button>
                </div>
                <span class="help-text">Actual weighed / counted inward quantity</span>
              </div>

              <!-- Warehouse Staging Bin via Combobox + Quick Tap Bins -->
              <div class="form-group">
                <span class="form-label-title">Default Warehouse Staging Bin</span>
                <div style="display:flex;gap:6px;flex-wrap:wrap;margin-bottom:6px;">
                  {#each quickBins as qb}
                    <button 
                      type="button" 
                      class="btn btn-sm {floorLocationBin === qb ? 'btn-primary' : 'btn-secondary'}" 
                      style="font-size:11px;padding:3px 8px;"
                      onclick={() => floorLocationBin = qb}
                    >
                      {qb}
                    </button>
                  {/each}
                </div>
                <Combobox 
                  options={locationOptions} 
                  bind:value={floorLocationBin} 
                  placeholder="Select warehouse bin e.g. A/R01/S01..." 
                  category="location"
                  allowCustom={false} 
                />
              </div>

              <div class="form-group">
                <label for="floor-by">Inward Counted By (Storekeeper)</label>
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

          <!-- Per-Bundle Physical Breakdown with Dynamic Staging Bin Dropdown & Sizes/Qtys -->
          <section class="card section-card">
            <div class="section-header">
              <div>
                <h3>Per-Bundle Inward Measurement &amp; Bin Allocation ({floorBundles.length} Bundles)</h3>
                <span class="subtitle">Step 1 Invoice Qty vs Step 2 Inward Qty, Size / Spec, Inward Rate, and Staging Bin dropdown per bundle</span>
              </div>
              <div class="btn-group">
                <button type="button" class="btn btn-secondary btn-sm touch-btn" onclick={distributeFloorCountEvenly}>
                  ➗ Distribute Evenly
                </button>
                <button type="button" class="btn btn-secondary btn-sm touch-btn" onclick={autoSumFloorBundles}>
                  ∑ Auto-Calculate Total
                </button>
              </div>
            </div>

            <div class="table-container">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>Roll / Bundle Code</th>
                    <th>Product / Part</th>
                    <th>Color</th>
                    <th>{floorKind === 'zips' ? 'Teeth Size / Spec' : 'Size / Spec'}</th>
                    <th>Invoice Qty</th>
                    <th>Inward Qty ({floorLotDetail.lot.uom})</th>
                    <th>Inward Rate (₹)</th>
                    <th>Staging Bin</th>
                    {#if isFloorFabric}
                      <th>Inward GSM</th>
                      <th>Inward Dia</th>
                    {/if}
                    <th>Inward Remarks</th>
                  </tr>
                </thead>
                <tbody>
                  {#each floorBundles as b, bIdx}
                    <tr>
                      <td class="code-pill">{b.bundleCode}</td>
                      <td><strong>{b.itemName || floorLotDetail.lot.fpccProduct}</strong></td>
                      <td>{b.colorName || '—'}</td>
                      <td>
                        {#if !isFloorFabric && floorBundleSizeSuggestions.length > 0}
                          <div style="min-width:140px;">
                            <Combobox 
                              options={floorBundleSizeSuggestions} 
                              bind:value={floorBundles[bIdx].sizeSpec} 
                              placeholder="Select allowed size..." 
                              allowCustom={false} 
                            />
                          </div>
                        {:else}
                          <input 
                            class="form-control" 
                            style="width: 140px; font-size:12px;" 
                            bind:value={floorBundles[bIdx].sizeSpec} 
                            placeholder="Size spec..." 
                          />
                        {/if}
                      </td>
                      <td>
                        <span class="code-pill" style="opacity:0.85;">{b.invoicedQty}</span>
                      </td>
                      <td>
                        <input 
                          type="number" 
                          step="0.01" 
                          class="form-control touch-input font-bold" 
                          style="width: 110px;" 
                          bind:value={floorBundles[bIdx].countedQty} 
                          required 
                        />
                      </td>
                      <td>
                        <input 
                          type="number" 
                          step="0.01" 
                          class="form-control" 
                          style="width: 90px; font-size:12px;" 
                          bind:value={floorBundles[bIdx].rate} 
                          placeholder="Rate" 
                        />
                      </td>
                      <td style="min-width:170px;">
                        <Combobox 
                          options={locationOptions} 
                          bind:value={floorBundles[bIdx].locationBin} 
                          placeholder="Select Staging Bin..." 
                          category="location"
                          allowCustom={false} 
                        />
                      </td>
                      {#if isFloorFabric}
                        <td>
                          <input 
                            class="form-control" 
                            style="width: 85px; font-size:12px;" 
                            bind:value={floorBundles[bIdx].measuredGsm} 
                            placeholder="Inw GSM" 
                          />
                        </td>
                        <td>
                          <input 
                            class="form-control" 
                            style="width: 85px; font-size:12px;" 
                            bind:value={floorBundles[bIdx].measuredDia} 
                            placeholder="Inw Dia" 
                          />
                        </td>
                      {/if}
                      <td>
                        <input 
                          class="form-control" 
                          style="font-size:12px;" 
                          bind:value={floorBundles[bIdx].comment} 
                          placeholder="Inward remarks..." 
                        />
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>

            <div class="submit-bar">
              <button type="submit" class="btn btn-primary btn-lg touch-btn-lg" disabled={isSubmitting}>
                {#if isSubmitting}
                  ⏳ Saving Inward Measurements...
                {:else}
                  ⚖️ Confirm Inward Measurements &amp; Advance to QC
                {/if}
              </button>
            </div>
          </section>
        </form>
      {/if}
    </div>
  {/if}

  <!-- ========================================================= -->
  <!-- STATION 1C: QC INSPECTION (PHONE VIEW WITH AUDIO & STT)   -->
  <!-- ========================================================= -->
  {#if activeStation === '1C'}
    <div class="station-container phone-container">
      <section class="card section-card">
        <div class="section-header">
          <div>
            <h3>Select Lot for QC Inspection</h3>
            <span class="subtitle">Mobile Phone Optimized • Binary Pass/Fail, Audio Notes & Tags</span>
          </div>
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
                  <span class="badge badge-primary">{lot.stage}</span>
                </div>
                <div class="chip-meta">
                  <div style="font-weight:700;color:var(--text-main);margin-bottom:2px;">
                    ✨ {lot.fpccProduct || lot.vendorProduct}
                  </div>
                  <strong>{lot.vendorName}</strong> • {lot.countedQty || lot.invoiceQty} {lot.uom}<br/>
                  QC Progress: {lot.inspectedBundles || 0}/{lot.noOfBundles || 1} inspected
                </div>
              </button>
            {/each}
          </div>
        {/if}
      </section>

      {#if qcLotDetail}
        <form onsubmit={handleQCSubmit} class="station-form">
          <section class="card section-card">
            <div class="section-header">
              <div>
                <h3>QC Inspection: {qcLotDetail.lot.parentCode}</h3>
                <span class="subtitle">✨ <strong>{qcLotDetail.lot.fpccProduct || qcLotDetail.lot.vendorProduct}</strong> | {qcLotDetail.lot.vendorName}</span>
              </div>
              <div class="btn-group">
                <button type="button" class="btn btn-secondary btn-sm" onclick={markAllQCPass}>
                  ✓ Quick Mark All PASS
                </button>
                <button type="button" class="btn btn-primary btn-sm" onclick={openAllZebraTags}>
                  🖨️ Thermal Tags ({qrCopiesPerBundle}x)
                </button>
              </div>
            </div>

            <div class="form-grid" style="margin-top: 14px;">
              <div class="form-group">
                <label for="qc-insp">QC Inspector Name</label>
                <input id="qc-insp" class="form-control" bind:value={qcInspector} required />
              </div>
              <div class="form-group">
                <span class="form-label-title">Lot Inspection Progress</span>
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

          <!-- Per-Roll Binary QC Cards (Phone-Optimized Vertical Cards) -->
          <section class="card section-card">
            <div class="section-header">
              <h3>Roll Inspection Cards ({qcBundles.length} Bundles)</h3>
              <span class="badge badge-secondary">Binary Touch + Audio Voice Notes</span>
            </div>

            <div class="qc-cards-list">
              {#each qcBundles as b, idx}
                <div class="qc-bundle-card phone-bundle-card {b.qcStatus === 'Rejected' ? 'rejected' : 'passed'}">
                  <div class="qc-bundle-header">
                    <div class="bundle-identity">
                      <span class="code-pill font-bold" style="font-size:15px;">{b.bundleCode}</span>
                      <div class="bundle-specs-tag">
                        <strong>{b.itemName || qcLotDetail.lot.fpccProduct}</strong> • {b.inwardQty} {qcLotDetail.lot.uom} • Bin: <strong>{b.locationBin}</strong>
                        {#if b.sizeSpec}<br/><span style="font-size:11px;color:var(--text-muted);">{b.sizeSpec}</span>{/if}
                      </div>
                    </div>

                    <!-- GIANT BINARY PASS/FAIL TOUCH BUTTONS FOR MOBILE -->
                    <div class="binary-qc-toggle giant-toggle">
                      <button 
                        type="button" 
                        class="qc-toggle-btn giant-btn pass {b.qcStatus === 'Passed' ? 'active' : ''}"
                        onclick={() => b.qcStatus = 'Passed'}
                      >
                        ✓ PASS
                      </button>
                      <button 
                        type="button" 
                        class="qc-toggle-btn giant-btn fail {b.qcStatus === 'Rejected' ? 'active' : ''}"
                        onclick={() => b.qcStatus = 'Rejected'}
                      >
                        ✕ REJECT
                      </button>
                    </div>
                  </div>

                  <div class="qc-bundle-body">
                    <!-- Photo Evidence Section -->
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
                            capture="environment"
                            style="display:none;" 
                            onchange={(e) => { const f = e.currentTarget.files?.[0]; if (f) handleQCImageUpload(idx, f); }} 
                          />
                          {#if b.isUploading}
                            <span>⏳ Uploading...</span>
                          {:else}
                            <span>📷 Camera Photo</span>
                          {/if}
                        </label>
                      {/if}
                    </div>

                    <!-- Audio Voice Note & Speech-to-Text Section -->
                    <div class="qc-audio-section">
                      {#if activeRecordingBundleIdx === idx}
                        <div class="recording-indicator pulsing">
                          <span class="record-dot">⏺</span>
                          <span class="record-text">Listening... ({recordingTimerSeconds}s)</span>
                          <button 
                            type="button" 
                            class="btn btn-danger btn-sm" 
                            onclick={() => stopAudioRecording(idx)}
                          >
                            ⏹ Stop & Save Voice Note
                          </button>
                        </div>
                      {:else}
                        <div style="display:flex;align-items:center;gap:8px;">
                          <button 
                            type="button" 
                            class="btn btn-secondary btn-sm mic-btn" 
                            onclick={() => startAudioRecording(idx)}
                          >
                            🎙️ Record Voice Note
                          </button>
                          {#if b.qcAudioBlobUrl || b.qcAudioUrl}
                            <audio 
                              src={b.qcAudioBlobUrl || b.qcAudioUrl} 
                              controls 
                              style="height:32px;max-width:180px;"
                            ></audio>
                          {/if}
                        </div>
                      {/if}
                    </div>

                    <!-- Defect Remarks with Live Speech Transcription -->
                    <div class="qc-remarks-section">
                      <Combobox 
                        options={qcRemarkOptions} 
                        bind:value={b.qcRemarks} 
                        placeholder="Type remarks or speak via microphone..." 
                        allowCustom={true} 
                      />
                    </div>

                    <div class="qc-tag-action">
                      <button type="button" class="btn btn-secondary btn-sm" onclick={() => openSingleZebraTag(b)}>
                        🏷️ Tag ({qrCopiesPerBundle}x)
                      </button>
                    </div>
                  </div>
                </div>
              {/each}
            </div>

            <div class="submit-bar sticky-mobile-submit">
              <button type="submit" class="btn btn-primary btn-lg giant-submit-btn" disabled={isSubmitting}>
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
    padding-bottom: 80px;
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
    background: rgba(37, 99, 235, 0.05);
  }

  .station-tab-btn.active {
    border-color: var(--primary-color);
    background: rgba(37, 99, 235, 0.09);
    box-shadow: 0 0 0 1px var(--primary-color);
  }

  .tab-icon {
    font-size: 26px;
  }

  .tab-title {
    font-size: 14px;
    font-weight: 700;
    color: var(--text-primary);
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .device-pill {
    font-size: 10px;
    font-weight: 800;
    text-transform: uppercase;
    padding: 2px 6px;
    border-radius: 4px;
    letter-spacing: 0.05em;
  }

  .device-pill.desktop {
    background: rgba(99, 102, 241, 0.15);
    color: #6366f1;
  }

  .device-pill.tablet {
    background: rgba(245, 158, 11, 0.15);
    color: #f59e0b;
  }

  .device-pill.phone {
    background: rgba(16, 185, 129, 0.15);
    color: #10b981;
  }

  .tab-desc {
    font-size: 12px;
    color: var(--text-secondary);
    margin-top: 2px;
  }

  .section-card {
    padding: 20px;
  }

  .invoice-header-card {
    border-left: 4px solid var(--primary-color);
  }

  .invoice-suggest-dropdown {
    margin-top: 8px;
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    box-shadow: 0 8px 24px rgba(0,0,0,0.15);
    max-height: 200px;
    overflow-y: auto;
  }

  .invoice-suggest-item {
    display: block;
    width: 100%;
    padding: 10px 14px;
    text-align: left;
    background: none;
    border: none;
    border-bottom: 1px solid var(--border-color);
    cursor: pointer;
  }

  .invoice-suggest-item:hover {
    background: rgba(37, 99, 235, 0.08);
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
    color: var(--text-main);
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

  .form-group label, .form-label-title {
    font-size: 11.5px;
    font-weight: 700;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .input-with-action {
    display: flex;
    gap: 6px;
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
    color: var(--text-primary);
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
    vertical-align: middle;
  }

  .data-table th {
    background: var(--bg-subtle);
    font-weight: 700;
    color: var(--text-secondary);
    text-transform: uppercase;
    font-size: 11px;
  }

  .submit-bar {
    margin-top: 24px;
    display: flex;
    justify-content: flex-end;
  }

  /* Tablet-specific Ergonomics */
  .touch-input {
    min-height: 44px;
    font-size: 15px;
  }

  .touch-btn {
    min-height: 40px;
    padding: 0 16px;
    font-size: 13px;
    font-weight: 700;
  }

  .touch-btn-lg {
    min-height: 52px;
    font-size: 16px;
    font-weight: 800;
    padding: 0 28px;
  }

  /* Mobile Phone QC Ergonomics */
  .phone-container {
    max-width: 640px;
    margin: 0 auto;
    width: 100%;
  }

  .phone-bundle-card {
    border-radius: var(--radius-md);
    padding: 16px;
  }

  .giant-toggle {
    display: flex;
    gap: 10px;
    border: none;
    width: 100%;
    margin-top: 10px;
  }

  .giant-btn {
    flex: 1;
    min-height: 56px;
    font-size: 16px;
    font-weight: 900;
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    justify-content: center;
    border: 2px solid transparent;
  }

  .giant-btn.pass {
    background: rgba(16, 185, 129, 0.15);
    color: #10b981;
    border-color: #10b981;
  }

  .giant-btn.pass.active {
    background: #10b981;
    color: #fff;
    box-shadow: 0 4px 14px rgba(16, 185, 129, 0.4);
  }

  .giant-btn.fail {
    background: rgba(239, 68, 68, 0.15);
    color: #ef4444;
    border-color: #ef4444;
  }

  .giant-btn.fail.active {
    background: #ef4444;
    color: #fff;
    box-shadow: 0 4px 14px rgba(239, 68, 68, 0.4);
  }

  .recording-indicator {
    display: flex;
    align-items: center;
    gap: 10px;
    background: rgba(239, 68, 68, 0.12);
    border: 1px solid rgba(239, 68, 68, 0.3);
    padding: 8px 12px;
    border-radius: var(--radius-sm);
  }

  .record-dot {
    color: #ef4444;
    font-size: 16px;
    animation: pulse-dot 1s infinite;
  }

  @keyframes pulse-dot {
    0% { opacity: 1; }
    50% { opacity: 0.2; }
    100% { opacity: 1; }
  }

  .record-text {
    font-size: 12px;
    font-weight: 700;
    color: #ef4444;
    flex: 1;
  }

  .mic-btn {
    font-weight: 700;
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  /* Discrepancy Sensor */
  .discrepancy-banner {
    border-radius: var(--radius-md);
    padding: 16px 20px;
    border-left: 6px solid;
  }

  .discrepancy-banner.danger {
    background: var(--alert-danger-bg);
    border-color: var(--alert-danger-border);
    color: var(--alert-danger-text);
  }

  .discrepancy-banner.warning {
    background: var(--alert-warning-bg);
    border-color: var(--alert-warning-border);
    color: var(--alert-warning-text);
  }

  .discrepancy-banner.success {
    background: var(--alert-success-bg);
    border-color: var(--alert-success-border);
    color: var(--alert-success-text);
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
    background: var(--alert-danger-bg);
    border: 1px solid var(--alert-danger-border);
    border-radius: var(--radius-md);
    padding: 16px;
  }

  .hold-actions-box h4 {
    color: var(--accent-rose);
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
    background: rgba(37, 99, 235, 0.08);
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
    color: var(--text-main);
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
    gap: 16px;
  }

  .qc-bundle-card {
    background: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    padding: 16px;
    transition: all 0.2s ease;
  }

  .qc-bundle-card.rejected {
    border-left: 6px solid var(--accent-rose);
    background: var(--alert-danger-bg);
  }

  .qc-bundle-card.passed {
    border-left: 6px solid var(--accent-emerald);
  }

  .qc-bundle-header {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 14px;
  }

  .bundle-identity {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .bundle-specs-tag {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .qc-bundle-body {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .photo-upload-dropzone {
    width: 100%;
    height: 44px;
    border: 1px dashed var(--border-color);
    border-radius: var(--radius-sm);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 13px;
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
    width: 72px;
    height: 72px;
  }

  .qc-image-preview img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: 6px;
    border: 1px solid var(--border-color);
  }

  .btn-remove-photo {
    position: absolute;
    top: -6px;
    right: -6px;
    background: var(--accent-rose);
    color: #fff;
    border: none;
    border-radius: 50%;
    width: 22px;
    height: 22px;
    font-size: 13px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .qc-remarks-section {
    width: 100%;
  }

  .sticky-mobile-submit {
    position: sticky;
    bottom: 16px;
    z-index: 10;
  }

  .giant-submit-btn {
    width: 100%;
    min-height: 52px;
    font-size: 16px;
    font-weight: 800;
    box-shadow: 0 4px 16px rgba(0,0,0,0.2);
  }

  .empty-state {
    text-align: center;
    padding: 30px;
    color: var(--text-secondary);
    font-size: 14px;
  }
</style>
