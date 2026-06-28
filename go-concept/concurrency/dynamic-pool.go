package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Define the missing Task type expected by your pool implementation
type Task func(ctx context.Context)

// DynamicWorkerPool definitions as provided
type DynamicWorkerPool struct {
	tasks         chan Task
	minWorkers    int32
	maxWorkers    int32
	activeWorkers int32
	idleTimeout   time.Duration
	wg            sync.WaitGroup
	quit          chan struct{}
}

func NewDynamicWorkerPool(min, max int32, idleTimeout time.Duration) *DynamicWorkerPool {
	return &DynamicWorkerPool{
		tasks:       make(chan Task, 1000),
		minWorkers:  min,
		maxWorkers:  max,
		idleTimeout: idleTimeout,
		quit:        make(chan struct{}),
	}
}

func (dwp *DynamicWorkerPool) Start(ctx context.Context) {
	for i := int32(0); i < dwp.minWorkers; i++ {
		dwp.spawnWorker(ctx, false)
	}
}

func (dwp *DynamicWorkerPool) spawnWorker(ctx context.Context, isDynamic bool) {
	atomic.AddInt32(&dwp.activeWorkers, 1)
	dwp.wg.Add(1)

	go func() {
		defer dwp.wg.Done()
		defer atomic.AddInt32(&dwp.activeWorkers, -1)

		timer := time.NewTimer(dwp.idleTimeout)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-dwp.quit:
				return
			case task := <-dwp.tasks:
				if !timer.Stop() && isDynamic {
					<-timer.C
				}
				task(ctx)
				if isDynamic {
					timer.Reset(dwp.idleTimeout)
				}
			case <-timer.C:
				if isDynamic {
					return // Scale down idle dynamic worker
				}
			}
		}
	}()
}

func (dwp *DynamicWorkerPool) Submit(ctx context.Context, task Task) {
	select {
	case dwp.tasks <- task:
		currentActive := atomic.LoadInt32(&dwp.activeWorkers)
		if len(dwp.tasks) > 10 && currentActive < dwp.maxWorkers {
			dwp.spawnWorker(ctx, true)
		}
	case <-ctx.Done():
	}
}

func (dwp *DynamicWorkerPool) Stop() {
	close(dwp.quit)
	dwp.wg.Wait()
}

func main() {
	// Create a cancellable context for running the pool
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize Pool: 2 baseline workers, max 10 workers, 500ms idle timeout
	minWorkers := int32(2)
	maxWorkers := int32(10)
	idleTimeout := 500 * time.Millisecond

	pool := NewDynamicWorkerPool(minWorkers, maxWorkers, idleTimeout)
	pool.Start(ctx)

	fmt.Printf("[Pool Initialized] Min Workers: %d, Max Workers: %d\n", minWorkers, maxWorkers)
	fmt.Printf("Active Workers at baseline: %d\n\n", atomic.LoadInt32(&pool.activeWorkers))

	// Helper helper factory to create dummy workload tasks
	createTask := func(id int) Task {
		return func(ctx context.Context) {
			// Simulate varying workload execution times
			time.Sleep(100 * time.Millisecond)
		}
	}

	// 2. Simulate a sudden spike in traffic
	// We flood the task channel instantly to bypass the backlog threshold (>10 tasks)
	// and force the allocation of dynamic scale-up workers.
	spikeCount := 40
	fmt.Printf("--- STAGE 1: Flooding pool with %d concurrent tasks ---\n", spikeCount)

	for i := 1; i <= spikeCount; i++ {
		pool.Submit(ctx, createTask(i))
	}

	// Give the submissions a microsecond to trigger the conditional spawn blocks
	time.Sleep(10 * time.Millisecond)
	fmt.Printf("Active Workers during heavy load spike: %d (Scaled up!)\n", atomic.LoadInt32(&pool.activeWorkers))

	// 3. Wait for the initial burst to drain completely
	fmt.Println("\n--- STAGE 2: Waiting for tasks to finish processing... ---")
	time.Sleep(1 * time.Second)
	fmt.Printf("Active Workers immediately after drain: %d\n", atomic.LoadInt32(&pool.activeWorkers))

	// 4. Observe the Scale Down
	// Wait longer than our configured 500ms idle timeout.
	// The dynamic workers should hit their timer branches and exit.
	fmt.Printf("\n--- STAGE 3: Idling for %v to test scale-down execution ---\n", idleTimeout*2)
	time.Sleep(idleTimeout * 2)
	fmt.Printf("Active Workers after idle cooldown: %d (Scaled back down to baseline!)\n", atomic.LoadInt32(&pool.activeWorkers))

	// 5. Clean teardown
	fmt.Println("\n[Teardown] Shutting down pool safely...")
	pool.Stop()
	fmt.Println("All pool worker routines stopped cleanly.")
}
