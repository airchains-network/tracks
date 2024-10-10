package config

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

const (
	PODSize                       = 25 // P0D Size
	defaultMoniker                = "tracks"
	DefaultTracksDir              = ".tracks"
	DefaultConfigDir              = "config"
	DefaultGenesisFileName        = "genesis.json"
	DefaultDataDir                = "data"
	DefaultConfigFileName         = "sequencer.toml"
	defaultSubscriptionBufferSize = 200
)

var (
	DefaultConfigFilePath  = filepath.Join(DefaultConfigDir, DefaultConfigFileName)
	DefaultGenesisFilePath = filepath.Join(DefaultTracksDir, DefaultConfigDir, DefaultGenesisFileName)
)

type Config struct {
	BaseConfig *BaseConfig      `toml:"base_config"`
	RPC        *RPCConfig       `toml:"rpc"`
	P2P        *P2PConfig       `toml:"p2p"`
	StateSync  *StateSyncConfig `toml:"state_sync"`
	Consensus  *ConsensusConfig `toml:"consensus"`
	DA         *DAConfig        `toml:"da"`
	Sequencer  *SequencerConfig `toml:"sequencer"`
	Prover     *ProverConfig    `toml:"prover"`
	Station    *StationConfig   `toml:"station"`
	Junction   *JunctionConfig  `toml:"junction"`
}

func DefaultConfig() *Config {
	return &Config{
		BaseConfig: DefaultBaseConfig(),
		RPC:        DefaultRPCConfig(),
		P2P:        DefaultP2PConfig(),
		StateSync:  NewStateSyncConfig(),
		Consensus:  DefaultConsensusConfig(),
		DA:         DefaultDAConfig(),
		Station:    DefaultStationConfig(),
		Junction:   DefaultJunctionConfig(),
		Sequencer:  DefaultSequencerConfig(),
		Prover:     DefaultProverConfig(),
	}
}

// SetRoot sets the RootDir for all Config structs
func (cfg *Config) SetRoot(root string) *Config {
	cfg.BaseConfig.RootDir = root
	cfg.RPC.RootDir = root
	cfg.P2P.RootDir = root
	cfg.Consensus.RootDir = root
	return cfg
}

type BaseConfig struct {
	Version     string `toml:"version"`
	RootDir     string `toml:"root_dir"`
	ProxyApp    string `toml:"proxy_app"`
	Moniker     string `toml:"moniker"`
	DBBackend   string `toml:"db_backend"`
	DBPath      string `toml:"db_path"`
	FilterPeers bool   `toml:"filter_peers"`
}

func DefaultBaseConfig() *BaseConfig {
	return &BaseConfig{
		Version:     "0.0.1",
		Moniker:     defaultMoniker,
		FilterPeers: false,
		DBBackend:   "goleveldb",
		DBPath:      DefaultDataDir,
	}
}

type RPCConfig struct {
	mu                        sync.RWMutex
	RootDir                   string        `toml:"root_dir"`
	ListenAddress             string        `toml:"listen_address"`
	CORSAllowedOrigins        []string      `toml:"cors_allowed_origins"`
	CORSAllowedMethods        []string      `toml:"cors_allowed_methods"`
	CORSAllowedHeaders        []string      `toml:"cors_allowed_headers"`
	GRPCListenAddress         string        `toml:"grpc_listen_address"`
	GRPCMaxOpenConnections    int           `toml:"grpc_max_open_connections"`
	Unsafe                    bool          `toml:"unsafe"`
	MaxOpenConnections        int           `toml:"max_open_connections"`
	MaxSubscriptionClients    int           `toml:"max_subscription_clients"`
	MaxSubscriptionsPerClient int           `toml:"max_subscriptions_per_client"`
	SubscriptionBufferSize    int           `toml:"subscription_buffer_size"`
	WebSocketWriteBufferSize  int           `toml:"web_socket_write_buffer_size"`
	CloseOnSlowClient         bool          `toml:"close_on_slow_client"`
	TimeoutBroadcastTxCommit  time.Duration `toml:"timeout_broadcast_tx_commit"`
	MaxBodyBytes              int64         `toml:"max_body_bytes"`
	MaxHeaderBytes            int           `toml:"max_header_bytes"`
	TLSCertFile               string        `toml:"tls_cert_file"`
	TLSKeyFile                string        `toml:"tls_key_file"`
	PprofListenAddress        string        `toml:"pprof_listen_address"`
}

// DefaultRPCConfig returns a default configuration for the RPC server
func DefaultRPCConfig() *RPCConfig {
	return &RPCConfig{
		ListenAddress:          "tcp://127.0.0.1:2322",
		CORSAllowedOrigins:     []string{},
		CORSAllowedMethods:     []string{http.MethodHead, http.MethodGet, http.MethodPost},
		CORSAllowedHeaders:     []string{"Origin", "Accept", "Content-Type", "X-Requested-With", "X-Server-Time"},
		GRPCListenAddress:      "",
		GRPCMaxOpenConnections: 900,

		Unsafe:             false,
		MaxOpenConnections: 900,

		MaxSubscriptionClients:    100,
		MaxSubscriptionsPerClient: 5,
		SubscriptionBufferSize:    defaultSubscriptionBufferSize,
		TimeoutBroadcastTxCommit:  10 * time.Second,
		WebSocketWriteBufferSize:  defaultSubscriptionBufferSize,

		MaxBodyBytes:   int64(1000000), // 1MB
		MaxHeaderBytes: 1 << 20,        // same as the net/http default

		TLSCertFile: "",
		TLSKeyFile:  "",
	}
}

// P2P COnfiguration
type P2PConfig struct {
	RootDir                 string   `toml:"root_dir,omitempty"`
	NodeId                  peer.ID  `toml:"node_id"`
	ListenAddress           string   `toml:"listen_address"`
	ExternalAddress         string   `toml:"external_address"`
	Seeds                   string   `toml:"seeds"`
	PersistentPeers         []string `toml:"persistent_peers"`
	CurrentlyConnectedPeers []string `toml:"currently_connected_peers"`
}

func DefaultP2PConfig() *P2PConfig {
	return &P2PConfig{
		RootDir:         DefaultDataDir,
		ListenAddress:   "tcp://0.0.0.0:2300",
		ExternalAddress: "",
		NodeId:          "",
		PersistentPeers: []string{},
	}
}

// StateSyncConfig holds configuration settings related to syncing pods.
type StateSyncConfig struct {
	Enable            bool          // Enable or disable pod syncing
	TempDir           string        // Directory for temporary storage during pod syncing
	RPCServers        []string      // List of RPC servers for fetching pods
	PodTrustPeriod    time.Duration // Period for which a pod is considered trusted
	PodTrustHeight    int64         // Height at which the pod's trust starts
	PodTrustHash      string        // Hash of a trusted pod to start syncing from
	PodDiscoveryTime  time.Duration // Time for discovering new pods
	PodRequestTimeout time.Duration // Timeout for pod requests
	PodChunkFetchers  int32         // Number of concurrent fetchers for pod chunks
}

// NewStateSyncConfig creates a new instance of StateSyncConfig with default values.
func NewStateSyncConfig() *StateSyncConfig {
	return &StateSyncConfig{
		Enable:            false,
		TempDir:           "./podsync_temp",
		RPCServers:        []string{},
		PodTrustPeriod:    168 * time.Hour, // 7 days
		PodTrustHeight:    0,
		PodTrustHash:      "",
		PodDiscoveryTime:  15 * time.Minute,
		PodRequestTimeout: 5 * time.Second,
		PodChunkFetchers:  4,
	}
}

// ConsensusConfig holds the configuration options for the consensus layer.
type ConsensusConfig struct {
	RootDir string `mapstructure:"home"`

	// Timing configurations for the pod consensus process
	TimeoutPropose             time.Duration
	TimeoutProposeDelta        time.Duration
	TimeoutPrevote             time.Duration
	TimeoutPrevoteDelta        time.Duration
	TimeoutPrecommit           time.Duration
	TimeoutPrecommitDelta      time.Duration
	TimeoutCommit              time.Duration
	SkipTimeoutCommit          bool
	ValidatePods               bool
	PodValidationSleepDuration time.Duration
	DoubleSignCheckHeight      int64
}

// DefaultConsensusConfig returns a default configuration for the consensus service.
func DefaultConsensusConfig() *ConsensusConfig {

	return &ConsensusConfig{
		TimeoutPropose:             3 * time.Second,
		TimeoutProposeDelta:        500 * time.Millisecond,
		TimeoutPrevote:             1 * time.Second,
		TimeoutPrevoteDelta:        500 * time.Millisecond,
		TimeoutPrecommit:           1 * time.Second,
		TimeoutPrecommitDelta:      500 * time.Millisecond,
		TimeoutCommit:              1 * time.Second,
		SkipTimeoutCommit:          false,
		ValidatePods:               true,
		PodValidationSleepDuration: 100 * time.Millisecond,
		DoubleSignCheckHeight:      0,
	}
}

type DAConfig struct {
	DaName    string
	DaType    string
	DaRPC     string
	DaKey     string
	DaVersion string
}

func DefaultDAConfig() *DAConfig {
	return &DAConfig{
		DaName:    "",
		DaType:    "",
		DaRPC:     "",
		DaKey:     "",
		DaVersion: "",
	}
}

type SequencerConfig struct {
	SequencerType      string
	SequencerRPC       string
	SequencerKey       string
	SequencerVersion   string
	SequencerNamespace string
}

func DefaultSequencerConfig() *SequencerConfig {
	return &SequencerConfig{
		SequencerType:      "",
		SequencerRPC:       "",
		SequencerKey:       "",
		SequencerVersion:   "",
		SequencerNamespace: "",
	}
}

type ProverConfig struct {
	ProverType    string
	ProverRPC     string
	ProverVersion string
	ProverKey     string
}

func DefaultProverConfig() *ProverConfig {
	return &ProverConfig{
		ProverType:    "",
		ProverRPC:     "",
		ProverVersion: "",
		ProverKey:     "",
	}
}

type StationConfig struct {
	StationType      string
	StationRPC       string
	StationAPI       string
	StationSchemaKey string
}

// DefaultStationConfig returns a default configuration for the station.
func DefaultStationConfig() *StationConfig {
	return &StationConfig{
		StationType:      "",
		StationRPC:       "",
		StationAPI:       "",
		StationSchemaKey: "",
	}
}

type JunctionConfig struct {
	JunctionRPC   string
	JunctionAPI   string
	StationId     string
	VRFPrivateKey string
	VRFPublicKey  string
	AddressPrefix string
	AccountName   string
	AccountPath   string
	Tracks        []string
}

// DefaultJunctionConfig returns a default configuration for the junction.
func DefaultJunctionConfig() *JunctionConfig {

	jsonRpc := ""
	JunctionAPI := ""
	stationId := ""
	VRFPrivateKey := ""
	VRFPublicKey := ""
	AddressPrefix := "air"
	var Tracks []string

	return &JunctionConfig{
		JunctionRPC:   jsonRpc,
		JunctionAPI:   JunctionAPI,
		StationId:     stationId,
		VRFPrivateKey: VRFPrivateKey,
		VRFPublicKey:  VRFPublicKey,
		AddressPrefix: AddressPrefix,
		AccountName:   "",
		AccountPath:   "",
		Tracks:        Tracks,
	}
}
