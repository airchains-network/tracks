package types

import "time"

type BatchTransaction struct {
	Tx         Tx         `json:"tx"`
	TxResponse TxResponse `json:"tx_response"`
}

type BatchTransactions struct {
	Tx         []Tx         `json:"tx"`
	TxResponse []TxResponse `json:"tx_response"`
}

type Tx struct {
	Body       BatchBody     `json:"body"`
	AuthInfo   BatchAuthInfo `json:"auth_info"`
	Signatures []string      `json:"signatures"`
}

type BatchBody struct {
	Messages                    []Message     `json:"messages"`
	Memo                        string        `json:"memo"`
	TimeoutHeight               string        `json:"timeout_height"`
	ExtensionOptions            []interface{} `json:"extension_options"`
	NonCriticalExtensionOptions []interface{} `json:"non_critical_extension_options"`
}

type Message struct {
	Type        string        `json:"@type"`
	FromAddress string        `json:"from_address"`
	ToAddress   string        `json:"to_address"`
	Amount      []BatchAmount `json:"amount"`
}

type BatchAmount struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type BatchAuthInfo struct {
	SignerInfos []BatchSignerInfo `json:"signer_infos"`
	Fee         BatchFee          `json:"fee"`
	Tip         interface{}       `json:"tip"`
}

type BatchSignerInfo struct {
	PublicKey BatchPublicKey `json:"public_key"`
	ModeInfo  BatchModeInfo  `json:"mode_info"`
	Sequence  string         `json:"sequence"`
}

type BatchPublicKey struct {
	Type string `json:"@type"`
	Key  string `json:"key"`
}

type BatchModeInfo struct {
	Single Single `json:"single"`
}

type Single struct {
	Mode string `json:"mode"`
}

type BatchFee struct {
	Amount   []interface{} `json:"amount"`
	GasLimit string        `json:"gas_limit"`
	Payer    string        `json:"payer"`
	Granter  string        `json:"granter"`
}

type TxResponse struct {
	Height    string    `json:"height"`
	TxHash    string    `json:"txhash"`
	Codespace string    `json:"codespace"`
	Code      int       `json:"code"`
	Data      string    `json:"data"`
	RawLog    string    `json:"raw_log"`
	Logs      []Log     `json:"logs"`
	Info      string    `json:"info"`
	GasWanted string    `json:"gas_wanted"`
	GasUsed   string    `json:"gas_used"`
	Tx        Tx        `json:"tx"`
	Timestamp time.Time `json:"timestamp"`
	Events    []Event   `json:"events"`
}

type Log struct {
	MsgIndex int          `json:"msg_index"`
	Log      string       `json:"log"`
	Events   []BatchEvent `json:"events"`
}

type BatchEvent struct {
	Type       string      `json:"type"`
	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Index bool   `json:"index"`
}
type Event struct {
	Type       string           `json:"type"`
	Attributes []EventAttribute `json:"attributes"`
}
type EventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Index bool   `json:"index"`
}

type GetTransactionStruct struct {
	To              string
	From            string
	Amount          string
	FromBalances    string
	ToBalances      string
	TransactionHash string
}
