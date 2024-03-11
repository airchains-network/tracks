package mock

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/syndtr/goleveldb/leveldb"
)

func MockDA(mdb *leveldb.DB, daData []byte, batchNumber int) (string, error) {

	hash := sha256.Sum256(daData)
	hashString := hex.EncodeToString(hash[:])

	mockData := types.MockDAStruck{
		DataBlob:    daData,
		BatchNumber: batchNumber,
		Commitment:  hashString,
	}

	byteMockData := []byte(fmt.Sprintf("%v", mockData))

	dbName := fmt.Sprintf("mockda-%d", batchNumber)
	dbErr := mdb.Put([]byte(dbName), byteMockData, nil)
	if dbErr != nil {
		return "", fmt.Errorf("error putting data into mock db: %v", dbErr)
	}

	_ = fmt.Sprintf("da_id : %d, commitment : %s", dbName, hashString)

	return dbName, nil
}
