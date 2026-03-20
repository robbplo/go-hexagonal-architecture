package system

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

type Clock struct{}

func NewClock() Clock {
	return Clock{}
}

func (Clock) Now() time.Time {
	return time.Now().UTC()
}

type IDGenerator struct{}

func NewIDGenerator() IDGenerator {
	return IDGenerator{}
}

func (IDGenerator) NewID(prefix string) string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		now := time.Now().UnixNano()
		return strings.TrimSpace(prefix) + "_" + strings.ToLower(hex.EncodeToString([]byte(time.Unix(0, now).UTC().Format("20060102150405"))))
	}
	return strings.TrimSpace(prefix) + "_" + hex.EncodeToString(buf)
}

type ApproxTokenCounter struct{}

func NewApproxTokenCounter() ApproxTokenCounter {
	return ApproxTokenCounter{}
}

func (ApproxTokenCounter) CountText(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}
	runes := len([]rune(text))
	tokens := runes / 4
	if runes%4 != 0 {
		tokens++
	}
	if tokens == 0 {
		return 1
	}
	return tokens
}
