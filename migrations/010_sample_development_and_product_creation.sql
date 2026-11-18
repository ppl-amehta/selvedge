-- Migration 010: Sample Development, Product Creation & Client Specific Pricing
-- Implements tabular projections, foreign key cascades, child-side indexes, and catalog sync

-- 1. Sample Development Root Table
CREATE TABLE IF NOT EXISTS sample_developments (
    id TEXT PRIMARY KEY,
    development_no TEXT UNIQUE NOT NULL,       -- 'D26110001'
    development_date TEXT NOT NULL,           -- 'YYYY-MM-DD'
    client_id TEXT NOT NULL,
    client_name TEXT NOT NULL,
    product_genre TEXT NOT NULL,              -- 'Apparel', 'Non-Apparel', etc.
    temp_product_name TEXT NOT NULL,
    image_url TEXT,                           -- Concept/sketch reference image
    product_description TEXT,                 -- Voice-transcribed or typed description
    temp_measurement_chart TEXT,              -- Formatted spec sheet / points of measure
    dev_comments TEXT,                        -- Merchandiser dev comments
    masters_comments TEXT,                    -- Pattern master voice transcribed comments
    rework_notes TEXT,                        -- Edits/rework notes if remaking sample
    sample_image_url TEXT,                    -- Physical sample photograph
    comments TEXT,                            -- General remarks
    status TEXT NOT NULL DEFAULT 'DRAFT',     -- 'DRAFT', 'IN_PROGRESS', 'SUBMITTED', 'REWORK_REQUESTED', 'APPROVED', 'REJECTED'
    meta_json JSON DEFAULT '{}',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sample_dev_client_date ON sample_developments(client_id, development_date DESC);
CREATE INDEX IF NOT EXISTS idx_sample_dev_status ON sample_developments(status);

-- 2. Sample Development Child Tables (Multi-Item Breakdowns)
CREATE TABLE IF NOT EXISTS sample_sizes (
    id TEXT PRIMARY KEY,
    development_id TEXT NOT NULL REFERENCES sample_developments(id) ON DELETE CASCADE,
    gender TEXT NOT NULL,
    size TEXT NOT NULL,
    qty INTEGER NOT NULL DEFAULT 1,
    sequence INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_sample_sizes_dev ON sample_sizes(development_id);

CREATE TABLE IF NOT EXISTS sample_fabrics (
    id TEXT PRIMARY KEY,
    development_id TEXT NOT NULL REFERENCES sample_developments(id) ON DELETE CASCADE,
    fabric_name TEXT NOT NULL,
    lot_no TEXT,
    average REAL NOT NULL DEFAULT 0.0,
    sequence INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_sample_fabrics_dev ON sample_fabrics(development_id);

CREATE TABLE IF NOT EXISTS sample_trims (
    id TEXT PRIMARY KEY,
    development_id TEXT NOT NULL REFERENCES sample_developments(id) ON DELETE CASCADE,
    trim_name TEXT NOT NULL,
    average REAL NOT NULL DEFAULT 0.0,
    sequence INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_sample_trims_dev ON sample_trims(development_id);

CREATE TABLE IF NOT EXISTS sample_zips (
    id TEXT PRIMARY KEY,
    development_id TEXT NOT NULL REFERENCES sample_developments(id) ON DELETE CASCADE,
    zip_name TEXT NOT NULL,
    length_and_qty TEXT,
    sequence INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_sample_zips_dev ON sample_zips(development_id);

CREATE TABLE IF NOT EXISTS sample_labels (
    id TEXT PRIMARY KEY,
    development_id TEXT NOT NULL REFERENCES sample_developments(id) ON DELETE CASCADE,
    label_name TEXT NOT NULL,
    sequence INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_sample_labels_dev ON sample_labels(development_id);

CREATE TABLE IF NOT EXISTS sample_pharmas (
    id TEXT PRIMARY KEY,
    development_id TEXT NOT NULL REFERENCES sample_developments(id) ON DELETE CASCADE,
    base_pharma_no TEXT,
    changes_made TEXT,
    new_pharma_no TEXT,
    sequence INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_sample_pharmas_dev ON sample_pharmas(development_id);

-- 3. Product Creation Root Table
CREATE TABLE IF NOT EXISTS master_products (
    id TEXT PRIMARY KEY,
    product_name TEXT UNIQUE NOT NULL,
    genre TEXT NOT NULL,
    sku_code TEXT UNIQUE,                     -- 'ET_PO_LINEFL'
    brand_label TEXT,
    why_we_made_this TEXT,
    garment_specifications TEXT,
    composition TEXT,
    mrp REAL NOT NULL DEFAULT 0.0,
    sample_development_id TEXT REFERENCES sample_developments(id) ON DELETE SET NULL,
    meta_json JSON DEFAULT '{}',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_master_products_genre ON master_products(genre);

-- 4. Product Creation Gender Variations & Specifications
CREATE TABLE IF NOT EXISTS master_product_genders (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL REFERENCES master_products(id) ON DELETE CASCADE,
    gender TEXT NOT NULL,
    pharma_no TEXT,
    measurement_chart_url TEXT,
    product_images_json JSON DEFAULT '[]',     -- Array of image URLs
    color_images_json JSON DEFAULT '[]',       -- Array of {color, front_url, back_url, side_url}
    sequence INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_product_genders_prod ON master_product_genders(product_id);

-- 5. Product Creation Component Materials (Fabrics, Trims, Zips BOM)
CREATE TABLE IF NOT EXISTS master_product_materials (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL REFERENCES master_products(id) ON DELETE CASCADE,
    material_type TEXT NOT NULL,               -- 'fabric', 'trim', 'zip'
    name TEXT NOT NULL,
    average REAL NOT NULL DEFAULT 0.0,
    gsm REAL,
    dia REAL,
    zip_length_qty TEXT,
    sequence INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_product_materials_prod ON master_product_materials(product_id);

-- 6. Client Specific Catalog & Negotiated Pricing
CREATE TABLE IF NOT EXISTS client_product_pricing (
    id TEXT PRIMARY KEY,
    client_id TEXT NOT NULL,
    client_name TEXT NOT NULL,
    product_id TEXT NOT NULL,
    product_name TEXT NOT NULL,
    client_image_url TEXT,
    contracted_price REAL NOT NULL DEFAULT 0.0,
    comments TEXT,
    effective_date TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_client_pricing_client ON client_product_pricing(client_id, product_id);
CREATE INDEX IF NOT EXISTS idx_client_pricing_effective ON client_product_pricing(effective_date DESC);
