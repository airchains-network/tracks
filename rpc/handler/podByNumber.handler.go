package handler

import (
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// assumed the logrus.Logger is defined globally which is a common practice
var Log *logrus.Logger

func HandleGetPodByNumber(c *gin.Context, Params []interface{}) {
	Log := logrus.New()
	batchDB := shared.Node.NodeConnections.GetPodsDatabaseConnection()
	podKey := fmt.Sprintf("pod-%.0f", Params[0])
	fmt.Println(podKey)
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
