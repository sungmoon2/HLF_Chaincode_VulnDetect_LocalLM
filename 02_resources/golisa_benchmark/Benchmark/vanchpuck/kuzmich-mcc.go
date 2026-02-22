package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"strconv"
	"strings"
)

const (
	Shipped  string = "SHIPPED"
	Received string = "RECEIVED"
)

type ChaincodeExercise struct {
}

type Asset struct {
	Title  string `json:"title"`
	Count  int    `json:"count"`
	Status string `json:"status"`
}

func (s *ChaincodeExercise) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (t *ChaincodeExercise) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	if function == "sendAsset" {
		return t.sendAsset(stub, args)
	}
	if function == "queryAsset" {
                return t.queryAsset(stub, args)
        }
//        if function == "getAsset" {
//                return t.getAsset(stub, args)
//        }
	return shim.Error("Invalid Smart Contract function name.")
}

func (t *ChaincodeExercise) sendAsset(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	title := strings.ToLower(args[0])
	count, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Asset count must be a numeric string")
	}
	asset := &Asset{title, count, Shipped}
	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(title, assetBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func getAsset(stub shim.ChaincodeStubInterface, args []string) (string, error) {
        if len(args) != 1 {
                return "", fmt.Errorf("Incorrect arguments. Expecting a key")
        }

        value, err := stub.GetState(args[0])
        if err != nil {
                return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
        }
        if value == nil {
                return "", fmt.Errorf("Asset not found: %s", args[0])
        }
        return string(value), nil
}

func (t *ChaincodeExercise) queryAsset(stub shim.ChaincodeStubInterface, args []string) peer.Response {

        if len(args) != 1 {
                return shim.Error("Incorrect number of arguments. Expecting 1")
        }

        assetAsBytes, _ := stub.GetState(args[0])
        return shim.Success(assetAsBytes)
}

func main() {
	if err := shim.Start(new(ChaincodeExercise)); err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
