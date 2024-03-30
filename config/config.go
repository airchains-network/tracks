package config

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

const (
	PODSize = 25 // P0D Size
)

const (
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
	BaseConfig `mapstructure:",squash"`
	RPC        *RPCConfig
	P2P        *P2PConfig
	StateSync  *StateSyncConfig
	Consensus  *ConsensusConfig
	DA         *DAConfig
	Station    *StationConfig
	Junction   *JunctionConfig
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
	Version     string `mapstructure:"version"`
	RootDir     string `mapstructure:"home"`
	ProxyApp    string `mapstructure:"proxy_app"`
	Moniker     string `mapstructure:"moniker"`
	DBBackend   string `mapstructure:"db_backend"`
	DBPath      string `mapstructure:"db_dir"`
	FilterPeers bool   `mapstructure:"filter_peers"`
}

func DefaultBaseConfig() BaseConfig {
	return BaseConfig{
		Version:     "0.0.1",
		Moniker:     defaultMoniker,
		FilterPeers: false,
		DBBackend:   "goleveldb",
		DBPath:      DefaultDataDir,
	}
}

type RPCConfig struct {
	mu                        sync.RWMutex
	RootDir                   string        `mapstructure:"home"`
	ListenAddress             string        `mapstructure:"laddr"`
	CORSAllowedOrigins        []string      `mapstructure:"cors_allowed_origins"`
	CORSAllowedMethods        []string      `mapstructure:"cors_allowed_methods"`
	CORSAllowedHeaders        []string      `mapstructure:"cors_allowed_headers"`
	GRPCListenAddress         string        `mapstructure:"grpc_laddr"`
	GRPCMaxOpenConnections    int           `mapstructure:"grpc_max_open_connections"`
	Unsafe                    bool          `mapstructure:"unsafe"`
	MaxOpenConnections        int           `mapstructure:"max_open_connections"`
	MaxSubscriptionClients    int           `mapstructure:"max_subscription_clients"`
	MaxSubscriptionsPerClient int           `mapstructure:"max_subscriptions_per_client"`
	SubscriptionBufferSize    int           `mapstructure:"experimental_subscription_buffer_size"`
	WebSocketWriteBufferSize  int           `mapstructure:"experimental_websocket_write_buffer_size"`
	CloseOnSlowClient         bool          `mapstructure:"experimental_close_on_slow_client"`
	TimeoutBroadcastTxCommit  time.Duration `mapstructure:"timeout_broadcast_tx_commit"`
	MaxBodyBytes              int64         `mapstructure:"max_body_bytes"`
	MaxHeaderBytes            int           `mapstructure:"max_header_bytes"`
	TLSCertFile               string        `mapstructure:"tls_cert_file"`
	TLSKeyFile                string        `mapstructure:"tls_key_file"`
	PprofListenAddress        string        `mapstructure:"pprof_laddr"`
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
	RootDir         string  `mapstructure:"home"`
	NodeId          peer.ID `mapstructure:"node_id"`
	ListenAddress   string  `mapstructure:"laddr"`
	ExternalAddress string  `mapstructure:"external_address"`
	Seeds           string  `mapstructure:"seeds"`
	PersistentPeers string  `mapstructure:"persistent_peers"`
}

func DefaultP2PConfig() *P2PConfig {
	return &P2PConfig{
		ListenAddress:   "tcp://0.0.0.0:2300",
		ExternalAddress: "",
		NodeId:          "",
	}
}

// StateSyncConfig holds configuration settings related to syncing pods.
type StateSyncConfig struct {
	Enable            bool          `mapstructure:"enable"`              // Enable or disable pod syncing
	TempDir           string        `mapstructure:"temp_dir"`            // Directory for temporary storage during pod syncing
	RPCServers        []string      `mapstructure:"rpc_servers"`         // List of RPC servers for fetching pods
	PodTrustPeriod    time.Duration `mapstructure:"pod_trust_period"`    // Period for which a pod is considered trusted
	PodTrustHeight    int64         `mapstructure:"pod_trust_height"`    // Height at which the pod's trust starts
	PodTrustHash      string        `mapstructure:"pod_trust_hash"`      // Hash of a trusted pod to start syncing from
	PodDiscoveryTime  time.Duration `mapstructure:"pod_discovery_time"`  // Time for discovering new pods
	PodRequestTimeout time.Duration `mapstructure:"pod_request_timeout"` // Timeout for pod requests
	PodChunkFetchers  int32         `mapstructure:"pod_chunk_fetchers"`  // Number of concurrent fetchers for pod chunks
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
	TimeoutPropose        time.Duration `mapstructure:"timeout_propose"`
	TimeoutProposeDelta   time.Duration `mapstructure:"timeout_propose_delta"`
	TimeoutPrevote        time.Duration `mapstructure:"timeout_prevote"`
	TimeoutPrevoteDelta   time.Duration `mapstructure:"timeout_prevote_delta"`
	TimeoutPrecommit      time.Duration `mapstructure:"timeout_precommit"`
	TimeoutPrecommitDelta time.Duration `mapstructure:"timeout_precommit_delta"`
	TimeoutCommit         time.Duration `mapstructure:"timeout_commit"`

	// Configuration to skip the commit timeout for faster consensus on pods
	SkipTimeoutCommit bool `mapstructure:"skip_timeout_commit"`

	// Pod-specific configurations
	ValidatePods               bool          `mapstructure:"validate_pods"`                 // Whether to validate pods before accepting them
	PodValidationSleepDuration time.Duration `mapstructure:"pod_validation_sleep_duration"` // Sleep duration between pod validations

	DoubleSignCheckHeight int64 `mapstructure:"double_sign_check_height"` // Height to check for double signing
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
	DaType string
	DaRPC  string
	DaKey  string
}

func DefaultDAConfig() *DAConfig {
	return &DAConfig{
		DaType: "",
		DaRPC:  "",
		DaKey:  "",
	}
}

type StationConfig struct {
	StationType string
	StationRPC  string
	StationAPI  string
}

// DefaultStationConfig returns a default configuration for the station.
func DefaultStationConfig() *StationConfig {
	return &StationConfig{
		StationType: "",
		StationRPC:  "",
		StationAPI:  "",
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
