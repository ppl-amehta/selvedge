package orders_test

import (
	"os"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ppl-amehta/selvedge/internal/erp/orders"
	"github.com/ppl-amehta/selvedge/internal/graph"
	"github.com/ppl-amehta/selvedge/internal/platform/db"
	"github.com/ppl-amehta/selvedge/internal/workflow"
)

func setupTestDB(t *testing.T) (*db.DB, *orders.Service, *graph.Store) {
	tmpDB := "/tmp/test_selvedge_orders.db"
	_ = os.Remove(tmpDB)

	database, err := db.Connect(tmpDB)
	if err != nil {
		t.Fatalf("Failed to open test db: %v", err)
	}

	migrationFiles := []string{
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
	}

	for _, mFile := range migrationFiles {
		content, err := os.ReadFile(mFile)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", mFile, err)
		}
		if err := database.RunMigrations(mFile, string(content)); err != nil {
			t.Fatalf("Migration %s failed: %v", mFile, err)
		}
	}

	gStore := graph.NewStore(database.DB)
	wf := workflow.NewEngine(database.DB)
	svc := orders.NewService(database.DB, gStore, wf)

	return database, svc, gStore
}

func TestMultiProductOrderBookingAndGraphLineage(t *testing.T) {
	database, svc, gStore := setupTestDB(t)
	defer database.Close()
	defer os.Remove("/tmp/test_selvedge_orders.db")

	// 1. Test Math Checker Violation (Unbalanced Matrix)
	unbalancedPayload := orders.CreateOrderPayload{
		OrderNo:     "S261026A",
		ClientName:  "Google India Pvt Ltd",
		TaxRate:     5.0,
		TaxBilling:  "Extra",
		FreightMode: "Surface",
		Items: []orders.OrderItemPayload{
			{
				Category:    "Apparel",
				SubCategory: "Polo Shirt",
				ProductName: "Classic Pique Polo",
				ColorName:   "Navy Blue",
				UnitPrice:   450.0,
				TotalQty:    100, // Target 100
				Matrix: map[string]any{
					"Men": map[string]any{
						"S": 20, "M": 30, // Sum = 50 != 100
					},
				},
			},
		},
	}

	_, err := svc.CreateOrder(unbalancedPayload)
	if err == nil {
		t.Fatalf("Expected math checker violation for matrix sum 50 != totalQty 100, but got success")
	}
	t.Logf("✓ Math Checker correctly blocked unbalanced order: %v", err)

	// 2. Test Personalization Alpha/Numeric Rule Violations
	invalidNamePayload := unbalancedPayload
	invalidNamePayload.Items[0].TotalQty = 50 // Balance matrix
	invalidNamePayload.Items[0].Personalizations = []orders.PersonalizationPayload{
		{
			Gender: "Men",
			Size:   "M",
			Name:   "ROHIT 123", // Invalid digits in alphabetical name
			Number: "45",
		},
	}

	_, err = svc.CreateOrder(invalidNamePayload)
	if err == nil {
		t.Fatalf("Expected failure for non-alphabetical personalization name 'ROHIT 123', got success")
	}
	t.Logf("✓ Personalization rule correctly rejected name with digits: %v", err)

	invalidNumPayload := invalidNamePayload
	invalidNumPayload.Items[0].Personalizations[0].Name = "ROHIT SHARMA"
	invalidNumPayload.Items[0].Personalizations[0].Number = "NO_FORTY_FIVE" // Invalid non-numeric

	_, err = svc.CreateOrder(invalidNumPayload)
	if err == nil {
		t.Fatalf("Expected failure for non-numerical personalization number, got success")
	}
	t.Logf("✓ Personalization rule correctly rejected non-numeric jersey number: %v", err)

	// 3. Test Full Multi-Product Valid Order Booking
	validPayload := orders.CreateOrderPayload{
		OrderNo:        "S261026B",
		SalesOrderNo:   "SO-GOOGLE-2026",
		OrderType:      "Bulk",
		BookingDate:    "2026-10-26",
		ClientName:     "Google India Pvt Ltd",
		ContactPerson:  "Rohan Mathur",
		Phone:          "+91 98450 11223",
		Email:          "orders@google.com",
		Gstin:          "29AABCG1234F1Z6",
		BillingAddress: "Google Campus, RMZ Infinity, Old Madras Road, Bengaluru 560016",
		TaxRate:        5.0,
		TaxBilling:     "Extra",
		FreightMode:    "Surface",
		FreightCost:    "Extra",
		SourceChannel:  "zoho",
		Items: []orders.OrderItemPayload{
			// Item 1: Apparel Polo with Embellishments & Personalizations
			{
				ID:          "it_polo_1",
				Category:    "Apparel",
				SubCategory: "Polo Shirt",
				ProductName: "Classic Pique Polo",
				ColorName:   "Navy Blue",
				FabricName:  "100% Cotton Pique 220 GSM",
				UnitPrice:   450.0,
				TotalQty:    100,
				ClientEtd:   "2026-11-10",
				Matrix: map[string]any{
					"Men":   map[string]any{"S": 30, "M": 40},
					"Women": map[string]any{"S": 15, "M": 15},
				}, // 30+40+15+15 = 100
				Brandings: []orders.BrandingPayload{
					{
						BrandingType:    "emb",
						Position:        "Left Chest",
						Size:            "3.5 x 2.0 inches",
						Placement:       "7cm below collar",
						Color:           "Metallic Gold",
						Comment:         "3D puff embroidery",
						BrandingFileURL: "/data/uploads/google_crest.ai",
					},
					{
						BrandingType:    "dtf",
						Position:        "Right Sleeve",
						Size:            "2 x 1 inches",
						Placement:       "3cm above hem",
						Color:           "White Pantone 000C",
						Comment:         "High stretch backing",
						BrandingFileURL: "/data/uploads/sleeve_badge.png",
					},
				},
				Personalizations: []orders.PersonalizationPayload{
					{
						Gender:  "Men",
						Size:    "M",
						Name:    "SUNDAR PICHAI",
						Number:  "1",
						Remarks: "Gold thread back name",
					},
					{
						Gender:  "Women",
						Size:    "S",
						Name:    "APARNA SEN",
						Number:  "7",
						Remarks: "Standard print",
					},
				},
			},
			// Item 2: Non-Apparel Drinkware
			{
				ID:          "it_bottle_2",
				Category:    "Non-Apparel",
				SubCategory: "Drinkware / Sippers",
				ProductName: "Vacuum Insulated Steel Sipper 750ml",
				ColorName:   "Matte Black",
				UnitPrice:   350.0,
				TotalQty:    50,
				ClientEtd:   "2026-11-15",
				Matrix: map[string]any{
					"General": map[string]any{
						"Standard / Free": 50,
					},
				},
				Brandings: []orders.BrandingPayload{
					{
						BrandingType: "other", // Laser engraving
						Position:     "Body Center",
						Size:         "4 x 1.5 inches",
						Comment:      "Precision rotary fiber laser engraving",
					},
				},
			},
		},
		MultipleShipping: true,
		ShippingLocations: []orders.ShippingDestPayload{
			{
				OrderItemID:        "it_polo_1",
				DestinationAddress: "Google Bengaluru Campus, RMZ Infinity Hub",
				Quantity:           100,
				FreightMode:        "Surface",
				Carrier:            "Bluedart Surface",
				DocketNo:           "BD-98231",
			},
			{
				OrderItemID:        "it_bottle_2",
				DestinationAddress: "Google Hyderabad Campus, DivyaSree Omega Hub",
				Quantity:           50,
				FreightMode:        "Air",
				Carrier:            "Bluedart Air",
				DocketNo:           "BD-AIR-5541",
			},
		},
	}

	bookRes, err := svc.CreateOrder(validPayload)
	if err != nil {
		t.Fatalf("Failed to create multi-product valid order: %v", err)
	}

	if bookRes["productionOrderNo"] != "S261026B" {
		t.Errorf("Expected productionOrderNo S261026B, got %v", bookRes["productionOrderNo"])
	}
	if bookRes["totalQty"] != 150 {
		t.Errorf("Expected totalQty 150, got %v", bookRes["totalQty"])
	}

	t.Logf("✓ Multi-product order created successfully: %v", bookRes)

	// 4. Test GetOrder Full Retrieval
	orderObj, err := svc.GetOrder("S261026B")
	if err != nil {
		t.Fatalf("Failed to retrieve order: %v", err)
	}

	itemsList, ok := orderObj["items"].([]map[string]any)
	if !ok || len(itemsList) != 2 {
		t.Fatalf("Expected 2 product line items in order, got %d", len(itemsList))
	}

	// Verify Brandings on Item 1
	item1 := itemsList[0]
	brandings1, ok := item1["brandings"].([]map[string]any)
	if !ok || len(brandings1) != 2 {
		t.Fatalf("Expected 2 brandings on Item 1, got %d", len(brandings1))
	}
	if brandings1[0]["brandingType"] != "emb" {
		t.Errorf("Expected brandingType 'emb', got %v", brandings1[0]["brandingType"])
	}

	// Verify Personalizations on Item 1
	pers1, ok := item1["personalizations"].([]map[string]any)
	if !ok || len(pers1) != 2 {
		t.Fatalf("Expected 2 personalizations on Item 1, got %d", len(pers1))
	}
	if pers1[0]["name"] != "SUNDAR PICHAI" || pers1[0]["number"] != "1" {
		t.Errorf("Personalization row 1 mismatch: %v", pers1[0])
	}

	// Verify Shipping Destinations
	shippings, ok := orderObj["shippingLocations"].([]map[string]any)
	if !ok || len(shippings) != 2 {
		t.Fatalf("Expected 2 shipping destinations, got %d", len(shippings))
	}

	t.Logf("✓ Order retrieval returned complete master-detail tree with %d items, %d brandings, %d personalizations, and %d shipping locations",
		len(itemsList), len(brandings1), len(pers1), len(shippings))

	// 5. Verify Unified Graph Structure
	orderNode, err := gStore.GetNode("ord_S261026B")
	if err != nil {
		t.Fatalf("Failed to fetch ORDER node from graph: %v", err)
	}
	if orderNode.NodeType != graph.TypeOrder {
		t.Errorf("Expected node_type ORDER, got %s", orderNode.NodeType)
	}

	// Verify graph edges from ORDER node
	itemEdges, err := gStore.Traverse("ord_S261026B", graph.EdgeHasItem, "OUT")
	if err != nil || len(itemEdges) != 2 {
		t.Fatalf("Expected 2 OUT edges for HAS_ITEM, got %d (err: %v)", len(itemEdges), err)
	}

	t.Logf("✓ Bi-temporal graph nodes and typed relationships successfully verified in Unified Graph SGi record")
}

func TestZohoProgrammaticFetch(t *testing.T) {
	database, svc, _ := setupTestDB(t)
	defer database.Close()
	defer os.Remove("/tmp/test_selvedge_orders.db")

	zohoRes, err := svc.FetchZohoOrder("SO-CRED-2026")
	if err != nil {
		t.Fatalf("Failed to fetch Zoho order: %v", err)
	}

	if zohoRes.SalesOrderNo != "SO-CRED-2026" {
		t.Errorf("Expected SO-CRED-2026, got %s", zohoRes.SalesOrderNo)
	}
	if len(zohoRes.Items) == 0 {
		t.Errorf("Expected line items in Zoho response, got 0")
	}
	t.Logf("✓ Zoho programmatic pull returned client %q with %d items", zohoRes.ClientName, len(zohoRes.Items))
}

func TestSuggestNextOrderNo_FormatMMYY0000(t *testing.T) {
	database, svc, _ := setupTestDB(t)
	defer database.Close()
	defer os.Remove("/tmp/test_selvedge_orders.db")

	no1, err := svc.SuggestNextOrderNo()
	if err != nil {
		t.Fatalf("SuggestNextOrderNo failed: %v", err)
	}

	// Should be 8 characters: 4 prefix + 4 digits
	if len(no1) != 8 {
		t.Fatalf("Expected 8 characters format MMYY0000, got %q (len %d)", no1, len(no1))
	}
	if !strings.HasSuffix(no1, "0001") {
		t.Fatalf("Expected first order number to end in 0001, got %q", no1)
	}
	t.Logf("✓ Verified auto-increment initial format: %s", no1)

	// Insert order with no1
	payload := orders.CreateOrderPayload{
		OrderNo:    no1,
		ClientName: "Test Client Ltd",
		Items: []orders.OrderItemPayload{
			{
				Category:    "Non Apparel",
				SubCategory: "Drinkwear",
				ProductName: "Ember Steel Mug 650ml",
				Size:        "650ml",
				UnitPrice:   650,
				TotalQty:    10,
			},
		},
	}
	_, err = svc.CreateOrder(payload)
	if err != nil {
		t.Fatalf("Failed to create order %s: %v", no1, err)
	}

	// Now next order should be 0002
	no2, err := svc.SuggestNextOrderNo()
	if err != nil {
		t.Fatalf("Second SuggestNextOrderNo failed: %v", err)
	}
	if !strings.HasSuffix(no2, "0002") {
		t.Fatalf("Expected second order number to end in 0002, got %q", no2)
	}
	t.Logf("✓ Verified auto-increment second format: %s", no2)
}

func TestOrderFabricCompositionAndBrandLabelGenre(t *testing.T) {
	database, svc, _ := setupTestDB(t)
	defer database.Close()
	defer os.Remove("/tmp/test_selvedge_orders.db")

	resOrder, err := svc.CreateOrder(orders.CreateOrderPayload{
		ClientName: "Test Client Ltd",
		Items: []orders.OrderItemPayload{
			{
				Category:          "Apparel",
				SubCategory:       "Polo",
				ProductName:       "Classic Pique Polo",
				ColorName:         "White",
				FabricName:        "Venus",
				FabricComposition: "95%Poly 5%SP",
				BrandLabel:        "ETTU",
				BrandLabelGenre:   "DTF",
				BrandLabelDesc:    "ETTU +Size Navy/aqua",
				AdditionalLabels: []orders.LabelDetailPayload{
					{LabelType: "washcare Label", Description: "made in india white/black", Genre: "Printed"},
				},
				UnitPrice: 499,
				TotalQty:  10,
				Matrix:    map[string]any{"Men": map[string]any{"M": 10}},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder with fabricComposition/brandLabelGenre failed: %v", err)
	}
	fetched, err := svc.GetOrder(resOrder["productionOrderNo"].(string))
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}
	fetchedItems := fetched["items"].([]map[string]any)
	if fetchedItems[0]["fabricComposition"] != "95%Poly 5%SP" {
		t.Errorf("expected fabricComposition 95%%Poly 5%%SP, got %v", fetchedItems[0]["fabricComposition"])
	}
	if fetchedItems[0]["brandLabelGenre"] != "DTF" {
		t.Errorf("expected brandLabelGenre DTF, got %v", fetchedItems[0]["brandLabelGenre"])
	}
	if fetchedItems[0]["brandings"] == nil || fetchedItems[0]["personalizations"] == nil {
		t.Errorf("expected non-nil brandings & personalizations slices for safe CopySpecsModal cloning")
	}
}
