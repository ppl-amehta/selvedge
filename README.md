# FPCC Selvedge Enterprise Edge Engine

> **10–20 Year Industrial System of Record for Apparel Manufacturing**  
> *Selvedge* (/ˈsɛlvɪdʒ/) — The self-finished edge of fabric preventing unraveling; here, an autonomous edge-computing Operating System for garment production.

---

## 🏛️ Architectural Foundation

Built in **Go 1.27 + SQLite WAL Mode + Chi Router** under an ultra-long-horizon foresight framework:
1. **Single Static Native Binary**: Compiles to a single zero-dependency binary (`GOOS=linux GOARCH=arm64 go build`). Runs on factory mini-PCs, Raspberry Pi 5 gateways, offline Zebra thermal barcode print stations, or edge servers with zero runtime bloat.
2. **Unified Graph System of Record**: Eliminates brittle relational foreign key sprawls by representing entities (`CLIENT`, `VENDOR`, `ORDER`, `PARENT_LOT`, `BUNDLE`, `FABRIC`, `COLOR`) as typed **Nodes** and domain relationships (`ORDERED_BY`, `SUPPLIED_BY`, `PART_OF_LOT`) as typed **Edges**.
3. **Bi-Temporal Append-Only (SGi Pattern)**: Every node and edge maintains `valid_from / valid_to` (when facts occurred in the real world) alongside `txn_from / txn_to` (when the system recorded or superseded the fact). Records are never destructively overwritten, enabling immutable audit trails and historical time-travel queries (`/api/graph/as-of`).
4. **Transactional Outbox & Smart Workflows**: Event-driven declarative automations (`domain_events`, `workflow_rules`, `action_effects`). Automatically drafts debit notes, freezes supplier payouts on dock shortages, and alerts planning stations on order booking.
5. **Fast Floor Projections**: Optimized read-model projection tables (`orders`, `parent_inward_lots`, `bundles`, `discrepancy_reports`) kept atomically synchronized with the graph for sub-millisecond floor queries.

---

## 🏭 Production Stations Implemented (Phased Greenfield Scope)

### Station 0: Production Order Taking
- **Format Generator**: Auto-generated sequential order numbers `S{YYMMDD}{A-Z}` (e.g. `S261019A`).
- **34-Point Spec Control**: Complete commercial, garment, sizing, and trim specs.
- **Modern Auto-Suggest Comboboxes**: Real-time typeahead querying catalog masters with seamless on-the-fly registration of new clients, products, fabrics, and colors.
- **Strict Enum Badges**: High-contrast pills for controlled manufacturing choices (Order Type, Category, Gender, Tax Rate, Tax Billing Mode, Freight Mode).
- **Instant Math**: Live size matrix calculator tallying total units, tax amounts, and grand total.
- **Multi-Shipping Logistics Split**: Allocate order quantities across branch offices, retail hubs, or stadium drops with live balance reconciliation.
- **Personalization Roster**: Sports jersey and corporate staff roster with direct **Bulk Paste from Excel / Google Sheets** parser.
- **Copy Specs**: One-click modal to clone specs from prior orders.

### Station 1: Raw Material Inward & Multi-Bundle QR Breakdown
- **Parent Lot Generator**: Sequential format `{YYMMDD}-{A-Z}` (e.g. `261019-A`).
- **Step A: Dock Receipt**: Gate scales, vendor lookup, invoice reconciliation, and contextual fabric/trim parameters.
- **Real-Time Discrepancy Reconciliation**: Automatic shortfall/excess detection. Shortages trigger red industrial warning banners, log financial exposure, draft debit notes, and freeze supplier payouts via the background workflow engine.
- **Step B: Physical Bundle Breakdown**: Split lots into individual rolls (`ParentCode-1/N ... N/N`) with individual scale weights, measured GSM, warehouse bin tags (`A/R01/S01`), and QC flags.
- **Factory QR Label Generation**: Pure client-side SVG 2D barcode generator for Zebra printers and mobile warehouse handhelds.
- **Remnant / Defect Return Workflow**: Return leftover bits or rejected rolls from cutting; automatically assigns `Bit` or `Rej` prefixes (e.g. `Bit261019-A-1/3`, `Rej261019-A-2/3`), updates available inventory, and alerts QC supervisors.

---

## 🚀 Quick Start

### Prerequisites
- Go 1.27+ (`brew install go`)
- GCC / CGo (standard macOS Xcode Command Line Tools)

### Build & Run
```bash
# 1. Compile native binary
go build -o selvedge ./cmd/selvedge

# 2. Run server on port 4000
./selvedge -port 4000 -db ./data/selvedge.db -web ./web/public
```

Open `http://localhost:4000` in any modern browser.

---

## 📡 Core API Endpoints

| Endpoint | Method | Description |
|---|---|---|
| `/api/health` | GET | System architecture and edge engine status |
| `/api/masters/all` | GET | Catalog entities, 301 colors, 225 fabrics, 270 locations |
| `/api/masters/register` | POST | Auto-register open-ended catalog items into projection + graph |
| `/api/orders/next-order-no` | GET | Next sequential order number (`S261019A`) |
| `/api/orders` | POST | Book production order with graph node & edge linking |
| `/api/orders/{orderNo}` | GET | Retrieve complete 34-point order specification |
| `/api/inventory/next-parent-code`| GET | Next parent lot code (`261019-A`) |
| `/api/inventory/inward` | POST | Commit Step A gate inward & Step B bundle roll breakdown |
| `/api/inventory/return-remnant` | POST | Process Remnant / Bit / Reject return workflow |
| `/api/inventory/discrepancies` | GET | List shortage & excess variance reports |
| `/api/graph/nodes/{nodeId}` | GET | Fetch bi-temporal node properties |
| `/api/graph/traverse` | GET | Traverse graph relationships along typed edges |
| `/api/workflows/dispatch` | POST | Trigger outbox event dispatch cycle |

---

## 📜 License
Proprietary — First Promotional Clothing Co. (FPCC).
