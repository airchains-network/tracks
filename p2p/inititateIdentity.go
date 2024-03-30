package p2p

import (
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
)

func GeneratePeerID() peer.ID {
	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/2300"),
		libp2p.Ping(false),
	)
	if err != nil {
		panic(err)
	}
	publicID := node.ID()
	return publicID
}
