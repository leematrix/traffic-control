package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
	"traffic-control/biz"
	"traffic-control/conf"
)

var serverIP = "0.0.0.0"
var recvSeqNum int64 = 0
var kcpSeqNum int64 = 0

func recvServerStart() {
	sAddr, err := net.ResolveUDPAddr("udp", serverIP+":"+strconv.Itoa(conf.RecvServerPort))
	if err != nil {
		panic(err)
	}

	go func() {
		conn, err := net.ListenUDP("udp", sAddr)
		if err != nil {
			panic(err)
		}
		defer conn.Close()
		log.Println("Start Recv Server", sAddr)
		for {
			var buf = make([]byte, 1500)
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println(err)
				continue
			}

			select {
			case biz.RecvChan <- biz.TcMessage{
				CreateTime: time.Duration(time.Now().UnixMilli()),
				Buf:        buf,
				BufLen:     int64(n),
				SeqNum:     recvSeqNum,
				IsKcp:      false}:
				{
					log.Printf("Received message [%d], len:%d, recvChan:%d, at %d.\n",
						recvSeqNum, n, len(biz.RecvChan), time.Now().UnixMilli())
					recvSeqNum++
				}
			default:
				log.Printf("recvChan full, len: %d, drop message.\n", len(biz.RecvChan))
				break
			}
		}
	}()
}

var KCPEnabled = false

func kcpRecvServerStart() {
	go func() {
		var bufLen = 1000
		buf := make([]byte, bufLen)

		count := 1
		addFlag := true
		tick := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-tick.C:
				if !KCPEnabled {
					count = 1
					continue
				}

				if count >= 100 {
					addFlag = false
				} else if count <= 1 {
					addFlag = true
				}
				if addFlag {
					count++
				} else {
					count--
				}
				t := time.Duration(time.Now().UnixMilli())
				for i := 0; i < count; i++ {
					select {
					case biz.RecvChan <- biz.TcMessage{
						CreateTime: t,
						Buf:        buf,
						BufLen:     int64(bufLen),
						SeqNum:     kcpSeqNum,
						IsKcp:      true}:
						{
							log.Printf("Received kcp message [%d], len:%d, recvChan:%d, i:%d, count:%d at %v.\n",
								kcpSeqNum, bufLen, len(biz.RecvChan), i, count, t.Milliseconds())
							kcpSeqNum++
						}
					default:
						log.Printf("recvChan full, len: %d, drop kcp message.\n", len(biz.RecvChan))
						break
					}
				}
			}
		}
	}()
}

func StartRecvServer() {
	recvServerStart()
	kcpRecvServerStart()
}
