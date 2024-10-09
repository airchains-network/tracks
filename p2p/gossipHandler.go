package p2p

import (
	"encoding/json"
	"fmt"
	"github.com/airchains-network/tracks/da/avail"
	"github.com/airchains-network/tracks/da/celestia"
	"github.com/airchains-network/tracks/da/eigen"
	mock "github.com/airchains-network/tracks/da/mockda"
	"github.com/airchains-network/tracks/junction"
	junction2 "github.com/airchains-network/tracks/junction/junction"
	junctionTypes "github.com/airchains-network/tracks/junction/junction/types"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/node/shared"
	"github.com/airchains-network/tracks/types"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const (
	LogPodDecodeSuccess     = "pod msg decoded successfully"
	LogPodMismatch          = "pod number is not matched, waiting for new pod"
	LogPodMatchSuccess      = "pod number matched successfully "
	LogPodSave              = "Saving current pod"
	LogPodGenNext           = "generating next pod"
	LogPodFail              = "Pod verification failed, its time for voting"
	LogPodExtractFail       = "Error in extracting PodVerifiedMsg"
	SleepDuration           = 3 * time.Second
	LogPodSubmitExtractFail = "Error in extracting PodSubmittedMsg"
	LogPodSubmitMatchFail   = "pod number is not matched, waiting for new pod"
	LogJunctionDetailsFail  = "can not get junctionDetails.json data: "
	LogJunctionAccFail      = "Can not get junction wallet address"
	LogPodVerifyTransact    = "Failed to Transact Verify pod"
	LogPodVerifyDone        = "pod verification transaction done"
	LogPodNotSubmitted      = "Pod is not submitted properly"
	LogPodVerifyFailed      = "Pod verification failed"
	LogMarshalPodVerified   = "Error in Marshaling PodVerifiedMsg"
	LogMarshalGossipMsg     = "Error marshaling gossip message"
)

type PodSubmittedMessageHandler struct {
	message PodSubmittedMsgData
}

type PodVerifiedMessageHandler struct {
	message PodVerifiedMsgData
}

type VRFInitiatedMessageHandler struct {
	message *VRFInitiatedMsgData
}
type VRNValidatedMessageHandler struct {
	message VRFVerifiedMsg
}

type AccountDetails struct {
	AccountPath   string
	AccountName   string
	AddressPrefix string
	Tracks        []string
	MyAddress     string
}

// ProcessGossipMessage takes in a node, data type, data byte slice, and a message broadcaster
// and processes the gossip message according to its data type.
// It checks if the data type is valid and if so, calls the corresponding message handler function.
// If the data type is unknown, it logs an error.
// Parameters:
// - node: The host node handling the gossip message
// - dataType: The type of gossip data being processed
// - dataByte: The byte slice representation of the gossip data
// - messageBroadcaster: The ID of the peer who broadcasted the message
func ProcessGossipMessage(node host.Host, dataType string, dataByte []byte, messageBroadcaster peer.ID) {
	messageHandlers := map[string]func([]byte){
		"vrfInitiated": func(dataByte []byte) {
			handler := NewVRFInitiatedMessageHandler(dataByte)
			if handler != nil {
				handler.HandleVRFInitiatedMessage()
			}
		},
		"vrnValidated": VRNValidatedMsgHandler,
		"podSubmitted": func(dataByte []byte) {
			handler := NewPodSubmittedMessageHandler(dataByte)
			if handler != nil {
				handler.HandlePodSubmissionMessage()
			}
		},
		"podVerified": func(dataByte []byte) {
			handler := NewPodVerifiedMessageHandler(dataByte)
			if handler != nil {
				handler.HandlePodMessage()
			}
		},
	}

	if handler, found := messageHandlers[dataType]; found {
		handler(dataByte)
	} else {
		logs.Log.Error("Unknown gossip data type found")
	}
}

// NewVRFInitiatedMessageHandler takes in a byte slice representing the VRFInitiated message,
// decodes it into a VRFInitiatedMsgData struct, and returns a pointer to a VRFInitiatedMessageHandler
// with the decoded message. If the decoding fails, it returns nil.
// Parameters:
// - dataByte: The byte slice representing the VRFInitiated message
// Returns:
// - *VRFInitiatedMessageHandler: A pointer to a VRFInitiatedMessageHandler with the decoded message

func NewVRFInitiatedMessageHandler(dataByte []byte) *VRFInitiatedMessageHandler {
	h := &VRFInitiatedMessageHandler{}
	h.message = decodeVRFInitiatedMsg(dataByte)
	if h.message == nil {
		return nil
	}
	return h
}

func (h *VRFInitiatedMessageHandler) HandleVRFInitiatedMessage() {
	waitUntilPodNumberMatched(h.message.PodNumber)

	accountDetails, err := getAccountDetails()
	if err != nil {
		logs.Log.Error(err.Error())
		return
	}

	// all nodes: update vrn init hash
	currentPodState := shared.GetPodState()
	VrfInitTxHash := h.message.VrfInitTxHash
	currentPodState.VRFInitiationTxHash = VrfInitTxHash
	shared.SetPodState(currentPodState)

	// selected node
	if h.message.SelectedTrackAddress == accountDetails.MyAddress {
		processVerifiedVRF(h.message, accountDetails)
	}
}

func decodeVRFInitiatedMsg(dataByte []byte) *VRFInitiatedMsgData {
	var VRFInitiatedMsg VRFInitiatedMsgData
	if err := json.Unmarshal(dataByte, &VRFInitiatedMsg); err != nil {
		logs.Log.Info("Error in extracting VRFInitiatedMsg")
		return nil
	}
	return &VRFInitiatedMsg
}

func waitUntilPodNumberMatched(podNumber uint64) {
	for {
		podState := shared.GetPodState()
		if podNumber == podState.LatestPodHeight {
			logs.Log.Info("Pod number matched")
			break
		}
		logs.Log.Warn("pod number is not matched, waiting for new pod")
		time.Sleep(3 * time.Second)
	}
}

func getAccountDetails() (*AccountDetails, error) {
	var ad AccountDetails

	var DetailsErr error
	_, _, ad.AccountPath, ad.AccountName, ad.AddressPrefix, ad.Tracks, DetailsErr = junction.GetJunctionDetails()

	if DetailsErr != nil {
		return nil, fmt.Errorf("can not get junctionDetails.json data: %s", DetailsErr)
	}

	ad.MyAddress, DetailsErr = junction.CheckIfAccountExists(ad.AccountName, ad.AccountPath, ad.AddressPrefix)
	if DetailsErr != nil {
		return nil, fmt.Errorf("Can not get junction wallet address")
	}

	return &ad, nil
}

func processVerifiedVRF(VRFInitiatedMsg *VRFInitiatedMsgData, ad *AccountDetails) {
	logs.Log.Info("This Track Address is selected to verify VRN")

	// verify
	VrfInitiatorAddress := VRFInitiatedMsg.VrfInitiatorAddress
	success := junction2.ValidateVRF(VrfInitiatorAddress)
	if !success {
		logs.Log.Error("Failed to Validate VRF")
		return
	}
	logs.Log.Info("validate vrf Transaction success")

	var vrfRecord *junctionTypes.VrfRecord
	vrfRecord = junction.QueryVRF()
	if vrfRecord == nil {
		logs.Log.Error("VRF record is nil")
		return
	}
	if !vrfRecord.IsVerified {
		logs.Log.Error("Verification of VRF is failed, need Voting for correct VRN")
		return
	}

	PodNumber := int(shared.GetPodState().LatestPodHeight)
	SelectedTrackAddress := ad.Tracks[vrfRecord.SelectedTrackIndex]
	VrnValidatedTxHash := shared.GetPodState().VRFValidationTxHash
	VRFVerifiedMsg := VRFVerifiedMsg{
		PodNumber:            uint64(PodNumber),
		SelectedTrackAddress: SelectedTrackAddress,
		VRFVerifiedTxHash:    VrnValidatedTxHash,
	}
	VRFVerifiedMsgByte, err := json.Marshal(VRFVerifiedMsg)
	if err != nil {
		logs.Log.Error("Error in Marshaling ProofVote Result")
		return
	}
	gossipMsg := types.GossipData{
		Type: "vrnValidated",
		Data: VRFVerifiedMsgByte,
	}
	gossipMsgByte, err := json.Marshal(gossipMsg)
	if err != nil {
		logs.Log.Error("Error marshaling gossip message")
		return
	}
	BroadcastMessage(CTX, Node, gossipMsgByte)

	if SelectedTrackAddress == ad.MyAddress {
		VRNValidatedMsgHandler(VRFVerifiedMsgByte)
	}
}

func VRNValidatedMsgHandler(dataByte []byte) {
	fmt.Println("VRN Validated Msg Handler called")
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	var VRNVerifiedMsg VRFVerifiedMsg
	if err := json.Unmarshal(dataByte, &VRNVerifiedMsg); err != nil {
		logs.Log.Error("Error in extracting VRFVerifiedMsg")
		return
	}

	for {
		podState := shared.GetPodState()
		if VRNVerifiedMsg.PodNumber == podState.LatestPodHeight {
			break
		}
		logs.Log.Warn("pod number is not matched, waiting for new pod")
		fmt.Println(VRNVerifiedMsg.PodNumber, podState.LatestPodHeight)
		time.Sleep(3 * time.Second)
	}

	// all nodes: update txHash of vrn validated
	VRFValidationTxHash := VRNVerifiedMsg.VRFVerifiedTxHash
	currentPodState := shared.GetPodState()
	currentPodState.VRFValidationTxHash = VRFValidationTxHash
	shared.SetPodState(currentPodState)

	// check if this node is selected to submit pod & da
	_, _, accountPath, accountName, addressPrefix, tracks, err := junction.GetJunctionDetails()
	if err != nil {
		logs.Log.Error("can not get junctionDetails.json data: " + err.Error())
		return
	}
	myAddress, err := junction.CheckIfAccountExists(accountName, accountPath, addressPrefix)
	if err != nil {
		logs.Log.Error("Can not get junction wallet address")
		return
	}
	// now check for this pod number, who is the selected track
	if VRNVerifiedMsg.SelectedTrackAddress == myAddress {
		// submit data to DA
		DaData := shared.GetPodState().Batch.TransactionHash
		daDataByte := []byte{}
		for _, str := range DaData {
			daDataByte = append(daDataByte, []byte(str)...)
		}

		PodNumber := int(shared.GetPodState().LatestPodHeight)
		connection := shared.Node.NodeConnections
		DaBatchSaver := connection.DataAvailabilityDatabaseConnection

		baseConfig, err := shared.LoadConfig()
		if err != nil {
			fmt.Println("Error loading configuration")
		}
		Datype := baseConfig.DA.DaType
		if Datype == "mock" {
			mdb := connection.MockDatabaseConnection
			daCheck, daCheckErr := mock.MockDA(mdb, daDataByte, PodNumber)
			if daCheckErr != nil {
				logs.Log.Error("Error in submitting data to DA")
				return
			}

			da := types.DAStruct{
				DAKey:             daCheck,
				DAClientName:      "mock-da",
				BatchNumber:       strconv.Itoa(PodNumber),
				PreviousStateHash: string(shared.GetPodState().PreviousPodHash),
				CurrentStateHash:  string(shared.GetPodState().TracksAppHash),
			}

			daStoreKey := fmt.Sprintf("da-%d", PodNumber)
			daStoreData, daStoreDataErr := json.Marshal(da)
			if daStoreDataErr != nil {
				logs.Log.Warn(fmt.Sprintf("Error in marshaling DA pointer : %s", daStoreDataErr.Error()))
			}

			storeErr := DaBatchSaver.Put([]byte(daStoreKey), daStoreData, nil)
			if storeErr != nil {
				logs.Log.Warn(fmt.Sprintf("Error in saving DA pointer in pod database : %s", storeErr.Error()))
			}

			log.Info().Str("module", "p2p").Msg("Data Saved in DA")

		} else if Datype == "avail" {
			daCheck, daCheckErr := avail.Avail(daDataByte, baseConfig.DA.DaRPC)
			if daCheckErr != nil {
				logs.Log.Warn("Error in submitting data to DA")
				return
			}

			da := types.DAStruct{
				DAKey:             daCheck,
				DAClientName:      "avail-da",
				BatchNumber:       strconv.Itoa(PodNumber),
				PreviousStateHash: string(shared.GetPodState().PreviousPodHash),
				CurrentStateHash:  string(shared.GetPodState().TracksAppHash),
			}

			daStoreKey := fmt.Sprintf("da-%d", PodNumber)
			daStoreData, daStoreDataErr := json.Marshal(da)
			if daStoreDataErr != nil {
				logs.Log.Warn(fmt.Sprintf("Error in marshaling DA pointer : %s", daStoreDataErr.Error()))
			}

			storeErr := DaBatchSaver.Put([]byte(daStoreKey), daStoreData, nil)
			if storeErr != nil {
				logs.Log.Warn(fmt.Sprintf("Error in saving DA pointer in pod database : %s", storeErr.Error()))
			}

			log.Info().Str("module", "p2p").Msg("Data Saved in DA")

		} else if Datype == "celestia" {
			daCheck, daCheckErr := celestia.Celestia(daDataByte, baseConfig.DA.DaRPC, baseConfig.DA.DaRPC)

			if daCheckErr != nil {
				logs.Log.Warn("Error in submitting data to DA")
				return
			}

			da := types.DAStruct{
				DAKey:             daCheck,
				DAClientName:      "celestia-da",
				BatchNumber:       strconv.Itoa(PodNumber),
				PreviousStateHash: string(shared.GetPodState().PreviousPodHash),
				CurrentStateHash:  string(shared.GetPodState().TracksAppHash),
			}

			daStoreKey := fmt.Sprintf("da-%d", PodNumber)
			daStoreData, daStoreDataErr := json.Marshal(da)
			if daStoreDataErr != nil {
				logs.Log.Warn(fmt.Sprintf("Error in marshaling DA pointer : %s", daStoreDataErr.Error()))
			}

			storeErr := DaBatchSaver.Put([]byte(daStoreKey), daStoreData, nil)
			if storeErr != nil {
				logs.Log.Warn(fmt.Sprintf("Error in saving DA pointer in pod database : %s", storeErr.Error()))
			}

			log.Info().Str("module", "p2p").Msg("Data Saved in DA")

		} else if Datype == "eigen" {
			daCheck, daCheckErr := eigen.Eigen(daDataByte,
				baseConfig.DA.DaRPC, baseConfig.DA.DaRPC,
			)

			if daCheckErr != nil {
				logs.Log.Warn("Error in submitting data to DA")
				return
			}

			da := types.DAStruct{
				DAKey:             daCheck,
				DAClientName:      "eigen-da",
				BatchNumber:       strconv.Itoa(PodNumber),
				PreviousStateHash: string(shared.GetPodState().PreviousPodHash),
				CurrentStateHash:  string(shared.GetPodState().TracksAppHash),
			}

			daStoreKey := fmt.Sprintf("da-%d", PodNumber)
			daStoreData, daStoreDataErr := json.Marshal(da)
			if daStoreDataErr != nil {
				logs.Log.Warn(fmt.Sprintf("Error in marshaling DA pointer : %s", daStoreDataErr.Error()))
			}

			storeErr := DaBatchSaver.Put([]byte(daStoreKey), daStoreData, nil)
			if storeErr != nil {
				logs.Log.Warn(fmt.Sprintf("Error in saving DA pointer in pod database : %s", storeErr.Error()))
			}

			log.Info().Str("module", "p2p").Msg("Data Saved in DA")

		} else {
			logs.Log.Error("Unknown layer. Please use 'avail' or 'celestia' as argument.")
			return
		}

		// submit pod to junction
		success := junction2.SubmitCurrentPod()
		if !success {
			logs.Log.Error("Failed to submit pod")
			return
		}

		var filteredTracks []string
		for _, track := range tracks {
			if track != myAddress {
				filteredTracks = append(filteredTracks, track)
			}
		}
		// Select a random peer from the filtered list
		SelectedTrackAddress := filteredTracks[rand.Intn(len(filteredTracks))]

		podNumber := shared.GetPodState().LatestPodHeight
		InitPodTxHash := shared.GetPodState().InitPodTxHash
		// broadcast pod submitted msg
		PodSubmittedMsg := PodSubmittedMsgData{
			PodNumber:            podNumber,
			SelectedTrackAddress: SelectedTrackAddress,
			InitPodTxHash:        InitPodTxHash,
		}
		PodSubmittedMsgByte, err := json.Marshal(PodSubmittedMsg)
		if err != nil {
			logs.Log.Error("Error in Marshaling PodSubmittedMsg")
			return
		}
		gossipMsg := types.GossipData{
			Type: "podSubmitted",
			Data: PodSubmittedMsgByte,
		}
		gossipMsgByte, err := json.Marshal(gossipMsg)
		if err != nil {
			logs.Log.Error("Error marshaling gossip message")
			return
		}
		BroadcastMessage(CTX, Node, gossipMsgByte)
	}
	return
}

// NewPodSubmittedMessageHandler takes in a byte slice representing the PodSubmitted message,
// decodes it into a PodSubmittedMsgData struct, and returns a pointer to a PodSubmittedMessageHandler
// with the decoded message. If the decoding fails, it returns nil.
func NewPodSubmittedMessageHandler(dataByte []byte) *PodSubmittedMessageHandler {
	h := &PodSubmittedMessageHandler{}
	if err := json.Unmarshal(dataByte, &h.message); err != nil {
		logs.Log.Error(LogPodSubmitExtractFail)
		return nil
	}
	return h
}

func (h *PodSubmittedMessageHandler) HandlePodSubmissionMessage() {
	for {
		podState := shared.GetPodState()
		if h.message.PodNumber == podState.LatestPodHeight {
			break
		}
		h.logAndSleep(LogPodSubmitMatchFail, h.message.PodNumber, podState.LatestPodHeight)
	}
	h.processPodSubmission()
}

func (h *PodSubmittedMessageHandler) logAndSleep(message string, number, height uint64) {
	logs.Log.Warn(message)
	fmt.Println(number, height)
	time.Sleep(SleepDuration)
}

func (h *PodSubmittedMessageHandler) processPodSubmission() {
	_, _, accountPath, accountName, addressPrefix, _, err := junction.GetJunctionDetails()
	if err != nil {
		logs.Log.Error(LogJunctionDetailsFail + err.Error())
		return
	}
	myAddress, err := junction.CheckIfAccountExists(accountName, accountPath, addressPrefix)
	if err != nil {
		logs.Log.Error(LogJunctionAccFail)
		return
	}

	// all nodes: update initPodTxHash
	currentPodState := shared.GetPodState()
	InitPodTxHash := h.message.InitPodTxHash
	currentPodState.InitPodTxHash = InitPodTxHash
	shared.SetPodState(currentPodState)

	if h.message.SelectedTrackAddress == myAddress {
		h.verifyAndBroadcastPod()
	}
}

func (h *PodSubmittedMessageHandler) verifyAndBroadcastPod() {
	success := junction2.VerifyCurrentPod()
	if !success {
		logs.Log.Error(LogPodVerifyTransact)
		return
	}
	logs.Log.Info(LogPodVerifyDone)
	podDetails := junction.QueryPod(h.message.PodNumber)
	if podDetails == nil {
		logs.Log.Error(LogPodNotSubmitted)
		return
	}
	if podDetails.IsVerified == false {
		logs.Log.Error(LogPodVerifyFailed)
		return
	}
	h.broadcastPodVerifiedMessage()
}

func (h *PodSubmittedMessageHandler) broadcastPodVerifiedMessage() {
	podNumber := shared.GetPodState().LatestPodHeight
	VerifyPodTxHash := shared.GetPodState().VerifyPodTxHash
	PodVerifiedMsg := PodVerifiedMsgData{
		PodNumber:          podNumber,
		VerificationResult: true,
		PodVerifiedTxHash:  VerifyPodTxHash,
	}
	PodVerifiedMsgByte, err := json.Marshal(PodVerifiedMsg)
	if err != nil {
		logs.Log.Error(LogMarshalPodVerified)
		return
	}
	gossipMsg := types.GossipData{
		Type: "podVerified",
		Data: PodVerifiedMsgByte,
	}
	gossipMsgByte, err := json.Marshal(gossipMsg)
	if err != nil {
		logs.Log.Error(LogMarshalGossipMsg)
		return
	}
	saveVerifiedPOD()
	BroadcastMessage(CTX, Node, gossipMsgByte)
	GenerateUnverifiedPods()
}

// NewPodVerifiedMessageHandler takes in a byte slice and returns a new instance of PodVerifiedMessageHandler
// with the message field populated.
// It attempts to decode the byte slice into the message field using JSON unmarshalling.
// If there is an error during decoding, it logs an error and returns nil.
// If decoding is successful, it logs a warning indicating the successful decoding and returns the handler.

func NewPodVerifiedMessageHandler(dataByte []byte) *PodVerifiedMessageHandler {
	h := &PodVerifiedMessageHandler{}
	if err := json.Unmarshal(dataByte, &h.message); err != nil {
		logs.Log.Error(LogPodExtractFail)
		return nil

	}
	logs.Log.Warn(LogPodDecodeSuccess)
	return h
}

func (h *PodVerifiedMessageHandler) HandlePodMessage() {
	// match the pod number
	for {
		podState := shared.GetPodState()
		if h.message.PodNumber == podState.LatestPodHeight {
			break
		}
		h.logAndSleep(LogPodMismatch, h.message.PodNumber, podState.LatestPodHeight)
	}

	// update pod verified
	podState := shared.GetPodState()
	VerifyPodTxHash := shared.GetPodState().VerifyPodTxHash
	podState.VerifyPodTxHash = VerifyPodTxHash
	shared.SetPodState(podState)

	logs.Log.Warn(LogPodMatchSuccess)
	h.handleVerificationResult()
}

func (h *PodVerifiedMessageHandler) logAndSleep(message string, number, height uint64) {
	logs.Log.Warn(message)
	fmt.Println(number, height)
	time.Sleep(SleepDuration)
}

func (h *PodVerifiedMessageHandler) handleVerificationResult() {
	// update verified hash in all nodes

	if h.message.VerificationResult {
		logs.Log.Info(LogPodSave)
		saveVerifiedPOD() // save the latest pod details and make next pod
		logs.Log.Info(LogPodGenNext)
		GenerateUnverifiedPods() // generate next pod
	} else {
		logs.Log.Error(LogPodFail)
	}
}
