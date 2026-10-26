-- ========================================================
-- FPCC SELVEDGE: ORDER PREPARATION & ORDER PLANNING
-- Station 2 (Pharma Tech Packs, BOM Engine, SOH Shortfall)
-- Station 3 (Routing Steps, Vendor Scheduling, Gantt Timeline)
-- ========================================================

-- 1. Pharma / Pattern Tech Pack Specifications per Order Item
CREATE TABLE IF NOT EXISTS pharma_patterns (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    order_item_id TEXT NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
    pharma_type TEXT NOT NULL,                     -- e.g. 'Body', 'Collar', 'Sleeve', 'Neck Collar', 'Placket', 'Pocket', 'Cuff', 'Rib'
    pharma_no TEXT NOT NULL,                       -- Pattern / CAD cut tech pack reference e.g. 'PH-POLO-2026-01'
    changes_to_pharma TEXT,                        -- Open remarks for pattern alterations or grading adjustments
    measurement_chart_json JSON,                   -- Array of {point, tolerance, xs, s, m, l, xl, xxl, xxxl}
    status TEXT NOT NULL DEFAULT 'APPROVED',       -- 'DRAFT', 'APPROVED', 'SUPERSEDED'
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_pharma_item ON pharma_patterns(order_item_id);
CREATE INDEX IF NOT EXISTS idx_pharma_order ON pharma_patterns(order_id);

-- 2. Bill of Materials (BOM) Lines per Order Item
CREATE TABLE IF NOT EXISTS bom_lines (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    order_item_id TEXT NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
    line_type TEXT NOT NULL CHECK(line_type IN ('FABRIC', 'ACCESSORY', 'PACKAGING', 'LABEL')),
    item_name TEXT NOT NULL,                       -- Combobox master name (e.g. 'Super Combed Pique 220', 'Vislon Zip 24"')
    sub_type TEXT,                                 -- 'Main Body', 'Collar Rib', 'Zipper', 'Button', 'Polybag', etc.
    consumption_avg REAL NOT NULL DEFAULT 0.0,     -- Per-piece consumption (e.g. 0.35 kg/pc or 1 pc)
    wastage_pct REAL NOT NULL DEFAULT 0.0,         -- Wastage percentage buffer (e.g. 3.0%)
    uom TEXT NOT NULL,                             -- 'Kgs', 'Pieces', 'Meters', 'Gross', 'Sets'
    unit_qty INTEGER NOT NULL DEFAULT 0,           -- Item total order quantity
    required_qty REAL NOT NULL DEFAULT 0.0,        -- Total computed: unit_qty * consumption_avg * (1 + wastage_pct/100)
    soh_qty REAL NOT NULL DEFAULT 0.0,             -- Live Stock On Hand snapshot from Inward warehouse
    allocated_qty REAL NOT NULL DEFAULT 0.0,       -- Reserved quantity from available stock
    shortfall_qty REAL NOT NULL DEFAULT 0.0,       -- Deficit shortfall: max(0, required_qty - soh_qty)
    rate_per_unit REAL NOT NULL DEFAULT 0.0,       -- Estimated unit cost
    comments TEXT,
    status TEXT NOT NULL DEFAULT 'ACTIVE',         -- 'ACTIVE', 'SUPERSEDED', 'CANCELLED'
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_bom_item ON bom_lines(order_item_id);
CREATE INDEX IF NOT EXISTS idx_bom_order ON bom_lines(order_id);
CREATE INDEX IF NOT EXISTS idx_bom_type ON bom_lines(line_type);

-- 3. Procurement Requisitions for Material Shortfalls
CREATE TABLE IF NOT EXISTS procurement_requisitions (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    bom_line_id TEXT NOT NULL REFERENCES bom_lines(id) ON DELETE CASCADE,
    material_name TEXT NOT NULL,
    material_type TEXT NOT NULL,                   -- 'FABRIC', 'ACCESSORY', 'PACKAGING'
    uom TEXT NOT NULL,
    required_qty REAL NOT NULL,
    order_buffer_qty REAL NOT NULL DEFAULT 0.0,
    total_po_qty REAL NOT NULL,
    vendor_name TEXT,                              -- Suggested procurement vendor
    estimated_rate REAL NOT NULL DEFAULT 0.0,
    estimated_amount REAL NOT NULL DEFAULT 0.0,
    required_by_date TEXT,                         -- Target date (must precede scheduled Cutting start)
    po_no TEXT,                                    -- Purchase Order reference code
    status TEXT NOT NULL DEFAULT 'DRAFT',          -- 'DRAFT', 'PO_SENT', 'RECEIVED', 'CANCELLED'
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_proc_order ON procurement_requisitions(order_id);
CREATE INDEX IF NOT EXISTS idx_proc_bom ON procurement_requisitions(bom_line_id);

-- 4. Order Routing & Production Scheduling Steps (Station 3)
CREATE TABLE IF NOT EXISTS routing_steps (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    order_item_id TEXT REFERENCES order_items(id) ON DELETE CASCADE,
    sequence_order INTEGER NOT NULL,               -- Gantt order index (1, 2, 3...)
    process_code TEXT NOT NULL,                    -- 'Cutting', 'Branding', 'Machine', 'Finishing', 'Packing'
    step_kind TEXT NOT NULL DEFAULT 'INTERNAL' CHECK(step_kind IN ('INTERNAL', 'VENDOR')),
    vendor_name TEXT,                              -- In-house Line (e.g. 'Line 1') or Job-Work Vendor
    duration_days INTEGER NOT NULL DEFAULT 1,
    planned_start_date TEXT NOT NULL,              -- YYYY-MM-DD
    planned_end_date TEXT NOT NULL,                -- YYYY-MM-DD
    planned_qty INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'PLANNED',        -- 'PLANNED', 'IN_PROGRESS', 'COMPLETED', 'DELAYED'
    depends_on_step_id TEXT REFERENCES routing_steps(id),
    comments TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_routing_order ON routing_steps(order_id);
CREATE INDEX IF NOT EXISTS idx_routing_item ON routing_steps(order_item_id);

-- 5. Order Planning Snapshots (Gantt vs Client ETD Comparison)
CREATE TABLE IF NOT EXISTS order_planning_snapshots (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    snapshot_version INTEGER NOT NULL DEFAULT 1,
    client_etd TEXT NOT NULL,
    planned_finish_date TEXT NOT NULL,
    slack_days INTEGER NOT NULL,                   -- Client ETD - Planned Finish Date (+ve = buffer, -ve = delay risk)
    is_overdue_risk INTEGER NOT NULL DEFAULT 0,    -- 1 if planned_finish_date > client_etd
    critical_path_summary TEXT,
    approved_by TEXT,
    created_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_plan_snap_order ON order_planning_snapshots(order_id);
