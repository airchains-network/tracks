package p2p

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"os"
)

func InititateIdentity(daType string, moniker string, stationType string) {

	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/2300"), // Listen on all interfaces and a random port
		libp2p.Ping(false), // Disable the built-in ping protocol
	)
	if err != nil {
		panic(err)
	}

	publicID := node.ID()
	fmt.Println("Node's public ID: ", publicID)
	privBytes, err := crypto.MarshalPrivateKey(node.Peerstore().PrivKey(node.ID()))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Node's Secret Key: %x\n", privBytes)
	defaultConfigSequencer := types.Sequencer{
		Moniker:     moniker,
		StationType: stationType,
		DAType:      daType,
		Identity:    publicID,
	}
	dataDIR := "sequencer"
	err = os.MkdirAll(dataDIR, 0755) // 0755 commonly used permissions for directories
	if err != nil {
		panic(err) // Handle the error properly
	}

	f, err := os.Create("sequencer/sequencer.toml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Encode the data to TOML and save it
	if err := toml.NewEncoder(f).Encode(defaultConfigSequencer); err != nil {
		panic(err)
	}

	// Write data to file
	err = os.WriteFile("sequencer/identity.info", privBytes, 0644) // 0644 is a common permission for read+write by owner and read-only by others
	if err != nil {
		panic(err) // Handle the error properly
	}
}
