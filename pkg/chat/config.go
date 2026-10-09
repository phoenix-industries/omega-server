package chat

import (
	"time"

	"golang.org/x/time/rate"
)

type Config struct {
	MaxMessagesBuffer int
	MessageLimiter    *rate.Limiter
	writeTimeout      time.Duration
	connectTimeout    time.Duration
	pingInterval      time.Duration
}

func NewConfig(maxBuf int, limiter *rate.Limiter) *Config {
	return &Config{
		MaxMessagesBuffer: maxBuf,
		MessageLimiter:    limiter,
	}
}

func DefaultConfig() *Config {
	return NewConfig(16, rate.NewLimiter(rate.Every(time.Millisecond*100), 8))
}
