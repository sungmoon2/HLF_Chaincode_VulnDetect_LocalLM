package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"crypto/sha512"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

type Hesh struct {
	HashS   string `json:"hash"`
}
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "hashChainCode"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "addHash" {
		return s.addHash (APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "getDatabase" {
		return s.getDatabase(APIstub)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	hashes := []Hesh{
		Hesh{HashS: fmt.Sprintf("%x", sha512.Sum512_256([]byte("Ryan")))},
		Hesh{HashS: fmt.Sprintf("%x", sha512.Sum512_256([]byte("Michael")))},
		Hesh{HashS: fmt.Sprintf("%x", sha512.Sum512_256([]byte("Raj")))},
		Hesh{HashS: fmt.Sprintf("%x", sha512.Sum512_256([]byte("NSA")))},
	}

	i := 0
	for i < len(hashes) {
		fmt.Println("i is ", i)
		hashAsBytes, _ := json.Marshal(hashes[i])
		APIstub.PutState("HASH"+strconv.Itoa(i), hashAsBytes)
		fmt.Println("Added", hashes[i])
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) addHash(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	temp := args[1]

	var hesh = Hesh{HashS: fmt.Sprintf("%x", sha512.Sum512_256([]byte(temp)))}

	hashAsBytes, _ := json.Marshal(hesh)
	APIstub.PutState(args[0], hashAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) getDatabase(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "HASH0"
	endKey := "HASH99999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getDatabase:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}


// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}

