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

	"github.com/stylehome/eshop/internal/pkg/concurrency"
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
	pool      *concurrency.Pool
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

/**
 * ============================================================================
 * CONCURRENCY PATTERN: Asynchronous Outbox Worker Pool
 * 
 * WHY A CONCURRENT DISPATCHER POOL?
 * Publishing events to Kafka brokers over the network involves socket I/O latency.
 * Running event dispatchers concurrently over a bounded pool of 10 workers ensures
 * high-throughput order event streaming (e.g. 5,000 events/sec) without stalling the poller loop.
 * ============================================================================
 */
func (w *Worker) Start(ctx context.Context) {
	w.pool = concurrency.NewPool(ctx, 10, 1000)
	ticker := time.NewTicker(200 * time.Millisecond)

	go func() {
		for {
			select {
			case <-ticker.C:
				w.flush(ctx)
			case <-w.stopChan:
				ticker.Stop()
				if w.pool != nil {
					w.pool.Stop()
				}
				return
			case <-ctx.Done():
				ticker.Stop()
				if w.pool != nil {
					w.pool.Stop()
				}
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
			idx := i
			topic := events.TopicOrderEvents
			if w.records[idx].AggregateType == "PAYMENT" {
				topic = events.TopicPaymentEvents
			}

			// Concurrently dispatch event via worker pool
			_ = w.pool.Submit(func(jobCtx context.Context) error {
				err := w.publisher.Publish(jobCtx, topic, w.records[idx].AggregateID, w.records[idx].Payload)
				if err == nil {
					w.mu.Lock()
					w.records[idx].Published = true
					w.mu.Unlock()
				}
				return err
			})
		}
	}
}
