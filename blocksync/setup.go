package blocksync

import (
	"encoding/json"
	"fmt"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/types"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"os"
	"path/filepath"
)

var txDbInstance *leveldb.DB
var blockDbInstance *leveldb.DB
var staticDbInstance *leveldb.DB
var stateDbInstance *leveldb.DB
var batchesDbInstance *leveldb.DB
var proofDbInstance *leveldb.DB
var publicWitnessDbInstance *leveldb.DB
var daDbInstance *leveldb.DB
var mockDbInstance *leveldb.DB

// InitTxDb This function initializes a LevelDB database for transactions and returns a boolean indicating
// whether the initialization was successful.
func InitTxDb() bool {
	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, ".tracks/data/leveldb/tx")

	txDB, err := leveldb.OpenFile(filePath, nil)
	if err != nil {
		log.Fatal("Failed to open transaction LevelDB:", err)
		return false
	}
	txDbInstance = txDB

	txnNumberByte, err := txDbInstance.Get([]byte("txnCount"), nil)
	if txnNumberByte == nil || err != nil {
		err = txDbInstance.Put([]byte("txnCount"), []byte("0"), nil)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in saving txnCount in txnDb : %s", err.Error()))
			//return false
			os.Exit(0)
		}
	}

	return true

}

// InitBlockDb This function initializes a LevelDB database for storing blocks and returns a boolean indicating
// whether the initialization was successful.
func InitBlockDb() bool {
	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, ".tracks/data/leveldb/blocks")
	blockDB, err := leveldb.OpenFile(filePath, nil)
	if err != nil {
		log.Fatal("Failed to open block LevelDB:", err)
		return false
	}

	blockDbInstance = blockDB

	// get

	blockNumberByte, err := blockDB.Get([]byte("blockCount"), nil)

	if blockNumberByte == nil || err != nil {
		err = blockDB.Put([]byte("blockCount"), []byte("0"), nil)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in saving blockCount in blockDatabase : %s", err.Error()))
			//return false
			os.Exit(0)
		}
	}

	return true
}

// InitStaticDb This function initializes a static LevelDB database and returns a boolean indicating whether the
// initialization was successful or not.
func InitStaticDb() bool {
	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, ".tracks/data/leveldb/static")
	staticDB, err := leveldb.OpenFile(filePath, nil)
	if err != nil {
		log.Fatal("Failed to open static LevelDB:", err)
		return false
	}
	staticDbInstance = staticDB
	return true
}

func InitStateDb() bool {
	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, ".tracks/data/leveldb/state")
	stateDB, err := leveldb.OpenFile(filePath, nil)
	if err != nil {
		log.Fatal("Failed to open state LevelDB:", err)
		return false
	}

	stateDbInstance = stateDB

	podStateByte, err := stateDB.Get([]byte("podState"), nil)
	if podStateByte == nil || err != nil {

		emptyPodState := types.PodState{
			LatestPodHeight:     1,
			LatestTxState:       "PreInit",
			LatestPodHash:       nil,
			PreviousPodHash:     nil,
			LatestPodProof:      nil,
			LatestPublicWitness: nil,
			Votes:               make(map[string]types.Votes),
			TracksAppHash:       nil,
			Batch:               nil,
			MasterTrackAppHash:  nil,
		}
		byteEmptyPodState, err := json.Marshal(emptyPodState)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in marshalling emptyPodState : %s", err.Error()))
			return false
		}

		err = stateDB.Put([]byte("podState"), byteEmptyPodState, nil)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in saving podState in pod database : %s", err.Error()))
			return false
		}

	}

	return true
}

// InitBatchesDb This function initializes a batches LevelDB database and returns a boolean indicating whether the
// initialization was successful or not.
func InitBatchesDb() bool {
	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, ".tracks/data/leveldb/batches")
	batchesDB, err := leveldb.OpenFile(filePath, nil)
	if err != nil {
		log.Fatal("Failed to open batches LevelDB:", err)
		return false
	}
	batchesDbInstance = batchesDB
	return true
}

// InitProofDb This function initializes a proof LevelDB database and returns a boolean indicating whether the
// initialization was successful or not.
func InitProofDb() bool {
	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, ".tracks/data/leveldb/proof")
	proofDB, err := leveldb.OpenFile(filePath, nil)
	if err != nil {
		log.Fatal("Failed to open proof LevelDB:", err)
		return false
	}
	proofDbInstance = proofDB
	return true
}

func InitPublicWitnessDb() bool {
	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, ".tracks/data/leveldb/publicWitness")
	publicWitnessDB, err := leveldb.OpenFile(filePath, nil)
	if err != nil {
		log.Fatal("Failed to open publicWitness LevelDB:", err)
		return false
	}
	publicWitnessDbInstance = publicWitnessDB
	return true
}

func InitDaDb() bool {
	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, ".tracks/data/leveldb/da")
	daDB, err := leveldb.OpenFile(filePath, nil)
	da := types.DAStruct{
		DAKey:             "0",
		DAClientName:      "0",
		BatchNumber:       "0",
		PreviousStateHash: "0",
		CurrentStateHash:  "0",
	}

	daBytes, err := json.Marshal(da)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling da : %s", err.Error()))
		return false
	}

	daDbInstance = daDB
	daBytes, err = daDbInstance.Get([]byte("batch_0"), nil)
	if daBytes == nil || err != nil {
		err = daDbInstance.Put([]byte("batch_0"), daBytes, nil)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in saving daBytes in da Database : %s", err.Error()))
			return false
		}
	}

	return true
}
func InitMockDb() bool {

	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, ".tracks/data/leveldb/mock")
	mockDB, err := leveldb.OpenFile(filePath, nil)
	if err != nil {
		log.Fatal("Failed to open mock LevelDB:", err)
		return false
	}
	mockDbInstance = mockDB
	return true
}

// InitDb This function  initializes three different databases and returns true if all of them are
// successfully initialized, otherwise it returns false.
func InitDb() bool {
	if !InitTxDb() {
		return false
	}
	if !InitBlockDb() {
		return false
	}
	if !InitStaticDb() {
		return false
	}
	if !InitStateDb() {
		return false
	}
	if !InitBatchesDb() {
		return false
	}
	if !InitProofDb() {
		return false
	}
	if !InitPublicWitnessDb() {
		return false
	}
	if !InitDaDb() {
		return false
	}
	if !InitMockDb() {
		return false
	}
	return true
}

// GetTxDbInstance This function returns the instance of the air-leveldb database.
func GetTxDbInstance() *leveldb.DB {
	return txDbInstance
}

// GetBlockDbInstance This function returns the instance of the block database.
func GetBlockDbInstance() *leveldb.DB {
	return blockDbInstance
}

// GetStaticDbInstance This function  is returning the instance of the LevelDB database that was
// initialized in the InitStaticDb function. This allows other parts of the code to access and use
// the LevelDB database instance for performing operations such as reading or writing data.
func GetStaticDbInstance() *leveldb.DB {
	return staticDbInstance
}

func GetStateDbInstance() *leveldb.DB {
	return stateDbInstance
}

// GetBatchesDbInstance This function  is returning the instance of the LevelDB database that was
// initialized in the InitBatchesDb function. This allows other parts of the code to access and use
// the LevelDB database instance for performing operations such as reading or writing data.
func GetBatchesDbInstance() *leveldb.DB {
	return batchesDbInstance
}

// GetProofDbInstance This function  is returning the instance of the LevelDB database that was
// initialized in the InitProofDb function. This allows other parts of the code to access and use
// the LevelDB database instance for performing operations such as reading or writing data.
func GetProofDbInstance() *leveldb.DB {
	return proofDbInstance
}

func GetPublicWitnessDbInstance() *leveldb.DB {
	return publicWitnessDbInstance
}

func GetDaDbInstance() *leveldb.DB {
	return daDbInstance
}

func GetMockDbInstance() *leveldb.DB {
	return mockDbInstance
}
