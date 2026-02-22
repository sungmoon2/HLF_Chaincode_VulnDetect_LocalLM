package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type IntermChaincode struct {
}

func (IntermChaincode *IntermChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {

	fmt.Println("Init executed")
	_, args := stub.GetFunctionAndParameters()

	fmt.Println("The Init triggered and args called are :-", args)
	return shim.Success([]byte("success"))
}

// Invoke method
func (IntermChaincode *IntermChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	// Get the function name and parameters
	function, args := stub.GetFunctionAndParameters()

	fmt.Println("Invoke executed : ", function, " args=", args)

	switch {

	// Query function
	case function == "createContainer":
		return IntermChaincode.createContainer(stub, args)
	case function == "getContainer":
		return IntermChaincode.getContainer(stub, args)
	case function == "createTruck":
		return IntermChaincode.createTruck(stub, args)
	case function == "getTruck":
		return IntermChaincode.getTruck(stub, args)
	case function == "assignTruck":
		return IntermChaincode.assignTruck(stub, args)
	case function == "loadContainer":
		return IntermChaincode.loadContainer(stub, args)
	case function == "readyContainer":
		return IntermChaincode.readyContainer(stub, args)
	case function == "clearContainer":
		return IntermChaincode.clearContainer(stub, args)
	}

	return shim.Error(("Bad Function Name = !!!"))
}

func main() {
	fmt.Println("Started the Interm Chaincode")
	err := shim.Start(new(IntermChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
