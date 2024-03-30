package command

import (
	"errors"
	"github.com/airchains-network/decentralized-sequencer/blocksync"
	logger "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/node"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/spf13/cobra"
)

func initSequencer() error {
	c, err := shared.LoadConfig()
	if err != nil {
		logger.Log.Error("Failed to load conf info")
		return err
	}
	if success := blocksync.InitDb(); !success {
		return errors.New("failed to initialize database")
	}
	logger.Log.Info("Database Initialized")

	if c.Junction.StationId == "" {
		return errors.New("create station before stating sequencer")
	}
	if c.Junction.VRFPublicKey == "" || c.Junction.VRFPrivateKey == "" {
		return errors.New("VRF keys not setup properly")
	}

	shared.NewNode(&c)

	return nil
}

var StationCmd = &cobra.Command{
	Use:   "start",
	Short: "start the sequencer nodes",
	Run: func(cmd *cobra.Command, args []string) {

		if err := initSequencer(); err != nil {
			logger.Log.Error(err.Error())
			logger.Log.Error("Error in initiating sequencer nodes due to above error")
			return
		}

		node.Start()

	},
}
