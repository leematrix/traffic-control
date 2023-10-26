package biz

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/time/rate"
)

type NativeLimiter struct {
	limiter *rate.Limiter
	ctx     context.Context
	cancel  context.CancelFunc
	sync.Mutex
	bandwidth int64
}

// 令牌桶就是想象有一个固定大小的桶，系统会以恒定速率向桶中放Token，桶满则暂时不放。
// 而用户则从桶中取Token，如果有剩余Token就可以一直取。
// 如果没有剩余Token，则需要等到系统中被放置了Token才行
func NewNativeLimiter(bandwidth int64) (Limiter, error) {
	if bandwidth < 0 {
		return nil, fmt.Errorf("invalid argument bandwidth %d", bandwidth)
	}

	limit := rate.Every(5 * time.Millisecond)
	//limit := rate.Limit(bandwidth)
	bucket := bandwidth
	limiter := rate.NewLimiter(limit, int(bucket))
	ctx, cancel := context.WithCancel(context.Background())
	return &NativeLimiter{
		limiter:   limiter,
		ctx:       ctx,
		cancel:    cancel,
		bandwidth: bucket,
	}, nil
}

func (bl *NativeLimiter) Wait(num int64) error {
	bandwidth := bl.bandwidth
	if bandwidth == 0 || num <= 0 {
		return nil
	}

	bl.Lock()
	defer bl.Unlock()

	if num <= bandwidth {
		log.Printf("limiter wait 01 run, num:%d, bandwidth:%d.\n", num, bandwidth)
		return bl.limiter.WaitN(bl.ctx, int(num))
	} else {
		log.Printf("limiter wait run.\n")
		for i := 0; i < (int(num / bandwidth)); i++ {
			bl.limiter.WaitN(bl.ctx, int(bandwidth))
			log.Printf("limiter wait [%d], bucket: %d.\n", i, bandwidth)
		}
		err := bl.limiter.WaitN(bl.ctx, int(num%bandwidth))
		log.Printf("limiter wait done.\n")
		return err
	}
}

func (bl *NativeLimiter) UpdateBandwidth(bandwidth int64) error {
	if bandwidth < 0 {
		return fmt.Errorf("invalid argument bandwidth %d", bandwidth)
	}

	bl.Lock()
	defer bl.Unlock()

	bl.limiter.SetLimit(rate.Limit(bandwidth))
	bl.limiter.SetBurst(int(bandwidth))
	atomic.StoreInt32((*int32)(unsafe.Pointer(&bl.bandwidth)), int32(bandwidth))
	return nil
}

func (bl *NativeLimiter) Stop() {
	bl.cancel()
}
