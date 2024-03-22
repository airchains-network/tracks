package types

type ChainInfoStruct struct {
	ChainInfo struct {
		ChainID string `json:"chainID"` // Chain ID of the chain which is taken from the user
		Key     string `json:"key"`     // Name of the account
		Moniker string `json:"moniker"` // Name of the chain
	} `json:"chainInfo"`
	DaInfo struct {
		DaSelected      string `json:"daSelected"`
		DaWalletAddress string `json:"daWalletAddress"`
		DaWalletKeypair string `json:"daWalletKeypair"`
	} `json:"daInfo"`
	SequencerInfo struct {
		SequencerType string `json:"sequencerType"`
	} `json:"sequencerInfo"`
}

type SettlementClientResponseStruct struct {
	Status      bool   `json:"status"`
	Data        string `json:"data"`
	Description string `json:"description"`
}

type SLVerificationKeyStruct struct {
	G1 struct {
		Alpha struct {
			X string `json:"X"`
			Y string `json:"Y"`
		} `json:"Alpha"`
		Beta struct {
			X string `json:"X"`
			Y string `json:"Y"`
		} `json:"Beta"`
		Delta struct {
			X string `json:"X"`
			Y string `json:"Y"`
		} `json:"Delta"`
		K []struct {
			X string `json:"X"`
			Y string `json:"Y"`
		} `json:"K"`
	} `json:"G1"`
	G2 struct {
		Beta struct {
			X struct {
				A0 string `json:"A0"`
				A1 string `json:"A1"`
			} `json:"X"`
			Y struct {
				A0 string `json:"A0"`
				A1 string `json:"A1"`
			} `json:"Y"`
		} `json:"Beta"`
		Delta struct {
			X struct {
				A0 string `json:"A0"`
				A1 string `json:"A1"`
			} `json:"X"`
			Y struct {
				A0 string `json:"A0"`
				A1 string `json:"A1"`
			} `json:"Y"`
		} `json:"Delta"`
		Gamma struct {
			X struct {
				A0 string `json:"A0"`
				A1 string `json:"A1"`
			} `json:"X"`
			Y struct {
				A0 string `json:"A0"`
				A1 string `json:"A1"`
			} `json:"Y"`
		} `json:"Gamma"`
	} `json:"G2"`
	CommitmentKey struct {
	} `json:"CommitmentKey"`
	PublicAndCommitmentCommitted []any `json:"PublicAndCommitmentCommitted"`
}

type SLProofStruct struct {
	Ar struct {
		X string `json:"X"`
		Y string `json:"Y"`
	} `json:"Ar"`
	Krs struct {
		X string `json:"X"`
		Y string `json:"Y"`
	} `json:"Krs"`
	Bs struct {
		X struct {
			A0 string `json:"A0"`
			A1 string `json:"A1"`
		} `json:"X"`
		Y struct {
			A0 string `json:"A0"`
			A1 string `json:"A1"`
		} `json:"Y"`
	} `json:"Bs"`
	Commitments   []any `json:"Commitments"`
	CommitmentPok struct {
		X int `json:"X"`
		Y int `json:"Y"`
	} `json:"CommitmentPok"`
}

type ExtraArg struct {
	SerializedRc []byte `json:"serializedRc"`
	Proof        []byte `json:"proof"`
	VrfOutput    []byte `json:"vrfOutput"`
}
