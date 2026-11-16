package migrations

import "embed"

// FS embeds all SQL schema migrations directly into the Go binary.
// This guarantees zero-dependency runtime execution in containers,
// serverless VMs (Fly.io, Koyeb, Render), and bare metal.
//
//go:embed *.sql
var FS embed.FS
