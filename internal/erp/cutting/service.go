package cutting

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

type CuttingOrderSummary struct {
	ID                string `json:"id"`
	OrderID           string `json:"orderId"`
	OrderItemID       string `json:"orderItemId"`
	ProductionOrderNo string `json:"productionOrderNo"`
	ClientName        string `json:"clientName"`
	ItemSequence      int    `json:"itemSequence"`
	Category          string `json:"category"`
	SubCategory       string `json:"subCategory"`
	ProductName       string `json:"productName"`
	ColorName         string `json:"colorName"`
	FabricName        string `json:"fabricName"`
	TargetQty         int    `json:"targetQty"`
	TotalCutQty       int    `json:"totalCutQty"`
	BalanceQty        int    `json:"balanceQty"`
	ClientETD         string `json:"clientEtd"`
	OrderStatus       string `json:"orderStatus"`
	CuttingJobStatus  string `json:"cuttingJobStatus"`
	PharmaNo          string `json:"pharmaNo"`
}

type SizeProgress struct {
	Size          string `json:"size"`
	TargetQty     int    `json:"targetQty"`
	PreviouslyCut int    `json:"previouslyCut"`
	TodayCut      int    `json:"todayCut"`
	TotalCut      int    `json:"totalCut"`
	Balance       int    `json:"balance"`
	CompletionPct float64 `json:"completionPct"`
}

type CutLayRecord struct {
	ID                string         `json:"id"`
	CuttingJobID      string         `json:"cuttingJobId"`
	LayNumber         int            `json:"layNumber"`
	RollCode          string         `json:"rollCode"`
	FabricName        string         `json:"fabricName"`
	RollWeightKg      float64        `json:"rollWeightKg"`
	MarkerName        string         `json:"markerName"`
	Ratio             map[string]int `json:"ratio"`
	Plies             int            `json:"plies"`
	TotalPiecesInLay  int            `json:"totalPiecesInLay"`
	LayWeightKg       float64        `json:"layWeightKg"`
	ActualLayAvgKg    float64        `json:"actualLayAvgKg"`
	PlannedBomAvgKg   float64        `json:"plannedBomAvgKg"`
	EfficiencyPct     float64        `json:"efficiencyPct"`
	Status            string         `json:"status"`
	OperatorName      string         `json:"operatorName"`
	CreatedAt         int64          `json:"createdAt"`
}

type CutBundleRecord struct {
	ID               string `json:"id"`
	CuttingJobID     string `json:"cuttingJobId"`
	CutLayID         string `json:"cutLayId"`
	BundleSequence   int    `json:"bundleSequence"`
	Size             string `json:"size"`
	Pieces           int    `json:"pieces"`
	SlipPrintedCount int    `json:"slipPrintedCount"`
	NextProcess      string `json:"nextProcess"`
	OperatorName     string `json:"operatorName"`
	CreatedAt        int64  `json:"createdAt"`
}

type RemnantRecord struct {
	ID             string  `json:"id"`
	Prefix         string  `json:"prefix"` // 'Bit' or 'Rej'
	SourceRollCode string  `json:"sourceRollCode"`
	CuttingJobID   string  `json:"cuttingJobId,omitempty"`
	WeightKg       float64 `json:"weightKg"`
	BinLocation    string  `json:"binLocation"`
	DefectReason   string  `json:"defectReason,omitempty"`
	OperatorName   string  `json:"operatorName"`
	CreatedAt      int64   `json:"createdAt"`
}

type CuttingJobDetail struct {
	Job struct {
		ID                string  `json:"id"`
		OrderID           string  `json:"orderId"`
		OrderItemID       string  `json:"orderItemId"`
		ProductionOrderNo string  `json:"productionOrderNo"`
		ProductName       string  `json:"productName"`
		ColorName         string  `json:"colorName"`
		FabricName        string  `json:"fabricName"`
		TargetQty         int     `json:"targetQty"`
		PharmaNo          string  `json:"pharmaNo"`
		PlannedBomAvg     float64 `json:"plannedBomAvg"`
		Status            string  `json:"status"`
		OperatorName      string  `json:"operatorName"`
	} `json:"job"`
	OrderHeader struct {
		ClientName string `json:"clientName"`
		ClientETD  string `json:"clientEtd"`
		OrderType  string `json:"orderType"`
	} `json:"orderHeader"`
	TotalTargetQty      int               `json:"totalTargetQty"`
	TotalCutQty         int               `json:"totalCutQty"`
	TotalBalanceQty     int               `json:"totalBalanceQty"`
	AverageLayAvgKg     float64           `json:"averageLayAvgKg"`
	PlannedBomAvgKg     float64           `json:"plannedBomAvgKg"`
	SizeProgressList    []SizeProgress    `json:"sizeProgressList"`
	IssuedRolls         []map[string]any  `json:"issuedRolls"`
	Lays                []CutLayRecord    `json:"lays"`
	Bundles             []CutBundleRecord `json:"bundles"`
	Remnants            []RemnantRecord   `json:"remnants"`
}

type RecordLayPayload struct {
	CuttingJobID string         `json:"cuttingJobId"`
	RollCode     string         `json:"rollCode"`
	FabricName   string         `json:"fabricName"`
	RollWeightKg float64        `json:"rollWeightKg"`
	MarkerName   string         `json:"markerName"`
	Ratio        map[string]int `json:"ratio"`
	Plies        int            `json:"plies"`
	LayWeightKg  float64        `json:"layWeightKg"`
	NextProcess  string         `json:"nextProcess"`
	OperatorName string         `json:"operatorName"`
}

type RecordRemnantPayload struct {
	Prefix         string  `json:"prefix"` // 'Bit' or 'Rej'
	SourceRollCode string  `json:"sourceRollCode"`
	CuttingJobID   string  `json:"cuttingJobId,omitempty"`
	CutLayID       string  `json:"cutLayId,omitempty"`
	WeightKg       float64 `json:"weightKg"`
	BinLocation    string  `json:"binLocation"`
	DefectReason   string  `json:"defectReason"`
	OperatorName   string  `json:"operatorName"`
}

type BundleSlipData struct {
	BundleCode        string `json:"bundleCode"`
	ProductionOrderNo string `json:"productionOrderNo"`
	ProductName       string `json:"productName"`
	ColorName         string `json:"colorName"`
	Size              string `json:"size"`
	Pieces            int    `json:"pieces"`
	LayNumber         int    `json:"layNumber"`
	NextProcess       string `json:"nextProcess"`
	OperatorName      string `json:"operatorName"`
	DateStr           string `json:"dateStr"`
	QRData            string `json:"qrData"`
	SlipPrintedCount  int    `json:"slipPrintedCount"`
}

// -------------------------------------------------------------
// LIST & GET CUTTING JOBS
// -------------------------------------------------------------

// ListCuttingOrders finds all production orders ready for cutting
func (s *Service) ListCuttingOrders() ([]CuttingOrderSummary, error) {
	query := `
		SELECT o.id, oi.id, o.production_order_no, o.client_name,
		       oi.item_sequence, oi.category, oi.sub_category,
		       oi.product_name, oi.color_name, COALESCE(oi.fabric_name, ''),
		       oi.total_qty, COALESCE(o.target_etd, ''), o.status
		FROM orders o
		JOIN order_items oi ON oi.order_id = o.id
		ORDER BY o.created_at DESC, oi.item_sequence ASC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders for cutting: %w", err)
	}
	defer rows.Close()

	var list []CuttingOrderSummary
	for rows.Next() {
		var it CuttingOrderSummary
		if err := rows.Scan(
			&it.OrderID, &it.OrderItemID, &it.ProductionOrderNo, &it.ClientName,
			&it.ItemSequence, &it.Category, &it.SubCategory, &it.ProductName,
			&it.ColorName, &it.FabricName, &it.TargetQty, &it.ClientETD, &it.OrderStatus,
		); err != nil {
			return nil, err
		}

		it.ID = fmt.Sprintf("CJ-%s-%d", it.ProductionOrderNo, it.ItemSequence)

		// Check if a cutting job exists
		var jobStatus, pharmaNo string
		err := s.db.QueryRow("SELECT status, COALESCE(pharma_no, '') FROM cutting_jobs WHERE order_item_id = ?", it.OrderItemID).Scan(&jobStatus, &pharmaNo)
		if err == nil {
			it.CuttingJobStatus = jobStatus
			it.PharmaNo = pharmaNo
		} else {
			it.CuttingJobStatus = "OPEN"
			// Check if Pharma exists in Station 2
			_ = s.db.QueryRow("SELECT COALESCE(pharma_no, '') FROM pharma_patterns WHERE order_item_id = ? LIMIT 1", it.OrderItemID).Scan(&it.PharmaNo)
		}

		// Calculate total cut pieces so far
		var cutSum sql.NullInt64
		_ = s.db.QueryRow(`
			SELECT SUM(cb.pieces) 
			FROM cut_bundles cb 
			JOIN cutting_jobs cj ON cj.id = cb.cutting_job_id 
			WHERE cj.order_item_id = ?
		`, it.OrderItemID).Scan(&cutSum)
		it.TotalCutQty = int(cutSum.Int64)
		it.BalanceQty = int(math.Max(0, float64(it.TargetQty-it.TotalCutQty)))

		list = append(list, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

// GetCuttingJob retrieves or initializes the cutting job with live counters
func (s *Service) GetCuttingJob(orderItemOrJobId string) (*CuttingJobDetail, error) {
	detail := &CuttingJobDetail{}

	// 1. Resolve order item ID & order ID
	var orderItemID, orderID string
	err := s.db.QueryRow(`
		SELECT id, order_id FROM order_items WHERE id = ?
		UNION
		SELECT order_item_id, order_id FROM cutting_jobs WHERE id = ?
		LIMIT 1
	`, orderItemOrJobId, orderItemOrJobId).Scan(&orderItemID, &orderID)
	if err != nil {
		// Try matching by production order no
		err = s.db.QueryRow(`
			SELECT oi.id, oi.order_id 
			FROM order_items oi
			JOIN orders o ON o.id = oi.order_id
			WHERE o.production_order_no = ?
			LIMIT 1
		`, orderItemOrJobId).Scan(&orderItemID, &orderID)
		if err != nil {
			return nil, fmt.Errorf("cutting job or order item not found: %w", err)
		}
	}

	// 2. Fetch Order Header & Item Details
	var matrixJSON string
	var itemSeq int
	err = s.db.QueryRow(`
		SELECT o.production_order_no, o.client_name, COALESCE(o.target_etd, ''), o.order_type,
		       oi.item_sequence, oi.product_name, oi.color_name, COALESCE(oi.fabric_name, ''),
		       oi.total_qty, COALESCE(oi.matrix_json, '{}')
		FROM orders o
		JOIN order_items oi ON oi.id = ? AND oi.order_id = o.id
	`, orderItemID).Scan(
		&detail.Job.ProductionOrderNo, &detail.OrderHeader.ClientName,
		&detail.OrderHeader.ClientETD, &detail.OrderHeader.OrderType,
		&itemSeq, &detail.Job.ProductName, &detail.Job.ColorName,
		&detail.Job.FabricName, &detail.Job.TargetQty, &matrixJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order item data: %w", err)
	}
	detail.Job.OrderID = orderID
	detail.Job.OrderItemID = orderItemID
	detail.TotalTargetQty = detail.Job.TargetQty

	// 3. Fetch or Create Cutting Job Record
	expectedJobID := fmt.Sprintf("CJ-%s-%d", detail.Job.ProductionOrderNo, itemSeq)
	err = s.db.QueryRow("SELECT id, status, COALESCE(pharma_no, ''), planned_bom_avg, COALESCE(operator_name, '') FROM cutting_jobs WHERE order_item_id = ?", orderItemID).Scan(
		&detail.Job.ID, &detail.Job.Status, &detail.Job.PharmaNo, &detail.Job.PlannedBomAvg, &detail.Job.OperatorName,
	)
	now := graph.NowMs()
	if err == sql.ErrNoRows {
		// Initialize Job
		detail.Job.ID = expectedJobID
		detail.Job.Status = "OPEN"

		// Query Pharma from Station 2 if exists
		_ = s.db.QueryRow("SELECT COALESCE(pharma_no, '') FROM pharma_patterns WHERE order_item_id = ? LIMIT 1", orderItemID).Scan(&detail.Job.PharmaNo)
		// Query Planned BOM average from Station 2 if exists
		_ = s.db.QueryRow("SELECT COALESCE(consumption_avg, 0.35) FROM bom_lines WHERE order_item_id = ? AND line_type = 'FABRIC' LIMIT 1", orderItemID).Scan(&detail.Job.PlannedBomAvg)
		if detail.Job.PlannedBomAvg <= 0 {
			detail.Job.PlannedBomAvg = 0.35
		}

		_, _ = s.db.Exec(`
			INSERT INTO cutting_jobs (
				id, order_id, order_item_id, production_order_no, product_name,
				color_name, fabric_name, target_qty, pharma_no, planned_bom_avg,
				status, operator_name, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'OPEN', 'Floor Operator', ?, ?)
		`, detail.Job.ID, orderID, orderItemID, detail.Job.ProductionOrderNo,
			detail.Job.ProductName, detail.Job.ColorName, detail.Job.FabricName,
			detail.Job.TargetQty, detail.Job.PharmaNo, detail.Job.PlannedBomAvg, now, now)

		// Bi-temporal graph node
		_, _ = s.graph.EnsureNode(nil, graph.TypeCuttingJob, map[string]any{
			"orderNo": detail.Job.ProductionOrderNo,
			"product": detail.Job.ProductName,
			"target":  detail.Job.TargetQty,
		}, detail.Job.ID)
	}
	detail.PlannedBomAvgKg = detail.Job.PlannedBomAvg

	// 4. Parse Target Matrix & Calculate Size-Wise Counters
	targetSizes := parseSizeMatrix(matrixJSON)

	// Fetch existing bundles cut for this job
	todayStartMs := getStartOfDayMs()
	type sizeCutSum struct {
		prevCut  int
		todayCut int
	}
	cutSums := make(map[string]*sizeCutSum)
	for s := range targetSizes {
		cutSums[s] = &sizeCutSum{}
	}

	bundleRows, err := s.db.Query(`
		SELECT id, cut_lay_id, bundle_sequence, size, pieces,
		       slip_printed_count, next_process, operator_name, created_at
		FROM cut_bundles
		WHERE cutting_job_id = ?
		ORDER BY created_at DESC
	`, detail.Job.ID)
	if err == nil {
		for bundleRows.Next() {
			var b CutBundleRecord
			b.CuttingJobID = detail.Job.ID
			if err := bundleRows.Scan(
				&b.ID, &b.CutLayID, &b.BundleSequence, &b.Size, &b.Pieces,
				&b.SlipPrintedCount, &b.NextProcess, &b.OperatorName, &b.CreatedAt,
			); err == nil {
				detail.Bundles = append(detail.Bundles, b)

				// Accumulate size counts
				cs, ok := cutSums[b.Size]
				if !ok {
					cs = &sizeCutSum{}
					cutSums[b.Size] = cs
				}
				if b.CreatedAt >= todayStartMs {
					cs.todayCut += b.Pieces
				} else {
					cs.prevCut += b.Pieces
				}
				detail.TotalCutQty += b.Pieces
			}
		}
		bundleRows.Close()
	}

	// Build Size Progress List
	detail.TotalBalanceQty = int(math.Max(0, float64(detail.TotalTargetQty-detail.TotalCutQty)))
	for size, tgt := range targetSizes {
		cs := cutSums[size]
		totalCut := cs.prevCut + cs.todayCut
		bal := int(math.Max(0, float64(tgt-totalCut)))
		pct := 0.0
		if tgt > 0 {
			pct = math.Round((float64(totalCut)/float64(tgt))*1000) / 10
		}
		detail.SizeProgressList = append(detail.SizeProgressList, SizeProgress{
			Size:          size,
			TargetQty:     tgt,
			PreviouslyCut: cs.prevCut,
			TodayCut:      cs.todayCut,
			TotalCut:      totalCut,
			Balance:       bal,
			CompletionPct: pct,
		})
	}

	// 5. Fetch Lays & Calculate Actual Average
	layRows, err := s.db.Query(`
		SELECT id, lay_number, roll_code, fabric_name, roll_weight_kg,
		       marker_name, ratio_json, plies, total_pieces_in_lay,
		       lay_weight_kg, actual_lay_avg_kg, planned_bom_avg_kg,
		       efficiency_pct, status, operator_name, created_at
		FROM cut_lays
		WHERE cutting_job_id = ?
		ORDER BY lay_number DESC
	`, detail.Job.ID)
	var totalLayWeight float64
	var totalLayPieces int
	if err == nil {
		for layRows.Next() {
			var l CutLayRecord
			var ratioJSON string
			l.CuttingJobID = detail.Job.ID
			if err := layRows.Scan(
				&l.ID, &l.LayNumber, &l.RollCode, &l.FabricName, &l.RollWeightKg,
				&l.MarkerName, &ratioJSON, &l.Plies, &l.TotalPiecesInLay,
				&l.LayWeightKg, &l.ActualLayAvgKg, &l.PlannedBomAvgKg,
				&l.EfficiencyPct, &l.Status, &l.OperatorName, &l.CreatedAt,
			); err == nil {
				_ = json.Unmarshal([]byte(ratioJSON), &l.Ratio)
				detail.Lays = append(detail.Lays, l)
				totalLayWeight += l.LayWeightKg
				totalLayPieces += l.TotalPiecesInLay
			}
		}
		layRows.Close()
	}
	if totalLayPieces > 0 {
		detail.AverageLayAvgKg = math.Round((totalLayWeight/float64(totalLayPieces))*1000) / 1000
	} else {
		detail.AverageLayAvgKg = detail.PlannedBomAvgKg
	}

	// 6. Fetch Issued Rolls
	rollRows, err := s.db.Query(`
		SELECT id, roll_code, fabric_name, COALESCE(color_name, ''),
		       issued_weight_kg, status, operator_name, issued_at
		FROM cutting_roll_issues
		WHERE cutting_job_id = ?
		ORDER BY issued_at DESC
	`, detail.Job.ID)
	if err == nil {
		for rollRows.Next() {
			var id, code, fab, col, stat, op string
			var wt float64
			var at int64
			if err := rollRows.Scan(&id, &code, &fab, &col, &wt, &stat, &op, &at); err == nil {
				detail.IssuedRolls = append(detail.IssuedRolls, map[string]any{
					"id": id, "rollCode": code, "fabricName": fab, "colorName": col,
					"issuedWeightKg": wt, "status": stat, "operatorName": op, "issuedAt": at,
				})
			}
		}
		rollRows.Close()
	}

	// 7. Fetch Remnants
	remRows, err := s.db.Query(`
		SELECT id, prefix, source_roll_code, weight_kg, bin_location,
		       COALESCE(defect_reason, ''), operator_name, created_at
		FROM remnant_returns
		WHERE cutting_job_id = ?
		ORDER BY created_at DESC
	`, detail.Job.ID)
	if err == nil {
		for remRows.Next() {
			var r RemnantRecord
			r.CuttingJobID = detail.Job.ID
			if err := remRows.Scan(
				&r.ID, &r.Prefix, &r.SourceRollCode, &r.WeightKg, &r.BinLocation,
				&r.DefectReason, &r.OperatorName, &r.CreatedAt,
			); err == nil {
				detail.Remnants = append(detail.Remnants, r)
			}
		}
		remRows.Close()
	}

	return detail, nil
}

// -------------------------------------------------------------
// ISSUE ROLL & RECORD CUT LAY
// -------------------------------------------------------------

// IssueRollToJob records a scanned fabric roll issued to the cutting floor
func (s *Service) IssueRollToJob(jobID, rollCode, operator string) (map[string]any, error) {
	rollCode = strings.TrimSpace(rollCode)
	if rollCode == "" {
		return nil, fmt.Errorf("roll barcode/code is required")
	}
	if operator == "" {
		operator = "Floor Operator"
	}

	// 1. Check roll details from bundles or parent lots
	var fabricName, colorName string
	var weightKg float64
	err := s.db.QueryRow(`
		SELECT b.item_name, COALESCE(b.color_name, ''), b.inward_weight_qty
		FROM bundles b
		WHERE b.bundle_code = ?
		LIMIT 1
	`, rollCode).Scan(&fabricName, &colorName, &weightKg)
	if err != nil {
		// Fallback check in parent_inward_lots
		err = s.db.QueryRow(`
			SELECT COALESCE(fpcc_fab_name, vendor_fab_name), COALESCE(fpcc_color, vendor_color),
			       CASE WHEN counted_qty > 0 THEN counted_qty ELSE physical_received_qty END
			FROM parent_inward_lots
			WHERE parent_code = ?
			LIMIT 1
		`, rollCode).Scan(&fabricName, &colorName, &weightKg)
		if err != nil {
			// If roll not yet inwarded, allow operator entry with fallback
			fabricName = "Super Combed Pique"
			colorName = "Default"
			weightKg = 45.0
		}
	}

	now := graph.NowMs()
	issueID := fmt.Sprintf("cri_%s_%s", rollCode, uuid.New().String()[:6])

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO cutting_roll_issues (
			id, cutting_job_id, roll_code, fabric_name, color_name,
			issued_weight_kg, status, operator_name, issued_at
		) VALUES (?, ?, ?, ?, ?, ?, 'ISSUED', ?, ?)
	`, issueID, jobID, rollCode, fabricName, colorName, weightKg, operator, now)
	if err != nil {
		return nil, fmt.Errorf("failed to record roll issue: %w", err)
	}

	// Update job status to ISSUED
	_, _ = tx.Exec("UPDATE cutting_jobs SET status = 'ISSUED', updated_at = ? WHERE id = ? AND status = 'OPEN'", now, jobID)

	// Bi-Temporal Graph Edge: (cutting_job) -[:CUT_FROM]-> (roll)
	_, _ = s.graph.InsertEdge(tx, jobID, rollCode, graph.EdgeCutFrom, map[string]any{
		"rollWeight": weightKg,
		"issuedAt":   now,
	})

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return map[string]any{
		"success":        true,
		"issueId":        issueID,
		"rollCode":       rollCode,
		"fabricName":     fabricName,
		"colorName":      colorName,
		"issuedWeightKg": weightKg,
	}, nil
}

// RecordCutLay records marker lay ratio, plies, computes actual average, and generates cut bundles
func (s *Service) RecordCutLay(p RecordLayPayload) (*CutLayRecord, error) {
	if p.CuttingJobID == "" {
		return nil, fmt.Errorf("cuttingJobId is required")
	}
	if p.Plies <= 0 {
		return nil, fmt.Errorf("plies must be greater than 0")
	}
	if len(p.Ratio) == 0 {
		return nil, fmt.Errorf("at least one size ratio is required")
	}
	if p.LayWeightKg <= 0 {
		return nil, fmt.Errorf("lay weight must be greater than 0")
	}
	if p.OperatorName == "" {
		p.OperatorName = "Floor Operator"
	}
	if p.NextProcess == "" {
		p.NextProcess = "Branding"
	}

	// 1. Calculate Pieces and Averages
	piecesPerLay := 0
	for _, r := range p.Ratio {
		piecesPerLay += r
	}
	if piecesPerLay <= 0 {
		return nil, fmt.Errorf("sum of ratio values must be greater than 0")
	}
	totalPieces := p.Plies * piecesPerLay
	actualLayAvg := math.Round((p.LayWeightKg/float64(totalPieces))*1000) / 1000

	// Fetch planned BOM average from job
	var plannedAvg float64
	var currentLayCount int
	err := s.db.QueryRow("SELECT planned_bom_avg FROM cutting_jobs WHERE id = ?", p.CuttingJobID).Scan(&plannedAvg)
	if err != nil || plannedAvg <= 0 {
		plannedAvg = 0.35
	}
	_ = s.db.QueryRow("SELECT COUNT(*) FROM cut_lays WHERE cutting_job_id = ?", p.CuttingJobID).Scan(&currentLayCount)
	layNumber := currentLayCount + 1

	efficiencyPct := 100.0
	if actualLayAvg > 0 {
		efficiencyPct = math.Round((plannedAvg/actualLayAvg)*1000) / 10
	}

	now := graph.NowMs()
	layID := fmt.Sprintf("CL-%s-%02d", p.CuttingJobID, layNumber)
	ratioJSON, _ := json.Marshal(p.Ratio)

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 2. Insert Cut Lay
	_, err = tx.Exec(`
		INSERT INTO cut_lays (
			id, cutting_job_id, lay_number, roll_code, fabric_name,
			roll_weight_kg, marker_name, ratio_json, plies, total_pieces_in_lay,
			lay_weight_kg, actual_lay_avg_kg, planned_bom_avg_kg, efficiency_pct,
			status, operator_name, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'CUT', ?, ?)
	`, layID, p.CuttingJobID, layNumber, p.RollCode, p.FabricName,
		p.RollWeightKg, p.MarkerName, string(ratioJSON), p.Plies, totalPieces,
		p.LayWeightKg, actualLayAvg, plannedAvg, efficiencyPct, p.OperatorName, now)
	if err != nil {
		return nil, fmt.Errorf("failed to insert cut lay: %w", err)
	}

	// 3. Generate Cut Bundles for each size in the lay
	bundleSeq := 1
	for size, ratioVal := range p.Ratio {
		if ratioVal <= 0 {
			continue
		}
		sizePieces := ratioVal * p.Plies
		bundleID := fmt.Sprintf("CB-%s-%02d-%s", time.Now().Format("060102"), (now%100)+int64(bundleSeq), size)

		_, err = tx.Exec(`
			INSERT INTO cut_bundles (
				id, cutting_job_id, cut_lay_id, bundle_sequence, size,
				pieces, slip_printed_count, next_process, operator_name, created_at
			) VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?, ?)
		`, bundleID, p.CuttingJobID, layID, bundleSeq, size, sizePieces, p.NextProcess, p.OperatorName, now)
		if err != nil {
			return nil, fmt.Errorf("failed to generate cut bundle for size %s: %w", size, err)
		}

		// Graph Node & Edge
		_, _ = s.graph.EnsureNode(tx, graph.TypeCutBundle, map[string]any{
			"size":   size,
			"pieces": sizePieces,
		}, bundleID)
		_, _ = s.graph.InsertEdge(tx, layID, bundleID, graph.EdgeYieldsBundle, map[string]any{
			"pieces": sizePieces,
		})

		bundleSeq++
	}

	// 4. Update Job Status to CUT
	_, _ = tx.Exec("UPDATE cutting_jobs SET status = 'CUT', updated_at = ? WHERE id = ?", now, p.CuttingJobID)

	// Bi-Temporal Graph Node for Lay
	_, _ = s.graph.EnsureNode(tx, graph.TypeCutLay, map[string]any{
		"layNumber":   layNumber,
		"totalPieces": totalPieces,
		"layAvg":      actualLayAvg,
	}, layID)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &CutLayRecord{
		ID:               layID,
		CuttingJobID:     p.CuttingJobID,
		LayNumber:        layNumber,
		RollCode:         p.RollCode,
		FabricName:       p.FabricName,
		RollWeightKg:     p.RollWeightKg,
		MarkerName:       p.MarkerName,
		Ratio:            p.Ratio,
		Plies:            p.Plies,
		TotalPiecesInLay: totalPieces,
		LayWeightKg:      p.LayWeightKg,
		ActualLayAvgKg:   actualLayAvg,
		PlannedBomAvgKg:  plannedAvg,
		EfficiencyPct:    efficiencyPct,
		Status:           "CUT",
		OperatorName:     p.OperatorName,
		CreatedAt:        now,
	}, nil
}

// -------------------------------------------------------------
// REMNANT & END-BIT HANDLER ('Bit' / 'Rej' prefixes)
// -------------------------------------------------------------

// RecordRemnant logs an end-bit or defect return to inventory
func (s *Service) RecordRemnant(p RecordRemnantPayload) (*RemnantRecord, error) {
	prefix := strings.TrimSpace(p.Prefix)
	if prefix != "Bit" && prefix != "Rej" {
		return nil, fmt.Errorf("prefix must be strictly 'Bit' (usable remnant) or 'Rej' (rejection)")
	}
	cleanRoll := strings.ReplaceAll(p.SourceRollCode, "-", "")
	cleanRoll = strings.ReplaceAll(cleanRoll, " ", "")
	cleanRoll = strings.ReplaceAll(cleanRoll, "/", "_")
	if cleanRoll == "" {
		cleanRoll = time.Now().Format("060102")
	}

	// Format required by Arjav notes: Bit240426A-1/6 or Rej240426A-1/6
	remnantID := fmt.Sprintf("%s%s", prefix, p.SourceRollCode)
	if p.WeightKg <= 0 {
		return nil, fmt.Errorf("remnant weight must be greater than 0")
	}
	if p.BinLocation == "" {
		p.BinLocation = "4.Cutting Room"
	}
	if p.OperatorName == "" {
		p.OperatorName = "Floor Operator"
	}

	now := graph.NowMs()
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO remnant_returns (
			id, prefix, source_roll_code, cutting_job_id, cut_lay_id,
			weight_kg, bin_location, defect_reason, operator_name, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			weight_kg = excluded.weight_kg,
			bin_location = excluded.bin_location,
			defect_reason = excluded.defect_reason,
			created_at = excluded.created_at
	`, remnantID, prefix, p.SourceRollCode, p.CuttingJobID, p.CutLayID,
		p.WeightKg, p.BinLocation, p.DefectReason, p.OperatorName, now)
	if err != nil {
		return nil, fmt.Errorf("failed to record remnant return: %w", err)
	}

	// Graph Node & Edge
	_, _ = s.graph.EnsureNode(tx, graph.TypeRemnantReturn, map[string]any{
		"prefix":   prefix,
		"source":   p.SourceRollCode,
		"weightKg": p.WeightKg,
		"bin":      p.BinLocation,
	}, remnantID)

	if p.CutLayID != "" {
		_, _ = s.graph.InsertEdge(tx, p.CutLayID, remnantID, graph.EdgeYieldsRemnant, map[string]any{
			"weightKg": p.WeightKg,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &RemnantRecord{
		ID:             remnantID,
		Prefix:         prefix,
		SourceRollCode: p.SourceRollCode,
		CuttingJobID:   p.CuttingJobID,
		WeightKg:       p.WeightKg,
		BinLocation:    p.BinLocation,
		DefectReason:   p.DefectReason,
		OperatorName:   p.OperatorName,
		CreatedAt:      now,
	}, nil
}

// -------------------------------------------------------------
// BUNDLE SLIP PRINTING
// -------------------------------------------------------------

// PrintBundleSlip increments print count and returns docket details for label printing
func (s *Service) PrintBundleSlip(bundleID string) (*BundleSlipData, error) {
	now := graph.NowMs()

	var data BundleSlipData
	var jobID string
	err := s.db.QueryRow(`
		SELECT cb.id, cb.cutting_job_id, cb.size, cb.pieces, cb.slip_printed_count,
		       cb.next_process, cb.operator_name, cl.lay_number
		FROM cut_bundles cb
		JOIN cut_lays cl ON cl.id = cb.cut_lay_id
		WHERE cb.id = ?
	`, bundleID).Scan(
		&data.BundleCode, &jobID, &data.Size, &data.Pieces, &data.SlipPrintedCount,
		&data.NextProcess, &data.OperatorName, &data.LayNumber,
	)
	if err != nil {
		return nil, fmt.Errorf("bundle %s not found: %w", bundleID, err)
	}

	_ = s.db.QueryRow(`
		SELECT production_order_no, product_name, color_name
		FROM cutting_jobs
		WHERE id = ?
	`, jobID).Scan(&data.ProductionOrderNo, &data.ProductName, &data.ColorName)

	// Increment print counter
	data.SlipPrintedCount++
	_, _ = s.db.Exec("UPDATE cut_bundles SET slip_printed_count = ?, printed_at = ? WHERE id = ?", data.SlipPrintedCount, now, bundleID)

	data.DateStr = time.Now().Format("02/01/2006 15:04")
	data.QRData = fmt.Sprintf("FPCC|CUT|%s|ORD:%s|SIZE:%s|QTY:%d|NEXT:%s",
		data.BundleCode, data.ProductionOrderNo, data.Size, data.Pieces, data.NextProcess)

	return &data, nil
}

// -------------------------------------------------------------
// HELPER UTILITIES
// -------------------------------------------------------------

func parseSizeMatrix(matrixJSON string) map[string]int {
	result := make(map[string]int)
	if matrixJSON == "" || matrixJSON == "{}" {
		return result
	}

	var raw map[string]any
	if err := json.Unmarshal([]byte(matrixJSON), &raw); err != nil {
		return result
	}

	// Check if nested by gender (e.g. {"Men": {"S": 10, "M": 20}, "Women": {"S": 5}})
	for _, val := range raw {
		if innerMap, ok := val.(map[string]any); ok {
			for sz, qtyVal := range innerMap {
				q := toInt(qtyVal)
				if q > 0 {
					result[sz] += q
				}
			}
		} else {
			// Flat map: {"S": 10, "M": 20}
			for sz, qtyVal := range raw {
				q := toInt(qtyVal)
				if q > 0 {
					result[sz] = q
				}
			}
			break
		}
	}
	return result
}

func toInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	default:
		return 0
	}
}

func getStartOfDayMs() int64 {
	t := time.Now()
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return start.UnixMilli()
}
