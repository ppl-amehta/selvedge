<script>
  let {
    isOpen = $bindable(false),
    bundleCode = '',
    item = '',
    color = '',
    qty = 0,
    uom = 'Kgs',
    location = '',
    qcStatus = 'Passed'
  } = $props();

  function printTag() {
    window.print();
  }

  // 2D Matrix generator for SVG Zebra label
  let rects = $derived.by(() => {
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
    for (let i = 0; i < bundleCode.length; i++) {
      hash ^= bundleCode.charCodeAt(i);
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
  });

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
    <div class="modal-box" style="max-width:420px;">
      <div class="modal-header">
        <h3>Factory Thermal QR Tag</h3>
        <button class="btn btn-secondary btn-sm" onclick={() => isOpen = false}>&times;</button>
      </div>
      <div class="modal-body" style="text-align:center;padding:24px;">
        <div class="qr-tag-card">
          <div class="qr-tag-header">FPCC APPAREL SYSTEM OF RECORD</div>
          <svg class="qr-svg" viewBox="0 0 125 125" xmlns="http://www.w3.org/2000/svg">
            <rect width="100%" height="100%" fill="#fff" />
            {#each rects as r}
              <rect x={r.x} y={r.y} width={r.w} height={r.h} fill="#000" />
            {/each}
          </svg>
          <div class="qr-code-text">{bundleCode}</div>
          <div class="qr-meta-text">
            <strong>{item}</strong> {#if color}• {color}{/if}<br/>
            <strong>{qty} {uom}</strong> • Bin: <strong>{location || 'A/R01/S01'}</strong><br/>
            QC: <span style="color:#059669;font-weight:800;">{qcStatus}</span>
          </div>
        </div>
      </div>
      <div class="modal-footer" style="justify-content:space-between;">
        <button class="btn btn-secondary" onclick={() => isOpen = false}>Close</button>
        <button class="btn btn-primary" onclick={printTag}>🖨️ Print Zebra Tag</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .qr-tag-card {
    background: #fff;
    color: #000;
    border: 2px solid #000;
    border-radius: var(--radius-md);
    padding: 16px;
    display: inline-flex;
    flex-direction: column;
    align-items: center;
    font-family: var(--font-mono);
  }

  .qr-tag-header {
    font-size: 11px;
    font-weight: 900;
    letter-spacing: 0.1em;
    border-bottom: 2px solid #000;
    width: 100%;
    text-align: center;
    padding-bottom: 4px;
    margin-bottom: 8px;
  }

  .qr-svg {
    width: 140px;
    height: 140px;
  }

  .qr-code-text {
    font-size: 15px;
    font-weight: 900;
    margin-top: 6px;
    letter-spacing: 0.05em;
  }

  .qr-meta-text {
    font-size: 10.5px;
    line-height: 1.35;
    margin-top: 4px;
    text-align: center;
  }
</style>
