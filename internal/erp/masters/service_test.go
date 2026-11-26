package masters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ppl-amehta/selvedge/internal/graph"
	"github.com/ppl-amehta/selvedge/internal/platform/db"
)

func TestMasterCatalogConsolidation(t *testing.T) {
	testDbPath := filepath.Join(os.TempDir(), "test_selvedge_masters.db")
	_ = os.Remove(testDbPath)
	defer os.Remove(testDbPath)

	database, err := db.Connect(testDbPath)
	if err != nil {
		t.Fatalf("failed to connect db: %v", err)
	}
	defer database.Close()

	// Apply migrations
	migrations := []string{
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
	for _, mFile := range migrations {
		content, err := os.ReadFile(mFile)
		if err != nil {
			t.Fatalf("failed to read migration %s: %v", mFile, err)
		}
		if err := database.RunMigrations(mFile, string(content)); err != nil {
			t.Fatalf("migration %s failed: %v", mFile, err)
		}
	}

	graphStore := graph.NewStore(database.DB)
	mastersService := NewService(database.DB, graphStore)

	masters, err := mastersService.GetAllMasters()
	if err != nil {
		t.Fatalf("GetAllMasters failed: %v", err)
	}

	// Verify non-null array responses for core master categories
	requiredKeys := []string{
		"vendors", "clients", "fabrics", "colors", "locations", "uoms", "products",
		"zips", "trims", "labels", "packaging", "product_categories", "subcategories",
		"genders", "branding_types", "inward_types", "cust_packaging", "order_processes",
		"size_scales", "sizes", "fabric_types", "zip_types", "packaging_types",
		"compositions", "gender_sizes", "non_apparel_sizes", "apparel_genres", "non_apparel_genres",
	}

	for _, k := range requiredKeys {
		val, exists := masters[k]
		if !exists || val == nil {
			t.Fatalf("expected key %s to exist in GetAllMasters response", k)
		}
	}

	// Verify specific counts
	uoms := masters["uoms"].([]MasterItem)
	if len(uoms) < 11 {
		t.Errorf("expected >= 11 uoms, got %d", len(uoms))
	}

	locations := masters["locations"].([]MasterItem)
	if len(locations) < 200 {
		t.Errorf("expected >= 200 locations, got %d", len(locations))
	}

	genders := masters["genders"].([]MasterItem)
	if len(genders) < 6 {
		t.Errorf("expected >= 6 genders, got %d", len(genders))
	}

	inwardTypes := masters["inward_types"].([]MasterItem)
	if len(inwardTypes) < 12 {
		t.Errorf("expected >= 12 inward types, got %d", len(inwardTypes))
	}

	vendors := masters["vendors"].([]Entity)
	if len(vendors) < 60 {
		t.Errorf("expected >= 60 vendors, got %d", len(vendors))
	}

	clients := masters["clients"].([]Entity)
	if len(clients) < 80 {
		t.Errorf("expected >= 80 clients, got %d", len(clients))
	}

	trimElastics := masters["trim_elastics"].([]MasterItem)
	if len(trimElastics) < 7 {
		t.Errorf("expected >= 7 trim_elastics from Excel, got %d", len(trimElastics))
	}
	packagingPolybagSizes := masters["packaging_polybag_sizes"].([]MasterItem)
	if len(packagingPolybagSizes) < 6 {
		t.Errorf("expected >= 6 packaging_polybag_sizes from Excel, got %d", len(packagingPolybagSizes))
	}
	cartonSizes := masters["packaging_carton_sizes"].([]MasterItem)
	for _, item := range cartonSizes {
		if item.DisplayName == "Carton" || item.DisplayName == "DHL" || item.DisplayName == "FPCC" || item.DisplayName == "blank" {
			t.Errorf("expected misplaced label %q cleaned from packaging_carton_sizes", item.DisplayName)
		}
	}

	t.Logf("✓ Successfully validated consolidated masters: %d vendors, %d clients, %d locations, %d UOMs, %d genders, %d inward types",
		len(vendors), len(clients), len(locations), len(uoms), len(genders), len(inwardTypes))
}
