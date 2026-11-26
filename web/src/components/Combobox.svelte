<script>
  import { api } from '../api.js';

  /** @type {{ id?: string, value?: any, category?: string, placeholder?: string, items?: any[], options?: any[] | null, displayKey?: string, valueKey?: string, allowCustom?: boolean, onSelect?: (val: any) => void, disabled?: boolean }} */
  let {
    id = undefined,
    value = $bindable(),
    category = 'item',
    placeholder = 'Type to search or select...',
    items = [],
    options = null,
    displayKey = 'display_name',
    valueKey = 'display_name',
    allowCustom = true,
    onSelect = () => {},
    disabled = false
  } = $props();

  const listboxId = 'cb-listbox-' + Math.random().toString(36).slice(2, 8);

  let isOpen = $state(false);
  let isFiltering = $state(false);
  let highlightedIndex = $state(-1);
  let isRegistering = $state(false);
  let inputEl = $state(null);
  let wrapperEl = $state(null);
  let dropdownEl = $state(null);
  let dropdownStyle = $state('');

  let safeVal = $derived(value !== null && value !== undefined ? String(value) : '');

  // Svelte 5 action to portal the dropdown directly to document.body,
  // bypassing any parent card, table, or stacking context trapping.
  function portal(node) {
    document.body.appendChild(node);
    return {
      destroy() {
        if (node.parentNode) {
          node.parentNode.removeChild(node);
        }
      }
    };
  }

  // Normalize items/options supporting strings, {label, value}, or master objects
  let normalizedList = $derived.by(() => {
    const list = options || items || [];
    return list.map(item => {
      if (typeof item === 'string' || typeof item === 'number') {
        const str = String(item);
        return { display_name: str, name: str, value: str, code: str };
      }
      const disp = item[displayKey] || item.label || item.name || item.code || item.value || '';
      const val = item[valueKey] !== undefined ? item[valueKey] : (item.value !== undefined ? item.value : disp);
      return {
        ...item,
        display_name: String(disp),
        value: String(val),
        code: item.code || ''
      };
    });
  });

  // Filtered options based on user typing
  let filteredItems = $derived.by(() => {
    if (!isFiltering) {
      return normalizedList.slice(0, 150);
    }
    const q = safeVal.toLowerCase().trim();
    if (!q) return normalizedList.slice(0, 150);
    return normalizedList.filter(item => {
      const text = item.display_name.toLowerCase();
      const code = (item.code || '').toLowerCase();
      const val = item.value.toLowerCase();
      const gstin = (item.gstin || '').toLowerCase();
      return text.includes(q) || code.includes(q) || val.includes(q) || gstin.includes(q);
    }).slice(0, 150);
  });

  let exactMatch = $derived.by(() => {
    const trimmed = safeVal.toLowerCase().trim();
    return normalizedList.some(i => i.display_name.toLowerCase() === trimmed || i.value.toLowerCase() === trimmed);
  });

  function updatePosition() {
    if (!inputEl) return;
    const rect = inputEl.getBoundingClientRect();
    if (rect.bottom < 0 || rect.top > window.innerHeight) {
      isOpen = false;
      return;
    }

    const spaceBelow = window.innerHeight - rect.bottom;
    const spaceAbove = rect.top;
    const maxHeight = 280;
    const flipUp = spaceBelow < 220 && spaceAbove > spaceBelow;

    const width = Math.max(rect.width, 240);
    let left = rect.left;
    if (left + width > window.innerWidth - 8) {
      left = Math.max(8, window.innerWidth - width - 8);
    } else if (left < 8) {
      left = 8;
    }

    if (flipUp) {
      const bottom = Math.max(8, window.innerHeight - rect.top + 4);
      const availHeight = Math.min(maxHeight, Math.max(100, spaceAbove - 16));
      dropdownStyle = `position: fixed !important; bottom: ${bottom}px !important; top: auto !important; left: ${left}px !important; width: ${width}px !important; max-height: ${availHeight}px !important; z-index: 9999999 !important;`;
    } else {
      const top = rect.bottom + 4;
      const availHeight = Math.min(maxHeight, Math.max(100, spaceBelow - 16));
      dropdownStyle = `position: fixed !important; top: ${top}px !important; bottom: auto !important; left: ${left}px !important; width: ${width}px !important; max-height: ${availHeight}px !important; z-index: 9999999 !important;`;
    }
  }

  // Click outside to dismiss dropdown reliably without blur races
  $effect(() => {
    if (typeof document === 'undefined') return;
    function handlePointerDown(e) {
      if (
        wrapperEl && !wrapperEl.contains(e.target) &&
        dropdownEl && !dropdownEl.contains(e.target)
      ) {
        isOpen = false;
        isFiltering = false;
      }
    }
    function handleScrollOrResize() {
      if (isOpen) updatePosition();
    }
    if (isOpen) {
      updatePosition();
      document.addEventListener('pointerdown', handlePointerDown, true);
      window.addEventListener('scroll', handleScrollOrResize, true);
      window.addEventListener('resize', handleScrollOrResize, true);
    }
    return () => {
      document.removeEventListener('pointerdown', handlePointerDown, true);
      window.removeEventListener('scroll', handleScrollOrResize, true);
      window.removeEventListener('resize', handleScrollOrResize, true);
    };
  });

  // Scroll highlighted item into view
  $effect(() => {
    if (isOpen && dropdownEl && highlightedIndex >= 0) {
      const el = dropdownEl.querySelector('.highlighted');
      if (el) el.scrollIntoView({ block: 'nearest' });
    }
  });

  function openDropdown(enableFilter = false) {
    updatePosition();
    isOpen = true;
    isFiltering = enableFilter;
    if (filteredItems.length > 0) {
      const idx = filteredItems.findIndex(i => i.value === safeVal || i.display_name === safeVal);
      highlightedIndex = idx >= 0 ? idx : 0;
    }
  }

  function toggleDropdown() {
    if (isOpen) {
      isOpen = false;
      isFiltering = false;
    } else {
      openDropdown(false);
      inputEl?.focus();
    }
  }

  function handleInput(e) {
    value = e.target.value;
    isFiltering = true;
    updatePosition();
    isOpen = true;
    highlightedIndex = 0;
  }

  function handleFocus() {
    openDropdown(false);
  }

  function selectOption(item) {
    const val = item.value || item.display_name || '';
    value = val;
    isOpen = false;
    isFiltering = false;
    highlightedIndex = -1;
    onSelect(item);
  }

  function clearSelection(e) {
    e.stopPropagation();
    value = '';
    isFiltering = false;
    openDropdown(false);
    inputEl?.focus();
  }

  async function handleRegisterNew() {
    const term = safeVal.trim();
    if (!term || isRegistering) return;

    try {
      isRegistering = true;
      const res = await api.registerMaster(category, term);
      const newItem = {
        id: res.id,
        display_name: res.display_name || res.name || term,
        value: res.display_name || res.name || term,
        name: res.display_name || res.name || term,
        code: res.code || term
      };
      if (options && Array.isArray(options)) {
        options.push(newItem);
      } else if (items && Array.isArray(items)) {
        items.push(newItem);
      }
      selectOption(newItem);
      if (window.showToast) {
        window.showToast(`Registered new ${category}: "${term}"`, 'success');
      }
    } catch (err) {
      if (window.showToast) {
        window.showToast(`Failed to register ${category}: ${err.message}`, 'error');
      }
    } finally {
      isRegistering = false;
    }
  }

  function handleKeydown(e) {
    if (e.key === 'Tab') {
      isOpen = false;
      isFiltering = false;
      return;
    }

    if (!isOpen) {
      if (e.key === 'ArrowDown' || e.key === 'Enter') {
        e.preventDefault();
        openDropdown(false);
      }
      return;
    }

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      highlightedIndex = Math.min(highlightedIndex + 1, filteredItems.length - 1);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      highlightedIndex = Math.max(highlightedIndex - 1, 0);
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (highlightedIndex >= 0 && filteredItems[highlightedIndex]) {
        selectOption(filteredItems[highlightedIndex]);
      } else if (allowCustom && safeVal.trim() && !exactMatch) {
        handleRegisterNew();
      }
    } else if (e.key === 'Escape') {
      isOpen = false;
      isFiltering = false;
    }
  }
</script>

<div class="combobox-wrapper" class:is-open={isOpen} bind:this={wrapperEl}>
  <div class="input-wrap">
    <input
      {id}
      type="text"
      bind:this={inputEl}
      class="input-control"
      {placeholder}
      {disabled}
      bind:value={value}
      onfocus={handleFocus}
      oninput={handleInput}
      onkeydown={handleKeydown}
      autocomplete="off"
      role="combobox"
      aria-controls={isOpen ? listboxId : undefined}
      aria-expanded={isOpen}
      aria-haspopup="listbox"
      aria-autocomplete="list"
    />
    <div class="actions-wrap">
      {#if safeVal && !disabled}
        <button
          type="button"
          class="clear-btn"
          tabindex="-1"
          onclick={clearSelection}
          title="Clear"
        >
          ✕
        </button>
      {/if}
      <button
        type="button"
        class="toggle-btn"
        tabindex="-1"
        onclick={toggleDropdown}
        title="Toggle choices"
      >
        ▾
      </button>
    </div>
  </div>

  {#if isOpen}
    <div
      id={listboxId}
      class="combobox-dropdown"
      use:portal
      bind:this={dropdownEl}
      role="listbox"
      style={dropdownStyle}
    >
      {#if filteredItems.length === 0 && !safeVal.trim()}
        <div class="combobox-empty">No options found</div>
      {:else if filteredItems.length === 0}
        <div class="combobox-empty">No matching options for "{safeVal.trim()}"</div>
      {:else}
        {#each filteredItems as item, idx (item.value || item.display_name || idx)}
          {@const isSelected = item.value === safeVal || item.display_name === safeVal}
          <div
            class="combobox-item"
            class:highlighted={idx === highlightedIndex}
            class:selected={isSelected}
            onclick={() => selectOption(item)}
            onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && selectOption(item)}
            role="option"
            aria-selected={isSelected}
            tabindex="-1"
          >
            <span class="item-text">{item.display_name}</span>
            {#if item.code && item.code !== item.display_name}
              <span class="item-meta">{item.code}</span>
            {:else if item.gstin}
              <span class="item-meta">{item.gstin}</span>
            {/if}
          </div>
        {/each}
      {/if}

      {#if allowCustom && safeVal.trim() && !exactMatch}
        <div
          class="combobox-add-new"
          class:disabled={isRegistering}
          onclick={handleRegisterNew}
          role="button"
          tabindex="0"
          onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && handleRegisterNew()}
        >
          <span>{isRegistering ? 'Registering...' : `+ Register "${safeVal.trim()}" as new ${category}`}</span>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .combobox-wrapper {
    position: relative;
    width: 100%;
  }

  .input-wrap {
    position: relative;
    display: flex;
    align-items: center;
    width: 100%;
  }

  .input-wrap input {
    width: 100%;
    padding-right: 52px;
    background-color: var(--bg-input, #0f172a);
    color: var(--text-primary, #f8fafc);
    border: 1px solid var(--border-color, #334155);
    border-radius: 6px;
    font-size: 0.85rem;
    box-sizing: border-box;
  }

  .input-wrap input:focus {
    outline: none;
    border-color: var(--accent-primary, #3b82f6);
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.25);
  }

  .actions-wrap {
    position: absolute;
    right: 6px;
    top: 50%;
    transform: translateY(-50%);
    display: flex;
    align-items: center;
    gap: 2px;
  }

  .clear-btn {
    background: transparent;
    border: none;
    color: var(--text-tertiary, #94a3b8);
    cursor: pointer;
    padding: 3px 6px;
    font-size: 10px;
    border-radius: 4px;
    line-height: 1;
    transition: color 0.15s ease, background-color 0.15s ease;
  }

  .clear-btn:hover {
    color: #ef4444;
    background: rgba(239, 68, 68, 0.15);
  }

  .toggle-btn {
    background: transparent;
    border: none;
    color: var(--text-tertiary, #94a3b8);
    cursor: pointer;
    padding: 4px 6px;
    font-size: 11px;
    border-radius: 4px;
    line-height: 1;
    transition: color 0.15s ease, background-color 0.15s ease;
  }

  .toggle-btn:hover {
    color: var(--text-primary, #f8fafc);
    background: rgba(255, 255, 255, 0.08);
  }

  :global(.combobox-dropdown) {
    position: fixed !important;
    max-width: min(460px, calc(100vw - 16px));
    background-color: var(--bg-surface, #0e1422) !important;
    border: 1px solid var(--accent-primary, #3b82f6) !important;
    border-radius: 8px !important;
    overflow-y: auto !important;
    z-index: 9999999 !important;
    box-shadow: 0 20px 48px rgba(0, 0, 0, 0.75), 0 8px 16px rgba(0, 0, 0, 0.4) !important;
    font-family: var(--font-sans, sans-serif);
  }

  :global(.combobox-item) {
    padding: 9px 14px;
    cursor: pointer;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    font-size: 13px;
    color: var(--text-primary, #f8fafc) !important;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    transition: background-color 0.1s ease, color 0.1s ease;
  }

  :global(.combobox-item:last-child) {
    border-bottom: none;
  }

  :global(.combobox-item:hover), :global(.combobox-item.highlighted) {
    background-color: var(--accent-primary, #2563eb) !important;
    color: #ffffff !important;
  }

  :global(.combobox-item.selected) {
    font-weight: 700;
    color: var(--accent-primary, #60a5fa) !important;
  }

  :global(.combobox-item:hover.selected), :global(.combobox-item.highlighted.selected) {
    color: #ffffff !important;
  }

  :global(.combobox-item .item-text) {
    flex: 1;
    white-space: normal;
    word-break: break-word;
  }

  :global(.combobox-item .item-meta) {
    font-size: 11px;
    color: var(--text-tertiary, #94a3b8);
    font-family: var(--font-mono, monospace);
    white-space: nowrap;
  }

  :global(.combobox-item:hover .item-meta), :global(.combobox-item.highlighted .item-meta) {
    color: rgba(255, 255, 255, 0.9) !important;
  }

  :global(.combobox-empty) {
    padding: 14px 16px;
    font-size: 12px;
    color: var(--text-tertiary, #94a3b8);
    text-align: center;
    font-style: italic;
  }

  :global(.combobox-add-new) {
    padding: 10px 14px;
    background-color: rgba(16, 185, 129, 0.15);
    border-top: 1px solid rgba(16, 185, 129, 0.3);
    color: #10b981;
    cursor: pointer;
    font-weight: 700;
    font-size: 12px;
  }

  :global(.combobox-add-new:hover) {
    background-color: rgba(16, 185, 129, 0.25);
  }

  :global(.combobox-add-new.disabled) {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
