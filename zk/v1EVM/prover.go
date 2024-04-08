package v1EVM

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/blocksync"
	"github.com/airchains-network/decentralized-sequencer/config"
	logs "github.com/airchains-network/decentralized-sequencer/log"
	"github.com/airchains-network/decentralized-sequencer/types"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"os"
)

type MyCircuit struct {
	To              [config.PODSize]frontend.Variable `gnark:",public"`
	From            [config.PODSize]frontend.Variable `gnark:",public"`
	Amount          [config.PODSize]frontend.Variable `gnark:",public"`
	TransactionHash [config.PODSize]frontend.Variable `gnark:",public"`
	FromBalances    [config.PODSize]frontend.Variable `gnark:",public"`
	ToBalances      [config.PODSize]frontend.Variable `gnark:",public"`
}

type TransactionSecond struct {
	To              string
	From            string
	Amount          string
	FromBalances    string
	ToBalances      string
	TransactionHash string
}

func getTransactionHash(tx TransactionSecond) string {

	h1 := sha256.Sum256([]byte(tx.To))
	h2 := sha256.Sum256([]byte(tx.From))
	h3 := sha256.Sum256([]byte(tx.Amount))
	h4 := sha256.Sum256([]byte(tx.FromBalances))
	h5 := sha256.Sum256([]byte(tx.ToBalances))
	h6 := sha256.Sum256([]byte(tx.TransactionHash))
	h := sha256.New()
	h.Write(h1[:])
	h.Write(h2[:])
	h.Write(h3[:])
	h.Write(h4[:])
	h.Write(h5[:])
	h.Write(h6[:])

	return hex.EncodeToString(h.Sum(nil))
}
func GetMerkleRootSecond(transactions []TransactionSecond) string {
	var merkleTree []string

	for _, tx := range transactions {
		merkleTree = append(merkleTree, getTransactionHash(tx))
	}

	for len(merkleTree) > 1 {
		var tempTree []string
		for i := 0; i < len(merkleTree); i += 2 {
			if i+1 == len(merkleTree) {
				tempTree = append(tempTree, merkleTree[i])
			} else {
				combinedHash := merkleTree[i] + merkleTree[i+1]
				h := sha256.New()
				h.Write([]byte(combinedHash))
				tempTree = append(tempTree, hex.EncodeToString(h.Sum(nil)))
			}
		}
		merkleTree = tempTree
	}

	return merkleTree[0]
}

func (circuit *MyCircuit) Define(api frontend.API) error {
	for i := 0; i < config.PODSize; i++ {
		api.AssertIsLessOrEqual(circuit.Amount[i], circuit.FromBalances[i]) //TODO  Here is one error1

		api.Sub(circuit.FromBalances[i], circuit.Amount[i])
		api.Add(circuit.ToBalances[i], circuit.Amount[i])

		updatedFromBalance := api.Sub(circuit.FromBalances[i], circuit.Amount[i])
		updatedToBalance := api.Add(circuit.ToBalances[i], circuit.Amount[i])

		api.AssertIsEqual(updatedFromBalance, api.Sub(circuit.FromBalances[i], circuit.Amount[i]))
		api.AssertIsEqual(updatedToBalance, api.Add(circuit.ToBalances[i], circuit.Amount[i]))
	}

	return nil
}

func ComputeCCS() constraint.ConstraintSystem {
	var circuit MyCircuit
	ccs, _ := frontend.Compile(ecc.BLS12_381.ScalarField(), r1cs.NewBuilder, &circuit)

	return ccs
}

func GenerateVerificationKey() (groth16.ProvingKey, groth16.VerifyingKey, error) {
	ccs := ComputeCCS()
	pk, vk, error := groth16.Setup(ccs)
	return pk, vk, error
}

func GenerateProof(inputData types.BatchStruct, batchNum int) (any, string, []byte, error) {
	ccs := ComputeCCS()
	logs.Log.Info("Generating proof for batch number:" + fmt.Sprintf("%d", batchNum))
	var transactions []TransactionSecond

	for i := 0; i < config.PODSize; i++ {

		transaction := TransactionSecond{
			To:              inputData.To[i],
			From:            inputData.From[i],
			Amount:          inputData.Amounts[i],
			FromBalances:    inputData.SenderBalances[i],
			ToBalances:      inputData.ReceiverBalances[i],
			TransactionHash: inputData.TransactionHash[i],
		}
		transactions = append(transactions, transaction)
	}

	currentStatusHash := GetMerkleRootSecond(transactions)

	if _, err := os.Stat("provingKey.txt"); os.IsNotExist(err) {
		fmt.Println("Proving key does not exist. Please run the command 'sequencer-sdk create-vk-pk' to generate the proving key")
		return nil, "", nil, err
	}

	//pk, err := ReadProvingKeyFromFile("provingKey.txt")
	homeDir, _ := os.UserHomeDir()
	provingKeyFile := homeDir + "/.tracks/config/provingKey.txt"
	pk, err := ReadProvingKeyFromFile(provingKeyFile)

	if err != nil {
		fmt.Println("Error reading proving key:", err)
		return nil, "", nil, err
	}

	var inputValueLength int

	fromLength := len(inputData.From)
	toLength := len(inputData.To)
	amountsLength := len(inputData.Amounts)
	txHashLength := len(inputData.TransactionHash)
	senderBalancesLength := len(inputData.SenderBalances)
	receiverBalancesLength := len(inputData.ReceiverBalances)
	messagesLength := len(inputData.Messages)
	txNoncesLength := len(inputData.TransactionNonces)
	accountNoncesLength := len(inputData.AccountNonces)

	if fromLength == toLength &&
		fromLength == amountsLength &&
		fromLength == txHashLength &&
		fromLength == senderBalancesLength &&
		fromLength == receiverBalancesLength &&
		fromLength == messagesLength &&
		fromLength == txNoncesLength &&
		fromLength == accountNoncesLength {
		inputValueLength = fromLength
	} else {
		fmt.Println("Error: Input data is not correct")
		return nil, "", nil, fmt.Errorf("input data is not correct")
	}

	if inputValueLength < config.PODSize {
		leftOver := config.PODSize - inputValueLength
		for i := 0; i < leftOver; i++ {
			inputData.From = append(inputData.From, "0")
			inputData.To = append(inputData.To, "0")
			inputData.Amounts = append(inputData.Amounts, "0")
			inputData.TransactionHash = append(inputData.TransactionHash, "0")
			inputData.SenderBalances = append(inputData.SenderBalances, "0")
			inputData.ReceiverBalances = append(inputData.ReceiverBalances, "0")
			inputData.Messages = append(inputData.Messages, "0")
			inputData.TransactionNonces = append(inputData.TransactionNonces, "0")
			inputData.AccountNonces = append(inputData.AccountNonces, "0")
		}
	}

	inputs := MyCircuit{
		To:              [config.PODSize]frontend.Variable{},
		From:            [config.PODSize]frontend.Variable{},
		Amount:          [config.PODSize]frontend.Variable{},
		TransactionHash: [config.PODSize]frontend.Variable{},
		FromBalances:    [config.PODSize]frontend.Variable{},
		ToBalances:      [config.PODSize]frontend.Variable{},
	}

	for i := 0; i < config.PODSize; i++ {
		inputs.To[i] = frontend.Variable(inputData.To[i])
		inputs.From[i] = frontend.Variable(inputData.From[i])
		inputs.Amount[i] = frontend.Variable(inputData.Amounts[i])
		inputs.TransactionHash[i] = frontend.Variable(inputData.TransactionHash[i])
		inputs.FromBalances[i] = frontend.Variable(inputData.SenderBalances[i])
		inputs.ToBalances[i] = frontend.Variable(inputData.ReceiverBalances[i])
	}

	witness, err := frontend.NewWitness(&inputs, ecc.BLS12_381.ScalarField())
	if err != nil {
		fmt.Printf("Error creating a witness: %v\n", err)
		return nil, "", nil, err
	}

	witnessVector := witness.Vector()

	publicWitness, _ := witness.Public()

	publicWitnessDb := blocksync.GetPublicWitnessDbInstance()
	publicWitnessDbKey := fmt.Sprintf("public_witness_%d", batchNum)
	publicWitnessDbValue, err := json.Marshal(publicWitness)
	if err != nil {
		fmt.Println("Error marshalling public witness:", err)
		return nil, "", nil, err
	}
	err = publicWitnessDb.Put([]byte(publicWitnessDbKey), publicWitnessDbValue, nil)
	if err != nil {
		fmt.Println("Error saving public witness:", err)
		return nil, "", nil, err
	}
	proof, err := groth16.Prove(ccs, pk, witness)
	if err != nil {
		fmt.Printf("Error generating proof: %v\n", err)
		return nil, "", nil, err
	}

	proofDb := blocksync.GetProofDbInstance()
	proofDbKey := fmt.Sprintf("proof_%d", batchNum)
	proofDbValue, err := json.Marshal(proof)
	if err != nil {
		fmt.Println("Error marshalling proof:", err)
		return nil, "", nil, err
	}
	err = proofDb.Put([]byte(proofDbKey), proofDbValue, nil)
	if err != nil {
		fmt.Println("Error saving proof:", err)
		return nil, "", nil, err
	}

	return witnessVector, currentStatusHash, proofDbValue, nil
}

func ReadProvingKeyFromFile(filename string) (groth16.ProvingKey, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	pk := groth16.NewProvingKey(ecc.BLS12_381)
	_, err = pk.ReadFrom(file)
	if err != nil {
		return nil, err
	}

	return pk, nil
}
