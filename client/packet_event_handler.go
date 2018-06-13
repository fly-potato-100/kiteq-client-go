package client

import (
	"errors"

	"github.com/blackbeans/kiteq-common/protocol"

	"github.com/blackbeans/turbo"
)

//远程操作的PacketHandler

type PacketHandler struct {
	turbo.BaseForwardHandler
}

func NewPacketHandler(name string) *PacketHandler {
	packetHandler := &PacketHandler{}
	packetHandler.BaseForwardHandler =	turbo. NewBaseForwardHandler(name, packetHandler)
	return packetHandler

}

func (self *PacketHandler) TypeAssert(event turbo.IEvent) bool {
	_, ok := self.cast(event)
	return ok
}

func (self *PacketHandler) cast(event turbo.IEvent) (val *turbo.PacketEvent, ok bool) {
	val, ok = event.(*turbo.PacketEvent)

	return
}

var INVALID_PACKET_ERROR = errors.New("INVALID PACKET ERROR")

func (self *PacketHandler) Process(ctx *turbo.DefaultPipelineContext, event turbo.IEvent) error {

	// log.DebugLog("kite_client_handler","PacketHandler|Process|%s|%t\n", self.GetName(), event)

	pevent, ok := self.cast(event)
	if !ok {
		return turbo.ERROR_INVALID_EVENT_TYPE
	}

	cevent, err := self.handlePacket(pevent)
	if nil != err {
		return err
	}
	ctx.SendForward(cevent)
	return nil
}

var eventSunk = &turbo.SunkEvent{}

//对于请求事件
func (self *PacketHandler) handlePacket(pevent *turbo.PacketEvent) (turbo.IEvent, error) {
	var err error
	var event turbo.IEvent
	packet := pevent.Packet
	//根据类型反解packet
	switch packet.Header.CmdType {
	//连接授权确认
	case protocol.CMD_CONN_AUTH:
		var auth protocol.ConnAuthAck
		err = protocol.UnmarshalPbMessage(packet.Data, &auth)
		if nil == err {
			pevent.RemoteClient.Attach(packet.Header.Opaque, &auth)
			event = &turbo.SunkEvent{}
		}

	//心跳
	case protocol.CMD_HEARTBEAT:
		var hearbeat protocol.HeartBeat
		err = protocol.UnmarshalPbMessage(packet.Data, &hearbeat)
		if nil == err {
			hb := &hearbeat
			// log.DebugLog("kite_client_handler","PacketHandler|handlePacket|HeartBeat|%t\n", hb)
			event = turbo.NewHeartbeatEvent(pevent.RemoteClient, packet.Header.Opaque, hb.GetVersion())
		}
		//消息持久化
	case protocol.CMD_MESSAGE_STORE_ACK:
		var pesisteAck protocol.MessageStoreAck
		err = protocol.UnmarshalPbMessage(packet.Data, &pesisteAck)
		if nil == err {
			//直接通知当前的client对应chan
			pevent.RemoteClient.Attach(packet.Header.Opaque, &pesisteAck)
			event = eventSunk
		}

	case protocol.CMD_TX_ACK:
		var txAck protocol.TxACKPacket
		err = protocol.UnmarshalPbMessage(packet.Data, &txAck)
		if nil == err {
			event = newAcceptEvent(protocol.CMD_TX_ACK, &txAck, pevent.RemoteClient, packet.Header.Opaque)
		}
	//发送的是bytesmessage
	case protocol.CMD_BYTES_MESSAGE:
		var msg protocol.BytesMessage
		err = protocol.UnmarshalPbMessage(packet.Data, &msg)
		if nil == err {
			event = newAcceptEvent(protocol.CMD_BYTES_MESSAGE, &msg, pevent.RemoteClient, packet.Header.Opaque)
		}
	//发送的是StringMessage
	case protocol.CMD_STRING_MESSAGE:
		var msg protocol.StringMessage
		err = protocol.UnmarshalPbMessage(packet.Data, &msg)
		if nil == err {
			event = newAcceptEvent(protocol.CMD_STRING_MESSAGE, &msg, pevent.RemoteClient, packet.Header.Opaque)
		}
	}

	return event, err

}
