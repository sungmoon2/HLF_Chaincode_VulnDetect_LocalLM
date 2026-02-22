//package simpleAsset
package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// AutoTraceChaincode example simple Chaincode implementation
type SimpleAsset struct {
}

// Init initializes chaincode
// ===========================
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) pb.Response {

	return shim.Success(nil)
}

func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error

	if fn == "set" {
		result, err = set(stub, args)
	} else {
		result, err = get(stub, args)
	}

	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(result))
}

func set(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("SET: Incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset %s", args[0])
	}
	return args[1], nil
}

func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("GET: Incorrect arguments. Expecting a key")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get asset %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found %s", args[0])
	}
	return string(value), nil
}

func main() {
	err := shim.Start(new(SimpleAsset))
	if err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
