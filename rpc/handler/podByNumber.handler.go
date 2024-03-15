package handler

import (
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/gin-gonic/gin"
	"strconv"
)

func HandleGetPodByNumber(c *gin.Context, Params []any) {

	podNumberHex, ok := Params[0].(string)
	if !ok {
		respondWithError(c, "Value is not a string")
	}

	currentPodNumberInt, err := strconv.ParseInt(podNumberHex, 0, 64)
	batchDB := shared.Node.NodeConnections.GetPodsDatabaseConnection()
	podKey := fmt.Sprintf("pod-%d", currentPodNumberInt)
	podDataByte, err := batchDB.Get([]byte(podKey), nil)
	if err != nil {
		respondWithError(c, "Failed to get pod data")
		return
	}

	podData := &shared.PodState{}
	err = json.Unmarshal(podDataByte, &podData)
	if err != nil {
		respondWithError(c, "Failed to unmarshal pod data")
		return
	}

	var value []any
	value = append(value, podData)
	// podData to any type

	respondWithSuccess(c, value, "success")
	return

}
