package model

// '{"jsonrpc":"2.0","method":"tracks_getPodsMaster","params":["0x2"],"id":1}'
type RequestBody struct {
	JsonRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	ID      string `json:"id"`
}

type ErrorMsg struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

type ResponseBody struct {
	JsonRPC string   `json:"jsonrpc"`
	ID      string   `json:"id"`
	Error   ErrorMsg `json:"error"`
	Result  []any    `json:"result"`
}

type RequestPodParams struct {
}
