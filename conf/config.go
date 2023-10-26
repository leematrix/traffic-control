package conf

type Option struct {
	OpenLog              bool `json:"openLog"`
	QueueCacheLen        int  `json:"queueCacheLen"`
	OpenTC               bool `json:"openTC"`
	LossRate             int  `json:"lossRate"`
	DelayMS              int  `json:"delayMS"`
	StartBitrate         int  `json:"startBitrate"`
	UpperBitrate         int  `json:"upperBitrate"`
	LowerBitrate         int  `json:"lowerBitrate"`
	AutoAdjustBwInterval int  `json:"autoAdjustBwInterval"`
}

var Options = Option{
	OpenLog:              true,
	QueueCacheLen:        1024,
	OpenTC:               true,
	LossRate:             5,
	DelayMS:              20,
	StartBitrate:         6000,
	UpperBitrate:         6000,
	LowerBitrate:         1500,
	AutoAdjustBwInterval: 10,
}

// Non-Configurable Items
var RecvServerPort = 8889
var HttpServerPort = 8099
var RealBandwidth = Options.StartBitrate
