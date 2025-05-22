package httpCases

import (
	"fmt"
	"redis-runner/app/utils"
	"time"
)

func printSummary(printType string, okCount, failCount, readCount, writeCount, rps int32, rts []time.Duration) {
	metrics := utils.CalculateStatistics(rts)
	metrics.RPS = rps
	endTips := fmt.Sprintf("\nsuccess: %d, error: %d\n", okCount, failCount)
	summary := fmt.Sprintf("\nSummary: \nqps: %v, avg: %.3fms, min: %.3fms, p90: %.3fms, p95: %.3fms, p99: %.3fms, max: %.3fms\n",
		metrics.RPS,
		float64(metrics.Avg)/float64(time.Millisecond),
		float64(metrics.Min)/float64(time.Millisecond),
		float64(metrics.P90)/float64(time.Millisecond),
		float64(metrics.P95)/float64(time.Millisecond),
		float64(metrics.P99)/float64(time.Millisecond),
		float64(metrics.Max)/float64(time.Millisecond))
	end := fmt.Sprintf("Completed.\n")
	arr := []string{endTips, summary, end}

	if printType == "console" {
		utils.PrintConsole(arr)
	} else {
		utils.PrintLogs(arr)
	}
}
