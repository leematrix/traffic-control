package biz

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"time"
	"traffic-control/conf"
	"traffic-control/data"
)

type uploaderMessage struct {
	RealBandwidth int `json:"realBandwidth"`
	RecvQueueLen  int `json:"recvQueueLen"`
}

var uploaderChan = make(chan uploaderMessage, 1024)

func reDialWs(ws *websocket.Conn) {
	if ws != nil {
		data.CloseWsClient(ws)
	}
	for {
		cli, err := data.NewWsClient(conf.Options.StatsServerUrl)
		if err == nil {
			ws = cli
			log.Println("ReDail websocket successful.")
			break
		} else {
			log.Println("ReDail websocket failed, uploader exit.")
			time.Sleep(1 * time.Second)
		}
	}
}

func startUploader() {
	go func() {
		ws, err := data.NewWsClient(conf.Options.StatsServerUrl)
		if err != nil {
			log.Printf("Failed to new ws client [%s], err: %v", conf.Options.StatsServerUrl, err)
			reDialWs(ws)
		}
		defer data.CloseWsClient(ws)

		uploaderSend(uploaderMessage{
			RealBandwidth: conf.RealBandwidth,
			RecvQueueLen:  len(RecvChan),
		})

		for {
			select {
			case msg := <-uploaderChan:
				result, err := json.Marshal(msg)
				if err != nil {
					continue
				}
				if err := ws.WriteMessage(websocket.TextMessage, result); err != nil {
					log.Println("Failed to upload msg to gateway, err: ", err)
					reDialWs(ws)
				} else {
					log.Println("Upload msg to gateway successful")
				}
			}
		}
	}()
}

func uploaderSend(msg uploaderMessage) {
	select {
	case uploaderChan <- msg:
		return
	default:
		log.Println("uploaderChan is full.")
	}
}
