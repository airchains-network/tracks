package p2p

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/tracks/config"
	logs "github.com/airchains-network/tracks/log"
	"github.com/airchains-network/tracks/types"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

func EspressoBatchSubmit(batchInput *types.BatchStruct, baseConfig *config.Config) (*types.EspressoData, error) {

	//hashes := batchInput.TransactionHash
	//for i := 0; i < len(hashes); i++ {
	//	fmt.Println(hashes[i])
	//}

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
		logs.Log.Error(fmt.Sprintf("Error making GET request: %v", err))
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response from the availability check
	availabilityBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error reading availability response body: %v", err))
		return nil, err
	}

	// Print the availability response
	//fmt.Printf("Availability Response: %s\n", availabilityBody)

	// Declare the variable to hold the parsed response
	var espressoTxResponseTemp types.EspressoTxResponseV1Temp
	espressoTxResponseTemp.Index = -1
	espressoTxResponseTemp.BlockHeight = -1

	//Unmarshal the JSON into the struct
	err = json.Unmarshal(availabilityBody, &espressoTxResponseTemp)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
		return nil, err
	}

	isFieldsValid := CheckFieldsV1(espressoTxResponseTemp)
	if !isFieldsValid {
		return nil, err
	}

	// todo if it give error then new schema may needed

	//fmt.Println(string(availabilityBody))
	//fmt.Println(espressoTxResponseTemp)
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
	fmt.Println(espressoSchemaV1)

	schemaObjectByte, err := json.Marshal(espressoSchemaV1)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error  marshaling JSON: %v", err))
		return nil, err
	}

	espressoData := types.EspressoData{
		Version: "v1.0.0",
		Data:    schemaObjectByte,
	}

	//type EspressoData struct {
	//	version string
	//	data []byte
	//}

	//

	//return &espressoSchemaV1, nil
	return &espressoData, nil
}
func saveStructAsJSON(filename string, data interface{}) error {
	// Marshal the struct into JSON
	jsonData, err := json.MarshalIndent(data, "", "  ") // Pretty print the JSON
	if err != nil {
		return fmt.Errorf("error marshaling struct to JSON: %v", err)
	}

	// Write the JSON to a file
	err = ioutil.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing JSON to file: %v", err)
	}

	return nil
}

// Function to convert []int to []string
func convertIntSliceToString(intSlice []int) string {
	var stringSlice []string
	for _, num := range intSlice {
		stringSlice = append(stringSlice, fmt.Sprintf("%d", num))
	}
	// Join the string slice into a single string separated by commas (or any delimiter you choose)
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(stringSlice, ",")))
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
	//if len(intSlice) == 0 {
	//	byteSlice = append(byteSlice, byte(0))
	//} else {
	for _, num := range intSlice {
		byteSlice = append(byteSlice, byte(num))
	}
	//}
	// Now, encode the []byte slice into a Base64 string
	return base64.StdEncoding.EncodeToString(byteSlice)
}

func saveEspressoPod(ldt *leveldb.DB, EspressoTxResponse *types.EspressoData, podNum int) error {
	// marshal
	EspressoTxResponseByte, err := json.Marshal(EspressoTxResponse)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling tx data : %s", err.Error()))
		return err
	}

	fmt.Println(string(EspressoTxResponseByte))
	podNumByte := []byte(strconv.Itoa(podNum))
	fmt.Println(podNumByte)

	err = ldt.Put(podNumByte, EspressoTxResponseByte, nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in saving tx data : %s", err.Error()))
		return err
	}
	fmt.Println("saved")
	return nil
}

// isEmpty checks if a value is empty (zero value)
func isEmpty(field interface{}) bool {
	v := reflect.ValueOf(field)

	// Check for empty values based on the kind of the field
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Slice:
		return v.Len() == 0
	}
	return false
}
func isNegative(num int) bool {
	if num < 0 {
		return true
	}
	return false
}

// CheckFields checks if any key fields are empty, prints them and returns false if any are empty
func CheckFieldsV1(espressoTx types.EspressoTxResponseV1Temp) bool {
	// Check each field in the struct, print if empty, and return false immediately if an empty field is found
	if isEmpty(espressoTx.Transaction.Namespace) {
		logs.Log.Error("Transaction.Namespace is empty")
		return false
	}
	if isEmpty(espressoTx.Transaction.Payload) {
		logs.Log.Error("Transaction.Payload is empty")
		return false
	}
	if isEmpty(espressoTx.Hash) {
		logs.Log.Error("Hash is empty")
		return false
	}
	if isNegative(espressoTx.Index) {
		logs.Log.Error("Index is empty")
		return false
	}
	if isEmpty(espressoTx.Proof.TxIndex) {
		logs.Log.Error("Proof.TxIndex is empty")
		return false
	}
	if isEmpty(espressoTx.Proof.PayloadNumTxs) {
		logs.Log.Error("Proof.PayloadNumTxs is empty")
		return false
	}
	if isEmpty(espressoTx.Proof.PayloadProofNumTxs.Proofs) {
		logs.Log.Error("Proof.PayloadProofNumTxs.Proofs is empty")
		return false
	}
	if isEmpty(espressoTx.Proof.PayloadProofNumTxs.PrefixBytes) {
		logs.Log.Debug("Proof.PayloadProofNumTxs.PrefixBytes is empty")
		//return false
	}
	if isEmpty(espressoTx.Proof.PayloadProofNumTxs.SuffixBytes) {
		logs.Log.Debug("Proof.PayloadProofNumTxs.SuffixBytes is empty")
		//return false
	}
	if isEmpty(espressoTx.Proof.PayloadTxTableEntries) {
		logs.Log.Error("Proof.PayloadTxTableEntries is empty")
		return false
	}
	if isEmpty(espressoTx.Proof.PayloadProofTxTableEntries.Proofs) {
		logs.Log.Error("Proof.PayloadProofTxTableEntries.Proofs is empty")
		return false
	}
	if isEmpty(espressoTx.Proof.PayloadProofTxTableEntries.PrefixBytes) {
		logs.Log.Debug("Proof.PayloadProofTxTableEntries.PrefixBytes is empty")
		//return false
	}
	if isEmpty(espressoTx.Proof.PayloadProofTxTableEntries.SuffixBytes) {
		logs.Log.Debug("Proof.PayloadProofTxTableEntries.SuffixBytes is empty")
		//return false
	}
	if isEmpty(espressoTx.Proof.PayloadProofTx.Proofs) {
		logs.Log.Error("Proof.PayloadProofTx.Proofs is empty")
		return false
	}
	if isEmpty(espressoTx.Proof.PayloadProofTx.PrefixBytes) {
		logs.Log.Debug("Proof.PayloadProofTx.PrefixBytes is empty")
		//return true
	}
	if isEmpty(espressoTx.Proof.PayloadProofTx.SuffixBytes) {
		logs.Log.Debug("Proof.PayloadProofTx.SuffixBytes is empty")
		//return true
	}
	if isEmpty(espressoTx.BlockHash) {
		logs.Log.Error("BlockHash is empty")
		return false
	}
	if isNegative(espressoTx.BlockHeight) {
		logs.Log.Error("BlockHeight is empty")
		return false
	}

	// If no fields are empty, return true
	return true
}
