package blocksync

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
)

func StartIndexer(wg *sync.WaitGroup, client *ethclient.Client, ctx context.Context, blockDatabaseConnection *leveldb.DB, txnDatabaseConnection *leveldb.DB, latestBlock int) {
	defer wg.Done()

	StoreEVMBlock(client, ctx, latestBlock, blockDatabaseConnection, txnDatabaseConnection)

}
