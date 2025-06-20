package chutils

import "sync"

func FanIn[T any](chs ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	out := make(chan T)

	for _, ch := range chs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for err := range ch {
				out <- err
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
