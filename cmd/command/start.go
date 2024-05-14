package command

import (
	"errors"
	"github.com/airchains-network/decentralized-sequencer/blocksync"
	logger "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/node"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/spf13/cobra"
)

func runSequencerCommand(_ *cobra.Command, _ []string) {
	if err := initSequencer(); err != nil {
		logger.Log.Error(err.Error())
		logger.Log.Error("Error in initiating sequencer nodes due to the above error")
		return
	}
	node.Start()
}

func initSequencer() error {
	config, err := shared.LoadConfig()
	if err != nil {
		logger.Log.Error("Failed to load conf info")
		return err
	}

	if success := blocksync.InitDb(); !success {
		return errors.New("failed to initialize database")
	}

	logger.Log.Info("Database Initialized")

	if config.Junction.StationId == "" {
		return errors.New("create station before stating sequencer")
	}

	if config.Junction.VRFPublicKey == "" || config.Junction.VRFPrivateKey == "" {
		return errors.New("VRF keys not setup properly")
	}

	shared.NewNode(config)

	return nil
}

var StationCmd = &cobra.Command{
	Use:   "start",
	Short: "start the sequencer nodes",
	Run:   runSequencerCommand,
}
