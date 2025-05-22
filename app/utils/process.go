package utils

import (
	"fmt"
	"sync/atomic"
	"time"
)

func PrintProgress(total int64, progress *int64) {
	ticker := time.NewTicker(time.Second)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			cur := atomic.LoadInt64(progress)
			percent := float64(cur) / float64(total) * 100
			fmt.Printf("\rProgress: %d / %d, %.2f%%", cur, total, percent)
			if cur >= total {
				break
			}
		}
		fmt.Printf("\n")
	}()
}
