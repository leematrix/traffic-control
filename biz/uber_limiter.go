package biz

import (
	"context"
	"go.uber.org/ratelimit"
	"sync"
)

type UberLimiter struct {
	limiter   ratelimit.Limiter
	ctx       context.Context
	cancel    context.CancelFunc
	bandwidth int
	sync.Mutex
}

func NewUberLimiter(bandwidth int64) (Limiter, error) {
	limiter := ratelimit.New(int(bandwidth))
	return &UberLimiter{
		limiter:   limiter,
		bandwidth: int(bandwidth),
	}, nil
}

func (l *UberLimiter) Wait(num int64) error {
	l.Lock()
	defer l.Unlock()

	for i := 0; int64(i) < num; i++ {
		l.limiter.Take()
	}
	return nil
}

func (l *UberLimiter) UpdateBandwidth(bandwidth int64) error {
	l.Lock()
	defer l.Unlock()

	l.limiter = ratelimit.New(int(bandwidth))
	l.bandwidth = int(bandwidth)
	return nil
}

func (l *UberLimiter) Stop() {
	l.cancel()
}
