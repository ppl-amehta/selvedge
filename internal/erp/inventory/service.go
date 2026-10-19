package inventory

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
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

type BundleInput struct {
	BundleSequence  int     `json:"bundleSequence"`
	BundleNo        int     `json:"bundleNo"`
	TotalBundles    int     `json:"totalBundles"`
	ItemType        string  `json:"itemType"`
	ItemName        string  `json:"itemName"`
	ColorName       string  `json:"colorName"`
	InwardWeightQty float64 `json:"inwardWeightQty"`
	MeasuredWeight  float64 `json:"measuredWeight"`
	MeasuredGsm     float64 `json:"measuredGsm"`
	LocationBin     string  `json:"locationBin"`
	Location        string  `json:"location"`
	QcStatus        string  `json:"qcStatus"`
	QcRemarks       string  `json:"qcRemarks"`
	Remarks         string  `json:"remarks"`
}

type InwardParentLotPayload struct {
	ParentCode          string        `json:"parentCode"`
	ParentLotCode       string        `json:"parentLotCode"`
	InwardDate          string        `json:"inwardDate"`
	VendorID            string        `json:"vendorId"`
	VendorName          string        `json:"vendorName"`
	InvoiceNo           string        `json:"invoiceNo"`
	InwardType          string        `json:"inwardType"` // Strict enum: Fabric, Zips, Packaging, Trims, Labels, Finished Goods
	UOM                 string        `json:"uom"`
	InvoiceQty          float64       `json:"invoiceQty"`
	InvoiceAmount       float64       `json:"invoiceAmount"`
	PhysicalReceivedQty float64       `json:"physicalReceivedQty"`
	VendorColor         string        `json:"vendorColor"`
	FpccColor           string        `json:"fpccColor"`
	FabricType          string        `json:"fabricType"`
	FpccFabName         string        `json:"fpccFabName"`
	VendorFabName       string        `json:"vendorFabName"`
	InvoiceGsm          float64       `json:"invoiceGsm"`
	Dia                 float64       `json:"dia"`
	Remarks             string        `json:"remarks"`
	Bundles             []BundleInput `json:"bundles"`
}

// SuggestNextParentCode generates format: YYMMDD-A
func (s *Service) SuggestNextParentCode() (string, error) {
	now := time.Now()
	prefix := now.Format("060102") + "-"

	var lastCode string
	err := s.db.QueryRow(`
		SELECT parent_code FROM parent_inward_lots
		WHERE parent_code LIKE ?
		ORDER BY parent_code DESC LIMIT 1
	`, prefix+"%").Scan(&lastCode)

	if err == sql.ErrNoRows {
		return prefix + "A", nil
	} else if err != nil {
		return prefix + "A", nil
	}

	nextChar := byte('A')
	if len(lastCode) > len(prefix) {
		lastChar := lastCode[len(prefix):]
		if len(lastChar) > 0 {
			nextChar = lastChar[0] + 1
		}
	}

	return fmt.Sprintf("%s%c", prefix, nextChar), nil
}

// InwardParentLot executes Step A & Step B, detects discrepancies, writes graph nodes and edges
func (s *Service) InwardParentLot(p InwardParentLotPayload) (map[string]any, error) {
	parentCode := strings.TrimSpace(p.ParentCode)
	if parentCode == "" {
		parentCode = strings.TrimSpace(p.ParentLotCode)
	}
	if parentCode == "" {
		var err error
		parentCode, err = s.SuggestNextParentCode()
		if err != nil {
			return nil, err
		}
	}

	vendorName := strings.TrimSpace(p.VendorID)
	if vendorName == "" {
		vendorName = strings.TrimSpace(p.VendorName)
	}
	if vendorName == "" || p.InvoiceQty <= 0 {
		return nil, fmt.Errorf("parentCode, vendorName, and positive invoiceQty are required")
	}

	inwardType := p.InwardType
	if inwardType == "" {
		inwardType = "Fabric"
	}
	uom := p.UOM
	if uom == "" {
		uom = "Kgs"
	}
	inwardDate := p.InwardDate
	if inwardDate == "" {
		inwardDate = time.Now().Format("2006-01-02")
	}

	now := graph.NowMs()
	bundleCount := len(p.Bundles)
	if bundleCount == 0 {
		bundleCount = 1
		weight := p.PhysicalReceivedQty
		if weight <= 0 {
			weight = p.InvoiceQty
		}
		p.Bundles = []BundleInput{{BundleNo: 1, MeasuredWeight: weight, Location: "A/R01/S01", QcStatus: "Passed"}}
	}

	// 1. Calculate physical sum across Step B bundles
	var physicalTotal float64
	for i := range p.Bundles {
		if p.Bundles[i].MeasuredWeight == 0 {
			p.Bundles[i].MeasuredWeight = p.Bundles[i].InwardWeightQty
		}
		if p.Bundles[i].Location == "" {
			p.Bundles[i].Location = p.Bundles[i].LocationBin
		}
		if p.Bundles[i].Remarks == "" {
			p.Bundles[i].Remarks = p.Bundles[i].QcRemarks
		}
		physicalTotal += p.Bundles[i].MeasuredWeight
	}
	if physicalTotal == 0 && p.PhysicalReceivedQty > 0 {
		physicalTotal = p.PhysicalReceivedQty
	}
	physicalTotal = math.Round(physicalTotal*1000) / 1000
	shortfallQty := math.Round((p.InvoiceQty-physicalTotal)*1000) / 1000

	unitRate := 0.0
	if p.InvoiceAmount > 0 && p.InvoiceQty > 0 {
		unitRate = math.Round((p.InvoiceAmount/p.InvoiceQty)*100) / 100
	}
	discrepancyAmt := math.Round(math.Abs(shortfallQty)*unitRate*100) / 100

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 2. Resolve or Auto-Register Vendor Entity
	var vendorID string
	err = tx.QueryRow("SELECT id FROM master_entities WHERE name = ? AND (entity_type = 'vendor' OR entity_type = 'both')", vendorName).Scan(&vendorID)
	if err == sql.ErrNoRows {
		vendorID = fmt.Sprintf("ent_vendor_%d", now%100000)
		_, err = tx.Exec(`
			INSERT INTO master_entities (id, name, entity_type, contact_person, address, created_at)
			VALUES (?, ?, 'vendor', 'Primary Vendor Contact', 'Direct Inward Supplier', ?)
		`, vendorID, vendorName, now)
		if err != nil {
			return nil, fmt.Errorf("failed to register vendor entity: %w", err)
		}
	}

	// 3. Insert Parent Lot Projection
	parentLotID := fmt.Sprintf("plt_%s", parentCode)
	materialName := p.FpccFabName
	if materialName == "" {
		materialName = p.VendorFabName
	}
	if materialName == "" {
		materialName = p.FabricType
	}
	if materialName == "" {
		materialName = inwardType
	}

	colorName := p.FpccColor
	if colorName == "" {
		colorName = p.VendorColor
	}
	if colorName == "" {
		colorName = "Standard"
	}

	_, err = tx.Exec(`
		INSERT INTO parent_inward_lots (
			id, parent_code, inward_date, vendor_id, vendor_name, invoice_no, inward_type,
			uom, invoice_qty, invoice_amount, physical_received_qty, shortfall_qty, discrepancy_amount,
			fabric_type, vendor_fab_name, fpcc_fab_name, vendor_color, fpcc_color,
			invoice_gsm, dia, remarks, created_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?
		)
	`, parentLotID, parentCode, inwardDate, vendorID, vendorName, p.InvoiceNo, inwardType,
		uom, p.InvoiceQty, p.InvoiceAmount, physicalTotal, shortfallQty, discrepancyAmt,
		p.FabricType, p.VendorFabName, p.FpccFabName, p.VendorColor, p.FpccColor,
		p.InvoiceGsm, p.Dia, p.Remarks, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert parent lot projection: %w", err)
	}

	// 4. Insert Graph Node: PARENT_LOT
	parentProps := map[string]any{
		"parent_code":   parentCode,
		"invoice_no":    p.InvoiceNo,
		"inward_type":   inwardType,
		"material_name": materialName,
		"color_name":    colorName,
		"invoice_qty":   p.InvoiceQty,
		"actual_qty":    physicalTotal,
		"shortfall_qty": shortfallQty,
	}
	_, err = s.graph.InsertNode(tx, graph.TypeParentLot, parentProps, parentLotID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert graph node for parent lot: %w", err)
	}

	// 5. Insert Graph Edge: (PARENT_LOT) -[:SUPPLIED_BY]-> (VENDOR)
	_, err = s.graph.InsertEdge(tx, parentLotID, vendorID, graph.EdgeSuppliedBy, map[string]any{
		"invoice_no": p.InvoiceNo,
		"inward_qty": physicalTotal,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert graph edge SUPPLIED_BY: %w", err)
	}

	// 6. Step B: Generate Child Bundles
	var createdBundles []map[string]any
	for i, b := range p.Bundles {
		bundleSeq := i + 1
		bundleCode := fmt.Sprintf("%s-%d/%d", parentCode, bundleSeq, bundleCount)
		bundleID := fmt.Sprintf("bnd_%s_%d", parentCode, bundleSeq)

		loc := strings.TrimSpace(b.Location)
		if loc == "" {
			loc = strings.TrimSpace(b.LocationBin)
		}
		if loc == "" {
			loc = "A/R01/S01"
		}
		qc := b.QcStatus
		if qc == "" {
			qc = "Passed"
		}

		// Auto-register open-ended location in master_items if new
		tx.Exec(`
			INSERT INTO master_items (id, category, code, display_name, meta_json, created_at)
			VALUES (?, 'location', ?, ?, '{}', ?)
			ON CONFLICT(category, code) DO NOTHING
		`, fmt.Sprintf("loc_%s", loc), strings.ToUpper(loc), strings.ToUpper(loc), now)

		qrPayloadJSON, _ := json.Marshal(map[string]any{
			"bundleCode":  bundleCode,
			"parentCode":  parentCode,
			"item":        materialName,
			"color":       colorName,
			"weight":      b.MeasuredWeight,
			"location":    loc,
			"qc":          qc,
			"generatedAt": time.Now().Format(time.RFC3339),
		})

		_, err = tx.Exec(`
			INSERT INTO bundles (
				id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
				item_type, item_name, color_name, inward_weight_qty, measured_gsm,
				location_bin, qc_status, qc_remarks, qr_payload, available_qty, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, bundleID, bundleCode, parentLotID, parentCode, bundleSeq, bundleCount,
			inwardType, materialName, colorName, b.MeasuredWeight, b.MeasuredGsm,
			loc, qc, b.Remarks, string(qrPayloadJSON), b.MeasuredWeight, now,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert bundle %s: %w", bundleCode, err)
		}

		// Insert Graph Node: BUNDLE
		bundleProps := map[string]any{
			"bundle_code": bundleCode,
			"weight":      b.MeasuredWeight,
			"gsm":         b.MeasuredGsm,
			"location":    loc,
			"qc_status":   qc,
		}
		_, _ = s.graph.InsertNode(tx, graph.TypeBundle, bundleProps, bundleID)

		// Insert Graph Edge: (BUNDLE) -[:PART_OF_LOT]-> (PARENT_LOT)
		_, _ = s.graph.InsertEdge(tx, bundleID, parentLotID, graph.EdgePartOfLot, map[string]any{"seq": bundleSeq})

		createdBundles = append(createdBundles, map[string]any{
			"bundleCode": bundleCode,
			"weight":     b.MeasuredWeight,
			"location":   loc,
			"qcStatus":   qc,
		})
	}

	// 7. Discrepancy Reporting
	var discrepancyInfo map[string]any
	if math.Abs(shortfallQty) > 0.001 {
		discReportID := fmt.Sprintf("disc_%s", parentCode)
		discReportNo := fmt.Sprintf("DISC-%s", parentCode)
		variancePct := math.Round((shortfallQty/p.InvoiceQty)*10000) / 100

		status := "SHORTAGE_FLAGGED"
		discRemarks := fmt.Sprintf("Invoice qty %.3f %s exceeds physical received %.3f %s (Shortage: %.3f)", p.InvoiceQty, uom, physicalTotal, uom, shortfallQty)
		if shortfallQty < 0 {
			status = "EXCESS_FLAGGED"
			discRemarks = fmt.Sprintf("Physical received %.3f %s exceeds invoice qty %.3f %s (Excess: %.3f)", physicalTotal, uom, p.InvoiceQty, uom, math.Abs(shortfallQty))
		}

		_, err = tx.Exec(`
			INSERT INTO discrepancy_reports (
				id, report_no, parent_code, vendor_name, invoice_no, inward_type,
				invoice_qty, actual_qty, variance_qty, variance_pct, financial_impact,
				status, remarks, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, discReportID, discReportNo, parentCode, vendorName, p.InvoiceNo, inwardType,
			p.InvoiceQty, physicalTotal, shortfallQty, variancePct, discrepancyAmt,
			status, discRemarks, now,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to log discrepancy report: %w", err)
		}

		discrepancyType := "SHORTAGE"
		if shortfallQty < 0 {
			discrepancyType = "EXCESS"
		}
		discrepancyInfo = map[string]any{
			"reportNo":        discReportNo,
			"type":            discrepancyType,
			"shortfallQty":    shortfallQty,
			"differenceQty":   -shortfallQty,
			"variancePct":     variancePct,
			"financialImpact": discrepancyAmt,
			"status":          status,
		}
	}

	// 8. Emit Event to Workflow Outbox (Triggers shortage rule if shortfallQty > 0)
	_, err = s.workflow.EmitEvent(tx, parentLotID, "INWARD_LOT_CREATED", map[string]any{
		"parent_code":     parentCode,
		"vendor_name":     vendorName,
		"inward_type":     inwardType,
		"invoice_qty":     p.InvoiceQty,
		"actual_qty":      physicalTotal,
		"shortfall_qty":   shortfallQty,
		"financialImpact": discrepancyAmt,
	}, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to emit INWARD_LOT_CREATED event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Trigger async event dispatch
	go func() {
		_, _ = s.workflow.ProcessPendingEvents()
	}()

	return map[string]any{
		"success":     true,
		"parentCode":  parentCode,
		"bundles":     createdBundles,
		"discrepancy": discrepancyInfo,
	}, nil
}

// ReturnRemnant processes end-of-roll bit returns or defect reject returns with 'Bit' and 'Rej' prefixes
func (s *Service) ReturnRemnant(bundleCode, returnType string, returnedWeightKg float64, returnLocation, remarks string) (map[string]any, error) {
	cleanCode := strings.TrimSpace(bundleCode)
	if cleanCode == "" || returnedWeightKg <= 0 {
		return nil, fmt.Errorf("valid bundleCode and returnedWeightKg > 0 required")
	}

	loc := strings.TrimSpace(returnLocation)
	if loc == "" {
		loc = "1.M/R/O"
	}

	prefix := ""
	rType := strings.ToLower(returnType)
	eventType := "ROLL_RETURNED_AS_REMNANT"

	if rType == "bits" {
		prefix = "Bit"
		eventType = "ROLL_RETURNED_AS_BITS"
	} else if rType == "reject" {
		prefix = "Rej"
		eventType = "ROLL_RETURNED_AS_REJECT"
	}

	updatedBundleCode := fmt.Sprintf("%s%s", prefix, cleanCode)

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Update bundle available quantity and location
	res, err := tx.Exec(`
		UPDATE bundles SET available_qty = ?, location_bin = ?
		WHERE bundle_code = ?
	`, returnedWeightKg, loc, cleanCode)
	if err != nil {
		return nil, err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("bundle %s not found in inventory", cleanCode)
	}

	// Emit workflow event
	_, _ = s.workflow.EmitEvent(tx, cleanCode, eventType, map[string]any{
		"bundle_code":         cleanCode,
		"updated_bundle_code": updatedBundleCode,
		"return_type":         returnType,
		"returned_weight_kg":  returnedWeightKg,
		"location":            loc,
		"remarks":             remarks,
	}, 0)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	go func() {
		_, _ = s.workflow.ProcessPendingEvents()
	}()

	return map[string]any{
		"success":           true,
		"originalBundleCode": cleanCode,
		"updatedBundleCode": updatedBundleCode,
		"returnType":        returnType,
		"returnedWeightKg":  returnedWeightKg,
		"location":          loc,
	}, nil
}

// ListDiscrepancies returns recorded discrepancy reports
func (s *Service) ListDiscrepancies() ([]map[string]any, error) {
	rows, err := s.db.Query(`
		SELECT report_no, parent_code, vendor_name, invoice_no, inward_type,
		       invoice_qty, actual_qty, variance_qty, financial_impact, status, remarks
		FROM discrepancy_reports ORDER BY created_at DESC LIMIT 50
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []map[string]any
	for rows.Next() {
		var repNo, pCode, vName, invNo, inwType, status, rem string
		var invQty, actQty, diffQty, finImpact float64
		_ = rows.Scan(&repNo, &pCode, &vName, &invNo, &inwType, &invQty, &actQty, &diffQty, &finImpact, &status, &rem)
		list = append(list, map[string]any{
			"report_no":          repNo,
			"parent_code":        pCode,
			"vendor_name":        vName,
			"invoice_no":         invNo,
			"inward_type":        inwType,
			"invoice_qty":        invQty,
			"actual_received_qty": actQty,
			"discrepancy_qty":    diffQty,
			"financial_impact":   finImpact,
			"status":             status,
			"remarks":            rem,
		})
	}
	return list, nil
}
