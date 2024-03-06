package prover

import (
	"encoding/json"
	"fmt"
	"github.com/airchains-network/evm-sequencer-node/common/logs"
	"os"
)

func CreateVkPk() {
	verificationKeyFile := "verificationKey.json"
	provingKeyFile := "provingKey.txt"

	if _, err := os.Stat(provingKeyFile); os.IsNotExist(err) {
		if _, err := os.Stat(verificationKeyFile); os.IsNotExist(err) {
			provingKey, verificationKey, err2 := GenerateVerificationKey()
			if err2 != nil {
				fmt.Println("Error generating verification key:", err2)
			}
			vkJSON, _ := json.Marshal(verificationKey)
			vkErr := os.WriteFile(verificationKeyFile, vkJSON, 0644)
			if vkErr != nil {
				fmt.Println("Error writing verification key to file:", vkErr)
			}
			file, err := os.Create(provingKeyFile)
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					fmt.Println("Error closing file:", err)
				}
			}(file)
			_, err = provingKey.WriteTo(file)
			if err != nil {
				fmt.Println("Error writing proving key to buffer:", err)
			}
		} else {
			return
		}
	}
	if _, err := os.Stat(verificationKeyFile); os.IsNotExist(err) {
		_, verificationKey, error := GenerateVerificationKey()
		if error != nil {
			fmt.Println("Error generating verification key:", error)
		}
		vkJSON, _ := json.Marshal(verificationKey)
		vkErr := os.WriteFile(verificationKeyFile, vkJSON, 0644)
		if vkErr != nil {
			fmt.Println("Error writing verification key to file:", vkErr)
		}
	} else {
		logs.Log.Info("Verification key already exists. No action needed.")
	}
	if _, err := os.Stat(provingKeyFile); os.IsNotExist(err) {
		provingKey, _, err2 := GenerateVerificationKey()
		if err2 != nil {
			fmt.Println("Error generating verification key:", err2)
		}
		file, err := os.Create(provingKeyFile)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				fmt.Println("Error closing file:", err)
			}
		}(file)
		_, err = provingKey.WriteTo(file)
		if err != nil {
			fmt.Println("Error writing proving key to buffer:", err)
		}
	} else {
		logs.Log.Info("Proving key already exists. No action needed.")
	}
}
