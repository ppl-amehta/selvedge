<script>
  import { api } from '../api.js';

  let {
    value = $bindable(''),
    category = 'item',
    placeholder = 'Type to search or add new...',
    items = [],
    options = null,
    displayKey = 'display_name',
    valueKey = 'display_name',
    allowCustom = true,
    onSelect = () => {},
    disabled = false
  } = $props();

  let isOpen = $state(false);
  let highlightedIndex = $state(-1);
  let inputEl;
  let dropdownEl = $state(null);

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
    const q = (value || '').toLowerCase().trim();
    if (!q) return normalizedList.slice(0, 50);
    return normalizedList.filter(item => {
      const text = item.display_name.toLowerCase();
      const code = (item.code || '').toLowerCase();
      const val = item.value.toLowerCase();
      return text.includes(q) || code.includes(q) || val.includes(q);
    }).slice(0, 50);
  });

  let exactMatch = $derived.by(() => {
    const trimmed = (value || '').toLowerCase().trim();
    return normalizedList.some(i => i.display_name.toLowerCase() === trimmed || i.value.toLowerCase() === trimmed);
  });

  function selectOption(item) {
    const val = item.value || item.display_name || '';
    value = val;
    isOpen = false;
    highlightedIndex = -1;
    onSelect(item);
  }

  async function handleRegisterNew() {
    const term = value.trim();
    if (!term) return;

    try {
      const res = await api.registerMaster(category, term);
      const newItem = {
        id: res.id,
        display_name: res.display_name || res.name || term,
        value: res.display_name || res.name || term,
        name: res.display_name || res.name || term,
        code: res.code || term
      };
      if (options) {
        options.push(newItem);
      } else {
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
    }
  }

  function handleKeydown(e) {
    if (!isOpen) {
      if (e.key === 'ArrowDown' || e.key === 'Enter') {
        isOpen = true;
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
      } else if (allowCustom && value.trim() && !exactMatch) {
        handleRegisterNew();
      }
    } else if (e.key === 'Escape') {
      isOpen = false;
    }
  }

  function handleBlur() {
    setTimeout(() => {
      isOpen = false;
    }, 200);
  }
</script>

<div class="combobox-wrapper">
  <div class="input-wrap">
    <input
      type="text"
      bind:this={inputEl}
      class="input-control"
      {placeholder}
      {disabled}
      bind:value={value}
      onfocus={() => isOpen = true}
      onblur={handleBlur}
      onkeydown={handleKeydown}
      autocomplete="off"
    />
    <button
      type="button"
      class="toggle-btn"
      tabindex="-1"
      onclick={() => {
        isOpen = !isOpen;
        if (isOpen) inputEl?.focus();
      }}
    >
      ▾
    </button>
  </div>

  {#if isOpen}
    <div class="combobox-dropdown" bind:this={dropdownEl} role="listbox">
      {#if filteredItems.length === 0 && !value.trim()}
        <div class="combobox-empty">No items available</div>
      {:else}
        {#each filteredItems as item, idx}
          <div
            class="combobox-item"
            class:highlighted={idx === highlightedIndex}
            onmousedown={() => selectOption(item)}
            role="option"
            aria-selected={idx === highlightedIndex}
            tabindex="-1"
          >
            <span>{item.display_name}</span>
            {#if item.code && item.code !== item.display_name}
              <span class="item-meta">{item.code}</span>
            {/if}
          </div>
        {/each}
      {/if}

      {#if allowCustom && value.trim() && !exactMatch}
        <div class="combobox-add-new" onmousedown={handleRegisterNew} role="button" tabindex="0">
          <span>+ Register <strong>"{value.trim()}"</strong> as new {category}</span>
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
  }

  .input-wrap input {
    padding-right: 32px;
  }

  .toggle-btn {
    position: absolute;
    right: 8px;
    background: transparent;
    border: none;
    color: var(--text-subtle);
    cursor: pointer;
    padding: 4px;
    font-size: 11px;
    transition: color 0.15s ease;
  }

  .toggle-btn:hover {
    color: var(--text-main);
  }

  .combobox-dropdown {
    position: absolute;
    top: calc(100% + 4px);
    left: 0;
    right: 0;
    background-color: var(--bg-surface);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    max-height: 240px;
    overflow-y: auto;
    z-index: 200;
    box-shadow: var(--shadow-panel);
  }

  .combobox-item {
    padding: 8px 14px;
    cursor: pointer;
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 13px;
    color: var(--text-main);
    transition: background-color 0.1s ease;
  }

  .combobox-item:hover, .combobox-item.highlighted {
    background-color: var(--accent-primary);
    color: #ffffff;
  }

  .item-meta {
    font-size: 11px;
    color: var(--text-subtle);
    font-family: var(--font-mono);
  }

  .combobox-item:hover .item-meta, .combobox-item.highlighted .item-meta {
    color: rgba(255, 255, 255, 0.85);
  }

  .combobox-empty {
    padding: 12px 14px;
    font-size: 12px;
    color: var(--text-subtle);
  }

  .combobox-add-new {
    padding: 10px 14px;
    background-color: rgba(16, 185, 129, 0.1);
    border-top: 1px solid var(--border-subtle);
    color: var(--accent-emerald);
    cursor: pointer;
    font-weight: 700;
    font-size: 12px;
  }

  .combobox-add-new:hover {
    background-color: rgba(16, 185, 129, 0.2);
  }
</style>
