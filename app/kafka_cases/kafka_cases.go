package kafkaCases

import (
	"redis-runner/app/runner"
	"redis-runner/app/utils"
	"strings"
	"sync"
	"time"
)

func DoProduceCase(printType string, n, c, d int) {
	value := strings.Repeat("X", d)
	okCount, failCount, readCount, writeCount, rps, rts := runner.OperationRun(n, c, func() (bool, bool, time.Duration) {
		return Produce("key", value)
	})
	metrics := utils.CalculateStatistics(rts)
	metrics.RPS = rps
	printSummary(printType, n, c, 0, d, 0, okCount, failCount, readCount, writeCount, metrics)
}

func DoConsumeCase(url, topic, groupId string, c int) {
	var wg sync.WaitGroup
	for i := 1; i <= c; i++ {
		reader := InitReaderCommand(url, topic, groupId)
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			Consume(reader, i)
		}(i)
	}
	wg.Wait()
}
