package main

import (
	"io"
	"log"
	"os"
	"traffic-control/biz"
	"traffic-control/conf"
	"traffic-control/server"
)

var file *os.File

func logInit() {
	if !conf.Options.OpenLog {
		return
	}

	var err error
	file, err = os.OpenFile("traffic-control.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}

	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	log.Println("Traffic Control Start.")

	logInit()
	biz.Start()
	server.StartRecvServer()
	server.StartHttpServer()
	log.Println("Traffic Control Exit.")
}
