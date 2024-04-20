package blocksync

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	stationTypes "github.com/airchains-network/decentralized-sequencer/types"
	utilis "github.com/airchains-network/decentralized-sequencer/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/syndtr/goleveldb/leveldb"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func insertTxnEVM(db *leveldb.DB, txns stationTypes.TransactionStruct, transactionNumber int) error {
	data, err := json.Marshal(txns)
	if err != nil {
		return err
	}

	txnsKey := fmt.Sprintf("txns-%d", transactionNumber+1)
	err = db.Put([]byte(txnsKey), data, nil)
	if err != nil {
		return err
	}

	err = db.Put([]byte("txnCount"), []byte(strconv.Itoa(transactionNumber+1)), nil)
	if err != nil {
		return err
	}

	return nil
}
func insertTxnWASM(db *leveldb.DB, txns []byte, transactionNumber int) error {

	txnsKey := fmt.Sprintf("txns-%d", transactionNumber+1)
	err := db.Put([]byte(txnsKey), txns, nil)
	if err != nil {
		return err
	}

	err = db.Put([]byte("txnCount"), []byte(strconv.Itoa(transactionNumber+1)), nil)
	if err != nil {
		return err
	}

	return nil
}

func StoreEVMTransactions(client *ethclient.Client, ctx context.Context, ldt *leveldb.DB, transactionHash string, blockNumber int, blockHash string) {
	blockNumberUint64, err := strconv.ParseUint(strconv.Itoa(blockNumber), 10, 64)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("error parsing block number to uint64:", err))
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

	var toAddress string
	if tx.To() == nil {
		toAddress = "0x000000000000000000000000000000000000000000"
	} else {
		toAddress = tx.To().Hex()
	}

	var txData = stationTypes.TransactionStruct{
		BlockHash:        blockHash,
		BlockNumber:      blockNumberUint64,
		From:             msg.Hex(),
		Gas:              utilis.ToString(tx.Gas()),
		GasPrice:         tx.GasPrice().String(),
		Hash:             tx.Hash().Hex(),
		Input:            string(tx.Data()),
		Nonce:            utilis.ToString(tx.Nonce()),
		R:                r.String(),
		S:                s.String(),
		To:               toAddress,
		TransactionIndex: utilis.ToString(receipt.TransactionIndex),
		Type:             fmt.Sprintf("%d", tx.Type()),
		V:                v.String(),
		Value:            tx.Value().String(),
	}

	// get transaction number from database
	transactionNumberBytes, err := ldt.Get([]byte("txnCount"), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Failed to get transaction number: %s" + err.Error()))
		os.Exit(0)
	}

	transactionNumber, err := strconv.Atoi(strings.TrimSpace(string(transactionNumberBytes)))
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Invalid transaction number : %s" + err.Error()))
		os.Exit(0)
	}

	insetTxnErr := insertTxnEVM(ldt, txData, transactionNumber)
	if insetTxnErr != nil {
		logs.Log.Error(fmt.Sprintf("Failed to insert transaction: %s" + insetTxnErr.Error()))
		os.Exit(0)
	}

}

func StoreWasmTransaction(txn []interface{}, db *leveldb.DB, JsonAPI string) {
	for i, tx := range txn {
		_ = i
		hash, err := ComputeTransactionHash(tx.(string))
		if err != nil {
			log.Println("Error computing transaction hash:", err)
			continue
		}
		rpcUrl := fmt.Sprintf("%s/cosmos/tx/v1beta1/txs/%s", JsonAPI, hash)
		respo, err := http.Get(rpcUrl)
		if err != nil {
			log.Println("HTTP request failed for transaction hash:", err)
			continue
		}
		if respo != nil {
			bodyTxnHash, err := io.ReadAll(respo.Body)
			err = respo.Body.Close()
			if err != nil {
				return
			}
			// get transaction number from database
			transactionNumberBytes, err := db.Get([]byte("txnCount"), nil)
			if err != nil {
				logs.Log.Error(fmt.Sprintf("Failed to get transaction number: %s" + err.Error()))
				os.Exit(0)
			}

			transactionNumber, err := strconv.Atoi(strings.TrimSpace(string(transactionNumberBytes)))
			if err != nil {
				logs.Log.Error(fmt.Sprintf("Invalid transaction number : %s" + err.Error()))
				os.Exit(0)
			}
			insetTxnErr := insertTxnWASM(db, bodyTxnHash, transactionNumber)
			if insetTxnErr != nil {
				logs.Log.Error(fmt.Sprintf("Failed to insert transaction: %s" + insetTxnErr.Error()))
				os.Exit(0)
			}

		} else {
			log.Println("Received nil response for transaction hash:", hash)
		}
	}
}

func ComputeTransactionHash(base64Tx string) (string, error) {
	txBytes, err := base64.StdEncoding.DecodeString(base64Tx)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(txBytes)
	txHash := hex.EncodeToString(hash[:])
	return txHash, nil
}
