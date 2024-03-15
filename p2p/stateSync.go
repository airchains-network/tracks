package p2p

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
)

// Fetches data from a LevelDB database and returns it in chunks
func fetchBlockDataInChunks(dbPath string) ([][]*PodData, error) {
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open LevelDB: %w", err)
	}
	defer db.Close()

	var pods []*PodData
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		// Decode the data into POD
		data := iter.Value()
		var pod *PodData
		decoder := gob.NewDecoder(bytes.NewReader(data))
		err = decoder.Decode(&pod)
		if err != nil {
			log.Printf("Error decoding data for key %s: %v", string(iter.Key()), err)
			continue
		}
		pods = append(pods, pod)
	}

	iter.Release()
	if err = iter.Error(); err != nil {
		return nil, fmt.Errorf("failed to iterate over LevelDB: %w", err)
	}

	var podsChunks [][]*PodData
	chunkSize := MaxChunkSize
	for i := 0; i < len(pods); i += chunkSize {
		end := i + chunkSize
		if end > len(pods) {
			end = len(pods)
		}
		podsChunks = append(podsChunks, pods[i:end])
	}

	return podsChunks, nil
}

// Send chunks one by one
func syncDataChunks(ctx context.Context, node host.Host, targetPeerID peer.ID, podsChunks [][]*PodData) {
	for i, chunk := range podsChunks {
		fmt.Printf("Sending chunk %d\n", i)

		s, err := node.NewStream(ctx, targetPeerID, protocol.ID(customProtocolID))
		if err != nil {
			log.Fatalf("Failed to open stream: %v", err)
		}
		defer s.Close()

		// Encode and send the chunk
		gobEncoder := gob.NewEncoder(s)
		err = gobEncoder.Encode(&chunk)
		if err != nil {
			log.Printf("Failed to send chunk %d: %v", i, err)
		}

		// TODO wait for a confirmation from the receiver
	}
}

func handleDataChunk(s network.Stream) {
	var podDataChunk []*PodData
	gobDecoder := gob.NewDecoder(s)
	err := gobDecoder.Decode(&podDataChunk)
	if err != nil {
		log.Fatalf("Failed to decode Pod data: %v", err)
	}

	storePodsData(podDataChunk)

	// Todo, you could send back a confirmation here
}

// Stores Pod to the LevelDB.
func storePodsData(pods []*PodData) error {
	db, err := leveldb.OpenFile("your_database_path", nil)
	if err != nil {
		return fmt.Errorf("failed to open LevelDB: %w", err)
	}
	defer db.Close()

	for _, pod := range pods {
		// Gob encode the POD
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(pod)
		if err != nil {
			log.Printf("Error encoding pod: %v", err)
			continue
		}

	}

	return nil
}
