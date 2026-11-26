package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ppl-amehta/selvedge/internal/erp/cutting"
	"github.com/ppl-amehta/selvedge/internal/erp/inventory"
	"github.com/ppl-amehta/selvedge/internal/erp/masters"
	"github.com/ppl-amehta/selvedge/internal/erp/orders"
	"github.com/ppl-amehta/selvedge/internal/erp/planning"
	"github.com/ppl-amehta/selvedge/internal/erp/products"
	"github.com/ppl-amehta/selvedge/internal/graph"
	"github.com/ppl-amehta/selvedge/internal/platform/db"
	"github.com/ppl-amehta/selvedge/internal/platform/router"
	"github.com/ppl-amehta/selvedge/internal/platform/storage"
	"github.com/ppl-amehta/selvedge/internal/workflow"
	"github.com/ppl-amehta/selvedge/migrations"
	"github.com/ppl-amehta/selvedge/web"
)

func main() {
	port := flag.Int("port", 4000, "HTTP server listening port")
	dbPath := flag.String("db", "./data/selvedge.db", "SQLite database path")
	webDir := flag.String("web", "./web/public", "Static web files directory")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		fmt.Sscanf(envPort, "%d", port)
	}

	isServerless := os.Getenv("VERCEL") != "" || os.Getenv("VERCEL_ENV") != ""

	// 0. Initialize Modern Structured slog Handler
	var handler slog.Handler
	if os.Getenv("LOG_FORMAT") == "json" || os.Getenv("FLY_MACHINE_ID") != "" || isServerless {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("starting Selvedge enterprise engine",
		"version", "1.0.0-production",
		"go_version", runtime.Version(),
		"architecture", "Go 1.27 + SQLite WAL + Svelte 5",
	)

	// 1. Database Connection
	database, err := db.Connect(*dbPath)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// 2. Run Database Migrations via Embedded FS (Dynamically Discovered & Lexically Sorted)
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		slog.Error("failed to discover embedded migrations", "error", err)
		os.Exit(1)
	}

	migrationNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrationNames = append(migrationNames, entry.Name())
		}
	}
	sort.Strings(migrationNames)

	appliedMigrations, err := database.ListAppliedMigrations()
	if err != nil {
		slog.Error("failed to inspect applied database migrations", "error", err)
		os.Exit(1)
	}

	for _, mFile := range migrationNames {
		canonicalKey := "migrations/" + filepath.Base(mFile)
		if appliedMigrations[canonicalKey] || appliedMigrations[filepath.Base(mFile)] {
			continue
		}
		content, err := migrations.FS.ReadFile(mFile)
		if err != nil {
			slog.Error("embedded migration file read error", "file", mFile, "error", err)
			os.Exit(1)
		}
		if err := database.RunMigrations(canonicalKey, string(content)); err != nil {
			slog.Error("database migration failed", "file", mFile, "error", err)
			os.Exit(1)
		}
	}

	// 3. Initialize Core Subsystems
	graphStore := graph.NewStore(database.DB)
	workflowEngine := workflow.NewEngine(database.DB)

	if err := workflowEngine.SeedDefaultRules(); err != nil {
		slog.Warn("failed to seed workflow rules", "error", err)
	}

	// 4. Initialize Domain ERP Services
	mastersService := masters.NewService(database.DB, graphStore)
	ordersService := orders.NewService(database.DB, graphStore, workflowEngine)
	inventoryService := inventory.NewService(database.DB, graphStore, workflowEngine)
	planningService := planning.NewService(database.DB, graphStore, workflowEngine)
	cuttingService := cutting.NewService(database.DB, graphStore, workflowEngine)
	productsService := products.NewService(database.DB, graphStore, workflowEngine)

	// 5. Start Background Outbox Poller (disabled on scale-to-zero serverless unless opted in;
	// all write services already trigger inline ProcessPendingEvents after commit)
	pollerCtx, cancelPoller := context.WithCancel(context.Background())
	var pollerWG sync.WaitGroup
	if !isServerless || os.Getenv("ENABLE_OUTBOX_POLLER") == "1" {
		pollerWG.Add(1)
		go func() {
			defer pollerWG.Done()
			ticker := time.NewTicker(3 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-pollerCtx.Done():
					return
				case <-ticker.C:
					count, err := workflowEngine.ProcessPendingEvents()
					if err != nil {
						slog.Error("outbox poller error", "error", err)
					} else if count > 0 {
						slog.Info("dispatched background domain events", "count", count)
					}
				}
			}
		}()
	} else {
		slog.Info("background outbox ticker disabled in serverless mode (inline dispatch active)")
	}

	// 6. Router Setup with Svelte 5 Embedded SPA & Storage Adapter
	svelteFS, err := web.GetFS()
	if err != nil {
		slog.Warn("embedded Svelte dist not found", "error", err)
	}

	dbMode := "sqlite"
	if !database.IsSQLite {
		dbMode = "turso"
	}
	uploadStore := storage.NewFromEnv("./data/uploads")
	slog.Info("configured storage backend", "storage", uploadStore.Mode(), "db_mode", dbMode)

	absWebDir, _ := filepath.Abs(*webDir)
	r := router.NewRouter(router.AppServices{
		DB:          database.DB,
		Graph:       graphStore,
		Workflow:    workflowEngine,
		Masters:     mastersService,
		Orders:      ordersService,
		Inventory:   inventoryService,
		Planning:    planningService,
		Cutting:     cuttingService,
		Products:    productsService,
		UploadStore: uploadStore,
		DBMode:      dbMode,
		WebFS:       svelteFS,
		WebDir:      absWebDir,
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      r,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 7. Graceful Server Startup & Shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	serverErrChan := make(chan error, 1)
	go func() {
		slog.Info("http server listening", "port", *port, "url", fmt.Sprintf("http://localhost:%d", *port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrChan <- err
		}
	}()

	select {
	case <-stopChan:
		slog.Info("initiating graceful shutdown signal...")
	case err := <-serverErrChan:
		slog.Error("http server crashed unexpectedly", "error", err)
	}

	// First stop accepting new HTTP requests
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	// Next, stop outbox poller and wait for active ticks to finish
	cancelPoller()
	pollerWG.Wait()
	slog.Info("outbox poller terminated cleanly")

	// Flush SQLite WAL safely with no concurrent writers
	if err := database.CheckpointWAL(shutdownCtx); err != nil {
		slog.Warn("wal shutdown checkpoint warning", "error", err)
	}

	slog.Info("Selvedge server terminated cleanly")
}
