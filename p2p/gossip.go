package p2p

import "context"

func ZKPGossip(Message []byte) {

	ctx := context.Background()
	BroadcastMessage(ctx, Node, Message)
}

func VrfGossip() {}
