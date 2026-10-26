package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ppl-amehta/selvedge/internal/erp/inventory"
	"github.com/ppl-amehta/selvedge/internal/erp/masters"
	"github.com/ppl-amehta/selvedge/internal/erp/orders"
	"github.com/ppl-amehta/selvedge/internal/graph"
	"github.com/ppl-amehta/selvedge/internal/platform/db"
	"github.com/ppl-amehta/selvedge/internal/platform/router"
	"github.com/ppl-amehta/selvedge/internal/workflow"
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

	log.Printf("==================================================")
	log.Printf(" FPCC SELVEDGE ENTERPRISE EDGE ENGINE")
	log.Printf(" Unified Graph System of Record & Smart Workflows")
	log.Printf(" Architecture: Go 1.27 + SQLite WAL + Chi SPA")
	log.Printf("==================================================")

	// 1. Database Connection
	database, err := db.Connect(*dbPath)
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to database: %v", err)
	}
	defer database.Close()

	// 2. Run Database Migrations
	migrationFiles := []string{
		"migrations/001_initial_graph.sql",
		"migrations/002_seed_masters.sql",
		"migrations/003_multi_product_orders.sql",
		"migrations/004_three_stage_inward_pipeline.sql",
	}

	for _, mFile := range migrationFiles {
		if content, err := os.ReadFile(mFile); err == nil {
			log.Printf("Applying migration: %s ...", mFile)
			if err := database.RunMigrations(mFile, string(content)); err != nil {
				log.Fatalf("FATAL: Migration failed on %s: %v", mFile, err)
			}
		} else {
			log.Printf("Notice: Migration file not found (skipping): %s", mFile)
		}
	}

	// 3. Initialize Core Subsystems
	graphStore := graph.NewStore(database.DB)
	workflowEngine := workflow.NewEngine(database.DB)

	if err := workflowEngine.SeedDefaultRules(); err != nil {
		log.Printf("Warning: Failed to seed workflow rules: %v", err)
	}

	// 4. Initialize Domain ERP Services
	mastersService := masters.NewService(database.DB, graphStore)
	ordersService := orders.NewService(database.DB, graphStore, workflowEngine)
	inventoryService := inventory.NewService(database.DB, graphStore, workflowEngine)

	// 5. Start Background Outbox Poller
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				count, err := workflowEngine.ProcessPendingEvents()
				if err != nil {
					log.Printf("Outbox poller error: %v", err)
				} else if count > 0 {
					log.Printf("⚡ Dispatched %d background domain event(s) through workflow rules", count)
				}
			}
		}
	}()

	// 6. Router Setup with Svelte 5 Embedded SPA
	svelteFS, err := web.GetFS()
	if err != nil {
		log.Printf("Notice: Embedded Svelte dist not found: %v", err)
	}

	absWebDir, _ := filepath.Abs(*webDir)
	r := router.NewRouter(router.AppServices{
		Graph:     graphStore,
		Workflow:  workflowEngine,
		Masters:   mastersService,
		Orders:    ordersService,
		Inventory: inventoryService,
		WebFS:     svelteFS,
		WebDir:    absWebDir,
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 7. Graceful Server Startup & Shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("✓ Server listening at: http://localhost:%d", *port)
		log.Printf("✓ Station 0 (Order Taking) & Station 1 (Inward) active on port %d", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("FATAL: HTTP server crashed: %v", err)
		}
	}()

	<-stopChan
	log.Println("Shutting down Selvedge gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Selvedge server terminated cleanly.")
}
