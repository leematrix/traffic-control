package biz

import (
	"context"
	"github.com/juju/ratelimit"
	"sync"
)

type JujuLimiter struct {
	bucket *ratelimit.Bucket
	ctx    context.Context
	cancel context.CancelFunc
	sync.RWMutex
	bucketCapacity int64
}

func NewJujuLimiter(bandwidth int64) (Limiter, error) {
	//fillInterval := 1 * time.Millisecond
	//fillTokenCount := int64(float64(bandwidth / 1000))
	//bucket := ratelimit.NewBucketWithQuantum(fillInterval, bandwidth, fillTokenCount)
	bucket := ratelimit.NewBucketWithRate(float64(bandwidth), bandwidth)
	return &JujuLimiter{
		bucket:         bucket,
		bucketCapacity: bandwidth,
	}, nil
}

func (l *JujuLimiter) Wait(num int64) error {
	l.RLock()
	defer l.RUnlock()
	l.bucket.Wait(num)
	return nil
}

func (l *JujuLimiter) UpdateBandwidth(bandwidth int64) error {
	l.Lock()
	defer l.Unlock()

	l.bucket = ratelimit.NewBucketWithRate(float64(bandwidth), bandwidth)
	l.bucketCapacity = bandwidth
	return nil
}

func (l *JujuLimiter) Stop() {
	l.cancel()
}
