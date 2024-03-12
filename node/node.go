package node

import (
	"context"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/blocksync"
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

/*
type Node struct {
	service.BaseService

	// config
	config        *cfg.Config
	genesisDoc    *types.GenesisDoc   // initial validator set
	privValidator types.PrivValidator // local node's validator key

	// network
	transport   *p2p.MultiplexTransport
	sw          *p2p.Switch  // p2p connections
	addrBook    pex.AddrBook // known peers
	nodeInfo    p2p.NodeInfo
	nodeKey     *p2p.NodeKey // our node privkey
	isListening bool

	// services
	eventBus          *types.EventBus // pub/sub for services
	stateStore        sm.Store
	blockStore        *store.BlockStore // store the blockchain to disk
	bcReactor         p2p.Reactor       // for block-syncing
	mempoolReactor    p2p.Reactor       // for gossipping transactions
	mempool           mempl.Mempool
	stateSync         bool                    // whether the node should state sync on startup
	stateSyncReactor  *statesync.Reactor      // for hosting and restoring state sync snapshots
	stateSyncProvider statesync.StateProvider // provides state data for bootstrapping a node
	stateSyncGenesis  sm.State                // provides the genesis state for state sync
	consensusState    *cs.State               // latest consensus state
	consensusReactor  *cs.Reactor             // for participating in the consensus
	pexReactor        *pex.Reactor            // for exchanging peer addresses
	evidencePool      *evidence.Pool          // tracking evidence
	proxyApp          proxy.AppConns          // connection to the application
	rpcListeners      []net.Listener          // rpc servers
	txIndexer         txindex.TxIndexer
	blockIndexer      indexer.BlockIndexer
	indexerService    *txindex.IndexerService
	prometheusSrv     *http.Server
	pprofSrv          *http.Server
}
*/

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
	client, err := ethclient.Dial("localhost:8545")
	if err != nil {
		logs.Log.Error("Error in connecting to the network")
		return
	}

	//DefaultUnverifiedData  define default pod data\

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
