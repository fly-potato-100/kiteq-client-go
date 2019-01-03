package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/blackbeans/kiteq-common/protocol"
	"github.com/blackbeans/kiteq-common/store"
	log "github.com/blackbeans/log4go"
	"github.com/golang/protobuf/proto"
	"io"
	"kiteq-client-go/client"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"kiteq-client-go/benchmark/listener"
)


var body []byte
var rander = rand.Reader // random function
func init() {
	body = make([]byte, 2*1024, 2*1024)
	// randomBits completely fills slice b with random data.
	if _, err := io.ReadFull(rander, body); err != nil {
		panic(err.Error()) // rand should never fail
	}
}

func buildBytesMessage(commit bool) *protocol.BytesMessage {
	//创建消息
	entity := &protocol.BytesMessage{}
	entity.Header = &protocol.Header{
		MessageId:    proto.String(store.MessageId()),
		Topic:        proto.String("user-profile"),
		MessageType:  proto.String("pay-succ"),
		ExpiredTime:  proto.Int64(time.Now().Add(24 * time.Hour).Unix()),
		DeliverLimit: proto.Int32(100),
		GroupId:      proto.String("go-kite-test"),
		Commit:       proto.Bool(commit),
		Fly:          proto.Bool(false),
		CreateTime:   proto.Int64(time.Now().Unix())}

	entity.Body = body

	return entity
}

func buildStringMessage(commit bool) *protocol.StringMessage {
	//创建消息
	entity := &protocol.StringMessage{}
	entity.Header = &protocol.Header{
		MessageId:    proto.String(store.MessageId()),
		Topic:        proto.String("user-profile"),
		MessageType:  proto.String("pay-succ"),
		ExpiredTime:  proto.Int64(-1),
		DeliverLimit: proto.Int32(100),
		GroupId:      proto.String("go-kite-test"),
		Commit:       proto.Bool(commit),
		Fly:          proto.Bool(false),
		CreateTime:   proto.Int64(time.Now().Unix())}

	entity.Body = proto.String("hello world")

	return entity
}

func main() {
	logxml := flag.String("logxml", "log/log_producer.xml", "-logxml=log/log_producer.xml")
	k := flag.Int("k", 1, "-k=1  //kiteclient num ")
	c := flag.Int("c", 1, "-c=100")
	tx := flag.Bool("tx", false, "-tx=true send Tx Message")
	zkhost := flag.String("registryUri", "zk://123.206.17.232:2181", "-registryUri=etcd://http://localhost:2379")
	flag.Parse()

	runtime.GOMAXPROCS(8)

	log.LoadConfiguration(*logxml)

	go func() {

		log.Info(http.ListenAndServe(":28000", nil))
	}()

	count := int32(0)
	lc := int32(0)

	fc := int32(0)
	flc := int32(0)

	go func() {
		for {

			tmp := count
			ftmp := fc

			time.Sleep(1 * time.Second)
			fmt.Printf("tps:%d/%d\n", (tmp - lc), (ftmp - flc))
			lc = tmp
			flc = ftmp
		}
	}()

	wg := &sync.WaitGroup{}
	stop := false
	clients := make([]*client.KiteQClient, 0, *k)
	for j := 0; j < *k; j++ {

		kiteClient := client.NewKiteQClient(*zkhost, "go-kite-test", "123456", &listener.DefaultListener{})
		kiteClient.SetTopics([]string{"user-profile"})
		kiteClient.Start()
		clients = append(clients, kiteClient)
		time.Sleep(3 * time.Second)
		fmt.Printf("Open Client %d\n", j)
		for i := 0; i < *c; i++ {
			go func(kite *client.KiteQClient) {
				wg.Add(1)
				for !stop {
					time.Sleep(1 * time.Second)
					fmt.Printf("Send One Msg.count=%v\n", count)
					if *tx {
						msg := buildBytesMessage(false)
						err := kite.SendTxBytesMessage(msg, doTranscation)
						if nil != err {
							fmt.Printf("SEND TxMESSAGE |FAIL|%s\n", err)
							atomic.AddInt32(&fc, 1)
						} else {
							atomic.AddInt32(&count, 1)
						}
					} else {
						txmsg := buildBytesMessage(true)
						err := kite.SendBytesMessage(txmsg)
						if nil != err {
							fmt.Printf("SEND MESSAGE |FAIL|%s\n", err)
							atomic.AddInt32(&fc, 1)
						} else {
							atomic.AddInt32(&count, 1)
						}
					}
				}
				wg.Done()

			}(kiteClient)
		}

		time.Sleep(10 * time.Second)

		var s = make(chan os.Signal, 1)
		signal.Notify(s, syscall.SIGKILL, syscall.SIGSEGV)
		//是否收到kill的命令
		for {
			cmd := <-s
			if cmd == syscall.SIGKILL {
				break
			} else if cmd == syscall.SIGSEGV {
				//如果为siguser1则进行dump内存
				unixtime := time.Now().Unix()
				path := "./heapdump-producer" + fmt.Sprintf("%d", unixtime)
				f, err := os.Create(path)
				if nil != err {
					continue
				} else {
					debug.WriteHeapDump(f.Fd())
				}
			}
		}

		wg.Wait()

		for _, k := range clients {
			k.Destory()
		}
	}
}

func doTranscation(message *protocol.QMessage) (bool, error) {
	return true, nil
}
