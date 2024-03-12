package node

import (
	"context"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/config"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/pods"
	"strconv"
	"strings"
	"sync"

	"github.com/airchains-network/decentralized-sequencer/blocksync"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/airchains-network/decentralized-sequencer/p2p"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/syndtr/goleveldb/leveldb"
)

type NodeS struct {
	config     *config.Config
	dbProvider Connections
	podState   shared.PodState
}

func InitializePodState() shared.PodState {
	return shared.PodState{
		LatestPodHeight:         0,
		LatestPodMerkleRootHash: nil,
		LatestPodProof:          nil,
		LatestPublicWitness:     nil,
		Votes:                   make(map[string]shared.Votes),
	}
}

func (n *NodeS) GetPodState() shared.PodState {
	return n.podState
}

func (n *NodeS) SetPodState(podState shared.PodState) {
	n.podState = podState
}

// NewNode creates a new NodeS instance with the necessary initializations.
func NewNode(config *config.Config, dbProvider Connections, podState shared.PodState) *NodeS {
	return &NodeS{config, dbProvider, podState}
}

func (n *NodeS) Start() {
	var wg1 sync.WaitGroup
	wg1.Add(2)

	go n.configureP2P(&wg1)
	go n.initializeDBAndStartIndexing(&wg1)

	wg1.Wait()
}

func (n *NodeS) configureP2P(wg *sync.WaitGroup) {
	defer wg.Done()
	p2p.P2PConfiguration()
}

func (n *NodeS) initializeDBAndStartIndexing(wg *sync.WaitGroup) {
	defer wg.Done()

	if !blocksync.InitDb() {
		logs.Log.Error("Error in initializing db")
		return
	}
	logs.Log.Info("Initialized the database")

	//dbConnections :=
	staticDB := n.dbProvider.StaticDatabaseConnection
	n.checkAndInitializeDBCounters(staticDB)

	latestBlock := n.getLatestBlock(n.dbProvider.BlockDatabaseConnection)
	client, err := ethclient.Dial("localhost:8545")
	if err != nil {
		logs.Log.Error("Error in connecting to the network")
		return
	}
	var ctx context.Context
	var wgnm *sync.WaitGroup
	wgnm.Add(2)

	go blocksync.StartIndexer(wgnm, client, context.Background(), n.dbProvider.BlockDatabaseConnection, n.dbProvider.TxnDatabaseConnection, latestBlock)
	go pods.BatchGeneration(wgnm, client, ctx, n.dbProvider.StaticDatabaseConnection, n.dbProvider.TxnDatabaseConnection, n.dbProvider.PodsDatabaseConnection, n.dbProvider.DataAvailabilityDatabaseConnection, n.GetLatestBatchIndex(staticDB))

	wgnm.Wait()
}

// Database connectiobnns
type Connections struct {
	BlockDatabaseConnection            *leveldb.DB
	TxnDatabaseConnection              *leveldb.DB
	PodsDatabaseConnection             *leveldb.DB
	DataAvailabilityDatabaseConnection *leveldb.DB
	StaticDatabaseConnection           *leveldb.DB
}

func InitializeDatabaseConnections() (connections Connections) { //} Connections {
	//var connections Connections
	connections.BlockDatabaseConnection = blocksync.GetBlockDbInstance()
	connections.TxnDatabaseConnection = blocksync.GetTxDbInstance()
	connections.PodsDatabaseConnection = blocksync.GetBatchesDbInstance()
	connections.DataAvailabilityDatabaseConnection = blocksync.GetDaDbInstance()
	connections.StaticDatabaseConnection = blocksync.GetStaticDbInstance()

	//n.dbProvider = connections
	return connections
}

func (n *NodeS) checkAndInitializeDBCounters(staticDB *leveldb.DB) {
	n.ensureCounter(staticDB, "batchStartIndex")
	n.ensureCounter(staticDB, "batchCount")
}

func (n *NodeS) ensureCounter(db *leveldb.DB, counterKey string) {
	if _, err := db.Get([]byte(counterKey), nil); err != nil {
		if err = db.Put([]byte(counterKey), []byte("0"), nil); err != nil {
			logs.Log.Error(fmt.Sprintf("Error in saving %s in static db: %s", counterKey, err.Error()))
			return
		}
	}
}

func (n *NodeS) getLatestBlock(blockDB *leveldb.DB) int {
	latestBlockBytes, err := blockDB.Get([]byte("blockCount"), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in getting blockCount from block db: %s", err.Error()))
		return 0
	}
	latestBlock, _ := strconv.Atoi(strings.TrimSpace(string(latestBlockBytes)))
	return latestBlock
}

func (n *NodeS) GetLatestBatchIndex(staticDB *leveldb.DB) []byte {
	batchStartIndex, err := staticDB.Get([]byte("batchStartIndex"), nil)
	if err != nil {
		err = staticDB.Put([]byte("batchStartIndex"), []byte("0"), nil)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in saving batchStartIndex in static db : %s", err.Error()))
			return nil
		}
	}
	return batchStartIndex
}
