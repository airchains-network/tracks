package command

import (
	"fmt"
	"github.com/airchains-network/tracks/config"
	logs "github.com/airchains-network/tracks/log"

	//logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/p2p"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

type Configs struct {
	moniker string

	stationType string
	stationRPC  string
	stationAPI  string

	daName    string
	daType    string
	daRPC     string
	daKey     string
	daVersion string

	sequencerType      string
	sequencerRPC       string
	sequencerKey       string
	sequencerVersion   string
	sequencerNamespace string

	proverType    string
	proverRPC     string
	proverVersion string
	proverKey     string
}

func InitConfigs(cmd *cobra.Command) (*Configs, error) {
	var configs Configs
	var err error

	configs.moniker, err = cmd.Flags().GetString("moniker")
	if err != nil {
		return nil, fmt.Errorf("failed to get flag 'moniker': %w", err)
	}

	configs.stationType, err = cmd.Flags().GetString("stationType")
	if err != nil {
		return nil, fmt.Errorf("failed to get flag 'stationType': %w", err)
	}

	configs.daType, err = cmd.Flags().GetString("daType")
	if err != nil {
		return nil, fmt.Errorf("failed to get flag 'daType': %w", err)
	}

	configs.daName, err = cmd.Flags().GetString("daName")
	if err != nil {
		return nil, fmt.Errorf("failed to get flag 'daName': %w", err)
	}

	validTypes := map[string]bool{
		"avail":    true,
		"celestia": true,
		"eigen":    true,
		"mock":     true,
	}
	if _, isValid := validTypes[configs.daType]; !isValid {
		logs.Log.Error("invalid daType. Must be one of: avail, celestia, eigen, mock")
		return nil, fmt.Errorf("invalid daType: %s", configs.daType)
	}

	configs.daRPC, err = cmd.Flags().GetString("daRpc")
	if err != nil {
		return nil, fmt.Errorf("failed to get flag 'daRpc': %w", err)
	}

	configs.daKey, err = cmd.Flags().GetString("daKey")
	if err != nil {
		return nil, fmt.Errorf("failed to get flag 'daKey': %w", err)
	}

	configs.stationRPC, err = cmd.Flags().GetString("stationRpc")
	if err != nil {
		return nil, fmt.Errorf("failed to get flag 'stationRPC': %w", err)
	}

	configs.stationAPI, err = cmd.Flags().GetString("stationAPI")
	if err != nil {
		return nil, fmt.Errorf("failed to get flag 'stationAPI': %w", err)
	}

	configs.sequencerType, err = cmd.Flags().GetString("sequencerType")
	if err != nil {
		return nil, fmt.Errorf("failed to get flag 'sequencerType': %w", err)
	}

	// check the other flags :
	if configs.sequencerType == "espresso" {

		configs.sequencerRPC, err = cmd.Flags().GetString("sequencerRpc")
		if err != nil {
			return nil, fmt.Errorf("failed to get flag 'sequencerRPC': %w", err)
		}

		configs.sequencerNamespace, err = cmd.Flags().GetString("sequencerNamespace")
		if err != nil {
			return nil, fmt.Errorf("failed to get flag 'sequencerNamespace': %w", err)
		}

		//configs.sequencerKey, err = cmd.Flags().GetString("sequencerKey")
		//if err != nil {
		//	return nil, fmt.Errorf("failed to get flag 'sequencerKey': %w", err)
		//}

		//configs.sequencerVersion, err = cmd.Flags().GetString("sequencerVersion")
		//if err != nil {
		//	return nil, fmt.Errorf("failed to get flag 'sequencerVersion': %w", err)
		//}
		//
		//configs.proverType, err = cmd.Flags().GetString("proverType")
		//if err != nil {
		//	return nil, fmt.Errorf("failed to get flag 'proverType': %w", err)
		//}
		//
		//configs.proverRPC, err = cmd.Flags().GetString("proverRPC")
		//if err != nil {
		//	return nil, fmt.Errorf("failed to get flag 'proverRPC': %w", err)
		//}
		//
		//configs.proverType, err = cmd.Flags().GetString("proverVersion")
		//if err != nil {
		//	return nil, fmt.Errorf("failed to get flag 'proverVersion': %w", err)
		//}
		//configs.proverType, err = cmd.Flags().GetString("proverKey")
		//if err != nil {
		//	return nil, fmt.Errorf("failed to get flag 'proverKey': %w", err)
		//}

		//configs.daVersion, err = cmd.Flags().GetString("daVersion")
		//if err != nil {
		//	return nil, fmt.Errorf("failed to get flag 'daVersion': %w", err)
		//}

		// todo: validTypes of sequencerType, proverType, daType && Versions

	}

	return &configs, nil
}

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "initialize the sequencer nodes",
	Run: func(cmd *cobra.Command, args []string) {
		configs, err := InitConfigs(cmd)
		if err != nil {
			logs.Log.Error(err.Error())
			return
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			logs.Log.Error("Failed to get user home directory:" + err.Error())
			return
		}

		tracksDir := filepath.Join(homeDir, config.DefaultTracksDir)

		conf := config.DefaultConfig()
		peerGen := p2p.NewPeerGenerator("/ip4/0.0.0.0/tcp/2300", false)
		peerID, err := peerGen.GeneratePeerID()

		conf.BaseConfig.RootDir = tracksDir
		conf.DA.DaName = configs.daName
		conf.DA.DaType = configs.daType
		conf.DA.DaRPC = configs.daRPC
		conf.DA.DaKey = configs.daKey
		conf.Station.StationType = configs.stationType
		conf.Station.StationRPC = configs.stationRPC
		conf.Station.StationAPI = configs.stationAPI
		conf.P2P.NodeId = peerID
		conf.SetRoot(conf.BaseConfig.RootDir)

		conf.Sequencer.SequencerType = configs.sequencerType
		if conf.Sequencer.SequencerType == "espresso" {
			conf.DA.DaVersion = "v1.0.0"
			conf.DA.DaName = configs.daName
			conf.Sequencer.SequencerKey = "mock"
			conf.Sequencer.SequencerVersion = "v1.0.0"
			conf.Sequencer.SequencerRPC = configs.sequencerRPC
			conf.Sequencer.SequencerNamespace = configs.sequencerNamespace
			conf.Prover.ProverType = "mock"
			conf.Prover.ProverRPC = "mock"
			conf.Prover.ProverVersion = "mock"
			conf.Prover.ProverKey = "mock"

			//conf.DA.DaVersion = configs.daVersion
			//conf.Sequencer.SequencerKey = configs.sequencerKey
			//conf.Sequencer.SequencerVersion = configs.sequencerVersion
			//conf.Sequencer.SequencerRPC = configs.sequencerRPC
			//conf.Prover.ProverType = configs.proverType
			//conf.Prover.ProverRPC = configs.proverRPC
			//conf.Prover.ProverVersion = configs.proverVersion
			//conf.Prover.ProverKey = configs.proverKey
		}

		success := config.CreateConfigFile(conf.BaseConfig.RootDir, conf)
		if !success {
			logs.Log.Error("Unable to generate a config file. Please check the error and try again.")
			return
		}

		logs.Log.Info("Track initialization successful")
	},
}
