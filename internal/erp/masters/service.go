package masters

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ppl-amehta/selvedge/internal/graph"
)

type Service struct {
	db    *sql.DB
	graph *graph.Store
}

func NewService(db *sql.DB, graphStore *graph.Store) *Service {
	return &Service{db: db, graph: graphStore}
}

type Entity struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	EntityType      string `json:"entity_type"`
	ContactPerson   string `json:"contact_person"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
	Gstin           string `json:"gstin"`
	Address         string `json:"address"`
	BillingAddress  string `json:"billing_address"`
	ShippingAddress string `json:"shipping_address"`
}

type MasterItem struct {
	ID          string         `json:"id"`
	Category    string         `json:"category"`
	Code        string         `json:"code"`
	DisplayName string         `json:"display_name"`
	Meta        map[string]any `json:"meta"`
}

// GetAllMasters returns all master dropdown items in one call
func (s *Service) GetAllMasters() (map[string]any, error) {
	// 1. Entities (Vendors & Clients)
	vendors, err := s.listEntities("vendor")
	if err != nil {
		vendors = make([]Entity, 0)
	}
	clients, err := s.listEntities("client")
	if err != nil {
		clients = make([]Entity, 0)
	}

	// 2. Fetch all master items in a single grouped query (single WAL read snapshot)
	categoriesMap := map[string][]MasterItem{
		"fabric":           make([]MasterItem, 0),
		"color":            make([]MasterItem, 0),
		"location":         make([]MasterItem, 0),
		"uom":              make([]MasterItem, 0),
		"product":          make([]MasterItem, 0),
		"zip":              make([]MasterItem, 0),
		"trim":             make([]MasterItem, 0),
		"label":            make([]MasterItem, 0),
		"packaging":        make([]MasterItem, 0),
		"product_category": make([]MasterItem, 0),
		"subcategory":      make([]MasterItem, 0),
		"gender":           make([]MasterItem, 0),
		"branding_type":    make([]MasterItem, 0),
		"inward_type":      make([]MasterItem, 0),
		"cust_packaging":   make([]MasterItem, 0),
		"order_process":    make([]MasterItem, 0),
		"size_scale":       make([]MasterItem, 0),
		"size":             make([]MasterItem, 0),
		"fabric_type":      make([]MasterItem, 0),
		"zip_type":         make([]MasterItem, 0),
		"zip_slider":       make([]MasterItem, 0),
		"zip_size":         make([]MasterItem, 0),
		"zip_closure":      make([]MasterItem, 0),
		"zip_finish":       make([]MasterItem, 0),
		"trim_type":        make([]MasterItem, 0),
		"label_brand":       make([]MasterItem, 0),
		"label_additional":  make([]MasterItem, 0),
		"label_genre":       make([]MasterItem, 0),
		"label_description": make([]MasterItem, 0),
		"washcare":          make([]MasterItem, 0),
		"packaging_type":    make([]MasterItem, 0),
		"packaging_size":    make([]MasterItem, 0),
		"composition":       make([]MasterItem, 0),
		"product_genre":     make([]MasterItem, 0),
	}

	rows, err := s.db.Query("SELECT id, category, code, display_name, coalesce(meta_json, '{}') FROM master_items ORDER BY display_name")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var m MasterItem
			var metaStr string
			if err := rows.Scan(&m.ID, &m.Category, &m.Code, &m.DisplayName, &metaStr); err == nil {
				m.Meta, _ = graph.DecodeProps(metaStr)
				if m.Meta == nil {
					m.Meta = make(map[string]any)
				}
				categoriesMap[m.Category] = append(categoriesMap[m.Category], m)
			}
		}
	}

	// 3. Gender to Exact Size Taxonomy (Authoritative Sheet 4 Ground Truth)
	genderSizes := map[string][]string{
		"Men":    {"XXS", "XS", "S", "M", "L", "XL", "2XL", "3XL", "4XL", "5XL", "Custom", "32", "34", "36", "38", "39", "40", "42", "44", "46", "48", "50", "52"},
		"Women":  {"XXS", "XS", "S", "M", "L", "XL", "2XL", "3XL", "4XL", "5XL", "Custom", "32", "34", "36", "38", "39", "40", "42", "44", "46", "48", "50", "52"},
		"Unisex": {"XXS", "XS", "S", "M", "L", "XL", "2XL", "3XL", "4XL", "5XL", "Custom", "32", "34", "36", "38", "39", "40", "42", "44", "46", "48", "50", "52"},
		"Infant": {"0-3", "3-6", "6-9", "9-12", "12-15", "15-18"},
		"Kids":   {"K16", "K18", "K20", "K22", "K24", "K26", "K28", "K30", "K32", "K34", "K36", "kids custom"},
		"Jr":     {"Jr-XS", "Jr-S", "Jr-M", "Jr-L", "Jr Custom"},
	}

	nonApparelSizes := map[string][]string{
		"Drinkwear":   {"100ml", "150ml", "200ml", "250ml", "300ml", "500ml", "750ml", "1Litre", "1.5 Litre", "2 Litre"},
		"Headgear":    {"Std", "Kids", "Large", "Small"},
		"Caps":        {"Std", "Kids", "Large", "Small"},
		"Bags":        {"Std"},
		"Stationary":  {"A4", "A3", "A5", "A6", "B5"},
		"Electronics": {"Std"},
		"Others":      {"Std"},
	}

	apparelGenres := []string{
		"Polo", "Tee", "Hoodies and Pullovers", "Jackets", "Bottoms", "Quarter zip",
		"Shirt", "Shorts", "Sweat Pants", "Sweat shirt", "Sweater", "Sweatshirt",
		"Tights", "Vest", "Dress", "Skorts", "Onsies", "Caps", "Traded apparel",
	}

	nonApparelGenres := []string{
		"Drinkwear", "Headgear", "Bags", "Stationary", "Electronics", "Others",
	}

	// 4. Recent Orders
	recentOrders, _ := s.listRecentOrders()
	if recentOrders == nil {
		recentOrders = make([]map[string]any, 0)
	}

	cat := func(key string) []MasterItem {
		if list := categoriesMap[key]; list != nil {
			return list
		}
		return make([]MasterItem, 0)
	}

	return map[string]any{
		"vendors":                 vendors,
		"clients":                 clients,
		"fabrics":                 cat("fabric"),
		"fabric_generics":         cat("fabric_generic"),
		"fabric_types":            cat("fabric_type"),
		"colors":                  cat("color"),
		"locations":               cat("location"),
		"uoms":                    cat("uom"),
		"products":                cat("product"),
		"zips":                    cat("zip"),
		"zip_types":               cat("zip_type"),
		"zip_sliders":             cat("zip_slider"),
		"zip_sizes":               cat("zip_size"),
		"zip_closures":            cat("zip_closure"),
		"zip_finishes":            cat("zip_finish"),
		"trims":                   cat("trim"),
		"trim_types":              cat("trim_type"),
		"trim_elastics":           cat("trim_elastic"),
		"trim_elastic_sizes":      cat("trim_elastic_size"),
		"trim_tapes":              cat("trim_tape"),
		"trim_tape_sizes":         cat("trim_tape_size"),
		"trim_cords":              cat("trim_cord"),
		"trim_cord_ends":          cat("trim_cord_end"),
		"trim_cord_sizes":         cat("trim_cord_size"),
		"trim_buttons":            cat("trim_button"),
		"trim_button_sizes":       cat("trim_button_size"),
		"trim_fusings":            cat("trim_fusing"),
		"trim_fusing_sizes":       cat("trim_fusing_size"),
		"trim_fusing_gsms":        cat("trim_fusing_gsm"),
		"trim_fusing_colors":      cat("trim_fusing_color"),
		"trim_threads":            cat("trim_thread"),
		"trim_thread_lengths":     cat("trim_thread_length"),
		"trim_thread_sizes":       cat("trim_thread_size"),
		"trim_other_colors":       cat("trim_other_color"),
		"trim_other_sizes":        cat("trim_other_size"),
		"labels":                  cat("label"),
		"label_brands":            cat("label_brand"),
		"label_additionals":       cat("label_additional"),
		"label_genres":            cat("label_genre"),
		"label_descriptions":      cat("label_description"),
		"washcare":                cat("washcare"),
		"packaging":               cat("packaging"),
		"packaging_types":         cat("packaging_type"),
		"packaging_sizes":         cat("packaging_size"),
		"packaging_polybag_brands": cat("packaging_polybag_brand"),
		"packaging_polybag_sizes": cat("packaging_polybag_size"),
		"packaging_tape_variants": cat("packaging_tape_variant"),
		"packaging_tape_sizes":    cat("packaging_tape_size"),
		"packaging_loop_materials": cat("packaging_loop_material"),
		"packaging_loop_sizes":    cat("packaging_loop_size"),
		"packaging_carton_brands": cat("packaging_carton_brand"),
		"packaging_carton_sizes":  cat("packaging_carton_size"),
		"packaging_shrink_sizes":  cat("packaging_shrink_size"),
		"packaging_courier_brands": cat("packaging_courier_brand"),
		"packaging_courier_sizes": cat("packaging_courier_size"),
		"product_categories":      cat("product_category"),
		"subcategories":           cat("subcategory"),
		"genders":                 cat("gender"),
		"branding_types":          cat("branding_type"),
		"inward_types":            cat("inward_type"),
		"cust_packaging":          cat("cust_packaging"),
		"order_processes":         cat("order_process"),
		"size_scales":             cat("size_scale"),
		"sizes":                   cat("size"),
		"gender_sizes":            genderSizes,
		"non_apparel_sizes":       nonApparelSizes,
		"compositions":            cat("composition"),
		"product_genres":          cat("product_genre"),
		"apparel_genres":          apparelGenres,
		"non_apparel_genres":      nonApparelGenres,
		"orders":                  recentOrders,
	}, nil
}

func (s *Service) listEntities(entityType string) ([]Entity, error) {
	query := "SELECT id, name, entity_type, coalesce(contact_person, ''), coalesce(phone, ''), coalesce(email, ''), coalesce(gstin, ''), coalesce(address, ''), coalesce(billing_address, ''), coalesce(shipping_address, '') FROM master_entities WHERE entity_type = ? OR entity_type = 'both' ORDER BY name"
	rows, err := s.db.Query(query, entityType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]Entity, 0)
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.Name, &e.EntityType, &e.ContactPerson, &e.Phone, &e.Email, &e.Gstin, &e.Address, &e.BillingAddress, &e.ShippingAddress); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, nil
}

func (s *Service) ListMasterItems(category string) ([]MasterItem, error) {
	query := "SELECT id, category, code, display_name, coalesce(meta_json, '{}') FROM master_items WHERE category = ? ORDER BY display_name"
	rows, err := s.db.Query(query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]MasterItem, 0)
	for rows.Next() {
		var m MasterItem
		var metaStr string
		if err := rows.Scan(&m.ID, &m.Category, &m.Code, &m.DisplayName, &metaStr); err != nil {
			return nil, err
		}
		m.Meta, _ = graph.DecodeProps(metaStr)
		if m.Meta == nil {
			m.Meta = make(map[string]any)
		}
		list = append(list, m)
	}
	return list, nil
}

func (s *Service) listRecentOrders() ([]map[string]any, error) {
	rows, err := s.db.Query(`
		SELECT production_order_no, sales_order_no, client_name, product_name, color_name, total_qty, unit_price, status
		FROM orders ORDER BY created_at DESC LIMIT 50
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]map[string]any, 0)
	for rows.Next() {
		var ordNo, soNo, client, prod, col, status string
		var qty int
		var price float64
		_ = rows.Scan(&ordNo, &soNo, &client, &prod, &col, &qty, &price, &status)
		list = append(list, map[string]any{
			"production_order_no": ordNo,
			"sales_order_no":      soNo,
			"client_name":         client,
			"product_name":        prod,
			"color_name":          col,
			"total_qty":           qty,
			"price":               price,
			"status":              status,
		})
	}
	return list, nil
}

// RegisterMaster registers a newly typed open-ended entry into both projections and unified graph
func (s *Service) RegisterMaster(category, value string, meta map[string]any) (map[string]any, error) {
	cleanVal := strings.TrimSpace(value)
	if cleanVal == "" {
		return nil, fmt.Errorf("value cannot be empty")
	}

	now := graph.NowMs()
	cat := strings.ToLower(category)

	switch cat {
	case "vendor", "client":
		var id string
		err := s.db.QueryRow("SELECT id FROM master_entities WHERE name = ? AND (entity_type = ? OR entity_type = 'both')", cleanVal, cat).Scan(&id)
		if err == nil {
			return map[string]any{"id": id, "name": cleanVal, "existed": true}, nil
		}

		id = fmt.Sprintf("ent_%s_%d", cat, now%100000)
		contact, _ := meta["contactPerson"].(string)
		phone, _ := meta["phone"].(string)
		email, _ := meta["email"].(string)
		gstin, _ := meta["gstin"].(string)
		address, _ := meta["address"].(string)
		billing, _ := meta["billingAddress"].(string)
		if billing == "" {
			billing = address
		}
		shipping, _ := meta["shippingAddress"].(string)
		if shipping == "" {
			shipping = address
		}

		_, err = s.db.Exec(`
			INSERT INTO master_entities (id, name, entity_type, contact_person, phone, email, gstin, address, billing_address, shipping_address, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, id, cleanVal, cat, contact, phone, email, gstin, address, billing, shipping, now)
		if err != nil {
			return nil, err
		}

		// Also register node in unified graph
		nodeType := graph.TypeVendor
		if cat == "client" {
			nodeType = graph.TypeClient
		}
		_, _ = s.graph.InsertNode(nil, nodeType, map[string]any{
			"name":           cleanVal,
			"contact_person": contact,
			"phone":          phone,
			"gstin":          gstin,
		}, id)

		return map[string]any{"id": id, "name": cleanVal, "created": true}, nil

	default:
		// Master item (fabric, color, product, location, uom)
		var id string
		err := s.db.QueryRow("SELECT id FROM master_items WHERE category = ? AND (code = ? OR display_name = ?)", cat, cleanVal, cleanVal).Scan(&id)
		if err == nil {
			return map[string]any{"id": id, "display_name": cleanVal, "existed": true}, nil
		}

		id = fmt.Sprintf("item_%s_%d", cat, now%100000)
		metaBytes, _ := json.Marshal(meta)
		_, err = s.db.Exec(`
			INSERT INTO master_items (id, category, code, display_name, meta_json, created_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, id, cat, cleanVal, cleanVal, string(metaBytes), now)
		if err != nil {
			return nil, err
		}

		// Also register node in unified graph
		_, _ = s.graph.InsertNode(nil, strings.ToUpper(cat), map[string]any{
			"code":         cleanVal,
			"display_name": cleanVal,
			"meta":         meta,
		}, id)

		return map[string]any{"id": id, "display_name": cleanVal, "created": true}, nil
	}
}
