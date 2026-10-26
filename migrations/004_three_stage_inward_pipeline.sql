-- ========================================================
-- FPCC SELVEDGE: 3-STAGE INWARD PIPELINE MIGRATION
-- Station 1A (Gate Entry) -> Station 1B (Floor Count) -> Station 1C (QC)
-- ========================================================

-- 1. Extend parent_inward_lots with workflow stages and conditional attribute payload
ALTER TABLE parent_inward_lots ADD COLUMN stage TEXT DEFAULT 'AWAITING_COUNT';
ALTER TABLE parent_inward_lots ADD COLUMN unit_rate REAL DEFAULT 0.0;
ALTER TABLE parent_inward_lots ADD COLUMN counted_qty REAL DEFAULT 0.0;
ALTER TABLE parent_inward_lots ADD COLUMN variance_qty REAL DEFAULT 0.0;
ALTER TABLE parent_inward_lots ADD COLUMN variance_pct REAL DEFAULT 0.0;
ALTER TABLE parent_inward_lots ADD COLUMN vendor_product TEXT;
ALTER TABLE parent_inward_lots ADD COLUMN fpcc_product TEXT;
ALTER TABLE parent_inward_lots ADD COLUMN no_of_bundles INTEGER DEFAULT 1;
ALTER TABLE parent_inward_lots ADD COLUMN attrs_json JSON DEFAULT '{}';

-- 2. Extend bundles with size-wise rate, diameter, and QC inspection metadata
ALTER TABLE bundles ADD COLUMN size_spec TEXT;
ALTER TABLE bundles ADD COLUMN rate REAL DEFAULT 0.0;
ALTER TABLE bundles ADD COLUMN measured_dia REAL;
ALTER TABLE bundles ADD COLUMN bundle_comment TEXT;
ALTER TABLE bundles ADD COLUMN qc_image_url TEXT;
ALTER TABLE bundles ADD COLUMN qc_inspected_at INTEGER;

-- 3. Floor Count & Reconciliation Audit Table (Station 1B)
CREATE TABLE IF NOT EXISTS floor_count_reconciliations (
    id TEXT PRIMARY KEY,
    lot_id TEXT NOT NULL REFERENCES parent_inward_lots(id) ON DELETE CASCADE,
    parent_code TEXT NOT NULL,
    invoiced_qty REAL NOT NULL,
    counted_qty REAL NOT NULL,
    variance_qty REAL NOT NULL,
    variance_pct REAL NOT NULL,
    financial_impact REAL NOT NULL,
    status TEXT NOT NULL,                     -- 'MATCHED', 'SHORTAGE_FLAGGED', 'EXCESS_FLAGGED'
    location_bin TEXT NOT NULL,
    size_lines_json JSON,
    counted_by TEXT,
    comments TEXT,
    recorded_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_reconcil_parent ON floor_count_reconciliations(parent_code);

-- 4. Bundle QC Inspections Append-Only Audit Table (Station 1C)
CREATE TABLE IF NOT EXISTS bundle_qc_inspections (
    id TEXT PRIMARY KEY,
    bundle_id TEXT NOT NULL REFERENCES bundles(id) ON DELETE CASCADE,
    parent_code TEXT NOT NULL,
    bundle_code TEXT NOT NULL,
    result TEXT NOT NULL,                     -- 'PASS', 'FAIL'
    remarks TEXT,
    image_url TEXT,
    inspected_by TEXT,
    inspected_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_qc_bundle ON bundle_qc_inspections(bundle_code);
CREATE INDEX IF NOT EXISTS idx_qc_parent ON bundle_qc_inspections(parent_code);
