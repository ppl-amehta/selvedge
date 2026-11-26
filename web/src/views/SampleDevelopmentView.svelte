<script>
  import { onMount } from 'svelte';
  import { api } from '../api.js';
  import Combobox from '../components/Combobox.svelte';
  import ProductImageUploader from '../components/ProductImageUploader.svelte';
  import VoiceNoteRecorder from '../components/VoiceNoteRecorder.svelte';

  let { masters = {} } = $props();

  // Active Station Tab: 'sample' | 'product' | 'client-pricing'
  let activeTab = $state('sample');

  // ============================================================
  // TAB 1: SAMPLE DEVELOPMENT STATE
  // ============================================================
  let sampleId = $state('');
  let developmentNo = $state('');
  let developmentDate = $state(new Date().toISOString().split('T')[0]);
  let clientId = $state('');
  let clientName = $state('');
  let sampleCategory = $state('Apparel'); // 'Apparel' | 'Non Apparel'
  let productGenre = $state('Polo');
  let tempProductName = $state('');
  let sampleColors = $state([]);
  let newSampleColor = $state('');
  let imageUrl = $state('');
  let productDescription = $state('');

  const defaultFabricAttrs = () => ({ fabricType: 'Fabric', fabricGeneric: '', composition: '', gsm: 200, dia: '', color: '' });
  const defaultTrimAttrs = (tType = 'Elastic') => ({ trimType: tType, size: '', color: '', cordEnd: '', gsm: '', length: '', tex: '' });
  const defaultZipAttrs = () => ({ zipType: '', teethSize: '', slider: '', closure: '', finish: '', tapeColor: '' });
  const defaultLabelAttrs = () => ({ labelCategory: 'Brand Label', labelGenre: 'Woven', labelDescription: '', baseColor: '', textColor: '' });

  // Multi-item child tables
  let sizes = $state([
    { id: '', gender: 'Men', size: 'M', qty: 1 }
  ]);
  let fabrics = $state([
    { id: '', fabricName: '', lotNo: '', average: 0.35, attrs: defaultFabricAttrs() }
  ]);
  let trims = $state([
    { id: '', trimName: '', average: 3.0, attrs: defaultTrimAttrs('Button') }
  ]);
  let zips = $state([
    { id: '', zipName: '', lengthAndQty: '', attrs: defaultZipAttrs() }
  ]);
  let labels = $state([
    { id: '', labelName: '', attrs: defaultLabelAttrs() }
  ]);
  let pharmas = $state([
    { id: '', basePharmaNo: '', changesMade: '', newPharmaNo: '' }
  ]);

  // Post-Sampling / Review Fields
  let tempMeasurementChart = $state('');
  let devComments = $state('');
  let mastersComments = $state('');
  let reworkNotes = $state('');
  let sampleImageUrl = $state('');
  let comments = $state('');
  let sampleStatus = $state('DRAFT');

  // Drawer / Sample List
  let sampleList = $state([]);
  let sampleSearchQuery = $state('');
  let isLoadingSamples = $state(false);
  let isSavingSample = $state(false);

  // ============================================================
  // TAB 2: PRODUCT CREATION STATE
  // ============================================================
  let prodId = $state('');
  let selectedSampleToPull = $state('');
  let creationDate = $state(new Date().toISOString().split('T')[0]);
  let prodGenre = $state('Polo');
  let prodName = $state('');
  let brandLabel = $state('');
  let skuCode = $state('');
  let whyWeMadeThis = $state('');
  let garmentSpecifications = $state('');
  let selectedGarmentSpecs = $state([]);
  let customGarmentSpecInput = $state('');
  let composition = $state('');
  let mrp = $state(1499.0);

  // Per-Gender / Non-Apparel Multi-Image, Pharma, Sizes and Color Variant states
  let newImgUrlInput = $state({});
  let newColorNameInput = $state({});
  let newColorImgInput = $state({});
  let uploadingTarget = $state(''); // Key identifying active file upload spinner

  // Standard Garment Specifications (Curated for Apparel & Non-Apparel Genres)
  const standardGarmentSpecsMap = {
    'Polo': 'Rib collar, 2-3 button placket, side vents, tennis tail, twin needle hem',
    'T-Shirt': 'Classic crew neck with internal neck tape, short sleeves, twin needle hem',
    'Round Neck T-Shirt': 'Classic crew neck with internal neck tape, short sleeves, twin needle hem',
    'V-Neck T-Shirt': 'Reinforced V-neck collar, short sleeves, regular fit',
    'Henley': '3-button placket, banded crew neck, set-in sleeves',
    'Hoodie': 'Drawcord adjustable hood, front kangaroo pocket, ribbed cuffs and hem',
    'Sweatshirt': 'Ribbed crew neck, ribbed cuffs and bottom band, set-in sleeves',
    'Jacket': 'Full-zip front, mock neck / stand collar, side welt pockets, elasticated cuffs',
    'Windcheater': 'Water-resistant nylon shell, hooded with elastic drawcord, zip front',
    'Shorts': 'Elasticated waistband with tubular drawcord, deep side pockets, above-knee cut',
    'Track Pants': 'Elasticated waistband with drawcord, side zipper pockets, tapered leg',
    'Joggers': 'Elasticated ribbed waistband, side pockets, articulated knee, ribbed ankle cuffs',
    'Leggings': 'High-rise wide waistband, 4-way stretch compression, flatlock anti-chafing seams',
    'Bags': 'Heavy-duty 16oz canvas, reinforced webbing grab handles, zippered main compartment',
    'Caps': '6-panel structured crown, curved visor, brass buckle adjustable back strap',
    'Socks': 'Ribbed arch compression band, reinforced cushioned heel and toe',
    'Water Bottles': 'Double-wall vacuum insulated stainless steel body, leak-proof twist cap',
    'Towels': '100% cotton terry weave, high-absorbency, double-stitched hems'
  };

  const standardGarmentSpecsList = [
    'Rib collar, 2-3 button placket, side vents, tennis tail, twin needle hem',
    'Classic crew neck with internal neck tape, short sleeves, twin needle hem',
    'Reinforced V-neck collar, short sleeves, regular fit',
    '3-button placket, banded crew neck, set-in sleeves',
    'Drawcord adjustable hood, front kangaroo pocket, ribbed cuffs and hem',
    'Ribbed crew neck, ribbed cuffs and bottom band, set-in sleeves',
    'Full-zip front, mock neck / stand collar, side welt pockets, elasticated cuffs',
    'Water-resistant nylon shell, hooded with elastic drawcord, zip front',
    'Elasticated waistband with tubular drawcord, deep side pockets, above-knee cut',
    'Elasticated waistband with drawcord, side zipper pockets, tapered leg',
    'Elasticated ribbed waistband, side pockets, articulated knee, ribbed ankle cuffs',
    'High-rise wide waistband, 4-way stretch compression, flatlock anti-chafing seams',
    'Heavy-duty 16oz canvas, reinforced webbing grab handles, zippered main compartment',
    '6-panel structured crown, curved visor, brass buckle adjustable back strap',
    'Ribbed arch compression band, reinforced cushioned heel and toe',
    'Double-wall vacuum insulated stainless steel body, leak-proof twist cap',
    '100% cotton terry weave, high-absorbency, double-stitched hems'
  ];

  // Gender variations
  let prodGenders = $state([
    {
      id: '',
      gender: 'Men',
      sizes: ['S', 'M', 'L', 'XL'],
      pharmaNo: '',
      pharmaNos: [''],
      measurementChartUrl: '',
      productImages: [],
      colorImages: []
    }
  ]);

  // Materials BOM
  let prodMaterials = $state([
    { id: '', materialType: 'fabric', name: '', average: 0.35, gsm: 220, dia: 32, zipLengthQty: '' },
    { id: '', materialType: 'trim', name: '', average: 3.0, gsm: 0, dia: 0, zipLengthQty: '' }
  ]);

  let productList = $state([]);
  let isPublishingProduct = $state(false);

  // ============================================================
  // TAB 3: CLIENT SPECIFIC PRICING STATE
  // ============================================================
  let cpClientId = $state('');
  let cpClientName = $state('');
  let cpProductId = $state('');
  let cpProductName = $state('');
  let cpProductDefaultImage = $state('');
  let cpClientImageUrl = $state('');
  let cpContractedPrice = $state(0.0);
  let cpComments = $state('');
  let cpEffectiveDate = $state(new Date().toISOString().split('T')[0]);

  let clientPricingList = $state([]);
  let cpFilterClient = $state('');
  let isSavingPricing = $state(false);

  // Authoritative Canonical Taxonomies
  const canonicalGenders = ['Men', 'Women', 'Unisex', 'Infant', 'Jr', 'Kids'];

  const fallbackGenderSizes = {
    Men: ['XXS', 'XS', 'S', 'M', 'L', 'XL', '2XL', '3XL', '4XL', '5XL', 'Custom', '32', '34', '36', '38', '39', '40', '42', '44', '46', '48', '50', '52'],
    Women: ['XXS', 'XS', 'S', 'M', 'L', 'XL', '2XL', '3XL', '4XL', '5XL', 'Custom', '32', '34', '36', '38', '39', '40', '42', '44', '46', '48', '50', '52'],
    Unisex: ['XXS', 'XS', 'S', 'M', 'L', 'XL', '2XL', '3XL', '4XL', '5XL', 'Custom', '32', '34', '36', '38', '39', '40', '42', '44', '46', '48', '50', '52'],
    Infant: ['0-3', '3-6', '6-9', '9-12', '12-15', '15-18'],
    Kids: ['K16', 'K18', 'K20', 'K22', 'K24', 'K26', 'K28', 'K30', 'K32', 'K34', 'K36', 'kids custom'],
    Jr: ['Jr-XS', 'Jr-S', 'Jr-M', 'Jr-L', 'Jr Custom']
  };

  // Non-Apparel Authoritative Size / Capacity Taxonomy (Sheet 4 from Excel)
  const nonApparelSizesMap = {
    'Drinkwear': ['100ml', '150ml', '200ml', '250ml', '300ml', '500ml', '750ml', '1 Litre', '1.5 Litre', '2 Litre'],
    'Headgear': ['Std', 'Kids', 'Large', 'Small'],
    'Caps': ['Std', 'Kids', 'Large', 'Small'],
    'Stationary': ['A4', 'A3', 'A5', 'A6', 'B5'],
    'Stationery': ['A4', 'A3', 'A5', 'A6', 'B5'],
    'Bags': ['15 Inch', '15.6 Inch', '10.5 Inch', 'Std'],
    'Electronics': ['Std'],
    'Others': ['Std']
  };

  const bottomwearGenres = ['Bottoms', 'Shorts', 'Sweat Pants', 'Tights', 'Skorts'];
  const zipIntensiveGenres = ['Jackets', 'Quarter zip', 'Hoodies and Pullovers', 'Sweater', 'Bottoms'];
  const buttonIntensiveGenres = ['Polo', 'Shirt'];

  let apparelGenres = $derived(
    (masters.apparel_genres && masters.apparel_genres.length > 0)
      ? masters.apparel_genres
      : [
          'Polo', 'Tee', 'Hoodies and Pullovers', 'Jackets', 'Bottoms', 'Quarter zip',
          'Shirt', 'Shorts', 'Sweat Pants', 'Sweat shirt', 'Sweater', 'Sweatshirt',
          'Tights', 'Vest', 'Dress', 'Skorts', 'Onsies', 'Caps', 'Traded apparel'
        ]
  );

  let nonApparelGenres = $derived(
    (masters.non_apparel_genres && masters.non_apparel_genres.length > 0)
      ? masters.non_apparel_genres
      : ['Drinkwear', 'Headgear', 'Bags', 'Stationary', 'Electronics', 'Others']
  );

  let isNonApparel = $derived(nonApparelGenres.includes(productGenre));
  let isBottomwear = $derived(bottomwearGenres.includes(productGenre));
  let isZipExpected = $derived(zipIntensiveGenres.includes(productGenre));
  let isButtonExpected = $derived(buttonIntensiveGenres.includes(productGenre));

  let isProdNonApparel = $derived(nonApparelGenres.includes(prodGenre));

  // Authoritative BOM & Material Engineering Recommendations by Genre
  const genreBOMRecommendations = {
    'Polo': {
      avgFabric: 0.35,
      suggestedFabric: 'Pique',
      trims: [
        { trimName: 'Buttons (2-3 pcs, 18L)', average: 3.0 },
        { trimName: 'Collar Fusing', average: 1.0 },
        { trimName: 'Sewing Thread Tex 30', average: 1.0 }
      ],
      zips: []
    },
    'Tee': {
      avgFabric: 0.22,
      suggestedFabric: 'Single Jersey',
      trims: [
        { trimName: 'Rib Neck Fusing / Tape', average: 1.0 },
        { trimName: 'Sewing Thread Tex 30', average: 1.0 }
      ],
      zips: []
    },
    'Shirt': {
      avgFabric: 0.28,
      suggestedFabric: 'Woven Cotton',
      trims: [
        { trimName: 'Shirt Buttons (7-9 pcs, 16L)', average: 8.0 },
        { trimName: 'Collar & Cuff Interlining Fusing', average: 1.0 },
        { trimName: 'Sewing Thread Tex 24', average: 1.0 }
      ],
      zips: []
    },
    'Bottoms': {
      avgFabric: 0.50,
      suggestedFabric: 'Twill / Knit Terry',
      trims: [
        { trimName: 'Waist Elastic 40mm', average: 0.85 },
        { trimName: 'Waist Drawcord / Rope', average: 1.2 },
        { trimName: 'Eyelets (2 pcs)', average: 2.0 },
        { trimName: 'Pocketing Fabric', average: 0.15 },
        { trimName: 'Sewing Thread Tex 40', average: 1.0 }
      ],
      zips: [
        { zipName: 'Nylon Fly Zip 15cm', lengthAndQty: 'Std (1pc)' }
      ]
    },
    'Shorts': {
      avgFabric: 0.28,
      suggestedFabric: 'French Terry / Poly Spandex',
      trims: [
        { trimName: 'Waist Elastic 35mm', average: 0.75 },
        { trimName: 'Drawstring', average: 1.1 },
        { trimName: 'Eyelets', average: 2.0 },
        { trimName: 'Sewing Thread Tex 40', average: 1.0 }
      ],
      zips: []
    },
    'Sweat Pants': {
      avgFabric: 0.55,
      suggestedFabric: 'Fleece / French Terry',
      trims: [
        { trimName: 'Waist Elastic 40mm', average: 0.85 },
        { trimName: 'Heavy Drawcord', average: 1.3 },
        { trimName: 'Eyelets (2 pcs)', average: 2.0 },
        { trimName: 'Rib Ankle Cuff', average: 0.2 },
        { trimName: 'Sewing Thread Tex 40', average: 1.0 }
      ],
      zips: []
    },
    'Hoodies and Pullovers': {
      avgFabric: 0.75,
      suggestedFabric: 'Fleece 280-320 GSM',
      trims: [
        { trimName: 'Hood Drawcord with Metal Tips', average: 1.4 },
        { trimName: 'Metal Eyelets', average: 2.0 },
        { trimName: 'Rib Waist & Cuff', average: 0.25 },
        { trimName: 'Sewing Thread Tex 40', average: 1.0 }
      ],
      zips: [
        { zipName: 'Vislon 5.0 Open End Front Zip (if zip hoodie)', lengthAndQty: 'S-L: 65cm, XL-3XL: 70cm' }
      ]
    },
    'Jackets': {
      avgFabric: 0.85,
      suggestedFabric: 'Bonded Box Fleece / Softshell',
      trims: [
        { trimName: 'Lining / Mesh Fabric', average: 0.45 },
        { trimName: 'Metal Snap Buttons', average: 6.0 },
        { trimName: 'Collar / Cuff Rib', average: 0.2 },
        { trimName: 'Sewing Thread Tex 40', average: 1.0 }
      ],
      zips: [
        { zipName: 'Metal 5.0 Open End Reversible Front Zip', lengthAndQty: 'S-L: 68cm, XL-3XL: 72cm' },
        { zipName: 'Concealed Pocket Zips 18cm (2 pcs)', lengthAndQty: '18cm (2pcs)' }
      ]
    },
    'Drinkwear': {
      avgFabric: 1.0,
      suggestedFabric: '18/8 Stainless Steel (Double Walled Vacuum)',
      trims: [
        { trimName: 'Silicone Leakproof Gasket Seal', average: 1.0 },
        { trimName: 'BPA-Free Polypropylene Lid', average: 1.0 }
      ],
      zips: []
    },
    'Stationary': {
      avgFabric: 1.0,
      suggestedFabric: '120 GSM Natural Shade Woodfree Paper',
      trims: [
        { trimName: 'Hardcover Greyboard 2mm', average: 1.0 },
        { trimName: 'Twin Ring Wire-O Binding', average: 1.0 },
        { trimName: 'Satin Bookmark Ribbon', average: 1.0 }
      ],
      zips: []
    },
    'Bags': {
      avgFabric: 1.0,
      suggestedFabric: '900D Water Repellent Poly Canvas',
      trims: [
        { trimName: 'Nylon Webbing Shoulder Straps (38mm)', average: 2.2 },
        { trimName: 'Quick Release POM Buckles', average: 2.0 },
        { trimName: 'Sewing Thread Tex 60', average: 1.0 }
      ],
      zips: [
        { zipName: 'Heavy Duty 8.0 Coil Main Compartment Zip', lengthAndQty: '60cm (1pc)' },
        { zipName: 'Front Accessory Pocket Zip 5.0', lengthAndQty: '25cm (1pc)' }
      ]
    }
  };

  function getSizesForSampling(g, genre) {
    if (nonApparelGenres.includes(genre)) {
      if (masters?.non_apparel_sizes?.[genre] && masters.non_apparel_sizes[genre].length > 0) {
        return masters.non_apparel_sizes[genre];
      }
      return nonApparelSizesMap[genre] || ['Std'];
    }

    // It's apparel! Check demographic scale
    if (g === 'Infant') {
      return masters?.gender_sizes?.['Infant'] || ['0-3', '3-6', '6-9', '9-12', '12-15', '15-18'];
    }
    if (g === 'Kids') {
      return masters?.gender_sizes?.['Kids'] || ['K16', 'K18', 'K20', 'K22', 'K24', 'K26', 'K28', 'K30', 'K32', 'K34', 'K36', 'kids custom'];
    }
    if (g === 'Jr') {
      return masters?.gender_sizes?.['Jr'] || ['Jr-XS', 'Jr-S', 'Jr-M', 'Jr-L', 'Jr Custom'];
    }

    // Adult scales (Men, Women, Unisex)
    const baseAdultAlpha = ['XXS', 'XS', 'S', 'M', 'L', 'XL', '2XL', '3XL', '4XL', '5XL', 'Custom'];
    const waistSizes = ['32', '34', '36', '38', '39', '40', '42', '44', '46', '48', '50', '52'];

    if (bottomwearGenres.includes(genre)) {
      return [...baseAdultAlpha, ...waistSizes];
    }
    return baseAdultAlpha;
  }

  function getSizesForGender(g) {
    return getSizesForSampling(g, productGenre);
  }

  function handleProductGenreChange(newGenre) {
    productGenre = newGenre;
    const isNowNonApparel = nonApparelGenres.includes(newGenre);
    sampleCategory = isNowNonApparel ? 'Non Apparel' : 'Apparel';

    if (isNowNonApparel) {
      const validSizes = getSizesForSampling('N/A', newGenre);
      sizes = sizes.map(s => ({
        ...s,
        gender: 'N/A',
        size: validSizes.includes(s.size) ? s.size : (validSizes[0] || 'Std')
      }));
    } else {
      sizes = sizes.map(s => {
        const targetGender = s.gender === 'N/A' ? 'Men' : s.gender;
        const validSizes = getSizesForSampling(targetGender, newGenre);
        return {
          ...s,
          gender: targetGender,
          size: validSizes.includes(s.size) ? s.size : (validSizes[0] || 'M')
        };
      });
    }
  }

  function handleSampleCategoryChange(newCat) {
    sampleCategory = newCat;
    if (newCat === 'Non Apparel') {
      const firstNonApp = nonApparelGenres[0] || 'Drinkwear';
      handleProductGenreChange(firstNonApp);
    } else {
      const firstApp = apparelGenres[0] || 'Polo';
      handleProductGenreChange(firstApp);
    }
  }

  function trimItemOptionsForType(tType) {
    const t = (tType || '').trim();
    if (t === 'Elastic') return masters.trim_elastics || [];
    if (t === 'Tapes') return masters.trim_tapes || [];
    if (t === 'Cord') return masters.trim_cords || [];
    if (t === 'Button') return masters.trim_buttons || [];
    if (t === 'Fusing') return masters.trim_fusings || [];
    if (t === 'Thread') return masters.trim_threads || [];
    if (t === 'Eyelets') return (masters.trims || []).filter(i => (i.display_name || '').toLowerCase().includes('eyelet'));
    if (t === 'Stopper') return (masters.trims || []).filter(i => (i.display_name || '').toLowerCase().includes('stopper'));
    return masters.trims || [];
  }

  function trimSizeOptionsForType(tType) {
    const t = (tType || '').trim();
    if (t === 'Elastic') return masters.trim_elastic_sizes || [];
    if (t === 'Tapes') return masters.trim_tape_sizes || [];
    if (t === 'Cord') return masters.trim_cord_sizes || [];
    if (t === 'Button') return masters.trim_button_sizes || [];
    if (t === 'Fusing') return masters.trim_fusing_sizes || [];
    if (t === 'Thread') return masters.trim_thread_sizes || [];
    return masters.trim_other_sizes || [];
  }

  function trimColorOptionsForType(tType) {
    const t = (tType || '').trim();
    if (t === 'Fusing') return masters.trim_fusing_colors || [];
    if (t === 'Other Trims' || t === 'Eyelets' || t === 'Stopper') return masters.trim_other_colors || masters.colors || [];
    return masters.colors || [];
  }

  function applyGenreBOMPreset(genre) {
    const preset = genreBOMRecommendations[genre];
    if (!preset) return;
    const isTargetNonApparel = nonApparelGenres.includes(genre);

    // Apply fabrics
    if (preset.suggestedFabric) {
      fabrics = [{
        id: '',
        fabricName: preset.suggestedFabric,
        lotNo: '',
        average: preset.avgFabric,
        attrs: defaultFabricAttrs()
      }];
    }

    // Apply trims
    if (preset.trims && preset.trims.length > 0) {
      trims = preset.trims.map(t => ({ id: '', trimName: t.trimName, average: t.average, attrs: defaultTrimAttrs('Other Trims') }));
    }

    // Apply zips
    if (preset.zips && preset.zips.length > 0) {
      zips = preset.zips.map(z => ({ id: '', zipName: z.zipName, lengthAndQty: z.lengthAndQty, attrs: defaultZipAttrs() }));
    } else if (!zipIntensiveGenres.includes(genre) && !isTargetNonApparel) {
      zips = [];
    }

    if (window.showToast) {
      window.showToast(`Preloaded standard BOM specifications for ${genre}`, 'success');
    }
  }

  let authoritativeCompositions = $derived.by(() => {
    if (masters.compositions && masters.compositions.length > 0) {
      return masters.compositions.map(c => c.display_name || c.code || c);
    }
    return [
      '95%Poly 5%SP', '40%Nylon 38%Poly 12%SP', '92%Poly 8%SP', '100%Poly',
      '88%Poly 12%SP', '60%Cot 30%Tencel 10%SP', '90%Poly 10%SP', '100%Cot',
      '90%Nylon 10%SP', '58%R.Cot 38%R.Poly 4%SP', '65%R.Cot 35%R.Poly',
      '85%Poly 15%SP', '52%Cot 48%Poly', '60%Modal 30%Tencel 10%SP',
      '95%Cot 5%SP', '60%Cot 40%Poly', '65%Poly 35%Cot',
      '50%A Cot 45% Siro Viscose 5%SP', '100%R.Poly', '50%Cot 45%Poly 5%SP',
      '70%Nylon 30%SP', '43%Cot 27%Viscose 30%Poly'
    ];
  });

  onMount(async () => {
    await refreshNextDevNo();
    await loadSampleList();
    await loadProductList();
    await loadClientPricingList();
  });

  async function refreshNextDevNo() {
    try {
      developmentNo = await api.getNextDevelopmentNo();
    } catch (e) {
      console.error('Failed to get next dev no:', e);
    }
  }

  async function loadSampleList() {
    try {
      isLoadingSamples = true;
      sampleList = await api.listSampleDevelopments(50, 0);
    } catch (e) {
      console.error('Failed to load sample list:', e);
    } finally {
      isLoadingSamples = false;
    }
  }

  async function loadProductList() {
    try {
      productList = await api.listMasterProducts(50, 0);
    } catch (e) {
      console.error('Failed to load product list:', e);
    }
  }

  async function loadClientPricingList() {
    try {
      clientPricingList = await api.listClientProductPricing(cpFilterClient);
    } catch (e) {
      console.error('Failed to load pricing list:', e);
    }
  }

  // Auto-generate SKU Code
  async function refreshSKU() {
    try {
      skuCode = await api.suggestSKU(brandLabel, prodGenre, prodName);
    } catch (e) {
      console.error('Failed to suggest SKU:', e);
    }
  }

  // ============================================================
  // TAB 1 HANDLERS: SAMPLE DEVELOPMENT
  // ============================================================
  function handleClientSelect(client) {
    if (!client) return;
    clientName = client.name || '';
    clientId = client.id || '';
  }

  function addSizeRow() {
    if (isNonApparel) {
      const valid = getSizesForSampling('N/A', productGenre);
      sizes = [...sizes, { id: '', gender: 'N/A', size: valid[0] || 'Std', qty: 1 }];
    } else {
      const valid = getSizesForSampling('Men', productGenre);
      sizes = [...sizes, { id: '', gender: 'Men', size: valid.includes('L') ? 'L' : (valid[0] || 'M'), qty: 1 }];
    }
  }
  function removeSizeRow(idx) {
    sizes = sizes.filter((_, i) => i !== idx);
  }

  function addSampleColor(val) {
    const name = typeof val === 'object' ? (val?.display_name || val?.code || '') : String(val || '');
    const clean = name.trim();
    if (!clean) return;
    if (!sampleColors.some(c => c.toLowerCase() === clean.toLowerCase())) {
      sampleColors = [...sampleColors, clean];
    }
    newSampleColor = '';
  }
  function removeSampleColor(idx) {
    sampleColors = sampleColors.filter((_, i) => i !== idx);
  }

  function addFabricRow() {
    fabrics = [...fabrics, { id: '', fabricName: '', lotNo: '', average: 0.35, attrs: defaultFabricAttrs() }];
  }
  function removeFabricRow(idx) {
    fabrics = fabrics.filter((_, i) => i !== idx);
  }

  function addTrimRow() {
    trims = [...trims, { id: '', trimName: '', average: 1.0, attrs: defaultTrimAttrs('Elastic') }];
  }
  function removeTrimRow(idx) {
    trims = trims.filter((_, i) => i !== idx);
  }

  function addZipRow() {
    zips = [...zips, { id: '', zipName: '', lengthAndQty: '', attrs: defaultZipAttrs() }];
  }
  function removeZipRow(idx) {
    zips = zips.filter((_, i) => i !== idx);
  }

  function addLabelRow() {
    labels = [...labels, { id: '', labelName: '', attrs: defaultLabelAttrs() }];
  }
  function removeLabelRow(idx) {
    labels = labels.filter((_, i) => i !== idx);
  }

  function addPharmaRow() {
    pharmas = [...pharmas, { id: '', basePharmaNo: '', changesMade: '', newPharmaNo: '' }];
  }
  function removePharmaRow(idx) {
    pharmas = pharmas.filter((_, i) => i !== idx);
  }

  async function saveSampleDevelopment(targetStatus) {
    if (!clientName) {
      alert('Please select a Client');
      return;
    }
    if (!tempProductName) {
      alert('Please enter a Temp Product Name');
      return;
    }

    try {
      isSavingSample = true;
      const statusToSave = targetStatus || sampleStatus || 'DRAFT';
      const payload = {
        id: sampleId,
        developmentNo: developmentNo,
        developmentDate: developmentDate,
        clientId: clientId,
        clientName: clientName,
        productGenre: productGenre,
        tempProductName: tempProductName,
        colors: sampleColors,
        imageUrl: imageUrl,
        productDescription: productDescription,
        tempMeasurementChart: tempMeasurementChart,
        devComments: devComments,
        mastersComments: mastersComments,
        reworkNotes: reworkNotes,
        sampleImageUrl: sampleImageUrl,
        comments: comments,
        status: statusToSave,
        sizes: sizes,
        fabrics: fabrics,
        trims: trims,
        zips: zips,
        labels: labels,
        pharmas: pharmas
      };

      const res = await api.createSampleDevelopment(payload);
      sampleId = res.id;
      sampleStatus = res.status;
      if (res.developmentNo) developmentNo = res.developmentNo;

      if (window.showToast) {
        window.showToast(`Sample Development ${developmentNo} saved (${sampleStatus})`, 'success');
      }
      await loadSampleList();
    } catch (err) {
      alert(err.message || 'Failed to save sample development');
    } finally {
      isSavingSample = false;
    }
  }

  async function loadSampleIntoForm(sample) {
    try {
      const detail = await api.getSampleDevelopment(sample.id);
      sampleId = detail.id;
      developmentNo = detail.developmentNo;
      developmentDate = detail.developmentDate;
      clientId = detail.clientId;
      clientName = detail.clientName;
      productGenre = detail.productGenre || 'Polo';
      sampleCategory = nonApparelGenres.includes(productGenre) ? 'Non Apparel' : 'Apparel';
      tempProductName = detail.tempProductName || '';
      sampleColors = Array.isArray(detail.colors) ? detail.colors : [];
      imageUrl = detail.imageUrl || '';
      productDescription = detail.productDescription || '';
      tempMeasurementChart = detail.tempMeasurementChart || '';
      devComments = detail.devComments || '';
      mastersComments = detail.mastersComments || '';
      reworkNotes = detail.reworkNotes || '';
      sampleImageUrl = detail.sampleImageUrl || '';
      comments = detail.comments || '';
      sampleStatus = detail.status || 'DRAFT';

      sizes = detail.sizes && detail.sizes.length > 0 ? detail.sizes : [{ id: '', gender: 'Men', size: 'M', qty: 1 }];
      fabrics = detail.fabrics && detail.fabrics.length > 0
        ? detail.fabrics.map(f => ({ ...f, attrs: { ...defaultFabricAttrs(), ...(f.attrs || {}) } }))
        : [{ id: '', fabricName: '', lotNo: '', average: 0.35, attrs: defaultFabricAttrs() }];
      trims = detail.trims && detail.trims.length > 0
        ? detail.trims.map(t => ({ ...t, attrs: { ...defaultTrimAttrs(t.attrs?.trimType || 'Other Trims'), ...(t.attrs || {}) } }))
        : [{ id: '', trimName: '', average: 3.0, attrs: defaultTrimAttrs('Button') }];
      zips = detail.zips && detail.zips.length > 0
        ? detail.zips.map(z => ({ ...z, attrs: { ...defaultZipAttrs(), ...(z.attrs || {}), zipType: z.attrs?.zipType || z.zipName || '' } }))
        : [{ id: '', zipName: '', lengthAndQty: '', attrs: defaultZipAttrs() }];
      labels = detail.labels && detail.labels.length > 0
        ? detail.labels.map(l => ({ ...l, attrs: { ...defaultLabelAttrs(), ...(l.attrs || {}) } }))
        : [{ id: '', labelName: '', attrs: defaultLabelAttrs() }];
      pharmas = detail.pharmas && detail.pharmas.length > 0 ? detail.pharmas : [{ id: '', basePharmaNo: '', changesMade: '', newPharmaNo: '' }];

      if (window.showToast) {
        window.showToast(`Loaded ${detail.developmentNo}`, 'info');
      }
    } catch (err) {
      alert('Failed to load sample: ' + err.message);
    }
  }

  function resetSampleForm() {
    sampleId = '';
    developmentDate = new Date().toISOString().split('T')[0];
    clientId = '';
    clientName = '';
    sampleCategory = 'Apparel';
    productGenre = 'Polo';
    tempProductName = '';
    sampleColors = [];
    imageUrl = '';
    productDescription = '';
    sizes = [{ id: '', gender: 'Men', size: 'M', qty: 1 }];
    fabrics = [{ id: '', fabricName: '', lotNo: '', average: 0.35, attrs: defaultFabricAttrs() }];
    trims = [{ id: '', trimName: '', average: 3.0, attrs: defaultTrimAttrs('Button') }];
    zips = [{ id: '', zipName: '', lengthAndQty: '', attrs: defaultZipAttrs() }];
    labels = [{ id: '', labelName: '', attrs: defaultLabelAttrs() }];
    pharmas = [{ id: '', basePharmaNo: '', changesMade: '', newPharmaNo: '' }];
    tempMeasurementChart = '';
    devComments = '';
    mastersComments = '';
    reworkNotes = '';
    sampleImageUrl = '';
    comments = '';
    sampleStatus = 'DRAFT';
    refreshNextDevNo();
  }

  function setGarmentSpecsFromString(str) {
    const raw = String(str || '').trim();
    if (!raw) {
      selectedGarmentSpecs = [];
      garmentSpecifications = '';
      return;
    }
    const parts = raw.includes(' | ')
      ? raw.split(/\s*\|\s*/).map(s => s.trim()).filter(Boolean)
      : [raw];
    selectedGarmentSpecs = Array.from(new Set(parts));
    garmentSpecifications = selectedGarmentSpecs.join(' | ');
  }

  function toggleGarmentSpec(spec) {
    if (selectedGarmentSpecs.includes(spec)) {
      selectedGarmentSpecs = selectedGarmentSpecs.filter(s => s !== spec);
    } else {
      selectedGarmentSpecs = [...selectedGarmentSpecs, spec];
    }
    garmentSpecifications = selectedGarmentSpecs.join(' | ');
  }

  function addCustomGarmentSpec() {
    const clean = (customGarmentSpecInput || '').trim();
    if (!clean) return;
    if (!selectedGarmentSpecs.includes(clean)) {
      selectedGarmentSpecs = [...selectedGarmentSpecs, clean];
    }
    garmentSpecifications = selectedGarmentSpecs.join(' | ');
    customGarmentSpecInput = '';
  }

  // Convert Sample to Master Product
  function convertSampleToProduct() {
    prodGenre = productGenre || 'Polo';
    prodName = tempProductName || '';
    selectedSampleToPull = sampleId;

    // Commercial & Attribute fields (CSV lines 51-54)
    whyWeMadeThis = productDescription || devComments || whyWeMadeThis || '';

    // Infer composition from primary fabric data (prefer attrs.composition)
    if (fabrics && fabrics.length > 0) {
      if (fabrics[0]?.attrs?.composition) {
        composition = fabrics[0].attrs.composition;
      } else if (fabrics[0].fabricName) {
        const fNameLower = fabrics[0].fabricName.toLowerCase();
        const matched = authoritativeCompositions.find(c => fNameLower.includes(c.toLowerCase()));
        composition = matched || fabrics[0].fabricName;
      }
    }

    // Infer garment specifications (multi-select initialized)
    setGarmentSpecsFromString(standardGarmentSpecsMap[productGenre] || standardGarmentSpecsList[0]);

    // Infer brand label from sample labels
    if (labels && labels.length > 0) {
      const brandObj = (masters.label_brands || []).find(b =>
        labels.some(l => (l.labelName || '').toLowerCase().includes((b.display_name || b.name || '').toLowerCase()))
      );
      if (brandObj) {
        brandLabel = brandObj.display_name || brandObj.name;
      }
    }

    if (!mrp || mrp <= 0) {
      mrp = 1499.0;
    }

    const baseImgs = [sampleImageUrl, imageUrl].filter(Boolean);
    const samplePharmaList = Array.from(new Set(
      (pharmas || []).flatMap(p => [p.newPharmaNo, p.basePharmaNo]).map(v => (v || '').trim()).filter(Boolean)
    ));
    const initPharmas = samplePharmaList.length > 0 ? samplePharmaList : [''];
    const plannedColorVariants = (sampleColors || []).map(c => ({ colorName: c, images: [] }));

    if (isNonApparel) {
      const chosenSizes = Array.from(new Set((sizes || []).map(sz => sz.size).filter(Boolean)));
      const validScale = getSizesForSampling('N/A', prodGenre);
      prodGenders = [{
        id: '',
        gender: 'N/A',
        sizes: chosenSizes.length > 0 ? chosenSizes : validScale.slice(0, 3),
        pharmaNo: initPharmas[0] || '',
        pharmaNos: [...initPharmas],
        measurementChartUrl: tempMeasurementChart || '',
        productImages: baseImgs,
        colorImages: plannedColorVariants.map(c => ({ ...c, images: [] }))
      }];
    } else {
      // Convert Genders + exact sizes chosen in Sample Development
      const uniqueGenders = [...new Set(sizes.map(s => s.gender).filter(g => g && g !== 'N/A'))];
      const targetGenders = uniqueGenders.length > 0 ? uniqueGenders : ['Men'];
      prodGenders = targetGenders.map(g => {
        const gSizes = Array.from(new Set((sizes || []).filter(sz => sz.gender === g).map(sz => sz.size).filter(Boolean)));
        return {
          id: '',
          gender: g,
          sizes: gSizes.length > 0 ? gSizes : ['S', 'M', 'L', 'XL'],
          pharmaNo: initPharmas[0] || '',
          pharmaNos: [...initPharmas],
          measurementChartUrl: tempMeasurementChart || '',
          productImages: [...baseImgs],
          colorImages: plannedColorVariants.map(c => ({ ...c, images: [] }))
        };
      });
    }

    // Convert Materials
    const newMaterials = [];
    fabrics.forEach(f => {
      if (f.fabricName) {
        newMaterials.push({
          id: '',
          materialType: 'fabric',
          name: f.fabricName,
          average: f.average || (isNonApparel ? 1.0 : 0.35),
          gsm: isNonApparel ? 0 : (Number(f.attrs?.gsm) || 220),
          dia: isNonApparel ? 0 : (parseFloat(String(f.attrs?.dia || '').replace(/[^0-9.]/g, '')) || 32),
          zipLengthQty: ''
        });
      }
    });
    trims.forEach(t => {
      if (t.trimName) {
        newMaterials.push({
          id: '',
          materialType: 'trim',
          name: t.trimName,
          average: t.average || 1.0,
          gsm: 0,
          dia: 0,
          zipLengthQty: ''
        });
      }
    });
    zips.forEach(z => {
      if (z.zipName) {
        newMaterials.push({
          id: '',
          materialType: 'zip',
          name: z.zipName,
          average: 1.0,
          gsm: 0,
          dia: 0,
          zipLengthQty: z.lengthAndQty || ''
        });
      }
    });
    if (newMaterials.length > 0) {
      prodMaterials = newMaterials;
    }

    refreshSKU();
    activeTab = 'product';
    if (window.showToast) {
      window.showToast('Sample specs & attributes transferred to Product Creation!', 'success');
    }
  }

  // ============================================================
  // TAB 2 HANDLERS: PRODUCT CREATION
  // ============================================================
  async function handlePullFromSampleChange(e) {
    const sId = e.target.value;
    if (!sId) return;
    try {
      const s = await api.getSampleDevelopment(sId);
      prodGenre = s.productGenre || 'Polo';
      prodName = s.tempProductName || '';
      selectedSampleToPull = s.id;

      // 1. Pull Product Story (Voice note / dev comments)
      whyWeMadeThis = s.productDescription || s.devComments || whyWeMadeThis || '';

      // 2. Infer Composition from sample primary fabric data
      if (s.fabrics && s.fabrics.length > 0) {
        if (s.fabrics[0]?.attrs?.composition) {
          composition = s.fabrics[0].attrs.composition;
        } else if (s.fabrics[0].fabricName) {
          const fNameLower = s.fabrics[0].fabricName.toLowerCase();
          const matched = authoritativeCompositions.find(c => fNameLower.includes(c.toLowerCase()));
          composition = matched || s.fabrics[0].fabricName;
        }
      }

      // 3. Infer Garment Specifications (multi-select)
      setGarmentSpecsFromString(standardGarmentSpecsMap[prodGenre] || standardGarmentSpecsList[0]);

      // 4. Infer Brand Label from sample labels
      if (s.labels && s.labels.length > 0) {
        const brandObj = (masters.label_brands || []).find(b =>
          s.labels.some(l => (l.labelName || '').toLowerCase().includes((b.display_name || b.name || '').toLowerCase()))
        );
        if (brandObj) {
          brandLabel = brandObj.display_name || brandObj.name;
        }
      }

      // 5. Default MRP
      if (!mrp || mrp <= 0) {
        mrp = 1499.0;
      }

      const pullIsNonApparel = nonApparelGenres.includes(prodGenre);
      const baseImgs = [s.sampleImageUrl, s.imageUrl].filter(Boolean);
      const pullPharmaList = Array.from(new Set(
        (s.pharmas || []).flatMap(p => [p.newPharmaNo, p.basePharmaNo]).map(v => (v || '').trim()).filter(Boolean)
      ));
      const initPharmas = pullPharmaList.length > 0 ? pullPharmaList : [''];
      const sColors = Array.isArray(s.colors) ? s.colors : [];
      if (sColors.length > 0) {
        sampleColors = [...sColors];
      }
      const plannedColorVariants = sColors.map(c => ({ colorName: c, images: [] }));

      if (pullIsNonApparel) {
        const chosenSizes = Array.from(new Set((s.sizes || []).map(sz => sz.size).filter(Boolean)));
        const validScale = getSizesForSampling('N/A', prodGenre);
        prodGenders = [{
          id: '',
          gender: 'N/A',
          sizes: chosenSizes.length > 0 ? chosenSizes : validScale.slice(0, 3),
          pharmaNo: initPharmas[0] || '',
          pharmaNos: [...initPharmas],
          measurementChartUrl: s.tempMeasurementChart || '',
          productImages: baseImgs,
          colorImages: plannedColorVariants.map(c => ({ ...c, images: [] }))
        }];
      } else {
        const uniqueGenders = [...new Set((s.sizes || []).map(sz => sz.gender).filter(g => g && g !== 'N/A'))];
        const targetGenders = uniqueGenders.length > 0 ? uniqueGenders : ['Men'];
        prodGenders = targetGenders.map(g => {
          const gSizes = Array.from(new Set((s.sizes || []).filter(sz => sz.gender === g).map(sz => sz.size).filter(Boolean)));
          return {
            id: '',
            gender: g,
            sizes: gSizes.length > 0 ? gSizes : ['S', 'M', 'L', 'XL'],
            pharmaNo: initPharmas[0] || '',
            pharmaNos: [...initPharmas],
            measurementChartUrl: s.tempMeasurementChart || '',
            productImages: [...baseImgs],
            colorImages: plannedColorVariants.map(c => ({ ...c, images: [] }))
          };
        });
      }

      const mats = [];
      (s.fabrics || []).forEach(f => {
        if (f.fabricName) {
          mats.push({
            id: '',
            materialType: 'fabric',
            name: f.fabricName,
            average: f.average || (pullIsNonApparel ? 1.0 : 0.35),
            gsm: 220,
            dia: 32,
            zipLengthQty: ''
          });
        }
      });
      (s.trims || []).forEach(t => {
        if (t.trimName) {
          mats.push({
            id: '',
            materialType: 'trim',
            name: t.trimName,
            average: t.average || 1.0,
            gsm: 0,
            dia: 0,
            zipLengthQty: ''
          });
        }
      });
      (s.zips || []).forEach(z => {
        if (z.zipName) {
          mats.push({
            id: '',
            materialType: 'zip',
            name: z.zipName,
            average: 1.0,
            gsm: 0,
            dia: 0,
            zipLengthQty: z.lengthAndQty || ''
          });
        }
      });
      if (mats.length > 0) {
        prodMaterials = mats;
      }

      refreshSKU();
      if (window.showToast) {
        window.showToast(`Imported all specs & product attributes from sample ${s.developmentNo}`, 'info');
      }
    } catch (err) {
      alert('Failed to pull from sample: ' + err.message);
    }
  }

  function addProductGender() {
    const plannedColorVariants = (sampleColors || []).map(c => ({ colorName: c, images: [] }));
    prodGenders = [...prodGenders, {
      id: '',
      gender: 'Women',
      sizes: ['S', 'M', 'L', 'XL'],
      pharmaNo: '',
      pharmaNos: [''],
      measurementChartUrl: '',
      productImages: [],
      colorImages: plannedColorVariants
    }];
  }
  function removeProductGender(idx) {
    prodGenders = prodGenders.filter((_, i) => i !== idx);
  }

  function toggleProductSize(gIdx, sz) {
    const g = prodGenders[gIdx];
    const current = Array.isArray(g.sizes) ? g.sizes : [];
    if (current.includes(sz)) {
      g.sizes = current.filter(s => s !== sz);
    } else {
      g.sizes = [...current, sz];
    }
    prodGenders = [...prodGenders];
  }

  function selectAllProductSizes(gIdx) {
    const g = prodGenders[gIdx];
    g.sizes = [...getSizesForSampling(g.gender, prodGenre)];
    prodGenders = [...prodGenders];
  }

  function clearProductSizes(gIdx) {
    prodGenders[gIdx].sizes = [];
    prodGenders = [...prodGenders];
  }

  function addGenderPharmaNo(gIdx) {
    const g = prodGenders[gIdx];
    if (!Array.isArray(g.pharmaNos)) {
      g.pharmaNos = g.pharmaNo ? [g.pharmaNo] : [];
    }
    g.pharmaNos = [...g.pharmaNos, ''];
    prodGenders = [...prodGenders];
  }

  function removeGenderPharmaNo(gIdx, pIdx) {
    const g = prodGenders[gIdx];
    const next = (g.pharmaNos || []).filter((_, i) => i !== pIdx);
    g.pharmaNos = next.length > 0 ? next : [''];
    g.pharmaNo = g.pharmaNos[0] || '';
    prodGenders = [...prodGenders];
  }

  function syncPlannedColorsToGender(gIdx) {
    const g = prodGenders[gIdx];
    if (!g.colorImages) g.colorImages = [];
    for (const c of sampleColors) {
      if (!g.colorImages.some(cv => cv.colorName.toLowerCase() === c.toLowerCase())) {
        g.colorImages.push({ colorName: c, images: [] });
      }
    }
    prodGenders = [...prodGenders];
  }

  // Image & Color Variant Helpers for Product Creation (supports direct multi-file upload + URL)
  async function handleProductImagesFileUpload(e, gIdx) {
    const files = Array.from(e.target.files || []);
    if (files.length === 0) return;
    uploadingTarget = `prod_${gIdx}`;
    try {
      const uploaded = [];
      for (const file of files) {
        const res = await api.uploadFile(file);
        if (res?.url) uploaded.push(res.url);
      }
      if (!prodGenders[gIdx].productImages) prodGenders[gIdx].productImages = [];
      prodGenders[gIdx].productImages = [...prodGenders[gIdx].productImages, ...uploaded];
      prodGenders = [...prodGenders];
      if (window.showToast && uploaded.length > 0) {
        window.showToast(`Uploaded ${uploaded.length} product image(s)`, 'success');
      }
    } catch (err) {
      if (window.showToast) window.showToast('Image upload failed: ' + err.message, 'error');
    } finally {
      uploadingTarget = '';
      e.target.value = '';
    }
  }

  async function handleColorImagesFileUpload(e, gIdx, colorIdx) {
    const files = Array.from(e.target.files || []);
    if (files.length === 0) return;
    uploadingTarget = `col_${gIdx}_${colorIdx}`;
    try {
      const uploaded = [];
      for (const file of files) {
        const res = await api.uploadFile(file);
        if (res?.url) uploaded.push(res.url);
      }
      const variant = prodGenders[gIdx].colorImages[colorIdx];
      if (!variant.images) variant.images = [];
      variant.images = [...variant.images, ...uploaded];
      prodGenders = [...prodGenders];
      if (window.showToast && uploaded.length > 0) {
        window.showToast(`Uploaded ${uploaded.length} image(s) for ${prodGenders[gIdx].gender} - ${variant.colorName}`, 'success');
      }
    } catch (err) {
      if (window.showToast) window.showToast('Color image upload failed: ' + err.message, 'error');
    } finally {
      uploadingTarget = '';
      e.target.value = '';
    }
  }

  function addGenderProductImage(gIdx) {
    const url = (newImgUrlInput[gIdx] || '').trim();
    if (!url) return;
    if (!prodGenders[gIdx].productImages) prodGenders[gIdx].productImages = [];
    prodGenders[gIdx].productImages = [...prodGenders[gIdx].productImages, url];
    newImgUrlInput[gIdx] = '';
  }

  function removeGenderProductImage(gIdx, imgIdx) {
    prodGenders[gIdx].productImages = prodGenders[gIdx].productImages.filter((_, i) => i !== imgIdx);
  }

  function addGenderColorVariant(gIdx, explicitVal) {
    const raw = explicitVal !== undefined
      ? (typeof explicitVal === 'object' ? (explicitVal?.display_name || explicitVal?.code || '') : explicitVal)
      : (newColorNameInput[gIdx] || '');
    const name = String(raw || '').trim();
    if (!name) return;
    if (!prodGenders[gIdx].colorImages) prodGenders[gIdx].colorImages = [];
    if (!prodGenders[gIdx].colorImages.some(cv => cv.colorName.toLowerCase() === name.toLowerCase())) {
      prodGenders[gIdx].colorImages = [...prodGenders[gIdx].colorImages, {
        colorName: name,
        images: []
      }];
    }
    newColorNameInput[gIdx] = '';
  }

  function removeGenderColorVariant(gIdx, colorIdx) {
    prodGenders[gIdx].colorImages = prodGenders[gIdx].colorImages.filter((_, i) => i !== colorIdx);
  }

  function addColorImage(gIdx, colorIdx) {
    const key = `${gIdx}_${colorIdx}`;
    const url = (newColorImgInput[key] || '').trim();
    if (!url) return;
    const variant = prodGenders[gIdx].colorImages[colorIdx];
    if (!variant.images) variant.images = [];
    variant.images = [...variant.images, url];
    newColorImgInput[key] = '';
    prodGenders = [...prodGenders];
  }

  function removeColorImage(gIdx, colorIdx, imgIdx) {
    prodGenders[gIdx].colorImages[colorIdx].images = prodGenders[gIdx].colorImages[colorIdx].images.filter((_, i) => i !== imgIdx);
    prodGenders = [...prodGenders];
  }

  // Load an existing published master product to edit all attribute fields
  async function selectProductToEdit(productId) {
    if (!productId) return;
    try {
      const p = await api.getMasterProduct(productId);
      prodId = p.id;
      prodName = p.productName || '';
      prodGenre = p.genre || 'Polo';
      skuCode = p.skuCode || '';
      brandLabel = p.brandLabel || '';
      whyWeMadeThis = p.whyWeMadeThis || '';
      setGarmentSpecsFromString(p.garmentSpecifications || standardGarmentSpecsMap[prodGenre] || '');
      composition = p.composition || '';
      mrp = p.mrp || 1499.0;
      selectedSampleToPull = p.sampleDevelopmentId || '';
      if (p.genders && p.genders.length > 0) {
        prodGenders = p.genders.map(g => {
          const pNos = (Array.isArray(g.pharmaNos) && g.pharmaNos.length > 0)
            ? g.pharmaNos
            : (g.pharmaNo ? [g.pharmaNo] : ['']);
          return {
            id: g.id || '',
            gender: g.gender || 'Men',
            sizes: Array.isArray(g.sizes) ? g.sizes : [],
            pharmaNo: pNos[0] || '',
            pharmaNos: pNos,
            measurementChartUrl: g.measurementChartUrl || '',
            productImages: g.productImages || [],
            colorImages: (g.colorImages || []).map(cv => ({
              colorName: cv.colorName || cv.color || '',
              images: Array.isArray(cv.images) ? cv.images : []
            }))
          };
        });
      }
      if (p.materials && p.materials.length > 0) {
        prodMaterials = p.materials.map(m => ({
          id: m.id || '',
          materialType: m.materialType || 'fabric',
          name: m.name || '',
          average: m.average || 0.35,
          gsm: m.gsm || 0,
          dia: m.dia || 0,
          zipLengthQty: m.zipLengthQty || ''
        }));
      }
      if (window.showToast) {
        window.showToast(`Loaded master product "${p.productName}" for editing`, 'info');
      }
    } catch (err) {
      alert('Failed to load master product: ' + err.message);
    }
  }

  function resetProductForm() {
    prodId = '';
    selectedSampleToPull = '';
    creationDate = new Date().toISOString().split('T')[0];
    prodGenre = 'Polo';
    prodName = '';
    brandLabel = '';
    skuCode = '';
    whyWeMadeThis = '';
    setGarmentSpecsFromString(standardGarmentSpecsMap['Polo']);
    composition = '100%Cot';
    mrp = 1499.0;
    prodGenders = [{
      id: '',
      gender: 'Men',
      sizes: ['S', 'M', 'L', 'XL'],
      pharmaNo: '',
      pharmaNos: [''],
      measurementChartUrl: '',
      productImages: [],
      colorImages: []
    }];
    prodMaterials = [
      { id: '', materialType: 'fabric', name: '', average: 0.35, gsm: 220, dia: 32, zipLengthQty: '' },
      { id: '', materialType: 'trim', name: '', average: 3.0, gsm: 0, dia: 0, zipLengthQty: '' }
    ];
    if (window.showToast) {
      window.showToast('Product form reset for new product creation', 'info');
    }
  }

  function addProductMaterial(type) {
    prodMaterials = [...prodMaterials, {
      id: '',
      materialType: type || 'fabric',
      name: '',
      average: type === 'fabric' ? 0.35 : 1.0,
      gsm: type === 'fabric' ? 220 : 0,
      dia: type === 'fabric' ? 32 : 0,
      zipLengthQty: ''
    }];
  }
  function removeProductMaterial(idx) {
    prodMaterials = prodMaterials.filter((_, i) => i !== idx);
  }

  async function publishMasterProduct() {
    if (!prodName) {
      alert('Product Name is required');
      return;
    }

    try {
      isPublishingProduct = true;
      if (!skuCode) {
        await refreshSKU();
      }
      const normalizedSpecs = selectedGarmentSpecs.length > 0
        ? selectedGarmentSpecs.join(' | ')
        : garmentSpecifications;
      const normalizedGenders = prodGenders.map(g => {
        const cleanPharmas = (g.pharmaNos || []).map(p => (p || '').trim()).filter(Boolean);
        return {
          ...g,
          sizes: Array.isArray(g.sizes) ? g.sizes : [],
          pharmaNos: cleanPharmas,
          pharmaNo: cleanPharmas[0] || (g.pharmaNo || '').trim()
        };
      });
      const payload = {
        id: prodId,
        productName: prodName,
        genre: prodGenre,
        skuCode: skuCode,
        brandLabel: brandLabel,
        whyWeMadeThis: whyWeMadeThis,
        garmentSpecifications: normalizedSpecs,
        composition: composition,
        mrp: Number(mrp) || 0.0,
        sampleDevelopmentId: selectedSampleToPull,
        genders: normalizedGenders,
        materials: prodMaterials
      };

      const res = await api.createMasterProduct(payload);
      prodId = res.id;
      if (window.showToast) {
        window.showToast(`Master Product "${res.productName}" published to factory catalog!`, 'success');
      }
      await loadProductList();
    } catch (err) {
      alert('Failed to create master product: ' + err.message);
    } finally {
      isPublishingProduct = false;
    }
  }

  // ============================================================
  // TAB 3 HANDLERS: CLIENT SPECIFIC PRICING
  // ============================================================
  function handlePricingClientSelect(c) {
    if (!c) return;
    cpClientId = c.id;
    cpClientName = c.name;
  }

  function handlePricingProductSelect(p) {
    if (!p) return;
    cpProductId = p.id;
    cpProductName = p.display_name || p.product_name || p.name || '';
  }

  async function saveClientPricing() {
    if (!cpClientName) {
      alert('Please select a Client');
      return;
    }
    if (!cpProductName) {
      alert('Please select a Product');
      return;
    }
    if (cpContractedPrice <= 0) {
      alert('Please enter a valid contracted price');
      return;
    }

    try {
      isSavingPricing = true;
      const payload = {
        clientId: cpClientId || cpClientName,
        clientName: cpClientName,
        productId: cpProductId || cpProductName,
        productName: cpProductName,
        clientImageUrl: cpClientImageUrl,
        contractedPrice: Number(cpContractedPrice) || 0,
        comments: cpComments,
        effectiveDate: cpEffectiveDate
      };
      await api.setClientProductPricing(payload);
      if (window.showToast) {
        window.showToast(`Client pricing agreement saved for ${cpClientName}`, 'success');
      }
      await loadClientPricingList();
      cpContractedPrice = 0.0;
      cpComments = '';
      cpClientImageUrl = '';
    } catch (err) {
      alert('Failed to save client pricing: ' + err.message);
    } finally {
      isSavingPricing = false;
    }
  }
</script>

<div class="station-container">
  <!-- STATION HEADER & TAB NAVIGATION -->
  <header class="station-header">
    <div class="title-group">
      <div class="station-pill">STATION: PRODUCT ENGINEERING</div>
      <h1 class="station-title">Sample Development & Product Creation</h1>
      <p class="station-desc">From initial sampling sketch & voice dictation to approved master catalog & client pricing</p>
    </div>

    <div class="tab-switcher">
      <button
        class="tab-btn {activeTab === 'sample' ? 'active' : ''}"
        onclick={() => activeTab = 'sample'}
      >
        <span class="tab-icon">🧪</span>
        <span>1. Sample Development</span>
      </button>

      <button
        class="tab-btn {activeTab === 'product' ? 'active' : ''}"
        onclick={() => activeTab = 'product'}
      >
        <span class="tab-icon">🏷️</span>
        <span>2. Product Creation (Catalog)</span>
      </button>

      <button
        class="tab-btn {activeTab === 'client-pricing' ? 'active' : ''}"
        onclick={() => activeTab = 'client-pricing'}
      >
        <span class="tab-icon">💼</span>
        <span>3. Client Specific Pricing</span>
      </button>
    </div>
  </header>

  <!-- ============================================================ -->
  <!-- TAB 1: SAMPLE DEVELOPMENT (CREATE SAMPLE)                     -->
  <!-- ============================================================ -->
  {#if activeTab === 'sample'}
    <div class="split-view">
      <!-- MAIN FORM -->
      <div class="form-col">
        <!-- 1. SINGLE FIELDS CARD -->
        <section class="card">
          <div class="card-header">
            <span class="badge">Header Specs</span>
            <h2 class="card-title">Development Request Details</h2>
            <div class="status-chip {sampleStatus.toLowerCase()}">{sampleStatus}</div>
          </div>

          <div class="form-grid">
            <div class="field-group">
              <label for="devNo">Development No</label>
              <div class="input-with-action">
                <input id="devNo" type="text" class="input" bind:value={developmentNo} readonly />
                <button type="button" class="btn btn-secondary btn-xs" onclick={refreshNextDevNo} title="Generate next sequence">
                  🔄
                </button>
              </div>
            </div>

            <div class="field-group">
              <label for="devDate">Date</label>
              <input id="devDate" type="date" class="input" bind:value={developmentDate} />
            </div>

            <div class="field-group span-2">
              <label for="clientSelect">Client</label>
              <Combobox
                items={masters.clients || []}
                bind:value={clientName}
                displayKey="name"
                valueKey="name"
                placeholder="Search 106 verified corporate clients..."
                onSelect={handleClientSelect}
              />
            </div>

            <div class="field-group">
              <label for="sampleCategory">Product Category</label>
              <select
                id="sampleCategory"
                class="input"
                bind:value={sampleCategory}
                onchange={(e) => handleSampleCategoryChange(e.currentTarget.value)}
              >
                <option value="Apparel">Apparel</option>
                <option value="Non Apparel">Non Apparel</option>
              </select>
            </div>

            <div class="field-group">
              <label for="productGenre">
                {sampleCategory === 'Apparel' ? 'Apparel Genre' : 'Non-Apparel Genre'}
              </label>
              <select
                id="productGenre"
                class="input"
                bind:value={productGenre}
                onchange={() => handleProductGenreChange(productGenre)}
              >
                {#if sampleCategory === 'Apparel'}
                  {#each apparelGenres as g}
                    <option value={g}>{g}</option>
                  {/each}
                {:else}
                  {#each nonApparelGenres as g}
                    <option value={g}>{g}</option>
                  {/each}
                {/if}
              </select>
            </div>

            <div class="field-group span-2">
              <label for="tempName">Temp Product Name</label>
              <input
                id="tempName"
                type="text"
                class="input"
                placeholder="e.g. AeroCool Poly Pique Contrast Polo"
                bind:value={tempProductName}
              />
            </div>

            <div class="field-group span-2">
              <span class="field-label" style="font-size:12px;font-weight:700;color:var(--text-secondary);">Planned Colorway(s) (Multi-Color Excel Master)</span>
              <div class="img-input-row" style="margin-bottom: 6px;">
                <div style="flex: 1;">
                  <Combobox
                    items={masters.colors || []}
                    bind:value={newSampleColor}
                    category="color"
                    displayKey="display_name"
                    valueKey="display_name"
                    placeholder="Select color from 486 master colors (Navy, Black, White...)..."
                    onSelect={(item) => addSampleColor(item)}
                  />
                </div>
                <button type="button" class="btn btn-secondary btn-xs" onclick={() => addSampleColor(newSampleColor)}>
                  + Add Color
                </button>
              </div>
              {#if sampleColors.length > 0}
                <div class="chips-wrap">
                  {#each sampleColors as col, cIdx}
                    <span class="spec-chip active">
                      🎨 {col}
                      <button type="button" class="chip-x" onclick={() => removeSampleColor(cIdx)} aria-label="Remove color {col}">✕</button>
                    </span>
                  {/each}
                </div>
              {/if}
            </div>
          </div>

          <!-- REFERENCE IMAGE UPLOADER -->
          <div class="media-row">
            <div class="uploader-wrap">
              <ProductImageUploader
                bind:imageUrl={imageUrl}
                label="Reference Sketch / Concept Image"
              />
            </div>
            <div class="voice-wrap">
              <VoiceNoteRecorder
                bind:text={productDescription}
                label="Product Description (Voice Note & Dictation)"
                placeholder="Dictate voice notes or type concept details, styling, fit preferences..."
                rows={4}
              />
            </div>
          </div>
        </section>

        <!-- 2. MULTI-ROW BREAKDOWNS CARD -->
        <section class="card">
          <div class="card-header">
            <span class="badge">Multi Breakdowns</span>
            <h2 class="card-title">Sampling Materials & Sizing Breakdown</h2>
          </div>

          <!-- A. GENDER, SIZE & QTY -->
          <div class="sub-table-block">
            <div class="sub-table-header">
              <span class="sub-title">1. {isNonApparel ? 'Capacity / Size & Quantity Breakdown (Non-Apparel)' : 'Gender, Size & Quantity Matrix (Apparel)'}</span>
              <button type="button" class="btn btn-secondary btn-xs" onclick={addSizeRow}>+ Add Size Row</button>
            </div>
            <table class="data-table">
              <thead>
                <tr>
                  {#if !isNonApparel}
                    <th style="width: 25%;">Gender / Demographic</th>
                  {/if}
                  <th style={isNonApparel ? "width: 60%;" : "width: 35%;"}>
                    {isNonApparel ? 'Size / Specification' : (isBottomwear ? 'Size (Alpha + Waist 32-52)' : 'Size Scale')}
                  </th>
                  <th style="width: 25%;">Qty (Pcs)</th>
                  <th style="width: 15%;"></th>
                </tr>
              </thead>
              <tbody>
                {#each sizes as sz, idx}
                  <tr>
                    {#if !isNonApparel}
                      <td>
                        <select
                          class="input table-input"
                          bind:value={sz.gender}
                          onchange={() => {
                            const valid = getSizesForSampling(sz.gender, productGenre);
                            if (!valid.includes(sz.size)) {
                              sz.size = valid[0];
                            }
                          }}
                        >
                          {#each canonicalGenders as g}
                            <option value={g}>{g}</option>
                          {/each}
                        </select>
                      </td>
                    {/if}
                    <td>
                      <select class="input table-input" bind:value={sz.size}>
                        {#each getSizesForSampling(sz.gender, productGenre) as s}
                          <option value={s}>{s}</option>
                        {/each}
                      </select>
                    </td>
                    <td>
                      <input type="number" class="input table-input" min="1" bind:value={sz.qty} />
                    </td>
                    <td class="text-right">
                      <button type="button" class="del-btn" onclick={() => removeSizeRow(idx)}>✕</button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>

          <!-- GENRE-DRIVEN MATERIAL & BOM RECOMMENDATION BANNER -->
          {#if genreBOMRecommendations[productGenre]}
            <div class="genre-recommendation-banner">
              <div class="genre-rec-header">
                <span class="genre-rec-badge">⚡ {productGenre} Specification Profile</span>
                <span class="genre-rec-text">
                  Recommended Avg: <strong>{genreBOMRecommendations[productGenre].avgFabric} {isNonApparel ? 'unit' : 'kg/pc'}</strong> ({genreBOMRecommendations[productGenre].suggestedFabric})
                </span>
              </div>
              <button
                type="button"
                class="btn btn-secondary btn-xs"
                onclick={() => applyGenreBOMPreset(productGenre)}
                title="Pre-populate standard fabric average, thread tex, and trims for {productGenre}"
              >
                ✨ Preload {productGenre} Materials & Trims
              </button>
            </div>
          {/if}

          <!-- B. FABRICS & LOTS -->
          <div class="sub-table-block">
            <div class="sub-table-header">
              <span class="sub-title">2. {isNonApparel ? 'Base / Shell Raw Materials & Consumption' : 'Fabrics (Type, Generic, Composition, GSM, Dia, Color, Lot No & Avg)'}</span>
              <button type="button" class="btn btn-secondary btn-xs" onclick={addFabricRow}>+ Add Material</button>
            </div>
            <div class="attr-cards-list">
              {#each fabrics as fb, idx}
                {@const fAttrs = fb.attrs || (fb.attrs = defaultFabricAttrs())}
                <div class="attr-row-card">
                  <div class="attr-row-top">
                    <span class="row-index-tag">#{idx + 1} {isNonApparel ? 'Material' : 'Fabric'}</span>
                    <button type="button" class="del-btn" onclick={() => removeFabricRow(idx)} aria-label="Remove material">✕ Remove</button>
                  </div>
                  <div class="attr-row-grid">
                    {#if !isNonApparel}
                      <div class="attr-field">
                        <span class="attr-label">Fabric Type</span>
                        <select class="input table-input" bind:value={fAttrs.fabricType}>
                          {#each (masters.fabric_types && masters.fabric_types.length ? masters.fabric_types : [{display_name:'Fabric'},{display_name:'Collar'},{display_name:'Cuff'},{display_name:'Rib'}]) as ft}
                            <option value={ft.display_name}>{ft.display_name}</option>
                          {/each}
                        </select>
                      </div>
                      <div class="attr-field">
                        <span class="attr-label">Generic Name</span>
                        <Combobox
                          items={masters.fabric_generics || []}
                          bind:value={fAttrs.fabricGeneric}
                          displayKey="display_name"
                          valueKey="display_name"
                          placeholder="Pique, Sinker, Fleece..."
                        />
                      </div>
                    {/if}
                    <div class="attr-field span-2-attr">
                      <span class="attr-label">{isNonApparel ? 'Material / Specification' : 'FPCC Fabric Name'}</span>
                      <Combobox
                        items={masters.fabrics || []}
                        bind:value={fb.fabricName}
                        displayKey="display_name"
                        valueKey="display_name"
                        placeholder={isNonApparel ? "e.g. Stainless Steel 304, Polycarbonate..." : "Select FPCC Fabric (Venus, Classic Fleece...)"}
                      />
                    </div>
                    {#if !isNonApparel}
                      <div class="attr-field span-2-attr">
                        <span class="attr-label">Fiber Composition</span>
                        <Combobox
                          items={masters.compositions || []}
                          bind:value={fAttrs.composition}
                          displayKey="display_name"
                          valueKey="display_name"
                          placeholder="Select Composition (e.g. 95%Poly 5%SP)..."
                        />
                      </div>
                    {/if}
                    <div class="attr-field">
                      <span class="attr-label">Color</span>
                      <Combobox
                        items={masters.colors || []}
                        bind:value={fAttrs.color}
                        category="color"
                        displayKey="display_name"
                        valueKey="display_name"
                        placeholder="Shade / Color..."
                      />
                    </div>
                    {#if !isNonApparel}
                      <div class="attr-field">
                        <span class="attr-label">GSM</span>
                        <input type="number" class="input table-input" bind:value={fAttrs.gsm} placeholder="e.g. 200" />
                      </div>
                      <div class="attr-field">
                        <span class="attr-label">Dia</span>
                        <input type="text" class="input table-input" bind:value={fAttrs.dia} placeholder='e.g. 60"' />
                      </div>
                    {/if}
                    <div class="attr-field">
                      <span class="attr-label">{isNonApparel ? 'Grade / Batch No' : 'Lot No'}</span>
                      <input type="text" class="input table-input" bind:value={fb.lotNo} placeholder="Lot / Shade #" />
                    </div>
                    <div class="attr-field">
                      <span class="attr-label">{isNonApparel ? 'Average (per unit)' : 'Average (kg/pc)'}</span>
                      <input type="number" step="0.01" class="input table-input" bind:value={fb.average} />
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          </div>

          <!-- C. TRIMS -->
          <div class="sub-table-block">
            <div class="sub-table-header">
              <span class="sub-title">3. {isNonApparel ? 'Hardware, Accessories & Consumables' : 'Trims & Dynamic Attributes (Type, Size, Color, GSM / Tex / Cord Ends & Avg)'}</span>
              <button type="button" class="btn btn-secondary btn-xs" onclick={addTrimRow}>+ Add Trim</button>
            </div>
            <div class="attr-cards-list">
              {#each trims as tr, idx}
                {@const tAttrs = tr.attrs || (tr.attrs = defaultTrimAttrs('Elastic'))}
                {@const tType = tAttrs.trimType || 'Other Trims'}
                {@const itemOpts = trimItemOptionsForType(tType)}
                {@const sizeOpts = trimSizeOptionsForType(tType)}
                {@const colorOpts = trimColorOptionsForType(tType)}
                <div class="attr-row-card">
                  <div class="attr-row-top">
                    <span class="row-index-tag">#{idx + 1} Trim • {tType}</span>
                    <button type="button" class="del-btn" onclick={() => removeTrimRow(idx)} aria-label="Remove trim">✕ Remove</button>
                  </div>
                  <div class="attr-row-grid">
                    <div class="attr-field">
                      <span class="attr-label">Trim Type</span>
                      <select
                        class="input table-input"
                        bind:value={tAttrs.trimType}
                        onchange={() => { tAttrs.size = ''; tAttrs.cordEnd = ''; tAttrs.gsm = ''; tAttrs.length = ''; }}
                      >
                        {#each (masters.trim_types && masters.trim_types.length ? masters.trim_types : ['Elastic','Tapes','Cord','Button','Fusing','Thread','Eyelets','Stopper','Other Trims'].map(v => ({display_name: v}))) as tt}
                          <option value={tt.display_name}>{tt.display_name}</option>
                        {/each}
                      </select>
                    </div>
                    <div class="attr-field span-2-attr">
                      <span class="attr-label">{tType} Name / Specification</span>
                      <Combobox
                        items={itemOpts}
                        bind:value={tr.trimName}
                        displayKey="display_name"
                        valueKey="display_name"
                        allowCustom={true}
                        placeholder="Select {tType} item..."
                      />
                    </div>
                    <div class="attr-field">
                      <span class="attr-label">{tType === 'Thread' ? 'Ticket / Tex Size' : 'Size / Width'}</span>
                      <Combobox
                        items={sizeOpts}
                        bind:value={tAttrs.size}
                        displayKey="display_name"
                        valueKey="display_name"
                        allowCustom={true}
                        placeholder="Select size..."
                      />
                    </div>
                    {#if tType === 'Cord'}
                      <div class="attr-field">
                        <span class="attr-label">Cord Ends</span>
                        <Combobox
                          items={masters.trim_cord_ends || []}
                          bind:value={tAttrs.cordEnd}
                          displayKey="display_name"
                          valueKey="display_name"
                          placeholder="metal silver, rubber, plastic..."
                        />
                      </div>
                    {:else if tType === 'Fusing'}
                      <div class="attr-field">
                        <span class="attr-label">Fusing GSM</span>
                        <Combobox
                          items={masters.trim_fusing_gsms || []}
                          bind:value={tAttrs.gsm}
                          displayKey="display_name"
                          valueKey="display_name"
                          placeholder="e.g. 50.0, 100.0..."
                        />
                      </div>
                    {:else if tType === 'Thread'}
                      <div class="attr-field">
                        <span class="attr-label">Cone / Tube Length</span>
                        <Combobox
                          items={masters.trim_thread_lengths || []}
                          bind:value={tAttrs.length}
                          displayKey="display_name"
                          valueKey="display_name"
                          placeholder="300 mtr, 5000 mtr..."
                        />
                      </div>
                    {/if}
                    <div class="attr-field">
                      <span class="attr-label">Color</span>
                      <Combobox
                        items={colorOpts}
                        bind:value={tAttrs.color}
                        category="color"
                        displayKey="display_name"
                        valueKey="display_name"
                        allowCustom={true}
                        placeholder="Color..."
                      />
                    </div>
                    <div class="attr-field">
                      <span class="attr-label">Average Consumption</span>
                      <input type="number" step="0.1" class="input table-input" bind:value={tr.average} placeholder="e.g. 3 pcs" />
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          </div>

          <!-- D. ZIPS -->
          <div class="sub-table-block">
            <div class="sub-table-header">
              <div style="display:flex;align-items:center;gap:8px;">
                <span class="sub-title">4. Zips (Type, Teeth Size, Slider, Closure, Finish, Tape Color, Length &amp; Qty)</span>
                {#if !isZipExpected && !isNonApparel}
                  <span class="optional-badge">Optional for {productGenre}</span>
                {/if}
                {#if isNonApparel && productGenre !== 'Bags'}
                  <span class="optional-badge">Optional for Non-Apparel</span>
                {/if}
              </div>
              <button type="button" class="btn btn-secondary btn-xs" onclick={addZipRow}>+ Add Zip</button>
            </div>
            <div class="attr-cards-list">
              {#each zips as zp, idx}
                {@const zAttrs = zp.attrs || (zp.attrs = defaultZipAttrs())}
                <div class="attr-row-card">
                  <div class="attr-row-top">
                    <span class="row-index-tag">#{idx + 1} Zip</span>
                    <button type="button" class="del-btn" onclick={() => removeZipRow(idx)} aria-label="Remove zip">✕ Remove</button>
                  </div>
                  <div class="attr-row-grid">
                    <div class="attr-field">
                      <span class="attr-label">Zip Type</span>
                      <Combobox
                        items={masters.zip_types || masters.zips || []}
                        bind:value={zp.zipName}
                        displayKey="display_name"
                        valueKey="display_name"
                        onSelect={(item) => { zAttrs.zipType = item?.display_name || zp.zipName; }}
                        placeholder="Coil, Vislon, Brass..."
                      />
                    </div>
                    <div class="attr-field">
                      <span class="attr-label">Teeth Size</span>
                      <Combobox
                        items={masters.zip_sizes || []}
                        bind:value={zAttrs.teethSize}
                        displayKey="display_name"
                        valueKey="display_name"
                        allowCustom={true}
                        placeholder="5.0, 8.0, 10.0..."
                      />
                    </div>
                    <div class="attr-field">
                      <span class="attr-label">Slider</span>
                      <Combobox
                        items={masters.zip_sliders || []}
                        bind:value={zAttrs.slider}
                        displayKey="display_name"
                        valueKey="display_name"
                        allowCustom={true}
                        placeholder="DA8DR4, Metal, Rubber..."
                      />
                    </div>
                    <div class="attr-field">
                      <span class="attr-label">Closure</span>
                      <select class="input table-input" bind:value={zAttrs.closure}>
                        <option value="">-- Closure --</option>
                        {#each (masters.zip_closures && masters.zip_closures.length ? masters.zip_closures : ['open ended','closed ended','2 way','roll'].map(v => ({display_name: v}))) as zc}
                          <option value={zc.display_name}>{zc.display_name}</option>
                        {/each}
                      </select>
                    </div>
                    <div class="attr-field">
                      <span class="attr-label">Teeth Finish / Color</span>
                      <Combobox
                        items={masters.zip_finishes || []}
                        bind:value={zAttrs.finish}
                        displayKey="display_name"
                        valueKey="display_name"
                        allowCustom={true}
                        placeholder="Silver, Gold, Matt..."
                      />
                    </div>
                    <div class="attr-field">
                      <span class="attr-label">Tape Color</span>
                      <Combobox
                        items={masters.colors || []}
                        bind:value={zAttrs.tapeColor}
                        category="color"
                        displayKey="display_name"
                        valueKey="display_name"
                        allowCustom={true}
                        placeholder="Tape Color..."
                      />
                    </div>
                    <div class="attr-field span-2-attr">
                      <span class="attr-label">Size-Wise Zip Length &amp; Qty</span>
                      <input type="text" class="input table-input" bind:value={zp.lengthAndQty} placeholder="e.g. S-L: 15cm (1pc), XL: 18cm (1pc)" />
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          </div>

          <!-- E. LABELS -->
          <div class="sub-table-block">
            <div class="sub-table-header">
              <span class="sub-title">5. Labels &amp; Branding (Brand / Additional, Label Genre, Description, Base &amp; Text Color)</span>
              <button type="button" class="btn btn-secondary btn-xs" onclick={addLabelRow}>+ Add Label</button>
            </div>
            <div class="attr-cards-list">
              {#each labels as lb, idx}
                {@const lAttrs = lb.attrs || (lb.attrs = defaultLabelAttrs())}
                <div class="attr-row-card">
                  <div class="attr-row-top">
                    <span class="row-index-tag">#{idx + 1} Label • {lAttrs.labelCategory || 'Brand Label'}</span>
                    <button type="button" class="del-btn" onclick={() => removeLabelRow(idx)} aria-label="Remove label">✕ Remove</button>
                  </div>
                  <div class="attr-row-grid">
                    <div class="attr-field">
                      <span class="attr-label">Label Category</span>
                      <select class="input table-input" bind:value={lAttrs.labelCategory}>
                        <option value="Brand Label">Brand Label</option>
                        <option value="Additional Label">Additional Label</option>
                      </select>
                    </div>
                    <div class="attr-field">
                      <span class="attr-label">Label Genre (Brand &amp; Additional)</span>
                      <select class="input table-input" bind:value={lAttrs.labelGenre}>
                        {#each (masters.label_genres && masters.label_genres.length ? masters.label_genres : ['Woven','Printed','DTF','Tareaway'].map(v => ({display_name: v}))) as lg}
                          <option value={lg.display_name}>{lg.display_name}</option>
                        {/each}
                      </select>
                    </div>
                    <div class="attr-field">
                      <span class="attr-label">{lAttrs.labelCategory === 'Additional Label' ? 'Additional Label Type' : 'Brand Label Name'}</span>
                      <Combobox
                        items={lAttrs.labelCategory === 'Additional Label' ? (masters.label_additionals || []) : (masters.label_brands || [])}
                        bind:value={lb.labelName}
                        displayKey="display_name"
                        valueKey="display_name"
                        allowCustom={true}
                        placeholder={lAttrs.labelCategory === 'Additional Label' ? 'Size, washcare, COO...' : 'ETTU, Jeep, OMG...'}
                      />
                    </div>
                    <div class="attr-field span-2-attr">
                      <span class="attr-label">Label Description / Artwork Spec</span>
                      <Combobox
                        items={masters.label_descriptions || masters.labels || []}
                        bind:value={lAttrs.labelDescription}
                        displayKey="display_name"
                        valueKey="display_name"
                        allowCustom={true}
                        placeholder="Select 51+ descriptions (ETTU +Size Navy/aqua...)"
                      />
                    </div>
                    <div class="attr-field">
                      <span class="attr-label">Base Color</span>
                      <Combobox
                        items={masters.colors || []}
                        bind:value={lAttrs.baseColor}
                        category="color"
                        displayKey="display_name"
                        valueKey="display_name"
                        allowCustom={true}
                        placeholder="Base color..."
                      />
                    </div>
                    <div class="attr-field">
                      <span class="attr-label">Text Color</span>
                      <Combobox
                        items={masters.colors || []}
                        bind:value={lAttrs.textColor}
                        category="color"
                        displayKey="display_name"
                        valueKey="display_name"
                        allowCustom={true}
                        placeholder="Text / logo color..."
                      />
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          </div>

          <!-- F. TEMP PHARMA NO -->
          <div class="sub-table-block">
            <div class="sub-table-header">
              <span class="sub-title">6. {isNonApparel ? 'Technical Drawing, Tooling & Mould Reference' : 'Temp Pharma No (Base Pattern & Grade Alterations)'}</span>
              <button type="button" class="btn btn-secondary btn-xs" onclick={addPharmaRow}>+ Add Ref</button>
            </div>
            <table class="data-table">
              <thead>
                <tr>
                  <th style="width: 30%;">{isNonApparel ? 'Base Spec / Mould #' : 'Base Pharma No'}</th>
                  <th style="width: 40%;">Changes Made / Alterations</th>
                  <th style="width: 20%;">{isNonApparel ? 'New Rev / Tooling #' : 'New Pharma No'}</th>
                  <th style="width: 10%;"></th>
                </tr>
              </thead>
              <tbody>
                {#each pharmas as ph, idx}
                  <tr>
                    <td>
                      <input type="text" class="input table-input" bind:value={ph.basePharmaNo} placeholder={isNonApparel ? "e.g. MOULD-SSB-500-V1" : "e.g. PH-POLO-STD-M"} />
                    </td>
                    <td>
                      <input type="text" class="input table-input" bind:value={ph.changesMade} placeholder={isNonApparel ? "e.g. Neck diameter increased by 2mm" : "e.g. Body length +1 inch, relaxed chest"} />
                    </td>
                    <td>
                      <input type="text" class="input table-input" bind:value={ph.newPharmaNo} placeholder={isNonApparel ? "e.g. MOULD-SSB-500-REV2" : "e.g. PH-POLO-2026-REV1"} />
                    </td>
                    <td class="text-right">
                      <button type="button" class="del-btn" onclick={() => removePharmaRow(idx)}>✕</button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </section>

        <!-- 3. COMMENTS & FINISHED SAMPLE CARD -->
        <section class="card">
          <div class="card-header">
            <span class="badge">Review & Rework</span>
            <h2 class="card-title">Specs, Comments & Finished Physical Sample</h2>
          </div>

          <div class="form-grid">
            <div class="field-group span-2">
              <label for="tempMeas">Temp Measurement Chart</label>
              <textarea
                id="tempMeas"
                class="input"
                rows="3"
                placeholder="Alphanumeric specs: e.g. CHEST: 40, LENGTH: 28, SHOULDER: 18, SLEEVE: 8.5..."
                bind:value={tempMeasurementChart}
              ></textarea>
            </div>

            <div class="field-group span-2">
              <label for="devComm">Dev Comments (Merchandiser / Tech)</label>
              <textarea
                id="devComm"
                class="input"
                rows="2"
                placeholder="Dev notes: if not approved, specify adjustments for remake..."
                bind:value={devComments}
              ></textarea>
            </div>

            <div class="field-group span-2">
              <VoiceNoteRecorder
                bind:text={mastersComments}
                label="Masters Comments (Voice Transcribed Pattern/Cutting Master Notes)"
                placeholder="Dictate Pattern Master / Cutting Master feedback..."
                rows={2}
              />
            </div>

            <div class="field-group span-2">
              <label for="rework">Rework Notes (If remaking sample with edits)</label>
              <textarea
                id="rework"
                class="input"
                rows="2"
                placeholder="Rework details: changes requested by client before bulk production..."
                bind:value={reworkNotes}
              ></textarea>
            </div>

            <div class="field-group span-2">
              <ProductImageUploader
                bind:imageUrl={sampleImageUrl}
                label="Sample Image (Finished Sample Photograph)"
              />
            </div>

            <div class="field-group span-2">
              <label for="genComments">General Comments</label>
              <textarea
                id="genComments"
                class="input"
                rows="2"
                placeholder="Additional sampling notes..."
                bind:value={comments}
              ></textarea>
            </div>
          </div>

          <!-- ACTION BUTTONS -->
          <div class="action-footer">
            <button type="button" class="btn btn-secondary" onclick={resetSampleForm}>
              ✕ Reset Form
            </button>

            <button
              type="button"
              class="btn btn-secondary"
              disabled={isSavingSample}
              onclick={() => saveSampleDevelopment('DRAFT')}
            >
              💾 Save Draft
            </button>

            <button
              type="button"
              class="btn btn-warning"
              disabled={isSavingSample}
              onclick={() => saveSampleDevelopment('SUBMITTED')}
            >
              📤 Submit Sample
            </button>

            <button
              type="button"
              class="btn btn-danger"
              disabled={isSavingSample}
              onclick={() => saveSampleDevelopment('REWORK_REQUESTED')}
            >
              ⚠️ Request Rework
            </button>

            <button
              type="button"
              class="btn btn-success"
              disabled={isSavingSample}
              onclick={() => saveSampleDevelopment('APPROVED')}
            >
              ✓ Approve Sample
            </button>

            <button
              type="button"
              class="btn btn-primary"
              onclick={convertSampleToProduct}
            >
              ➡️ Convert to Master Product
            </button>
          </div>
        </section>
      </div>

      <!-- SIDEBAR: RECENT SAMPLES DRAWER -->
      <aside class="sidebar-col">
        <div class="card sidebar-card">
          <div class="sidebar-header">
            <h3>Recent Development Samples</h3>
            <button class="refresh-btn" onclick={loadSampleList} title="Refresh sample list">🔄</button>
          </div>

          <input
            type="text"
            class="input sidebar-search"
            placeholder="Search dev no, client, product..."
            bind:value={sampleSearchQuery}
          />

          <div class="sample-list-scroll">
            {#if isLoadingSamples}
              <div class="loading-state">Loading samples...</div>
            {:else}
              {#each sampleList.filter(s =>
                !sampleSearchQuery ||
                s.developmentNo?.toLowerCase().includes(sampleSearchQuery.toLowerCase()) ||
                s.clientName?.toLowerCase().includes(sampleSearchQuery.toLowerCase()) ||
                s.tempProductName?.toLowerCase().includes(sampleSearchQuery.toLowerCase())
              ) as s}
                <div
                  class="sample-card {sampleId === s.id ? 'active' : ''}"
                  onclick={() => loadSampleIntoForm(s)}
                  role="button"
                  tabindex="0"
                  onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && loadSampleIntoForm(s)}
                >
                  <div class="sc-top">
                    <span class="sc-no">{s.developmentNo}</span>
                    <span class="status-chip {s.status?.toLowerCase()}">{s.status}</span>
                  </div>
                  <div class="sc-prod">{s.tempProductName || 'Untitled Sample'}</div>
                  <div class="sc-client">{s.clientName}</div>
                  <div class="sc-meta">
                    <span>{s.developmentDate}</span>
                    <span>{s.totalQty} pcs</span>
                  </div>
                </div>
              {/each}
            {/if}
          </div>
        </div>
      </aside>
    </div>

  <!-- ============================================================ -->
  <!-- TAB 2: PRODUCT CREATION (CREATE MASTER PRODUCT)              -->
  <!-- ============================================================ -->
  {:else if activeTab === 'product'}
    <div class="single-col-container">
      <section class="card">
        <div class="card-header">
          <div style="display:flex;align-items:center;gap:10px;">
            <span class="badge">Master Specification</span>
            <h2 class="card-title">
              {prodId ? `Edit Product: ${prodName}` : 'Create / Edit Master Product & Attribute Specifications'}
            </h2>
          </div>
          <button type="button" class="btn btn-secondary btn-xs" onclick={resetProductForm}>
            + New Blank Product
          </button>
        </div>

        <!-- PULL FROM SAMPLE & LOAD EXISTING PRODUCT CONTROLS -->
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;margin-bottom:16px;">
          <!-- 1. Pull from Sample -->
          <div class="pull-box" style="margin-bottom:0;">
            <div class="pull-label">
              <span class="pull-icon">⚡</span>
              <span>Import Approved Sample:</span>
            </div>
            <select class="input pull-select" onchange={handlePullFromSampleChange}>
              <option value="">-- Select Approved Sample to Pull Specs --</option>
              {#each sampleList as smp}
                <option value={smp.id} selected={selectedSampleToPull === smp.id}>
                  {smp.developmentNo} • {smp.tempProductName} ({smp.clientName}) [{smp.status}]
                </option>
              {/each}
            </select>
          </div>

          <!-- 2. Edit Existing Master Product -->
          <div class="pull-box" style="margin-bottom:0;background:rgba(16,185,129,0.06);border-color:rgba(16,185,129,0.25);">
            <div class="pull-label" style="color:#10b981;">
              <span class="pull-icon">📝</span>
              <span>Edit Existing Product:</span>
            </div>
            <select class="input pull-select" onchange={(e) => selectProductToEdit(e.currentTarget.value)}>
              <option value="">-- Select Published Product to Edit --</option>
              {#each productList as p}
                <option value={p.id} selected={prodId === p.id}>
                  {p.skuCode || 'SKU'} • {p.productName} [{p.genre}]
                </option>
              {/each}
            </select>
          </div>
        </div>

        <div class="form-grid">
          <div class="field-group">
            <label for="prodCreateDate">Creation Date</label>
            <input id="prodCreateDate" type="date" class="input" bind:value={creationDate} />
          </div>

          <div class="field-group">
            <label for="prodGenre">Product Genre</label>
            <select
              id="prodGenre"
              class="input"
              bind:value={prodGenre}
              onchange={() => {
                refreshSKU();
                if (nonApparelGenres.includes(prodGenre)) {
                  /** @type {Record<string, any>} */
                  const prev = prodGenders[0] || {};
                  const valid = getSizesForSampling('N/A', prodGenre);
                  prodGenders = [{
                    id: prev.id || '',
                    gender: 'N/A',
                    sizes: valid.slice(0, 3),
                    pharmaNo: prev.pharmaNo || '',
                    pharmaNos: Array.isArray(prev.pharmaNos) && prev.pharmaNos.length ? prev.pharmaNos : (prev.pharmaNo ? [prev.pharmaNo] : ['']),
                    measurementChartUrl: prev.measurementChartUrl || '',
                    productImages: prev.productImages || [],
                    colorImages: prev.colorImages || []
                  }];
                } else if (prodGenders[0]?.gender === 'N/A') {
                  prodGenders = [{
                    id: '',
                    gender: 'Men',
                    sizes: ['S', 'M', 'L', 'XL'],
                    pharmaNo: '',
                    pharmaNos: [''],
                    measurementChartUrl: '',
                    productImages: [],
                    colorImages: (sampleColors || []).map(c => ({ colorName: c, images: [] }))
                  }];
                }
              }}
            >
              <optgroup label="Apparel Genres">
                {#each apparelGenres as g}
                  <option value={g}>{g}</option>
                {/each}
              </optgroup>
              <optgroup label="Non-Apparel Genres">
                {#each nonApparelGenres as g}
                  <option value={g}>{g}</option>
                {/each}
              </optgroup>
            </select>
          </div>

          <div class="field-group span-2">
            <label for="prodNameInput">Product Name</label>
            <input
              id="prodNameInput"
              type="text"
              class="input"
              placeholder="e.g. AeroCool Poly Pique Contrast Polo"
              bind:value={prodName}
              onblur={refreshSKU}
            />
          </div>

          <div class="field-group">
            <label for="brandLabelSelect">Brand Label</label>
            <Combobox
              items={masters.label_brands || []}
              bind:value={brandLabel}
              displayKey="display_name"
              valueKey="display_name"
              placeholder="Select brand label..."
              onSelect={refreshSKU}
            />
          </div>

          <div class="field-group">
            <label for="skuCode">SKU Code (Generated: ET_PO_LINEFL)</label>
            <div class="input-with-action">
              <input id="skuCode" type="text" class="input" bind:value={skuCode} placeholder="Auto-generated" />
              <button type="button" class="btn btn-secondary btn-xs" onclick={refreshSKU} title="Regenerate SKU">
                ⚡
              </button>
            </div>
          </div>
        </div>

        <!-- ============================================================ -->
        <!-- COMMERCIAL & CATALOG ATTRIBUTES (CSV LINES 51-54)            -->
        <!-- ============================================================ -->
        <div class="sub-table-block" style="margin-top: 20px;">
          <div class="sub-table-header">
            <div style="display: flex; align-items: center; gap: 8px;">
              <span class="sub-title">✨ Commercial & Catalog Specifications (Product Story, Garment Specs, Composition & MRP)</span>
              <span class="badge" style="background: rgba(59, 130, 246, 0.2); color: #60a5fa; border: 1px solid rgba(59, 130, 246, 0.4);">
                CSV Lines 51-54 • 100% Editable
              </span>
            </div>
          </div>

          <div style="padding: 16px;">
            <div class="form-grid">
              <!-- Line 51: Why We Made This -->
              <div class="field-group span-2">
                <VoiceNoteRecorder
                  bind:text={whyWeMadeThis}
                  label="Why We Made This (Product Story & Design Intent) [CSV Line 51]"
                  placeholder="Voice dictate or type why we made this product, technical performance details, fit story..."
                  rows={3}
                />
                <span class="field-hint" style="font-size: 11px; color: var(--text-tertiary, #94a3b8); margin-top: 4px; display: block;">
                  Pulls from sample voice description or dev comments. Merchandiser can freely edit or voice-record new story.
                </span>
              </div>

              <!-- Line 52: Garment Specifications (Multi-Select + Custom) -->
              <div class="field-group span-2">
                <label for="customGarmentSpecInput" style="font-weight: 700; color: #f8fafc;">
                  Garment Specifications (Multi-Select + Custom) [CSV Line 52]
                </label>
                <div class="spec-multiselect">
                  {#each standardGarmentSpecsList as specOpt}
                    {@const isSpecSelected = selectedGarmentSpecs.includes(specOpt)}
                    <button
                      type="button"
                      class="spec-chip {isSpecSelected ? 'active' : ''}"
                      onclick={() => toggleGarmentSpec(specOpt)}
                    >
                      {isSpecSelected ? '✓ ' : '+ '}{specOpt}
                    </button>
                  {/each}
                  {#each selectedGarmentSpecs.filter(s => !standardGarmentSpecsList.includes(s)) as customSpec}
                    <button
                      type="button"
                      class="spec-chip active"
                      onclick={() => toggleGarmentSpec(customSpec)}
                    >
                      ✓ {customSpec} ✕
                    </button>
                  {/each}
                </div>
                <div class="img-input-row" style="margin-top: 8px; margin-bottom: 4px;">
                  <input
                    id="customGarmentSpecInput"
                    type="text"
                    class="input table-input"
                    bind:value={customGarmentSpecInput}
                    placeholder="Add custom garment specification..."
                    onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addCustomGarmentSpec(); } }}
                  />
                  <button type="button" class="btn btn-secondary btn-xs" onclick={addCustomGarmentSpec}>+ Add Spec</button>
                </div>
                <span class="field-hint" style="font-size: 11px; color: var(--text-tertiary, #94a3b8); margin-top: 2px; display: block;">
                  {selectedGarmentSpecs.length} selected: {selectedGarmentSpecs.length ? selectedGarmentSpecs.join(' | ') : 'None yet'}
                </span>
              </div>

              <!-- Line 53: Composition -->
              <div class="field-group">
                <label for="composition" style="font-weight: 700; color: #f8fafc;">
                  Composition (22 Authoritative Dropdown Compositions) [CSV Line 53]
                </label>
                <Combobox
                  items={authoritativeCompositions}
                  bind:value={composition}
                  placeholder="Select or enter composition (e.g. 60%Cot 40%Poly)..."
                  allowCustom={true}
                />
                <span class="field-hint" style="font-size: 11px; color: var(--text-tertiary, #94a3b8); margin-top: 4px; display: block;">
                  Auto-inferred from fabric data ({prodMaterials.find(m => m.materialType === 'fabric')?.name || 'No fabric'}), editable via combobox.
                </span>
              </div>

              <!-- Line 54: MRP -->
              <div class="field-group">
                <label for="mrp" style="font-weight: 700; color: #f8fafc;">
                  Maximum Retail Price (MRP ₹) [CSV Line 54]
                </label>
                <div style="position: relative;">
                  <span style="position: absolute; left: 12px; top: 50%; transform: translateY(-50%); color: #94a3b8; font-weight: 700;">₹</span>
                  <input
                    id="mrp"
                    type="number"
                    step="1"
                    min="0"
                    class="input"
                    style="padding-left: 28px; font-weight: 700; font-size: 1rem; color: #38bdf8;"
                    bind:value={mrp}
                    placeholder="1499"
                  />
                </div>
                <span class="field-hint" style="font-size: 11px; color: var(--text-tertiary, #94a3b8); margin-top: 4px; display: block;">
                  Standard recommended retail price in INR (editable).
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- ============================================================ -->
        <!-- GENDER VARIATIONS / NON-APPAREL PROFILE                      -->
        <!-- ============================================================ -->
        {#if isProdNonApparel}
          {@const naScale = getSizesForSampling('N/A', prodGenre)}
          {@const naG = prodGenders[0] || { gender: 'N/A', sizes: [], pharmaNos: [''], productImages: [], colorImages: [] }}
          <div class="sub-table-block" style="margin-top: 20px;">
            <div class="sub-table-header">
              <span class="sub-title">📦 Non-Apparel Product Profile, Sizes, Mould Refs &amp; Images (No Gender)</span>
            </div>
            <div class="gender-card">
              <!-- Choose Sizes / Capacity in Product Development -->
              <div class="size-selector-box">
                <div class="size-selector-header">
                  <span style="color: var(--color-text-muted, #94a3b8); font-weight: 600; font-size: 12px;">
                    Choose Applicable Sizes / Capacities ({(naG.sizes || []).length} / {naScale.length} selected):
                  </span>
                  <div style="display:flex;gap:6px;">
                    <button type="button" class="btn btn-secondary btn-xs" onclick={() => selectAllProductSizes(0)}>Select All</button>
                    <button type="button" class="btn btn-secondary btn-xs" onclick={() => clearProductSizes(0)}>Clear</button>
                  </div>
                </div>
                <div class="size-picker-grid">
                  {#each naScale as sz}
                    {@const activeSz = (naG.sizes || []).includes(sz)}
                    <button type="button" class="size-toggle-btn {activeSz ? 'active' : ''}" onclick={() => toggleProductSize(0, sz)}>
                      {activeSz ? '✓ ' : ''}{sz}
                    </button>
                  {/each}
                </div>
              </div>

              <div class="form-grid" style="margin-top: 12px;">
                <div class="field-group">
                  <div style="display:flex;justify-content:space-between;align-items:center;">
                    <label for="na-pharma-0">Mould / Tooling / Pharma Nos (Multiple)</label>
                    <button type="button" class="btn btn-secondary btn-xs" onclick={() => addGenderPharmaNo(0)}>+ Add Ref</button>
                  </div>
                  <div class="pharma-list">
                    {#each (naG.pharmaNos && naG.pharmaNos.length ? naG.pharmaNos : ['']) as _, pIdx}
                      <div class="img-input-row" style="margin-bottom:4px;">
                        <input
                          id="na-pharma-{pIdx}"
                          type="text"
                          class="input table-input"
                          placeholder="e.g. MOULD-SSB-500-V{pIdx + 1}"
                          bind:value={prodGenders[0].pharmaNos[pIdx]}
                          oninput={() => { prodGenders[0].pharmaNo = prodGenders[0].pharmaNos[0] || ''; }}
                        />
                        {#if (prodGenders[0].pharmaNos || []).length > 1}
                          <button type="button" class="del-btn" onclick={() => removeGenderPharmaNo(0, pIdx)}>✕</button>
                        {/if}
                      </div>
                    {/each}
                  </div>
                </div>
                  <div class="field-group">
                    <ProductImageUploader
                      bind:imageUrl={
                        () => prodGenders[0]?.measurementChartUrl || '',
                        (v) => { if (prodGenders[0]) prodGenders[0].measurementChartUrl = v; }
                      }
                      label="Technical Drawing / Measurement Spec (Upload or URL)"
                    />
                  </div>
                </div>

                <!-- Product Images (Non-Apparel) -->
                <div class="image-management-section">
                  <div class="image-sub-title">
                    <span>📸 Product Images (Direct Upload + Links) [CSV Line 36]</span>
                    <label class="upload-file-label btn btn-primary btn-xs">
                      {uploadingTarget === 'prod_0' ? 'Uploading...' : '⬆️ Upload Images'}
                      <input type="file" multiple accept="image/*" style="display:none;" onchange={(e) => handleProductImagesFileUpload(e, 0)} />
                    </label>
                  </div>
                  <div class="img-input-row">
                    <input
                      type="text"
                      class="input table-input"
                      placeholder="Or paste image URL (e.g. /uploads/bag_front.jpg)"
                      bind:value={newImgUrlInput[0]}
                      onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addGenderProductImage(0); } }}
                    />
                    <button type="button" class="btn btn-secondary btn-xs" onclick={() => addGenderProductImage(0)}>+ Add URL</button>
                  </div>
                  {#if prodGenders[0]?.productImages && prodGenders[0].productImages.length > 0}
                    <div class="img-thumbnails-row">
                      {#each prodGenders[0].productImages as imgUrl, imgIdx}
                        <div class="img-thumbnail-chip">
                          <img src={imgUrl} alt="Product view" onerror={(e) => { /** @type {HTMLImageElement} */ (e.currentTarget).style.display = 'none'; }} />
                          <button type="button" class="img-remove-btn" onclick={() => removeGenderProductImage(0, imgIdx)} title="Remove image">✕</button>
                        </div>
                      {/each}
                    </div>
                  {/if}

                  <!-- Color Available Images (Non-Apparel Finishes) -->
                  <div class="image-sub-title" style="margin-top: 14px;">
                    <span>🎨 Color / Finish Available Images (Planned Colorways) [CSV Line 37]</span>
                    {#if sampleColors.length > 0}
                      <button type="button" class="btn btn-secondary btn-xs" onclick={() => syncPlannedColorsToGender(0)}>
                        🔄 Sync Planned Colors ({sampleColors.length})
                      </button>
                    {/if}
                  </div>
                  <div class="img-input-row">
                    <div style="flex:1;">
                      <Combobox
                        items={masters.colors || []}
                        bind:value={
                          () => newColorNameInput[0] || '',
                          (v) => { newColorNameInput = { ...newColorNameInput, 0: v }; }
                        }
                        category="color"
                        displayKey="display_name"
                        valueKey="display_name"
                        allowCustom={true}
                        placeholder="Select Color from Master (Black, Silver, Navy...)"
                        onSelect={(item) => addGenderColorVariant(0, item)}
                      />
                    </div>
                    <button type="button" class="btn btn-secondary btn-xs" onclick={() => addGenderColorVariant(0)}>+ Add Color Variant</button>
                  </div>
                {#if prodGenders[0]?.colorImages && prodGenders[0].colorImages.length > 0}
                  {#each prodGenders[0].colorImages as colorVar, cIdx}
                    <div class="color-variant-box">
                      <div class="color-variant-header">
                        <span class="color-badge">{colorVar.colorName} ({(colorVar.images || []).length} images)</span>
                        <div style="display:flex;gap:8px;align-items:center;">
                          <label class="upload-file-label btn btn-primary btn-xs">
                            {uploadingTarget === `col_0_${cIdx}` ? 'Uploading...' : `⬆️ Upload ${colorVar.colorName} Images`}
                            <input type="file" multiple accept="image/*" style="display:none;" onchange={(e) => handleColorImagesFileUpload(e, 0, cIdx)} />
                          </label>
                          <button type="button" class="del-btn" onclick={() => removeGenderColorVariant(0, cIdx)}>✕ Remove Color</button>
                        </div>
                      </div>
                      <div class="img-input-row">
                        <input
                          type="text"
                          class="input table-input"
                          placeholder="Or paste URL for {colorVar.colorName} (front, back, side)"
                          bind:value={newColorImgInput[`0_${cIdx}`]}
                          onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addColorImage(0, cIdx); } }}
                        />
                        <button type="button" class="btn btn-secondary btn-xs" onclick={() => addColorImage(0, cIdx)}>+ Add URL</button>
                      </div>
                      {#if colorVar.images && colorVar.images.length > 0}
                        <div class="img-thumbnails-row" style="margin-bottom: 0;">
                          {#each colorVar.images as cImg, cImgIdx}
                            <div class="img-thumbnail-chip">
                              <img src={cImg} alt="{colorVar.colorName} view" onerror={(e) => { /** @type {HTMLImageElement} */ (e.currentTarget).style.display = 'none'; }} />
                              <button type="button" class="img-remove-btn" onclick={() => removeColorImage(0, cIdx, cImgIdx)}>✕</button>
                            </div>
                          {/each}
                        </div>
                      {/if}
                    </div>
                  {/each}
                {/if}
              </div>
            </div>
          </div>
        {:else}
          <div class="sub-table-block" style="margin-top: 20px;">
            <div class="sub-table-header">
              <span class="sub-title">Gender Variations, Sizes, Multiple Pharma Nos, Measurements &amp; (Gender × Color) Images</span>
              <button type="button" class="btn btn-secondary btn-xs" onclick={addProductGender}>+ Add Gender</button>
            </div>

            {#each prodGenders as g, idx}
              {@const validGenderSizes = getSizesForSampling(g.gender, prodGenre)}
              <div class="gender-card">
                <div class="gender-card-header">
                  <div style="display:flex;align-items:center;gap:10px;">
                    <span class="gender-tag">{g.gender}</span>
                    <select
                      class="input table-input"
                      style="width: 140px;"
                      bind:value={g.gender}
                      onchange={() => {
                        const allowed = getSizesForSampling(g.gender, prodGenre);
                        g.sizes = (g.sizes || []).filter(s => allowed.includes(s));
                        if (g.sizes.length === 0) g.sizes = allowed.slice(0, 4);
                      }}
                    >
                      {#each canonicalGenders as cg}
                        <option value={cg}>{cg}</option>
                      {/each}
                    </select>
                  </div>
                  <button type="button" class="del-btn" onclick={() => removeProductGender(idx)}>✕ Remove Gender</button>
                </div>

                <!-- Interactive Per-Gender Size Chooser -->
                <div class="size-selector-box">
                  <div class="size-selector-header">
                    <span style="color: var(--color-text-muted, #94a3b8); font-weight: 600; font-size: 11.5px;">
                      Choose Sizes for {g.gender} ({(g.sizes || []).length} / {validGenderSizes.length} selected):
                    </span>
                    <div style="display:flex;gap:6px;">
                      <button type="button" class="btn btn-secondary btn-xs" onclick={() => selectAllProductSizes(idx)}>Select All</button>
                      <button type="button" class="btn btn-secondary btn-xs" onclick={() => clearProductSizes(idx)}>Clear</button>
                    </div>
                  </div>
                  <div class="size-picker-grid">
                    {#each validGenderSizes as sz}
                      {@const activeSz = (g.sizes || []).includes(sz)}
                      <button type="button" class="size-toggle-btn {activeSz ? 'active' : ''}" onclick={() => toggleProductSize(idx, sz)}>
                        {activeSz ? '✓ ' : ''}{sz}
                      </button>
                    {/each}
                  </div>
                </div>

                <div class="form-grid" style="margin-top: 10px;">
                  <div class="field-group">
                    <div style="display:flex;justify-content:space-between;align-items:center;">
                      <label for="pharmaNo-{idx}-0">Pharma Nos for {g.gender} (Multiple) [CSV Line 40]</label>
                      <button type="button" class="btn btn-secondary btn-xs" onclick={() => addGenderPharmaNo(idx)}>+ Add Pharma No</button>
                    </div>
                    <div class="pharma-list">
                      {#each (g.pharmaNos && g.pharmaNos.length ? g.pharmaNos : ['']) as _, pIdx}
                        <div class="img-input-row" style="margin-bottom:4px;">
                          <input
                            id="pharmaNo-{idx}-{pIdx}"
                            type="text"
                            class="input table-input"
                            placeholder="e.g. PH-{g.gender[0]}-POLO-V{pIdx + 1}"
                            bind:value={g.pharmaNos[pIdx]}
                            oninput={() => { g.pharmaNo = g.pharmaNos[0] || ''; }}
                          />
                          {#if (g.pharmaNos || []).length > 1}
                            <button type="button" class="del-btn" onclick={() => removeGenderPharmaNo(idx, pIdx)}>✕</button>
                          {/if}
                        </div>
                      {/each}
                    </div>
                  </div>
                  <div class="field-group">
                    <ProductImageUploader
                      bind:imageUrl={
                        () => prodGenders[idx]?.measurementChartUrl || '',
                        (v) => { if (prodGenders[idx]) prodGenders[idx].measurementChartUrl = v; }
                      }
                      label="Measurement Chart for {g.gender} (Upload or URL) [CSV Line 41]"
                    />
                  </div>
                </div>

                <!-- Product Images (diff for gender, upload multiple) [CSV Line 36] -->
                <div class="image-management-section">
                  <div class="image-sub-title">
                    <span>📸 Product Images for {g.gender} (Direct File Upload &amp; URLs) [CSV Line 36]</span>
                    <label class="upload-file-label btn btn-primary btn-xs">
                      {uploadingTarget === `prod_${idx}` ? 'Uploading...' : `⬆️ Upload ${g.gender} Images`}
                      <input type="file" multiple accept="image/*" style="display:none;" onchange={(e) => handleProductImagesFileUpload(e, idx)} />
                    </label>
                  </div>
                  <div class="img-input-row">
                    <input
                      type="text"
                      class="input table-input"
                      placeholder="Or paste URL for {g.gender} (e.g. /uploads/{g.gender.toLowerCase()}_polo_front.jpg)"
                      bind:value={newImgUrlInput[idx]}
                      onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addGenderProductImage(idx); } }}
                    />
                    <button type="button" class="btn btn-secondary btn-xs" onclick={() => addGenderProductImage(idx)}>+ Add URL</button>
                  </div>
                  {#if g.productImages && g.productImages.length > 0}
                    <div class="img-thumbnails-row">
                      {#each g.productImages as imgUrl, imgIdx}
                        <div class="img-thumbnail-chip">
                          <img src={imgUrl} alt="{g.gender} product view" onerror={(e) => { /** @type {HTMLImageElement} */ (e.currentTarget).style.display = 'none'; }} />
                          <button type="button" class="img-remove-btn" onclick={() => removeGenderProductImage(idx, imgIdx)} title="Remove image">✕</button>
                        </div>
                      {/each}
                    </div>
                  {/if}

                  <!-- Color Available Images (every gender-color combination planned) [CSV Line 37] -->
                  <div class="image-sub-title" style="margin-top: 14px;">
                    <span>🎨 Images for Each {g.gender} × Color Planned (Front, Back, Side Uploads) [CSV Line 37]</span>
                    {#if sampleColors.length > 0}
                      <button type="button" class="btn btn-secondary btn-xs" onclick={() => syncPlannedColorsToGender(idx)}>
                        🔄 Sync Sample Colors ({sampleColors.length})
                      </button>
                    {/if}
                  </div>
                  <div class="img-input-row">
                    <div style="flex:1;">
                      <Combobox
                        items={masters.colors || []}
                        bind:value={
                          () => newColorNameInput[idx] || '',
                          (v) => { newColorNameInput = { ...newColorNameInput, [idx]: v }; }
                        }
                        category="color"
                        displayKey="display_name"
                        valueKey="display_name"
                        allowCustom={true}
                        placeholder="Add planned Color for {g.gender} from Excel Master..."
                        onSelect={(item) => addGenderColorVariant(idx, item)}
                      />
                    </div>
                    <button type="button" class="btn btn-secondary btn-xs" onclick={() => addGenderColorVariant(idx)}>+ Add Color</button>
                  </div>
                  {#if g.colorImages && g.colorImages.length > 0}
                    {#each g.colorImages as colorVar, cIdx}
                      <div class="color-variant-box">
                        <div class="color-variant-header">
                          <span class="color-badge">{g.gender} × {colorVar.colorName} ({(colorVar.images || []).length} uploaded)</span>
                          <div style="display:flex;gap:8px;align-items:center;">
                            <label class="upload-file-label btn btn-primary btn-xs">
                              {uploadingTarget === `col_${idx}_${cIdx}` ? 'Uploading...' : `⬆️ Upload ${g.gender} ${colorVar.colorName} Views`}
                              <input type="file" multiple accept="image/*" style="display:none;" onchange={(e) => handleColorImagesFileUpload(e, idx, cIdx)} />
                            </label>
                            <button type="button" class="del-btn" onclick={() => removeGenderColorVariant(idx, cIdx)}>✕ Remove Color</button>
                          </div>
                        </div>
                        <div class="img-input-row">
                          <input
                            type="text"
                            class="input table-input"
                            placeholder="Or paste image URL for {g.gender} / {colorVar.colorName} (front, back, side)"
                            bind:value={newColorImgInput[`${idx}_${cIdx}`]}
                            onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addColorImage(idx, cIdx); } }}
                          />
                          <button type="button" class="btn btn-secondary btn-xs" onclick={() => addColorImage(idx, cIdx)}>+ Add URL</button>
                        </div>
                        {#if colorVar.images && colorVar.images.length > 0}
                          <div class="img-thumbnails-row" style="margin-bottom: 0;">
                            {#each colorVar.images as cImg, cImgIdx}
                              <div class="img-thumbnail-chip">
                                <img src={cImg} alt="{g.gender} {colorVar.colorName} view" onerror={(e) => { /** @type {HTMLImageElement} */ (e.currentTarget).style.display = 'none'; }} />
                                <button type="button" class="img-remove-btn" onclick={() => removeColorImage(idx, cIdx, cImgIdx)}>✕</button>
                              </div>
                            {/each}
                          </div>
                        {/if}
                      </div>
                    {/each}
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        {/if}

        <!-- ============================================================ -->
        <!-- COMPONENT MATERIALS BOM (CSV LINES 42-49)                    -->
        <!-- ============================================================ -->
        <div class="sub-table-block" style="margin-top: 20px;">
          <div class="sub-table-header">
            <span class="sub-title">Component Materials BOM (Fabrics, Trims & Zips) [CSV Lines 42-49]</span>
            <div style="display:flex;gap:8px;">
              <button type="button" class="btn btn-secondary btn-xs" onclick={() => addProductMaterial('fabric')}>+ Fabric</button>
              <button type="button" class="btn btn-secondary btn-xs" onclick={() => addProductMaterial('trim')}>+ Trim</button>
              <button type="button" class="btn btn-secondary btn-xs" onclick={() => addProductMaterial('zip')}>+ Zip</button>
            </div>
          </div>

          <table class="data-table">
            <thead>
              <tr>
                <th style="width: 15%;">Type</th>
                <th style="width: 35%;">Component Name</th>
                <th style="width: 12%;">Avg (Consump)</th>
                <th style="width: 12%;">GSM</th>
                <th style="width: 12%;">Dia</th>
                <th style="width: 14%;"></th>
              </tr>
            </thead>
            <tbody>
              {#each prodMaterials as m, idx}
                <tr>
                  <td>
                    <span class="type-pill {m.materialType}">{m.materialType.toUpperCase()}</span>
                  </td>
                  <td>
                    {#if m.materialType === 'fabric'}
                      <Combobox
                        items={masters.fabrics || []}
                        bind:value={m.name}
                        displayKey="display_name"
                        valueKey="display_name"
                        placeholder="Select fabric..."
                      />
                    {:else if m.materialType === 'trim'}
                      <Combobox
                        items={masters.trims || []}
                        bind:value={m.name}
                        displayKey="display_name"
                        valueKey="display_name"
                        placeholder="Select trim..."
                      />
                    {:else if m.materialType === 'zip'}
                      <Combobox
                        items={masters.zip_types || masters.zips || []}
                        bind:value={m.name}
                        displayKey="display_name"
                        valueKey="display_name"
                        placeholder="Select zip..."
                      />
                    {:else}
                      <input type="text" class="input table-input" bind:value={m.name} placeholder="Material name" />
                    {/if}
                  </td>
                  <td>
                    <input type="number" step="0.01" class="input table-input" bind:value={m.average} />
                  </td>
                  <td>
                    <input type="number" class="input table-input" bind:value={m.gsm} disabled={m.materialType !== 'fabric'} />
                  </td>
                  <td>
                    <input type="number" class="input table-input" bind:value={m.dia} disabled={m.materialType !== 'fabric'} />
                  </td>
                  <td class="text-right">
                    <button type="button" class="del-btn" onclick={() => removeProductMaterial(idx)}>✕</button>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>

        <!-- PUBLISH ACTION -->
        <div class="action-footer">
          <button
            type="button"
            class="btn btn-primary"
            style="min-width: 240px;"
            disabled={isPublishingProduct}
            onclick={publishMasterProduct}
          >
            {isPublishingProduct ? 'Publishing to Catalog...' : '✨ Publish to Master Catalog'}
          </button>
        </div>
      </section>
    </div>

  <!-- ============================================================ -->
  <!-- TAB 3: CLIENT SPECIFIC CATALOG & PRICING                     -->
  <!-- ============================================================ -->
  {:else if activeTab === 'client-pricing'}
    <div class="split-view">
      <!-- ENTRY FORM -->
      <div class="form-col">
        <section class="card">
          <div class="card-header">
            <span class="badge">Commercial Matrix</span>
            <h2 class="card-title">Client-Specific Contracted Pricing</h2>
          </div>

          <div class="form-grid">
            <div class="field-group">
              <label for="cpDate">Effective Date</label>
              <input id="cpDate" type="date" class="input" bind:value={cpEffectiveDate} />
            </div>

            <div class="field-group">
              <label for="cpClient">Client Name</label>
              <Combobox
                items={masters.clients || []}
                bind:value={cpClientName}
                displayKey="name"
                valueKey="name"
                placeholder="Select client..."
                onSelect={handlePricingClientSelect}
              />
            </div>

            <div class="field-group span-2">
              <label for="cpProduct">Master Product</label>
              <Combobox
                items={masters.products || []}
                bind:value={cpProductName}
                displayKey="display_name"
                valueKey="display_name"
                placeholder="Choose from master catalog..."
                onSelect={handlePricingProductSelect}
              />
            </div>

            <div class="field-group">
              <label for="cpPrice">Contracted Unit Price (₹)</label>
              <input
                id="cpPrice"
                type="number"
                step="0.5"
                class="input"
                placeholder="e.g. 525.00"
                bind:value={cpContractedPrice}
              />
            </div>

            <div class="field-group span-2">
              <ProductImageUploader
                bind:imageUrl={cpClientImageUrl}
                label="Client Product Image (Custom Co-Branding / Embroidery Mockup)"
              />
            </div>

            <div class="field-group span-2">
              <label for="cpNotes">Pricing Comments / Volume Slab</label>
              <textarea
                id="cpNotes"
                class="input"
                rows="2"
                placeholder="e.g. Minimum order 250 pcs, includes chest embroidery, FOB warehouse"
                bind:value={cpComments}
              ></textarea>
            </div>
          </div>

          <div class="action-footer">
            <button
              type="button"
              class="btn btn-primary"
              disabled={isSavingPricing}
              onclick={saveClientPricing}
            >
              {isSavingPricing ? 'Saving Agreement...' : '💾 Save Client Pricing Agreement'}
            </button>
          </div>
        </section>
      </div>

      <!-- PRICING DIRECTORY -->
      <aside class="sidebar-col">
        <div class="card sidebar-card">
          <div class="sidebar-header">
            <h3>Active Pricing Agreements</h3>
            <button class="refresh-btn" onclick={loadClientPricingList} title="Refresh">🔄</button>
          </div>

          <div class="sample-list-scroll">
            {#each clientPricingList as cp}
              <div class="sample-card">
                <div class="sc-top">
                  <span class="sc-no">{cp.clientName}</span>
                  <span class="price-chip">₹{cp.contractedPrice.toFixed(2)}</span>
                </div>
                <div class="sc-prod">{cp.productName}</div>
                {#if cp.comments}
                  <div class="sc-client" style="font-style: italic;">"{cp.comments}"</div>
                {/if}
                <div class="sc-meta">
                  <span>Effective: {cp.effectiveDate}</span>
                </div>
              </div>
            {/each}
          </div>
        </div>
      </aside>
    </div>
  {/if}
</div>

<style>
  .station-container {
    max-width: 1440px;
    margin: 0 auto;
    padding: 16px 24px;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .station-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    gap: 20px;
    flex-wrap: wrap;
    border-bottom: 1px solid var(--border-color, #334155);
    padding-bottom: 16px;
  }

  .station-pill {
    font-size: 0.72rem;
    font-weight: 700;
    color: var(--accent-color, #3b82f6);
    text-transform: uppercase;
    letter-spacing: 0.08em;
    margin-bottom: 4px;
  }

  .station-title {
    font-size: 1.5rem;
    font-weight: 800;
    color: var(--text-primary, #f8fafc);
    margin: 0;
  }

  .station-desc {
    font-size: 0.85rem;
    color: var(--text-tertiary, #94a3b8);
    margin: 4px 0 0;
  }

  /* TAB SWITCHER */
  .tab-switcher {
    display: flex;
    background: var(--bg-surface, #1e293b);
    border: 1px solid var(--border-color, #334155);
    border-radius: 8px;
    padding: 4px;
    gap: 4px;
  }

  .tab-btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    border-radius: 6px;
    border: none;
    background: transparent;
    color: var(--text-secondary, #94a3b8);
    font-size: 0.85rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .tab-btn:hover {
    color: var(--text-primary, #f8fafc);
  }

  .tab-btn.active {
    background: var(--accent-color, #3b82f6);
    color: #ffffff;
    box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
  }

  /* LAYOUT */
  .split-view {
    display: grid;
    grid-template-columns: 1fr 340px;
    gap: 20px;
    align-items: start;
  }

  .single-col-container {
    max-width: 1040px;
    margin: 0 auto;
    width: 100%;
  }

  .form-col {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .card {
    background: var(--bg-card, #1e293b);
    border: 1px solid var(--border-color, #334155);
    border-radius: 10px;
    padding: 20px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }

  .card-header {
    display: flex;
    align-items: center;
    gap: 12px;
    border-bottom: 1px solid var(--border-color, #334155);
    padding-bottom: 12px;
    margin-bottom: 16px;
  }

  .badge {
    font-size: 0.7rem;
    font-weight: 700;
    text-transform: uppercase;
    background: rgba(59, 130, 246, 0.15);
    color: var(--accent-color, #3b82f6);
    border: 1px solid rgba(59, 130, 246, 0.3);
    padding: 2px 8px;
    border-radius: 4px;
  }

  .card-title {
    font-size: 1.1rem;
    font-weight: 700;
    color: var(--text-primary, #f8fafc);
    margin: 0;
    flex: 1;
  }

  .status-chip {
    font-size: 0.72rem;
    font-weight: 700;
    padding: 3px 10px;
    border-radius: 12px;
    text-transform: uppercase;
    background: #334155;
    color: #94a3b8;
  }

  .status-chip.approved {
    background: rgba(34, 197, 94, 0.15);
    color: #22c55e;
    border: 1px solid #22c55e;
  }

  .status-chip.submitted {
    background: rgba(234, 179, 8, 0.15);
    color: #eab308;
    border: 1px solid #eab308;
  }

  .status-chip.rework_requested {
    background: rgba(239, 68, 68, 0.15);
    color: #ef4444;
    border: 1px solid #ef4444;
  }

  .form-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 14px;
  }

  .span-2 {
    grid-column: span 2;
  }

  .field-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .field-group label {
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--text-secondary, #94a3b8);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .input {
    width: 100%;
    padding: 8px 12px;
    font-size: 0.85rem;
    background: var(--bg-input, #0f172a);
    color: var(--text-primary, #f8fafc);
    border: 1px solid var(--border-color, #334155);
    border-radius: 6px;
    box-sizing: border-box;
  }

  .input:focus {
    outline: none;
    border-color: var(--accent-color, #3b82f6);
  }

  .input-with-action {
    display: flex;
    gap: 6px;
  }

  .media-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid var(--border-color, #334155);
  }

  /* SUB TABLES */
  .sub-table-block {
    margin-top: 16px;
    border: 1px solid var(--border-color, #334155);
    border-radius: 8px;
    overflow: visible;
    background: var(--bg-surface, #0f172a);
    position: relative;
  }

  .sub-table-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 14px;
    background: rgba(255, 255, 255, 0.02);
    border-bottom: 1px solid var(--border-color, #334155);
  }

  .sub-title {
    font-size: 0.82rem;
    font-weight: 700;
    color: var(--text-primary, #f8fafc);
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }

  .data-table {
    width: 100%;
    border-collapse: separate;
    border-spacing: 0;
    font-size: 0.85rem;
  }

  .data-table th {
    padding: 8px 12px;
    font-size: 0.72rem;
    font-weight: 600;
    text-transform: uppercase;
    color: var(--text-secondary, #94a3b8);
    text-align: left;
    border-bottom: 1px solid var(--border-color, #334155);
    background: rgba(0, 0, 0, 0.2);
  }

  .data-table td {
    padding: 6px 10px;
    border-bottom: 1px solid var(--border-color, #1e293b);
    position: relative;
    overflow: visible;
  }

  .table-input {
    padding: 6px 10px;
    font-size: 0.82rem;
  }

  /* NATIVE SELECT ACCESSIBILITY & VISIBILITY STYLING */
  select.input, select.table-input, select.pull-select {
    appearance: none;
    -webkit-appearance: none;
    background-image: url("data:image/svg+xml;charset=UTF-8,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%2338bdf8' stroke-width='2.5' stroke-linecap='round' stroke-linejoin='round'%3e%3cpolyline points='6 9 12 15 18 9'%3e%3c/polyline%3e%3c/svg%3e");
    background-repeat: no-repeat;
    background-position: right 10px center;
    background-size: 14px;
    padding-right: 32px;
    cursor: pointer;
    color-scheme: dark;
    background-color: var(--bg-input, #0f172a);
    color: var(--text-primary, #f8fafc);
    border: 1px solid var(--border-color, #334155);
  }

  select.input:focus, select.table-input:focus, select.pull-select:focus {
    border-color: var(--accent-primary, #3b82f6);
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.25);
  }

  select.input option, select.table-input option, select.pull-select option {
    background-color: #1e293b;
    color: #f8fafc;
    padding: 10px 14px;
    font-size: 13px;
  }

  select.input optgroup {
    background-color: #0f172a;
    color: #38bdf8;
    font-weight: 700;
    padding: 8px 10px;
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .del-btn {
    background: transparent;
    border: none;
    color: #ef4444;
    cursor: pointer;
    font-size: 0.85rem;
    padding: 4px 8px;
    border-radius: 4px;
  }

  .del-btn:hover {
    background: rgba(239, 68, 68, 0.15);
  }

  .text-right {
    text-align: right;
  }

  /* BUTTONS & FOOTER */
  .action-footer {
    display: flex;
    gap: 10px;
    justify-content: flex-end;
    flex-wrap: wrap;
    margin-top: 20px;
    padding-top: 16px;
    border-top: 1px solid var(--border-color, #334155);
  }

  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 8px 16px;
    font-size: 0.85rem;
    font-weight: 600;
    border-radius: 6px;
    border: none;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .btn-primary { background: #3b82f6; color: #ffffff; }
  .btn-primary:hover { background: #2563eb; }

  .btn-secondary { background: #334155; color: #f8fafc; }
  .btn-secondary:hover { background: #475569; }

  .btn-success { background: #16a34a; color: #ffffff; }
  .btn-success:hover { background: #15803d; }

  .btn-warning { background: #ca8a04; color: #ffffff; }
  .btn-warning:hover { background: #a16207; }

  .btn-danger { background: #dc2626; color: #ffffff; }
  .btn-danger:hover { background: #b91c1c; }

  .btn-xs { padding: 4px 8px; font-size: 0.75rem; }

  /* SIDEBAR DRAWER */
  .sidebar-card {
    display: flex;
    flex-direction: column;
    height: 780px;
  }

  .sidebar-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .sidebar-header h3 {
    font-size: 0.95rem;
    font-weight: 700;
    color: var(--text-primary, #f8fafc);
    margin: 0;
  }

  .refresh-btn {
    background: transparent;
    border: none;
    cursor: pointer;
    font-size: 0.9rem;
  }

  .sidebar-search {
    margin-bottom: 12px;
  }

  .sample-list-scroll {
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 8px;
    flex: 1;
  }

  .sample-card {
    background: var(--bg-surface, #0f172a);
    border: 1px solid var(--border-color, #334155);
    border-radius: 6px;
    padding: 10px 12px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .sample-card:hover {
    border-color: var(--accent-color, #3b82f6);
    background: rgba(59, 130, 246, 0.05);
  }

  .sample-card.active {
    border-color: var(--accent-color, #3b82f6);
    box-shadow: 0 0 0 1px var(--accent-color, #3b82f6);
  }

  .sc-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 4px;
  }

  .sc-no {
    font-weight: 700;
    font-size: 0.8rem;
    color: var(--accent-color, #3b82f6);
  }

  .sc-prod {
    font-size: 0.82rem;
    font-weight: 600;
    color: var(--text-primary, #f8fafc);
  }

  .sc-client {
    font-size: 0.75rem;
    color: var(--text-secondary, #94a3b8);
    margin-top: 2px;
  }

  .sc-meta {
    display: flex;
    justify-content: space-between;
    font-size: 0.7rem;
    color: var(--text-tertiary, #64748b);
    margin-top: 6px;
  }

  .price-chip {
    font-size: 0.75rem;
    font-weight: 700;
    color: #22c55e;
  }

  .pull-box {
    display: flex;
    align-items: center;
    gap: 12px;
    background: rgba(59, 130, 246, 0.08);
    border: 1px solid rgba(59, 130, 246, 0.25);
    border-radius: 6px;
    padding: 10px 14px;
    margin-bottom: 16px;
  }

  .pull-label {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 0.82rem;
    font-weight: 600;
    color: var(--accent-color, #3b82f6);
    white-space: nowrap;
  }

  .pull-select {
    flex: 1;
    font-weight: 500;
  }

  .gender-card {
    background: rgba(0, 0, 0, 0.15);
    border: 1px solid var(--border-color, #334155);
    border-radius: 6px;
    padding: 12px;
    margin-bottom: 10px;
  }

  .gender-card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .gender-tag {
    font-size: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
    color: #e2e8f0;
  }

  .type-pill {
    font-size: 0.65rem;
    font-weight: 700;
    padding: 2px 6px;
    border-radius: 4px;
  }

  .type-pill.fabric { background: rgba(59, 130, 246, 0.15); color: #3b82f6; }
  .type-pill.trim { background: rgba(234, 179, 8, 0.15); color: #eab308; }
  .type-pill.zip { background: rgba(168, 85, 247, 0.15); color: #a855f7; }

  .loading-state {
    padding: 20px;
    text-align: center;
    color: var(--text-tertiary, #64748b);
    font-size: 0.85rem;
  }

  .genre-recommendation-banner {
    margin-top: 14px;
    margin-bottom: 6px;
    padding: 10px 14px;
    background: linear-gradient(135deg, rgba(37, 99, 235, 0.12) 0%, rgba(30, 41, 59, 0.45) 100%);
    border: 1px solid rgba(59, 130, 246, 0.35);
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
  }

  .genre-rec-header {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }

  .genre-rec-badge {
    background: #2563eb;
    color: #ffffff;
    font-size: 11px;
    font-weight: 700;
    padding: 3px 8px;
    border-radius: 4px;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .genre-rec-text {
    font-size: 12px;
    color: var(--text-secondary, #cbd5e1);
  }

  .genre-rec-text strong {
    color: #38bdf8;
  }

  .optional-badge {
    font-size: 10px;
    padding: 2px 6px;
    background: rgba(255, 255, 255, 0.08);
    color: var(--text-tertiary, #94a3b8);
    border-radius: 4px;
    font-weight: 500;
    text-transform: none;
  }

  .sub-table-block.dimmed {
    opacity: 0.8;
  }

  /* PRODUCT CREATION IMAGES & COLOR VARIANT STYLING */
  .image-management-section {
    margin-top: 14px;
    padding-top: 14px;
    border-top: 1px solid rgba(255, 255, 255, 0.08);
  }

  .image-sub-title {
    font-size: 0.78rem;
    font-weight: 700;
    color: #38bdf8;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    margin-bottom: 8px;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .img-input-row {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-bottom: 10px;
  }

  .img-thumbnails-row {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-bottom: 14px;
  }

  .img-thumbnail-chip {
    position: relative;
    width: 64px;
    height: 64px;
    border-radius: 6px;
    overflow: hidden;
    border: 1px solid rgba(59, 130, 246, 0.4);
    background: #090d16;
  }

  .img-thumbnail-chip img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .img-remove-btn {
    position: absolute;
    top: 2px;
    right: 2px;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: rgba(239, 68, 68, 0.9);
    color: white;
    border: none;
    cursor: pointer;
    font-size: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    line-height: 1;
    transition: background 0.15s ease;
  }

  .img-remove-btn:hover {
    background: #dc2626;
  }

  .color-variant-box {
    background: rgba(15, 23, 42, 0.7);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 6px;
    padding: 10px 12px;
    margin-bottom: 8px;
  }

  .color-variant-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }

  .color-badge {
    font-size: 11px;
    font-weight: 700;
    color: #f8fafc;
    background: rgba(59, 130, 246, 0.2);
    padding: 2px 8px;
    border-radius: 4px;
    border: 1px solid rgba(59, 130, 246, 0.35);
  }

  .chips-wrap,
  .spec-multiselect {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .spec-chip {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 5px 10px;
    font-size: 11.5px;
    font-weight: 500;
    border-radius: 6px;
    border: 1px solid rgba(255, 255, 255, 0.14);
    background: rgba(15, 23, 42, 0.65);
    color: #cbd5e1;
    cursor: pointer;
    text-align: left;
  }

  .spec-chip.active {
    background: rgba(59, 130, 246, 0.2);
    border-color: #3b82f6;
    color: #f8fafc;
    font-weight: 600;
  }

  .chip-x {
    background: none;
    border: none;
    color: #fda4af;
    cursor: pointer;
    font-size: 11px;
    padding: 0 2px;
  }

  .attr-cards-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 10px 14px 14px;
  }

  .attr-row-card {
    background: rgba(15, 23, 42, 0.6);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
    padding: 10px 12px;
  }

  .attr-row-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }

  .row-index-tag {
    font-size: 11px;
    font-weight: 700;
    color: #38bdf8;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .attr-row-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(175px, 1fr));
    gap: 8px 10px;
  }

  .attr-field {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .attr-field.span-2-attr {
    grid-column: span 2;
  }

  .attr-label {
    font-size: 10.5px;
    font-weight: 700;
    text-transform: uppercase;
    color: #94a3b8;
    letter-spacing: 0.03em;
  }

  .size-selector-box {
    margin-top: 8px;
    padding: 8px 10px;
    background: rgba(0, 0, 0, 0.22);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 6px;
  }

  .size-selector-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
    gap: 8px;
    flex-wrap: wrap;
  }

  .size-picker-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .size-toggle-btn {
    padding: 4px 10px;
    border-radius: 5px;
    font-size: 11.5px;
    font-weight: 600;
    font-family: monospace;
    border: 1px solid rgba(255, 255, 255, 0.14);
    background: rgba(15, 23, 42, 0.7);
    color: #94a3b8;
    cursor: pointer;
  }

  .size-toggle-btn.active {
    background: rgba(16, 185, 129, 0.2);
    border-color: #10b981;
    color: #ecfdf5;
  }

  .pharma-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-top: 4px;
  }

  .upload-file-label {
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }

  @media (max-width: 1080px) {
    .split-view {
      grid-template-columns: 1fr;
    }
    .media-row {
      grid-template-columns: 1fr;
    }
    .attr-field.span-2-attr {
      grid-column: span 1;
    }
  }
</style>
