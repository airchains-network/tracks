package config

import (
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
)

// TODO Change this to MUTEX for better performance and Struct Types
const (
	StationRPC           = "http://localhost:8545" //Station RPC
	StationID            = "1"                     //Station Unique ID
	StationName          = "Station1"              // NAME OF THE STATION
	StationType          = "1"                     //EVM, SVM ,COSMWASM
	PODSize              = 25                      // P0D Size
	StationBlockDuration = 5                       // In Seconds
	JunctionRPC          = "1"                     // Junction RPC
	DAType               = "mock"                  // Data Availability Type  : -Eigen , Avail , Celestia,Mock
	DARpc                = "localhost:8080"        // Data Availability RPC
)

type Config struct {
	mu                   sync.RWMutex // Mutex for safe concurrent access
	LatestPodNumber      uint64
	Peers                []string
	LatestProof          []byte
	PreviousProof        []byte
	StaticDb             *leveldb.DB
	StationRPC           string
	StationID            string
	StationName          string
	StationType          string
	PODSize              int
	StationBlockDuration int
	JunctionRPC          string
	DAType               string
	DARpc                string
}

// NewConfig creates a new Config instance with default values
func NewConfig() *Config {
	return &Config{
		LatestPodNumber:      0,   // Example default value
		Peers:                nil, // No default peers
		LatestProof:          nil, // No default latest proof
		PreviousProof:        nil, // No default previous proof
		StaticDb:             nil, // No default DB, must be set up separately
		StationRPC:           "http://localhost:8545",
		StationID:            "1",
		StationName:          "Station1",
		StationType:          "1",
		PODSize:              25,
		StationBlockDuration: 5,
		JunctionRPC:          "1",
		DAType:               "mock",
		DARpc:                "localhost:8080",
	}
}

//type LatestUnverifiedPodData struct {
//	mu    sync.Mutex
//	count uint64
//}
//
//func (pod *LatestUnverifiedPodData) IncrementUnverifiedPod() {
//	pod.mu.Lock()   // Lock the mutex before accessing count
//	pod.count++     // Critical section: modify count
//	pod.mu.Unlock() // Unlock the mutex after accessing count
//}
//
//func (pod *LatestUnverifiedPodData) ValueUnverifiedPod() uint64 {
//	pod.mu.Lock()         // Lock the mutex before accessing count
//	defer pod.mu.Unlock() // Unlock the mutex after accessing count using defer
//	return pod.count      // Critical section: read count
//}
//
//type LatestUnverifiedProofData struct {
//	mu   sync.Mutex
//	data []byte
//}
//
//func (proof *LatestUnverifiedProofData) UpdateUnverifiedProof(p []byte) {
//	proof.mu.Lock()   // Lock the mutex before accessing count
//	proof.data = p    // Critical section: modify count
//	proof.mu.Unlock() // Unlock the mutex after accessing count
//}
//
//func (proof *LatestUnverifiedProofData) ValueUnverifiedProof() []byte {
//	proof.mu.Lock()         // Lock the mutex before accessing count
//	defer proof.mu.Unlock() // Unlock the mutex after accessing count using defer
//	return proof.data       // Critical section: read count
//}
//
//// call from main with default values... e.g. pod := NewLatestUnverifiedPodData(10) // Start count at 10
//func NewLatestUnverifiedPodData(initialCount uint64) *LatestUnverifiedPodData {
//	return &LatestUnverifiedPodData{
//		count: initialCount,
//	}
//}
//func NewLatestVerified() *LatestUnverifiedProofData {
//	return &LatestUnverifiedProofData{
//		data: nil,
//	}
//}

// Combined struct for unverified pod data and proof

type LatestUnverifiedData struct {
	Mtx       sync.Mutex
	Count     uint64
	ProofData []byte
}

func NewLatestUnverifiedData(initialCount uint64, initialProofData []byte) *LatestUnverifiedData {
	return &LatestUnverifiedData{
		Count:     initialCount,
		ProofData: initialProofData,
	}
}

type ProofUpdater interface {
	UpdateUnverifiedProof([]byte)
}

type CountIncrementer interface {
	IncrementUnverifiedPod()
}

func (data *LatestUnverifiedData) IncrementUnverifiedPod() {
	data.Mtx.Lock()
	data.Count++
	data.Mtx.Unlock()
}

func (data *LatestUnverifiedData) UpdateUnverifiedProof(proof []byte) {
	data.Mtx.Lock()
	data.ProofData = proof
	data.Mtx.Unlock()
}
