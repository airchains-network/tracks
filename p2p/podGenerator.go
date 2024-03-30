package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/config"
	mock "github.com/airchains-network/decentralized-sequencer/da/mockda"
	"github.com/airchains-network/decentralized-sequencer/junction"
	junctionTypes "github.com/airchains-network/decentralized-sequencer/junction/types"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/node/shared"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/airchains-network/decentralized-sequencer/utilis"
	v1 "github.com/airchains-network/decentralized-sequencer/zk/v1"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/syndtr/goleveldb/leveldb"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	StationRPC string
)

func BatchGeneration(wg *sync.WaitGroup) {
	defer wg.Done()

	GenerateUnverifiedPods()
}

func GenerateUnverifiedPods() {
	fmt.Println("generating new pod")

	lds := shared.Node.NodeConnections.GetStaticDatabaseConnection()
	ldt := shared.Node.NodeConnections.GetTxnDatabaseConnection()
	fmt.Println(lds)

	fmt.Println("1")
	ConfirmendTransactionIndex, err := lds.Get([]byte("batchStartIndex"), nil)
	if err != nil {
		logs.Log.Warn("ConfirmendTransactionIndex not found in static db")
		err = lds.Put([]byte("batchStartIndex"), []byte("0"), nil)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in saving batchStartIndex in static db : %s", err.Error()))
			os.Exit(0)
		}
	}
	fmt.Println("2")

	currentPodNumber, err := lds.Get([]byte("batchCount"), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in getting sssssss from static db : %s", err.Error()))
		os.Exit(0)
	}
	fmt.Println("3")

	previousStateData, err := getPodStateFromDatabase()
	if err != nil {
		logs.Log.Error("Error in getting previous station data")
		os.Exit(0)
	}
	//fmt.Println("get this data from database:", previousStateData)
	PreviousTrackAppHash := previousStateData.TracksAppHash
	if PreviousTrackAppHash == nil {
		PreviousTrackAppHash = []byte("nil")
	}

	SelectedMaster := MasterTracksSelection(Node, string(PreviousTrackAppHash))
	decodedMaster, err := peer.Decode(SelectedMaster)

	currentPodNumberInt, _ := strconv.Atoi(strings.TrimSpace(string(currentPodNumber)))
	batchNumber := currentPodNumberInt + 1
	fmt.Println("generating new pod")
	fmt.Println(batchNumber)

	//var batchInput *types.BatchStruct
	Witness, uZKP, MRH, batchInput, err := createPOD(ldt, ConfirmendTransactionIndex, currentPodNumber)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in creating POD : %s", err.Error()))
		os.Exit(0)
	}

	TrackAppHash := generatePodHash(Witness, uZKP, MRH, currentPodNumber)
	//podState := shared.GetPodState()
	//tempMasterTrackAppHash := podState.MasterTrackAppHash

	// update pod state as per latest pod
	updateNewPodState(TrackAppHash, Witness, uZKP, MRH, uint64(batchNumber), batchInput)

	// Here the MasterTrack Will Broadcast the uZKP in the Network
	if decodedMaster == Node.ID() {
		fmt.Println("I am Master")
		// make master's vote true by default
		podState := shared.GetPodState()
		currentVotes := podState.Votes
		currentVotes[decodedMaster.String()] = shared.Votes{
			PeerID: decodedMaster.String(),
			Vote:   true,
		}
		podState.Votes = currentVotes
		shared.SetPodState(podState)
		Peers := getAllPeers(Node)
		peerCount := len(Peers)
		if peerCount == 1 {
			// no peers connected so submit & verify VRF & Pod by own and update local database too.
			DaData := shared.GetPodState().Batch.TransactionHash
			daDataByte := []byte{}
			for _, str := range DaData {
				daDataByte = append(daDataByte, []byte(str)...)
			}
			ZkProof := shared.GetPodState().LatestPodProof
			PodNumber := int(shared.GetPodState().LatestPodHeight)

			finalizeDA := types.FinalizeDA{
				CompressedHash: DaData,
				Proof:          ZkProof,
				PodNumber:      PodNumber,
			}
			_, err := json.Marshal(finalizeDA)
			if err != nil {
				logs.Log.Error("Error marshaling da data: " + err.Error())
				return
			}

			success, addr := junction.InitVRF()
			if !success {
				logs.Log.Error("Failed to Init VRF")
				return
			}
			logs.Log.Info("VRF initiated")

			success = junction.ValidateVRF(addr)
			if !success {
				logs.Log.Error("Failed to Validate VRF")
				return
			}
			logs.Log.Info("validate vrf success")

			// check if VRF is successfully validated
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

			// DA submit
			connection := shared.Node.NodeConnections
			mdb := connection.GetDataAvailabilityDatabaseConnection()
			dbName, err := mock.MockDA(mdb, daDataByte, PodNumber) // (mockda-%d", batchNumber), nil
			if err != nil {
				logs.Log.Error("Error in submitting data to DA")
				return
			}
			_ = dbName
			logs.Log.Info("data in DA submitted")

			// submit pod
			success = junction.SubmitCurrentPod()
			if !success {
				logs.Log.Error("Failed to submit pod")
				return
			}
			logs.Log.Info("pod submitted")

			// verify pod
			success = junction.VerifyCurrentPod()
			if !success {
				logs.Log.Error("Failed to Transact Verify pod")
				return
			}
			logs.Log.Info("pod verification transaction done")
			// todo : query if verification return true or false...

			// if (no peers connected): update database and make next pod without voting process
			saveVerifiedPOD() // save data to database
			//
			//peerListLocked = false
			//peerListLock.Unlock()
			//peerListLock.Lock()
			//for _, peerInfo := range incomingPeers.GetPeers() {
			//	peerList.AddPeer(peerInfo)
			//}
			//peerListLock.Unlock()

			GenerateUnverifiedPods() // generate next pod

		} else {
			// call VRF -> if success: broadcast peerId of node who will verify VRF [not randomly]

			PodNumber := int(shared.GetPodState().LatestPodHeight)

			success, addr := junction.InitVRF()
			if !success {
				logs.Log.Error("Failed to Init VRF")
				return
			}
			logs.Log.Info("VRF initiated")

			// get own address
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

			// choose one verifiable random node to verify the VRF
			// Filter out the peer with own Id
			var filteredTracks []string
			for _, track := range tracks {
				if track != myAddress {
					filteredTracks = append(filteredTracks, track)
				}
			}
			// Select a random peer from the filtered list
			selectedTrackAddress := filteredTracks[rand.Intn(len(filteredTracks))]
			fmt.Println("Selected random address:", selectedTrackAddress)

			// send verify VRF message to selected node
			VRFInitiatedMsg := VRFInitiatedMsg{
				PodNumber:            uint64(PodNumber),
				SelectedTrackAddress: selectedTrackAddress,
				VrfInitiatorAddress:  addr,
			}

			VRFInitiatedMsgByte, err := json.Marshal(VRFInitiatedMsg)
			if err != nil {
				logs.Log.Error("Error in Marshaling ProofVote Result")
				return
			}
			gossipMsg := types.GossipData{
				Type: "vrfInitiated",
				Data: VRFInitiatedMsgByte,
			}
			gossipMsgByte, err := json.Marshal(gossipMsg)
			if err != nil {
				logs.Log.Error("Error marshaling gossip message")
				return
			}
			BroadcastMessage(CTX, Node, gossipMsgByte)

			// after VRF is verified: broadcast pod Submitter id/address to all.
			// pod submitter will then broadcast the message that pod is submitted, and choose a node [not randomly] to verify the pod
			// after pod is verify, the node will broadcast the message that pod is verified, and all nodes will update the database and generate next pod

			// vrfHandlerGossip -> all nodes will got who is selected node
			//// Preparing the Message that master track will gossip to the Network
			//proofData := ProofData{
			//	PodNumber:    uint64(batchNumber),
			//	TrackAppHash: TrackAppHash,
			//}
			//
			//// Marshal the proofData
			//proofDataByte, err := json.Marshal(proofData)
			//if err != nil {
			//	logs.Log.Error(fmt.Sprintf("Error in marshalling proof data : %s", err.Error()))
			//}
			//
			//gossipMsg := types.GossipData{
			//	Type: "proof",
			//	Data: proofDataByte,
			//}
			//
			//gossipMsgByte, err := json.Marshal(gossipMsg)
			//if err != nil {
			//	logs.Log.Error("Error marshaling gossip message")
			//	return
			//}
			//
			//logs.Log.Info("Sending proof result: %s")
			//BroadcastMessage(context.Background(), Node, gossipMsgByte)
		}

	}

	//else {
	//	if podState.MasterTrackAppHash != nil {
	//		fmt.Println(TrackAppHash)
	//		fmt.Println(tempMasterTrackAppHash)
	//		currentPodData := shared.GetPodState()
	//		if bytes.Equal(TrackAppHash, tempMasterTrackAppHash) {
	//			SendValidProof(CTX, currentPodData.LatestPodHeight, decodedMaster)
	//			return
	//		} else {
	//			SendInvalidProofError(CTX, currentPodData.LatestPodHeight, decodedMaster)
	//			return
	//		}
	//	} else {
	//		// pod state is nil, means master track has not yet broadcasted the proof
	//		// don't need to do anything..
	//	}
	//}

}

func createPOD(ldt *leveldb.DB, batchStartIndex []byte, limit []byte) (witness []byte, unverifiedProof []byte, MRH []byte, podData *types.BatchStruct, err error) {
	baseConfig, err := shared.LoadConfig()
	if err != nil {
		return
	}
	limitInt, _ := strconv.Atoi(strings.TrimSpace(string(limit)))

	batchStartIndexInt, _ := strconv.Atoi(strings.TrimSpace(string(batchStartIndex)))

	fmt.Println(limitInt)
	fmt.Println(batchStartIndexInt)

	var batch types.BatchStruct

	var From []string
	var To []string
	var Amounts []string
	var TransactionHash []string
	var SenderBalances []string
	var ReceiverBalances []string
	var Messages []string
	var TransactionNonces []string
	var AccountNonces []string

	for i := batchStartIndexInt; i < (config.PODSize * (limitInt + 1)); i++ {

		findKey := fmt.Sprintf("txns-%d", i+1)
		txData, err := ldt.Get([]byte(findKey), nil)
		if err != nil {
			i--
			time.Sleep(1 * time.Second)
			continue
		}
		var tx types.TransactionStruct
		err = json.Unmarshal(txData, &tx)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in unmarshalling tx data : %s", err.Error()))
			os.Exit(0)
		}

		senderBalancesCheck, err := utilis.GetBalance(tx.From, tx.BlockNumber-1, baseConfig.Station.StationRPC)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting sender balance : %s", err.Error()))
			os.Exit(0)
		}

		receiverBalancesCheck, err := utilis.GetBalance(tx.To, tx.BlockNumber-1, baseConfig.Station.StationRPC)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting reciver balance : %s", err.Error()))
			os.Exit(0)
		}

		accountNouceCheck, err := utilis.GetAccountNonce(context.Background(), tx.Hash, tx.BlockNumber)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting account nonce : %s", err.Error()))
			os.Exit(0)
		}

		From = append(From, tx.From)
		To = append(To, tx.To)
		Amounts = append(Amounts, tx.Value)
		TransactionHash = append(TransactionHash, tx.Hash)
		SenderBalances = append(SenderBalances, senderBalancesCheck)
		ReceiverBalances = append(ReceiverBalances, receiverBalancesCheck)
		Messages = append(Messages, tx.Input)
		TransactionNonces = append(TransactionNonces, tx.Nonce)
		AccountNonces = append(AccountNonces, accountNouceCheck)
	}

	batch.From = From
	batch.To = To
	batch.Amounts = Amounts
	batch.TransactionHash = TransactionHash
	batch.SenderBalances = SenderBalances
	batch.ReceiverBalances = ReceiverBalances
	batch.Messages = Messages
	batch.TransactionNonces = TransactionNonces
	batch.AccountNonces = AccountNonces
	fmt.Println("batch data", batch)
	witnessVector, currentStatusHash, proofByte, pkErr := v1.GenerateProof(batch, limitInt+1)
	fmt.Println("Witness Vector: ", currentStatusHash)
	if pkErr != nil {
		logs.Log.Error(fmt.Sprintf("Error in generating proof : %s", pkErr.Error()))
		return nil, nil, nil, nil, pkErr
	}
	logs.Log.Warn(fmt.Sprintf("Successfully generated  Unverified proof for Batch %s in the latest phase", strconv.Itoa(limitInt+1)))

	// marshal witnessVector
	witnessVectorByte, err := json.Marshal(witnessVector)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling witness vector : %s", err.Error()))
	}

	// string to []byte currentStatusHash
	currentStatusHashByte, err := json.Marshal(currentStatusHash)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling current status hash : %s", err.Error()))
		os.Exit(0)
	}

	return witnessVectorByte, proofByte, currentStatusHashByte, &batch, nil
}

func saveVerifiedPOD() {

	podState := shared.GetPodState()
	batchInput := podState.Batch
	currentPodNumber := podState.LatestPodHeight
	currentPodNumberInt := int(currentPodNumber)

	batchByte, err := json.Marshal(batchInput)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling batch data : %s", err.Error()))
		os.Exit(0)
	}

	podByte, err := json.Marshal(podState)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling pod data : %s", err.Error()))
		os.Exit(0)
	}

	ldbatch := shared.Node.NodeConnections.GetDataAvailabilityDatabaseConnection()
	lds := shared.Node.NodeConnections.GetStaticDatabaseConnection()

	batchKey := fmt.Sprintf("batch-%d", currentPodNumberInt)
	err = ldbatch.Put([]byte(batchKey), batchByte, nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in writing batch data to file : %s", err.Error()))
		os.Exit(0)
	}
	err = lds.Put([]byte("batchStartIndex"), []byte(strconv.Itoa(config.PODSize*(currentPodNumberInt))), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in updating batchStartIndex in static db : %s", err.Error()))
		os.Exit(0)
	}

	err = lds.Put([]byte("batchCount"), []byte(strconv.Itoa(currentPodNumberInt)), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in updating batchCount in static db : %s", err.Error()))
		os.Exit(0)
	}

	batchDB := shared.Node.NodeConnections.GetPodsDatabaseConnection()
	podKey := fmt.Sprintf("pod-%d", currentPodNumberInt)
	err = batchDB.Put([]byte(podKey), podByte, nil)
	if err != nil {
		panic("Failed to update pod data: " + err.Error())
	}

	err = os.WriteFile("data/batchCount.txt", []byte(strconv.Itoa(currentPodNumberInt)), 0666)
	if err != nil {
		panic("Failed to update batch number: " + err.Error())
	}

	podState.MasterTrackAppHash = nil
	shared.SetPodState(podState)
	logs.Log.Warn("presend pod data saved to database")
}

func generatePodHash(Witness, uZKP, MRH []byte, podNumber []byte) []byte {
	return MRH
}

func updateNewPodState(CombinedPodHash, Witness, uZKP, MRH []byte, podNumber uint64, batchInput *types.BatchStruct) {
	var podState *shared.PodState
	// empty votes
	votes := make(map[string]shared.Votes)

	podState = &shared.PodState{
		LatestPodHeight:     podNumber,
		LatestPodHash:       MRH,
		PreviousPodHash:     shared.GetPodState().LatestPodHash,
		LatestPodProof:      uZKP,
		LatestPublicWitness: Witness,
		Votes:               votes,
		TracksAppHash:       CombinedPodHash,
		Batch:               batchInput,
	}

	// save pod state to database
	shared.SetPodState(podState)

	// save pod data in local state
	updatePodStateInDatabase(podState)

}

func updatePodStateInDatabase(podState *shared.PodState) {
	stateConnection := shared.Node.NodeConnections.GetStateDatabaseConnection()

	podStateByte, err := json.Marshal(podState)
	if err != nil {
		logs.Log.Error(err.Error())
		os.Exit(0)
	}

	err = stateConnection.Put([]byte("podState"), podStateByte, nil)
	if err != nil {
		logs.Log.Error(err.Error())
		os.Exit(0)
	}
}

func getPodStateFromDatabase() (*types.PodState, error) {
	var podStateData *types.PodState
	stateConnection := shared.Node.NodeConnections.GetStateDatabaseConnection()

	podStateDataByte, err := stateConnection.Get([]byte("podState"), nil)
	if err != nil {
		logs.Log.Error("error in getting pod state data from database")
		return nil, err
	}
	err = json.Unmarshal(podStateDataByte, &podStateData)
	if err != nil {
		logs.Log.Error("error in unmarshal pod state data")
		return nil, err
	}

	return podStateData, nil

}
