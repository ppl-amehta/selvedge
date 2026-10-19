package graph

import (
	"encoding/json"
	"time"
)

// Standard Node Types
const (
	TypeClient     = "CLIENT"
	TypeVendor     = "VENDOR"
	TypeOrder      = "ORDER"
	TypeParentLot  = "PARENT_LOT"
	TypeBundle     = "BUNDLE"
	TypeFabric     = "FABRIC"
	TypeColor      = "COLOR"
	TypeLocation   = "LOCATION"
	TypeItemMaster = "ITEM_MASTER"
)

// Standard Edge Types
const (
	EdgeOrderedBy    = "ORDERED_BY"     // (Order) -> (Client)
	EdgeSuppliedBy   = "SUPPLIED_BY"    // (ParentLot) -> (Vendor)
	EdgePartOfLot    = "PART_OF_LOT"    // (Bundle) -> (ParentLot)
	EdgeStoredAt     = "STORED_AT"      // (Bundle) -> (Location)
	EdgeAllocatedTo  = "ALLOCATED_TO"   // (Fabric/Trim) -> (Order)
	EdgeIssuedTo     = "ISSUED_TO"      // (Bundle) -> (Order)
	EdgeReturnedFrom = "RETURNED_FROM"  // (Remnant/Bit) -> (Order)
)

// Node represents a vertex in the unified enterprise graph with bi-temporal validity
type Node struct {
	ID        string         `json:"id"`
	NodeType  string         `json:"node_type"`
	Props     map[string]any `json:"props"`
	ValidFrom int64          `json:"valid_from"` // Real-world event time (ms)
	ValidTo   *int64         `json:"valid_to"`   // Real-world expiry time (null = active)
	TxnFrom   int64          `json:"txn_from"`   // System recording time (ms)
	TxnTo     *int64         `json:"txn_to"`     // System superseding time (null = current belief)
}

// Edge represents a directed, typed relationship between two nodes
type Edge struct {
	ID        string         `json:"id"`
	SrcID     string         `json:"src_id"`
	DstID     string         `json:"dst_id"`
	EdgeType  string         `json:"edge_type"`
	Props     map[string]any `json:"props"`
	ValidFrom int64          `json:"valid_from"`
	ValidTo   *int64         `json:"valid_to"`
	TxnFrom   int64          `json:"txn_from"`
	TxnTo     *int64         `json:"txn_to"`
}

// Helper to get current Unix time in milliseconds
func NowMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// Helper to encode props into JSON string
func EncodeProps(props map[string]any) (string, error) {
	if props == nil {
		return "{}", nil
	}
	bytes, err := json.Marshal(props)
	if err != nil {
		return "{}", err
	}
	return string(bytes), nil
}

// Helper to decode JSON string into props map
func DecodeProps(jsonStr string) (map[string]any, error) {
	var props map[string]any
	if jsonStr == "" {
		return make(map[string]any), nil
	}
	if err := json.Unmarshal([]byte(jsonStr), &props); err != nil {
		return make(map[string]any), err
	}
	return props, nil
}
