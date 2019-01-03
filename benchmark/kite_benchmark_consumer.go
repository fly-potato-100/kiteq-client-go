package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/blackbeans/kiteq-common/registry/bind"
	log "github.com/blackbeans/log4go"
	 "kiteq-client-go/client"
	"kiteq-client-go/benchmark/listener"
)



func main() {
	logxml := flag.String("logxml", "log/log_consumer.xml", "-logxml=log/log_consumer.xml")
	zkhost := flag.String("registryUri", "zk://123.206.17.232:2181", "-registryUri=etcd://http://localhost:2379")
	flag.Parse()
	runtime.GOMAXPROCS(8)

	log.LoadConfiguration(*logxml)
	go func() {

		log.Info(http.ListenAndServe(":38000", nil))
	}()

	lis := &listener.DefaultListener{}
	go lis.Monitor()

	kite := client.NewKiteQClient(*zkhost, "s-mts-test1", "123456", lis)
	kite.SetBindings([]*bind.Binding{
		bind.Bind_Direct("s-mts-test1", "user-profile", "pay-succ", 8000, true),
	})
	kite.Start()

	var s = make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGKILL, syscall.SIGABRT)
	//是否收到kill的命令
	for {
		cmd := <-s
		if cmd == syscall.SIGKILL {
			break
		} else if cmd == syscall.SIGABRT {
			//如果为siguser1则进行dump内存
			unixtime := time.Now().Unix()
			path := "./heapdump-consumer" + fmt.Sprintf("%d", unixtime)
			f, err := os.Create(path)
			if nil != err {
				continue
			} else {
				debug.WriteHeapDump(f.Fd())
			}
		}
	}
	kite.Destory()
}
