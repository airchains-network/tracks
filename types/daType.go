package types

type DAConfigType struct {
	DALayer    string
	DARpc      string
	AccountKey string
	DAAuth     string
}

type AvailSuccessResponse struct {
	BlockNumber int    `json:"block_number"`
	BlockHash   string `json:"block_hash"`
	Hash        string `json:"hash"`
	Index       int    `json:"index"`
}

type CelestiaSuccessResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  int    `json:"result"`
	ID      int    `json:"id"`
}

type CelestiaErrorResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type MockDAStruck struct {
	DataBlob    []byte
	BatchNumber int
	Commitment  string
}

type DAStruct struct {
	DAKey             string `json:"da_key"`
	DAClientName      string `json:"da_client_name"`
	BatchNumber       string `json:"batch_number"`
	PreviousStateHash string `json:"previous_state_hash"`
	CurrentStateHash  string `json:"current_state_hash"`
}

type FinalizeDA struct {
	CompressedHash []string
	Proof          []byte
	PodNumber      int
}
