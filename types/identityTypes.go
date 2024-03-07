package types

import "github.com/libp2p/go-libp2p/core/peer"

type Sequencer struct {
	Moniker     string
	StationType string
	DAType      string
	Identity    peer.ID
	Key         []byte
}
