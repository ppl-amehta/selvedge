-- Migration 011: Enrich Gender Sizes, Compositions, and Dropdown Relationships from Excel Ground Truth

-- 1. Remove corrupted Excel date serials and float-suffixed sizes
DELETE FROM master_items WHERE category = 'size' AND code IN (
    '46087.0', '46182.0', '46277.0', '46371.0',
    '32.0', '34.0', '36.0', '38.0', '39.0', '40.0', '42.0', '44.0', '46.0', '48.0', '50.0', '52.0'
);

-- 2. Clean Seed / Replace Sizes with Explicit Gender Scale in meta_json
-- A. Adult Apparel Sizes (Men, Women, Unisex)
INSERT OR REPLACE INTO master_items (id, category, code, display_name, meta_json, created_at) VALUES
('mi_sz_xxs', 'size', 'XXS', 'XXS', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"alpha"}', 1794916474761),
('mi_sz_xs', 'size', 'XS', 'XS', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"alpha"}', 1794916474761),
('mi_sz_s', 'size', 'S', 'S', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"alpha"}', 1794916474761),
('mi_sz_m', 'size', 'M', 'M', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"alpha"}', 1794916474761),
('mi_sz_l', 'size', 'L', 'L', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"alpha"}', 1794916474761),
('mi_sz_xl', 'size', 'XL', 'XL', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"alpha"}', 1794916474761),
('mi_sz_2xl', 'size', '2XL', '2XL', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"alpha"}', 1794916474761),
('mi_sz_3xl', 'size', '3XL', '3XL', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"alpha"}', 1794916474761),
('mi_sz_4xl', 'size', '4XL', '4XL', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"alpha"}', 1794916474761),
('mi_sz_5xl', 'size', '5XL', '5XL', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"alpha"}', 1794916474761),
('mi_sz_custom', 'size', 'Custom', 'Custom', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"alpha"}', 1794916474761),
('mi_sz_num_32', 'size', '32', '32', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"numeric"}', 1794916474761),
('mi_sz_num_34', 'size', '34', '34', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"numeric"}', 1794916474761),
('mi_sz_num_36', 'size', '36', '36', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"numeric"}', 1794916474761),
('mi_sz_num_38', 'size', '38', '38', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"numeric"}', 1794916474761),
('mi_sz_num_39', 'size', '39', '39', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"numeric"}', 1794916474761),
('mi_sz_num_40', 'size', '40', '40', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"numeric"}', 1794916474761),
('mi_sz_num_42', 'size', '42', '42', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"numeric"}', 1794916474761),
('mi_sz_num_44', 'size', '44', '44', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"numeric"}', 1794916474761),
('mi_sz_num_46', 'size', '46', '46', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"numeric"}', 1794916474761),
('mi_sz_num_48', 'size', '48', '48', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"numeric"}', 1794916474761),
('mi_sz_num_50', 'size', '50', '50', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"numeric"}', 1794916474761),
('mi_sz_num_52', 'size', '52', '52', '{"scale":"adult","genders":["Men","Women","Unisex"],"type":"numeric"}', 1794916474761);

-- B. Infant Sizes (Months)
INSERT OR REPLACE INTO master_items (id, category, code, display_name, meta_json, created_at) VALUES
('mi_sz_inf_0_3', 'size', '0-3', '0-3', '{"scale":"infant","genders":["Infant"],"unit":"months"}', 1794916474761),
('mi_sz_inf_3_6', 'size', '3-6', '3-6', '{"scale":"infant","genders":["Infant"],"unit":"months"}', 1794916474761),
('mi_sz_inf_6_9', 'size', '6-9', '6-9', '{"scale":"infant","genders":["Infant"],"unit":"months"}', 1794916474761),
('mi_sz_inf_9_12', 'size', '9-12', '9-12', '{"scale":"infant","genders":["Infant"],"unit":"months"}', 1794916474761),
('mi_sz_inf_12_15', 'size', '12-15', '12-15', '{"scale":"infant","genders":["Infant"],"unit":"months"}', 1794916474761),
('mi_sz_inf_15_18', 'size', '15-18', '15-18', '{"scale":"infant","genders":["Infant"],"unit":"months"}', 1794916474761);

-- C. Kids Sizes
INSERT OR REPLACE INTO master_items (id, category, code, display_name, meta_json, created_at) VALUES
('mi_sz_k16', 'size', 'K16', 'K16', '{"scale":"kids","genders":["Kids"]}', 1794916474761),
('mi_sz_k18', 'size', 'K18', 'K18', '{"scale":"kids","genders":["Kids"]}', 1794916474761),
('mi_sz_k20', 'size', 'K20', 'K20', '{"scale":"kids","genders":["Kids"]}', 1794916474761),
('mi_sz_k22', 'size', 'K22', 'K22', '{"scale":"kids","genders":["Kids"]}', 1794916474761),
('mi_sz_k24', 'size', 'K24', 'K24', '{"scale":"kids","genders":["Kids"]}', 1794916474761),
('mi_sz_k26', 'size', 'K26', 'K26', '{"scale":"kids","genders":["Kids"]}', 1794916474761),
('mi_sz_k28', 'size', 'K28', 'K28', '{"scale":"kids","genders":["Kids"]}', 1794916474761),
('mi_sz_k30', 'size', 'K30', 'K30', '{"scale":"kids","genders":["Kids"]}', 1794916474761),
('mi_sz_k32', 'size', 'K32', 'K32', '{"scale":"kids","genders":["Kids"]}', 1794916474761),
('mi_sz_k34', 'size', 'K34', 'K34', '{"scale":"kids","genders":["Kids"]}', 1794916474761),
('mi_sz_k36', 'size', 'K36', 'K36', '{"scale":"kids","genders":["Kids"]}', 1794916474761),
('mi_sz_kids_custom', 'size', 'kids custom', 'kids custom', '{"scale":"kids","genders":["Kids"]}', 1794916474761);

-- D. Junior Sizes
INSERT OR REPLACE INTO master_items (id, category, code, display_name, meta_json, created_at) VALUES
('mi_sz_jr_xs', 'size', 'Jr-XS', 'Jr-XS', '{"scale":"junior","genders":["Jr"]}', 1794916474761),
('mi_sz_jr_s', 'size', 'Jr-S', 'Jr-S', '{"scale":"junior","genders":["Jr"]}', 1794916474761),
('mi_sz_jr_m', 'size', 'Jr-M', 'Jr-M', '{"scale":"junior","genders":["Jr"]}', 1794916474761),
('mi_sz_jr_l', 'size', 'Jr-L', 'Jr-L', '{"scale":"junior","genders":["Jr"]}', 1794916474761),
('mi_sz_jr_custom', 'size', 'Jr Custom', 'Jr Custom', '{"scale":"junior","genders":["Jr"]}', 1794916474761);

-- E. Non-Apparel Sizes
INSERT OR REPLACE INTO master_items (id, category, code, display_name, meta_json, created_at) VALUES
('mi_sz_100ml', 'size', '100ml', '100ml', '{"category":"Drinkwear"}', 1794916474761),
('mi_sz_150ml', 'size', '150ml', '150ml', '{"category":"Drinkwear"}', 1794916474761),
('mi_sz_200ml', 'size', '200ml', '200ml', '{"category":"Drinkwear"}', 1794916474761),
('mi_sz_250ml', 'size', '250ml', '250ml', '{"category":"Drinkwear"}', 1794916474761),
('mi_sz_300ml', 'size', '300ml', '300ml', '{"category":"Drinkwear"}', 1794916474761),
('mi_sz_500ml', 'size', '500ml', '500ml', '{"category":"Drinkwear"}', 1794916474761),
('mi_sz_750ml', 'size', '750ml', '750ml', '{"category":"Drinkwear"}', 1794916474761),
('mi_sz_1litre', 'size', '1Litre', '1Litre', '{"category":"Drinkwear"}', 1794916474761),
('mi_sz_1_5_litre', 'size', '1.5 Litre', '1.5 Litre', '{"category":"Drinkwear"}', 1794916474761),
('mi_sz_2_litre', 'size', '2 Litre', '2 Litre', '{"category":"Drinkwear"}', 1794916474761),
('mi_sz_head_std', 'size', 'Std', 'Std', '{"category":"Headgear"}', 1794916474761),
('mi_sz_head_kids', 'size', 'Kids', 'Kids', '{"category":"Headgear"}', 1794916474761),
('mi_sz_head_large', 'size', 'Large', 'Large', '{"category":"Headgear"}', 1794916474761),
('mi_sz_head_small', 'size', 'Small', 'Small', '{"category":"Headgear"}', 1794916474761),
('mi_sz_stat_a4', 'size', 'A4', 'A4', '{"category":"Stationary"}', 1794916474761),
('mi_sz_stat_a3', 'size', 'A3', 'A3', '{"category":"Stationary"}', 1794916474761),
('mi_sz_stat_a5', 'size', 'A5', 'A5', '{"category":"Stationary"}', 1794916474761),
('mi_sz_stat_a6', 'size', 'A6', 'A6', '{"category":"Stationary"}', 1794916474761),
('mi_sz_stat_b5', 'size', 'B5', 'B5', '{"category":"Stationary"}', 1794916474761),
('mi_sz_std_free', 'size', 'Standard / Free', 'Standard / Free', '{"category":"General"}', 1794916474761);

-- 3. Seed 22 Authoritative Compositions from Sheet 2 Column O
INSERT OR REPLACE INTO master_items (id, category, code, display_name, meta_json, created_at) VALUES
('mi_cmp_95poly_5sp', 'composition', '95%Poly 5%SP', '95%Poly 5%SP', '{}', 1794916474761),
('mi_cmp_40nylon_38poly_12sp', 'composition', '40%Nylon 38%Poly 12%SP', '40%Nylon 38%Poly 12%SP', '{}', 1794916474761),
('mi_cmp_92poly_8sp', 'composition', '92%Poly 8%SP', '92%Poly 8%SP', '{}', 1794916474761),
('mi_cmp_100poly', 'composition', '100%Poly', '100%Poly', '{}', 1794916474761),
('mi_cmp_88poly_12sp', 'composition', '88%Poly 12%SP', '88%Poly 12%SP', '{}', 1794916474761),
('mi_cmp_60cot_30tencel_10sp', 'composition', '60%Cot 30%Tencel 10%SP', '60%Cot 30%Tencel 10%SP', '{}', 1794916474761),
('mi_cmp_90poly_10sp', 'composition', '90%Poly 10%SP', '90%Poly 10%SP', '{}', 1794916474761),
('mi_cmp_100cot', 'composition', '100%Cot', '100%Cot', '{}', 1794916474761),
('mi_cmp_90nylon_10sp', 'composition', '90%Nylon 10%SP', '90%Nylon 10%SP', '{}', 1794916474761),
('mi_cmp_58rcot_38rpoly_4sp', 'composition', '58%R.Cot 38%R.Poly 4%SP', '58%R.Cot 38%R.Poly 4%SP', '{}', 1794916474761),
('mi_cmp_65rcot_35rpoly', 'composition', '65%R.Cot 35%R.Poly', '65%R.Cot 35%R.Poly', '{}', 1794916474761),
('mi_cmp_85poly_15sp', 'composition', '85%Poly 15%SP', '85%Poly 15%SP', '{}', 1794916474761),
('mi_cmp_52cot_48poly', 'composition', '52%Cot 48%Poly', '52%Cot 48%Poly', '{}', 1794916474761),
('mi_cmp_60modal_30tencel_10sp', 'composition', '60%Modal 30%Tencel 10%SP', '60%Modal 30%Tencel 10%SP', '{}', 1794916474761),
('mi_cmp_95cot_5sp', 'composition', '95%Cot 5%SP', '95%Cot 5%SP', '{}', 1794916474761),
('mi_cmp_60cot_40poly', 'composition', '60%Cot 40%Poly', '60%Cot 40%Poly', '{}', 1794916474761),
('mi_cmp_65poly_35cot', 'composition', '65%Poly 35%Cot', '65%Poly 35%Cot', '{}', 1794916474761),
('mi_cmp_50acot_45siro_5sp', 'composition', '50%A Cot 45% Siro Viscose 5%SP', '50%A Cot 45% Siro Viscose 5%SP', '{}', 1794916474761),
('mi_cmp_100rpoly', 'composition', '100%R.Poly', '100%R.Poly', '{}', 1794916474761),
('mi_cmp_50cot_45poly_5sp', 'composition', '50%Cot 45%Poly 5%SP', '50%Cot 45%Poly 5%SP', '{}', 1794916474761),
('mi_cmp_70nylon_30sp', 'composition', '70%Nylon 30%SP', '70%Nylon 30%SP', '{}', 1794916474761),
('mi_cmp_43cot_27visc_30poly', 'composition', '43%Cot 27%Viscose 30%Poly', '43%Cot 27%Viscose 30%Poly', '{}', 1794916474761);

-- 4. Seed Product Genres from Sheet 2
INSERT OR REPLACE INTO master_items (id, category, code, display_name, meta_json, created_at) VALUES
-- Apparel Genres
('mi_gen_polo', 'product_genre', 'Polo', 'Polo', '{"category":"Apparel"}', 1794916474761),
('mi_gen_tee', 'product_genre', 'Tee', 'Tee', '{"category":"Apparel"}', 1794916474761),
('mi_gen_hoodies', 'product_genre', 'Hoodies and Pullovers', 'Hoodies and Pullovers', '{"category":"Apparel"}', 1794916474761),
('mi_gen_jackets', 'product_genre', 'Jackets', 'Jackets', '{"category":"Apparel"}', 1794916474761),
('mi_gen_bottoms', 'product_genre', 'Bottoms', 'Bottoms', '{"category":"Apparel"}', 1794916474761),
('mi_gen_qzip', 'product_genre', 'Quarter zip', 'Quarter zip', '{"category":"Apparel"}', 1794916474761),
('mi_gen_shirt', 'product_genre', 'Shirt', 'Shirt', '{"category":"Apparel"}', 1794916474761),
('mi_gen_shorts', 'product_genre', 'Shorts', 'Shorts', '{"category":"Apparel"}', 1794916474761),
('mi_gen_sweatpants', 'product_genre', 'Sweat Pants', 'Sweat Pants', '{"category":"Apparel"}', 1794916474761),
('mi_gen_sweatshirt', 'product_genre', 'Sweat shirt', 'Sweat shirt', '{"category":"Apparel"}', 1794916474761),
('mi_gen_sweater', 'product_genre', 'Sweater', 'Sweater', '{"category":"Apparel"}', 1794916474761),
('mi_gen_tights', 'product_genre', 'Tights', 'Tights', '{"category":"Apparel"}', 1794916474761),
('mi_gen_vest', 'product_genre', 'Vest', 'Vest', '{"category":"Apparel"}', 1794916474761),
('mi_gen_dress', 'product_genre', 'Dress', 'Dress', '{"category":"Apparel"}', 1794916474761),
('mi_gen_skorts', 'product_genre', 'Skorts', 'Skorts', '{"category":"Apparel"}', 1794916474761),
('mi_gen_onsies', 'product_genre', 'Onsies', 'Onsies', '{"category":"Apparel"}', 1794916474761),
('mi_gen_caps', 'product_genre', 'Caps', 'Caps', '{"category":"Apparel"}', 1794916474761),
('mi_gen_traded', 'product_genre', 'Traded apparel', 'Traded apparel', '{"category":"Apparel"}', 1794916474761),
-- Non-Apparel Genres
('mi_gen_drinkwear', 'product_genre', 'Drinkwear', 'Drinkwear', '{"category":"Non Apparel"}', 1794916474761),
('mi_gen_headgear', 'product_genre', 'Headgear', 'Headgear', '{"category":"Non Apparel"}', 1794916474761),
('mi_gen_bags', 'product_genre', 'Bags', 'Bags', '{"category":"Non Apparel"}', 1794916474761),
('mi_gen_stationary', 'product_genre', 'Stationary', 'Stationary', '{"category":"Non Apparel"}', 1794916474761),
('mi_gen_electronics', 'product_genre', 'Electronics', 'Electronics', '{"category":"Non Apparel"}', 1794916474761),
('mi_gen_others', 'product_genre', 'Others', 'Others', '{"category":"Non Apparel"}', 1794916474761);
