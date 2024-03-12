package utilis

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestGetBalance(t *testing.T) {
	address := "0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693" //0xBdb56Cf303763cBAC0F610D5ec6B811Aa5d91693
	blockNumber := uint64(1100)
	payload := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"method": "eth_getBalance",
		"params": ["%s", "%s"],
		"id": 1
	}`, address, "0x"+strconv.FormatUint(blockNumber, 16))

	resp, err := http.Post("http://192.168.1.106:8545", "application/json", strings.NewReader(payload))
	if err != nil {
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
	}

	var jsonResponse map[string]interface{}
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
	}

	//if errMsg, ok := jsonResponse["error"]; ok {
	//
	//}

	if result, ok := jsonResponse["result"].(string); ok {
		balance, success := new(big.Int).SetString(result[2:], 16)
		if !success {
		}
		t.Log(balance)
		fmt.Printf("%T\n", balance)
		l := balance.String()
		fmt.Printf("%T\n", l)
		t.Log(l)

		choot := fmt.Sprintf("%d", balance)
		t.Log(choot)
	} else {
	}
}
