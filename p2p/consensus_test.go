package p2p_test

import (
	"crypto/sha256"
	"fmt"
	"github.com/libp2p/go-libp2p/core/peer"
	"math/big"
	"testing"
)

// Simulating peerList implementation
type MockPeerList struct {
	peers []peer.ID
}

func (m *MockPeerList) GetPeers() []peer.ID {
	return m.peers
}

// Simulating host.Host object
type MockHost struct {
	id peer.ID
}

// Define the ID who is the Host
func (mh *MockHost) ID() peer.ID {
	return mh.id
}

// Start of function MasterTracksSelection
// Please Merge accordingly with your existing function
func MasterTracksSelection(host *MockHost, sharedInput string, peerList *MockPeerList) string {

	peers := peerList.GetPeers()
	numPeers := len(peers)
	if numPeers == 0 {
		fmt.Println("No peers available.")
		return ""
	}

	// Compute the SHA256 hash of the sharedInput
	h := sha256.New()
	h.Write([]byte(sharedInput))
	hashed := h.Sum(nil)

	hashedInt := new(big.Int)
	hashedInt.SetBytes(hashed)

	randomIndex := hashedInt.Mod(hashedInt, big.NewInt(int64(numPeers)))

	randomPeer := peers[int(randomIndex.Int64())]

	for randomPeer == host.ID() && numPeers > 1 {

		h = sha256.New()
		h.Write([]byte(sharedInput))
		hashed = h.Sum(nil)

		hashedInt = new(big.Int)
		hashedInt.SetBytes(hashed)

		randomIndex = hashedInt.Mod(hashedInt, big.NewInt(int64(numPeers)))
		randomPeer = peers[int(randomIndex.Int64())]
	}

	fmt.Printf("Selected peer ID: %s\n", randomPeer.String())

	return randomPeer.String()
}

func TestMasterTracksSelection(t *testing.T) {
	// Set up mock host and peers
	mockHost := &MockHost{id: "host1"}
	peers := []peer.ID{"peer1", "peer2", "peer3"}
	mockPeerList := &MockPeerList{peers: peers}

	sharedInput := "sharedInputForHash"

	// Call MasterTracksSelection
	selectedPeer := MasterTracksSelection(mockHost, sharedInput, mockPeerList)

	fmt.Println("Selected peer:", selectedPeer)
}

func TestMasterTracksSelectionMultiNode(t *testing.T) {
	// Set up mock hosts and peers
	hosts := []*MockHost{
		{peer.ID("host1")},
		{peer.ID("host2")},
		{peer.ID("host3")},
	}

	peerLists := []*MockPeerList{
		{peers: []peer.ID{"host1", "peer1", "peer2", "host2", "peer3", "peer4", "host3", "peer5"}},
	}

	sharedInput := "sharedInputForHash"

	// Call MasterTracksSelection for each node
	for i := 0; i < len(hosts); i++ {
		fmt.Printf("\nNode %d:\n", i+1)
		selectedPeer := MasterTracksSelection(hosts[i], sharedInput, peerLists[0])
		fmt.Println("Selected peer:", selectedPeer)
	}
}
