package query

import (
	"fmt"
	"github.com/airchains-network/tracks/junction/trackgate"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/node/shared"
	"github.com/spf13/cobra"
	"strconv"
)

var ListStation = &cobra.Command{
	Use:   "list-station",
	Short: "Station details for a espresso station",
	Run: func(cmd *cobra.Command, args []string) {

		conf, err := shared.LoadConfig()
		if err != nil {
			logs.Log.Error("Failed to load conf info")
			return
		}

		// code for taking flag values

		offsetString := cmd.Flag("offset").Value.String()
		limitString := cmd.Flag("limit").Value.String()
		reverse := cmd.Flag("reverse").Value.String()

		offsetInt, _ := strconv.Atoi(offsetString)
		offset := uint64(offsetInt)
		limitInt, _ := strconv.Atoi(limitString)
		limit := uint64(limitInt)
		reverseBool, _ := strconv.ParseBool(reverse)
		fmt.Println(reverseBool)

		if conf.Sequencer.SequencerType == "espresso" {
			success := trackgate.ListStation(conf, offset, limit, reverseBool)
			if !success {
				logs.Log.Error("Failed to list station due to above error")
				return
			}

		}
	}}
