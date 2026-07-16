// Package service contains application business orchestration used by handlers.
//
// Keep this package focused on workflow-level logic. Shared pure helpers should
// live under backend/internal/app/support, while low-coupling domains can move
// behind compatibility shims one domain at a time.
package service
