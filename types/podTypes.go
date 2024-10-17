package types

type BatchStruct struct {
	From              []string
	To                []string
	Amounts           []string
	TransactionHash   []string
	SenderBalances    []string
	ReceiverBalances  []string
	Messages          []string
	TransactionNonces []string
	AccountNonces     []string
}

type Votes struct {
	PeerID string // TODO change this type to proper Peer ID Type
	Vote   bool
}
type PodState struct {
	LatestPodHeight     uint64
	LatestTxState       string // InitVRF / VerifyVRF / InitPod / VerifyPod
	LatestPodHash       []byte
	PreviousPodHash     []byte
	LatestPodProof      []byte
	LatestPublicWitness []byte
	Votes               map[string]Votes
	TracksAppHash       []byte
	Batch               *BatchStruct
	MasterTrackAppHash  []byte
}

type TrackgatePodState struct {
	LatestPodHeight uint64
	LatestTxState   string // InitPod / SubmitPod / StorePod
	LatestPodHash   []byte
	PreviousPodHash []byte
	TracksAppHash   []byte
	Batch           *BatchStruct
}
