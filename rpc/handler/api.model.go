package handler

type RequestBody struct {
	JsonRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      string        `json:"id"`
}

type ErrorDetails struct {
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"` // Additional error information
}

type ResponseBody struct {
	JsonRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Error   *ErrorDetails `json:"error"`
	Result  interface{}   `json:"result"` // use interface{} for arbitrary result structure
}
type ErrorMsg struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}
