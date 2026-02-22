/**
 *  Xooa Get Set Delete blockchain Logger
 *
 *  Copyright 2018 Xooa
 *
 *  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at:
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software distributed under the License is distributed
 *  on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License
 *  for the specific language governing permissions and limitations under the License.
 */
/*
 * Original source via IBM Corp:
 *  https://hyperledger-fabric.readthedocs.io/en/release-1.2/chaincode4ade.html#pulling-it-all-together
 *
 * Modifications from: Kavi Sarna:
 *  https://github.com/xooa/
 *
 * Changes:
 *  Logs to Xooa blockchain platform
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type Result struct {
	Payload   []Payload   `json:"result"`
	ErrorKeys []ErrorKeys `json:"errors"`
}

type Payload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ErrorKeys struct {
	Key   string `json:"key"`
	Error string `json:"error"`
}

var logger = shim.NewLogger("GetSetDelCC")

// SimpleAsset implements a simple chaincode to manage an asset
type SimpleAsset struct {
}

// Init is called during chaincode instantiation to initialize any data.
// Note that chaincode upgrade also calls this function to reset or to migrate data.
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either updating the state or retreiving the state created by Init function.
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	function, args := stub.GetFunctionAndParameters()

	if function == "set" {
		return t.set(stub, args)

	} else if function == "get" {
		return t.get(stub, args)

	} else if function == "del" {
		return t.del(stub, args)

	} else if function == "getHistory" {
		return t.getHistory(stub, args)

	}

	logger.Error("Function declaration not found for ", function)

	response := shim.Error("Invalid function name " + function + " for 'Invoke'.")
	response.Status = 404
	return response
}

// get queries using key. It retrieves the latest state(value) of the key.
func (t *SimpleAsset) get(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	logger.Debug("get() called.")
	logger.Debug(args)

	//Function call with no arguments will return an error response and will not commit anything to ledger
	if len(args) == 0 {

		logger.Debug("Get method invoked without arguments. Returning an error response.")
		response := shim.Error("No arguments were available. Nothing was committed to ledger.")
		response.Status = 400
		return response
	}

	// Function call with argument "Xooa Test" is used for testing function names.
	// We return success wihout committing the call to the ledger
	if len(args) == 1 && args[0] == "Xooa Test" {

		logger.Debug("Method test call with argument 'Xooa Test'. Nothing will be committed to ledger.")
		response := shim.Success([]byte("Method test call. Nothing will be committed to ledger."))
		response.Status = 200
		return response
	}

	var payload = []Payload{}
	var errors = []ErrorKeys{}

	for i := 0; i < len(args); i++ {

		key := args[i]

		logger.Debug(key)

		// Get value form the ledger
		valuesAsBytes, err := stub.GetState(key)

		if err != nil {
			logger.Error("Error occured while calling GetState() method of shim - ", err.Error())
			errors = append(errors, ErrorKeys{Key: key, Error: err.Error()})
		}

		// if len(args) == 1 && valuesAsBytes == nil {

		// 	logger.Error("No value found for key", key)
		// 	response := shim.Error("No value found for key " + key)
		// 	response.Status = 404
		// 	return response
		// }

		if valuesAsBytes == nil {

			errors = append(errors, ErrorKeys{Key: key, Error: "404 - No value found for key."})

		} else {

			payload = append(payload, Payload{Key: key, Value: string(valuesAsBytes)})
		}
	}

	var result = Result{}
	result.Payload = payload
	result.ErrorKeys = errors

	resultJson, err := json.Marshal(result)

	if err != nil {

		logger.Error("Error occured while marshalling payload to json - ", err.Error())
		response := shim.Error("Error occured while marshalling payload to json - " + err.Error())
		response.Status = 400
		return response
	}

	response := shim.Success(resultJson)
	response.Status = 200

	return response
}

// set stores the event on the ledger.
// For each key, it will override the current state with the new one.
func (t *SimpleAsset) set(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	logger.Debug("set() called.")
	logger.Debug(args)

	//Function call with no arguments will return an error response and will not commit anything to ledger.
	if len(args) == 0 {

		logger.Debug("Set method invoked without arguments. Returning an error response.")
		response := shim.Error("No arguments were available. Nothing was committed to ledger.")
		response.Status = 400
		return response
	}

	// Function call with argument "Xooa Test" is used for testing function names entered.
	// We return success wihout committing the call to the ledger.
	if len(args) == 1 && args[0] == "Xooa Test" {

		logger.Debug("Method test call with argument 'Xooa Test'. Nothing will be committed to ledger.")
		response := shim.Success([]byte("Method test call. Nothing will be committed to ledger."))
		response.Status = 200
		return response
	}

	var payload = []Payload{}
	var errors = []ErrorKeys{}

	for i := 0; i < len(args); i += 2 {

		value := ""

		if i+1 < len(args) {

			value = args[i+1]
		}

		key := args[i]
		valuesAsBytes := []byte(value)

		logger.Debug(key)
		logger.Debug(value)

		if value == "" || valuesAsBytes == nil {

			logger.Debug("Not storing anything as no value for key.")
			errors = append(errors, ErrorKeys{Key: key, Error: "No value for key."})

		} else {

			err := stub.PutState(key, valuesAsBytes)

			if err != nil {

				logger.Error("Error occured while calling PutState() - ", err.Error())
				errors = append(errors, ErrorKeys{Key: key, Error: err.Error()})

			} else {

				payload = append(payload, Payload{Key: key, Value: string(valuesAsBytes)})
			}
		}
	}

	var result = Result{}
	result.Payload = payload
	result.ErrorKeys = errors

	resultJson, err := json.Marshal(result)

	if err != nil {

		logger.Error("Error occured while marshalling payload to json - ", err.Error())
		response := shim.Error("Error occured while marshalling payload to json - " + err.Error())
		response.Status = 400
		return response
	}

	response := shim.Success([]byte(resultJson))
	response.Status = 200
	return response
}

// del deletes the keys from the ledger.
// For each key, it will mark them as deleted and remove their state.
func (t *SimpleAsset) del(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	logger.Debug("del() called.")
	logger.Debug(args)

	//Function call with no arguments will return an error response and will not commit anything to ledger.
	if len(args) == 0 {

		logger.Debug("Del method invoked without arguments. Returning an error response.")
		response := shim.Error("No arguments were available. Nothing was committed to ledger.")
		response.Status = 400
		return response
	}

	// Function call with argument "Xooa Test" is used for testing function names.
	// We return success wihout committing the call to the ledger.
	if len(args) == 1 && args[0] == "Xooa Test" {

		logger.Debug("Method test call with argument 'Xooa Test'. Nothing will be committed to ledger.")
		response := shim.Success([]byte("Method test call. Nothing will be committed to ledger."))
		response.Status = 200
		return response
	}

	var payload = []string{}

	for i := 0; i < len(args); i++ {

		key := args[i]

		logger.Debug(key)

		err := stub.DelState(key)

		if err != nil {

			logger.Error("Error occured while calling DelState() - ", err.Error())
			response := shim.Error("Error occured while calling DelState() - " + err.Error())
			response.Status = 400
			return response
		}

		payload = append(payload, key)
	}

	result, err := json.Marshal(payload)

	if err != nil {

		logger.Error("Error occured while marshalling payload to json - ", err.Error())
		response := shim.Error("Error occured while marshalling payload to json - " + err.Error())
		response.Status = 400
		return response
	}

	response := shim.Success([]byte(result))
	response.Status = 200
	return response
}

// getHistory queries the entity using the key.
// It retrieves all the changes to the entity happened over time.
func (t *SimpleAsset) getHistory(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	logger.Debug("getHistory() called.")
	logger.Debug(args)

	//Function call with no arguments will return an error response and will not commit anything to ledger.
	if len(args) == 0 {

		logger.Debug("getHistory method invoked without arguments. Returning an error response.")
		response := shim.Error("No arguments were available. Nothing was committed to ledger.")
		response.Status = 400
		return response
	}

	// Function call with argument "Xooa Test" is used for testing function names.
	// We return success wihout committing the call to the ledger.
	if len(args) == 1 && args[0] == "Xooa Test" {

		logger.Debug("Method test call with argument 'Xooa Test'. Nothing will be committed to ledger.")
		response := shim.Success([]byte("Method test call. Nothing will be committed to ledger."))
		response.Status = 200
		return response
	}

	key := args[0]

	resultsIterator, err := stub.GetHistoryForKey(myCompositeKey)

	if err != nil {

		logger.Error("Error occured while calling GetHistoryForKey() - ", err.Error())
		response := shim.Error("Error occured while calling GetHistoryForKey() - " + err.Error())
		response.Status = 400
		return response
	}

	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the key.
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false

	for resultsIterator.HasNext() {

		response, err := resultsIterator.Next()

		if err != nil {
			logger.Debug("Error occured while iterating over resultsIterator - ", err.Error())
			errorResponse := shim.Error("Error occured while iterating resultsIterator - " + err.Error())
			errorResponse.Status = 400
			return errorResponse
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
		buffer.WriteString(string(response.Value))
		buffer.WriteString("\"")

		bArrayMemberAlreadyWritten = true
	}

	buffer.WriteString("]")

	response := shim.Success(buffer.Bytes())
	response.Status = 200
	return response
}

func (t *SimpleAsset) queryData(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	logger.Debug("queryData() called.")
	logger.Debug(args)

	//Function call with no arguments will return an error response and will not commit anything to ledger.
	if len(args) == 0 {

		logger.Debug("queryData method invoked without arguments. Returning an error response.")
		response := shim.Error("No arguments were available. Nothing was committed to ledger.")
		response.Status = 400
		return response
	}

	// Function call with argument "Xooa Test" is used for testing function names.
	// We return success wihout committing the call to the ledger.
	if len(args) == 1 && args[0] == "Xooa Test" {

		logger.Debug("Method test call with argument 'Xooa Test'. Nothing will be committed to ledger.")
		response := shim.Success([]byte("Method test call. Nothing will be committed to ledger."))
		response.Status = 200
		return response
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)

	if err != nil {

		return err
	}

	response := shim.Success(queryResults)
	response.Status = 200
	return response
}

func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	logger.Debug("getQueryResultForQueryString() called.")
	logger.Debug(queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)

	if err != nil {

		logger.Debug("Error occured while calling GetQueryResult() - ", err.Error())
		response := shim.Error("Error occured while calling GetQueryResult() - " + err.Error())
		response.Status = 400
		return nil, response
	}

	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the key.
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false

	for resultsIterator.HasNext() {

		response, err := resultsIterator.Next()

		if err != nil {

			logger.Debug("Error occured while iterating over resultsIterator - ", err.Error())
			errorResponse := shim.Error("Error occured while iterating resultsIterator - " + err.Error())
			errorResponse.Status = 400
			return nil, errorResponse
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
		buffer.WriteString(string(response.Value))
		buffer.WriteString("\"")

		bArrayMemberAlreadyWritten = true
	}

	buffer.WriteString("]")

	return buffer.bytes(), nil
}

// main function starts up the chaincode in the container during instantiate.
func main() {

	if err := shim.Start(new(SimpleAsset)); err != nil {

		logger.Error("Error starting SimpleAsset smartcontract: ", err)
		fmt.Printf("Error starting smartcontract: %s", err)
	}
}
