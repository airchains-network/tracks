package p2p

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	multiaddr "github.com/multiformats/go-multiaddr"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	connectedPeers = make(map[peer.ID]peer.AddrInfo)
	mutex          = &sync.Mutex{}
	Node           host.Host
	CTX            context.Context
)

const (
	identityFilePath = "sequencer/identity.info"
	customProtocolID = "/station/tracks/0.0.1"
)

func startNode(ctx context.Context) (host.Host, error) {
	privKey, err := loadPrivateKey(identityFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/2300"),
		libp2p.Identity(privKey),
		libp2p.Ping(false),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	registerConnectionHandlers(node)
	return node, nil
}

func loadPrivateKey(filepath string) (crypto.PrivKey, error) {
	serializedPrivKey, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return crypto.UnmarshalPrivateKey(serializedPrivKey)
}

func registerConnectionHandlers(node host.Host) {
	node.Network().Notify(&network.NotifyBundle{
		ConnectedF:    onConnected,
		DisconnectedF: onDisconnected,
	})
}

func onConnected(n network.Network, c network.Conn) {
	mutex.Lock()
	defer mutex.Unlock()
	peerInfo := peer.AddrInfo{ID: c.RemotePeer(), Addrs: []multiaddr.Multiaddr{c.RemoteMultiaddr()}}
	connectedPeers[c.RemotePeer()] = peerInfo
	fmt.Printf("Connected to %s\n", c.RemotePeer())
}

func onDisconnected(n network.Network, c network.Conn) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(connectedPeers, c.RemotePeer())
	fmt.Printf("Disconnected from %s\n", c.RemotePeer())
}

func printNodeInfo(node host.Host) {
	fmt.Println("Listen addresses:", node.Addrs())
	fmt.Println("Node ID:", node.ID())
	printNodePrivateKey(node)
}

func printNodePrivateKey(node host.Host) {
	privKey := node.Peerstore().PrivKey(node.ID())
	if privKey != nil {
		privBytes, err := crypto.MarshalPrivateKey(privKey)
		if err == nil {
			fmt.Printf("Node's Private Key: %x\n", privBytes)
		}
	}
}

func connectToPeer(ctx context.Context, node host.Host, addrStr string) error {
	peerInfo, err := parseAddrToPeerInfo(addrStr)
	if err != nil {
		return err
	}
	fmt.Printf("Connecting to %s\n", addrStr)
	return node.Connect(ctx, peerInfo)
}

func parseAddrToPeerInfo(addrStr string) (peer.AddrInfo, error) {
	addr, err := multiaddr.NewMultiaddr(addrStr)
	if err != nil {
		return peer.AddrInfo{}, fmt.Errorf("parsing multiaddr failed: %w", err)
	}
	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return peer.AddrInfo{}, fmt.Errorf("creating peer info failed: %w", err)
	}
	return *peerInfo, nil
}

func setupStreamHandler(node host.Host) {
	node.SetStreamHandler(protocol.ID(customProtocolID), streamHandler)
}

func streamHandler(s network.Stream) {
	defer s.Close()
	handleStreamData(s)
}

func handleStreamData(s network.Stream) {
	const initialBufSize = 8192
	buf := make([]byte, initialBufSize)

	for {
		n, err := s.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Failed to read from stream:", err)
			} else {
				fmt.Println("Stream closed by sender")
			}
			break
		}
		buf = resizeBufferIfNeeded(buf, n)
		dataType, dataByte, err := DecodeGossipData(buf[:n])
		if err != nil {
			fmt.Println("Error in getting data type:", err)
			return
		}
		//fmt.Println("Data Type:", dataType)
		//fmt.Printf("Data: %s\n", dataByte)
		//shared.SetPodState(shared.PodState{
		//	LatestPodHeight:         1000000,
		//	LatestPodMerkleRootHash: nil,
		//})
		fmt.Println("current pod data:")
		ProcessGossipMessage(Node, CTX, dataType, dataByte)
	}
}

func resizeBufferIfNeeded(buf []byte, n int) []byte {
	if n == len(buf) {
		newSize := len(buf) * 2
		return make([]byte, newSize)
	}
	return buf
}
func sendMessage(ctx context.Context, node host.Host, peerID peer.ID, message []byte) error {
	s, err := node.NewStream(ctx, peerID, protocol.ID(customProtocolID))
	fmt.Printf("Sending message to %s\n", peerID)
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}
	defer s.Close()
	_, err = s.Write(message)
	if err != nil {
		return fmt.Errorf("failed to write message to stream: %w", err)
	}

	return nil
}

func BroadcastMessage(ctx context.Context, host host.Host, message []byte) {
	mutex.Lock()
	defer mutex.Unlock()

	for peerID := range connectedPeers {
		if peerID == host.ID() {
			continue
		}
		if err := sendMessage(ctx, host, peerID, message); err != nil {
			fmt.Printf("Error broadcasting message to %s: %s\n", peerID, err)
		}
	}
}

func P2PConfiguration() {

	// create state

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	CTX = ctx
	node, err := startNode(ctx)
	if err != nil {
		panic(fmt.Errorf("error starting node: %w", err))
	}
	Node = node
	defer Node.Close()

	printNodeInfo(Node)
	setupStreamHandler(Node)

	handlePeerConnections(ctx, Node)
	waitForShutdownSignal()
}

func handlePeerConnections(ctx context.Context, node host.Host) {
	if len(os.Args) > 2 {
		peerAddrStr := os.Args[2]
		if err := connectToPeer(ctx, node, peerAddrStr); err != nil {
			fmt.Println("Error connecting to peer:", err)
		}
	}
}

func waitForShutdownSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM)
	<-ch
	fmt.Println("Received signal, shutting down...")
}
