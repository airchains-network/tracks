package node

import (
	"context"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/blocksync"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/airchains-network/decentralized-sequencer/p2p"
	"github.com/airchains-network/decentralized-sequencer/rpc"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
	"time"
)

func init() {
	viper.SetConfigName("sequencer")            // name of config file (without extension)
	viper.SetConfigType("toml")                 // REQUIRED if the config file does not have the extension in the name // path to look for the config file in
	viper.AddConfigPath("$HOME/.tracks/config") // call multiple times to add many search paths 	// optionally look for config in the working directory
	err := viper.ReadInConfig()                 // Find and read the config file
	if err != nil {                             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func Start() {
	var wg1 sync.WaitGroup
	wg1.Add(2)
	go configureP2P(&wg1)
	go time.AfterFunc(3*time.Second, func() {
		beginDBIndexingOperations(&wg1)
	})
	wg1.Wait()
}
func configureP2P(wg *sync.WaitGroup) {
	defer wg.Done()
	p2p.P2PConfiguration()
}
func beginDBIndexingOperations(wg *sync.WaitGroup) {
	fmt.Println("Connected to the network. Starting the indexing process...")

	defer wg.Done()
	connection := shared.Node.NodeConnections

	staticDB := connection.GetStaticDatabaseConnection()
	blockDB := connection.GetBlockDatabaseConnection()
	txnDB := connection.GetTxnDatabaseConnection()

	shared.CheckAndInitializeDBCounters(staticDB)
	latestBlock := shared.GetLatestBlock(blockDB)
	fmt.Println("This is the JSON ", viper.GetString("station.stationRPC"))
	fmt.Println("This is the Station Type ", viper.GetString("station.stationType"))
	client, err := ethclient.Dial(viper.GetString("station.stationRPC"))
	if err != nil {
		logs.Log.Error("Error in connecting to the network")
		return
	}

	initializeCounter(staticDB, "batchCount")
	initializeCounter(staticDB, "batchStartIndex")

	var ctx context.Context
	ctx = context.Background()
	var wgnm *sync.WaitGroup
	wgnm = &sync.WaitGroup{}
	wgnm.Add(3)

	go blocksync.StartIndexer(wgnm, client, ctx, blockDB, txnDB, latestBlock)
	go p2p.BatchGeneration(wgnm)
	go rpc.StartRPC(wgnm)
	wgnm.Wait()
}

func initializeCounter(staticDB *leveldb.DB, counterName string) {
	_, err := staticDB.Get([]byte(counterName), nil)
	if err != nil {
		err = staticDB.Put([]byte(counterName), []byte("0"), nil)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in saving %s in static db: %s", counterName, err.Error()))
		}
	}
}
