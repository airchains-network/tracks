package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/libp2p/go-libp2p/core/host"
)

func ProcessGossipMessage(node host.Host, ctx context.Context, dataType string, dataByte []byte) {
	fmt.Println("Processing gossip message")
	switch dataType {
	case "proof":
		ProofHandler(node, ctx, dataByte)
		return
	case "proofResult":
		ProofResultHandler(node, ctx, dataByte)
		return
	default:
		return
	}
}

func ProofHandler(node host.Host, ctx context.Context, dataByte []byte) {

	var ProofData types.ProofData
	err := json.Unmarshal(dataByte, &ProofData)
	if err != nil {
		panic("error in unmarshling proof")
	}

	//ReceivedProof := ProofData.Proof
	//podNumber := ProofData.PodNumber

	// Proof Local Pod
	//LocalGeneratedProof := pods.LatestUnverifiedValue.ValueUnverifiedProof()

	//proofResult := types.ProofResult{}
	//if bytes.Equal(ReceivedProof, LocalGeneratedProof) {
	//	proofResult = types.ProofResult{
	//		PodNumber: ProofData.PodNumber,
	//		Success:   true,
	//	}
	//} else {
	//	proofResult = types.ProofResult{
	//		PodNumber: ProofData.PodNumber,
	//		Success:   false,
	//	}
	//}
	//ProofResult
	// marshal proof result
	proofResultByte, err := json.Marshal(proofResult)
	if err != nil {
		panic("error in mashaling proof result")
	}

	gossipMsg := types.GossipData{
		Type: "proofResult",
		Data: proofResultByte,
	}

	gossipMsgByte, err := json.Marshal(gossipMsg)
	if err != nil {
		panic("error in marshing proof result")
	}

	fmt.Printf("Sending proof result to %s\n", gossipMsgByte)

	BroadcastMessage(ctx, node, gossipMsgByte)
	return
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
