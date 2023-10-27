package data

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

func NewWsClient(url string) (wsCli *websocket.Conn, err error) {
	dialer := websocket.Dialer{}
	reqHeader := make(http.Header)
	ws, _, err := dialer.Dial(url, reqHeader)
	if err != nil {
		log.Println("Failed to dial gateway, err: ", err)
		return nil, err
	}
	log.Println("New ws client successful.")
	return ws, nil
}

func CloseWsClient(cli *websocket.Conn) {
	if cli != nil {
		if err := cli.Close(); err != nil {
			log.Println("close ws client failed, err: ", err)
		}
	}
}
