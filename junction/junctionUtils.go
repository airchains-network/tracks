package junction

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/tracks/config"
	junctionTypes "github.com/airchains-network/tracks/junction/types"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/node/shared"
	"github.com/airchains-network/tracks/types"
	"github.com/consensys/gnark/backend/groth16"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"io"
	"math/big"
	"os"
	"path/filepath"
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
		logs.Log.Error("Error marshaling to JSON:" + err.Error())
		return false
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.Log.Error("Error in getting home dir path: " + err.Error())
		return false
	}

	GenesisFilePath := filepath.Join(homeDir, config.DefaultGenesisFilePath)

	// Write the JSON data to a file
	err = os.WriteFile(GenesisFilePath, jsonBytes, 0644)
	if err != nil {
		logs.Log.Error("Error writing JSON to file:" + err.Error())
		return false
	} else {
		logs.Log.Info(GenesisFilePath + " created")
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.Log.Error("Error in getting home dir path: " + err.Error())
	}

	ConfigFilePath := filepath.Join(homeDir, config.DefaultTracksDir, config.DefaultConfigDir)
	VRFPubKeyPath := filepath.Join(ConfigFilePath, "vrfPubKey.txt")
	file, err := os.Create(VRFPubKeyPath)
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.Log.Error("Error in getting home dir path: " + err.Error())
	}
	ConfigFilePath := filepath.Join(homeDir, config.DefaultTracksDir, config.DefaultConfigDir)
	VRFPrivKeyPath := filepath.Join(ConfigFilePath, "vrfPrivKey.txt")
	file, err := os.Create(VRFPrivKeyPath)
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.Log.Error("Error in getting home dir path: " + err.Error())
	}

	ConfigFilePath := filepath.Join(homeDir, config.DefaultTracksDir, config.DefaultConfigDir)
	VRFPrivKeyPath := filepath.Join(ConfigFilePath, "vrfPrivKey.txt")
	file, err := os.Open(VRFPrivKeyPath)
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

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.Log.Error("Error in getting home dir path: " + err.Error())
	}

	ConfigFilePath := filepath.Join(homeDir, config.DefaultTracksDir, config.DefaultConfigDir)
	VRFPubKeyPath := filepath.Join(ConfigFilePath, "vrfPubKey.txt")

	// get private Key
	file, err := os.Open(VRFPubKeyPath)
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

func GetJunctionDetails() (JsonRPC, StationId, AccountPath, AccountName, AddressPrefix string, Tracks []string, err error) {
	// Specify the file path and name

	baseConfig, err := shared.LoadConfig()
	if err != nil {
		errorMsg := fmt.Errorf("error in aloading config file")
		return "", "", "", "", "", Tracks, errorMsg
	}

	JsonRPC = baseConfig.Junction.JunctionRPC
	StationId = baseConfig.Junction.StationId
	AccountPath = baseConfig.Junction.AccountPath
	AccountName = baseConfig.Junction.AccountName
	AddressPrefix = baseConfig.Junction.AddressPrefix
	Tracks = baseConfig.Junction.Tracks

	if JsonRPC == "" {
		errorMsg := fmt.Errorf("JsonRPC should not be empty at config file")
		return "", "", "", "", "", Tracks, errorMsg
	}
	if StationId == "" {
		errorMsg := fmt.Errorf("StationId should not be empty at config file. Create station first")
		return "", "", "", "", "", Tracks, errorMsg
	}
	if AccountPath == "" {
		errorMsg := fmt.Errorf("AccountPath should not be empty at config file")
		return "", "", "", "", "", Tracks, errorMsg
	}
	if AccountName == "" {
		errorMsg := fmt.Errorf("AccountName should not be empty at config file")
		return "", "", "", "", "", Tracks, errorMsg
	}
	if AddressPrefix == "" {
		errorMsg := fmt.Errorf("AddressPrefix should not be empty at config file")
		return "", "", "", "", "", Tracks, errorMsg
	}

	if len(Tracks) == 0 {
		errorMsg := fmt.Errorf("tracks filed should not be empty in config file")
		return "", "", "", "", "", Tracks, errorMsg
	}

	// Return the StationId from the unmarshaled data
	return JsonRPC, StationId, AccountPath, AccountName, AddressPrefix, Tracks, nil
}
