package query

import (
	"github.com/airchains-network/tracks/junction/trackgate"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/node/shared"
	"github.com/spf13/cobra"
	"strconv"
)

var ListStationEngagements = &cobra.Command{
	Use:   "list-station-engagements",
	Short: "List station engagements for a espresso station",
	Run: func(cmd *cobra.Command, args []string) {

		conf, err := shared.LoadConfig()
		if err != nil {
			logs.Log.Error("Failed to load conf info")
			return
		}

		// code for taking flag values

		offsetString := cmd.Flag("offset").Value.String()
		limitString := cmd.Flag("limit").Value.String()
		order := cmd.Flag("order").Value.String()

		offsetInt, _ := strconv.Atoi(offsetString)
		offset := uint64(offsetInt)
		limitInt, _ := strconv.Atoi(limitString)
		limit := uint64(limitInt)

		if conf.Sequencer.SequencerType == "espresso" {
			success := trackgate.ListEngagements(conf, order, offset, limit)
			if !success {
				logs.Log.Error("Failed to list engagements due to above error")
				return
			}

		}
	}}
