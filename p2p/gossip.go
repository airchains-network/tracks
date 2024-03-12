package p2p

import "context"

//type Gossiper interface {
//	ZKPGossip(proofDataByte []byte)
//}
//
//type P2P struct {
//}
//func (p *P2P)

func ZKPGossip(Message []byte) {
	ctx := context.Background()
	BroadcastMessage(ctx, Node, Message)
}

func VrfGossip() {}
