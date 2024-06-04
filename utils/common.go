package utils

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
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
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling JSON response: %w", err)
	}

	if errMsg, ok := jsonResponse["error"]; ok {
		return "", fmt.Errorf("error from Ethereum node: %v", errMsg)
	}

	if result, ok := jsonResponse["result"].(string); ok {
		balance, success := new(big.Int).SetString(result[2:], 16)
		if !success {
			return "", fmt.Errorf("Failed to parse balance")
		}
		return balance.String(), nil
	} else {
		return "", fmt.Errorf("failed to parse balance")
	}
}

func GetAccountNonce(ctx context.Context, address string, blockNumber uint64, stationRPC string) (string, error) {
	client, err := rpc.Dial(stationRPC)
	if err != nil {
		return "0", fmt.Errorf("error dialing RPC: %w", err)
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

func ImportAccountByMnemonic(accountName string, accountPath string, mnemonic string) {
	registry, err := cosmosaccount.New(cosmosaccount.WithHome(accountPath))
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error creating account registry: %v", err))
		return
	}

	algos, _ := registry.Keyring.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), algos)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error getting signing algorithm: %v", err))
		return
	}

	registryPath := hd.CreateHDPath(sdktypes.GetConfig().GetCoinType(), 0, 0).String()
	record, err := registry.Keyring.NewAccount(accountName, mnemonic, "", registryPath, algo)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error Importing account: %v", err))
		return
	}
	account := cosmosaccount.Account{
		Name:   accountName,
		Record: record,
	}

	// account to address
	accountAddr, err := account.Address("air")
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error getting address: %v", err))
		return
	}
	logs.Log.Info(fmt.Sprintf("Account imported: " + accountAddr))

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
	if favourableSet[0] < lowerBound || favourableSet[1] > upperBound {
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

func Bech32Decoder(value string) string {
	_, bytes, err := bech32.Decode(value)

	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error decoding Bech32 value: %v", err))
	}

	decodedBigInt := new(big.Int).SetBytes(bytes)
	return decodedBigInt.String()
}

func TXHashCheck(value string) string {
	byteSlice, err := hex.DecodeString(value)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error decoding hex value: %v", err))
	}

	decodedBigInt := new(big.Int).SetBytes(byteSlice)
	return decodedBigInt.String()
}
func AccountBalanceCheck(walletAddress string, blockHeight string, JsonAPI string) string {

	height, err := strconv.Atoi(blockHeight)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error converting block height to integer: %v", err))
	}

	res, resErr := http.Get(
		fmt.Sprintf(
			"%s/cosmos/bank/v1beta1/balances/%s?height=%d", JsonAPI, walletAddress, height-1,
		),
	)
	if resErr != nil {
		logs.Log.Error(fmt.Sprintf("Error making HTTP request: %v", resErr))
	}
	defer res.Body.Close()

	var accountBalance struct {
		Balances []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"balances"`
		Pagination struct {
			NextKey string `json:"next_key"`
			Total   string `json:"total"`
		} `json:"pagination"`
	}

	decodeError := json.NewDecoder(res.Body).Decode(&accountBalance)
	if decodeError != nil {
		logs.Log.Error(fmt.Sprintf("Error decoding JSON response: %v", decodeError))
	}

	return accountBalance.Balances[0].Amount
}

func AccountNounceCheck(walletAddress string, JsonAPI string) string {
	res, resErr := http.Get(
		fmt.Sprintf(
			"%s/cosmos/auth/v1beta1/accounts/%s", JsonAPI, walletAddress,
		),
	)
	if resErr != nil {
		logs.Log.Error(fmt.Sprintf("Error making HTTP request: %v", resErr))
	}
	defer res.Body.Close()

	var accountNounce struct {
		Account struct {
			Type    string `json:"@type"`
			Address string `json:"address"`
			PubKey  struct {
				Type string `json:"@type"`
				Key  string `json:"key"`
			} `json:"pub_key"`
			AccountNumber string `json:"account_number"`
			Sequence      string `json:"sequence"`
		} `json:"account"`
	}

	decodeError := json.NewDecoder(res.Body).Decode(&accountNounce)
	if decodeError != nil {
		logs.Log.Error(fmt.Sprintf("Error decoding JSON response: %v", decodeError))
	}

	return accountNounce.Account.Sequence
}
