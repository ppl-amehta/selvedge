// Category-linked Inward specification, master product filtering, and size constraints.

const DEFAULT_INWARD_TYPES = [
  'Fabric', 'Zips', 'packaging', 'Trims', 'Label', 'Apparel',
  'Other Inwards', 'Electronic', 'Drinkwear', 'Headgear', 'Bags', 'Stationary'
];

const DEFAULT_ZIP_TEETH_SIZES = ['5.0', '6.0', '7.0', '8.0', '10.0', '12.0', '14.0', '16.0', 'na'];

const DEFAULT_APPAREL_GENRES = [
  'Polo', 'Tee', 'Hoodies and Pullovers', 'Jackets', 'Bottoms', 'Quarter zip',
  'Shirt', 'Shorts', 'Sweat Pants', 'Sweat shirt', 'Sweater', 'Sweatshirt',
  'Tights', 'Vest', 'Dress', 'Skorts', 'Onsies', 'Caps', 'Traded apparel'
];

const DEFAULT_NON_APPAREL_GENRES = [
  'Drinkwear', 'Headgear', 'Bags', 'Stationary', 'Electronics', 'Others'
];

const DEFAULT_GENDER_SIZES = {
  Men: ['XXS', 'XS', 'S', 'M', 'L', 'XL', '2XL', '3XL', '4XL', '5XL', 'Custom', '32', '34', '36', '38', '39', '40', '42', '44', '46', '48', '50', '52'],
  Women: ['XXS', 'XS', 'S', 'M', 'L', 'XL', '2XL', '3XL', '4XL', '5XL', 'Custom', '32', '34', '36', '38', '39', '40', '42', '44', '46', '48', '50', '52'],
  Unisex: ['XXS', 'XS', 'S', 'M', 'L', 'XL', '2XL', '3XL', '4XL', '5XL', 'Custom', '32', '34', '36', '38', '39', '40', '42', '44', '46', '48', '50', '52'],
  Infant: ['0-3', '3-6', '6-9', '9-12', '12-15', '15-18'],
  Kids: ['K16', 'K18', 'K20', 'K22', 'K24', 'K26', 'K28', 'K30', 'K32', 'K34', 'K36', 'kids custom'],
  Jr: ['Jr-XS', 'Jr-S', 'Jr-M', 'Jr-L', 'Jr Custom']
};

const DEFAULT_NON_APPAREL_SIZES = {
  Drinkwear: ['100ml', '150ml', '200ml', '250ml', '300ml', '500ml', '750ml', '1Litre', '1.5 Litre', '2 Litre'],
  Headgear: ['Std', 'Kids', 'Large', 'Small'],
  Caps: ['Std', 'Kids', 'Large', 'Small'],
  Bags: ['Std'],
  Stationary: ['A4', 'A3', 'A5', 'A6', 'B5'],
  Electronics: ['Std'],
  Others: ['Std']
};

const APPAREL_GENRE_SUBCATEGORIES = {
  'polo': ['polo'],
  'tee': ['tee', 't-shirt', 'tshirt'],
  'hoodies and pullovers': ['hoodie', 'hoodies and pullovers', 'pullover'],
  'jackets': ['jacket', 'jackets'],
  'bottoms': ['jogger', 'pant', 'sweat pants', 'shorts', 'tights', 'bottoms'],
  'quarter zip': ['quarter zip'],
  'shirt': ['shirt'],
  'shorts': ['shorts'],
  'sweat pants': ['sweat pants', 'jogger', 'pant'],
  'sweat shirt': ['sweat shirt', 'sweatshirt'],
  'sweatshirt': ['sweatshirt', 'sweat shirt'],
  'sweater': ['sweater'],
  'tights': ['tights'],
  'vest': ['vest'],
  'dress': ['dress'],
  'skorts': ['skorts'],
  'onsies': ['onsies', 'onesies'],
  'caps': ['headgear', 'caps'],
  'traded apparel': ['polo', 'tee', 'shirt', 'hoodie', 'jacket', 'jogger', 'pant', 'shorts', 'sweatshirt', 'vest', 'dress']
};

const NOISE_LABELS = new Set(['size', 'elastic size', 'cord size']);

/**
 * Filter Excel column-header artifacts and case-insensitive duplicates.
 * @param {any[]} list
 * @returns {string[]}
 */
export function cleanDisplayNames(list) {
  const seen = new Set();
  const out = [];
  for (const item of list || []) {
    const raw = typeof item === 'string' || typeof item === 'number' ? String(item) : (item?.display_name || item?.label || item?.name || '');
    const trimmed = raw.trim();
    if (!trimmed) continue;
    const lower = trimmed.toLowerCase();
    if (NOISE_LABELS.has(lower) || seen.has(lower)) continue;
    seen.add(lower);
    out.push(trimmed);
  }
  return out;
}

/**
 * Canonical kind classification for an inward type string.
 * Empty input yields 'unset' so clearing the combobox never silently falls back to 'fabric'.
 * @param {string} inwardType
 * @returns {'unset'|'fabric'|'zips'|'trims'|'packaging'|'label'|'apparel'|'non_apparel'|'other'}
 */
export function normInwardKind(inwardType) {
  const k = (inwardType || '').trim().toLowerCase();
  if (!k) return 'unset';
  if (k.startsWith('fabric')) return 'fabric';
  if (k === 'zips' || k === 'zip') return 'zips';
  if (k === 'trims' || k === 'trim') return 'trims';
  if (k === 'packaging') return 'packaging';
  if (k === 'label' || k === 'labels') return 'label';
  if (k === 'apparel' || k === 'finished goods') return 'apparel';
  if (['drinkwear', 'bags', 'bag', 'headgear', 'caps', 'electronic', 'electronics', 'stationary', 'stationery'].includes(k)) {
    return 'non_apparel';
  }
  return 'other';
}

/**
 * Resolve canonical Non-Apparel product genre from either inwardType or user-selected genre.
 * @param {string} value
 * @returns {string}
 */
export function normalizeNonApparelGenre(value) {
  const k = (value || '').trim().toLowerCase();
  if (!k) return '';
  if (k === 'drinkwear') return 'Drinkwear';
  if (k === 'bags' || k === 'bag') return 'Bags';
  if (k === 'headgear' || k === 'caps') return 'Headgear';
  if (k === 'electronic' || k === 'electronics') return 'Electronics';
  if (k === 'stationary' || k === 'stationery') return 'Stationary';
  if (k === 'others' || k === 'other' || k === 'other inwards') return 'Others';
  return '';
}

/**
 * Recommend default UOM when switching inward types or trim subtypes.
 * @param {string} inwardType
 * @param {string} [trimType]
 * @returns {string}
 */
export function defaultUomForInward(inwardType, trimType = '') {
  const kind = normInwardKind(inwardType);
  if (kind === 'fabric') return 'Kgs';
  if (kind === 'trims') {
    if (trimType === 'Thread') return '1000 mt tube';
    if (trimType === 'Elastic' || trimType === 'Tapes' || trimType === 'Cord') return 'Meter';
  }
  return 'Pcs';
}

/**
 * Available inward type dropdown options.
 * @param {Record<string, any>} masters
 * @returns {string[]}
 */
export function getInwardTypeOptions(masters) {
  const list = cleanDisplayNames(masters?.inward_types || []);
  return list.length > 0 ? list : DEFAULT_INWARD_TYPES;
}

/**
 * Sort Zip teeth sizes numerically ascending, pushing non-numeric values like 'na' last.
 * @param {Record<string, any>} masters
 * @returns {string[]}
 */
export function getZipTeethSizes(masters) {
  const raw = cleanDisplayNames(masters?.zip_sizes || []);
  const list = raw.length > 0 ? raw : [...DEFAULT_ZIP_TEETH_SIZES];
  return list.sort((a, b) => {
    const na = parseFloat(a);
    const nb = parseFloat(b);
    const aNum = Number.isFinite(na) && /^[\d.]+$/.test(a);
    const bNum = Number.isFinite(nb) && /^[\d.]+$/.test(b);
    if (aNum && bNum) return na - nb;
    if (aNum) return -1;
    if (bNum) return 1;
    return a.localeCompare(b);
  });
}

/**
 * Packaging variants scoped to the chosen packaging type.
 * @param {Record<string, any>} masters
 * @param {string} pType
 * @returns {string[]}
 */
export function getPackagingVariants(masters, pType) {
  const p = (pType || '').toLowerCase();
  if (p.includes('poly')) return cleanDisplayNames(masters?.packaging_polybag_brands);
  if (p.includes('tape')) return cleanDisplayNames(masters?.packaging_tape_variants);
  if (p.includes('loop')) return cleanDisplayNames(masters?.packaging_loop_materials);
  if (p.includes('carton')) return cleanDisplayNames(masters?.packaging_carton_brands);
  if (p.includes('courier')) return cleanDisplayNames(masters?.packaging_courier_brands);
  if (p.includes('hangtag') || p.includes('sticker')) return ['Blank', 'ETTU', 'Gone Golfing', 'FPCC'];
  return ['Blank', 'FPCC'];
}

/**
 * Packaging dimensions/sizes scoped strictly to the chosen packaging type.
 * @param {Record<string, any>} masters
 * @param {string} pType
 * @returns {string[]}
 */
export function getPackagingSizes(masters, pType) {
  const p = (pType || '').toLowerCase();
  if (p.includes('poly')) return cleanDisplayNames(masters?.packaging_polybag_sizes);
  if (p.includes('tape')) return cleanDisplayNames(masters?.packaging_tape_sizes);
  if (p.includes('loop')) return cleanDisplayNames(masters?.packaging_loop_sizes);
  if (p.includes('carton')) return cleanDisplayNames(masters?.packaging_carton_sizes);
  if (p.includes('shrink')) return cleanDisplayNames(masters?.packaging_shrink_sizes);
  if (p.includes('courier')) return cleanDisplayNames(masters?.packaging_courier_sizes);
  const generic = cleanDisplayNames(masters?.packaging_sizes);
  return generic.length > 0 ? generic : ['Std', 'Small', 'Medium', 'Large', 'Custom'];
}

/**
 * Packaging master items scoped by packaging type keyword.
 * @param {Record<string, any>} masters
 * @param {string} pType
 * @returns {{value: string, label: string}[]}
 */
export function getPackagingItems(masters, pType) {
  const all = (masters?.packaging || []).map(x => ({ value: x.display_name, label: x.display_name }));
  const p = (pType || '').toLowerCase().trim();
  if (!p || p === 'other') return all;
  const token = p.includes('poly')
    ? 'poly'
    : p.includes('carton') && !p.includes('tape')
      ? 'carton'
      : p.includes('tape')
        ? 'tape'
        : p.includes('courier')
          ? 'courier'
          : p.includes('hangtag')
            ? 'hangtag'
            : p.includes('sticker')
              ? 'sticker'
              : p;
  const matched = all.filter(item => item.label.toLowerCase().includes(token));
  return matched.length > 0 ? matched : all;
}

/**
 * Trim master items strictly scoped to trimType.
 * @param {Record<string, any>} masters
 * @param {string} tType
 * @returns {{value: string, label: string}[]}
 */
export function getTrimItems(masters, tType) {
  const t = (tType || '').trim();
  const mapObj = {
    Elastic: masters?.trim_elastics,
    Tapes: masters?.trim_tapes,
    Cord: masters?.trim_cords,
    Button: masters?.trim_buttons,
    Fusing: masters?.trim_fusings,
    Thread: masters?.trim_threads
  };
  if (mapObj[t]) {
    return (mapObj[t] || []).map(x => ({ value: x.display_name, label: x.display_name }));
  }
  if (t === 'Eyelets' || t === 'Stopper') {
    const key = t.toLowerCase().slice(0, -1); // eyelet / stopper
    return (masters?.trims || [])
      .filter(i => (i.display_name || '').toLowerCase().includes(key))
      .map(x => ({ value: x.display_name, label: x.display_name }));
  }
  return (masters?.trims || []).map(x => ({ value: x.display_name, label: x.display_name }));
}

/**
 * Trim sizes strictly scoped to trimType, removing Excel column-header noise.
 * @param {Record<string, any>} masters
 * @param {string} tType
 * @returns {string[]}
 */
export function getTrimSizes(masters, tType) {
  const t = (tType || '').trim();
  let list = [];
  if (t === 'Elastic') list = cleanDisplayNames(masters?.trim_elastic_sizes);
  else if (t === 'Tapes') list = cleanDisplayNames(masters?.trim_tape_sizes);
  else if (t === 'Cord') list = cleanDisplayNames(masters?.trim_cord_sizes);
  else if (t === 'Button') list = cleanDisplayNames(masters?.trim_button_sizes);
  else if (t === 'Fusing') list = cleanDisplayNames(masters?.trim_fusing_sizes);
  else if (t === 'Thread') list = cleanDisplayNames(masters?.trim_thread_sizes);
  else list = cleanDisplayNames(masters?.trim_other_sizes);
  return list.length > 0 ? list : ['Std'];
}

/**
 * Trim color choices scoped to trimType.
 * @param {Record<string, any>} masters
 * @param {string} tType
 * @param {{value: string, label: string}[]} colorOptions
 * @returns {{value: string, label: string}[]}
 */
export function getTrimColors(masters, tType, colorOptions = []) {
  const t = (tType || '').trim();
  if (t === 'Fusing') return (masters?.trim_fusing_colors || []).map(x => ({ value: x.display_name, label: x.display_name }));
  if (t === 'Other Trims' || t === 'Eyelets' || t === 'Stopper') {
    return (masters?.trim_other_colors || []).map(x => ({ value: x.display_name, label: x.display_name }));
  }
  return colorOptions;
}

/**
 * Apparel / Non-Apparel genre dropdown choices based on inward kind.
 * @param {Record<string, any>} masters
 * @param {string} inwardType
 * @returns {string[]}
 */
export function getGenreOptionsForInward(masters, inwardType) {
  const kind = normInwardKind(inwardType);
  if (kind === 'apparel') {
    const list = masters?.apparel_genres;
    return Array.isArray(list) && list.length > 0 ? list : DEFAULT_APPAREL_GENRES;
  }
  if (kind === 'non_apparel') {
    const locked = normalizeNonApparelGenre(inwardType);
    return locked ? [locked] : DEFAULT_NON_APPAREL_GENRES;
  }
  const nonApp = masters?.non_apparel_genres;
  return Array.isArray(nonApp) && nonApp.length > 0 ? nonApp : DEFAULT_NON_APPAREL_GENRES;
}

/**
 * Master product dropdown options strictly constrained to the selected Inward Type and sub-type.
 * Never leaks Apparel into Drinkwear/Bags/Zips/Fabric, nor Non-Apparel into Apparel.
 * @param {Record<string, any>} masters
 * @param {string} inwardType
 * @param {Record<string, any>} [ctx]
 * @returns {{value: string, label: string, subcategory?: string, category?: string}[]}
 */
export function getProductOptionsForInward(masters, inwardType, ctx = {}) {
  const kind = normInwardKind(inwardType);
  if (kind === 'unset') return [];

  if (kind === 'fabric') {
    return (masters?.fabrics || []).map(f => ({
      value: f.display_name,
      label: f.display_name,
      fabricType: f.meta?.fabricType || ''
    }));
  }

  if (kind === 'zips') {
    const dbZips = (masters?.zips || []).map(z => ({ value: z.display_name, label: z.display_name }));
    if (dbZips.length > 0) return dbZips;
    const zType = (ctx.zipType || 'Metal').trim();
    const teeth = (ctx.teethSize || '5.0').trim();
    const closure = (ctx.closure || 'open ended').trim();
    const types = cleanDisplayNames(masters?.zip_types);
    const closures = cleanDisplayNames(masters?.zip_closures);
    const primary = `${zType} #${teeth} ${closure}`.trim();
    const out = [{ value: primary, label: primary }];
    const seen = new Set([primary.toLowerCase()]);
    for (const c of closures) {
      const s = `${zType} #${teeth} ${c}`.trim();
      if (!seen.has(s.toLowerCase())) {
        seen.add(s.toLowerCase());
        out.push({ value: s, label: s });
      }
    }
    for (const t of types) {
      const s = `${t} #${teeth} ${closure}`.trim();
      if (!seen.has(s.toLowerCase())) {
        seen.add(s.toLowerCase());
        out.push({ value: s, label: s });
      }
    }
    return out;
  }

  if (kind === 'packaging') {
    return getPackagingItems(masters, ctx.packagingType || '');
  }

  if (kind === 'trims') {
    return getTrimItems(masters, ctx.trimType || '');
  }

  if (kind === 'label') {
    const list = ctx.labelCategory === 'Additional Label'
      ? (masters?.label_additionals || [])
      : (masters?.label_brands || []);
    return list.map(l => ({ value: l.display_name, label: l.display_name }));
  }

  const allProducts = masters?.products || [];

  if (kind === 'apparel') {
    const apparelProducts = allProducts.filter(p => {
      const cat = (p.meta?.category || '').trim().toLowerCase();
      const sub = (p.meta?.subcategory || p.meta?.genre || '').trim().toLowerCase();
      if (cat === 'apparel') return true;
      // Include newly created Apparel products that stored meta.genre without meta.category
      if (!cat && p.meta?.genre) {
        return !DEFAULT_NON_APPAREL_GENRES.map(g => g.toLowerCase()).includes(sub);
      }
      return false;
    });

    const genreKey = (ctx.genre || '').trim().toLowerCase();
    const formatItem = p => ({
      value: p.display_name,
      label: p.display_name,
      category: 'Apparel',
      subcategory: p.meta?.subcategory || p.meta?.genre || ''
    });

    if (!genreKey || genreKey === 'all' || genreKey === 'all apparel') {
      return apparelProducts.map(formatItem);
    }

    // Special case: Caps genre lives under Non Apparel Headgear if Apparel has none
    if (genreKey === 'caps') {
      const capItems = allProducts.filter(p => (p.meta?.subcategory || '').trim().toLowerCase() === 'headgear');
      if (capItems.length > 0) return capItems.map(formatItem);
    }

    const targets = APPAREL_GENRE_SUBCATEGORIES[genreKey] || [genreKey];
    const filtered = apparelProducts.filter(p => {
      const sub = (p.meta?.subcategory || p.meta?.genre || '').trim().toLowerCase();
      return targets.includes(sub) || sub === genreKey;
    });
    return filtered.map(formatItem);
  }

  // non_apparel or other ('Other Inwards')
  const nonApparelProducts = allProducts.filter(p => {
    const cat = (p.meta?.category || '').trim().toLowerCase();
    return cat === 'non apparel' || cat === 'non-apparel';
  });
  const lockedGenre = normalizeNonApparelGenre(inwardType);
  const chosenGenre = lockedGenre || normalizeNonApparelGenre(ctx.genre || '') || (ctx.genre || '').trim();
  const formatNonApparel = p => ({
    value: p.display_name,
    label: p.display_name,
    category: 'Non Apparel',
    subcategory: p.meta?.subcategory || ''
  });

  if (!chosenGenre || chosenGenre.toLowerCase() === 'all') {
    return nonApparelProducts.map(formatNonApparel);
  }

  const targetLower = chosenGenre.toLowerCase();
  return nonApparelProducts
    .filter(p => (p.meta?.subcategory || '').trim().toLowerCase() === targetLower)
    .map(formatNonApparel);
}

/**
 * Allowed bundle/lot sizes strictly constrained to the selected Inward Type, genre, gender, and subtype.
 * Guarantees Drinkwear sizes never appear on Apparel, Stationary sizes never appear on Bags, etc.
 * @param {Record<string, any>} masters
 * @param {string} inwardType
 * @param {Record<string, any>} [ctx]
 * @returns {string[]}
 */
export function getAllowedSizesForInward(masters, inwardType, ctx = {}) {
  const kind = normInwardKind(inwardType);
  if (kind === 'unset' || kind === 'fabric') return [];
  if (kind === 'zips') return getZipTeethSizes(masters);
  if (kind === 'trims') return getTrimSizes(masters, ctx.trimType || '');
  if (kind === 'packaging') return getPackagingSizes(masters, ctx.packagingType || '');
  if (kind === 'label') {
    const genres = cleanDisplayNames(masters?.label_genres);
    return genres.length > 0 ? genres : ['Woven', 'Printed', 'DTF', 'Tareaway'];
  }
  if (kind === 'apparel') {
    const gMap = (masters?.gender_sizes && Object.keys(masters.gender_sizes).length > 0)
      ? masters.gender_sizes
      : DEFAULT_GENDER_SIZES;
    const gender = (ctx.gender || '').trim();
    if (gender && gender !== 'All Apparel' && Array.isArray(gMap[gender])) {
      return gMap[gender];
    }
    const seen = new Set();
    const union = [];
    for (const key of Object.keys(gMap)) {
      for (const s of gMap[key] || []) {
        if (!seen.has(s)) {
          seen.add(s);
          union.push(s);
        }
      }
    }
    return union;
  }

  // non_apparel or other
  const naMap = (masters?.non_apparel_sizes && Object.keys(masters.non_apparel_sizes).length > 0)
    ? masters.non_apparel_sizes
    : DEFAULT_NON_APPAREL_SIZES;
  const genreKey = normalizeNonApparelGenre(inwardType) || normalizeNonApparelGenre(ctx.genre || '') || 'Others';
  return naMap[genreKey] || naMap['Others'] || ['Std'];
}

/**
 * Pick default bundle sizeSpec for the active inward configuration.
 * @param {Record<string, any>} masters
 * @param {string} inwardType
 * @param {Record<string, any>} [ctx]
 * @returns {string}
 */
export function getDefaultSizeSpecForInward(masters, inwardType, ctx = {}) {
  const kind = normInwardKind(inwardType);
  if (kind === 'zips') return ctx.teethSize || getZipTeethSizes(masters)[0] || '';
  if (kind === 'trims') return ctx.trimSize || getTrimSizes(masters, ctx.trimType || '')[0] || '';
  if (kind === 'packaging') return ctx.packagingSize || getPackagingSizes(masters, ctx.packagingType || '')[0] || '';
  if (kind === 'label') return ctx.labelGenre || cleanDisplayNames(masters?.label_genres)[0] || 'Woven';
  const allowed = getAllowedSizesForInward(masters, inwardType, ctx);
  if (kind === 'apparel' && allowed.includes('M')) return 'M';
  return allowed[0] || '';
}

/**
 * Reconcile bundle sizeSpec when inward type, subtype, or gender changes so incompatible sizes cannot stick.
 * @param {string} currentSpec
 * @param {string[]} allowedSizes
 * @param {string} defaultSpec
 * @returns {string}
 */
export function reconcileBundleSizeSpec(currentSpec, allowedSizes, defaultSpec) {
  const trimmed = (currentSpec || '').trim();
  if (allowedSizes.length === 0) return trimmed || defaultSpec;
  if (trimmed && allowedSizes.includes(trimmed)) return trimmed;
  return defaultSpec;
}
