package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/chennqqi/qqwry"
)

type IPQueryServer struct {
	Port string
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	datFile := flag.String("qqwry", "./qqwry.dat", "IP纯真库路径")
	port := flag.String("port", "8080", "HTTP 请求监听端口号")
	flag.Parse()

	startTime := time.Now()
	err := qqwry.Init(*datFile)
	if err != nil {
		log.Panic(err)
	}
	count, _ := qqwry.Count()
	du := time.Since(startTime)
	log.Println("IP 库加载完成 共加载:", count, " 条 IP 记录, 所花时间:", du)

	var server IPQueryServer
	server.Port = *port
	server.Run()
}

func (s *IPQueryServer) Run() {
	http.HandleFunc("/", s.handleIPQuery)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", s.Port), nil))
}

// 查找 IP 地址的接口
func (s *IPQueryServer) handleIPQuery(w http.ResponseWriter, r *http.Request) {
	ip := r.FormValue("ip")

	if ip == "" {
		http.Error(w, ("please input ip"), http.StatusBadRequest)
		return
	}

	ips := strings.Split(ip, ",")

	qqWry, _ := qqwry.NewQQwry()

	rs := map[string]qqwry.ResultQQwry{}
	if len(ips) > 0 {
		for _, v := range ips {
			q, err := qqWry.Query(v)
			if err != nil {
				continue
			}
			rs[v] = q
		}
	}
	resp, _ := json.MarshalIndent(rs, " ", "\t")
	w.Header().Set("Content-Type", "application/json; charset=UTF8")
	io.WriteString(w, string(resp))
}
