-- Migration 014: Normalize legacy 'Fabric Knitted Roll' inward_type and backfill missing bundle rows on seeded lots

UPDATE parent_inward_lots SET inward_type = 'Fabric' WHERE inward_type LIKE 'Fabric %';
UPDATE inward_invoices SET inward_type = 'Fabric' WHERE inward_type LIKE 'Fabric %';
UPDATE bundles SET item_type = 'Fabric' WHERE item_type LIKE 'Fabric %';

-- Backfill Trims lot 261116-A (3 bundles) if missing
INSERT OR IGNORE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, invoiced_qty, inward_weight_qty, measured_gsm, measured_dia,
    location_bin, qc_status, qc_remarks, qr_payload, available_qty, size_spec, rate, attrs_json, created_at
)
SELECT
    'bnd_261116_a_1', '261116-A-1/3', p.id, p.parent_code, 1, 3,
    'Trims', COALESCE(NULLIF(p.fpcc_product, ''), 'Snap Button 18L'), COALESCE(NULLIF(p.fpcc_color, ''), 'Black'),
    33.333, 30.0, 0.0, 0.0, 'A/R02/S01', 'Pending', '', '{}', 30.0, '18L', 12.0, '{"trimType":"Button"}', strftime('%s','now')
FROM parent_inward_lots p WHERE p.parent_code = '261116-A'
  AND NOT EXISTS (SELECT 1 FROM bundles b WHERE b.parent_code = '261116-A' AND b.bundle_sequence = 1);

INSERT OR IGNORE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, invoiced_qty, inward_weight_qty, measured_gsm, measured_dia,
    location_bin, qc_status, qc_remarks, qr_payload, available_qty, size_spec, rate, attrs_json, created_at
)
SELECT
    'bnd_261116_a_2', '261116-A-2/3', p.id, p.parent_code, 2, 3,
    'Trims', COALESCE(NULLIF(p.fpcc_product, ''), 'Snap Button 18L'), COALESCE(NULLIF(p.fpcc_color, ''), 'Black'),
    33.333, 30.0, 0.0, 0.0, 'A/R02/S01', 'Pending', '', '{}', 30.0, '18L', 12.0, '{"trimType":"Button"}', strftime('%s','now')
FROM parent_inward_lots p WHERE p.parent_code = '261116-A'
  AND NOT EXISTS (SELECT 1 FROM bundles b WHERE b.parent_code = '261116-A' AND b.bundle_sequence = 2);

INSERT OR IGNORE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, invoiced_qty, inward_weight_qty, measured_gsm, measured_dia,
    location_bin, qc_status, qc_remarks, qr_payload, available_qty, size_spec, rate, attrs_json, created_at
)
SELECT
    'bnd_261116_a_3', '261116-A-3/3', p.id, p.parent_code, 3, 3,
    'Trims', COALESCE(NULLIF(p.fpcc_product, ''), 'Snap Button 18L'), COALESCE(NULLIF(p.fpcc_color, ''), 'Black'),
    33.334, 30.0, 0.0, 0.0, 'A/R02/S01', 'Pending', '', '{}', 30.0, '18L', 12.0, '{"trimType":"Button"}', strftime('%s','now')
FROM parent_inward_lots p WHERE p.parent_code = '261116-A'
  AND NOT EXISTS (SELECT 1 FROM bundles b WHERE b.parent_code = '261116-A' AND b.bundle_sequence = 3);

-- Backfill Fabric lot 261117-A (1 bundle) if missing
INSERT OR IGNORE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, invoiced_qty, inward_weight_qty, measured_gsm, measured_dia,
    location_bin, qc_status, qc_remarks, qr_payload, available_qty, size_spec, rate, attrs_json, created_at
)
SELECT
    'bnd_261117_a_1', '261117-A-1/1', p.id, p.parent_code, 1, 1,
    'Fabric', COALESCE(NULLIF(p.fpcc_product, ''), NULLIF(p.fpcc_fab_name, ''), '2441 Galactic Polo'),
    COALESCE(NULLIF(p.fpcc_color, ''), 'Navy Blue'),
    COALESCE(p.invoice_qty, 850.0), 0.0, COALESCE(p.invoice_gsm, 180.0), COALESCE(p.dia, 60.0),
    'A/R01/S01', 'Pending', '', '{}', 0.0, 'Fabric (Dia: 60")', 350.0, '{}', strftime('%s','now')
FROM parent_inward_lots p WHERE p.parent_code = '261117-A'
  AND NOT EXISTS (SELECT 1 FROM bundles b WHERE b.parent_code = '261117-A');
