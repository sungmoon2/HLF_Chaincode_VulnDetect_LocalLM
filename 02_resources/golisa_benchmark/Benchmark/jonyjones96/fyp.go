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

// Define the car structure, with 4 properties.  Structure tags are used by encoding/json library
type DataBlock struct {
	SourceID       string `json:"sourceid"`
	OrganisationID string `json:"organisationid"`
	MetaData       string `json:"metadata"`
	Date           string `json:"date"`
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
	//console.log("Function")
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryBlock" {
		return s.queryBlock(APIstub, args)
	} else if function == "insertBlock" {
		return s.insertBlock(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "queryAllBlocks" {
		return s.queryAllBlocks(APIstub)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryBlock(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	blockAsBytes, _ := APIstub.GetState(args[0])
	//messageAsBytes := []byte("args[0] = " + args[0])
	//return shim.Success(messageAsBytes)
	return shim.Success(blockAsBytes)
}

func (s *SmartContract) insertBlock(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	var block = DataBlock{SourceID: args[1], OrganisationID: args[2], MetaData: args[3], Date: args[4]}

	blockAsBytes, _ := json.Marshal(block)
	APIstub.PutState(args[0], blockAsBytes)

	return shim.Success(nil)

}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	dataBlocks := []DataBlock{
		DataBlock{SourceID: "1", OrganisationID: "Facebook.com", MetaData: "Name", Date: "01/03/2018"},
		DataBlock{SourceID: "Facebook.com", OrganisationID: "Google.com", MetaData: "Name", Date: "02/03/2018"},
		DataBlock{SourceID: "Google.com", OrganisationID: "Twitter.com", MetaData: "Name", Date: "02/03/2018"},
		DataBlock{SourceID: "Google.com", OrganisationID: "AirBnB", MetaData: "Name", Date: "03/03/2018"},
	}

	i := 0
	for i < len(dataBlocks) {
		fmt.Println("i is ", i)
		blockAsBytes, _ := json.Marshal(dataBlocks[i])
		APIstub.PutState("BLOCK"+strconv.Itoa(i), blockAsBytes)
		fmt.Println("Added", dataBlocks[i])
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) queryAllBlocks(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "BLOCK0"
	endKey := "BLOCK999"

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

	fmt.Printf("- queryAllCars:\n%s\n", buffer.String())

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
