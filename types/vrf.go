package types

type RequestCommitmentV2Plus struct {
	BlockNum         uint64
	StationId        string
	UpperBound       uint64
	RequesterAddress string
	//ExtraArgs        []byte
}
