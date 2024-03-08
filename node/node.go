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
	"os"
	"strconv"
	"strings"
	"sync"
)

func Node() {
	connectResult := p2p.P2PConfiguration()
	if connectResult {
		response := blocksync.InitDb()
		if !response {
			logs.Log.Error("Error in initializing db")
		}
		logs.Log.Info("Initialized the database")
		ctx := context.Background()
		var (
			BlockDatabaseConnection            = blocksync.GetBlockDbInstance()
			TxnDatabaseConnection              = blocksync.GetTxDbInstance()
			PodsDatabaseConnection             = blocksync.GetBatchesDbInstance()
			DataAvailabilityDatabaseConnection = blocksync.GetDaDbInstance()
			StaticDatabaseConnection           = blocksync.GetStaticDbInstance()
		)
		fmt.Println("staticDatabaseConnection", StaticDatabaseConnection)

		batchStartIndex, err := StaticDatabaseConnection.Get([]byte("batchStartIndex"), nil)

		if err != nil {
			err = StaticDatabaseConnection.Put([]byte("batchStartIndex"), []byte("0"), nil)
			if err != nil {
				logs.Log.Error(fmt.Sprintf("Error in saving batchStartIndex in static db : %s", err.Error()))
				os.Exit(0)
			}
		}

		_, err = StaticDatabaseConnection.Get([]byte("batchCount"), nil)
		if err != nil {
			err = StaticDatabaseConnection.Put([]byte("batchCount"), []byte("0"), nil)
			if err != nil {
				logs.Log.Error(fmt.Sprintf("Error in saving batchCount in static db : %s", err.Error()))
				os.Exit(0)
			}
		}

		latestBlockBytes, err := BlockDatabaseConnection.Get([]byte("blockCount"), nil)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting blockCount from block db : %s", err.Error()))
			os.Exit(0) //TODO : Handle this error
		}

		latestBlock, _ := strconv.Atoi(strings.TrimSpace(string(latestBlockBytes)))
		fmt.Println("latestBlock", latestBlock)

		client, err := ethclient.Dial(stationConfig.StationRPC)
		if err != nil {
			fmt.Println("Error in connecting to the network")
		}
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			blocksync.StartIndexer(&wg, client, ctx, BlockDatabaseConnection, TxnDatabaseConnection, latestBlock)
		}()
		go func() {
			defer wg.Done()
			pods.BatchGeneration(&wg, client, ctx, StaticDatabaseConnection, TxnDatabaseConnection, PodsDatabaseConnection, DataAvailabilityDatabaseConnection, batchStartIndex)
		}()
		wg.Wait()
	} else {
		logs.Log.Error("Failed to connect to the network")
	}
}
