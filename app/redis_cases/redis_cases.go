package redisCases

import (
	"log"
	"redis-runner/app/runner"
	"redis-runner/app/utils"
	"sync"
	"time"
)

func Start() {
	var err error
	redisConfigs, err = LoadConfig()
	if err != nil {
		log.Fatalf("redis-config load failed: %v", err)
	}
	ConfigConnect(redisConfigs)
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
		doSetGetRandomCase("logs", n, c, r, d, R, ttl)
	case "set":
		doSetStringCase("logs", n, c, r, d, R, ttl)
	case "get":
		doGetStringCase("logs", n, c)
	case "del":
		doDelCase("logs", n, c, r)
	case "pub":
		doPubCase("logs", n, c, d)
	case "sub":
		SubCaseCommand(mode, c)
	case "default":
	default:
		doGetStringCase("logs", n, c)
	}
}

func SetGetRandomCaseCommand(n, c, r, d, R, ttl int) {
	doSetGetRandomCase("console", n, c, r, d, R, ttl)
}

func SetCaseCommand(n, c, r, d, R, ttl int) {
	doSetStringCase("console", n, c, r, d, R, ttl)
}

func GetCaseCommand(n, c int) {
	doGetStringCase("console", n, c)
}

func DelCaseCommand(n, c, r int) {
	doDelCase("console", n, c, r)
}

func PubCaseCommand(n, c, d int) {
	doPubCase("console", n, c, d)
}

func SubCaseCommand(mode string, c int) {
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

func doSetGetRandomCase(printType string, n, c, r, d, R, ttl int) {
	okCount, failCount, readCount, writeCount, rps, rts := runner.OperationRun(n, c, func() (bool, bool, time.Duration) {
		return DoSetGetStringRandomOperation(n, r, d, R, ttl)
	})
	metrics := utils.CalculateStatistics(rts)
	metrics.RPS = rps
	printSummary(printType, n, c, R, d, ttl, okCount, failCount, readCount, writeCount, metrics)
}

func doSetStringCase(printType string, n, c, r, d, R, ttl int) {
	okCount, failCount, readCount, writeCount, rps, rts := runner.OperationRun(n, c, func() (bool, bool, time.Duration) {
		return DoSetStringOperation(n, r, d, R, ttl)
	})
	metrics := utils.CalculateStatistics(rts)
	metrics.RPS = rps
	printSummary(printType, n, c, R, d, ttl, okCount, failCount, readCount, writeCount, metrics)
}

func doGetStringCase(printType string, n, c int) {
	okCount, failCount, readCount, writeCount, rps, rts := runner.OperationRun(n, c, func() (bool, bool, time.Duration) {
		return DoGetStringOperation()
	})
	metrics := utils.CalculateStatistics(rts)
	metrics.RPS = rps
	printSummary(printType, n, c, 100, 0, 0, okCount, failCount, readCount, writeCount, metrics)
}

func doDelCase(printType string, n, c, r int) {
	okCount, failCount, readCount, writeCount, rps, rts := runner.OperationRun(n, c, func() (bool, bool, time.Duration) {
		return DoDeleteOperation(n, r)
	})
	metrics := utils.CalculateStatistics(rts)
	metrics.RPS = rps
	printSummary(printType, n, c, 0, 0, 0, okCount, failCount, readCount, writeCount, metrics)
}

func doPubCase(printType string, n, c, d int) {
	okCount, failCount, readCount, writeCount, rps, rts := runner.OperationRun(n, c, func() (bool, bool, time.Duration) {
		return DoPubOperation(d)
	})
	metrics := utils.CalculateStatistics(rts)
	metrics.RPS = rps
	printSummary(printType, n, c, 0, d, 0, okCount, failCount, readCount, writeCount, metrics)
}
