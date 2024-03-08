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
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	multiaddr "github.com/multiformats/go-multiaddr"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var connectedPeers = make(map[peer.ID]peer.AddrInfo)
var mutex = &sync.Mutex{} // For synchronizing access to connectedPeers
var Node host.Host

func startNode(ctx context.Context) (host.Host, error) {
	filepath := "sequencer/identity.info"
	// Load the file into a byte slice
	serializedPrivKey, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Println("Error reading file:", err)
	}

	// Unmarshal the private key bytes into a libp2p PrivKey object
	privKey, err := crypto.UnmarshalPrivateKey(serializedPrivKey)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal private key: %w", err))
	}
	//fmt.Printf("Node's Private Key: %x\n", privKey)

	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/2300"), // Listen on all interfaces and a random port
		libp2p.Identity(privKey),                          // Use the private key to identify this node
		libp2p.Ping(false),                                // Disable the built-in ping protocol
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	// Register connection handler to update connectedPeers
	node.Network().Notify(&network.NotifyBundle{
		ConnectedF: func(n network.Network, c network.Conn) {
			mutex.Lock()
			peerInfo := peer.AddrInfo{
				ID:    c.RemotePeer(),
				Addrs: []multiaddr.Multiaddr{c.RemoteMultiaddr()},
			}
			connectedPeers[c.RemotePeer()] = peerInfo
			mutex.Unlock()
			fmt.Printf("Connected to %s\n", c.RemotePeer())
		},
		DisconnectedF: func(n network.Network, c network.Conn) {
			mutex.Lock()
			delete(connectedPeers, c.RemotePeer())
			mutex.Unlock()
			fmt.Printf("Disconnected from %s\n", c.RemotePeer())
		},
	})

	return node, nil
}

func printNodeInfo(node host.Host) {
	fmt.Println("Listen addresses:", node.Addrs())
	peerInfo := peer.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}
	addrs, err := peer.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil {
		fmt.Println("Failed to obtain p2p addresses:", err)
		return
	}
	for _, addr := range addrs {
		fmt.Println("libp2p node address:", addr)
	}

	privKey := node.Peerstore().PrivKey(node.ID())
	if privKey == nil {
	}
	fmt.Printf("Node's Private Key: %x\n", privKey)
	// Convert the private key to a bytes representation (for demonstration purposes)
	privBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
	}

	// Print the private key bytes
	fmt.Printf("Node's Private Key: %x\n", privBytes)

}

func connectToPeer(ctx context.Context, node host.Host, addrStr string) error {
	addr, err := multiaddr.NewMultiaddr(addrStr)
	if err != nil {
		return fmt.Errorf("parsing multiaddr failed: %w", err)
	}
	peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
	if err != nil {
		return fmt.Errorf("creating peer info failed: %w", err)
	}
	if err := node.Connect(ctx, *peerInfo); err != nil {
		return fmt.Errorf("connecting to peer failed: %w", err)
	}
	return nil
}

const customProtocolID = "/station/tracks/0.0.1"

func setupStreamHandler(node host.Host) {
	node.SetStreamHandler(protocol.ID(customProtocolID), func(s network.Stream) {
		defer s.Close()
		//buf := bufio.NewReader(s)
		//str, err := buf.ReadString('\n')
		//if err != nil {
		//	fmt.Println("Failed to read from stream:", err)
		//	os.Exit(1)
		//}

		//var receivedNumber int
		//_, err = fmt.Sscanf(str, "Random number: %d", &receivedNumber)
		//if err != nil {
		//	fmt.Println("Failed to parse received number:", err)
		//	os.Exit(1)
		//}
		//
		//fmt.Printf("Received random number: %d\n", receivedNumber)

		buf := make([]byte, 1024) // Adjust the buffer size as needed.

		for {
			n, err := s.Read(buf)
			if err == io.EOF {
				fmt.Println("Stream closed by sender")
				break
			}
			if err != nil {
				fmt.Println("Failed to read from stream:", err)
				return // Exit the handler on read error.
			}

			// Here, buf[:n] contains the received bytes.
			// Process the received bytes according to your application's needs.
			fmt.Printf("Received bytes: %v\n", buf[:n])
		}
	})
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

// Function to broadcast a message to all connected peers
func BroadcastMessage(ctx context.Context, host host.Host, message []byte) {
	mutex.Lock()
	fmt.Println("Broadcasting message to all connected peers")
	defer mutex.Unlock()
	for peerID, _ := range connectedPeers {
		if len(connectedPeers) == 1 {
			fmt.Println("Only 1 peer to broadcast message ")
		}
		if peerID == host.ID() {
			fmt.Println("Skipping message to self")
			continue // Skip sending message to self
		}
		if err := sendMessage(ctx, host, peerID, message); err != nil {
			fmt.Printf("Error broadcasting message to %s: %s\n", peerID, err)
		}
	}
}

func P2PConfiguration() bool {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	node, err := startNode(ctx)
	Node = node
	if err != nil {
		panic(err)
		return false
	}
	defer Node.Close()
	printNodeInfo(Node)
	setupStreamHandler(Node)

	if len(os.Args) > 2 {
		// Connect to the specified peer and get its ID for pinging
		peerAddrStr := os.Args[1]
		err := connectToPeer(ctx, node, peerAddrStr)
		if err != nil {
			fmt.Println("Error connecting to peer:", err)
			return false
		}

		// Extract the peer ID from the multiaddress for pinging
		addr, err := multiaddr.NewMultiaddr(peerAddrStr)
		if err != nil {
			fmt.Println("Failed to parse multiaddress:", err)
			return false
		}
		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			fmt.Println("Failed to extract peer info from address:", err)
			return false
		}
		peerID := peerInfo.ID

		// Attach the ping service and handler
		pingService := ping.NewPingService(node)

		// Start pinging the peer
		go func() {
			fmt.Println("Sending ping messages to", peerID)
			for {
				pingMessage := pingService.Ping(ctx, peerID)
				if pingMessage == nil {
					fmt.Println("Ping failed:", err)
				} else {
					sendMessage(ctx, node, peerID, []byte("Hello from the other side"))
					fmt.Println("Ping successful to", peerID)
				}
				time.Sleep(3 * time.Second)
			}
		}()
	} else {
		// Start the leader election process
		fmt.Println()
	}
	// Wait for a SIGINT (Ctrl+C) or SIGTERM signal to shut down gracefully.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	fmt.Println("Received signal, shutting down...")
	return true
}

var lastSelectedIndex int = -1

func selectLeader(h host.Host, peers []peer.ID) peer.ID {
	mutex.Lock()
	defer mutex.Unlock()

	rand.Seed(time.Now().UnixNano())

	if len(peers) == 0 {
		return h.ID() // Fallback to self if no other peers are available
	}

	// Ensure we cycle through peers and check availability
	for i := 0; i < len(peers); i++ {
		lastSelectedIndex = (lastSelectedIndex + 1) % len(peers)
		potentialLeader := peers[lastSelectedIndex]

		//if isPeerAvailable(h, potentialLeader) {
		return potentialLeader
		//}
	}

	return h.ID() // Fallback to self if no peers are available/active
}
