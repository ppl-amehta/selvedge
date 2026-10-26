-- ========================================================
-- FPCC SELVEDGE: MULTI-PRODUCT ORDERS & ADVANCED EMBELLISHMENTS
-- Bi-temporal Graph Projections & Master-Detail Entities
-- ========================================================

-- 1. Product Line Items per Order
CREATE TABLE IF NOT EXISTS order_items (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    item_sequence INTEGER NOT NULL,
    category TEXT NOT NULL,                   -- 'Apparel' or 'Non-Apparel'
    sub_category TEXT NOT NULL,               -- e.g. 'Polo', 'Hoodie', 'Drinkware', 'Bags'
    product_name TEXT NOT NULL,
    color_name TEXT NOT NULL,
    fabric_name TEXT,
    unit_price REAL NOT NULL DEFAULT 0.0,
    total_qty INTEGER NOT NULL DEFAULT 0,
    client_etd TEXT,
    client_remarks TEXT,
    product_image_url TEXT,
    main_label TEXT,
    other_label TEXT,
    cust_packaging TEXT,
    matrix_json JSON,                         -- Full Gender-Size-Qty Matrix breakdown
    created_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);

-- 2. Embellishments / Brandings per Product Line Item
CREATE TABLE IF NOT EXISTS order_item_brandings (
    id TEXT PRIMARY KEY,
    order_item_id TEXT NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
    order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    branding_type TEXT NOT NULL,              -- Strict enum: 'emb', 'print', 'dtf', 'unik', 'other'
    position TEXT,                            -- e.g. 'Left Chest', 'Center Chest', 'Back Nape'
    size TEXT,                                -- e.g. '3.5 x 2.0 inches'
    placement TEXT,                           -- e.g. 'Standard 7cm down from collar'
    color TEXT,                               -- e.g. 'Gold Metallic', 'White Pantone 000C'
    comment TEXT,                             -- Merchandiser notes
    branding_file_url TEXT,                   -- Uploaded design vector / artwork URL
    created_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_brandings_item ON order_item_brandings(order_item_id);
CREATE INDEX IF NOT EXISTS idx_brandings_order ON order_item_brandings(order_id);

-- 3. Shipping Destinations Linked to Order Items
CREATE TABLE IF NOT EXISTS order_shipping_destinations (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    order_item_id TEXT,                       -- NULL for whole order, or specific order_items(id)
    destination_address TEXT NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    freight_mode TEXT NOT NULL DEFAULT 'Surface', -- 'Surface', 'Air', 'Courier', 'Self-Pickup'
    carrier TEXT,
    docket_no TEXT,
    target_etd TEXT,
    created_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_shipping_dest_order ON order_shipping_destinations(order_id);

-- 4. Per-Piece Personalization (Strict 1 Alphabetical Name + 1 Numerical Number)
CREATE TABLE IF NOT EXISTS order_personalizations (
    id TEXT PRIMARY KEY,
    order_item_id TEXT NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
    order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    gender TEXT NOT NULL,                     -- e.g. 'Men', 'Women', 'Unisex', 'Kids'
    size TEXT NOT NULL,                       -- e.g. 'M', 'L', 'XL'
    name TEXT NOT NULL,                       -- Strict Alphabetical (A-Z, space, hyphen)
    number TEXT NOT NULL,                     -- Strict Numerical (0-999)
    remarks TEXT,
    created_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_personalization_item ON order_personalizations(order_item_id);
CREATE INDEX IF NOT EXISTS idx_personalization_order ON order_personalizations(order_id);

-- 5. Add zoho_synced column and grand totals to orders table if not existing
ALTER TABLE orders ADD COLUMN source_channel TEXT DEFAULT 'manual';
ALTER TABLE orders ADD COLUMN total_order_value REAL DEFAULT 0.0;
