// sample POST request:
// {"jsonrpc":"2.0","method":"tracks_getLatestPod","params":[],"id":1}

// Sample Response:
/*
{
	"jsonrpc":"2.0",
	"id":"1",
	"error":{"code":200,"message":""},
	"result":[
		417
	]
}
*/

package handler

import (
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func HandleGetBatchCount(c *gin.Context, Params []any) {

	// currently it have no use
	_ = Params

	staticDB := shared.Node.NodeConnections.GetStaticDatabaseConnection()
	currentPodNumber, err := staticDB.Get([]byte("batchCount"), nil)
	if err != nil {
		respondWithError(c, "Error in getting current pod number from database")
	}
	currentPodNumberInt, _ := strconv.Atoi(strings.TrimSpace(string(currentPodNumber)))
	responseData := []any{currentPodNumberInt}
	respondWithSuccess(c, responseData, "success")
	return

}
