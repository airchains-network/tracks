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
	"os"
	"sync"
	"time"
)

func Start() {
	var wg1 sync.WaitGroup
	wg1.Add(2)

	go configureP2P(&wg1)
	go time.AfterFunc(3*time.Second, func() {
		initializeDBAndStartIndexing(&wg1)
	})

	wg1.Wait()
}

func configureP2P(wg *sync.WaitGroup) {
	defer wg.Done()
	p2p.P2PConfiguration()
}

func initializeDBAndStartIndexing(wg *sync.WaitGroup) {
	defer wg.Done()

	staticDB := shared.Node.NodeConnections.GetStaticDatabaseConnection()
	shared.CheckAndInitializeDBCounters(staticDB)

	latestBlock := shared.GetLatestBlock(shared.Node.NodeConnections.BlockDatabaseConnection)
	client, err := ethclient.Dial("http://192.168.1.106:8545")
	if err != nil {
		logs.Log.Error("Error in connecting to the network")
		return
	}

	_, err = staticDB.Get([]byte("batchCount"), nil)
	if err != nil {
		err = staticDB.Put([]byte("batchCount"), []byte("0"), nil)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in saving batchCount in static db : %s", err.Error()))
			os.Exit(0)
		}
	}

	_, err = staticDB.Get([]byte("batchStartIndex"), nil)
	if err != nil {
		fmt.Println("batchStartIndex not found")
		err = staticDB.Put([]byte("batchStartIndex"), []byte("0"), nil)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in saving batchStartIndex in static db : %s", err.Error()))
			os.Exit(0)
		}
	}

	var ctx context.Context
	ctx = context.Background()
	var wgnm *sync.WaitGroup
	wgnm = &sync.WaitGroup{}
	wgnm.Add(3)

	//go configureP2P(wgnm)
	go blocksync.StartIndexer(wgnm, client, ctx, shared.Node.NodeConnections.BlockDatabaseConnection, shared.Node.NodeConnections.TxnDatabaseConnection, latestBlock)
	go p2p.BatchGeneration(wgnm)
	go rpc.StartRPC(wgnm)

	wgnm.Wait()
}
