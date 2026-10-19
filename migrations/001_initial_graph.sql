-- ========================================================
-- FPCC SELVEDGE: UNIFIED GRAPH & BI-TEMPORAL SCHEMA
-- 10-20 Year Industrial System of Record
-- ========================================================

-- 1. UNIFIED GRAPH NODES (Append-Only SGi Pattern)
-- Current valid nodes have txn_to IS NULL and valid_to IS NULL
CREATE TABLE IF NOT EXISTS graph_nodes (
    id TEXT NOT NULL,
    node_type TEXT NOT NULL,                -- e.g. 'CLIENT', 'VENDOR', 'ORDER', 'PARENT_LOT', 'BUNDLE', 'FABRIC', 'COLOR'
    props_json JSON NOT NULL,               -- Dynamic entity payload & domain attributes
    valid_from INTEGER NOT NULL,            -- Valid-time: when fact occurred in real world (Unix ms)
    valid_to INTEGER,                       -- Valid-time: when fact ended in real world (Unix ms)
    txn_from INTEGER NOT NULL,              -- Transaction-time: when system recorded the fact (Unix ms)
    txn_to INTEGER,                         -- Transaction-time: when system superseded the fact (NULL = current)
    PRIMARY KEY (id, txn_from)
);

CREATE INDEX IF NOT EXISTS idx_cur_nodes ON graph_nodes(node_type, id) WHERE txn_to IS NULL;
CREATE INDEX IF NOT EXISTS idx_nodes_temporal ON graph_nodes(txn_from, txn_to);

-- 2. UNIFIED GRAPH EDGES (Typed Relationships & Lineage)
-- Examples: (ORDER)-[:ORDERED_BY]->(CLIENT), (BUNDLE)-[:CUT_FROM]->(PARENT_LOT)
CREATE TABLE IF NOT EXISTS graph_edges (
    id TEXT NOT NULL,
    src_id TEXT NOT NULL,                   -- Source Node ID
    dst_id TEXT NOT NULL,                   -- Destination Node ID
    edge_type TEXT NOT NULL,                -- e.g. 'ORDERED_BY', 'SUPPLIED_BY', 'PART_OF_LOT', 'STORED_AT', 'ISSUED_TO'
    props_json JSON,                        -- Edge properties (e.g. quantity, role, sequence)
    valid_from INTEGER NOT NULL,
    valid_to INTEGER,
    txn_from INTEGER NOT NULL,
    txn_to INTEGER,                         -- NULL = current active edge
    PRIMARY KEY (id, txn_from)
);

CREATE INDEX IF NOT EXISTS idx_cur_edges ON graph_edges(edge_type, src_id, dst_id) WHERE txn_to IS NULL;
CREATE INDEX IF NOT EXISTS idx_edges_src ON graph_edges(src_id) WHERE txn_to IS NULL;
CREATE INDEX IF NOT EXISTS idx_edges_dst ON graph_edges(dst_id) WHERE txn_to IS NULL;

-- 3. TRANSACTIONAL EVENT OUTBOX (Change Data Capture & Automation Trigger Feed)
CREATE TABLE IF NOT EXISTS domain_events (
    id TEXT PRIMARY KEY,
    aggregate_id TEXT NOT NULL,             -- Node or Order ID
    event_type TEXT NOT NULL,               -- e.g. 'ORDER_BOOKED', 'LOT_INWARDED', 'DISCREPANCY_FLAGGED', 'ROLL_RETURNED'
    payload_json JSON NOT NULL,             -- Event snapshot
    depth INTEGER NOT NULL DEFAULT 0,       -- Chained automation depth guard (max 10)
    occurred_at INTEGER NOT NULL,
    dispatched_at INTEGER                   -- NULL = pending dispatch
);

CREATE INDEX IF NOT EXISTS idx_events_pending ON domain_events(dispatched_at, occurred_at);

-- 4. SMART WORKFLOW RULES (Rippling-Style Declarative Automations)
CREATE TABLE IF NOT EXISTS workflow_rules (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    trigger_event TEXT NOT NULL,            -- Event type triggering the rule
    condition_json JSON,                    -- Predicates e.g. {"field": "discrepancy_qty", "op": ">", "val": 0}
    actions_json JSON NOT NULL,             -- Actions array e.g. [{"action": "NOTIFY", "role": "merchandiser"}]
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL
);

-- 5. WORKFLOW ACTION EFFECTS (Idempotent Exactly-Once Execution Table)
CREATE TABLE IF NOT EXISTS action_effects (
    event_id TEXT NOT NULL,
    rule_id TEXT NOT NULL,
    action_idx INTEGER NOT NULL,
    status TEXT NOT NULL,                   -- 'EXECUTED', 'FAILED', 'WAITING_APPROVAL'
    result_json JSON,
    executed_at INTEGER NOT NULL,
    PRIMARY KEY (event_id, rule_id, action_idx)
);

-- ========================================================
-- TABULAR PROJECTIONS (High-Speed Hot Paths for Floor Stations)
-- Maintained atomically in sync with graph mutations
-- ========================================================

-- Production Orders Projection
CREATE TABLE IF NOT EXISTS orders (
    id TEXT PRIMARY KEY,
    production_order_no TEXT UNIQUE NOT NULL,
    sales_order_no TEXT,
    order_type TEXT NOT NULL,               -- Strict enum: 'Bulk', 'Sample', 'FOC', 'Repeat', 'Development'
    booking_date TEXT NOT NULL,
    client_id TEXT NOT NULL,
    client_name TEXT NOT NULL,
    contact_person TEXT,
    phone TEXT,
    email TEXT,
    billing_address TEXT,
    shipping_address TEXT,
    gstin TEXT,
    category TEXT NOT NULL,                 -- Strict enum: 'Apparel', 'Non-Apparel'
    sub_category TEXT NOT NULL,             -- Modern Combobox: 'Polo', 'Hoodie', etc.
    product_name TEXT NOT NULL,             -- Modern Combobox from master
    color_name TEXT NOT NULL,               -- Modern Combobox from master
    fabric_name TEXT,                       -- Modern Combobox from master
    gender TEXT NOT NULL,                   -- Strict enum: 'Men', 'Women', 'Unisex', etc.
    unit_price REAL NOT NULL DEFAULT 0.0,
    tax_rate REAL NOT NULL DEFAULT 5.0,     -- Strict enum: 5, 12, 18
    tax_billing TEXT NOT NULL DEFAULT 'Extra', -- Strict enum: 'Extra', 'Included'
    freight_mode TEXT NOT NULL DEFAULT 'Surface', -- Strict enum: 'Surface', 'Air', 'Courier'
    freight_cost TEXT NOT NULL DEFAULT 'Extra', -- Strict enum: 'Inclusive', 'Extra', 'To be paid by client'
    total_qty INTEGER NOT NULL DEFAULT 0,
    size_breakdown_json JSON,               -- e.g. {"S": 100, "M": 200, "L": 150}
    multiple_shipping INTEGER DEFAULT 0,
    shipping_locations_json JSON,           -- Array of {destination, quantity, carrier, docketNo, etd}
    personalization_json JSON,              -- Array of {size, name, number, remarks}
    branding_type TEXT,
    branding_file_url TEXT,
    main_label TEXT,
    other_label TEXT,
    cust_packaging TEXT,
    client_comment TEXT,
    target_etd TEXT,
    actual_disp_date TEXT,
    carrier TEXT,
    docket_no TEXT,
    status TEXT NOT NULL DEFAULT 'BOOKED',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Inward Parent Lots Projection (Step A)
CREATE TABLE IF NOT EXISTS parent_inward_lots (
    id TEXT PRIMARY KEY,
    parent_code TEXT UNIQUE NOT NULL,       -- Sequential: YYMMDD-A
    inward_date TEXT NOT NULL,
    vendor_id TEXT NOT NULL,
    vendor_name TEXT NOT NULL,
    invoice_no TEXT NOT NULL,
    inward_type TEXT NOT NULL,              -- Strict enum: 'Fabric', 'Zips', 'Packaging', 'Trims', 'Labels', 'Finished Goods'
    uom TEXT NOT NULL DEFAULT 'Kgs',        -- Combobox: Kgs, Meters, Pcs, etc.
    invoice_qty REAL NOT NULL,
    invoice_amount REAL NOT NULL DEFAULT 0.0,
    physical_received_qty REAL NOT NULL,
    shortfall_qty REAL NOT NULL DEFAULT 0.0,
    discrepancy_amount REAL NOT NULL DEFAULT 0.0,
    fabric_type TEXT,
    vendor_fab_name TEXT,
    fpcc_fab_name TEXT,
    vendor_color TEXT,
    fpcc_color TEXT,
    invoice_gsm REAL,
    dia REAL,
    remarks TEXT,
    meta_json JSON,
    created_at INTEGER NOT NULL
);

-- Bundle Rolls Projection (Step B)
CREATE TABLE IF NOT EXISTS bundles (
    id TEXT PRIMARY KEY,
    bundle_code TEXT UNIQUE NOT NULL,       -- Sequential: ParentCode-1/N ... N/N
    parent_lot_id TEXT NOT NULL REFERENCES parent_inward_lots(id),
    parent_code TEXT NOT NULL,
    bundle_sequence INTEGER NOT NULL,
    total_bundles INTEGER NOT NULL,
    item_type TEXT NOT NULL,
    item_name TEXT NOT NULL,
    color_name TEXT,
    inward_weight_qty REAL NOT NULL,
    measured_gsm REAL,
    location_bin TEXT NOT NULL,             -- Warehouse rack code e.g. 'A/R01/S01'
    qc_status TEXT NOT NULL DEFAULT 'Passed', -- Strict enum: 'Passed', 'Rejected', 'Pending'
    qc_remarks TEXT,
    qr_payload JSON NOT NULL,
    available_qty REAL NOT NULL,
    created_at INTEGER NOT NULL
);

-- Discrepancy Reports Projection
CREATE TABLE IF NOT EXISTS discrepancy_reports (
    id TEXT PRIMARY KEY,
    report_no TEXT UNIQUE NOT NULL,
    parent_code TEXT NOT NULL,
    vendor_name TEXT NOT NULL,
    invoice_no TEXT NOT NULL,
    inward_type TEXT NOT NULL,
    invoice_qty REAL NOT NULL,
    actual_qty REAL NOT NULL,
    variance_qty REAL NOT NULL,
    variance_pct REAL NOT NULL,
    financial_impact REAL NOT NULL,
    status TEXT NOT NULL,                   -- 'SHORTAGE_FLAGGED', 'EXCESS_FLAGGED', 'RESOLVED'
    remarks TEXT,
    created_at INTEGER NOT NULL
);

-- Master Entities (Vendors & Clients with auto-registration)
CREATE TABLE IF NOT EXISTS master_entities (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    entity_type TEXT NOT NULL,              -- 'client', 'vendor', 'both'
    contact_person TEXT,
    phone TEXT,
    email TEXT,
    gstin TEXT,
    address TEXT,
    created_at INTEGER NOT NULL
);

-- Master Items (Fabrics, Colors, Products, Locations, UOMs)
CREATE TABLE IF NOT EXISTS master_items (
    id TEXT PRIMARY KEY,
    category TEXT NOT NULL,                 -- 'fabric', 'color', 'product', 'location', 'uom', 'zip', 'trim', 'label', 'packaging'
    code TEXT NOT NULL,
    display_name TEXT NOT NULL,
    meta_json JSON,
    created_at INTEGER NOT NULL,
    UNIQUE(category, code)
);
