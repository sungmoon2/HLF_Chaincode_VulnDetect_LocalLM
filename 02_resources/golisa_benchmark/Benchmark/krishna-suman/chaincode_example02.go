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

//WARNING - this chaincode's ID is hard-coded in chaincode_example04 to illustrate one way of
//calling chaincode from a chaincode. If this example is modified, chaincode_example04.go has
//to be modified as well with the new ID of chaincode_example02.
//chaincode_example05 show's how chaincode ID can be passed in as a parameter instead of
//hard-coding.

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ex02 Init")
	_, args := stub.GetFunctionAndParameters()
	var a, b string    // Entities
	var aval, bval int // Asset holdings
	var err error

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// Initialize the chaincode
	a = args[0]
	aval, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	b = args[2]
	bval, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	fmt.Printf("a's bal = %d, b's bal = %d\n", aval, bval)

	// Write the state to the ledger
	err = stub.PutState(a, []byte(strconv.Itoa(aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(b, []byte(strconv.Itoa(bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("myc Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// Make payment of x units from a to b
		return t.invoke(stub, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		return t.delete(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemtned in invoke
		return t.query(stub, args)
	} else if function == "add" {
		// the old "Query" is now implemtned in invoke
		return t.add(stub, args)
	}else if function == "getAll" {
		// the old "Query" is now implemtned in invoke
		return t.getAll(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\"")
}

// Transaction makes payment of x units from a to b
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var a, b string    // Entities
	var aval, bval int // Asset holdings
	var x int          // Transaction value
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	a = args[0]
	b = args[1]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(a)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(b)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	x, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	aval = aval - x
	bval = bval + x
	fmt.Printf("aval = %d, bval = %d\n", aval, bval)

	// Write the state back to the ledger
	err = stub.PutState(a, []byte(strconv.Itoa(aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(b, []byte(strconv.Itoa(bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	a := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(a)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var a string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	a = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(a)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + a + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + a + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + a + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

// add new entity or customer
func (t *SimpleChaincode) add(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var c string // Entities
	var err error
	var x int

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person and amount to add")
	}
	fmt.Printf("arguments:",args[0],args[1])
	c = args[0]
	x, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting c integer value")
	}
	// Get the state from the ledger
	Avalbytes, err := stub.GetState(c)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + c + "\"}"
		return shim.Error(jsonResp)
	}
	if Avalbytes != nil {
		fmt.Printf(c,"already exists, hence cannot add")
		jsonResp := "{\"Error\":\"Nil amount for " + c + "\"}"
		return shim.Error(jsonResp)
	}else{
		fmt.Printf(c,"is not present, hence add")
	}
	err = stub.PutState(c, []byte(strconv.Itoa(x)))
	if err != nil {
		return shim.Error(err.Error())
	}
	jsonResp := "{\"Name\":\"" + c + "\",\"Amount\":\"" + string(x) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

// add new entity or customer
func (t *SimpleChaincode) getAll(stub shim.ChaincodeStubInterface, args []string) pb.Response {	
	resultsIterator, err := stub.GetStateByRange("a", "z")
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
	fmt.Printf("- query all persons:\n%s\n", buffer.String())
	return shim.Success(buffer.Bytes())
}


func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
