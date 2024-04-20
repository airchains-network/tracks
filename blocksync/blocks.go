package blocksync

import (
	"context"
	"encoding/json"
	"fmt"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/airchains-network/decentralized-sequencer/utils"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/syndtr/goleveldb/leveldb"
	"io"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

func StoreEVMBlock(client *ethclient.Client, ctx context.Context, blockIndex int, ldb *leveldb.DB, ldt *leveldb.DB) {

	blockData, err := client.BlockByNumber(ctx, big.NewInt(int64(blockIndex)))
	if err != nil {

		errMessage := fmt.Sprintf("Failed to get block data for block number %d: %s", blockIndex, err)
		logs.Log.Error(errMessage)
		logs.Log.Info("Waiting for the next station block..")
		time.Sleep(3 * time.Second)
		StoreEVMBlock(client, ctx, blockIndex, ldb, ldt)
	}

	var block = types.BlockStruct{
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

	for i := 0; i < block.TransactionCount; i++ {
		StoreEVMTransactions(client, ctx, ldt, transactions[i].Hash().String(), blockIndex, block.Hash)
	}

	blockCount := blockIndex + 1
	err = ldb.Put([]byte("blockCount"), []byte(strconv.Itoa(blockCount)), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in saving latestBlock in block db : %s", err.Error()))
	}
	StoreEVMBlock(client, ctx, blockIndex+1, ldb, ldt)
}

func getLastProcessedBlock(db *leveldb.DB) int {
	lastBlockKey := []byte("lastProcessedBlock")
	data, err := db.Get(lastBlockKey, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			// If not found, return 0 indicating start from the beginning
			return 0
		}

	}
	lastBlockNum, _ := strconv.Atoi(string(data))
	return lastBlockNum
}

func StoreWasmBlock(ldb *leveldb.DB, ldt *leveldb.DB, JsonRPC string, JsonAPI string) {
	rpcUrl := fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/latest", JsonAPI)
	res, resErr := http.Get(rpcUrl)
	if resErr != nil {
		logs.Log.Error(resErr.Error())
	}
	defer res.Body.Close()
	bodyBlockHeight, bodyBlockHeightErr := io.ReadAll(res.Body)
	if bodyBlockHeightErr != nil {
		logs.Log.Error(bodyBlockHeightErr.Error())
	}

	var blockHeight BlockObject
	error := json.Unmarshal(bodyBlockHeight, &blockHeight)
	if error != nil {
		logs.Log.Error(error.Error())

	}
	latestBlock := blockHeight.Block.Header.Height

	numLatestBlock, err := strconv.Atoi(latestBlock)
	if err != nil {
		logs.Log.Error(err.Error())

	}
	lastBlock := getLastProcessedBlock(ldb)
	startBlock := lastBlock + 1

	OldWasmBlocks(JsonRPC, JsonAPI, startBlock, numLatestBlock, ldb, ldt)
	NewWasmBlocks(JsonRPC, JsonAPI, numLatestBlock, ldb, ldt)

}

func OldWasmBlocks(JsonRPC string, JsonAPI string, startBlock int, numLatestBlock int, db *leveldb.DB, txnDB *leveldb.DB) {
	for i := startBlock; i <= numLatestBlock; i++ {
		rpcUrl := fmt.Sprintf("%s/block?height=%d", JsonRPC, i)
		resp, err := http.Get(rpcUrl)
		if err != nil {
			logs.Log.Error(err.Error())

		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close() // Close the response body here
		if err != nil {
			logs.Log.Error(err.Error())
		}

		var blockData Response

		jsonErr := json.Unmarshal(body, &blockData)
		if jsonErr != nil {
			logs.Log.Error(err.Error())

		}
		if len(blockData.Result.Block.Data.Txs) > 0 {
			StoreWasmTransaction(blockData.Result.Block.Data.Txs, txnDB, JsonAPI)
		}

		var responseMap map[string]interface{}
		if err := json.Unmarshal(body, &responseMap); err != nil {
			logs.Log.Error("Error While Unmarshalling " + err.Error())

		}
		if result, ok := responseMap["result"]; ok {
			resultJSON, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				logs.Log.Error("Error marshalling JSON" + err.Error())
			}

			blockKey := []byte("Block" + strconv.Itoa(i))

			if err = db.Put(blockKey, resultJSON, nil); err != nil {
				logs.Log.Error(err.Error())
			}

			blockCount := i + 1
			err = db.Put([]byte("blockCount"), []byte(strconv.Itoa(blockCount)), nil)
			if err != nil {
				logs.Log.Error(fmt.Sprintf("Error in saving latestBlock in block db : %s", err.Error()))
			}
		}

	}
}

func GetWasmCurrentBlock(JsonAPI string) (BlockObject, error) {
	rpcUrl := fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/blocks/latest", JsonAPI)
	res, err := http.Get(rpcUrl)
	if err != nil {
		return BlockObject{}, err
	}
	defer res.Body.Close()

	var data BlockObject
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return BlockObject{}, err
	}

	return data, nil
}

func watchWasmBlocks(JsonRPC string, JsonAPI string, currentBlockHeight int, db *leveldb.DB, txnDB *leveldb.DB) {
	var currentBlock BlockObject
	for {
		latestBlock, err := GetWasmCurrentBlock(JsonAPI)
		if err != nil {
			logs.Log.Debug(err.Error())
			time.Sleep(2 * time.Second)
			continue
		}

		if currentBlock.Block.Header.Height == latestBlock.Block.Header.Height {
			time.Sleep(7 * time.Second)
			continue
		}
		logs.Log.Info("New blocks Found")
		rpcUrl := fmt.Sprintf("%s/block?height=%s", JsonRPC, latestBlock.Block.Header.Height)
		resp, err := http.Get(rpcUrl)
		if err != nil {
			logs.Log.Error(err.Error())
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			logs.Log.Error(err.Error())
			continue
		}

		var blockData Response

		err = json.Unmarshal(body, &blockData)
		if err != nil {
			logs.Log.Error(err.Error())
		}

		if len(blockData.Result.Block.Data.Txs) > 0 {

			StoreWasmTransaction(blockData.Result.Block.Data.Txs, txnDB, JsonAPI)
		}

		var responseMap map[string]interface{}
		if err := json.Unmarshal(body, &responseMap); err != nil {
			log.Fatal("Error unmarshalling JSON:", err)
		}

		if result, ok := responseMap["result"]; ok {
			resultJSON, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				log.Fatal("Error marshalling JSON:", err)
			}

			blockKey := []byte("Block" + latestBlock.Block.Header.Height)
			if err = db.Put(blockKey, resultJSON, nil); err != nil {
				log.Println("Error saving block to LevelDB:", err)
			}

			height, err := strconv.Atoi(latestBlock.Block.Header.Height)
			if err != nil {
				log.Fatal("Error converting block height to integer:", err)
			}
			blockCount := height + 1
			err = db.Put([]byte("blockCount"), []byte(strconv.Itoa(blockCount)), nil)
			if err != nil {
				logs.Log.Error(fmt.Sprintf("Error in saving latestBlock in block db : %s", err.Error()))
			}
		} else {
			fmt.Println("Result key not found in response")
		}

		currentBlock = latestBlock
	}
}

func NewWasmBlocks(JsonRPC string, JsonAPI string, currentBlock int, db *leveldb.DB, txnDB *leveldb.DB) {
	watchWasmBlocks(JsonRPC, JsonAPI, currentBlock, db, txnDB)
}
