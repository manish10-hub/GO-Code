## Always pass context.Context and respect ctx.Done()

func worker(ctx context.Context, jobs <-chan string) {
	for {
		select {
		case <-ctx.Done():
			log.Println("worker shutting down")
			return

		case job, ok := <-jobs:
			if !ok {
				return
			}

			process(job)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	jobs := make(chan string)

	go worker(ctx, jobs)

	// shutdown
	cancel()
}


## Use WaitGroup to signal shutdown and wait for completion

func worker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Println("worker stopped")
			return

		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	go worker(ctx, &wg)

	// trigger shutdown
	cancel()

	// wait for worker to exit
	wg.Wait()
}

## Set timeouts on all blocking operations

ctx, cancel := context.WithTimeout(
	context.Background(),
	5*time.Second,
)
defer cancel()

req, err := http.NewRequestWithContext(
	ctx,
	http.MethodGet,
	"https://example.com",
	nil,
)
if err != nil {
	return err
}

resp, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer resp.Body.Close()