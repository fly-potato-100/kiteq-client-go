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
	"sync/atomic"
	"syscall"
	"time"

	"github.com/blackbeans/kiteq-common/protocol"
	"github.com/blackbeans/kiteq-common/registry/bind"
	log "github.com/blackbeans/log4go"
	"github.com/blackbeans/kiteq-client-go/client"
)

type defaultListener struct {
	count int32
	lc    int32
}

func (self *defaultListener) monitor() {
	for {
		tmp := self.count
		ftmp := self.lc

		time.Sleep(1 * time.Second)
		fmt.Printf("tps:%d\n", (tmp - ftmp))
		self.lc = tmp
	}
}



func (self *defaultListener) OnMessage(msg *protocol.QMessage) bool {
//	log.Info("defaultListener|OnMessage|%s", msg.GetHeader().GetMessageId())
	atomic.AddInt32(&self.count, 1)
	return true
}

func (self *defaultListener) OnMessageCheck(tx *protocol.TxResponse) error {
	log.Info("defaultListener|OnMessageCheck", tx.MessageId)
	tx.Commit()
	return nil
}

func main() {
	logxml := flag.String("logxml", "../log/log_consumer.xml", "-logxml=../log/log_consumer.xml")
	zkhost := flag.String("registryUri", "etcd://http://localhost:2379", "-registryUri=etcd://http://localhost:2379")
	flag.Parse()
	runtime.GOMAXPROCS(8)

	log.LoadConfiguration(*logxml)
	go func() {

		log.Info(http.ListenAndServe(":38000", nil))
	}()

	lis := &defaultListener{}
	go lis.monitor()

	kite := client.NewKiteQClient(*zkhost, "s-mts-test1", "123456", lis)
	kite.SetBindings([]*bind.Binding{
		bind.Bind_Direct("s-mts-test1", "trade", "pay-succ", 8000, true),
	})
	kite.Start()

	var s = make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGKILL, syscall.SIGUSR1)
	//是否收到kill的命令
	for {
		cmd := <-s
		if cmd == syscall.SIGKILL {
			break
		} else if cmd == syscall.SIGUSR1 {
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
