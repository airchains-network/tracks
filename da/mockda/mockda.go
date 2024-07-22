package mock

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/airchains-network/tracks/types"
	"github.com/syndtr/goleveldb/leveldb"
)

// MockDA is a function that mocks the functionality of storing data in a mock database (leveldb). It takes the following parameters:
// - mdb: a pointer to a leveldb.DB instance representing the mock database
// - daData: a byte slice containing the data to be stored
// - batchNumber: an integer representing the batch number
//
// The function performs the following steps:
// 1. Computes the SHA256 hash of daData.
// 2. Encodes the hash as a string using hexadecimal encoding.
// 3. Creates a new types.MockDAStruck instance with the daData, batchNumber, and computed hashString.
// 4. Converts the mockData into a byte slice.
// 5. Generates a unique database name based on the batchNumber.
// 6. Stores the byteMockData in the mock database using the dbName as the key.
// 7. Returns the dbName and nil error if the operation is successful.
// 8. Otherwise, returns an empty string and an error message indicating the failure.
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
