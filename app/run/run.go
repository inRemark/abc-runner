package run

import (
	"fmt"
	"redis-runner/app/utils"
	"sync"
	"sync/atomic"
	"time"
)

type OperationFunc func() (bool, bool, time.Duration)

func OperationRun(n, c int, operationFunc OperationFunc) (okCount, failCount, readCount, writeCount, rps int32, rts []time.Duration) {

	okCount = 0
	failCount = 0
	readCount = 0
	writeCount = 0

	taskChan := make(chan struct{}, n)
	for i := 0; i < n; i++ {
		taskChan <- struct{}{}
	}
	close(taskChan)

	var progress int64 = 0
	utils.PrintProgress(int64(n), &progress)

	localRTS := make([][]time.Duration, n)
	for i := 0; i < n; i++ {
		localRTS[i] = make([]time.Duration, 0)
	}

	var wg sync.WaitGroup
	wg.Add(c)
	start := time.Now()
	for i := 0; i < c; i++ {
		go func(idx int) {
			defer wg.Done()
			for range taskChan {
				isOk, isRead, rt := operationFunc()
				if isOk {
					atomic.AddInt32(&okCount, 1)
				} else {
					atomic.AddInt32(&failCount, 1)
				}
				if isRead {
					atomic.AddInt32(&readCount, 1)
				} else {
					atomic.AddInt32(&writeCount, 1)
				}
				localRTS[idx] = append(localRTS[idx], rt)
				atomic.AddInt64(&progress, 1)
			}
		}(i)
	}
	wg.Wait()
	since := time.Since(start)
	rps = int32(float64(n) / since.Seconds())

	rts = make([]time.Duration, 0, n)
	for _, localSlice := range localRTS {
		rts = append(rts, localSlice...)
	}

	fmt.Printf("\rProgress: %d / %d, %.2f%%\n", int(readCount)+int(writeCount), n, 100.0)
	return okCount, failCount, readCount, writeCount, rps, rts
}
