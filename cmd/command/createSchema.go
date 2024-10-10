package command

import (
	"github.com/airchains-network/tracks/junction/trackgate"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/node/shared"
	"github.com/spf13/cobra"
)

var CreateSchema = &cobra.Command{
	Use:   "create-schema",
	Short: "Create schema for Espresso Sequencer",
	Run: func(cmd *cobra.Command, args []string) {

		conf, err := shared.LoadConfig()
		if err != nil {
			logs.Log.Error("Failed to load conf info")
			return
		}
		if conf.Sequencer.SequencerType == "espresso" {
			success := trackgate.SchemaCreation(conf)
			if !success {
				logs.Log.Error("Failed to create new station due to above error")
				return
			}

		}
	}}
