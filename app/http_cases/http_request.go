package httpCases

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func doGet(url string, headers, bodies map[string]string) (isOk, isRead bool, rt time.Duration) {
	isOk, isRead, rt = HttpRequest(url, http.MethodGet, headers, bodies)
	return isOk, isRead, rt
}

func doPost(url string, headers, bodies map[string]string) (isOk, isRead bool, rt time.Duration) {
	isOk, isRead, rt = HttpRequest(url, http.MethodPost, headers, bodies)
	return isOk, isRead, rt
}

func HttpRequest(url, method string, headers, bodies map[string]string) (isOk, isRead bool, rt time.Duration) {
	req := makeReq(url, method, headers, bodies)
	if req.Method == "GET" {
		isRead = true
	} else {
		isRead = false
	}
	isOk, rt = doRequest(req)
	return isOk, isRead, rt
}

func makeReq(url, method string, headers, bodies map[string]string) (req *http.Request) {
	jsonByte, err := json.Marshal(bodies)
	if err != nil {
		log.Printf("json.Marshal() failed with '%s'\n", err)
		return nil
	}
	body := bytes.NewBuffer(jsonByte)
	req, err = http.NewRequest(method, url, body)
	if err != nil {
		log.Printf("http.NewRequest() failed with '%s'\n", err)
		return nil
	}
	for key, value := range headers {
		fmt.Printf("%s -> %s\n", key, value)
		req.Header.Set(key, value)
	}
	return req
}

func doRequest(req *http.Request) (isOk bool, rt time.Duration) {
	isOk = false
	rt = 0.0
	start := time.Now()
	resp, err := GetClient().Do(req)
	rt = time.Since(start)
	fmt.Printf("Response Status: %v\n", resp)
	if err != nil {
		fmt.Printf("client.Do() failed with '%s'\n", err)
		return isOk, rt
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Body.Close() failed with '%s'\n", err)
		}
	}(resp.Body)
	log.Printf("Response Status: %v\n", resp)
	fmt.Printf("Response Status: %v\n", resp)
	if resp.StatusCode == http.StatusOK {
		isOk = true
	}
	return isOk, rt
}
