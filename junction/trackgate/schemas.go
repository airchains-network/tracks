package trackgate

import (
	"github.com/airchains-network/tracks/types"
)

var VersionNameV1 = "1.0.0"
var SchemaV1 = types.SchemaDef{
	Fields: map[string]interface{}{
		"espresso_tx_response_v_1": map[string]interface{}{
			"transaction": map[string]interface{}{
				"namespace": "int",
				"payload":   "string",
			},
			"hash":  "string",
			"index": "int",
			"proof": map[string]interface{}{
				"tx_index":        "bytes",
				"payload_num_txs": "bytes",
				"payload_proof_num_txs": map[string]interface{}{
					"proofs":       "string",
					"prefix_bytes": "bytes",
					"suffix_bytes": "bytes",
				},
				"payload_tx_table_entries": "bytes",
				"payload_proof_tx_table_entries": map[string]interface{}{
					"proofs":       "string",
					"prefix_bytes": "bytes",
					"suffix_bytes": "bytes",
				},
				"payload_proof_tx": map[string]interface{}{
					"proofs":       "string",
					"prefix_bytes": "bytes",
					"suffix_bytes": "bytes",
				},
			},
			"block_hash":   "string",
			"block_height": "int",
		},
		"station_id": "string",
		"pod_number": "int",
	},
}
