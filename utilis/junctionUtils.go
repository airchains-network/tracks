package utilis

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	junctionTypes "github.com/airchains-network/decentralized-sequencer/junction/types"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/consensys/gnark/backend/groth16"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"io"
	"math/big"
	"os"
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

func SerializeRequestCommitmentV2Plus(rc types.RequestCommitmentV2Plus) ([]byte, error) {
	var buf bytes.Buffer

	// Encode the blockNum
	err := binary.Write(&buf, binary.BigEndian, rc.BlockNum)
	if err != nil {
		return nil, fmt.Errorf("failed to encode blockNum: %w", err)
	}

	// Encode the stationId as a fixed size or prefixed with its length
	// Here, we choose to prefix with length for simplicity
	if err := binary.Write(&buf, binary.BigEndian, uint64(len(rc.StationId))); err != nil {
		return nil, fmt.Errorf("failed to encode stationId length: %w", err)
	}
	buf.WriteString(rc.StationId)

	// Encode the upperBound
	err = binary.Write(&buf, binary.BigEndian, rc.UpperBound)
	if err != nil {
		return nil, fmt.Errorf("failed to encode upperBound: %w", err)
	}

	// Encode the requesterAddress as a fixed size or prefixed with its length
	if err := binary.Write(&buf, binary.BigEndian, uint64(len(rc.RequesterAddress))); err != nil {
		return nil, fmt.Errorf("failed to encode requesterAddress length: %w", err)
	}
	buf.WriteString(rc.RequesterAddress)

	// Encode the extraArgs
	//buf.WriteByte(rc.ExtraArgs)

	return buf.Bytes(), nil
}

func GenerateVRFProof(suite kyber.Group, privateKey kyber.Scalar, data []byte, nonce int64) ([]byte, []byte, error) {
	// Convert nonce to a deterministic scalar
	nonceBytes := big.NewInt(nonce).Bytes()
	nonceScalar := suite.Scalar().SetBytes(nonceBytes)

	// Generate proof like in a Schnorr signature: R = g^k, s = k + e*x
	R := suite.Point().Mul(nonceScalar, nil) // R = g^k
	hash := sha256.New()
	rBytes, _ := R.MarshalBinary()
	hash.Write(rBytes)
	hash.Write(data)
	e := suite.Scalar().SetBytes(hash.Sum(nil))                             // e = H(R||data)
	s := suite.Scalar().Add(nonceScalar, suite.Scalar().Mul(e, privateKey)) // s = k + e*x

	// The VRF output (pseudo-random value) is hash of R combined with data
	vrfHash := sha256.New()
	vrfHash.Write(rBytes)         // Incorporate R
	vrfHash.Write(data)           // Incorporate input data
	vrfOutput := vrfHash.Sum(nil) // This is the deterministic "random" output

	// Serialize R and s into the proof
	sBytes, _ := s.MarshalBinary()
	proof := append(rBytes, sBytes...)

	return proof, vrfOutput, nil
}

type JunctionDetails struct {
	JsonRPC       string
	StationId     string
	AccountPath   string
	AccountName   string
	AddressPrefix string
}

func SetJunctionDetails(JsonRPC, StationId, AccountPath, AccountName, AddressPrefix string) (success bool) {
	stationData := JunctionDetails{
		JsonRPC:       JsonRPC,
		StationId:     StationId,
		AccountPath:   AccountPath,
		AccountName:   AccountName,
		AddressPrefix: AddressPrefix,
	}

	// Marshal the data into JSON
	stationDataBytes, err := json.MarshalIndent(stationData, "", "    ")
	if err != nil {
		//logs.Log.Error("Error marshaling to JSON:" + err.Error())
		return false
	}

	// Specify the file path and name
	filePath := "data/stationData.json"

	// Write the JSON data to a file
	err = os.WriteFile(filePath, stationDataBytes, 0644)
	if err != nil {
		logs.Log.Error("Error writing JSON to file:" + err.Error())
		return false
	} else {
		logs.Log.Info(filePath + " created")
	}

	return true
}

func GetJunctionDetails() (JsonRPC, StationId, AccountPath, AccountName, AddressPrefix string, err error) {
	// Specify the file path and name
	filePath := "data/stationData.json"

	// Read the JSON data from the file
	jsonBytes, err := os.ReadFile(filePath)
	if err != nil {
		logs.Log.Error("Error reading JSON file:" + err.Error())
		return "", "", "", "", "", err
	}

	// Unmarshal the JSON data into GenesisDataType struct
	var junctionData JunctionDetails
	err = json.Unmarshal(jsonBytes, &junctionData)
	if err != nil {
		logs.Log.Error("Error unmarshaling JSON:" + err.Error())
		return "", "", "", "", "", err
	}

	JsonRPC = junctionData.JsonRPC
	StationId = junctionData.StationId
	AccountPath = junctionData.AccountPath
	AccountName = junctionData.AccountName
	AddressPrefix = junctionData.AddressPrefix

	if JsonRPC == "" || StationId == "" || AccountPath == "" || AccountName == "" {
		logs.Log.Error("Some fields are empty in data/stationData.json")
		errorMsg := fmt.Errorf("Some fields are empty in data/stationData.json")
		return "", "", "", "", "", errorMsg
	}

	// Return the StationId from the unmarshaled data
	return JsonRPC, StationId, AccountPath, AccountName, AddressPrefix, nil
}
