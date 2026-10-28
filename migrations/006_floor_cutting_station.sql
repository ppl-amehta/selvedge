-- ========================================================
-- FPCC SELVEDGE: FLOOR CUTTING STATION MIGRATION
-- Station 4: Floor Operator Cutting, Marker Palla Average,
-- Daily Size-Wise Cut Tracking, Remnant End-Bits, Bundle Slips
-- ========================================================

-- 1. Cutting Jobs per Production Order / Item
CREATE TABLE IF NOT EXISTS cutting_jobs (
    id TEXT PRIMARY KEY,                           -- 'CJ-' || production_order_no || '-' || seq
    order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    order_item_id TEXT REFERENCES order_items(id) ON DELETE CASCADE,
    production_order_no TEXT NOT NULL,
    product_name TEXT NOT NULL,
    color_name TEXT NOT NULL,
    fabric_name TEXT,
    target_qty INTEGER NOT NULL DEFAULT 0,
    pharma_no TEXT,                                -- Linked Pharma pattern tech pack
    planned_bom_avg REAL NOT NULL DEFAULT 0.0,     -- BOM planned kg/pc
    status TEXT NOT NULL DEFAULT 'OPEN'            -- 'OPEN', 'ISSUED', 'LAYING', 'CUT', 'BUNDLED', 'CLOSED'
        CHECK(status IN ('OPEN', 'ISSUED', 'LAYING', 'CUT', 'BUNDLED', 'CLOSED', 'CANCELLED')),
    operator_name TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cj_order ON cutting_jobs(order_id);
CREATE INDEX IF NOT EXISTS idx_cj_prod_no ON cutting_jobs(production_order_no);
CREATE INDEX IF NOT EXISTS idx_cj_status ON cutting_jobs(status);

-- 2. Physical Fabric Lays on Cutting Table (Marker Palla Average)
CREATE TABLE IF NOT EXISTS cut_lays (
    id TEXT PRIMARY KEY,                           -- 'CL-' || cutting_job_id || '-' || lay_number
    cutting_job_id TEXT NOT NULL REFERENCES cutting_jobs(id) ON DELETE CASCADE,
    lay_number INTEGER NOT NULL,                   -- 1, 2, 3...
    roll_code TEXT NOT NULL,                       -- Scanned Inward Roll e.g. '261026-M-1/2'
    fabric_name TEXT NOT NULL,
    roll_weight_kg REAL NOT NULL,                  -- Initial roll weight
    marker_name TEXT NOT NULL,                     -- e.g. 'Palla-Ratio-S-M-L'
    ratio_json JSON NOT NULL,                      -- e.g. {"S": 1, "M": 2, "L": 2, "XL": 1}
    plies INTEGER NOT NULL,                        -- Number of fabric layers laid
    total_pieces_in_lay INTEGER NOT NULL,          -- plies * sum(ratio)
    lay_weight_kg REAL NOT NULL,                   -- Measured weight of this lay
    actual_lay_avg_kg REAL NOT NULL,               -- lay_weight_kg / total_pieces_in_lay
    planned_bom_avg_kg REAL NOT NULL,              -- Target from BOM
    efficiency_pct REAL NOT NULL DEFAULT 100.0,    -- (planned / actual) * 100
    status TEXT NOT NULL DEFAULT 'CUT'
        CHECK(status IN ('LAYING', 'CUT', 'RECONCILED', 'CANCELLED')),
    operator_name TEXT NOT NULL,
    created_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cut_lays_job ON cut_lays(cutting_job_id);
CREATE INDEX IF NOT EXISTS idx_cut_lays_roll ON cut_lays(roll_code);

-- 3. Cut Bundles for Downstream Dispatch (Station 5: Branding / Machine)
CREATE TABLE IF NOT EXISTS cut_bundles (
    id TEXT PRIMARY KEY,                           -- 'CB-' || YYMMDD || '-' || seq (barcodeable)
    cutting_job_id TEXT NOT NULL REFERENCES cutting_jobs(id) ON DELETE CASCADE,
    cut_lay_id TEXT NOT NULL REFERENCES cut_lays(id) ON DELETE CASCADE,
    bundle_sequence INTEGER NOT NULL,
    size TEXT NOT NULL,
    pieces INTEGER NOT NULL,
    slip_printed_count INTEGER NOT NULL DEFAULT 0,
    printed_at INTEGER,
    next_process TEXT NOT NULL DEFAULT 'Branding'
        CHECK(next_process IN ('Branding', 'Machine', 'Finishing', 'Stock')),
    operator_name TEXT NOT NULL,
    created_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_bundles_job ON cut_bundles(cutting_job_id);
CREATE INDEX IF NOT EXISTS idx_bundles_lay ON cut_bundles(cut_lay_id);
CREATE INDEX IF NOT EXISTS idx_bundles_size ON cut_bundles(size);

-- 4. Remnants / End-Bits Returned to Inventory (Arjav Notes: 'Bit' / 'Rej' prefixes)
CREATE TABLE IF NOT EXISTS remnant_returns (
    id TEXT PRIMARY KEY,                           -- 'Bit' || clean_roll_code OR 'Rej' || clean_roll_code
    prefix TEXT NOT NULL CHECK(prefix IN ('Bit', 'Rej')),
    source_roll_code TEXT NOT NULL,                -- e.g. '261026-M-1/2'
    cutting_job_id TEXT REFERENCES cutting_jobs(id),
    cut_lay_id TEXT REFERENCES cut_lays(id),
    weight_kg REAL NOT NULL,
    bin_location TEXT NOT NULL DEFAULT '4.Cutting Room',
    defect_reason TEXT,                            -- Reason if 'Rej' (holes, dye shade variance, slub)
    operator_name TEXT NOT NULL,
    created_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_remnant_source ON remnant_returns(source_roll_code);
CREATE INDEX IF NOT EXISTS idx_remnant_job ON remnant_returns(cutting_job_id);

-- 5. Roll Issues to Floor Cutting Table
CREATE TABLE IF NOT EXISTS cutting_roll_issues (
    id TEXT PRIMARY KEY,
    cutting_job_id TEXT NOT NULL REFERENCES cutting_jobs(id) ON DELETE CASCADE,
    roll_code TEXT NOT NULL,
    fabric_name TEXT NOT NULL,
    color_name TEXT,
    issued_weight_kg REAL NOT NULL,
    status TEXT NOT NULL DEFAULT 'ISSUED'          -- 'ISSUED', 'IN_USE', 'CONSUMED', 'RETURNED'
        CHECK(status IN ('ISSUED', 'IN_USE', 'CONSUMED', 'RETURNED')),
    operator_name TEXT NOT NULL,
    issued_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_roll_issues_job ON cutting_roll_issues(cutting_job_id);
CREATE INDEX IF NOT EXISTS idx_roll_issues_roll ON cutting_roll_issues(roll_code);
