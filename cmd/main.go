package main

import (
	"bufio"
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	multiaddr "github.com/multiformats/go-multiaddr"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func setupNode(ctx context.Context) (host.Host, error) {
	// Creating a new libp2p host
	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"), // Listen on all interfaces and a random port
		libp2p.Ping(false), // Disable the built-in ping protocol
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}
	return node, nil
}

func printNodeInfo(node host.Host) {
	// Print the node's listening addresses
	fmt.Println("Listen addresses:", node.Addrs())

	// Print the node's PeerInfo in multiaddr format
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

		// Read message from the stream
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

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := setupNode(ctx)
	if err != nil {
		panic(err)
	}
	defer node.Close() // Ensure the node is closed on exit

	printNodeInfo(node)
	setupStreamHandler(node)

	if len(os.Args) > 1 {
		// Connect to the specified peer and get its ID for pinging
		peerAddrStr := os.Args[1]
		err := connectToPeer(ctx, node, peerAddrStr)
		if err != nil {
			fmt.Println("Error connecting to peer:", err)
			return
		}

		// Extract the peer ID from the multiaddress for pinging
		addr, err := multiaddr.NewMultiaddr(peerAddrStr)
		if err != nil {
			fmt.Println("Failed to parse multiaddress:", err)
			return
		}
		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			fmt.Println("Failed to extract peer info from address:", err)
			return
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
					sendMessage(ctx, node, peerID, "Hello from "+node.ID().String())
					fmt.Println("Ping successful to", peerID)
				}
				time.Sleep(3 * time.Second)
			}
		}()
	}

	// Wait for a SIGINT (Ctrl+C) or SIGTERM signal to shut down gracefully.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	fmt.Println("Received signal, shutting down...")
}
