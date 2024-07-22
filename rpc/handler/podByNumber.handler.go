package handler

import (
	"encoding/json"
	"fmt"
	"github.com/airchains-network/tracks/node/shared"
	"github.com/airchains-network/tracks/types"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

// assumed the logrus.Logger is defined globally which is a common practice
var Log *logrus.Logger

func HandleGetPodByNumber(c *gin.Context, Params []interface{}) {
	Log := logrus.New()
	batchDB := shared.Node.NodeConnections.GetPodsDatabaseConnection()
	daDB := shared.Node.NodeConnections.GetDataAvailabilityDatabaseConnection()
	podKey := fmt.Sprintf("pod-%.0f", Params[0])
	daKey := fmt.Sprintf("da-%.0f", Params[0])

	fmt.Println(daKey)
	podDataByte, err := batchDB.Get([]byte(podKey), nil)
	if err != nil {
		Log.Error("Failed to get pod data: ", err)
		respondWithError(c, Log, 3, "Failed to get pod data", 500)
		return
	}
	daDataByte, err := daDB.Get([]byte(daKey), nil)
	if err != nil {
		Log.Error("Failed to get pod data: ", err)
		respondWithError(c, Log, 3, "Failed to get da data", 500)
		return
	}

	podData := &shared.PodState{}
	err = json.Unmarshal(podDataByte, &podData)
	if err != nil {
		Log.Error("Failed to unmarshal pod data: ", err)
		respondWithError(c, Log, 4, "Failed to unmarshal pod data", 500)
		return
	}
	daData := &types.DAStruct{}
	err = json.Unmarshal(daDataByte, &daData)
	if err != nil {
		Log.Error("Failed to unmarshal da data: ", err)
		respondWithError(c, Log, 4, "Failed to unmarshal da data", 500)
		return
	}

	var responseData struct {
		LatestPodHeight     uint64
		LatestPodHash       []byte
		PreviousPodHash     []byte
		LatestPodProof      []byte
		LatestPublicWitness []byte
		Votes               map[string]shared.Votes
		TracksAppHash       []byte
		Batch               *types.BatchStruct
		MasterTrackAppHash  []byte
		Timestamp           *time.Time `json:"timestamp,omitempty"`
		DAKey               string
		DAClientName        string
		InitPodTxHash       string
		VerifyPodTxHash     string
		VRFValidationTxHash string
		VRFInitiationTxHash string
	}

	responseData.LatestPodHeight = podData.LatestPodHeight
	responseData.LatestPodHash = podData.LatestPodHash
	responseData.PreviousPodHash = podData.PreviousPodHash
	responseData.LatestPodProof = podData.LatestPodProof
	responseData.LatestPublicWitness = podData.LatestPublicWitness
	responseData.Votes = podData.Votes
	responseData.TracksAppHash = podData.TracksAppHash
	responseData.Batch = podData.Batch
	responseData.MasterTrackAppHash = podData.MasterTrackAppHash
	responseData.Timestamp = podData.Timestamp
	responseData.DAKey = daData.DAKey
	responseData.DAClientName = daData.DAClientName
	responseData.InitPodTxHash = podData.InitPodTxHash
	responseData.VerifyPodTxHash = podData.VerifyPodTxHash
	responseData.VRFValidationTxHash = podData.VRFValidationTxHash
	responseData.VRFInitiationTxHash = podData.VRFInitiationTxHash

	respondWithSuccess(c, Log, responseData, "success")
	return
}
