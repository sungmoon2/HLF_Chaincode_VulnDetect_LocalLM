package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type SimpleChaincode struct {
}

func (c *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	args := stub.GetStringArgs()
	if len(args) != 1 {
		return shim.Error("Incorrect arguments. Expecting a key and a value")
	}
	err := stub.PutState("key", []byte(args[0]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset"))
	}
	return shim.Success(nil)
}

func (c *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	fn, args := stub.GetFunctionAndParameters()
	var err error
	var result string

	if fn == "set" {
		result, err = set(stub, args)
	} else {
		result, err = get(stub)
	}
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(result))
}

func set(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}
	err := stub.PutState("key", []byte(args[0]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}

	return args[0], nil
}

func get(stub shim.ChaincodeStubInterface) (string, error) {

	value, err := stub.GetState("key")
	if err != nil {
		return "", fmt.Errorf("Failed to set asset")
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found")
	}
	return string(value), nil
}

func main() {
	if err := shim.Start(new(SimpleChaincode)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
