package node

import (
	"context"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/pods"
	"github.com/syndtr/goleveldb/leveldb"

	//"fmt"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	//"strconv"
	//"strings"
	"sync"

	"github.com/airchains-network/decentralized-sequencer/blocksync"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/airchains-network/decentralized-sequencer/p2p"
	"github.com/ethereum/go-ethereum/ethclient"
	//"github.com/syndtr/goleveldb/leveldb"
)

func Start() {
	var wg1 sync.WaitGroup
	wg1.Add(2)

	go configureP2P(&wg1)
	go initializeDBAndStartIndexing(&wg1)

	wg1.Wait()
}

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
	dbProvider := InitializeDatabaseConnections()

	staticDB := dbProvider.StaticDatabaseConnection
	shared.CheckAndInitializeDBCounters(staticDB)

	latestBlock := shared.GetLatestBlock(dbProvider.BlockDatabaseConnection)
	client, err := ethclient.Dial("http://192.168.1.106:8545")
	if err != nil {
		logs.Log.Error("Error in connecting to the network")
		return
	}

	fmt.Println("HJey", dbProvider)

	var ctx context.Context
	ctx = context.Background()
	fmt.Println("HJey", ctx)
	var wgnm *sync.WaitGroup
	wgnm = &sync.WaitGroup{}
	wgnm.Add(2)
	latestBatch := shared.GetLatestBatchIndex(staticDB)

	go blocksync.StartIndexer(wgnm, client, ctx, dbProvider.BlockDatabaseConnection, dbProvider.TxnDatabaseConnection, latestBlock)
	go pods.BatchGeneration(wgnm, client, ctx, dbProvider.StaticDatabaseConnection, dbProvider.TxnDatabaseConnection, dbProvider.PodsDatabaseConnection, dbProvider.DataAvailabilityDatabaseConnection, latestBatch)

	wgnm.Wait()
}

// Database connectiobnns
type DatabaseConnections struct {
	BlockDatabaseConnection            *leveldb.DB
	TxnDatabaseConnection              *leveldb.DB
	PodsDatabaseConnection             *leveldb.DB
	DataAvailabilityDatabaseConnection *leveldb.DB
	StaticDatabaseConnection           *leveldb.DB
}

func InitializeDatabaseConnections() DatabaseConnections {
	var connections DatabaseConnections

	connections.BlockDatabaseConnection = blocksync.GetBlockDbInstance()

	connections.TxnDatabaseConnection = blocksync.GetTxDbInstance()

	connections.PodsDatabaseConnection = blocksync.GetBatchesDbInstance()

	connections.DataAvailabilityDatabaseConnection = blocksync.GetDaDbInstance()

	connections.StaticDatabaseConnection = blocksync.GetStaticDbInstance()
	fmt.Println(connections)
	return connections
}
