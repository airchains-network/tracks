package handler

import (
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
)

const BatchCountKey = "batchCount"

func HandleGetBatchCount(c *gin.Context, Params []any) {
	logger := logrus.New()
	staticDB := shared.Node.NodeConnections.GetStaticDatabaseConnection()
	currentPodNumber, err := staticDB.Get([]byte(BatchCountKey), nil)
	if err != nil {
		logger.WithField("batchCountKey", BatchCountKey).Error("Failed to get current pod number from database: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in getting current pod number from database"})
		return
	}
	currentPodNumberInt, err := strconv.Atoi(strings.TrimSpace(string(currentPodNumber)))
	if err != nil {
		logger.WithField("currentPodNumber", string(currentPodNumber)).Error("Failed to convert current pod number to integer: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error in converting current pod number to integer"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"batchCount": currentPodNumberInt})
}
