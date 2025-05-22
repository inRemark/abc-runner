package httpCases

import (
	"crypto/tls"
	"net/http"
	"time"
)

var client *http.Client

func InitHttpClient(connectTimeout int) {
	timeout := 6
	if connectTimeout > 0 {
		timeout = connectTimeout
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS11,
			MaxVersion: tls.VersionTLS13,
		},
	}

	client = &http.Client{Transport: tr}
	client.Timeout = time.Duration(timeout) * time.Second
}

func GetClient() *http.Client {
	return client
}
