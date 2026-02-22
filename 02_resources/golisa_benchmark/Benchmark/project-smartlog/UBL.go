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
	UBL-API SmartContract v0.8
*/

package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"fmt"
	b64 "encoding/base64"
	sc "github.com/hyperledger/fabric/protos/peer"
	"strconv"
)

// Define the Smart Contract structure
type SmartContract struct {
}

// Define the Message structure. Structure tags are used by encoding/json library
type Message struct {
	DocumentID string `json:"documentID"`
	OrganisationID string `json:"organisationID"`
	SupplyChainID string `json:"supplyChainID"`
	ContainerID string `json:"containerID"`
	SenderParty string `json:"senderParty"`
	RFIDTransportEquipment string `json:"RFIDTransportEquipment"`
	RFIDTransportHandlingUnit string `json:"RFIDTransportHandlingUnit"`
	StatusTypeCode string `json:"statusTypeCode"`
	EncryptedMessage string `json:"encryptedMessage"`
	Participants []Participant `json:"participants"`
	Timestamp string `json:"timestamp"`
	CarrierAssignedID string `json:"carrierAssignedID"`
	ShippingOrderID string `json:"shippingOrderID"`
	EmptyFullIndicator string `json:"emptyFullIndicator"`
	ContentType string `json:"contentType"`
	ContentTypeSchemeVersion string `json:"contentTypeSchemeVersion"`
	StatusLocationID string `json:"statusLocationId"`
}

type Participant struct {
	MSPID string `json:"MSPID"`
	EncryptedKey string `json:"enryptedKey"`
}


var logger = shim.NewLogger("UBLChaincode")

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
	fmt.Println("Invoking chaincode function:" + function)
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "addMessage" {
		return s.addMessage(APIstub, args)
	} else if function == "getMessage" {
		return s.getMessage(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name:" + function)
}

/*.
 *	Adds new encrypted message with participants.
	Argument order:
	#1. Message scruct in JSON. example: {"encryptedMessage": "message","participants": [{"MSPID": "Org1MSP","encryptedKey": "KEY"},{"MSPID": "Org2MSP","encryptedKey": "KEY2"}]
 */
func (s *SmartContract) addMessage(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	logger.Info("addMessage called")

	//	If no name was given as an argument, return error
	if len(args) != 2 {
		argsCount := strconv.Itoa(len(args))
		return shim.Error("Incorrect number of arguments, expecting 3. Args given:" + argsCount)
	}

	data := args[0]
    key := args[1]

	//Decode base64 data
	decodedBytes, err := b64.StdEncoding.DecodeString(data)
	if(err != nil) {
		return shim.Error("Error decoding base64:" + err.Error())
	}

	var message Message

	err2 := json.Unmarshal(decodedBytes, &message) //unmarshal it aka JSON.parse()

	if err2 != nil {
		return shim.Error("Error unmarshallling! Not valid JSON:" + err2.Error())
	}

	// Save message behind key
	//newMessageAsBytes, _ := json.Marshal(message)
	APIstub.PutState(key, decodedBytes)

	return shim.Success([]byte("Message succesfully saved"))
}

/*
 *	Returns single message behind key from ledger if it exists.
 */
func (s *SmartContract) getMessage(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	//	If no key was given as an argument, return error
	if len(args) != 1 {
		argsCount := strconv.Itoa(len(args))
		return shim.Error("Incorrect number of arguments, expecting 1. Args given:" + argsCount)
	}

	//	Try to find message by key from ledger
	messageAsBytes, err := APIstub.GetState(args[0])
	//Not found
	if err != nil {
		return shim.Error("Message could not be found behind key:" + args[0])
	}

	//	Return message
	return shim.Success(messageAsBytes)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {
	logger.SetLevel(shim.LogInfo)
	logger.Info("Starting Chaincode")
	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}

}
