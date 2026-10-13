package tee

import "time"

type KeyBindingPolicy struct {
	Measurement string
	Debug       bool
	Nonce       []byte
	MaxAge      time.Duration
	Now         time.Time
}
