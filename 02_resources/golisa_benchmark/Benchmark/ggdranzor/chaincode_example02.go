/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/*
 * The sample smart contract for documentation topic:
 * Writing Your First Blockchain Application
 */

package main

/* Imports
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
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

// Define the request structure, with 4 properties.  Structure tags are used by encoding/json library
type Request struct {

	ProviderID  string `json:"ProviderID"`
	PatientID string `json:"PatientID"`
	Category  string `json:"Category"`
    Status  string `json:"Status"`

}

/*
 * The Init method is called when the Smart Contract "fabcar" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "fabcar"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "publishRequest" {
		return s.publishRequest(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "response" {
		return s.response(APIstub, args)
	} else if function == "queryPatientRequests" {
		return s.queryPatientRequests(APIstub)
	} else if function == "revoke" {
		return s.revoke(APIstub, args)
	}
	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) publishRequest(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}
    //var Status = "pending"
    var request = Request{ProviderID: args[1], PatientID: args[2], Category: args[3], Status: "pending" }

	reqAsBytes, _ := json.Marshal(request)
    APIstub.PutState(args[0], reqAsBytes)
	
    return shim.Success(nil)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	requests := []Request{
		Request{ProviderID: "PR0", PatientID: "PA3", Category: "lifestyle", Status: "pending"},
		Request{ProviderID: "PR1", PatientID: "PA2", Category: "history", Status: "accepted"},
		Request{ProviderID: "PR2", PatientID: "PA1", Category: "medication", Status: "denied"},
		Request{ProviderID: "PR3", PatientID: "PA0", Category: "history", Status: "pending"},
		Request{ProviderID: "PR0", PatientID: "PA3", Category: "lifestyle", Status: "accepted"},
		Request{ProviderID: "PR1", PatientID: "PA2", Category: "history", Status: "accepted"},
		Request{ProviderID: "PR2", PatientID: "PA1", Category: "medication", Status: "accepted"},
		Request{ProviderID: "PR3", PatientID: "PA0", Category: "lifestyle", Status: "denied"},
		Request{ProviderID: "PR0", PatientID: "PA3", Category: "history", Status: "pending"},
		Request{ProviderID: "PR1", PatientID: "PA2", Category: "medication", Status: "pending"},
	}

	i := 0
	for i < len(requests) {
		fmt.Println("i is ", i)
		reqAsBytes, _ := json.Marshal(requests[i])
		APIstub.PutState("REQ"+strconv.Itoa(i), reqAsBytes)
		fmt.Println("Added", requests[i])
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) response(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

    //var request = Request{ProviderID: args[1], PatientID: args[2], Category: args[3], Status: args[4]}
    reqAsBytes, _ := APIstub.GetState(args[0])
	request := Request{}

	json.Unmarshal(reqAsBytes, &request)
    if request.PatientID == args[2] { 
	request.Status = args[1]

	reqAsBytes, _ = json.Marshal(request)
	APIstub.PutState(args[0], reqAsBytes)
    }
	return shim.Success(nil)
}

func (s *SmartContract) revoke(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

    //var request = Request{ProviderID: args[1], PatientID: args[2], Category: args[3], Status: args[4]}
    reqAsBytes, _ := APIstub.GetState(args[0])
	request := Request{}

	json.Unmarshal(reqAsBytes, &request)
    if (request.Status == "accepted" && request.PatientID == args[1]) { 
	request.Status = "revoked"

	reqAsBytes, _ = json.Marshal(request)
	APIstub.PutState(args[0], reqAsBytes)
    } else {
	return shim.Error("Cannot revoke.")}
	return shim.Success(nil)
}


func (s *SmartContract) queryPatientRequests(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "REQ0"
	endKey := "REQ9"

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

	fmt.Printf("- queryPatientRequests:\n%s\n", buffer.String())

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
