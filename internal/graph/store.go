package graph

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// InsertNode adds a new node to the unified graph
func (s *Store) InsertNode(tx *sql.Tx, nodeType string, props map[string]any, customID string) (*Node, error) {
	id := customID
	if id == "" {
		id = fmt.Sprintf("node_%s_%s", nodeType, uuid.New().String()[:8])
	}

	now := NowMs()
	propsStr, err := EncodeProps(props)
	if err != nil {
		return nil, fmt.Errorf("failed to encode props: %w", err)
	}

	query := `
		INSERT INTO graph_nodes (id, node_type, props_json, valid_from, valid_to, txn_from, txn_to)
		VALUES (?, ?, ?, ?, NULL, ?, NULL)
	`
	var execer interface {
		Exec(query string, args ...any) (sql.Result, error)
	} = s.db
	if tx != nil {
		execer = tx
	}

	if _, err := execer.Exec(query, id, nodeType, propsStr, now, now); err != nil {
		return nil, fmt.Errorf("failed to insert graph node: %w", err)
	}

	return &Node{
		ID:        id,
		NodeType:  nodeType,
		Props:     props,
		ValidFrom: now,
		ValidTo:   nil,
		TxnFrom:   now,
		TxnTo:     nil,
	}, nil
}

// EnsureNode checks if an active node with the specified ID already exists.
// If it exists, it returns the existing node with stored properties and timestamps.
// If it does not exist, it inserts the node.
func (s *Store) EnsureNode(tx *sql.Tx, nodeType string, props map[string]any, customID string) (*Node, error) {
	if customID == "" {
		return s.InsertNode(tx, nodeType, props, "")
	}

	var execer interface {
		QueryRow(query string, args ...any) *sql.Row
	}
	if tx != nil {
		execer = tx
	} else {
		execer = s.db
	}

	var existing Node
	var propsStr string
	err := execer.QueryRow(`
		SELECT id, node_type, props_json, valid_from, valid_to, txn_from, txn_to
		FROM graph_nodes
		WHERE id = ? AND txn_to IS NULL AND valid_to IS NULL
	`, customID).Scan(&existing.ID, &existing.NodeType, &propsStr, &existing.ValidFrom, &existing.ValidTo, &existing.TxnFrom, &existing.TxnTo)
	if err == nil {
		decodedProps, _ := DecodeProps(propsStr)
		existing.Props = decodedProps
		return &existing, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to query existing node %s: %w", customID, err)
	}

	return s.InsertNode(tx, nodeType, props, customID)
}

// UpdateNode applies the append-only bi-temporal SGi pattern:
// Closes the previous active version by setting txn_to = now,
// and inserts a new version with txn_from = now.
func (s *Store) UpdateNode(tx *sql.Tx, nodeID string, newProps map[string]any) (*Node, error) {
	now := NowMs()
	var execer interface {
		Exec(query string, args ...any) (sql.Result, error)
		QueryRow(query string, args ...any) *sql.Row
	}
	if tx != nil {
		execer = tx
	} else {
		execer = s.db
	}

	// 1. Get current active node type and valid_from
	var nodeType string
	var validFrom int64
	err := execer.QueryRow("SELECT node_type, valid_from FROM graph_nodes WHERE id = ? AND txn_to IS NULL", nodeID).Scan(&nodeType, &validFrom)
	if err != nil {
		return nil, fmt.Errorf("node %s not found or already closed: %w", nodeID, err)
	}

	// 2. Supersede old record
	if _, err := execer.Exec("UPDATE graph_nodes SET txn_to = ? WHERE id = ? AND txn_to IS NULL", now, nodeID); err != nil {
		return nil, fmt.Errorf("failed to supersede old node: %w", err)
	}

	// 3. Insert new record
	propsStr, err := EncodeProps(newProps)
	if err != nil {
		return nil, fmt.Errorf("failed to encode new props: %w", err)
	}

	insertQuery := `
		INSERT INTO graph_nodes (id, node_type, props_json, valid_from, valid_to, txn_from, txn_to)
		VALUES (?, ?, ?, ?, NULL, ?, NULL)
	`
	if _, err := execer.Exec(insertQuery, nodeID, nodeType, propsStr, validFrom, now); err != nil {
		return nil, fmt.Errorf("failed to insert updated node: %w", err)
	}

	return &Node{
		ID:        nodeID,
		NodeType:  nodeType,
		Props:     newProps,
		ValidFrom: validFrom,
		ValidTo:   nil,
		TxnFrom:   now,
		TxnTo:     nil,
	}, nil
}

// InsertEdge creates a typed relationship between two nodes
func (s *Store) InsertEdge(tx *sql.Tx, srcID, dstID, edgeType string, props map[string]any) (*Edge, error) {
	edgeID := fmt.Sprintf("edge_%s_%s", edgeType, uuid.New().String()[:8])
	now := NowMs()
	propsStr, err := EncodeProps(props)
	if err != nil {
		return nil, fmt.Errorf("failed to encode edge props: %w", err)
	}

	query := `
		INSERT INTO graph_edges (id, src_id, dst_id, edge_type, props_json, valid_from, valid_to, txn_from, txn_to)
		VALUES (?, ?, ?, ?, ?, ?, NULL, ?, NULL)
	`
	var execer interface {
		Exec(query string, args ...any) (sql.Result, error)
	} = s.db
	if tx != nil {
		execer = tx
	}

	if _, err := execer.Exec(query, edgeID, srcID, dstID, edgeType, propsStr, now, now); err != nil {
		return nil, fmt.Errorf("failed to insert graph edge: %w", err)
	}

	return &Edge{
		ID:        edgeID,
		SrcID:     srcID,
		DstID:     dstID,
		EdgeType:  edgeType,
		Props:     props,
		ValidFrom: now,
		ValidTo:   nil,
		TxnFrom:   now,
		TxnTo:     nil,
	}, nil
}

// GetNode retrieves the current active version of a node
func (s *Store) GetNode(nodeID string) (*Node, error) {
	query := "SELECT id, node_type, props_json, valid_from, valid_to, txn_from, txn_to FROM graph_nodes WHERE id = ? AND txn_to IS NULL"
	var n Node
	var propsStr string
	err := s.db.QueryRow(query, nodeID).Scan(&n.ID, &n.NodeType, &propsStr, &n.ValidFrom, &n.ValidTo, &n.TxnFrom, &n.TxnTo)
	if err != nil {
		return nil, err
	}
	props, err := DecodeProps(propsStr)
	if err != nil {
		return nil, err
	}
	n.Props = props
	return &n, nil
}

// Traverse finds nodes connected to startNodeID along active edges of specified type
func (s *Store) Traverse(startNodeID, edgeType, direction string) ([]*Node, error) {
	var query string
	if direction == "OUT" {
		query = `
			SELECT n.id, n.node_type, n.props_json, n.valid_from, n.valid_to, n.txn_from, n.txn_to
			FROM graph_edges e
			JOIN graph_nodes n ON e.dst_id = n.id AND n.txn_to IS NULL
			WHERE e.src_id = ? AND e.edge_type = ? AND e.txn_to IS NULL
		`
	} else {
		query = `
			SELECT n.id, n.node_type, n.props_json, n.valid_from, n.valid_to, n.txn_from, n.txn_to
			FROM graph_edges e
			JOIN graph_nodes n ON e.src_id = n.id AND n.txn_to IS NULL
			WHERE e.dst_id = ? AND e.edge_type = ? AND e.txn_to IS NULL
		`
	}

	rows, err := s.db.Query(query, startNodeID, edgeType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		var n Node
		var propsStr string
		if err := rows.Scan(&n.ID, &n.NodeType, &propsStr, &n.ValidFrom, &n.ValidTo, &n.TxnFrom, &n.TxnTo); err != nil {
			return nil, err
		}
		props, err := DecodeProps(propsStr)
		if err != nil {
			return nil, err
		}
		n.Props = props
		nodes = append(nodes, &n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return nodes, nil
}

// AsOfNode performs time-travel query: retrieves belief state of node at timestamp t
func (s *Store) AsOfNode(nodeID string, timestampMs int64) (*Node, error) {
	query := `
		SELECT id, node_type, props_json, valid_from, valid_to, txn_from, txn_to
		FROM graph_nodes
		WHERE id = ? AND txn_from <= ? AND (txn_to IS NULL OR txn_to > ?)
		ORDER BY txn_from DESC
		LIMIT 1
	`
	var n Node
	var propsStr string
	err := s.db.QueryRow(query, nodeID, timestampMs, timestampMs).Scan(
		&n.ID, &n.NodeType, &propsStr, &n.ValidFrom, &n.ValidTo, &n.TxnFrom, &n.TxnTo,
	)
	if err != nil {
		return nil, err
	}
	props, err := DecodeProps(propsStr)
	if err != nil {
		return nil, err
	}
	n.Props = props
	return &n, nil
}
