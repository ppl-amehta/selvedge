<script>
  let {
    personalization = $bindable([])
  } = $props();

  let isImportModalOpen = $state(false);
  let rawImportText = $state('');
  let parsedPreview = $state([]);
  let importFileName = $state('');

  const sizes = ['XS', 'S', 'M', 'L', 'XL', '2XL', '3XL', '4XL', '5XL', 'Free'];
  const genders = ['Men', 'Women', 'Unisex', 'Kids'];

  const reAlpha = /^[A-Za-z\s.'-]+$/;
  const reNumeric = /^[0-9]+$/;

  function validateRow(row) {
    const errors = [];
    const name = (row.name || '').trim();
    const number = (row.number || '').trim();

    if (name && !reAlpha.test(name)) {
      errors.push('Name must be alphabetical only (no numbers/symbols)');
    }
    if (number && !reNumeric.test(number)) {
      errors.push('Number must be digits only (0-999)');
    }
    return errors;
  }

  function addRow(gender = 'Men', size = 'M', name = '', number = '', remarks = '') {
    personalization = [
      ...personalization,
      {
        id: 'pers_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6),
        gender,
        size,
        name,
        number,
        remarks
      }
    ];
  }

  function removeRow(idx) {
    personalization = personalization.filter((_, i) => i !== idx);
  }

  function handleParseText(text) {
    if (!text.trim()) {
      parsedPreview = [];
      return;
    }

    const lines = text.trim().split(/\r?\n/);
    const results = [];

    lines.forEach((line, idx) => {
      const clean = line.trim();
      if (!clean) return;

      // Detect delimiter: tab or comma
      let cols = clean.includes('\t') ? clean.split('\t') : clean.split(',');
      cols = cols.map(c => c.trim().replace(/^["']|["']$/g, ''));

      // Check if header line
      if (idx === 0 && (cols[0]?.toLowerCase().includes('gender') || cols[2]?.toLowerCase().includes('name'))) {
        return; // skip header
      }

      // Expected 4 columns: [Gender] [Size] [Name] [Number]
      let g = cols[0] || 'Men';
      let s = cols[1] || 'M';
      let n = cols[2] || '';
      let num = cols[3] || '';

      // Auto-correct common gender casing
      const matchedGender = genders.find(item => item.toLowerCase() === g.toLowerCase());
      if (matchedGender) g = matchedGender;

      // Auto-correct size casing
      const matchedSize = sizes.find(item => item.toLowerCase() === s.toLowerCase());
      if (matchedSize) s = matchedSize;

      const item = { gender: g, size: s, name: n, number: num, remarks: '' };
      const errs = validateRow(item);

      results.push({
        ...item,
        isValid: errs.length === 0,
        errors: errs
      });
    });

    parsedPreview = results;
  }

  function handleFileUpload(e) {
    const file = e.target.files?.[0];
    if (!file) return;
    importFileName = file.name;

    const reader = new FileReader();
    reader.onload = (event) => {
      rawImportText = event.target?.result || '';
      handleParseText(rawImportText);
    };
    reader.readAsText(file);
  }

  function applyValidImports() {
    const validRows = parsedPreview.filter(r => r.isValid).map(r => ({
      id: 'pers_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6),
      gender: r.gender,
      size: r.size,
      name: r.name,
      number: r.number,
      remarks: r.remarks || ''
    }));

    if (validRows.length === 0) {
      if (window.showToast) window.showToast('No valid records to import', 'error');
      return;
    }

    personalization = [...personalization.filter(p => p.name || p.number), ...validRows];
    isImportModalOpen = false;
    rawImportText = '';
    parsedPreview = [];
    importFileName = '';

    if (window.showToast) {
      window.showToast(`Imported ${validRows.length} personalization records`, 'success');
    }
  }

  $effect(() => {
    if (personalization.length === 0) {
      addRow();
    }
  });
</script>

<div class="roster-box">
  <div class="roster-header">
    <div class="title-group">
      <span class="label">Per-Piece Personalization & Roster</span>
      <span class="count-pill">{personalization.filter(r => r.name || r.number).length} Customized Pcs</span>
      <span class="rule-hint">1 Alphabetical Name • 1 Numerical Number</span>
    </div>

    <div class="actions-group">
      <button type="button" class="btn btn-secondary btn-sm" onclick={() => { isImportModalOpen = true; handleParseText(rawImportText); }}>
        📊 4-Column Bulk Import (CSV / Sheets)
      </button>
      <button type="button" class="btn btn-secondary btn-sm" onclick={() => addRow()}>
        + Add Row
      </button>
    </div>
  </div>

  <div class="table-scroll-container">
    <table class="roster-table">
      <thead>
        <tr>
          <th style="width: 100px;">Gender</th>
          <th style="width: 90px;">Size</th>
          <th>Alphabetical Name (Letters only)</th>
          <th style="width: 130px;">Number (Digits 0-999)</th>
          <th>Remarks / Sleeve Placement</th>
          <th style="width: 44px;"></th>
        </tr>
      </thead>
      <tbody>
        {#each personalization as row, idx}
          {@const errs = validateRow(row)}
          {@const hasNameErr = row.name && !reAlpha.test(row.name)}
          {@const hasNumErr = row.number && !reNumeric.test(row.number)}
          <tr>
            <td>
              <select class="select-control compact" bind:value={row.gender}>
                {#each genders as g}
                  <option value={g}>{g}</option>
                {/each}
              </select>
            </td>
            <td>
              <select class="select-control compact" bind:value={row.size}>
                {#each sizes as s}
                  <option value={s}>{s}</option>
                {/each}
              </select>
            </td>
            <td>
              <input
                type="text"
                class="input-control compact {hasNameErr ? 'input-error' : ''}"
                placeholder="e.g. VIRAT KOHLI"
                bind:value={row.name}
              />
              {#if hasNameErr}
                <span class="field-error">Letters only (no numbers)</span>
              {/if}
            </td>
            <td>
              <input
                type="text"
                class="input-control compact input-mono {hasNumErr ? 'input-error' : ''}"
                placeholder="e.g. 18"
                bind:value={row.number}
              />
              {#if hasNumErr}
                <span class="field-error">Digits only</span>
              {/if}
            </td>
            <td>
              <input
                type="text"
                class="input-control compact"
                placeholder="Left Sleeve, Center..."
                bind:value={row.remarks}
              />
            </td>
            <td>
              <button
                type="button"
                class="del-row-btn"
                onclick={() => removeRow(idx)}
                title="Remove entry"
              >
                ✕
              </button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>

<!-- 4-Column Bulk Import Modal -->
{#if isImportModalOpen}
  <div class="modal-backdrop" onclick={() => isImportModalOpen = false} role="presentation">
    <div class="modal-card" onclick={(e) => e.stopPropagation()} role="dialog">
      <div class="modal-head">
        <div class="modal-title-group">
          <h3>4-Column Bulk Personalization Import</h3>
          <span class="modal-sub">Accepts 4 columns: <code>[Gender] [Size] [Name] [Number]</code> from CSV or Excel</span>
        </div>
        <button type="button" class="close-btn" onclick={() => isImportModalOpen = false}>✕</button>
      </div>

      <div class="modal-body">
        <div class="import-options-bar">
          <label class="file-pick-btn">
            📁 Select CSV File
            <input type="file" accept=".csv, .tsv, .txt" class="hidden-file-input" onchange={handleFileUpload} />
          </label>
          {#if importFileName}
            <span class="file-name-tag">Loaded: {importFileName}</span>
          {/if}
          <span class="divider">or</span>
          <span class="hint">Copy cells from Google Sheets / Excel and paste below:</span>
        </div>

        <textarea
          class="raw-input-textarea"
          rows="5"
          placeholder="Men	M	ROHIT SHARMA	45
Women	S	SMRITI MANDHANA	18
Men	L	JASPRIT BUMRAH	93"
          bind:value={rawImportText}
          oninput={(e) => handleParseText(e.target.value)}
        ></textarea>

        <!-- Live Validation Preview -->
        {#if parsedPreview.length > 0}
          <div class="preview-header">
            <span class="prev-title">Parsed Rows Preview ({parsedPreview.length} records)</span>
            <div class="prev-stats">
              <span class="stat-valid">✓ {parsedPreview.filter(r => r.isValid).length} Valid</span>
              {#if parsedPreview.filter(r => !r.isValid).length > 0}
                <span class="stat-invalid">⚠ {parsedPreview.filter(r => !r.isValid).length} Disqualified</span>
              {/if}
            </div>
          </div>

          <div class="preview-scroll">
            <table class="preview-table">
              <thead>
                <tr>
                  <th>Status</th>
                  <th>Gender</th>
                  <th>Size</th>
                  <th>Name</th>
                  <th>Number</th>
                  <th>Rule Check</th>
                </tr>
              </thead>
              <tbody>
                {#each parsedPreview as row}
                  <tr class={row.isValid ? 'valid-row' : 'invalid-row'}>
                    <td>
                      <span class="status-badge {row.isValid ? 'ok' : 'err'}">
                        {row.isValid ? '✓ Valid' : '✕ Error'}
                      </span>
                    </td>
                    <td>{row.gender}</td>
                    <td>{row.size}</td>
                    <td><strong>{row.name}</strong></td>
                    <td class="numeric-cell">{row.number}</td>
                    <td class="err-cell">
                      {#if row.errors?.length > 0}
                        {row.errors.join(', ')}
                      {:else}
                        <span style="color:var(--accent-emerald);">Passed</span>
                      {/if}
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </div>

      <div class="modal-foot">
        <button type="button" class="btn btn-secondary" onclick={() => isImportModalOpen = false}>
          Cancel
        </button>
        <button
          type="button"
          class="btn btn-primary"
          onclick={applyValidImports}
          disabled={parsedPreview.filter(r => r.isValid).length === 0}
        >
          Import {parsedPreview.filter(r => r.isValid).length} Valid Records
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .roster-box {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 16px;
    margin-top: 12px;
  }

  .roster-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
    flex-wrap: wrap;
    gap: 8px;
  }

  .title-group {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .label {
    font-size: 11px;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-main);
  }

  .count-pill {
    font-size: 10px;
    background: rgba(16, 185, 129, 0.12);
    border: 1px solid rgba(16, 185, 129, 0.3);
    color: var(--accent-emerald);
    padding: 2px 8px;
    border-radius: 9999px;
    font-weight: 700;
  }

  .rule-hint {
    font-size: 10.5px;
    color: var(--text-muted);
  }

  .actions-group {
    display: flex;
    gap: 6px;
  }

  .table-scroll-container {
    overflow-x: auto;
  }

  .roster-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 12px;
  }

  .roster-table th {
    background: var(--bg-input);
    color: var(--text-muted);
    font-size: 10.5px;
    font-weight: 700;
    text-align: left;
    padding: 8px 10px;
    border-bottom: 1px solid var(--border-subtle);
  }

  .roster-table td {
    padding: 6px 8px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.03);
  }

  .compact {
    height: 32px !important;
    font-size: 12px !important;
    padding: 4px 8px !important;
  }

  .input-error {
    border-color: var(--accent-rose) !important;
    background: rgba(244, 63, 94, 0.05) !important;
  }

  .field-error {
    display: block;
    font-size: 9.5px;
    color: var(--accent-rose);
    margin-top: 2px;
  }

  .del-row-btn {
    background: transparent;
    border: none;
    color: var(--text-subtle);
    font-size: 12px;
    cursor: pointer;
    padding: 4px 8px;
    border-radius: 4px;
  }

  .del-row-btn:hover {
    color: var(--accent-rose);
    background: rgba(244, 63, 94, 0.1);
  }

  /* Modal */
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 20px;
  }

  .modal-card {
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-lg);
    width: 100%;
    max-width: 720px;
    max-height: 90vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.6);
  }

  .modal-head {
    padding: 16px 20px;
    border-bottom: 1px solid var(--border-subtle);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .modal-title-group h3 {
    font-size: 15px;
    font-weight: 700;
    color: var(--text-main);
  }

  .modal-sub {
    font-size: 11.5px;
    color: var(--text-muted);
  }

  .modal-sub code {
    background: rgba(255, 255, 255, 0.1);
    padding: 1px 4px;
    border-radius: 3px;
    font-family: var(--font-mono);
  }

  .close-btn {
    background: transparent;
    border: none;
    color: var(--text-muted);
    font-size: 16px;
    cursor: pointer;
  }

  .modal-body {
    padding: 16px 20px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .import-options-bar {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 11.5px;
  }

  .file-pick-btn {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    padding: 6px 12px;
    border-radius: var(--radius-sm);
    cursor: pointer;
    font-weight: 600;
    color: var(--text-main);
  }

  .file-name-tag {
    font-size: 11px;
    color: var(--accent-primary);
  }

  .hidden-file-input { display: none; }
  .divider { color: var(--text-subtle); }
  .hint { color: var(--text-muted); }

  .raw-input-textarea {
    width: 100%;
    background: var(--bg-input);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    padding: 10px;
    color: var(--text-main);
    font-family: var(--font-mono);
    font-size: 11.5px;
    line-height: 1.4;
    resize: vertical;
  }

  .preview-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 4px;
  }

  .prev-title {
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    color: var(--text-muted);
  }

  .prev-stats {
    display: flex;
    gap: 8px;
    font-size: 11.5px;
    font-weight: 600;
  }

  .stat-valid { color: var(--accent-emerald); }
  .stat-invalid { color: var(--accent-rose); }

  .preview-scroll {
    max-height: 200px;
    overflow-y: auto;
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
  }

  .preview-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 11.5px;
  }

  .preview-table th {
    background: var(--bg-surface);
    padding: 6px 8px;
    text-align: left;
    font-size: 10px;
    color: var(--text-subtle);
    position: sticky;
    top: 0;
  }

  .preview-table td {
    padding: 5px 8px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.03);
  }

  .valid-row { background: rgba(16, 185, 129, 0.02); }
  .invalid-row { background: rgba(244, 63, 94, 0.05); }

  .status-badge {
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 10px;
    font-weight: 700;
  }

  .status-badge.ok {
    background: rgba(16, 185, 129, 0.15);
    color: var(--accent-emerald);
  }

  .status-badge.err {
    background: rgba(244, 63, 94, 0.15);
    color: var(--accent-rose);
  }

  .err-cell { font-size: 10.5px; color: var(--accent-rose); }
  .numeric-cell { font-family: var(--font-mono); }

  .modal-foot {
    padding: 14px 20px;
    border-top: 1px solid var(--border-subtle);
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }
</style>
