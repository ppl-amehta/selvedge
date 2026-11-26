-- Migration 013: Enrich Master Dropdown Sub-Attributes, Schema Support for Multi-Attribute Sampling/Product/Orders/Inward
-- Authoritative Excel source: Arjav_Drop_Downs__(1).xlsx (Trims, Zips, Label, Packaging, Inward Drop downs, fabrics)

-- ============================================================================
-- 1. CLEAN MISPLACED HEADER / BRAND STRINGS IN MASTER SIZE CATEGORIES
-- ============================================================================
DELETE FROM master_items WHERE category = 'trim_fusing_size' AND display_name = 'size';
DELETE FROM master_items WHERE category = 'packaging_carton_size' AND display_name IN ('Carton', 'DHL', 'FPCC', 'blank');
DELETE FROM master_items WHERE category = 'packaging_courier_size' AND display_name IN ('Courier Bag', 'ETTU', 'FPCC', 'Blank');

-- Clean non-numeric/duplicate fusing/trim size noise if any, keep exact Excel strings

-- ============================================================================
-- 2. SEED ALL EXCEL TRIM SUB-ITEMS & TYPE-SPECIFIC ATTRIBUTES
-- ============================================================================

-- Elastic Names (Trims Col D)
INSERT INTO master_items (id, category, code, display_name, meta_json, created_at) VALUES
('mi_te_mask',   'trim_elastic', 'MASK_ELASTIC',    'Mask Elastic',    '{"trim_type":"Elastic","source":"excel_trims"}', strftime('%s','now')),
('mi_te_fold',   'trim_elastic', 'FOLDED_ELASTIC',  'Folded Elastic',  '{"trim_type":"Elastic","source":"excel_trims"}', strftime('%s','now')),
('mi_te_s8r',    'trim_elastic', 'SINDU_8R',        'Sindu elastic 8R','{"trim_type":"Elastic","source":"excel_trims"}', strftime('%s','now')),
('mi_te_s16r',   'trim_elastic', 'SINDU_16R',       'Sindu elastic 16R','{"trim_type":"Elastic","source":"excel_trims"}', strftime('%s','now')),
('mi_te_cord',   'trim_elastic', 'CORD_ELASTIC',    'Cord Elastic',    '{"trim_type":"Elastic","source":"excel_trims"}', strftime('%s','now')),
('mi_te_fuse',   'trim_elastic', 'FUSING_ELASTIC',  'Fusing Elastic',  '{"trim_type":"Elastic","source":"excel_trims"}', strftime('%s','now')),
('mi_te_thread', 'trim_elastic', 'THREAD_ELASTIC',  'Thread Elastic',  '{"trim_type":"Elastic","source":"excel_trims"}', strftime('%s','now'))
ON CONFLICT(category, code) DO UPDATE SET display_name = excluded.display_name, meta_json = excluded.meta_json;

-- Tapes Names (Trims Col H)
INSERT INTO master_items (id, category, code, display_name, meta_json, created_at) VALUES
('mi_tt_cotton',   'trim_tape', 'TAPE_COTTON',     'Cotton',       '{"trim_type":"Tapes","source":"excel_trims"}', strftime('%s','now')),
('mi_tt_twill',    'trim_tape', 'TAPE_TWILL',      'Twill',        '{"trim_type":"Tapes","source":"excel_trims"}', strftime('%s','now')),
('mi_tt_gross',    'trim_tape', 'TAPE_GROSSGRAIN', 'Gross grain',  '{"trim_type":"Tapes","source":"excel_trims"}', strftime('%s','now')),
('mi_tt_poly',     'trim_tape', 'TAPE_POLY',       'Polyster',     '{"trim_type":"Tapes","source":"excel_trims"}', strftime('%s','now')),
('mi_tt_refl',     'trim_tape', 'TAPE_REFLECTIVE', 'Reflective',   '{"trim_type":"Tapes","source":"excel_trims"}', strftime('%s','now')),
('mi_tt_mob',      'trim_tape', 'TAPE_MOBILIN',    'mobilin tape', '{"trim_type":"Tapes","source":"excel_trims"}', strftime('%s','now')),
('mi_tt_velvet',   'trim_tape', 'TAPE_VELVET',     'Velvet Tape',  '{"trim_type":"Tapes","source":"excel_trims"}', strftime('%s','now'))
ON CONFLICT(category, code) DO UPDATE SET display_name = excluded.display_name, meta_json = excluded.meta_json;

-- Cord Types (Trims Col N) & Cord Ends (Trims Col O)
INSERT INTO master_items (id, category, code, display_name, meta_json, created_at) VALUES
('mi_tc_flat',     'trim_cord', 'FLAT_CORD',   'Flat cord',  '{"trim_type":"Cord","source":"excel_trims"}', strftime('%s','now')),
('mi_tc_round',    'trim_cord', 'ROUND_CORD',  'round cord', '{"trim_type":"Cord","source":"excel_trims"}', strftime('%s','now')),
('mi_tc_end',      'trim_cord', 'END_CORD',    'end cord',   '{"trim_type":"Cord","source":"excel_trims"}', strftime('%s','now')),
('mi_tce_msilver', 'trim_cord_end', 'METAL_SILVER', 'metal silver', '{"trim_type":"Cord","source":"excel_trims"}', strftime('%s','now')),
('mi_tce_mgold',   'trim_cord_end', 'METAL_GOLD',   'metal gold',   '{"trim_type":"Cord","source":"excel_trims"}', strftime('%s','now')),
('mi_tce_rubber',  'trim_cord_end', 'RUBBER',       'rubber',       '{"trim_type":"Cord","source":"excel_trims"}', strftime('%s','now')),
('mi_tce_plastic', 'trim_cord_end', 'PLASTIC',      'plastic',      '{"trim_type":"Cord","source":"excel_trims"}', strftime('%s','now'))
ON CONFLICT(category, code) DO UPDATE SET display_name = excluded.display_name, meta_json = excluded.meta_json;

-- Buttons (Trims Col T)
INSERT INTO master_items (id, category, code, display_name, meta_json, created_at) VALUES
('mi_tb_ring',   'trim_button', 'RING_BTN',     'Ring buttons',     '{"trim_type":"Button","source":"excel_trims"}', strftime('%s','now')),
('mi_tb_snap',   'trim_button', 'SNAP_BTN',     'snap button',      '{"trim_type":"Button","source":"excel_trims"}', strftime('%s','now')),
('mi_tb_poly',   'trim_button', 'POLY_BTN',     'polyster buttons', '{"trim_type":"Button","source":"excel_trims"}', strftime('%s','now')),
('mi_tb_shell',  'trim_button', 'SHELL_BTN',    'shell buttons',    '{"trim_type":"Button","source":"excel_trims"}', strftime('%s','now')),
('mi_tb_wooden', 'trim_button', 'WOODEN_BTN',   'wooden buttons',   '{"trim_type":"Button","source":"excel_trims"}', strftime('%s','now'))
ON CONFLICT(category, code) DO UPDATE SET display_name = excluded.display_name, meta_json = excluded.meta_json;

-- Fusing Names (Trims Col Y), Fusing GSM (Trims Col AC), Fusing Colors (Trims Col AE)
INSERT INTO master_items (id, category, code, display_name, meta_json, created_at) VALUES
('mi_tf_dhoti',   'trim_fusing', 'DHOTI',           'Dhoti',                       '{"trim_type":"Fusing","source":"excel_trims"}', strftime('%s','now')),
('mi_tf_canvass', 'trim_fusing', 'CANVASS',         'Canvass',                     '{"trim_type":"Fusing","source":"excel_trims"}', strftime('%s','now')),
('mi_tf_9233a',   'trim_fusing', 'FREUD_9233A',     '9/1/9233 Freudenberg-plkt',   '{"trim_type":"Fusing","source":"excel_trims"}', strftime('%s','now')),
('mi_tf_9233b',   'trim_fusing', 'FREUD_9233_98',   '9233-98 Freudenberg-plkt',    '{"trim_type":"Fusing","source":"excel_trims"}', strftime('%s','now')),
('mi_tf_9520a',   'trim_fusing', 'FREUD_9520_1000', '9520-1000 Freudenberg collar','{"trim_type":"Fusing","source":"excel_trims"}', strftime('%s','now')),
('mi_tf_9520b',   'trim_fusing', 'FREUD_9520_3108', '9520-3108 Freudenberg collar','{"trim_type":"Fusing","source":"excel_trims"}', strftime('%s','now')),
('mi_tf_welco1',  'trim_fusing', 'WELCO_1182',      '3630R/1182 welco collar',     '{"trim_type":"Fusing","source":"excel_trims"}', strftime('%s','now')),
('mi_tf_welco2',  'trim_fusing', 'WELCO_9999',      '3630R/9999 welco collar',     '{"trim_type":"Fusing","source":"excel_trims"}', strftime('%s','now')),
('mi_tf_mdw',     'trim_fusing', 'MICRODOT_W',      'microdot white',              '{"trim_type":"Fusing","source":"excel_trims"}', strftime('%s','now')),
('mi_tf_mdg',     'trim_fusing', 'MICRODOT_G',      'microdot grey',               '{"trim_type":"Fusing","source":"excel_trims"}', strftime('%s','now')),

('mi_tfg_20',  'trim_fusing_gsm', 'GSM_20',  '20.0',  '{"trim_type":"Fusing"}', strftime('%s','now')),
('mi_tfg_35',  'trim_fusing_gsm', 'GSM_35',  '35.0',  '{"trim_type":"Fusing"}', strftime('%s','now')),
('mi_tfg_45',  'trim_fusing_gsm', 'GSM_45',  '45.0',  '{"trim_type":"Fusing"}', strftime('%s','now')),
('mi_tfg_50',  'trim_fusing_gsm', 'GSM_50',  '50.0',  '{"trim_type":"Fusing"}', strftime('%s','now')),
('mi_tfg_55',  'trim_fusing_gsm', 'GSM_55',  '55.0',  '{"trim_type":"Fusing"}', strftime('%s','now')),
('mi_tfg_100', 'trim_fusing_gsm', 'GSM_100', '100.0', '{"trim_type":"Fusing"}', strftime('%s','now')),
('mi_tfg_140', 'trim_fusing_gsm', 'GSM_140', '140.0', '{"trim_type":"Fusing"}', strftime('%s','now')),
('mi_tfg_145', 'trim_fusing_gsm', 'GSM_145', '145.0', '{"trim_type":"Fusing"}', strftime('%s','now')),
('mi_tfg_150', 'trim_fusing_gsm', 'GSM_150', '150.0', '{"trim_type":"Fusing"}', strftime('%s','now')),

('mi_tfc_blk', 'trim_fusing_color', 'BLACK', 'black', '{"trim_type":"Fusing"}', strftime('%s','now')),
('mi_tfc_wht', 'trim_fusing_color', 'WHITE', 'white', '{"trim_type":"Fusing"}', strftime('%s','now')),
('mi_tfc_gry', 'trim_fusing_color', 'GREY',  'grey',  '{"trim_type":"Fusing"}', strftime('%s','now'))
ON CONFLICT(category, code) DO UPDATE SET display_name = excluded.display_name, meta_json = excluded.meta_json;

-- Thread Types (Trims Col AG) & Lengths (Trims Col AI)
INSERT INTO master_items (id, category, code, display_name, meta_json, created_at) VALUES
('mi_th_polo',  'trim_thread', 'POLO_POLY', 'Polo Poly',    '{"trim_type":"Thread","source":"excel_trims"}', strftime('%s','now')),
('mi_th_fil',   'trim_thread', 'FILAMENT',  'Filament',     '{"trim_type":"Thread","source":"excel_trims"}', strftime('%s','now')),
('mi_th_3ply',  'trim_thread', '3PLY',      '3 ply thread', '{"trim_type":"Thread","source":"excel_trims"}', strftime('%s','now')),
('mi_thl_300',  'trim_thread_length', '300M',   '300 mtr',   '{"trim_type":"Thread"}', strftime('%s','now')),
('mi_thl_1000', 'trim_thread_length', '1000M',  '1000 mtr',  '{"trim_type":"Thread"}', strftime('%s','now')),
('mi_thl_5000', 'trim_thread_length', '5000M',  '5000 mtr',  '{"trim_type":"Thread"}', strftime('%s','now')),
('mi_thl_10k',  'trim_thread_length', '10000M', '10000 mtr', '{"trim_type":"Thread"}', strftime('%s','now'))
ON CONFLICT(category, code) DO UPDATE SET display_name = excluded.display_name, meta_json = excluded.meta_json;

-- Other Trims Colors (Col AO) & Sizes (Col AQ)
INSERT INTO master_items (id, category, code, display_name, meta_json, created_at) VALUES
('mi_toc_blk', 'trim_other_color', 'BLACK',  'black',  '{"source":"excel_trims"}', strftime('%s','now')),
('mi_toc_wht', 'trim_other_color', 'WHITE',  'white',  '{"source":"excel_trims"}', strftime('%s','now')),
('mi_toc_slv', 'trim_other_color', 'SILVER', 'silver', '{"source":"excel_trims"}', strftime('%s','now')),
('mi_toc_gld', 'trim_other_color', 'GOLD',   'gold',   '{"source":"excel_trims"}', strftime('%s','now')),
('mi_tos_05',  'trim_other_size',  '0_5IN',  '0.5"',   '{"source":"excel_trims"}', strftime('%s','now')),
('mi_tos_15',  'trim_other_size',  '1_5IN',  '1.5"',   '{"source":"excel_trims"}', strftime('%s','now')),
('mi_tos_25',  'trim_other_size',  '2_5IN',  '2.5"',   '{"source":"excel_trims"}', strftime('%s','now'))
ON CONFLICT(category, code) DO UPDATE SET display_name = excluded.display_name, meta_json = excluded.meta_json;

-- Packaging Sub-Attributes (Packaging sheet Col G-L)
INSERT INTO master_items (id, category, code, display_name, meta_json, created_at) VALUES
('mi_pkb_jeep',  'packaging_polybag_brand', 'JEEP',  'Jeep',  '{"packaging_type":"Polybag"}', strftime('%s','now')),
('mi_pkb_ettu',  'packaging_polybag_brand', 'ETTU',  'Ettu',  '{"packaging_type":"Polybag"}', strftime('%s','now')),
('mi_pkb_blank', 'packaging_polybag_brand', 'BLANK', 'Blank', '{"packaging_type":"Polybag"}', strftime('%s','now')),

('mi_pks_3x4',   'packaging_polybag_size', '3X4',   '3x4"',   '{"packaging_type":"Polybag"}', strftime('%s','now')),
('mi_pks_10x16', 'packaging_polybag_size', '10X16', '10x16"', '{"packaging_type":"Polybag"}', strftime('%s','now')),
('mi_pks_12x16', 'packaging_polybag_size', '12X16', '12x16"', '{"packaging_type":"Polybag"}', strftime('%s','now')),
('mi_pks_12x17', 'packaging_polybag_size', '12X17', '12x17"', '{"packaging_type":"Polybag"}', strftime('%s','now')),
('mi_pks_16x20', 'packaging_polybag_size', '16X20', '16x20"', '{"packaging_type":"Polybag"}', strftime('%s','now')),
('mi_pks_30x50', 'packaging_polybag_size', '30X50', '30x50"', '{"packaging_type":"Polybag"}', strftime('%s','now')),

('mi_pktv_brn', 'packaging_tape_variant', 'BROWN', 'brown tape',       '{"packaging_type":"Carton Tape"}', strftime('%s','now')),
('mi_pktv_trn', 'packaging_tape_variant', 'TRANS', 'transparent tape', '{"packaging_type":"Carton Tape"}', strftime('%s','now')),
('mi_pktv_ett', 'packaging_tape_variant', 'ETTU',  'ettu tape',        '{"packaging_type":"Carton Tape"}', strftime('%s','now')),

('mi_pkts_025', 'packaging_tape_size', '0_25', '1/4"', '{"packaging_type":"Carton Tape"}', strftime('%s','now')),
('mi_pkts_050', 'packaging_tape_size', '0_50', '1/2"', '{"packaging_type":"Carton Tape"}', strftime('%s','now')),
('mi_pkts_1',   'packaging_tape_size', '1IN',  '1"',   '{"packaging_type":"Carton Tape"}', strftime('%s','now')),
('mi_pkts_2',   'packaging_tape_size', '2IN',  '2"',   '{"packaging_type":"Carton Tape"}', strftime('%s','now')),
('mi_pkts_3',   'packaging_tape_size', '3IN',  '3"',   '{"packaging_type":"Carton Tape"}', strftime('%s','now')),

('mi_pklm_poly', 'packaging_loop_material', 'POLY', 'Polyster', '{"packaging_type":"Loop Pins"}', strftime('%s','now')),
('mi_pklm_cot',  'packaging_loop_material', 'COT',  'cotton',   '{"packaging_type":"Loop Pins"}', strftime('%s','now')),
('mi_pklm_pls',  'packaging_loop_material', 'PLS',  'plastic',  '{"packaging_type":"Loop Pins"}', strftime('%s','now')),

('mi_pkls_3', 'packaging_loop_size', '3IN', '3"', '{"packaging_type":"Loop Pins"}', strftime('%s','now')),
('mi_pkls_5', 'packaging_loop_size', '5IN', '5"', '{"packaging_type":"Loop Pins"}', strftime('%s','now')),
('mi_pkls_7', 'packaging_loop_size', '7IN', '7"', '{"packaging_type":"Loop Pins"}', strftime('%s','now')),
('mi_pkls_9', 'packaging_loop_size', '9IN', '9"', '{"packaging_type":"Loop Pins"}', strftime('%s','now')),

('mi_pkcb_blk',  'packaging_carton_brand',  'BLANK', 'blank', '{"packaging_type":"Carton"}', strftime('%s','now')),
('mi_pkcb_fpcc', 'packaging_carton_brand',  'FPCC',  'FPCC',  '{"packaging_type":"Carton"}', strftime('%s','now')),
('mi_pkcb_dhl',  'packaging_carton_brand',  'DHL',   'DHL',   '{"packaging_type":"Carton"}', strftime('%s','now')),

('mi_pksr_18',   'packaging_shrink_size',   '18IN',  '18"',   '{"packaging_type":"Shrink wrap"}', strftime('%s','now')),
('mi_pksr_20',   'packaging_shrink_size',   '20IN',  '20"',   '{"packaging_type":"Shrink wrap"}', strftime('%s','now')),

('mi_pkcrb_ettu',  'packaging_courier_brand', 'ETTU',  'ETTU',  '{"packaging_type":"Courier Bag"}', strftime('%s','now')),
('mi_pkcrb_fpcc',  'packaging_courier_brand', 'FPCC',  'FPCC',  '{"packaging_type":"Courier Bag"}', strftime('%s','now')),
('mi_pkcrb_blank', 'packaging_courier_brand', 'BLANK', 'Blank', '{"packaging_type":"Courier Bag"}', strftime('%s','now'))
ON CONFLICT(category, code) DO UPDATE SET display_name = excluded.display_name, meta_json = excluded.meta_json;

-- ============================================================================
-- 3. SCHEMA EXTENSIONS FOR SAMPLE, PRODUCT CREATION, ORDERS & INWARD PIPELINE
-- ============================================================================

ALTER TABLE sample_developments ADD COLUMN colors_json JSON DEFAULT '[]';
ALTER TABLE sample_fabrics ADD COLUMN attrs_json JSON DEFAULT '{}';
ALTER TABLE sample_trims ADD COLUMN attrs_json JSON DEFAULT '{}';
ALTER TABLE sample_zips ADD COLUMN attrs_json JSON DEFAULT '{}';
ALTER TABLE sample_labels ADD COLUMN attrs_json JSON DEFAULT '{}';

ALTER TABLE master_product_genders ADD COLUMN sizes_json JSON DEFAULT '[]';
ALTER TABLE master_product_genders ADD COLUMN pharma_nos_json JSON DEFAULT '[]';

ALTER TABLE order_items ADD COLUMN fabric_composition TEXT DEFAULT '';
ALTER TABLE order_items ADD COLUMN brand_label TEXT DEFAULT '';
ALTER TABLE order_items ADD COLUMN brand_label_genre TEXT DEFAULT '';
ALTER TABLE order_items ADD COLUMN brand_label_desc TEXT DEFAULT '';
ALTER TABLE order_items ADD COLUMN additional_labels_json JSON DEFAULT '[]';

ALTER TABLE inward_invoices ADD COLUMN inward_type TEXT DEFAULT '';

ALTER TABLE bundles ADD COLUMN invoiced_qty REAL NOT NULL DEFAULT 0.0;
ALTER TABLE bundles ADD COLUMN attrs_json JSON DEFAULT '{}';

-- Backfill inward_invoices.inward_type from parent_inward_lots where unambiguous
UPDATE inward_invoices
SET inward_type = COALESCE(
    (SELECT l.inward_type FROM parent_inward_lots l WHERE l.invoice_id = inward_invoices.id LIMIT 1),
    (SELECT l.inward_type FROM parent_inward_lots l WHERE l.invoice_no = inward_invoices.invoice_no LIMIT 1),
    ''
)
WHERE COALESCE(inward_type, '') = '';

-- Backfill bundles.invoiced_qty from parent lot invoice_qty divided by bundle count where 0
UPDATE bundles
SET invoiced_qty = ROUND(
    COALESCE((SELECT l.invoice_qty FROM parent_inward_lots l WHERE l.id = bundles.parent_lot_id), 0.0)
    / MAX(COALESCE(NULLIF(bundles.total_bundles, 0), 1), 1),
    3
)
WHERE invoiced_qty = 0.0;

-- Backfill order_items.brand_label from main_label if empty
UPDATE order_items
SET brand_label = COALESCE(main_label, '')
WHERE COALESCE(brand_label, '') = '' AND COALESCE(main_label, '') != '';
