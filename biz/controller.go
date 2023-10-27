package biz

import (
	"fmt"
	"log"
	"math/rand"
	"time"
	"traffic-control/conf"
)

type TcMessage struct {
	CreateTime time.Duration
	Buf        []byte
	BufLen     int64
	SeqNum     int64
	IsKcp      bool
}

var RecvChan = make(chan TcMessage, conf.Options.QueueCacheLen)
var AdjustTicker *time.Ticker

var rateLimiter Limiter

var lossCount = 0

func execLoss() bool {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num := r.Intn(100)
	return (num % 100) < conf.Options.LossRate
}

func execDelay(ts time.Duration) {
	delayTimeMs := int64(conf.Options.DelayMS) - (int64(time.Now().UnixMilli()) - int64(ts))
	if delayTimeMs > 0 {
		time.Sleep(time.Duration(delayTimeMs) * time.Millisecond)
	}
}

func execLimitRate(msgLen int64) {
	err := rateLimiter.Wait(msgLen)
	if err != nil {
		fmt.Println("err:", err)
	}
}

func controlStrategy(msg TcMessage) bool {
	if conf.Options.OpenTC {
		// 丢包
		if execLoss() {
			lossCount++
			log.Printf("exec loss strategy, count:%d.\n", lossCount)
			return false
		}

		// 延时
		execDelay(msg.CreateTime)

		// 限速
		execLimitRate(msg.BufLen)

		return true
	}
	return true
}

func KbpsToBPS(kbps int64) int64 {
	return kbps * 1000 / 8
}

func autoAdjustBandwidth() {
	go func() {
		rateLimiter, _ = NewJujuLimiter(KbpsToBPS(int64(conf.Options.StartBitrate)))
		defer rateLimiter.Stop()

		log.Printf("Start bandwidth: [%d] kbps.\n", conf.Options.StartBitrate)
		AdjustTicker = time.NewTicker(time.Duration(conf.Options.AutoAdjustBwInterval) * time.Second)
		for {
			select {
			case <-AdjustTicker.C:
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				conf.RealBandwidth = r.Intn(conf.Options.UpperBitrate)
				if conf.RealBandwidth < conf.Options.LowerBitrate {
					conf.RealBandwidth = conf.Options.LowerBitrate
				}
				uploaderSend(uploaderMessage{
					RealBandwidth: conf.RealBandwidth,
					RecvQueueLen:  len(RecvChan),
				})
				err := rateLimiter.UpdateBandwidth(KbpsToBPS(int64(conf.RealBandwidth)))
				if err != nil {
					log.Printf("Update Bandwidth err:%v", err)
				}
				log.Printf("Update bandwidth: [%d] kbps, recv chan len: [%d], send chan len:[%d].\n",
					conf.RealBandwidth, len(RecvChan), len(sendChan))
			}
		}
	}()
}

func updateRecvQueueLen() {
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				uploaderSend(uploaderMessage{
					RealBandwidth: conf.RealBandwidth,
					RecvQueueLen:  len(RecvChan),
				})
			}
		}
	}()
}

func mainLoop() {
	go func() {
		for {
			select {
			case msg := <-RecvChan:
				if controlStrategy(msg) {
					sendChan <- msg
				}
			}
		}
	}()
}

func Start() {
	startUploader()
	updateRecvQueueLen()
	autoAdjustBandwidth()
	relayServerStart()
	mainLoop()
}
