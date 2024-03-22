package v1

import (
	"encoding/json"
	"fmt"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"os"
)

// Deprecated
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

// CreateVkPkNew generates and saves a new Proving Key and Verification Key if either file doesn't exist
func CreateVkPkNew() {
	provingKeyFile := "provingKey.txt"
	verificationKeyFile := "verificationKey.json"

	_, err1 := os.Stat(provingKeyFile)
	_, err2 := os.Stat(verificationKeyFile)

	// If either file doesn't exist, generate and save new keys
	if os.IsNotExist(err1) || os.IsNotExist(err2) {
		provingKey, verificationKey, err := GenerateVerificationKey()
		if err != nil {
			return
		}

		// Save Proving Key
		pkFile, err := os.Create(provingKeyFile)
		if err != nil {
			logs.Log.Error("Unable to create Proving Key file" + err.Error())
			return
		}
		_, err = provingKey.WriteTo(pkFile)
		pkFile.Close()
		if err != nil {
			logs.Log.Error("Unable to write Proving Key" + err.Error())
			return
		}

		// Save Verification Key
		file, _ := json.MarshalIndent(verificationKey, "", " ")
		err = os.WriteFile(verificationKeyFile, file, 0644)
		if err != nil {
			logs.Log.Error("Unable to write Verification Key to file" + err.Error())
		}
		logs.Log.Info("Proving key and Verification key generated and saved successfully\n")
	} else {
		logs.Log.Info("Both Proving key and Verification key already exist. No action needed.")
	}
}

func GetVkPk() (groth16.ProvingKey, groth16.VerifyingKey, error) {
	provingKeyFile := "provingKey.txt"
	verificationKeyFile := "verificationKey.json"

	// Read Proving Key
	pk, err := ReadProvingKeyFromFile2(provingKeyFile)
	if err != nil {
		logs.Log.Error("Failed to read Proving Key")
		return nil, nil, err
	}

	vk, err := ReadVerificationKeyFromFile(verificationKeyFile)
	if err != nil {
		logs.Log.Error("Failed to read Verification Key")
		return nil, nil, err
	}

	return pk, vk, nil
}

func ReadProvingKeyFromFile2(filename string) (groth16.ProvingKey, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	pk := groth16.NewProvingKey(ecc.BLS12_381)
	_, err = pk.ReadFrom(file)
	if err != nil {
		return nil, err
	}

	return pk, nil
}
func ReadVerificationKeyFromFile(filename string) (groth16.VerifyingKey, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	vk := groth16.NewVerifyingKey(ecc.BLS12_381)
	err = json.Unmarshal(file, vk)
	if err != nil {
		return nil, err
	}

	return vk, nil
}

//func ReadVerificationKeyFromFile(filename string) (groth16.VerifyingKey, error) {
//	file, err := os.Open(filename)
//	if err != nil {
//		return nil, err
//	}
//	defer file.Close()
//
//	vk := groth16.NewVerifyingKey(ecc.BLS12_381)
//	_, err = vk.ReadFrom(file)
//	if err != nil {
//		return nil, err
//	}
//
//	return vk, nil
//}
