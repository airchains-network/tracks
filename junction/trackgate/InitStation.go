package trackgate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/airchains-network/tracks/config"
	junction2 "github.com/airchains-network/tracks/junction"
	logs "github.com/airchains-network/tracks/log"
	"github.com/google/uuid"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	//internalCodeTypes "github.com/ComputerKeeda/trackgate-go-client/type"
	"github.com/airchains-network/tracks/junction/trackgate/types"
	trackTypes "github.com/airchains-network/tracks/types"
)

func InitStation(accountName, accountPath, jsonRPC string, bootstrapNode []string, conf *config.Config, addressPrefix string) bool {

	ctx := context.Background()

	client, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(addressPrefix), cosmosclient.WithNodeAddress(jsonRPC), cosmosclient.WithHome(accountPath))
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

	randomStationInfoDetails := trackTypes.StationInfoDetails{
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

	randomStationInfoDetailsByte, err := json.Marshal(randomStationInfoDetails)
	if err != nil {
		log.Fatal(err)
		return false
	}

	msg := &types.MsgInitStation{
		Submitter:   newTempAddr,
		StationId:   stationId,
		StationInfo: randomStationInfoDetailsByte,
		Operators:   []string{newTempAddr},
	}

	var OpsVotingPower []uint64
	totalPower := uint64(100)
	numOps := len(msg.Operators)
	// Calculate the equal share for each track
	equalShare := totalPower / uint64(numOps)
	// Calculate the remainder
	remainder := totalPower % uint64(numOps)
	// Distribute the equal share to each track
	for i := 0; i < numOps; i++ {
		if remainder > 0 {
			// For each track, until the remainder is exhausted,
			// add an extra unit of power to make the total sum 100.
			OpsVotingPower = append(OpsVotingPower, equalShare+1)
			remainder-- // Decrement the remainder until it's 0
		} else {
			// Once the remainder is exhausted, append the equal share.
			OpsVotingPower = append(OpsVotingPower, equalShare)
		}
	}

	// Broadcast a transaction from account `charlie` with the message
	// to create a post store response in txResp
	txResp, err := client.BroadcastTx(ctx, newTempAccount, msg)
	if err != nil {
		fmt.Println("txResp above")
		fmt.Println(txResp)
		fmt.Println("txResp below")
		log.Fatal(err.Error())
		return false
	}

	// Print response from broadcasting a transaction
	fmt.Print("MsgCreatePost:\n\n")
	fmt.Println(txResp)
	timestamp := time.Now().String()
	successGenesis := junction2.CreateGenesisTrackGateJson(randomStationInfoDetails, stationId, msg.Operators, OpsVotingPower, txResp.TxHash, timestamp, newTempAddr)
	if !successGenesis {
		return false
	}

	// create VRF Keys
	vrfPrivateKey, vrfPublicKey := junction2.NewKeyPair()
	vrfPrivateKeyHex := vrfPrivateKey.String()
	vrfPublicKeyHex := vrfPublicKey.String()
	if vrfPrivateKeyHex != "" {
		junction2.SetVRFPrivKey(vrfPrivateKeyHex)
	} else {
		logs.Log.Error("Error saving VRF private key")
		return false
	}
	if vrfPublicKeyHex != "" {
		junction2.SetVRFPubKey(vrfPublicKeyHex)
	} else {
		logs.Log.Error("Error saving VRF public key")
		return false
	}
	logs.Log.Info("Successfully Created VRF public and private Keys")

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

	//fmt.Println(conf.P2P)

	// Update the values

	conf.P2P.PersistentPeers = bootstrapNode
	conf.Junction.JunctionRPC = jsonRPC
	conf.Junction.JunctionAPI = ""
	conf.Junction.StationId = stationId
	conf.Junction.VRFPrivateKey = vrfPrivateKeyHex
	conf.Junction.VRFPublicKey = vrfPublicKeyHex
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

	//// Instantiate a query client for your `blog` blockchain
	//queryClient := types.NewQueryClient(client.Context())
	//
	//queryResp, err := queryClient.GetExtTrackStation(ctx, &types.QueryGetExtTrackStationRequest{Id: stationId})
	//if err != nil {
	//	log.Fatal(err)
	//	return false
	//}
	//
	//data := queryResp.Station
	//fmt.Println("Operator: ", data.Operators)
	//fmt.Println("LatestPod: ", data.LatestPod)
	//fmt.Println("LatestMerkleRootHash: ", data.LatestMerkleRootHash)
	//fmt.Println("Name: ", data.Name)
	//fmt.Println("Id: ", data.Id)
	//fmt.Println("StationType: ", data.StationType)
	//fmt.Println("FheEnabled: ", data.FheEnabled)
	//var seqData SequencerDetails
	//seqDataErr := json.Unmarshal(data.SequencerDetails, &seqData)
	//if seqDataErr != nil {
	//	log.Fatal(seqDataErr)
	//	return false
	//}
	//var daData DADetails
	//daDataErr := json.Unmarshal(data.DaDetails, &daData)
	//if daDataErr != nil {
	//	log.Fatal(daDataErr)
	//	return false
	//}
	//var proverData ProverDetails
	//proverDataErr := json.Unmarshal(data.ProverDetails, &proverData)
	//if proverDataErr != nil {
	//	log.Fatal(proverDataErr)
	//	return false
	//}
	//seqDataInd, err := json.MarshalIndent(seqData, "", "    ")
	//if err != nil {
	//	log.Fatalf("Failed to marshal JSON: %v", err)
	//	return false
	//}
	//daDataInd, err := json.MarshalIndent(daData, "", "    ")
	//if err != nil {
	//	log.Fatalf("Failed to marshal JSON: %v", err)
	//	return false
	//}
	//proverDataInd, err := json.MarshalIndent(proverData, "", "    ")
	//if err != nil {
	//	log.Fatalf("Failed to marshal JSON: %v", err)
	//	return false
	//}
	//fmt.Println("SequencerDetails: ", string(seqDataInd))
	//fmt.Println("DaDetails: ", string(daDataInd))
	//fmt.Println("ProverDetails: ", string(proverDataInd))
	/*
	  Operators            []string
	    LatestPod            uint64
	    LatestMerkleRootHash string
	    Name                 string
	    Id                   string
	    StationType          string
	    FheEnabled           bool
	    SequencerDetails     []byte
	    DaDetails            []byte
	    ProverDetails        []byte
	*/

	//queryResp2, err := queryClient.ListExtTrackStations(ctx, &types.QueryListExtTrackStationsRequest{})
	//if err != nil {
	//	log.Fatal(err)
	//}
	//// Use json.MarshalIndent to print with indentation
	//jsonData, err = json.MarshalIndent(queryResp2, "", "    ")
	//if err != nil {
	//	log.Fatalf("Failed to marshal JSON: %v", err)
	//}
	//
	//fmt.Print("\n\nAll posts:\n\n")
	//fmt.Println(string(jsonData))
	return true
}

// Function to generate a random string in the format "stationId-xxxx"
func generateRandomString() (string, string) {
	// Seed the random number generator
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate a random four-digit number
	randomNumber := rand.Intn(9000) + 1000 // Generates a number between 1000 and 9999

	// Concatenate the string parts
	randomStationId := fmt.Sprintf("stationId-%d", randomNumber)
	randomStationInfo := fmt.Sprintf("stationInfo-%d", randomNumber)

	return randomStationId, randomStationInfo
}
