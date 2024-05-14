package blocksync

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"time"

	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/airchains-network/decentralized-sequencer/utils"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/syndtr/goleveldb/leveldb"
)

func StoreEVMBlock(client *ethclient.Client, ctx context.Context, blockIndex int, ldb *leveldb.DB, ldt *leveldb.DB) {
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	blockData, err := client.BlockByNumber(ctx, big.NewInt(int64(blockIndex)))
	if err != nil {

		// errMessage := fmt.Sprintf("Failed to get block data for block number %d: %s", blockIndex, err)
		// log.Warn().Str("module", "blocksync").Err(err).Msg(errMessage)
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
		log.Error().Str("module", "blocksync").Err(err).Msg(errMessage)

	}
	key := fmt.Sprintf("block_%s", block.Number)
	err = ldb.Put([]byte(key), data, nil)
	if err != nil {
		errMessage := fmt.Sprintf("Error inserting block data into database: %s", err)
		log.Error().Str("module", "blocksync").Err(err).Msg(errMessage)
	}

	transactions := blockData.Transactions()
	if transactions == nil {
		log.Info().Str("module", "blocksync").Msg(fmt.Sprintf("No Txn In  %s", block.Number))
	}

	for i := 0; i < block.TransactionCount; i++ {
		StoreEVMTransactions(client, ctx, ldt, transactions[i].Hash().String(), blockIndex, block.Hash)
	}

	blockCount := blockIndex + 1
	err = ldb.Put([]byte("blockCount"), []byte(strconv.Itoa(blockCount)), nil)
	if err != nil {
		errMessage := fmt.Sprintf("Error inserting block count into database: %s", err)
		log.Error().Str("module", "blocksync").Err(err).Msg(errMessage)
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
		log.Error().Str("module", "blocksync").Err(resErr).Msg("")
	}
	defer res.Body.Close()
	bodyBlockHeight, bodyBlockHeightErr := io.ReadAll(res.Body)
	if bodyBlockHeightErr != nil {
		log.Error().Str("module", "blocksync").Err(bodyBlockHeightErr).Msg("")
	}

	var blockHeight BlockObject
	error := json.Unmarshal(bodyBlockHeight, &blockHeight)
	if error != nil {
		log.Error().Str("module", "blocksync").Err(error).Msg("")
	}
	latestBlock := blockHeight.Block.Header.Height

	numLatestBlock, err := strconv.Atoi(latestBlock)
	if err != nil {
		log.Error().Str("module", "blocksync").Err(err).Msg("")
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
			log.Error().Str("module", "blocksync").Err(jsonErr).Msg("")
		}
		if len(blockData.Result.Block.Data.Txs) > 0 {
			StoreWasmTransaction(blockData.Result.Block.Data.Txs, txnDB, JsonAPI)
		}

		var responseMap map[string]interface{}
		if err := json.Unmarshal(body, &responseMap); err != nil {
			logs.Log.Error(fmt.Sprintf("Error in response: %v", err))
			continue
		}
		if result, ok := responseMap["result"]; ok {
			resultJSON, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				log.Error().Str("module", "blocksync").Err(err).Msg("")
			}

			blockKey := []byte("Block" + strconv.Itoa(i))

			if err = db.Put(blockKey, resultJSON, nil); err != nil {
				log.Error().Str("module", "blocksync").Err(err).Msg("")
			}

			blockCount := i + 1
			err = db.Put([]byte("blockCount"), []byte(strconv.Itoa(blockCount)), nil)
			if err != nil {
				log.Error().Str("module", "blocksync").Err(err).Msg("")
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
		log.Info().Str("module", "blocksync").Msg("New Block Found")
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
			log.Fatal().Str("body", string(body)).Msg(err.Error())
		}

		if result, ok := responseMap["result"]; ok {
			resultJSON, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				log.Fatal().Str("body", string(body)).Msg(err.Error())
			}

			blockKey := []byte("Block" + latestBlock.Block.Header.Height)
			if err = db.Put(blockKey, resultJSON, nil); err != nil {

			}

			height, err := strconv.Atoi(latestBlock.Block.Header.Height)
			if err != nil {
				log.Error().Str("module", "blocksync").Err(err).Msg("")
				continue
			}
			blockCount := height + 1
			err = db.Put([]byte("blockCount"), []byte(strconv.Itoa(blockCount)), nil)
			if err != nil {

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

// * SVM chain

func StoreSVMBlock(ldb *leveldb.DB, ldt *leveldb.DB, JsonRPC, JsonAPI string) {
	initSVMRPC(JsonRPC)

	latestIndex, latestIndexErr := SVMLatestBlockCheck()
	if latestIndexErr != nil {
		log.Error().Str("module", "blocksync").Err(latestIndexErr).Msg("")
		//return fmt.Errorf("error while fetching latest block: %v", latestIndexErr)
	}

	lastBlock := getLastProcessedBlock(ldb)
	startBlock := lastBlock + 1

	OldSVMBlock(startBlock, latestIndex, ldb, ldt)
	NewSVMBlock(latestIndex, ldb, ldt)
}

func OldSVMBlock(startIndex, latestIndex int, ldb, ldt *leveldb.DB) {
	for i := startIndex; i <= latestIndex; i++ {
		SVMBlockStore(i, ldb, ldt)
		fmt.Println("old block store : ", i)
	}
}

func NewSVMBlock(currentIndex int, ldb, ldt *leveldb.DB) {
	for {
		latestIndex, latestIndexErr := SVMLatestBlockCheck()
		if latestIndexErr != nil {
			log.Error().Str("module", "blocksync").Err(latestIndexErr).Msg("")
			//return fmt.Errorf("error while fetching latest block: %v", latestIndexErr)
		}

		if currentIndex == latestIndex {
			fmt.Println("wait for new block...")
			time.Sleep(2 * time.Second)
			continue
		} else {
			for i := currentIndex; i < latestIndex; i++ {
				SVMBlockStore(latestIndex, ldb, ldt)
				fmt.Println("new block store : ", i)
			}
			currentIndex = latestIndex
		}
	}

}

func SVMBlockStore(blockNumber int, ldb, ldt *leveldb.DB) {
	res, resErr := SVMBlockCall(blockNumber)
	if resErr != nil {
		log.Error().Str("module", "blocksync").Err(resErr).Msg("")
	}

	resJson, err := json.Marshal(res.Result)
	if err != nil {
		log.Error().Str("module", "blocksync").Err(err).Msg("")
	}

	blockKey := []byte("Block" + strconv.Itoa(blockNumber))
	if err = ldb.Put(blockKey, resJson, nil); err != nil {
		log.Error().Str("module", "blocksync").Err(err).Msg("")
	}

	for i := 0; i < len(res.Result.Transactions); i++ {
		StoreSVMTransaction(ldt, res.Result.Transactions[i])
	}

	blockCount := blockNumber + 1
	err = ldb.Put([]byte("blockCount"), []byte(strconv.Itoa(blockCount)), nil)
	if err != nil {
		log.Error().Str("module", "blocksync").Err(err).Msg("")
	}
}
