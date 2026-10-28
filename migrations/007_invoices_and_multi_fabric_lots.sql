-- ========================================================
-- FPCC SELVEDGE: INWARD INVOICES & MULTI-PRODUCT FABRIC LOTS
-- Migration 007
-- ========================================================

-- 1. Inward Invoices Master Table
CREATE TABLE IF NOT EXISTS inward_invoices (
    id TEXT PRIMARY KEY,
    invoice_no TEXT NOT NULL,
    vendor_name TEXT NOT NULL,
    invoice_date TEXT NOT NULL,
    invoice_qty REAL NOT NULL DEFAULT 0.0,
    invoice_amount REAL NOT NULL DEFAULT 0.0,
    uom TEXT DEFAULT 'Kgs',
    currency TEXT DEFAULT 'INR',
    remarks TEXT,
    created_at INTEGER NOT NULL,
    UNIQUE(vendor_name, invoice_no)
);

CREATE INDEX IF NOT EXISTS idx_inward_invoices_no ON inward_invoices(invoice_no);
CREATE INDEX IF NOT EXISTS idx_inward_invoices_vendor ON inward_invoices(vendor_name);

-- 2. Link parent_inward_lots to inward_invoices and multi-product fabric support
ALTER TABLE parent_inward_lots ADD COLUMN invoice_id TEXT;
ALTER TABLE parent_inward_lots ADD COLUMN fabric_products_json JSON DEFAULT '[]';
ALTER TABLE parent_inward_lots ADD COLUMN qr_copies_per_bundle INTEGER DEFAULT 3;

-- 3. Extend QC inspections with audio notes & speech-to-text transcript
ALTER TABLE bundle_qc_inspections ADD COLUMN audio_url TEXT;
ALTER TABLE bundle_qc_inspections ADD COLUMN transcript TEXT;

-- 4. Extend bundles with audio note & transcript
ALTER TABLE bundles ADD COLUMN qc_audio_url TEXT;
ALTER TABLE bundles ADD COLUMN qc_transcript TEXT;

-- 5. Backfill existing invoices from parent_inward_lots
INSERT OR IGNORE INTO inward_invoices (id, invoice_no, vendor_name, invoice_date, invoice_qty, invoice_amount, uom, currency, remarks, created_at)
SELECT 
    'inv_' || hex(randomblob(6)),
    invoice_no,
    vendor_name,
    inward_date,
    SUM(invoice_qty),
    SUM(invoice_amount),
    uom,
    'INR',
    'Historical invoice sync',
    strftime('%s', 'now')
FROM parent_inward_lots
WHERE invoice_no IS NOT NULL AND invoice_no != ''
GROUP BY vendor_name, invoice_no;
