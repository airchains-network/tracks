package prover

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/decentralized-sequencer/blocksync"
	"github.com/airchains-network/decentralized-sequencer/config"
	"github.com/airchains-network/decentralized-sequencer/types"
	"math/rand"
	"os"
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	tedwards "github.com/consensys/gnark-crypto/ecc/twistededwards"
	"github.com/consensys/gnark-crypto/hash"
	cryptoEddsa "github.com/consensys/gnark-crypto/signature/eddsa"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark/std/algebra/native/twistededwards"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/signature/eddsa"
)

type MyCircuit struct {
	To              [config.PODSize]frontend.Variable `gnark:",public"`
	From            [config.PODSize]frontend.Variable `gnark:",public"`
	Amount          [config.PODSize]frontend.Variable `gnark:",public"`
	TransactionHash [config.PODSize]frontend.Variable `gnark:",public"`
	FromBalances    [config.PODSize]frontend.Variable `gnark:",public"`
	ToBalances      [config.PODSize]frontend.Variable `gnark:",public"`
	Messages        [config.PODSize]frontend.Variable `gnark:",public"`
	PublicKeys      [config.PODSize]eddsa.PublicKey   `gnark:",public"`
	Signatures      [config.PODSize]eddsa.Signature   `gnark:",public"`
}

func getTransactionHash(tx types.GetTransactionStruct) string {
	record := tx.To + tx.From + tx.Amount + tx.FromBalances + tx.ToBalances + tx.TransactionHash
	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

func GetMerkleRootCheck(transactions []types.GetTransactionStruct) string {
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

func GetMerkleRoot(api frontend.API, leaves [config.PODSize]frontend.Variable) frontend.Variable {

	if len(leaves) == 0 {
		return nil
	}

	// If there is only one input hash, return it
	if len(leaves) == 1 {
		return leaves[0]
	}

	var hashes []frontend.Variable
	for _, hashStr := range leaves {
		hashes = append(hashes, hashStr)
	}

	for len(hashes) > 1 {
		if len(hashes)%2 == 1 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}
		var nextLevel []frontend.Variable
		for i := 0; i < len(hashes); i += 2 {
			// combined := append(hashes[i], hashes[i+1]...)
			combined := api.Add(hashes[i], hashes[i+1])
			hash := MiMC7(api, 18, combined, combined)
			nextLevel = append(nextLevel, hash)
		}
		hashes = nextLevel
	}

	return hashes[0]
}

func (circuit *MyCircuit) Define(api frontend.API) error {
	var leaves [config.PODSize]frontend.Variable
	for i := 0; i < config.PODSize; i++ {

		//Signature Verification
		curve, err := twistededwards.NewEdCurve(api, tedwards.ID(ecc.BLS12_381))
		if err != nil {
			fmt.Println("Error creating a curve")
			return err
		}
		mimc, err := mimc.NewMiMC(api)
		if err != nil {
			return err
		}
		err = eddsa.Verify(curve, circuit.Signatures[i], circuit.Messages[i], circuit.PublicKeys[i], &mimc)
		if err != nil {
			fmt.Println("Error verifying signature")
			return err
		}
		// fmt.Println(i, ". Signature verified successfully!")
		transactionInputs := []frontend.Variable{
			circuit.To[i],
			circuit.From[i],
			circuit.Amount[i],
			circuit.FromBalances[i],
			circuit.ToBalances[i],
			circuit.TransactionHash[i],
		}

		TxLeaf := Poseidon(api, transactionInputs)
		leaves[i] = TxLeaf
		// Ensure sender's balance >= amount
		api.AssertIsLessOrEqual(circuit.Amount[i], circuit.FromBalances[i])

		// Deduct amount from sender and add to receiver within the circuit
		api.Sub(circuit.FromBalances[i], circuit.Amount[i])
		api.Add(circuit.ToBalances[i], circuit.Amount[i])

		updatedFromBalance := api.Sub(circuit.FromBalances[i], circuit.Amount[i])
		updatedToBalance := api.Add(circuit.ToBalances[i], circuit.Amount[i])

		// Ensure the updated balances are correct
		api.AssertIsEqual(updatedFromBalance, api.Sub(circuit.FromBalances[i], circuit.Amount[i]))
		api.AssertIsEqual(updatedToBalance, api.Add(circuit.ToBalances[i], circuit.Amount[i]))
	}

	_ = GetMerkleRoot(api, leaves)
	return nil
}

func ComputeCCS() constraint.ConstraintSystem {
	var circuit MyCircuit
	ccs, _ := frontend.Compile(ecc.BLS12_381.ScalarField(), r1cs.NewBuilder, &circuit)

	return ccs
}

func GenerateVerificationKey() (groth16.ProvingKey, groth16.VerifyingKey, error) {
	ccs := ComputeCCS()
	// groth16 zkSNARK: Setup
	pk, vk, error := groth16.Setup(ccs)
	return pk, vk, error
}

// GenerateProof generates a proof for the given input data
// and returns the proof and the error
// batchDbCount is the number of batches in the database and it will be passed as batchNum here
func GenerateProof(inputData types.BatchStruct, batchNum int) (any, string, []byte, error) {
	ccs := ComputeCCS()
	var transactions []types.GetTransactionStruct
	for i := 0; i < config.PODSize; i++ {
		transaction := types.GetTransactionStruct{
			To:              inputData.To[i],
			From:            inputData.From[i],
			Amount:          inputData.Amounts[i],
			FromBalances:    inputData.SenderBalances[i],
			ToBalances:      inputData.ReceiverBalances[i],
			TransactionHash: inputData.TransactionHash[i],
		}
		transactions = append(transactions, transaction)
	}
	currentStatusHash := GetMerkleRootCheck(transactions)
	homeDir, _ := os.UserHomeDir()
	provingKeyFile := homeDir + "/.tracks/config/provingKey.txt"
	pk, err := ReadProvingKeyFromFile(provingKeyFile)
	if err != nil {
		fmt.Println("Error reading proving key:", err)
		return nil, "", nil, err
	}
	if err != nil {
		fmt.Println("Error reading proving key:", err)
		return nil, "", nil, err
	}
	seed := time.Now().Unix()
	randomness := rand.New(rand.NewSource(seed))
	hFunc := hash.MIMC_BLS12_381.New()
	snarkField, err := twistededwards.GetSnarkField(tedwards.BLS12_381)
	if err != nil {
		fmt.Println("Error getting snark field")
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
		Signatures:      [config.PODSize]eddsa.Signature{},
		PublicKeys:      [config.PODSize]eddsa.PublicKey{},
		Messages:        [config.PODSize]frontend.Variable{},
	}

	for i := 0; i < config.PODSize; i++ {
		inputs.To[i] = frontend.Variable(inputData.To[i])
		inputs.From[i] = frontend.Variable(inputData.From[i])
		inputs.Amount[i] = frontend.Variable(inputData.Amounts[i])
		inputs.TransactionHash[i] = frontend.Variable(inputData.TransactionHash[i])
		inputs.FromBalances[i] = frontend.Variable(inputData.SenderBalances[i])
		inputs.ToBalances[i] = frontend.Variable(inputData.ReceiverBalances[i])
		// msg := []byte(inputData.Messages[i])
		msg := make([]byte, len(snarkField.Bytes()))

		inputs.Messages[i] = msg
		// create a eddsa key pair
		privateKey, err := cryptoEddsa.New(tedwards.ID(ecc.BLS12_381), randomness)
		if err != nil {
			fmt.Println("Not able to generate private keys")
		}
		publicKey := privateKey.Public()
		signature, err := privateKey.Sign(msg, hFunc)
		if err != nil {
			fmt.Println("Error signing the message")
			return nil, "", nil, err

		}
		// Public key
		_publicKey := publicKey.Bytes()

		inputs.PublicKeys[i].Assign(tedwards.BLS12_381, _publicKey[:32])
		inputs.Signatures[i].Assign(tedwards.BLS12_381, signature)
	}

	// witness definition
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
	// Open the file for reading
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	pk := groth16.NewProvingKey(ecc.BLS12_381)
	// Read the proving key from the file
	_, err = pk.ReadFrom(file)
	if err != nil {
		return nil, err
	}

	return pk, nil
}
