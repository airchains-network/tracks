package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/rs/zerolog/log"
)

// podStateManager shared.PodStateManager,
func ProcessGossipMessage(node host.Host, ctx context.Context, dataType string, dataByte []byte) {
	fmt.Println("Processing gossip message")
	switch dataType {
	case "proof":
		// podStateManager
		ProofHandler(node, ctx, dataByte)
		return
	case "proofResult":
		ProofResultHandler(node, ctx, dataByte)
		return
	default:
		return
	}
}

// ProofHandler processes the proof received in a P2P message.
func ProofHandler(node host.Host, ctx context.Context, dataByte []byte) {
	var proofData types.ProofData
	if err := json.Unmarshal(dataByte, &proofData); err != nil {
		logs.Log.Info("Error unmarshaling proof: %v")
		return
	}

	currentPodData := shared.GetPodState()
	log.Printf("Current pod data: %+v", currentPodData)

	// check voting for new pod,if yes then update the pod state
	updatePodState(proofData)

	podConnection := shared.Node.NodeConnections.GetPodsDatabaseConnection()
	_ = podConnection

	proofResult := createProofResult(proofData)
	sendProofResult(ctx, node, proofResult)

}

// updatePodState updates the pod's state based on the proof data received.
func updatePodState(proofData types.ProofData) {
	currentPodData := shared.GetPodState()
	currentPodData.LatestPodHeight = 1000000 // Example modification, should be based on actual proof data
	shared.SetPodState(currentPodData)
}

// createProofResult creates a proof result based on the proof data received.
func createProofResult(proofData types.ProofData) types.ProofResult {
	// Logic to determine the success or failure of the proof validation
	return types.ProofResult{
		PodNumber: proofData.PodNumber,
		Success:   true, // This should be determined by actual validation logic
	}
}

// sendProofResult marshals and sends the proof result to the P2P network.
func sendProofResult(ctx context.Context, node host.Host, proofResult types.ProofResult) {
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

func ProofResultHandler(node host.Host, ctx context.Context, dataByte []byte) {

	var proofResult types.ProofResult
	err := json.Unmarshal(dataByte, &proofResult)
	if err != nil {
		panic("error in unmarshling proof result")
	}

	fmt.Printf("Proof result received: %v\n", proofResult)

	// TODO: Handle database as per the proof received

}
