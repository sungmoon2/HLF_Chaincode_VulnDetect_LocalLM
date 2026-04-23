package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// Smart contract - Chaincode structure definition
// ===============================================
type EthDataLoaderChaincode struct {
}

type ethData struct {
	ObjectType string `json:"docType"`  // .
	Eth_id     string `json:"eth_id"`   // ethData index
	From       string `json:"from"`     // address of the sender
	Gas        string `json:"gas"`      // gas provided by the sender
	GasPrice   string `json:"gasPrice"` // gas price provided by the sender in Wei
	Hash       string `json:"hash"`     // hash of the transaction
	To         string `json:"to"`       // address of the receiver. null when its a contract creation transaction
	Value      string `value"`          // value transferred in Wei
}

// =============================================================================
// 			       Smart Contract/Chaincode Main Function
// =============================================================================

func main() {
	err := shim.Start(new(EthDataLoaderChaincode))
	if err != nil {
		fmt.Printf("Error starting ethDataLoader chaincode: %s", err)
	}
}

// Init - initializes chaincode
// =============================

func (t *EthDataLoaderChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println(" EtheDataloader Chaincode Initilized.")
	return shim.Success(nil)
}

// Invoke - the entry point for function invocations
// =====================================================

func (t *EthDataLoaderChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Invoke method gets called ")
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" { // create an eth data instance
		return t.invoke(stub, args)
	} else if function == "query" { //read an eth data instance
		return t.query(stub, args)
	}

	fmt.Println("invoke did not find function: " + function)
	return shim.Error("Unknown function name. Expecting \"invoke\"\"query\"")
}

// ===================================================================
// invoke - set an EthTransaction Data instance from chaincode state
// ===================================================================
func (t *EthDataLoaderChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("- start init EthTranction Data")

	// Eth_id:"0001", From: "0x122", Gas: "1000", GasPrice: "9000", hash: "0xakbkb1245", To: "0x234", Value: "9000000"
	// "0001", "0x122", "1000",  "9000", "0xakbkb1245", "0x234", "9000000"
	// "0002", "0x122", "1000",  "9000", "0xakbkb1245", "0x234", "9000000"

	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	eth_id := args[0]
	from := args[1]
	gas := args[2]
	gasPrice := args[3]
	hash := args[4]
	to := args[5]
	value := args[6]

	objectType := "ethData"
	ethData := &ethData{objectType, eth_id, from, gas, gasPrice, hash, to, value}
	ethDataJSONasBytes, err := json.Marshal(ethData)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(eth_id, ethDataJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end init ethTransaction data")
	return shim.Success(nil)
}

// ==============================================================
// query - read an Ethereum Data instance from chaincode state
// ==============================================================

func (t *EthDataLoaderChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var eth_id, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting only from address to query")
	}

	eth_id = args[0]
	valAsBytes, err := stub.GetState(eth_id)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state of: " + eth_id + "\"}"
		return shim.Error(jsonResp)
	} else if valAsBytes == nil {
		jsonResp = "{\"Error: \" \"The following eth data instance does not exist: " + eth_id + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valAsBytes)
}
