package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// AssetManagement struct
type AssetManagement struct{}

// Init method
func (t *AssetManagement) Init(stub shim.ChaincodeStubInterface) peer.Response {
	_, args := stub.GetFunctionAndParameters()
	if len(args) != 2 {
		return shim.Error("len(args) must be 2!")
	}
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error("Error while saving data...")
	}
	fmt.Println("Init success!")
	return shim.Success(nil)
}

// Invoke method
func (t *AssetManagement) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	var result string
	var err error
	if function == "set" {
		result, err = set(stub, args)
	} else if function == "get" {
		result, err = get(stub, args)
	} else {
		errMessage := fmt.Sprintf("undefined function %s,please use set or get", function)
		return shim.Error(errMessage)
	}
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("invoke success!")
	return shim.Success([]byte(result))
}

func set(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("len(args) must be 2 when use set function")
	}
	err := stub.PutState(args[0], []byte(args[1]))
	return "", err
}

func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("len(args) must be 1 when use get function")
	}
	result, err := stub.GetState(args[0])
	return string(result), err
}

func main() {
	err := shim.Start(new(AssetManagement))
	if err != nil {
		fmt.Printf("Error while starting AssetManagement chaincode : %v", err)
	}
}
