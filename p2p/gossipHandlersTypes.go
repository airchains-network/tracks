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

type VRFInitiatedMsg struct {
	PodNumber            uint64
	selectedTrackAddress string
	VrfInitiatorAddress  string
}

type VRFVerifiedMsg struct {
	PodNumber            uint64
	selectedTrackAddress string
}

type PodSubmittedMsgData struct {
	PodNumber            uint64
	selectedTrackAddress string
}

type PodVerifiedMsgData struct {
	PodNumber          uint64
	VerificationResult bool
}
