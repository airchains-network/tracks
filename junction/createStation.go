package junction

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pelletier/go-toml"
	"math"
	"strconv"

	//"github.com/BurntSushi/toml"
	"github.com/airchains-network/decentralized-sequencer/config"
	junctionTypes "github.com/airchains-network/decentralized-sequencer/junction/types"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/types"
	utilis "github.com/airchains-network/decentralized-sequencer/utils"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	"os"
	"path/filepath"

	"time"
)

type JunctionConfig struct {
	JunctionRPC   string
	JunctionAPI   string
	StationID     string
	VRFPrivateKey string
	VRFPublicKey  string
	AddressPrefix string
	Tracks        []string
}

func CreateStation(extraArg junctionTypes.StationArg, stationId string, stationInfo types.StationInfo, accountName, accountPath, jsonRPC string, verificationKey groth16.VerifyingKey, addressPrefix string, tracks []string, bootstrapNode []string) bool {

	verificationKeyByte, err := json.Marshal(verificationKey)
	if err != nil {
		logs.Log.Error("Failed to unmarshal Verification key" + err.Error())
		return false
	}

	extraArgBytes, err := json.Marshal(extraArg)
	if err != nil {
		logs.Log.Error("Error marshalling extra arg")
		return false
	}

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

	success, amount, err := CheckBalance(jsonRPC, newTempAddr)
	if err != nil || !success {
		logs.Log.Error("Error checking balance")
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
	// Voting powers: almost equal.
	var tracksVotingPower []uint64
	totalPower := uint64(100)
	numTracks := len(tracks)
	// Calculate the equal share for each track
	equalShare := totalPower / uint64(numTracks)
	// Calculate the remainder
	remainder := totalPower % uint64(numTracks)
	// Distribute the equal share to each track
	for i := 0; i < numTracks; i++ {
		if remainder > 0 {
			// For each track, until the remainder is exhausted,
			// add an extra unit of power to make the total sum 100.
			tracksVotingPower = append(tracksVotingPower, equalShare+1)
			remainder-- // Decrement the remainder until it's 0
		} else {
			// Once the remainder is exhausted, append the equal share.
			tracksVotingPower = append(tracksVotingPower, equalShare)
		}
	}

	ctx := context.Background()
	gas := utilis.GenerateRandomWithFavour(611, 1200, [2]int{612, 1000}, 0.7)
	gasFees := fmt.Sprintf("%damf", gas)
	accountClient, err := cosmosclient.New(ctx, cosmosclient.WithAddressPrefix(addressPrefix), cosmosclient.WithNodeAddress(jsonRPC), cosmosclient.WithHome(accountPath), cosmosclient.WithGas("auto"), cosmosclient.WithFees(gasFees))
	if err != nil {
		logs.Log.Error("Error creating account client")
		return false
	}

	stationData := junctionTypes.MsgInitStation{
		Creator:           newTempAddr,
		Tracks:            tracks,
		VerificationKey:   verificationKeyByte,
		StationId:         stationId,
		StationInfo:       stationInfo.StationType,
		TracksVotingPower: tracksVotingPower,
		ExtraArg:          extraArgBytes,
	}

	txResp, err := accountClient.BroadcastTx(ctx, newTempAccount, &stationData)
	if err != nil {
		logs.Log.Error("Error in broadcasting transaction")
		logs.Log.Error(err.Error())
		return false
	}
	logs.Log.Info("tx hash: " + txResp.TxHash)

	timestamp := time.Now().String()
	successGenesis := CreateGenesisJson(stationInfo, verificationKey, stationId, tracks, tracksVotingPower, txResp.TxHash, timestamp, extraArg, newTempAddr)
	if !successGenesis {
		return false
	}

	// create VRF Keys
	vrfPrivateKey, vrfPublicKey := NewKeyPair()
	vrfPrivateKeyHex := vrfPrivateKey.String()
	vrfPublicKeyHex := vrfPublicKey.String()
	if vrfPrivateKeyHex != "" {
		SetVRFPrivKey(vrfPrivateKeyHex)
	} else {
		logs.Log.Error("Error saving VRF private key")
		return false
	}
	if vrfPublicKeyHex != "" {
		SetVRFPubKey(vrfPublicKeyHex)
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

	//var check struct {
	//	Version string `toml:"version"`
	//	RPC     struct {
	//		Laddr                                string        `toml:"laddr"`
	//		CorsAllowedOrigins                   []interface{} `toml:"cors_allowed_origins"`
	//		CorsAllowedMethods                   []string      `toml:"cors_allowed_methods"`
	//		CorsAllowedHeaders                   []string      `toml:"cors_allowed_headers"`
	//		GrpcLaddr                            string        `toml:"grpc_laddr"`
	//		GrpcMaxOpenConnections               int           `toml:"grpc_max_open_connections"`
	//		Unsafe                               bool          `toml:"unsafe"`
	//		MaxOpenConnections                   int           `toml:"max_open_connections"`
	//		MaxSubscriptionClients               int           `toml:"max_subscription_clients"`
	//		MaxSubscriptionsPerClient            int           `toml:"max_subscriptions_per_client"`
	//		ExperimentalSubscriptionBufferSize   int           `toml:"experimental_subscription_buffer_size"`
	//		ExperimentalWebsocketWriteBufferSize int           `toml:"experimental_websocket_write_buffer_size"`
	//		ExperimentalCloseOnSlowClient        bool          `toml:"experimental_close_on_slow_client"`
	//		TimeoutBroadcastTxCommit             string        `toml:"timeout_broadcast_tx_commit"`
	//		MaxBodyBytes                         int           `toml:"max_body_bytes"`
	//		MaxHeaderBytes                       int           `toml:"max_header_bytes"`
	//		TLSCertFile                          string        `toml:"tls_cert_file"`
	//		TLSKeyFile                           string        `toml:"tls_key_file"`
	//		PprofLaddr                           string        `toml:"pprof_laddr"`
	//	} `toml:"rpc"`
	//	P2P struct {
	//		RootDir         string        `toml:"root_dir"`
	//		NodeID          string        `toml:"node_id"`
	//		ListenAddress   string        `toml:"listen_address"`
	//		ExternalAddress string        `toml:"external_address"`
	//		Seeds           string        `toml:"seeds"`
	//		PersistentPeers []interface{} `toml:"persistent_peers"`
	//	} `toml:"p2p"`
	//	Statesync struct {
	//		Enable            bool          `toml:"enable"`
	//		TempDir           string        `toml:"temp_dir"`
	//		RPCServers        []interface{} `toml:"rpc_servers"`
	//		PodTrustPeriod    string        `toml:"pod_trust_period"`
	//		PodTrustHeight    int           `toml:"pod_trust_height"`
	//		PodTrustHash      string        `toml:"pod_trust_hash"`
	//		PodDiscoveryTime  string        `toml:"pod_discovery_time"`
	//		PodRequestTimeout string        `toml:"pod_request_timeout"`
	//		PodChunkFetchers  int           `toml:"pod_chunk_fetchers"`
	//	} `toml:"statesync"`
	//	Consensus struct {
	//		TimeoutPropose        string `toml:"timeout_propose"`
	//		TimeoutProposeDelta   string `toml:"timeout_propose_delta"`
	//		TimeoutPrevote        string `toml:"timeout_prevote"`
	//		TimeoutPrevoteDelta   string `toml:"timeout_prevote_delta"`
	//		TimeoutPrecommit      string `toml:"timeout_precommit"`
	//		TimeoutPrecommitDelta string `toml:"timeout_precommit_delta"`
	//		TimeoutCommit         string `toml:"timeout_commit"`
	//		SkipTimeoutCommit     bool   `toml:"skip_timeout_commit"`
	//		DoubleSignCheckHeight int    `toml:"double_sign_check_height"`
	//	} `toml:"consensus"`
	//	Da struct {
	//		DaType string `toml:"daType"`
	//		DaRPC  string `toml:"daRPC"`
	//		DaKey  string `toml:"daKey"`
	//	} `toml:"da"`
	//	Station struct {
	//		StationType string `toml:"stationType"`
	//		StationRPC  string `toml:"stationRPC"`
	//		StationAPI  string `toml:"stationAPI"`
	//	} `toml:"station"`
	//	Junction struct {
	//		JunctionRPC   string        `toml:"junctionRPC"`
	//		JunctionAPI   string        `toml:"junctionAPI"`
	//		StationID     string        `toml:"stationId"`
	//		VRFPrivateKey string        `toml:"VRFPrivateKey"`
	//		VRFPublicKey  string        `toml:"VRFPublicKey"`
	//		AddressPrefix string        `toml:"AddressPrefix"`
	//		Tracks        []interface{} `toml:"Tracks"`
	//	} `toml:"junction"`
	//}
	//
	//if err = toml.Unmarshal(bytes, &check); err != nil {
	//	logs.Log.Error(fmt.Sprintf("Error unmarshalling config: %v", err))
	//	return false
	//}
	//
	//fmt.Println(check)

	var conf config.Config // JunctionConfig
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
	conf.Junction.Tracks = tracks

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
