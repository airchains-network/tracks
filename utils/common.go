package utils

import (
	"context"
	"encoding/json"
	"fmt"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
)

func GetBalance(address string, blockNumber uint64, stationRPC string) (string, error) {
	payload := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"method": "eth_getBalance",
		"params": ["%s", "0x%s"],
		"id": 1
	}`, address, strconv.FormatUint(blockNumber, 16))

	resp, err := http.Post(stationRPC, "application/json", strings.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("http post error: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response body: %w", err)
	}

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return "", fmt.Errorf("Error unmarshalling JSON response: %w", err)
	}

	if errMsg, ok := jsonResponse["error"]; ok {
		return "", fmt.Errorf("Error from Ethereum node: %v", errMsg)
	}

	if result, ok := jsonResponse["result"].(string); ok {
		balance, success := new(big.Int).SetString(result[2:], 16)
		if !success {
			return "", fmt.Errorf("Failed to parse balance")
		}
		return balance.String(), nil
	} else {
		return "", fmt.Errorf("Failed to parse balance")
	}
}

func GetAccountNonce(ctx context.Context, address string, blockNumber uint64, stationRPC string) (string, error) {
	client, err := rpc.Dial(stationRPC)
	if err != nil {
		return "0", fmt.Errorf("Error dialing RPC: %w", err)
	}

	accountAddress := common.HexToAddress(address)
	formattedBlockNumber := "0x" + strconv.FormatUint(blockNumber, 16)
	var result string
	err = client.CallContext(ctx, &result, "eth_getTransactionCount", accountAddress, formattedBlockNumber)

	if err != nil {
		return "0", fmt.Errorf("error getting transaction count: %w", err)
	}

	return result, nil
}

func ToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []interface{}:
		jsonBytes, _ := json.Marshal(v)
		return string(jsonBytes)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func CreateAccount(accountName string, accountPath string) {
	registry, err := cosmosaccount.New(cosmosaccount.WithHome(accountPath))
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error creating account registry: %v", err))
		return
	}

	account, mnemonic, err := registry.Create(accountName)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error creating account: %v", err))
		return
	}

	newCreatedAccount, err := registry.GetByName(accountName)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error getting account: %v", err))
		return
	}

	newCreatedAccountAddr, err := newCreatedAccount.Address("air")
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error getting address: %v", err))
		return
	}

	logs.Log.Info(fmt.Sprintf("Account created: %s", account.Name))
	logs.Log.Info(fmt.Sprintf("Mnemonic: %s", mnemonic))
	logs.Log.Info(fmt.Sprintf("Address: %s", newCreatedAccountAddr))
	logs.Log.Info("Please save this mnemonic key for account recovery")
	logs.Log.Info("Please save this address for future reference")

}

func GenerateRandomWithFavour(lowerBound, upperBound int, favourableSet [2]int, favourableProbability float64) int {

	if lowerBound > upperBound || favourableProbability < 0 || favourableProbability > 1 {
		fmt.Println("Invalid parameters")
		return 0
	}

	// Calculate total range and the favourable range
	totalRange := upperBound - lowerBound + 1
	favourableRange := favourableSet[1] - favourableSet[0] + 1

	if favourableRange <= 0 || favourableRange > totalRange {
		fmt.Println("Invalid favourable set")
		return 0
	}

	// Check if the favourable set is within the total range
	if favourableSet[0] < lowerBound || favourableSet[1] > upperBound || favourableRange <= 0 {
		fmt.Println("Invalid favourable set")
		return 0
	}

	// Calculate the number of favourable outcomes based on the probability
	favourableOutcomes := int(favourableProbability * float64(totalRange))
	if favourableOutcomes < favourableRange {
		favourableOutcomes = favourableRange
	}

	// Generate a random number and adjust for favourable outcomes
	randNum := rand.Intn(totalRange)
	if randNum < favourableOutcomes {
		// Map the first `favourableOutcomes` to the favourable range
		randNum = randNum%favourableRange + favourableSet[0]
	} else {
		// Adjust the random number to exclude the favourable range and map to the rest of the range
		randNum = randNum%favourableOutcomes + lowerBound
		if randNum >= favourableSet[0] && randNum <= favourableSet[1] {
			randNum = favourableSet[1] + 1 + (randNum - favourableSet[0])
		}
	}

	return randNum
}
