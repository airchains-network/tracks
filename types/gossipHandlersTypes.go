// data types for gossip messages
package types

type ProofData struct {
	PodNumber uint64 `json:"podnumber"`
	Proof     []byte `json:"proof"`
}

type ProofResult struct {
	PodNumber uint64 `json:"podnumber"`
	Success   bool   `json:"success"`
}
