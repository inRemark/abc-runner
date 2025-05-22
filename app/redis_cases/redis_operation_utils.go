package redisCases

import (
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"
)

func KeysRandom(r, n int) string {
	if r == 0 {
		keyNum := atomic.AddInt64(&globalCounter, 1) - 1
		return "i:" + strconv.FormatInt(keyNum, 10)
	}
	// r > 1 rand key in [0, r)
	r = rand.New(rand.NewSource(time.Now().UnixNano())).Intn(r)
	return "r:" + strconv.Itoa(r)
}

func KeysRandomFromGenKeys() string {
	if len(GenKeys) == 0 {
		return "r:0"
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return GenKeys[r.Intn(len(GenKeys))]
}
