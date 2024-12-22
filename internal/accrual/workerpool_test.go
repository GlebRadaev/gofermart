package accrual

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerPool(t *testing.T) {
	tests := []struct {
		name           string
		numTasks       int
		numWorkers     int
		expectedErrors int
	}{
		{
			name:           "Test worker pool with simple tasks",
			numTasks:       5,
			numWorkers:     2,
			expectedErrors: 0,
		},
		{
			name:           "Test worker pool with error in task",
			numTasks:       2,
			numWorkers:     2,
			expectedErrors: 1,
		},
		// {
		// 	name:           "Test worker pool with canceled context",
		// 	numTasks:       1,
		// 	numWorkers:     1,
		// 	expectedErrors: 0,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wp := NewWorkerPool(tt.numWorkers)
			defer wp.Close()

			var mu sync.Mutex
			var taskExecutionCount int
			var errorCount int
			var wg sync.WaitGroup

			// if tt.name == "Test worker pool with canceled context" {
			// 	ctx, cancel := context.WithCancel(context.Background())
			// 	cancel()

			// 	err := wp.AddTask(ctx, func() error {
			// 		t.Error("Task should not be executed")
			// 		time.Sleep(100 * time.Millisecond)
			// 		return nil
			// 	})
			// 	assert.Error(t, err)
			// 	assert.Equal(t, context.Canceled, err)
			// 	return
			// }

			for i := 0; i < tt.numTasks; i++ {
				wg.Add(1)
				task := func(i int) func() error {
					return func() error {
						defer wg.Done()
						if i == tt.numTasks-1 && tt.expectedErrors > 0 {
							mu.Lock()
							errorCount++
							mu.Unlock()
							return assert.AnError
						}
						time.Sleep(200 * time.Millisecond)
						mu.Lock()
						taskExecutionCount++
						mu.Unlock()
						return nil
					}
				}(i)

				err := wp.AddTask(context.Background(), task)
				require.NoError(t, err, "failed to add task to pool")
			}

			wg.Wait()

			assert.Equal(t, tt.numTasks-tt.expectedErrors, taskExecutionCount, "number of executed tasks does not match")
			assert.Equal(t, tt.expectedErrors, errorCount, "number of errors does not match")

			wp.Close()
		})
	}
}
