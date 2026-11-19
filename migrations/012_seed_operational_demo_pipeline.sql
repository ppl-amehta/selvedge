-- =============================================================================
-- Migration 012: Seed Precise, Nuanced Operational Pipeline Across Selvedge ERP
-- Strictly incorporates authoritative Excel entities, fabrics, trims, zips,
-- and compositions for realistic end-to-end multi-station execution.
-- =============================================================================
-- Clean up seeded rows in strict reverse dependency order to ensure idempotency with foreign keys ON
DELETE FROM remnant_returns WHERE id LIKE 'Bit2611%' OR id LIKE 'Rej2611%';
DELETE FROM cut_bundles WHERE id LIKE 'CB-261118%';
DELETE FROM cut_lays WHERE id LIKE 'CL-CJ-%';
DELETE FROM cutting_roll_issues WHERE id LIKE 'cri_%';
DELETE FROM cutting_jobs WHERE id LIKE 'CJ-%';
DELETE FROM procurement_requisitions WHERE id LIKE 'pr_1126%';
DELETE FROM bom_lines WHERE id LIKE 'bom_1126%';
DELETE FROM order_planning_snapshots WHERE id LIKE 'plan_1126%';
DELETE FROM order_personalizations WHERE id LIKE 'pers_item_%';
DELETE FROM order_item_brandings WHERE id LIKE 'br_item_%';
DELETE FROM order_items WHERE id LIKE 'item_1126%';
DELETE FROM orders WHERE id LIKE 'ord_1126%';
DELETE FROM bundles WHERE id LIKE 'bnd_2611%';
DELETE FROM discrepancy_reports WHERE id LIKE 'dr_2611%';
DELETE FROM parent_inward_lots WHERE id LIKE 'lot_2611%';
DELETE FROM inward_invoices WHERE id LIKE 'inv_%';
DELETE FROM client_product_pricing WHERE id LIKE 'prc_%';
DELETE FROM master_product_materials WHERE id LIKE 'mm_prod_%' OR product_id = 'prod_19b7dec6';
DELETE FROM master_product_genders WHERE id LIKE 'mg_prod_%' OR product_id = 'prod_19b7dec6';
DELETE FROM master_products WHERE id LIKE 'prod_%';
DELETE FROM sample_pharmas WHERE id LIKE 'ph_smp_%' OR development_id = 'smp_0f5812c1';
DELETE FROM sample_labels WHERE id LIKE 'lb_smp_%' OR development_id = 'smp_0f5812c1';
DELETE FROM sample_zips WHERE id LIKE 'zp_smp_%' OR development_id = 'smp_0f5812c1';
DELETE FROM sample_trims WHERE id LIKE 'tr_smp_%' OR development_id = 'smp_0f5812c1';
DELETE FROM sample_fabrics WHERE id LIKE 'fb_smp_%' OR development_id = 'smp_0f5812c1';
DELETE FROM sample_sizes WHERE id LIKE 'sz_smp_%' OR development_id = 'smp_0f5812c1';
DELETE FROM sample_developments WHERE id LIKE 'smp_%';
-- 1. Sample Developments & Children
INSERT OR REPLACE INTO sample_developments (
    id, development_no, development_date, client_id, client_name, product_genre,
    temp_product_name, sample_image_url, product_description, dev_comments,
    masters_comments, rework_notes, status, created_at, updated_at
) VALUES (
    'smp_polo_001', 'D26110001', '2026-11-10', 'ent_cli_atc_tires_private_limited_9768afc0', 'ATC Tires Private Limited',
    'Polo', 'High-Density Pique Corporate Polo', 'https://images.unsplash.com/photo-1581655353564-df123a1eb820?w=600',
    'Premium corporate polo with tipped collar and embroidered chest crest for management staff.', 'Collar rib tension verified. Bio-wash washfastness rating 4.5. Approved by client merchandising.',
    'Armhole drop deepened by 0.5cm for ease of movement. Base pattern pharma PH-POL-2401 finalized.', '',
    'APPROVED', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_polo_001_men_s_0', 'smp_polo_001', 'Men', 'S', 1, 0);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_polo_001_men_m_1', 'smp_polo_001', 'Men', 'M', 2, 1);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_polo_001_men_l_2', 'smp_polo_001', 'Men', 'L', 2, 2);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_polo_001_men_xl_3', 'smp_polo_001', 'Men', 'XL', 1, 3);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_polo_001_men_2xl_4', 'smp_polo_001', 'Men', '2XL', 1, 4);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_polo_001_women_xs_5', 'smp_polo_001', 'Women', 'XS', 1, 5);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_polo_001_women_s_6', 'smp_polo_001', 'Women', 'S', 2, 6);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_polo_001_women_m_7', 'smp_polo_001', 'Women', 'M', 2, 7);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_polo_001_women_l_8', 'smp_polo_001', 'Women', 'L', 1, 8);
INSERT OR REPLACE INTO sample_fabrics (id, development_id, fabric_name, lot_no, average, sequence) VALUES ('fb_smp_polo_001_0', 'smp_polo_001', '2441 Galactic Polo', 'L-2611-01', 0.28, 0);
INSERT OR REPLACE INTO sample_trims (id, development_id, trim_name, average, sequence) VALUES ('tr_smp_polo_001_0', 'smp_polo_001', 'FPCC Signature Laser Button', 3.0, 0);
INSERT OR REPLACE INTO sample_trims (id, development_id, trim_name, average, sequence) VALUES ('tr_smp_polo_001_1', 'smp_polo_001', 'High Density Woven Fusing', 0.05, 1);
INSERT OR REPLACE INTO sample_labels (id, development_id, label_name, sequence) VALUES ('lb_smp_polo_001_0', 'smp_polo_001', 'Brand Label - ETTU', 0);
INSERT OR REPLACE INTO sample_labels (id, development_id, label_name, sequence) VALUES ('lb_smp_polo_001_1', 'smp_polo_001', 'Taffeta Wash Care Label', 1);
INSERT OR REPLACE INTO sample_pharmas (id, development_id, base_pharma_no, changes_made, new_pharma_no, sequence) VALUES ('ph_smp_polo_001_0', 'smp_polo_001', 'PH-POL-BASE-01', 'Armhole +0.5cm, collar length +1cm', 'PH-POL-2401', 0);
INSERT OR REPLACE INTO sample_developments (
    id, development_no, development_date, client_id, client_name, product_genre,
    temp_product_name, sample_image_url, product_description, dev_comments,
    masters_comments, rework_notes, status, created_at, updated_at
) VALUES (
    'smp_hoodie_002', 'D26110002', '2026-11-11', 'ent_cli_amara_raja_energy_mobility_limited_95ac4e91', 'Amara Raja Energy & Mobility Limited',
    'Hoodies and Pullovers', 'Heavyweight Fleece Tech Zip Hoodie', 'https://images.unsplash.com/photo-1556905055-8f358a7a47b2?w=600',
    '320 GSM brushed fleece hooded sweatshirt with antique metal zip and kangaroo pocket.', 'Sample sent via courier for fabric handfeel and drawstring tipping approval.',
    'Hood 3-panel profile fits head contours without pulling back neck.', '',
    'SUBMITTED', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_hoodie_002_unisex_s_0', 'smp_hoodie_002', 'Unisex', 'S', 1, 0);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_hoodie_002_unisex_m_1', 'smp_hoodie_002', 'Unisex', 'M', 2, 1);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_hoodie_002_unisex_l_2', 'smp_hoodie_002', 'Unisex', 'L', 2, 2);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_hoodie_002_unisex_xl_3', 'smp_hoodie_002', 'Unisex', 'XL', 2, 3);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_hoodie_002_unisex_2xl_4', 'smp_hoodie_002', 'Unisex', '2XL', 1, 4);
INSERT OR REPLACE INTO sample_fabrics (id, development_id, fabric_name, lot_no, average, sequence) VALUES ('fb_smp_hoodie_002_0', 'smp_hoodie_002', '2111 bonded box fleece', 'L-FL-88', 0.65, 0);
INSERT OR REPLACE INTO sample_trims (id, development_id, trim_name, average, sequence) VALUES ('tr_smp_hoodie_002_0', 'smp_hoodie_002', 'Rollax Powder', 1.0, 0);
INSERT OR REPLACE INTO sample_zips (id, development_id, zip_name, length_and_qty, sequence) VALUES ('zp_smp_hoodie_002_0', 'smp_hoodie_002', 'Brass', 'S-L: 65cm (1pc), XL-2XL: 70cm (1pc)', 0);
INSERT OR REPLACE INTO sample_labels (id, development_id, label_name, sequence) VALUES ('lb_smp_hoodie_002_0', 'smp_hoodie_002', 'Brand Label - American Athlete', 0);
INSERT OR REPLACE INTO sample_labels (id, development_id, label_name, sequence) VALUES ('lb_smp_hoodie_002_1', 'smp_hoodie_002', 'Care Label', 1);
INSERT OR REPLACE INTO sample_pharmas (id, development_id, base_pharma_no, changes_made, new_pharma_no, sequence) VALUES ('ph_smp_hoodie_002_0', 'smp_hoodie_002', 'PH-HOD-BASE-02', 'Lengthened sleeve rib by 1.5cm', 'PH-HOD-3205', 0);
INSERT OR REPLACE INTO sample_developments (
    id, development_no, development_date, client_id, client_name, product_genre,
    temp_product_name, sample_image_url, product_description, dev_comments,
    masters_comments, rework_notes, status, created_at, updated_at
) VALUES (
    'smp_tee_003', 'D26110003', '2026-11-12', 'ent_cli_acma_mobility_foundation_4d29096f', 'Acma Mobility Foundation',
    'Tee', 'EV Mobility Crew Neck Bio-Wash Tee', 'https://images.unsplash.com/photo-1521572267360-ee0c2909d518?w=600',
    '180 GSM 100% combed cotton jersey crew neck tee with silicon wash finish for conference delegates.', 'Knit lab dip shade matching in progress against Pantone 19-4052 Classic Blue.',
    'Standard regular fit pattern PH-TEE-REG-01 applied.', '',
    'IN_PROGRESS', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_tee_003_men_s_0', 'smp_tee_003', 'Men', 'S', 2, 0);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_tee_003_men_m_1', 'smp_tee_003', 'Men', 'M', 3, 1);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_tee_003_men_l_2', 'smp_tee_003', 'Men', 'L', 3, 2);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_tee_003_men_xl_3', 'smp_tee_003', 'Men', 'XL', 2, 3);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_tee_003_women_s_4', 'smp_tee_003', 'Women', 'S', 2, 4);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_tee_003_women_m_5', 'smp_tee_003', 'Women', 'M', 2, 5);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_tee_003_women_l_6', 'smp_tee_003', 'Women', 'L', 2, 6);
INSERT OR REPLACE INTO sample_fabrics (id, development_id, fabric_name, lot_no, average, sequence) VALUES ('fb_smp_tee_003_0', 'smp_tee_003', '30s Sinker Crew Bio Wash', 'L-SW-104', 0.22, 0);
INSERT OR REPLACE INTO sample_trims (id, development_id, trim_name, average, sequence) VALUES ('tr_smp_tee_003_0', 'smp_tee_003', '30s Sinker Crew Bio Wash -Rib', 0.04, 0);
INSERT OR REPLACE INTO sample_labels (id, development_id, label_name, sequence) VALUES ('lb_smp_tee_003_0', 'smp_tee_003', 'Brand Label - Citreon', 0);
INSERT OR REPLACE INTO sample_pharmas (id, development_id, base_pharma_no, changes_made, new_pharma_no, sequence) VALUES ('ph_smp_tee_003_0', 'smp_tee_003', 'PH-TEE-REG-01', 'Base regular crew', 'PH-TEE-REG-01', 0);
INSERT OR REPLACE INTO sample_developments (
    id, development_no, development_date, client_id, client_name, product_genre,
    temp_product_name, sample_image_url, product_description, dev_comments,
    masters_comments, rework_notes, status, created_at, updated_at
) VALUES (
    'smp_jacket_004', 'D26110004', '2026-11-13', 'ent_cli_apex_trading_d985c8b1', 'Apex Trading',
    'Jackets', 'Water-Resistant All-Weather Windbreaker', 'https://images.unsplash.com/photo-1548883354-7622d03aca27?w=600',
    'Lightweight lined utility jacket with concealed hood and breathable mesh lining.', 'Client requested looser shoulder pitch and waterproof reverse coil zippers.',
    'Pattern rework in progress: widen chest across back by 2cm and re-grade sleeve head.', '1. Increase bicep circumference by 1.5cm. 2. Replace plastic zip with Reverse Coil #5. 3. Adjust cuff elastic tension.',
    'REWORK_REQUESTED', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_jacket_004_men_m_0', 'smp_jacket_004', 'Men', 'M', 2, 0);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_jacket_004_men_l_1', 'smp_jacket_004', 'Men', 'L', 2, 1);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_jacket_004_men_xl_2', 'smp_jacket_004', 'Men', 'XL', 2, 2);
INSERT OR REPLACE INTO sample_fabrics (id, development_id, fabric_name, lot_no, average, sequence) VALUES ('fb_smp_jacket_004_0', 'smp_jacket_004', '3347 shark plain', 'L-SHK-09', 0.48, 0);
INSERT OR REPLACE INTO sample_trims (id, development_id, trim_name, average, sequence) VALUES ('tr_smp_jacket_004_0', 'smp_jacket_004', '357 mesh', 0.18, 0);
INSERT OR REPLACE INTO sample_zips (id, development_id, zip_name, length_and_qty, sequence) VALUES ('zp_smp_jacket_004_0', 'smp_jacket_004', 'Coil', 'M-L: 72cm (1pc), XL: 75cm (1pc)', 0);
INSERT OR REPLACE INTO sample_labels (id, development_id, label_name, sequence) VALUES ('lb_smp_jacket_004_0', 'smp_jacket_004', 'Brand Label - Jeep', 0);
INSERT OR REPLACE INTO sample_pharmas (id, development_id, base_pharma_no, changes_made, new_pharma_no, sequence) VALUES ('ph_smp_jacket_004_0', 'smp_jacket_004', 'PH-JKT-101', 'Chest across back +2cm, bicep +1.5cm', 'PH-JKT-101R', 0);
INSERT OR REPLACE INTO sample_developments (
    id, development_no, development_date, client_id, client_name, product_genre,
    temp_product_name, sample_image_url, product_description, dev_comments,
    masters_comments, rework_notes, status, created_at, updated_at
) VALUES (
    'smp_bag_005', 'D26110005', '2026-11-14', 'ent_cli_bapl_rototech_private_limited_6d9806e8', 'Bapl Rototech Private Limited',
    'Bags', 'Heavy-Duty Industrial Canvas Tote Bag', 'https://images.unsplash.com/photo-1544816155-12df9643f363?w=600',
    '16oz heavy cotton duck canvas multi-compartment promotional tote with reinforced cross-box stitched handles.', 'Tensile drop test passed 25kg load capacity. Approved for bulk corporate gifting.',
    'CAD cutting marker nested at 89% yield. Handle strap reinforcement template tool #TOTE-16 finalized.', '',
    'APPROVED', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_bag_005_one size_one size_0', 'smp_bag_005', 'One Size', 'One Size', 5, 0);
INSERT OR REPLACE INTO sample_fabrics (id, development_id, fabric_name, lot_no, average, sequence) VALUES ('fb_smp_bag_005_0', 'smp_bag_005', '30s cotton slub', 'L-CNV-03', 0.35, 0);
INSERT OR REPLACE INTO sample_trims (id, development_id, trim_name, average, sequence) VALUES ('tr_smp_bag_005_0', 'smp_bag_005', 'Stainoff', 1.0, 0);
INSERT OR REPLACE INTO sample_zips (id, development_id, zip_name, length_and_qty, sequence) VALUES ('zp_smp_bag_005_0', 'smp_bag_005', 'Metal', 'Interior Pocket: 20cm (1pc)', 0);
INSERT OR REPLACE INTO sample_labels (id, development_id, label_name, sequence) VALUES ('lb_smp_bag_005_0', 'smp_bag_005', 'Brand Label - ETTU', 0);
INSERT OR REPLACE INTO sample_pharmas (id, development_id, base_pharma_no, changes_made, new_pharma_no, sequence) VALUES ('ph_smp_bag_005_0', 'smp_bag_005', 'TOOL-TOTE-16', 'Handle drop lengthened to 62cm', 'TOOL-TOTE-16B', 0);
INSERT OR REPLACE INTO sample_developments (
    id, development_no, development_date, client_id, client_name, product_genre,
    temp_product_name, sample_image_url, product_description, dev_comments,
    masters_comments, rework_notes, status, created_at, updated_at
) VALUES (
    'smp_bottle_006', 'D26110006', '2026-11-15', 'ent_cli_aapka_khuranas_india_llp_0e766966', 'Aapka Khuranas India Llp',
    'Drinkwear', 'Matte Vacuum Insulated Flask 750ml', 'https://images.unsplash.com/photo-1602143407151-7111542de6e8?w=600',
    'Double-wall food grade SS304 thermal hydration bottle with bamboo screw cap and silicone loop.', 'Initial tooling evaluation for laser etching logo vs rotary screen printing.',
    'Die tooling drawing DWG-BTL-750 reviewed with mold supplier.', '',
    'DRAFT', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_bottle_006_one size_750ml_0', 'smp_bottle_006', 'One Size', '750ml', 2, 0);
INSERT OR REPLACE INTO sample_fabrics (id, development_id, fabric_name, lot_no, average, sequence) VALUES ('fb_smp_bottle_006_0', 'smp_bottle_006', 'Stainless Steel 304 Grade', 'LOT-SS-2601', 1.0, 0);
INSERT OR REPLACE INTO sample_trims (id, development_id, trim_name, average, sequence) VALUES ('tr_smp_bottle_006_0', 'smp_bottle_006', 'Maax Stain Lifter', 1.0, 0);
INSERT OR REPLACE INTO sample_labels (id, development_id, label_name, sequence) VALUES ('lb_smp_bottle_006_0', 'smp_bottle_006', 'Brand Label - Waterdrop', 0);
INSERT OR REPLACE INTO sample_pharmas (id, development_id, base_pharma_no, changes_made, new_pharma_no, sequence) VALUES ('ph_smp_bottle_006_0', 'smp_bottle_006', 'DWG-BTL-750-A', 'Cap thread pitch adjusted to 3.5mm', 'DWG-BTL-750-B', 0);
INSERT OR REPLACE INTO sample_developments (
    id, development_no, development_date, client_id, client_name, product_genre,
    temp_product_name, sample_image_url, product_description, dev_comments,
    masters_comments, rework_notes, status, created_at, updated_at
) VALUES (
    'smp_bottoms_007', 'D26110007', '2026-11-16', 'ent_cli_altamar_global_exports_8462c771', 'Altamar Global Exports',
    'Sweat Pants', 'Pro-Athletic French Terry Tapered Jogger', 'https://images.unsplash.com/photo-1552902865-b72c031ac5ea?w=600',
    '280 GSM cotton poly loopback jogger with ribbed ankle cuffs and concealed side seam zip pockets.', 'Fit trial on athletic mannequin confirmed knee articulation and pocket depth.',
    'Gusset crotch panel inserted to eliminate seam strain during movement.', '',
    'APPROVED', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_bottoms_007_men_s_0', 'smp_bottoms_007', 'Men', 'S', 1, 0);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_bottoms_007_men_m_1', 'smp_bottoms_007', 'Men', 'M', 2, 1);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_bottoms_007_men_l_2', 'smp_bottoms_007', 'Men', 'L', 2, 2);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_bottoms_007_men_xl_3', 'smp_bottoms_007', 'Men', 'XL', 1, 3);
INSERT OR REPLACE INTO sample_fabrics (id, development_id, fabric_name, lot_no, average, sequence) VALUES ('fb_smp_bottoms_007_0', 'smp_bottoms_007', '3360 ettu pant', 'L-PNT-54', 0.42, 0);
INSERT OR REPLACE INTO sample_trims (id, development_id, trim_name, average, sequence) VALUES ('tr_smp_bottoms_007_0', 'smp_bottoms_007', 'Maax Rust Wash 300', 1.0, 0);
INSERT OR REPLACE INTO sample_zips (id, development_id, zip_name, length_and_qty, sequence) VALUES ('zp_smp_bottoms_007_0', 'smp_bottoms_007', 'Invisible', 'Pockets: 18cm (2pcs)', 0);
INSERT OR REPLACE INTO sample_labels (id, development_id, label_name, sequence) VALUES ('lb_smp_bottoms_007_0', 'smp_bottoms_007', 'Brand Label - American Athlete', 0);
INSERT OR REPLACE INTO sample_pharmas (id, development_id, base_pharma_no, changes_made, new_pharma_no, sequence) VALUES ('ph_smp_bottoms_007_0', 'smp_bottoms_007', 'PH-JOG-202', 'Inseam -1cm, crotch gusset width +2cm', 'PH-JOG-202-V2', 0);
INSERT OR REPLACE INTO sample_developments (
    id, development_no, development_date, client_id, client_name, product_genre,
    temp_product_name, sample_image_url, product_description, dev_comments,
    masters_comments, rework_notes, status, created_at, updated_at
) VALUES (
    'smp_performance_008', 'D26110008', '2026-11-17', 'ent_cli_atc_tires_private_limited_9768afc0', 'ATC Tires Private Limited',
    'Tee', 'Sublimation Poly Mesh Marathon Tee', 'https://images.unsplash.com/photo-1503342217505-b0a15ec3261c?w=600',
    '135 GSM lightweight pinhole mesh quick-dry tee for factory sports day event.', 'Rejected during QC heat press trial: color migration and seam puckering under heat transfer.',
    'Fabric elongation under high temperature caused neckline distortion. Sample archived.', 'Do not proceed. Thermal stability failed.',
    'REJECTED', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_performance_008_men_m_0', 'smp_performance_008', 'Men', 'M', 2, 0);
INSERT OR REPLACE INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES ('sz_smp_performance_008_men_l_1', 'smp_performance_008', 'Men', 'L', 2, 1);
INSERT OR REPLACE INTO sample_fabrics (id, development_id, fabric_name, lot_no, average, sequence) VALUES ('fb_smp_performance_008_0', 'smp_performance_008', '2901poly pointal mesh', 'L-MSH-12', 0.16, 0);
INSERT OR REPLACE INTO sample_trims (id, development_id, trim_name, average, sequence) VALUES ('tr_smp_performance_008_0', 'smp_performance_008', 'LTK - 43 An effective multipurpose stain remover', 1.0, 0);
INSERT OR REPLACE INTO sample_labels (id, development_id, label_name, sequence) VALUES ('lb_smp_performance_008_0', 'smp_performance_008', 'Brand Label - OMG', 0);
INSERT OR REPLACE INTO sample_pharmas (id, development_id, base_pharma_no, changes_made, new_pharma_no, sequence) VALUES ('ph_smp_performance_008_0', 'smp_performance_008', 'PH-TEE-SP-04', 'Standard sports mesh', 'PH-TEE-SP-04-FAIL', 0);
-- 2. Master Catalog Products & Attributes
INSERT OR REPLACE INTO master_products (
    id, product_name, genre, sku_code, brand_label, why_we_made_this,
    garment_specifications, composition, mrp, sample_development_id, created_at, updated_at
) VALUES (
    'prod_atc_polo_001', 'High-Density Pique Corporate Polo', 'Polo', 'ET_PO_HDPIQ', 'ETTU',
    'Engineered for enterprise executives who demand a sharp silhouette from the boardroom to client site visits. Structured pique knit with zero pilling and crisp collar retention after 50+ industrial washes.', '220 GSM combed poly-cotton pique knit; 3-button placket with FPCC signature laser buttons; flat-knit rib collar with Lycra core; taped neckline and side vents.', '60%Cot 40%Poly',
    1499.0, 'smp_polo_001', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO master_product_genders (id, product_id, gender, pharma_no, product_images_json, color_images_json, sequence) VALUES ('mg_prod_atc_polo_001_men_0', 'prod_atc_polo_001', 'Men', 'PH-POL-2401', '["https://images.unsplash.com/photo-1581655353564-df123a1eb820?w=800"]', '[{"color": "Navy Blue", "front_url": "https://images.unsplash.com/photo-1581655353564-df123a1eb820?w=800"}, {"color": "Charcoal Grey", "front_url": "https://images.unsplash.com/photo-1618354691373-d851c5c3a990?w=800"}]', 0);
INSERT OR REPLACE INTO master_product_genders (id, product_id, gender, pharma_no, product_images_json, color_images_json, sequence) VALUES ('mg_prod_atc_polo_001_women_1', 'prod_atc_polo_001', 'Women', 'PH-POL-2401-W', '["https://images.unsplash.com/photo-1598033129183-c4f50c736f10?w=800"]', '[{"color": "Navy Blue", "front_url": "https://images.unsplash.com/photo-1598033129183-c4f50c736f10?w=800"}]', 1);
INSERT OR REPLACE INTO master_product_materials (id, product_id, material_type, name, average, gsm, dia, zip_length_qty, sequence) VALUES ('mm_prod_atc_polo_001_0', 'prod_atc_polo_001', 'fabric', '2441 Galactic Polo', 0.28, 220, 32, NULL, 0);
INSERT OR REPLACE INTO master_product_materials (id, product_id, material_type, name, average, gsm, dia, zip_length_qty, sequence) VALUES ('mm_prod_atc_polo_001_1', 'prod_atc_polo_001', 'trim', 'FPCC Signature Laser Button', 3.0, NULL, NULL, NULL, 1);
INSERT OR REPLACE INTO master_product_materials (id, product_id, material_type, name, average, gsm, dia, zip_length_qty, sequence) VALUES ('mm_prod_atc_polo_001_2', 'prod_atc_polo_001', 'trim', 'High Density Woven Fusing', 0.05, NULL, NULL, NULL, 2);
INSERT OR REPLACE INTO master_products (
    id, product_name, genre, sku_code, brand_label, why_we_made_this,
    garment_specifications, composition, mrp, sample_development_id, created_at, updated_at
) VALUES (
    'prod_amara_hoodie_002', 'Heavyweight Fleece Tech Zip Hoodie', 'Hoodies and Pullovers', 'AA_HO_TECHZ', 'American Athlete',
    'Built for mobility engineers operating in testing bays and winter morning plant shifts. Heavyweight thermal insulation with ergonomic raglan sleeves.', '320 GSM 3-thread brushed box fleece; antique brass #5 full-zip; ergonomic 3-panel hood with metal eyelets; deep kangaroo pockets; 2x2 spandex rib.', '52%Cot 48%Poly',
    2499.0, 'smp_hoodie_002', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO master_product_genders (id, product_id, gender, pharma_no, product_images_json, color_images_json, sequence) VALUES ('mg_prod_amara_hoodie_002_unisex_0', 'prod_amara_hoodie_002', 'Unisex', 'PH-HOD-3205', '["https://images.unsplash.com/photo-1556905055-8f358a7a47b2?w=800"]', '[{"color": "Jet Black", "front_url": "https://images.unsplash.com/photo-1556905055-8f358a7a47b2?w=800"}, {"color": "Heather Grey", "front_url": "https://images.unsplash.com/photo-1556905055-8f358a7a47b2?w=800"}]', 0);
INSERT OR REPLACE INTO master_product_materials (id, product_id, material_type, name, average, gsm, dia, zip_length_qty, sequence) VALUES ('mm_prod_amara_hoodie_002_0', 'prod_amara_hoodie_002', 'fabric', '2111 bonded box fleece', 0.65, 320, 34, NULL, 0);
INSERT OR REPLACE INTO master_product_materials (id, product_id, material_type, name, average, gsm, dia, zip_length_qty, sequence) VALUES ('mm_prod_amara_hoodie_002_1', 'prod_amara_hoodie_002', 'zip', 'Brass', 1.0, NULL, NULL, '65cm-70cm (1pc)', 1);
INSERT OR REPLACE INTO master_product_materials (id, product_id, material_type, name, average, gsm, dia, zip_length_qty, sequence) VALUES ('mm_prod_amara_hoodie_002_2', 'prod_amara_hoodie_002', 'trim', 'Rollax Powder', 1.0, NULL, NULL, NULL, 2);
INSERT OR REPLACE INTO master_products (
    id, product_name, genre, sku_code, brand_label, why_we_made_this,
    garment_specifications, composition, mrp, sample_development_id, created_at, updated_at
) VALUES (
    'prod_acma_tee_003', 'EV Mobility Crew Neck Bio-Wash Tee', 'Tee', 'CI_TE_EVCREW', 'Citreon',
    'Designed for all-day conference comfort and active workshop participation. Signature super-combed cotton and double bio-wash process ensures zero skin friction.', '180 GSM single jersey 100% super-combed cotton; enzyme bio-washed; seamless tubular body construction with 1x1 Lycra ribbed neckline.', '100%Cot',
    799.0, 'smp_tee_003', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO master_product_genders (id, product_id, gender, pharma_no, product_images_json, color_images_json, sequence) VALUES ('mg_prod_acma_tee_003_men_0', 'prod_acma_tee_003', 'Men', 'PH-TEE-REG-01', '["https://images.unsplash.com/photo-1521572267360-ee0c2909d518?w=800"]', '[{"color": "Royal Blue", "front_url": "https://images.unsplash.com/photo-1521572267360-ee0c2909d518?w=800"}]', 0);
INSERT OR REPLACE INTO master_product_genders (id, product_id, gender, pharma_no, product_images_json, color_images_json, sequence) VALUES ('mg_prod_acma_tee_003_women_1', 'prod_acma_tee_003', 'Women', 'PH-TEE-REG-01W', '["https://images.unsplash.com/photo-1503342217505-b0a15ec3261c?w=800"]', '[{"color": "Royal Blue", "front_url": "https://images.unsplash.com/photo-1503342217505-b0a15ec3261c?w=800"}]', 1);
INSERT OR REPLACE INTO master_product_materials (id, product_id, material_type, name, average, gsm, dia, zip_length_qty, sequence) VALUES ('mm_prod_acma_tee_003_0', 'prod_acma_tee_003', 'fabric', '30s Sinker Crew Bio Wash', 0.22, 180, 30, NULL, 0);
INSERT OR REPLACE INTO master_product_materials (id, product_id, material_type, name, average, gsm, dia, zip_length_qty, sequence) VALUES ('mm_prod_acma_tee_003_1', 'prod_acma_tee_003', 'trim', '30s Sinker Crew Bio Wash -Rib', 0.04, NULL, NULL, NULL, 1);
INSERT OR REPLACE INTO master_products (
    id, product_name, genre, sku_code, brand_label, why_we_made_this,
    garment_specifications, composition, mrp, sample_development_id, created_at, updated_at
) VALUES (
    'prod_apex_windbreaker_004', 'Water-Resistant All-Weather Windbreaker', 'Jackets', 'JP_JK_ALLWTH', 'Jeep',
    'Created for rugged field reps facing sudden tropical downpours and mountain terrain drafts. Features a lightweight micro-ripstop shell with DWR finish.', 'Durable water-repellent micro-ripstop shell; internal breathable mesh back vent; 3-piece stowaway hood; waterproof reverse coil zippers.', '100%Poly',
    3299.0, 'smp_jacket_004', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO master_product_genders (id, product_id, gender, pharma_no, product_images_json, color_images_json, sequence) VALUES ('mg_prod_apex_windbreaker_004_men_0', 'prod_apex_windbreaker_004', 'Men', 'PH-JKT-101R', '["https://images.unsplash.com/photo-1548883354-7622d03aca27?w=800"]', '[{"color": "Olive Green", "front_url": "https://images.unsplash.com/photo-1548883354-7622d03aca27?w=800"}]', 0);
INSERT OR REPLACE INTO master_product_materials (id, product_id, material_type, name, average, gsm, dia, zip_length_qty, sequence) VALUES ('mm_prod_apex_windbreaker_004_0', 'prod_apex_windbreaker_004', 'fabric', '3347 shark plain', 0.48, 160, 58, NULL, 0);
INSERT OR REPLACE INTO master_product_materials (id, product_id, material_type, name, average, gsm, dia, zip_length_qty, sequence) VALUES ('mm_prod_apex_windbreaker_004_1', 'prod_apex_windbreaker_004', 'trim', '357 mesh', 0.18, NULL, NULL, NULL, 1);
INSERT OR REPLACE INTO master_product_materials (id, product_id, material_type, name, average, gsm, dia, zip_length_qty, sequence) VALUES ('mm_prod_apex_windbreaker_004_2', 'prod_apex_windbreaker_004', 'zip', 'Coil', 2.0, NULL, NULL, 'Front 72cm (1pc), Pockets 18cm (2pcs)', 2);
INSERT OR REPLACE INTO master_products (
    id, product_name, genre, sku_code, brand_label, why_we_made_this,
    garment_specifications, composition, mrp, sample_development_id, created_at, updated_at
) VALUES (
    'prod_bapl_tote_005', 'Heavy-Duty Industrial Canvas Tote Bag', 'Bags', 'ET_BG_CANVAS', 'ETTU',
    'Indestructible, sustainable alternative to synthetic goodie bags. Heavy-duty 16oz duck cotton canvas with box-X stitching on handles, holding 25kg.', '45cm (H) x 40cm (W) x 12cm (Gusset); 16oz raw untreated heavy cotton duck canvas; 38mm wide heavy cotton webbing shoulder handles; antique brass magnetic snap.', '100%Cot',
    899.0, 'smp_bag_005', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO master_product_genders (id, product_id, gender, pharma_no, product_images_json, color_images_json, sequence) VALUES ('mg_prod_bapl_tote_005_one size_0', 'prod_bapl_tote_005', 'One Size', 'TOOL-TOTE-16B', '["https://images.unsplash.com/photo-1544816155-12df9643f363?w=800"]', '[{"color": "Natural Ecru", "front_url": "https://images.unsplash.com/photo-1544816155-12df9643f363?w=800"}]', 0);
INSERT OR REPLACE INTO master_product_materials (id, product_id, material_type, name, average, gsm, dia, zip_length_qty, sequence) VALUES ('mm_prod_bapl_tote_005_0', 'prod_bapl_tote_005', 'fabric', '30s cotton slub', 0.35, 450, 58, NULL, 0);
INSERT OR REPLACE INTO master_product_materials (id, product_id, material_type, name, average, gsm, dia, zip_length_qty, sequence) VALUES ('mm_prod_bapl_tote_005_1', 'prod_bapl_tote_005', 'zip', 'Metal', 1.0, NULL, NULL, '20cm (1pc)', 1);
INSERT OR REPLACE INTO master_product_materials (id, product_id, material_type, name, average, gsm, dia, zip_length_qty, sequence) VALUES ('mm_prod_bapl_tote_005_2', 'prod_bapl_tote_005', 'trim', 'Stainoff', 1.0, NULL, NULL, NULL, 2);
INSERT OR REPLACE INTO master_products (
    id, product_name, genre, sku_code, brand_label, why_we_made_this,
    garment_specifications, composition, mrp, sample_development_id, created_at, updated_at
) VALUES (
    'prod_jogger_006', 'Pro-Athletic French Terry Tapered Jogger', 'Sweat Pants', 'AA_SP_TERRY', 'American Athlete',
    'Engineered for movement freedom, warm-up drills, and off-duty travel. Incorporates an articulated diamond crotch gusset that distributes seam stress.', '280 GSM cotton poly French terry; tapered athletic leg with wide rib cuffs; deep concealed zip welt pockets; wide elasticated waist.', '95%Cot 5%SP',
    1899.0, 'smp_bottoms_007', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO master_product_genders (id, product_id, gender, pharma_no, product_images_json, color_images_json, sequence) VALUES ('mg_prod_jogger_006_men_0', 'prod_jogger_006', 'Men', 'PH-JOG-202-V2', '["https://images.unsplash.com/photo-1552902865-b72c031ac5ea?w=800"]', '[{"color": "Anthracite", "front_url": "https://images.unsplash.com/photo-1552902865-b72c031ac5ea?w=800"}]', 0);
INSERT OR REPLACE INTO master_product_materials (id, product_id, material_type, name, average, gsm, dia, zip_length_qty, sequence) VALUES ('mm_prod_jogger_006_0', 'prod_jogger_006', 'fabric', '3360 ettu pant', 0.42, 280, 32, NULL, 0);
INSERT OR REPLACE INTO master_product_materials (id, product_id, material_type, name, average, gsm, dia, zip_length_qty, sequence) VALUES ('mm_prod_jogger_006_1', 'prod_jogger_006', 'zip', 'Invisible', 2.0, NULL, NULL, '18cm (2pcs)', 1);
-- 3. Client Specific Contract Pricing
INSERT OR REPLACE INTO client_product_pricing (
    id, client_id, client_name, product_id, product_name, contracted_price, comments, effective_date, created_at, updated_at
) VALUES (
    'prc_atc_01', 'ent_cli_atc_tires_private_limited_9768afc0', 'ATC Tires Private Limited', 'prod_atc_polo_001', 'High-Density Pique Corporate Polo', 685.0, 'Annual management apparel contract (5,000 pcs commitment)', '2026-11-01', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO client_product_pricing (
    id, client_id, client_name, product_id, product_name, contracted_price, comments, effective_date, created_at, updated_at
) VALUES (
    'prc_amara_02', 'ent_cli_amara_raja_energy_mobility_limited_95ac4e91', 'Amara Raja Energy & Mobility Limited', 'prod_amara_hoodie_002', 'Heavyweight Fleece Tech Zip Hoodie', 1250.0, 'Winter plant crew order tier (2,500 pcs)', '2026-11-01', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO client_product_pricing (
    id, client_id, client_name, product_id, product_name, contracted_price, comments, effective_date, created_at, updated_at
) VALUES (
    'prc_acma_03', 'ent_cli_acma_mobility_foundation_4d29096f', 'Acma Mobility Foundation', 'prod_acma_tee_003', 'EV Mobility Crew Neck Bio-Wash Tee', 340.0, 'National EV Summit official delegate t-shirt', '2026-11-01', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO client_product_pricing (
    id, client_id, client_name, product_id, product_name, contracted_price, comments, effective_date, created_at, updated_at
) VALUES (
    'prc_apex_04', 'ent_cli_apex_trading_d985c8b1', 'Apex Trading', 'prod_apex_windbreaker_004', 'Water-Resistant All-Weather Windbreaker', 1620.0, 'Export outerwear agreement', '2026-11-01', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO client_product_pricing (
    id, client_id, client_name, product_id, product_name, contracted_price, comments, effective_date, created_at, updated_at
) VALUES (
    'prc_bapl_05', 'ent_cli_bapl_rototech_private_limited_6d9806e8', 'Bapl Rototech Private Limited', 'prod_bapl_tote_005', 'Heavy-Duty Industrial Canvas Tote Bag', 395.0, 'Trade expo corporate kit bag', '2026-11-01', 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO client_product_pricing (
    id, client_id, client_name, product_id, product_name, contracted_price, comments, effective_date, created_at, updated_at
) VALUES (
    'prc_alt_06', 'ent_cli_altamar_global_exports_8462c771', 'Altamar Global Exports', 'prod_jogger_006', 'Pro-Athletic French Terry Tapered Jogger', 890.0, 'Fitness franchise recurring order', '2026-11-01', 1795074719192, 1795074719192
);
-- 4. Commercial & Production Orders (Strict MMYY0000 Sequence)
INSERT OR REPLACE INTO orders (
    id, production_order_no, sales_order_no, order_type, booking_date, client_id,
    client_name, contact_person, phone, email, billing_address, shipping_address,
    gstin, category, sub_category, product_name, color_name, fabric_name, gender,
    unit_price, tax_rate, tax_billing, freight_mode, freight_cost, total_qty,
    size_breakdown_json, multiple_shipping, shipping_locations_json, target_etd,
    actual_disp_date, carrier, docket_no, status, total_order_value, created_at, updated_at
) VALUES (
    'ord_11260001', '11260001', 'SO-2026-ATC-01', 'Bulk', '2026-11-10',
    'ent_cli_atc_tires_private_limited_9768afc0', 'ATC Tires Private Limited', 'Vikram Malhotra',
    '+91 98201 44552', 'v.malhotra@atctires.com', 'Plot B-12, MIDC Industrial Area, Tarapur, Maharashtra 401506', 'Plot B-12, MIDC Industrial Area, Tarapur, Maharashtra 401506',
    '27AACCN5773A1ZL', 'Apparel', 'Polo', 'High-Density Pique Corporate Polo',
    'Navy Blue', '2441 Galactic Polo', 'Men', 685.0, 5.0,
    'Extra', 'Surface', 'Inclusive', 800,
    '{"Men": {"S": 50, "M": 120, "L": 130, "XL": 60, "2XL": 20}, "Women": {"S": 10, "M": 10}}', 0, NULL, '2026-12-05', NULL, NULL, NULL,
    'BOOKED', 432000.0, 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO order_items (
    id, order_id, item_sequence, category, sub_category, product_name,
    color_name, fabric_name, unit_price, total_qty, matrix_json, created_at
) VALUES (
    'item_11260001_1', 'ord_11260001', 1, 'Apparel', 'Polo',
    'High-Density Pique Corporate Polo', 'Navy Blue', '2441 Galactic Polo',
    685.0, 400, '{"Men": {"S": 50, "M": 120, "L": 130, "XL": 60, "2XL": 20}, "Women": {"S": 10, "M": 10}}', 1795074719192
);
INSERT OR REPLACE INTO order_item_brandings (
    id, order_item_id, order_id, branding_type, position, size, placement, color, comment, created_at
) VALUES (
    'br_item_11260001_1_0', 'item_11260001_1', 'ord_11260001', 'emb', 'Left Chest', '3.5 inches', 'Direct', 'White / Gold Poly', 'High-density stitch embroidery file atc_crest.dst', 1795074719192
);
INSERT OR REPLACE INTO order_items (
    id, order_id, item_sequence, category, sub_category, product_name,
    color_name, fabric_name, unit_price, total_qty, matrix_json, created_at
) VALUES (
    'item_11260001_2', 'ord_11260001', 2, 'Non-Apparel', 'Bags',
    'Heavy-Duty Industrial Canvas Tote Bag', 'Natural Ecru', '30s cotton slub',
    395.0, 400, '{"One Size": {"One Size": 400}}', 1795074719192
);
INSERT OR REPLACE INTO order_item_brandings (
    id, order_item_id, order_id, branding_type, position, size, placement, color, comment, created_at
) VALUES (
    'br_item_11260001_2_0', 'item_11260001_2', 'ord_11260001', 'print', 'Front Center', '10x10 inches', 'Center', 'Navy Blue', 'Screen print 2-pass plastisol ink with flash cure', 1795074719192
);
INSERT OR REPLACE INTO orders (
    id, production_order_no, sales_order_no, order_type, booking_date, client_id,
    client_name, contact_person, phone, email, billing_address, shipping_address,
    gstin, category, sub_category, product_name, color_name, fabric_name, gender,
    unit_price, tax_rate, tax_billing, freight_mode, freight_cost, total_qty,
    size_breakdown_json, multiple_shipping, shipping_locations_json, target_etd,
    actual_disp_date, carrier, docket_no, status, total_order_value, created_at, updated_at
) VALUES (
    'ord_11260002', '11260002', 'SO-2026-AR-88', 'Bulk', '2026-11-12',
    'ent_cli_amara_raja_energy_mobility_limited_95ac4e91', 'Amara Raja Energy & Mobility Limited', 'K. Srinivas',
    '+91 87722 65000', 'ksrinivas@amararaja.com', 'Renigunta Road, Karakambadi, Tirupati, Andhra Pradesh 517520', 'Renigunta Road, Karakambadi, Tirupati, Andhra Pradesh 517520',
    '37AABCA9264E3ZM', 'Apparel', 'Hoodies and Pullovers', 'Heavyweight Fleece Tech Zip Hoodie',
    'Jet Black', '2111 bonded box fleece', 'Unisex', 1250.0, 12.0,
    'Extra', 'Surface', 'Extra', 250,
    '{"Men": {"M": 80, "L": 100, "XL": 50, "2XL": 20}}', 0, NULL, '2026-12-15', NULL, NULL, NULL,
    'PLANNED', 312500.0, 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO order_items (
    id, order_id, item_sequence, category, sub_category, product_name,
    color_name, fabric_name, unit_price, total_qty, matrix_json, created_at
) VALUES (
    'item_11260002_1', 'ord_11260002', 1, 'Apparel', 'Hoodies and Pullovers',
    'Heavyweight Fleece Tech Zip Hoodie', 'Jet Black', '2111 bonded box fleece',
    1250.0, 250, '{"Men": {"M": 80, "L": 100, "XL": 50, "2XL": 20}}', 1795074719192
);
INSERT OR REPLACE INTO order_item_brandings (
    id, order_item_id, order_id, branding_type, position, size, placement, color, comment, created_at
) VALUES (
    'br_item_11260002_1_0', 'item_11260002_1', 'ord_11260002', 'emb', 'Left Bicep', '2.5 inches', 'Sleeve', 'Silver 3D', 'Raised silicone 3D heat seal badge', 1795074719192
);
INSERT OR REPLACE INTO order_personalizations (
    id, order_item_id, order_id, gender, size, name, number, remarks, created_at
) VALUES (
    'pers_item_11260002_1_0', 'item_11260002_1', 'ord_11260002', 'Men', 'L', 'S. RAMESH', '01', 'R&D Lead', 1795074719192
);
INSERT OR REPLACE INTO order_personalizations (
    id, order_item_id, order_id, gender, size, name, number, remarks, created_at
) VALUES (
    'pers_item_11260002_1_1', 'item_11260002_1', 'ord_11260002', 'Men', 'XL', 'V. SHARMA', '02', 'EV Battery Head', 1795074719192
);
INSERT OR REPLACE INTO order_personalizations (
    id, order_item_id, order_id, gender, size, name, number, remarks, created_at
) VALUES (
    'pers_item_11260002_1_2', 'item_11260002_1', 'ord_11260002', 'Men', 'M', 'A. MEHTA', '07', 'Plant Operations', 1795074719192
);
INSERT OR REPLACE INTO order_personalizations (
    id, order_item_id, order_id, gender, size, name, number, remarks, created_at
) VALUES (
    'pers_item_11260002_1_3', 'item_11260002_1', 'ord_11260002', 'Men', 'L', 'R. REDDY', '10', 'Quality Assurance', 1795074719192
);
INSERT OR REPLACE INTO orders (
    id, production_order_no, sales_order_no, order_type, booking_date, client_id,
    client_name, contact_person, phone, email, billing_address, shipping_address,
    gstin, category, sub_category, product_name, color_name, fabric_name, gender,
    unit_price, tax_rate, tax_billing, freight_mode, freight_cost, total_qty,
    size_breakdown_json, multiple_shipping, shipping_locations_json, target_etd,
    actual_disp_date, carrier, docket_no, status, total_order_value, created_at, updated_at
) VALUES (
    'ord_11260003', '11260003', 'SO-2026-ACMA-12', 'Bulk', '2026-11-08',
    'ent_cli_acma_mobility_foundation_4d29096f', 'Acma Mobility Foundation', 'Pooja Deshmukh',
    '+91 99302 11984', 'pooja.d@acma.in', 'Trade Centre, BKC, Bandra East, Mumbai, Maharashtra 400051', 'Trade Centre, BKC, Bandra East, Mumbai, Maharashtra 400051',
    '27ABACA9032G1ZU', 'Apparel', 'Tee', 'EV Mobility Crew Neck Bio-Wash Tee',
    'Royal Blue', '30s Sinker Crew Bio Wash', 'Men', 340.0, 5.0,
    'Extra', 'Courier', 'To be paid by client', 600,
    '{"Men": {"S": 100, "M": 200, "L": 180, "XL": 70, "2XL": 10}, "Women": {"XS": 10, "S": 15, "M": 15}}', 0, NULL, '2026-11-28', NULL, NULL, NULL,
    'IN_CUTTING', 204000.0, 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO order_items (
    id, order_id, item_sequence, category, sub_category, product_name,
    color_name, fabric_name, unit_price, total_qty, matrix_json, created_at
) VALUES (
    'item_11260003_1', 'ord_11260003', 1, 'Apparel', 'Tee',
    'EV Mobility Crew Neck Bio-Wash Tee', 'Royal Blue', '30s Sinker Crew Bio Wash',
    340.0, 600, '{"Men": {"S": 100, "M": 200, "L": 180, "XL": 70, "2XL": 10}, "Women": {"XS": 10, "S": 15, "M": 15}}', 1795074719192
);
INSERT OR REPLACE INTO order_item_brandings (
    id, order_item_id, order_id, branding_type, position, size, placement, color, comment, created_at
) VALUES (
    'br_item_11260003_1_0', 'item_11260003_1', 'ord_11260003', 'print', 'Front Chest', '9x3 inches', 'Chest', 'Multi-Color', 'Direct to Film (DTF) full vibrant color print', 1795074719192
);
INSERT OR REPLACE INTO orders (
    id, production_order_no, sales_order_no, order_type, booking_date, client_id,
    client_name, contact_person, phone, email, billing_address, shipping_address,
    gstin, category, sub_category, product_name, color_name, fabric_name, gender,
    unit_price, tax_rate, tax_billing, freight_mode, freight_cost, total_qty,
    size_breakdown_json, multiple_shipping, shipping_locations_json, target_etd,
    actual_disp_date, carrier, docket_no, status, total_order_value, created_at, updated_at
) VALUES (
    'ord_11260004', '11260004', 'SO-2026-APX-04', 'Development', '2026-11-05',
    'ent_cli_apex_trading_d985c8b1', 'Apex Trading', 'Tariq Mansoor',
    '+91 97110 56230', 'tariq@apextrading.co.in', 'Seepz-SEZ, Andheri East, Mumbai 400096', 'Seepz-SEZ, Andheri East, Mumbai 400096',
    '27AQOPM7456J1Z3', 'Apparel', 'Jackets', 'Water-Resistant All-Weather Windbreaker',
    'Olive Green', '3347 shark plain', 'Men', 1620.0, 12.0,
    'Extra', 'Air', 'Inclusive', 180,
    '{"Men": {"M": 60, "L": 80, "XL": 40}}', 0, NULL, '2026-11-25', NULL, NULL, NULL,
    'IN_PRODUCTION', 291600.0, 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO order_items (
    id, order_id, item_sequence, category, sub_category, product_name,
    color_name, fabric_name, unit_price, total_qty, matrix_json, created_at
) VALUES (
    'item_11260004_1', 'ord_11260004', 1, 'Apparel', 'Jackets',
    'Water-Resistant All-Weather Windbreaker', 'Olive Green', '3347 shark plain',
    1620.0, 180, '{"Men": {"M": 60, "L": 80, "XL": 40}}', 1795074719192
);
INSERT OR REPLACE INTO order_item_brandings (
    id, order_item_id, order_id, branding_type, position, size, placement, color, comment, created_at
) VALUES (
    'br_item_11260004_1_0', 'item_11260004_1', 'ord_11260004', 'print', 'Right Chest', '3x2 inches', 'Chest', 'Silver Reflective', '3M Scotchlite reflective transfer', 1795074719192
);
INSERT OR REPLACE INTO orders (
    id, production_order_no, sales_order_no, order_type, booking_date, client_id,
    client_name, contact_person, phone, email, billing_address, shipping_address,
    gstin, category, sub_category, product_name, color_name, fabric_name, gender,
    unit_price, tax_rate, tax_billing, freight_mode, freight_cost, total_qty,
    size_breakdown_json, multiple_shipping, shipping_locations_json, target_etd,
    actual_disp_date, carrier, docket_no, status, total_order_value, created_at, updated_at
) VALUES (
    'ord_11260005', '11260005', 'SO-2026-BAPL-09', 'Bulk', '2026-11-15',
    'ent_cli_bapl_rototech_private_limited_6d9806e8', 'Bapl Rototech Private Limited', 'Sunil Nambiar',
    '+91 75540 82211', 'sunil.n@baplrototech.com', 'Sector F, Sanwer Road Industrial Area, Indore, Madhya Pradesh 452015', 'Multi-destination consignment',
    '23AAGCB4215M1ZT', 'Non-Apparel', 'Bags', 'Heavy-Duty Industrial Canvas Tote Bag',
    'Natural Ecru', '30s cotton slub', 'Unisex', 395.0, 5.0,
    'Included', 'Surface', 'Inclusive', 500,
    '{"One Size": {"One Size": 500}}', 1, '[{"destination": "Pune Plant: Gate 3, Chakan MIDC, Phase II, Pune 410501", "quantity": 300, "carrier": "VRL Logistics", "docketNo": "VRL-PN-9921", "etd": "2026-12-20"}, {"destination": "Chennai Warehouse: Ambattur Industrial Estate, Chennai 600058", "quantity": 200, "carrier": "Safexpress", "docketNo": "SFX-CH-4412", "etd": "2026-12-22"}]', '2026-12-20', NULL, NULL, NULL,
    'BOOKED', 197500.0, 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO order_items (
    id, order_id, item_sequence, category, sub_category, product_name,
    color_name, fabric_name, unit_price, total_qty, matrix_json, created_at
) VALUES (
    'item_11260005_1', 'ord_11260005', 1, 'Non-Apparel', 'Bags',
    'Heavy-Duty Industrial Canvas Tote Bag', 'Natural Ecru', '30s cotton slub',
    395.0, 500, '{"One Size": {"One Size": 500}}', 1795074719192
);
INSERT OR REPLACE INTO order_item_brandings (
    id, order_item_id, order_id, branding_type, position, size, placement, color, comment, created_at
) VALUES (
    'br_item_11260005_1_0', 'item_11260005_1', 'ord_11260005', 'print', 'Front Pocket', '6x4 inches', 'Pocket', 'Forest Green', 'Silk screen print with company monogram', 1795074719192
);
INSERT OR REPLACE INTO orders (
    id, production_order_no, sales_order_no, order_type, booking_date, client_id,
    client_name, contact_person, phone, email, billing_address, shipping_address,
    gstin, category, sub_category, product_name, color_name, fabric_name, gender,
    unit_price, tax_rate, tax_billing, freight_mode, freight_cost, total_qty,
    size_breakdown_json, multiple_shipping, shipping_locations_json, target_etd,
    actual_disp_date, carrier, docket_no, status, total_order_value, created_at, updated_at
) VALUES (
    'ord_11260006', '11260006', 'SO-2026-ALT-99', 'Repeat', '2026-10-20',
    'ent_cli_altamar_global_exports_8462c771', 'Altamar Global Exports', 'Aditi Roy',
    '+91 98331 77620', 'aditi.roy@altamarglobal.com', 'Express Towers, Nariman Point, Mumbai 400021', 'Express Towers, Nariman Point, Mumbai 400021',
    '27AQXPP8924F1Z1', 'Apparel', 'Sweat Pants', 'Pro-Athletic French Terry Tapered Jogger',
    'Anthracite', '3360 ettu pant', 'Men', 890.0, 12.0,
    'Extra', 'Courier', 'Inclusive', 350,
    '{"Men": {"S": 50, "M": 150, "L": 100, "XL": 50}}', 0, NULL, '2026-11-10', '2026-11-09', 'Blue Dart Apex', 'BD-8829104',
    'COMPLETED', 311500.0, 1795074719192, 1795074719192
);
INSERT OR REPLACE INTO order_items (
    id, order_id, item_sequence, category, sub_category, product_name,
    color_name, fabric_name, unit_price, total_qty, matrix_json, created_at
) VALUES (
    'item_11260006_1', 'ord_11260006', 1, 'Apparel', 'Sweat Pants',
    'Pro-Athletic French Terry Tapered Jogger', 'Anthracite', '3360 ettu pant',
    890.0, 350, '{"Men": {"S": 50, "M": 150, "L": 100, "XL": 50}}', 1795074719192
);
INSERT OR REPLACE INTO order_item_brandings (
    id, order_item_id, order_id, branding_type, position, size, placement, color, comment, created_at
) VALUES (
    'br_item_11260006_1_0', 'item_11260006_1', 'ord_11260006', 'emb', 'Left Thigh', '2.5 inches', 'Thigh', 'White Poly', 'Woven athletic emblem badge', 1795074719192
);
-- 5. Three-Stage Inward Pipeline & Discrepancies
INSERT OR REPLACE INTO inward_invoices (
    id, invoice_no, vendor_name, invoice_date, invoice_qty, invoice_amount, uom, currency, remarks, created_at
) VALUES (
    'inv_ark_261101', 'INV-ARK-261101', 'Arkap Knits Pvt Ltd / Rr Knits', '2026-11-17',
    850.0, 314500.0, 'Kgs', 'INR', 'Tax Invoice with GST e-Way bill compliance verified', 1795074719192
);
INSERT OR REPLACE INTO parent_inward_lots (
    id, parent_code, inward_date, vendor_id, vendor_name, invoice_no, inward_type,
    uom, invoice_qty, invoice_amount, physical_received_qty, shortfall_qty,
    discrepancy_amount, fabric_type, vendor_fab_name, fpcc_fab_name, vendor_color,
    fpcc_color, invoice_gsm, dia, stage, unit_rate, counted_qty, variance_qty,
    invoice_id, created_at
) VALUES (
    'lot_261117_a', '261117-A', '2026-11-17', 'ent_ven_arkap_knits_pvt_ltd_rr_knits_1f90c965',
    'Arkap Knits Pvt Ltd / Rr Knits', 'INV-ARK-261101', 'Fabric Knitted Roll',
    'Kgs', 850.0, 314500.0, 850.0, 0.0, 0.0,
    'Knitted', '2441 Galactic Polo', '2441 Galactic Polo', 'Navy Blue', 'Navy Blue',
    220, 32, 'AWAITING_COUNT', 370.0, 850.0, 0.0,
    'inv_ark_261101', 1795074719192
);
INSERT OR REPLACE INTO inward_invoices (
    id, invoice_no, vendor_name, invoice_date, invoice_qty, invoice_amount, uom, currency, remarks, created_at
) VALUES (
    'inv_and_8842', 'INV-AND-8842', 'Anandghan Technologies', '2026-11-16',
    620.0, 248000.0, 'Kgs', 'INR', 'Tax Invoice with GST e-Way bill compliance verified', 1795074719192
);
INSERT OR REPLACE INTO parent_inward_lots (
    id, parent_code, inward_date, vendor_id, vendor_name, invoice_no, inward_type,
    uom, invoice_qty, invoice_amount, physical_received_qty, shortfall_qty,
    discrepancy_amount, fabric_type, vendor_fab_name, fpcc_fab_name, vendor_color,
    fpcc_color, invoice_gsm, dia, stage, unit_rate, counted_qty, variance_qty,
    invoice_id, created_at
) VALUES (
    'lot_261116_b', '261116-B', '2026-11-16', 'ent_ven_anandghan_technologies_f6aa1f26',
    'Anandghan Technologies', 'INV-AND-8842', 'Fabric Knitted Roll',
    'Kgs', 620.0, 248000.0, 620.0, 0.0, 0.0,
    'Knitted', '2111 bonded box fleece', '2111 bonded box fleece', 'Jet Black', 'Jet Black',
    320, 34, 'AWAITING_QC', 400.0, 620.0, 0.0,
    'inv_and_8842', 1795074719192
);
INSERT OR REPLACE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, inward_weight_qty, measured_gsm, location_bin,
    qc_status, qr_payload, available_qty, rate, measured_dia, created_at
) VALUES (
    'bnd_261116-b-1_4', '261116-B-1/4', 'lot_261116_b', '261116-B', 1, 4,
    'fabric', '2111 bonded box fleece', 'Jet Black', 155.2, 320, 'BAY-INW-02',
    'Pending', '{"bundleCode": "261116-B-1/4", "parentCode": "261116-B", "vendor": "Anandghan Technologies", "fabric": "2111 bonded box fleece", "color": "Jet Black", "weight": 155.2, "gsm": 320, "dia": 34, "bin": "BAY-INW-02"}', 155.2, 400.0, 34, 1795074719192
);
INSERT OR REPLACE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, inward_weight_qty, measured_gsm, location_bin,
    qc_status, qr_payload, available_qty, rate, measured_dia, created_at
) VALUES (
    'bnd_261116-b-2_4', '261116-B-2/4', 'lot_261116_b', '261116-B', 2, 4,
    'fabric', '2111 bonded box fleece', 'Jet Black', 154.8, 320, 'BAY-INW-02',
    'Pending', '{"bundleCode": "261116-B-2/4", "parentCode": "261116-B", "vendor": "Anandghan Technologies", "fabric": "2111 bonded box fleece", "color": "Jet Black", "weight": 154.8, "gsm": 320, "dia": 34, "bin": "BAY-INW-02"}', 154.8, 400.0, 34, 1795074719192
);
INSERT OR REPLACE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, inward_weight_qty, measured_gsm, location_bin,
    qc_status, qr_payload, available_qty, rate, measured_dia, created_at
) VALUES (
    'bnd_261116-b-3_4', '261116-B-3/4', 'lot_261116_b', '261116-B', 3, 4,
    'fabric', '2111 bonded box fleece', 'Jet Black', 156.0, 320, 'BAY-INW-02',
    'Pending', '{"bundleCode": "261116-B-3/4", "parentCode": "261116-B", "vendor": "Anandghan Technologies", "fabric": "2111 bonded box fleece", "color": "Jet Black", "weight": 156.0, "gsm": 320, "dia": 34, "bin": "BAY-INW-02"}', 156.0, 400.0, 34, 1795074719192
);
INSERT OR REPLACE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, inward_weight_qty, measured_gsm, location_bin,
    qc_status, qr_payload, available_qty, rate, measured_dia, created_at
) VALUES (
    'bnd_261116-b-4_4', '261116-B-4/4', 'lot_261116_b', '261116-B', 4, 4,
    'fabric', '2111 bonded box fleece', 'Jet Black', 154.0, 320, 'BAY-INW-02',
    'Pending', '{"bundleCode": "261116-B-4/4", "parentCode": "261116-B", "vendor": "Anandghan Technologies", "fabric": "2111 bonded box fleece", "color": "Jet Black", "weight": 154.0, "gsm": 320, "dia": 34, "bin": "BAY-INW-02"}', 154.0, 400.0, 34, 1795074719192
);
INSERT OR REPLACE INTO inward_invoices (
    id, invoice_no, vendor_name, invoice_date, invoice_qty, invoice_amount, uom, currency, remarks, created_at
) VALUES (
    'inv_anl_9021', 'INV-ANL-9021', 'Anil International', '2026-11-15',
    500.0, 185000.0, 'Kgs', 'INR', 'Tax Invoice with GST e-Way bill compliance verified', 1795074719192
);
INSERT OR REPLACE INTO parent_inward_lots (
    id, parent_code, inward_date, vendor_id, vendor_name, invoice_no, inward_type,
    uom, invoice_qty, invoice_amount, physical_received_qty, shortfall_qty,
    discrepancy_amount, fabric_type, vendor_fab_name, fpcc_fab_name, vendor_color,
    fpcc_color, invoice_gsm, dia, stage, unit_rate, counted_qty, variance_qty,
    invoice_id, created_at
) VALUES (
    'lot_261115_c', '261115-C', '2026-11-15', 'ent_ven_anil_international_289e7a92',
    'Anil International', 'INV-ANL-9021', 'Fabric Knitted Roll',
    'Kgs', 500.0, 185000.0, 472.5, 27.5, 10175.0,
    'Knitted', '30s Sinker Crew Bio Wash', '30s Sinker Crew Bio Wash', 'Royal Blue', 'Royal Blue',
    180, 30, 'DISCREPANCY_HOLD', 370.0, 472.5, 27.5,
    'inv_anl_9021', 1795074719192
);
INSERT OR REPLACE INTO discrepancy_reports (
    id, report_no, parent_code, vendor_name, invoice_no, inward_type,
    invoice_qty, actual_qty, variance_qty, variance_pct, financial_impact,
    status, remarks, created_at
) VALUES (
    'dr_261115-c', 'DR-2611-001', '261115-C', 'Anil International',
    'INV-ANL-9021', 'Fabric Knitted Roll', 500.0, 472.5, 27.5, 5.5,
    10175.0, 'PENDING_SUPPLIER_CREDIT_NOTE', 'Scale weighment short by 27.5 kgs across 3 rolls. Debit note recommended.', 1795074719192
);
INSERT OR REPLACE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, inward_weight_qty, measured_gsm, location_bin,
    qc_status, qr_payload, available_qty, rate, measured_dia, created_at
) VALUES (
    'bnd_261115-c-1_3', '261115-C-1/3', 'lot_261115_c', '261115-C', 1, 3,
    'fabric', '30s Sinker Crew Bio Wash', 'Royal Blue', 156.0, 180, 'BAY-INW-03',
    'Hold', '{"bundleCode": "261115-C-1/3", "parentCode": "261115-C", "vendor": "Anil International", "fabric": "30s Sinker Crew Bio Wash", "color": "Royal Blue", "weight": 156.0, "gsm": 180, "dia": 30, "bin": "BAY-INW-03"}', 156.0, 370.0, 30, 1795074719192
);
INSERT OR REPLACE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, inward_weight_qty, measured_gsm, location_bin,
    qc_status, qr_payload, available_qty, rate, measured_dia, created_at
) VALUES (
    'bnd_261115-c-2_3', '261115-C-2/3', 'lot_261115_c', '261115-C', 2, 3,
    'fabric', '30s Sinker Crew Bio Wash', 'Royal Blue', 158.5, 180, 'BAY-INW-03',
    'Hold', '{"bundleCode": "261115-C-2/3", "parentCode": "261115-C", "vendor": "Anil International", "fabric": "30s Sinker Crew Bio Wash", "color": "Royal Blue", "weight": 158.5, "gsm": 180, "dia": 30, "bin": "BAY-INW-03"}', 158.5, 370.0, 30, 1795074719192
);
INSERT OR REPLACE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, inward_weight_qty, measured_gsm, location_bin,
    qc_status, qr_payload, available_qty, rate, measured_dia, created_at
) VALUES (
    'bnd_261115-c-3_3', '261115-C-3/3', 'lot_261115_c', '261115-C', 3, 3,
    'fabric', '30s Sinker Crew Bio Wash', 'Royal Blue', 158.0, 180, 'BAY-INW-03',
    'Hold', '{"bundleCode": "261115-C-3/3", "parentCode": "261115-C", "vendor": "Anil International", "fabric": "30s Sinker Crew Bio Wash", "color": "Royal Blue", "weight": 158.0, "gsm": 180, "dia": 30, "bin": "BAY-INW-03"}', 158.0, 370.0, 30, 1795074719192
);
INSERT OR REPLACE INTO inward_invoices (
    id, invoice_no, vendor_name, invoice_date, invoice_qty, invoice_amount, uom, currency, remarks, created_at
) VALUES (
    'inv_amb_4412', 'INV-AMB-4412', 'Ambica Elastics', '2026-11-12',
    1200.0, 456000.0, 'Kgs', 'INR', 'Tax Invoice with GST e-Way bill compliance verified', 1795074719192
);
INSERT OR REPLACE INTO parent_inward_lots (
    id, parent_code, inward_date, vendor_id, vendor_name, invoice_no, inward_type,
    uom, invoice_qty, invoice_amount, physical_received_qty, shortfall_qty,
    discrepancy_amount, fabric_type, vendor_fab_name, fpcc_fab_name, vendor_color,
    fpcc_color, invoice_gsm, dia, stage, unit_rate, counted_qty, variance_qty,
    invoice_id, created_at
) VALUES (
    'lot_261112_d', '261112-D', '2026-11-12', 'ent_ven_ambica_elastics_6b8870c2',
    'Ambica Elastics', 'INV-AMB-4412', 'Fabric Knitted Roll',
    'Kgs', 1200.0, 456000.0, 1200.0, 0.0, 0.0,
    'Knitted', '3360 ettu pant', '3360 ettu pant', 'Anthracite', 'Anthracite',
    280, 32, 'ACCEPTED', 380.0, 1200.0, 0.0,
    'inv_amb_4412', 1795074719192
);
INSERT OR REPLACE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, inward_weight_qty, measured_gsm, location_bin,
    qc_status, qr_payload, available_qty, rate, measured_dia, created_at
) VALUES (
    'bnd_261112-d-1_6', '261112-D-1/6', 'lot_261112_d', '261112-D', 1, 6,
    'fabric', '3360 ettu pant', 'Anthracite', 202.0, 280, 'BIN-A1-04',
    'Passed', '{"bundleCode": "261112-D-1/6", "parentCode": "261112-D", "vendor": "Ambica Elastics", "fabric": "3360 ettu pant", "color": "Anthracite", "weight": 202.0, "gsm": 280, "dia": 32, "bin": "BIN-A1-04"}', 65.0, 380.0, 32, 1795074719192
);
INSERT OR REPLACE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, inward_weight_qty, measured_gsm, location_bin,
    qc_status, qr_payload, available_qty, rate, measured_dia, created_at
) VALUES (
    'bnd_261112-d-2_6', '261112-D-2/6', 'lot_261112_d', '261112-D', 2, 6,
    'fabric', '3360 ettu pant', 'Anthracite', 198.5, 280, 'BIN-A1-04',
    'Passed', '{"bundleCode": "261112-D-2/6", "parentCode": "261112-D", "vendor": "Ambica Elastics", "fabric": "3360 ettu pant", "color": "Anthracite", "weight": 198.5, "gsm": 280, "dia": 32, "bin": "BIN-A1-04"}', 64.1, 380.0, 32, 1795074719192
);
INSERT OR REPLACE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, inward_weight_qty, measured_gsm, location_bin,
    qc_status, qr_payload, available_qty, rate, measured_dia, created_at
) VALUES (
    'bnd_261112-d-3_6', '261112-D-3/6', 'lot_261112_d', '261112-D', 3, 6,
    'fabric', '3360 ettu pant', 'Anthracite', 201.0, 280, 'BIN-A1-05',
    'Passed', '{"bundleCode": "261112-D-3/6", "parentCode": "261112-D", "vendor": "Ambica Elastics", "fabric": "3360 ettu pant", "color": "Anthracite", "weight": 201.0, "gsm": 280, "dia": 32, "bin": "BIN-A1-05"}', 201.0, 380.0, 32, 1795074719192
);
INSERT OR REPLACE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, inward_weight_qty, measured_gsm, location_bin,
    qc_status, qr_payload, available_qty, rate, measured_dia, created_at
) VALUES (
    'bnd_261112-d-4_6', '261112-D-4/6', 'lot_261112_d', '261112-D', 4, 6,
    'fabric', '3360 ettu pant', 'Anthracite', 199.5, 280, 'BIN-A1-05',
    'Passed', '{"bundleCode": "261112-D-4/6", "parentCode": "261112-D", "vendor": "Ambica Elastics", "fabric": "3360 ettu pant", "color": "Anthracite", "weight": 199.5, "gsm": 280, "dia": 32, "bin": "BIN-A1-05"}', 199.5, 380.0, 32, 1795074719192
);
INSERT OR REPLACE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, inward_weight_qty, measured_gsm, location_bin,
    qc_status, qr_payload, available_qty, rate, measured_dia, created_at
) VALUES (
    'bnd_261112-d-5_6', '261112-D-5/6', 'lot_261112_d', '261112-D', 5, 6,
    'fabric', '3360 ettu pant', 'Anthracite', 200.0, 280, 'BIN-A1-06',
    'Passed', '{"bundleCode": "261112-D-5/6", "parentCode": "261112-D", "vendor": "Ambica Elastics", "fabric": "3360 ettu pant", "color": "Anthracite", "weight": 200.0, "gsm": 280, "dia": 32, "bin": "BIN-A1-06"}', 200.0, 380.0, 32, 1795074719192
);
INSERT OR REPLACE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, inward_weight_qty, measured_gsm, location_bin,
    qc_status, qr_payload, available_qty, rate, measured_dia, created_at
) VALUES (
    'bnd_261112-d-6_6', '261112-D-6/6', 'lot_261112_d', '261112-D', 6, 6,
    'fabric', '3360 ettu pant', 'Anthracite', 199.0, 280, 'BIN-A1-06',
    'Passed', '{"bundleCode": "261112-D-6/6", "parentCode": "261112-D", "vendor": "Ambica Elastics", "fabric": "3360 ettu pant", "color": "Anthracite", "weight": 199.0, "gsm": 280, "dia": 32, "bin": "BIN-A1-06"}', 199.0, 380.0, 32, 1795074719192
);
INSERT OR REPLACE INTO inward_invoices (
    id, invoice_no, vendor_name, invoice_date, invoice_qty, invoice_amount, uom, currency, remarks, created_at
) VALUES (
    'inv_ash_1109', 'INV-ASH-1109', 'Ashapura Industries', '2026-11-10',
    750.0, 262500.0, 'Kgs', 'INR', 'Tax Invoice with GST e-Way bill compliance verified', 1795074719192
);
INSERT OR REPLACE INTO parent_inward_lots (
    id, parent_code, inward_date, vendor_id, vendor_name, invoice_no, inward_type,
    uom, invoice_qty, invoice_amount, physical_received_qty, shortfall_qty,
    discrepancy_amount, fabric_type, vendor_fab_name, fpcc_fab_name, vendor_color,
    fpcc_color, invoice_gsm, dia, stage, unit_rate, counted_qty, variance_qty,
    invoice_id, created_at
) VALUES (
    'lot_261110_e', '261110-E', '2026-11-10', 'ent_ven_ashapura_industries_43e4ae85',
    'Ashapura Industries', 'INV-ASH-1109', 'Fabric Knitted Roll',
    'Kgs', 750.0, 262500.0, 750.0, 0.0, 0.0,
    'Knitted', '30s cotton slub', '30s cotton slub', 'Natural Ecru', 'Natural Ecru',
    450, 58, 'ACCEPTED', 350.0, 750.0, 0.0,
    'inv_ash_1109', 1795074719192
);
INSERT OR REPLACE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, inward_weight_qty, measured_gsm, location_bin,
    qc_status, qr_payload, available_qty, rate, measured_dia, created_at
) VALUES (
    'bnd_261110-e-1_4', '261110-E-1/4', 'lot_261110_e', '261110-E', 1, 4,
    'fabric', '30s cotton slub', 'Natural Ecru', 188.0, 450, 'BIN-C2-01',
    'Passed', '{"bundleCode": "261110-E-1/4", "parentCode": "261110-E", "vendor": "Ashapura Industries", "fabric": "30s cotton slub", "color": "Natural Ecru", "weight": 188.0, "gsm": 450, "dia": 58, "bin": "BIN-C2-01"}', 188.0, 350.0, 58, 1795074719192
);
INSERT OR REPLACE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, inward_weight_qty, measured_gsm, location_bin,
    qc_status, qr_payload, available_qty, rate, measured_dia, created_at
) VALUES (
    'bnd_261110-e-2_4', '261110-E-2/4', 'lot_261110_e', '261110-E', 2, 4,
    'fabric', '30s cotton slub', 'Natural Ecru', 187.5, 450, 'BIN-C2-01',
    'Passed', '{"bundleCode": "261110-E-2/4", "parentCode": "261110-E", "vendor": "Ashapura Industries", "fabric": "30s cotton slub", "color": "Natural Ecru", "weight": 187.5, "gsm": 450, "dia": 58, "bin": "BIN-C2-01"}', 187.5, 350.0, 58, 1795074719192
);
INSERT OR REPLACE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, inward_weight_qty, measured_gsm, location_bin,
    qc_status, qr_payload, available_qty, rate, measured_dia, created_at
) VALUES (
    'bnd_261110-e-3_4', '261110-E-3/4', 'lot_261110_e', '261110-E', 3, 4,
    'fabric', '30s cotton slub', 'Natural Ecru', 186.0, 450, 'BIN-C2-01',
    'Passed', '{"bundleCode": "261110-E-3/4", "parentCode": "261110-E", "vendor": "Ashapura Industries", "fabric": "30s cotton slub", "color": "Natural Ecru", "weight": 186.0, "gsm": 450, "dia": 58, "bin": "BIN-C2-01"}', 186.0, 350.0, 58, 1795074719192
);
INSERT OR REPLACE INTO bundles (
    id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
    item_type, item_name, color_name, inward_weight_qty, measured_gsm, location_bin,
    qc_status, qr_payload, available_qty, rate, measured_dia, created_at
) VALUES (
    'bnd_261110-e-4_4', '261110-E-4/4', 'lot_261110_e', '261110-E', 4, 4,
    'fabric', '30s cotton slub', 'Natural Ecru', 188.5, 450, 'BIN-C2-01',
    'Passed', '{"bundleCode": "261110-E-4/4", "parentCode": "261110-E", "vendor": "Ashapura Industries", "fabric": "30s cotton slub", "color": "Natural Ecru", "weight": 188.5, "gsm": 450, "dia": 58, "bin": "BIN-C2-01"}', 188.5, 350.0, 58, 1795074719192
);
-- 6. Order Prep & Planning (Snapshots, BOMs, Procurement Requisitions)
-- Order 11260001 (ATC Tires Polo + Tote)
INSERT OR REPLACE INTO order_planning_snapshots (
    id, order_id, snapshot_version, client_etd, planned_finish_date, slack_days, is_overdue_risk, critical_path_summary, created_at
) VALUES (
    'plan_11260001', 'ord_11260001', 1, '2026-12-05', '2026-11-28', 7, 0, 'Sufficient slack: fabric allocated from stock, cutting begins Nov 20', 1795074719192
);
INSERT OR REPLACE INTO bom_lines (
    id, order_id, order_item_id, line_type, item_name, sub_type, consumption_avg, wastage_pct, uom, unit_qty, required_qty, soh_qty, allocated_qty, shortfall_qty, rate_per_unit, status, created_at, updated_at
) VALUES 
('bom_11260001_1', 'ord_11260001', 'item_11260001_1', 'FABRIC', '2441 Galactic Polo', 'Knitted', 0.28, 3.0, 'Kgs', 400, 115.36, 850.0, 115.36, 0.0, 370.0, 'ACTIVE', 1795074719192, 1795074719192),
('bom_11260001_2', 'ord_11260001', 'item_11260001_1', 'ACCESSORY', 'FPCC Signature Laser Button', 'Placket', 3.0, 5.0, 'Pcs', 400, 1260.0, 5000.0, 1260.0, 0.0, 2.5, 'ACTIVE', 1795074719192, 1795074719192),
('bom_11260001_3', 'ord_11260001', 'item_11260001_1', 'LABEL', 'Brand Label - ETTU', 'Neck', 1.0, 2.0, 'Pcs', 400, 408.0, 2000.0, 408.0, 0.0, 4.0, 'ACTIVE', 1795074719192, 1795074719192),
('bom_11260001_4', 'ord_11260001', 'item_11260001_2', 'FABRIC', '30s cotton slub', 'Woven Canvas', 0.35, 4.0, 'Kgs', 400, 145.6, 750.0, 145.6, 0.0, 350.0, 'ACTIVE', 1795074719192, 1795074719192);

-- Order 11260002 (Amara Raja Tech Hoodie) - Has Shortfall & Raised POs
INSERT OR REPLACE INTO order_planning_snapshots (
    id, order_id, snapshot_version, client_etd, planned_finish_date, slack_days, is_overdue_risk, critical_path_summary, created_at
) VALUES (
    'plan_11260002', 'ord_11260002', 1, '2026-12-15', '2026-12-10', 5, 0, 'Procurement requisitions raised for fleece & brass zips. Critical path slack: 5 days', 1795074719192
);
INSERT OR REPLACE INTO bom_lines (
    id, order_id, order_item_id, line_type, item_name, sub_type, consumption_avg, wastage_pct, uom, unit_qty, required_qty, soh_qty, allocated_qty, shortfall_qty, rate_per_unit, status, created_at, updated_at
) VALUES 
('bom_11260002_1', 'ord_11260002', 'item_11260002_1', 'FABRIC', '2111 bonded box fleece', 'Fleece', 0.65, 4.0, 'Kgs', 250, 169.0, 0.0, 0.0, 169.0, 400.0, 'ACTIVE', 1795074719192, 1795074719192),
('bom_11260002_2', 'ord_11260002', 'item_11260002_1', 'ACCESSORY', 'Brass', 'Full Front Zip', 1.0, 2.0, 'Pcs', 250, 255.0, 0.0, 0.0, 255.0, 45.0, 'ACTIVE', 1795074719192, 1795074719192);

INSERT OR REPLACE INTO procurement_requisitions (
    id, order_id, bom_line_id, material_name, material_type, uom, required_qty, order_buffer_qty, total_po_qty, vendor_name, estimated_rate, estimated_amount, required_by_date, po_no, status, created_at, updated_at
) VALUES 
('pr_11260002_1', 'ord_11260002', 'bom_11260002_1', '2111 bonded box fleece', 'FABRIC', 'Kgs', 169.0, 11.0, 180.0, 'Anandghan Technologies', 400.0, 72000.0, '2026-11-25', 'PO-TEX-2611-04', 'PO_ISSUED', 1795074719192, 1795074719192),
('pr_11260002_2', 'ord_11260002', 'bom_11260002_2', 'Brass', 'ACCESSORY', 'Pcs', 255.0, 20.0, 275.0, 'Arihant Plastics', 45.0, 12375.0, '2026-11-26', 'PO-TRM-2611-12', 'APPROVED', 1795074719192, 1795074719192);

-- Order 11260003 (ACMA EV Crew Neck Tee) - In Cutting
INSERT OR REPLACE INTO order_planning_snapshots (
    id, order_id, snapshot_version, client_etd, planned_finish_date, slack_days, is_overdue_risk, critical_path_summary, created_at
) VALUES (
    'plan_11260003', 'ord_11260003', 1, '2026-11-28', '2026-11-22', 6, 0, 'Materials issued 100% to cutting floor. Lay 1 completed.', 1795074719192
);
INSERT OR REPLACE INTO bom_lines (
    id, order_id, order_item_id, line_type, item_name, sub_type, consumption_avg, wastage_pct, uom, unit_qty, required_qty, soh_qty, allocated_qty, shortfall_qty, rate_per_unit, status, created_at, updated_at
) VALUES 
('bom_11260003_1', 'ord_11260003', 'item_11260003_1', 'FABRIC', '30s Sinker Crew Bio Wash', 'Single Jersey', 0.22, 3.0, 'Kgs', 600, 135.96, 472.5, 135.96, 0.0, 370.0, 'ACTIVE', 1795074719192, 1795074719192),
('bom_11260003_2', 'ord_11260003', 'item_11260003_1', 'ACCESSORY', '30s Sinker Crew Bio Wash -Rib', 'Neck Rib', 0.04, 5.0, 'Kgs', 600, 25.2, 80.0, 25.2, 0.0, 420.0, 'ACTIVE', 1795074719192, 1795074719192);
-- 7. Floor Cutting Station (Jobs, Lays, Rolls, Bundles, Remnants)
INSERT OR REPLACE INTO cutting_jobs (
    id, order_id, order_item_id, production_order_no, product_name, color_name, fabric_name, target_qty, pharma_no, planned_bom_avg, status, operator_name, created_at, updated_at
) VALUES 
('CJ-11260003-1', 'ord_11260003', 'item_11260003_1', '11260003', 'EV Mobility Crew Neck Bio-Wash Tee', 'Royal Blue', '30s Sinker Crew Bio Wash', 600, 'PH-TEE-REG-01', 0.22, 'BUNDLED', 'Manoj Sharma (Lead Cutter)', 1795074719192, 1795074719192),
('CJ-11260004-1', 'ord_11260004', 'item_11260004_1', '11260004', 'Water-Resistant All-Weather Windbreaker', 'Olive Green', '3347 shark plain', 180, 'PH-JKT-101R', 0.48, 'OPEN', 'Ramesh Yadav', 1795074719192, 1795074719192);

INSERT OR REPLACE INTO cutting_roll_issues (
    id, cutting_job_id, roll_code, fabric_name, color_name, issued_weight_kg, status, operator_name, issued_at
) VALUES 
('cri_11260003_1', 'CJ-11260003-1', '261112-D-1/6', '3360 ettu pant', 'Anthracite', 68.5, 'CONSUMED', 'Manoj Sharma', 1795074719192),
('cri_11260003_2', 'CJ-11260003-1', '261112-D-2/6', '3360 ettu pant', 'Anthracite', 67.2, 'CONSUMED', 'Manoj Sharma', 1795074719192);

INSERT OR REPLACE INTO cut_lays (
    id, cutting_job_id, lay_number, roll_code, fabric_name, roll_weight_kg, marker_name, ratio_json, plies, total_pieces_in_lay, lay_weight_kg, actual_lay_avg_kg, planned_bom_avg_kg, efficiency_pct, status, operator_name, created_at
) VALUES (
    'CL-CJ-11260003-1-1', 'CJ-11260003-1', 1, '261112-D-1/6', '30s Sinker Crew Bio Wash', 68.5, 'MKR-TEE-ROYAL-01', '{"S": 1, "M": 2, "L": 2, "XL": 1}', 50, 300, 66.2, 0.2207, 0.22, 99.7, 'CUT', 'Manoj Sharma', 1795074719192
);

INSERT OR REPLACE INTO cut_bundles (
    id, cutting_job_id, cut_lay_id, bundle_sequence, size, pieces, slip_printed_count, printed_at, next_process, operator_name, created_at
) VALUES 
('CB-261118-01', 'CJ-11260003-1', 'CL-CJ-11260003-1-1', 1, 'S', 50, 1, 1795074719192, 'Branding', 'Manoj Sharma', 1795074719192),
('CB-261118-02', 'CJ-11260003-1', 'CL-CJ-11260003-1-1', 2, 'M', 50, 1, 1795074719192, 'Branding', 'Manoj Sharma', 1795074719192),
('CB-261118-03', 'CJ-11260003-1', 'CL-CJ-11260003-1-1', 3, 'M', 50, 1, 1795074719192, 'Branding', 'Manoj Sharma', 1795074719192),
('CB-261118-04', 'CJ-11260003-1', 'CL-CJ-11260003-1-1', 4, 'L', 50, 1, 1795074719192, 'Branding', 'Manoj Sharma', 1795074719192),
('CB-261118-05', 'CJ-11260003-1', 'CL-CJ-11260003-1-1', 5, 'L', 50, 1, 1795074719192, 'Branding', 'Manoj Sharma', 1795074719192),
('CB-261118-06', 'CJ-11260003-1', 'CL-CJ-11260003-1-1', 6, 'XL', 50, 1, 1795074719192, 'Branding', 'Manoj Sharma', 1795074719192);

INSERT OR REPLACE INTO remnant_returns (
    id, prefix, source_roll_code, cutting_job_id, cut_lay_id, weight_kg, bin_location, defect_reason, operator_name, created_at
) VALUES 
('Bit261112-D-1', 'Bit', '261112-D-1/6', 'CJ-11260003-1', 'CL-CJ-11260003-1-1', 2.3, '4.Cutting Room', 'End of roll usable remnant returned to fabric stock', 'Manoj Sharma', 1795074719192),
('Rej261112-D-2', 'Rej', '261112-D-2/6', 'CJ-11260003-1', 'CL-CJ-11260003-1-1', 0.8, '4.Cutting Room', 'Slub defect at roll start trimmed out', 'Manoj Sharma', 1795074719192);
-- 8. Unified Graph Nodes & Bi-temporal Edges
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('ord_11260001', 'ORDER', '{"production_order_no": "11260001", "client_name": "ATC Tires Private Limited", "total_qty": 800, "status": "BOOKED"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_ord_11260001_cli', 'ord_11260001', 'ent_cli_atc_tires_private_limited_9768afc0', 'ORDERED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('ord_11260002', 'ORDER', '{"production_order_no": "11260002", "client_name": "Amara Raja Energy & Mobility Limited", "total_qty": 250, "status": "PLANNED"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_ord_11260002_cli', 'ord_11260002', 'ent_cli_amara_raja_energy_mobility_limited_95ac4e91', 'ORDERED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('ord_11260003', 'ORDER', '{"production_order_no": "11260003", "client_name": "Acma Mobility Foundation", "total_qty": 600, "status": "IN_CUTTING"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_ord_11260003_cli', 'ord_11260003', 'ent_cli_acma_mobility_foundation_4d29096f', 'ORDERED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('ord_11260004', 'ORDER', '{"production_order_no": "11260004", "client_name": "Apex Trading", "total_qty": 180, "status": "IN_PRODUCTION"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_ord_11260004_cli', 'ord_11260004', 'ent_cli_apex_trading_d985c8b1', 'ORDERED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('ord_11260005', 'ORDER', '{"production_order_no": "11260005", "client_name": "Bapl Rototech Private Limited", "total_qty": 500, "status": "BOOKED"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_ord_11260005_cli', 'ord_11260005', 'ent_cli_bapl_rototech_private_limited_6d9806e8', 'ORDERED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('ord_11260006', 'ORDER', '{"production_order_no": "11260006", "client_name": "Altamar Global Exports", "total_qty": 350, "status": "COMPLETED"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_ord_11260006_cli', 'ord_11260006', 'ent_cli_altamar_global_exports_8462c771', 'ORDERED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('smp_polo_001', 'SAMPLE_DEVELOPMENT', '{"development_no": "D26110001", "genre": "Polo", "status": "APPROVED"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_smp_polo_001_cli', 'smp_polo_001', 'ent_cli_atc_tires_private_limited_9768afc0', 'REQUESTED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('smp_hoodie_002', 'SAMPLE_DEVELOPMENT', '{"development_no": "D26110002", "genre": "Hoodies and Pullovers", "status": "SUBMITTED"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_smp_hoodie_002_cli', 'smp_hoodie_002', 'ent_cli_amara_raja_energy_mobility_limited_95ac4e91', 'REQUESTED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('smp_tee_003', 'SAMPLE_DEVELOPMENT', '{"development_no": "D26110003", "genre": "Tee", "status": "IN_PROGRESS"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_smp_tee_003_cli', 'smp_tee_003', 'ent_cli_acma_mobility_foundation_4d29096f', 'REQUESTED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('smp_jacket_004', 'SAMPLE_DEVELOPMENT', '{"development_no": "D26110004", "genre": "Jackets", "status": "REWORK_REQUESTED"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_smp_jacket_004_cli', 'smp_jacket_004', 'ent_cli_apex_trading_d985c8b1', 'REQUESTED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('smp_bag_005', 'SAMPLE_DEVELOPMENT', '{"development_no": "D26110005", "genre": "Bags", "status": "APPROVED"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_smp_bag_005_cli', 'smp_bag_005', 'ent_cli_bapl_rototech_private_limited_6d9806e8', 'REQUESTED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('smp_bottle_006', 'SAMPLE_DEVELOPMENT', '{"development_no": "D26110006", "genre": "Drinkwear", "status": "DRAFT"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_smp_bottle_006_cli', 'smp_bottle_006', 'ent_cli_aapka_khuranas_india_llp_0e766966', 'REQUESTED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('smp_bottoms_007', 'SAMPLE_DEVELOPMENT', '{"development_no": "D26110007", "genre": "Sweat Pants", "status": "APPROVED"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_smp_bottoms_007_cli', 'smp_bottoms_007', 'ent_cli_altamar_global_exports_8462c771', 'REQUESTED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('smp_performance_008', 'SAMPLE_DEVELOPMENT', '{"development_no": "D26110008", "genre": "Tee", "status": "REJECTED"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_smp_performance_008_cli', 'smp_performance_008', 'ent_cli_atc_tires_private_limited_9768afc0', 'REQUESTED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('prod_atc_polo_001', 'PRODUCT_DEFINITION', '{"product_name": "High-Density Pique Corporate Polo", "sku_code": "ET_PO_HDPIQ", "genre": "Polo", "mrp": 1499.0}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_prod_atc_polo_001_smp', 'prod_atc_polo_001', 'smp_polo_001', 'DERIVED_FROM_SAMPLE', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('prod_amara_hoodie_002', 'PRODUCT_DEFINITION', '{"product_name": "Heavyweight Fleece Tech Zip Hoodie", "sku_code": "AA_HO_TECHZ", "genre": "Hoodies and Pullovers", "mrp": 2499.0}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_prod_amara_hoodie_002_smp', 'prod_amara_hoodie_002', 'smp_hoodie_002', 'DERIVED_FROM_SAMPLE', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('prod_acma_tee_003', 'PRODUCT_DEFINITION', '{"product_name": "EV Mobility Crew Neck Bio-Wash Tee", "sku_code": "CI_TE_EVCREW", "genre": "Tee", "mrp": 799.0}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_prod_acma_tee_003_smp', 'prod_acma_tee_003', 'smp_tee_003', 'DERIVED_FROM_SAMPLE', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('prod_apex_windbreaker_004', 'PRODUCT_DEFINITION', '{"product_name": "Water-Resistant All-Weather Windbreaker", "sku_code": "JP_JK_ALLWTH", "genre": "Jackets", "mrp": 3299.0}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_prod_apex_windbreaker_004_smp', 'prod_apex_windbreaker_004', 'smp_jacket_004', 'DERIVED_FROM_SAMPLE', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('prod_bapl_tote_005', 'PRODUCT_DEFINITION', '{"product_name": "Heavy-Duty Industrial Canvas Tote Bag", "sku_code": "ET_BG_CANVAS", "genre": "Bags", "mrp": 899.0}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_prod_bapl_tote_005_smp', 'prod_bapl_tote_005', 'smp_bag_005', 'DERIVED_FROM_SAMPLE', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('prod_jogger_006', 'PRODUCT_DEFINITION', '{"product_name": "Pro-Athletic French Terry Tapered Jogger", "sku_code": "AA_SP_TERRY", "genre": "Sweat Pants", "mrp": 1899.0}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_prod_jogger_006_smp', 'prod_jogger_006', 'smp_bottoms_007', 'DERIVED_FROM_SAMPLE', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('lot_261117_a', 'PARENT_LOT', '{"parent_code": "261117-A", "vendor": "Arkap Knits Pvt Ltd / Rr Knits", "stage": "AWAITING_COUNT", "fabric": "2441 Galactic Polo"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_lot_261117_a_ven', 'lot_261117_a', 'ent_ven_arkap_knits_pvt_ltd_rr_knits_1f90c965', 'SUPPLIED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('lot_261116_b', 'PARENT_LOT', '{"parent_code": "261116-B", "vendor": "Anandghan Technologies", "stage": "AWAITING_QC", "fabric": "2111 bonded box fleece"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_lot_261116_b_ven', 'lot_261116_b', 'ent_ven_anandghan_technologies_f6aa1f26', 'SUPPLIED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('lot_261115_c', 'PARENT_LOT', '{"parent_code": "261115-C", "vendor": "Anil International", "stage": "DISCREPANCY_HOLD", "fabric": "30s Sinker Crew Bio Wash"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_lot_261115_c_ven', 'lot_261115_c', 'ent_ven_anil_international_289e7a92', 'SUPPLIED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('lot_261112_d', 'PARENT_LOT', '{"parent_code": "261112-D", "vendor": "Ambica Elastics", "stage": "ACCEPTED", "fabric": "3360 ettu pant"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_lot_261112_d_ven', 'lot_261112_d', 'ent_ven_ambica_elastics_6b8870c2', 'SUPPLIED_BY', '{}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_nodes (id, node_type, props_json, valid_from, txn_from) VALUES ('lot_261110_e', 'PARENT_LOT', '{"parent_code": "261110-E", "vendor": "Ashapura Industries", "stage": "ACCEPTED", "fabric": "30s cotton slub"}', 1795074719192, 1795074719192);
INSERT OR REPLACE INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, txn_from) VALUES ('edge_lot_261110_e_ven', 'lot_261110_e', 'ent_ven_ashapura_industries_43e4ae85', 'SUPPLIED_BY', '{}', 1795074719192, 1795074719192);
