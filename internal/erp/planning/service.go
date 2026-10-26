package planning

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ppl-amehta/selvedge/internal/graph"
	"github.com/ppl-amehta/selvedge/internal/workflow"
)

type Service struct {
	db       *sql.DB
	graph    *graph.Store
	workflow *workflow.Engine
}

func NewService(db *sql.DB, graphStore *graph.Store, workflowEngine *workflow.Engine) *Service {
	return &Service{
		db:       db,
		graph:    graphStore,
		workflow: workflowEngine,
	}
}

// -------------------------------------------------------------
// DATA STRUCTS
// -------------------------------------------------------------

type MeasurementPoint struct {
	Point     string  `json:"point"`
	Tolerance string  `json:"tolerance"`
	XS        float64 `json:"xs"`
	S         float64 `json:"s"`
	M         float64 `json:"m"`
	L         float64 `json:"l"`
	XL        float64 `json:"xl"`
	XXL       float64 `json:"xxl"`
	XXXL      float64 `json:"xxxl"`
}

type PharmaPattern struct {
	ID                 string             `json:"id"`
	OrderID            string             `json:"orderId"`
	OrderItemID        string             `json:"orderItemId"`
	PharmaType         string             `json:"pharmaType"`
	PharmaNo           string             `json:"pharmaNo"`
	ChangesToPharma    string             `json:"changesToPharma"`
	MeasurementChart   []MeasurementPoint `json:"measurementChart"`
	Status             string             `json:"status"`
	CreatedAt          int64              `json:"createdAt"`
	UpdatedAt          int64              `json:"updatedAt"`
}

type PharmaPayload struct {
	OrderID          string             `json:"orderId"`
	OrderItemID      string             `json:"orderItemId"`
	PharmaType       string             `json:"pharmaType"`
	PharmaNo         string             `json:"pharmaNo"`
	ChangesToPharma  string             `json:"changesToPharma"`
	MeasurementChart []MeasurementPoint `json:"measurementChart"`
	Status           string             `json:"status"`
}

type BomLine struct {
	ID             string  `json:"id"`
	OrderID        string  `json:"orderId"`
	OrderItemID    string  `json:"orderItemId"`
	LineType       string  `json:"lineType"`
	ItemName       string  `json:"itemName"`
	SubType        string  `json:"subType"`
	ConsumptionAvg float64 `json:"consumptionAvg"`
	WastagePct     float64 `json:"wastagePct"`
	UOM            string  `json:"uom"`
	UnitQty        int     `json:"unitQty"`
	RequiredQty    float64 `json:"requiredQty"`
	SohQty         float64 `json:"sohQty"`
	AllocatedQty   float64 `json:"allocatedQty"`
	ShortfallQty   float64 `json:"shortfallQty"`
	RatePerUnit    float64 `json:"ratePerUnit"`
	Comments       string  `json:"comments"`
	Status         string  `json:"status"`
	CreatedAt      int64   `json:"createdAt"`
	UpdatedAt      int64   `json:"updatedAt"`
}

type BomLinePayload struct {
	ID             string  `json:"id"`
	OrderID        string  `json:"orderId"`
	OrderItemID    string  `json:"orderItemId"`
	LineType       string  `json:"lineType"`
	ItemName       string  `json:"itemName"`
	SubType        string  `json:"subType"`
	ConsumptionAvg float64 `json:"consumptionAvg"`
	WastagePct     float64 `json:"wastagePct"`
	UOM            string  `json:"uom"`
	UnitQty        int     `json:"unitQty"`
	RatePerUnit    float64 `json:"ratePerUnit"`
	Comments       string  `json:"comments"`
}

type ProcurementRequisition struct {
	ID              string  `json:"id"`
	OrderID         string  `json:"orderId"`
	BomLineID       string  `json:"bomLineId"`
	MaterialName    string  `json:"materialName"`
	MaterialType    string  `json:"materialType"`
	UOM             string  `json:"uom"`
	RequiredQty     float64 `json:"requiredQty"`
	OrderBufferQty  float64 `json:"orderBufferQty"`
	TotalPoQty      float64 `json:"totalPoQty"`
	VendorName      string  `json:"vendorName"`
	EstimatedRate   float64 `json:"estimatedRate"`
	EstimatedAmount float64 `json:"estimatedAmount"`
	RequiredByDate  string  `json:"requiredByDate"`
	PoNo            string  `json:"poNo"`
	Status          string  `json:"status"`
	CreatedAt       int64   `json:"createdAt"`
	UpdatedAt       int64   `json:"updatedAt"`
}

type RaisePOPayload struct {
	OrderID        string  `json:"orderId"`
	BomLineID      string  `json:"bomLineId"`
	MaterialName   string  `json:"materialName"`
	MaterialType   string  `json:"materialType"`
	UOM            string  `json:"uom"`
	RequiredQty    float64 `json:"requiredQty"`
	OrderBufferQty float64 `json:"orderBufferQty"`
	VendorName     string  `json:"vendorName"`
	EstimatedRate  float64 `json:"estimatedRate"`
	RequiredByDate string  `json:"requiredByDate"`
}

type RoutingStep struct {
	ID               string `json:"id"`
	OrderID          string `json:"orderId"`
	OrderItemID      string `json:"orderItemId,omitempty"`
	SequenceOrder    int    `json:"sequenceOrder"`
	ProcessCode      string `json:"processCode"`
	StepKind         string `json:"stepKind"`
	VendorName       string `json:"vendorName"`
	DurationDays     int    `json:"durationDays"`
	PlannedStartDate string `json:"plannedStartDate"`
	PlannedEndDate   string `json:"plannedEndDate"`
	PlannedQty       int    `json:"plannedQty"`
	Status           string `json:"status"`
	DependsOnStepID  string `json:"dependsOnStepId,omitempty"`
	Comments         string `json:"comments,omitempty"`
	CreatedAt        int64  `json:"createdAt"`
	UpdatedAt        int64  `json:"updatedAt"`
}

type RoutingStepPayload struct {
	ID               string `json:"id"`
	OrderItemID      string `json:"orderItemId,omitempty"`
	SequenceOrder    int    `json:"sequenceOrder"`
	ProcessCode      string `json:"processCode"`
	StepKind         string `json:"stepKind"`
	VendorName       string `json:"vendorName"`
	DurationDays     int    `json:"durationDays"`
	PlannedStartDate string `json:"plannedStartDate"`
	PlannedEndDate   string `json:"plannedEndDate"`
	PlannedQty       int    `json:"plannedQty"`
	Status           string `json:"status"`
	DependsOnStepID  string `json:"dependsOnStepId,omitempty"`
	Comments         string `json:"comments,omitempty"`
}

type OrderPrepItem struct {
	ID              string           `json:"id"`
	ItemSequence    int              `json:"itemSequence"`
	Category        string           `json:"category"`
	SubCategory     string           `json:"subCategory"`
	ProductName     string           `json:"productName"`
	ColorName       string           `json:"colorName"`
	FabricName      string           `json:"fabricName"`
	UnitPrice       float64          `json:"unitPrice"`
	TotalQty        int              `json:"totalQty"`
	ClientETD       string           `json:"clientEtd"`
	ClientRemarks   string           `json:"clientRemarks"`
	ProductImageURL string           `json:"productImageUrl"`
	MainLabel       string           `json:"mainLabel"`
	OtherLabel      string           `json:"otherLabel"`
	CustPackaging   string           `json:"custPackaging"`
	Matrix          map[string]any   `json:"matrix"`
	PharmaPatterns  []PharmaPattern  `json:"pharmaPatterns"`
	BomLines        []BomLine        `json:"bomLines"`
	ShortfallCount  int              `json:"shortfallCount"`
}

type OrderPrepAndPlanData struct {
	Order struct {
		ID                string  `json:"id"`
		ProductionOrderNo string  `json:"productionOrderNo"`
		SalesOrderNo      string  `json:"salesOrderNo"`
		OrderType         string  `json:"orderType"`
		BookingDate       string  `json:"bookingDate"`
		ClientName        string  `json:"clientName"`
		ContactPerson     string  `json:"contactPerson"`
		Status            string  `json:"status"`
		TotalQty          int     `json:"totalQty"`
		TotalOrderValue   float64 `json:"totalOrderValue"`
		ClientETD         string  `json:"clientEtd"`
	} `json:"order"`
	Items                []OrderPrepItem          `json:"items"`
	Requisitions         []ProcurementRequisition `json:"requisitions"`
	RoutingSteps         []RoutingStep            `json:"routingSteps"`
	GanttSummary         GanttTimelineSummary     `json:"ganttSummary"`
}

type GanttTimelineSummary struct {
	EarliestStart    string `json:"earliestStart"`
	LatestFinish     string `json:"latestFinish"`
	ClientETD        string `json:"clientEtd"`
	SlackDays        int    `json:"slackDays"`
	IsOverdueRisk    bool   `json:"isOverdueRisk"`
	TotalDurationDays int   `json:"totalDurationDays"`
	StatusText       string `json:"statusText"`
}

type OrderPrepListItem struct {
	ID                string  `json:"id"`
	ProductionOrderNo string  `json:"productionOrderNo"`
	SalesOrderNo      string  `json:"salesOrderNo"`
	ClientName        string  `json:"clientName"`
	OrderType         string  `json:"orderType"`
	BookingDate       string  `json:"bookingDate"`
	Status            string  `json:"status"`
	TotalQty          int     `json:"totalQty"`
	ClientETD         string  `json:"clientEtd"`
	ItemCount         int     `json:"itemCount"`
	ShortfallLines    int     `json:"shortfallLines"`
}

type OrderStateAdvanceResult struct {
	Success      bool   `json:"success"`
	OrderID      string `json:"orderId"`
	PreviousState string `json:"previousState"`
	NewState     string `json:"newState"`
	Message      string `json:"message"`
}

// -------------------------------------------------------------
// CORE PREP & PLANNING METHODS
// -------------------------------------------------------------

// ListOrdersForPrep lists all orders ready for prep or in prep/planning
func (s *Service) ListOrdersForPrep() ([]OrderPrepListItem, error) {
	query := `
		SELECT o.id, o.production_order_no, COALESCE(o.sales_order_no, ''), o.client_name,
		       o.order_type, o.booking_date, o.status, o.total_qty, COALESCE(o.target_etd, '')
		FROM orders o
		ORDER BY o.created_at DESC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders for prep: %w", err)
	}
	defer rows.Close()

	var list []OrderPrepListItem
	for rows.Next() {
		var item OrderPrepListItem
		if err := rows.Scan(
			&item.ID, &item.ProductionOrderNo, &item.SalesOrderNo, &item.ClientName,
			&item.OrderType, &item.BookingDate, &item.Status, &item.TotalQty, &item.ClientETD,
		); err != nil {
			return nil, err
		}

		// Count items and shortfalls
		_ = s.db.QueryRow("SELECT COUNT(*) FROM order_items WHERE order_id = ?", item.ID).Scan(&item.ItemCount)
		_ = s.db.QueryRow("SELECT COUNT(*) FROM bom_lines WHERE order_id = ? AND shortfall_qty > 0", item.ID).Scan(&item.ShortfallLines)

		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

// CalculateSOH checks the live stock on hand in the warehouse for a material
func (s *Service) CalculateSOH(itemName string) float64 {
	trimmed := strings.TrimSpace(itemName)
	if trimmed == "" {
		return 0
	}

	likePattern := "%" + strings.ToLower(trimmed) + "%"

	// 1. Check parent inward lots (fabric / trims / packaging)
	var lotSoh float64
	err := s.db.QueryRow(`
		SELECT COALESCE(SUM(
			CASE 
				WHEN counted_qty > 0 THEN counted_qty 
				ELSE physical_received_qty 
			END
		), 0)
		FROM parent_inward_lots
		WHERE COALESCE(stage, '') != 'CANCELLED' 
		  AND (LOWER(fpcc_fab_name) LIKE ? OR LOWER(fpcc_product) LIKE ? OR LOWER(vendor_product) LIKE ?)
	`, likePattern, likePattern, likePattern).Scan(&lotSoh)
	if err == nil && lotSoh > 0 {
		return math.Round(lotSoh*100) / 100
	}

	// 2. Check bundles
	var bundleSoh float64
	err = s.db.QueryRow(`
		SELECT COALESCE(SUM(inward_weight_qty), 0)
		FROM bundles
		WHERE COALESCE(qc_status, '') != 'Failed' AND (LOWER(item_name) LIKE ? OR LOWER(item_type) LIKE ?)
	`, likePattern, likePattern).Scan(&bundleSoh)
	if err == nil && bundleSoh > 0 {
		return math.Round(bundleSoh*100) / 100
	}

	return 0.0
}

// GetOrderPrepAndPlanning retrieves full hierarchical preparation and planning data
func (s *Service) GetOrderPrepAndPlanning(orderID string) (*OrderPrepAndPlanData, error) {
	data := &OrderPrepAndPlanData{}

	// 1. Order Header
	err := s.db.QueryRow(`
		SELECT id, production_order_no, COALESCE(sales_order_no, ''), order_type,
		       booking_date, client_name, COALESCE(contact_person, ''), status,
		       total_qty, COALESCE(total_order_value, 0.0), COALESCE(target_etd, '')
		FROM orders WHERE id = ?
	`, orderID).Scan(
		&data.Order.ID, &data.Order.ProductionOrderNo, &data.Order.SalesOrderNo,
		&data.Order.OrderType, &data.Order.BookingDate, &data.Order.ClientName,
		&data.Order.ContactPerson, &data.Order.Status, &data.Order.TotalQty,
		&data.Order.TotalOrderValue, &data.Order.ClientETD,
	)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// 2. Order Items
	itemRows, err := s.db.Query(`
		SELECT id, item_sequence, category, sub_category, product_name, color_name,
		       COALESCE(fabric_name, ''), unit_price, total_qty, COALESCE(client_etd, ''),
		       COALESCE(client_remarks, ''), COALESCE(product_image_url, ''),
		       COALESCE(main_label, ''), COALESCE(other_label, ''),
		       COALESCE(cust_packaging, ''), COALESCE(matrix_json, '{}')
		FROM order_items
		WHERE order_id = ?
		ORDER BY item_sequence ASC
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var it OrderPrepItem
		var matrixJSON string
		if err := itemRows.Scan(
			&it.ID, &it.ItemSequence, &it.Category, &it.SubCategory, &it.ProductName,
			&it.ColorName, &it.FabricName, &it.UnitPrice, &it.TotalQty, &it.ClientETD,
			&it.ClientRemarks, &it.ProductImageURL, &it.MainLabel, &it.OtherLabel,
			&it.CustPackaging, &matrixJSON,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(matrixJSON), &it.Matrix)

		// 2a. Pharma Patterns for this item
		pRows, pErr := s.db.Query(`
			SELECT id, order_id, order_item_id, pharma_type, pharma_no,
			       COALESCE(changes_to_pharma, ''), COALESCE(measurement_chart_json, '[]'),
			       status, created_at, updated_at
			FROM pharma_patterns
			WHERE order_item_id = ?
			ORDER BY created_at ASC
		`, it.ID)
		if pErr == nil {
			for pRows.Next() {
				var p PharmaPattern
				var chartJSON string
				if err := pRows.Scan(
					&p.ID, &p.OrderID, &p.OrderItemID, &p.PharmaType, &p.PharmaNo,
					&p.ChangesToPharma, &chartJSON, &p.Status, &p.CreatedAt, &p.UpdatedAt,
				); err == nil {
					_ = json.Unmarshal([]byte(chartJSON), &p.MeasurementChart)
					it.PharmaPatterns = append(it.PharmaPatterns, p)
				}
			}
			pRows.Close()
		}

		// 2b. BOM Lines for this item
		bRows, bErr := s.db.Query(`
			SELECT id, order_id, order_item_id, line_type, item_name,
			       COALESCE(sub_type, ''), consumption_avg, wastage_pct, uom,
			       unit_qty, required_qty, soh_qty, allocated_qty, shortfall_qty,
			       rate_per_unit, COALESCE(comments, ''), status, created_at, updated_at
			FROM bom_lines
			WHERE order_item_id = ?
			ORDER BY line_type ASC, item_name ASC
		`, it.ID)
		if bErr == nil {
			for bRows.Next() {
				var b BomLine
				if err := bRows.Scan(
					&b.ID, &b.OrderID, &b.OrderItemID, &b.LineType, &b.ItemName,
					&b.SubType, &b.ConsumptionAvg, &b.WastagePct, &b.UOM,
					&b.UnitQty, &b.RequiredQty, &b.SohQty, &b.AllocatedQty, &b.ShortfallQty,
					&b.RatePerUnit, &b.Comments, &b.Status, &b.CreatedAt, &b.UpdatedAt,
				); err == nil {
					// Dynamically refresh SOH & shortfall
					liveSOH := s.CalculateSOH(b.ItemName)
					if liveSOH != b.SohQty {
						b.SohQty = liveSOH
						b.ShortfallQty = math.Max(0, math.Round((b.RequiredQty-b.SohQty)*100)/100)
						_, _ = s.db.Exec("UPDATE bom_lines SET soh_qty = ?, shortfall_qty = ? WHERE id = ?", b.SohQty, b.ShortfallQty, b.ID)
					}
					if b.ShortfallQty > 0 {
						it.ShortfallCount++
					}
					it.BomLines = append(it.BomLines, b)
				}
			}
			bRows.Close()
		}

		data.Items = append(data.Items, it)
	}

	// 3. Procurement Requisitions
	reqRows, err := s.db.Query(`
		SELECT id, order_id, bom_line_id, material_name, material_type, uom,
		       required_qty, order_buffer_qty, total_po_qty, COALESCE(vendor_name, ''),
		       estimated_rate, estimated_amount, COALESCE(required_by_date, ''),
		       COALESCE(po_no, ''), status, created_at, updated_at
		FROM procurement_requisitions
		WHERE order_id = ?
		ORDER BY created_at DESC
	`, orderID)
	if err == nil {
		defer reqRows.Close()
		for reqRows.Next() {
			var r ProcurementRequisition
			if err := reqRows.Scan(
				&r.ID, &r.OrderID, &r.BomLineID, &r.MaterialName, &r.MaterialType,
				&r.UOM, &r.RequiredQty, &r.OrderBufferQty, &r.TotalPoQty, &r.VendorName,
				&r.EstimatedRate, &r.EstimatedAmount, &r.RequiredByDate, &r.PoNo,
				&r.Status, &r.CreatedAt, &r.UpdatedAt,
			); err == nil {
				data.Requisitions = append(data.Requisitions, r)
			}
		}
	}

	// 4. Routing Steps
	stepRows, err := s.db.Query(`
		SELECT id, order_id, COALESCE(order_item_id, ''), sequence_order,
		       process_code, step_kind, COALESCE(vendor_name, ''), duration_days,
		       planned_start_date, planned_end_date, planned_qty, status,
		       COALESCE(depends_on_step_id, ''), COALESCE(comments, ''),
		       created_at, updated_at
		FROM routing_steps
		WHERE order_id = ?
		ORDER BY sequence_order ASC
	`, orderID)
	if err == nil {
		defer stepRows.Close()
		for stepRows.Next() {
			var st RoutingStep
			if err := stepRows.Scan(
				&st.ID, &st.OrderID, &st.OrderItemID, &st.SequenceOrder,
				&st.ProcessCode, &st.StepKind, &st.VendorName, &st.DurationDays,
				&st.PlannedStartDate, &st.PlannedEndDate, &st.PlannedQty, &st.Status,
				&st.DependsOnStepID, &st.Comments, &st.CreatedAt, &st.UpdatedAt,
			); err == nil {
				data.RoutingSteps = append(data.RoutingSteps, st)
			}
		}
	}

	// 5. Compute Gantt Timeline Summary & ETD Comparison
	data.GanttSummary = s.computeGanttSummary(data.RoutingSteps, data.Order.ClientETD)

	return data, nil
}

func (s *Service) computeGanttSummary(steps []RoutingStep, clientETD string) GanttTimelineSummary {
	summary := GanttTimelineSummary{
		ClientETD: clientETD,
	}
	if len(steps) == 0 {
		summary.StatusText = "No routing steps scheduled yet"
		return summary
	}

	earliest := steps[0].PlannedStartDate
	latest := steps[0].PlannedEndDate
	totalDays := 0

	for _, st := range steps {
		totalDays += st.DurationDays
		if st.PlannedStartDate < earliest && st.PlannedStartDate != "" {
			earliest = st.PlannedStartDate
		}
		if st.PlannedEndDate > latest && st.PlannedEndDate != "" {
			latest = st.PlannedEndDate
		}
	}

	summary.EarliestStart = earliest
	summary.LatestFinish = latest
	summary.TotalDurationDays = totalDays

	if clientETD != "" && latest != "" {
		tClient, errC := time.Parse("2006-01-02", clientETD)
		tFinish, errF := time.Parse("2006-01-02", latest)
		if errC == nil && errF == nil {
			diffDays := int(tClient.Sub(tFinish).Hours() / 24)
			summary.SlackDays = diffDays
			if diffDays < 0 {
				summary.IsOverdueRisk = true
				summary.StatusText = fmt.Sprintf("⚠️ Overdue Risk: Exceeds client ETD by %d days!", -diffDays)
			} else {
				summary.IsOverdueRisk = false
				summary.StatusText = fmt.Sprintf("✓ On Schedule (%d days safety buffer before ETD)", diffDays)
			}
		}
	} else {
		summary.StatusText = fmt.Sprintf("Scheduled duration: %d days", totalDays)
	}

	return summary
}

// -------------------------------------------------------------
// PHARMA PATTERN HANDLERS
// -------------------------------------------------------------

// SavePharma creates or updates a pharma pattern tech pack spec
func (s *Service) SavePharma(orderItemID string, p PharmaPayload) (*PharmaPattern, error) {
	pharmaType := strings.TrimSpace(p.PharmaType)
	pharmaNo := strings.TrimSpace(p.PharmaNo)
	if pharmaType == "" || pharmaNo == "" {
		return nil, fmt.Errorf("pharmaType and pharmaNo are required")
	}

	// Fetch OrderID and unit info if not passed
	var orderID string
	err := s.db.QueryRow("SELECT order_id FROM order_items WHERE id = ?", orderItemID).Scan(&orderID)
	if err != nil {
		return nil, fmt.Errorf("order item not found: %w", err)
	}

	now := graph.NowMs()
	pharmaID := fmt.Sprintf("ph_%s_%s", uuid.New().String()[:8], strings.ToLower(pharmaType))

	chartJSON, _ := json.Marshal(p.MeasurementChart)
	status := p.Status
	if status == "" {
		status = "APPROVED"
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert Pharma Record
	_, err = tx.Exec(`
		INSERT INTO pharma_patterns (
			id, order_id, order_item_id, pharma_type, pharma_no,
			changes_to_pharma, measurement_chart_json, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, pharmaID, orderID, orderItemID, pharmaType, pharmaNo,
		p.ChangesToPharma, string(chartJSON), status, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to insert pharma pattern: %w", err)
	}

	// Bi-Temporal Graph Vertex & Edge
	props := map[string]any{
		"pharmaType": pharmaType,
		"pharmaNo":   pharmaNo,
		"status":     status,
		"changes":    p.ChangesToPharma,
	}
	if _, err := s.graph.EnsureNode(tx, graph.TypePharmaPattern, props, pharmaID); err != nil {
		return nil, fmt.Errorf("failed to insert pharma graph node: %w", err)
	}

	if _, err := s.graph.InsertEdge(tx, orderItemID, pharmaID, graph.EdgeHasPharma, map[string]any{
		"pharmaType": pharmaType,
	}); err != nil {
		return nil, fmt.Errorf("failed to link pharma edge: %w", err)
	}

	// Update order status to IN_PREP if currently BOOKED
	_, _ = tx.Exec("UPDATE orders SET status = 'IN_PREP', updated_at = ? WHERE id = ? AND status = 'BOOKED'", now, orderID)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &PharmaPattern{
		ID:               pharmaID,
		OrderID:          orderID,
		OrderItemID:      orderItemID,
		PharmaType:       pharmaType,
		PharmaNo:         pharmaNo,
		ChangesToPharma:  p.ChangesToPharma,
		MeasurementChart: p.MeasurementChart,
		Status:           status,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

// -------------------------------------------------------------
// BOM ENGINE HANDLERS
// -------------------------------------------------------------

// SaveBomLine adds or updates a BOM material line, calculates required quantity & checks SOH
func (s *Service) SaveBomLine(orderItemID string, p BomLinePayload) (*BomLine, error) {
	itemName := strings.TrimSpace(p.ItemName)
	lineType := strings.ToUpper(strings.TrimSpace(p.LineType))
	if itemName == "" {
		return nil, fmt.Errorf("itemName is required")
	}
	if lineType == "" {
		lineType = "FABRIC"
	}

	var orderID string
	var unitQty int
	err := s.db.QueryRow("SELECT order_id, total_qty FROM order_items WHERE id = ?", orderItemID).Scan(&orderID, &unitQty)
	if err != nil {
		return nil, fmt.Errorf("order item not found: %w", err)
	}
	if p.UnitQty > 0 {
		unitQty = p.UnitQty
	}

	uom := strings.TrimSpace(p.UOM)
	if uom == "" {
		if lineType == "FABRIC" {
			uom = "Kgs"
		} else {
			uom = "Pieces"
		}
	}

	// Requirement Math: Total_Qty * Consumption_Avg * (1 + Wastage/100)
	rawReq := float64(unitQty) * p.ConsumptionAvg * (1.0 + (p.WastagePct / 100.0))
	requiredQty := math.Round(rawReq*100) / 100

	// Check live warehouse SOH
	sohQty := s.CalculateSOH(itemName)
	shortfallQty := math.Max(0, math.Round((requiredQty-sohQty)*100)/100)
	allocatedQty := math.Min(sohQty, requiredQty)

	now := graph.NowMs()
	lineID := p.ID
	if lineID == "" {
		lineID = fmt.Sprintf("bom_%s_%s", strings.ToLower(lineType), uuid.New().String()[:8])
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO bom_lines (
			id, order_id, order_item_id, line_type, item_name, sub_type,
			consumption_avg, wastage_pct, uom, unit_qty, required_qty,
			soh_qty, allocated_qty, shortfall_qty, rate_per_unit, comments,
			status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'ACTIVE', ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			item_name = excluded.item_name,
			sub_type = excluded.sub_type,
			consumption_avg = excluded.consumption_avg,
			wastage_pct = excluded.wastage_pct,
			uom = excluded.uom,
			unit_qty = excluded.unit_qty,
			required_qty = excluded.required_qty,
			soh_qty = excluded.soh_qty,
			allocated_qty = excluded.allocated_qty,
			shortfall_qty = excluded.shortfall_qty,
			rate_per_unit = excluded.rate_per_unit,
			comments = excluded.comments,
			updated_at = excluded.updated_at
	`, lineID, orderID, orderItemID, lineType, itemName, p.SubType,
		p.ConsumptionAvg, p.WastagePct, uom, unitQty, requiredQty,
		sohQty, allocatedQty, shortfallQty, p.RatePerUnit, p.Comments, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to save bom line: %w", err)
	}

	// Graph Vertex & Edge
	props := map[string]any{
		"lineType":     lineType,
		"itemName":     itemName,
		"requiredQty":  requiredQty,
		"sohQty":       sohQty,
		"shortfallQty": shortfallQty,
		"uom":          uom,
	}
	if _, err := s.graph.EnsureNode(tx, graph.TypeBomLine, props, lineID); err != nil {
		return nil, fmt.Errorf("failed to insert bom line graph node: %w", err)
	}

	if _, err := s.graph.InsertEdge(tx, orderItemID, lineID, graph.EdgeRequiresBom, map[string]any{
		"lineType": lineType,
		"qty":      requiredQty,
	}); err != nil {
		return nil, fmt.Errorf("failed to link bom line edge: %w", err)
	}

	// Update order to IN_PREP if currently BOOKED
	_, _ = tx.Exec("UPDATE orders SET status = 'IN_PREP', updated_at = ? WHERE id = ? AND status = 'BOOKED'", now, orderID)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &BomLine{
		ID:             lineID,
		OrderID:        orderID,
		OrderItemID:    orderItemID,
		LineType:       lineType,
		ItemName:       itemName,
		SubType:        p.SubType,
		ConsumptionAvg: p.ConsumptionAvg,
		WastagePct:     p.WastagePct,
		UOM:            uom,
		UnitQty:        unitQty,
		RequiredQty:    requiredQty,
		SohQty:         sohQty,
		AllocatedQty:   allocatedQty,
		ShortfallQty:   shortfallQty,
		RatePerUnit:    p.RatePerUnit,
		Comments:       p.Comments,
		Status:         "ACTIVE",
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// DeleteBomLine removes a BOM line
func (s *Service) DeleteBomLine(lineID string) error {
	_, err := s.db.Exec("DELETE FROM bom_lines WHERE id = ?", lineID)
	return err
}

// -------------------------------------------------------------
// PROCUREMENT REQUISITION HANDLER
// -------------------------------------------------------------

// RaiseProcurementRequisition converts a BOM shortfall into a procurement requisition / PO
func (s *Service) RaiseProcurementRequisition(orderID string, p RaisePOPayload) (*ProcurementRequisition, error) {
	materialName := strings.TrimSpace(p.MaterialName)
	if materialName == "" {
		return nil, fmt.Errorf("materialName is required")
	}

	totalPoQty := p.RequiredQty + p.OrderBufferQty
	if totalPoQty <= 0 {
		return nil, fmt.Errorf("total procurement quantity must be greater than 0")
	}

	now := graph.NowMs()
	reqID := fmt.Sprintf("po_req_%s_%s", strings.ToLower(p.MaterialType), uuid.New().String()[:8])
	poNo := fmt.Sprintf("PO-%s-%02d", time.Now().Format("060102"), (now%100)+1)

	estimatedAmount := math.Round(totalPoQty*p.EstimatedRate*100) / 100

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var bomLineID any = nil
	if p.BomLineID != "" {
		bomLineID = p.BomLineID
	}

	_, err = tx.Exec(`
		INSERT INTO procurement_requisitions (
			id, order_id, bom_line_id, material_name, material_type, uom,
			required_qty, order_buffer_qty, total_po_qty, vendor_name,
			estimated_rate, estimated_amount, required_by_date, po_no,
			status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'DRAFT', ?, ?)
	`, reqID, orderID, bomLineID, materialName, p.MaterialType, p.UOM,
		p.RequiredQty, p.OrderBufferQty, totalPoQty, p.VendorName,
		p.EstimatedRate, estimatedAmount, p.RequiredByDate, poNo, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to insert procurement requisition: %w", err)
	}

	// Graph Vertex & Edge
	props := map[string]any{
		"material": materialName,
		"poQty":    totalPoQty,
		"poNo":     poNo,
		"vendor":   p.VendorName,
	}
	if _, err := s.graph.EnsureNode(tx, graph.TypeProcurementRequisition, props, reqID); err != nil {
		return nil, fmt.Errorf("failed to insert procurement graph node: %w", err)
	}

	if p.BomLineID != "" {
		if _, err := s.graph.InsertEdge(tx, p.BomLineID, reqID, graph.EdgeHasRequisition, map[string]any{
			"poNo": poNo,
		}); err != nil {
			return nil, fmt.Errorf("failed to link requisition edge: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &ProcurementRequisition{
		ID:              reqID,
		OrderID:         orderID,
		BomLineID:       p.BomLineID,
		MaterialName:    materialName,
		MaterialType:    p.MaterialType,
		UOM:             p.UOM,
		RequiredQty:     p.RequiredQty,
		OrderBufferQty:  p.OrderBufferQty,
		TotalPoQty:      totalPoQty,
		VendorName:      p.VendorName,
		EstimatedRate:   p.EstimatedRate,
		EstimatedAmount: estimatedAmount,
		RequiredByDate:  p.RequiredByDate,
		PoNo:            poNo,
		Status:          "DRAFT",
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// -------------------------------------------------------------
// ORDER PLANNING & ROUTING HANDLERS (Station 3)
// -------------------------------------------------------------

// SaveRoutingSteps replaces or saves the routing sequence for an order
func (s *Service) SaveRoutingSteps(orderID string, steps []RoutingStepPayload) ([]RoutingStep, error) {
	if len(steps) == 0 {
		return nil, fmt.Errorf("at least one routing step is required")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Clear unexecuted steps for clean re-planning
	_, err = tx.Exec("DELETE FROM routing_steps WHERE order_id = ? AND status = 'PLANNED'", orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to clear previous planned steps: %w", err)
	}

	now := graph.NowMs()
	var savedSteps []RoutingStep
	var prevStepID string

	for i, st := range steps {
		stepID := st.ID
		if stepID == "" {
			stepID = fmt.Sprintf("step_%s_%d_%s", strings.ToLower(st.ProcessCode), i+1, uuid.New().String()[:6])
		}

		duration := st.DurationDays
		if duration <= 0 {
			duration = 1
		}

		kind := st.StepKind
		if kind == "" {
			kind = "INTERNAL"
		}

		status := st.Status
		if status == "" {
			status = "PLANNED"
		}

		seq := i + 1
		dependsOnStr := prevStepID
		if st.DependsOnStepID != "" {
			dependsOnStr = st.DependsOnStepID
		}
		var dependsOnVal any = nil
		if dependsOnStr != "" {
			dependsOnVal = dependsOnStr
		}

		var orderItemID any = nil
		if st.OrderItemID != "" {
			orderItemID = st.OrderItemID
		}

		_, err = tx.Exec(`
			INSERT INTO routing_steps (
				id, order_id, order_item_id, sequence_order, process_code,
				step_kind, vendor_name, duration_days, planned_start_date,
				planned_end_date, planned_qty, status, depends_on_step_id,
				comments, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, stepID, orderID, orderItemID, seq, st.ProcessCode,
			kind, st.VendorName, duration, st.PlannedStartDate,
			st.PlannedEndDate, st.PlannedQty, status, dependsOnVal,
			st.Comments, now, now)
		if err != nil {
			return nil, fmt.Errorf("failed to insert routing step #%d: %w", seq, err)
		}

		// Graph Node & Edges
		props := map[string]any{
			"process":   st.ProcessCode,
			"vendor":    st.VendorName,
			"duration":  duration,
			"startDate": st.PlannedStartDate,
			"endDate":   st.PlannedEndDate,
			"status":    status,
		}
		if _, err := s.graph.EnsureNode(tx, graph.TypeRoutingStep, props, stepID); err != nil {
			return nil, fmt.Errorf("failed to insert step graph node: %w", err)
		}

		if _, err := s.graph.InsertEdge(tx, orderID, stepID, graph.EdgeHasRouting, map[string]any{
			"sequence": seq,
		}); err != nil {
			return nil, fmt.Errorf("failed to link routing edge: %w", err)
		}

		// Sequence DAG edge
		if prevStepID != "" {
			_, _ = s.graph.InsertEdge(tx, prevStepID, stepID, graph.EdgeSequencedAfter, map[string]any{
				"dependency": "FINISH_TO_START",
			})
		}

		prevStepID = stepID
		savedSteps = append(savedSteps, RoutingStep{
			ID:               stepID,
			OrderID:          orderID,
			OrderItemID:      st.OrderItemID,
			SequenceOrder:    seq,
			ProcessCode:      st.ProcessCode,
			StepKind:         kind,
			VendorName:       st.VendorName,
			DurationDays:     duration,
			PlannedStartDate: st.PlannedStartDate,
			PlannedEndDate:   st.PlannedEndDate,
			PlannedQty:       st.PlannedQty,
			Status:           status,
			DependsOnStepID:  dependsOnStr,
			Comments:         st.Comments,
			CreatedAt:        now,
			UpdatedAt:        now,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return savedSteps, nil
}

// -------------------------------------------------------------
// STATE MACHINE TRANSITIONS
// -------------------------------------------------------------

// AdvanceOrderState transitions the order through the formal workflow gates:
// BOOKED -> IN_PREP -> PREP_COMPLETE -> PLANNED -> READY_FOR_EXECUTION
func (s *Service) AdvanceOrderState(orderID string, targetState string, reason string) (*OrderStateAdvanceResult, error) {
	targetState = strings.ToUpper(strings.TrimSpace(targetState))

	var currentStatus string
	var clientETD string
	err := s.db.QueryRow("SELECT status, COALESCE(target_etd, '') FROM orders WHERE id = ?", orderID).Scan(&currentStatus, &clientETD)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Idempotent: already at target state
	if currentStatus == targetState {
		return &OrderStateAdvanceResult{
			Success:       true,
			OrderID:       orderID,
			PreviousState: currentStatus,
			NewState:      targetState,
			Message:       fmt.Sprintf("Order is already at status %s", targetState),
		}, nil
	}

	now := graph.NowMs()
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	switch targetState {
	case "IN_PREP":
		// Allowed from BOOKED
		if currentStatus != "BOOKED" && currentStatus != "IN_PREP" {
			return nil, fmt.Errorf("cannot transition to IN_PREP from status %s", currentStatus)
		}

	case "PREP_COMPLETE":
		// Guard: Every order item must have at least 1 Pharma spec and at least 1 Fabric BOM line
		var itemCount int
		_ = tx.QueryRow("SELECT COUNT(*) FROM order_items WHERE order_id = ?", orderID).Scan(&itemCount)
		if itemCount == 0 {
			return nil, fmt.Errorf("cannot advance to PREP_COMPLETE: order has no line items")
		}

		var itemsWithoutPharma int
		_ = tx.QueryRow(`
			SELECT COUNT(*) FROM order_items oi
			WHERE oi.order_id = ?
			  AND NOT EXISTS (SELECT 1 FROM pharma_patterns pp WHERE pp.order_item_id = oi.id)
		`, orderID).Scan(&itemsWithoutPharma)
		if itemsWithoutPharma > 0 {
			return nil, fmt.Errorf("cannot advance to PREP_COMPLETE: %d product item(s) are missing Pharma pattern tech packs", itemsWithoutPharma)
		}

		var itemsWithoutFabric int
		_ = tx.QueryRow(`
			SELECT COUNT(*) FROM order_items oi
			WHERE oi.order_id = ?
			  AND NOT EXISTS (SELECT 1 FROM bom_lines bl WHERE bl.order_item_id = oi.id AND bl.line_type = 'FABRIC')
		`, orderID).Scan(&itemsWithoutFabric)
		if itemsWithoutFabric > 0 {
			return nil, fmt.Errorf("cannot advance to PREP_COMPLETE: %d product item(s) are missing Fabric BOM requirements", itemsWithoutFabric)
		}

	case "PLANNED":
		// Guard: Must have routing steps with dates
		var stepCount int
		_ = tx.QueryRow("SELECT COUNT(*) FROM routing_steps WHERE order_id = ?", orderID).Scan(&stepCount)
		if stepCount == 0 {
			return nil, fmt.Errorf("cannot advance to PLANNED: no routing steps scheduled")
		}

		// Check latest finish date vs client ETD
		var latestFinish string
		_ = tx.QueryRow("SELECT MAX(planned_end_date) FROM routing_steps WHERE order_id = ?", orderID).Scan(&latestFinish)

		slackDays := 0
		isOverdue := 0
		if clientETD != "" && latestFinish != "" {
			tClient, _ := time.Parse("2006-01-02", clientETD)
			tFinish, _ := time.Parse("2006-01-02", latestFinish)
			slackDays = int(tClient.Sub(tFinish).Hours() / 24)
			if slackDays < 0 {
				isOverdue = 1
			}
		}

		// Snapshot planning
		snapID := fmt.Sprintf("plan_snap_%s_%s", orderID, uuid.New().String()[:6])
		_, err = tx.Exec(`
			INSERT INTO order_planning_snapshots (
				id, order_id, snapshot_version, client_etd, planned_finish_date,
				slack_days, is_overdue_risk, critical_path_summary, approved_by, created_at
			) VALUES (?, ?, 1, ?, ?, ?, ?, ?, 'Planning Lead', ?)
		`, snapID, orderID, clientETD, latestFinish, slackDays, isOverdue, reason, now)
		if err != nil {
			return nil, fmt.Errorf("failed to record planning snapshot: %w", err)
		}

	case "READY_FOR_EXECUTION":
		// Guard: Must be PLANNED
		if currentStatus != "PLANNED" && currentStatus != "READY_FOR_EXECUTION" {
			return nil, fmt.Errorf("cannot advance to READY_FOR_EXECUTION from status %s (must be PLANNED first)", currentStatus)
		}

	default:
		return nil, fmt.Errorf("unrecognized target state: %s", targetState)
	}

	res, err := tx.Exec("UPDATE orders SET status = ?, updated_at = ? WHERE id = ? AND status = ?", targetState, now, orderID, currentStatus)
	if err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("concurrent modification detected: order status changed from %s", currentStatus)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &OrderStateAdvanceResult{
		Success:       true,
		OrderID:       orderID,
		PreviousState: currentStatus,
		NewState:      targetState,
		Message:       fmt.Sprintf("Order successfully transitioned from %s to %s", currentStatus, targetState),
	}, nil
}
