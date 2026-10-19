// FPCC Selvedge Enterprise Edge Engine - App Main Orchestrator
import { api } from './api.js';
import { initOrdersView } from './views/orders.js';
import { initInwardView } from './views/inward.js';

// Global Toast System
window.showToast = function(message, type = 'info') {
  let container = document.getElementById('toast-container');
  if (!container) {
    container = document.createElement('div');
    container.id = 'toast-container';
    container.className = 'toast-container';
    document.body.appendChild(container);
  }

  const toast = document.createElement('div');
  toast.className = `toast ${type}`;
  toast.innerHTML = `
    <span>${type === 'success' ? '✓' : type === 'error' ? '✕' : 'ℹ'}</span>
    <div>${message}</div>
  `;
  container.appendChild(toast);

  setTimeout(() => {
    toast.style.opacity = '0';
    toast.style.transform = 'translateX(100%)';
    toast.style.transition = 'all 0.3s ease';
    setTimeout(() => toast.remove(), 300);
  }, 4000);
};

document.addEventListener('DOMContentLoaded', async () => {
  console.log('⚡ Initializing FPCC Selvedge Enterprise Edge UI...');

  try {
    // 1. Check System Health
    const health = await api.health();
    document.getElementById('sys-status-text').textContent = 'ONLINE (Go 1.27 SQLite WAL)';

    // 2. Fetch Master Datasets
    const masters = await api.getMasters();
    console.log('✓ Master catalogs loaded:', Object.keys(masters).map(k => `${k}: ${masters[k] ? masters[k].length : 0}`).join(', '));

    // 3. Initialize Views
    const ordersContainer = document.getElementById('view-orders');
    const inwardContainer = document.getElementById('view-inward');

    initOrdersView({
      container: ordersContainer,
      masters,
      onOrderBooked: (res) => {
        console.log('Order booked into graph:', res);
      }
    });

    initInwardView({
      container: inwardContainer,
      masters,
      onInwardCompleted: (res) => {
        console.log('Inward lot committed to graph:', res);
      }
    });

    // 4. Tab Navigation Switching
    const navButtons = document.querySelectorAll('.station-btn');
    const viewSections = document.querySelectorAll('.view-section');

    navButtons.forEach(btn => {
      btn.addEventListener('click', () => {
        const targetView = btn.dataset.view;
        navButtons.forEach(b => b.classList.remove('active'));
        btn.classList.add('active');

        viewSections.forEach(sec => {
          if (sec.id === targetView) {
            sec.classList.add('active');
          } else {
            sec.classList.remove('active');
          }
        });
      });
    });

    // 5. Unified Graph Inspector Modal
    const graphModal = document.getElementById('modal-graph-inspector');
    const btnOpenGraph = document.getElementById('btn-open-graph');
    if (btnOpenGraph && graphModal) {
      btnOpenGraph.addEventListener('click', () => {
        graphModal.classList.add('open');
      });

      document.getElementById('btn-query-graph').addEventListener('click', async () => {
        const nodeId = document.getElementById('inp-graph-node-id').value.trim();
        const display = document.getElementById('graph-inspector-result');
        if (!nodeId) return;

        display.textContent = 'Querying graph kernel...';
        try {
          const res = await api.getGraphNode(nodeId);
          display.textContent = JSON.stringify(res.data, null, 2);
        } catch (err) {
          display.textContent = `Graph query error: ${err.message}`;
        }
      });
    }

    console.log('✓ Selvedge Station 0 and Station 1 UI ready.');
  } catch (err) {
    console.error('Initialization error:', err);
    window.showToast(`System initialization error: ${err.message}`, 'error');
  }
});
