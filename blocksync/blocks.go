package blocksync

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/config"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/airchains-network/decentralized-sequencer/utilis"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/syndtr/goleveldb/leveldb"
	"math/big"
	"os"
	"strconv"
	"time"
)

func StoreEVMBlock(client *ethclient.Client, ctx context.Context, blockIndex int, ldb *leveldb.DB, ldt *leveldb.DB) {
	blockData, err := client.BlockByNumber(ctx, big.NewInt(int64(blockIndex)))
	if err != nil {
		//errMessage := fmt.Sprintf("Failed to get block data for block number %d: %s", blockIndex, err)
		//logs.Log.Error(errMessage)
		logs.Log.Info("Waiting for the next station block..")
		time.Sleep(config.StationBlockDuration * time.Second)
		StoreEVMBlock(client, ctx, blockIndex, ldb, ldt)
	}

	block := types.BlockStruct{
		BaseFeePerGas:    utilis.ToString(blockData.Header().BaseFee),
		Difficulty:       utilis.ToString(blockData.Difficulty().String()),
		ExtraData:        utilis.ToString(blockData.Extra()),
		GasLimit:         utilis.ToString(blockData.GasLimit()),
		GasUsed:          utilis.ToString(blockData.GasUsed()),
		Hash:             utilis.ToString(blockData.Hash().String()),
		LogsBloom:        utilis.ToString(blockData.Bloom()),
		Miner:            utilis.ToString(blockData.Coinbase().String()),
		MixHash:          utilis.ToString(blockData.MixDigest().String()),
		Nonce:            utilis.ToString(blockData.Nonce()),
		Number:           utilis.ToString(blockData.Number().String()),
		ParentHash:       utilis.ToString(blockData.ParentHash().String()),
		ReceiptsRoot:     utilis.ToString(blockData.ReceiptHash().String()),
		Sha3Uncles:       utilis.ToString(blockData.UncleHash()),
		Size:             utilis.ToString(blockData.Size()),
		StateRoot:        utilis.ToString(blockData.Root().String()),
		Timestamp:        utilis.ToString(blockData.Time()),
		TotalDifficulty:  utilis.ToString(blockData.Difficulty().String()),
		TransactionCount: blockData.Transactions().Len(), // Assuming transactionCount is already defined and holds the appropriate value
		TransactionsRoot: utilis.ToString(blockData.TxHash().String()),
		Uncles:           utilis.ToString(blockData.Uncles()),
	}

	data, err := json.Marshal(block)
	if err != nil {
		errMessage := fmt.Sprintf("Error marshalling block data: %s", err)
		logs.Log.Error(errMessage)
	}
	key := fmt.Sprintf("block_%s", block.Number)
	err = ldb.Put([]byte(key), data, nil)
	if err != nil {
		errMessage := fmt.Sprintf("Error inserting block data into database: %s", err)
		logs.Log.Error(errMessage)
	}

	transactions := blockData.Transactions()
	if transactions == nil {
		fmt.Println("No transactions found in block number", blockIndex)
	}

	//infoMessage := fmt.Sprintf("Block number %d has %d transactions", blockIndex, transactions.Len())
	//logs.Log.Info(infoMessage)

	for i := 0; i < block.TransactionCount; i++ {
		StoreEVMTransactions(client, ctx, ldt, transactions[i].Hash().String(), blockIndex, block.Hash)
	}

	err = os.WriteFile("data/blockCount.txt", []byte(strconv.Itoa(blockIndex+1)), 0666)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in saving blockCount in static db : %s", err.Error()))
		os.Exit(0)
	}
	blockCount := blockIndex + 1
	err = ldb.Put([]byte("blockCount"), []byte(strconv.Itoa(blockCount)), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in saving latestBlock in block db : %s", err.Error()))
	}
	StoreEVMBlock(client, ctx, blockIndex+1, ldb, ldt)
}

//TODO ADD COSMWASM BLOCK Saver also
