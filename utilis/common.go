package utilis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"io"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

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

func GetBalance(address string, blockNumber uint64) (string, error) {
	payload := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"method": "eth_getBalance",
		"params": ["%s", "%s"],
		"id": 1
	}`, address, "0x"+strconv.FormatUint(blockNumber, 16))

	resp, err := http.Post("locahost:8545", "application/json", strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return "", err
	}

	if errMsg, ok := jsonResponse["error"]; ok {
		return "", fmt.Errorf("error from Ethereum node: %v", errMsg)
	}

	if result, ok := jsonResponse["result"].(string); ok {
		balance, success := new(big.Int).SetString(result[2:], 16)
		if !success {
			return "", fmt.Errorf("failed to parse balance")
		}

		return balance.String(), nil
		//etherBalance := new(big.Float).Quo(new(big.Float).SetInt(balance), new(big.Float).SetInt64(1e18))
		//formattedBalance := fmt.Sprintf("%.18f", etherBalance) // Format the balance
		//return etherBalance, nil
	} else {
		log.Fatal("Failed to parse balance")
		return "", err
	}
}

func GetAccountNonce(ctx context.Context, address string, blockNumber uint64) (string, error) {
	client, err := rpc.Dial("locahost:8545")
	if err != nil {
		return "0", err
	}
	accountAddress := common.HexToAddress(address)
	formatedBlockNumber := "0x" + strconv.FormatUint(blockNumber, 16)
	var result string
	err = client.CallContext(ctx, &result, "eth_getTransactionCount", accountAddress, formatedBlockNumber)
	if err != nil {
		fmt.Println("Error getting transaction count:", err)
		return "0", err
	}

	return result, nil
}
