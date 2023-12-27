package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	multiaddr "github.com/multiformats/go-multiaddr"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/192.168.1.106/tcp/2300"),
		libp2p.Ping(false),
	)
	if err != nil {
		panic(err)
	}
	// 	// attach a ping service to the node
	pingService := &ping.PingService{Host: node}
	node.SetStreamHandler(ping.ID, pingService.PingHandler)

	// print the node's listening addresses
	fmt.Println("Listen addresses:", node.Addrs())
	// print the node's PeerInfo in multiaddr format
	peerInfo := peerstore.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}
	addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	fmt.Println("libp2p node address:", addrs[0])

	if len(os.Args) > 1 {

		go func() {
			for {
				addr, err := multiaddr.NewMultiaddr(os.Args[1])
				if err != nil {
					panic(err)
				}
				peer, err := peerstore.AddrInfoFromP2pAddr(addr)
				if err != nil {
					panic(err)
				}
				if err := node.Connect(context.Background(), *peer); err != nil {
					panic(err)
				}
				fmt.Println("sending  broadcast messages every 3 seconds to", addr)
				ch := pingService.Ping(context.Background(), peer.ID)

				res := <-ch
				fmt.Println("got ping response!", "RTT:", res.RTT)
				stream, err := node.NewStream(context.Background(), peer.ID, ping.ID)
				if err != nil {
					fmt.Println("Error creating stream:", err)

				}
				message := "Hello, this is a broadcasted message!"
				_, err = stream.Write([]byte(message))
				if err != nil {
					fmt.Println("Error writing to stream:", err)
				}
				stream.Close()
				time.Sleep(3 * time.Second)
			}
		}()
		select {}

	} else {
		// wait for a SIGINT or SIGTERM signal
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		fmt.Println("Received signal, shutting down...")
	}

	// shut the node down
	if err := node.Close(); err != nil {
		panic(err)
	}
}
