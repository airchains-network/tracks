package handler

import (
	"encoding/json"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func NewHandler(logger *logrus.Logger) *Handler {
	return &Handler{Log: logger}
}

type Handler struct {
	Log *logrus.Logger
}

func (h *Handler) getStateData() ([]byte, error) {
	stateConnection := shared.Node.NodeConnections.GetStateDatabaseConnection()
	return stateConnection.Get([]byte("podState"), nil)
}

func (h *Handler) unmarshalPodStateData(data []byte, out *types.PodState) error {
	return json.Unmarshal(data, out)
}

func (h *Handler) HandleGetLatestPod(c *gin.Context) {
	Log := logrus.New()
	podStateData, err := h.getStateData()
	if err != nil {
		h.Log.Error("Error in getting pod state data from database: ", err)
		respondWithError(c, Log, 500, "Internal Server Error", 500)
		return
	}

	var podState types.PodState
	err = h.unmarshalPodStateData(podStateData, &podState)
	if err != nil {
		h.Log.Error("Error in unmarshalling pod state data: ", err)
		respondWithError(c, Log, 500, "Internal Server Error", 500)
		return
	}

	responseData := []interface{}{podState}
	respondWithSuccess(c, Log, responseData, "success")
	return
}
