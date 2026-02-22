package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type SmartContract struct {

}

type Customer struct {
	SSN string `json:"ssn"`
	IdImage string `json:"idimage"`
	FirstName string `json:"firstname"`
	LastName string `json:"lastname"`
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()

	if function == "queryCustomer" {
		return s.queryCustomer(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "insertCustomer" {
		return s.insertCustomer(APIstub, args)
	}
	return shim.Error("Invalid Custom smart contract function name")
}

func (s *SmartContract) queryCustomer(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of args. Expects 1")
	}

	customerBytes, _ := APIstub.GetState(args[0])
	if customerBytes == nil {
		return shim.Error("Could not locate customer")
	}
	return shim.Success(customerBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}


func (s *SmartContract) insertCustomer(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 4 {
		return shim.Error("Incorrect arguments, Want 4")
	}

	var customer = Customer{IdImage: args[1], FirstName: args[2], LastName: args[3]}
	var ssn = args[0]
	customerBytes, _ := json.Marshal(customer)
	err := APIstub.PutState(ssn, customerBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to record customer: %s", ssn))
	}

	return shim.Success(nil)
}


/*
 * main function *
calls the Start function
The main function starts the chaincode in the container during instantiation.
 */
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Customer Smart Contract: %s", err)
	}
}
