package outbox

/**
 * ============================================================================
 * PACKAGE: internal/pkg/outbox
 * PATTERN: Transactional Outbox Pattern
 * TECHNOLOGY: PostgreSQL + Kafka CDC / Polling Publisher
 * 
 * WHY THE TRANSACTIONAL OUTBOX PATTERN?
 * Problem: If an order service saves an order to PostgreSQL and then attempts
 * to send a message to Kafka, network failure or service crash can cause dual-write
 * inconsistency (DB committed, but Kafka event lost).
 * 
 * Solution: The domain event is inserted into PostgreSQL `outbox_events` table
 * within the EXACT SAME ACID transaction as the order. This background publisher
 * (or Debezium CDC) reads unpublished events, dispatches them to Kafka, and marks
 * them as `published = TRUE`.
 * ============================================================================
 */

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/stylehome/eshop/internal/pkg/events"
)

type OutboxRecord struct {
	ID            string          `json:"id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	Published     bool            `json:"published"`
	CreatedAt     time.Time       `json:"created_at"`
}

type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload []byte) error
}

type MemoryEventPublisher struct {
	mu     sync.Mutex
	Events []OutboxRecord
}

func NewMemoryPublisher() *MemoryEventPublisher {
	return &MemoryEventPublisher{
		Events: make([]OutboxRecord, 0),
	}
}

func (m *MemoryEventPublisher) Publish(ctx context.Context, topic string, key string, payload []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	log.Printf("[KAFKA DISPATCH] topic=%s key=%s bytes=%d", topic, key, len(payload))
	return nil
}

type Worker struct {
	publisher EventPublisher
	records   []OutboxRecord
	mu        sync.Mutex
	stopChan  chan struct{}
}

func NewWorker(pub EventPublisher) *Worker {
	return &Worker{
		publisher: pub,
		records:   make([]OutboxRecord, 0),
		stopChan:  make(chan struct{}),
	}
}

// StageEvent stores the outbox event in the transactional queue
func (w *Worker) StageEvent(rec OutboxRecord) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.records = append(w.records, rec)
}

// Start begins the asynchronous poller
func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(200 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				w.flush(ctx)
			case <-w.stopChan:
				ticker.Stop()
				return
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	close(w.stopChan)
}

func (w *Worker) flush(ctx context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i := range w.records {
		if !w.records[i].Published {
			topic := events.TopicOrderEvents
			if w.records[i].AggregateType == "PAYMENT" {
				topic = events.TopicPaymentEvents
			}
			err := w.publisher.Publish(ctx, topic, w.records[i].AggregateID, w.records[i].Payload)
			if err == nil {
				w.records[i].Published = true
			}
		}
	}
}
