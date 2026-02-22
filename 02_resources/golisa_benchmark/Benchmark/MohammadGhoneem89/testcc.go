/*
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type participant struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database

	Name     string `json:"name"` //the fieldtags are needed to keep case from bouncing around
	Category string `json:"category"`
	Balance  int    `json:"balance"`
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ==========================================================================
// Init initializes chaincode
// ==========================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// =======================================================================================
// Invoke method - the entry point for Invocations
// =======================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "initParty" { //create a new Participant
		return t.initParty(stub, args)
	} else if function == "transferPoints" { //Transfer points from.... to.....
		return t.transferPoints(stub, args)
	} else if function == "readParty" { //read a Participant
		return t.readParty(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ===========================================================================================================
// initParty - create a new Participant, store into chaincode state
// ===========================================================================================================
func (t *SimpleChaincode) initParty(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//   0       1       2
	// "A", "Normal", "500"
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3 (Name, Type, Balance)")
	}

	// ==== Input checking ====
	fmt.Println("- start init participant")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}

	partyName := args[0]
	category := strings.ToLower(args[1])
	balance, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	// ==== Check if Participant already exists ====
	partyAsBytes, err := stub.GetState(partyName)
	if err != nil {
		return shim.Error("Failed to get Participant: " + err.Error())
	} else if partyAsBytes != nil {
		fmt.Println("This Participant already exists: " + partyName)
		return shim.Error("This Participant already exists: " + partyName)
	}

	// ==== Create Participant object and marshal to JSON ====
	objectType := "participant"
	participant := &participant{objectType, partyName, category, balance}
	partyJSONasBytes, err := json.Marshal(participant)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save Participant to state ===
	err = stub.PutState(partyName, partyJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Participant saved. Return success ====
	fmt.Println("- end init participant")
	return shim.Success(nil)
}

// ==============================================================================================
// readParty - read a participant from chaincode state
// ==============================================================================================
func (t *SimpleChaincode) readParty(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the Participant to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name) //get the participant from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"particpant does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// ==========================================================================================================
// transfer a Points from one participant to another
// ==========================================================================================================
func (t *SimpleChaincode) transferPoints(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0       1      2
	// "from",  "to", "Points"
	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	sender := args[0]
	receiver := args[1]
	points := args[2]
	taxAuth := "TaxAuth"

	pointsAsInt, err := strconv.Atoi(points)
	// need to check for err

	fmt.Println("- start Points Transfer ", sender, receiver, points)

	// Get the sender object
	senderAsBytes, err := stub.GetState(sender)
	if err != nil {
		return shim.Error("Failed to get Sender:" + err.Error())
	} else if senderAsBytes == nil {
		return shim.Error("Sender does not exist")
	}

	// Get the receiver object.
	receiverAsBytes, err := stub.GetState(receiver)
	if err != nil {
		return shim.Error("Failed to get Receiver:" + err.Error())
	} else if senderAsBytes == nil {
		return shim.Error("Receiver does not exist")
	}

	// Get the Tax_Authority object.
	taxAuthAsBytes, err := stub.GetState(taxAuth)
	if err != nil {
		return shim.Error("Failed to get Tax_Authority:" + err.Error())
	} else if senderAsBytes == nil {
		return shim.Error("Tax Authority does not exist")
	}

	// convert Tax Authority to json
	taxAuthority := participant{}
	err = json.Unmarshal(taxAuthAsBytes, &taxAuthority)
	if err != nil {
		return shim.Error(err.Error())
	}
	// convert Sender Object to json
	senderTransfer := participant{}
	err = json.Unmarshal(senderAsBytes, &senderTransfer)
	if err != nil {
		return shim.Error(err.Error())
	}

	// convert Receiver Object to json
	receiverTranfer := participant{}
	err = json.Unmarshal(receiverAsBytes, &receiverTranfer)
	if err != nil {
		return shim.Error(err.Error())
	}

	// check if sender or receiver of type Tax-Auth => Terminate the transaction.
	if senderTransfer.Category == "TaxAuth" {
		return shim.Error("Tax Authority cann't Participate in any transaction")
	} else if receiverTranfer.Category == "TaxAuth" {
		return shim.Error("Tax Authority cann't Participate in any transaction")
	}

	// check if the sender has enough points to send => Terminate the transaction.
	if senderTransfer.Balance < pointsAsInt {
		return shim.Error("There is no enough balance in the sender account")
	}

	// withdraw the points from sender account
	senderTransfer.Balance = senderTransfer.Balance - pointsAsInt //withdraw the points from sender account

	// Check receiver type to calculate the taxes
	if receiverTranfer.Category == "TaxExempt" {
		receiverTranfer.Balance = receiverTranfer.Balance + pointsAsInt // credit the points to Receiver account with no taxex
	} else {
		//calculate tax amount.
		taxPoints := (pointsAsInt * 2) / 100
		pointsAfterTax := pointsAsInt - taxPoints

		//transfer points to receiver after cutting the tax amount.
		receiverTranfer.Balance = receiverTranfer.Balance + pointsAfterTax

		//tranfer the tax to tax authority.
		taxAuthority.Balance = taxAuthority.Balance + taxPoints

		// Save the new values to the chain
		authorityJSONasBytes, _ := json.Marshal(taxAuthority)
		err = stub.PutState(taxAuth, authorityJSONasBytes) //rewrite the Tax Authority with the new balance.
		if err != nil {
			return shim.Error(err.Error())
		}

	}

	// Save the new values to the chain
	senderJSONasBytes, _ := json.Marshal(senderTransfer)
	err = stub.PutState(sender, senderJSONasBytes) //rewrite the participant
	if err != nil {
		return shim.Error(err.Error())
	}

	receiverJSONasBytes, _ := json.Marshal(receiverTranfer)
	err = stub.PutState(receiver, receiverJSONasBytes) //rewrite the participant
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end transferPoints (success)")
	return shim.Success(nil)

}
