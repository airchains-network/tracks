package types

type BatchStruct struct {
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
