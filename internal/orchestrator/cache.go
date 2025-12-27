package orchestrator

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

var (
	resultCache = make(map[string]Result)
	mu          sync.RWMutex
)

func GetCache(code string) (Result, bool) {
	hash := generateHash(code)

	mu.Lock()
	val, ok := resultCache[hash]
	mu.Unlock()

	return val, ok
}

func PutCache(code string, res Result) {
	hash := generateHash(code)

	mu.Lock()
	resultCache[hash] = res
	mu.Unlock()
}

func generateHash(key string) string {
	hash := sha256.New()
	hash.Write([]byte(key))

	return hex.EncodeToString(hash.Sum(nil))
}
