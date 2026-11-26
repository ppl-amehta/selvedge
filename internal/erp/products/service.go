package products

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ppl-amehta/selvedge/internal/graph"
	"github.com/ppl-amehta/selvedge/internal/workflow"
)

var (
	reAlphanumeric = regexp.MustCompile(`[^A-Za-z0-9]`)
)

type Service struct {
	db       *sql.DB
	graph    *graph.Store
	workflow *workflow.Engine
}

func NewService(db *sql.DB, graphStore *graph.Store, wf *workflow.Engine) *Service {
	return &Service{
		db:       db,
		graph:    graphStore,
		workflow: wf,
	}
}

// -------------------------------------------------------------
// STRUCTS: SAMPLE DEVELOPMENT
// -------------------------------------------------------------

type SampleSizeItem struct {
	ID       string `json:"id"`
	Gender   string `json:"gender"`
	Size     string `json:"size"`
	Qty      int    `json:"qty"`
	Sequence int    `json:"sequence"`
}

type SampleFabricItem struct {
	ID         string         `json:"id"`
	FabricName string         `json:"fabricName"`
	LotNo      string         `json:"lotNo"`
	Average    float64        `json:"average"`
	Attrs      map[string]any `json:"attrs,omitempty"`
	Sequence   int            `json:"sequence"`
}

type SampleTrimItem struct {
	ID       string         `json:"id"`
	TrimName string         `json:"trimName"`
	Average  float64        `json:"average"`
	Attrs    map[string]any `json:"attrs,omitempty"`
	Sequence int            `json:"sequence"`
}

type SampleZipItem struct {
	ID           string         `json:"id"`
	ZipName      string         `json:"zipName"`
	LengthAndQty string         `json:"lengthAndQty"`
	Attrs        map[string]any `json:"attrs,omitempty"`
	Sequence     int            `json:"sequence"`
}

type SampleLabelItem struct {
	ID        string         `json:"id"`
	LabelName string         `json:"labelName"`
	Attrs     map[string]any `json:"attrs,omitempty"`
	Sequence  int            `json:"sequence"`
}

type SamplePharmaItem struct {
	ID           string `json:"id"`
	BasePharmaNo string `json:"basePharmaNo"`
	ChangesMade  string `json:"changesMade"`
	NewPharmaNo  string `json:"newPharmaNo"`
	Sequence     int    `json:"sequence"`
}

type SampleDevelopmentPayload struct {
	ID                   string             `json:"id"`
	DevelopmentNo        string             `json:"developmentNo"`
	DevelopmentDate      string             `json:"developmentDate"`
	ClientID             string             `json:"clientId"`
	ClientName           string             `json:"clientName"`
	ProductGenre         string             `json:"productGenre"`
	TempProductName      string             `json:"tempProductName"`
	Colors               []string           `json:"colors,omitempty"`
	ImageURL             string             `json:"imageUrl"`
	ProductDescription   string             `json:"productDescription"`
	TempMeasurementChart string             `json:"tempMeasurementChart"`
	DevComments          string             `json:"devComments"`
	MastersComments      string             `json:"mastersComments"`
	ReworkNotes          string             `json:"reworkNotes"`
	SampleImageURL       string             `json:"sampleImageUrl"`
	Comments             string             `json:"comments"`
	Status               string             `json:"status"`
	Sizes                []SampleSizeItem   `json:"sizes"`
	Fabrics              []SampleFabricItem `json:"fabrics"`
	Trims                []SampleTrimItem   `json:"trims"`
	Zips                 []SampleZipItem    `json:"zips"`
	Labels               []SampleLabelItem  `json:"labels"`
	Pharmas              []SamplePharmaItem `json:"pharmas"`
}

type SampleDevelopmentDetail struct {
	SampleDevelopmentPayload
	CreatedAt int64 `json:"createdAt"`
	UpdatedAt int64 `json:"updatedAt"`
}

type SampleDevelopmentSummary struct {
	ID              string `json:"id"`
	DevelopmentNo   string `json:"developmentNo"`
	DevelopmentDate string `json:"developmentDate"`
	ClientID        string `json:"clientId"`
	ClientName      string `json:"clientName"`
	ProductGenre    string `json:"productGenre"`
	TempProductName string `json:"tempProductName"`
	ImageURL        string `json:"imageUrl"`
	SampleImageURL  string `json:"sampleImageUrl"`
	Status          string `json:"status"`
	TotalQty        int    `json:"totalQty"`
	CreatedAt       int64  `json:"createdAt"`
	UpdatedAt       int64  `json:"updatedAt"`
}

// -------------------------------------------------------------
// STRUCTS: PRODUCT CREATION
// -------------------------------------------------------------

type ProductGenderItem struct {
	ID                  string           `json:"id"`
	Gender              string           `json:"gender"`
	Sizes               []string         `json:"sizes"`
	PharmaNo            string           `json:"pharmaNo"`
	PharmaNos           []string         `json:"pharmaNos"`
	MeasurementChartURL string           `json:"measurementChartUrl"`
	ProductImages       []string         `json:"productImages"`
	ColorImages         []map[string]any `json:"colorImages"`
	Sequence            int              `json:"sequence"`
}

type ProductMaterialItem struct {
	ID           string  `json:"id"`
	MaterialType string  `json:"materialType"` // 'fabric', 'trim', 'zip'
	Name         string  `json:"name"`
	Average      float64 `json:"average"`
	GSM          float64 `json:"gsm"`
	Dia          float64 `json:"dia"`
	ZipLengthQty string  `json:"zipLengthQty"`
	Sequence     int     `json:"sequence"`
}

type CreateProductPayload struct {
	ID                    string                `json:"id"`
	ProductName           string                `json:"productName"`
	Genre                 string                `json:"genre"`
	SKUCode               string                `json:"skuCode"`
	BrandLabel            string                `json:"brandLabel"`
	WhyWeMadeThis         string                `json:"whyWeMadeThis"`
	GarmentSpecifications string                `json:"garmentSpecifications"`
	Composition           string                `json:"composition"`
	MRP                   float64               `json:"mrp"`
	SampleDevelopmentID   string                `json:"sampleDevelopmentId"`
	Genders               []ProductGenderItem   `json:"genders"`
	Materials             []ProductMaterialItem `json:"materials"`
}

type MasterProductDetail struct {
	CreateProductPayload
	CreatedAt int64 `json:"createdAt"`
	UpdatedAt int64 `json:"updatedAt"`
}

// -------------------------------------------------------------
// STRUCTS: CLIENT SPECIFIC PRICING
// -------------------------------------------------------------

type ClientPricingPayload struct {
	ID              string  `json:"id"`
	ClientID        string  `json:"clientId"`
	ClientName      string  `json:"clientName"`
	ProductID       string  `json:"productId"`
	ProductName     string  `json:"productName"`
	ClientImageURL  string  `json:"clientImageUrl"`
	ContractedPrice float64 `json:"contractedPrice"`
	Comments        string  `json:"comments"`
	EffectiveDate   string  `json:"effectiveDate"`
}

type ClientPricingDetail struct {
	ClientPricingPayload
	CreatedAt int64 `json:"createdAt"`
	UpdatedAt int64 `json:"updatedAt"`
}

// -------------------------------------------------------------
// METHODS: SAMPLE DEVELOPMENT
// -------------------------------------------------------------

type rowQueryer interface {
	QueryRow(query string, args ...any) *sql.Row
}

func fetchNextDevNo(q rowQueryer) (string, error) {
	now := time.Now()
	prefix := fmt.Sprintf("D%02d%02d", now.Year()%100, int(now.Month()))
	pattern := prefix + "____"

	var lastNo sql.NullString
	query := "SELECT development_no FROM sample_developments WHERE development_no LIKE ? AND length(development_no) = ? ORDER BY development_no DESC LIMIT 1"
	if err := q.QueryRow(query, pattern, len(prefix)+4).Scan(&lastNo); err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to scan latest development no: %w", err)
	}

	nextSeq := 1
	if lastNo.Valid && len(lastNo.String) >= len(prefix)+4 {
		seqPart := lastNo.String[len(prefix):]
		if n, err := strconv.Atoi(seqPart); err == nil {
			nextSeq = n + 1
		}
	}

	return fmt.Sprintf("%s%04d", prefix, nextSeq), nil
}

// SuggestNextDevelopmentNo computes the next sequence 'D' + YYMM + 0001 (e.g. D26110001)
func (s *Service) SuggestNextDevelopmentNo() (string, error) {
	return fetchNextDevNo(s.db)
}

// allocateNextDevNoTx allocates the next sequence inside an immediate transaction
func (s *Service) allocateNextDevNoTx(tx *sql.Tx) (string, error) {
	return fetchNextDevNo(tx)
}

// CreateSampleDevelopment registers a new sample or updates an existing draft
func (s *Service) CreateSampleDevelopment(payload SampleDevelopmentPayload) (map[string]any, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	sampleID := payload.ID
	if sampleID == "" {
		sampleID = "smp_" + uuid.New().String()
	}

	// Development No allocation: if updating existing sample, preserve its original development_no
	devNo := strings.TrimSpace(payload.DevelopmentNo)
	if sampleID != "" {
		var existingDevNo string
		if err := tx.QueryRow("SELECT development_no FROM sample_developments WHERE id = ?", sampleID).Scan(&existingDevNo); err == nil && existingDevNo != "" {
			devNo = existingDevNo
		}
	}
	if devNo == "" {
		allocated, err := s.allocateNextDevNoTx(tx)
		if err != nil {
			return nil, err
		}
		devNo = allocated
	}

	devDate := strings.TrimSpace(payload.DevelopmentDate)
	if devDate == "" {
		devDate = time.Now().Format("2006-01-02")
	}

	status := strings.TrimSpace(payload.Status)
	if status == "" {
		status = "DRAFT"
	}

	nowMs := graph.NowMs()

	nonNilStrs := func(in []string) []string {
		if in == nil {
			return []string{}
		}
		return in
	}
	colorsBytes, _ := json.Marshal(nonNilStrs(payload.Colors))
	marshalAttrs := func(m map[string]any) string {
		if len(m) == 0 {
			return "{}"
		}
		b, err := json.Marshal(m)
		if err != nil {
			return "{}"
		}
		return string(b)
	}

	// Upsert sample_developments
	sampleQuery := `
		INSERT INTO sample_developments (
			id, development_no, development_date, client_id, client_name,
			product_genre, temp_product_name, colors_json, image_url, product_description,
			temp_measurement_chart, dev_comments, masters_comments, rework_notes,
			sample_image_url, comments, status, meta_json, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '{}', ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			development_date = excluded.development_date,
			client_id = excluded.client_id,
			client_name = excluded.client_name,
			product_genre = excluded.product_genre,
			temp_product_name = excluded.temp_product_name,
			colors_json = excluded.colors_json,
			image_url = excluded.image_url,
			product_description = excluded.product_description,
			temp_measurement_chart = excluded.temp_measurement_chart,
			dev_comments = excluded.dev_comments,
			masters_comments = excluded.masters_comments,
			rework_notes = excluded.rework_notes,
			sample_image_url = excluded.sample_image_url,
			comments = excluded.comments,
			status = excluded.status,
			updated_at = excluded.updated_at
	`
	if _, err := tx.Exec(sampleQuery,
		sampleID, devNo, devDate, payload.ClientID, payload.ClientName,
		payload.ProductGenre, payload.TempProductName, string(colorsBytes), payload.ImageURL, payload.ProductDescription,
		payload.TempMeasurementChart, payload.DevComments, payload.MastersComments, payload.ReworkNotes,
		payload.SampleImageURL, payload.Comments, status, nowMs, nowMs,
	); err != nil {
		return nil, fmt.Errorf("failed to upsert sample_developments: %w", err)
	}

	// Clear and re-insert child records for clean idempotent saves
	tablesToClear := []string{
		"sample_sizes", "sample_fabrics", "sample_trims",
		"sample_zips", "sample_labels", "sample_pharmas",
	}
	for _, tbl := range tablesToClear {
		if _, err := tx.Exec("DELETE FROM "+tbl+" WHERE development_id = ?", sampleID); err != nil {
			return nil, fmt.Errorf("failed to clear %s: %w", tbl, err)
		}
	}

	// 1. Sizes (contiguous sequence)
	sizeStmt, err := tx.Prepare("INSERT INTO sample_sizes (id, development_id, gender, size, qty, sequence) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}
	defer sizeStmt.Close()
	sizeSeq := 0
	for _, sz := range payload.Sizes {
		if strings.TrimSpace(sz.Size) == "" {
			continue
		}
		sid := "ssz_" + uuid.New().String()
		qty := sz.Qty
		if qty <= 0 {
			qty = 1
		}
		if _, err := sizeStmt.Exec(sid, sampleID, sz.Gender, sz.Size, qty, sizeSeq); err != nil {
			return nil, fmt.Errorf("failed to insert sample size: %w", err)
		}
		sizeSeq++
	}

	// 2. Fabrics
	fabStmt, err := tx.Prepare("INSERT INTO sample_fabrics (id, development_id, fabric_name, lot_no, average, attrs_json, sequence) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}
	defer fabStmt.Close()
	fabSeq := 0
	for _, f := range payload.Fabrics {
		fabName := strings.TrimSpace(f.FabricName)
		if fabName == "" && f.Attrs != nil {
			if v, ok := f.Attrs["fabricType"].(string); ok {
				fabName = strings.TrimSpace(v)
			}
		}
		if fabName == "" && len(f.Attrs) == 0 {
			continue
		}
		fid := "sfb_" + uuid.New().String()
		if _, err := fabStmt.Exec(fid, sampleID, fabName, f.LotNo, f.Average, marshalAttrs(f.Attrs), fabSeq); err != nil {
			return nil, fmt.Errorf("failed to insert sample fabric: %w", err)
		}
		fabSeq++
	}

	// 3. Trims
	trimStmt, err := tx.Prepare("INSERT INTO sample_trims (id, development_id, trim_name, average, attrs_json, sequence) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}
	defer trimStmt.Close()
	trimSeq := 0
	for _, t := range payload.Trims {
		trimName := strings.TrimSpace(t.TrimName)
		if trimName == "" && t.Attrs != nil {
			if v, ok := t.Attrs["trimItem"].(string); ok && strings.TrimSpace(v) != "" {
				trimName = strings.TrimSpace(v)
			} else if v, ok := t.Attrs["trimType"].(string); ok {
				trimName = strings.TrimSpace(v)
			}
		}
		if trimName == "" && len(t.Attrs) == 0 {
			continue
		}
		tid := "trm_" + uuid.New().String()
		if _, err := trimStmt.Exec(tid, sampleID, trimName, t.Average, marshalAttrs(t.Attrs), trimSeq); err != nil {
			return nil, fmt.Errorf("failed to insert sample trim: %w", err)
		}
		trimSeq++
	}

	// 4. Zips
	zipStmt, err := tx.Prepare("INSERT INTO sample_zips (id, development_id, zip_name, length_and_qty, attrs_json, sequence) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}
	defer zipStmt.Close()
	zipSeq := 0
	for _, z := range payload.Zips {
		zipName := strings.TrimSpace(z.ZipName)
		if zipName == "" && z.Attrs != nil {
			if v, ok := z.Attrs["zipType"].(string); ok {
				zipName = strings.TrimSpace(v)
			}
		}
		if zipName == "" && len(z.Attrs) == 0 {
			continue
		}
		zid := "szp_" + uuid.New().String()
		if _, err := zipStmt.Exec(zid, sampleID, zipName, z.LengthAndQty, marshalAttrs(z.Attrs), zipSeq); err != nil {
			return nil, fmt.Errorf("failed to insert sample zip: %w", err)
		}
		zipSeq++
	}

	// 5. Labels
	lblStmt, err := tx.Prepare("INSERT INTO sample_labels (id, development_id, label_name, attrs_json, sequence) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}
	defer lblStmt.Close()
	lblSeq := 0
	for _, l := range payload.Labels {
		lblName := strings.TrimSpace(l.LabelName)
		if lblName == "" && l.Attrs != nil {
			if v, ok := l.Attrs["labelGenre"].(string); ok {
				lblName = strings.TrimSpace(v)
			}
		}
		if lblName == "" && len(l.Attrs) == 0 {
			continue
		}
		lid := "slb_" + uuid.New().String()
		if _, err := lblStmt.Exec(lid, sampleID, lblName, marshalAttrs(l.Attrs), lblSeq); err != nil {
			return nil, fmt.Errorf("failed to insert sample label: %w", err)
		}
		lblSeq++
	}

	// 6. Pharmas
	phrStmt, err := tx.Prepare("INSERT INTO sample_pharmas (id, development_id, base_pharma_no, changes_made, new_pharma_no, sequence) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}
	defer phrStmt.Close()
	phrSeq := 0
	for _, p := range payload.Pharmas {
		if strings.TrimSpace(p.BasePharmaNo) == "" && strings.TrimSpace(p.NewPharmaNo) == "" {
			continue
		}
		pid := "sph_" + uuid.New().String()
		if _, err := phrStmt.Exec(pid, sampleID, p.BasePharmaNo, p.ChangesMade, p.NewPharmaNo, phrSeq); err != nil {
			return nil, fmt.Errorf("failed to insert sample pharma: %w", err)
		}
		phrSeq++
	}

	// 7. Graph Lineage Node & Edge
	if s.graph != nil {
		nodeProps := map[string]any{
			"development_no":    devNo,
			"client_name":       payload.ClientName,
			"temp_product_name": payload.TempProductName,
			"genre":             payload.ProductGenre,
			"status":            status,
		}
		_, _ = s.graph.InsertNode(tx, "SAMPLE_DEVELOPMENT", nodeProps, sampleID)

		if payload.ClientID != "" {
			_, _ = s.graph.InsertEdge(tx, sampleID, payload.ClientID, "REQUESTED_BY", map[string]any{
				"development_no": devNo,
				"created_at":     nowMs,
			})
		}
	}

	// 8. Domain Event Outbox
	if s.workflow != nil {
		_, _ = s.workflow.EmitEvent(tx, sampleID, "sample.development.saved", map[string]any{
			"developmentNo":   devNo,
			"clientName":      payload.ClientName,
			"tempProductName": payload.TempProductName,
			"status":          status,
		}, 0)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit sample transaction: %w", err)
	}

	return map[string]any{
		"success":       true,
		"id":            sampleID,
		"developmentNo": devNo,
		"status":        status,
		"message":       "Sample development recorded successfully",
	}, nil
}

// GetSampleDevelopment retrieves full sample master-detail tree
func (s *Service) GetSampleDevelopment(idOrNo string) (*SampleDevelopmentDetail, error) {
	var d SampleDevelopmentDetail
	var colorsJSON string
	query := `
		SELECT
			id, development_no, development_date, client_id, client_name,
			coalesce(product_genre, ''), coalesce(temp_product_name, ''),
			coalesce(colors_json, '[]'),
			coalesce(image_url, ''), coalesce(product_description, ''),
			coalesce(temp_measurement_chart, ''), coalesce(dev_comments, ''),
			coalesce(masters_comments, ''), coalesce(rework_notes, ''),
			coalesce(sample_image_url, ''), coalesce(comments, ''),
			status, created_at, updated_at
		FROM sample_developments
		WHERE id = ? OR development_no = ?
	`
	err := s.db.QueryRow(query, idOrNo, idOrNo).Scan(
		&d.ID, &d.DevelopmentNo, &d.DevelopmentDate, &d.ClientID, &d.ClientName,
		&d.ProductGenre, &d.TempProductName, &colorsJSON, &d.ImageURL, &d.ProductDescription,
		&d.TempMeasurementChart, &d.DevComments, &d.MastersComments, &d.ReworkNotes,
		&d.SampleImageURL, &d.Comments, &d.Status, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	d.Colors = make([]string, 0)
	if colorsJSON != "" {
		_ = json.Unmarshal([]byte(colorsJSON), &d.Colors)
	}
	if d.Colors == nil {
		d.Colors = make([]string, 0)
	}

	// Child collections
	d.Sizes = make([]SampleSizeItem, 0)
	szRows, err := s.db.Query("SELECT id, gender, size, qty, sequence FROM sample_sizes WHERE development_id = ? ORDER BY sequence", d.ID)
	if err == nil {
		defer szRows.Close()
		for szRows.Next() {
			var sz SampleSizeItem
			if err := szRows.Scan(&sz.ID, &sz.Gender, &sz.Size, &sz.Qty, &sz.Sequence); err == nil {
				d.Sizes = append(d.Sizes, sz)
			}
		}
	}

	d.Fabrics = make([]SampleFabricItem, 0)
	fbRows, err := s.db.Query("SELECT id, fabric_name, coalesce(lot_no, ''), average, coalesce(attrs_json, '{}'), sequence FROM sample_fabrics WHERE development_id = ? ORDER BY sequence", d.ID)
	if err == nil {
		defer fbRows.Close()
		for fbRows.Next() {
			var fb SampleFabricItem
			var attrsStr string
			if err := fbRows.Scan(&fb.ID, &fb.FabricName, &fb.LotNo, &fb.Average, &attrsStr, &fb.Sequence); err == nil {
				fb.Attrs = make(map[string]any)
				_ = json.Unmarshal([]byte(attrsStr), &fb.Attrs)
				d.Fabrics = append(d.Fabrics, fb)
			}
		}
	}

	d.Trims = make([]SampleTrimItem, 0)
	trRows, err := s.db.Query("SELECT id, trim_name, average, coalesce(attrs_json, '{}'), sequence FROM sample_trims WHERE development_id = ? ORDER BY sequence", d.ID)
	if err == nil {
		defer trRows.Close()
		for trRows.Next() {
			var tr SampleTrimItem
			var attrsStr string
			if err := trRows.Scan(&tr.ID, &tr.TrimName, &tr.Average, &attrsStr, &tr.Sequence); err == nil {
				tr.Attrs = make(map[string]any)
				_ = json.Unmarshal([]byte(attrsStr), &tr.Attrs)
				d.Trims = append(d.Trims, tr)
			}
		}
	}

	d.Zips = make([]SampleZipItem, 0)
	zpRows, err := s.db.Query("SELECT id, zip_name, coalesce(length_and_qty, ''), coalesce(attrs_json, '{}'), sequence FROM sample_zips WHERE development_id = ? ORDER BY sequence", d.ID)
	if err == nil {
		defer zpRows.Close()
		for zpRows.Next() {
			var zp SampleZipItem
			var attrsStr string
			if err := zpRows.Scan(&zp.ID, &zp.ZipName, &zp.LengthAndQty, &attrsStr, &zp.Sequence); err == nil {
				zp.Attrs = make(map[string]any)
				_ = json.Unmarshal([]byte(attrsStr), &zp.Attrs)
				d.Zips = append(d.Zips, zp)
			}
		}
	}

	d.Labels = make([]SampleLabelItem, 0)
	lbRows, err := s.db.Query("SELECT id, label_name, coalesce(attrs_json, '{}'), sequence FROM sample_labels WHERE development_id = ? ORDER BY sequence", d.ID)
	if err == nil {
		defer lbRows.Close()
		for lbRows.Next() {
			var lb SampleLabelItem
			var attrsStr string
			if err := lbRows.Scan(&lb.ID, &lb.LabelName, &attrsStr, &lb.Sequence); err == nil {
				lb.Attrs = make(map[string]any)
				_ = json.Unmarshal([]byte(attrsStr), &lb.Attrs)
				d.Labels = append(d.Labels, lb)
			}
		}
	}

	d.Pharmas = make([]SamplePharmaItem, 0)
	phRows, err := s.db.Query("SELECT id, coalesce(base_pharma_no, ''), coalesce(changes_made, ''), coalesce(new_pharma_no, ''), sequence FROM sample_pharmas WHERE development_id = ? ORDER BY sequence", d.ID)
	if err == nil {
		defer phRows.Close()
		for phRows.Next() {
			var ph SamplePharmaItem
			if err := phRows.Scan(&ph.ID, &ph.BasePharmaNo, &ph.ChangesMade, &ph.NewPharmaNo, &ph.Sequence); err == nil {
				d.Pharmas = append(d.Pharmas, ph)
			}
		}
	}

	return &d, nil
}

// ListSampleDevelopments queries summaries with pagination
func (s *Service) ListSampleDevelopments(limit, offset int) ([]SampleDevelopmentSummary, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT
			d.id, d.development_no, d.development_date, d.client_id, d.client_name,
			coalesce(d.product_genre, ''), coalesce(d.temp_product_name, ''),
			coalesce(d.image_url, ''), coalesce(d.sample_image_url, ''), d.status,
			coalesce((SELECT SUM(qty) FROM sample_sizes WHERE development_id = d.id), 0) AS total_qty,
			d.created_at, d.updated_at
		FROM sample_developments d
		ORDER BY d.created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]SampleDevelopmentSummary, 0)
	for rows.Next() {
		var item SampleDevelopmentSummary
		if err := rows.Scan(
			&item.ID, &item.DevelopmentNo, &item.DevelopmentDate, &item.ClientID, &item.ClientName,
			&item.ProductGenre, &item.TempProductName, &item.ImageURL, &item.SampleImageURL, &item.Status,
			&item.TotalQty, &item.CreatedAt, &item.UpdatedAt,
		); err == nil {
			list = append(list, item)
		}
	}
	return list, nil
}

// UpdateSampleDevelopmentStatus modifies status, records rework notes and timestamps
func (s *Service) UpdateSampleDevelopmentStatus(id string, status string, reworkNotes string) error {
	nowMs := graph.NowMs()
	query := `
		UPDATE sample_developments
		SET status = ?, rework_notes = CASE WHEN ? != '' THEN ? ELSE rework_notes END, updated_at = ?
		WHERE id = ? OR development_no = ?
	`
	_, err := s.db.Exec(query, status, reworkNotes, reworkNotes, nowMs, id, id)
	return err
}

// -------------------------------------------------------------
// METHODS: PRODUCT CREATION & SKU GENERATION
// -------------------------------------------------------------

// SuggestSKU generates ET_PO_LINEFL format:
// First 2 chars of brand, 2 chars of genre, 6 chars of product name, sanitized and uppercase.
// Resolves collisions against master_products.
func (s *Service) SuggestSKU(brandLabel, genre, productName string) string {
	bClean := reAlphanumeric.ReplaceAllString(strings.ToUpper(strings.TrimSpace(brandLabel)), "")
	gClean := reAlphanumeric.ReplaceAllString(strings.ToUpper(strings.TrimSpace(genre)), "")
	pClean := reAlphanumeric.ReplaceAllString(strings.ToUpper(strings.TrimSpace(productName)), "")

	if len(bClean) < 2 {
		bClean = (bClean + "FP")[:2]
	} else {
		bClean = bClean[:2]
	}

	if len(gClean) < 2 {
		gClean = (gClean + "AP")[:2]
	} else {
		gClean = gClean[:2]
	}

	if len(pClean) < 6 {
		pClean = (pClean + "XXXXXX")[:6]
	} else {
		pClean = pClean[:6]
	}

	baseSKU := fmt.Sprintf("%s_%s_%s", bClean, gClean, pClean)

	// Check collisions in master_products
	candidate := baseSKU
	suffix := 1
	for {
		var exists bool
		_ = s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM master_products WHERE sku_code = ?)", candidate).Scan(&exists)
		if !exists {
			return candidate
		}
		candidate = fmt.Sprintf("%s-%02d", baseSKU, suffix)
		suffix++
		if suffix > 99 {
			break
		}
	}

	return candidate
}

// CreateMasterProduct saves a finished product and mirrors it to master_items
func (s *Service) CreateMasterProduct(payload CreateProductPayload) (map[string]any, error) {
	prodName := strings.TrimSpace(payload.ProductName)
	if prodName == "" {
		return nil, fmt.Errorf("product name is required")
	}

	genre := strings.TrimSpace(payload.Genre)
	if genre == "" {
		genre = "Apparel"
	}

	sku := strings.TrimSpace(payload.SKUCode)
	if sku == "" {
		sku = s.SuggestSKU(payload.BrandLabel, genre, prodName)
	}

	prodID := payload.ID
	if prodID == "" {
		prodID = "prod_" + uuid.New().String()[:8]
	}

	nowMs := graph.NowMs()

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Insert into master_products
	prodQuery := `
		INSERT INTO master_products (
			id, product_name, genre, sku_code, brand_label, why_we_made_this,
			garment_specifications, composition, mrp, sample_development_id,
			meta_json, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '{}', ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			product_name = excluded.product_name,
			genre = excluded.genre,
			sku_code = excluded.sku_code,
			brand_label = excluded.brand_label,
			why_we_made_this = excluded.why_we_made_this,
			garment_specifications = excluded.garment_specifications,
			composition = excluded.composition,
			mrp = excluded.mrp,
			sample_development_id = excluded.sample_development_id,
			updated_at = excluded.updated_at
	`
	sampleDevID := sql.NullString{String: payload.SampleDevelopmentID, Valid: payload.SampleDevelopmentID != ""}
	if _, err := tx.Exec(prodQuery,
		prodID, prodName, genre, sku, payload.BrandLabel, payload.WhyWeMadeThis,
		payload.GarmentSpecifications, payload.Composition, payload.MRP, sampleDevID,
		nowMs, nowMs,
	); err != nil {
		return nil, fmt.Errorf("failed to upsert master_products: %w", err)
	}

	// 2. Clear and insert Genders
	if _, err := tx.Exec("DELETE FROM master_product_genders WHERE product_id = ?", prodID); err != nil {
		return nil, err
	}
	genderStmt, err := tx.Prepare(`
		INSERT INTO master_product_genders (
			id, product_id, gender, sizes_json, pharma_no, pharma_nos_json,
			measurement_chart_url, product_images_json, color_images_json, sequence
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, err
	}
	defer genderStmt.Close()

	for i, g := range payload.Genders {
		gid := g.ID
		if gid == "" {
			gid = "mpg_" + uuid.New().String()[:8]
		}
		prodImgs := g.ProductImages
		if prodImgs == nil {
			prodImgs = []string{}
		}
		colorImgs := g.ColorImages
		if colorImgs == nil {
			colorImgs = []map[string]any{}
		}
		sizes := g.Sizes
		if sizes == nil {
			sizes = []string{}
		}
		pharmaNos := make([]string, 0, len(g.PharmaNos))
		for _, pn := range g.PharmaNos {
			if clean := strings.TrimSpace(pn); clean != "" {
				pharmaNos = append(pharmaNos, clean)
			}
		}
		pharmaSingle := strings.TrimSpace(g.PharmaNo)
		if len(pharmaNos) == 0 && pharmaSingle != "" {
			pharmaNos = []string{pharmaSingle}
		}
		if pharmaSingle == "" && len(pharmaNos) > 0 {
			pharmaSingle = pharmaNos[0]
		}
		sizesJSON, _ := json.Marshal(sizes)
		pharmaNosJSON, _ := json.Marshal(pharmaNos)
		prodImgsJSON, _ := json.Marshal(prodImgs)
		colorImgsJSON, _ := json.Marshal(colorImgs)
		if _, err := genderStmt.Exec(
			gid, prodID, g.Gender, string(sizesJSON), pharmaSingle, string(pharmaNosJSON),
			g.MeasurementChartURL, string(prodImgsJSON), string(colorImgsJSON), i,
		); err != nil {
			return nil, fmt.Errorf("failed to insert product gender: %w", err)
		}
	}

	// 3. Clear and insert Materials (Fabrics, Trims, Zips)
	if _, err := tx.Exec("DELETE FROM master_product_materials WHERE product_id = ?", prodID); err != nil {
		return nil, err
	}
	matStmt, err := tx.Prepare(`
		INSERT INTO master_product_materials (
			id, product_id, material_type, name, average, gsm, dia, zip_length_qty, sequence
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, err
	}
	defer matStmt.Close()

	for i, m := range payload.Materials {
		if strings.TrimSpace(m.Name) == "" {
			continue
		}
		mid := m.ID
		if mid == "" {
			mid = "mpm_" + uuid.New().String()[:8]
		}
		if _, err := matStmt.Exec(mid, prodID, m.MaterialType, m.Name, m.Average, m.GSM, m.Dia, m.ZipLengthQty, i); err != nil {
			return nil, fmt.Errorf("failed to insert product material: %w", err)
		}
	}

	// 4. Synchronize into master_items so it is immediately active in dropdowns!
	masterItemID := "mi_prod_" + prodID
	metaObj := map[string]any{
		"genre":       genre,
		"sku":         sku,
		"mrp":         payload.MRP,
		"brand_label": payload.BrandLabel,
		"source":      "product_creation",
	}
	metaBytes, _ := json.Marshal(metaObj)

	syncQuery := `
		INSERT INTO master_items (id, category, code, display_name, meta_json, created_at)
		VALUES (?, 'product', ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			code = excluded.code,
			display_name = excluded.display_name,
			meta_json = excluded.meta_json
	`
	if _, err := tx.Exec(syncQuery, masterItemID, sku, prodName, string(metaBytes), nowMs); err != nil {
		return nil, fmt.Errorf("failed to sync product into master_items: %w", err)
	}

	// 5. Graph node & lineage
	if s.graph != nil {
		prodProps := map[string]any{
			"product_name": prodName,
			"genre":        genre,
			"sku_code":     sku,
			"mrp":          payload.MRP,
			"created_at":   nowMs,
		}
		_, _ = s.graph.InsertNode(tx, "PRODUCT_DEFINITION", prodProps, prodID)

		if payload.SampleDevelopmentID != "" {
			_, _ = s.graph.InsertEdge(tx, prodID, payload.SampleDevelopmentID, "DERIVED_FROM_SAMPLE", map[string]any{
				"sku_code":   sku,
				"created_at": nowMs,
			})
		}
	}

	// 6. Domain Outbox Event
	if s.workflow != nil {
		_, _ = s.workflow.EmitEvent(tx, prodID, "product.master.created", map[string]any{
			"productName": prodName,
			"skuCode":     sku,
			"genre":       genre,
		}, 0)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit product creation tx: %w", err)
	}

	return map[string]any{
		"success":     true,
		"id":          prodID,
		"productName": prodName,
		"skuCode":     sku,
		"message":     "Master product created and synchronized with catalog",
	}, nil
}

// GetMasterProduct queries product details with genders and materials
func (s *Service) GetMasterProduct(idOrName string) (*MasterProductDetail, error) {
	var p MasterProductDetail
	query := `
		SELECT
			id, product_name, genre, coalesce(sku_code, ''), coalesce(brand_label, ''),
			coalesce(why_we_made_this, ''), coalesce(garment_specifications, ''),
			coalesce(composition, ''), mrp, coalesce(sample_development_id, ''),
			created_at, updated_at
		FROM master_products
		WHERE id = ? OR product_name = ?
	`
	err := s.db.QueryRow(query, idOrName, idOrName).Scan(
		&p.ID, &p.ProductName, &p.Genre, &p.SKUCode, &p.BrandLabel,
		&p.WhyWeMadeThis, &p.GarmentSpecifications, &p.Composition, &p.MRP,
		&p.SampleDevelopmentID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Genders
	p.Genders = make([]ProductGenderItem, 0)
	gRows, err := s.db.Query(`
		SELECT id, gender, coalesce(sizes_json, '[]'), coalesce(pharma_no, ''), coalesce(pharma_nos_json, '[]'),
		       coalesce(measurement_chart_url, ''), coalesce(product_images_json, '[]'),
		       coalesce(color_images_json, '[]'), sequence
		FROM master_product_genders
		WHERE product_id = ?
		ORDER BY sequence
	`, p.ID)
	if err == nil {
		defer gRows.Close()
		for gRows.Next() {
			var g ProductGenderItem
			var sizesJSON, pharmaNosJSON, prodImgsJSON, colorImgsJSON string
			if err := gRows.Scan(&g.ID, &g.Gender, &sizesJSON, &g.PharmaNo, &pharmaNosJSON, &g.MeasurementChartURL, &prodImgsJSON, &colorImgsJSON, &g.Sequence); err == nil {
				g.Sizes = make([]string, 0)
				g.PharmaNos = make([]string, 0)
				g.ProductImages = make([]string, 0)
				g.ColorImages = make([]map[string]any, 0)
				_ = json.Unmarshal([]byte(sizesJSON), &g.Sizes)
				_ = json.Unmarshal([]byte(pharmaNosJSON), &g.PharmaNos)
				_ = json.Unmarshal([]byte(prodImgsJSON), &g.ProductImages)
				_ = json.Unmarshal([]byte(colorImgsJSON), &g.ColorImages)
				if len(g.PharmaNos) == 0 && strings.TrimSpace(g.PharmaNo) != "" {
					g.PharmaNos = []string{strings.TrimSpace(g.PharmaNo)}
				}
				p.Genders = append(p.Genders, g)
			}
		}
	}

	// Materials
	p.Materials = make([]ProductMaterialItem, 0)
	mRows, err := s.db.Query(`
		SELECT id, material_type, name, average, coalesce(gsm, 0), coalesce(dia, 0), coalesce(zip_length_qty, ''), sequence
		FROM master_product_materials
		WHERE product_id = ?
		ORDER BY sequence
	`, p.ID)
	if err == nil {
		defer mRows.Close()
		for mRows.Next() {
			var m ProductMaterialItem
			if err := mRows.Scan(&m.ID, &m.MaterialType, &m.Name, &m.Average, &m.GSM, &m.Dia, &m.ZipLengthQty, &m.Sequence); err == nil {
				p.Materials = append(p.Materials, m)
			}
		}
	}

	return &p, nil
}

// ListMasterProducts queries list of created master products
func (s *Service) ListMasterProducts(limit, offset int) ([]MasterProductDetail, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT
			id, product_name, genre, coalesce(sku_code, ''), coalesce(brand_label, ''),
			coalesce(why_we_made_this, ''), coalesce(garment_specifications, ''),
			coalesce(composition, ''), mrp, coalesce(sample_development_id, ''),
			created_at, updated_at
		FROM master_products
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]MasterProductDetail, 0)
	for rows.Next() {
		var p MasterProductDetail
		if err := rows.Scan(
			&p.ID, &p.ProductName, &p.Genre, &p.SKUCode, &p.BrandLabel,
			&p.WhyWeMadeThis, &p.GarmentSpecifications, &p.Composition, &p.MRP,
			&p.SampleDevelopmentID, &p.CreatedAt, &p.UpdatedAt,
		); err == nil {
			list = append(list, p)
		}
	}
	return list, nil
}

// -------------------------------------------------------------
// METHODS: CLIENT SPECIFIC PRICING
// -------------------------------------------------------------

// SetClientProductPricing creates or updates negotiated pricing for a client
func (s *Service) SetClientProductPricing(payload ClientPricingPayload) (map[string]any, error) {
	if strings.TrimSpace(payload.ClientID) == "" || strings.TrimSpace(payload.ProductID) == "" {
		return nil, fmt.Errorf("clientId and productId are required")
	}

	id := payload.ID
	if id == "" {
		id = "cpp_" + uuid.New().String()[:8]
	}

	effDate := strings.TrimSpace(payload.EffectiveDate)
	if effDate == "" {
		effDate = time.Now().Format("2006-01-02")
	}

	nowMs := graph.NowMs()

	query := `
		INSERT INTO client_product_pricing (
			id, client_id, client_name, product_id, product_name,
			client_image_url, contracted_price, comments, effective_date,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			client_name = excluded.client_name,
			product_name = excluded.product_name,
			client_image_url = excluded.client_image_url,
			contracted_price = excluded.contracted_price,
			comments = excluded.comments,
			effective_date = excluded.effective_date,
			updated_at = excluded.updated_at
	`
	if _, err := s.db.Exec(query,
		id, payload.ClientID, payload.ClientName, payload.ProductID, payload.ProductName,
		payload.ClientImageURL, payload.ContractedPrice, payload.Comments, effDate,
		nowMs, nowMs,
	); err != nil {
		return nil, fmt.Errorf("failed to save client pricing: %w", err)
	}

	// Graph link
	if s.graph != nil {
		_, _ = s.graph.InsertEdge(nil, payload.ClientID, payload.ProductID, "HAS_CLIENT_PRICING", map[string]any{
			"price":          payload.ContractedPrice,
			"effective_date": effDate,
			"updated_at":     nowMs,
		})
	}

	return map[string]any{
		"success": true,
		"id":      id,
		"message": "Client pricing recorded successfully",
	}, nil
}

// ListClientProductPricing retrieves client-specific prices
func (s *Service) ListClientProductPricing(clientID string) ([]ClientPricingDetail, error) {
	query := `
		SELECT
			id, client_id, client_name, product_id, product_name,
			coalesce(client_image_url, ''), contracted_price, coalesce(comments, ''),
			effective_date, created_at, updated_at
		FROM client_product_pricing
	`
	var rows *sql.Rows
	var err error
	if clientID != "" {
		query += " WHERE client_id = ? ORDER BY effective_date DESC"
		rows, err = s.db.Query(query, clientID)
	} else {
		query += " ORDER BY effective_date DESC"
		rows, err = s.db.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]ClientPricingDetail, 0)
	for rows.Next() {
		var c ClientPricingDetail
		if err := rows.Scan(
			&c.ID, &c.ClientID, &c.ClientName, &c.ProductID, &c.ProductName,
			&c.ClientImageURL, &c.ContractedPrice, &c.Comments, &c.EffectiveDate,
			&c.CreatedAt, &c.UpdatedAt,
		); err == nil {
			list = append(list, c)
		}
	}
	return list, nil
}
