// In shared/core package
package shared

import (
	"encoding/json"
	"fmt"
	"github.com/airchains-network/tracks/blocksync"
	"github.com/airchains-network/tracks/config"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/types"
	"github.com/pelletier/go-toml"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	Node *NodeS
	mu   sync.Mutex

	// txStates
	TxStatePreInit   = "PreInit"
	TxStateInitVRF   = "InitVRF"
	TxStateVerifyVRF = "VerifyVRF"
	TxStateSubmitPod = "InitPod"
	TxStateVerifyPod = "VerifyPod"
)

type Votes struct {
	PeerID string // TODO change this type to proper Peer ID Type
	//Commitment string
	Vote bool
}
type PodState struct {
	LatestPodHeight     uint64
	LatestTxState       string // InitVRF / VerifyVRF / InitPod / VerifyPod
	LatestPodHash       []byte
	PreviousPodHash     []byte
	LatestPodProof      []byte
	LatestPublicWitness []byte
	Votes               map[string]Votes
	TracksAppHash       []byte
	Batch               *types.BatchStruct
	MasterTrackAppHash  []byte
	Timestamp           *time.Time `json:"timestamp,omitempty"`

	VRFInitiationTxHash string
	VRFValidationTxHash string
	InitPodTxHash       string
	VerifyPodTxHash     string
}

type TrackgatePodState struct {
	LatestPodHeight uint64
	LatestTxState   string // InitVRF / VerifyVRF / InitPod / VerifyPod
	LatestPodHash   []byte
	PreviousPodHash []byte
	Votes           map[string]Votes
	TracksAppHash   []byte
	Batch           *types.BatchStruct
	Timestamp       *time.Time `json:"timestamp,omitempty"`
}

type Connections struct {
	mu                                 sync.Mutex
	BlockDatabaseConnection            *leveldb.DB
	TxnDatabaseConnection              *leveldb.DB
	PodsDatabaseConnection             *leveldb.DB
	DataAvailabilityDatabaseConnection *leveldb.DB
	StaticDatabaseConnection           *leveldb.DB
	EspressoDatabaseConnection         *leveldb.DB
	StateDatabaseConnection            *leveldb.DB
	MockDatabaseConnection             *leveldb.DB
	PublicWitnessConnection            *leveldb.DB
}

type NodeS struct {
	Config          *config.Config
	podState        *PodState
	NodeConnections *Connections
}

func InitializePodState(stateConnection *leveldb.DB) *PodState {

	// sync pod state from database
	podStateByte, err := stateConnection.Get([]byte("podState"), nil)
	if err != nil {
		fmt.Println(err)
		logs.Log.Error("Pod should be already initiated/updated by now")
		os.Exit(0)
	}
	var podState *PodState
	err = json.Unmarshal(podStateByte, &podState)
	if err != nil {
		logs.Log.Error("Error in unmarshal  pod state")
		os.Exit(0)
	}
	return podState
}
func GetPodState() *PodState {
	mu.Lock()
	defer mu.Unlock()
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
		StateDatabaseConnection:            blocksync.GetStateDbInstance(),
		TxnDatabaseConnection:              blocksync.GetTxDbInstance(),
		PodsDatabaseConnection:             blocksync.GetBatchesDbInstance(),
		DataAvailabilityDatabaseConnection: blocksync.GetDaDbInstance(),
		StaticDatabaseConnection:           blocksync.GetStaticDbInstance(),
		EspressoDatabaseConnection:         blocksync.GetEspressoDbInstance(),
		MockDatabaseConnection:             blocksync.GetMockDbInstance(),
		PublicWitnessConnection:            blocksync.GetPublicWitnessDbInstance(),
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

func (c *Connections) GetEspressoDatabaseConnection() *leveldb.DB {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.EspressoDatabaseConnection
}

func (c *Connections) GetStateDatabaseConnection() *leveldb.DB {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.StateDatabaseConnection
}

func (c *Connections) GetPublicWitnessDbInstance() *leveldb.DB {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.PublicWitnessConnection
}

func CheckAndInitializeDBCounters(staticDB *leveldb.DB) {
	ensureCounter(staticDB, "batchStartIndex")
	ensureCounter(staticDB, "batchCount")
}

func ensureCounter(db *leveldb.DB, counterKey string) {
	//fmt.Println(db)
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

func NewNode(conf *config.Config) {

	NodeConnections := InitializeDatabaseConnections()
	stateConnection := NodeConnections.GetStateDatabaseConnection()
	podState := InitializePodState(stateConnection)

	Node = &NodeS{
		Config:          conf,
		podState:        podState,
		NodeConnections: NodeConnections,
	}
}

func LoadConfig() (cnf *config.Config, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("%v", err) // Return error, perhaps log it as well
	}
	configDir := filepath.Join(homeDir, config.DefaultTracksDir, config.DefaultConfigDir)

	_, err = os.Stat(configDir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("config directory not found: %s", configDir)
	}
	//
	//viper.AddConfigPath(configDir)
	//viper.SetConfigName("sequencer")
	//viper.SetConfigType("toml")
	//
	//if err = viper.ReadInConfig(); err != nil {
	//	return nil, err
	//}
	//
	//if err = viper.Unmarshal(&config); err != nil {
	//	return nil, err
	//}
	//
	//fmt.Println(config)
	ConfigFilePath := filepath.Join(configDir, config.DefaultConfigFileName)
	bytes, err := os.ReadFile(ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("error in reading config file: %s : %v", ConfigFilePath, err)
	}

	var conf config.Config // JunctionConfig
	if err = toml.Unmarshal(bytes, &conf); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %v", err)
	}

	return &conf, nil
}
