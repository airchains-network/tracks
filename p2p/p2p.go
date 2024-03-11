//package p2p
//
//import (
//	"context"
//	"fmt"
//	"github.com/libp2p/go-libp2p"
//	"github.com/libp2p/go-libp2p/core/crypto"
//	"github.com/libp2p/go-libp2p/core/host"
//	"github.com/libp2p/go-libp2p/core/network"
//	"github.com/libp2p/go-libp2p/core/peer"
//	"github.com/libp2p/go-libp2p/core/protocol"
//	multiaddr "github.com/multiformats/go-multiaddr"
//	"io"
//	"os"
//	"os/signal"
//	"sync"
//	"syscall"
//)
//
//var connectedPeers = make(map[peer.ID]peer.AddrInfo)
//var mutex = &sync.Mutex{} // For synchronizing access to connectedPeers
//var Node host.Host
//
//func startNode(ctx context.Context) (host.Host, error) {
//	filepath := "sequencer/identity.info"
//	// Load the file into a byte slice
//	serializedPrivKey, err := os.ReadFile(filepath)
//	if err != nil {
//		fmt.Println("Error reading file:", err)
//	}
//
//	// Unmarshal the private key bytes into a libp2p PrivKey object
//	privKey, err := crypto.UnmarshalPrivateKey(serializedPrivKey)
//	if err != nil {
//		panic(fmt.Errorf("failed to unmarshal private key: %w", err))
//	}
//	//fmt.Printf("Node's Private Key: %x\n", privKey)
//
//	node, err := libp2p.New(
//		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/2300"), // Listen on all interfaces and a random port
//		libp2p.Identity(privKey),                          // Use the private key to identify this node
//		libp2p.Ping(false),                                // Disable the built-in ping protocol
//	)
//	if err != nil {
//		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
//	}
//
//	// Register connection handler to update connectedPeers
//	node.Network().Notify(&network.NotifyBundle{
//		ConnectedF: func(n network.Network, c network.Conn) {
//			mutex.Lock()
//			peerInfo := peer.AddrInfo{
//				ID:    c.RemotePeer(),
//				Addrs: []multiaddr.Multiaddr{c.RemoteMultiaddr()},
//			}
//			connectedPeers[c.RemotePeer()] = peerInfo
//			mutex.Unlock()
//			fmt.Printf("Connected to %s\n", c.RemotePeer())
//		},
//		DisconnectedF: func(n network.Network, c network.Conn) {
//			mutex.Lock()
//			delete(connectedPeers, c.RemotePeer())
//			mutex.Unlock()
//			fmt.Printf("Disconnected from %s\n", c.RemotePeer())
//		},
//	})
//
//	return node, nil
//}
//
//func printNodeInfo(node host.Host) {
//	fmt.Println("Listen addresses:", node.Addrs())
//	peerInfo := peer.AddrInfo{
//		ID:    node.ID(),
//		Addrs: node.Addrs(),
//	}
//	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
//	if err != nil {
//		fmt.Println("Failed to obtain p2p addresses:", err)
//		return
//	}
//	for _, addr := range addrs {
//		fmt.Println("libp2p node address:", addr)
//	}
//
//	privKey := node.Peerstore().PrivKey(node.ID())
//	if privKey == nil {
//	}
//	fmt.Printf("Node's Private Key: %x\n", privKey)
//	// Convert the private key to a bytes representation (for demonstration purposes)
//	privBytes, err := crypto.MarshalPrivateKey(privKey)
//	if err != nil {
//	}
//
//	// Print the private key bytes
//	fmt.Printf("Node's Private Key: %x\n", privBytes)
//
//}
//
//func connectToPeer(ctx context.Context, node host.Host, addrStr string) error {
//	addr, err := multiaddr.NewMultiaddr(addrStr)
//	fmt.Printf("Connecting to %s\n", addr)
//	if err != nil {
//		return fmt.Errorf("parsing multiaddr failed: %w", err)
//	}
//	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
//	if err != nil {
//		return fmt.Errorf("creating peer info failed: %w", err)
//	}
//	if err := node.Connect(ctx, *peerInfo); err != nil {
//		return fmt.Errorf("connecting to peer failed: %w", err)
//	}
//	return nil
//}
//
//const customProtocolID = "/station/tracks/0.0.1"
//
//func setupStreamHandler(node host.Host, ctx context.Context) {
//	node.SetStreamHandler(protocol.ID(customProtocolID), func(s network.Stream) {
//		defer s.Close()
//		const initialBufSize = 8192 // Initial buffer size
//		var buf []byte = make([]byte, initialBufSize)
//		var ReceivedDataType string
//		var ReceivedDataBytes []byte
//		for {
//			n, err := s.Read(buf)
//
//			if err == io.EOF {
//				fmt.Println("Stream closed by sender")
//				break
//			}
//			if err != nil {
//				fmt.Println("Failed to read from stream:", err)
//				return // Exit the handler on read error.
//			}
//
//			fmt.Printf("Received bytes: %v\n", buf[:n])
//
//			// If the buffer was filled, increase its size for the next read
//			if n == len(buf) {
//				newSize := len(buf) * 2 // Double the buffer size
//				buf = make([]byte, newSize)
//			}
//			dataType, dataByte, err := DecodeGossipData(buf[:n])
//			if err != nil {
//				fmt.Println("Error in getting data type:", err)
//				return
//			}
//			fmt.Println("Data Type:", dataType)
//			ReceivedDataType = dataType
//			ReceivedDataBytes = dataByte
//		}
//
//		ProcessGossipMessage(node, ctx, ReceivedDataType, ReceivedDataBytes)
//
//	})
//}
//
//func sendMessage(ctx context.Context, node host.Host, peerID peer.ID, message []byte) error {
//	s, err := node.NewStream(ctx, peerID, protocol.ID(customProtocolID))
//	fmt.Printf("Sending message to %s\n", peerID)
//	if err != nil {
//		return fmt.Errorf("failed to open stream: %w", err)
//	}
//	defer s.Close()
//
//	// before sending message, check the data type of the message
//	dataType, dataByte, err := DecodeGossipData(message)
//	if err != nil {
//		fmt.Println("Error in getting data type:", err)
//		return err
//	}
//	fmt.Println("Data type:", dataType)
//	_ = dataByte
//
//	_, err = s.Write(message)
//	if err != nil {
//		return fmt.Errorf("failed to write message to stream: %w", err)
//	}
//
//	return nil
//}
//
//// Function to broadcast a message to all connected peers
//func BroadcastMessage(ctx context.Context, host host.Host, message []byte) {
//	mutex.Lock()
//	fmt.Println("Broadcasting message to all connected peers")
//	defer mutex.Unlock()
//	for peerID, _ := range connectedPeers {
//		if len(connectedPeers) == 1 {
//			fmt.Println("Only 1 peer to broadcast message ")
//		}
//		if peerID == host.ID() {
//			fmt.Println("Skipping message to self")
//			continue // Skip sending message to self
//		}
//		if err := sendMessage(ctx, host, peerID, message); err != nil {
//			fmt.Printf("Error broadcasting message to %s: %s\n", peerID, err)
//		}
//	}
//}
//
//func P2PConfiguration() {
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	node, err := startNode(ctx)
//	Node = node
//	if err != nil {
//		panic(err)
//		//return false
//	}
//	defer Node.Close()
//
//	printNodeInfo(Node)
//
//	setupStreamHandler(Node, ctx)
//
//	if len(os.Args) > 2 {
//		// Connect to the specified peer and get its ID for pinging
//		peerAddrStr := os.Args[2]
//		err := connectToPeer(ctx, node, peerAddrStr)
//		if err != nil {
//			fmt.Println("Error connecting to peer:", err)
//			//return false
//		}
//
//		//
//		//// Attach the ping service and handler
//		//pingService := ping.NewPingService(node)
//
//		// Start pinging the peer
//
//	} else {
//		// Start the leader election process
//		fmt.Println()
//	}
//	//Wait for a SIGINT (Ctrl+C) or SIGTERM signal to shut down gracefully.
//	ch := make(chan os.Signal, 1)
//	signal.Notify(ch, syscall.SIGTERM)
//	<-ch
//
//	fmt.Println("Received signal, shutting down...")
//}

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
		fmt.Printf("Received bytes: %v\n", buf[:n])
		buf = resizeBufferIfNeeded(buf, n)
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
