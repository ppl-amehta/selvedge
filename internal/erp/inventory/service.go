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

type FabricProductItem struct {
	ID            string  `json:"id"`
	FabricType    string  `json:"fabricType"`
	FpccProduct   string  `json:"fpccProduct"`
	VendorProduct string  `json:"vendorProduct"`
	Color         string  `json:"color"`
	UOM           string  `json:"uom"`
	Dia           string  `json:"dia,omitempty"`
	GSM           float64 `json:"gsm,omitempty"`
	Composition   string  `json:"composition,omitempty"`
	Rate          float64 `json:"rate,omitempty"`
	RollCount     int     `json:"rollCount"`
}

type InwardInvoice struct {
	ID            string   `json:"id"`
	InvoiceNo     string   `json:"invoiceNo"`
	VendorName    string   `json:"vendorName"`
	InvoiceDate   string   `json:"invoiceDate"`
	InwardType    string   `json:"inwardType"`
	InvoiceQty    float64  `json:"invoiceQty"`
	InvoiceAmount float64  `json:"invoiceAmount"`
	UOM           string   `json:"uom"`
	Currency      string   `json:"currency"`
	Remarks       string   `json:"remarks"`
	CreatedAt     int64    `json:"createdAt"`
	TaggedLots    []string `json:"taggedLots,omitempty"`
	TotalLotsQty  float64  `json:"totalLotsQty,omitempty"`
}

type GateBundleStub struct {
	Sequence    int            `json:"sequence"`
	InvoicedQty float64        `json:"invoicedQty,omitempty"`
	SizeSpec    string         `json:"sizeSpec,omitempty"`
	Rate        float64        `json:"rate,omitempty"`
	Attrs       map[string]any `json:"attrs,omitempty"`
}

type GateInwardPayload struct {
	ParentCode     string              `json:"parentCode"`
	InwardDate     string              `json:"inwardDate"`
	InwardType     string              `json:"inwardType"` // Fabric, Zips, packaging, Trims, Label, Apparel, etc.
	VendorName     string              `json:"vendorName"`
	InvoiceNo      string              `json:"invoiceNo"`
	InvoiceQty     float64             `json:"invoiceQty"`
	InvoiceAmount  float64             `json:"invoiceAmount"`
	UOM            string              `json:"uom"`
	VendorProduct  string              `json:"vendorProduct"`
	FpccProduct    string              `json:"fpccProduct"`
	VendorColor    string              `json:"vendorColor"`
	FpccColor      string              `json:"fpccColor"`
	NoOfBundles    int                 `json:"noOfBundles"`
	Comment        string              `json:"comment"`
	Attrs          map[string]any      `json:"attrs"`   // Conditional specs
	Bundles        []GateBundleStub    `json:"bundles"` // Optional bundle-level size, qty & rate
	FabricProducts []FabricProductItem `json:"fabricProducts,omitempty"` // Multiple fabric products per lot
	QRCopies       int                 `json:"qrCopies,omitempty"`       // Default 3
}

type BundleStubInfo struct {
	BundleCode  string         `json:"bundleCode"`
	Sequence    int            `json:"sequence"`
	Total       int            `json:"total"`
	InvoicedQty float64        `json:"invoicedQty"`
	SizeSpec    string         `json:"sizeSpec,omitempty"`
	Rate        float64        `json:"rate,omitempty"`
	Attrs       map[string]any `json:"attrs,omitempty"`
	Status      string         `json:"status"`
}

type GateInwardResult struct {
	Success    bool             `json:"success"`
	ParentCode string           `json:"parentCode"`
	Stage      LotStage         `json:"stage"`
	Bundles    []BundleStubInfo `json:"bundles"`
}

type FloorCountBundle struct {
	BundleCode  string         `json:"bundleCode"`
	Sequence    int            `json:"sequence"`
	InvoicedQty float64        `json:"invoicedQty,omitempty"`
	CountedQty  float64        `json:"countedQty"`
	LocationBin string         `json:"locationBin"`
	MeasuredGsm float64        `json:"measuredGsm,omitempty"`
	MeasuredDia float64        `json:"measuredDia,omitempty"`
	SizeSpec    string         `json:"sizeSpec,omitempty"`
	Rate        float64        `json:"rate,omitempty"`
	Comment     string         `json:"comment,omitempty"`
	Attrs       map[string]any `json:"attrs,omitempty"`
}

type FloorCountPayload struct {
	ParentCode  string             `json:"parentCode"`
	InvoicedQty float64            `json:"invoicedQty"`
	CountedQty  float64            `json:"countedQty"`
	LocationBin string             `json:"locationBin"`
	CountedBy   string             `json:"countedBy"`
	Comments    string             `json:"comments"`
	HoldOnShort bool               `json:"holdOnShort"`
	Attrs       map[string]any     `json:"attrs,omitempty"`
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
	AudioURL    string `json:"audioUrl,omitempty"`
	Transcript  string `json:"transcript,omitempty"`
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
	ID                string              `json:"id"`
	ParentCode        string              `json:"parentCode"`
	InwardDate        string              `json:"inwardDate"`
	InwardType        string              `json:"inwardType"`
	VendorName        string              `json:"vendorName"`
	InvoiceID         string              `json:"invoiceId,omitempty"`
	InvoiceNo         string              `json:"invoiceNo"`
	InvoiceQty        float64             `json:"invoiceQty"`
	InvoiceAmount     float64             `json:"invoiceAmount"`
	CountedQty        float64             `json:"countedQty"`
	VarianceQty       float64             `json:"varianceQty"`
	VariancePct       float64             `json:"variancePct"`
	UOM               string              `json:"uom"`
	VendorProduct     string              `json:"vendorProduct"`
	FpccProduct       string              `json:"fpccProduct"`
	VendorColor       string              `json:"vendorColor"`
	FpccColor         string              `json:"fpccColor"`
	Stage             string              `json:"stage"`
	NoOfBundles       int                 `json:"noOfBundles"`
	InspectedBundles  int                 `json:"inspectedBundles"`
	PassedBundles     int                 `json:"passedBundles"`
	RejectedBundles   int                 `json:"rejectedBundles"`
	Remarks           string              `json:"remarks"`
	Attrs             map[string]any      `json:"attrs"`
	FabricProducts    []FabricProductItem `json:"fabricProducts,omitempty"`
	QRCopiesPerBundle int                 `json:"qrCopiesPerBundle,omitempty"`
	CreatedAt         int64               `json:"createdAt"`
}

type BundleDetail struct {
	ID             string         `json:"id"`
	BundleCode     string         `json:"bundleCode"`
	ParentCode     string         `json:"parentCode"`
	BundleSequence int            `json:"bundleSequence"`
	TotalBundles   int            `json:"totalBundles"`
	ItemType       string         `json:"itemType"`
	ItemName       string         `json:"itemName"`
	ColorName      string         `json:"colorName"`
	InvoicedQty    float64        `json:"invoicedQty"`
	InwardQty      float64        `json:"inwardQty"`
	MeasuredGsm    float64        `json:"measuredGsm"`
	MeasuredDia    float64        `json:"measuredDia"`
	SizeSpec       string         `json:"sizeSpec"`
	Rate           float64        `json:"rate"`
	Attrs          map[string]any `json:"attrs,omitempty"`
	LocationBin    string         `json:"locationBin"`
	QcStatus       string         `json:"qcStatus"`
	QcRemarks      string         `json:"qcRemarks"`
	QcImageUrl     string         `json:"qcImageUrl"`
	QcAudioUrl     string         `json:"qcAudioUrl"`
	QcTranscript   string         `json:"qcTranscript"`
	QcInspectedAt  int64          `json:"qcInspectedAt"`
	BundleComment  string         `json:"bundleComment"`
	QRPayload      string         `json:"qrPayload"`
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

	if p.InvoiceQty <= 0 && len(p.Bundles) > 0 {
		var bSum float64
		for _, b := range p.Bundles {
			bSum += b.InvoicedQty
		}
		if bSum > 0 {
			p.InvoiceQty = math.Round(bSum*1000) / 1000
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
	diaStr := ""
	if diaVal, ok := p.Attrs["dia"]; ok {
		switch v := diaVal.(type) {
		case float64:
			dia = v
			diaStr = fmt.Sprintf("%.0f\"", v)
		case string:
			diaStr = v
			fmt.Sscanf(strings.ReplaceAll(v, `"`, ""), "%f", &dia)
		}
	}

	qrCopies := p.QRCopies
	if qrCopies < 1 {
		qrCopies = 3
	} else if qrCopies > 10 {
		qrCopies = 10
	}
	if numBundles > 200 {
		numBundles = 200
	}

	// Fabric Multi-Product or Single Product Handling
	var fabricProductsJSON []byte
	var fabricProducts []FabricProductItem
	if inwardType == "Fabric" {
		if len(p.FabricProducts) > 0 {
			fabricProducts = p.FabricProducts
			totalRolls := 0
			for _, fp := range fabricProducts {
				rCount := fp.RollCount
				if rCount <= 0 {
					rCount = 1
				}
				totalRolls += rCount
			}
			if totalRolls > 0 {
				numBundles = totalRolls
			}
		} else {
			fabricProducts = []FabricProductItem{
				{
					ID:            "fp_1",
					FabricType:    fabricType,
					FpccProduct:   fpccFabName,
					VendorProduct: vendorFabName,
					Color:         p.FpccColor,
					UOM:           uom,
					Dia:           diaStr,
					GSM:           invoiceGsm,
					Rate:          unitRate,
					RollCount:     numBundles,
				},
			}
		}
		var err error
		fabricProductsJSON, err = json.Marshal(fabricProducts)
		if err != nil {
			return GateInwardResult{}, fmt.Errorf("failed to serialize fabric products: %w", err)
		}
	} else {
		if len(p.FabricProducts) > 1 {
			return GateInwardResult{}, fmt.Errorf("non-fabric inward types strictly support one product type")
		}
		fabricProductsJSON = []byte("[]")
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

	// 1b. Resolve or Auto-Register Top-Level Inward Invoice (Enforce ONLY ONE Inward Type per Invoice)
	var invoiceID, existingInvType string
	err = tx.QueryRow("SELECT id, COALESCE(inward_type, '') FROM inward_invoices WHERE vendor_name = ? AND invoice_no = ?", vendorName, invoiceNo).Scan(&invoiceID, &existingInvType)
	if err == sql.ErrNoRows {
		invoiceID = fmt.Sprintf("inv_%s", randomSuffix())
		_, err = tx.Exec(`
			INSERT INTO inward_invoices (id, invoice_no, vendor_name, invoice_date, inward_type, invoice_qty, invoice_amount, uom, currency, remarks, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'INR', ?, ?)
		`, invoiceID, invoiceNo, vendorName, inwardDate, inwardType, p.InvoiceQty, p.InvoiceAmount, uom, p.Comment, now)
		if err != nil {
			return GateInwardResult{}, fmt.Errorf("failed to register inward invoice: %w", err)
		}
	} else if err != nil {
		return GateInwardResult{}, fmt.Errorf("failed to check existing inward invoice: %w", err)
	} else {
		cleanExistingType := strings.TrimSpace(existingInvType)
		if cleanExistingType != "" && !strings.EqualFold(cleanExistingType, inwardType) {
			return GateInwardResult{}, fmt.Errorf("invoice %s is already assigned to Inward Type '%s'; only one inward type per invoice is allowed", invoiceNo, cleanExistingType)
		}
		if cleanExistingType == "" {
			_, _ = tx.Exec("UPDATE inward_invoices SET inward_type = ? WHERE id = ?", inwardType, invoiceID)
		}
	}

	// 2. Insert Parent Lot with stage 'AWAITING_COUNT'
	_, err = tx.Exec(`
		INSERT INTO parent_inward_lots (
			id, parent_code, inward_date, vendor_id, vendor_name, invoice_no, inward_type,
			uom, invoice_qty, invoice_amount, physical_received_qty, shortfall_qty, discrepancy_amount,
			fabric_type, vendor_fab_name, fpcc_fab_name, vendor_color, fpcc_color,
			invoice_gsm, dia, remarks, stage, unit_rate, counted_qty, variance_qty, variance_pct,
			vendor_product, fpcc_product, no_of_bundles, attrs_json, created_at,
			invoice_id, fabric_products_json, qr_copies_per_bundle
		) VALUES (
			?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, 0.0, 0.0, 0.0,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, 0.0, 0.0, 0.0,
			?, ?, ?, ?, ?,
			?, ?, ?
		)
	`, parentLotID, parentCode, inwardDate, vendorID, vendorName, invoiceNo, inwardType,
		uom, p.InvoiceQty, p.InvoiceAmount,
		fabricType, vendorFabName, fpccFabName, p.VendorColor, p.FpccColor,
		invoiceGsm, dia, p.Comment, string(StageAwaitingCount), unitRate,
		p.VendorProduct, p.FpccProduct, numBundles, string(attrsJSON), now,
		invoiceID, string(fabricProductsJSON), qrCopies,
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

	// 4b. Insert Graph Node for INVOICE & Edge: (INVOICE) -[:CONTAINS_LOT]-> (PARENT_LOT)
	_, _ = s.graph.EnsureNode(tx, "INVOICE", map[string]any{
		"invoice_no": invoiceNo,
		"vendor":     vendorName,
	}, invoiceID)
	_, _ = s.graph.InsertEdge(tx, invoiceID, parentLotID, "CONTAINS_LOT", map[string]any{
		"inward_type": inwardType,
	})

	// Map bundle stubs by explicit Sequence
	stubSpecs := make(map[int]GateBundleStub)
	for _, stub := range p.Bundles {
		if stub.Sequence > 0 {
			stubSpecs[stub.Sequence] = stub
		}
	}

	// 5. Generate Bundle Stubs (ParentCode-1/N .. N/N) with Pending status
	var createdStubs []BundleStubInfo
	bundleIndex := 0
	isFabric := strings.EqualFold(inwardType, "Fabric")
	defaultBundleInvoiced := math.Round((p.InvoiceQty/float64(numBundles))*1000) / 1000

	marshalBundleAttrs := func(m map[string]any) string {
		if len(m) == 0 {
			return "{}"
		}
		b, err := json.Marshal(m)
		if err != nil {
			return "{}"
		}
		return string(b)
	}

	if isFabric && len(fabricProducts) > 0 {
		for _, fp := range fabricProducts {
			rolls := fp.RollCount
			if rolls <= 0 {
				rolls = 1
			}
			prodItemName := fp.FpccProduct
			if prodItemName == "" {
				prodItemName = fp.VendorProduct
			}
			if prodItemName == "" {
				prodItemName = fp.FabricType
			}
			prodColor := fp.Color
			if prodColor == "" {
				prodColor = colorName
			}
			prodRate := fp.Rate
			if prodRate <= 0 {
				prodRate = unitRate
			}
			prodGsm := fp.GSM
			if prodGsm <= 0 {
				prodGsm = invoiceGsm
			}

			for r := 0; r < rolls; r++ {
				bundleIndex++
				bundleCode := fmt.Sprintf("%s-%d/%d", parentCode, bundleIndex, numBundles)
				bundleID := fmt.Sprintf("bnd_%s_%d", parentCode, bundleIndex)

				sizeSpec := fp.FabricType
				if fp.Dia != "" {
					sizeSpec += fmt.Sprintf(" (Dia: %s)", fp.Dia)
				}
				rate := prodRate
				invoicedQty := defaultBundleInvoiced
				var bAttrs map[string]any
				if stub, ok := stubSpecs[bundleIndex]; ok {
					if stub.SizeSpec != "" {
						sizeSpec = stub.SizeSpec
					}
					if stub.Rate > 0 {
						rate = stub.Rate
					}
					if stub.InvoicedQty > 0 {
						invoicedQty = stub.InvoicedQty
					}
					if len(stub.Attrs) > 0 {
						bAttrs = stub.Attrs
					}
				} else if bundleIndex <= len(p.Bundles) {
					st := p.Bundles[bundleIndex-1]
					if st.InvoicedQty > 0 {
						invoicedQty = st.InvoicedQty
					}
					if len(st.Attrs) > 0 {
						bAttrs = st.Attrs
					}
				}

				qrPayloadJSON, _ := json.Marshal(map[string]any{
					"bundleCode":  bundleCode,
					"parentCode":  parentCode,
					"item":        prodItemName,
					"color":       prodColor,
					"sizeSpec":    sizeSpec,
					"invoicedQty": invoicedQty,
					"location":    "A/R01/S01",
					"qcStatus":    "Pending",
					"generatedAt": time.Now().Format(time.RFC3339),
				})

				_, err = tx.Exec(`
					INSERT INTO bundles (
						id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
						item_type, item_name, color_name, invoiced_qty, inward_weight_qty, measured_gsm,
						location_bin, qc_status, qc_remarks, qr_payload, available_qty,
						size_spec, rate, attrs_json, created_at
					) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0.0, ?, 'A/R01/S01', 'Pending', '', ?, 0.0, ?, ?, ?, ?)
				`, bundleID, bundleCode, parentLotID, parentCode, bundleIndex, numBundles,
					inwardType, prodItemName, prodColor, invoicedQty, prodGsm,
					string(qrPayloadJSON), sizeSpec, rate, marshalBundleAttrs(bAttrs), now,
				)
				if err != nil {
					return GateInwardResult{}, fmt.Errorf("failed to insert fabric bundle stub %s: %w", bundleCode, err)
				}

				// Graph Node: BUNDLE
				bundleProps := map[string]any{
					"bundle_code":  bundleCode,
					"sequence":     bundleIndex,
					"total":        numBundles,
					"qc_status":    "Pending",
					"item_name":    prodItemName,
					"invoiced_qty": invoicedQty,
				}
				if _, err := s.graph.InsertNode(tx, graph.TypeBundle, bundleProps, bundleID); err != nil {
					return GateInwardResult{}, fmt.Errorf("failed to insert bundle graph node: %w", err)
				}

				if _, err := s.graph.InsertEdge(tx, bundleID, parentLotID, graph.EdgePartOfLot, map[string]any{"sequence": bundleIndex}); err != nil {
					return GateInwardResult{}, fmt.Errorf("failed to insert edge PART_OF_LOT: %w", err)
				}

				createdStubs = append(createdStubs, BundleStubInfo{
					BundleCode:  bundleCode,
					Sequence:    bundleIndex,
					Total:       numBundles,
					InvoicedQty: invoicedQty,
					SizeSpec:    sizeSpec,
					Rate:        rate,
					Attrs:       bAttrs,
					Status:      "Pending",
				})
			}
		}
	} else {
		for i := 1; i <= numBundles; i++ {
			bundleCode := fmt.Sprintf("%s-%d/%d", parentCode, i, numBundles)
			bundleID := fmt.Sprintf("bnd_%s_%d", parentCode, i)

			sizeSpec := ""
			rate := unitRate
			invoicedQty := defaultBundleInvoiced
			var bAttrs map[string]any
			if stub, ok := stubSpecs[i]; ok {
				if stub.SizeSpec != "" {
					sizeSpec = stub.SizeSpec
				}
				if stub.Rate > 0 {
					rate = stub.Rate
				}
				if stub.InvoicedQty > 0 {
					invoicedQty = stub.InvoicedQty
				}
				if len(stub.Attrs) > 0 {
					bAttrs = stub.Attrs
				}
			} else if i <= len(p.Bundles) {
				st := p.Bundles[i-1]
				if st.SizeSpec != "" {
					sizeSpec = st.SizeSpec
				}
				if st.Rate > 0 {
					rate = st.Rate
				}
				if st.InvoicedQty > 0 {
					invoicedQty = st.InvoicedQty
				}
				if len(st.Attrs) > 0 {
					bAttrs = st.Attrs
				}
			}

			initGsm := 0.0
			if isFabric {
				initGsm = invoiceGsm
			}

			qrPayloadJSON, _ := json.Marshal(map[string]any{
				"bundleCode":  bundleCode,
				"parentCode":  parentCode,
				"item":        itemName,
				"color":       colorName,
				"sizeSpec":    sizeSpec,
				"invoicedQty": invoicedQty,
				"location":    "A/R01/S01",
				"qcStatus":    "Pending",
				"generatedAt": time.Now().Format(time.RFC3339),
			})

			_, err = tx.Exec(`
				INSERT INTO bundles (
					id, bundle_code, parent_lot_id, parent_code, bundle_sequence, total_bundles,
					item_type, item_name, color_name, invoiced_qty, inward_weight_qty, measured_gsm,
					location_bin, qc_status, qc_remarks, qr_payload, available_qty,
					size_spec, rate, attrs_json, created_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0.0, ?, 'A/R01/S01', 'Pending', '', ?, 0.0, ?, ?, ?, ?)
			`, bundleID, bundleCode, parentLotID, parentCode, i, numBundles,
				inwardType, itemName, colorName, invoicedQty, initGsm,
				string(qrPayloadJSON), sizeSpec, rate, marshalBundleAttrs(bAttrs), now,
			)
			if err != nil {
				return GateInwardResult{}, fmt.Errorf("failed to insert bundle stub %s: %w", bundleCode, err)
			}

			// Graph Node: BUNDLE
			bundleProps := map[string]any{
				"bundle_code":  bundleCode,
				"sequence":     i,
				"total":        numBundles,
				"qc_status":    "Pending",
				"invoiced_qty": invoicedQty,
			}
			if _, err := s.graph.InsertNode(tx, graph.TypeBundle, bundleProps, bundleID); err != nil {
				return GateInwardResult{}, fmt.Errorf("failed to insert bundle graph node: %w", err)
			}

			// Graph Edge: (BUNDLE) -[:PART_OF_LOT]-> (PARENT_LOT)
			if _, err := s.graph.InsertEdge(tx, bundleID, parentLotID, graph.EdgePartOfLot, map[string]any{"sequence": i}); err != nil {
				return GateInwardResult{}, fmt.Errorf("failed to insert edge PART_OF_LOT: %w", err)
			}

			createdStubs = append(createdStubs, BundleStubInfo{
				BundleCode:  bundleCode,
				Sequence:    i,
				Total:       numBundles,
				InvoicedQty: invoicedQty,
				SizeSpec:    sizeSpec,
				Rate:        rate,
				Attrs:       bAttrs,
				Status:      "Pending",
			})
		}
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
	var currentStage, rawLotAttrs, curFabricType string
	var unitRate, billedInvoiceQty, curInvoiceGsm, curDia float64
	err = tx.QueryRow(`
		SELECT id, vendor_name, invoice_no, inward_type, uom, stage, unit_rate, invoice_qty,
		       COALESCE(attrs_json, '{}'), COALESCE(fabric_type, ''), COALESCE(invoice_gsm, 0), COALESCE(dia, 0)
		FROM parent_inward_lots WHERE parent_code = ?
	`, parentCode).Scan(&lotID, &vendorName, &invoiceNo, &inwardType, &uom, &currentStage, &unitRate, &billedInvoiceQty,
		&rawLotAttrs, &curFabricType, &curInvoiceGsm, &curDia)
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

	// Merge Step 2 measured/verified attributes if provided
	mergedAttrsJSON := rawLotAttrs
	if strings.TrimSpace(mergedAttrsJSON) == "" {
		mergedAttrsJSON = "{}"
	}
	if len(p.Attrs) > 0 {
		lotAttrs := make(map[string]any)
		_ = json.Unmarshal([]byte(rawLotAttrs), &lotAttrs)
		for k, v := range p.Attrs {
			lotAttrs[k] = v
		}
		if raw, mErr := json.Marshal(lotAttrs); mErr == nil {
			mergedAttrsJSON = string(raw)
		}
		if strings.EqualFold(inwardType, "Fabric") {
			if ft, ok := p.Attrs["fabricType"].(string); ok && strings.TrimSpace(ft) != "" {
				curFabricType = strings.TrimSpace(ft)
			}
			if gsmVal, ok := p.Attrs["invoiceGsm"]; ok {
				if f := asFloat64(gsmVal); f > 0 {
					curInvoiceGsm = f
				}
			}
			if diaVal, ok := p.Attrs["dia"]; ok {
				if f := asFloat64(diaVal); f > 0 {
					curDia = f
				}
			}
		}
	}

	// 5. Atomic CAS Update on Parent Inward Lot
	res, err := tx.Exec(`
		UPDATE parent_inward_lots
		SET stage = ?, invoice_qty = ?, counted_qty = ?, variance_qty = ?, variance_pct = ?,
		    physical_received_qty = ?, shortfall_qty = ?, discrepancy_amount = ?,
		    attrs_json = ?, fabric_type = ?, invoice_gsm = ?, dia = ?
		WHERE parent_code = ? AND stage IN (?, ?)
	`, string(nextStage), invoicedQty, countedTotal, varianceQty, variancePct,
		countedTotal, (invoicedQty - countedTotal), financialImpact,
		mergedAttrsJSON, curFabricType, curInvoiceGsm, curDia, parentCode,
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

		bAttrsStr := ""
		if len(b.Attrs) > 0 {
			if raw, mErr := json.Marshal(b.Attrs); mErr == nil {
				bAttrsStr = string(raw)
			}
		}
		measuredGsm := b.MeasuredGsm
		measuredDia := b.MeasuredDia
		if !strings.EqualFold(inwardType, "Fabric") {
			measuredGsm = 0
			measuredDia = 0
		}

		_, err = tx.Exec(`
			UPDATE bundles
			SET inward_weight_qty = ?, available_qty = ?, location_bin = ?,
			    invoiced_qty = CASE WHEN ? > 0 THEN ? ELSE invoiced_qty END,
			    measured_gsm = ?, measured_dia = ?, size_spec = COALESCE(NULLIF(?, ''), size_spec),
			    rate = CASE WHEN ? > 0 THEN ? ELSE rate END,
			    attrs_json = CASE WHEN ? != '' THEN ? ELSE attrs_json END,
			    bundle_comment = ?
			WHERE id = ?
		`, b.CountedQty, b.CountedQty, loc,
			b.InvoicedQty, b.InvoicedQty,
			measuredGsm, measuredDia, b.SizeSpec,
			b.Rate, b.Rate,
			bAttrsStr, bAttrsStr,
			b.Comment, bID,
		)
		if err != nil {
			return FloorCountResult{}, fmt.Errorf("failed to update bundle %s: %w", b.BundleCode, err)
		}
		bundlesUpdated++

		// Bi-Temporal Graph Edge: Close previous STORED_AT and open new STORED_AT
		locNodeID := fmt.Sprintf("loc_%s", strings.ToUpper(loc))

		// Ensure LOCATION node exists
		if _, err := s.graph.EnsureNode(tx, graph.TypeLocation, map[string]any{"code": loc}, locNodeID); err != nil {
			return FloorCountResult{}, fmt.Errorf("failed to ensure location graph node: %w", err)
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
				id, bundle_id, parent_code, bundle_code, result, remarks, image_url, audio_url, transcript, inspected_by, inspected_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, qcInspID, bundleID, parentCode, cleanBundleCode, qcResultRecord, item.Remarks, item.ImageURL, item.AudioURL, item.Transcript, inspector, now)
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
			SET qc_status = ?, qc_remarks = ?, qc_image_url = ?, qc_audio_url = ?, qc_transcript = ?, qc_inspected_at = ?, qr_payload = ?
			WHERE id = ?
		`, resStatus, item.Remarks, item.ImageURL, item.AudioURL, item.Transcript, now, string(qrPayloadJSON), bundleID)
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
// INWARD INVOICES & DETAIL QUERIES
// -------------------------------------------------------------

func (s *Service) SearchInvoices(query string) ([]InwardInvoice, error) {
	clean := strings.TrimSpace(query)
	sqlQuery := `
		SELECT i.id, i.invoice_no, i.vendor_name, i.invoice_date,
		       COALESCE(NULLIF(i.inward_type, ''), (SELECT p.inward_type FROM parent_inward_lots p WHERE p.invoice_no = i.invoice_no LIMIT 1), ''),
		       i.invoice_qty, i.invoice_amount,
		       COALESCE(i.uom, 'Kgs'), COALESCE(i.currency, 'INR'), COALESCE(i.remarks, ''), i.created_at,
		       COALESCE((SELECT GROUP_CONCAT(p.parent_code) FROM parent_inward_lots p WHERE p.invoice_no = i.invoice_no), ''),
		       COALESCE((SELECT SUM(p.invoice_qty) FROM parent_inward_lots p WHERE p.invoice_no = i.invoice_no), 0)
		FROM inward_invoices i
	`
	var rows *sql.Rows
	var err error
	if clean != "" {
		sqlQuery += ` WHERE i.invoice_no LIKE ? OR i.vendor_name LIKE ? ORDER BY i.created_at DESC LIMIT 30`
		rows, err = s.db.Query(sqlQuery, "%"+clean+"%", "%"+clean+"%")
	} else {
		sqlQuery += ` ORDER BY i.created_at DESC LIMIT 30`
		rows, err = s.db.Query(sqlQuery)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []InwardInvoice
	for rows.Next() {
		var inv InwardInvoice
		var rawLots string
		if err := rows.Scan(
			&inv.ID, &inv.InvoiceNo, &inv.VendorName, &inv.InvoiceDate, &inv.InwardType, &inv.InvoiceQty, &inv.InvoiceAmount,
			&inv.UOM, &inv.Currency, &inv.Remarks, &inv.CreatedAt, &rawLots, &inv.TotalLotsQty,
		); err != nil {
			return nil, err
		}
		if rawLots != "" {
			inv.TaggedLots = strings.Split(rawLots, ",")
		} else {
			inv.TaggedLots = []string{}
		}
		list = append(list, inv)
	}
	if list == nil {
		list = []InwardInvoice{}
	}
	return list, nil
}

func (s *Service) GetInvoice(invoiceNo string) (*InwardInvoice, error) {
	clean := strings.TrimSpace(invoiceNo)
	var inv InwardInvoice
	var rawLots string
	err := s.db.QueryRow(`
		SELECT i.id, i.invoice_no, i.vendor_name, i.invoice_date,
		       COALESCE(NULLIF(i.inward_type, ''), (SELECT p.inward_type FROM parent_inward_lots p WHERE p.invoice_no = i.invoice_no LIMIT 1), ''),
		       i.invoice_qty, i.invoice_amount,
		       COALESCE(i.uom, 'Kgs'), COALESCE(i.currency, 'INR'), COALESCE(i.remarks, ''), i.created_at,
		       COALESCE((SELECT GROUP_CONCAT(p.parent_code) FROM parent_inward_lots p WHERE p.invoice_no = i.invoice_no), ''),
		       COALESCE((SELECT SUM(p.invoice_qty) FROM parent_inward_lots p WHERE p.invoice_no = i.invoice_no), 0)
		FROM inward_invoices i
		WHERE i.invoice_no = ?
		LIMIT 1
	`, clean).Scan(
		&inv.ID, &inv.InvoiceNo, &inv.VendorName, &inv.InvoiceDate, &inv.InwardType, &inv.InvoiceQty, &inv.InvoiceAmount,
		&inv.UOM, &inv.Currency, &inv.Remarks, &inv.CreatedAt, &rawLots, &inv.TotalLotsQty,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if rawLots != "" {
		inv.TaggedLots = strings.Split(rawLots, ",")
	} else {
		inv.TaggedLots = []string{}
	}
	return &inv, nil
}

// -------------------------------------------------------------
// LISTING & DETAIL QUERIES
// -------------------------------------------------------------

func asFloat64(v any) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case string:
		var parsed float64
		cleaned := strings.ReplaceAll(strings.TrimSpace(val), `"`, "")
		_, _ = fmt.Sscanf(cleaned, "%f", &parsed)
		return parsed
	default:
		return 0
	}
}

func attrMissing(attrs map[string]any, key string) bool {
	v, ok := attrs[key]
	if !ok || v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s) == ""
	}
	if f, ok := v.(float64); ok {
		return f == 0
	}
	return false
}

// mergeLotColumnFallback backfills Attrs and FabricProducts from top-level SQL columns
// so seeded or legacy lots never lose GSM, Dia, Fabric Type, or Rate in UI stations.
func mergeLotColumnFallback(item *LotSummary, fabricType string, invoiceGsm, dia, unitRate float64, vendorFabName, fpccFabName string) {
	if item.Attrs == nil {
		item.Attrs = make(map[string]any)
	}
	if fabricType != "" && attrMissing(item.Attrs, "fabricType") {
		item.Attrs["fabricType"] = fabricType
	}
	if invoiceGsm > 0 && attrMissing(item.Attrs, "invoiceGsm") {
		item.Attrs["invoiceGsm"] = invoiceGsm
	}
	if dia > 0 && attrMissing(item.Attrs, "dia") {
		item.Attrs["dia"] = dia
	}
	if unitRate > 0 && attrMissing(item.Attrs, "rate") {
		item.Attrs["rate"] = unitRate
	}
	if vendorFabName != "" && attrMissing(item.Attrs, "vendorFabName") {
		item.Attrs["vendorFabName"] = vendorFabName
	}
	if fpccFabName != "" && attrMissing(item.Attrs, "fpccFabName") {
		item.Attrs["fpccFabName"] = fpccFabName
	}

	if strings.EqualFold(strings.TrimSpace(item.InwardType), "Fabric") && len(item.FabricProducts) == 0 {
		comp, _ := item.Attrs["composition"].(string)
		diaStr := ""
		if dia > 0 {
			diaStr = fmt.Sprintf("%.0f\"", dia)
		}
		fpName := fpccFabName
		if fpName == "" {
			fpName = item.FpccProduct
		}
		vpName := vendorFabName
		if vpName == "" {
			vpName = item.VendorProduct
		}
		if fabricType != "" || invoiceGsm > 0 || dia > 0 || fpName != "" || vpName != "" {
			rolls := item.NoOfBundles
			if rolls <= 0 {
				rolls = 1
			}
			item.FabricProducts = []FabricProductItem{
				{
					ID:            "fp_1",
					FabricType:    fabricType,
					FpccProduct:   fpName,
					VendorProduct: vpName,
					Color:         item.FpccColor,
					UOM:           item.UOM,
					Dia:           diaStr,
					GSM:           invoiceGsm,
					Composition:   comp,
					Rate:          unitRate,
					RollCount:     rolls,
				},
			}
		}
	}
}

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
			COALESCE(p.remarks, ''), COALESCE(p.attrs_json, '{}'),
			COALESCE(p.invoice_id, ''), COALESCE(p.fabric_products_json, '[]'), COALESCE(p.qr_copies_per_bundle, 3),
			COALESCE(p.fabric_type, ''), COALESCE(p.invoice_gsm, 0), COALESCE(p.dia, 0),
			COALESCE(p.unit_rate, 0), COALESCE(p.vendor_fab_name, ''), COALESCE(p.fpcc_fab_name, ''),
			p.created_at
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
		var rawAttrs, rawFabProds, fabricType, vendorFabName, fpccFabName string
		var invoiceGsm, dia, unitRate float64
		err := rows.Scan(
			&item.ID, &item.ParentCode, &item.InwardDate, &item.InwardType, &item.VendorName, &item.InvoiceNo,
			&item.InvoiceQty, &item.InvoiceAmount, &item.CountedQty, &item.VarianceQty, &item.VariancePct,
			&item.UOM, &item.VendorProduct, &item.FpccProduct,
			&item.VendorColor, &item.FpccColor,
			&item.Stage, &item.NoOfBundles,
			&item.InspectedBundles, &item.PassedBundles, &item.RejectedBundles,
			&item.Remarks, &rawAttrs,
			&item.InvoiceID, &rawFabProds, &item.QRCopiesPerBundle,
			&fabricType, &invoiceGsm, &dia, &unitRate, &vendorFabName, &fpccFabName,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		item.Attrs = make(map[string]any)
		if rawAttrs != "" {
			_ = json.Unmarshal([]byte(rawAttrs), &item.Attrs)
		}
		item.FabricProducts = []FabricProductItem{}
		if rawFabProds != "" {
			_ = json.Unmarshal([]byte(rawFabProds), &item.FabricProducts)
		}
		mergeLotColumnFallback(&item, fabricType, invoiceGsm, dia, unitRate, vendorFabName, fpccFabName)
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
	var rawAttrs, rawFabProds, fabricType, vendorFabName, fpccFabName string
	var invoiceGsm, dia, unitRate float64
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
			COALESCE(p.remarks, ''), COALESCE(p.attrs_json, '{}'),
			COALESCE(p.invoice_id, ''), COALESCE(p.fabric_products_json, '[]'), COALESCE(p.qr_copies_per_bundle, 3),
			COALESCE(p.fabric_type, ''), COALESCE(p.invoice_gsm, 0), COALESCE(p.dia, 0),
			COALESCE(p.unit_rate, 0), COALESCE(p.vendor_fab_name, ''), COALESCE(p.fpcc_fab_name, ''),
			p.created_at
		FROM parent_inward_lots p
		WHERE p.parent_code = ?
	`, cleanCode).Scan(
		&item.ID, &item.ParentCode, &item.InwardDate, &item.InwardType, &item.VendorName, &item.InvoiceNo,
		&item.InvoiceQty, &item.InvoiceAmount, &item.CountedQty, &item.VarianceQty, &item.VariancePct,
		&item.UOM, &item.VendorProduct, &item.FpccProduct,
		&item.VendorColor, &item.FpccColor,
		&item.Stage, &item.NoOfBundles,
		&item.InspectedBundles, &item.PassedBundles, &item.RejectedBundles,
		&item.Remarks, &rawAttrs,
		&item.InvoiceID, &rawFabProds, &item.QRCopiesPerBundle,
		&fabricType, &invoiceGsm, &dia, &unitRate, &vendorFabName, &fpccFabName,
		&item.CreatedAt,
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
	item.FabricProducts = []FabricProductItem{}
	if rawFabProds != "" {
		_ = json.Unmarshal([]byte(rawFabProds), &item.FabricProducts)
	}
	mergeLotColumnFallback(&item, fabricType, invoiceGsm, dia, unitRate, vendorFabName, fpccFabName)

	// 2. Fetch Bundles
	bundleRows, err := s.db.Query(`
		SELECT 
			id, bundle_code, parent_code, bundle_sequence, total_bundles,
			item_type, item_name, COALESCE(color_name, ''), COALESCE(invoiced_qty, 0), inward_weight_qty,
			COALESCE(measured_gsm, 0), COALESCE(measured_dia, 0),
			COALESCE(size_spec, ''), COALESCE(rate, 0), COALESCE(attrs_json, '{}'), location_bin,
			qc_status, COALESCE(qc_remarks, ''), COALESCE(qc_image_url, ''),
			COALESCE(qc_audio_url, ''), COALESCE(qc_transcript, ''),
			COALESCE(qc_inspected_at, 0), COALESCE(bundle_comment, ''), COALESCE(qr_payload, '{}')
		FROM bundles
		WHERE parent_code = ?
		ORDER BY bundle_sequence ASC
	`, cleanCode)
	if err != nil {
		return LotDetail{}, err
	}
	defer bundleRows.Close()

	bundles := make([]BundleDetail, 0)
	for bundleRows.Next() {
		var b BundleDetail
		var bAttrsStr string
		err := bundleRows.Scan(
			&b.ID, &b.BundleCode, &b.ParentCode, &b.BundleSequence, &b.TotalBundles,
			&b.ItemType, &b.ItemName, &b.ColorName, &b.InvoicedQty, &b.InwardQty,
			&b.MeasuredGsm, &b.MeasuredDia,
			&b.SizeSpec, &b.Rate, &bAttrsStr, &b.LocationBin,
			&b.QcStatus, &b.QcRemarks, &b.QcImageUrl,
			&b.QcAudioUrl, &b.QcTranscript,
			&b.QcInspectedAt, &b.BundleComment, &b.QRPayload,
		)
		if err != nil {
			return LotDetail{}, err
		}
		b.Attrs = make(map[string]any)
		if bAttrsStr != "" {
			_ = json.Unmarshal([]byte(bAttrsStr), &b.Attrs)
		}
		if b.InvoicedQty == 0 && item.InvoiceQty > 0 && b.TotalBundles > 0 {
			b.InvoicedQty = math.Round((item.InvoiceQty/float64(b.TotalBundles))*1000) / 1000
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
	reconciliations := make([]map[string]any, 0)
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
	inspections := make([]map[string]any, 0)
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
