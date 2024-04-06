package p2p

import (
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
)

type PeerGenerator struct {
	ListenAddr string
	PingConfig bool
}

func NewPeerGenerator(listenAddr string, pingConfig bool) *PeerGenerator {
	return &PeerGenerator{
		ListenAddr: listenAddr,
		PingConfig: pingConfig,
	}
}

func (pg *PeerGenerator) GeneratePeerID() (peer.ID, error) {
	node, err := libp2p.New(
		libp2p.ListenAddrStrings(pg.ListenAddr),
		libp2p.Ping(pg.PingConfig),
	)
	if err != nil {
		return "", err
	}
	return node.ID(), nil
}
