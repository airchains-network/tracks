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
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func BatchGeneration(wg *sync.WaitGroup, sequencerType string) {
	defer wg.Done()
	if sequencerType == "espresso" {

		//// config: sequencer version,
		//storedVersion := config.Sequencer.Version
		//CurrentEspressoVersion := "v4.0.0"
		//
		//if storedVersion != CurrentEspressoVersion {
		//	// new schema -> schema update
		//	// .toml update
		//}
		TrackgatePodGenerator()
	} else {
		GenerateUnverifiedPods()
	}
}

func GenerateUnverifiedPods() {
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().
		Str("module", "p2p").
		Msg("Generating New unverified pods")

	connection := shared.Node.NodeConnections
	staticDBConnection := connection.GetStaticDatabaseConnection()
	txnDBConnection := connection.GetTxnDatabaseConnection()

	rawConfirmedTransactionIndex, err := GetValueOrDefault(staticDBConnection, []byte(BatchStartIndexKey), []byte("0"))
	CheckErrorAndExit(err, "Error in getting confirmedTransactionIndex from static db", 0)

	rawCurrentPodNumber, err := GetValueOrDefault(staticDBConnection, []byte(BatchCountKey), []byte("0"))
	CheckErrorAndExit(err, "Error in getting currentPodNumber from static db", 0)

	//previousStateData
	podStateData, err := GetPodStateFromDatabase()
	CheckErrorAndExit(err, "Error in getting previous station data", 0)

	var (
		previousTrackAppHash []byte
		trackAppHash         []byte
		witness              []byte
		uZKP                 []byte
		MRH                  []byte
		batchInput           *types.BatchStruct
		txState              string
		batchNumber          int
	)

	currentPodNumber, _ := strconv.Atoi(strings.TrimSpace(string(rawCurrentPodNumber)))
	txState = podStateData.LatestTxState
	if txState == "" {
		txState = shared.TxStatePreInit
	}

	if currentPodNumber == 0 {
		currentPodNumber = 1
	}

	podData := junction.QueryPod(uint64(currentPodNumber))
	if podData != nil {
		if podData.IsVerified == true {
			currentPodNumber++
		}
	}

	batchNumber = currentPodNumber
	log.Info().Str("module", "p2p").Msg(fmt.Sprintf("Processing Pod Number: %d", batchNumber))

	if podStateData.LatestTxState == shared.TxStatePreInit {
		txState = shared.TxStateInitVRF
		previousTrackAppHash = podStateData.TracksAppHash
		if previousTrackAppHash == nil {
			previousTrackAppHash = []byte("nil")
		}

		baseCfg, err := shared.LoadConfig()
		if err != nil {
			log.Error().Str("module", "p2p").Msg("Error in loading config")
		}
		stationVariant := baseCfg.Station.StationType
		stationVariantLowerCase := strings.ToLower(stationVariant)

		if stationVariantLowerCase == "evm" {
			witness, uZKP, MRH, batchInput, err = createEVMPOD(txnDBConnection, rawConfirmedTransactionIndex, rawCurrentPodNumber)
			CheckErrorAndExit(err, "Error in creating POD", 0)
		} else if stationVariantLowerCase == "wasm" {
			witness, uZKP, MRH, batchInput, err = createWasmPOD(txnDBConnection, rawConfirmedTransactionIndex, rawCurrentPodNumber)
			CheckErrorAndExit(err, "Error in creating POD", 0)
		}

		trackAppHash = generatePodHash(witness, uZKP, MRH, rawCurrentPodNumber)
		updateNewPodState(trackAppHash, witness, uZKP, MRH, uint64(batchNumber), batchInput, txState)
	} else {
		trackAppHash = podStateData.TracksAppHash
		witness = podStateData.LatestPublicWitness
		uZKP = podStateData.LatestPodProof
		MRH = podStateData.LatestPodHash
		pMRH := podStateData.PreviousPodHash
		batchInput = podStateData.Batch

		storeNewPodState(trackAppHash, witness, uZKP, pMRH, MRH, uint64(batchNumber), batchInput, txState)
	}

	selectedMaster := MasterTracksSelection(Node, string(previousTrackAppHash))
	decodedMaster, err := peer.Decode(selectedMaster)
	CheckErrorAndExit(err, "Error in decoding master", 0)

	if decodedMaster == Node.ID() {
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
			DaData := shared.GetPodState().Batch.TransactionHash
			var daDataByte []byte
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

			addr, err := junction.GetAddress()
			if err != nil {
				logs.Log.Error("Error in getting address")
				return
			}

			if shared.GetPodState().LatestTxState == shared.TxStateInitVRF {
				success, _ := junction2.InitVRF()
				if !success {
					logs.Log.Error("Failed to Init VRF")
					return
				}
				updateTxState(shared.TxStateVerifyVRF)
			} else {
				log.Debug().Str("module", "p2p").Msg("VRF is already initiated, moving to next step")
			}

			//os.Exit(0)

			if shared.GetPodState().LatestTxState == shared.TxStateVerifyVRF {
				success := junction2.ValidateVRF(addr)
				if !success {
					logs.Log.Error("Failed to Validate VRF")
					return
				}
				updateTxState(shared.TxStateSubmitPod)

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
			} else {
				log.Debug().Str("module", "p2p").Msg("VRF is already validated, moving to next step")
			}
			//os.Exit(0)

			if shared.GetPodState().LatestTxState == shared.TxStateSubmitPod {

				DaBatchSaver := connection.DataAvailabilityDatabaseConnection
				baseConfig, err := shared.LoadConfig()
				if err != nil {
					fmt.Println("Error loading configuration")
				}
				Datype := baseConfig.DA.DaType
				var (
					daCheck    string
					daCheckErr error
				)

				if Datype == "mock" {
					mdb := connection.MockDatabaseConnection

					for {
						daCheck, daCheckErr = mock.MockDA(mdb, daDataByte, PodNumber)
						if daCheckErr != nil {
							logs.Log.Error("Error in submitting data to DA")
							logs.Log.Debug("Retrying Mock DA after 10 seconds")
							time.Sleep(10 * time.Second)
						} else {
							break
						}
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
						logs.Log.Debug(fmt.Sprintf("Error in marshaling DA pointer : %s", daStoreDataErr.Error()))
					}

					storeErr := DaBatchSaver.Put([]byte(daStoreKey), daStoreData, nil)
					if storeErr != nil {
						logs.Log.Debug(fmt.Sprintf("Error in saving DA pointer in pod database : %s", storeErr.Error()))
					}

					log.Info().Str("module", "p2p").Msg("Data Saved in DA")

				} else if Datype == "avail" {

					for {
						daCheck, daCheckErr = avail.Avail(daDataByte, baseConfig.DA.DaRPC)
						if daCheckErr != nil {
							logs.Log.Debug("Error in submitting data to DA " + daCheckErr.Error())
							logs.Log.Debug("Retrying Avail DA after 10 seconds")
							time.Sleep(10 * time.Second)
						} else {
							break
						}
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
						logs.Log.Debug(fmt.Sprintf("Error in marshaling DA pointer : %s", daStoreDataErr.Error()))
						return
					}

					storeErr := DaBatchSaver.Put([]byte(daStoreKey), daStoreData, nil)
					if storeErr != nil {
						logs.Log.Debug(fmt.Sprintf("Error in saving DA pointer in pod database : %s", storeErr.Error()))
						return
					}

					log.Info().Str("module", "p2p").Msg("Data Saved in DA")

				} else if Datype == "celestia" {
					for {
						daCheck, daCheckErr = celestia.Celestia(daDataByte, baseConfig.DA.DaRPC, baseConfig.DA.DaRPC)
						if daCheckErr != nil {
							logs.Log.Debug("Error in submitting data to DA " + daCheckErr.Error())
							logs.Log.Debug("Retrying Celestia DA after 10 seconds")
							time.Sleep(10 * time.Second)
						} else {
							break
						}
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
						logs.Log.Debug(fmt.Sprintf("Error in marshaling DA pointer : %s", daStoreDataErr.Error()))
						return
					}

					storeErr := DaBatchSaver.Put([]byte(daStoreKey), daStoreData, nil)
					if storeErr != nil {
						logs.Log.Debug(fmt.Sprintf("Error in saving DA pointer in pod database : %s", storeErr.Error()))
						return
					}

					log.Info().Str("module", "p2p").Msg("Data Saved in DA")

				} else if Datype == "eigen" {

					for {
						daCheck, daCheckErr = eigen.Eigen(daDataByte, baseConfig.DA.DaRPC, baseConfig.DA.DaRPC)
						if daCheckErr != nil {
							logs.Log.Debug("Error in submitting data to DA " + daCheckErr.Error())
							logs.Log.Debug("Retrying Eigen DA after 10 seconds")
							time.Sleep(10 * time.Second)
						} else {
							break
						}
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
						logs.Log.Debug(fmt.Sprintf("Error in marshaling DA pointer : %s", daStoreDataErr.Error()))
						return
					}

					storeErr := DaBatchSaver.Put([]byte(daStoreKey), daStoreData, nil)
					if storeErr != nil {
						logs.Log.Debug(fmt.Sprintf("Error in saving DA pointer in pod database : %s", storeErr.Error()))
						return
					}

					log.Info().Str("module", "p2p").Msg("Data Saved in DA")

				} else {
					logs.Log.Error("Unknown layer. Please use 'avail' or 'celestia' as argument.")
					return
				}

				// submit pod
				success := junction2.SubmitCurrentPod()
				if !success {
					logs.Log.Error("Failed to submit pod")
					return
				}
				updateTxState(shared.TxStateVerifyPod)
			} else {
				log.Warn().Str("module", "p2p").Msg("Pod already submitted, moving to next step")
			}
			//os.Exit(0)

			// verify pod
			if shared.GetPodState().LatestTxState == shared.TxStateVerifyPod {
				success := junction2.VerifyCurrentPod()
				if !success {
					logs.Log.Error("Failed to Transact Verify pod")
					return
				}
				updateTxState(shared.TxStatePreInit)
			} else {
				log.Error().Str("module", "p2p").Msg("Database Error. LatestTxState should equal to TxStatePreInit at this point")
				log.Error().Str("module", "p2p").Msg("LatestTxState: " + shared.GetPodState().LatestTxState)
				return // stop sequencer, there is some error
			}

			saveVerifiedPOD()        // save data to database
			GenerateUnverifiedPods() // generate next pod
		} else {
			PodNumber := int(shared.GetPodState().LatestPodHeight)
			success, addr := junction2.InitVRF()
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

			// get txHash of vrfInit
			VrfInitTxHash := shared.GetPodState().VRFInitiationTxHash
			// send verify VRF message to selected node
			VRFInitiatedMsg := VRFInitiatedMsgData{
				PodNumber:            uint64(PodNumber),
				SelectedTrackAddress: selectedTrackAddress,
				VrfInitTxHash:        VrfInitTxHash,
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
		}
	}

}
