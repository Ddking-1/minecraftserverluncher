// Package database owns the SQLite connection, schema initialization and
// the persisted models (accounts, devices, sessions, access tokens).
//
// The schema is initialized idempotently on startup. All token digests
// stored here are SHA-256 hashes of opaque tokens; raw tokens never touch
// the database.
package database

// Session CRUD lives in accounts.go. This file is a placeholder for any
// future session-specific extensions that are not already covered there.
