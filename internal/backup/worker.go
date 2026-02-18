package backup

import (
	"context"
	"sync"
)

func runWorkerPool(ctx context.Context, workerCount int, tables []string, job func(string) error) error {
	jobs := make(chan string)
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for table := range jobs {
				if err := job(table); err != nil {
					select {
					case errChan <- err:
					default:
					}
					return
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, t := range tables {
			select {
			case <-ctx.Done():
				return
			case jobs <- t:
			}
		}
	}()

	wg.Wait()
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}
