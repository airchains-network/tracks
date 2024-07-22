package node

import (
	"context"
	"fmt"
	"github.com/airchains-network/tracks/blocksync"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/node/shared"
	"github.com/airchains-network/tracks/p2p"
	"github.com/airchains-network/tracks/rpc"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
	"time"
)

func Start() {
	var wg1 sync.WaitGroup
	wg1.Add(2)
	go configureP2P(&wg1)

	go func() {
		time.Sleep(5 * time.Second)
		ticker := time.NewTicker(4 * time.Second) // adjust the check frequency as needed
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if p2p.PeerConnectionStatus(p2p.Node) {
					beginDBIndexingOperations(&wg1)
					return
				}
			}
		}
	}()
	wg1.Wait()
}

func configureP2P(wg *sync.WaitGroup) {
	defer wg.Done()
	p2p.P2PConfiguration()
}

func beginDBIndexingOperations(wg *sync.WaitGroup) {
	defer wg.Done()
	connection := shared.Node.NodeConnections
	staticDB := connection.GetStaticDatabaseConnection()
	blockDB := connection.GetBlockDatabaseConnection()
	txnDB := connection.GetTxnDatabaseConnection()
	shared.CheckAndInitializeDBCounters(staticDB)
	latestBlock := shared.GetLatestBlock(blockDB)
	baseConfig, err := shared.LoadConfig()

	client, err := ethclient.Dial(baseConfig.Station.StationRPC) // viper.GetString("station.stationRPC"))
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
	//wgnm.Add(1)
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
