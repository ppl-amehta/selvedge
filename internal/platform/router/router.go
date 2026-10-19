package router

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/ppl-amehta/selvedge/internal/erp/inventory"
	"github.com/ppl-amehta/selvedge/internal/erp/masters"
	"github.com/ppl-amehta/selvedge/internal/erp/orders"
	"github.com/ppl-amehta/selvedge/internal/graph"
	"github.com/ppl-amehta/selvedge/internal/workflow"
)

type AppServices struct {
	Graph     *graph.Store
	Workflow  *workflow.Engine
	Masters   *masters.Service
	Orders    *orders.Service
	Inventory *inventory.Service
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
		r.Get("/next-parent-code", func(w http.ResponseWriter, req *http.Request) {
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
		})

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

	// 5. Unified Graph Query Endpoints
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

	// 7. Static Web Serving for the 2 Screens (Order Taking & Inward)
	if svc.WebDir != "" {
		if _, err := os.Stat(svc.WebDir); err == nil {
			fileServer := http.FileServer(http.Dir(svc.WebDir))
			r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				// If file doesn't exist, serve index.html for SPA
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
