package inventory_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ppl-amehta/selvedge/internal/erp/inventory"
	"github.com/ppl-amehta/selvedge/internal/graph"
	"github.com/ppl-amehta/selvedge/internal/workflow"
)

func setupTestDB(t *testing.T) (*sql.DB, *inventory.Service) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "selvedge-inventory-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON&_synchronous=NORMAL&_txlock=immediate")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })

	// Load schema migrations
	schemaFiles := []string{
		"../../../migrations/001_initial_graph.sql",
		"../../../migrations/002_seed_masters.sql",
		"../../../migrations/003_multi_product_orders.sql",
		"../../../migrations/004_three_stage_inward_pipeline.sql",
		"../../../migrations/005_order_prep_and_planning.sql",
		"../../../migrations/006_floor_cutting_station.sql",
		"../../../migrations/007_invoices_and_multi_fabric_lots.sql",
		"../../../migrations/008_consolidate_master_dropdowns.sql",
		"../../../migrations/009_clean_foundational_excel_truth.sql",
		"../../../migrations/010_sample_development_and_product_creation.sql",
		"../../../migrations/011_enrich_gender_sizes_and_dropdown_relationships.sql",
		"../../../migrations/013_enrich_attributes_and_schema.sql",
		"../../../migrations/014_backfill_seeded_bundles_and_fabric_types.sql",
	}

	for _, sf := range schemaFiles {
		content, err := os.ReadFile(sf)
		if err != nil {
			t.Fatalf("failed to read migration %s: %v", sf, err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			t.Fatalf("failed to execute migration %s: %v", sf, err)
		}
	}

	graphStore := graph.NewStore(db)
	wfEngine := workflow.NewEngine(db)
	svc := inventory.NewService(db, graphStore, wfEngine)

	return db, svc
}

func TestSuggestNextParentCode(t *testing.T) {
	_, svc := setupTestDB(t)

	code, err := svc.SuggestNextParentCode()
	if err != nil {
		t.Fatalf("SuggestNextParentCode failed: %v", err)
	}
	if len(code) < 8 || code[6] != '-' {
		t.Errorf("Unexpected format for parent code: %s", code)
	}
}

func TestThreeStageInwardPipeline_Fabric(t *testing.T) {
	_, svc := setupTestDB(t)

	// 1. Station 1A: Gate Inward Booking
	gateRes, err := svc.GateInward(inventory.GateInwardPayload{
		ParentCode:    "261026-A",
		InwardDate:    "2026-10-26",
		InwardType:    "Fabric",
		VendorName:    "Vardhman Textiles Ltd",
		InvoiceNo:     "INV-2026-8891",
		InvoiceQty:    1000.0,
		InvoiceAmount: 250000.0,
		UOM:           "Kgs",
		VendorProduct: "100% Cotton Single Jersey 180 GSM",
		FpccProduct:   "100% Cotton Single Jersey 180 GSM",
		VendorColor:   "Jet Black",
		FpccColor:     "Jet Black",
		NoOfBundles:   4,
		Comment:       "Gate truck arrival roll batch",
		Attrs: map[string]any{
			"fabricType":  "Knitted",
			"invoiceGsm":  180.0,
			"dia":         60.0,
			"composition": "100% Super Combed Cotton",
		},
	})
	if err != nil {
		t.Fatalf("GateInward failed: %v", err)
	}

	if gateRes.ParentCode != "261026-A" {
		t.Errorf("Expected parent code 261026-A, got %s", gateRes.ParentCode)
	}
	if gateRes.Stage != inventory.StageAwaitingCount {
		t.Errorf("Expected initial stage AWAITING_COUNT, got %s", gateRes.Stage)
	}
	if len(gateRes.Bundles) != 4 {
		t.Fatalf("Expected 4 bundle stubs, got %d", len(gateRes.Bundles))
	}
	if gateRes.Bundles[0].BundleCode != "261026-A-1/4" {
		t.Errorf("Expected bundle code 261026-A-1/4, got %s", gateRes.Bundles[0].BundleCode)
	}

	// 2. Station 1B: Floor Count with Shortage Discrepancy
	floorRes, err := svc.RecordFloorCount(inventory.FloorCountPayload{
		ParentCode:  "261026-A",
		InvoicedQty: 1000.0,
		CountedQty:  960.0, // Shortage of 40 Kgs (-4%)
		LocationBin: "A/R01/S01",
		CountedBy:   "Storekeeper Ramesh",
		Comments:    "Dock weighbridge roll count",
		HoldOnShort: true,
		Bundles: []inventory.FloorCountBundle{
			{BundleCode: "261026-A-1/4", Sequence: 1, CountedQty: 240.0, LocationBin: "A/R01/S01", MeasuredGsm: 182.0, MeasuredDia: 60.0},
			{BundleCode: "261026-A-2/4", Sequence: 2, CountedQty: 240.0, LocationBin: "A/R01/S02", MeasuredGsm: 180.0, MeasuredDia: 60.0},
			{BundleCode: "261026-A-3/4", Sequence: 3, CountedQty: 240.0, LocationBin: "A/R01/S03", MeasuredGsm: 179.0, MeasuredDia: 60.0},
			{BundleCode: "261026-A-4/4", Sequence: 4, CountedQty: 240.0, LocationBin: "A/R01/S04", MeasuredGsm: 181.0, MeasuredDia: 60.0},
		},
	})
	if err != nil {
		t.Fatalf("RecordFloorCount failed: %v", err)
	}

	if floorRes.Status != "SHORTAGE_FLAGGED" {
		t.Errorf("Expected SHORTAGE_FLAGGED, got %s", floorRes.Status)
	}
	if floorRes.VarianceQty != -40.0 {
		t.Errorf("Expected variance -40.0, got %f", floorRes.VarianceQty)
	}
	if floorRes.Stage != inventory.StageDiscrepancyHold {
		t.Errorf("Expected DISCREPANCY_HOLD stage on shortage, got %s", floorRes.Stage)
	}
	if floorRes.FinancialImpact != 10000.0 { // 40 Kgs * 250 unit rate
		t.Errorf("Expected financial impact 10000.0, got %f", floorRes.FinancialImpact)
	}

	// 3. Discrepancy Resolution: Approve shortage to proceed to QC
	resolveRes, err := svc.ResolveDiscrepancy("261026-A", "APPROVE_SHORTAGE", "Vendor agreed to issue credit note CN-442")
	if err != nil {
		t.Fatalf("ResolveDiscrepancy failed: %v", err)
	}
	if resolveRes["newStage"] != string(inventory.StageAwaitingQC) {
		t.Errorf("Expected stage AWAITING_QC after resolution, got %v", resolveRes["newStage"])
	}

	// 4. Station 1C: Bundle QC Inspection
	qcRes, err := svc.RecordBundleQC(inventory.BundleQCPayload{
		ParentCode: "261026-A",
		Inspector:  "QC Lead Priya",
		Inspections: []inventory.BundleQCItem{
			{BundleCode: "261026-A-1/4", Result: "PASS", Remarks: "Flawless shade and feel", ImageURL: "/data/uploads/qc_1.jpg"},
			{BundleCode: "261026-A-2/4", Result: "PASS", Remarks: "Within GSM tolerance", ImageURL: "/data/uploads/qc_2.jpg"},
			{BundleCode: "261026-A-3/4", Result: "PASS", Remarks: "Even knitting dia", ImageURL: "/data/uploads/qc_3.jpg"},
			{BundleCode: "261026-A-4/4", Result: "FAIL", Remarks: "Minor oil streak near selvedge edge", ImageURL: "/data/uploads/qc_4.jpg"},
		},
	})
	if err != nil {
		t.Fatalf("RecordBundleQC failed: %v", err)
	}

	if !qcRes.LotCompleted {
		t.Errorf("Expected lotCompleted = true after inspecting all bundles")
	}
	if qcRes.Stage != inventory.StageQCComplete {
		t.Errorf("Expected stage QC_COMPLETE, got %s", qcRes.Stage)
	}
	if qcRes.PassCount != 3 || qcRes.FailCount != 1 {
		t.Errorf("Expected 3 pass, 1 fail; got pass=%d, fail=%d", qcRes.PassCount, qcRes.FailCount)
	}

	// 5. Verify GetLot details
	detail, err := svc.GetLot("261026-A")
	if err != nil {
		t.Fatalf("GetLot failed: %v", err)
	}
	if detail.Lot.Stage != string(inventory.StageQCComplete) {
		t.Errorf("Expected lot detail stage QC_COMPLETE, got %s", detail.Lot.Stage)
	}
	if len(detail.Bundles) != 4 {
		t.Errorf("Expected 4 bundles in detail, got %d", len(detail.Bundles))
	}
	if detail.Bundles[3].QcStatus != "Rejected" {
		t.Errorf("Expected 4th bundle qcStatus Rejected, got %s", detail.Bundles[3].QcStatus)
	}
	if len(detail.Reconciliations) == 0 {
		t.Errorf("Expected at least 1 reconciliation record")
	}
	if len(detail.Inspections) != 4 {
		t.Errorf("Expected 4 inspection records, got %d", len(detail.Inspections))
	}
}

func TestThreeStageInwardPipeline_Zips(t *testing.T) {
	_, svc := setupTestDB(t)

	// Station 1A: Zips Gate Inward with bundle specs (size and rate per bundle)
	gateRes, err := svc.GateInward(inventory.GateInwardPayload{
		ParentCode:    "261026-B",
		InwardDate:    "2026-10-26",
		InwardType:    "Zips",
		VendorName:    "YKK India Pvt Ltd",
		InvoiceNo:     "YKK-5510",
		InvoiceQty:    500.0,
		InvoiceAmount: 25000.0,
		UOM:           "Pcs",
		VendorProduct: "No 5 Metal Brass Antique",
		FpccProduct:   "No 5 Metal Brass Antique",
		NoOfBundles:   2,
		Comment:       "Zippers for winter jacket collection",
		Attrs: map[string]any{
			"zipType":    "Metal",
			"slider":     "Auto-lock",
			"closure":    "Open-end",
			"tapeColor":  "Black",
			"teethColor": "Antique Brass",
		},
		Bundles: []inventory.GateBundleStub{
			{Sequence: 1, SizeSpec: "65 cm", Rate: 50.0},
			{Sequence: 2, SizeSpec: "70 cm", Rate: 50.0},
		},
	})
	if err != nil {
		t.Fatalf("GateInward Zips failed: %v", err)
	}
	if gateRes.Bundles[0].SizeSpec != "65 cm" || gateRes.Bundles[1].SizeSpec != "70 cm" {
		t.Errorf("Bundle size specs not preserved: %+v", gateRes.Bundles)
	}

	// Station 1B: Floor count matching invoice exactly
	floorRes, err := svc.RecordFloorCount(inventory.FloorCountPayload{
		ParentCode:  "261026-B",
		InvoicedQty: 500.0,
		CountedQty:  500.0,
		LocationBin: "B/R02/S05",
		CountedBy:   "Storekeeper",
		Comments:    "Perfect count",
		Bundles: []inventory.FloorCountBundle{
			{BundleCode: "261026-B-1/2", Sequence: 1, CountedQty: 250.0, LocationBin: "B/R02/S05", SizeSpec: "65 cm", Rate: 50.0},
			{BundleCode: "261026-B-2/2", Sequence: 2, CountedQty: 250.0, LocationBin: "B/R02/S05", SizeSpec: "70 cm", Rate: 50.0},
		},
	})
	if err != nil {
		t.Fatalf("RecordFloorCount Zips failed: %v", err)
	}
	if floorRes.Status != "MATCHED" {
		t.Errorf("Expected MATCHED, got %s", floorRes.Status)
	}
	if floorRes.Stage != inventory.StageAwaitingQC {
		t.Errorf("Expected stage AWAITING_QC, got %s", floorRes.Stage)
	}
}

func TestMultiProductFabric_And_InvoiceSplitting(t *testing.T) {
	_, svc := setupTestDB(t)

	// 1. Inward multi-product fabric consignment under single parent code 261028-A and invoice INV-SPLIT-999
	res1, err := svc.GateInward(inventory.GateInwardPayload{
		ParentCode:    "261028-A",
		InwardDate:    "2026-10-28",
		InwardType:    "Fabric",
		VendorName:    "Apex Textiles Ltd",
		InvoiceNo:     "INV-SPLIT-999",
		InvoiceQty:    150.0,
		InvoiceAmount: 52500.0,
		UOM:           "Kgs",
		FabricProducts: []inventory.FabricProductItem{
			{
				ID:            "fp_body",
				FabricType:    "Single Jersey",
				FpccProduct:   "Single Jersey 180 GSM",
				VendorProduct: "Super Combed SJ",
				Color:         "Navy Blue",
				UOM:           "Kgs",
				Dia:           "60\"",
				GSM:           180,
				Rate:          350.0,
				RollCount:     2,
			},
			{
				ID:            "fp_rib",
				FabricType:    "Rib 1x1 / 2x2",
				FpccProduct:   "1x1 Collar Rib 220 GSM",
				VendorProduct: "Collar Rib",
				Color:         "Navy Blue",
				UOM:           "Kgs",
				Dia:           "34\"",
				GSM:           220,
				Rate:          380.0,
				RollCount:     1,
			},
		},
		QRCopies: 3,
	})
	if err != nil {
		t.Fatalf("GateInward multi-fabric failed: %v", err)
	}
	if len(res1.Bundles) != 3 {
		t.Fatalf("Expected 3 bundles across multi-fabric products, got %d", len(res1.Bundles))
	}
	if res1.Bundles[0].BundleCode != "261028-A-1/3" || res1.Bundles[2].BundleCode != "261028-A-3/3" {
		t.Errorf("Unexpected bundle codes: %+v", res1.Bundles)
	}

	// Verify default auto-distribution of per-bundle invoicedQty across the 3 bundles (150 / 3 = 50)
	if res1.Bundles[0].InvoicedQty != 50.0 || res1.Bundles[2].InvoicedQty != 50.0 {
		t.Errorf("Expected default 50.0 invoicedQty per bundle, got %+v", res1.Bundles)
	}

	// 2. Inward a second Fabric lot (rib/cuff component) tagged to the SAME invoice INV-SPLIT-999 -> MUST SUCCEED
	res2, err := svc.GateInward(inventory.GateInwardPayload{
		ParentCode:    "261028-B",
		InwardDate:    "2026-10-28",
		InwardType:    "Fabric",
		VendorName:    "Apex Textiles Ltd",
		InvoiceNo:     "INV-SPLIT-999",
		InvoiceQty:    50.0,
		InvoiceAmount: 19000.0,
		UOM:           "Kgs",
		VendorProduct: "Collar Trim Fabric",
		FpccProduct:   "Venus Collar",
		NoOfBundles:   2,
	})
	if err != nil {
		t.Fatalf("GateInward second Fabric lot under same invoice failed: %v", err)
	}
	if res2.ParentCode != "261028-B" {
		t.Errorf("Expected 261028-B, got %s", res2.ParentCode)
	}
	if res2.Bundles[0].InvoicedQty != 25.0 || res2.Bundles[1].InvoicedQty != 25.0 {
		t.Errorf("Expected 25.0 invoicedQty per bundle for 50/2, got %+v", res2.Bundles)
	}

	// 2b. Attempting to mix a different Inward Type ('Zips') on the same invoice MUST BE BLOCKED
	_, err = svc.GateInward(inventory.GateInwardPayload{
		ParentCode:    "261028-Z",
		InwardDate:    "2026-10-28",
		InwardType:    "Zips",
		VendorName:    "Apex Textiles Ltd",
		InvoiceNo:     "INV-SPLIT-999",
		InvoiceQty:    200.0,
		InvoiceAmount: 10000.0,
		UOM:           "Pcs",
		VendorProduct: "YKK Metal #5",
		FpccProduct:   "Metal Zip 65cm",
		NoOfBundles:   2,
	})
	if err == nil || !strings.Contains(err.Error(), "only one inward type per invoice") {
		t.Fatalf("Expected error blocking mixed inward types on same invoice, got: %v", err)
	}

	// 3. Search Invoices by invoiceNo prefix
	invoices, err := svc.SearchInvoices("SPLIT")
	if err != nil {
		t.Fatalf("SearchInvoices failed: %v", err)
	}
	if len(invoices) != 1 {
		t.Fatalf("Expected 1 invoice, got %d", len(invoices))
	}
	inv := invoices[0]
	if inv.InvoiceNo != "INV-SPLIT-999" {
		t.Errorf("Expected invoice INV-SPLIT-999, got %s", inv.InvoiceNo)
	}
	if inv.InwardType != "Fabric" {
		t.Errorf("Expected invoice InwardType Fabric, got %s", inv.InwardType)
	}
	if len(inv.TaggedLots) != 2 {
		t.Errorf("Expected 2 tagged lots under invoice, got %d (%v)", len(inv.TaggedLots), inv.TaggedLots)
	}

	// 4. Test non-fabric rejection if >1 product provided (on its own Zips invoice)
	_, err = svc.GateInward(inventory.GateInwardPayload{
		ParentCode:    "261028-C",
		InwardDate:    "2026-10-28",
		InwardType:    "Zips",
		VendorName:    "Apex Textiles Ltd",
		InvoiceNo:     "INV-ZIPS-ONLY-1",
		InvoiceQty:    10.0,
		UOM:           "Pcs",
		FabricProducts: []inventory.FabricProductItem{
			{FabricType: "Zip A", RollCount: 1},
			{FabricType: "Zip B", RollCount: 1},
		},
	})
	if err == nil {
		t.Errorf("Expected error for non-fabric multi-product inward, got nil")
	}

	// 5. Test QC Inspection with audio URL & speech transcript
	qcRes, err := svc.RecordBundleQC(inventory.BundleQCPayload{
		ParentCode: "261028-A",
		Inspector:  "QC Specialist Ramesh",
		Inspections: []inventory.BundleQCItem{
			{
				BundleCode: "261028-A-1/3",
				Result:     "PASS",
				Remarks:    "Clean knit, no fly contamination",
				AudioURL:   "/uploads/qc-audio-1.webm",
				Transcript: "Roll checked under lightbox, zero defects found",
			},
		},
	})
	if err != nil {
		t.Fatalf("RecordBundleQC with audio failed: %v", err)
	}
	if qcRes.PassCount != 1 {
		t.Errorf("Expected 1 pass, got %d", qcRes.PassCount)
	}

	// Verify GetLot returns audio and transcript
	lotDetail, err := svc.GetLot("261028-A")
	if err != nil {
		t.Fatalf("GetLot failed: %v", err)
	}
	if len(lotDetail.Bundles) != 3 {
		t.Fatalf("Expected 3 bundles, got %d", len(lotDetail.Bundles))
	}
	b1 := lotDetail.Bundles[0]
	if b1.QcAudioUrl != "/uploads/qc-audio-1.webm" {
		t.Errorf("Expected QcAudioUrl to be /uploads/qc-audio-1.webm, got '%s'", b1.QcAudioUrl)
	}
	if b1.QcTranscript != "Roll checked under lightbox, zero defects found" {
		t.Errorf("Expected transcript match, got '%s'", b1.QcTranscript)
	}
	if lotDetail.Lot.QRCopiesPerBundle != 3 {
		t.Errorf("Expected QRCopiesPerBundle 3, got %d", lotDetail.Lot.QRCopiesPerBundle)
	}
}

func TestLegacySeededFabricAttrsFallbackAndFloorAttrsPersist(t *testing.T) {
	db, svc := setupTestDB(t)

	// Insert a legacy/seeded Fabric lot whose SQL columns hold GSM/Dia/FabricType while attrs_json is '{}'
	_, err := db.Exec(`
		INSERT INTO master_entities (id, name, entity_type, created_at)
		VALUES ('ent_v_legacy', 'Legacy Knit Mills', 'vendor', 1700000000)
		ON CONFLICT(id) DO NOTHING
	`)
	if err != nil {
		t.Fatalf("failed to seed vendor: %v", err)
	}
	_, err = db.Exec(`
		INSERT INTO parent_inward_lots (
			id, parent_code, inward_date, vendor_id, vendor_name, invoice_no, inward_type,
			uom, invoice_qty, invoice_amount, physical_received_qty,
			fabric_type, vendor_fab_name, fpcc_fab_name, vendor_color, fpcc_color,
			invoice_gsm, dia, stage, unit_rate, vendor_product, fpcc_product, no_of_bundles, attrs_json, fabric_products_json, created_at
		) VALUES (
			'plt_261199-Z', '261199-Z', '2026-11-17', 'ent_v_legacy', 'Legacy Knit Mills', 'INV-LEG-01', 'Fabric',
			'Kgs', 250.0, 87500.0, 0.0,
			'Knitted', '2441 Galactic Polo', '2441 Galactic Polo', 'Navy', 'Navy',
			220.0, 32.0, 'AWAITING_COUNT', 350.0, '2441 Galactic Polo', '2441 Galactic Polo', 1, '{}', '[]', 1700000000
		)
	`)
	if err != nil {
		t.Fatalf("failed to seed legacy lot: %v", err)
	}
	_, err = db.Exec(`
		INSERT INTO bundles (
			id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
			item_type, item_name, color_name, invoiced_qty, inward_weight_qty, available_qty,
			measured_gsm, measured_dia, size_spec, location_bin, qc_status, qr_payload, created_at
		) VALUES (
			'bnd_261199_z_1', '261199-Z-1/1', 'plt_261199-Z', '261199-Z', 1, 1,
			'Fabric', '2441 Galactic Polo', 'Navy', 250.0, 250.0, 250.0,
			220.0, 32.0, 'Knitted (Dia: 32", 220 GSM)', 'A/R01/S01', 'Pending', '{}', 1700000000
		)
	`)
	if err != nil {
		t.Fatalf("failed to seed legacy bundle: %v", err)
	}

	// Verify GetLot backfills GSM/Dia/FabricType into Attrs and FabricProducts
	seededDetail, err := svc.GetLot("261199-Z")
	if err != nil {
		t.Fatalf("GetLot 261199-Z failed: %v", err)
	}
	if seededDetail.Lot.Attrs["fabricType"] != "Knitted" {
		t.Errorf("Expected backfilled fabricType 'Knitted', got %v", seededDetail.Lot.Attrs["fabricType"])
	}
	if seededDetail.Lot.Attrs["invoiceGsm"] != 220.0 {
		t.Errorf("Expected backfilled invoiceGsm 220, got %v", seededDetail.Lot.Attrs["invoiceGsm"])
	}
	if seededDetail.Lot.Attrs["dia"] != 32.0 {
		t.Errorf("Expected backfilled dia 32, got %v", seededDetail.Lot.Attrs["dia"])
	}
	if len(seededDetail.Lot.FabricProducts) != 1 || seededDetail.Lot.FabricProducts[0].GSM != 220.0 {
		t.Errorf("Expected synthesized FabricProducts[0] GSM 220, got %+v", seededDetail.Lot.FabricProducts)
	}

	// Verify ListLots also returns backfilled Attrs for 261199-Z
	lots, err := svc.ListLots("ALL")
	if err != nil {
		t.Fatalf("ListLots failed: %v", err)
	}
	found := false
	for _, l := range lots {
		if l.ParentCode == "261199-Z" {
			found = true
			if l.Attrs["invoiceGsm"] != 220.0 || l.Attrs["dia"] != 32.0 {
				t.Errorf("Expected ListLots 261199-Z attrs to contain GSM=220 & Dia=32, got %+v", l.Attrs)
			}
		}
	}
	if !found {
		t.Fatalf("Expected 261199-Z in ListLots result")
	}

	// Verify Station 1B floor count persists updated Step 2 Attrs onto parent_inward_lots
	_, err = svc.RecordFloorCount(inventory.FloorCountPayload{
		ParentCode:  "261199-Z",
		InvoicedQty: 250.0,
		CountedQty:  250.0,
		LocationBin: "A/R01/S01",
		CountedBy:   "Tester",
		Attrs: map[string]any{
			"composition": "100%Cot",
			"fabricType":  "Pique",
		},
		Bundles: []inventory.FloorCountBundle{
			{
				BundleCode:  "261199-Z-1/1",
				Sequence:    1,
				InvoicedQty: 250.0,
				CountedQty:  250.0,
				LocationBin: "A/R01/S01",
				MeasuredGsm: 222,
				MeasuredDia: 32,
			},
		},
	})
	if err != nil {
		t.Fatalf("RecordFloorCount with Attrs failed: %v", err)
	}

	updatedDetail, err := svc.GetLot("261199-Z")
	if err != nil {
		t.Fatalf("GetLot post-floor failed: %v", err)
	}
	if updatedDetail.Lot.Attrs["composition"] != "100%Cot" || updatedDetail.Lot.Attrs["fabricType"] != "Pique" {
		t.Errorf("Expected persisted Step 2 attrs, got %+v", updatedDetail.Lot.Attrs)
	}
}
