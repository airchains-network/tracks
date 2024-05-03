package celestia

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/types"
	"io"
	"net/http"
	"strconv"
)

const (
	namespaceVersion    = 0
	leadingZeroBytes    = 18
	userSpecifiedBytes  = 11
	totalNamespaceBytes = leadingZeroBytes + userSpecifiedBytes
)

func Celestia(daData []byte, daRpc string, rpcAUTH string) (string, error) {

	//namespace := GenerateNamespace()
	encodedDataString := base64.StdEncoding.EncodeToString(daData)

	//* Create the payload struct
	payload := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "blob.Submit",
		"params": []interface{}{
			[]interface{}{map[string]interface{}{
				"namespace":     "AAAAAAAAAAAAAAAAAAAAAAAAAICj+khUlIv2W7g=",
				"data":          encodedDataString,
				"share_version": 0,
				"commitment":    "AD5EzbG0/EMvpw0p8NIjMVnoCP4Bv6K+V6gjmwdXUKU=",
			}},
			0.05,
		},
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("DA Body Unable to Marshell: %v", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", daRpc, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return "", fmt.Errorf("HTTP Req Error: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", rpcAUTH))

	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API call failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode == 401 {
		return "", fmt.Errorf("unauthorized Token")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	fmt.Println("Raw API Response:", string(body)) // Log the raw response

	// Try parsing as SuccessResponse first
	var successResponse types.CelestiaSuccessResponse
	err = json.Unmarshal(body, &successResponse)
	if err != nil {
		fmt.Println("Error unmarshalling SuccessResponse:", err) // Detailed error
		var errorResponse types.CelestiaErrorResponse
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			fmt.Println("Error unmarshalling ErrorResponse:", err) // Detailed error
			return "", fmt.Errorf("failed to unmarshal response: %v", err)
		} else {
			return "", fmt.Errorf(errorResponse.Error.Message)
		}
	} else {
		return strconv.Itoa(successResponse.Result), nil
	}
}
