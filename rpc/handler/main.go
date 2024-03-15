package handler

import (
	"github.com/airchains-network/decentralized-sequencer/rpc/model"
	"github.com/gin-gonic/gin"
)

func RouterHandler(c *gin.Context) {

	// Parse the request body into a struct
	var requestBody *model.RequestBody
	err := c.BindJSON(&requestBody)
	if err != nil {
		respondWithError(c, "Invalid Request Body")
		return
	}

	JsonRPC := requestBody.JsonRPC
	Method := requestBody.Method
	Params := requestBody.Params
	ID := requestBody.ID

	if JsonRPC != "2.0" {
		respondWithError(c, "Invalid JsonRPC ID, it should be 2.0")
		return
	}
	if ID != "1" {
		respondWithError(c, "ID should be 1")
		return
	}

	switch Method {
	case "tracks_getLatestPod":
		HandleGetLatestPod(c, Params)
	case "tracks_batchCount":
		HandleGetBatchCount(c, Params)
	case "tracks_getPodByNumber":
		HandleGetPodByNumber(c, Params)
	default:

		respondWithError(c, "No method exists with name "+Method)
	}

	// POST Request's:-
	//DONE: '{"jsonrpc":"2.0","method":"tracks_getLatestPod","params":[],"id":1}'
	//DONE: '{"jsonrpc":"2.0","method":"tracks_batchCount","params":[],"id":1}'
	//DONE: '{"jsonrpc":"2.0","method":"tracks_getPodByNumber","params":["0x123ab"],"id":1}'

	// ! ???????????????????
	//'{"jsonrpc":"2.0","method":"tracks_getPodMaster","params":["0x2"],"id":1}'
	// '{"jsonrpc":"2.0","method":"tracks_getPodsJunctionDetails","params":["0x2"],"id":1}'
	// '{"jsonrpc":"2.0","method":"tracks_getPodsDataAvailibiltyDetails","params":["0x2"],"id":1}'

}
