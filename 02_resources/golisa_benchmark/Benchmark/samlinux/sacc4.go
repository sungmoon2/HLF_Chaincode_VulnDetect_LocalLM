/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"bytes"
	"fmt"
  "time"
  "strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// SimpleAsset implements a simple chaincode to manage an asset
type SimpleAsset struct {
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Get the args from the transaction proposal
	args := stub.GetStringArgs()
	if len(args) != 2 {
		return shim.Error("Incorrect arguments. Expecting a key and a value")
	}

	// Set up any variables or assets here by calling stub.PutState()

	// We store the key and the value on the ledger
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
	}
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()
	//fmt.Println("invoke is running " + fn)

	if fn == "set" {
		return set(stub, args)
	} else if fn == "all" {
		return getAll(stub)
	} else if  fn == "history" {
		return getAllTx(stub, args)
	} else { // assume 'get' even if fn is nil
		return get(stub, args)
	}
}

// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
func set(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error("Failed to set asset: %s" + args[0])
	}

	// set chaincode event
	eventPayload := "A new value was set: " + args[0]
	payloadAsBytes := []byte(eventPayload)
	eventErr := stub.SetEvent("setEvent", payloadAsBytes)
	if eventErr != nil {
		return shim.Error("Failed to emit set event")
	}
	return shim.Success(nil)
}

// Get returns the value of the specified asset key
func get(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect arguments. Expecting a key")
	}

	valAsbytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get asset: %s with error: " + args[0])
	}
	if valAsbytes == nil {
		return shim.Error("Asset not found: " + args[0])
	}
	return shim.Success(valAsbytes)
}

// =========================================================================================
// GetAll returns all keys
// =========================================================================================
func getAll(stub shim.ChaincodeStubInterface) peer.Response {

	// this is my query string
	queryString := fmt.Sprintf("{\"selector\": {\"_id\": {\"$gt\": null}},\"fields\": [\"_id\"]}")

	queryResults, err := getQueryResultForQueryString(stub, queryString)

	if err != nil {
		return shim.Error("Failed to query")
	}

	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	//fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	//fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

// ===========================================================================================
// constructQueryResponseFromIterator constructs a JSON array containing query results from
// a given result iterator
// ===========================================================================================
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		//buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		//buffer.WriteString(string(queryResponse.Value))

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return &buffer, nil
}

// =========================================================================================
// getKeyHistory returns the history to a given key
// =========================================================================================
func getAllTx (stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	keyName := args[0]

	//fmt.Printf("- start getKeyHistory: %s\n", keyName)

	resultsIterator, err := stub.GetHistoryForKey(keyName)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		
		buffer.WriteString("\"")
		if response.IsDelete {
			buffer.WriteString("null")
		} else {	
			buffer.WriteString(string(response.Value))
		}
		buffer.WriteString("\"")
		
		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	//fmt.Printf("- getKeyHistory returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}
