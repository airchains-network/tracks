package handler

import (
	"github.com/airchains-network/decentralized-sequencer/rpc/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

// respondWithError sends a JSON error response
func respondWithError(c *gin.Context, errorMsg string) {
	response := model.ResponseBody{
		JsonRPC: "2.0",
		ID:      "1",
		Error: model.ErrorMsg{
			Code:    500,
			Message: errorMsg,
		},
		Result: nil,
	}

	c.JSON(http.StatusBadRequest, response)
	return
}

// respondWithJSON sends a JSON response
func respondWithSuccess(c *gin.Context, resData []any, description string) {
	response := model.ResponseBody{
		JsonRPC: "2.0",
		ID:      "1",
		Error: model.ErrorMsg{
			Code:    200,
			Message: "",
		},
		Result: resData,
	}
	c.JSON(http.StatusOK, response)
	return
}
