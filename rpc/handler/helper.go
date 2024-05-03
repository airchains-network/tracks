package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

func NewErrorResponse(code int64, message string) *ErrorDetails {
	return &ErrorDetails{
		Code:    code,
		Message: message,
	}
}

// respondWithError sends a JSON error response
func respondWithError(c *gin.Context, logger *logrus.Logger, errorCode int64, errorMsg string, httpCode int) {
	response := ResponseBody{
		JsonRPC: "2.0",
		ID:      "1",
		Error:   NewErrorResponse(errorCode, errorMsg),
	}

	c.JSON(httpCode, response)
}

// respondWithJSON sends a JSON response
func respondWithSuccess(c *gin.Context, logger *logrus.Logger, resData interface{}, description string) {
	response := ResponseBody{
		JsonRPC: "2.0",
		ID:      "1",
		Result:  resData,
	}
	c.JSON(http.StatusOK, response)
}
