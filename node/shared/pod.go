// In shared/core package
package shared

import (
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/blocksync"
	"github.com/airchains-network/decentralized-sequencer/config"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/syndtr/goleveldb/leveldb"
	"strconv"
	"strings"
	"sync"
)

var (
	Node *NodeS
	mu   sync.Mutex
)

type Votes struct {
	PeerID string // TODO change this type to proper Peer ID Type
	//Commitment string
	Vote bool
}
type PodState struct {
	LatestPodHeight     uint64
	LatestPodHash       []byte
	LatestPodProof      []byte
	LatestPublicWitness []byte
	Votes               map[string]Votes
	TracksAppHash       []byte
	Batch               *types.BatchStruct
}
type Connections struct {
	mu                                 sync.Mutex
	BlockDatabaseConnection            *leveldb.DB
	TxnDatabaseConnection              *leveldb.DB
	PodsDatabaseConnection             *leveldb.DB
	DataAvailabilityDatabaseConnection *leveldb.DB
	StaticDatabaseConnection           *leveldb.DB
}

type NodeS struct {
	config          *config.Config
	podState        *PodState
	NodeConnections *Connections
}

func InitializePodState() *PodState {
	return &PodState{
		LatestPodHeight:     0,
		LatestPodHash:       nil,
		LatestPodProof:      nil,
		LatestPublicWitness: nil,
		Votes:               make(map[string]Votes),
		TracksAppHash:       nil,
		Batch:               nil,
	}
}
func GetPodState() *PodState {
	mu.Lock()
	defer mu.Unlock()
	//fmt.Println(Node.podState)
	return Node.podState
}

func SetPodState(podState *PodState) {
	mu.Lock()
	defer mu.Unlock()
	Node.podState = podState
}

func InitializeDatabaseConnections() *Connections {
	return &Connections{
		BlockDatabaseConnection:            blocksync.GetBlockDbInstance(),
		TxnDatabaseConnection:              blocksync.GetTxDbInstance(),
		PodsDatabaseConnection:             blocksync.GetBatchesDbInstance(),
		DataAvailabilityDatabaseConnection: blocksync.GetDaDbInstance(),
		StaticDatabaseConnection:           blocksync.GetStaticDbInstance(),
	}
}
func (c *Connections) GetBlockDatabaseConnection() *leveldb.DB {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.BlockDatabaseConnection
}

func (c *Connections) GetTxnDatabaseConnection() *leveldb.DB {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.TxnDatabaseConnection
}

func (c *Connections) GetPodsDatabaseConnection() *leveldb.DB {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.PodsDatabaseConnection
}

func (c *Connections) GetDataAvailabilityDatabaseConnection() *leveldb.DB {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.DataAvailabilityDatabaseConnection
}

func (c *Connections) GetStaticDatabaseConnection() *leveldb.DB {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.StaticDatabaseConnection
}

func CheckAndInitializeDBCounters(staticDB *leveldb.DB) {
	ensureCounter(staticDB, "batchStartIndex")
	ensureCounter(staticDB, "batchCount")
}

func ensureCounter(db *leveldb.DB, counterKey string) {
	fmt.Println(db)
	if _, err := db.Get([]byte(counterKey), nil); err != nil {
		if err = db.Put([]byte(counterKey), []byte("0"), nil); err != nil {
			logs.Log.Error(fmt.Sprintf("Error in saving %s in static db: %s", counterKey, err.Error()))
			return
		}
	}
}

func GetLatestBlock(blockDB *leveldb.DB) int {
	latestBlockBytes, err := blockDB.Get([]byte("blockCount"), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in getting blockCount from block db: %s", err.Error()))
		return 0
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
			return nil
		}
	}
	return batchStartIndex
}

func NewNode(conf *config.Config) {
	if !blocksync.InitDb() {
		logs.Log.Error("Error in initializing db")
		return
	}
	logs.Log.Info("Initialized the database")

	NodeConnections := InitializeDatabaseConnections()
	podState := InitializePodState()
	Node = &NodeS{
		config:          conf,
		podState:        podState,
		NodeConnections: NodeConnections,
	}
}

//func GetConfig() *config.Config {
//	return Node.config
//}
