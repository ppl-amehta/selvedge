<script>
  import { api } from '../api.js';

  let {
    isOpen = $bindable(false)
  } = $props();

  let queryNodeId = $state('');
  let nodeResult = $state(null);
  let traverseResult = $state(null);
  let isLoading = $state(false);
  let errorMsg = $state('');

  async function handleInspect() {
    if (!queryNodeId.trim()) return;
    isLoading = true;
    errorMsg = '';
    nodeResult = null;
    traverseResult = null;

    try {
      const res = await api.getGraphNode(queryNodeId.trim());
      nodeResult = res.data;

      // Also attempt traversal along common edges
      const trav = await api.traverseGraph(queryNodeId.trim(), '');
      traverseResult = trav.data;
    } catch (err) {
      errorMsg = err.message;
    } finally {
      isLoading = false;
    }
  }
</script>

{#if isOpen}
  <div class="modal-overlay">
    <div class="modal-box" style="max-width:780px;">
      <div class="modal-header">
        <h3>Unified Graph Kernel Explorer</h3>
        <button class="btn btn-secondary btn-sm" onclick={() => isOpen = false}>&times;</button>
      </div>
      <div class="modal-body">
        <p style="font-size:12px;color:var(--text-muted);margin-bottom:12px;">
          Inspect real-time bi-temporal graph properties (valid_from/to, txn_from/to) and edge relationships:
        </p>

        <div style="display:flex;gap:8px;margin-bottom:16px;">
          <input
            type="text"
            class="input-control input-mono"
            placeholder="e.g. ord_S261019A or plt_261019-A or ent-client-titan"
            bind:value={queryNodeId}
            onkeydown={(e) => e.key === 'Enter' && handleInspect()}
          />
          <button type="button" class="btn btn-primary" onclick={handleInspect} disabled={isLoading}>
            {isLoading ? 'Querying...' : 'Inspect Node'}
          </button>
        </div>

        {#if errorMsg}
          <div style="color:var(--accent-rose);padding:10px;background:rgba(244,63,94,0.1);border-radius:6px;font-size:12px;">
            {errorMsg}
          </div>
        {/if}

        {#if nodeResult}
          <div style="margin-bottom:14px;">
            <div style="font-size:11px;font-weight:700;color:var(--text-muted);text-transform:uppercase;margin-bottom:4px;">
              Node: <span style="color:var(--accent-primary);">{nodeResult.id}</span> ({nodeResult.node_type})
            </div>
            <pre class="code-block">{JSON.stringify(nodeResult, null, 2)}</pre>
          </div>
        {/if}

        {#if traverseResult && traverseResult.length > 0}
          <div>
            <div style="font-size:11px;font-weight:700;color:var(--text-muted);text-transform:uppercase;margin-bottom:4px;">
              Connected Lineage Edges ({traverseResult.length} targets)
            </div>
            <pre class="code-block" style="color:#93c5fd;">{JSON.stringify(traverseResult, null, 2)}</pre>
          </div>
        {/if}
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" onclick={() => isOpen = false}>Close</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .code-block {
    background: #060911;
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-sm);
    padding: 14px;
    font-family: var(--font-mono);
    font-size: 11.5px;
    max-height: 280px;
    overflow: auto;
    color: #a7f3d0;
  }
</style>
