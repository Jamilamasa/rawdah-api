package migrations

import "embed"

// Files contains SQL migrations embedded into the binary.
//
//go:embed *.up.sql
var Files embed.FS
