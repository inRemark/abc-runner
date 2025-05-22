package redisCases

import (
	"math/rand"
	"strings"
	"time"
)

var GenKeys []string
var globalCounter int64

func DoSetGetRandomOperation(mode string, n, r, d, R, ttl int) (isOk, isRead bool, rt time.Duration) {
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	isOk, isRead, rt = true, false, time.Duration(0)
	if rd.Intn(100) < R {
		rt = doRead(mode)
		isRead = true
	} else {
		rt = doWriteString(mode, r, n, d, ttl)
	}
	return isOk, isRead, rt
}

func DoGetOperation(mode string) (isOk, isRead bool, rt time.Duration) {
	rt = doRead(mode)
	return true, true, rt
}

func DoSetOperation(mode string, n, r, d, R, ttl int) (isOk, isRead bool, rt time.Duration) {
	rt = doWriteString(mode, r, n, d, ttl)
	return true, false, rt
}

func DoDeleteOperation(mode string, n, r int) (isOk, isRead bool, rt time.Duration) {
	randomKey := KeysRandom(r, n)
	start := time.Now()
	Del(mode, randomKey)
	rt = time.Since(start)
	return true, false, rt
}

func DoPubOperation(mode string, d int) (isOk, isRead bool, rt time.Duration) {
	value := strings.Repeat("X", d)
	start := time.Now()
	Pub(mode, value)
	rt = time.Since(start)
	return true, false, rt
}

func doWriteString(mode string, r, n, d, ttl int) (rt time.Duration) {
	randomKey := KeysRandom(r, n)
	value := strings.Repeat("X", d)
	start := time.Now()
	Set(mode, randomKey, value, time.Duration(ttl)*time.Second)
	rt = time.Since(start)
	GenKeys = append(GenKeys, randomKey)
	return rt
}

func doRead(mode string) (rt time.Duration) {
	randomKey := KeysRandomFromGenKeys()
	start := time.Now()
	Get(mode, randomKey)
	rt = time.Since(start)
	return rt
}
