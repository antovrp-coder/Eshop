package concurrency

/**
 * ============================================================================
 * PACKAGE: internal/pkg/concurrency
 * TECHNOLOGY: Go Concurrency Primitives (Goroutines, Channels, sync.WaitGroup, Context)
 * PATTERN: Bounded Worker Pool & Fan-Out / Fan-In
 * 
 * WHY CONCURRENCY CONTROL IS CRITICAL IN E-COMMERCE:
 * 1. Backpressure & Resource Protection: Unbounded goroutine creation under high traffic
 *    (e.g., 50k simultaneous orders) can exhaust memory and file descriptors.
 *    A Bounded Worker Pool limits concurrent execution to a safe threshold (e.g., N CPU cores * 4).
 * 2. Non-Blocking Execution: Asynchronous domain tasks (sending confirmation emails,
 *    updating search indexes, publishing Kafka events) run concurrently without blocking
 *    the user's checkout response.
 * 3. Graceful Shutdown: Worker pools monitor context cancellation and drain in-flight jobs
 *    before terminating during container deployments.
 * ============================================================================
 */

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

var (
	ErrPoolClosed   = errors.New("worker pool is closed")
	ErrJobTimeout   = errors.New("job execution exceeded context deadline")
)

// Job represents a unit of work executed by the pool
type Job func(ctx context.Context) error

// Pool manages a bounded set of worker goroutines processing jobs from a channel
type Pool struct {
	workerCount int
	jobQueue    chan Job
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	closed      bool
	mu          sync.Mutex
}

// NewPool initializes and starts the worker goroutines
func NewPool(ctx context.Context, workerCount int, queueCapacity int) *Pool {
	poolCtx, cancel := context.WithCancel(ctx)
	p := &Pool{
		workerCount: workerCount,
		jobQueue:    make(chan Job, queueCapacity),
		ctx:         poolCtx,
		cancel:      cancel,
	}

	p.start()
	return p
}

func (p *Pool) start() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker is the concurrent loop continuously pulling jobs from the channel
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			// Context canceled, drain remaining jobs or exit
			return
		case job, ok := <-p.jobQueue:
			if !ok {
				return // Channel closed
			}
			p.executeJob(id, job)
		}
	}
}

func (p *Pool) executeJob(workerID int, job Job) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[CONCURRENCY PANIC] Worker %d recovered from panic: %v", workerID, r)
		}
	}()

	jobCtx, cancel := context.WithTimeout(p.ctx, 10*time.Second)
	defer cancel()

	if err := job(jobCtx); err != nil {
		log.Printf("[CONCURRENCY ERROR] Worker %d job failed: %v", workerID, err)
	}
}

// Submit queues a job for execution. Returns error if pool is closed or queue is full
func (p *Pool) Submit(job Job) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return ErrPoolClosed
	}
	p.mu.Unlock()

	select {
	case p.jobQueue <- job:
		return nil
	case <-p.ctx.Done():
		return p.ctx.Err()
	}
}

// Stop closes the queue and waits for all active workers to finish gracefully
func (p *Pool) Stop() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.mu.Unlock()

	close(p.jobQueue)
	p.cancel()
	p.wg.Wait()
}

// FanOutFanIn executes multiple tasks concurrently and collects their results
func FanOutFanIn[T any, R any](
	ctx context.Context,
	items []T,
	workerFn func(ctx context.Context, item T) (R, error),
) ([]R, []error) {
	results := make([]R, len(items))
	errs := make([]error, len(items))

	var wg sync.WaitGroup
	wg.Add(len(items))

	for i, item := range items {
		go func(idx int, it T) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errs[idx] = errors.New("goroutine panic recovered")
				}
			}()

			res, err := workerFn(ctx, it)
			if err != nil {
				errs[idx] = err
			} else {
				results[idx] = res
			}
		}(i, item)
	}

	wg.Wait()
	return results, errs
}
