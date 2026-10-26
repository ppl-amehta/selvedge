<script>
  import { api } from '../api.js';

  let {
    brandings = $bindable([])
  } = $props();

  const brandingTypes = [
    { value: 'emb', label: 'Embroidery (emb)' },
    { value: 'print', label: 'Screen Print (print)' },
    { value: 'dtf', label: 'Direct to Film (dtf)' },
    { value: 'unik', label: 'Unik / High Density (unik)' },
    { value: 'other', label: 'Other Embellishment (other)' }
  ];

  let uploadingIndex = $state(-1);

  function addBranding() {
    brandings = [
      ...brandings,
      {
        id: 'bnd_' + Date.now() + '_' + Math.random().toString(36).substring(2, 6),
        brandingType: 'emb',
        position: 'Left Chest',
        size: '',
        placement: '',
        color: '',
        comment: '',
        brandingFileUrl: ''
      }
    ];
  }

  function removeBranding(idx) {
    brandings = brandings.filter((_, i) => i !== idx);
  }

  async function handleFileUpload(e, idx) {
    const file = e.target.files?.[0];
    if (!file) return;

    try {
      uploadingIndex = idx;
      const res = await api.uploadFile(file);
      brandings[idx].brandingFileUrl = res.url;
      if (window.showToast) {
        window.showToast(`Artwork design uploaded: ${res.filename}`, 'success');
      }
    } catch (err) {
      if (window.showToast) {
        window.showToast('Failed to upload artwork: ' + err.message, 'error');
      }
    } finally {
      uploadingIndex = -1;
    }
  }
</script>

<div class="embellishments-container">
  <div class="emb-header">
    <div class="title-group">
      <span class="label">Brandings & Embellishments</span>
      <span class="count-tag">{brandings.length} Configured</span>
    </div>
    <button type="button" class="btn btn-secondary btn-sm" onclick={addBranding}>
      + Add Embellishment
    </button>
  </div>

  {#if brandings.length === 0}
    <div class="empty-state">
      <span>No embellishments added to this product. Click <strong>+ Add Embellishment</strong> to specify embroidery, print, DTF, or badges.</span>
    </div>
  {:else}
    <div class="cards-list">
      {#each brandings as item, idx}
        <div class="emb-card">
          <div class="card-header">
            <span class="item-num">#{idx + 1} Embellishment</span>
            <button type="button" class="del-btn" onclick={() => removeBranding(idx)} title="Remove embellishment">
              ✕ Remove
            </button>
          </div>

          <div class="grid-fields">
            <!-- Branding Type -->
            <div class="field">
              <label for="emb-type-{item.id}">Type</label>
              <select id="emb-type-{item.id}" class="select-control" bind:value={item.brandingType}>
                {#each brandingTypes as t}
                  <option value={t.value}>{t.label}</option>
                {/each}
              </select>
            </div>

            <!-- Position -->
            <div class="field">
              <label for="emb-pos-{item.id}">Position</label>
              <input
                id="emb-pos-{item.id}"
                type="text"
                class="input-control"
                placeholder="e.g. Left Chest, Right Sleeve"
                bind:value={item.position}
              />
            </div>

            <!-- Size -->
            <div class="field">
              <label for="emb-size-{item.id}">Size</label>
              <input
                id="emb-size-{item.id}"
                type="text"
                class="input-control"
                placeholder="e.g. 3.5 x 2.0 inches"
                bind:value={item.size}
              />
            </div>

            <!-- Placement -->
            <div class="field">
              <label for="emb-place-{item.id}">Placement Specs</label>
              <input
                id="emb-place-{item.id}"
                type="text"
                class="input-control"
                placeholder="e.g. 7cm down from shoulder collar"
                bind:value={item.placement}
              />
            </div>

            <!-- Color -->
            <div class="field">
              <label for="emb-color-{item.id}">Color / Pantone</label>
              <input
                id="emb-color-{item.id}"
                type="text"
                class="input-control"
                placeholder="e.g. PMS 286C, Gold Thread"
                bind:value={item.color}
              />
            </div>

            <!-- Artwork Design File Upload -->
            <div class="field file-field">
              <label for="emb-file-{item.id}">Design Artwork File</label>
              {#if item.brandingFileUrl}
                <div class="file-chip">
                  <a href={item.brandingFileUrl} target="_blank" rel="noreferrer" class="file-link">
                    📎 {item.brandingFileUrl.split('/').pop()}
                  </a>
                  <button type="button" class="remove-file-btn" onclick={() => item.brandingFileUrl = ''} title="Remove file">
                    ✕
                  </button>
                </div>
              {:else}
                <label class="file-upload-label {uploadingIndex === idx ? 'uploading' : ''}">
                  <input
                    id="emb-file-{item.id}"
                    type="file"
                    class="hidden-file-input"
                    onchange={(e) => handleFileUpload(e, idx)}
                  />
                  <span>{uploadingIndex === idx ? 'Uploading...' : '📁 Upload Vector / Artwork'}</span>
                </label>
              {/if}
            </div>

            <!-- Comment / Instructions -->
            <div class="field full-width">
              <label for="emb-comm-{item.id}">Merchandiser Remarks / Execution Comment</label>
              <input
                id="emb-comm-{item.id}"
                type="text"
                class="input-control"
                placeholder="Special needle density, 3D foam fill, screen mesh count, or backing details..."
                bind:value={item.comment}
              />
            </div>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .embellishments-container {
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 16px;
    margin-top: 12px;
  }

  .emb-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
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

  .count-tag {
    font-size: 10px;
    background: rgba(37, 99, 235, 0.12);
    border: 1px solid rgba(37, 99, 235, 0.3);
    color: #93c5fd;
    padding: 2px 8px;
    border-radius: 9999px;
    font-weight: 700;
  }

  .empty-state {
    padding: 16px;
    background: var(--bg-input);
    border: 1px dashed var(--border-subtle);
    border-radius: var(--radius-sm);
    color: var(--text-muted);
    font-size: 12px;
    text-align: center;
  }

  .cards-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .emb-card {
    background: var(--bg-input);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    padding: 12px 14px;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
    padding-bottom: 6px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .item-num {
    font-size: 11px;
    font-weight: 800;
    color: var(--accent-primary);
    text-transform: uppercase;
  }

  .del-btn {
    background: transparent;
    border: none;
    color: var(--accent-rose);
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
  }

  .grid-fields {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 10px;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .field label {
    font-size: 10.5px;
    font-weight: 600;
    color: var(--text-muted);
  }

  .full-width {
    grid-column: 1 / -1;
  }

  .file-chip {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    padding: 5px 10px;
    border-radius: var(--radius-sm);
    height: 36px;
  }

  .file-link {
    font-size: 11px;
    color: var(--accent-primary);
    text-decoration: none;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 150px;
  }

  .remove-file-btn {
    background: transparent;
    border: none;
    color: var(--accent-rose);
    cursor: pointer;
    font-size: 11px;
  }

  .file-upload-label {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-surface);
    border: 1px dashed var(--border-subtle);
    border-radius: var(--radius-sm);
    height: 36px;
    cursor: pointer;
    font-size: 11px;
    color: var(--text-muted);
    transition: all 0.15s ease;
  }

  .file-upload-label:hover {
    border-color: var(--accent-primary);
    color: #fff;
  }

  .hidden-file-input {
    display: none;
  }
</style>
