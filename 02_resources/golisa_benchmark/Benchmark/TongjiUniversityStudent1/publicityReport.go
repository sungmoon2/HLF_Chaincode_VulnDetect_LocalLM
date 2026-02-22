/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

//两公示一公告举报

import (
	"fmt"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()
	var tableTime string  // ctreation time of the publicity Report list
	var operatorID string // ID of the operator
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2.")
	}

	// Initialize the chaincode
	tableTime = args[0]
	operatorID = args[1]

	// Write the state to the ledger
	err = stub.PutState(tableTime, []byte(operatorID))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke (stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "add" {		
		return t.add(stub, args)
	} else if function == "comfirm" {
		return t.comfirm(stub, args)
	}else if function == "query"{
		return t.query(stub,args)
	}
	return shim.Error("Invalid invoke function name. ")
}
func (t *SimpleChaincode) add (stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	reportID := args[0]
	reportInfo := args[1]
	Temp, errs := stub.GetState(reportID)

	if errs != nil {
		return shim.Error(errs.Error())
	}
	if Temp != nil {
        	return shim.Error("This ID already exists")
        }

	err := stub.PutState(reportID, []byte(reportInfo))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
	}

func (t *SimpleChaincode) comfirm (stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	reportID := args[0]
	reportInfo := args[1]
	Temp, errs := stub.GetState(reportID)
	if errs != nil {
		return shim.Error("List is not here")
	}
	if Temp == nil {
		return shim.Error("Entity not found")
	}
	err := stub.PutState(reportID, []byte(reportInfo))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)

 }
// Query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var reportID string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	reportID = args[0]

	// Get the state from the ledger
	reportInfo, err := stub.GetState(reportID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + reportID + "\"}"
		return shim.Error(jsonResp)
	}

	if reportInfo == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + reportID + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + reportID + "\",\"Amount\":\"" + string(reportInfo) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(reportInfo)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
