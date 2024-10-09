package types

import (
	"github.com/airchains-network/tracks/junction/junction/types"
)

type StationInfo struct {
	StationType string `json:"stationType"`
	//DaType      string `json:"daType"`
}
type SequencerDetails struct {
	Name    string
	Version string
}

// DADetails represents the Data Availability details of a blockchain component, including its name, type, etc.
/*
{
	Name: mocha4,goldberg,turing....
	Type: celestia,avail,eigen
	Version: git commit version or tag version
}
*/
type DADetails struct {
	Name    string
	Type    string
	Version string
}

// ProverDetails represents metadata about a prover, including its name and version.
/*
	Name: gevulot,sindri,....
	Version: git commit version or tag version
*/
type ProverDetails struct {
	Name    string
	Version string
}

type StationInfoDetails struct {
	StationName      string
	Type             string
	FheEnabled       bool
	Operators        []string // List of node operators or validator identities
	SequencerDetails SequencerDetails
	DADetails        DADetails
	ProverDetails    ProverDetails
}

type GenesisDataType struct {
	StationId          string
	Creator            string
	CreationTime       string
	TxHash             string
	Tracks             []string
	TracksVotingPowers []uint64
	VerificationKey    interface{}
	ExtraArg           types.StationArg
	StationInfo        StationInfo
}

type GenesisTrackGateDataType struct {
	StationId             string
	Submitter             string
	CreationTime          string
	TxHash                string
	Operators             []string
	OperatorsVotingPowers []uint64
	//ExtraArg              types.StationArg
	StationInfo StationInfoDetails
}
