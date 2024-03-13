// data types for gossip messages
package p2p

import (
	"github.com/airchains-network/decentralized-sequencer/node/shared"
)

type ProofData struct {
	PodNumber    uint64 `json:"podnumber"`
	TrackAppHash []byte `json:"proof"`
}

type ProofResult struct {
	PodNumber uint64 `json:"podnumber"`
	Success   bool   `json:"success"`
}

type VoteResult struct {
	TrueCount          int
	FalseCount         int
	TrueVotePercentage float64
	Votes              map[string]shared.Votes
	Success            bool
}
