package node

import (
	"context"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/blocksync"
	stationConfig "github.com/airchains-network/decentralized-sequencer/config"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/p2p"
	"github.com/airchains-network/decentralized-sequencer/pods"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"strconv"
	"strings"
	"sync"
)

func Node() {
	var wg1 sync.WaitGroup
	wg1.Add(2)

	go configureP2P(&wg1)
	go initializeDBAndStartIndexing(&wg1)

	wg1.Wait()
}

// Removed (data stationConfig.LatestUnverifiedData)
func configureP2P(wg *sync.WaitGroup) {
	defer wg.Done()
	p2p.P2PConfiguration()
}

func initializeDBAndStartIndexing(wg *sync.WaitGroup) {
	defer wg.Done()

	if !blocksync.InitDb() {
		logs.Log.Error("Error in initializing db")
		return
	}
	logs.Log.Info("Initialized the database")

	dbConnections := initializeDatabaseConnections()
	staticDB := dbConnections.StaticDatabaseConnection
	checkAndInitializeDBCounters(staticDB)

	latestBlock := getLatestBlock(dbConnections.BlockDatabaseConnection)
	client, err := ethclient.Dial(stationConfig.StationRPC)
	if err != nil {
		logs.Log.Error("Error in connecting to the network")
		return
	}

	//DefaultUnverifiedData  define default pod data
	pods.DefaultUnverifiedData()

	var wg2 sync.WaitGroup
	wg2.Add(2)
	go blocksync.StartIndexer(&wg2, client, context.Background(), dbConnections.BlockDatabaseConnection, dbConnections.TxnDatabaseConnection, latestBlock)
	go pods.BatchGeneration(&wg2, client, context.Background(), staticDB, dbConnections.TxnDatabaseConnection, dbConnections.PodsDatabaseConnection, dbConnections.DataAvailabilityDatabaseConnection, GetLatestBatchIndex(staticDB))

	wg2.Wait()
}

func initializeDatabaseConnections() (connections struct {
	BlockDatabaseConnection            *leveldb.DB
	TxnDatabaseConnection              *leveldb.DB
	PodsDatabaseConnection             *leveldb.DB
	DataAvailabilityDatabaseConnection *leveldb.DB
	StaticDatabaseConnection           *leveldb.DB
}) {
	connections.BlockDatabaseConnection = blocksync.GetBlockDbInstance()
	connections.TxnDatabaseConnection = blocksync.GetTxDbInstance()
	connections.PodsDatabaseConnection = blocksync.GetBatchesDbInstance()
	connections.DataAvailabilityDatabaseConnection = blocksync.GetDaDbInstance()
	connections.StaticDatabaseConnection = blocksync.GetStaticDbInstance()
	return
}

func checkAndInitializeDBCounters(staticDB *leveldb.DB) {
	ensureCounter(staticDB, "batchStartIndex")
	ensureCounter(staticDB, "batchCount")
}

func ensureCounter(db *leveldb.DB, counterKey string) {
	if _, err := db.Get([]byte(counterKey), nil); err != nil {
		if err = db.Put([]byte(counterKey), []byte("0"), nil); err != nil {
			logs.Log.Error(fmt.Sprintf("Error in saving %s in static db: %s", counterKey, err.Error()))
			os.Exit(0)
		}
	}
}

func getLatestBlock(blockDB *leveldb.DB) int {
	latestBlockBytes, err := blockDB.Get([]byte("blockCount"), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in getting blockCount from block db: %s", err.Error()))
		os.Exit(0) // Consider proper error handling instead of os.Exit
	}
	latestBlock, _ := strconv.Atoi(strings.TrimSpace(string(latestBlockBytes)))
	return latestBlock
}

func GetLatestBatchIndex(staticDB *leveldb.DB) []byte {
	batchStartIndex, err := staticDB.Get([]byte("batchStartIndex"), nil)

	if err != nil {
		err = staticDB.Put([]byte("batchStartIndex"), []byte("0"), nil)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in saving batchStartIndex in static db : %s", err.Error()))
			os.Exit(0)
		}
	}
	return batchStartIndex
}
