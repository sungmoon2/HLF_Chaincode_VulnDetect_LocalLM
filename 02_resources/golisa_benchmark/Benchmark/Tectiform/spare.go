/*
 * Copyright Tectiform Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// Spare implements a simple chaincode to manage an asset 
type Spare struct {
}

// Part is key asset
type Part struct {
	Name    string
	Barcode string
	Comment string
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *Spare) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Get the args from the transaction proposal
	//args := stub.GetStringArgs()
	//if len(args) != 2 {
	//	return shim.Error("Incorrect arguments. Expecting a key and a value")
	//}

	// Set up any variables or assets here by calling stub.PutState()

	// We store the key and the value on the ledger
	//err := stub.PutState(args[0], []byte(args[1]))
	//if err != nil {
	//	return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
	return shim.Success(nil)
}

func (t *Spare) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error
	if fn == "addPartRecord" {
		result, err = addPartRecord(stub, args)
	} else if fn == "getPartRecord" {
		result, err = getPartRecord(stub, args)
	} else {
		return shim.Error("Transaction not supported") //result, err = get(stub, args)
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(result))
}

func addPartRecord(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 4 {
		return "", fmt.Errorf("Incorrect arguments. Expecting Key, Name, Barcode, Comment")
	}

	var part = Part{Name: args[1], Barcode: args[2], Comment: args[3]}
	partAsBytes, _ := json.Marshal(part)
	err := stub.PutState(args[0], partAsBytes)
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	return args[0], nil
}

func getPartRecord(stub shim.ChaincodeStubInterface, args []string) (string, error) {
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

func main() {
	if err := shim.Start(new(Spare)); err != nil {
		fmt.Printf("Error starting Spare chaincode: %s", err)
	}
}
