package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type SmartContract struct {}

func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	args := stub.GetStringArgs()
	s.create(stub, args)
	return shim.Success(nil)
}

// func (s *SmartContract) Query(stub shim.ChaincodeStubInterface) pb.Response {
// 	return shim.Success(nil)
// }

func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := stub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "create" {
		return s.create(stub, args)
	} else if function == "query" {
		return s.query(stub, args)
	} else if function == "update" {
		return s.update(stub, args)
	} else if function == "delete" {
		return s.delete(stub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) create(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	_key  	:= args[0]
	_data	:= args[1]

	if(_data == ""){
		return shim.Error("Incorrect string of data")
	}

	existAsBytes,err := stub.GetState(_key)
	if string(existAsBytes) != "" {
		fmt.Println("Failed to create key, Duplicate key.")
		return shim.Error("Failed to create key, Duplicate key.")
	}

	err = stub.PutState(_key, []byte(_data))
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Printf("create %s %s \n", _key, string(_data))

	return shim.Success(nil)
}

func (s *SmartContract) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	_key  	:= args[0]
	_data, err := stub.GetState(_key)
	if err != nil {
		return shim.Error(err.Error())
	}
	if string(_data) == "" {
		return shim.Error("The key isn't exist.")
	}else{
		fmt.Printf("query %s %s \n", _key, string(_data))
	}

	return shim.Success(_data)
}

func (s *SmartContract) update(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	_key  	:= args[0]
	_data	:= args[1]

	if(_data == ""){
		return shim.Error("Incorrect string of data")
	}

	err := stub.PutState(_key, []byte(_data))
	if err != nil {
		return shim.Error(err.Error())
	}else{
		fmt.Printf("update %s %s \n", _key, string(_data))
	}
	
	return shim.Success(nil)
}

// Deletes an entity from state
func (t *SmartContract) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	_key := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(_key)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

func main() {

	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}