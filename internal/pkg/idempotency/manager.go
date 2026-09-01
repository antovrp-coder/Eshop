package idempotency

/**
 * ============================================================================
 * PACKAGE: internal/pkg/idempotency
 * TECHNOLOGY: Redis In-Memory Locks / In-Memory State Machine
 * PATTERN: Idempotency Key Gatekeeper
 * 
 * WHY IDEMPOTENCY IS MANDATORY IN E-COMMERCE:
 * In unreliable network environments (mobile networks, retrying clients), duplicate HTTP
 * requests or gRPC calls can result in double-charging credit cards or double-decrementing stock.
 * 
 * THE IDEMPOTENCY PATTERN:
 * 1. The client supplies an `Idempotency-Key` header with every mutating action.
 * 2. The server acquires a lock on `idempotency:<key>`.
 * 3. If key is in-flight (PENDING), second request is rejected with 409 Conflict.
 * 4. Once executed, the resulting response is cached with a 24-hour TTL.
 * 5. Subsequent duplicate requests receive the cached result with zero side-effects.
 * ============================================================================
 */

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrDuplicateRequest = errors.New("idempotency key is currently being processed")
	ErrKeyExpired        = errors.New("idempotency key has expired")
)

type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "PENDING"
	StatusCompleted ExecutionStatus = "COMPLETED"
	StatusFailed    ExecutionStatus = "FAILED"
)

type Record struct {
	Key       string
	Status    ExecutionStatus
	Response  []byte
	CreatedAt time.Time
	ExpiresAt time.Time
}

type Manager struct {
	mu      sync.RWMutex
	records map[string]*Record
	ttl     time.Duration
}

func NewManager(ttl time.Duration) *Manager {
	m := &Manager{
		records: make(map[string]*Record),
		ttl:     ttl,
	}

	// Background cleanup for expired keys
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		for range ticker.C {
			m.cleanup()
		}
	}()

	return m
}

// LockOrGet checks if key is already processed or locks it for current execution.
func (m *Manager) LockOrGet(ctx context.Context, key string) (*Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if rec, exists := m.records[key]; exists {
		if now.After(rec.ExpiresAt) {
			delete(m.records, key)
		} else {
			if rec.Status == StatusPending {
				return nil, ErrDuplicateRequest
			}
			return rec, nil // Return cached result
		}
	}

	// Create pending record
	newRec := &Record{
		Key:       key,
		Status:    StatusPending,
		CreatedAt: now,
		ExpiresAt: now.Add(m.ttl),
	}
	m.records[key] = newRec
	return nil, nil // Nil record indicates caller should execute the logic
}

// Complete saves the execution response for future replays.
func (m *Manager) Complete(ctx context.Context, key string, response []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if rec, exists := m.records[key]; exists {
		rec.Status = StatusCompleted
		rec.Response = response
	}
}

// Fail unlocks the key if the downstream action failed so the client can retry.
func (m *Manager) Fail(ctx context.Context, key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.records, key)
}

func (m *Manager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for k, rec := range m.records {
		if now.After(rec.ExpiresAt) {
			delete(m.records, k)
		}
	}
}
