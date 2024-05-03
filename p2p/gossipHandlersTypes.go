package p2p

type VRFInitiatedMsgData struct {
	PodNumber            uint64
	SelectedTrackAddress string
	VrfInitiatorAddress  string
	VrfInitTxHash        string
}

type VRFVerifiedMsg struct {
	PodNumber            uint64
	SelectedTrackAddress string
	VRFVerifiedTxHash    string
}

type PodSubmittedMsgData struct {
	PodNumber            uint64
	SelectedTrackAddress string
	InitPodTxHash        string
}

type PodVerifiedMsgData struct {
	PodNumber          uint64
	VerificationResult bool
	PodVerifiedTxHash  string
}
