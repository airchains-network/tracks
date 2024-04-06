package types

type DAConfigType struct {
	DALayer    string
	DARpc      string
	AccountKey string
	DAAuth     string
}

type AvailSuccessResponse struct {
	BlockNumber int
	BlockHash   string
	Hash        string
	Index       int
}

type CelestiaSuccessResponse struct {
	Jsonrpc string
	Result  int
	ID      int
}

type CelestiaErrorResponse struct {
	Jsonrpc string
	ID      int
	Error   struct {
		Code    int
		Message string
	} `json:"error"`
}

type MockDAStruck struct {
	DataBlob    []byte
	BatchNumber int
	Commitment  string
}

type DAStruct struct {
	DAKey             string
	DAClientName      string
	BatchNumber       string
	PreviousStateHash string
	CurrentStateHash  string
}

type FinalizeDA struct {
	CompressedHash []string
	Proof          []byte
	PodNumber      int
}
