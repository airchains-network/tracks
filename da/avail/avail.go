package avail

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/types"
	"io"
	"net/http"
)

func Avail(daData []byte, daRpc string) (string, error) {

	encodedString := base64.StdEncoding.EncodeToString(daData)
	//* Create the payload struct
	payload := map[string]interface{}{
		"data": encodedString,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("DA Body Unable to Marshell: %v", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v2/submit", daRpc), bytes.NewBuffer(payloadJSON))
	if err != nil {
		return "", fmt.Errorf("HTTP Req Error: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API call failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return "", fmt.Errorf("forbidden Method Status : %v, Call : %s", response.StatusCode, response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	var successResponse types.AvailSuccessResponse
	err = json.Unmarshal(body, &successResponse)
	if err != nil {
		return "", fmt.Errorf("error parsing response: %s", response.Body)
	}
	return successResponse.BlockHash, nil

}
