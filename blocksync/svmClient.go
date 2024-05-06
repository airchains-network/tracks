package blocksync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/types/svmTypes"
	"github.com/deadlium/deadlogs"
	"io"
	"net/http"
	"sync"
)

var SVMChainRPCUrl string

func initSVMRPC(JsonRPC string) {
	SVMChainRPCUrl = JsonRPC
}

func svmRPCCall(method string, value any) ([]byte, error) {
	payload := SVMPayLoad(method, value)

	jsonPayload, jsonPayloadErr := json.Marshal(payload)
	if jsonPayloadErr != nil {
		return nil, fmt.Errorf("error marshaling JSON: %v", jsonPayloadErr)
	}

	client := &http.Client{}

	req, reqErr := http.NewRequest("POST", SVMChainRPCUrl, bytes.NewBuffer(jsonPayload))
	if reqErr != nil {
		return nil, fmt.Errorf("error creating request: %v", reqErr)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, respErr := client.Do(req)
	if respErr != nil {
		return nil, fmt.Errorf("error sending request: %v", respErr)
	}
	defer func(Body io.ReadCloser) {
		bodyErr := Body.Close()
		if bodyErr != nil {
			deadlogs.Warn(fmt.Sprintf("error while closing body: %v", bodyErr))
		}
	}(resp.Body)

	body, bodyErr := io.ReadAll(resp.Body)
	if bodyErr != nil {
		return nil, fmt.Errorf("error reading response: %v", bodyErr)
	}

	return body, nil
}

func SVMLatestBlockCheck() (int, error) {
	res, resErr := svmRPCCall("getSlot", nil)
	if resErr != nil {
		return 0, fmt.Errorf("error rpc call: %v", resErr)
	}

	var latestBlock svmTypes.LatestSlotStruct
	latestBlockErr := json.Unmarshal(res, &latestBlock)
	if latestBlockErr != nil {
		return 0, fmt.Errorf("error decoding response: %v", latestBlockErr)
	}

	return latestBlock.Result, nil
}

func SVMBlockCall(height int) (*svmTypes.BlockResponseStruct, error) {

	res, resErr := svmRPCCall("getBlock", height)
	if resErr != nil {
		return nil, fmt.Errorf("error rpc call: %v", resErr)
	}

	var blockData svmTypes.BlockResponseStruct
	blockDataErr := json.Unmarshal(res, &blockData)
	if blockDataErr != nil {
		return nil, fmt.Errorf("error decoding response: %v", blockDataErr)
	}

	return &blockData, nil
}

func SVMBlockLeaderCall(height int) (*svmTypes.SlotLeaderResponseStruct, error) {

	res, resErr := svmRPCCall("getSlotLeaders", height)
	if resErr != nil {
		return nil, fmt.Errorf("error rpc call: %v", resErr)
	}

	var leaderData svmTypes.SlotLeaderResponseStruct
	blockLeaderErr := json.Unmarshal(res, &leaderData)
	if blockLeaderErr != nil {
		return nil, fmt.Errorf("error decoding response: %v", blockLeaderErr)
	}

	return &leaderData, nil
}

func SVMAccountListCall() ([]svmTypes.AccountDetailsResponseStruct, error) {

	var leaderCircleData svmTypes.LargeAccountStruct
	var leaderNonCircleData svmTypes.LargeAccountStruct

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		res, resErr := svmRPCCall("getLargestAccounts", "circulating")
		if resErr != nil {
			deadlogs.Warn(fmt.Sprintf("error rpc call: %v", resErr))
		}

		blockLeaderErr := json.Unmarshal(res, &leaderCircleData)
		if blockLeaderErr != nil {
			deadlogs.Warn(fmt.Sprintf("error decoding response: %v", blockLeaderErr))
		}
	}()

	go func() {
		defer wg.Done()
		res, resErr := svmRPCCall("getLargestAccounts", "nonCirculating")
		if resErr != nil {
			deadlogs.Warn(fmt.Sprintf("error rpc call: %v", resErr))
		}

		blockLeaderErr := json.Unmarshal(res, &leaderNonCircleData)
		if blockLeaderErr != nil {
			deadlogs.Warn(fmt.Sprintf("error decoding response: %v", blockLeaderErr))
		}
	}()

	wg.Wait()

	var accountArray []string
	for _, account := range leaderCircleData.Result.Value {
		accountArray = append(accountArray, account.Address)
	}

	for _, account := range leaderNonCircleData.Result.Value {
		accountArray = append(accountArray, account.Address)
	}

	details, detailsErr := SVMAccountDetailsCall(accountArray)
	if detailsErr != nil {
		return nil, fmt.Errorf("error fetching account details: %v", detailsErr)
	}

	var accountDetails []svmTypes.AccountDetailsResponseStruct

	for i, account := range details.Result.Value {
		accountDetails = append(accountDetails, svmTypes.AccountDetailsResponseStruct{
			Address: accountArray[i],
			Value:   account,
		})
	}

	return accountDetails, nil
}

func SVMAccountDetailsCall(address []string) (*svmTypes.AccountDetailsStruct, error) {

	res, resErr := svmRPCCall("getMultipleAccounts", address)
	if resErr != nil {
		return nil, fmt.Errorf("error rpc call: %v", resErr)
	}

	var leaderData svmTypes.AccountDetailsStruct
	blockLeaderErr := json.Unmarshal(res, &leaderData)
	if blockLeaderErr != nil {
		return nil, fmt.Errorf("error decoding response: %v", blockLeaderErr)
	}

	return &leaderData, nil
}

func SVMPayLoad(method string, value any) svmTypes.PayloadStruct {
	if method == "getSlot" {
		return svmTypes.PayloadStruct{
			JsonRPC: "2.0",
			ID:      1,
			Method:  method,
			Params:  make([]interface{}, 0),
		}
	}

	if method == "getSlotLeaders" {
		return svmTypes.PayloadStruct{
			JsonRPC: "2.0",
			ID:      1,
			Method:  method,
			Params: []interface{}{
				value,
				1,
			},
		}
	}

	if method == "getLargestAccounts" {
		return svmTypes.PayloadStruct{
			JsonRPC: "2.0",
			ID:      1,
			Method:  method,
			Params: []interface{}{
				struct {
					Filter string `json:"filter"`
				}{
					Filter: value.(string),
				},
			},
		}
	}

	return svmTypes.PayloadStruct{
		JsonRPC: "2.0",
		ID:      1,
		Method:  method,
		Params: []interface{}{
			value,
			svmTypes.Params{
				Encoding:                       "jsonParsed",
				MaxSupportedTransactionVersion: 0,
				TransactionDetails:             "full",
				Rewards:                        true,
			},
		},
	}
}
