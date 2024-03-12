package pods

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/config"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/p2p"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/airchains-network/decentralized-sequencer/utilis"
	v1 "github.com/airchains-network/decentralized-sequencer/zk/v1"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

//type LatestUnverifiedData struct {
//	count     uint64
//	proofData []byte
//}
//
//var (
//	LatestUnverifiedValue LatestUnverifiedData
//	mtx                   sync.Mutex
//)
//
//func NewLatestUnverifiedData(initialCount uint64, initialProofData []byte) *LatestUnverifiedData {
//	return &LatestUnverifiedData{
//		count:     initialCount,
//		proofData: initialProofData,
//	}
//}
//
//func DefaultUnverifiedData() {
//	mtx.Lock() // Lock the mutex before accessing count
//	LatestUnverifiedValue = *NewLatestUnverifiedData(0, []byte{})
//	mtx.Unlock() // Unlock the mutex after accessing count
//}
//
//// IncrementUnverifiedPod increments the count safely
//func (data *LatestUnverifiedData) IncrementUnverifiedPod() {
//	mtx.Lock()   // Lock the mutex before accessing count
//	data.count++ // Critical section: modify count
//	mtx.Unlock() // Unlock the mutex after accessing count
//}
//
//// ValueUnverifiedPod returns the current count safely
//func (data *LatestUnverifiedData) ValueUnverifiedPod() uint64 {
//	mtx.Lock()         // Lock the mutex before accessing count
//	defer mtx.Unlock() // Unlock the mutex after accessing count using defer
//	return data.count  // Critical section: read count
//}
//
//// UpdateUnverifiedProof updates the proof data safely
//func (data *LatestUnverifiedData) UpdateUnverifiedProof(proof []byte) {
//	mtx.Lock()             // Lock the mutex before accessing proof data
//	data.proofData = proof // Critical section: modify proof data
//	mtx.Unlock()           // Unlock the mutex after accessing proof data
//}
//
//// ValueUnverifiedProof returns the current proof data safely
//func (data *LatestUnverifiedData) ValueUnverifiedProof() []byte {
//	mtx.Lock()            // Lock the mutex before accessing proof data
//	defer mtx.Unlock()    // Unlock the mutex after accessing proof data using defer
//	return data.proofData // Critical section: read proof data
//}
//type Gossiper interface {
//	ZKPGossip(proofDataByte []byte)
//}

func BatchGeneration(wg *sync.WaitGroup, client *ethclient.Client, ctx context.Context, lds *leveldb.DB, ldt *leveldb.DB, ldbatch *leveldb.DB, ldda *leveldb.DB, batchStartIndex []byte) {

	defer wg.Done()

	limit, err := lds.Get([]byte("batchCount"), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in getting batchCount from static db : %s", err.Error()))
		os.Exit(0)
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

		senderBalancesCheck, err := utilis.GetBalance(tx.From, (tx.BlockNumber - 1))
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting sender balance : %s", err.Error()))
			os.Exit(0)
		}

		receiverBalancesCheck, err := utilis.GetBalance(tx.To, (tx.BlockNumber - 1))
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting reciver balance : %s", err.Error()))
			os.Exit(0)
		}

		accountNouceCheck, err := utilis.GetAccountNonce(ctx, tx.Hash, tx.BlockNumber)
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
		AccountNonces = append(AccountNonces, accountNouceCheck)
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
		//TODO : Add proper error handling
		//BatchGenerationV2(wg, client, ctx, lds, ldt, ldbatch, ldda, []byte(strconv.Itoa(config.PODSize*(limitInt+1))))
	}
	logs.Log.Warn(fmt.Sprintf("Successfully generated proof for Batch %s in the latest phase", strconv.Itoa(limitInt+1)))
	fmt.Println("Witness Vector: ", witnessVector)
	fmt.Println("Current Status Hash: ", currentStatusHash)

	proofGossip := types.ProofData{
		Proof:     proofByte,
		PodNumber: uint64(limitInt + 1),
	}
	proofByteGossip, err := json.Marshal(proofGossip)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling proof data : %s", err.Error()))
		os.Exit(0)
	}

	proofData := types.GossipData{
		Type: "proof",
		Data: proofByteGossip,
	}
	// proofData to byte
	ProofDataByte, err := json.Marshal(proofData)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling proof data : %s", err.Error()))
		os.Exit(0)
	}

	err = lds.Put([]byte("podPool"), ProofDataByte, nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in updating ProofDataByte in static db : %s", err.Error()))
		os.Exit(0)
	}

	// check who is the validator of this POD
	// if validator is this node:

	p2p.ZKPGossip(ProofDataByte)

	// after receiving the proofResponse, the data is saved in databas(e
	// check latest verified pod from database.
	//for {}

}
