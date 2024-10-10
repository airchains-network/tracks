package types

type Payload struct {
	Namespace int    `json:"namespace"`
	Payload   string `json:"payload"`
}

type EspressoTxResponseV1Temp struct {
	Transaction struct {
		Namespace int    `json:"namespace"`
		Payload   string `json:"payload"`
	} `json:"transaction"`
	Hash  string `json:"hash"`
	Index int    `json:"index"`
	Proof struct {
		TxIndex            []int `json:"tx_index"`
		PayloadNumTxs      []int `json:"payload_num_txs"`
		PayloadProofNumTxs struct {
			Proofs      string `json:"proofs"`
			PrefixBytes []int  `json:"prefix_bytes"`
			SuffixBytes []int  `json:"suffix_bytes"`
		} `json:"payload_proof_num_txs"`
		PayloadTxTableEntries      []int `json:"payload_tx_table_entries"`
		PayloadProofTxTableEntries struct {
			Proofs      string `json:"proofs"`
			PrefixBytes []int  `json:"prefix_bytes"`
			SuffixBytes []int  `json:"suffix_bytes"`
		} `json:"payload_proof_tx_table_entries"`
		PayloadProofTx struct {
			Proofs      string `json:"proofs"`
			PrefixBytes []int  `json:"prefix_bytes"`
			SuffixBytes []int  `json:"suffix_bytes"`
		} `json:"payload_proof_tx"`
	} `json:"proof"`
	BlockHash   string `json:"block_hash"`
	BlockHeight int    `json:"block_height"`
}

type EspressoTxResponseV1 struct {
	Transaction struct {
		Namespace int    `json:"namespace"`
		Payload   string `json:"payload"`
	} `json:"transaction"`
	Hash  string `json:"hash"`
	Index int    `json:"index"`
	Proof struct {
		TxIndex            string `json:"tx_index"`
		PayloadNumTxs      string `json:"payload_num_txs"`
		PayloadProofNumTxs struct {
			Proofs      string `json:"proofs"`
			PrefixBytes string `json:"prefix_bytes"`
			SuffixBytes string `json:"suffix_bytes"`
		} `json:"payload_proof_num_txs"`
		PayloadTxTableEntries      string `json:"payload_tx_table_entries"`
		PayloadProofTxTableEntries struct {
			Proofs      string `json:"proofs"`
			PrefixBytes string `json:"prefix_bytes"`
			SuffixBytes string `json:"suffix_bytes"`
		} `json:"payload_proof_tx_table_entries"`
		PayloadProofTx struct {
			Proofs      string `json:"proofs"`
			PrefixBytes string `json:"prefix_bytes"`
			SuffixBytes string `json:"suffix_bytes"`
		} `json:"payload_proof_tx"`
	} `json:"proof"`
	BlockHash   string `json:"block_hash"`
	BlockHeight int    `json:"block_height"`
}

// EspressoSchemaV1 Define the combined struct to match the JSON structure
type EspressoSchemaV1 struct {
	EspressoTxResponseV1 struct {
		Transaction struct {
			Namespace int    `json:"namespace"`
			Payload   string `json:"payload"`
		} `json:"transaction"`
		Hash  string `json:"hash"`
		Index int    `json:"index"`
		Proof struct {
			TxIndex            string `json:"tx_index"`
			PayloadNumTxs      string `json:"payload_num_txs"`
			PayloadProofNumTxs struct {
				Proofs      string `json:"proofs"`
				PrefixBytes string `json:"prefix_bytes"`
				SuffixBytes string `json:"suffix_bytes"`
			} `json:"payload_proof_num_txs"`
			PayloadTxTableEntries      string `json:"payload_tx_table_entries"`
			PayloadProofTxTableEntries struct {
				Proofs      string `json:"proofs"`
				PrefixBytes string `json:"prefix_bytes"`
				SuffixBytes string `json:"suffix_bytes"`
			} `json:"payload_proof_tx_table_entries"`
			PayloadProofTx struct {
				Proofs      string `json:"proofs"`
				PrefixBytes string `json:"prefix_bytes"`
				SuffixBytes string `json:"suffix_bytes"`
			} `json:"payload_proof_tx"`
		} `json:"proof"`
		BlockHash   string `json:"block_hash"`
		BlockHeight int    `json:"block_height"`
	} `json:"espresso_tx_response_v_1"`
	StationId string `json:"station_id"`
}

type SchemaDef struct {
	Fields map[string]interface{} // Changed to interface{} to allow nested definitions
}
