package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})
}

func TestRunAdditional(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("m <= 0 ignores errors and runs all tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				atomic.AddInt32(&runTasksCount, 1)
				return fmt.Errorf("error on task")
			})
		}

		err := Run(tasks, 5, 0)
		require.NoError(t, err)
		require.Equal(t, int32(tasksCount), runTasksCount)
	})

	t.Run("concurrency without time.Sleep (require.Eventually)", func(t *testing.T) {
		const tasksCount = 100
		tasks := make([]Task, 0, tasksCount)

		var current, maxConcurrency int32
		release := make(chan struct{})

		for i := 0; i < tasksCount; i++ {
			tasks = append(tasks, func() error {
				cur := atomic.AddInt32(&current, 1)
				for {
					old := atomic.LoadInt32(&maxConcurrency)
					if cur <= old || atomic.CompareAndSwapInt32(&maxConcurrency, old, cur) {
						break
					}
				}
				<-release
				atomic.AddInt32(&current, -1)
				return nil
			})
		}

		errCh := make(chan error, 1)
		go func() {
			errCh <- Run(tasks, 10, 1)
		}()

		require.Eventually(t, func() bool {
			return atomic.LoadInt32(&maxConcurrency) >= 10
		}, 2*time.Second, 20*time.Millisecond)

		close(release)
		err := <-errCh
		require.NoError(t, err)
	})
}
