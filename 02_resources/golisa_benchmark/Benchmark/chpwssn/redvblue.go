package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

type Flag struct {
	Owner string `json:"owner"`
	Count int64  `json:"count"`
}

/*
 * The Init method is called when the Smart Contract "redvblue" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "redvblue"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "scorePoint" {
		return s.scorePoint(APIstub, args)
	} else if function == "queryAllFlags" {
		return s.queryAllFlags(APIstub)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	flags := []Flag{
		Flag{Owner: "Red", Count: 0},
		Flag{Owner: "Blue", Count: 0},
	}

	i := 0
	for i < len(flags) {
		fmt.Println("i is ", i)
		flagAsBytes, _ := json.Marshal(flags[i])
		APIstub.PutState("FLAG"+strconv.Itoa(i), flagAsBytes)
		fmt.Println("Added", flags[i])
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) queryAllFlags(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "FLAG0"
	endKey := "FLAG999"

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

	fmt.Printf("- queryAllFlags:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) scorePoint(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	flagAsBytes, _ := APIstub.GetState(args[0])
	flag := Flag{}

	json.Unmarshal(flagAsBytes, &flag)
	score, _ := strconv.ParseInt(args[1], 10, 32)
	flag.Count = flag.Count + score

	flagAsBytes, _ = json.Marshal(flag)
	APIstub.PutState(args[0], flagAsBytes)

	return shim.Success(nil)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
