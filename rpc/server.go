package rpc

import (
	"context"
	"github.com/airchains-network/decentralized-sequencer/rpc/handler"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Server struct {
	httpServer *http.Server
	router     *gin.Engine
	Log        *logrus.Logger
}

func NewServer() *Server {
	server := &Server{
		Log:    logrus.New(),
		router: gin.Default(),
	}

	server.router.POST("/", func(c *gin.Context) {
		handler.RouterHandler(c)
	})

	server.httpServer = &http.Server{
		Addr:    ":2024",
		Handler: server.router,
	}

	return server
}

func (s *Server) Start() error {
	go func() {

		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			s.Log.WithField("error", err).Error("Error starting server")
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.Log.WithField("error", err).Error("Server forced to shutdown")
	}
}

func StartRPC(wg *sync.WaitGroup) {
	defer wg.Done()

	server := NewServer()

	if err := server.Start(); err != nil {
		server.Log.WithField("error", err).Fatal("Failed to start server")
	}

	server.Log.Info("Server Started Successfully on Port 2024")

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGTERM)

	<-stopChan
	server.Log.Info("Received shutdown signal")

}
