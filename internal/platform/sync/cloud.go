package sync

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ppl-amehta/selvedge/internal/graph"
)

// CloudSyncConfig configures the edge-to-cloud HQ replication bridge
type CloudSyncConfig struct {
	HQEndpoint   string        // URL for Central Cloud Postgres / Turso gateway
	APIKey       string        // Factory device authentication token
	BatchSize    int           // Max events per sync batch
	SyncInterval time.Duration // Interval between sync sweeps
}

type CloudSyncer struct {
	db     *sql.DB
	cfg    CloudSyncConfig
	client *http.Client
}

func NewCloudSyncer(db *sql.DB, cfg CloudSyncConfig) *CloudSyncer {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
	}
	if cfg.SyncInterval <= 0 {
		cfg.SyncInterval = 10 * time.Second
	}

	return &CloudSyncer{
		db:  db,
		cfg: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Start begins the background edge-to-cloud outbox synchronization loop
func (s *CloudSyncer) Start(ctx context.Context) {
	if s.cfg.HQEndpoint == "" {
		log.Println("Notice: Cloud HQ sync disabled (HQ_ENDPOINT not configured)")
		return
	}

	log.Printf("✓ Cloud HQ Syncer active: streaming events to %s", s.cfg.HQEndpoint)

	ticker := time.NewTicker(s.cfg.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.syncBatch(); err != nil {
				log.Printf("Cloud HQ sync warning: %v (queued locally in SQLite WAL)", err)
			}
		}
	}
}

func (s *CloudSyncer) syncBatch() error {
	// Query events marked as dispatched locally but pending cloud HQ forward
	rows, err := s.db.Query(`
		SELECT id, aggregate_id, event_type, payload_json, occurred_at
		FROM domain_events
		WHERE dispatched_at IS NOT NULL
		ORDER BY occurred_at ASC
		LIMIT ?
	`, s.cfg.BatchSize)
	if err != nil {
		return err
	}
	defer rows.Close()

	type EventItem struct {
		ID          string         `json:"id"`
		AggregateID string         `json:"aggregate_id"`
		EventType   string         `json:"event_type"`
		Payload     map[string]any `json:"payload"`
		OccurredAt  int64          `json:"occurred_at"`
	}

	var batch []EventItem
	for rows.Next() {
		var item EventItem
		var payloadStr string
		if err := rows.Scan(&item.ID, &item.AggregateID, &item.EventType, &payloadStr, &item.OccurredAt); err != nil {
			return err
		}
		props, _ := graph.DecodeProps(payloadStr)
		item.Payload = props
		batch = append(batch, item)
	}

	if len(batch) == 0 {
		return nil
	}

	bodyBytes, err := json.Marshal(map[string]any{
		"factory_id": "edge-plant-01",
		"timestamp":  graph.NowMs(),
		"events":     batch,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", s.cfg.HQEndpoint+"/api/sync/events", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("hq gateway unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("hq gateway responded with HTTP %d", resp.StatusCode)
	}

	log.Printf("☁️ Synced %d domain event(s) to Central Cloud HQ", len(batch))
	return nil
}
