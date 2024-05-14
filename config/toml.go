package config

import (
	"bytes"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/utils"
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

// CreateConfigFile creates the root, config, and data directories if they don't exist,
// and panics if it fails.
func CreateConfigFile(rootDir string, config *Config) (success bool) {
	if err := utils.EnsureDir(rootDir, DefaultDirPerm); err != nil {
		logs.Log.Error(err.Error())
		return false
	}
	if err := utils.EnsureDir(filepath.Join(rootDir, DefaultConfigDir), DefaultDirPerm); err != nil {
		logs.Log.Error(err.Error())
		return false
	}
	if err := utils.EnsureDir(filepath.Join(rootDir, DefaultDataDir), DefaultDirPerm); err != nil {
		logs.Log.Error(err.Error())
		return false
	}

	configFilePath := filepath.Join(rootDir, DefaultConfigFilePath)

	// Write default config file if missing.
	if !utils.FileExists(configFilePath) {
		writeDefaultConfigFile(configFilePath, config)
		return true
	} else {
		logs.Log.Error("Config file already exists at \n" + configFilePath + "\nPlease remove it and try again.")
		return false
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

	utils.MustWriteFile(configFilePath, buffer.Bytes(), 0644)
}

/*
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml
*/

const defaultConfigTemplate = `[base_config]
db_backend="goleveldb"
db_path="{{ .BaseConfig.DBPath }}"
filter_peers={{ .BaseConfig.FilterPeers }}
moniker="{{ .BaseConfig.Moniker }}"
proxy_app="{{ .BaseConfig.ProxyApp }}"
root_dir="{{ .BaseConfig.RootDir }}"
version="0.0.1"

[consensus]
double_sign_check_height = {{ .Consensus.DoubleSignCheckHeight }}
pod_validation_sleep_duration = "{{ .Consensus.PodValidationSleepDuration }}"
skip_timeout_commit = {{ .Consensus.SkipTimeoutCommit }}
timeout_commit = "{{ .Consensus.TimeoutCommit }}"
timeout_precommit = "{{ .Consensus.TimeoutPrecommit }}"
timeout_precommit_delta = "{{ .Consensus.TimeoutPrecommitDelta }}"
timeout_prevote = "{{ .Consensus.TimeoutPrevote }}"
timeout_prevote_delta = "{{ .Consensus.TimeoutPrevoteDelta }}"
timeout_propose = "{{ .Consensus.TimeoutPropose }}"
timeout_propose_delta = "{{ .Consensus.TimeoutProposeDelta }}"
validatePods = {{ .Consensus.ValidatePods }}

[da]
daKey = "{{ .DA.DaKey }}"
daRPC = "{{ .DA.DaRPC }}"
daType = "{{ .DA.DaType }}"

[junction]
accountName = "{{ .Junction.AccountName }}"
accountPath = "{{ .Junction.AccountPath }}"
AddressPrefix = "{{ .Junction.AddressPrefix }}"
junctionAPI =  "{{ .Junction.JunctionAPI }}"
junctionRPC =  "{{ .Junction.JunctionRPC }}"
stationId = "{{ .Junction.StationId }}"
Tracks = {{ .Junction.Tracks }}
VRFPrivateKey = "{{ .Junction.VRFPrivateKey }}"
VRFPublicKey = "{{ .Junction.VRFPublicKey }}"

[p2p]
currently_connected_peers = {{ .P2P.CurrentlyConnectedPeers }}
external_address = "{{ .P2P.ExternalAddress }}"
listen_address = "{{ .P2P.ListenAddress }}"
node_id = "{{ .P2P.NodeId }}"
persistent_peers = {{ .P2P.PersistentPeers }}
root_dir = "{{ .P2P.RootDir }}"
seeds = "{{ .P2P.Seeds }}"

[rpc]
close_on_slow_client = {{ .RPC.CloseOnSlowClient }}
cors_allowed_headers = [{{ range .RPC.CORSAllowedHeaders }} "{{ . }}", {{ end }}]
cors_allowed_methods = [{{ range .RPC.CORSAllowedMethods }} "{{ . }}", {{ end }}]
cors_allowed_origins = [{{ range .RPC.CORSAllowedOrigins }} "{{ . }}", {{ end }}]
grpc_listen_address = "{{ .RPC.GRPCListenAddress }}"
grpc_max_open_connections = {{ .RPC.GRPCMaxOpenConnections }}
listen_address="{{ .RPC.ListenAddress }}"
max_body_bytes = {{ .RPC.MaxBodyBytes }}
max_header_bytes = {{ .RPC.MaxHeaderBytes }}
max_open_connections = {{ .RPC.MaxOpenConnections }}
max_subscription_clients = {{ .RPC.MaxSubscriptionClients }}
max_subscriptions_per_client = {{ .RPC.MaxSubscriptionsPerClient }}
pprof_listen_address = "{{ .RPC.PprofListenAddress }}"
root_dir="{{ .RPC.RootDir }}"
subscription_buffer_size = {{ .RPC.SubscriptionBufferSize }}
timeout_broadcast_tx_commit = "{{ .RPC.TimeoutBroadcastTxCommit }}"
tls_cert_file = "{{ .RPC.TLSCertFile }}"
tls_key_file = "{{ .RPC.TLSKeyFile }}"
unsafe = {{ .RPC.Unsafe }}
web_socket_write_buffer_size = {{ .RPC.WebSocketWriteBufferSize }}

[statesync]
enable = {{ .StateSync.Enable }}
pod_chunk_fetchers = {{ .StateSync.PodChunkFetchers }}
pod_discovery_time = "{{ .StateSync.PodDiscoveryTime }}"
pod_request_timeout = "{{ .StateSync.PodRequestTimeout }}"
pod_trust_hash = "{{ .StateSync.PodTrustHash }}"
pod_trust_height = {{ .StateSync.PodTrustHeight }}
pod_trust_period = "{{ .StateSync.PodTrustPeriod }}"
rpc_servers = [{{ range .StateSync.RPCServers }} "{{ . }}", {{ end }}]
temp_dir = "{{ .StateSync.TempDir }}"

[station]
stationAPI = "{{ .Station.StationAPI }}"
stationRPC = "{{ .Station.StationRPC }}"
stationType = "{{ .Station.StationType }}"

`
