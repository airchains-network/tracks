package blocksync

import (
	"context"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/config"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"path/filepath"
	"sync"
)

func StartIndexer(wg *sync.WaitGroup, client *ethclient.Client, ctx context.Context, blockDatabaseConnection *leveldb.DB, txnDatabaseConnection *leveldb.DB, latestBlock int) {
	wg.Done()
	bsgConfig, err := LoadConfig()
	if err != nil {
		fmt.Println(err)
	}

	if bsgConfig.Station.StationType == "EVM" || bsgConfig.Station.StationType == "evm" {
		StoreEVMBlock(client, ctx, latestBlock, blockDatabaseConnection, txnDatabaseConnection)
	} else if bsgConfig.Station.StationType == "WASM" || bsgConfig.Station.StationType == "wasm" {
		JsonRPC := bsgConfig.Station.StationRPC
		JsonAPI := bsgConfig.Station.StationAPI
		StoreWasmBlock(blockDatabaseConnection, txnDatabaseConnection, JsonRPC, JsonAPI)
	} else if bsgConfig.Station.StationType == "SVM" || bsgConfig.Station.StationType == "svm" {
		JsonRPC := bsgConfig.Station.StationRPC
		JsonAPI := bsgConfig.Station.StationAPI
		StoreSVMBlock(blockDatabaseConnection, txnDatabaseConnection, JsonRPC, JsonAPI)
	}

}

func LoadConfig() (config config.Config, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config, err // Return error, perhaps log it as well
	}
	configDir := filepath.Join(homeDir, ".tracks/config")

	_, err = os.Stat(configDir)
	if os.IsNotExist(err) {
		return config, fmt.Errorf("config directory not found: %s", configDir)
	}

	viper.AddConfigPath(configDir)
	viper.SetConfigName("sequencer")
	viper.SetConfigType("toml")

	if err = viper.ReadInConfig(); err != nil {
		return config, err
	}

	err = viper.Unmarshal(&config)
	return config, err
}
