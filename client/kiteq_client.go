package client

import (
	"github.com/blackbeans/kiteq-client-go/client/core"
	"github.com/blackbeans/kiteq-client-go/client/listener"
	"github.com/blackbeans/kiteq-common/protocol"
	"github.com/blackbeans/kiteq-common/registry/bind"
)

type KiteQClient struct {
	kclientManager *core.KiteClientManager
}

func (self *KiteQClient) Start() {
	self.kclientManager.Start()
}

func NewKiteQClient(zkAddr, groupId, secretKey string, listener listener.IListener) *KiteQClient {
	return &KiteQClient{
		kclientManager: core.NewKiteClientManager(zkAddr, groupId, secretKey, listener)}
}

func (self *KiteQClient) SetTopics(topics []string) {
	self.kclientManager.SetPublishTopics(topics)
}

func (self *KiteQClient) SetBindings(bindings []*bind.Binding) {
	self.kclientManager.SetBindings(bindings)

}

func (self *KiteQClient) SendTxStringMessage(msg *protocol.StringMessage, transcation core.DoTranscation) error {
	message := protocol.NewQMessage(msg)
	return self.kclientManager.SendTxMessage(message, transcation)
}

func (self *KiteQClient) SendTxBytesMessage(msg *protocol.BytesMessage, transcation core.DoTranscation) error {
	message := protocol.NewQMessage(msg)
	return self.kclientManager.SendTxMessage(message, transcation)
}

func (self *KiteQClient) SendStringMessage(msg *protocol.StringMessage) error {
	message := protocol.NewQMessage(msg)
	return self.kclientManager.SendMessage(message)
}

func (self *KiteQClient) SendBytesMessage(msg *protocol.BytesMessage) error {
	message := protocol.NewQMessage(msg)
	return self.kclientManager.SendMessage(message)
}

func (self *KiteQClient) Destory() {
	self.kclientManager.Destory()
}
