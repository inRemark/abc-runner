package redisCases

import (
	"testing"
)

func TestRedisCommand(t *testing.T) {
	mode, h, p, a, db, n, c, d, r, R, ttl :=
		"standalone", "127.0.0.1", 6371, "pwd@redis", 0, 100000, 50, 5, 200000, 100, 300
	CommandConnect(mode, h, a, p, db)
	DoSetGetRandomCaseCommand(mode, n, c, r, d, R, ttl)
}
