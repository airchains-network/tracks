package p2p

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/tracks/config"
	"github.com/airchains-network/tracks/types"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func EspressoBatchSubmit(batchInput *types.BatchStruct, baseConfig *config.Config) (*types.EspressoSchemaV1, error) {

	espressoRPC := baseConfig.Sequencer.SequencerRPC
	//base64 encode og 25 batch input to form payload
	inputBytes, err := json.Marshal(batchInput)
	if err != nil {
		return nil, err
	}

	payload := base64.StdEncoding.EncodeToString(inputBytes)
	namespace, err := strconv.ParseUint(baseConfig.Sequencer.SequencerNamespace, 10, 64)
	if err != nil {
		return nil, err
	}
	data := types.Payload{
		Namespace: int(namespace),
		Payload:   payload,
	}

	// Convert the payload to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
		return nil, err
	}

	// Make the POST request
	submitURL := fmt.Sprintf("%s/v0/submit/submit", espressoRPC)

	resp, err := http.Post(submitURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error making POST request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	// Assuming the response body contains the transaction hash as plain text
	txHash := string(body)
	txHash = strings.ReplaceAll(txHash, `"`, "")

	fmt.Printf("Transaction Hash: %s\n", txHash)

	// Now make the GET request to check availability using the returned transaction hash
	availabilityURL := fmt.Sprintf("%s/v0/availability/transaction/hash/%s", espressoRPC, txHash)

	// Make the GET request
	resp, err = http.Get(availabilityURL)
	if err != nil {
		log.Fatalf("Error making GET request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response from the availability check
	availabilityBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading availability response body: %v", err)
		return nil, err
	}

	// Print the availability response
	fmt.Printf("Availability Response: %s\n", availabilityBody)

	// Declare the variable to hold the parsed response
	var espressoTxResponseTemp types.EspressoTxResponseV1Temp

	// Unmarshal the JSON into the struct
	err = json.Unmarshal(availabilityBody, &espressoTxResponseTemp)
	if err != nil {
		// todo if it give error then new schema may needed
		log.Fatalf("Error unmarshaling JSON: %v", err)
		return nil, err
	}
	var espressoTxResponse types.EspressoTxResponseV1
	// Copy data from the temporary struct to the final struct, converting []int to []string where needed
	espressoTxResponse.Transaction.Namespace = espressoTxResponseTemp.Transaction.Namespace
	espressoTxResponse.Transaction.Payload = espressoTxResponseTemp.Transaction.Payload
	espressoTxResponse.Hash = espressoTxResponseTemp.Hash
	espressoTxResponse.Index = espressoTxResponseTemp.Index

	// Convert slices and assign
	espressoTxResponse.Proof.TxIndex = convertIntSliceToString(espressoTxResponseTemp.Proof.TxIndex)
	espressoTxResponse.Proof.PayloadNumTxs = convertIntSliceToString(espressoTxResponseTemp.Proof.PayloadNumTxs)
	espressoTxResponse.Proof.PayloadProofNumTxs.PrefixBytes = convertIntSliceToBase64(espressoTxResponseTemp.Proof.PayloadProofNumTxs.PrefixBytes)
	espressoTxResponse.Proof.PayloadProofNumTxs.SuffixBytes = convertIntSliceToBase64(espressoTxResponseTemp.Proof.PayloadProofNumTxs.SuffixBytes)
	espressoTxResponse.Proof.PayloadTxTableEntries = convertIntSliceToString(espressoTxResponseTemp.Proof.PayloadTxTableEntries)
	espressoTxResponse.Proof.PayloadProofTxTableEntries.PrefixBytes = convertIntSliceToBase64(espressoTxResponseTemp.Proof.PayloadProofTxTableEntries.PrefixBytes)
	espressoTxResponse.Proof.PayloadProofTxTableEntries.SuffixBytes = convertIntSliceToBase64(espressoTxResponseTemp.Proof.PayloadProofTxTableEntries.SuffixBytes)
	espressoTxResponse.Proof.PayloadProofTx.PrefixBytes = convertIntSliceToBase64(espressoTxResponseTemp.Proof.PayloadProofTx.PrefixBytes)
	espressoTxResponse.Proof.PayloadProofTx.SuffixBytes = convertIntSliceToBase64(espressoTxResponseTemp.Proof.PayloadProofTx.SuffixBytes)

	// Handle the rest of the values
	espressoTxResponse.BlockHash = espressoTxResponseTemp.BlockHash
	espressoTxResponse.BlockHeight = espressoTxResponseTemp.BlockHeight

	// Now, espressoTxResponse contains all values as strings, and it's ready to use

	var espressoSchemaV1 = types.EspressoSchemaV1{EspressoTxResponseV1: espressoTxResponse, StationId: baseConfig.Junction.StationId}

	return &espressoSchemaV1, nil
}

// Function to convert []int to []string
func convertIntSliceToString(intSlice []int) string {
	var stringSlice []string
	for _, num := range intSlice {
		stringSlice = append(stringSlice, fmt.Sprintf("%d", num))
	}
	// Join the string slice into a single string separated by commas (or any delimiter you choose)
	return strings.Join(stringSlice, ",")
}

func convertIntSliceToBytes(intSlice []int) []byte {
	var byteSlice []byte
	for _, num := range intSlice {
		// Convert the integer to its byte representation
		byteSlice = append(byteSlice, byte(num))
	}
	return byteSlice
}

// Function to convert []int to a Base64-encoded string
func convertIntSliceToBase64(intSlice []int) string {
	// First, convert []int to []byte
	var byteSlice []byte
	for _, num := range intSlice {
		byteSlice = append(byteSlice, byte(num))
	}
	// Now, encode the []byte slice into a Base64 string
	return base64.StdEncoding.EncodeToString(byteSlice)
}
