package infrastructure

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

type RandomIDGenerator struct {
	fallback atomic.Uint64
}

func NewRandomIDGenerator() *RandomIDGenerator {
	return &RandomIDGenerator{}
}

func (generator *RandomIDGenerator) NewID(prefix string) string {
	prefix = strings.Trim(strings.TrimSpace(prefix), "-")
	if prefix == "" {
		prefix = "workflow"
	}
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err == nil {
		return prefix + "-" + hex.EncodeToString(buffer)
	}
	sequence := generator.fallback.Add(1)
	return fmt.Sprintf("%s-%x-%x", prefix, time.Now().UnixNano(), sequence)
}
