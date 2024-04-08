package p2p

import (
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"io/ioutil"
	"os"
	"path/filepath"
)

type PeerGenerator struct {
	ListenAddr string
	PingConfig bool
	PrivKey    crypto.PrivKey
	Node       host.Host
}

func NewPeerGenerator(listenAddr string, pingConfig bool) *PeerGenerator {
	return &PeerGenerator{
		ListenAddr: listenAddr,
		PingConfig: pingConfig,
	}
}

func (pg *PeerGenerator) GeneratePeerID() (peer.ID, error) {
	node, err := libp2p.New(
		libp2p.ListenAddrStrings(pg.ListenAddr),
		libp2p.Ping(pg.PingConfig),
	)
	if err != nil {
		return "", err
	}

	pg.PrivKey = node.Peerstore().PrivKey(node.ID())
	pg.Node = node
	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, ".tracks/config/sequencer")
	err = savePrivateKey(filepath.Join(filePath, "identity.info"), pg.PrivKey)
	if err != nil {
		return "", err
	}

	return node.ID(), nil
}

func savePrivateKey(filePath string, privKey crypto.PrivKey) error {
	privateKeyBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("unable to marshal private key: %v", err)
	}

	err = ioutil.WriteFile(filePath, privateKeyBytes, 0644)
	if err != nil {
		return fmt.Errorf("unable to write private key to file: %v", err)
	}

	return nil
}
