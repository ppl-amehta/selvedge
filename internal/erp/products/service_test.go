package products

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ppl-amehta/selvedge/internal/graph"
	"github.com/ppl-amehta/selvedge/internal/platform/db"
	"github.com/ppl-amehta/selvedge/internal/workflow"
)

func setupTestDB(t *testing.T) (*db.DB, *Service) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_products.db")

	database, err := db.Connect(dbPath)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// Apply migrations 001 through 011, 013
	migrationFiles := []string{
		"migrations/001_initial_graph.sql",
		"migrations/002_seed_masters.sql",
		"migrations/003_multi_product_orders.sql",
		"migrations/004_three_stage_inward_pipeline.sql",
		"migrations/005_order_prep_and_planning.sql",
		"migrations/006_floor_cutting_station.sql",
		"migrations/007_invoices_and_multi_fabric_lots.sql",
		"migrations/008_consolidate_master_dropdowns.sql",
		"migrations/009_clean_foundational_excel_truth.sql",
		"migrations/010_sample_development_and_product_creation.sql",
		"migrations/011_enrich_gender_sizes_and_dropdown_relationships.sql",
		"migrations/013_enrich_attributes_and_schema.sql",
	}

	for _, mFile := range migrationFiles {
		content, err := os.ReadFile(filepath.Join("../../../", mFile))
		if err != nil {
			t.Fatalf("failed to read migration file %s: %v", mFile, err)
		}
		if err := database.RunMigrations(mFile, string(content)); err != nil {
			t.Fatalf("failed to run migration %s: %v", mFile, err)
		}
	}

	gStore := graph.NewStore(database.DB)
	wfEngine := workflow.NewEngine(database.DB)
	svc := NewService(database.DB, gStore, wfEngine)

	return database, svc
}

func TestSuggestNextDevelopmentNo(t *testing.T) {
	db, svc := setupTestDB(t)
	defer db.Close()

	now := time.Now()
	expectedPrefix := fmt.Sprintf("D%02d%02d", now.Year()%100, int(now.Month()))
	expectedFirst := expectedPrefix + "0001"

	no, err := svc.SuggestNextDevelopmentNo()
	if err != nil {
		t.Fatalf("unexpected error suggesting dev no: %v", err)
	}
	if no != expectedFirst {
		t.Fatalf("expected initial dev no %s, got %s", expectedFirst, no)
	}

	// Insert a sample
	_, err = svc.CreateSampleDevelopment(SampleDevelopmentPayload{
		DevelopmentNo:   expectedFirst,
		ClientName:      "Amara Raja Energy & Mobility Limited",
		ClientID:        "ent_amara_raja",
		ProductGenre:    "Apparel",
		TempProductName: "Performance Aerocool Polo",
	})
	if err != nil {
		t.Fatalf("failed to create sample: %v", err)
	}

	// Next suggest should be 0002
	next, err := svc.SuggestNextDevelopmentNo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedSecond := expectedPrefix + "0002"
	if next != expectedSecond {
		t.Fatalf("expected next dev no %s, got %s", expectedSecond, next)
	}
}

func TestSuggestSKU(t *testing.T) {
	db, svc := svcWithDB(setupTestDB(t))
	defer db.Close()

	sku := svc.SuggestSKU("FPCC Elite", "Apparel", "Luxury Classic Polo")
	// FP + AP + LUXURY -> FP_AP_LUXURY
	if sku != "FP_AP_LUXURY" {
		t.Fatalf("expected SKU FP_AP_LUXURY, got %s", sku)
	}

	// Insert product with this SKU
	_, err := svc.CreateMasterProduct(CreateProductPayload{
		ProductName: "Luxury Classic Polo",
		Genre:       "Apparel",
		SKUCode:     sku,
		BrandLabel:  "FPCC Elite",
	})
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	// Collision check: next suggest should append -01
	skuCollision := svc.SuggestSKU("FPCC Elite", "Apparel", "Luxury Classic Polo")
	if skuCollision != "FP_AP_LUXURY-01" {
		t.Fatalf("expected collision resolved SKU FP_AP_LUXURY-01, got %s", skuCollision)
	}
}

func svcWithDB(d *db.DB, s *Service) (*db.DB, *Service) {
	return d, s
}

func TestSampleDevelopmentLifecycle(t *testing.T) {
	db, svc := setupTestDB(t)
	defer db.Close()

	payload := SampleDevelopmentPayload{
		DevelopmentDate:      "2026-11-18",
		ClientID:             "cli_atc_tires",
		ClientName:           "ATC Tires Private Limited",
		ProductGenre:         "Apparel",
		TempProductName:      "Engineers High-Vis Tech Polo",
		ImageURL:             "/data/uploads/concept_sketch_1.jpg",
		ProductDescription:   "Heavy duty moisture wicking polo with 3M reflective tape on sleeves",
		TempMeasurementChart: "CHEST: 42, LENGTH: 28, SHOULDER: 18.5, SLEEVE: 9.0",
		DevComments:          "Ensure contrast tipping on rib collar is exactly 2mm",
		MastersComments:      "Armhole curve lowered by 0.5 cm for industrial mobility",
		ReworkNotes:          "",
		SampleImageURL:       "/data/uploads/sample_photo_finished.jpg",
		Comments:             "Urgent client milestone sample for factory workers trial",
		Status:               "SUBMITTED",
		Sizes: []SampleSizeItem{
			{Gender: "Men", Size: "M", Qty: 2},
			{Gender: "Men", Size: "L", Qty: 2},
			{Gender: "Women", Size: "S", Qty: 1},
		},
		Fabrics: []SampleFabricItem{
			{FabricName: "Honey Comb Poly Cotton", LotNo: "LOT-HC-992", Average: 0.32},
		},
		Trims: []SampleTrimItem{
			{TrimName: "FPCC Signature Laser Button", Average: 3.0},
			{TrimName: "3M Reflective Piping", Average: 0.65},
		},
		Zips: []SampleZipItem{
			{ZipName: "YKK Concealed Chest Pocket Zip", LengthAndQty: "15cm - 1pc"},
		},
		Labels: []SampleLabelItem{
			{LabelName: "ATC Tires Co-Branded Heat Transfer Label"},
			{LabelName: "Tearaway Washcare Label"},
		},
		Pharmas: []SamplePharmaItem{
			{BasePharmaNo: "PH-POLO-STD-M", ChangesMade: "Girth +1 inch, armhole dropped", NewPharmaNo: "PH-ATC-IND-2026"},
		},
	}

	res, err := svc.CreateSampleDevelopment(payload)
	if err != nil {
		t.Fatalf("failed to create sample development: %v", err)
	}

	sampleID := res["id"].(string)
	devNo := res["developmentNo"].(string)

	// Fetch detail
	detail, err := svc.GetSampleDevelopment(sampleID)
	if err != nil {
		t.Fatalf("failed to get sample development: %v", err)
	}

	if detail.DevelopmentNo != devNo {
		t.Errorf("expected devNo %s, got %s", devNo, detail.DevelopmentNo)
	}
	if len(detail.Sizes) != 3 {
		t.Errorf("expected 3 sizes, got %d", len(detail.Sizes))
	}
	if len(detail.Fabrics) != 1 || detail.Fabrics[0].Average != 0.32 {
		t.Errorf("expected fabric average 0.32, got %v", detail.Fabrics)
	}
	if len(detail.Trims) != 2 {
		t.Errorf("expected 2 trims, got %d", len(detail.Trims))
	}
	if len(detail.Zips) != 1 {
		t.Errorf("expected 1 zip, got %d", len(detail.Zips))
	}
	if len(detail.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(detail.Labels))
	}
	if len(detail.Pharmas) != 1 || detail.Pharmas[0].NewPharmaNo != "PH-ATC-IND-2026" {
		t.Errorf("expected new pharma PH-ATC-IND-2026, got %v", detail.Pharmas)
	}

	// Update status
	if err := svc.UpdateSampleDevelopmentStatus(sampleID, "APPROVED", "Approved without changes by client VP"); err != nil {
		t.Fatalf("failed to update status: %v", err)
	}

	updated, _ := svc.GetSampleDevelopment(sampleID)
	if updated.Status != "APPROVED" {
		t.Errorf("expected status APPROVED, got %s", updated.Status)
	}
	if updated.ReworkNotes != "Approved without changes by client VP" {
		t.Errorf("expected rework notes updated, got %s", updated.ReworkNotes)
	}
}

func TestProductCreationAndCatalogSync(t *testing.T) {
	db, svc := setupTestDB(t)
	defer db.Close()

	payload := CreateProductPayload{
		ProductName:           "Frontier Hydrophobic Tech Polo",
		Genre:                 "Apparel",
		BrandLabel:            "FPCC Precision",
		WhyWeMadeThis:         "Engineered for industrial wet-lab conditions with water-repellent yarn",
		GarmentSpecifications: "Raglan cut, flat-knit rib collar, double needle reinforced hem",
		Composition:           "65% Combed Cotton, 35% Recycled Poly",
		MRP:                   1499.00,
		Genders: []ProductGenderItem{
			{
				Gender:              "Men",
				PharmaNo:            "PH-M-TECH-01",
				MeasurementChartURL: "/data/uploads/specs_men_tech.pdf",
				ProductImages:       []string{"/data/uploads/men_front.jpg", "/data/uploads/men_back.jpg"},
			},
			{
				Gender:              "Women",
				PharmaNo:            "PH-W-TECH-01",
				MeasurementChartURL: "/data/uploads/specs_women_tech.pdf",
				ProductImages:       []string{"/data/uploads/women_front.jpg"},
			},
		},
		Materials: []ProductMaterialItem{
			{MaterialType: "fabric", Name: "Hydrophobic Pique", Average: 0.34, GSM: 220, Dia: 32},
			{MaterialType: "trim", Name: "Gunmetal Snap Buttons", Average: 3},
		},
	}

	res, err := svc.CreateMasterProduct(payload)
	if err != nil {
		t.Fatalf("failed to create master product: %v", err)
	}

	prodID := res["id"].(string)

	// Fetch detail
	prod, err := svc.GetMasterProduct(prodID)
	if err != nil {
		t.Fatalf("failed to get product: %v", err)
	}
	if prod.ProductName != "Frontier Hydrophobic Tech Polo" {
		t.Errorf("unexpected product name: %s", prod.ProductName)
	}
	if len(prod.Genders) != 2 {
		t.Errorf("expected 2 gender rows, got %d", len(prod.Genders))
	}
	if len(prod.Materials) != 2 {
		t.Errorf("expected 2 material rows, got %d", len(prod.Materials))
	}

	// Verify catalog synchronization into master_items
	var catName string
	err = db.DB.QueryRow("SELECT display_name FROM master_items WHERE id = ?", "mi_prod_"+prodID).Scan(&catName)
	if err != nil {
		t.Fatalf("product failed to sync into master_items: %v", err)
	}
	if catName != "Frontier Hydrophobic Tech Polo" {
		t.Errorf("expected master_items display_name to match, got %s", catName)
	}
}

func TestClientProductPricing(t *testing.T) {
	db, svc := setupTestDB(t)
	defer db.Close()

	payload := ClientPricingPayload{
		ClientID:        "ent_atc_tires",
		ClientName:      "ATC Tires Private Limited",
		ProductID:       "prod_frontier_polo",
		ProductName:     "Frontier Hydrophobic Tech Polo",
		ClientImageURL:  "/data/uploads/atc_custom_embroidery_mock.jpg",
		ContractedPrice: 725.50,
		Comments:        "Volume rate contract for 5000+ units annual commit",
		EffectiveDate:   "2026-11-18",
	}

	_, err := svc.SetClientProductPricing(payload)
	if err != nil {
		t.Fatalf("failed to set client pricing: %v", err)
	}

	list, err := svc.ListClientProductPricing("ent_atc_tires")
	if err != nil {
		t.Fatalf("failed to list client pricing: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 pricing record, got %d", len(list))
	}
	if list[0].ContractedPrice != 725.50 {
		t.Errorf("expected price 725.50, got %f", list[0].ContractedPrice)
	}
}

func TestEnrichedSampleAttrsAndProductMultiSizesPharmas(t *testing.T) {
	db, svc := setupTestDB(t)
	defer db.Close()

	sampleRes, err := svc.CreateSampleDevelopment(SampleDevelopmentPayload{
		ClientID:        "ent_amara_raja",
		ClientName:      "Amara Raja Energy & Mobility Limited",
		ProductGenre:    "Jackets",
		TempProductName: "Tech Softshell Jacket",
		Colors:          []string{"Angora", "Classic Navy"},
		Fabrics: []SampleFabricItem{
			{
				FabricName: "Venus",
				LotNo:      "LOT-881",
				Average:    0.75,
				Attrs:      map[string]any{"fabricType": "Fabric", "fabricGeneric": "Bonded", "composition": "95%Poly 5%SP", "gsm": 260, "dia": "60\""},
			},
		},
		Trims: []SampleTrimItem{
			{
				TrimName: "Flat cord",
				Average:  1.4,
				Attrs:    map[string]any{"trimType": "Cord", "size": "54\"", "cordEnd": "metal silver", "color": "Black"},
			},
		},
		Zips: []SampleZipItem{
			{
				ZipName:      "Waterproof",
				LengthAndQty: "S-L: 68cm",
				Attrs:        map[string]any{"zipType": "Waterproof", "teethSize": "8.0", "slider": "DA8DR4", "closure": "open ended", "finish": "Matt silver", "tapeColor": "Black"},
			},
		},
		Labels: []SampleLabelItem{
			{
				LabelName: "ETTU",
				Attrs:     map[string]any{"labelCategory": "Brand Label", "labelGenre": "Woven", "labelDescription": "ETTU +Size Navy/aqua", "baseColor": "White", "textColor": "Classic Navy"},
			},
		},
		Pharmas: []SamplePharmaItem{
			{BasePharmaNo: "PH-JKT-01", NewPharmaNo: "PH-JKT-01-R1"},
			{BasePharmaNo: "PH-JKT-02", NewPharmaNo: "PH-JKT-02-R1"},
		},
	})
	if err != nil {
		t.Fatalf("CreateSampleDevelopment failed: %v", err)
	}

	sDetail, err := svc.GetSampleDevelopment(sampleRes["id"].(string))
	if err != nil {
		t.Fatalf("GetSampleDevelopment failed: %v", err)
	}
	if len(sDetail.Colors) != 2 || sDetail.Colors[1] != "Classic Navy" {
		t.Errorf("expected 2 sample colors, got %+v", sDetail.Colors)
	}
	if sDetail.Zips[0].Attrs["teethSize"] != "8.0" {
		t.Errorf("expected zip teethSize 8.0, got %+v", sDetail.Zips[0].Attrs)
	}
	if sDetail.Trims[0].Attrs["cordEnd"] != "metal silver" {
		t.Errorf("expected trim cordEnd 'metal silver', got %+v", sDetail.Trims[0].Attrs)
	}
	if sDetail.Labels[0].Attrs["labelGenre"] != "Woven" {
		t.Errorf("expected labelGenre Woven, got %+v", sDetail.Labels[0].Attrs)
	}

	prodRes, err := svc.CreateMasterProduct(CreateProductPayload{
		ProductName:           "Tech Softshell Jacket Pro",
		Genre:                 "Jackets",
		BrandLabel:            "ETTU",
		GarmentSpecifications: "Full-zip front | Drawcord adjustable hood",
		Composition:           "95%Poly 5%SP",
		MRP:                   3499,
		SampleDevelopmentID:   sDetail.ID,
		Genders: []ProductGenderItem{
			{
				Gender:    "Men",
				Sizes:     []string{"S", "M", "L", "XL"},
				PharmaNos: []string{"PH-JKT-01-R1", "PH-JKT-02-R1"},
				ColorImages: []map[string]any{
					{"colorName": "Angora", "images": []string{"/uploads/m_angora_1.jpg"}},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateMasterProduct failed: %v", err)
	}
	prodDetail, err := svc.GetMasterProduct(prodRes["id"].(string))
	if err != nil {
		t.Fatalf("GetMasterProduct failed: %v", err)
	}
	g0 := prodDetail.Genders[0]
	if len(g0.Sizes) != 4 || g0.Sizes[2] != "L" {
		t.Errorf("expected 4 chosen sizes [S M L XL], got %+v", g0.Sizes)
	}
	if len(g0.PharmaNos) != 2 || g0.PharmaNos[1] != "PH-JKT-02-R1" {
		t.Errorf("expected 2 pharmaNos, got %+v", g0.PharmaNos)
	}
}
