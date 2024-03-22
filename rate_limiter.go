package main

import "time"

type RateLimiter struct {
	ch *time.Ticker
}

func NewRateLimiter(rps int) *RateLimiter {
	ch := time.NewTicker(time.Duration(rps) * time.Second)
	return &RateLimiter{
		ch: ch,
	}
}

func (r *RateLimiter) allow() bool {
	select {
	case <-r.ch.C:
		return false
	default:
		return true
	}
}
