package trackgate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/airchains-network/tracks/config"
	junction2 "github.com/airchains-network/tracks/junction"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/node/shared"
	"github.com/google/uuid"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	//internalCodeTypes "github.com/ComputerKeeda/trackgate-go-client/type"
	"github.com/airchains-network/tracks/junction/trackgate/types"
	trackTypes "github.com/airchains-network/tracks/types"
)

func InitStation(accountName, accountPath, jsonRPC string, bootstrapNode []string, addressPrefix string) bool {

	conf, err := shared.LoadConfig()
	if err != nil {
		logs.Log.Error("Failed to load conf info")
		return false
	}
	ctx := context.Background()

	client, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(addressPrefix), cosmosclient.WithNodeAddress(jsonRPC), cosmosclient.WithHome(accountPath), cosmosclient.WithGas("auto"), cosmosclient.WithFees("1000amf"))
	if err != nil {
		log.Fatal(err)
	}
	// connect junction

	registry, err := cosmosaccount.New(cosmosaccount.WithHome(accountPath))
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error creating account registry: %v", err))
		return false
	}

	newTempAccount, err := registry.GetByName(accountName)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error getting account: %v", err))
		return false
	}

	newTempAddr, err := newTempAccount.Address(addressPrefix)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error getting address: %v", err))
		return false
	}
	logs.Log.Info("tracks address: " + newTempAddr)

	success, amount, err := junction2.CheckBalance(jsonRPC, newTempAddr)
	if err != nil || !success {
		logs.Log.Error(err.Error())
		return false
	}
	if amount < 100 {
		logs.Log.Error("Not enough balance on " + newTempAddr + " to create station")
		return false
	}
	amountDiv, err := strconv.ParseFloat(strconv.FormatInt(amount, 10), 64)
	if err != nil {
		fmt.Println(err)
		return false
	}

	dividedAmount := amountDiv / math.Pow(10, 6)
	dividedAmountStr := strconv.FormatFloat(dividedAmount, 'f', 6, 64)

	logs.Log.Info("Currently user have " + dividedAmountStr + "AMF")

	// addr string, client cosmosclient.Client, ctx context.Context, account cosmosaccount.Account
	stationId := uuid.New().String()
	stationName := "test"

	stationInfo := trackTypes.StationInfoDetails{
		StationName: stationName,
		Type:        conf.Station.StationType,
		FheEnabled:  false,
		Operators:   []string{newTempAddr},
		SequencerDetails: trackTypes.SequencerDetails{
			Name:    conf.Sequencer.SequencerType,
			Version: conf.Sequencer.SequencerVersion,
		},
		DADetails: trackTypes.DADetails{
			Name:    conf.DA.DaName,
			Type:    conf.DA.DaType,
			Version: conf.DA.DaVersion,
		},
		ProverDetails: trackTypes.ProverDetails{
			Name:    conf.Prover.ProverType,
			Version: conf.Prover.ProverVersion,
		},
	}

	stationInfoByte, err := json.Marshal(stationInfo)
	if err != nil {
		log.Fatal(err)
		return false
	}

	msg := &types.MsgInitStation{
		Submitter:   newTempAddr,
		StationId:   stationId,
		StationInfo: stationInfoByte,
		Operators:   []string{newTempAddr},
	}

	// Broadcast a transaction from account `charlie` with the message
	// to create a post store response in txResp
	txResp, err := client.BroadcastTx(ctx, newTempAccount, msg)
	if err != nil {
		fmt.Println("txResp above")
		fmt.Printf("Error broadcasting transaction: %v\n", err)
		fmt.Printf("txResp: %+v\n", txResp)
		fmt.Println("txResp below")
		log.Fatal(err.Error())
		return false
	}

	// Print response from broadcasting a transaction
	fmt.Print("MsgCreatePost:\n\n")
	fmt.Println(txResp)
	timestamp := time.Now().String()
	successGenesis := junction2.CreateGenesisTrackGateJson(newTempAddr, stationId, stationInfo, msg.Operators, txResp.TxHash, timestamp)
	if !successGenesis {
		return false
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.Log.Error("Error in getting home dir path: " + err.Error())
		return false
	}

	ConfigFilePath := filepath.Join(homeDir, config.DefaultTracksDir, config.DefaultConfigDir, config.DefaultConfigFileName)
	bytes, err := os.ReadFile(ConfigFilePath)
	if err != nil {
		logs.Log.Error("Error reading sequencer.toml")
		return false
	}

	//var conf config.Config // JunctionConfig
	if err = toml.Unmarshal(bytes, &conf); err != nil {
		logs.Log.Error(fmt.Sprintf("Error unmarshalling config: %v", err))
		return false
	}

	// Update the values

	conf.P2P.PersistentPeers = bootstrapNode
	conf.Junction.JunctionRPC = jsonRPC
	conf.Junction.JunctionAPI = ""
	conf.Junction.StationId = stationId
	conf.Junction.AddressPrefix = "air"
	conf.Junction.AccountPath = accountPath
	conf.Junction.AccountName = accountName
	conf.Junction.Tracks = msg.Operators

	// Marshal the struct to TOML
	f, err := os.Create(ConfigFilePath)
	if err != nil {
		logs.Log.Error("Error creating file")
		return false
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			logs.Log.Error("Error closing file")
		}
	}(f)
	newData := toml.NewEncoder(f)
	if err := newData.Encode(conf); err != nil {
		logs.Log.Error(fmt.Sprintf("Error encoding config: %v", err))
		return false
	}

	return true

}
