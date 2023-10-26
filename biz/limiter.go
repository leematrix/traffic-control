package biz

type Limiter interface {
	Wait(num int64) error
	UpdateBandwidth(num int64) error
	Stop()
}
