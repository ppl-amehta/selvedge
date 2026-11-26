package cutting

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ppl-amehta/selvedge/internal/graph"
	"github.com/ppl-amehta/selvedge/internal/workflow"
)

func setupTestDB(t *testing.T) (*sql.DB, *Service) {
	t.Helper()
	dbPath := "/tmp/test_selvedge_cutting.db"
	_ = os.Remove(dbPath)

	database, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=ON")
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	migrations := []string{
		"../../../migrations/001_initial_graph.sql",
		"../../../migrations/002_seed_masters.sql",
		"../../../migrations/003_multi_product_orders.sql",
		"../../../migrations/004_three_stage_inward_pipeline.sql",
		"../../../migrations/005_order_prep_and_planning.sql",
		"../../../migrations/006_floor_cutting_station.sql",
	}

	for _, m := range migrations {
		content, err := os.ReadFile(m)
		if err != nil {
			t.Fatalf("failed to read migration %s: %v", m, err)
		}
		if _, err := database.Exec(string(content)); err != nil {
			t.Fatalf("failed to execute migration %s: %v", m, err)
		}
	}

	graphStore := graph.NewStore(database)
	wfEngine := workflow.NewEngine(database)
	svc := NewService(database, graphStore, wfEngine)

	return database, svc
}

func TestCuttingFloorWorkflow(t *testing.T) {
	database, svc := setupTestDB(t)
	defer database.Close()

	orderID := "ord_TEST_CUT_001"
	orderItemID := "item_TEST_CUT_001"
	now := graph.NowMs()

	// 1. Seed Order and Item with Gender-Size Matrix (100 pieces total)
	_, err := database.Exec(`
		INSERT INTO orders (
			id, production_order_no, sales_order_no, order_type, booking_date,
			client_id, client_name, category, sub_category, product_name,
			color_name, gender, unit_price, total_qty, target_etd, status,
			created_at, updated_at
		) VALUES (?, 'P-CUT-01', 'SO-CUT-01', 'Bulk', '2026-10-27',
			'c-001', 'CRED Technologies', 'Apparel', 'Polo', 'Pique Polo',
			'Navy Blue', 'Men', 450.0, 100, '2026-11-20', 'READY_FOR_EXECUTION', ?, ?)
	`, orderID, now, now)
	if err != nil {
		t.Fatalf("failed to insert test order: %v", err)
	}

	matrixJSON := `{"Men": {"S": 15, "M": 35, "L": 35, "XL": 15}}`
	_, err = database.Exec(`
		INSERT INTO order_items (
			id, order_id, item_sequence, category, sub_category, product_name,
			color_name, fabric_name, unit_price, total_qty, client_etd, matrix_json, created_at
		) VALUES (?, ?, 1, 'Apparel', 'Polo', 'Classic Pique Polo',
			'Navy Blue', 'Super Combed Pique 220', 450.0, 100, '2026-11-20', ?, ?)
	`, orderItemID, orderID, matrixJSON, now)
	if err != nil {
		t.Fatalf("failed to insert test order item: %v", err)
	}

	// 2. Test GetCuttingJob (Initialization)
	jobDetail, err := svc.GetCuttingJob(orderItemID)
	if err != nil {
		t.Fatalf("GetCuttingJob failed: %v", err)
	}
	if jobDetail.Job.ID != "CJ-P-CUT-01-1" {
		t.Fatalf("expected job id CJ-P-CUT-01-1, got %s", jobDetail.Job.ID)
	}
	if jobDetail.TotalTargetQty != 100 {
		t.Fatalf("expected total target 100, got %d", jobDetail.TotalTargetQty)
	}
	if len(jobDetail.SizeProgressList) != 4 {
		t.Fatalf("expected 4 size targets (S, M, L, XL), got %d", len(jobDetail.SizeProgressList))
	}
	t.Logf("✓ Cutting job initialized: %s with target 100 pcs across 4 sizes", jobDetail.Job.ID)

	// 3. Test IssueRollToJob
	rollCode := "261026-M-1/2"
	issueRes, err := svc.IssueRollToJob(jobDetail.Job.ID, rollCode, "Operator Ramesh")
	if err != nil {
		t.Fatalf("IssueRollToJob failed: %v", err)
	}
	if issueRes["rollCode"] != rollCode {
		t.Fatalf("unexpected issue roll output: %+v", issueRes)
	}
	t.Logf("✓ Fabric roll issued to cutting job: %s (%f kg)", rollCode, issueRes["issuedWeightKg"])

	// 4. Test RecordCutLay (Marker Lay Ratio Calculation)
	// Ratio: S:1, M:2, L:2, XL:1 = 6 pcs per lay.
	// Plies: 10 layers.
	// Total Pieces: 10 * 6 = 60 pieces.
	// Lay Weight: 21.0 kg -> actual lay avg = 21.0 / 60 = 0.35 kg/pc.
	layRec, err := svc.RecordCutLay(RecordLayPayload{
		CuttingJobID: jobDetail.Job.ID,
		RollCode:     rollCode,
		FabricName:   "Super Combed Pique 220",
		RollWeightKg: 45.0,
		MarkerName:   "Marker-S1-M2-L2-XL1",
		Ratio:        map[string]int{"S": 1, "M": 2, "L": 2, "XL": 1},
		Plies:        10,
		LayWeightKg:  21.0,
		NextProcess:  "Branding",
		OperatorName: "Operator Ramesh",
	})
	if err != nil {
		t.Fatalf("RecordCutLay failed: %v", err)
	}
	if layRec.TotalPiecesInLay != 60 {
		t.Fatalf("expected 60 pieces in lay, got %d", layRec.TotalPiecesInLay)
	}
	if layRec.ActualLayAvgKg != 0.35 {
		t.Fatalf("expected lay avg 0.35, got %f", layRec.ActualLayAvgKg)
	}
	t.Logf("✓ Cut Lay recorded: %s (Lay #%d: %d pieces, Actual Avg: %f kg/pc, Efficiency: %f%%)",
		layRec.ID, layRec.LayNumber, layRec.TotalPiecesInLay, layRec.ActualLayAvgKg, layRec.EfficiencyPct)

	// 5. Verify Generated Bundles
	updatedJob, err := svc.GetCuttingJob(jobDetail.Job.ID)
	if err != nil {
		t.Fatalf("failed to refetch cutting job: %v", err)
	}
	if len(updatedJob.Bundles) != 4 {
		t.Fatalf("expected 4 cut bundles generated, got %d", len(updatedJob.Bundles))
	}
	if updatedJob.TotalCutQty != 60 {
		t.Fatalf("expected total cut qty 60, got %d", updatedJob.TotalCutQty)
	}
	if updatedJob.TotalBalanceQty != 40 {
		t.Fatalf("expected balance to cut 40, got %d", updatedJob.TotalBalanceQty)
	}
	t.Logf("✓ Cut bundles verified: Total cut = 60, Balance to cut = 40 pcs")

	// 6. Test Remnant / End-Bit Recording ('Bit' vs 'Rej')
	// Case A: Usable End-Bit
	bitRemnant, err := svc.RecordRemnant(RecordRemnantPayload{
		Prefix:         "Bit",
		SourceRollCode: rollCode,
		CuttingJobID:   jobDetail.Job.ID,
		CutLayID:       layRec.ID,
		WeightKg:       2.5,
		BinLocation:    "4.Cutting Room",
		OperatorName:   "Operator Ramesh",
	})
	if err != nil {
		t.Fatalf("RecordRemnant Bit failed: %v", err)
	}
	expectedBitID := "Bit261026-M-1/2"
	if bitRemnant.ID != expectedBitID {
		t.Fatalf("expected Bit ID %s, got %s", expectedBitID, bitRemnant.ID)
	}
	t.Logf("✓ Usable End-Bit recorded: %s (%f kg)", bitRemnant.ID, bitRemnant.WeightKg)

	// Case B: Rejection / Defect
	rejRemnant, err := svc.RecordRemnant(RecordRemnantPayload{
		Prefix:         "Rej",
		SourceRollCode: rollCode,
		CuttingJobID:   jobDetail.Job.ID,
		CutLayID:       layRec.ID,
		WeightKg:       1.2,
		BinLocation:    "4.Cutting Room",
		DefectReason:   "Severe slub and dye banding in final 3 meters",
		OperatorName:   "Operator Ramesh",
	})
	if err != nil {
		t.Fatalf("RecordRemnant Rej failed: %v", err)
	}
	expectedRejID := "Rej261026-M-1/2"
	if rejRemnant.ID != expectedRejID {
		t.Fatalf("expected Rej ID %s, got %s", expectedRejID, rejRemnant.ID)
	}
	t.Logf("✓ Defective fabric recorded: %s (%f kg, Reason: %s)", rejRemnant.ID, rejRemnant.WeightKg, rejRemnant.DefectReason)

	// 7. Test Bundle Slip Printing
	firstBundle := updatedJob.Bundles[0]
	slipData, err := svc.PrintBundleSlip(firstBundle.ID)
	if err != nil {
		t.Fatalf("PrintBundleSlip failed: %v", err)
	}
	if slipData.SlipPrintedCount != 1 {
		t.Fatalf("expected slip printed count 1, got %d", slipData.SlipPrintedCount)
	}
	if slipData.QRData == "" {
		t.Fatalf("expected non-empty QR data")
	}
	t.Logf("✓ Bundle slip printed: Code=%s, Size=%s, Qty=%d, QRData=%s",
		slipData.BundleCode, slipData.Size, slipData.Pieces, slipData.QRData)
}

func TestCuttingJobLookupBySyntheticIDAndCaseInsensitiveRemnantPrefix(t *testing.T) {
	database, svc := setupTestDB(t)
	defer database.Close()

	now := graph.NowMs()
	_, err := database.Exec(`
		INSERT INTO orders (
			id, production_order_no, sales_order_no, order_type, booking_date,
			client_id, client_name, category, sub_category, product_name,
			color_name, gender, unit_price, total_qty, target_etd, status,
			created_at, updated_at
		) VALUES ('ord_CJ_LOOKUP', '11260001', 'SO-1126', 'Bulk', '2026-11-26',
			'c-001', 'ATC Tires Private Limited', 'Apparel', 'Polo', '3 Way polo',
			'Navy Blue', 'Men', 450.0, 800, '2026-12-15', 'BOOKED', ?, ?)
	`, now, now)
	if err != nil {
		t.Fatalf("insert order failed: %v", err)
	}

	for _, item := range []struct {
		id, prod, color string
		seq, qty        int
		matrix          string
	}{
		{"oi_cj_1", "3 Way polo", "Navy Blue", 1, 400, `{"Men":{"M":200,"L":200}}`},
		{"oi_cj_2", "Jute Tote Bag", "Natural", 2, 400, `{"General":{"Std":400}}`},
	} {
		_, err = database.Exec(`
			INSERT INTO order_items (
				id, order_id, item_sequence, category, sub_category, product_name,
				color_name, fabric_name, unit_price, total_qty, client_etd, matrix_json, created_at
			) VALUES (?, 'ord_CJ_LOOKUP', ?, 'Apparel', 'Polo', ?, ?, 'Pique', 450.0, ?, '2026-12-15', ?, ?)
		`, item.id, item.seq, item.prod, item.color, item.qty, item.matrix, now)
		if err != nil {
			t.Fatalf("insert order item %s failed: %v", item.id, err)
		}
	}

	// Synthetic IDs from ListCuttingOrders must initialize and return each distinct item cleanly.
	detail2, err := svc.GetCuttingJob("CJ-11260001-2")
	if err != nil {
		t.Fatalf("GetCuttingJob(CJ-11260001-2) failed: %v", err)
	}
	if detail2.Job.OrderItemID != "oi_cj_2" || detail2.Job.ProductName != "Jute Tote Bag" || detail2.TotalTargetQty != 400 {
		t.Fatalf("unexpected synthetic lookup result: %+v", detail2.Job)
	}

	rem, err := svc.RecordRemnant(RecordRemnantPayload{
		Prefix:         "BIT",
		SourceRollCode: "261126-A-1/2",
		CuttingJobID:   detail2.Job.ID,
		WeightKg:       1.5,
	})
	if err != nil {
		t.Fatalf("RecordRemnant with uppercase prefix failed: %v", err)
	}
	if rem.Prefix != "Bit" {
		t.Fatalf("expected canonical prefix 'Bit', got %q", rem.Prefix)
	}
}
