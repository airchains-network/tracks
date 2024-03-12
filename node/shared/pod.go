// In shared/core package
package shared

import (
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/blocksync"
	"github.com/airchains-network/decentralized-sequencer/config"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/syndtr/goleveldb/leveldb"
	"strconv"
	"strings"
)

var (
	Node *NodeS
)

type Votes struct {
	PeerID     string // TODO change this type to proper Peer ID Type
	Commitment string // this is the hash of the commitment that the peer has voted for the pod and the Signaturwe will done using thePrivate Keys of the Peers
	Vote       bool
}
type PodState struct {
	LatestPodHeight         int
	LatestPodMerkleRootHash []byte
	LatestPodProof          []byte
	LatestPublicWitness     []byte
	Votes                   map[string]Votes
}
type Connections struct {
	BlockDatabaseConnection            *leveldb.DB
	TxnDatabaseConnection              *leveldb.DB
	PodsDatabaseConnection             *leveldb.DB
	DataAvailabilityDatabaseConnection *leveldb.DB
	StaticDatabaseConnection           *leveldb.DB
}

type NodeS struct {
	config   *config.Config
	podState PodState
}

func InitializePodState() PodState {
	return PodState{
		LatestPodHeight:         0,
		LatestPodMerkleRootHash: nil,
		LatestPodProof:          nil,
		LatestPublicWitness:     nil,
		Votes:                   make(map[string]Votes),
	}
}
func GetPodState() PodState {
	fmt.Println(Node.podState)
	return Node.podState

}

func SetPodState(podState PodState) {
	Node.podState = podState
}

func GetConfig() *config.Config {
	return Node.config
}

func InitializeDatabaseConnections() (connections Connections) {

	connections.BlockDatabaseConnection = blocksync.GetBlockDbInstance()
	connections.TxnDatabaseConnection = blocksync.GetTxDbInstance()
	connections.PodsDatabaseConnection = blocksync.GetBatchesDbInstance()
	connections.DataAvailabilityDatabaseConnection = blocksync.GetDaDbInstance()
	connections.StaticDatabaseConnection = blocksync.GetStaticDbInstance()
	return connections
}

func CheckAndInitializeDBCounters(staticDB *leveldb.DB) {
	ensureCounter(staticDB, "batchStartIndex")
	ensureCounter(staticDB, "batchCount")
}

func ensureCounter(db *leveldb.DB, counterKey string) {
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

func NewNode(conf *config.Config, podState PodState) *NodeS {
	return &NodeS{
		config:   conf,
		podState: podState,
	}
}

//type PodStateManager interface {
//	UpdatePodState(podState PodState) error
//	GetPodState() PodState
//}
