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
		BatchGeneration(wg, client, ctx, lds, ldt, ldbatch, ldda, []byte(strconv.Itoa(config.PODSize*(limitInt+1))))
	}
	logs.Log.Warn(fmt.Sprintf("Successfully generated proof for Batch %s in the latest phase", strconv.Itoa(limitInt+1)))
	fmt.Println("Witness Vector: ", witnessVector)
	fmt.Println("Proof Byte: ", proofByte)
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
	p2p.ZKPGossip(ProofDataByte)

	err = lds.Put([]byte("podPool"), ProofDataByte, nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in updating ProofDataByte in static db : %s", err.Error()))
		os.Exit(0)
	}

	// check who is the validator of this POD
	// if validator is this node:

	//p2p.ZKPGossip(ProofDataByte)
	//daKeyHash, err := DaCall(batch.TransactionHash, client, ctx, currentStatusHash, limitInt+1, ldda)
	//if err != nil {
	//	logs.Log.Error(fmt.Sprintf("Error in adding Da client : %s", err.Error()))
	//	logs.Log.Warn("Trying again...")
	//	time.Sleep(3 * time.Second)
	//	BatchGeneration(wg, client, ctx, lds, ldt, ldbatch, ldda, []byte(strconv.Itoa(config.PODSize*(limitInt+1))))
	//}
	//
	logs.Log.Warn(fmt.Sprintf("Successfully added Da client for Batch %s in the latest phase", "x"))

	//addBatchRes := settlement_client.AddBatch(witnessVector, limitInt+1, lds)
	//if addBatchRes == "nil" {
	//	logs.Log.Error(fmt.Sprintf("Error in adding batch to settlement client : %s", addBatchRes))
	//	os.Exit(0)
	//}
	//
	//status := settlement_client.VerifyBatch(limitInt+1, proofByte, ldda, lds)
	//if !status {
	//	logs.Log.Error(fmt.Sprintf("Error in verifying batch to settlement client : %s", status))
	//	os.Exit(0)
	//}
	//
	logs.Log.Warn(fmt.Sprintf("Successfully generated proof for Batch %s in the latest phase", strconv.Itoa(limitInt+1)))
	//
	batchJSON, err := json.Marshal(batch)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling batch data : %s", err.Error()))
		os.Exit(0)
	}

	batchKey := fmt.Sprintf("batch-%d", limitInt+1)
	err = ldbatch.Put([]byte(batchKey), batchJSON, nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in writing batch data to file : %s", err.Error()))
		os.Exit(0)
	}

	err = lds.Put([]byte("batchStartIndex"), []byte(strconv.Itoa(config.PODSize*(limitInt+1))), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in updating batchStartIndex in static db : %s", err.Error()))
		os.Exit(0)
	}

	err = lds.Put([]byte("batchCount"), []byte(strconv.Itoa(limitInt+1)), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in updating batchCount in static db : %s", err.Error()))
		os.Exit(0)
	}

	err = os.WriteFile("data/batchCount.txt", []byte(strconv.Itoa(limitInt+1)), 0666)
	if err != nil {
		panic("Failed to update batch number: " + err.Error())
	}

	//logs.Log.Warn(fmt.Sprintf("Successfully saved Batch %s in the latest phase", strconv.Itoa(limitInt+1)))

	BatchGeneration(wg, client, ctx, lds, ldt, ldbatch, ldda, []byte(strconv.Itoa(config.PODSize*(limitInt+1))))
}
