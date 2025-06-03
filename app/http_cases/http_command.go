package httpCases

import (
	"flag"
	"fmt"
	"net/http"
	"redis-runner/app/runner"
	"time"
)

func HttpCommand(args []string) {
	flags := flag.NewFlagSet("http", flag.ExitOnError)
	url := flags.String("url", "http://localhost:8080", "request url")
	method := flags.String("m", "GET", "request method (GET/POST)")
	co := flags.Int("co", 10, "Connection timeout in seconds")
	n := flags.Int("n", 1000, "total number of requests")
	c := flags.Int("c", 10, "number of parallel connections")
	err := flags.Parse(args)
	if err != nil {
		return
	}
	fmt.Printf("Http Info: url: %s, method: %s, n: %d, c: %d\n", *url, *method, *n, *c)

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	bodies := map[string]string{
		"name":    "Hello",
		"bio":     "world",
		"address": "eath",
	}
	InitHttpClient(*co)
	okCount, failCount, readCount, writeCount, rps, rts := runner.OperationRun(*n, *c, func() (bool, bool, time.Duration) {
		if *method == "GET" || *method == "Get" || *method == "get" {
			return HttpRequest(*url, http.MethodGet, headers, bodies)
		} else if *method == "POST" || *method == "Post" || *method == "post" {
			return HttpRequest(*url, http.MethodPost, headers, bodies)
		}
		return false, false, 0
	})
	printSummary("console", okCount, failCount, readCount, writeCount, rps, rts)
}
