<script>
  import { api } from '../api.js';

  let {
    imageUrl = $bindable(''),
    label = 'Product Image'
  } = $props();

  let isUploading = $state(false);
  let uploadError = $state('');
  let fileInput = $state();

  async function handleFileSelect(e) {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!file.type.startsWith('image/')) {
      uploadError = 'Please upload a valid image (JPEG, PNG, WEBP, SVG)';
      return;
    }

    try {
      isUploading = true;
      uploadError = '';
      const res = await api.uploadFile(file);
      imageUrl = res.url;
      if (window.showToast) {
        window.showToast('Product image uploaded successfully', 'success');
      }
    } catch (err) {
      uploadError = err.message || 'Image upload failed';
    } finally {
      isUploading = false;
    }
  }

  function handleRemove() {
    imageUrl = '';
    if (fileInput) fileInput.value = '';
  }
</script>

<div class="uploader-container">
  <div class="label-row">
    <span class="uploader-title">{label}</span>
    {#if imageUrl}
      <button type="button" class="remove-btn" onclick={handleRemove} title="Remove image">
        ✕ Remove
      </button>
    {/if}
  </div>

  {#if imageUrl}
    <div class="preview-card">
      <img src={imageUrl} alt="Product Preview" class="preview-img" />
      <div class="preview-meta">
        <a href={imageUrl} target="_blank" rel="noreferrer" class="view-link">
          🔍 View Full Resolution
        </a>
      </div>
    </div>
  {:else}
    <div
      class="dropzone {isUploading ? 'uploading' : ''}"
      onclick={() => fileInput?.click()}
      onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && fileInput?.click()}
      role="button"
      tabindex="0"
    >
      <input
        type="file"
        accept="image/*"
        class="hidden-file-input"
        bind:this={fileInput}
        onchange={handleFileSelect}
      />
      {#if isUploading}
        <div class="spinner"></div>
        <span class="drop-text">Uploading image to database...</span>
      {:else}
        <span class="icon">📷</span>
        <span class="drop-text">Click or drop image to upload</span>
        <span class="hint">JPG, PNG, WebP or SVG (Max 16MB)</span>
      {/if}
    </div>
  {/if}

  {#if uploadError}
    <div class="error-msg">{uploadError}</div>
  {/if}
</div>

<style>
  .uploader-container {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .label-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .uploader-title {
    font-size: 11px;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--text-muted);
  }

  .remove-btn {
    background: transparent;
    border: none;
    color: var(--accent-rose);
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
    padding: 2px 6px;
    border-radius: var(--radius-sm);
  }

  .remove-btn:hover {
    background: rgba(244, 63, 94, 0.1);
  }

  .preview-card {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 8px;
    background: var(--bg-surface);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
  }

  .preview-img {
    width: 64px;
    height: 64px;
    object-fit: cover;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border-subtle);
    background: var(--bg-input);
  }

  .preview-meta {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .view-link {
    font-size: 11.5px;
    color: var(--accent-primary);
    text-decoration: none;
    font-weight: 600;
  }

  .view-link:hover {
    text-decoration: underline;
  }

  .dropzone {
    height: 72px;
    border: 1.5px dashed var(--border-subtle);
    border-radius: var(--radius-md);
    background: var(--bg-input);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.15s ease;
    padding: 8px;
    text-align: center;
  }

  .dropzone:hover {
    border-color: var(--accent-primary);
    background: rgba(37, 99, 235, 0.04);
  }

  .dropzone.uploading {
    opacity: 0.7;
    cursor: wait;
  }

  .hidden-file-input {
    display: none;
  }

  .icon {
    font-size: 18px;
    margin-bottom: 2px;
  }

  .drop-text {
    font-size: 11.5px;
    font-weight: 600;
    color: var(--text-main);
  }

  .hint {
    font-size: 10px;
    color: var(--text-subtle);
  }

  .error-msg {
    font-size: 11px;
    color: var(--accent-rose);
  }

  .spinner {
    width: 16px;
    height: 16px;
    border: 2px solid rgba(255, 255, 255, 0.2);
    border-top-color: var(--accent-primary);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
    margin-bottom: 4px;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>
