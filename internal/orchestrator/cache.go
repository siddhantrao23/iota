package orchestrator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
)

var (
	resultCache = make(map[string]Result)
	mu          sync.RWMutex
)

func GetCache(runtime string, args json.RawMessage) (Result, bool) {
	hash := generateHash(runtime, args)

	mu.Lock()
	val, ok := resultCache[hash]
	mu.Unlock()

	return val, ok
}

func PutCache(runtime string, args json.RawMessage, res Result) {
	hash := generateHash(runtime, args)

	mu.Lock()
	resultCache[hash] = res
	mu.Unlock()
}

func generateHash(runtime string, args json.RawMessage) string {
	h := sha256.New()
	h.Write([]byte(runtime))
	h.Write([]byte{0})
	h.Write([]byte(args))

	return hex.EncodeToString(h.Sum(nil))
}
