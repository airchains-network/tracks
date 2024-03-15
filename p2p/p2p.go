package p2p

import (
	"context"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	multiaddr "github.com/multiformats/go-multiaddr"
	"io"
	"math/big"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
)

const (
	identityFilePath         = "sequencer/identity.info"
	customProtocolID         = "/station/tracks/0.0.1"
	customDataSyncProtocolID = "/tracks/dataSync/0.0.1"
	MaxChunkSize             = 100
)

var (
	incomingPeers  = NewPeerList()
	peerListLocked = false
	peerListLock   = sync.Mutex{}
)

type PeerList struct {
	peers []peer.AddrInfo
}

type NodeStateSync struct {
	TrackAppHash string
	PODNumber    int
}

type PodData struct {
	PODs []*PodData
}

func (p *PeerList) AddPeer(peerInfo peer.AddrInfo) {
	p.peers = append(p.peers, peerInfo)
	sort.Slice(p.peers, func(i, j int) bool {
		return p.peers[i].ID.String() < p.peers[j].ID.String()
	})
}

func (p *PeerList) GetPeers() []peer.AddrInfo {
	return p.peers
}

func NewPeerList() *PeerList {
	return &PeerList{}
}

var (
	// ConnectedPeers = make(map[peer.ID]peer.AddrInfo) // Not used anymore
	mutex    = &sync.Mutex{}
	Node     host.Host
	CTX      context.Context
	peerList = NewPeerList()
)

// Your other functions...
func onConnected(n network.
	Network, c network.Conn) {
	peerListLock.Lock()
	defer peerListLock.Unlock()

	peerInfo := peer.AddrInfo{ID: c.RemotePeer(), Addrs: []multiaddr.Multiaddr{c.RemoteMultiaddr()}}

	if peerListLocked {
		incomingPeers.AddPeer(peerInfo)
	} else {
		peerList.AddPeer(peerInfo)
	}

	fmt.Printf("Connected to %s\n", c.RemotePeer())

	// Send a data sync request to the newly connected peer
	sendDataSyncRequest(Node, CTX, c.RemotePeer())
}
func onDisconnected(n network.Network, c network.Conn) {
	mutex.Lock()
	defer mutex.Unlock()
	// delete(ConnectedPeers, c.RemotePeer()) // Not used anymore
	fmt.Printf("Disconnected from %s\n", c.RemotePeer())
}

//// Get all peers, including the current node
//func getAllPeers(node host.Host) []peer.AddrInfo {
//	peers := peerList.GetPeers()
//	ownPeerInfo := peer.AddrInfo{ID: node.ID(), Addrs: node.Addrs()}
//	peers = append(peers, ownPeerInfo)
//	return peers
//}

func getAllPeers(node host.Host) []peer.AddrInfo {
	peers := peerList.GetPeers()
	ownPeerInfo := peer.AddrInfo{ID: node.ID(), Addrs: node.Addrs()}
	peers = append(peers, ownPeerInfo)

	// Sort peers by ID
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].ID.String() < peers[j].ID.String()
	})

	return peers
}

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

func printNodeInfo(node host.Host) {
	fmt.Println("Listen addresses:", node.Addrs())
	fmt.Println("Node ID:", node.ID())
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
	node.SetStreamHandler(protocol.ID(customDataSyncProtocolID), dataSyncHandler) // Data Sync handler
}

func streamHandler(s network.Stream) {
	defer s.Close()
	if s.Protocol() == protocol.ID(customDataSyncProtocolID) {
		// Handle block data message
		blockData := "sdsd"
		gobDecoder := gob.NewDecoder(s)
		err := gobDecoder.Decode(&blockData)
		if err != nil {
			logs.Log.Error("Failed to decode data")
		}
		return
	}
	handleStreamData(s)
}

func handleStreamData(s network.Stream) {
	const initialBufSize = 8192
	buf := make([]byte, initialBufSize)
	messageBroadcaster := s.Conn().RemotePeer()
	//fmt.Println(messageBroadcaster)
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
		fmt.Println("Data Type Received from other Peer :", dataType)

		ProcessGossipMessage(Node, CTX, dataType, dataByte, messageBroadcaster)
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

//func BroadcastMessage(ctx context.Context, host host.Host, message []byte) {
//	mutex.Lock()
//	defer mutex.Unlock()
//
//	for peerID := range peerList.GetPeers() {
//		if peerID == host.ID() {
//			continue
//		}
//		if err := sendMessage(ctx, host, peerID, message); err != nil {
//			fmt.Printf("Error broadcasting message to %s: %s\n", peerID, err)
//		}
//	}
//}

func BroadcastMessage(ctx context.Context, host host.Host, message []byte) {
	mutex.Lock()
	defer mutex.Unlock()

	peers := peerList.GetPeers()
	if len(peers) == 0 {
		fmt.Println("No connected peers to send the message to.")
		return
	}

	for _, peerInfo := range peers {
		if peerInfo.ID == host.ID() {
			continue
		}
		if err := sendMessage(ctx, host, peerInfo.ID, message); err != nil {
			fmt.Printf("Error broadcasting message to %s: %s\n", peerInfo.ID, err)
		}
	}
}

func P2PConfiguration() {

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

func MasterTracksSelection(host host.Host, sharedInput string) string {
	peers := getAllPeers(host)
	numPeers := len(peers)
	if numPeers == 0 {
		fmt.Println("No peers available.")
		return ""
	}
	fmt.Printf(sharedInput)

	// Compute the SHA256 hash of the sharedInput
	h := sha256.New()
	h.Write([]byte(sharedInput))
	hashed := h.Sum(nil)

	// Convert the hash to a big.Int
	hashedInt := new(big.Int)
	hashedInt.SetBytes(hashed)

	// Use modulus to get an index within the range of numPeers.
	// randomIndex will always be in the range of 0 to numPeers-1 (inclusive).
	randomIndex := hashedInt.Mod(hashedInt, big.NewInt(int64(numPeers)))
	randomPeer := peers[int(randomIndex.Int64())]

	// Need to re-compute hash and index if the randomly selected peer is the host itself
	h = sha256.New()
	h.Write([]byte(sharedInput))
	hashed = h.Sum(nil)

	hashedInt = new(big.Int)
	hashedInt.SetBytes(hashed)

	randomIndex = hashedInt.Mod(hashedInt, big.NewInt(int64(numPeers)))
	randomPeer = peers[int(randomIndex.Int64())]

	fmt.Printf("Selected peer ID: %s\n", randomPeer.ID.String())

	return randomPeer.ID.String()
}

func sendDataSyncRequest(node host.Host, ctx context.Context, peerID peer.ID) {
	s, err := node.NewStream(ctx, peerID, protocol.ID(customDataSyncProtocolID))
	if err != nil {
		logs.Log.Error("Failed to open stream")
	}
	defer s.Close()

	_, err = s.Write([]byte("SYNC_REQUEST"))
	if err != nil {
		logs.Log.Error("Failed to write to stream")
	}
}

func dataSyncHandler(s network.Stream) {
	defer s.Close()

	buf := make([]byte, 256)
	n, err := s.Read(buf)
	if err != nil {
		logs.Log.Error("Failed to read from stream")
	}

	if string(buf[:n]) == "SYNC_REQUEST" {
		// If it's a sync request, fetch data and send it
		data := "sds" // TODO: Implement this function
		gobEncoder := gob.NewEncoder(s)
		err := gobEncoder.Encode(data)
		if err != nil {
			logs.Log.Error("Failed to encode data")
		}
	}
}
