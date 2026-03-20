package migrations

import "embed"

// Files contains SQL migration assets bundled into the binary.
//
//go:embed *.sql
var Files embed.FS
