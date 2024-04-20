package blocksync

import "time"

type BlockObject struct {
	Block struct {
		Header struct {
			ChainID string `json:"chain_id"`
			Height  string `json:"height"`
		} `json:"header"`
	} `json:"block"`
}
type Root struct {
	BlockID struct {
		Hash  string `json:"hash"`
		Parts struct {
			Total int    `json:"total"`
			Hash  string `json:"hash"`
		} `json:"parts"`
	} `json:"block_id"`
	Block struct {
		Header struct {
			Version struct {
				Block string `json:"block"`
			} `json:"version"`
			ChainID            string  `json:"chain_id"`
			Height             string  `json:"height"`
			Time               string  `json:"time"`
			LastBlockID        BlockID `json:"last_block_id"`
			LastCommitHash     string  `json:"last_commit_hash"`
			DataHash           string  `json:"data_hash"`
			ValidatorsHash     string  `json:"validators_hash"`
			NextValidatorsHash string  `json:"next_validators_hash"`
			ConsensusHash      string  `json:"consensus_hash"`
			AppHash            string  `json:"app_hash"`
			LastResultsHash    string  `json:"last_results_hash"`
			EvidenceHash       string  `json:"evidence_hash"`
			ProposerAddress    string  `json:"proposer_address"`
		} `json:"header"`
		Data struct {
			Txs []interface{} `json:"txs"`
		} `json:"data"`
		Evidence struct {
			Evidence []interface{} `json:"evidence"`
		} `json:"evidence"`
		LastCommit struct {
			Height  string `json:"height"`
			Round   int    `json:"round"`
			BlockID struct {
				Hash  string `json:"hash"`
				Parts struct {
					Total int    `json:"total"`
					Hash  string `json:"hash"`
				} `json:"parts"`
			} `json:"block_id"`
			Signatures []struct {
				BlockIDFlag      int    `json:"block_id_flag"`
				ValidatorAddress string `json:"validator_address"`
				Timestamp        string `json:"timestamp"`
				Signature        string `json:"signature"`
			} `json:"signatures"`
		} `json:"last_commit"`
	} `json:"block"`
}
type BlockID struct {
	Hash  string `json:"hash"`
	Parts struct {
		Total int    `json:"total"`
		Hash  string `json:"hash"`
	} `json:"parts"`
}
type Response struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		BlockID struct {
			Hash  string `json:"hash"`
			Parts struct {
				Total int    `json:"total"`
				Hash  string `json:"hash"`
			} `json:"parts"`
		} `json:"block_id"`
		Block struct {
			Header struct {
				Version struct {
					Block string `json:"block"`
				} `json:"version"`
				ChainID            string  `json:"chain_id"`
				Height             string  `json:"height"`
				Time               string  `json:"time"`
				LastBlockID        BlockID `json:"last_block_id"`
				LastCommitHash     string  `json:"last_commit_hash"`
				DataHash           string  `json:"data_hash"`
				ValidatorsHash     string  `json:"validators_hash"`
				NextValidatorsHash string  `json:"next_validators_hash"`
				ConsensusHash      string  `json:"consensus_hash"`
				AppHash            string  `json:"app_hash"`
				LastResultsHash    string  `json:"last_results_hash"`
				EvidenceHash       string  `json:"evidence_hash"`
				ProposerAddress    string  `json:"proposer_address"`
			} `json:"header"`
			Data struct {
				Txs []interface{} `json:"txs"`
			} `json:"data"`
			Evidence struct {
				Evidence []interface{} `json:"evidence"`
			} `json:"evidence"`
			LastCommit struct {
				Height  string `json:"height"`
				Round   int    `json:"round"`
				BlockID struct {
					Hash  string `json:"hash"`
					Parts struct {
						Total int    `json:"total"`
						Hash  string `json:"hash"`
					} `json:"parts"`
				} `json:"block_id"`
				Signatures []struct {
					BlockIDFlag      int    `json:"block_id_flag"`
					ValidatorAddress string `json:"validator_address"`
					Timestamp        string `json:"timestamp"`
					Signature        string `json:"signature"`
				} `json:"signatures"`
			} `json:"last_commit"`
		} `json:"block"`
	} `json:"result"`
}

// ! transactions
type Amount struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type PublicKey struct {
	Type string `json:"@type"`
	Key  string `json:"key"`
}

type ModeInfo struct {
	Single struct {
		Mode string `json:"mode"`
	} `json:"single"`
}

type SignerInfo struct {
	PublicKey PublicKey `json:"public_key"`
	ModeInfo  ModeInfo  `json:"mode_info"`
	Sequence  string    `json:"sequence"`
}

type Fee struct {
	Amount   []Amount `json:"amount"`
	GasLimit string   `json:"gas_limit"`
	Payer    string   `json:"payer"`
	Granter  string   `json:"granter"`
}

type AuthInfo struct {
	SignerInfos []SignerInfo `json:"signer_infos"`
	Fee         Fee          `json:"fee"`
	Tip         interface{}  `json:"tip"`
}

type Body struct {
	Messages                    []interface{} `json:"messages"`
	Memo                        string        `json:"memo"`
	TimeoutHeight               string        `json:"timeout_height"`
	ExtensionOptions            []interface{} `json:"extension_options"`
	NonCriticalExtensionOptions []interface{} `json:"non_critical_extension_options"`
}

type EventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Index bool   `json:"index"`
}

type Event struct {
	Type       string           `json:"type"`
	Attributes []EventAttribute `json:"attributes"`
}

type Transaction struct {
	TxResponse struct {
		Height    string        `json:"height"`
		TxHash    string        `json:"txhash"`
		Codespace string        `json:"codespace"`
		Code      int           `json:"code"`
		Data      string        `json:"data"`
		RawLog    string        `json:"raw_log"`
		Logs      []interface{} `json:"logs"`
		Info      string        `json:"info"`
		GasWanted string        `json:"gas_wanted"`
		GasUsed   string        `json:"gas_used"`
		Tx        struct {
			Type       string   `json:"@type"`
			Body       Body     `json:"body"`
			AuthInfo   AuthInfo `json:"auth_info"`
			Signatures []string `json:"signatures"`
		} `json:"tx"`
		Timestamp time.Time `json:"timestamp"`
		Events    []Event   `json:"events"`
	} `json:"tx_response"`
}
