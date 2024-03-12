package blocksync

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
)

func StartIndexer(wg *sync.WaitGroup, client *ethclient.Client, ctx context.Context, blockDatabaseConnection *leveldb.DB, txnDatabaseConnection *leveldb.DB, latestBlock int) {

	wg.Done()
	//Add  Cosmos and  SVM  Save Blocks also
	StoreEVMBlock(client, ctx, latestBlock, blockDatabaseConnection, txnDatabaseConnection)

}
