package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/ppl-amehta/selvedge/internal/erp/inventory"
	"github.com/ppl-amehta/selvedge/internal/erp/masters"
	"github.com/ppl-amehta/selvedge/internal/erp/orders"
	"github.com/ppl-amehta/selvedge/internal/erp/planning"
	"github.com/ppl-amehta/selvedge/internal/graph"
	"github.com/ppl-amehta/selvedge/internal/workflow"
)

type AppServices struct {
	Graph     *graph.Store
	Workflow  *workflow.Engine
	Masters   *masters.Service
	Orders    *orders.Service
	Inventory *inventory.Service
	Planning  *planning.Service
	WebFS     fs.FS
	WebDir    string
}

func NewRouter(svc AppServices) http.Handler {
	r := chi.NewRouter()

	setupMiddleware(r)

	// 1. Health
	r.Get("/api/health", func(w http.ResponseWriter, req *http.Request) {
		renderJSON(w, http.StatusOK, map[string]any{
			"status":    "ok",
			"system":    "FPCC Selvedge Enterprise Edge Engine",
			"version":   "1.0.0-greenfield",
			"arch":      "Unified Graph + Bi-Temporal Outbox (Go 1.27)",
			"timestamp": graph.NowMs(),
		})
	})

	// 2. Masters Endpoints
	r.Route("/api/masters", func(r chi.Router) {
		r.Get("/all", func(w http.ResponseWriter, req *http.Request) {
			data, err := svc.Masters.GetAllMasters()
			if err != nil {
				renderError(w, http.StatusInternalServerError, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": data})
		})

		r.Post("/register", func(w http.ResponseWriter, req *http.Request) {
			var body struct {
				Type  string         `json:"type"`
				Value string         `json:"value"`
				Meta  map[string]any `json:"meta"`
			}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				renderError(w, http.StatusBadRequest, "Invalid JSON payload")
				return
			}
			res, err := svc.Masters.RegisterMaster(body.Type, body.Value, body.Meta)
			if err != nil {
				renderError(w, http.StatusInternalServerError, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, res)
		})
	})

	// 3. Orders Endpoints
	r.Route("/api/orders", func(r chi.Router) {
		r.Get("/next-order-no", func(w http.ResponseWriter, req *http.Request) {
			no, err := svc.Orders.SuggestNextOrderNo()
			if err != nil {
				renderError(w, http.StatusInternalServerError, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{
				"success":          true,
				"suggestedOrderNo": no,
				"nextOrderNo":      no,
			})
		})

		r.Get("/zoho/fetch", func(w http.ResponseWriter, req *http.Request) {
			soNo := req.URL.Query().Get("salesOrderNo")
			if soNo == "" {
				soNo = req.URL.Query().Get("so")
			}
			if soNo == "" {
				renderError(w, http.StatusBadRequest, "Missing salesOrderNo query parameter")
				return
			}
			data, err := svc.Orders.FetchZohoOrder(soNo)
			if err != nil {
				renderError(w, http.StatusBadRequest, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": data})
		})

		r.Post("/", func(w http.ResponseWriter, req *http.Request) {
			var payload orders.CreateOrderPayload
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				renderError(w, http.StatusBadRequest, "Invalid JSON payload")
				return
			}
			res, err := svc.Orders.CreateOrder(payload)
			if err != nil {
				renderError(w, http.StatusBadRequest, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, res)
		})

		r.Get("/{orderNo}", func(w http.ResponseWriter, req *http.Request) {
			orderNo := chi.URLParam(req, "orderNo")
			ord, err := svc.Orders.GetOrder(orderNo)
			if err != nil {
				renderError(w, http.StatusNotFound, "Order not found: "+err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": ord})
		})

		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			list, err := svc.Orders.ListOrders(50, 0)
			if err != nil {
				renderError(w, http.StatusInternalServerError, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": list})
		})
	})

	// 4. Inventory Endpoints
	r.Route("/api/inventory", func(r chi.Router) {
		suggestParentCodeHandler := func(w http.ResponseWriter, req *http.Request) {
			code, err := svc.Inventory.SuggestNextParentCode()
			if err != nil {
				renderError(w, http.StatusInternalServerError, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{
				"success":        true,
				"suggestedCode":  code,
				"nextParentCode": code,
			})
		}

		r.Get("/next-parent-code", suggestParentCodeHandler)
		r.Get("/suggest-parent-code", suggestParentCodeHandler)

		r.Post("/inward", func(w http.ResponseWriter, req *http.Request) {
			var payload inventory.InwardParentLotPayload
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				renderError(w, http.StatusBadRequest, "Invalid JSON payload")
				return
			}
			res, err := svc.Inventory.InwardParentLot(payload)
			if err != nil {
				renderError(w, http.StatusBadRequest, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, res)
		})

		// --- 3-Stage Inward Pipeline Endpoints ---

		// Station 1A: Gate Entry & Inward Booking
		r.Post("/gate-inward", func(w http.ResponseWriter, req *http.Request) {
			var payload inventory.GateInwardPayload
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				renderError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
				return
			}
			res, err := svc.Inventory.GateInward(payload)
			if err != nil {
				renderError(w, http.StatusBadRequest, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, res)
		})

		// Station 1B: Floor Count & Discrepancy Reconciliation
		r.Post("/floor-count", func(w http.ResponseWriter, req *http.Request) {
			var payload inventory.FloorCountPayload
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				renderError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
				return
			}
			res, err := svc.Inventory.RecordFloorCount(payload)
			if err != nil {
				renderError(w, http.StatusBadRequest, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, res)
		})

		// Station 1C: Bundle QC Inspection & Tagging
		r.Post("/qc-inspect", func(w http.ResponseWriter, req *http.Request) {
			var payload inventory.BundleQCPayload
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				renderError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
				return
			}
			res, err := svc.Inventory.RecordBundleQC(payload)
			if err != nil {
				renderError(w, http.StatusBadRequest, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, res)
		})

		// Discrepancy Resolution
		r.Post("/resolve-discrepancy", func(w http.ResponseWriter, req *http.Request) {
			var body struct {
				ParentCode string `json:"parentCode"`
				Action     string `json:"action"`
				Remarks    string `json:"remarks"`
			}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				renderError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
				return
			}
			res, err := svc.Inventory.ResolveDiscrepancy(body.ParentCode, body.Action, body.Remarks)
			if err != nil {
				renderError(w, http.StatusBadRequest, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, res)
		})

		// Lots Listing with Stage Filter
		r.Get("/lots", func(w http.ResponseWriter, req *http.Request) {
			stage := req.URL.Query().Get("stage")
			lots, err := svc.Inventory.ListLots(stage)
			if err != nil {
				renderError(w, http.StatusInternalServerError, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": lots})
		})

		// Lot Detail & Bundle Breakdown
		r.Get("/lots/{parentCode}", func(w http.ResponseWriter, req *http.Request) {
			parentCode := chi.URLParam(req, "parentCode")
			detail, err := svc.Inventory.GetLot(parentCode)
			if err != nil {
				renderError(w, http.StatusNotFound, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": detail})
		})

		r.Post("/return-remnant", func(w http.ResponseWriter, req *http.Request) {
			var body struct {
				BundleCode       string  `json:"bundleCode"`
				ReturnType       string  `json:"returnType"`
				ReturnedWeightKg float64 `json:"returnedWeightKg"`
				ReturnLocation   string  `json:"returnLocation"`
				Remarks          string  `json:"remarks"`
			}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				renderError(w, http.StatusBadRequest, "Invalid JSON payload")
				return
			}
			res, err := svc.Inventory.ReturnRemnant(body.BundleCode, body.ReturnType, body.ReturnedWeightKg, body.ReturnLocation, body.Remarks)
			if err != nil {
				renderError(w, http.StatusBadRequest, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": res})
		})

		r.Get("/discrepancies", func(w http.ResponseWriter, req *http.Request) {
			list, err := svc.Inventory.ListDiscrepancies()
			if err != nil {
				renderError(w, http.StatusInternalServerError, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": list})
		})
	})

	// 5. Order Preparation & Planning Endpoints (Station 2 & 3)
	r.Route("/api/planning", func(r chi.Router) {
		r.Get("/orders", func(w http.ResponseWriter, req *http.Request) {
			list, err := svc.Planning.ListOrdersForPrep()
			if err != nil {
				renderError(w, http.StatusInternalServerError, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": list})
		})

		r.Get("/orders/{orderId}", func(w http.ResponseWriter, req *http.Request) {
			orderID := chi.URLParam(req, "orderId")
			data, err := svc.Planning.GetOrderPrepAndPlanning(orderID)
			if err != nil {
				renderError(w, http.StatusNotFound, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": data})
		})

		r.Post("/pharma", func(w http.ResponseWriter, req *http.Request) {
			var payload planning.PharmaPayload
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				renderError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
				return
			}
			res, err := svc.Planning.SavePharma(payload.OrderItemID, payload)
			if err != nil {
				renderError(w, http.StatusBadRequest, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": res})
		})

		r.Post("/bom", func(w http.ResponseWriter, req *http.Request) {
			var payload planning.BomLinePayload
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				renderError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
				return
			}
			res, err := svc.Planning.SaveBomLine(payload.OrderItemID, payload)
			if err != nil {
				renderError(w, http.StatusBadRequest, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": res})
		})

		r.Delete("/bom/{lineId}", func(w http.ResponseWriter, req *http.Request) {
			lineID := chi.URLParam(req, "lineId")
			if err := svc.Planning.DeleteBomLine(lineID); err != nil {
				renderError(w, http.StatusInternalServerError, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "deleted": lineID})
		})

		r.Post("/raise-po", func(w http.ResponseWriter, req *http.Request) {
			var payload planning.RaisePOPayload
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				renderError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
				return
			}
			res, err := svc.Planning.RaiseProcurementRequisition(payload.OrderID, payload)
			if err != nil {
				renderError(w, http.StatusBadRequest, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": res})
		})

		r.Post("/routing-steps", func(w http.ResponseWriter, req *http.Request) {
			var body struct {
				OrderID string                    `json:"orderId"`
				Steps   []planning.RoutingStepPayload `json:"steps"`
			}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				renderError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
				return
			}
			res, err := svc.Planning.SaveRoutingSteps(body.OrderID, body.Steps)
			if err != nil {
				renderError(w, http.StatusBadRequest, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": res})
		})

		r.Post("/advance-state", func(w http.ResponseWriter, req *http.Request) {
			var body struct {
				OrderID     string `json:"orderId"`
				TargetState string `json:"targetState"`
				Reason      string `json:"reason"`
			}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				renderError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
				return
			}
			res, err := svc.Planning.AdvanceOrderState(body.OrderID, body.TargetState, body.Reason)
			if err != nil {
				renderError(w, http.StatusBadRequest, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": res})
		})
	})

	// 6. Unified Graph Query Endpoints
	r.Route("/api/graph", func(r chi.Router) {
		r.Get("/nodes/{nodeId}", func(w http.ResponseWriter, req *http.Request) {
			nodeId := chi.URLParam(req, "nodeId")
			node, err := svc.Graph.GetNode(nodeId)
			if err != nil {
				renderError(w, http.StatusNotFound, "Node not found")
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": node})
		})

		r.Get("/traverse", func(w http.ResponseWriter, req *http.Request) {
			startNode := req.URL.Query().Get("start")
			edgeType := req.URL.Query().Get("edge")
			dir := req.URL.Query().Get("direction")
			if dir == "" {
				dir = "OUT"
			}
			nodes, err := svc.Graph.Traverse(startNode, edgeType, dir)
			if err != nil {
				renderError(w, http.StatusInternalServerError, err.Error())
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": nodes})
		})

		r.Get("/as-of", func(w http.ResponseWriter, req *http.Request) {
			nodeId := req.URL.Query().Get("nodeId")
			tsStr := req.URL.Query().Get("timestamp")
			ts, _ := strconv.ParseInt(tsStr, 10, 64)
			node, err := svc.Graph.AsOfNode(nodeId, ts)
			if err != nil {
				renderError(w, http.StatusNotFound, "No version found at timestamp")
				return
			}
			renderJSON(w, http.StatusOK, map[string]any{"success": true, "data": node})
		})
	})

	// 6. Workflow Dispatch Endpoint
	r.Post("/api/workflows/dispatch", func(w http.ResponseWriter, req *http.Request) {
		count, err := svc.Workflow.ProcessPendingEvents()
		if err != nil {
			renderError(w, http.StatusInternalServerError, err.Error())
			return
		}
		renderJSON(w, http.StatusOK, map[string]any{"success": true, "dispatchedCount": count})
	})

	// 7. File Upload Endpoint (Images, Artwork, Specs)
	r.Post("/api/upload", func(w http.ResponseWriter, req *http.Request) {
		if err := req.ParseMultipartForm(32 << 20); err != nil {
			renderError(w, http.StatusBadRequest, "File too large or invalid multipart form")
			return
		}

		file, handler, err := req.FormFile("file")
		if err != nil {
			renderError(w, http.StatusBadRequest, "Missing 'file' field in multipart form")
			return
		}
		defer file.Close()

		uploadDir := "./data/uploads"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			renderError(w, http.StatusInternalServerError, "Failed to create uploads directory")
			return
		}

		cleanBase := strings.ReplaceAll(filepath.Base(handler.Filename), " ", "_")
		cleanFilename := fmt.Sprintf("%d_%s", time.Now().UnixNano()/1e6, cleanBase)
		dstPath := filepath.Join(uploadDir, cleanFilename)

		dst, err := os.Create(dstPath)
		if err != nil {
			renderError(w, http.StatusInternalServerError, "Failed to create destination file: "+err.Error())
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			renderError(w, http.StatusInternalServerError, "Failed to write file contents: "+err.Error())
			return
		}

		fileURL := "/data/uploads/" + cleanFilename
		renderJSON(w, http.StatusOK, map[string]any{
			"success":  true,
			"url":      fileURL,
			"filename": cleanFilename,
			"size":     handler.Size,
		})
	})

	// Serve Static Uploads
	_ = os.MkdirAll("./data/uploads", 0755)
	uploadsFS := http.Dir("./data/uploads")
	r.Handle("/data/uploads/*", http.StripPrefix("/data/uploads/", http.FileServer(uploadsFS)))

	// 8. Svelte 5 Frontend SPA Serving (Embedded in binary or disk fallback)
	if svc.WebFS != nil {
		indexHTML, _ := fs.ReadFile(svc.WebFS, "index.html")
		fileServer := http.FileServerFS(svc.WebFS)

		r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			p := strings.TrimPrefix(path.Clean(req.URL.Path), "/")
			if p == "" || p == "index.html" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("Cache-Control", "no-cache")
				http.ServeContent(w, req, "index.html", time.Time{}, bytes.NewReader(indexHTML))
				return
			}
			if _, err := fs.Stat(svc.WebFS, p); err != nil {
				// Unknown route -> SPA fallback
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("Cache-Control", "no-cache")
				http.ServeContent(w, req, "index.html", time.Time{}, bytes.NewReader(indexHTML))
				return
			}
			if strings.HasPrefix(p, "assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			fileServer.ServeHTTP(w, req)
		}))
	} else if svc.WebDir != "" {
		if _, err := os.Stat(svc.WebDir); err == nil {
			fileServer := http.FileServer(http.Dir(svc.WebDir))
			r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				path := svc.WebDir + req.URL.Path
				if _, err := os.Stat(path); os.IsNotExist(err) {
					http.ServeFile(w, req, svc.WebDir+"/index.html")
					return
				}
				fileServer.ServeHTTP(w, req)
			}))
		}
	}

	return r
}

func setupMiddleware(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

func renderJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func renderError(w http.ResponseWriter, status int, message string) {
	renderJSON(w, status, map[string]any{
		"success": false,
		"error":   message,
	})
}
