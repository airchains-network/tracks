package types

type BlockStruct struct {
	BaseFeePerGas    string `json:"basefeepergas"`
	Difficulty       string `json:"difficulty"`
	ExtraData        string `json:"extradata"`
	GasLimit         string `json:"gaslimit"`
	GasUsed          string `json:"gasused"`
	Hash             string `json:"hash"`
	LogsBloom        string `json:"logsbloom"`
	Miner            string `json:"miner"`
	MixHash          string `json:"mixhash"`
	Nonce            string `json:"nonce"`
	Number           string `json:"number"`
	ParentHash       string `json:"parenthash"`
	ReceiptsRoot     string `json:"receiptsroot"`
	Sha3Uncles       string `json:"sha3uncles"`
	Size             string `json:"size"`
	StateRoot        string `json:"stateroot"`
	Timestamp        string `json:"timestamp"`
	TotalDifficulty  string `json:"totaldifficulty"`
	TransactionCount int    `json:"transactioncount"`
	TransactionsRoot string `json:"transactionsroot"`
	Uncles           string `json:"uncles"`
}

type TransactionStruct struct {
	BlockHash        string `json:"blockHash"`
	BlockNumber      uint64 `json:"blockNumber"`
	From             string `json:"from"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Hash             string `json:"hash"`
	Input            string `json:"input"`
	Nonce            string `json:"nonce"`
	R                string `json:"r"`
	S                string `json:"s"`
	To               string `json:"to"`
	TransactionIndex string `json:"transactionIndex"`
	Type             string `json:"type"`
	V                string `json:"v"`
	Value            string `json:"value"`
}

type PodStruct struct {
	From              []string `json:"from"`
	To                []string `json:"to"`
	Amounts           []string `json:"amounts"`
	TransactionHash   []string `json:"tx_hashes"`
	SenderBalances    []string `json:"sender_balances"`
	ReceiverBalances  []string `json:"receiver_balances"`
	Messages          []string `json:"messages"`
	TransactionNonces []string `json:"tx_nonces"`
	AccountNonces     []string `json:"account_nonces"`
}

type DAStruct struct {
	DAKey             string `json:"da_key"`
	DAClientName      string `json:"da_client_name"`
	BatchNumber       string `json:"batch_number"`
	PreviousStateHash string `json:"previous_state_hash"`
	CurrentStateHash  string `json:"current_state_hash"`
}

// SettlementLayerChainInfoStruct ChainInfoStruct is the struct for chainInfo.json file
type SettlementLayerChainInfoStruct struct {
	ChainId   string `json:"chain_id"`
	ChainName string `json:"chain_name"`
}
