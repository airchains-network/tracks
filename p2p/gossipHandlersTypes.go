package p2p

type VRFInitiatedMsgData struct {
	PodNumber            uint64
	SelectedTrackAddress string
	VrfInitiatorAddress  string
}

type VRFVerifiedMsg struct {
	PodNumber            uint64
	SelectedTrackAddress string
}

type PodSubmittedMsgData struct {
	PodNumber            uint64
	SelectedTrackAddress string
}

type PodVerifiedMsgData struct {
	PodNumber          uint64
	VerificationResult bool
}
