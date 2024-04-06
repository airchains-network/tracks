package v1

import (
	"encoding/json"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"os"
)

// Deprecated

// CreateVkPkNew generates and saves a new Proving Key and Verification Key if either file doesn't exist
func CreateVkPkNew() {
	homeDir, _ := os.UserHomeDir()
	provingKeyFile := homeDir + "/.tracks/config/provingKey.txt"
	verificationKeyFile := homeDir + "/.tracks/config/verificationKey.json"

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
	homeDir, _ := os.UserHomeDir()
	provingKeyFile := homeDir + "/.tracks/config/provingKey.txt"
	verificationKeyFile := homeDir + "/.tracks/config/verificationKey.json"

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
