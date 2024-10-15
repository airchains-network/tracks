package p2p

import (
	"encoding/json"
	"fmt"
	//"github.com/airchains-network/tracks/config"
	"github.com/airchains-network/tracks/da/avail"
	"github.com/airchains-network/tracks/da/celestia"
	"github.com/airchains-network/tracks/da/eigen"
	mock "github.com/airchains-network/tracks/da/mockda"
	"github.com/airchains-network/tracks/junction/trackgate"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/node/shared"
	"github.com/airchains-network/tracks/types"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"strings"
	"time"
)

func TrackgatePodGenerator() {
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().
		Str("module", "p2p").
		Msg("Generating New unverified pods")

	connection := shared.Node.NodeConnections

	staticDBConnection := connection.GetStaticDatabaseConnection()
	espressoDBConnection := connection.GetEspressoDatabaseConnection()
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
		batchInput           *types.BatchStruct
		txState              string
		trackAppHash         []byte
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

	//podData := junction.QueryPod(uint64(currentPodNumber))
	//if podData != nil {
	//	if podData.IsVerified == true {
	//		currentPodNumber++
	//	}
	//}

	batchNumber = currentPodNumber
	log.Info().Str("module", "p2p").Msg(fmt.Sprintf("Processing Pod Number: %d", batchNumber))

	// create batch
	//if podStateData.LatestTxState == shared.TxStatePreInit {
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
		batchInput, err = createEVMBatch(txnDBConnection, rawConfirmedTransactionIndex, rawCurrentPodNumber)
		CheckErrorAndExit(err, "Error in creating POD", 0)
	} else if stationVariantLowerCase == "wasm" {
		batchInput, err = createWasmBatch(txnDBConnection, rawConfirmedTransactionIndex, rawCurrentPodNumber)
		CheckErrorAndExit(err, "Error in creating POD", 0)
	}

	trackAppHash = generateBatchHash(rawCurrentPodNumber)
	updateNewBatchState(trackAppHash, uint64(batchNumber), batchInput, txState)
	//}
	//
	//else {
	//	trackAppHash = podStateData.TracksAppHash
	//	batchInput = podStateData.Batch
	//
	//	storeNewBatchState(trackAppHash, uint64(batchNumber), batchInput, txState)
	//}

	//	multi node
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

			baseConfig, err := shared.LoadConfig()
			if err != nil {
				fmt.Println("Error loading configuration")
			}

			if shared.GetPodState().LatestTxState == shared.TxStateSubmitPod {

				DaBatchSaver := connection.DataAvailabilityDatabaseConnection

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

			}

			// espresso data submit
			EspressoTxResponse, err := EspressoBatchSubmit(batchInput, baseConfig, PodNumber)
			if err != nil {
				logs.Log.Error("Error in submitting data to Espresso")
				return
			}

			// schema engage
			success := trackgate.SchemaEngage(baseConfig, PodNumber, EspressoTxResponse.Data)
			if !success {
				logs.Log.Error("Failed to submit pod")
				return
			} else {
				logs.Log.Info("Successfully submitted pod")
			}

			saveEspressoPod(espressoDBConnection, EspressoTxResponse, PodNumber)

			saveVerifiedPOD()
			TrackgatePodGenerator()
			//os.Exit(0)
		} else {
			log.Warn().Str("module", "p2p").Msg("Pod already submitted, moving to next step")
		}

	}

}
