package main

import (
	"bufio"
	"context"
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	multiaddr "github.com/multiformats/go-multiaddr"
	"os"
	"os/signal"
	"syscall"
)

var connectedPeers = make(map[peer.ID]peer.AddrInfo)
var mutex = &sync.Mutex{} // For synchronizing access to connectedPeers

func setupNode(ctx context.Context) (host.Host, error) {
	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"), // Listen on all interfaces and a random port
		libp2p.Ping(false), // Disable the built-in ping protocol
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

const customProtocolID = "/myapp/message/1.0.0"

func setupStreamHandler(node host.Host) {
	node.SetStreamHandler(protocol.ID(customProtocolID), func(s network.Stream) {
		defer s.Close()
		buf := bufio.NewReader(s)
		str, err := buf.ReadString('\n')
		if err != nil {
			fmt.Println("Failed to read from stream:", err)
			return
		}
		fmt.Printf("Received message: %s\n", str)
	})
}

func sendMessage(ctx context.Context, node host.Host, peerID peer.ID, message string) error {
	s, err := node.NewStream(ctx, peerID, protocol.ID(customProtocolID))
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}
	defer s.Close()

	_, err = s.Write([]byte(message + "\n"))
	if err != nil {
		return fmt.Errorf("failed to write message to stream: %w", err)
	}

	return nil
}

// Function to broadcast a message to all connected peers
func broadcastMessage(ctx context.Context, host host.Host, message string) {
	mutex.Lock()
	defer mutex.Unlock()
	for peerID, _ := range connectedPeers {
		if peerID == host.ID() {
			continue // Skip sending message to self
		}
		if err := sendMessage(ctx, host, peerID, message); err != nil {
			fmt.Printf("Error broadcasting message to %s: %s\n", peerID, err)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := setupNode(ctx)
	if err != nil {
		panic(err)
	}
	defer node.Close()

	printNodeInfo(node)
	setupStreamHandler(node)

	// Handling user input in a separate goroutine
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Enter message to broadcast (Type 'exit' to quit):")
		for scanner.Scan() {
			input := scanner.Text()
			if input == "exit" {
				break
			}
			broadcastMessage(ctx, node, input)
		}
		cancel() // Cancel the context to exit the main program
	}()

	// Argument to connect to a peer if provided
	if len(os.Args) > 1 {
		peerAddrStr := os.Args[1]
		if err := connectToPeer(ctx, node, peerAddrStr); err != nil {
			fmt.Println("Error connecting to peer:", err)
		}
	}

	// Wait for a SIGINT (Ctrl+C) or SIGTERM signal to shut down gracefully
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ch:
		fmt.Println("Received signal, shutting down...")
	case <-ctx.Done():
		fmt.Println("Context cancelled, shutting down...")
	}
}
