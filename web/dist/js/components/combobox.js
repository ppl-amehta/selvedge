// FPCC Selvedge Modern Auto-Suggest Combobox with Open-Ended Registration
import { api } from '../api.js';

export function createCombobox({
  container,
  category,
  placeholder = 'Type or search...',
  items = [],
  displayKey = 'display_name',
  valueKey = 'display_name',
  metaKey = null,
  allowCustom = true,
  onSelect = () => {},
  initialValue = ''
}) {
  let currentItems = [...items];
  let selectedValue = initialValue;
  let highlightedIndex = -1;

  container.classList.add('combobox-wrapper');
  container.innerHTML = `
    <div class="combobox-input-group">
      <input type="text" class="input-control combobox-input" placeholder="${placeholder}" value="${initialValue}" autocomplete="off" />
      <button type="button" class="combobox-toggle" title="Toggle list">▼</button>
    </div>
    <div class="combobox-dropdown"></div>
  `;

  const input = container.querySelector('.combobox-input');
  const toggleBtn = container.querySelector('.combobox-toggle');
  const dropdown = container.querySelector('.combobox-dropdown');

  function renderDropdown(filteredList, query = '') {
    dropdown.innerHTML = '';
    highlightedIndex = -1;

    if (filteredList.length === 0 && !query) {
      dropdown.innerHTML = `<div class="combobox-item" style="color:var(--text-subtle);">No items available</div>`;
    } else {
      filteredList.slice(0, 50).forEach((item, index) => {
        const div = document.createElement('div');
        div.className = 'combobox-item';
        div.dataset.index = index;

        const labelText = item[displayKey] || item.name || item.code || '';
        const metaText = metaKey && item[metaKey] ? item[metaKey] : (item.code && item.code !== labelText ? item.code : '');

        div.innerHTML = `
          <span>${highlightMatch(labelText, query)}</span>
          ${metaText ? `<span class="item-meta">${metaText}</span>` : ''}
        `;

        div.addEventListener('click', () => {
          selectItem(item);
        });

        dropdown.appendChild(div);
      });
    }

    // Open-ended new entry registration
    const trimmed = query.trim();
    const exactMatch = currentItems.some(i => (i[displayKey] || i.name || '').toLowerCase() === trimmed.toLowerCase());

    if (allowCustom && trimmed && !exactMatch) {
      const addDiv = document.createElement('div');
      addDiv.className = 'combobox-add-new';
      addDiv.innerHTML = `<span>+ Register "<strong>${escapeHtml(trimmed)}</strong>" as new ${category}</span>`;
      addDiv.addEventListener('click', async () => {
        await handleRegisterNew(trimmed);
      });
      dropdown.appendChild(addDiv);
    }

    dropdown.classList.add('open');
  }

  function highlightMatch(text, query) {
    if (!query) return escapeHtml(text);
    const escaped = escapeHtml(text);
    const regex = new RegExp(`(${query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
    return escaped.replace(regex, '<strong style="color:#60a5fa;text-decoration:underline;">$1</strong>');
  }

  function escapeHtml(str) {
    return String(str).replace(/[&<>"']/g, m => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[m]);
  }

  async function handleRegisterNew(name) {
    try {
      const res = await api.registerMaster(category, name);
      const newItem = {
        id: res.id,
        [displayKey]: res.display_name || res.name || name,
        [valueKey]: res.display_name || res.name || name,
      };
      currentItems.push(newItem);
      selectItem(newItem);
      if (window.showToast) {
        window.showToast(`Registered new ${category}: "${name}"`, 'success');
      }
    } catch (err) {
      if (window.showToast) {
        window.showToast(`Failed to register ${category}: ${err.message}`, 'error');
      }
    }
  }

  function selectItem(item) {
    const val = item[valueKey] || item[displayKey] || item.name || '';
    selectedValue = val;
    input.value = val;
    dropdown.classList.remove('open');
    onSelect(item);
  }

  function filterItems(query) {
    const q = query.toLowerCase().trim();
    if (!q) return currentItems;
    return currentItems.filter(item => {
      const text = (item[displayKey] || item.name || item.code || '').toLowerCase();
      const code = (item.code || '').toLowerCase();
      return text.includes(q) || code.includes(q);
    });
  }

  input.addEventListener('focus', () => {
    const filtered = filterItems(input.value);
    renderDropdown(filtered, input.value);
  });

  input.addEventListener('input', () => {
    selectedValue = input.value;
    const filtered = filterItems(input.value);
    renderDropdown(filtered, input.value);
  });

  input.addEventListener('keydown', (e) => {
    const itemsElements = dropdown.querySelectorAll('.combobox-item, .combobox-add-new');
    if (!dropdown.classList.contains('open') || itemsElements.length === 0) {
      if (e.key === 'ArrowDown') {
        renderDropdown(filterItems(input.value), input.value);
      }
      return;
    }

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      highlightedIndex = Math.min(highlightedIndex + 1, itemsElements.length - 1);
      updateHighlight(itemsElements);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      highlightedIndex = Math.max(highlightedIndex - 1, 0);
      updateHighlight(itemsElements);
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (highlightedIndex >= 0 && itemsElements[highlightedIndex]) {
        itemsElements[highlightedIndex].click();
      }
    } else if (e.key === 'Escape') {
      dropdown.classList.remove('open');
    }
  });

  function updateHighlight(elements) {
    elements.forEach((el, idx) => {
      if (idx === highlightedIndex) {
        el.classList.add('highlighted');
        el.scrollIntoView({ block: 'nearest' });
      } else {
        el.classList.remove('highlighted');
      }
    });
  }

  toggleBtn.addEventListener('click', (e) => {
    e.stopPropagation();
    if (dropdown.classList.contains('open')) {
      dropdown.classList.remove('open');
    } else {
      input.focus();
      renderDropdown(currentItems, '');
    }
  });

  document.addEventListener('click', (e) => {
    if (!container.contains(e.target)) {
      dropdown.classList.remove('open');
    }
  });

  return {
    getValue() {
      return input.value.trim();
    },
    setValue(val) {
      selectedValue = val;
      input.value = val;
    },
    setItems(newItems) {
      currentItems = [...newItems];
    },
    clear() {
      selectedValue = '';
      input.value = '';
    }
  };
}
