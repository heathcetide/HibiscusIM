package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type IdemStore interface {
	Set(key string, ttl time.Duration) bool // return true if set, false if exists
}

type memoryIdemStore struct {
	mu sync.Mutex
	m  map[string]time.Time
}

func newMemoryIdemStore() *memoryIdemStore { return &memoryIdemStore{m: make(map[string]time.Time)} }

func (s *memoryIdemStore) Set(key string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if exp, ok := s.m[key]; ok && exp.After(now) {
		return false
	}
	s.m[key] = now.Add(ttl)
	return true
}

// 清理过期键（可选）
func (s *memoryIdemStore) gc() {
	for {
		time.Sleep(1 * time.Minute)
		now := time.Now()
		s.mu.Lock()
		for k, exp := range s.m {
			if exp.Before(now) {
				delete(s.m, k)
			}
		}
		s.mu.Unlock()
	}
}

type IdempotencyConfig struct {
	HeaderName string        // Idempotency-Key 的请求头名
	TTL        time.Duration // 决定一段时间内重复请求的拒绝窗口
	Store      IdemStore     // 可选外部存储（如 Redis）
}

func IdempotencyMiddleware(cfg IdempotencyConfig) gin.HandlerFunc {
	if cfg.HeaderName == "" {
		cfg.HeaderName = "Idempotency-Key"
	}
	if cfg.TTL <= 0 {
		cfg.TTL = 10 * time.Minute
	}
	store := cfg.Store
	if store == nil {
		mem := newMemoryIdemStore()
		store = mem
		go mem.gc()
	}
	return func(c *gin.Context) {
		key := strings.TrimSpace(c.GetHeader(cfg.HeaderName))
		if key == "" {
			// 兜底以请求体生成哈希作为幂等键
			b, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(strings.NewReader(string(b)))
			h := sha256.Sum256(b)
			key = hex.EncodeToString(h[:])
		}
		if !store.Set(key, cfg.TTL) {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "duplicate request"})
			return
		}
		c.Next()
	}
}
