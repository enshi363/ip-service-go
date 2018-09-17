package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// IPDB 纯真数据库
var IPDB *QQwry

// RegString 中国城市
var RegString string

func loadChinaCity(f string) {
	var confFile, err = os.Open(f)
	if err != nil {
		panic(err)
	}
	var dat map[string]map[string]string
	decoder := json.NewDecoder(confFile)
	err = decoder.Decode(&dat)
	if err != nil {
		panic(err)
	}
	var s = []string{}
	for _, v := range dat {
		s = append(s, v["province"])
		s = append(s, v["name"])
		// log.Printf("key[%s] value[%s]\n", k, v)
	}
	RegString = strings.Join(s, "|")
	// log.Println()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var datafile = flag.String("c", "./qqwry.dat", "ip数据文件路径也可以是一个url地址")
	var port = flag.String("p", ":8080", "服务端口")
	var chinaCityDataFile = flag.String("cc", "./china_city.json", "中国省市数据")
	var baseURI = flag.String("b", "/", "base uri")
	flag.Parse()
	IPDB = NewQQwry(*datafile)
	IPDB.LoadIPData()
	loadChinaCity(*chinaCityDataFile)
	var s *http.Server

	s = &http.Server{
		Addr:           *port,
		Handler:        HanlderRoutes(*baseURI),
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   7 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	// go func() {

	// }()
	go func() {
		processQuit := make(chan os.Signal, 1)
		signal.Notify(processQuit,
			os.Interrupt,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGKILL,
			syscall.SIGQUIT)
		<-processQuit
		log.Println("Shutdown Server ...")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown:", err)
		}
		log.Println("Server exiting")
	}()
	go func() {
		reloadData := make(chan os.Signal, 1)
		for {
			signal.Notify(reloadData, syscall.SIGUSR1)
			<-reloadData
			log.Println("重新加载ip数据库")
			IPDB.LoadIPData()
			loadChinaCity(*chinaCityDataFile)
			log.Println("加载完毕")
		}
	}()
	s.ListenAndServe()

}
