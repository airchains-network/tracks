package p2p

import (
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
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

	privateKey, err := crypto.MarshalPrivateKey(node.Peerstore().PrivKey(node.ID()))
	if err != nil {
		panic(err)
	}
	pg.Node = node
	homeDir, _ := os.UserHomeDir()
	filePath := filepath.Join(homeDir, ".tracks/config")
	err = savePrivateKey(filepath.Join(filePath, "identity.info"), privateKey)
	if err != nil {
		return "", err
	}

	return node.ID(), nil
}

func savePrivateKey(filePath string, privateKey []byte) error {

	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("unable to create directory: %v", err)
	}
	fmt.Println(privateKey)
	err := os.WriteFile(filePath, privateKey, 0644)
	if err != nil {
		return fmt.Errorf("unable to write private key to file: %v", err)
	}
	serializedPrivKey, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
	}
	a, err := crypto.UnmarshalPrivateKey(serializedPrivKey)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(a)
	return nil
}
