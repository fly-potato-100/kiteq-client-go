package listener

import (
	"time"
	"fmt"
	"github.com/blackbeans/kiteq-common/protocol"
	"sync/atomic"
	"github.com/labstack/gommon/log"
)

type DefaultListener struct {
	count int32
	lc    int32
}

func (self *DefaultListener) Monitor() {
	for {
		tmp := self.count
		ftmp := self.lc

		time.Sleep(1 * time.Second)
		fmt.Printf("tps:%d\n", (tmp - ftmp))
		self.lc = tmp
	}
}



func (self *DefaultListener) OnMessage(msg *protocol.QMessage) bool {
	//	log.Info("DefaultListener|OnMessage|%s", msg.GetHeader().GetMessageId())
	atomic.AddInt32(&self.count, 1)
	return true
}

func (self *DefaultListener) OnMessageCheck(tx *protocol.TxResponse) error {
	log.Info("DefaultListener|OnMessageCheck", tx.MessageId)
	tx.Commit()
	return nil
}
