package httpCases

import (
	"net/http"
	"redis-runner/app/runner"
	"testing"
	"time"
)

func initHttpClient() {
	InitHttpClient(10)
}

func TestHttpRunGet(t *testing.T) {
	url := "https://bing.com"
	total := 10
	parallels := 2
	initHttpClient()
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	okCount, failCount, readCount, writeCount, rps, rts := runner.OperationRun(total, parallels, func() (bool, bool, time.Duration) {
		return HttpRequest(url, http.MethodGet, headers, nil)
	})

	printSummary("console", okCount, failCount, readCount, writeCount, rps, rts)
}

func TestHttpRunPost(t *testing.T) {
	url := "https://fr-if2.muji.com.cn:943/Maps_Fr_CouPonCodes_test/TinyFuncServlet"
	total := 1
	parallels := 1
	initHttpClient()
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	bodies := map[string]string{
		"receiver":     "Name",
		"receiverType": "Email",
		"businessType": "LOGIN",
	}
	okCount, failCount, readCount, writeCout, rps, rts := runner.OperationRun(total, parallels, func() (bool, bool, time.Duration) {
		return HttpRequest(url, http.MethodPost, headers, bodies)
	})

	printSummary("logs", okCount, failCount, readCount, writeCout, rps, rts)
}

func TestHttpsPost(t *testing.T) {
	url := "https://fr-if2.muji.com.cn:943/Maps_Fr_CouPonCodes_test/TinyFuncServlet"
	initHttpClient()
	headers := map[string]string{}
	bodies := map[string]string{}
	HttpRequest(url, http.MethodPost, headers, bodies)

}
