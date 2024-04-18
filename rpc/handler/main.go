package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Adding the 'RouterHandler' function explained earlier here:

func RouterHandler(c *gin.Context) {
	Log := logrus.New()

	var requestBody *RequestBody
	err := c.BindJSON(&requestBody)
	if err != nil {
		respondWithError(c, Log, 1, "Invalid Request Body", 500)
		return
	}

	JsonRPC := requestBody.JsonRPC
	ID := requestBody.ID

	if JsonRPC != "2.0" {
		respondWithError(c, Log, 2, "Invalid JsonRPC ID, it should be 2.0", 500)
		return
	}
	if ID != 1 {
		respondWithError(c, Log, 3, "ID should be 1", 500)
		return
	}

	handler := NewHandler(Log)

	switch requestBody.Method {
	case "tracks_getLatestPod":
		handler.HandleGetLatestPod(c)
	case "tracks_batchCount":
		HandleGetBatchCount(c, requestBody.Params) // Assuming this is defined
	case "tracks_getPodByNumber":
		HandleGetPodByNumber(c, requestBody.Params) // Assuming this is defined
	default:
		errorMsg := "No method exists with the name " + requestBody.Method
		respondWithError(c, Log, 4, errorMsg, 404)
	}
}
