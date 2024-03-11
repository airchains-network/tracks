package p2p

import (
	"encoding/json"
	"github.com/airchains-network/decentralized-sequencer/types"
)

func DecodeGossipData(data []byte) (string, []byte, error) {

	var gossipData types.GossipData
	err := json.Unmarshal(data, &gossipData)
	if err != nil {
		return "", nil, err
	}

	return gossipData.Type, gossipData.Data, nil
}
