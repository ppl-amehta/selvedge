package orders

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ppl-amehta/selvedge/internal/graph"
	"github.com/ppl-amehta/selvedge/internal/workflow"
)

type Service struct {
	db       *sql.DB
	graph    *graph.Store
	workflow *workflow.Engine
}

func NewService(db *sql.DB, graphStore *graph.Store, wf *workflow.Engine) *Service {
	return &Service{db: db, graph: graphStore, workflow: wf}
}

type CreateOrderPayload struct {
	OrderNo             string         `json:"orderNo"`
	ProductionOrderNo   string         `json:"productionOrderNo"`
	FpccOrderNo         string         `json:"fpccOrderNo"`
	SalesOrderNo        string         `json:"salesOrderNo"`
	OrderType           string         `json:"orderType"`
	BookingDate         string         `json:"bookingDate"`
	Date                string         `json:"date"`
	ClientID            string         `json:"clientId"`
	ClientName          string         `json:"clientName"`
	ClientContactPerson string         `json:"clientContactPerson"`
	ContactPerson       string         `json:"contactPerson"`
	Phone               string         `json:"phone"`
	Email               string         `json:"email"`
	BillingAddress      string         `json:"billingAddress"`
	ShippingAddress     string         `json:"shippingAddress"`
	Gstin               string         `json:"gstin"`
	GstNo               string         `json:"gstNo"`
	Category            string         `json:"category"`
	SubCategory         string         `json:"subCategory"`
	ProductName         string         `json:"productName"`
	ColorName           string         `json:"colorName"`
	FabricName          string         `json:"fabricName"`
	Gender              string         `json:"gender"`
	UnitPrice           float64        `json:"unitPrice"`
	UnitRate            float64        `json:"unitRate"`
	TaxRate             float64        `json:"taxRate"`
	TaxBilling          string         `json:"taxBilling"`
	FreightMode         string         `json:"freightMode"`
	FreightCost         string         `json:"freightCost"`
	TotalQty            int            `json:"totalQty"`
	SizeBreakdown       map[string]int `json:"sizeBreakdown"`
	MultipleShipping    any            `json:"multipleShipping"`
	ShippingLocations   []any          `json:"shippingLocations"`
	Personalization     []any          `json:"personalization"`
	BrandingType        string         `json:"brandingType"`
	BrandingFileUrl     string         `json:"brandingFileUrl"`
	BrandingFile        string         `json:"brandingFile"`
	MainLabel           string         `json:"mainLabel"`
	OtherLabel          string         `json:"otherLabel"`
	CustPackaging       string         `json:"custPackaging"`
	ClientComment       string         `json:"clientComment"`
	TargetEtd           string         `json:"targetEtd"`
	Etd                 string         `json:"etd"`
}

// SuggestNextOrderNo generates format: S{YYMMDD}{A}
func (s *Service) SuggestNextOrderNo() (string, error) {
	now := time.Now()
	prefix := fmt.Sprintf("S%s", now.Format("060102"))

	var lastOrderNo string
	err := s.db.QueryRow(`
		SELECT production_order_no FROM orders
		WHERE production_order_no LIKE ?
		ORDER BY production_order_no DESC LIMIT 1
	`, prefix+"%").Scan(&lastOrderNo)

	if err == sql.ErrNoRows {
		return prefix + "A", nil
	} else if err != nil {
		return prefix + "A", nil
	}

	// Extract trailing letter and increment
	nextChar := byte('A')
	if len(lastOrderNo) > len(prefix) {
		lastChar := lastOrderNo[len(prefix):]
		if len(lastChar) > 0 {
			nextChar = lastChar[0] + 1
		}
	}

	return fmt.Sprintf("%s%c", prefix, nextChar), nil
}

// CreateOrder books an order, updates projections, and writes to unified graph
func (s *Service) CreateOrder(p CreateOrderPayload) (map[string]any, error) {
	orderNo := strings.TrimSpace(p.OrderNo)
	if orderNo == "" {
		orderNo = strings.TrimSpace(p.ProductionOrderNo)
	}
	if orderNo == "" {
		orderNo = strings.TrimSpace(p.FpccOrderNo)
	}
	if orderNo == "" {
		var err error
		orderNo, err = s.SuggestNextOrderNo()
		if err != nil {
			return nil, err
		}
	}

	clientName := strings.TrimSpace(p.ClientName)
	productName := strings.TrimSpace(p.ProductName)
	colorName := strings.TrimSpace(p.ColorName)
	if clientName == "" || productName == "" || colorName == "" {
		return nil, fmt.Errorf("clientName, productName, and colorName are required")
	}

	contactPerson := strings.TrimSpace(p.ContactPerson)
	if contactPerson == "" {
		contactPerson = strings.TrimSpace(p.ClientContactPerson)
	}

	gstin := strings.TrimSpace(p.Gstin)
	if gstin == "" {
		gstin = strings.TrimSpace(p.GstNo)
	}

	unitPrice := p.UnitPrice
	if unitPrice == 0 {
		unitPrice = p.UnitRate
	}

	targetEtd := p.TargetEtd
	if targetEtd == "" {
		targetEtd = p.Etd
	}

	now := graph.NowMs()
	todayStr := time.Now().Format("2006-01-02")
	bookingDate := p.BookingDate
	if bookingDate == "" {
		bookingDate = p.Date
	}
	if bookingDate == "" {
		bookingDate = todayStr
	}

	brandingFile := p.BrandingFileUrl
	if brandingFile == "" {
		brandingFile = p.BrandingFile
	}

	multiShipping := 0
	switch v := p.MultipleShipping.(type) {
	case bool:
		if v {
			multiShipping = 1
		}
	case float64:
		if v > 0 {
			multiShipping = 1
		}
	case int:
		if v > 0 {
			multiShipping = 1
		}
	}

	sizeBreakdownJSON, _ := json.Marshal(p.SizeBreakdown)
	shippingLocsJSON, _ := json.Marshal(p.ShippingLocations)
	personalizationJSON, _ := json.Marshal(p.Personalization)

	// Transaction: Projection + Graph Node + Graph Edge + Event Outbox
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Resolve or Create Client Entity
	var clientID string
	err = tx.QueryRow("SELECT id FROM master_entities WHERE name = ? AND (entity_type = 'client' OR entity_type = 'both')", clientName).Scan(&clientID)
	if err == sql.ErrNoRows {
		clientID = fmt.Sprintf("ent_client_%d", now%100000)
		_, err = tx.Exec(`
			INSERT INTO master_entities (id, name, entity_type, contact_person, phone, email, gstin, address, created_at)
			VALUES (?, ?, 'client', ?, ?, ?, ?, ?, ?)
		`, clientID, clientName, contactPerson, p.Phone, p.Email, gstin, p.ShippingAddress, now)
		if err != nil {
			return nil, fmt.Errorf("failed to register client entity: %w", err)
		}
	}

	// 2. Insert into Orders Projection Table
	orderID := fmt.Sprintf("ord_%s", orderNo)
	query := `
		INSERT INTO orders (
			id, production_order_no, sales_order_no, order_type, booking_date,
			client_id, client_name, contact_person, phone, email, billing_address, shipping_address, gstin,
			category, sub_category, product_name, color_name, fabric_name, gender,
			unit_price, tax_rate, tax_billing, freight_mode, freight_cost,
			total_qty, size_breakdown_json, multiple_shipping, shipping_locations_json, personalization_json,
			branding_type, branding_file_url, main_label, other_label, cust_packaging,
			client_comment, target_etd, status, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, 'BOOKED', ?, ?
		)
	`
	_, err = tx.Exec(query,
		orderID, orderNo, p.SalesOrderNo, p.OrderType, bookingDate,
		clientID, clientName, contactPerson, p.Phone, p.Email, p.BillingAddress, p.ShippingAddress, gstin,
		p.Category, p.SubCategory, productName, colorName, p.FabricName, p.Gender,
		unitPrice, p.TaxRate, p.TaxBilling, p.FreightMode, p.FreightCost,
		p.TotalQty, string(sizeBreakdownJSON), multiShipping, string(shippingLocsJSON), string(personalizationJSON),
		p.BrandingType, brandingFile, p.MainLabel, p.OtherLabel, p.CustPackaging,
		p.ClientComment, targetEtd, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order projection: %w", err)
	}

	// 3. Insert Node in Unified Graph
	orderProps := map[string]any{
		"production_order_no": orderNo,
		"sales_order_no":      p.SalesOrderNo,
		"order_type":          p.OrderType,
		"product_name":        productName,
		"color_name":          colorName,
		"fabric_name":         p.FabricName,
		"total_qty":           p.TotalQty,
		"unit_price":          unitPrice,
		"size_breakdown":      p.SizeBreakdown,
		"target_etd":          targetEtd,
		"status":              "BOOKED",
	}
	_, err = s.graph.InsertNode(tx, graph.TypeOrder, orderProps, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert graph node for order: %w", err)
	}

	// 4. Insert Edge: (ORDER) -[:ORDERED_BY]-> (CLIENT)
	_, err = s.graph.InsertEdge(tx, orderID, clientID, graph.EdgeOrderedBy, map[string]any{
		"booking_date": bookingDate,
		"total_qty":    p.TotalQty,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert graph edge ORDERED_BY: %w", err)
	}

	// 5. Emit Event to Workflow Outbox
	_, err = s.workflow.EmitEvent(tx, orderID, "ORDER_BOOKED", map[string]any{
		"order_id":            orderID,
		"production_order_no": orderNo,
		"client_name":         clientName,
		"product_name":        productName,
		"total_qty":           p.TotalQty,
		"unit_price":          p.UnitPrice,
	}, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to emit ORDER_BOOKED event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Trigger async event dispatch
	go func() {
		_, _ = s.workflow.ProcessPendingEvents()
	}()

	return map[string]any{
		"success":           true,
		"productionOrderNo": orderNo,
		"id":                orderID,
	}, nil
}

// GetOrder retrieves an order by production_order_no
func (s *Service) GetOrder(orderNo string) (map[string]any, error) {
	var o struct {
		ID                  string
		OrderNo             string
		SalesOrderNo        sql.NullString
		OrderType           string
		BookingDate         string
		ClientID            string
		ClientName          string
		ContactPerson       sql.NullString
		Phone               sql.NullString
		Email               sql.NullString
		BillingAddress      sql.NullString
		ShippingAddress     sql.NullString
		Gstin               sql.NullString
		Category            string
		SubCategory         string
		ProductName         string
		ColorName           string
		FabricName          sql.NullString
		Gender              string
		UnitPrice           float64
		TaxRate             float64
		TaxBilling          string
		FreightMode         string
		FreightCost         string
		TotalQty            int
		SizeBreakdownStr    sql.NullString
		MultipleShipping    int
		ShippingLocsStr     sql.NullString
		PersonalizationStr  sql.NullString
		BrandingType        sql.NullString
		BrandingFileUrl     sql.NullString
		MainLabel           sql.NullString
		OtherLabel          sql.NullString
		CustPackaging       sql.NullString
		ClientComment       sql.NullString
		TargetEtd           sql.NullString
		Status              string
	}

	query := `
		SELECT id, production_order_no, sales_order_no, order_type, booking_date,
		       client_id, client_name, contact_person, phone, email, billing_address, shipping_address, gstin,
		       category, sub_category, product_name, color_name, fabric_name, gender,
		       unit_price, tax_rate, tax_billing, freight_mode, freight_cost,
		       total_qty, size_breakdown_json, multiple_shipping, shipping_locations_json, personalization_json,
		       branding_type, branding_file_url, main_label, other_label, cust_packaging,
		       client_comment, target_etd, status
		FROM orders WHERE production_order_no = ?
	`
	err := s.db.QueryRow(query, orderNo).Scan(
		&o.ID, &o.OrderNo, &o.SalesOrderNo, &o.OrderType, &o.BookingDate,
		&o.ClientID, &o.ClientName, &o.ContactPerson, &o.Phone, &o.Email, &o.BillingAddress, &o.ShippingAddress, &o.Gstin,
		&o.Category, &o.SubCategory, &o.ProductName, &o.ColorName, &o.FabricName, &o.Gender,
		&o.UnitPrice, &o.TaxRate, &o.TaxBilling, &o.FreightMode, &o.FreightCost,
		&o.TotalQty, &o.SizeBreakdownStr, &o.MultipleShipping, &o.ShippingLocsStr, &o.PersonalizationStr,
		&o.BrandingType, &o.BrandingFileUrl, &o.MainLabel, &o.OtherLabel, &o.CustPackaging,
		&o.ClientComment, &o.TargetEtd, &o.Status,
	)
	if err != nil {
		return nil, err
	}

	var sizeBreakdown map[string]int
	if o.SizeBreakdownStr.Valid {
		_ = json.Unmarshal([]byte(o.SizeBreakdownStr.String), &sizeBreakdown)
	}

	var shippingLocs []any
	if o.ShippingLocsStr.Valid {
		_ = json.Unmarshal([]byte(o.ShippingLocsStr.String), &shippingLocs)
	}

	var personalization []any
	if o.PersonalizationStr.Valid {
		_ = json.Unmarshal([]byte(o.PersonalizationStr.String), &personalization)
	}

	return map[string]any{
		"id":                  o.ID,
		"production_order_no": o.OrderNo,
		"sales_order_no":      o.SalesOrderNo.String,
		"order_type":          o.OrderType,
		"booking_date":        o.BookingDate,
		"client_id":           o.ClientID,
		"client_name":         o.ClientName,
		"contact_person":      o.ContactPerson.String,
		"phone":               o.Phone.String,
		"email":               o.Email.String,
		"billing_address":     o.BillingAddress.String,
		"shipping_address":    o.ShippingAddress.String,
		"gstin":               o.Gstin.String,
		"category":            o.Category,
		"sub_category":        o.SubCategory,
		"product_name":        o.ProductName,
		"color_name":          o.ColorName,
		"fabric_name":         o.FabricName.String,
		"gender":              o.Gender,
		"unit_price":          o.UnitPrice,
		"tax_rate":            o.TaxRate,
		"tax_billing":         o.TaxBilling,
		"freight_mode":        o.FreightMode,
		"freight_cost":        o.FreightCost,
		"total_qty":           o.TotalQty,
		"sizeBreakdown":       sizeBreakdown,
		"multiple_shipping":   o.MultipleShipping,
		"shippingLocations":   shippingLocs,
		"personalization":     personalization,
		"branding_type":       o.BrandingType.String,
		"branding_file_url":   o.BrandingFileUrl.String,
		"main_label":          o.MainLabel.String,
		"other_label":         o.OtherLabel.String,
		"cust_packaging":      o.CustPackaging.String,
		"client_comment":      o.ClientComment.String,
		"target_etd":          o.TargetEtd.String,
		"status":              o.Status,
	}, nil
}

// ListOrders returns recent orders
func (s *Service) ListOrders(limit, offset int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`
		SELECT id, production_order_no, sales_order_no, client_name, product_name, color_name, fabric_name, total_qty, unit_price, status, target_etd
		FROM orders ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []map[string]any
	for rows.Next() {
		var id, ordNo, soNo, client, prod, col, fab, status, etd string
		var qty int
		var price float64
		_ = rows.Scan(&id, &ordNo, &soNo, &client, &prod, &col, &fab, &qty, &price, &status, &etd)
		list = append(list, map[string]any{
			"id":                  id,
			"production_order_no": ordNo,
			"sales_order_no":      soNo,
			"client_name":         client,
			"product_name":        prod,
			"color_name":          col,
			"fabric_name":         fab,
			"total_qty":           qty,
			"price":               price,
			"status":              status,
			"target_etd":          etd,
		})
	}
	return list, nil
}
