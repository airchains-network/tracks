package svmTypes

type Params struct {
	Encoding                       any    `json:"encoding"`
	MaxSupportedTransactionVersion int    `json:"maxSupportedTransactionVersion"`
	TransactionDetails             string `json:"transactionDetails"`
	Rewards                        bool   `json:"rewards"`
}

type PayloadStruct struct {
	JsonRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Error   Error  `json:"error"`
	ID      int    `json:"id"`
}

type LatestSlotStruct struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  int    `json:"result"`
	ID      int    `json:"id"`
}

type BlockResponseStruct struct {
	JsonRpc string `json:"jsonrpc"`
	Result  struct {
		BlockHeight       int                    `json:"blockHeight"`
		BlockTime         int                    `json:"blockTime"`
		Blockhash         string                 `json:"blockhash"`
		ParentSlot        int                    `json:"parentSlot"`
		PreviousBlockhash string                 `json:"previousBlockhash"`
		Transactions      []SVMTransactionStruct `json:"transactions"`
	} `json:"result"`
	ID int `json:"id"`
}

type SVMTransactionStruct struct {
	Meta struct {
		ComputeUnitsConsumed int           `json:"computeUnitsConsumed"`
		Err                  interface{}   `json:"err"`
		Fee                  int           `json:"fee"`
		InnerInstructions    []interface{} `json:"innerInstructions"`
		LogMessages          []string      `json:"logMessages"`
		PostBalances         []interface{} `json:"postBalances"`
		PostTokenBalances    []interface{} `json:"postTokenBalances"`
		PreBalances          []interface{} `json:"preBalances"`
		PreTokenBalances     []interface{} `json:"preTokenBalances"`
		Rewards              interface{}   `json:"rewards"`
		Status               struct {
			Ok interface{} `json:"Ok"`
		} `json:"status"`
	} `json:"meta"`
	Transaction struct {
		Message struct {
			AccountKeys []struct {
				Pubkey   string `json:"pubkey"`
				Signer   bool   `json:"signer"`
				Source   string `json:"source"`
				Writable bool   `json:"writable"`
			} `json:"accountKeys"`
			Instructions []struct {
				Parsed struct {
					Info struct {
						VoteAccount     string `json:"voteAccount"`
						VoteAuthority   string `json:"voteAuthority"`
						VoteStateUpdate struct {
							Hash     string `json:"hash"`
							Lockouts []struct {
								ConfirmationCount int `json:"confirmation_count"`
								Slot              int `json:"slot"`
							} `json:"lockouts"`
							Root      int `json:"root"`
							Timestamp int `json:"timestamp"`
						} `json:"voteStateUpdate"`
					} `json:"info"`
					Type string `json:"type"`
				} `json:"parsed"`
				Program     string      `json:"program"`
				ProgramID   string      `json:"programId"`
				StackHeight interface{} `json:"stackHeight"`
			} `json:"instructions"`
			RecentBlockhash string `json:"recentBlockhash"`
		} `json:"message"`
		Signatures []string `json:"signatures"`
	} `json:"transaction"`
	Version string `json:"version"`
}

type SlotLeaderResponseStruct struct {
	JsonRpc string   `json:"jsonrpc"`
	Result  []string `json:"result"`
	ID      int      `json:"id"`
}

type LargeAccountStruct struct {
	JsonRpc string `json:"jsonrpc"`
	Result  struct {
		Context struct {
			APIVersion string `json:"apiVersion"`
			Slot       int    `json:"slot"`
		} `json:"context"`
		Value []struct {
			Address string      `json:"address"`
			Lamport interface{} `json:"lamports"`
		} `json:"value"`
	} `json:"result"`
	ID int `json:"id"`
}

type AccountDetailsStruct struct {
	JsonRpc string `json:"jsonrpc"`
	Result  struct {
		Context struct {
			APIVersion string `json:"apiVersion"`
			Slot       int    `json:"slot"`
		} `json:"context"`
		Value []AccountDetailValueStruct `json:"value"`
	} `json:"result"`
	ID int `json:"id"`
}

type AccountDetailValueStruct struct {
	Data       interface{} `json:"data"`
	Executable bool        `json:"executable"`
	Lamport    interface{} `json:"lamports"`
	Owner      string      `json:"owner"`
	RentEpoch  interface{} `json:"rentEpoch"`
	Space      int         `json:"space"`
}

type AccountDetailsResponseStruct struct {
	Address string                   `json:"address"`
	Value   AccountDetailValueStruct `json:"data"`
}
