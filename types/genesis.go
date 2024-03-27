package types

import (
	"github.com/airchains-network/decentralized-sequencer/junction/types"
)

type StationInfo struct {
	StationType string `json:"stationType"`
	//DaType      string `json:"daType"`
}

type GenesisDataType struct {
	StationId          string           `json:"stationId"`
	Creator            string           `json:"creator"`
	CreationTime       string           `json:"creationTime"`
	TxHash             string           `json:"txHash"`
	Tracks             []string         `json:"tracks"`
	TracksVotingPowers []uint64         `json:"tracksVotingPowers"`
	VerificationKey    interface{}      `json:"verificationKey"`
	ExtraArg           types.StationArg `json:"extraArg"`
	StationInfo        StationInfo      `json:"stationInfo"`
}
