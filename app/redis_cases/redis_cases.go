package redisCases

import (
	"log"
	"redis-runner/app/run"
	"redis-runner/app/utils"
	"sync"
	"time"
)

func Start(tc string) {
	ConfigConnect()
	var err error
	redisConfigs, err := LoadConfig()
	if err != nil {
		log.Fatalf("redis-config load failed: %v", err)
	}
	mode, n, c, r, d, R, ttl, testCase := redisConfigs.Mode,
		redisConfigs.BenchMark.Total,
		redisConfigs.BenchMark.Parallels,
		redisConfigs.BenchMark.RandomKeys,
		redisConfigs.BenchMark.DataSize,
		redisConfigs.BenchMark.ReadPercent,
		redisConfigs.BenchMark.TTL,
		redisConfigs.BenchMark.Case

	switch testCase {
	case "set_get_random":
		doSetGetRandomCase(mode, "logs", n, c, r, d, R, ttl)
	case "set":
		doSetCase(mode, "logs", n, c, r, d, R, ttl)
	case "get":
		doGetCase(mode, "logs", n, c)
	case "del":
		doDelCase(mode, "logs", n, c, r)
	case "pub":
		doPubCase(mode, "logs", n, c, d)
	case "sub":
		doSubCase(mode, c)
	case "default":
	default:
		doGetCase(mode, "logs", n, c)
	}
}

func doSetGetRandomCase(printType, mode string, n, c, r, d, R, ttl int) {
	okCount, failCount, readCount, writeCount, rps, rts := run.OperationRun(n, c, func() (bool, bool, time.Duration) {
		return DoSetGetRandomOperation(mode, n, r, d, R, ttl)
	})
	metrics := utils.CalculateStatistics(rts)
	metrics.RPS = rps
	printSummary(printType, n, c, R, d, ttl, okCount, failCount, readCount, writeCount, metrics)
}

func doSetCase(printType, mode string, n, c, r, d, R, ttl int) {
	okCount, failCount, readCount, writeCount, rps, rts := run.OperationRun(n, c, func() (bool, bool, time.Duration) {
		return DoSetOperation(mode, n, r, d, R, ttl)
	})
	metrics := utils.CalculateStatistics(rts)
	metrics.RPS = rps
	printSummary(printType, n, c, R, d, ttl, okCount, failCount, readCount, writeCount, metrics)
}

func doGetCase(printType, mode string, n, c int) {
	okCount, failCount, readCount, writeCount, rps, rts := run.OperationRun(n, c, func() (bool, bool, time.Duration) {
		return DoGetOperation(mode)
	})
	metrics := utils.CalculateStatistics(rts)
	metrics.RPS = rps
	printSummary(printType, n, c, 100, 0, 0, okCount, failCount, readCount, writeCount, metrics)
}

func doDelCase(printType, mode string, n, c, r int) {
	okCount, failCount, readCount, writeCount, rps, rts := run.OperationRun(n, c, func() (bool, bool, time.Duration) {
		return DoDeleteOperation(mode, n, r)
	})
	metrics := utils.CalculateStatistics(rts)
	metrics.RPS = rps
	printSummary(printType, n, c, 0, 0, 0, okCount, failCount, readCount, writeCount, metrics)
}

func doPubCase(printType, mode string, n, c, d int) {
	okCount, failCount, readCount, writeCount, rps, rts := run.OperationRun(n, c, func() (bool, bool, time.Duration) {
		return DoPubOperation(mode, d)
	})
	metrics := utils.CalculateStatistics(rts)
	metrics.RPS = rps
	printSummary(printType, n, c, 0, d, 0, okCount, failCount, readCount, writeCount, metrics)
}

func doSubCase(mode string, c int) {
	var wg sync.WaitGroup
	for i := 1; i <= c; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			Sub(mode, id)
		}(i)
	}
	wg.Wait()
}

func DoSetGetRandomCaseCommand(mode string, n, c, r, d, R, ttl int) {
	doSetGetRandomCase("console", mode, n, c, r, d, R, ttl)
}

func DoSetCaseCommand(mode string, n, c, r, d, R, ttl int) {
	doSetCase("console", mode, n, c, r, d, R, ttl)
}

func DoGetCaseCommand(mode string, n, c int) {
	doGetCase("console", mode, n, c)
}

func DoDelCaseCommand(mode string, n, c, r int) {
	doDelCase("console", mode, n, c, r)
}

func DoPubCaseCommand(mode string, n, c, d int) {
	doPubCase("console", mode, n, c, d)
}

func DoSubCaseCommand(mode string, c int) {
	doSubCase(mode, c)
}
