package config

import (
	"bytes"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/utilis"
	"path/filepath"
	"strings"
	"text/template"
)

// DefaultDirPerm is the default permissions used when creating directories.
const DefaultDirPerm = 0700

var configTemplate *template.Template

func init() {
	var err error
	tmpl := template.New("configFileTemplate").Funcs(template.FuncMap{
		"StringsJoin": strings.Join,
	})
	if configTemplate, err = tmpl.Parse(defaultConfigTemplate); err != nil {
		panic(err)
	}
}

/****** these are for production settings ***********/

// EnsureRoot creates the root, config, and data directories if they don't exist,
// and panics if it fails.
func EnsureRoot(rootDir string, config *Config) {
	fmt.Println(rootDir)
	if err := utilis.EnsureDir(rootDir, DefaultDirPerm); err != nil {
		panic(err.Error())
	}
	if err := utilis.EnsureDir(filepath.Join(rootDir, DefaultConfigDir), DefaultDirPerm); err != nil {
		panic(err.Error())
	}
	if err := utilis.EnsureDir(filepath.Join(rootDir, DefaultDataDir), DefaultDirPerm); err != nil {
		panic(err.Error())
	}

	configFilePath := filepath.Join(rootDir, defaultConfigFilePath)

	// Write default config file if missing.
	if !utilis.FileExists(configFilePath) {
		writeDefaultConfigFile(configFilePath, config)
	}
}

func writeDefaultConfigFile(configFilePath string, config *Config) {
	WriteConfigFile(configFilePath, config)
}

func WriteConfigFile(configFilePath string, config *Config) {
	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, config); err != nil {
		panic(err)
	}

	utilis.MustWriteFile(configFilePath, buffer.Bytes(), 0644)
}

const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

version = "0.1"  

[rpc]
laddr = "{{ .RPC.ListenAddress }}"
cors_allowed_origins = [{{ range .RPC.CORSAllowedOrigins }} "{{ . }}", {{ end }}]
cors_allowed_methods = [{{ range .RPC.CORSAllowedMethods }} "{{ . }}", {{ end }}]
cors_allowed_headers = [{{ range .RPC.CORSAllowedHeaders }} "{{ . }}", {{ end }}]
grpc_laddr = "{{ .RPC.GRPCListenAddress }}"
grpc_max_open_connections = {{ .RPC.GRPCMaxOpenConnections }}
unsafe = {{ .RPC.Unsafe }}
max_open_connections = {{ .RPC.MaxOpenConnections }}
max_subscription_clients = {{ .RPC.MaxSubscriptionClients }}
max_subscriptions_per_client = {{ .RPC.MaxSubscriptionsPerClient }}
experimental_subscription_buffer_size = {{ .RPC.SubscriptionBufferSize }}
experimental_websocket_write_buffer_size = {{ .RPC.WebSocketWriteBufferSize }}
experimental_close_on_slow_client = {{ .RPC.CloseOnSlowClient }}
timeout_broadcast_tx_commit = "{{ .RPC.TimeoutBroadcastTxCommit }}"
max_body_bytes = {{ .RPC.MaxBodyBytes }}
max_header_bytes = {{ .RPC.MaxHeaderBytes }}
tls_cert_file = "{{ .RPC.TLSCertFile }}"
tls_key_file = "{{ .RPC.TLSKeyFile }}"
pprof_laddr = "{{ .RPC.PprofListenAddress }}"

[p2p]
laddr = "{{ .P2P.ListenAddress }}"
external_address = "{{ .P2P.ExternalAddress }}"
seeds = "{{ .P2P.Seeds }}"
persistent_peers = "{{ .P2P.PersistentPeers }}"
max_num_inbound_peers = {{ .P2P.MaxNumInboundPeers }}
max_num_outbound_peers = {{ .P2P.MaxNumOutboundPeers }}
unconditional_peer_ids = "{{ .P2P.UnconditionalPeerIDs }}"
persistent_peers_max_dial_period = "{{ .P2P.PersistentPeersMaxDialPeriod }}"
flush_throttle_timeout = "{{ .P2P.FlushThrottleTimeout }}"
max_packet_msg_payload_size = {{ .P2P.MaxPacketMsgPayloadSize }}
send_rate = {{ .P2P.SendRate }}
recv_rate = {{ .P2P.RecvRate }}
pex = {{ .P2P.PexReactor }}
seed_mode = {{ .P2P.SeedMode }}
private_peer_ids = "{{ .P2P.PrivatePeerIDs }}"
allow_duplicate_ip = {{ .P2P.AllowDuplicateIP }}
handshake_timeout = "{{ .P2P.HandshakeTimeout }}"
dial_timeout = "{{ .P2P.DialTimeout }}"

[statesync]
enable = {{ .StateSync.Enable }}
temp_dir = "{{ .StateSync.TempDir }}"
rpc_servers = [{{ range .StateSync.RPCServers }} "{{ . }}", {{ end }}]
pod_trust_period = "{{ .StateSync.PodTrustPeriod }}"
pod_trust_height = {{ .StateSync.PodTrustHeight }}
pod_trust_hash = "{{ .StateSync.PodTrustHash }}"
pod_discovery_time = "{{ .StateSync.PodDiscoveryTime }}"
pod_request_timeout = "{{ .StateSync.PodRequestTimeout }}"
pod_chunk_fetchers = {{ .StateSync.PodChunkFetchers }}

[consensus]
timeout_propose = "{{ .Consensus.TimeoutPropose }}"
timeout_propose_delta = "{{ .Consensus.TimeoutProposeDelta }}"
timeout_prevote = "{{ .Consensus.TimeoutPrevote }}"
timeout_prevote_delta = "{{ .Consensus.TimeoutPrevoteDelta }}"
timeout_precommit = "{{ .Consensus.TimeoutPrecommit }}"
timeout_precommit_delta = "{{ .Consensus.TimeoutPrecommitDelta }}"
timeout_commit = "{{ .Consensus.TimeoutCommit }}"
skip_timeout_commit = {{ .Consensus.SkipTimeoutCommit }}

double_sign_check_height = {{ .Consensus.DoubleSignCheckHeight }}

# Data Availability Layer Configuration
[da]
daType = "{{ .DA.DaType }}"
daRPC = "{{ .DA.DaRPC }}"

# Station Configuration
[station]
stationType = "{{ .Station.StationType }}"
stationRPC = "{{ .Station.StationRPC }}"

# Junction Configuration
[junction]
junctionRPC =  "{{ .Junction.JunctionRPC }}"
junctionAPI =  "{{ .Junction.JunctionAPI }}"
# Add specific configuration keys and values for Junction here

`
