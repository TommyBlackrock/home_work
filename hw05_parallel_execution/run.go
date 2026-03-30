package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		n = 1
	}

	if len(tasks) == 0 {
		return nil
	}

	stopOnErrors := m > 0

	taskChan := make(chan Task, n)
	resultChan := make(chan error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				resultChan <- task()
			}
		}()
	}

	taskIndex := 0
	started := 0
	for taskIndex < len(tasks) && started < n {
		taskChan <- tasks[taskIndex]
		taskIndex++
		started++
	}

	errCount := 0
	finished := 0
	limitExceeded := false

	for finished < started {
		err := <-resultChan
		finished++

		if err != nil && stopOnErrors {
			errCount++
			if errCount >= m {
				limitExceeded = true
			}
		}

		if !limitExceeded && taskIndex < len(tasks) {
			taskChan <- tasks[taskIndex]
			taskIndex++
			started++
		}
	}

	close(taskChan)
	wg.Wait()
	close(resultChan)

	if limitExceeded {
		return ErrErrorsLimitExceeded
	}

	return nil
}
