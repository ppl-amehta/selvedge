package planning

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
	dbPath := "/tmp/test_selvedge_planning.db"
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

func TestOrderPrepAndPlanningPipeline(t *testing.T) {
	database, svc := setupTestDB(t)
	defer database.Close()

	// 1. Seed an Order and Order Item
	orderID := "ord_TEST_PREP_001"
	orderItemID := "item_TEST_PREP_001"
	now := graph.NowMs()

	_, err := database.Exec(`
		INSERT INTO orders (
			id, production_order_no, sales_order_no, order_type, booking_date,
			client_id, client_name, category, sub_category, product_name,
			color_name, gender, unit_price, total_qty, target_etd, status,
			created_at, updated_at
		) VALUES (?, 'P-PREP-01', 'SO-PREP-01', 'Bulk', '2026-10-26',
			'c-001', 'Swiggy Instamart', 'Apparel', 'Polo', 'Classic Pique Polo',
			'Navy Blue', 'Men', 450.0, 100, '2026-11-15', 'BOOKED', ?, ?)
	`, orderID, now, now)
	if err != nil {
		t.Fatalf("failed to insert test order: %v", err)
	}

	_, err = database.Exec(`
		INSERT INTO order_items (
			id, order_id, item_sequence, category, sub_category, product_name,
			color_name, fabric_name, unit_price, total_qty, client_etd, created_at
		) VALUES (?, ?, 1, 'Apparel', 'Polo', 'Classic Pique Polo',
			'Navy Blue', 'Super Combed Pique 220', 450.0, 100, '2026-11-15', ?)
	`, orderItemID, orderID, now)
	if err != nil {
		t.Fatalf("failed to insert test order item: %v", err)
	}

	// 2. Test Guard: Advancing to PREP_COMPLETE should FAIL because no Pharma and no BOM lines exist
	_, err = svc.AdvanceOrderState(orderID, "PREP_COMPLETE", "premature test")
	if err == nil {
		t.Fatalf("expected error advancing to PREP_COMPLETE without pharma/bom, but got nil")
	}
	t.Logf("✓ Guard correctly blocked PREP_COMPLETE: %v", err)

	// 3. Test SavePharma
	pharma, err := svc.SavePharma(orderItemID, PharmaPayload{
		OrderID:         orderID,
		OrderItemID:     orderItemID,
		PharmaType:      "Body",
		PharmaNo:        "PH-POLO-2026-B1",
		ChangesToPharma: "Add 1cm to body length for shrinkage allowance",
		MeasurementChart: []MeasurementPoint{
			{Point: "Chest", Tolerance: "± 0.5 cm", S: 38, M: 40, L: 42, XL: 44, XXL: 46},
			{Point: "Body Length", Tolerance: "± 0.7 cm", S: 27, M: 28, L: 29, XL: 30, XXL: 31},
			{Point: "Sleeve Length", Tolerance: "± 0.5 cm", S: 8.5, M: 9, L: 9.5, XL: 10, XXL: 10.5},
		},
		Status: "APPROVED",
	})
	if err != nil {
		t.Fatalf("SavePharma failed: %v", err)
	}
	if pharma.ID == "" || pharma.PharmaNo != "PH-POLO-2026-B1" {
		t.Fatalf("unexpected pharma output: %+v", pharma)
	}
	t.Logf("✓ Pharma pattern created: %s (%s)", pharma.ID, pharma.PharmaNo)

	// Verify order status transitioned to IN_PREP
	var orderStatus string
	_ = database.QueryRow("SELECT status FROM orders WHERE id = ?", orderID).Scan(&orderStatus)
	if orderStatus != "IN_PREP" {
		t.Fatalf("expected order status IN_PREP, got: %s", orderStatus)
	}

	// 4. Test SaveBomLine (Fabric)
	// 100 pcs * 0.35 kg/pc * (1 + 3% wastage) = 100 * 0.35 * 1.03 = 36.05 kg
	bomFab, err := svc.SaveBomLine(orderItemID, BomLinePayload{
		OrderID:        orderID,
		OrderItemID:    orderItemID,
		LineType:       "FABRIC",
		ItemName:       "Super Combed Pique 220",
		SubType:        "Main Body",
		ConsumptionAvg: 0.35,
		WastagePct:     3.0,
		UOM:            "Kgs",
		UnitQty:        100,
		RatePerUnit:    380.0,
		Comments:       "High quality pique knit",
	})
	if err != nil {
		t.Fatalf("SaveBomLine Fabric failed: %v", err)
	}
	if bomFab.RequiredQty != 36.05 {
		t.Fatalf("expected RequiredQty 36.05, got %f", bomFab.RequiredQty)
	}
	t.Logf("✓ Fabric BOM Line created: %s (Required: %f %s, Shortfall: %f %s)",
		bomFab.ID, bomFab.RequiredQty, bomFab.UOM, bomFab.ShortfallQty, bomFab.UOM)

	// 5. Test SaveBomLine (Trims - Buttons)
	bomTrim, err := svc.SaveBomLine(orderItemID, BomLinePayload{
		OrderID:        orderID,
		OrderItemID:    orderItemID,
		LineType:       "ACCESSORY",
		ItemName:       "Button 18L Chalk",
		SubType:        "Placket",
		ConsumptionAvg: 3.0, // 3 buttons per polo
		WastagePct:     5.0,
		UOM:            "Pieces",
		UnitQty:        100,
		RatePerUnit:    0.85,
		Comments:       "Engraved 4-hole buttons",
	})
	if err != nil {
		t.Fatalf("SaveBomLine Trim failed: %v", err)
	}
	// 100 * 3.0 * 1.05 = 315 pcs
	if bomTrim.RequiredQty != 315.0 {
		t.Fatalf("expected RequiredQty 315, got %f", bomTrim.RequiredQty)
	}
	t.Logf("✓ Trim BOM Line created: %s (Required: %f %s)", bomTrim.ID, bomTrim.RequiredQty, bomTrim.UOM)

	// 6. Test RaiseProcurementRequisition for the fabric shortfall
	poReq, err := svc.RaiseProcurementRequisition(orderID, RaisePOPayload{
		OrderID:        orderID,
		BomLineID:      bomFab.ID,
		MaterialName:   bomFab.ItemName,
		MaterialType:   "FABRIC",
		UOM:            bomFab.UOM,
		RequiredQty:    bomFab.ShortfallQty,
		OrderBufferQty: 5.0, // +5kg buffer
		VendorName:     "Arvind Mills Ltd",
		EstimatedRate:  380.0,
		RequiredByDate: "2026-11-01",
	})
	if err != nil {
		t.Fatalf("RaiseProcurementRequisition failed: %v", err)
	}
	if poReq.ID == "" || poReq.TotalPoQty != (bomFab.ShortfallQty+5.0) {
		t.Fatalf("unexpected poReq: %+v", poReq)
	}
	t.Logf("✓ Procurement Requisition raised: %s (PO No: %s, Total PO Qty: %f %s)",
		poReq.ID, poReq.PoNo, poReq.TotalPoQty, poReq.UOM)

	// 7. Advance Order State to PREP_COMPLETE (Now it should PASS!)
	advRes, err := svc.AdvanceOrderState(orderID, "PREP_COMPLETE", "Pharma and BOM verified")
	if err != nil {
		t.Fatalf("failed to advance to PREP_COMPLETE: %v", err)
	}
	if advRes.NewState != "PREP_COMPLETE" {
		t.Fatalf("expected PREP_COMPLETE, got: %s", advRes.NewState)
	}
	t.Logf("✓ Order advanced to PREP_COMPLETE: %s", advRes.Message)

	// 8. Test SaveRoutingSteps (Order Planning / Scheduling)
	routingSteps, err := svc.SaveRoutingSteps(orderID, []RoutingStepPayload{
		{
			SequenceOrder:    1,
			ProcessCode:      "Cutting",
			StepKind:         "INTERNAL",
			VendorName:       "Cutting Room Line A",
			DurationDays:     2,
			PlannedStartDate: "2026-11-02",
			PlannedEndDate:   "2026-11-03",
			PlannedQty:       100,
			Comments:         "Precision lay cutting",
		},
		{
			SequenceOrder:    2,
			ProcessCode:      "Branding",
			StepKind:         "VENDOR",
			VendorName:       "Rajesh Embroidery Works",
			DurationDays:     3,
			PlannedStartDate: "2026-11-04",
			PlannedEndDate:   "2026-11-06",
			PlannedQty:       100,
			Comments:         "Left chest logo 6000 stitches",
		},
		{
			SequenceOrder:    3,
			ProcessCode:      "Machine",
			StepKind:         "INTERNAL",
			VendorName:       "Stitching Floor Line 2",
			DurationDays:     4,
			PlannedStartDate: "2026-11-07",
			PlannedEndDate:   "2026-11-10",
			PlannedQty:       100,
			Comments:         "Overlock + Flatlock placket assembly",
		},
		{
			SequenceOrder:    4,
			ProcessCode:      "Finishing",
			StepKind:         "INTERNAL",
			VendorName:       "Finishing Section",
			DurationDays:     2,
			PlannedStartDate: "2026-11-11",
			PlannedEndDate:   "2026-11-12",
			PlannedQty:       100,
			Comments:         "Thread trim, buttoning, steam iron",
		},
		{
			SequenceOrder:    5,
			ProcessCode:      "Packing",
			StepKind:         "INTERNAL",
			VendorName:       "Packing Section",
			DurationDays:     1,
			PlannedStartDate: "2026-11-13",
			PlannedEndDate:   "2026-11-13",
			PlannedQty:       100,
			Comments:         "Individual polybag + 7-ply master carton",
		},
	})
	if err != nil {
		t.Fatalf("SaveRoutingSteps failed: %v", err)
	}
	if len(routingSteps) != 5 {
		t.Fatalf("expected 5 routing steps, got %d", len(routingSteps))
	}
	t.Logf("✓ Routing steps saved (5 steps from Cutting to Packing)")

	// 9. Advance Order State to PLANNED (should snapshot and check ETD)
	planRes, err := svc.AdvanceOrderState(orderID, "PLANNED", "Routing confirmed with internal lines and vendor")
	if err != nil {
		t.Fatalf("failed to advance to PLANNED: %v", err)
	}
	if planRes.NewState != "PLANNED" {
		t.Fatalf("expected PLANNED, got: %s", planRes.NewState)
	}
	t.Logf("✓ Order advanced to PLANNED: %s", planRes.Message)

	// 10. Fetch Full Order Prep & Planning Data
	fullData, err := svc.GetOrderPrepAndPlanning(orderID)
	if err != nil {
		t.Fatalf("GetOrderPrepAndPlanning failed: %v", err)
	}
	if len(fullData.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(fullData.Items))
	}
	if len(fullData.Items[0].PharmaPatterns) != 1 {
		t.Fatalf("expected 1 pharma pattern, got %d", len(fullData.Items[0].PharmaPatterns))
	}
	if len(fullData.Items[0].BomLines) != 2 {
		t.Fatalf("expected 2 bom lines, got %d", len(fullData.Items[0].BomLines))
	}
	if len(fullData.RoutingSteps) != 5 {
		t.Fatalf("expected 5 routing steps, got %d", len(fullData.RoutingSteps))
	}
	if fullData.GanttSummary.SlackDays != 2 { // Client ETD 2026-11-15 - Latest Finish 2026-11-13 = 2 days
		t.Fatalf("expected SlackDays 2, got %d", fullData.GanttSummary.SlackDays)
	}
	if fullData.GanttSummary.IsOverdueRisk {
		t.Fatalf("expected no overdue risk, but got true")
	}
	t.Logf("✓ Full Order Prep & Planning tree verified: SlackDays=%d, StatusText=%s",
		fullData.GanttSummary.SlackDays, fullData.GanttSummary.StatusText)

	// 11. Advance to READY_FOR_EXECUTION
	execRes, err := svc.AdvanceOrderState(orderID, "READY_FOR_EXECUTION", "Approved for factory floor release")
	if err != nil {
		t.Fatalf("failed to advance to READY_FOR_EXECUTION: %v", err)
	}
	if execRes.NewState != "READY_FOR_EXECUTION" {
		t.Fatalf("expected READY_FOR_EXECUTION, got: %s", execRes.NewState)
	}
	t.Logf("✓ Order successfully reached terminal gate READY_FOR_EXECUTION!")
}
