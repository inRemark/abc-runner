package redisCases

import (
	"math/rand"
	"strings"
	"time"
)

var GenKeys []string
var globalCounter int64

func DoSetGetStringRandomOperation(n, r, d, R, ttl int) (isOk, isRead bool, rt time.Duration) {
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	isOk, isRead, rt = true, false, time.Duration(0)
	if rd.Intn(100) < R {
		rt = doReadString()
		isRead = true
	} else {
		rt = doWriteString(r, n, d, ttl)
	}
	return isOk, isRead, rt
}

func DoGetStringOperation() (isOk, isRead bool, rt time.Duration) {
	rt = doReadString()
	return true, true, rt
}

func DoSetStringOperation(n, r, d, R, ttl int) (isOk, isRead bool, rt time.Duration) {
	rt = doWriteString(r, n, d, ttl)
	return true, false, rt
}

func DoDeleteOperation(n, r int) (isOk, isRead bool, rt time.Duration) {
	randomKey := KeysRandom(r, n)
	start := time.Now()
	Del(randomKey)
	rt = time.Since(start)
	return true, false, rt
}

func DoPubOperation(d int) (isOk, isRead bool, rt time.Duration) {
	value := strings.Repeat("X", d)
	start := time.Now()
	Pub(value)
	rt = time.Since(start)
	return true, false, rt
}

func doWriteString(r, n, d, ttl int) (rt time.Duration) {
	randomKey := KeysRandom(r, n)
	value := strings.Repeat("X", d)
	start := time.Now()
	Set(randomKey, value, time.Duration(ttl)*time.Second)
	rt = time.Since(start)
	GenKeys = append(GenKeys, randomKey)
	return rt
}

func doReadString() (rt time.Duration) {
	randomKey := KeysRandomFromGenKeys()
	start := time.Now()
	Get(randomKey)
	rt = time.Since(start)
	return rt
}

func doHSet(key string, d int) (rt time.Duration) {
	field := KeysRandomFromGenKeys()
	value := strings.Repeat("X", d)
	start := time.Now()
	HSet(key, field, value)
	rt = time.Since(start)
	return rt
}
