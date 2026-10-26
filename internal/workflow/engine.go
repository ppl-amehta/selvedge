package workflow

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/ppl-amehta/selvedge/internal/graph"
)

// Condition defines a declarative predicate against event/node payloads
type Condition struct {
	Field string `json:"field"` // e.g. "discrepancy_qty", "status", "variance_pct"
	Op    string `json:"op"`    // ">", ">=", "==", "!=", "<", "<="
	Val   any    `json:"val"`   // comparison target
}

// Action defines an automated or human-gated task triggered by a rule
type Action struct {
	ActionType string         `json:"action_type"` // e.g. "DRAFT_DEBIT_NOTE", "HOLD_PAYMENT", "NOTIFY_ROLE", "STATUS_UPDATE"
	Params     map[string]any `json:"params"`
}

// Rule defines a Rippling-style automation triggered by graph events
type Rule struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	TriggerEvent string      `json:"trigger_event"` // e.g. "INWARD_LOT_CREATED", "ORDER_BOOKED"
	Condition    *Condition  `json:"condition"`
	Actions      []Action    `json:"actions"`
	Enabled      bool        `json:"enabled"`
}

type Engine struct {
	db *sql.DB
}

func NewEngine(db *sql.DB) *Engine {
	return &Engine{db: db}
}

// EmitEvent atomically inserts an event into the transactional outbox
func (e *Engine) EmitEvent(tx *sql.Tx, aggregateID, eventType string, payload map[string]any, depth int) (string, error) {
	if depth > 10 {
		return "", fmt.Errorf("workflow execution exceeded maximum depth guard (10)")
	}

	eventID := fmt.Sprintf("evt_%s", uuid.New().String()[:12])
	payloadStr, err := graph.EncodeProps(payload)
	if err != nil {
		return "", err
	}

	now := graph.NowMs()
	query := `
		INSERT INTO domain_events (id, aggregate_id, event_type, payload_json, depth, occurred_at, dispatched_at)
		VALUES (?, ?, ?, ?, ?, ?, NULL)
	`
	var execer interface {
		Exec(query string, args ...any) (sql.Result, error)
	} = e.db
	if tx != nil {
		execer = tx
	}

	if _, err := execer.Exec(query, eventID, aggregateID, eventType, payloadStr, depth, now); err != nil {
		return "", fmt.Errorf("failed to emit domain event: %w", err)
	}

	return eventID, nil
}

// SeedDefaultRules installs the factory automation rules into workflow_rules
func (e *Engine) SeedDefaultRules() error {
	rules := []Rule{
		{
			ID:           "rule-shortage-hold",
			Name:         "Inward Shortage Debit Note & Payment Freeze",
			TriggerEvent: "INWARD_LOT_CREATED",
			Condition: &Condition{
				Field: "shortfall_qty",
				Op:    ">",
				Val:   0.0,
			},
			Actions: []Action{
				{
					ActionType: "DRAFT_DEBIT_NOTE",
					Params: map[string]any{
						"reason": "Physical received quantity less than invoice",
						"status": "DRAFT",
					},
				},
				{
					ActionType: "HOLD_PAYMENT",
					Params: map[string]any{
						"reason": "Inward discrepancy flagged on dock",
					},
				},
				{
					ActionType: "NOTIFY_ROLE",
					Params: map[string]any{
						"role":    "merchandiser",
						"message": "Vendor invoice discrepancy detected; review shortage variance",
					},
				},
			},
			Enabled: true,
		},
		{
			ID:           "rule-floor-count-shortage-hold",
			Name:         "Floor Count Shortage Debit Note & Payment Freeze",
			TriggerEvent: "INWARD_FLOOR_COUNTED",
			Condition: &Condition{
				Field: "shortfall_qty",
				Op:    ">",
				Val:   0.0,
			},
			Actions: []Action{
				{
					ActionType: "DRAFT_DEBIT_NOTE",
					Params: map[string]any{
						"reason": "Physical floor count quantity less than billed invoice",
						"status": "DRAFT",
					},
				},
				{
					ActionType: "HOLD_PAYMENT",
					Params: map[string]any{
						"reason": "Inward floor count shortage discrepancy flagged",
					},
				},
				{
					ActionType: "NOTIFY_ROLE",
					Params: map[string]any{
						"role":    "merchandiser",
						"message": "Vendor invoice floor shortage detected; debit note drafted",
					},
				},
			},
			Enabled: true,
		},
		{
			ID:           "rule-order-bom-init",
			Name:         "Order Booking Automatic BOM Allocation",
			TriggerEvent: "ORDER_BOOKED",
			Actions: []Action{
				{
					ActionType: "NOTIFY_ROLE",
					Params: map[string]any{
						"role":    "planning",
						"message": "New production order booked; initiate BOM reservation",
					},
				},
				{
					ActionType: "SCHEDULE_STAGE",
					Params: map[string]any{
						"initial_stage": "Cutting",
					},
				},
			},
			Enabled: true,
		},
		{
			ID:           "rule-reject-roll-flag",
			Name:         "Defective Roll Rejection Alert",
			TriggerEvent: "ROLL_RETURNED_AS_REJECT",
			Actions: []Action{
				{
					ActionType: "NOTIFY_ROLE",
					Params: map[string]any{
						"role":    "qc_supervisor",
						"message": "Defect roll returned from cutting; review Lycra needle damage",
					},
				},
			},
			Enabled: true,
		},
	}

	for _, r := range rules {
		condBytes, _ := json.Marshal(r.Condition)
		actBytes, _ := json.Marshal(r.Actions)
		now := graph.NowMs()

		query := `
			INSERT INTO workflow_rules (id, name, trigger_event, condition_json, actions_json, enabled, created_at)
			VALUES (?, ?, ?, ?, ?, 1, ?)
			ON CONFLICT(id) DO UPDATE SET
				name = excluded.name,
				condition_json = excluded.condition_json,
				actions_json = excluded.actions_json
		`
		if _, err := e.db.Exec(query, r.ID, r.Name, r.TriggerEvent, string(condBytes), string(actBytes), now); err != nil {
			return err
		}
	}

	log.Println("✓ Workflow automation rules initialized (Shortage debit note, BOM allocation, Defect alert)")
	return nil
}

// ProcessPendingEvents polls the outbox and evaluates active rules against events
func (e *Engine) ProcessPendingEvents() (int, error) {
	rows, err := e.db.Query(`
		SELECT id, aggregate_id, event_type, payload_json, depth, occurred_at
		FROM domain_events
		WHERE dispatched_at IS NULL
		ORDER BY occurred_at ASC
		LIMIT 50
	`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type PendingEvent struct {
		ID          string
		AggregateID string
		EventType   string
		Payload     map[string]any
		Depth       int
		OccurredAt  int64
	}

	var events []PendingEvent
	for rows.Next() {
		var pe PendingEvent
		var payloadStr string
		if err := rows.Scan(&pe.ID, &pe.AggregateID, &pe.EventType, &payloadStr, &pe.Depth, &pe.OccurredAt); err != nil {
			return 0, err
		}
		props, _ := graph.DecodeProps(payloadStr)
		pe.Payload = props
		events = append(events, pe)
	}

	processedCount := 0
	now := graph.NowMs()

	for _, evt := range events {
		// Find matching active rules
		ruleRows, err := e.db.Query(`
			SELECT id, name, condition_json, actions_json
			FROM workflow_rules
			WHERE trigger_event = ? AND enabled = 1
		`, evt.EventType)
		if err != nil {
			continue
		}

		type MatchedRuleItem struct {
			rID     string
			rName   string
			actions []Action
		}
		var matchedRules []MatchedRuleItem

		for ruleRows.Next() {
			var rID, rName, condStr, actStr string
			if err := ruleRows.Scan(&rID, &rName, &condStr, &actStr); err != nil {
				continue
			}

			// Evaluate condition
			if condStr != "" && condStr != "null" {
				var cond Condition
				if err := json.Unmarshal([]byte(condStr), &cond); err == nil {
					if !evaluateCondition(cond, evt.Payload) {
						continue // Condition not met
					}
				}
			}

			// Parse actions
			var actions []Action
			if err := json.Unmarshal([]byte(actStr), &actions); err != nil {
				continue
			}

			matchedRules = append(matchedRules, MatchedRuleItem{
				rID:     rID,
				rName:   rName,
				actions: actions,
			})
		}
		ruleRows.Close()

		// Execute actions with idempotent recording into action_effects
		for _, mr := range matchedRules {
			for idx, action := range mr.actions {
				resJSON, _ := json.Marshal(map[string]any{
					"executed": true,
					"rule":     mr.rName,
					"action":   action.ActionType,
					"params":   action.Params,
					"at":       time.Now().Format(time.RFC3339),
				})

				e.db.Exec(`
					INSERT INTO action_effects (event_id, rule_id, action_idx, status, result_json, executed_at)
					VALUES (?, ?, ?, 'EXECUTED', ?, ?)
					ON CONFLICT(event_id, rule_id, action_idx) DO NOTHING
				`, evt.ID, mr.rID, idx, string(resJSON), now)
			}
		}

		// Mark event as dispatched
		e.db.Exec("UPDATE domain_events SET dispatched_at = ? WHERE id = ?", now, evt.ID)
		processedCount++
	}

	return processedCount, nil
}

// evaluateCondition tests if a payload matches the condition predicate
func evaluateCondition(c Condition, payload map[string]any) bool {
	val, ok := payload[c.Field]
	if !ok {
		return false
	}

	switch c.Op {
	case ">":
		vNum, vOk := toFloat(val)
		targetNum, tOk := toFloat(c.Val)
		return vOk && tOk && vNum > targetNum
	case ">=":
		vNum, vOk := toFloat(val)
		targetNum, tOk := toFloat(c.Val)
		return vOk && tOk && vNum >= targetNum
	case "<":
		vNum, vOk := toFloat(val)
		targetNum, tOk := toFloat(c.Val)
		return vOk && tOk && vNum < targetNum
	case "<=":
		vNum, vOk := toFloat(val)
		targetNum, tOk := toFloat(c.Val)
		return vOk && tOk && vNum <= targetNum
	case "==":
		return fmt.Sprintf("%v", val) == fmt.Sprintf("%v", c.Val)
	case "!=":
		return fmt.Sprintf("%v", val) != fmt.Sprintf("%v", c.Val)
	default:
		return false
	}
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}
