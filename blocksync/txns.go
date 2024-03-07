package blocksync

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	stationTypes "github.com/airchains-network/decentralized-sequencer/types"
	"github.com/emirpasic/gods/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"strconv"
	"strings"
	"time"
)

func insertTxn(db *leveldb.DB, txns stationTypes.TransactionStruct, transactionNumber int) error {
	data, err := json.Marshal(txns)
	if err != nil {
		return err
	}

	txnsKey := fmt.Sprintf("txns-%d", transactionNumber+1)
	err = db.Put([]byte(txnsKey), data, nil)
	if err != nil {
		return err
	}
	err = os.WriteFile("data/transactionCount.txt", []byte(strconv.Itoa(transactionNumber+1)), 0666)
	if err != nil {
		return err
	}

	return nil
}

func StoreEVMTransactions(client *ethclient.Client, ctx context.Context, ldt *leveldb.DB, transactionHash string, blockNumber int, blockHash string) {
	fmt.Println("Storing EVM Transactions")
	blockNumberUint64, err := strconv.ParseUint(strconv.Itoa(blockNumber), 10, 64)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error parsing block number to uint64:", err))
		time.Sleep(2 * time.Second)
		logs.Log.Info("Retrying in 2s...")
		StoreEVMTransactions(client, ctx, ldt, transactionHash, blockNumber, blockHash)
	}

	txHash := common.HexToHash(transactionHash)
	tx, isPending, err := client.TransactionByHash(ctx, txHash)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Failed to get transaction by hash: %s", err))
		os.Exit(0)
	}

	if isPending {
		logs.Log.Warn("Transaction is pending")
		logs.Log.Info(fmt.Sprintf("Transaction type: %d\n", tx.Type()))
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Failed to get the network ID: %v", err))
		os.Exit(0)
	}
	msg, err := types.Sender(types.NewLondonSigner(chainID), tx)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Failed to derive the sender address: %v", err))
		os.Exit(0)
	}

	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Failed to fetch the transaction receipt: %v", err))
		os.Exit(0)
	}

	v, r, s := tx.RawSignatureValues()

	txData := stationTypes.TransactionStruct{
		BlockHash:        blockHash,
		BlockNumber:      blockNumberUint64,
		From:             msg.Hex(),
		Gas:              utils.ToString(tx.Gas()),
		GasPrice:         tx.GasPrice().String(),
		Hash:             tx.Hash().Hex(),
		Input:            string(tx.Data()),
		Nonce:            utils.ToString(tx.Nonce()),
		R:                r.String(),
		S:                s.String(),
		To:               tx.To().Hex(),
		TransactionIndex: utils.ToString(receipt.TransactionIndex),
		Type:             fmt.Sprintf("%d", tx.Type()),
		V:                v.String(),
		Value:            tx.Value().String(),
	}

	fileOpen, err := os.Open("data/transactionCount.txt")
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Failed to read file: %s" + err.Error()))
		os.Exit(0)
	}
	defer fileOpen.Close()

	scanner := bufio.NewScanner(fileOpen)

	transactionNumberBytes := ""

	for scanner.Scan() {
		transactionNumberBytes = scanner.Text()
	}

	transactionNumber, err := strconv.Atoi(strings.TrimSpace(string(transactionNumberBytes)))
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Invalid transaction number : %s" + err.Error()))
		os.Exit(0)
	}

	insetTxnErr := insertTxn(ldt, txData, transactionNumber)
	if insetTxnErr != nil {
		logs.Log.Error(fmt.Sprintf("Failed to insert transaction: %s" + insetTxnErr.Error()))
		os.Exit(0)
	}

	logs.Log.Debug(fmt.Sprintf("Successfully saved Transation %s in the latest phase", txHash))

}

//TODO add COSMOS txn saver
