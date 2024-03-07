package blocksync

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/config"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/emirpasic/gods/utils"
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
		errMessage := fmt.Sprintf("Failed to get block data for block number %d: %s", blockIndex, err)
		logs.Log.Error(errMessage)
		logs.Log.Info("Waiting for the next block..")
		time.Sleep(config.StationBlockDuration * time.Second)
		StoreEVMBlock(client, ctx, blockIndex, ldb, ldt)
	}

	block := types.BlockStruct{
		BaseFeePerGas:    utils.ToString(blockData.Header().BaseFee),
		Difficulty:       utils.ToString(blockData.Difficulty().String()),
		ExtraData:        utils.ToString(blockData.Extra()),
		GasLimit:         utils.ToString(blockData.GasLimit()),
		GasUsed:          utils.ToString(blockData.GasUsed()),
		Hash:             utils.ToString(blockData.Hash().String()),
		LogsBloom:        utils.ToString(blockData.Bloom()),
		Miner:            utils.ToString(blockData.Coinbase().String()),
		MixHash:          utils.ToString(blockData.MixDigest().String()),
		Nonce:            utils.ToString(blockData.Nonce()),
		Number:           utils.ToString(blockData.Number().String()),
		ParentHash:       utils.ToString(blockData.ParentHash().String()),
		ReceiptsRoot:     utils.ToString(blockData.ReceiptHash().String()),
		Sha3Uncles:       utils.ToString(blockData.UncleHash()),
		Size:             utils.ToString(blockData.Size()),
		StateRoot:        utils.ToString(blockData.Root().String()),
		Timestamp:        utils.ToString(blockData.Time()),
		TotalDifficulty:  utils.ToString(blockData.Difficulty().String()),
		TransactionCount: blockData.Transactions().Len(), // Assuming transactionCount is already defined and holds the appropriate value
		TransactionsRoot: utils.ToString(blockData.TxHash().String()),
		Uncles:           utils.ToString(blockData.Uncles()),
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

	infoMessage := fmt.Sprintf("Block number %d has %d transactions", blockIndex, transactions.Len())
	logs.Log.Info(infoMessage)

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
	test, err := ldb.Get([]byte("blockCount"), nil)
	fmt.Println("Block Count: ", string(test))
	StoreEVMBlock(client, ctx, blockIndex+1, ldb, ldt)
}

//TODO ADD COSMWASM BLOCK Saver also
