package orders

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/ppl-amehta/selvedge/internal/graph"
	"github.com/ppl-amehta/selvedge/internal/workflow"
)

var (
	reAlphaOnly   = regexp.MustCompile(`^[A-Za-z\s.'-]+$`)
	reNumericOnly = regexp.MustCompile(`^[0-9]+$`)
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
	OrderNo             string                 `json:"orderNo"`
	ProductionOrderNo   string                 `json:"productionOrderNo"`
	FpccOrderNo         string                 `json:"fpccOrderNo"`
	SalesOrderNo        string                 `json:"salesOrderNo"`
	OrderType           string                 `json:"orderType"`
	BookingDate         string                 `json:"bookingDate"`
	Date                string                 `json:"date"`
	ClientID            string                 `json:"clientId"`
	ClientName          string                 `json:"clientName"`
	ClientContactPerson string                 `json:"clientContactPerson"`
	ContactPerson       string                 `json:"contactPerson"`
	Phone               string                 `json:"phone"`
	Email               string                 `json:"email"`
	BillingAddress      string                 `json:"billingAddress"`
	ShippingAddress     string                 `json:"shippingAddress"`
	Gstin               string                 `json:"gstin"`
	GstNo               string                 `json:"gstNo"`
	TaxRate             float64                `json:"taxRate"`
	TaxBilling          string                 `json:"taxBilling"`
	FreightMode         string                 `json:"freightMode"`
	FreightCost         string                 `json:"freightCost"`
	SourceChannel       string                 `json:"sourceChannel"`
	ClientComment       string                 `json:"clientComment"`
	TargetEtd           string                 `json:"targetEtd"`
	Etd                 string                 `json:"etd"`

	// Multi-Product Line Items
	Items []OrderItemPayload `json:"items"`

	// Multi-Shipping Destinations
	ShippingLocations []ShippingDestPayload `json:"shippingLocations"`
	MultipleShipping  any                   `json:"multipleShipping"`

	// Legacy single-item fields (for backward compatibility)
	Category        string                   `json:"category"`
	SubCategory     string                   `json:"subCategory"`
	ProductName     string                   `json:"productName"`
	ColorName       string                   `json:"colorName"`
	FabricName      string                   `json:"fabricName"`
	Gender          string                   `json:"gender"`
	UnitPrice       float64                  `json:"unitPrice"`
	UnitRate        float64                  `json:"unitRate"`
	TotalQty        int                      `json:"totalQty"`
	SizeBreakdown   map[string]int           `json:"sizeBreakdown"`
	Personalization []PersonalizationPayload `json:"personalization"`
	BrandingType    string                   `json:"brandingType"`
	BrandingFileUrl string                   `json:"brandingFileUrl"`
	BrandingFile    string                   `json:"brandingFile"`
	MainLabel       string                   `json:"mainLabel"`
	OtherLabel      string                   `json:"otherLabel"`
	CustPackaging   string                   `json:"custPackaging"`
}

type OrderItemPayload struct {
	ID               string                   `json:"id"`
	ItemSequence     int                      `json:"itemSequence"`
	Category         string                   `json:"category"`
	SubCategory      string                   `json:"subCategory"`
	ProductName      string                   `json:"productName"`
	ColorName        string                   `json:"colorName"`
	FabricName       string                   `json:"fabricName"`
	UnitPrice        float64                  `json:"unitPrice"`
	TotalQty         int                      `json:"totalQty"`
	ClientEtd        string                   `json:"clientEtd"`
	ClientRemarks    string                   `json:"clientRemarks"`
	ProductImageUrl  string                   `json:"productImageUrl"`
	MainLabel        string                   `json:"mainLabel"`
	OtherLabel       string                   `json:"otherLabel"`
	CustPackaging    string                   `json:"custPackaging"`
	Matrix           any                      `json:"matrix"`
	Brandings        []BrandingPayload        `json:"brandings"`
	Personalizations []PersonalizationPayload `json:"personalizations"`
}

type BrandingPayload struct {
	ID              string `json:"id"`
	BrandingType    string `json:"brandingType"` // 'emb', 'print', 'dtf', 'unik', 'other'
	Position        string `json:"position"`
	Size            string `json:"size"`
	Placement       string `json:"placement"`
	Color           string `json:"color"`
	Comment         string `json:"comment"`
	BrandingFileURL string `json:"brandingFileUrl"`
}

type PersonalizationPayload struct {
	ID      string `json:"id"`
	Gender  string `json:"gender"`
	Size    string `json:"size"`
	Name    string `json:"name"`
	Number  string `json:"number"`
	Remarks string `json:"remarks"`
}

type ShippingDestPayload struct {
	ID                 string `json:"id"`
	OrderItemID        string `json:"orderItemId"`
	DestinationAddress string `json:"destinationAddress"`
	Destination        string `json:"destination"` // legacy fallback
	Quantity           int    `json:"quantity"`
	FreightMode        string `json:"freightMode"`
	Carrier            string `json:"carrier"`
	DocketNo           string `json:"docketNo"`
	TargetEtd          string `json:"targetEtd"`
}

// SuggestNextOrderNo generates format: S{YYMMDD}{A..Z}
func (s *Service) SuggestNextOrderNo() (string, error) {
	now := time.Now()
	prefix := fmt.Sprintf("S%s", now.Format("060102"))

	var lastOrderNo string
	err := s.db.QueryRow(`
		SELECT production_order_no FROM orders
		WHERE production_order_no LIKE ?
		ORDER BY production_order_no DESC LIMIT 1
	`, prefix+"%").Scan(&lastOrderNo)

	if errors.Is(err, sql.ErrNoRows) {
		return prefix + "A", nil
	} else if err != nil {
		return "", fmt.Errorf("database query failed for next order no: %w", err)
	}

	suffix := strings.TrimPrefix(lastOrderNo, prefix)
	if len(suffix) == 0 {
		return prefix + "A", nil
	}

	nextSuffix := incrementSuffix(suffix)
	return prefix + nextSuffix, nil
}

func incrementSuffix(s string) string {
	b := []byte(s)
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] < 'Z' {
			b[i]++
			return string(b)
		}
		b[i] = 'A'
	}
	return "A" + string(b)
}

// computeMatrixSum recursively traverses map[string]any or map[string]int to calculate sum
func computeMatrixSum(matrix any) (int, error) {
	if matrix == nil {
		return 0, nil
	}
	total := 0
	switch m := matrix.(type) {
	case map[string]int:
		for _, v := range m {
			if v < 0 {
				return 0, fmt.Errorf("negative quantity not allowed: %d", v)
			}
			total += v
		}
	case map[string]float64:
		for _, v := range m {
			if v < 0 || float64(int(v)) != v {
				return 0, fmt.Errorf("fractional or negative quantity not allowed: %v", v)
			}
			total += int(v)
		}
	case map[string]any:
		for _, v := range m {
			switch val := v.(type) {
			case float64:
				if val < 0 || float64(int(val)) != val {
					return 0, fmt.Errorf("fractional or negative quantity not allowed: %v", val)
				}
				total += int(val)
			case int:
				if val < 0 {
					return 0, fmt.Errorf("negative quantity not allowed: %d", val)
				}
				total += val
			case map[string]any:
				sub, err := computeMatrixSum(val)
				if err != nil {
					return 0, err
				}
				total += sub
			}
		}
	}
	return total, nil
}

// CreateOrder books an order with multi-product items, runs math lock-up checker,
// writes bi-temporal graph nodes and edges, and registers workflow events.
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
	if clientName == "" {
		return nil, fmt.Errorf("clientName is required")
	}

	contactPerson := strings.TrimSpace(p.ContactPerson)
	if contactPerson == "" {
		contactPerson = strings.TrimSpace(p.ClientContactPerson)
	}

	gstin := strings.TrimSpace(p.Gstin)
	if gstin == "" {
		gstin = strings.TrimSpace(p.GstNo)
	}

	targetEtd := p.TargetEtd
	if targetEtd == "" {
		targetEtd = p.Etd
	}

	now := time.Now()
	nowMs := now.UnixNano() / int64(time.Millisecond)
	todayStr := now.Format("2006-01-02")
	bookingDate := p.BookingDate
	if bookingDate == "" {
		bookingDate = p.Date
	}
	if bookingDate == "" {
		bookingDate = todayStr
	}

	// 1. Normalize Items: fallback to legacy fields if items slice is empty
	items := p.Items
	if len(items) == 0 && p.ProductName != "" {
		legacyMatrix := make(map[string]any)
		for k, v := range p.SizeBreakdown {
			legacyMatrix[k] = v
		}
		var legacyBrandings []BrandingPayload
		if p.BrandingType != "" {
			legacyBrandings = append(legacyBrandings, BrandingPayload{
				ID:              fmt.Sprintf("bnd_%d_%d", nowMs, rand.Intn(10000)),
				BrandingType:    "emb",
				Placement:       p.BrandingType,
				BrandingFileURL: p.BrandingFileUrl,
			})
		}
		items = append(items, OrderItemPayload{
			ID:               fmt.Sprintf("item_%s_1", orderNo),
			ItemSequence:     1,
			Category:         p.Category,
			SubCategory:      p.SubCategory,
			ProductName:      p.ProductName,
			ColorName:        p.ColorName,
			FabricName:       p.FabricName,
			UnitPrice:        p.UnitPrice,
			TotalQty:         p.TotalQty,
			ClientEtd:        targetEtd,
			ClientRemarks:    p.ClientComment,
			MainLabel:        p.MainLabel,
			OtherLabel:       p.OtherLabel,
			CustPackaging:    p.CustPackaging,
			Matrix:           legacyMatrix,
			Brandings:        legacyBrandings,
			Personalizations: p.Personalization,
		})
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("order must contain at least one product item")
	}

	// 2. Strict Math Checker & Personalization Rule Verification
	totalOrderQty := 0
	totalOrderValue := 0.0

	for i, item := range items {
		if strings.TrimSpace(item.ProductName) == "" {
			return nil, fmt.Errorf("item #%d product name is required", i+1)
		}
		if item.TotalQty <= 0 {
			return nil, fmt.Errorf("item #%d (%s) total quantity must be greater than 0", i+1, item.ProductName)
		}

		// Math Checker: Matrix sum must strictly equal TotalQty
		matrixSum, err := computeMatrixSum(item.Matrix)
		if err != nil {
			return nil, fmt.Errorf("item #%d (%s) matrix error: %w", i+1, item.ProductName, err)
		}
		if matrixSum > 0 && matrixSum != item.TotalQty {
			return nil, fmt.Errorf("math checker violation on item #%d (%s): matrix sum (%d) does not match total quantity (%d). action locked",
				i+1, item.ProductName, matrixSum, item.TotalQty)
		}

		// Personalization rules: Only 1 alphabetical Name and 1 numerical Number
		for j, pers := range item.Personalizations {
			name := strings.TrimSpace(pers.Name)
			number := strings.TrimSpace(pers.Number)
			if name != "" && !reAlphaOnly.MatchString(name) {
				return nil, fmt.Errorf("item #%d personalization row #%d has invalid name %q: must contain alphabetical characters only",
					i+1, j+1, name)
			}
			if number != "" && !reNumericOnly.MatchString(number) {
				return nil, fmt.Errorf("item #%d personalization row #%d has invalid number %q: must contain numerical digits only",
					i+1, j+1, number)
			}
		}

		// Branding validation: type must be emb, print, dtf, unik, or other
		for j, bnd := range item.Brandings {
			bType := strings.ToLower(strings.TrimSpace(bnd.BrandingType))
			if bType == "" {
				bType = "other"
			}
			switch bType {
			case "emb", "print", "dtf", "unik", "other":
				items[i].Brandings[j].BrandingType = bType
			default:
				items[i].Brandings[j].BrandingType = "other"
			}
		}

		totalOrderQty += item.TotalQty
		totalOrderValue += float64(item.TotalQty) * item.UnitPrice
	}

	// Calculate grand commercial valuation
	grandValuation := totalOrderValue
	if p.TaxBilling == "Extra" {
		grandValuation += totalOrderValue * (p.TaxRate / 100.0)
	}

	sourceChannel := p.SourceChannel
	if sourceChannel == "" {
		sourceChannel = "manual"
	}

	// Transaction: Atomic Write across Projections, Graph Nodes, and Outbox Events
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Resolve or Register Client Entity
	var clientID string
	err = tx.QueryRow("SELECT id FROM master_entities WHERE name = ? AND (entity_type = 'client' OR entity_type = 'both') LIMIT 1", clientName).Scan(&clientID)
	if errors.Is(err, sql.ErrNoRows) {
		clientID = fmt.Sprintf("ent_client_%d_%d", nowMs, rand.Intn(10000))
		_, err = tx.Exec(`
			INSERT INTO master_entities (id, name, entity_type, contact_person, phone, email, gstin, address, created_at)
			VALUES (?, ?, 'client', ?, ?, ?, ?, ?, ?)
		`, clientID, clientName, contactPerson, p.Phone, p.Email, gstin, p.ShippingAddress, nowMs)
		if err != nil {
			return nil, fmt.Errorf("failed to register client entity: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to query client entity: %w", err)
	}

	// 2. Insert Order Header Projection
	orderID := fmt.Sprintf("ord_%s", orderNo)
	firstItem := items[0]
	displayProductName := firstItem.ProductName
	if len(items) > 1 {
		displayProductName = fmt.Sprintf("%s + %d item(s)", firstItem.ProductName, len(items)-1)
	}

	firstMatrixJSON, err := json.Marshal(firstItem.Matrix)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal first item matrix: %w", err)
	}
	shippingLocsJSON, err := json.Marshal(p.ShippingLocations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal shipping locations: %w", err)
	}

	orderQuery := `
		INSERT INTO orders (
			id, production_order_no, sales_order_no, order_type, booking_date,
			client_id, client_name, contact_person, phone, email, billing_address, shipping_address, gstin,
			category, sub_category, product_name, color_name, fabric_name, gender,
			unit_price, tax_rate, tax_billing, freight_mode, freight_cost,
			total_qty, size_breakdown_json, multiple_shipping, shipping_locations_json, personalization_json,
			branding_type, branding_file_url, main_label, other_label, cust_packaging,
			client_comment, target_etd, status, source_channel, total_order_value, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, 'BOOKED', ?, ?, ?, ?
		)
	`
	_, err = tx.Exec(orderQuery,
		orderID, orderNo, p.SalesOrderNo, p.OrderType, bookingDate,
		clientID, clientName, contactPerson, p.Phone, p.Email, p.BillingAddress, p.ShippingAddress, gstin,
		firstItem.Category, firstItem.SubCategory, displayProductName, firstItem.ColorName, firstItem.FabricName, "Multi",
		firstItem.UnitPrice, p.TaxRate, p.TaxBilling, p.FreightMode, p.FreightCost,
		totalOrderQty, string(firstMatrixJSON), len(p.ShippingLocations) > 1, string(shippingLocsJSON), "[]",
		firstItem.MainLabel, firstItem.ProductImageUrl, firstItem.MainLabel, firstItem.OtherLabel, firstItem.CustPackaging,
		p.ClientComment, targetEtd, sourceChannel, grandValuation, nowMs, nowMs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order projection: %w", err)
	}

	// 3. Insert into Unified Graph (ORDER Node)
	orderNodeProps := map[string]any{
		"production_order_no": orderNo,
		"sales_order_no":      p.SalesOrderNo,
		"order_type":          p.OrderType,
		"client_name":         clientName,
		"total_qty":           totalOrderQty,
		"total_order_value":   grandValuation,
		"source_channel":      sourceChannel,
		"item_count":          len(items),
		"status":              "BOOKED",
	}
	_, err = s.graph.InsertNode(tx, graph.TypeOrder, orderNodeProps, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order graph node: %w", err)
	}

	// Edge: (ORDER) -[:ORDERED_BY]-> (CLIENT)
	_, err = s.graph.InsertEdge(tx, orderID, clientID, graph.EdgeOrderedBy, map[string]any{
		"booking_date": bookingDate,
		"total_qty":    totalOrderQty,
		"total_value":  grandValuation,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert graph edge ORDERED_BY: %w", err)
	}

	// 4. Insert Product Line Items, Brandings, and Personalizations
	for idx, item := range items {
		itemID := item.ID
		if itemID == "" {
			itemID = fmt.Sprintf("item_%s_%d_%d", orderNo, idx+1, nowMs%10000)
		}
		itemMatrixJSON, err := json.Marshal(item.Matrix)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal item #%d matrix: %w", idx+1, err)
		}

		itemQuery := `
			INSERT INTO order_items (
				id, order_id, item_sequence, category, sub_category, product_name,
				color_name, fabric_name, unit_price, total_qty, client_etd, client_remarks,
				product_image_url, main_label, other_label, cust_packaging, matrix_json, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err = tx.Exec(itemQuery,
			itemID, orderID, idx+1, item.Category, item.SubCategory, item.ProductName,
			item.ColorName, item.FabricName, item.UnitPrice, item.TotalQty, item.ClientEtd, item.ClientRemarks,
			item.ProductImageUrl, item.MainLabel, item.OtherLabel, item.CustPackaging, string(itemMatrixJSON), nowMs,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert order item #%d: %w", idx+1, err)
		}

		// Graph Node: ORDER_ITEM
		itemNodeProps := map[string]any{
			"order_id":          orderID,
			"sequence":          idx + 1,
			"product_name":      item.ProductName,
			"category":          item.Category,
			"sub_category":      item.SubCategory,
			"color_name":        item.ColorName,
			"unit_price":        item.UnitPrice,
			"total_qty":         item.TotalQty,
			"product_image_url": item.ProductImageUrl,
		}
		_, err = s.graph.InsertNode(tx, graph.TypeOrderItem, itemNodeProps, itemID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert graph node for item #%d: %w", idx+1, err)
		}

		// Edge: (ORDER) -[:HAS_ITEM]-> (ORDER_ITEM)
		_, err = s.graph.InsertEdge(tx, orderID, itemID, graph.EdgeHasItem, map[string]any{
			"sequence":  idx + 1,
			"total_qty": item.TotalQty,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to insert edge HAS_ITEM: %w", err)
		}

		// Brandings
		for bIdx, bnd := range item.Brandings {
			bndID := bnd.ID
			if bndID == "" {
				bndID = fmt.Sprintf("bnd_%s_%d_%d_%d", orderNo, idx+1, bIdx+1, nowMs%10000)
			}
			bndQuery := `
				INSERT INTO order_item_brandings (
					id, order_item_id, order_id, branding_type, position, size,
					placement, color, comment, branding_file_url, created_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`
			_, err = tx.Exec(bndQuery,
				bndID, itemID, orderID, bnd.BrandingType, bnd.Position, bnd.Size,
				bnd.Placement, bnd.Color, bnd.Comment, bnd.BrandingFileURL, nowMs,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to insert branding #%d on item #%d: %w", bIdx+1, idx+1, err)
			}

			// Graph Node: BRANDING
			bndProps := map[string]any{
				"order_id":      orderID,
				"order_item_id": itemID,
				"branding_type": bnd.BrandingType,
				"position":      bnd.Position,
				"file_url":      bnd.BrandingFileURL,
			}
			_, _ = s.graph.InsertNode(tx, graph.TypeBranding, bndProps, bndID)
			_, _ = s.graph.InsertEdge(tx, itemID, bndID, graph.EdgeHasBranding, map[string]any{
				"branding_type": bnd.BrandingType,
			})
		}

		// Personalizations
		for pIdx, pers := range item.Personalizations {
			if strings.TrimSpace(pers.Name) == "" && strings.TrimSpace(pers.Number) == "" {
				continue
			}
			persID := pers.ID
			if persID == "" {
				persID = fmt.Sprintf("pers_%s_%d_%d_%d", orderNo, idx+1, pIdx+1, nowMs%10000)
			}
			persQuery := `
				INSERT INTO order_personalizations (
					id, order_item_id, order_id, gender, size, name, number, remarks, created_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`
			_, err = tx.Exec(persQuery,
				persID, itemID, orderID, pers.Gender, pers.Size, pers.Name, pers.Number, pers.Remarks, nowMs,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to insert personalization #%d on item #%d: %w", pIdx+1, idx+1, err)
			}

			// Graph Node: PERSONALIZATION
			persProps := map[string]any{
				"gender": pers.Gender,
				"size":   pers.Size,
				"name":   pers.Name,
				"number": pers.Number,
			}
			_, _ = s.graph.InsertNode(tx, graph.TypePersonalization, persProps, persID)
			_, _ = s.graph.InsertEdge(tx, itemID, persID, graph.EdgeHasPersonalization, map[string]any{
				"size": pers.Size,
			})
		}
	}

	// 5. Insert Multi-Shipping Destinations
	for sIdx, dest := range p.ShippingLocations {
		addr := strings.TrimSpace(dest.DestinationAddress)
		if addr == "" {
			addr = strings.TrimSpace(dest.Destination)
		}
		if addr == "" {
			continue
		}
		destID := dest.ID
		if destID == "" {
			destID = fmt.Sprintf("dest_%s_%d_%d", orderNo, sIdx+1, nowMs%10000)
		}
		fMode := dest.FreightMode
		if fMode == "" {
			fMode = p.FreightMode
		}
		destQuery := `
			INSERT INTO order_shipping_destinations (
				id, order_id, order_item_id, destination_address, quantity, freight_mode, carrier, docket_no, target_etd, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		var itemIDVal *string
		if dest.OrderItemID != "" {
			itemIDVal = &dest.OrderItemID
		}
		_, err = tx.Exec(destQuery, destID, orderID, itemIDVal, addr, dest.Quantity, fMode, dest.Carrier, dest.DocketNo, dest.TargetEtd, nowMs)
		if err != nil {
			return nil, fmt.Errorf("failed to insert shipping destination #%d: %w", sIdx+1, err)
		}

		// Graph Node: SHIPPING_DESTINATION
		destProps := map[string]any{
			"address":  addr,
			"quantity": dest.Quantity,
			"freight":  fMode,
			"carrier":  dest.Carrier,
		}
		_, _ = s.graph.InsertNode(tx, graph.TypeShippingDestination, destProps, destID)
		_, _ = s.graph.InsertEdge(tx, orderID, destID, graph.EdgeShipsTo, map[string]any{
			"quantity": dest.Quantity,
		})
	}

	// 6. Emit Domain Event to Outbox
	_, err = s.workflow.EmitEvent(tx, orderID, "ORDER_BOOKED", map[string]any{
		"order_id":            orderID,
		"production_order_no": orderNo,
		"client_name":         clientName,
		"total_qty":           totalOrderQty,
		"grand_valuation":     grandValuation,
		"item_count":          len(items),
		"source_channel":      sourceChannel,
	}, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to emit ORDER_BOOKED event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Background asynchronous event processing
	go func() {
		_, _ = s.workflow.ProcessPendingEvents()
	}()

	return map[string]any{
		"success":           true,
		"productionOrderNo": orderNo,
		"id":                orderID,
		"totalQty":          totalOrderQty,
		"grandValuation":    grandValuation,
		"itemCount":         len(items),
	}, nil
}

// GetOrder retrieves full order details including line items, embellishments,
// personalizations, and multi-shipping destinations.
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
		SourceChannel       sql.NullString
		TotalOrderValue     sql.NullFloat64
	}

	query := `
		SELECT id, production_order_no, sales_order_no, order_type, booking_date,
		       client_id, client_name, contact_person, phone, email, billing_address, shipping_address, gstin,
		       category, sub_category, product_name, color_name, fabric_name, gender,
		       unit_price, tax_rate, tax_billing, freight_mode, freight_cost,
		       total_qty, size_breakdown_json, multiple_shipping, shipping_locations_json, personalization_json,
		       branding_type, branding_file_url, main_label, other_label, cust_packaging,
		       client_comment, target_etd, status, source_channel, total_order_value
		FROM orders WHERE production_order_no = ?
	`
	err := s.db.QueryRow(query, orderNo).Scan(
		&o.ID, &o.OrderNo, &o.SalesOrderNo, &o.OrderType, &o.BookingDate,
		&o.ClientID, &o.ClientName, &o.ContactPerson, &o.Phone, &o.Email, &o.BillingAddress, &o.ShippingAddress, &o.Gstin,
		&o.Category, &o.SubCategory, &o.ProductName, &o.ColorName, &o.FabricName, &o.Gender,
		&o.UnitPrice, &o.TaxRate, &o.TaxBilling, &o.FreightMode, &o.FreightCost,
		&o.TotalQty, &o.SizeBreakdownStr, &o.MultipleShipping, &o.ShippingLocsStr, &o.PersonalizationStr,
		&o.BrandingType, &o.BrandingFileUrl, &o.MainLabel, &o.OtherLabel, &o.CustPackaging,
		&o.ClientComment, &o.TargetEtd, &o.Status, &o.SourceChannel, &o.TotalOrderValue,
	)
	if err != nil {
		return nil, err
	}

	// 1. Fetch Line Items from order_items
	itemRows, err := s.db.Query(`
		SELECT id, item_sequence, category, sub_category, product_name, color_name,
		       fabric_name, unit_price, total_qty, client_etd, client_remarks,
		       product_image_url, main_label, other_label, cust_packaging, matrix_json
		FROM order_items WHERE order_id = ? ORDER BY item_sequence ASC
	`, o.ID)

	var items []map[string]any
	if err == nil {
		defer itemRows.Close()
		for itemRows.Next() {
			var itID, itCat, itSub, itProd, itColor string
			var itSeq, itQty int
			var itPrice float64
			var itFab, itEtd, itRem, itImg, itMain, itOther, itPkg, itMatrix sql.NullString

			_ = itemRows.Scan(
				&itID, &itSeq, &itCat, &itSub, &itProd, &itColor,
				&itFab, &itPrice, &itQty, &itEtd, &itRem,
				&itImg, &itMain, &itOther, &itPkg, &itMatrix,
			)

			var matrixObj any
			if itMatrix.Valid && itMatrix.String != "" {
				_ = json.Unmarshal([]byte(itMatrix.String), &matrixObj)
			}

			// Fetch Brandings for this item
			var brandings []map[string]any
			bRows, bErr := s.db.Query(`
				SELECT id, branding_type, position, size, placement, color, comment, branding_file_url
				FROM order_item_brandings WHERE order_item_id = ?
			`, itID)
			if bErr == nil {
				for bRows.Next() {
					var bID, bType string
					var bPos, bSize, bPlace, bCol, bComm, bFile sql.NullString
					_ = bRows.Scan(&bID, &bType, &bPos, &bSize, &bPlace, &bCol, &bComm, &bFile)
					brandings = append(brandings, map[string]any{
						"id":              bID,
						"brandingType":    bType,
						"position":        bPos.String,
						"size":            bSize.String,
						"placement":       bPlace.String,
						"color":           bCol.String,
						"comment":         bComm.String,
						"brandingFileUrl": bFile.String,
					})
				}
				bRows.Close()
			}

			// Fetch Personalizations for this item
			var personalizations []map[string]any
			pRows, pErr := s.db.Query(`
				SELECT id, gender, size, name, number, remarks
				FROM order_personalizations WHERE order_item_id = ?
			`, itID)
			if pErr == nil {
				for pRows.Next() {
					var pID, pGender, pSize, pName, pNumber string
					var pRemarks sql.NullString
					_ = pRows.Scan(&pID, &pGender, &pSize, &pName, &pNumber, &pRemarks)
					personalizations = append(personalizations, map[string]any{
						"id":      pID,
						"gender":  pGender,
						"size":    pSize,
						"name":    pName,
						"number":  pNumber,
						"remarks": pRemarks.String,
					})
				}
				pRows.Close()
			}

			items = append(items, map[string]any{
				"id":               itID,
				"itemSequence":     itSeq,
				"category":         itCat,
				"subCategory":      itSub,
				"productName":      itProd,
				"colorName":        itColor,
				"fabricName":       itFab.String,
				"unitPrice":        itPrice,
				"totalQty":         itQty,
				"clientEtd":        itEtd.String,
				"clientRemarks":    itRem.String,
				"productImageUrl":  itImg.String,
				"mainLabel":        itMain.String,
				"otherLabel":       itOther.String,
				"custPackaging":    itPkg.String,
				"matrix":           matrixObj,
				"brandings":        brandings,
				"personalizations": personalizations,
			})
		}
	}

	// 2. Fetch Shipping Destinations
	var shippingDests []map[string]any
	sRows, err := s.db.Query(`
		SELECT id, order_item_id, destination_address, quantity, freight_mode, carrier, docket_no, target_etd
		FROM order_shipping_destinations WHERE order_id = ?
	`, o.ID)
	if err == nil {
		defer sRows.Close()
		for sRows.Next() {
			var sID, sAddr, sFMode string
			var sItem sql.NullString
			var sQty int
			var sCarrier, sDocket, sEtd sql.NullString
			_ = sRows.Scan(&sID, &sItem, &sAddr, &sQty, &sFMode, &sCarrier, &sDocket, &sEtd)
			shippingDests = append(shippingDests, map[string]any{
				"id":                 sID,
				"orderItemId":        sItem.String,
				"destinationAddress": sAddr,
				"destination":        sAddr,
				"quantity":           sQty,
				"freightMode":        sFMode,
				"carrier":            sCarrier.String,
				"docketNo":           sDocket.String,
				"targetEtd":          sEtd.String,
			})
		}
	}

	// Fallback to legacy single-item view if items is empty
	if len(items) == 0 {
		var sizeBreakdown map[string]int
		if o.SizeBreakdownStr.Valid {
			_ = json.Unmarshal([]byte(o.SizeBreakdownStr.String), &sizeBreakdown)
		}
		items = append(items, map[string]any{
			"id":              fmt.Sprintf("item_%s_1", o.OrderNo),
			"itemSequence":    1,
			"category":        o.Category,
			"subCategory":     o.SubCategory,
			"productName":     o.ProductName,
			"colorName":       o.ColorName,
			"fabricName":      o.FabricName.String,
			"unitPrice":       o.UnitPrice,
			"totalQty":        o.TotalQty,
			"matrix":          sizeBreakdown,
			"mainLabel":       o.MainLabel.String,
			"otherLabel":      o.OtherLabel.String,
			"custPackaging":   o.CustPackaging.String,
			"productImageUrl": o.BrandingFileUrl.String,
		})
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
		"tax_rate":            o.TaxRate,
		"tax_billing":         o.TaxBilling,
		"freight_mode":        o.FreightMode,
		"freight_cost":        o.FreightCost,
		"total_qty":           o.TotalQty,
		"total_order_value":   o.TotalOrderValue.Float64,
		"source_channel":      o.SourceChannel.String,
		"client_comment":      o.ClientComment.String,
		"target_etd":          o.TargetEtd.String,
		"status":              o.Status,
		"items":               items,
		"shippingLocations":   shippingDests,
	}, nil
}

// ListOrders returns recent orders with summary totals
func (s *Service) ListOrders(limit, offset int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`
		SELECT id, production_order_no, sales_order_no, client_name, product_name, color_name, fabric_name,
		       total_qty, unit_price, status, target_etd, COALESCE(total_order_value, 0.0), COALESCE(source_channel, 'manual')
		FROM orders ORDER BY created_at DESC LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []map[string]any
	for rows.Next() {
		var id, ordNo, client, prod, col, status string
		var soNo, fab, etd, source sql.NullString
		var qty int
		var price, totalVal float64
		_ = rows.Scan(&id, &ordNo, &soNo, &client, &prod, &col, &fab, &qty, &price, &status, &etd, &totalVal, &source)
		list = append(list, map[string]any{
			"id":                  id,
			"production_order_no": ordNo,
			"sales_order_no":      soNo.String,
			"client_name":         client,
			"product_name":        prod,
			"color_name":          col,
			"fabric_name":         fab.String,
			"total_qty":           qty,
			"unit_price":          price,
			"total_order_value":   totalVal,
			"source_channel":      source.String,
			"status":              status,
			"target_etd":          etd.String,
		})
	}
	return list, nil
}
