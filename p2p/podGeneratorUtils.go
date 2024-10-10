package p2p

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/tracks/config"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/node/shared"
	"github.com/airchains-network/tracks/types"
	utilis "github.com/airchains-network/tracks/utils"
	v1 "github.com/airchains-network/tracks/zk/v1EVM"
	v1Wasm "github.com/airchains-network/tracks/zk/v1WASM"
	"github.com/rs/zerolog/log"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	Mock               = "mock"
	Avail              = "avail"
	Celestia           = "celestia"
	Eigen              = "eigen"
	BatchCountKey      = "batchCount"
	BatchStartIndexKey = "batchStartIndex"
)

func CheckErrorAndExit(err error, message string, exitCode int) {
	if err != nil {
		log.Log().Str("module", "p2p").Msg("error: " + message)
		os.Exit(exitCode)
	}
}

func GetValueOrDefault(db *leveldb.DB, key []byte, defaultValue []byte) ([]byte, error) {
	val, err := db.Get(key, nil)
	if err != nil {
		logs.Log.Warn(fmt.Sprintf("%s not found in static db", string(key)))
		err = db.Put(key, defaultValue, nil)
		CheckErrorAndExit(err, fmt.Sprintf("Error in saving %s in static db", string(key)), 0)
	}
	return val, nil
}

func retryGetBalance(address string, blockNumber int, rpcURL string) (string, error) {
	var balance string
	err := utilis.Retry(func() error {
		var err error
		balance, err = utilis.GetBalance(address, uint64(blockNumber), rpcURL)
		return err
	})
	return balance, err
}

func retryGetAccountNonce(ctx context.Context, txHash string, blockNumber int, rpcURL string) (string, error) {
	var nonce string
	err := utilis.Retry(func() error {
		var err error
		nonce, err = utilis.GetAccountNonce(ctx, txHash, uint64(blockNumber), rpcURL)
		return err
	})
	return nonce, err
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

		senderBalancesCheck, err := retryGetBalance(tx.From, int(tx.BlockNumber-1), baseConfig.Station.StationRPC)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting sender balance : %s", err.Error()))
		}

		receiverBalancesCheck, err := retryGetBalance(tx.To, int(tx.BlockNumber-1), baseConfig.Station.StationRPC)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting reciver balance : %s", err.Error()))
		}

		accountNonceCheck, err := retryGetAccountNonce(context.Background(), tx.Hash, int(tx.BlockNumber), baseConfig.Station.StationRPC)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting account nonce : %s", err.Error()))
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
	log.Info().Str("module", "p2p").Str("Pod Number", strconv.Itoa(limitInt+1)).Msg("Successfully generated  Unverified proof")

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

func createWasmBatch(ldt *leveldb.DB, batchStartIndex []byte, limit []byte) (podData *types.BatchStruct, err error) {
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

	return &batch, nil

}

func createEVMBatch(ldt *leveldb.DB, batchStartIndex []byte, limit []byte) (podData *types.BatchStruct, err error) {
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

		senderBalancesCheck, err := retryGetBalance(tx.From, int(tx.BlockNumber-1), baseConfig.Station.StationRPC)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting sender balance : %s", err.Error()))
		}

		receiverBalancesCheck, err := retryGetBalance(tx.To, int(tx.BlockNumber-1), baseConfig.Station.StationRPC)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting reciver balance : %s", err.Error()))
		}

		accountNonceCheck, err := retryGetAccountNonce(context.Background(), tx.Hash, int(tx.BlockNumber), baseConfig.Station.StationRPC)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting account nonce : %s", err.Error()))
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

	return &batch, nil

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

func generateBatchHash(podNumber []byte) []byte {
	hash := sha256.New()
	hash.Write(podNumber)
	return hash.Sum(nil)
}

func storeNewPodState(CombinedPodHash, Witness, uZKP, previousMRH, MRH []byte, podNumber uint64, batchInput *types.BatchStruct, txState string) {
	var podState *shared.PodState
	votes := make(map[string]shared.Votes)
	podState = &shared.PodState{
		LatestPodHeight:     podNumber,
		LatestTxState:       txState,
		LatestPodHash:       MRH,
		PreviousPodHash:     previousMRH,
		LatestPodProof:      uZKP,
		LatestPublicWitness: Witness,
		Votes:               votes,
		TracksAppHash:       CombinedPodHash,
		Batch:               batchInput,
	}
	shared.SetPodState(podState)
	updatePodStateInDatabase(podState)
}
func storeNewBatchState(CombinedPodHash []byte, podNumber uint64, batchInput *types.BatchStruct, txState string) {
	var podState *shared.PodState
	votes := make(map[string]shared.Votes)
	podState = &shared.PodState{
		LatestPodHeight: podNumber,
		LatestTxState:   txState,
		Votes:           votes,
		TracksAppHash:   CombinedPodHash,
		Batch:           batchInput,
	}
	shared.SetPodState(podState)
	updatePodStateInDatabase(podState)
}
func updateNewPodState(CombinedPodHash, Witness, uZKP, MRH []byte, podNumber uint64, batchInput *types.BatchStruct, txState string) {
	var podState *shared.PodState
	votes := make(map[string]shared.Votes)
	podState = &shared.PodState{
		LatestPodHeight:     podNumber,
		LatestTxState:       txState,
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

func updateNewBatchState(CombinedPodHash []byte, podNumber uint64, batchInput *types.BatchStruct, txState string) {
	var podState *shared.PodState
	votes := make(map[string]shared.Votes)
	podState = &shared.PodState{
		LatestPodHeight: podNumber,
		LatestTxState:   txState,
		PreviousPodHash: shared.GetPodState().LatestPodHash,
		Votes:           votes,
		TracksAppHash:   CombinedPodHash,
		Batch:           batchInput,
	}
	shared.SetPodState(podState)
	updatePodStateInDatabase(podState)
}
func updateTxState(txState string) {
	podState := shared.GetPodState()
	podState.LatestTxState = txState
	shared.SetPodState(podState)
	updatePodStateInDatabase(podState)
}

func SetPodNumber(currentPodNumber uint64) {
	podState := shared.GetPodState()
	podState.LatestPodHeight = currentPodNumber
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
func GetPodStateFromDatabase() (*types.PodState, error) {
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
