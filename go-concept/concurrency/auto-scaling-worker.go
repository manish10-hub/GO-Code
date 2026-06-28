package main
// Online Go compiler to run Golang program online
// Print "Start small. Ship something." message

package main
import (
    "fmt"
    "context"
    "sync"
    "sync/atomic"
    "time"
    )

/*
 q:"Design a worker pool with dynamic scaling, backpressure, and graceful shutdown for a high-throughput job processing service.",
    a:`This is a classic architecture question that tests understanding of goroutine lifecycle, backpressure, and graceful degradation.

1. Bounded job queue (buffered channel) — defines the backpressure point. When full, producers block or receive a rejection error depending on SLA.
  2. Dynamic worker count — controlled by a semaphore or by spawning/retiring goroutines based on queue depth.
  3. Graceful shutdown — close the job channel, drain it, wait for all workers to finish with sync.WaitGroup.
  4. Metrics — expose queue depth, active workers, processing latency for auto-scaling triggers.
  5. Error handling — dead letter queue for jobs that fail after N retries.
  
  */

                Producer
                    │
                    ▼
          ┌──────────────────┐
          │  Buffered Queue   │
          └──────────────────┘
                    │
                    ▼
            AutoScaler Loop
                    │
     ┌──────────────┴──────────────┐
     │                             │
Queue > High                Queue Empty
     │                             │
Spawn Workers              Retire Idle Workers
     │                             │
     ▼                             ▼
 Worker 1
 Worker 2
 Worker 3
 ...
 Worker N
     │
     ▼
  Process Jobs
     │
     ▼
Metrics + Retry + DLQ


  
 type job struct {
     ID int
 }
 
type  Pool struct{
    jobs chan job
    ctx context.Context
    cancel context.CancelFunc
    minWorker int
    maxWorker int
    wg sync.WaitGroup
    activeWorkers atomic.Int32
}

func newPool(queueSize int, minWorkers int, maxWorkers int) *Pool{
    if minWorkers <= 0 { 
        panic("invalid minWorkers") 
    } 
    if maxWorkers < minWorkers { 
        panic("maxWorkers must be >= minWorkers") 
    }
    ctx, cancel := context.WithCancel(context.Background())
    pool := &Pool{
        jobs: make(chan job, queueSize),
        minWorker: minWorkers,
        maxWorker: maxWorkers,
        ctx: ctx,
        cancel: cancel,
    }
    
    for i :=0; i<minWorkers; i++{
        pool.StartWorker()
    }
    
    go pool.AutoScale()
    return pool
}
func (p *Pool) StartWorker(){
    id := p.activeWorkers.Add(1)
    p.wg.Add(1)
    
    go func(workerID int32){
        defer p.wg.Done()
        defer p.activeWorkers.Add(-1)
        fmt.Printf("[worker-%d] started", workerID)
        idelTimer := time.NewTimer(30 * time.Second)
        defer idelTimer.Stop()
        for{
            select{
                case <-p.ctx.Done():
                    return
                case job, ok := <- p.jobs:
                    if !ok{
                        return
                    }
                    if !idleTimer.Stop() { 
                        select { 
                            case <-idleTimer.C: 
                            default: 
                        } 
                    } 
                    idleTimer.Reset(30 * time.Second)
                    p.process(job)
                    
                case <-idelTimer.C:
                    if p.activeWorkers.Load() > int32(p.minWorker) && len(p.jobs) == 0{
                        return
                    }
            }
        }
        
    }(id)
}
func (p *Pool) process(job job){
    fmt.Printf("processing job %d \n", job.ID)
    time.Sleep(100 * time.Millisecond)
}

func (p *Pool) AutoScale(){
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    for{
        select{
            case <-ticker.C :
                queueCap := cap(p.jobs)
                queuelen := len(p.jobs)
                if queueCap < 1{
                    return
                }
                per := (queuelen*100)/queueCap
                if per > 70 && p.activeWorkers.Load() < int32(p.maxWorker){
                    p.StartWorker()
                }
            case <-p.ctx.Done():
                return
        }
    }
    
}
func (p *Pool) Submit(jobId job) error{
    select {
        case p.jobs <- jobId:
            return nil
        default:
            return fmt.Errorf("queue full")
    }
    
}
func (p *Pool) Shutdown(){
    fmt.Println("stopping producers")

	close(p.jobs)

	fmt.Println("cancelling workers")

	p.wg.Wait()

	fmt.Println("shutdown complete")
    p.cancel()

	fmt.Println("waiting for workers")
    
}
func main(){
    pool := newPool(100, 2, 10)
    
    go func() {
		for i := 1; i <= 1000; i++ {
			job := job{ID: i}
			err := pool.Submit(job)
			if err != nil {
				fmt.Printf("job=%d rejected: %v",job.ID,err)
				time.Sleep(500 * time.Millisecond)
				continue
			}
			time.Sleep(50 * time.Millisecond)
		}
	}()
	time.Sleep(60 * time.Second)
	pool.Shutdown()
}

