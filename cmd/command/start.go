package command

import (
	"errors"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/blocksync"
	"github.com/airchains-network/decentralized-sequencer/config"
	logger "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/node"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/spf13/cobra"
)

func initSequencer() error {
	c := config.DefaultConfig()
	fmt.Println(c)
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

	shared.NewNode(c)

	return nil
}

var StationCmd = &cobra.Command{
	Use:   "start",
	Short: "start the sequencer nodes",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initSequencer(); err != nil {
			logger.Log.Error("Error in initiating sequencer nodes")
			logger.Log.Error(err.Error())
			return
		}

		node.Start()
	},
}
