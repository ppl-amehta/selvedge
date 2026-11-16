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
	ID            string `json:"id"`
	Name          string `json:"name"`
	EntityType    string `json:"entity_type"`
	ContactPerson string `json:"contact_person"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Gstin         string `json:"gstin"`
	Address       string `json:"address"`
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
		"label_brand":      make([]MasterItem, 0),
		"label_additional": make([]MasterItem, 0),
		"label_genre":      make([]MasterItem, 0),
		"washcare":         make([]MasterItem, 0),
		"packaging_type":   make([]MasterItem, 0),
		"packaging_size":   make([]MasterItem, 0),
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

	// 3. Recent Orders
	recentOrders, _ := s.listRecentOrders()
	if recentOrders == nil {
		recentOrders = make([]map[string]any, 0)
	}

	return map[string]any{
		"vendors":            vendors,
		"clients":            clients,
		"fabrics":            categoriesMap["fabric"],
		"colors":             categoriesMap["color"],
		"locations":          categoriesMap["location"],
		"uoms":               categoriesMap["uom"],
		"products":           categoriesMap["product"],
		"zips":               categoriesMap["zip"],
		"trims":              categoriesMap["trim"],
		"labels":             categoriesMap["label"],
		"packaging":          categoriesMap["packaging"],
		"product_categories": categoriesMap["product_category"],
		"subcategories":      categoriesMap["subcategory"],
		"genders":            categoriesMap["gender"],
		"branding_types":     categoriesMap["branding_type"],
		"inward_types":       categoriesMap["inward_type"],
		"cust_packaging":     categoriesMap["cust_packaging"],
		"order_processes":    categoriesMap["order_process"],
		"size_scales":        categoriesMap["size_scale"],
		"sizes":              categoriesMap["size"],
		"fabric_types":       categoriesMap["fabric_type"],
		"zip_types":          categoriesMap["zip_type"],
		"zip_sliders":        categoriesMap["zip_slider"],
		"zip_sizes":          categoriesMap["zip_size"],
		"zip_closures":       categoriesMap["zip_closure"],
		"zip_finishes":       categoriesMap["zip_finish"],
		"trim_types":         categoriesMap["trim_type"],
		"label_brands":       categoriesMap["label_brand"],
		"label_additionals":  categoriesMap["label_additional"],
		"label_genres":       categoriesMap["label_genre"],
		"washcare":           categoriesMap["washcare"],
		"packaging_types":    categoriesMap["packaging_type"],
		"packaging_sizes":    categoriesMap["packaging_size"],
		"orders":             recentOrders,
	}, nil
}

func (s *Service) listEntities(entityType string) ([]Entity, error) {
	query := "SELECT id, name, entity_type, coalesce(contact_person, ''), coalesce(phone, ''), coalesce(email, ''), coalesce(gstin, ''), coalesce(address, '') FROM master_entities WHERE entity_type = ? OR entity_type = 'both' ORDER BY name"
	rows, err := s.db.Query(query, entityType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]Entity, 0)
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.Name, &e.EntityType, &e.ContactPerson, &e.Phone, &e.Email, &e.Gstin, &e.Address); err != nil {
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

		_, err = s.db.Exec(`
			INSERT INTO master_entities (id, name, entity_type, contact_person, phone, email, gstin, address, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, id, cleanVal, cat, contact, phone, email, gstin, address, now)
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
