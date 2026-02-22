package main

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type SimpleChaincode struct {
}

type User struct {
	User   string `json:"user"`
	Amount string `json:"amount"`
}

var userNumber int = 0

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	t.mint(stub, []string{"Total", "0"})
	userNumber = 1
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "mint" {
		return t.mint(stub, args)
	} else if function == "balance" {
		return t.printBalance(stub, args)
	} else if function == "withdraw" {
		return t.withdraw(stub, args, false)
	} else if function == "transfer" {
		return t.transfer(stub, args)
	} else if function == "totalAmount" {
		return t.printBalance(stub, []string{"Total"})
	} else if function == "queryAllUsers" {
		return t.queryAllUsers(stub)
	}

	return shim.Error("Invalid invoke function name. Expecting \"mint\" \"transfer\" \"totalAmount\" \"balanceOf\" \"withdraw\" \"queryAllUsers\"")
}

func (t *SimpleChaincode) mint(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	user := args[0]
	amount, _ := strconv.Atoi(args[1])

	if user == "Total" && userNumber > 0 {
		return shim.Error("Invalid user name.")
	}
	if amount < 0 {
		return shim.Error("Amount can't be negative.")
	}
	if balanceOf(stub, []string{user}) != nil {
		return shim.Error("The user already exists.")
	}

	// {"USER"+N: user name}, {user name: amount}
	if err := stub.PutState("USER"+strconv.Itoa(userNumber), []byte(user)); err != nil {
		return shim.Error(err.Error())
	}
	if err := stub.PutState(user, []byte(strconv.Itoa(amount))); err != nil {
		return shim.Error(err.Error())
	}
	userNumber++

	t.addTotalAmount(stub, amount)

	return shim.Success(nil)
}

func (t *SimpleChaincode) addTotalAmount(stub shim.ChaincodeStubInterface, amount int) pb.Response {

	totalAmountAsBytes := balanceOf(stub, []string{"Total"})
	totalAmount, _ := strconv.Atoi(string(totalAmountAsBytes))
	totalAmount += amount

	if err := stub.PutState("Total", []byte(strconv.Itoa(totalAmount))); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func balanceOf(stub shim.ChaincodeStubInterface, args []string) []byte {
	user := args[0]

	balanceAsBytes, err := stub.GetState(user)
	if err != nil {
		return nil
	}

	return balanceAsBytes
}

func (t *SimpleChaincode) printBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	user := args[0]
	balance := string(balanceOf(stub, []string{user}))

	var buffer bytes.Buffer
	if user == "Total" {
		buffer.WriteString(user + " amount: " + balance)
	} else {
		buffer.WriteString(user + "'s balance: " + balance)
	}

	//	fmt.Printf("%s's balance: %d", user, balance)

	return shim.Success(buffer.Bytes())
}

func (t *SimpleChaincode) withdraw(stub shim.ChaincodeStubInterface, args []string, isTransferred bool) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	user := args[0]
	amount, _ := strconv.Atoi(args[1])

	if amount < 0 {
		return shim.Error("Amount can't be negative.")
	}

	balanceAsBytes := balanceOf(stub, []string{user})

	balance, _ := strconv.Atoi(string(balanceAsBytes))
	balance = balance - amount

	if balance < 0 {
		return shim.Error("The withdrawal amount is more than the balance.")
	}

	if err := stub.PutState(user, []byte(strconv.Itoa(balance))); err != nil {
		return shim.Error(err.Error())
	}

	if isTransferred == false {
		t.addTotalAmount(stub, -amount)
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	info := []string{args[0], args[2]}
	t.withdraw(stub, info, true)

	beneficiary := args[1]

	balance, _ := strconv.Atoi(string(balanceOf(stub, []string{beneficiary})))
	amount, _ := strconv.Atoi(args[2])
	balance = balance + amount

	if err := stub.PutState(beneficiary, []byte(strconv.Itoa(balance))); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) queryAllUsers(stub shim.ChaincodeStubInterface) pb.Response {
	startKey := "USER1"
	endKey := "USER999"

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer

	buffer.WriteString("{\"User\": ")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(", ")
		}
		buffer.WriteString("\"")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("\"")

		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("}")

	//	fmt.Printf("- queryAllUsers:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
