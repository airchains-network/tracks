package utilis

import (
	"encoding/json"
	"fmt"
	junctionTypes "github.com/airchains-network/decentralized-sequencer/junction/types"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/consensys/gnark/backend/groth16"
	"go.dedis.ch/kyber/v3"
	"io"
	"os"

	"go.dedis.ch/kyber/v3/group/edwards25519"
)

func CreateGenesisJson(stationInfo types.StationInfo, verificationKey groth16.VerifyingKey, stationId string, tracks []string, tracksVotingPower []uint64, txHash string, transactionTime string, extraArg junctionTypes.StationArg, creator string) (success bool) {

	genesisData := types.GenesisDataType{
		StationId:          stationId,
		Creator:            creator,
		CreationTime:       transactionTime,
		TxHash:             txHash,
		Tracks:             tracks,
		TracksVotingPowers: tracksVotingPower,
		VerificationKey:    verificationKey,
		ExtraArg:           extraArg,
		StationInfo:        stationInfo,
	}

	// Marshal the data into JSON
	jsonBytes, err := json.MarshalIndent(genesisData, "", "    ")
	if err != nil {
		//logs.Log.Error("Error marshaling to JSON:" + err.Error())
		return false
	}

	// Specify the file path and name
	filePath := "data/genesis.json"

	// Write the JSON data to a file
	err = os.WriteFile(filePath, jsonBytes, 0644)
	if err != nil {
		//logs.Log.Error("Error writing JSON to file:" + err.Error())
		return false
	} else {
		logs.Log.Info(filePath + " created")
	}

	return true
}

func GetStationIdFromGenesis() (stationId string) {
	// Specify the file path and name
	filePath := "data/genesis.json"

	// Read the JSON data from the file
	jsonBytes, err := os.ReadFile(filePath)
	if err != nil {
		logs.Log.Error("Error reading JSON file:" + err.Error())
		return ""
	}

	// Unmarshal the JSON data into GenesisDataType struct
	var genesisData types.GenesisDataType
	err = json.Unmarshal(jsonBytes, &genesisData)
	if err != nil {
		logs.Log.Error("Error unmarshaling JSON:" + err.Error())
		return ""
	}

	// Return the StationId from the unmarshaled data
	return genesisData.StationId
}

func NewKeyPair() (privateKeyX kyber.Scalar, publicKeyX kyber.Point) {
	// Initialize the Kyber suite for Edwards25519 curve
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Generate a new private key
	privateKey := suite.Scalar().Pick(suite.RandomStream())

	// Derive the public key from the private key
	publicKey := suite.Point().Mul(privateKey, nil)

	return privateKey, publicKey
}

func SetVRFPubKey(pubKey string) {
	// Create or open the file chainId.txt
	file, err := os.Create("data/vrfPubKey.txt")
	if err != nil {
		// Handle the error if the file cannot be created
		logs.Log.Error(fmt.Sprintf("error creating vrfPubKey.txt: %v", err))
		return
	}
	defer file.Close()

	// Write the stationId to the file
	_, err = file.WriteString(pubKey)
	if err != nil {
		// Handle the error if the file cannot be written to
		logs.Log.Error(fmt.Sprintf("error writing to vrfPubKey.txt: %v", err))
		return
	}

	// Save the file
	err = file.Sync()
	if err != nil {
		// Handle the error if the file cannot be saved
		logs.Log.Error(fmt.Sprintf("error saving vrfPubKey.txt: %v", err))
		return
	}

	// Print the stationId
	logs.Log.Info(fmt.Sprintf("vrfPubKey ID: %s", pubKey))
}

func SetVRFPrivKey(privateKey string) {
	// Create or open the file chainId.txt
	file, err := os.Create("data/vrfPrivKey.txt")
	if err != nil {
		// Handle the error if the file cannot be created
		logs.Log.Error(fmt.Sprintf("error creating vrfPrivKey.txt: %v", err))
		return
	}
	defer file.Close()

	// Write the stationId to the file
	_, err = file.WriteString(privateKey)
	if err != nil {
		// Handle the error if the file cannot be written to
		logs.Log.Error(fmt.Sprintf("error writing to vrfPrivKey.txt: %v", err))
		return
	}

	// Save the file
	err = file.Sync()
	if err != nil {
		// Handle the error if the file cannot be saved
		logs.Log.Error(fmt.Sprintf("error saving vrfPrivKey.txt: %v", err))
		return
	}

	// Print the stationId
	logs.Log.Info(fmt.Sprintf("vrfPrivKey ID: %s", privateKey))
}

func GetVRFPrivateKey() (privateKey string) {
	// get private Key
	file, err := os.Open("data/vrfPrivKey.txt")
	if err != nil {
		logs.Log.Error("Can not get VRF private key: " + err.Error())
		return ""
	}
	defer file.Close()
	buf := make([]byte, 1024) // Buffer size of 1024 bytes
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break // End of file reached
		}
		if err != nil {
			logs.Log.Error("Can not get VRF private key: " + err.Error())
			return ""
		}
		privateKey = string(buf[:n])
	}

	return privateKey
}

func GetVRFPubKey() (pubKey string) {
	// get private Key
	file, err := os.Open("data/vrfPubKey.txt")
	if err != nil {
		logs.Log.Error("Can not get VRF public key: " + err.Error())
		return ""
	}
	defer file.Close()
	buf := make([]byte, 1024) // Buffer size of 1024 bytes
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break // End of file reached
		}
		if err != nil {
			logs.Log.Error("Can not get VRF public key: " + err.Error())
			return ""
		}
		pubKey = string(buf[:n])
	}

	return pubKey
}
