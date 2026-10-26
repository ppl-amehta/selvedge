package inventory

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ppl-amehta/selvedge/internal/graph"
	"github.com/ppl-amehta/selvedge/internal/workflow"
)

type LotStage string

const (
	StageAwaitingCount   LotStage = "AWAITING_COUNT"
	StageAwaitingQC      LotStage = "AWAITING_QC"
	StageDiscrepancyHold LotStage = "DISCREPANCY_HOLD"
	StageQCComplete      LotStage = "QC_COMPLETE"
	StageClosed          LotStage = "CLOSED"
)

func randomSuffix() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type Service struct {
	db       *sql.DB
	graph    *graph.Store
	workflow *workflow.Engine
}

func NewService(db *sql.DB, graphStore *graph.Store, wf *workflow.Engine) *Service {
	return &Service{db: db, graph: graphStore, workflow: wf}
}

// -------------------------------------------------------------
// 3-STAGE INWARD PIPELINE STRUCTS
// -------------------------------------------------------------

type GateBundleStub struct {
	Sequence int     `json:"sequence"`
	SizeSpec string  `json:"sizeSpec,omitempty"`
	Rate     float64 `json:"rate,omitempty"`
}

type GateInwardPayload struct {
	ParentCode    string           `json:"parentCode"`
	InwardDate    string           `json:"inwardDate"`
	InwardType    string           `json:"inwardType"` // Fabric, Zips, Packaging, Trims, Finished Goods, Labels
	VendorName    string           `json:"vendorName"`
	InvoiceNo     string           `json:"invoiceNo"`
	InvoiceQty    float64          `json:"invoiceQty"`
	InvoiceAmount float64          `json:"invoiceAmount"`
	UOM           string           `json:"uom"`
	VendorProduct string           `json:"vendorProduct"`
	FpccProduct   string           `json:"fpccProduct"`
	VendorColor   string           `json:"vendorColor"`
	FpccColor     string           `json:"fpccColor"`
	NoOfBundles   int              `json:"noOfBundles"`
	Comment       string           `json:"comment"`
	Attrs         map[string]any   `json:"attrs"`   // Conditional specs
	Bundles       []GateBundleStub `json:"bundles"` // Optional bundle-level size & rate
}

type BundleStubInfo struct {
	BundleCode string  `json:"bundleCode"`
	Sequence   int     `json:"sequence"`
	Total      int     `json:"total"`
	SizeSpec   string  `json:"sizeSpec,omitempty"`
	Rate       float64 `json:"rate,omitempty"`
	Status     string  `json:"status"`
}

type GateInwardResult struct {
	Success    bool             `json:"success"`
	ParentCode string           `json:"parentCode"`
	Stage      LotStage         `json:"stage"`
	Bundles    []BundleStubInfo `json:"bundles"`
}

type FloorCountBundle struct {
	BundleCode  string  `json:"bundleCode"`
	Sequence    int     `json:"sequence"`
	CountedQty  float64 `json:"countedQty"`
	LocationBin string  `json:"locationBin"`
	MeasuredGsm float64 `json:"measuredGsm,omitempty"`
	MeasuredDia float64 `json:"measuredDia,omitempty"`
	SizeSpec    string  `json:"sizeSpec,omitempty"`
	Rate        float64 `json:"rate,omitempty"`
	Comment     string  `json:"comment,omitempty"`
}

type FloorCountPayload struct {
	ParentCode  string             `json:"parentCode"`
	InvoicedQty float64            `json:"invoicedQty"`
	CountedQty  float64            `json:"countedQty"`
	LocationBin string             `json:"locationBin"`
	CountedBy   string             `json:"countedBy"`
	Comments    string             `json:"comments"`
	HoldOnShort bool               `json:"holdOnShort"`
	SizeLines   []map[string]any   `json:"sizeLines,omitempty"`
	Bundles     []FloorCountBundle `json:"bundles"`
}

type FloorCountResult struct {
	Success         bool     `json:"success"`
	ParentCode      string   `json:"parentCode"`
	Stage           LotStage `json:"stage"`
	InvoicedQty     float64  `json:"invoicedQty"`
	CountedQty      float64  `json:"countedQty"`
	VarianceQty     float64  `json:"varianceQty"`
	VariancePct     float64  `json:"variancePct"`
	FinancialImpact float64  `json:"financialImpact"`
	Status          string   `json:"status"` // 'MATCHED', 'SHORTAGE_FLAGGED', 'EXCESS_FLAGGED'
	BundlesUpdated  int      `json:"bundlesUpdated"`
}

type BundleQCItem struct {
	BundleCode  string `json:"bundleCode"`
	Result      string `json:"result"` // 'PASS' or 'FAIL'
	Remarks     string `json:"remarks"`
	ImageURL    string `json:"imageUrl"`
	InspectedBy string `json:"inspectedBy"`
}

type BundleQCPayload struct {
	ParentCode  string         `json:"parentCode"`
	Inspector   string         `json:"inspector"`
	Inspections []BundleQCItem `json:"inspections"`
}

type BundleQCResult struct {
	Success        bool     `json:"success"`
	ParentCode     string   `json:"parentCode"`
	Stage          LotStage `json:"stage"`
	InspectedCount int      `json:"inspectedCount"`
	TotalBundles   int      `json:"totalBundles"`
	LotCompleted   bool     `json:"lotCompleted"`
	PassCount      int      `json:"passCount"`
	FailCount      int      `json:"failCount"`
}

type LotSummary struct {
	ID               string         `json:"id"`
	ParentCode       string         `json:"parentCode"`
	InwardDate       string         `json:"inwardDate"`
	InwardType       string         `json:"inwardType"`
	VendorName       string         `json:"vendorName"`
	InvoiceNo        string         `json:"invoiceNo"`
	InvoiceQty       float64        `json:"invoiceQty"`
	InvoiceAmount    float64        `json:"invoiceAmount"`
	CountedQty       float64        `json:"countedQty"`
	VarianceQty      float64        `json:"varianceQty"`
	VariancePct      float64        `json:"variancePct"`
	UOM              string         `json:"uom"`
	VendorProduct    string         `json:"vendorProduct"`
	FpccProduct      string         `json:"fpccProduct"`
	VendorColor      string         `json:"vendorColor"`
	FpccColor        string         `json:"fpccColor"`
	Stage            string         `json:"stage"`
	NoOfBundles      int            `json:"noOfBundles"`
	InspectedBundles int            `json:"inspectedBundles"`
	PassedBundles    int            `json:"passedBundles"`
	RejectedBundles  int            `json:"rejectedBundles"`
	Remarks          string         `json:"remarks"`
	Attrs            map[string]any `json:"attrs"`
	CreatedAt        int64          `json:"createdAt"`
}

type BundleDetail struct {
	ID             string  `json:"id"`
	BundleCode     string  `json:"bundleCode"`
	ParentCode     string  `json:"parentCode"`
	BundleSequence int     `json:"bundleSequence"`
	TotalBundles   int     `json:"totalBundles"`
	ItemType       string  `json:"itemType"`
	ItemName       string  `json:"itemName"`
	ColorName      string  `json:"colorName"`
	InwardQty      float64 `json:"inwardQty"`
	MeasuredGsm    float64 `json:"measuredGsm"`
	MeasuredDia    float64 `json:"measuredDia"`
	SizeSpec       string  `json:"sizeSpec"`
	Rate           float64 `json:"rate"`
	LocationBin    string  `json:"locationBin"`
	QcStatus       string  `json:"qcStatus"`
	QcRemarks      string  `json:"qcRemarks"`
	QcImageUrl     string  `json:"qcImageUrl"`
	QcInspectedAt  int64   `json:"qcInspectedAt"`
	BundleComment  string  `json:"bundleComment"`
	QRPayload      string  `json:"qrPayload"`
}

type LotDetail struct {
	Lot             LotSummary       `json:"lot"`
	Bundles         []BundleDetail   `json:"bundles"`
	Reconciliations []map[string]any `json:"reconciliations"`
	Inspections     []map[string]any `json:"inspections"`
}

// -------------------------------------------------------------
// LEGACY COMPATIBILITY STRUCTS
// -------------------------------------------------------------

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
	InwardType          string        `json:"inwardType"`
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

// -------------------------------------------------------------
// CODE GENERATION
// -------------------------------------------------------------

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
		return "", fmt.Errorf("failed to query last parent code: %w", err)
	}

	nextChar := byte('A')
	if len(lastCode) > len(prefix) {
		lastChar := lastCode[len(prefix):]
		if len(lastChar) > 0 {
			if lastChar[0] >= 'A' && lastChar[0] < 'Z' {
				nextChar = lastChar[0] + 1
			} else {
				return fmt.Sprintf("%s%s", prefix, randomSuffix()[:2]), nil
			}
		}
	}

	return fmt.Sprintf("%s%c", prefix, nextChar), nil
}

// -------------------------------------------------------------
// STATION 1A: GATE ENTRY & INWARD BOOKING
// -------------------------------------------------------------

func (s *Service) GateInward(p GateInwardPayload) (GateInwardResult, error) {
	parentCode := strings.TrimSpace(p.ParentCode)
	if parentCode == "" {
		var err error
		parentCode, err = s.SuggestNextParentCode()
		if err != nil {
			return GateInwardResult{}, fmt.Errorf("failed to generate parent code: %w", err)
		}
	}

	vendorName := strings.TrimSpace(p.VendorName)
	if vendorName == "" || p.InvoiceQty <= 0 {
		return GateInwardResult{}, fmt.Errorf("vendorName and positive invoiceQty (> 0) are required")
	}

	invoiceNo := strings.TrimSpace(p.InvoiceNo)
	if invoiceNo == "" {
		return GateInwardResult{}, fmt.Errorf("invoiceNo is required")
	}

	inwardType := strings.TrimSpace(p.InwardType)
	if inwardType == "" {
		inwardType = "Fabric"
	}

	uom := strings.TrimSpace(p.UOM)
	if uom == "" {
		uom = "Kgs"
	}

	inwardDate := strings.TrimSpace(p.InwardDate)
	if inwardDate == "" {
		inwardDate = time.Now().Format("2006-01-02")
	}

	numBundles := p.NoOfBundles
	if numBundles <= 0 {
		numBundles = 1
	}

	unitRate := 0.0
	if p.InvoiceAmount > 0 && p.InvoiceQty > 0 {
		unitRate = math.Round((p.InvoiceAmount/p.InvoiceQty)*100) / 100
	}

	if p.Attrs == nil {
		p.Attrs = make(map[string]any)
	}
	attrsJSON, err := json.Marshal(p.Attrs)
	if err != nil {
		return GateInwardResult{}, fmt.Errorf("failed to serialize attrs: %w", err)
	}

	// Backward compatibility field extraction for Fabric
	fabricType := ""
	if ft, ok := p.Attrs["fabricType"].(string); ok {
		fabricType = ft
	}
	vendorFabName := p.VendorProduct
	if vfn, ok := p.Attrs["vendorFabName"].(string); ok && vfn != "" {
		vendorFabName = vfn
	}
	fpccFabName := p.FpccProduct
	if ffn, ok := p.Attrs["fpccFabName"].(string); ok && ffn != "" {
		fpccFabName = ffn
	}
	var invoiceGsm float64
	if gsmVal, ok := p.Attrs["invoiceGsm"]; ok {
		if gsmF, ok := gsmVal.(float64); ok {
			invoiceGsm = gsmF
		}
	}
	var dia float64
	if diaVal, ok := p.Attrs["dia"]; ok {
		if diaF, ok := diaVal.(float64); ok {
			dia = diaF
		}
	}

	now := graph.NowMs()
	parentLotID := fmt.Sprintf("plt_%s", parentCode)

	tx, err := s.db.Begin()
	if err != nil {
		return GateInwardResult{}, err
	}
	defer tx.Rollback()

	// 1. Resolve or Auto-Register Vendor atomically
	var vendorID string
	err = tx.QueryRow("SELECT id FROM master_entities WHERE name = ? AND (entity_type = 'vendor' OR entity_type = 'both')", vendorName).Scan(&vendorID)
	if err == sql.ErrNoRows {
		vendorID = fmt.Sprintf("ent_vendor_%s", randomSuffix())
		_, err = tx.Exec(`
			INSERT INTO master_entities (id, name, entity_type, contact_person, address, created_at)
			VALUES (?, ?, 'vendor', 'Primary Vendor Contact', 'Direct Inward Supplier', ?)
			ON CONFLICT(id) DO NOTHING
		`, vendorID, vendorName, now)
		if err != nil {
			return GateInwardResult{}, fmt.Errorf("failed to register vendor entity: %w", err)
		}
	}

	// 2. Insert Parent Lot with stage 'AWAITING_COUNT'
	_, err = tx.Exec(`
		INSERT INTO parent_inward_lots (
			id, parent_code, inward_date, vendor_id, vendor_name, invoice_no, inward_type,
			uom, invoice_qty, invoice_amount, physical_received_qty, shortfall_qty, discrepancy_amount,
			fabric_type, vendor_fab_name, fpcc_fab_name, vendor_color, fpcc_color,
			invoice_gsm, dia, remarks, stage, unit_rate, counted_qty, variance_qty, variance_pct,
			vendor_product, fpcc_product, no_of_bundles, attrs_json, created_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, 0.0, 0.0, 0.0,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, 0.0, 0.0, 0.0,
			?, ?, ?, ?, ?
		)
	`, parentLotID, parentCode, inwardDate, vendorID, vendorName, invoiceNo, inwardType,
		uom, p.InvoiceQty, p.InvoiceAmount,
		fabricType, vendorFabName, fpccFabName, p.VendorColor, p.FpccColor,
		invoiceGsm, dia, p.Comment, string(StageAwaitingCount), unitRate,
		p.VendorProduct, p.FpccProduct, numBundles, string(attrsJSON), now,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return GateInwardResult{}, fmt.Errorf("parent code %s already exists in system", parentCode)
		}
		return GateInwardResult{}, fmt.Errorf("failed to insert parent lot: %w", err)
	}

	// 3. Insert Graph Node: PARENT_LOT
	itemName := p.FpccProduct
	if itemName == "" {
		itemName = p.VendorProduct
	}
	if itemName == "" {
		itemName = inwardType
	}
	colorName := p.FpccColor
	if colorName == "" {
		colorName = p.VendorColor
	}
	if colorName == "" {
		colorName = "Standard"
	}

	parentProps := map[string]any{
		"parent_code":   parentCode,
		"invoice_no":    invoiceNo,
		"inward_type":   inwardType,
		"item_name":     itemName,
		"color_name":    colorName,
		"invoice_qty":   p.InvoiceQty,
		"no_of_bundles": numBundles,
		"stage":         string(StageAwaitingCount),
	}
	_, err = s.graph.InsertNode(tx, graph.TypeParentLot, parentProps, parentLotID)
	if err != nil {
		return GateInwardResult{}, fmt.Errorf("failed to insert graph node for parent lot: %w", err)
	}

	// 4. Insert Graph Edge: (PARENT_LOT) -[:SUPPLIED_BY]-> (VENDOR)
	_, err = s.graph.InsertEdge(tx, parentLotID, vendorID, graph.EdgeSuppliedBy, map[string]any{
		"invoice_no":  invoiceNo,
		"invoice_qty": p.InvoiceQty,
	})
	if err != nil {
		return GateInwardResult{}, fmt.Errorf("failed to insert graph edge SUPPLIED_BY: %w", err)
	}

	// Map bundle stubs by explicit Sequence
	stubSpecs := make(map[int]GateBundleStub)
	for _, stub := range p.Bundles {
		if stub.Sequence > 0 {
			stubSpecs[stub.Sequence] = stub
		}
	}

	// 5. Generate Bundle Stubs (ParentCode-1/N .. N/N) with Pending status
	var createdStubs []BundleStubInfo
	for i := 1; i <= numBundles; i++ {
		bundleCode := fmt.Sprintf("%s-%d/%d", parentCode, i, numBundles)
		bundleID := fmt.Sprintf("bnd_%s_%d", parentCode, i)

		sizeSpec := ""
		rate := unitRate
		if stub, ok := stubSpecs[i]; ok {
			if stub.SizeSpec != "" {
				sizeSpec = stub.SizeSpec
			}
			if stub.Rate > 0 {
				rate = stub.Rate
			}
		} else if i <= len(p.Bundles) {
			if p.Bundles[i-1].SizeSpec != "" {
				sizeSpec = p.Bundles[i-1].SizeSpec
			}
			if p.Bundles[i-1].Rate > 0 {
				rate = p.Bundles[i-1].Rate
			}
		}

		qrPayloadJSON, _ := json.Marshal(map[string]any{
			"bundleCode":  bundleCode,
			"parentCode":  parentCode,
			"item":        itemName,
			"color":       colorName,
			"sizeSpec":    sizeSpec,
			"location":    "A/R01/S01",
			"qcStatus":    "Pending",
			"generatedAt": time.Now().Format(time.RFC3339),
		})

		_, err = tx.Exec(`
			INSERT INTO bundles (
				id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
				item_type, item_name, color_name, inward_weight_qty, measured_gsm,
				location_bin, qc_status, qc_remarks, qr_payload, available_qty,
				size_spec, rate, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0.0, ?, 'A/R01/S01', 'Pending', '', ?, 0.0, ?, ?, ?)
		`, bundleID, bundleCode, parentLotID, parentCode, i, numBundles,
			inwardType, itemName, colorName, invoiceGsm,
			string(qrPayloadJSON), sizeSpec, rate, now,
		)
		if err != nil {
			return GateInwardResult{}, fmt.Errorf("failed to insert bundle stub %s: %w", bundleCode, err)
		}

		// Graph Node: BUNDLE
		bundleProps := map[string]any{
			"bundle_code": bundleCode,
			"sequence":    i,
			"total":       numBundles,
			"qc_status":   "Pending",
		}
		if _, err := s.graph.InsertNode(tx, graph.TypeBundle, bundleProps, bundleID); err != nil {
			return GateInwardResult{}, fmt.Errorf("failed to insert bundle graph node: %w", err)
		}

		// Graph Edge: (BUNDLE) -[:PART_OF_LOT]-> (PARENT_LOT)
		if _, err := s.graph.InsertEdge(tx, bundleID, parentLotID, graph.EdgePartOfLot, map[string]any{"sequence": i}); err != nil {
			return GateInwardResult{}, fmt.Errorf("failed to insert edge PART_OF_LOT: %w", err)
		}

		createdStubs = append(createdStubs, BundleStubInfo{
			BundleCode: bundleCode,
			Sequence:   i,
			Total:      numBundles,
			SizeSpec:   sizeSpec,
			Rate:       rate,
			Status:     "Pending",
		})
	}

	// 6. Workflow Event: INWARD_LOT_BOOKED
	if _, err := s.workflow.EmitEvent(tx, parentLotID, "INWARD_LOT_BOOKED", map[string]any{
		"parent_code":   parentCode,
		"vendor_name":   vendorName,
		"inward_type":   inwardType,
		"invoice_qty":   p.InvoiceQty,
		"no_of_bundles": numBundles,
	}, 0); err != nil {
		return GateInwardResult{}, fmt.Errorf("failed to emit INWARD_LOT_BOOKED event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return GateInwardResult{}, err
	}

	go func() {
		_, _ = s.workflow.ProcessPendingEvents()
	}()

	return GateInwardResult{
		Success:    true,
		ParentCode: parentCode,
		Stage:      StageAwaitingCount,
		Bundles:    createdStubs,
	}, nil
}

// -------------------------------------------------------------
// STATION 1B: FLOOR COUNT & DISCREPANCY RECONCILIATION
// -------------------------------------------------------------

func (s *Service) RecordFloorCount(p FloorCountPayload) (FloorCountResult, error) {
	parentCode := strings.TrimSpace(p.ParentCode)
	if parentCode == "" {
		return FloorCountResult{}, fmt.Errorf("parentCode is required")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return FloorCountResult{}, err
	}
	defer tx.Rollback()

	// 1. Fetch current lot with stage check
	var lotID, vendorName, invoiceNo, inwardType, uom string
	var currentStage string
	var unitRate, billedInvoiceQty float64
	err = tx.QueryRow(`
		SELECT id, vendor_name, invoice_no, inward_type, uom, stage, unit_rate, invoice_qty
		FROM parent_inward_lots WHERE parent_code = ?
	`, parentCode).Scan(&lotID, &vendorName, &invoiceNo, &inwardType, &uom, &currentStage, &unitRate, &billedInvoiceQty)
	if err == sql.ErrNoRows {
		return FloorCountResult{}, fmt.Errorf("lot %s not found", parentCode)
	} else if err != nil {
		return FloorCountResult{}, err
	}

	// CAS guard: Only lots in AWAITING_COUNT or DISCREPANCY_HOLD can be floor counted
	if currentStage != string(StageAwaitingCount) && currentStage != string(StageDiscrepancyHold) {
		return FloorCountResult{}, fmt.Errorf("lot %s is in stage %s; floor count can only be recorded for lots awaiting count", parentCode, currentStage)
	}

	invoicedQty := p.InvoicedQty
	if invoicedQty <= 0 {
		invoicedQty = billedInvoiceQty
	}

	// 2. Compute Physical Counted Total
	countedTotal := p.CountedQty
	var bundleSum float64
	if len(p.Bundles) > 0 {
		for _, b := range p.Bundles {
			bundleSum += b.CountedQty
		}
		if countedTotal <= 0 || math.Abs(countedTotal-bundleSum) > 0.05 {
			countedTotal = bundleSum
		}
	}
	countedTotal = math.Round(countedTotal*1000) / 1000

	// 3. Discrepancy Engine
	varianceQty := math.Round((countedTotal-invoicedQty)*1000) / 1000
	variancePct := 0.0
	if invoicedQty > 0 {
		variancePct = math.Round((varianceQty/invoicedQty)*10000) / 100
	}
	financialImpact := math.Round(math.Abs(varianceQty)*unitRate*100) / 100

	status := "MATCHED"
	if varianceQty < -0.05 {
		status = "SHORTAGE_FLAGGED"
	} else if varianceQty > 0.05 {
		status = "EXCESS_FLAGGED"
	}

	nextStage := StageAwaitingQC
	if status == "SHORTAGE_FLAGGED" && p.HoldOnShort {
		nextStage = StageDiscrepancyHold
	}

	now := graph.NowMs()
	recID := fmt.Sprintf("fcr_%s_%s", parentCode, randomSuffix())
	sizeLinesJSON, _ := json.Marshal(p.SizeLines)

	locationBin := strings.TrimSpace(p.LocationBin)
	if locationBin == "" {
		locationBin = "A/R01/S01"
	}

	// 4. Insert Reconciliation Audit Row
	_, err = tx.Exec(`
		INSERT INTO floor_count_reconciliations (
			id, lot_id, parent_code, invoiced_qty, counted_qty, variance_qty, variance_pct,
			financial_impact, status, location_bin, size_lines_json, counted_by, comments, recorded_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, recID, lotID, parentCode, invoicedQty, countedTotal, varianceQty, variancePct,
		financialImpact, status, locationBin, string(sizeLinesJSON), p.CountedBy, p.Comments, now,
	)
	if err != nil {
		return FloorCountResult{}, fmt.Errorf("failed to insert floor count reconciliation: %w", err)
	}

	// 5. Atomic CAS Update on Parent Inward Lot
	res, err := tx.Exec(`
		UPDATE parent_inward_lots
		SET stage = ?, invoice_qty = ?, counted_qty = ?, variance_qty = ?, variance_pct = ?,
		    physical_received_qty = ?, shortfall_qty = ?, discrepancy_amount = ?
		WHERE parent_code = ? AND stage IN (?, ?)
	`, string(nextStage), invoicedQty, countedTotal, varianceQty, variancePct,
		countedTotal, (invoicedQty - countedTotal), financialImpact, parentCode,
		string(StageAwaitingCount), string(StageDiscrepancyHold),
	)
	if err != nil {
		return FloorCountResult{}, fmt.Errorf("failed to update parent lot projection: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return FloorCountResult{}, fmt.Errorf("lot %s stage update conflict; lot may have been modified by another station", parentCode)
	}

	// 6. Update Bundles with Physical Measurements & Locations
	bundlesUpdated := 0
	for _, b := range p.Bundles {
		loc := strings.TrimSpace(b.LocationBin)
		if loc == "" {
			loc = locationBin
		}

		// Auto-register location in master_items
		_, _ = tx.Exec(`
			INSERT INTO master_items (id, category, code, display_name, meta_json, created_at)
			VALUES (?, 'location', ?, ?, '{}', ?)
			ON CONFLICT(category, code) DO NOTHING
		`, fmt.Sprintf("loc_%s", loc), strings.ToUpper(loc), strings.ToUpper(loc), now)

		// Fetch existing bundle row to ensure accurate bundle_id
		var bID string
		var bSeq int
		err = tx.QueryRow("SELECT id, bundle_sequence FROM bundles WHERE bundle_code = ? AND parent_code = ?", b.BundleCode, parentCode).Scan(&bID, &bSeq)
		if err != nil {
			return FloorCountResult{}, fmt.Errorf("bundle %s not found in lot %s: %w", b.BundleCode, parentCode, err)
		}

		_, err = tx.Exec(`
			UPDATE bundles
			SET inward_weight_qty = ?, available_qty = ?, location_bin = ?,
			    measured_gsm = ?, measured_dia = ?, size_spec = COALESCE(NULLIF(?, ''), size_spec),
			    rate = CASE WHEN ? > 0 THEN ? ELSE rate END, bundle_comment = ?
			WHERE id = ?
		`, b.CountedQty, b.CountedQty, loc,
			b.MeasuredGsm, b.MeasuredDia, b.SizeSpec,
			b.Rate, b.Rate, b.Comment, bID,
		)
		if err != nil {
			return FloorCountResult{}, fmt.Errorf("failed to update bundle %s: %w", b.BundleCode, err)
		}
		bundlesUpdated++

		// Bi-Temporal Graph Edge: Close previous STORED_AT and open new STORED_AT
		locNodeID := fmt.Sprintf("loc_%s", strings.ToUpper(loc))

		// Ensure LOCATION node exists
		if _, err := s.graph.InsertNode(tx, graph.TypeLocation, map[string]any{"code": loc}, locNodeID); err != nil {
			return FloorCountResult{}, fmt.Errorf("failed to insert location graph node: %w", err)
		}

		// Close previous STORED_AT edges for this bundle
		_, err = tx.Exec(`
			UPDATE graph_edges SET txn_to = ?, valid_to = ?
			WHERE src_id = ? AND edge_type = ? AND txn_to IS NULL
		`, now, now, bID, graph.EdgeStoredAt)
		if err != nil {
			return FloorCountResult{}, fmt.Errorf("failed to close previous STORED_AT edge: %w", err)
		}

		// Create fresh STORED_AT edge
		if _, err := s.graph.InsertEdge(tx, bID, locNodeID, graph.EdgeStoredAt, map[string]any{
			"counted_qty": b.CountedQty,
			"location":    loc,
		}); err != nil {
			return FloorCountResult{}, fmt.Errorf("failed to insert STORED_AT edge: %w", err)
		}
	}

	// 7. If Shortage or Excess, Record in discrepancy_reports table with correct invoiceNo
	if status != "MATCHED" {
		discReportID := fmt.Sprintf("disc_%s", parentCode)
		discReportNo := fmt.Sprintf("DISC-%s", parentCode)
		discRemarks := fmt.Sprintf("Invoiced %.3f %s vs Counted %.3f %s (Variance: %.3f, Status: %s)",
			invoicedQty, uom, countedTotal, uom, varianceQty, status)

		_, err = tx.Exec(`
			INSERT INTO discrepancy_reports (
				id, report_no, parent_code, vendor_name, invoice_no, inward_type,
				invoice_qty, actual_qty, variance_qty, variance_pct, financial_impact,
				status, remarks, created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				invoice_qty = excluded.invoice_qty,
				actual_qty = excluded.actual_qty,
				variance_qty = excluded.variance_qty,
				variance_pct = excluded.variance_pct,
				financial_impact = excluded.financial_impact,
				status = excluded.status,
				remarks = excluded.remarks
		`, discReportID, discReportNo, parentCode, vendorName, invoiceNo, inwardType,
			invoicedQty, countedTotal, varianceQty, variancePct, financialImpact,
			status, discRemarks, now,
		)
		if err != nil {
			return FloorCountResult{}, fmt.Errorf("failed to log discrepancy report: %w", err)
		}
	}

	// 8. Emit Workflow Event: INWARD_FLOOR_COUNTED with shortfall_qty for trigger rule
	shortfallQty := math.Max(0, invoicedQty-countedTotal)
	if _, err := s.workflow.EmitEvent(tx, lotID, "INWARD_FLOOR_COUNTED", map[string]any{
		"parent_code":      parentCode,
		"vendor_name":      vendorName,
		"invoiced_qty":     invoicedQty,
		"counted_qty":      countedTotal,
		"shortfall_qty":    shortfallQty,
		"variance_qty":     varianceQty,
		"variance_pct":     variancePct,
		"status":           status,
		"financial_impact": financialImpact,
	}, 0); err != nil {
		return FloorCountResult{}, fmt.Errorf("failed to emit INWARD_FLOOR_COUNTED event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return FloorCountResult{}, err
	}

	go func() {
		_, _ = s.workflow.ProcessPendingEvents()
	}()

	return FloorCountResult{
		Success:         true,
		ParentCode:      parentCode,
		Stage:           nextStage,
		InvoicedQty:     invoicedQty,
		CountedQty:      countedTotal,
		VarianceQty:     varianceQty,
		VariancePct:     variancePct,
		FinancialImpact: financialImpact,
		Status:          status,
		BundlesUpdated:  bundlesUpdated,
	}, nil
}

// -------------------------------------------------------------
// STATION 1C: QC INSPECTION & ZEBRA QR TAGGING
// -------------------------------------------------------------

func (s *Service) RecordBundleQC(p BundleQCPayload) (BundleQCResult, error) {
	parentCode := strings.TrimSpace(p.ParentCode)
	if parentCode == "" {
		return BundleQCResult{}, fmt.Errorf("parentCode is required")
	}

	if len(p.Inspections) == 0 {
		return BundleQCResult{}, fmt.Errorf("at least one bundle inspection is required")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return BundleQCResult{}, err
	}
	defer tx.Rollback()

	now := graph.NowMs()
	passCount := 0
	failCount := 0
	processedCount := 0

	for _, item := range p.Inspections {
		cleanBundleCode := strings.TrimSpace(item.BundleCode)
		if cleanBundleCode == "" {
			continue
		}

		// Normalize QC Result
		resStatus := "Passed"
		upperResult := strings.ToUpper(strings.TrimSpace(item.Result))
		qcResultRecord := "PASS"
		if upperResult == "FAIL" || upperResult == "REJECTED" || upperResult == "NO" {
			resStatus = "Rejected"
			qcResultRecord = "FAIL"
			failCount++
		} else {
			resStatus = "Passed"
			qcResultRecord = "PASS"
			passCount++
		}

		inspector := strings.TrimSpace(item.InspectedBy)
		if inspector == "" {
			inspector = strings.TrimSpace(p.Inspector)
		}
		if inspector == "" {
			inspector = "QC Inspector"
		}

		// Fenced query: bundle MUST belong to this parentCode
		var bundleID, itemName, colorName, locBin string
		var inwardQty float64
		err = tx.QueryRow(`
			SELECT id, item_name, color_name, location_bin, inward_weight_qty
			FROM bundles WHERE bundle_code = ? AND parent_code = ?
		`, cleanBundleCode, parentCode).Scan(&bundleID, &itemName, &colorName, &locBin, &inwardQty)
		if err != nil {
			return BundleQCResult{}, fmt.Errorf("bundle %s does not belong to lot %s or does not exist: %w", cleanBundleCode, parentCode, err)
		}

		// 1. Append Inspection Audit Row with random entropy
		qcInspID := fmt.Sprintf("qc_%s_%s", cleanBundleCode, randomSuffix())
		_, err = tx.Exec(`
			INSERT INTO bundle_qc_inspections (
				id, bundle_id, parent_code, bundle_code, result, remarks, image_url, inspected_by, inspected_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, qcInspID, bundleID, parentCode, cleanBundleCode, qcResultRecord, item.Remarks, item.ImageURL, inspector, now)
		if err != nil {
			return BundleQCResult{}, fmt.Errorf("failed to record inspection for %s: %w", cleanBundleCode, err)
		}

		// 2. Regenerate 2D QR Code Payload
		qrPayloadJSON, _ := json.Marshal(map[string]any{
			"bundleCode":  cleanBundleCode,
			"parentCode":  parentCode,
			"item":        itemName,
			"color":       colorName,
			"qty":         inwardQty,
			"location":    locBin,
			"qcStatus":    resStatus,
			"inspector":   inspector,
			"inspectedAt": time.Now().Format(time.RFC3339),
		})

		// 3. Update Bundle Status & QC Metadata
		_, err = tx.Exec(`
			UPDATE bundles
			SET qc_status = ?, qc_remarks = ?, qc_image_url = ?, qc_inspected_at = ?, qr_payload = ?
			WHERE id = ?
		`, resStatus, item.Remarks, item.ImageURL, now, string(qrPayloadJSON), bundleID)
		if err != nil {
			return BundleQCResult{}, fmt.Errorf("failed to update bundle %s: %w", cleanBundleCode, err)
		}

		// 4. Unified Graph: Create QC_INSPECTION node and INSPECTED_BY edge
		qcNodeID := fmt.Sprintf("qc_node_%s", cleanBundleCode)
		if _, err := s.graph.InsertNode(tx, graph.TypeQcInspection, map[string]any{
			"bundle_code": cleanBundleCode,
			"result":      resStatus,
			"inspector":   inspector,
			"image_url":   item.ImageURL,
			"remarks":     item.Remarks,
		}, qcNodeID); err != nil {
			return BundleQCResult{}, fmt.Errorf("failed to insert graph node for QC inspection: %w", err)
		}

		if _, err := s.graph.InsertEdge(tx, bundleID, qcNodeID, graph.EdgeInspectedBy, map[string]any{
			"result":       resStatus,
			"inspected_at": now,
		}); err != nil {
			return BundleQCResult{}, fmt.Errorf("failed to insert graph edge INSPECTED_BY: %w", err)
		}

		processedCount++
	}

	// 5. Atomic check: Are all bundles in this lot inspected?
	var pendingCount, totalBundles int
	err = tx.QueryRow(`
		SELECT COUNT(*), (SELECT COUNT(*) FROM bundles WHERE parent_code = ?)
		FROM bundles WHERE parent_code = ? AND qc_status = 'Pending'
	`, parentCode, parentCode).Scan(&pendingCount, &totalBundles)
	if err != nil {
		return BundleQCResult{}, err
	}

	lotCompleted := (pendingCount == 0 && totalBundles > 0)
	currentLotStage := StageAwaitingQC
	if lotCompleted {
		currentLotStage = StageQCComplete
		// Advance lot stage atomically using CAS
		_, err = tx.Exec(`
			UPDATE parent_inward_lots
			SET stage = ?
			WHERE parent_code = ? AND stage IN (?, ?)
		`, string(StageQCComplete), parentCode, string(StageAwaitingQC), string(StageDiscrepancyHold))
		if err != nil {
			return BundleQCResult{}, fmt.Errorf("failed to advance lot stage to QC_COMPLETE: %w", err)
		}

		// Emit workflow event
		lotID := fmt.Sprintf("plt_%s", parentCode)
		if _, err := s.workflow.EmitEvent(tx, lotID, "INWARD_QC_COMPLETED", map[string]any{
			"parent_code":   parentCode,
			"total_bundles": totalBundles,
			"pass_count":    passCount,
			"fail_count":    failCount,
		}, 0); err != nil {
			return BundleQCResult{}, fmt.Errorf("failed to emit INWARD_QC_COMPLETED event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return BundleQCResult{}, err
	}

	go func() {
		_, _ = s.workflow.ProcessPendingEvents()
	}()

	return BundleQCResult{
		Success:        true,
		ParentCode:     parentCode,
		Stage:          currentLotStage,
		InspectedCount: processedCount,
		TotalBundles:   totalBundles,
		LotCompleted:   lotCompleted,
		PassCount:      passCount,
		FailCount:      failCount,
	}, nil
}

// -------------------------------------------------------------
// DISCREPANCY RESOLUTION WORKFLOW
// -------------------------------------------------------------

func (s *Service) ResolveDiscrepancy(parentCode, action, remarks string) (map[string]any, error) {
	cleanCode := strings.TrimSpace(parentCode)
	if cleanCode == "" {
		return nil, fmt.Errorf("parentCode is required")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var currentStage string
	err = tx.QueryRow("SELECT stage FROM parent_inward_lots WHERE parent_code = ?", cleanCode).Scan(&currentStage)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lot %s not found", cleanCode)
	} else if err != nil {
		return nil, err
	}

	if currentStage != string(StageDiscrepancyHold) {
		return nil, fmt.Errorf("lot %s is in stage %s; only lots on DISCREPANCY_HOLD can be resolved", cleanCode, currentStage)
	}

	targetStage := StageAwaitingQC
	switch strings.ToUpper(action) {
	case "APPROVE_SHORTAGE", "ACCEPT_VARIANCE", "PROCEED":
		targetStage = StageAwaitingQC
	case "FORCE_RECOUNT", "RECOUNT":
		targetStage = StageAwaitingCount
	default:
		return nil, fmt.Errorf("unknown resolution action '%s' (allowed: APPROVE_SHORTAGE, FORCE_RECOUNT)", action)
	}

	now := graph.NowMs()
	res, err := tx.Exec(`
		UPDATE parent_inward_lots
		SET stage = ?, remarks = CASE WHEN ? != '' THEN remarks || ' | Resolution: ' || ? ELSE remarks END
		WHERE parent_code = ? AND stage = ?
	`, string(targetStage), remarks, remarks, cleanCode, string(StageDiscrepancyHold))
	if err != nil {
		return nil, err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return nil, fmt.Errorf("lot %s stage conflict during resolution", cleanCode)
	}

	// Update discrepancy reports
	_, _ = tx.Exec(`
		UPDATE discrepancy_reports
		SET status = 'RESOLVED', remarks = remarks || ' [Resolved: ' || ? || ']'
		WHERE parent_code = ?
	`, remarks, cleanCode)

	// Emit workflow event
	lotID := fmt.Sprintf("plt_%s", cleanCode)
	if _, err := s.workflow.EmitEvent(tx, lotID, "DISCREPANCY_RESOLVED", map[string]any{
		"parent_code":  cleanCode,
		"action":       action,
		"target_stage": string(targetStage),
		"resolved_at":  now,
	}, 0); err != nil {
		return nil, fmt.Errorf("failed to emit DISCREPANCY_RESOLVED event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	go func() {
		_, _ = s.workflow.ProcessPendingEvents()
	}()

	return map[string]any{
		"success":    true,
		"parentCode": cleanCode,
		"newStage":   string(targetStage),
		"action":     action,
	}, nil
}

// -------------------------------------------------------------
// LISTING & DETAIL QUERIES
// -------------------------------------------------------------

func (s *Service) ListLots(stageFilter string) ([]LotSummary, error) {
	query := `
		SELECT 
			p.id, p.parent_code, p.inward_date, p.inward_type, p.vendor_name, p.invoice_no,
			p.invoice_qty, p.invoice_amount, p.counted_qty, p.variance_qty, p.variance_pct,
			p.uom, COALESCE(p.vendor_product, p.vendor_fab_name, ''),
			COALESCE(p.fpcc_product, p.fpcc_fab_name, ''),
			COALESCE(p.vendor_color, ''), COALESCE(p.fpcc_color, ''),
			COALESCE(p.stage, 'AWAITING_COUNT'), COALESCE(p.no_of_bundles, 1),
			(SELECT COUNT(*) FROM bundles b WHERE b.parent_code = p.parent_code AND b.qc_status != 'Pending') AS inspected_count,
			(SELECT COUNT(*) FROM bundles b WHERE b.parent_code = p.parent_code AND b.qc_status = 'Passed') AS passed_count,
			(SELECT COUNT(*) FROM bundles b WHERE b.parent_code = p.parent_code AND b.qc_status = 'Rejected') AS rejected_count,
			COALESCE(p.remarks, ''), COALESCE(p.attrs_json, '{}'), p.created_at
		FROM parent_inward_lots p
	`
	var args []any
	cleanFilter := strings.TrimSpace(stageFilter)
	if cleanFilter != "" && cleanFilter != "ALL" {
		query += " WHERE p.stage = ?"
		args = append(args, cleanFilter)
	}
	query += " ORDER BY p.created_at DESC LIMIT 100"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []LotSummary
	for rows.Next() {
		var item LotSummary
		var rawAttrs string
		err := rows.Scan(
			&item.ID, &item.ParentCode, &item.InwardDate, &item.InwardType, &item.VendorName, &item.InvoiceNo,
			&item.InvoiceQty, &item.InvoiceAmount, &item.CountedQty, &item.VarianceQty, &item.VariancePct,
			&item.UOM, &item.VendorProduct, &item.FpccProduct,
			&item.VendorColor, &item.FpccColor,
			&item.Stage, &item.NoOfBundles,
			&item.InspectedBundles, &item.PassedBundles, &item.RejectedBundles,
			&item.Remarks, &rawAttrs, &item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		item.Attrs = make(map[string]any)
		if rawAttrs != "" {
			_ = json.Unmarshal([]byte(rawAttrs), &item.Attrs)
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating lots: %w", err)
	}

	return results, nil
}

func (s *Service) GetLot(parentCode string) (LotDetail, error) {
	cleanCode := strings.TrimSpace(parentCode)
	if cleanCode == "" {
		return LotDetail{}, fmt.Errorf("parentCode is required")
	}

	// 1. Fetch Lot Summary
	var item LotSummary
	var rawAttrs string
	err := s.db.QueryRow(`
		SELECT 
			p.id, p.parent_code, p.inward_date, p.inward_type, p.vendor_name, p.invoice_no,
			p.invoice_qty, p.invoice_amount, p.counted_qty, p.variance_qty, p.variance_pct,
			p.uom, COALESCE(p.vendor_product, p.vendor_fab_name, ''),
			COALESCE(p.fpcc_product, p.fpcc_fab_name, ''),
			COALESCE(p.vendor_color, ''), COALESCE(p.fpcc_color, ''),
			COALESCE(p.stage, 'AWAITING_COUNT'), COALESCE(p.no_of_bundles, 1),
			(SELECT COUNT(*) FROM bundles b WHERE b.parent_code = p.parent_code AND b.qc_status != 'Pending'),
			(SELECT COUNT(*) FROM bundles b WHERE b.parent_code = p.parent_code AND b.qc_status = 'Passed'),
			(SELECT COUNT(*) FROM bundles b WHERE b.parent_code = p.parent_code AND b.qc_status = 'Rejected'),
			COALESCE(p.remarks, ''), COALESCE(p.attrs_json, '{}'), p.created_at
		FROM parent_inward_lots p
		WHERE p.parent_code = ?
	`, cleanCode).Scan(
		&item.ID, &item.ParentCode, &item.InwardDate, &item.InwardType, &item.VendorName, &item.InvoiceNo,
		&item.InvoiceQty, &item.InvoiceAmount, &item.CountedQty, &item.VarianceQty, &item.VariancePct,
		&item.UOM, &item.VendorProduct, &item.FpccProduct,
		&item.VendorColor, &item.FpccColor,
		&item.Stage, &item.NoOfBundles,
		&item.InspectedBundles, &item.PassedBundles, &item.RejectedBundles,
		&item.Remarks, &rawAttrs, &item.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return LotDetail{}, fmt.Errorf("lot %s not found", cleanCode)
	} else if err != nil {
		return LotDetail{}, err
	}
	item.Attrs = make(map[string]any)
	if rawAttrs != "" {
		_ = json.Unmarshal([]byte(rawAttrs), &item.Attrs)
	}

	// 2. Fetch Bundles
	bundleRows, err := s.db.Query(`
		SELECT 
			id, bundle_code, parent_code, bundle_sequence, total_bundles,
			item_type, item_name, COALESCE(color_name, ''), inward_weight_qty,
			COALESCE(measured_gsm, 0), COALESCE(measured_dia, 0),
			COALESCE(size_spec, ''), COALESCE(rate, 0), location_bin,
			qc_status, COALESCE(qc_remarks, ''), COALESCE(qc_image_url, ''),
			COALESCE(qc_inspected_at, 0), COALESCE(bundle_comment, ''), COALESCE(qr_payload, '{}')
		FROM bundles
		WHERE parent_code = ?
		ORDER BY bundle_sequence ASC
	`, cleanCode)
	if err != nil {
		return LotDetail{}, err
	}
	defer bundleRows.Close()

	var bundles []BundleDetail
	for bundleRows.Next() {
		var b BundleDetail
		err := bundleRows.Scan(
			&b.ID, &b.BundleCode, &b.ParentCode, &b.BundleSequence, &b.TotalBundles,
			&b.ItemType, &b.ItemName, &b.ColorName, &b.InwardQty,
			&b.MeasuredGsm, &b.MeasuredDia,
			&b.SizeSpec, &b.Rate, &b.LocationBin,
			&b.QcStatus, &b.QcRemarks, &b.QcImageUrl,
			&b.QcInspectedAt, &b.BundleComment, &b.QRPayload,
		)
		if err != nil {
			return LotDetail{}, err
		}
		bundles = append(bundles, b)
	}
	if err := bundleRows.Err(); err != nil {
		return LotDetail{}, fmt.Errorf("error reading bundles: %w", err)
	}

	// 3. Fetch Floor Count Reconciliations
	recRows, err := s.db.Query(`
		SELECT id, parent_code, invoiced_qty, counted_qty, variance_qty, variance_pct,
		       financial_impact, status, location_bin, COALESCE(size_lines_json, '[]'),
		       COALESCE(counted_by, ''), COALESCE(comments, ''), recorded_at
		FROM floor_count_reconciliations
		WHERE parent_code = ?
		ORDER BY recorded_at DESC
	`, cleanCode)
	var reconciliations []map[string]any
	if err == nil {
		defer recRows.Close()
		for recRows.Next() {
			var id, pCode, status, loc, sizeJSON, countedBy, comm string
			var invQ, cntQ, varQ, varP, finI float64
			var recAt int64
			if err := recRows.Scan(&id, &pCode, &invQ, &cntQ, &varQ, &varP, &finI, &status, &loc, &sizeJSON, &countedBy, &comm, &recAt); err == nil {
				var sizeLines any
				_ = json.Unmarshal([]byte(sizeJSON), &sizeLines)
				reconciliations = append(reconciliations, map[string]any{
					"id":              id,
					"parentCode":      pCode,
					"invoicedQty":     invQ,
					"countedQty":      cntQ,
					"varianceQty":     varQ,
					"variancePct":     varP,
					"financialImpact": finI,
					"status":          status,
					"locationBin":     loc,
					"sizeLines":       sizeLines,
					"countedBy":       countedBy,
					"comments":        comm,
					"recordedAt":      recAt,
				})
			}
		}
	}

	// 4. Fetch QC Inspections
	inspRows, err := s.db.Query(`
		SELECT id, bundle_id, bundle_code, result, COALESCE(remarks, ''),
		       COALESCE(image_url, ''), COALESCE(inspected_by, ''), inspected_at
		FROM bundle_qc_inspections
		WHERE parent_code = ?
		ORDER BY inspected_at DESC
	`, cleanCode)
	var inspections []map[string]any
	if err == nil {
		defer inspRows.Close()
		for inspRows.Next() {
			var id, bId, bCode, res, rem, img, insp string
			var inspAt int64
			if err := inspRows.Scan(&id, &bId, &bCode, &res, &rem, &img, &insp, &inspAt); err == nil {
				inspections = append(inspections, map[string]any{
					"id":          id,
					"bundleId":    bId,
					"bundleCode":  bCode,
					"result":      res,
					"remarks":     rem,
					"imageUrl":    img,
					"inspectedBy": insp,
					"inspectedAt": inspAt,
				})
			}
		}
	}

	return LotDetail{
		Lot:             item,
		Bundles:         bundles,
		Reconciliations: reconciliations,
		Inspections:     inspections,
	}, nil
}

// -------------------------------------------------------------
// LEGACY COMPATIBILITY METHODS
// -------------------------------------------------------------

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

	var vendorID string
	err = tx.QueryRow("SELECT id FROM master_entities WHERE name = ? AND (entity_type = 'vendor' OR entity_type = 'both')", vendorName).Scan(&vendorID)
	if err == sql.ErrNoRows {
		vendorID = fmt.Sprintf("ent_vendor_%s", randomSuffix())
		_, err = tx.Exec(`
			INSERT INTO master_entities (id, name, entity_type, contact_person, address, created_at)
			VALUES (?, ?, 'vendor', 'Primary Vendor Contact', 'Direct Inward Supplier', ?)
		`, vendorID, vendorName, now)
		if err != nil {
			return nil, fmt.Errorf("failed to register vendor entity: %w", err)
		}
	}

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
			invoice_gsm, dia, remarks, stage, unit_rate, counted_qty, variance_qty, variance_pct,
			vendor_product, fpcc_product, no_of_bundles, attrs_json, created_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, 'QC_COMPLETE', ?, ?, ?, ?,
			?, ?, ?, '{}', ?
		)
	`, parentLotID, parentCode, inwardDate, vendorID, vendorName, p.InvoiceNo, inwardType,
		uom, p.InvoiceQty, p.InvoiceAmount, physicalTotal, shortfallQty, discrepancyAmt,
		p.FabricType, p.VendorFabName, p.FpccFabName, p.VendorColor, p.FpccColor,
		p.InvoiceGsm, p.Dia, p.Remarks, unitRate, physicalTotal, (physicalTotal - p.InvoiceQty),
		(physicalTotal-p.InvoiceQty)/p.InvoiceQty*100, p.VendorFabName, p.FpccFabName, bundleCount, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert parent lot: %w", err)
	}

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
	_, _ = s.graph.InsertNode(tx, graph.TypeParentLot, parentProps, parentLotID)
	_, _ = s.graph.InsertEdge(tx, parentLotID, vendorID, graph.EdgeSuppliedBy, map[string]any{
		"invoice_no": p.InvoiceNo,
		"inward_qty": physicalTotal,
	})

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

		_, _ = tx.Exec(`
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

		bundleProps := map[string]any{
			"bundle_code": bundleCode,
			"weight":      b.MeasuredWeight,
			"gsm":         b.MeasuredGsm,
			"location":    loc,
			"qc_status":   qc,
		}
		_, _ = s.graph.InsertNode(tx, graph.TypeBundle, bundleProps, bundleID)
		_, _ = s.graph.InsertEdge(tx, bundleID, parentLotID, graph.EdgePartOfLot, map[string]any{"seq": bundleSeq})

		createdBundles = append(createdBundles, map[string]any{
			"bundleCode": bundleCode,
			"weight":     b.MeasuredWeight,
			"location":   loc,
			"qcStatus":   qc,
		})
	}

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
			ON CONFLICT(id) DO UPDATE SET
				invoice_qty = excluded.invoice_qty,
				actual_qty = excluded.actual_qty,
				variance_qty = excluded.variance_qty,
				variance_pct = excluded.variance_pct,
				financial_impact = excluded.financial_impact,
				status = excluded.status,
				remarks = excluded.remarks
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
		"success":            true,
		"originalBundleCode": cleanCode,
		"updatedBundleCode":  updatedBundleCode,
		"returnType":         returnType,
		"returnedWeightKg":   returnedWeightKg,
		"location":           loc,
	}, nil
}

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
			"report_no":           repNo,
			"parent_code":         pCode,
			"vendor_name":         vName,
			"invoice_no":          invNo,
			"inward_type":         inwType,
			"invoice_qty":         invQty,
			"actual_received_qty": actQty,
			"discrepancy_qty":     diffQty,
			"financial_impact":    finImpact,
			"status":              status,
			"remarks":             rem,
		})
	}
	return list, rows.Err()
}
