// In shared/core package
package shared

type Votes struct {
	PeerID     string // TODO change this type to proper Peer ID Type
	Commitment string // this is the hash of the commitment that the peer has voted for the pod and the Signaturwe will done using thePrivate Keys of the Peers
	Vote       bool
}
type PodState struct {
	LatestPodHeight         int
	LatestPodMerkleRootHash []byte
	LatestPodProof          []byte
	LatestPublicWitness     []byte
	Votes                   map[string]Votes
}

type PodStateManager interface {
	UpdatePodState(podState PodState) error
	GetPodState() PodState
}
