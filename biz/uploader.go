package biz

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
	"traffic-control/conf"
	"traffic-control/data"
)

type UploaderMessage struct {
	RealBandwidth int `json:"realBandwidth"`
	RecvQueueLen  int `json:"recvQueueLen"`
}

var uploaderChan = make(chan UploaderMessage, 1024)

func reDialWs() *websocket.Conn {
	for {
		close(uploaderChan)
		uploaderChan = make(chan UploaderMessage, 1024)
		url := fmt.Sprintf("ws://%s:8090/ws/tc", conf.Options.StatsServerAddr)
		cli, err := data.NewWsClient(url)
		if err == nil {
			log.Println("ReDail websocket successful.")
			return cli
		} else {
			log.Println("ReDail websocket failed, uploader exit.")
			time.Sleep(1 * time.Second)
		}
	}
}

func startUploader() {
	go func() {
		url := fmt.Sprintf("ws://%s:8090/ws/tc", conf.Options.StatsServerAddr)
		ws, err := data.NewWsClient(url)
		if err != nil {
			log.Printf("Failed to new ws client [%s], err: %v", url, err)
			ws = reDialWs()
		}
		defer data.CloseWsClient(ws)

		UploaderSend(UploaderMessage{
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
					data.CloseWsClient(ws)
					ws = reDialWs()
				} else {
					//log.Println("Upload msg to gateway successful, len:", len(uploaderChan))
				}
			}
		}
	}()
}

func UploaderSend(msg UploaderMessage) {
	select {
	case uploaderChan <- msg:
	case <-time.After(1 * time.Millisecond):
		log.Println("uploaderChan is full.")
	}
}
