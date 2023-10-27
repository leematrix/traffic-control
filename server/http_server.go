package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"
	"traffic-control/biz"
	"traffic-control/conf"
)

func getConf(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("static/conf.html")
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w, nil)
}

func setOptions(cfg *conf.Option) {
	conf.Options = *cfg
	close(biz.RecvChan)
	biz.RecvChan = make(chan biz.TcMessage, conf.Options.QueueCacheLen)
	if biz.AdjustTicker != nil {
		biz.AdjustTicker = time.NewTicker(time.Duration(conf.Options.AutoAdjustBwInterval) * time.Second)
	}
	log.Println("Update options:", conf.Options)
}

func setConf(w http.ResponseWriter, r *http.Request) {
	// 读取请求的body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	// 关闭请求的body
	defer r.Body.Close()

	// 打印请求的body
	log.Println("Request body:", string(body))

	cfg := conf.Option{}
	err = json.Unmarshal(body, &cfg)
	if err != nil {
		log.Println("Failed to unmarshal:", err)
		return
	}

	setOptions(&cfg)

	w.WriteHeader(http.StatusOK)
}

func StartHttpServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/conf", getConf)
	mux.HandleFunc("/setConf", setConf)

	s := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", conf.HttpServerPort),
		Handler: mux,
	}

	log.Println("Start Http Server", s.Addr)

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
