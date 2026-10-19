<script>
  let {
    personalization = $bindable([])
  } = $props();

  let isPasteModalOpen = $state(false);
  let pasteRawText = $state('');

  const sizes = ['XS', 'S', 'M', 'L', 'XL', '2XL', '3XL', '4XL', '5XL', 'Free'];

  function addRow(size = 'M', name = '', number = '', remarks = '') {
    personalization = [
      ...personalization,
      { size, name, number, remarks }
    ];
  }

  function removeRow(idx) {
    personalization = personalization.filter((_, i) => i !== idx);
  }

  function handleParsePaste() {
    const raw = pasteRawText.trim();
    if (!raw) return;

    const lines = raw.split(/\r?\n/);
    const parsed = [];

    lines.forEach(line => {
      if (!line.trim()) return;
      const parts = line.split('\t').map(p => p.trim());
      parsed.push({
        size: parts[0] || 'M',
        name: parts[1] || '',
        number: parts[2] || '',
        remarks: parts[3] || ''
      });
    });

    personalization = [...personalization, ...parsed];
    pasteRawText = '';
    isPasteModalOpen = false;

    if (window.showToast) {
      window.showToast(`Imported ${parsed.length} player/staff roster records`, 'success');
    }
  }

  $effect(() => {
    if (personalization.length === 0) {
      addRow();
    }
  });
</script>

<div class="roster-box">
  <div class="header">
    <div class="title-group">
      <span class="title">Personalization Roster (Jersey / Corporate Customization)</span>
      <span class="count-badge">{personalization.filter(r => r.name || r.number).length} Customized Items</span>
    </div>
    <button type="button" class="btn btn-secondary btn-sm" onclick={() => isPasteModalOpen = true}>
      📋 Bulk Paste from Excel / Sheets
    </button>
  </div>

  <div class="data-table-container" style="margin-top:12px;">
    <table class="data-table">
      <thead>
        <tr>
          <th style="width:110px;">Size</th>
          <th>Player / Employee Name</th>
          <th style="width:120px;">Jersey No</th>
          <th>Custom Requirements / Logo Placement</th>
          <th style="width:50px;"></th>
        </tr>
      </thead>
      <tbody>
        {#each personalization as row, idx}
          <tr>
            <td>
              <select class="select-control" bind:value={row.size}>
                {#each sizes as s}
                  <option value={s}>{s}</option>
                {/each}
              </select>
            </td>
            <td>
              <input
                type="text"
                class="input-control"
                placeholder="Name printed on back"
                bind:value={row.name}
              />
            </td>
            <td>
              <input
                type="text"
                class="input-control input-mono"
                placeholder="No."
                bind:value={row.number}
              />
            </td>
            <td>
              <input
                type="text"
                class="input-control"
                placeholder="e.g. Captain / Left sleeve sponsor"
                bind:value={row.remarks}
              />
            </td>
            <td>
              <button
                type="button"
                class="btn btn-danger btn-sm"
                onclick={() => removeRow(idx)}
              >
                &times;
              </button>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>

  <button type="button" class="btn btn-secondary btn-sm" onclick={() => addRow()} style="margin-top:12px;">
    + Add Personalization Row
  </button>
</div>

{#if isPasteModalOpen}
  <div class="modal-overlay">
    <div class="modal-box">
      <div class="modal-header">
        <h3>Bulk Paste Personalization from Spreadsheet</h3>
        <button class="btn btn-secondary btn-sm" onclick={() => isPasteModalOpen = false}>&times;</button>
      </div>
      <div class="modal-body">
        <p style="font-size:12px;color:var(--text-muted);margin-bottom:10px;">
          Copy rows directly from Excel or Google Sheets (columns: <strong>[Size] [Tab] [Name] [Tab] [Number] [Tab] [Remarks]</strong>):
        </p>
        <textarea
          class="textarea-control"
          rows="8"
          placeholder="M	Virat Kohli	18	Captain&#10;L	Rohit Sharma	45	Opener&#10;XL	Jasprit Bumrah	93	Fast Bowler"
          bind:value={pasteRawText}
          style="font-family:var(--font-mono);font-size:12px;"
        ></textarea>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => isPasteModalOpen = false}>Cancel</button>
        <button class="btn btn-primary" onclick={handleParsePaste}>Import Rows</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .roster-box {
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    background-color: var(--bg-surface);
    padding: 16px;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .title-group {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .title {
    font-size: 13px;
    font-weight: 700;
    color: var(--text-main);
  }

  .count-badge {
    font-size: 11px;
    font-family: var(--font-mono);
    color: var(--text-subtle);
  }
</style>
