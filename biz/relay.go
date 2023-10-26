package biz

import (
	"log"
	"net"
	"strconv"
)

var RelayIP = "127.0.0.1"
var RelayPort = 8887
var KcpRelayPort = 18887
var sendChan = make(chan TcMessage, 1024)

func relayServerStart() {
	addr := RelayIP + ":" + strconv.Itoa(RelayPort)
	kcpAddr := RelayIP + ":" + strconv.Itoa(KcpRelayPort)
	go func() {
		conn, err := net.Dial("udp", addr)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		kcpConn, err := net.Dial("udp", kcpAddr)
		if err != nil {
			panic(err)
		}
		defer kcpConn.Close()

		for {
			select {
			case msg := <-sendChan:
				if msg.IsKcp {
					log.Printf("Sent kcp message [%d] to %v, msgLen: %d, chanLen: %d\n",
						msg.SeqNum, addr, msg.BufLen, len(sendChan))
				} else {
					_, err := conn.Write(msg.Buf[:msg.BufLen])
					if err != nil {
						log.Printf("send msg err:%v", err)
						log.Println("clear recv and send channel.")
						flag := false
						for {
							select {
							case <-RecvChan:
							case <-sendChan:
							default:
								flag = true
							}
							if flag {
								break
							}
						}
						log.Println("ReDial addr:", addr)
						conn, _ = net.Dial("udp", addr)
					}
					log.Printf("Sent message [%d] to %v, msgLen: %d, chanLen: %d\n",
						msg.SeqNum, addr, msg.BufLen, len(sendChan))
				}
			}
		}
	}()
}
