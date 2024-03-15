package rpc

import (
	"github.com/airchains-network/decentralized-sequencer/rpc/handler"
	"github.com/gin-gonic/gin"
	"sync"
)

func StartRPC(wg *sync.WaitGroup) {
	defer wg.Done()

	router := gin.Default()
	router.POST("/", func(c *gin.Context) {
		handler.RouterHandler(c)
	})

	// Start serving the application on port 8080
	router.Run(":2024")
}
