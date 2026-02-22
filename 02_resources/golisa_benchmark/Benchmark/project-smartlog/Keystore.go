/*
 * Copyright 2016-2019
 *
 * Interreg Central Baltic 2014-2020 funded project
 * Smart Logistics and Freight Villages Initiative, CB426
 *
 * Kouvola Innovation Oy, FINLAND
 * Region Örebro County, SWEDEN
 * Tallinn University of Technology, ESTONIA
 * Foundation Valga County Development Agency, ESTONIA
 * Transport and Telecommunication Institute, LATVIA
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/*
	Keystore SmartContract v0.1
*/

package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"fmt"
	"strconv"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

// Define the Message structure, with 2 properties.  Structure tags are used by encoding/json library
type Organisation struct {
	MSPID string `json:"mspID"`
	PublicKey string `json:"publicKey"`
}

/*
 * The Init method is called when the Smart Contract "UBL" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "UBL"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "addOrganisation" {
		return s.addOrganisation(APIstub, args)
	} else if function == "getOrganisation" {
		return s.getOrganisation(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

/*. 
 *	Add organisation public key and mspid to ledger. Key is stored in base64 format.
 */
func (s *SmartContract) addOrganisation(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	//	If no name was given as an argument, return error
	if len(args) != 2 {
		argsCount := strconv.Itoa(len(args))
		return shim.Error("Incorrect number of arguments, expecting 2. Args given:" + argsCount)
	}
	
	//	Create new message with given argument
	var organisation = Organisation{MSPID: args[0], PublicKey: args[1]}
	
	// Save organisation to ledger. Key is organisation MSPID.
	newMessageAsBytes, _ := json.Marshal(organisation)
	APIstub.PutState(args[0], newMessageAsBytes)

	//	Return new list as bytes, so we can be sure that the organisation was added.
	return shim.Success(newMessageAsBytes)
}

/*
 *	Returns single public key of organisation from ledger if it exists.
 */
func (s *SmartContract) getOrganisation(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	
	//	If no key was given as an argument, return error
	if len(args) != 1 {
		argsCount := strconv.Itoa(len(args))
		return shim.Error("Incorrect number of arguments, expecting 1. Args given:" + argsCount)
	}
	
	//	Try to find message by key from ledger
	organisationAsBytes, err := APIstub.GetState(args[0])
	//Not found
	if err != nil {
		return shim.Error("Message could not be found behind key:" + args[0])
	}
	
	//	Return message
	return shim.Success(organisationAsBytes)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
