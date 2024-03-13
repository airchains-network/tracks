package p2p

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog/log"
)

// podStateManager shared.PodStateManager,
func ProcessGossipMessage(node host.Host, ctx context.Context, dataType string, dataByte []byte) {
	fmt.Println("Processing gossip message")
	switch dataType {
	case "proof":
		// podStateManager
		messageBroadcasterProofHandler(node, ctx, dataByte, messageBroadcaster)
		return
	case "proofResult":
		ProofResultHandler(node, ctx, dataByte, messageBroadcaster)
		return
	//case "finalizationRequest":

	default:
		return
	}
}

// ProofHandler processes the proof received in a P2P message.
func messageBroadcasterProofHandler(node host.Host, ctx context.Context, dataByte []byte, messageBroadcaster peer.ID) {
	var proofData ProofData
	if err := json.Unmarshal(dataByte, &proofData); err != nil {
		logs.Log.Info("Error unmarshaling proof: %v")
		return
	}

	currentPodData := shared.GetPodState()
	receivedTrackAppHash := proofData.TrackAppHash
	receivedPodNumber := proofData.PodNumber

	// match pod numbers
	if currentPodData.LatestPodHeight != receivedPodNumber {
		SendWrongPodNumberError(ctx, receivedPodNumber, messageBroadcaster)
		return
	}

	// match track app hash
	if bytes.Equal(currentPodData.TracksAppHash, receivedTrackAppHash) {
		SendValidProof(ctx, receivedPodNumber, messageBroadcaster)
		return
	} else {
		SendInvalidProofError(ctx, receivedPodNumber, messageBroadcaster)
		return
	}

}

// updatePodState updates the pod's state based on the proof data received.
func updatePodState(proofData ProofData) {
	currentPodData := shared.GetPodState()
	currentPodData.LatestPodHeight = 1000000 // Example modification, should be based on actual proof data
	shared.SetPodState(currentPodData)
}

// createProofResult creates a proof result based on the proof data received.
func createProofResult(proofData ProofData) ProofResult {
	// Logic to determine the success or failure of the proof validation
	return ProofResult{
		PodNumber: proofData.PodNumber,
		Success:   true, // This should be determined by actual validation logic
	}
}

// sendProofResult marshals and sends the proof result to the P2P network.
func sendProofResult(ctx context.Context, node host.Host, proofResult ProofResult) {
	proofResultByte, err := json.Marshal(proofResult)
	if err != nil {
		log.Printf("Error marshaling proof result: %v", err)
		return
	}

	gossipMsg := types.GossipData{
		Type: "proofResult",
		Data: proofResultByte,
	}

	gossipMsgByte, err := json.Marshal(gossipMsg)
	if err != nil {
		log.Printf("Error marshaling gossip message: %v", err)
		return
	}

	log.Printf("Sending proof result: %s", gossipMsgByte)
	BroadcastMessage(ctx, node, gossipMsgByte)
}

// ProofResultHandler processes the proof result received in a P2P message.
func ProofResultHandler(node host.Host, ctx context.Context, dataByte []byte, messageBroadcaster peer.ID) {

	var proofResult ProofResult
	err := json.Unmarshal(dataByte, &proofResult)
	if err != nil {
		panic("error in unmarshling proof result")
	}

	// update pod state votes based on proof result
	updatePodStateVotes(proofResult, messageBroadcaster)

	// count votes of all nodes, if 2/3 votes are true, then
	voteResult, isVotesEnough := calculateVotes()

	// if votes are enough
	if isVotesEnough {
		// if votes are enough and 2/3 votes are true
		if voteResult {
			// TODO SubmitPodToDA()
			// TODO SubmitPodToJunction()

			saveVerifiedPOD()        // save data to database
			GenerateUnverifiedPods() // generate next pod
		} else {
			// TODO: ?????????  what todo if verification failed: discuss with rahul and shubham
		}
	}
	// else: votes are not enough yet, so do nothing....
}
