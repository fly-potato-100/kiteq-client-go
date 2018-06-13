package kiteq_client_go

import (
	"errors"
	"github.com/blackbeans/kiteq-common/protocol"
	log "github.com/blackbeans/log4go"
	"github.com/blackbeans/turbo"
	"time"
)

//握手包
func handshake(ga *turbo.GroupAuth, remoteClient *turbo.TClient) (bool, error) {

	for i := 0; i < 3; i++ {
		p := protocol.MarshalConnMeta(ga.GroupId, ga.SecretKey, int32(ga.WarmingupSec))
		rpacket := turbo.NewPacket(protocol.CMD_CONN_META, p)
		resp, err := remoteClient.WriteAndGet(*rpacket, 5*time.Second)
		if nil != err {
			//两秒后重试
			time.Sleep(2 * time.Second)
			log.WarnLog("kite_client", "kiteClient|handShake|FAIL|%s|%s\n", ga.GroupId, err)
		} else {
			authAck, ok := resp.(*protocol.ConnAuthAck)
			if !ok {
				return false, errors.New("Unmatches Handshake Ack Type! ")
			} else {
				if authAck.GetStatus() {
					log.InfoLog("kite_client", "kiteClient|handShake|SUCC|%s|%s\n", ga.GroupId, authAck.GetFeedback())
					return true, nil
				} else {
					log.WarnLog("kite_client", "kiteClient|handShake|FAIL|%s|%s\n", ga.GroupId, authAck.GetFeedback())
					return false, errors.New("Auth FAIL![" + authAck.GetFeedback() + "]")
				}
			}
		}
	}

	return false, errors.New("handshake fail! [" + remoteClient.RemoteAddr() + "]")
}