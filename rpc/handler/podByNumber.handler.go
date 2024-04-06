package handler

import (
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"strconv"
)

// assumed the logrus.Logger is defined globally which is a common practice
var Log *logrus.Logger

func HandleGetPodByNumber(c *gin.Context, Params []interface{}) {
	Log := logrus.New()
	podNumberHex, ok := Params[0].(string)
	if !ok {
		respondWithError(c, Log, 1, "Value is not a string", 400)
		return
	}

	currentPodNumberInt, err := strconv.ParseInt(podNumberHex, 0, 64)
	if err != nil {
		Log.Error("Error parsing current pod number: ", err)
		respondWithError(c, Log, 2, "Failed to parse pod number", 500)
		return
	}

	batchDB := shared.Node.NodeConnections.GetPodsDatabaseConnection()
	podKey := fmt.Sprintf("pod-%d", currentPodNumberInt)
	podDataByte, err := batchDB.Get([]byte(podKey), nil)
	if err != nil {
		Log.Error("Failed to get pod data: ", err)
		respondWithError(c, Log, 3, "Failed to get pod data", 500)
		return
	}

	podData := &shared.PodState{}
	err = json.Unmarshal(podDataByte, &podData)
	if err != nil {
		Log.Error("Failed to unmarshal pod data: ", err)
		respondWithError(c, Log, 4, "Failed to unmarshal pod data", 500)
		return
	}

	responseData := []interface{}{podData}
	respondWithSuccess(c, Log, responseData, "success")
	return
}
