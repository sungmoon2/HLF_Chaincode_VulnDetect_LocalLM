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
	Transport chain v0.1
*/

package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"fmt"
	"strconv"
	sc "github.com/hyperledger/fabric/protos/peer"
    "strings"
)

// Define the Smart Contract structure
type SmartContract struct {
}

// Define the Chain structure, with 2 properties.  Structure tags are used by encoding/json library
type Chain struct {
	Id string `json:"id"`
	Participants []string `json:"participants"`
}

/*
 * The Init method is called when the Smart Contract "UBL" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "transportChain"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "setTransportChain" {
		return s.setTransportChain(APIstub, args)
	} else if function == "getTransportChain" {
		return s.getTransportChain(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

/*. 
 *	Adds new message to list and puts new list to ledger.
 */
func (s *SmartContract) setTransportChain(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	//	If no name was given as an argument, return error
	if len(args) != 2 {
		argsCount := strconv.Itoa(len(args))
		return shim.Error("Incorrect number of arguments, expecting 3. Args given:" + argsCount)
	}
	
	var chains []Chain
	
	chainsAsBytes, _ := APIstub.GetState("chains")
	
	err := json.Unmarshal(chainsAsBytes, &chains) //unmarshal it aka JSON.parse()
	
	//	Chains is empty, maybe this is the first time it is used?
	if (len(chains) != 0) {
		
		//	Empty but still got error, so something was really wrong
		if err != nil {
			return shim.Error(err.Error())
		}
		
	}
	
	var parts = strings.Split(args[1], ",")
	
	//	Create new chain with given argument
	var newChain = Chain{Id: args[0], Participants: parts}
	
	var i = -1
	
	// index is the index where we are
    // element is the element from someSlice for where we are
	for index, chain := range chains {
    
		//	Old chain was found so store its index
	    if (chain.Id == args[0]) {
	    	i = index
	    	break
	    }
	}
	
	//	If index of the chain was found
	if (i != -1) {
		chains[i] = newChain
		newChainsAsBytes, _ := json.Marshal(chains)
		APIstub.PutState("chains", newChainsAsBytes)
		
		// Marshal to json-bytes
		newChainAsBytes, _ := json.Marshal(newChain)
	
		//	Return new list as bytes, so we can be sure that the Message was added
		return shim.Success(newChainAsBytes)
	}
	
	chains = append(chains, newChain)
	
	newChainsAsBytes, _ := json.Marshal(chains)
	APIstub.PutState("chains", newChainsAsBytes)
	
	// Save chain behind key
	newChainAsBytes, _ := json.Marshal(newChain)

	//	Return new list as bytes, so we can be sure that the Message was added
	return shim.Success(newChainAsBytes)
}

/*
 *	Returns a list of keys that were used to save Message to ledger.
 */
func (s *SmartContract) getTransportChain(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	
	//	If no id was given as an argument, return error
	if len(args) != 1 {
		argsCount := strconv.Itoa(len(args))
		return shim.Error("Incorrect number of arguments, expecting 1. Args given:" + argsCount)
	}
	
	var chains []Chain
	chainsAsBytes, _ := APIstub.GetState("chains")
	
	//	Get all chains
	if (args[0] == "") {
		return shim.Success(chainsAsBytes)
	}
	
	err := json.Unmarshal(chainsAsBytes, &chains) //unmarshal it aka JSON.parse()
	
	//	Could not parse to Chain-struct
	if err != nil {
		return shim.Error(err.Error())
	}
	
	// index is the index where we are
    // element is the element from someSlice for where we are
	for _, chain := range chains {
    
		//	Right chain was found
	    if (chain.Id == args[0]) {
	    	chainAsBytes, _ := json.Marshal(chain)
	    	return shim.Success(chainAsBytes)
	    }
	}
	
	return shim.Success(nil)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
