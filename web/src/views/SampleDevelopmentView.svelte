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
  let productGenre = $state('Polo');
  let tempProductName = $state('');
  let imageUrl = $state('');
  let productDescription = $state('');

  // Multi-item child tables
  let sizes = $state([
    { id: '', gender: 'Men', size: 'M', qty: 1 }
  ]);
  let fabrics = $state([
    { id: '', fabricName: '', lotNo: '', average: 0.35 }
  ]);
  let trims = $state([
    { id: '', trimName: '', average: 3.0 }
  ]);
  let zips = $state([
    { id: '', zipName: '', lengthAndQty: '' }
  ]);
  let labels = $state([
    { id: '', labelName: '' }
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
  let composition = $state('');
  let mrp = $state(999.0);

  // Gender variations
  let prodGenders = $state([
    {
      id: '',
      gender: 'Men',
      pharmaNo: '',
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

  function getSizesForGender(g) {
    if (masters?.gender_sizes?.[g] && masters.gender_sizes[g].length > 0) {
      return masters.gender_sizes[g];
    }
    return fallbackGenderSizes[g] || fallbackGenderSizes['Men'];
  }

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
    sizes = [...sizes, { id: '', gender: 'Men', size: 'L', qty: 1 }];
  }
  function removeSizeRow(idx) {
    sizes = sizes.filter((_, i) => i !== idx);
  }

  function addFabricRow() {
    fabrics = [...fabrics, { id: '', fabricName: '', lotNo: '', average: 0.35 }];
  }
  function removeFabricRow(idx) {
    fabrics = fabrics.filter((_, i) => i !== idx);
  }

  function addTrimRow() {
    trims = [...trims, { id: '', trimName: '', average: 1.0 }];
  }
  function removeTrimRow(idx) {
    trims = trims.filter((_, i) => i !== idx);
  }

  function addZipRow() {
    zips = [...zips, { id: '', zipName: '', lengthAndQty: '' }];
  }
  function removeZipRow(idx) {
    zips = zips.filter((_, i) => i !== idx);
  }

  function addLabelRow() {
    labels = [...labels, { id: '', labelName: '' }];
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
      tempProductName = detail.tempProductName || '';
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
      fabrics = detail.fabrics && detail.fabrics.length > 0 ? detail.fabrics : [{ id: '', fabricName: '', lotNo: '', average: 0.35 }];
      trims = detail.trims && detail.trims.length > 0 ? detail.trims : [{ id: '', trimName: '', average: 3.0 }];
      zips = detail.zips && detail.zips.length > 0 ? detail.zips : [{ id: '', zipName: '', lengthAndQty: '' }];
      labels = detail.labels && detail.labels.length > 0 ? detail.labels : [{ id: '', labelName: '' }];
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
    productGenre = 'Polo';
    tempProductName = '';
    imageUrl = '';
    productDescription = '';
    sizes = [{ id: '', gender: 'Men', size: 'M', qty: 1 }];
    fabrics = [{ id: '', fabricName: '', lotNo: '', average: 0.35 }];
    trims = [{ id: '', trimName: '', average: 3.0 }];
    zips = [{ id: '', zipName: '', lengthAndQty: '' }];
    labels = [{ id: '', labelName: '' }];
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

  // Convert Sample to Master Product
  function convertSampleToProduct() {
    prodGenre = productGenre || 'Polo';
    prodName = tempProductName || '';
    selectedSampleToPull = sampleId;

    // Convert Genders
    const uniqueGenders = [...new Set(sizes.map(s => s.gender).filter(Boolean))];
    if (uniqueGenders.length > 0) {
      prodGenders = uniqueGenders.map(g => ({
        id: '',
        gender: g,
        pharmaNo: pharmas[0]?.newPharmaNo || pharmas[0]?.basePharmaNo || '',
        measurementChartUrl: '',
        productImages: imageUrl ? [imageUrl] : [],
        colorImages: []
      }));
    }

    // Convert Materials
    const newMaterials = [];
    fabrics.forEach(f => {
      if (f.fabricName) {
        newMaterials.push({
          id: '',
          materialType: 'fabric',
          name: f.fabricName,
          average: f.average || 0.35,
          gsm: 220,
          dia: 32,
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
      window.showToast('Sample specs transferred to Product Creation!', 'success');
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

      const uniqueGenders = [...new Set((s.sizes || []).map(sz => sz.gender).filter(Boolean))];
      if (uniqueGenders.length > 0) {
        prodGenders = uniqueGenders.map(g => ({
          id: '',
          gender: g,
          pharmaNo: s.pharmas?.[0]?.newPharmaNo || s.pharmas?.[0]?.basePharmaNo || '',
          measurementChartUrl: '',
          productImages: s.sampleImageUrl ? [s.sampleImageUrl] : (s.imageUrl ? [s.imageUrl] : []),
          colorImages: []
        }));
      }

      const mats = [];
      (s.fabrics || []).forEach(f => {
        if (f.fabricName) {
          mats.push({ id: '', materialType: 'fabric', name: f.fabricName, average: f.average || 0.35, gsm: 220, dia: 32, zipLengthQty: '' });
        }
      });
      (s.trims || []).forEach(t => {
        if (t.trimName) {
          mats.push({ id: '', materialType: 'trim', name: t.trimName, average: t.average || 1.0, gsm: 0, dia: 0, zipLengthQty: '' });
        }
      });
      (s.zips || []).forEach(z => {
        if (z.zipName) {
          mats.push({ id: '', materialType: 'zip', name: z.zipName, average: 1.0, gsm: 0, dia: 0, zipLengthQty: z.lengthAndQty || '' });
        }
      });
      if (mats.length > 0) {
        prodMaterials = mats;
      }

      refreshSKU();
      if (window.showToast) {
        window.showToast(`Imported specs from ${s.developmentNo}`, 'info');
      }
    } catch (err) {
      alert('Failed to pull from sample: ' + err.message);
    }
  }

  function addProductGender() {
    prodGenders = [...prodGenders, {
      id: '',
      gender: 'Women',
      pharmaNo: '',
      measurementChartUrl: '',
      productImages: [],
      colorImages: []
    }];
  }
  function removeProductGender(idx) {
    prodGenders = prodGenders.filter((_, i) => i !== idx);
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
      const payload = {
        id: prodId,
        productName: prodName,
        genre: prodGenre,
        skuCode: skuCode,
        brandLabel: brandLabel,
        whyWeMadeThis: whyWeMadeThis,
        garmentSpecifications: garmentSpecifications,
        composition: composition,
        mrp: parseFloat(mrp) || 0.0,
        sampleDevelopmentId: selectedSampleToPull,
        genders: prodGenders,
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
        contractedPrice: parseFloat(cpContractedPrice),
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
              <label for="productGenre">Product Genre</label>
              <select id="productGenre" class="input" bind:value={productGenre}>
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
              <label for="tempName">Temp Product Name</label>
              <input
                id="tempName"
                type="text"
                class="input"
                placeholder="e.g. AeroCool Poly Pique Contrast Polo"
                bind:value={tempProductName}
              />
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
              <span class="sub-title">1. Gender, Size & Quantity Matrix</span>
              <button type="button" class="btn btn-secondary btn-xs" onclick={addSizeRow}>+ Add Size Row</button>
            </div>
            <table class="data-table">
              <thead>
                <tr>
                  <th style="width: 25%;">Gender</th>
                  <th style="width: 35%;">Size</th>
                  <th style="width: 25%;">Qty (Pcs)</th>
                  <th style="width: 15%;"></th>
                </tr>
              </thead>
              <tbody>
                {#each sizes as sz, idx}
                  <tr>
                    <td>
                      <select
                        class="input table-input"
                        bind:value={sz.gender}
                        onchange={() => {
                          const valid = getSizesForGender(sz.gender);
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
                    <td>
                      <select class="input table-input" bind:value={sz.size}>
                        {#each getSizesForGender(sz.gender) as s}
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

          <!-- B. FABRICS & LOTS -->
          <div class="sub-table-block">
            <div class="sub-table-header">
              <span class="sub-title">2. Fabrics, Lot & Consumption Average (kg/pc)</span>
              <button type="button" class="btn btn-secondary btn-xs" onclick={addFabricRow}>+ Add Fabric</button>
            </div>
            <table class="data-table">
              <thead>
                <tr>
                  <th style="width: 45%;">Fabric Name</th>
                  <th style="width: 25%;">Lot No (to be added)</th>
                  <th style="width: 20%;">Average (kg/pc)</th>
                  <th style="width: 10%;"></th>
                </tr>
              </thead>
              <tbody>
                {#each fabrics as fb, idx}
                  <tr>
                    <td>
                      <Combobox
                        items={masters.fabrics || []}
                        bind:value={fb.fabricName}
                        displayKey="display_name"
                        valueKey="display_name"
                        placeholder="Select 89 real FPCC fabrics..."
                      />
                    </td>
                    <td>
                      <input type="text" class="input table-input" bind:value={fb.lotNo} placeholder="Lot / Shade #" />
                    </td>
                    <td>
                      <input type="number" step="0.01" class="input table-input" bind:value={fb.average} />
                    </td>
                    <td class="text-right">
                      <button type="button" class="del-btn" onclick={() => removeFabricRow(idx)}>✕</button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>

          <!-- C. TRIMS -->
          <div class="sub-table-block">
            <div class="sub-table-header">
              <span class="sub-title">3. Trims & Average Consumption</span>
              <button type="button" class="btn btn-secondary btn-xs" onclick={addTrimRow}>+ Add Trim</button>
            </div>
            <table class="data-table">
              <thead>
                <tr>
                  <th style="width: 60%;">Trim Name</th>
                  <th style="width: 30%;">Average Consumption</th>
                  <th style="width: 10%;"></th>
                </tr>
              </thead>
              <tbody>
                {#each trims as tr, idx}
                  <tr>
                    <td>
                      <Combobox
                        items={masters.trims || []}
                        bind:value={tr.trimName}
                        displayKey="display_name"
                        valueKey="display_name"
                        placeholder="Select trims (buttons, fusing, thread)..."
                      />
                    </td>
                    <td>
                      <input type="number" step="0.1" class="input table-input" bind:value={tr.average} placeholder="e.g. 3 pcs" />
                    </td>
                    <td class="text-right">
                      <button type="button" class="del-btn" onclick={() => removeTrimRow(idx)}>✕</button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>

          <!-- D. ZIPS -->
          <div class="sub-table-block">
            <div class="sub-table-header">
              <span class="sub-title">4. Zips (Type, Size, Length & Qty)</span>
              <button type="button" class="btn btn-secondary btn-xs" onclick={addZipRow}>+ Add Zip</button>
            </div>
            <table class="data-table">
              <thead>
                <tr>
                  <th style="width: 50%;">Zip Name / Specification</th>
                  <th style="width: 40%;">Size-Wise Zip Length & Qty</th>
                  <th style="width: 10%;"></th>
                </tr>
              </thead>
              <tbody>
                {#each zips as zp, idx}
                  <tr>
                    <td>
                      <Combobox
                        items={masters.zip_types || masters.zips || []}
                        bind:value={zp.zipName}
                        displayKey="display_name"
                        valueKey="display_name"
                        placeholder="Select zip (Nylon, Vislon, Brass)..."
                      />
                    </td>
                    <td>
                      <input type="text" class="input table-input" bind:value={zp.lengthAndQty} placeholder="e.g. S-L: 15cm (1pc), XL: 18cm (1pc)" />
                    </td>
                    <td class="text-right">
                      <button type="button" class="del-btn" onclick={() => removeZipRow(idx)}>✕</button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>

          <!-- E. LABELS -->
          <div class="sub-table-block">
            <div class="sub-table-header">
              <span class="sub-title">5. Labels & Branding</span>
              <button type="button" class="btn btn-secondary btn-xs" onclick={addLabelRow}>+ Add Label</button>
            </div>
            <table class="data-table">
              <thead>
                <tr>
                  <th style="width: 90%;">Label Name / Specification</th>
                  <th style="width: 10%;"></th>
                </tr>
              </thead>
              <tbody>
                {#each labels as lb, idx}
                  <tr>
                    <td>
                      <Combobox
                        items={masters.label_descriptions || masters.labels || []}
                        bind:value={lb.labelName}
                        displayKey="display_name"
                        valueKey="display_name"
                        placeholder="Select Brand Label / Additional Label..."
                      />
                    </td>
                    <td class="text-right">
                      <button type="button" class="del-btn" onclick={() => removeLabelRow(idx)}>✕</button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>

          <!-- F. TEMP PHARMA NO -->
          <div class="sub-table-block">
            <div class="sub-table-header">
              <span class="sub-title">6. Temp Pharma No (Base Pattern & Grade Alterations)</span>
              <button type="button" class="btn btn-secondary btn-xs" onclick={addPharmaRow}>+ Add Pharma</button>
            </div>
            <table class="data-table">
              <thead>
                <tr>
                  <th style="width: 30%;">Base Pharma No</th>
                  <th style="width: 40%;">Changes Made / Alterations</th>
                  <th style="width: 20%;">New Pharma No</th>
                  <th style="width: 10%;"></th>
                </tr>
              </thead>
              <tbody>
                {#each pharmas as ph, idx}
                  <tr>
                    <td>
                      <input type="text" class="input table-input" bind:value={ph.basePharmaNo} placeholder="e.g. PH-POLO-STD-M" />
                    </td>
                    <td>
                      <input type="text" class="input table-input" bind:value={ph.changesMade} placeholder="e.g. Body length +1 inch, relaxed chest" />
                    </td>
                    <td>
                      <input type="text" class="input table-input" bind:value={ph.newPharmaNo} placeholder="e.g. PH-POLO-2026-REV1" />
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
          <span class="badge">Master Specification</span>
          <h2 class="card-title">Create Master Product from Approved Sample</h2>
        </div>

        <!-- PULL FROM SAMPLE HELPER -->
        <div class="pull-box">
          <div class="pull-label">
            <span class="pull-icon">⚡</span>
            <span>Import from Approved Sample Development:</span>
          </div>
          <select class="input pull-select" onchange={handlePullFromSampleChange}>
            <option value="">-- Choose Approved Sample to Pull Specs --</option>
            {#each sampleList as smp}
              <option value={smp.id} selected={selectedSampleToPull === smp.id}>
                {smp.developmentNo} • {smp.tempProductName} ({smp.clientName}) [{smp.status}]
              </option>
            {/each}
          </select>
        </div>

        <div class="form-grid">
          <div class="field-group">
            <label for="prodCreateDate">Creation Date</label>
            <input id="prodCreateDate" type="date" class="input" bind:value={creationDate} />
          </div>

          <div class="field-group">
            <label for="prodGenre">Product Genre</label>
            <select id="prodGenre" class="input" bind:value={prodGenre} onchange={refreshSKU}>
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

        <!-- GENDER VARIATIONS -->
        <div class="sub-table-block" style="margin-top: 20px;">
          <div class="sub-table-header">
            <span class="sub-title">Gender Variations, Pharma No & Images</span>
            <button type="button" class="btn btn-secondary btn-xs" onclick={addProductGender}>+ Add Gender</button>
          </div>

          {#each prodGenders as g, idx}
            <div class="gender-card">
              <div class="gender-card-header">
                <div style="display:flex;align-items:center;gap:10px;">
                  <span class="gender-tag">{g.gender}</span>
                  <select class="input table-input" style="width: 140px;" bind:value={g.gender}>
                    {#each canonicalGenders as cg}
                      <option value={cg}>{cg}</option>
                    {/each}
                  </select>
                </div>
                <button type="button" class="del-btn" onclick={() => removeProductGender(idx)}>✕ Remove Gender</button>
              </div>

              <!-- Authoritative Per-Gender Size Scale Preview -->
              <div class="gender-sizes-preview" style="margin-top: 8px; padding: 6px 10px; background: rgba(0,0,0,0.2); border: 1px solid rgba(255,255,255,0.06); border-radius: 6px; font-size: 11px;">
                <span style="color: var(--color-text-muted, #94a3b8); font-weight: 600; margin-right: 6px;">
                  Valid Size Scale ({getSizesForGender(g.gender).length} sizes):
                </span>
                <span style="color: #818cf8; font-family: monospace; font-weight: 600;">
                  {getSizesForGender(g.gender).join(', ')}
                </span>
              </div>

              <div class="form-grid" style="margin-top: 10px;">
                <div class="field-group">
                  <label for="pharmaNo-{idx}">Pharma No (diff for gender)</label>
                  <input id="pharmaNo-{idx}" type="text" class="input" placeholder="e.g. PH-M-POLO-2026" bind:value={g.pharmaNo} />
                </div>
                <div class="field-group">
                  <label for="measChart-{idx}">Measurement Chart Upload / Link</label>
                  <input id="measChart-{idx}" type="text" class="input" placeholder="e.g. /data/uploads/chart.pdf" bind:value={g.measurementChartUrl} />
                </div>
              </div>
            </div>
          {/each}
        </div>

        <!-- COMPONENT MATERIALS BOM -->
        <div class="sub-table-block" style="margin-top: 20px;">
          <div class="sub-table-header">
            <span class="sub-title">Component Materials BOM (Fabrics, Trims & Zips)</span>
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

        <!-- COMMERCIALS & SPECS -->
        <div class="form-grid" style="margin-top: 20px;">
          <div class="field-group span-2">
            <label for="whyMade">Why We Made This (Product Story)</label>
            <textarea
              id="whyMade"
              class="input"
              rows="2"
              placeholder="e.g. Engineered for extreme durability with breathable pique knit and reinforced seams..."
              bind:value={whyWeMadeThis}
            ></textarea>
          </div>

          <div class="field-group">
            <label for="garmentSpecs">Garment Specifications</label>
            <input
              id="garmentSpecs"
              type="text"
              class="input"
              placeholder="e.g. Rib collar, 3-button placket, side vents"
              bind:value={garmentSpecifications}
            />
          </div>

          <div class="field-group">
            <label for="composition">Composition (22 Authoritative Compositions)</label>
            <Combobox
              items={authoritativeCompositions}
              bind:value={composition}
              placeholder="Select or enter composition (e.g. 60%Cot 40%Poly)..."
            />
          </div>

          <div class="field-group">
            <label for="mrp">Maximum Retail Price (MRP ₹)</label>
            <input
              id="mrp"
              type="number"
              step="1"
              class="input"
              bind:value={mrp}
            />
          </div>
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

  select.input optgroup, select.table-input optgroup, select.pull-select optgroup {
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

  @media (max-width: 1080px) {
    .split-view {
      grid-template-columns: 1fr;
    }
    .media-row {
      grid-template-columns: 1fr;
    }
  }
</style>
