package kafkaCases

import (
	"fmt"
	"redis-runner/app/utils"
	"time"
)

func printSummary(printType string, n, c, R, d, ttl int, okCount, failCount, readCount, writeCount int32, metrics utils.ClientMetrics) {
	endTips := fmt.Sprintf("\nAll %d request have completed. \n", n)
	parameters := fmt.Sprintf("Parameters: \nTotal: %d, Parallel: %d, ReadPercent: %d DataSize: %d, TTL: %d\n", n, c, R, d, ttl)
	statistics := fmt.Sprintf("Statistics: \nRead count: %d, Write count: %d\n", readCount, writeCount)
	summary := fmt.Sprintf("Summary: \nrps: %v, avg: %.3fms, min: %.3fms, p90: %.3fms, p95: %.3fms, p99: %.3fms, max: %.3fms\n",
		metrics.RPS,
		float64(metrics.Avg)/float64(time.Millisecond),
		float64(metrics.Min)/float64(time.Millisecond),
		float64(metrics.P90)/float64(time.Millisecond),
		float64(metrics.P95)/float64(time.Millisecond),
		float64(metrics.P99)/float64(time.Millisecond),
		float64(metrics.Max)/float64(time.Millisecond))
	end := fmt.Sprintf("Completed.\n")
	arr := []string{endTips, parameters, statistics, summary, end}
	if printType == "console" {
		utils.PrintConsole(arr)
	} else {
		utils.PrintLogs(arr)
	}
}
