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
  }
};
