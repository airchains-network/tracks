package p2p

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/config"
	"github.com/airchains-network/decentralized-sequencer/da/avail"
	"github.com/airchains-network/decentralized-sequencer/da/celestia"
	"github.com/airchains-network/decentralized-sequencer/da/eigen"
	mock "github.com/airchains-network/decentralized-sequencer/da/mockda"
	"github.com/airchains-network/decentralized-sequencer/junction"
	junctionTypes "github.com/airchains-network/decentralized-sequencer/junction/types"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/airchains-network/decentralized-sequencer/types"
	utilis "github.com/airchains-network/decentralized-sequencer/utils"
	v1 "github.com/airchains-network/decentralized-sequencer/zk/v1EVM"
	v1Wasm "github.com/airchains-network/decentralized-sequencer/zk/v1WASM"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/syndtr/goleveldb/leveldb"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func BatchGeneration(wg *sync.WaitGroup) {
	defer wg.Done()

	GenerateUnverifiedPods()
}

const (
	Mock               = "mock"
	Avail              = "avail"
	Celestia           = "celestia"
	Eigen              = "eigen"
	BatchCountKey      = "batchCount"
	BatchStartIndexKey = "batchStartIndex"
)

func checkErrorAndExit(err error, message string, exitCode int) {
	if err != nil {
		log.Log().Str("module", "p2p").Msg("error: " + message)
		os.Exit(exitCode)
	}
}

func getValueOrDefault(db *leveldb.DB, key []byte, defaultValue []byte) ([]byte, error) {
	val, err := db.Get(key, nil)
	log := logs.Log
	if err != nil {
		log.Warn(fmt.Sprintf("%s not found in static db", string(key)))
		err = db.Put(key, defaultValue, nil)
		checkErrorAndExit(err, fmt.Sprintf("Error in saving %s in static db", string(key)), 0)
	}
	return val, nil
}

func GenerateUnverifiedPods() {
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().
		Str("module", "p2p").
		Msg("Generating New unverified pods")

	connection := shared.Node.NodeConnections
	staticDBConnection := connection.GetStaticDatabaseConnection()
	txnDBConnection := connection.GetTxnDatabaseConnection()

	rawConfirmedTransactionIndex, err := getValueOrDefault(staticDBConnection, []byte(BatchStartIndexKey), []byte("0"))
	checkErrorAndExit(err, "Error in getting confirmedTransactionIndex from static db", 0)

	rawCurrentPodNumber, err := getValueOrDefault(staticDBConnection, []byte(BatchCountKey), []byte("0"))
	checkErrorAndExit(err, "Error in getting currentPodNumber from static db", 0)

	previousStateData, err := getPodStateFromDatabase()
	checkErrorAndExit(err, "Error in getting previous station data", 0)

	previousTrackAppHash := previousStateData.TracksAppHash
	if previousTrackAppHash == nil {
		previousTrackAppHash = []byte("nil")
	}

	selectedMaster := MasterTracksSelection(Node, string(previousTrackAppHash))

	decodedMaster, err := peer.Decode(selectedMaster)
	checkErrorAndExit(err, "Error in decoding master", 0)

	currentPodNumber, _ := strconv.Atoi(strings.TrimSpace(string(rawCurrentPodNumber)))
	batchNumber := currentPodNumber + 1

	baseCfg, err := shared.LoadConfig()
	if err != nil {
		log.Error().Str("module", "p2p").Msg("Error in loading config")
	}
	stationVariant := baseCfg.Station.StationType
	stationVariantLowerCase := strings.ToLower(stationVariant)
	var witness []byte
	var uZKP []byte
	var mRH []byte
	var batchInput *types.BatchStruct
	if stationVariantLowerCase == "evm" {
		witness, uZKP, mRH, batchInput, err = createEVMPOD(txnDBConnection, rawConfirmedTransactionIndex, rawCurrentPodNumber)
		checkErrorAndExit(err, "Error in creating POD", 0)
	} else if stationVariantLowerCase == "wasm" {
		witness, uZKP, mRH, batchInput, err = createWasmPOD(txnDBConnection, rawConfirmedTransactionIndex, rawCurrentPodNumber)
		checkErrorAndExit(err, "Error in creating POD", 0)
	}

	trackAppHash := generatePodHash(witness, uZKP, mRH, rawCurrentPodNumber)
	updateNewPodState(trackAppHash, witness, uZKP, mRH, uint64(batchNumber), batchInput)

	if decodedMaster == Node.ID() {
		podState := shared.GetPodState()
		currentVotes := podState.Votes
		currentVotes[decodedMaster.String()] = shared.Votes{
			PeerID: decodedMaster.String(),
			Vote:   true,
		}
		podState.Votes = currentVotes
		shared.SetPodState(podState)
		Peers := getAllPeers(Node)
		peerCount := len(Peers)
		if peerCount == 1 {
			DaData := shared.GetPodState().Batch.TransactionHash
			var daDataByte []byte
			for _, str := range DaData {
				daDataByte = append(daDataByte, []byte(str)...)
			}
			ZkProof := shared.GetPodState().LatestPodProof
			PodNumber := int(shared.GetPodState().LatestPodHeight)

			finalizeDA := types.FinalizeDA{
				CompressedHash: DaData,
				Proof:          ZkProof,
				PodNumber:      PodNumber,
			}
			_, err := json.Marshal(finalizeDA)
			if err != nil {
				logs.Log.Error("Error marshaling da data: " + err.Error())
				return
			}

			success, addr := junction.InitVRF()
			if !success {
				logs.Log.Error("Failed to Init VRF")
				return
			}
			log.Info().Str("module", "p2p").Msg("VRF Initiated Successfully ")
			success = junction.ValidateVRF(addr)
			if !success {
				logs.Log.Error("Failed to Validate VRF")
				return
			}
			log.Info().Str("module", "p2p").Msg("VRF Validated Successfully")
			// check if VRF is successfully validated
			var vrfRecord *junctionTypes.VrfRecord
			vrfRecord = junction.QueryVRF()
			if vrfRecord == nil {
				logs.Log.Error("VRF record is nil")
				return
			}
			if !vrfRecord.IsVerified {
				logs.Log.Error("Verification of VRF is failed, need Voting for correct VRN")
				return
			}

			// DA submit
			connection := shared.Node.NodeConnections
			mdb := connection.GetDataAvailabilityDatabaseConnection()
			dbName, err := mock.MockDA(mdb, daDataByte, PodNumber)
			if err != nil {
				logs.Log.Error("Error in submitting data to DA")
				return
			}
			_ = dbName

			DaBatchSaver := connection.DataAvailabilityDatabaseConnection
			baseConfig, err := shared.LoadConfig()
			if err != nil {
				fmt.Println("Error loading configuration")
			}
			Datype := baseConfig.DA.DaType
			if Datype == "mock" {
				mdb := connection.MockDatabaseConnection
				daCheck, daCheckErr := mock.MockDA(mdb, daDataByte, PodNumber)
				if daCheckErr != nil {
					logs.Log.Error("Error in submitting data to DA")
					return
				}

				da := types.DAStruct{
					DAKey:             daCheck,
					DAClientName:      "mock-da",
					BatchNumber:       strconv.Itoa(PodNumber),
					PreviousStateHash: string(shared.GetPodState().PreviousPodHash),
					CurrentStateHash:  string(shared.GetPodState().TracksAppHash),
				}

				daStoreKey := fmt.Sprintf("da-%d", PodNumber)
				daStoreData, daStoreDataErr := json.Marshal(da)
				if daStoreDataErr != nil {
					logs.Log.Warn(fmt.Sprintf("Error in marshaling DA pointer : %s", daStoreDataErr.Error()))
				}

				storeErr := DaBatchSaver.Put([]byte(daStoreKey), daStoreData, nil)
				if storeErr != nil {
					logs.Log.Warn(fmt.Sprintf("Error in saving DA pointer in pod database : %s", storeErr.Error()))
				}

				log.Info().Str("module", "p2p").Msg("Data Saved in DA")

			} else if Datype == "avail" {
				daCheck, daCheckErr := avail.Avail(daDataByte, baseConfig.DA.DaRPC)
				if daCheckErr != nil {
					logs.Log.Warn("Error in submitting data to DA")
					return
				}

				da := types.DAStruct{
					DAKey:             daCheck,
					DAClientName:      "avail-da",
					BatchNumber:       strconv.Itoa(PodNumber),
					PreviousStateHash: string(shared.GetPodState().PreviousPodHash),
					CurrentStateHash:  string(shared.GetPodState().TracksAppHash),
				}

				daStoreKey := fmt.Sprintf("da-%d", PodNumber)
				daStoreData, daStoreDataErr := json.Marshal(da)
				if daStoreDataErr != nil {
					logs.Log.Warn(fmt.Sprintf("Error in marshaling DA pointer : %s", daStoreDataErr.Error()))
				}

				storeErr := DaBatchSaver.Put([]byte(daStoreKey), daStoreData, nil)
				if storeErr != nil {
					logs.Log.Warn(fmt.Sprintf("Error in saving DA pointer in pod database : %s", storeErr.Error()))
				}

				log.Info().Str("module", "p2p").Msg("Data Saved in DA")

			} else if Datype == "celestia" {
				daCheck, daCheckErr := celestia.Celestia(daDataByte, baseConfig.DA.DaRPC, baseConfig.DA.DaRPC)

				if daCheckErr != nil {
					logs.Log.Warn("Error in submitting data to DA")
					return
				}

				da := types.DAStruct{
					DAKey:             daCheck,
					DAClientName:      "celestia-da",
					BatchNumber:       strconv.Itoa(PodNumber),
					PreviousStateHash: string(shared.GetPodState().PreviousPodHash),
					CurrentStateHash:  string(shared.GetPodState().TracksAppHash),
				}

				daStoreKey := fmt.Sprintf("da-%d", PodNumber)
				daStoreData, daStoreDataErr := json.Marshal(da)
				if daStoreDataErr != nil {
					logs.Log.Warn(fmt.Sprintf("Error in marshaling DA pointer : %s", daStoreDataErr.Error()))
				}

				storeErr := DaBatchSaver.Put([]byte(daStoreKey), daStoreData, nil)
				if storeErr != nil {
					logs.Log.Warn(fmt.Sprintf("Error in saving DA pointer in pod database : %s", storeErr.Error()))
				}

				log.Info().Str("module", "p2p").Msg("Data Saved in DA")

			} else if Datype == "eigen" {
				daCheck, daCheckErr := eigen.Eigen(daDataByte,
					baseConfig.DA.DaRPC, baseConfig.DA.DaRPC,
				)

				if daCheckErr != nil {
					logs.Log.Warn("Error in submitting data to DA")
					return
				}

				da := types.DAStruct{
					DAKey:             daCheck,
					DAClientName:      "eigen-da",
					BatchNumber:       strconv.Itoa(PodNumber),
					PreviousStateHash: string(shared.GetPodState().PreviousPodHash),
					CurrentStateHash:  string(shared.GetPodState().TracksAppHash),
				}

				daStoreKey := fmt.Sprintf("da-%d", PodNumber)
				daStoreData, daStoreDataErr := json.Marshal(da)
				if daStoreDataErr != nil {
					logs.Log.Warn(fmt.Sprintf("Error in marshaling DA pointer : %s", daStoreDataErr.Error()))
				}

				storeErr := DaBatchSaver.Put([]byte(daStoreKey), daStoreData, nil)
				if storeErr != nil {
					logs.Log.Warn(fmt.Sprintf("Error in saving DA pointer in pod database : %s", storeErr.Error()))
				}

				log.Info().Str("module", "p2p").Msg("Data Saved in DA")

			} else {
				logs.Log.Error("Unknown layer. Please use 'avail' or 'celestia' as argument.")
				return
			}

			// submit pod
			success = junction.SubmitCurrentPod()
			if !success {
				logs.Log.Error("Failed to submit pod")
				return
			}
			log.Info().Str("module", "p2p").Msg("Pod Submitted  Successfully")

			// verify pod
			success = junction.VerifyCurrentPod()
			if !success {
				logs.Log.Error("Failed to Transact Verify pod")
				return
			}
			logs.Log.Info("pod verification transaction done")

			saveVerifiedPOD() // save data to database

			GenerateUnverifiedPods() // generate next pod

		} else {

			PodNumber := int(shared.GetPodState().LatestPodHeight)

			success, addr := junction.InitVRF()
			if !success {
				logs.Log.Error("Failed to Init VRF")
				return
			}
			logs.Log.Info("VRF initiated")

			// get own address
			_, _, accountPath, accountName, addressPrefix, tracks, err := junction.GetJunctionDetails()
			if err != nil {
				logs.Log.Error("can not get junctionDetails.json data: " + err.Error())
				return
			}
			myAddress, err := junction.CheckIfAccountExists(accountName, accountPath, addressPrefix)
			if err != nil {
				logs.Log.Error("Can not get junction wallet address")
				return
			}

			// choose one verifiable random node to verify the VRF
			// Filter out the peer with own Id
			var filteredTracks []string
			for _, track := range tracks {
				if track != myAddress {
					filteredTracks = append(filteredTracks, track)
				}
			}
			// Select a random peer from the filtered list
			selectedTrackAddress := filteredTracks[rand.Intn(len(filteredTracks))]
			fmt.Println("Selected random address:", selectedTrackAddress)

			// send verify VRF message to selected node
			VRFInitiatedMsg := VRFInitiatedMsgData{
				PodNumber:            uint64(PodNumber),
				SelectedTrackAddress: selectedTrackAddress,
				VrfInitiatorAddress:  addr,
			}

			VRFInitiatedMsgByte, err := json.Marshal(VRFInitiatedMsg)
			if err != nil {
				logs.Log.Error("Error in Marshaling ProofVote Result")
				return
			}
			gossipMsg := types.GossipData{
				Type: "vrfInitiated",
				Data: VRFInitiatedMsgByte,
			}
			gossipMsgByte, err := json.Marshal(gossipMsg)
			if err != nil {
				logs.Log.Error("Error marshaling gossip message")
				return
			}
			BroadcastMessage(CTX, Node, gossipMsgByte)
		}
	}

}

func createEVMPOD(ldt *leveldb.DB, batchStartIndex []byte, limit []byte) (witness []byte, unverifiedProof []byte, MRH []byte, podData *types.BatchStruct, err error) {
	baseConfig, err := shared.LoadConfig()
	if err != nil {
		return
	}
	limitInt, _ := strconv.Atoi(strings.TrimSpace(string(limit)))

	batchStartIndexInt, _ := strconv.Atoi(strings.TrimSpace(string(batchStartIndex)))

	var batch types.BatchStruct

	var From []string
	var To []string
	var Amounts []string
	var TransactionHash []string
	var SenderBalances []string
	var ReceiverBalances []string
	var Messages []string
	var TransactionNonces []string
	var AccountNonces []string

	for i := batchStartIndexInt; i < (config.PODSize * (limitInt + 1)); i++ {

		findKey := fmt.Sprintf("txns-%d", i+1)
		txData, err := ldt.Get([]byte(findKey), nil)
		if err != nil {
			i--
			time.Sleep(1 * time.Second)
			continue
		}
		var tx types.TransactionStruct
		err = json.Unmarshal(txData, &tx)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in unmarshalling tx data : %s", err.Error()))
			os.Exit(0)
		}

		senderBalancesCheck, err := utilis.GetBalance(tx.From, tx.BlockNumber-1, baseConfig.Station.StationRPC)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting sender balance : %s", err.Error()))
			os.Exit(0)
		}

		receiverBalancesCheck, err := utilis.GetBalance(tx.To, tx.BlockNumber-1, baseConfig.Station.StationRPC)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting reciver balance : %s", err.Error()))
			os.Exit(0)
		}

		accountNonceCheck, err := utilis.GetAccountNonce(context.Background(), tx.Hash, tx.BlockNumber, baseConfig.Station.StationRPC)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting account nonce : %s", err.Error()))
			os.Exit(0)
		}

		From = append(From, tx.From)
		To = append(To, tx.To)
		Amounts = append(Amounts, tx.Value)
		TransactionHash = append(TransactionHash, tx.Hash)
		SenderBalances = append(SenderBalances, senderBalancesCheck)
		ReceiverBalances = append(ReceiverBalances, receiverBalancesCheck)
		Messages = append(Messages, tx.Input)
		TransactionNonces = append(TransactionNonces, tx.Nonce)
		AccountNonces = append(AccountNonces, accountNonceCheck)
	}

	batch.From = From
	batch.To = To
	batch.Amounts = Amounts
	batch.TransactionHash = TransactionHash
	batch.SenderBalances = SenderBalances
	batch.ReceiverBalances = ReceiverBalances
	batch.Messages = Messages
	batch.TransactionNonces = TransactionNonces
	batch.AccountNonces = AccountNonces
	witnessVector, currentStatusHash, proofByte, pkErr := v1.GenerateProof(batch, limitInt+1)
	if pkErr != nil {
		logs.Log.Error(fmt.Sprintf("Error in generating proof : %s", pkErr.Error()))
		return nil, nil, nil, nil, pkErr
	}
	logs.Log.Warn(fmt.Sprintf("Successfully generated  Unverified proof for Batch %s in the latest phase", strconv.Itoa(limitInt+1)))

	// marshal witnessVector
	witnessVectorByte, err := json.Marshal(witnessVector)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling witness vector : %s", err.Error()))
	}

	// string to []byte currentStatusHash
	currentStatusHashByte, err := json.Marshal(currentStatusHash)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling current status hash : %s", err.Error()))
		os.Exit(0)
	}

	return witnessVectorByte, proofByte, currentStatusHashByte, &batch, nil
}

func createWasmPOD(ldt *leveldb.DB, batchStartIndex []byte, limit []byte) (witness []byte, unverifiedProof []byte, MRH []byte, podData *types.BatchStruct, err error) {
	baseConfig, err := shared.LoadConfig()
	if err != nil {
		return
	}
	limitInt, _ := strconv.Atoi(strings.TrimSpace(string(limit)))
	batchStartIndexInt, _ := strconv.Atoi(strings.TrimSpace(string(batchStartIndex)))

	var batch types.BatchStruct

	var From []string
	var To []string
	var Amounts []string
	var TransactionHash []string
	var SenderBalances []string
	var ReceiverBalances []string
	var Messages []string
	var TransactionNonces []string
	var AccountNonces []string

	for i := batchStartIndexInt; i < (config.PODSize * (limitInt + 1)); i++ {
		findKey := fmt.Sprintf("txns-%d", i+1)
		txData, err := ldt.Get([]byte(findKey), nil)

		if err != nil {
			i--
			time.Sleep(1 * time.Second)
			continue
		}

		var txn types.BatchTransaction
		err = json.Unmarshal(txData, &txn)
		if err != nil {
			logs.Log.Info(fmt.Sprintf("Error in unmarshalling tx data : %s", err.Error()))
		}
		fromCheck := utilis.Bech32Decoder(txn.Tx.Body.Messages[0].FromAddress)
		toCheck := utilis.Bech32Decoder(txn.Tx.Body.Messages[0].ToAddress)
		transactionHashCheck := utilis.TXHashCheck(txn.TxResponse.TxHash)

		senderBalancesCheck := utilis.AccountBalanceCheck(txn.Tx.Body.Messages[0].FromAddress, txn.TxResponse.Height, baseConfig.Station.StationAPI)
		receiverBalancesCheck := utilis.AccountBalanceCheck(txn.Tx.Body.Messages[0].ToAddress, txn.TxResponse.Height, baseConfig.Station.StationAPI)
		accountNoncesCheck := utilis.AccountNounceCheck(txn.Tx.Body.Messages[0].FromAddress, baseConfig.Station.StationAPI)

		From = append(From, fromCheck)
		To = append(To, toCheck)
		Amounts = append(Amounts, txn.Tx.Body.Messages[0].Amount[0].Amount)
		SenderBalances = append(SenderBalances, senderBalancesCheck)
		ReceiverBalances = append(ReceiverBalances, receiverBalancesCheck)
		TransactionHash = append(TransactionHash, transactionHashCheck)
		Messages = append(Messages, fmt.Sprint(txn.Tx.Body.Messages[0]))
		TransactionNonces = append(TransactionNonces, "0")
		AccountNonces = append(AccountNonces, accountNoncesCheck)
	}

	batch.From = From
	batch.To = To
	batch.Amounts = Amounts
	batch.TransactionHash = TransactionHash
	batch.SenderBalances = SenderBalances
	batch.ReceiverBalances = ReceiverBalances
	batch.Messages = Messages
	batch.TransactionNonces = TransactionNonces
	batch.AccountNonces = AccountNonces

	// add prover here
	witnessVector, currentStatusHash, proofByte, pkErr := v1Wasm.GenerateProof(batch, limitInt+1)
	if pkErr != nil {
		logs.Log.Error("Error in generating proof : %s" + pkErr.Error())
		os.Exit(0)
	}

	witnessVectorByte, err := json.Marshal(witnessVector)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling witness vector : %s", err.Error()))
	}

	// string to []byte currentStatusHash
	currentStatusHashByte, err := json.Marshal(currentStatusHash)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling current status hash : %s", err.Error()))
		os.Exit(0)
	}

	return witnessVectorByte, proofByte, currentStatusHashByte, &batch, nil

}

func saveVerifiedPOD() {

	podState := shared.GetPodState()
	batchTimestamp := time.Now()
	podState.Timestamp = &batchTimestamp
	currentPodNumber := podState.LatestPodHeight
	currentPodNumberInt := int(currentPodNumber)

	lds := shared.Node.NodeConnections.GetStaticDatabaseConnection()

	err := lds.Put([]byte("batchStartIndex"), []byte(strconv.Itoa(config.PODSize*(currentPodNumberInt))), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in updating batchStartIndex in static db : %s", err.Error()))
		os.Exit(0)
	}

	err = lds.Put([]byte("batchCount"), []byte(strconv.Itoa(currentPodNumberInt)), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in updating batchCount in static db : %s", err.Error()))
		os.Exit(0)
	}

	batchDB := shared.Node.NodeConnections.GetPodsDatabaseConnection()
	podKey := fmt.Sprintf("pod-%d", currentPodNumberInt)

	batchInputWithTimestampBytes, err := json.Marshal(podState)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling batch data : %s", err.Error()))
		os.Exit(0)
	}
	err = batchDB.Put([]byte(podKey), batchInputWithTimestampBytes, nil)
	if err != nil {
		panic("Failed to update pod data: " + err.Error())
	}

	podState.MasterTrackAppHash = nil
	shared.SetPodState(podState)

	log.Info().Str("module", "p2p").Msg("Present Pod has been saved Locally")
}

func generatePodHash(Witness, uZKP, MRH []byte, podNumber []byte) []byte {
	hash := sha256.New()
	hash.Write(Witness)
	hash.Write(uZKP)
	hash.Write(MRH)
	hash.Write(podNumber)
	return hash.Sum(nil)
}

func updateNewPodState(CombinedPodHash, Witness, uZKP, MRH []byte, podNumber uint64, batchInput *types.BatchStruct) {
	var podState *shared.PodState
	votes := make(map[string]shared.Votes)
	podState = &shared.PodState{
		LatestPodHeight:     podNumber,
		LatestPodHash:       MRH,
		PreviousPodHash:     shared.GetPodState().LatestPodHash,
		LatestPodProof:      uZKP,
		LatestPublicWitness: Witness,
		Votes:               votes,
		TracksAppHash:       CombinedPodHash,
		Batch:               batchInput,
	}
	shared.SetPodState(podState)
	updatePodStateInDatabase(podState)
}

func updatePodStateInDatabase(podState *shared.PodState) {
	stateConnection := shared.Node.NodeConnections.GetStateDatabaseConnection()

	podStateByte, err := json.Marshal(podState)
	if err != nil {
		logs.Log.Error(err.Error())
		os.Exit(0)
	}

	err = stateConnection.Put([]byte("podState"), podStateByte, nil)
	if err != nil {
		logs.Log.Error(err.Error())
	}
}

func getPodStateFromDatabase() (*types.PodState, error) {
	var podStateData *types.PodState
	stateConnection := shared.Node.NodeConnections.GetStateDatabaseConnection()

	podStateDataByte, err := stateConnection.Get([]byte("podState"), nil)
	if err != nil {
		logs.Log.Error("error in getting pod state data from database")
		return nil, err
	}
	err = json.Unmarshal(podStateDataByte, &podStateData)
	if err != nil {
		logs.Log.Error("error in unmarshal pod state data")
		return nil, err
	}

	return podStateData, nil

}
