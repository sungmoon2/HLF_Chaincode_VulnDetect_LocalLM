package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define simple contract structure
type SimpleChaincode struct {
}

// Define the BankAccount struct
type BankAccount struct {
	AccountNumber string `json:"accountnumber"`
	Amount        int    `json:"amount"`
	AccountHolder string `json:"accountholder"`
}

func (s *SimpleChaincode) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SimpleChaincode) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Simple Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()

	//Route to appropriate handler function
	if function == "queryAccount" {
		return s.queryAccount(APIstub, args)
	} else if function == "createAccount" {
		return s.createAccount(APIstub, args)
	}

	return shim.Error("Invalid smart contract function name.")
}

func (s *SimpleChaincode) queryAccount(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		shim.Error("Incorrect number of arguments. Expecting 1")
	}
	bankAccountAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(bankAccountAsBytes)
}

func (s *SimpleChaincode) createAccount(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}
	amount, _ := strconv.Atoi(args[2])
	account := BankAccount{AccountNumber: args[1], Amount: amount, AccountHolder: args[3]}

	accountAsBytes, _ := json.Marshal(account)
	APIstub.PutState(args[0], accountAsBytes)

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error creating new smart contract: %s", err)
	}
}
