<script>
  let {
    isOpen = $bindable(false),
    bundleCode = '',
    item = '',
    color = '',
    qty = 0,
    uom = 'Kgs',
    location = '',
    qcStatus = 'Passed',
    tagList = []
  } = $props();

  let activeList = $derived.by(() => {
    if (tagList && tagList.length > 0) return tagList;
    return [{ bundleCode, item, color, qty, uom, location, qcStatus }];
  });

  function printTag() {
    window.print();
  }

  // 2D Matrix generator for SVG Zebra label
  function generateRects(code) {
    const size = 25;
    const m = Array(size).fill(0).map(() => Array(size).fill(0));

    // Finder patterns
    drawFinder(m, 0, 0);
    drawFinder(m, 0, size - 7);
    drawFinder(m, size - 7, 0);

    // Timing
    for (let i = 8; i < size - 8; i++) {
      m[6][i] = i % 2 === 0 ? 1 : 0;
      m[i][6] = i % 2 === 0 ? 1 : 0;
    }

    let hash = 2166136261;
    const str = code || 'SELVEDGE-QR';
    for (let i = 0; i < str.length; i++) {
      hash ^= str.charCodeAt(i);
      hash = Math.imul(hash, 16777619);
    }

    let bitIdx = 0;
    for (let r = 0; r < size; r++) {
      for (let c = 0; c < size; c++) {
        if (isReserved(r, c, size)) continue;
        const bit = ((hash >> (bitIdx % 31)) & 1) ^ ((r + c) % 2 === 0 ? 1 : 0);
        m[r][c] = bit;
        bitIdx++;
        if (bitIdx % 31 === 0) hash = Math.imul(hash ^ 0x5bd1e995, 16777619);
      }
    }

    const rList = [];
    const cellSize = 5;
    for (let r = 0; r < size; r++) {
      for (let c = 0; c < size; c++) {
        if (m[r][c] === 1) {
          rList.push({ x: c * cellSize, y: r * cellSize, w: cellSize, h: cellSize });
        }
      }
    }
    return rList;
  }

  function drawFinder(m, top, left) {
    for (let r = 0; r < 7; r++) {
      for (let c = 0; c < 7; c++) {
        if (
          r === 0 || r === 6 || c === 0 || c === 6 ||
          (r >= 2 && r <= 4 && c >= 2 && c <= 4)
        ) {
          m[top + r][left + c] = 1;
        }
      }
    }
  }

  function isReserved(r, c, size) {
    if (r < 8 && c < 8) return true;
    if (r < 8 && c >= size - 8) return true;
    if (r >= size - 8 && c < 8) return true;
    if (r === 6 || c === 6) return true;
    return false;
  }
</script>

{#if isOpen}
  <div class="modal-overlay">
    <div class="modal-box" style="max-width: {activeList.length > 1 ? '780px' : '420px'}; max-height: 90vh; overflow-y: auto;">
      <div class="modal-header">
        <h3>Factory Thermal QR Tags ({activeList.length} Labels)</h3>
        <button class="btn btn-secondary btn-sm" onclick={() => isOpen = false}>&times;</button>
      </div>
      <div class="modal-body" style="padding: 20px;">
        <div class="tags-grid">
          {#each activeList as tag}
            {@const rects = generateRects(tag.bundleCode)}
            <div class="qr-tag-card">
              <div class="qr-tag-header">FPCC APPAREL SYSTEM OF RECORD</div>
              <svg class="qr-svg" viewBox="0 0 125 125" xmlns="http://www.w3.org/2000/svg">
                <rect width="100%" height="100%" fill="#fff" />
                {#each rects as r}
                  <rect x={r.x} y={r.y} width={r.w} height={r.h} fill="#000" />
                {/each}
              </svg>
              <div class="qr-code-text">{tag.bundleCode}</div>
              <div class="qr-meta-text">
                <strong>{tag.item}</strong> {#if tag.color}• {tag.color}{/if}<br/>
                {#if tag.sizeSpec}Size: <strong>{tag.sizeSpec}</strong> • {/if}
                <strong>{tag.qty} {tag.uom || 'Kgs'}</strong> • Bin: <strong>{tag.location || 'A/R01/S01'}</strong><br/>
                QC: <span style="color: {tag.qcStatus === 'Rejected' ? '#ef4444' : '#059669'}; font-weight: 800;">{tag.qcStatus || 'Passed'}</span>
              </div>
            </div>
          {/each}
        </div>
      </div>
      <div class="modal-footer" style="justify-content: space-between;">
        <button class="btn btn-secondary" onclick={() => isOpen = false}>Close</button>
        <button class="btn btn-primary" onclick={printTag}>🖨️ Print Zebra Thermal Tags ({activeList.length})</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .tags-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
    justify-content: center;
  }

  .qr-tag-card {
    background: #fff;
    color: #000;
    border: 2px solid #000;
    border-radius: var(--radius-md);
    padding: 14px;
    display: inline-flex;
    flex-direction: column;
    align-items: center;
    font-family: var(--font-mono);
    width: 200px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  }

  .qr-tag-header {
    font-size: 9px;
    font-weight: 900;
    letter-spacing: 0.1em;
    border-bottom: 2px solid #000;
    width: 100%;
    text-align: center;
    padding-bottom: 4px;
    margin-bottom: 8px;
  }

  .qr-svg {
    width: 120px;
    height: 120px;
  }

  .qr-code-text {
    font-size: 13px;
    font-weight: 900;
    margin-top: 6px;
    letter-spacing: 0.05em;
  }

  .qr-meta-text {
    font-size: 10px;
    line-height: 1.35;
    margin-top: 4px;
    text-align: center;
    word-break: break-word;
  }

  @media print {
    body * {
      visibility: hidden;
    }
    .tags-grid, .tags-grid * {
      visibility: visible;
    }
    .tags-grid {
      position: absolute;
      left: 0;
      top: 0;
      display: block;
    }
    .qr-tag-card {
      page-break-after: always;
      margin-bottom: 20px;
    }
  }
</style>
