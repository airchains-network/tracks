package command

import (
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/config"
	logger "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/airchains-network/decentralized-sequencer/p2p"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/spf13/cobra"
	"strconv"
)

var Rollback = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback the sequencer nodes",
	Run:   runRollbackCommand,
}

func runRollbackCommand(_ *cobra.Command, _ []string) {

	// for DB config and connection
	if err := initSequencer(); err != nil {
		logger.Log.Error(err.Error())
		logger.Log.Error("Error in initiating sequencer nodes due to the above error")
		return
	}

	batchDB := shared.Node.NodeConnections.GetPodsDatabaseConnection()
	staticDBConnection := shared.Node.NodeConnections.GetStaticDatabaseConnection()
	stateConnection := shared.Node.NodeConnections.GetStateDatabaseConnection()

	podStateData, err := p2p.GetPodStateFromDatabase()
	if err != nil {
		logger.Log.Error("Error in getting pod state data from database")
		return
	}

	if podStateData.LatestPodHeight < 2 {
		logger.Log.Error("Cannot rollback as there is only one pod")
		return
	}

	processingPodNumber := podStateData.LatestPodHeight
	requiredPodNumberInt := int(processingPodNumber - 1)
	podKey := fmt.Sprintf("pod-%d", requiredPodNumberInt)
	oldPodStateByte, err := batchDB.Get([]byte(podKey), nil)
	if err != nil {
		logger.Log.Error("Error in getting old pod state data from database")
		return
	}
	// unmarshal
	var oldPodStateData types.PodState
	err = json.Unmarshal(oldPodStateByte, &oldPodStateData)
	if err != nil {
		logger.Log.Error("Error in unmarshalling old pod state data")
		return
	}

	err = stateConnection.Put([]byte("podState"), oldPodStateByte, nil)
	if err != nil {
		logger.Log.Error("Error in updating podState in state db")
		return
	}

	err = staticDBConnection.Put([]byte("batchStartIndex"), []byte(strconv.Itoa(config.PODSize*(requiredPodNumberInt))), nil)
	if err != nil {
		logger.Log.Error("Error in updating batchStartIndex in static db")
		return
	}

	err = staticDBConnection.Put([]byte("batchCount"), []byte(strconv.Itoa(requiredPodNumberInt)), nil)
	if err != nil {
		logger.Log.Error("Error in updating batchCount in static db")
		return
	}

	strOld := fmt.Sprintf("%d", requiredPodNumberInt)
	strNew := fmt.Sprintf("%d", processingPodNumber)
	Msg := fmt.Sprintf("Rollback Successfull. Pod: %s -> %s", strNew, strOld)
	logger.Log.Info(Msg)
}
