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
		return nil, err
	}
	clients, err := s.listEntities("client")
	if err != nil {
		return nil, err
	}

	// 2. Master items by category
	fabrics, _ := s.listMasterItems("fabric")
	colors, _ := s.listMasterItems("color")
	locations, _ := s.listMasterItems("location")
	uoms, _ := s.listMasterItems("uom")
	products, _ := s.listMasterItems("product")
	zips, _ := s.listMasterItems("zip")
	trims, _ := s.listMasterItems("trim")
	labels, _ := s.listMasterItems("label")
	packaging, _ := s.listMasterItems("packaging")

	// 3. Recent Orders
	recentOrders, _ := s.listRecentOrders()

	return map[string]any{
		"vendors":   vendors,
		"clients":   clients,
		"fabrics":   fabrics,
		"colors":    colors,
		"locations": locations,
		"uoms":      uoms,
		"products":  products,
		"zips":      zips,
		"trims":     trims,
		"labels":    labels,
		"packaging": packaging,
		"orders":    recentOrders,
	}, nil
}

func (s *Service) listEntities(entityType string) ([]Entity, error) {
	query := "SELECT id, name, entity_type, coalesce(contact_person, ''), coalesce(phone, ''), coalesce(email, ''), coalesce(gstin, ''), coalesce(address, '') FROM master_entities WHERE entity_type = ? OR entity_type = 'both' ORDER BY name"
	rows, err := s.db.Query(query, entityType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Entity
	for rows.Next() {
		var e Entity
		if err := rows.Scan(&e.ID, &e.Name, &e.EntityType, &e.ContactPerson, &e.Phone, &e.Email, &e.Gstin, &e.Address); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, nil
}

func (s *Service) listMasterItems(category string) ([]MasterItem, error) {
	query := "SELECT id, category, code, display_name, coalesce(meta_json, '{}') FROM master_items WHERE category = ? ORDER BY display_name"
	rows, err := s.db.Query(query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []MasterItem
	for rows.Next() {
		var m MasterItem
		var metaStr string
		if err := rows.Scan(&m.ID, &m.Category, &m.Code, &m.DisplayName, &metaStr); err != nil {
			return nil, err
		}
		m.Meta, _ = graph.DecodeProps(metaStr)
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

	var list []map[string]any
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
