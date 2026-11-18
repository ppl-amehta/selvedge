// FPCC Selvedge Unified API Client
const API_BASE = '/api';

export const api = {
  async health() {
    const res = await fetch(`${API_BASE}/health`);
    return res.json();
  },

  async getMasters() {
    const res = await fetch(`${API_BASE}/masters/all`);
    const json = await res.json();
    return json.data || {};
  },

  async registerMaster(type, value, meta = {}) {
    const res = await fetch(`${API_BASE}/masters/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ type, value, meta }),
    });
    return res.json();
  },

  async getNextOrderNo() {
    const res = await fetch(`${API_BASE}/orders/next-order-no`);
    const json = await res.json();
    return json.nextOrderNo || '';
  },

  async fetchZohoOrder(salesOrderNo) {
    const res = await fetch(`${API_BASE}/orders/zoho/fetch?salesOrderNo=${encodeURIComponent(salesOrderNo)}`);
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to fetch sales order from Zoho');
    }
    return json.data;
  },

  async uploadFile(file) {
    const formData = new FormData();
    formData.append('file', file);
    const res = await fetch(`${API_BASE}/upload`, {
      method: 'POST',
      body: formData,
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'File upload failed');
    }
    return json;
  },

  async createOrder(payload) {
    const res = await fetch(`${API_BASE}/orders`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to create production order');
    }
    return json;
  },

  async getOrder(orderNo) {
    const res = await fetch(`${API_BASE}/orders/${encodeURIComponent(orderNo)}`);
    return res.json();
  },

  async listOrders() {
    const res = await fetch(`${API_BASE}/orders`);
    const json = await res.json();
    return json.data || [];
  },

  async getNextParentCode() {
    const res = await fetch(`${API_BASE}/inventory/next-parent-code`);
    const json = await res.json();
    return json.nextParentCode || '';
  },

  async searchInvoices(query = '') {
    const url = query ? `${API_BASE}/inventory/invoices?q=${encodeURIComponent(query)}` : `${API_BASE}/inventory/invoices`;
    const res = await fetch(url);
    const json = await res.json();
    return json.data || [];
  },

  async getInvoice(invoiceNo) {
    const res = await fetch(`${API_BASE}/inventory/invoices/${encodeURIComponent(invoiceNo)}`);
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || `Failed to fetch invoice ${invoiceNo}`);
    }
    return json.data;
  },

  async inwardParentLot(payload) {
    const res = await fetch(`${API_BASE}/inventory/inward`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to inward parent lot');
    }
    return json;
  },

  async gateInward(payload) {
    const res = await fetch(`${API_BASE}/inventory/gate-inward`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to submit gate inward');
    }
    return json;
  },

  async recordFloorCount(payload) {
    const res = await fetch(`${API_BASE}/inventory/floor-count`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to record floor count');
    }
    return json;
  },

  async recordBundleQC(payload) {
    const res = await fetch(`${API_BASE}/inventory/qc-inspect`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to submit bundle QC inspection');
    }
    return json;
  },

  async resolveDiscrepancy(parentCode, action, remarks = '') {
    const res = await fetch(`${API_BASE}/inventory/resolve-discrepancy`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ parentCode, action, remarks }),
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to resolve discrepancy');
    }
    return json;
  },

  async listLots(stage = '') {
    const url = stage ? `${API_BASE}/inventory/lots?stage=${encodeURIComponent(stage)}` : `${API_BASE}/inventory/lots`;
    const res = await fetch(url);
    const json = await res.json();
    return json.data || [];
  },

  async getLot(parentCode) {
    const res = await fetch(`${API_BASE}/inventory/lots/${encodeURIComponent(parentCode)}`);
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || `Failed to fetch lot ${parentCode}`);
    }
    return json.data;
  },

  async returnRemnant(payload) {
    const res = await fetch(`${API_BASE}/inventory/return-remnant`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to return remnant');
    }
    return json;
  },

  async listDiscrepancies() {
    const res = await fetch(`${API_BASE}/inventory/discrepancies`);
    const json = await res.json();
    return json.data || [];
  },

  async getGraphNode(nodeId) {
    const res = await fetch(`${API_BASE}/graph/nodes/${encodeURIComponent(nodeId)}`);
    return res.json();
  },

  async traverseGraph(start, edge, direction = 'OUT') {
    const params = new URLSearchParams({ start, edge: edge || '', direction });
    const res = await fetch(`${API_BASE}/graph/traverse?${params.toString()}`);
    return res.json();
  },

  async dispatchWorkflows() {
    const res = await fetch(`${API_BASE}/workflows/dispatch`, { method: 'POST' });
    return res.json();
  },

  // --- Station 2 & 3: Order Preparation & Planning ---
  async listOrdersForPrep() {
    const res = await fetch(`${API_BASE}/planning/orders`);
    const json = await res.json();
    return json.data || [];
  },

  async getOrderPrepAndPlan(orderId) {
    const res = await fetch(`${API_BASE}/planning/orders/${encodeURIComponent(orderId)}`);
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || `Failed to fetch prep details for order ${orderId}`);
    }
    return json.data;
  },

  async savePharma(payload) {
    const res = await fetch(`${API_BASE}/planning/pharma`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to save pharma tech pack');
    }
    return json.data;
  },

  async saveBomLine(payload) {
    const res = await fetch(`${API_BASE}/planning/bom`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to save BOM line');
    }
    return json.data;
  },

  async deleteBomLine(lineId) {
    const res = await fetch(`${API_BASE}/planning/bom/${encodeURIComponent(lineId)}`, {
      method: 'DELETE'
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to delete BOM line');
    }
    return json;
  },

  async raiseProcurementPO(payload) {
    const res = await fetch(`${API_BASE}/planning/raise-po`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to raise procurement requisition');
    }
    return json.data;
  },

  async saveRoutingSteps(orderId, steps) {
    const res = await fetch(`${API_BASE}/planning/routing-steps`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ orderId, steps })
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to save routing steps');
    }
    return json.data;
  },

  async advanceOrderState(orderId, targetState, reason = '') {
    const res = await fetch(`${API_BASE}/planning/advance-state`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ orderId, targetState, reason })
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || `Failed to advance order state to ${targetState}`);
    }
    return json.data;
  },

  // --- Station 4: Floor Cutting Station ---
  async listCuttingOrders() {
    const res = await fetch(`${API_BASE}/cutting/orders`);
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to list cutting orders');
    }
    return json.data || [];
  },

  async getCuttingJob(idOrOrderNo) {
    const res = await fetch(`${API_BASE}/cutting/job/${encodeURIComponent(idOrOrderNo)}`);
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || `Failed to fetch cutting job ${idOrOrderNo}`);
    }
    return json.data;
  },

  async issueRollToJob(payload) {
    const res = await fetch(`${API_BASE}/cutting/issue-roll`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to issue roll to cutting job');
    }
    return json;
  },

  async recordCutLay(payload) {
    const res = await fetch(`${API_BASE}/cutting/record-lay`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to record cut lay');
    }
    return json.data;
  },

  async recordCuttingRemnant(payload) {
    const res = await fetch(`${API_BASE}/cutting/record-remnant`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to record remnant roll');
    }
    return json.data;
  },

  async getBundleSlip(bundleId) {
    const res = await fetch(`${API_BASE}/cutting/bundle-slip/${encodeURIComponent(bundleId)}`, {
      method: 'POST'
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || `Failed to fetch bundle slip ${bundleId}`);
    }
    return json.data;
  },

  // --- Sample Development & Product Creation ---

  async getNextDevelopmentNo() {
    const res = await fetch(`${API_BASE}/samples/next-development-no`);
    const json = await res.json();
    return json.nextDevelopmentNo || '';
  },

  async createSampleDevelopment(payload) {
    const res = await fetch(`${API_BASE}/samples`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to save sample development');
    }
    return json;
  },

  async getSampleDevelopment(id) {
    const res = await fetch(`${API_BASE}/samples/${encodeURIComponent(id)}`);
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Sample development not found');
    }
    return json.data;
  },

  async listSampleDevelopments(limit = 50, offset = 0) {
    const res = await fetch(`${API_BASE}/samples?limit=${limit}&offset=${offset}`);
    const json = await res.json();
    return json.data || [];
  },

  async updateSampleStatus(id, status, reworkNotes = '') {
    const res = await fetch(`${API_BASE}/samples/${encodeURIComponent(id)}/status`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ status, reworkNotes })
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to update sample status');
    }
    return json;
  },

  async suggestSKU(brandLabel, genre, productName) {
    const url = `${API_BASE}/products/suggest-sku?brandLabel=${encodeURIComponent(brandLabel || '')}&genre=${encodeURIComponent(genre || '')}&productName=${encodeURIComponent(productName || '')}`;
    const res = await fetch(url);
    const json = await res.json();
    return json.skuCode || '';
  },

  async createMasterProduct(payload) {
    const res = await fetch(`${API_BASE}/products`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to create master product');
    }
    return json;
  },

  async getMasterProduct(id) {
    const res = await fetch(`${API_BASE}/products/${encodeURIComponent(id)}`);
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Master product not found');
    }
    return json.data;
  },

  async listMasterProducts(limit = 50, offset = 0) {
    const res = await fetch(`${API_BASE}/products?limit=${limit}&offset=${offset}`);
    const json = await res.json();
    return json.data || [];
  },

  async setClientProductPricing(payload) {
    const res = await fetch(`${API_BASE}/client-pricing`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    const json = await res.json();
    if (!res.ok || json.success === false) {
      throw new Error(json.error || 'Failed to set client pricing');
    }
    return json;
  },

  async listClientProductPricing(clientId = '') {
    const url = clientId ? `${API_BASE}/client-pricing?clientId=${encodeURIComponent(clientId)}` : `${API_BASE}/client-pricing`;
    const res = await fetch(url);
    const json = await res.json();
    return json.data || [];
  }
};
