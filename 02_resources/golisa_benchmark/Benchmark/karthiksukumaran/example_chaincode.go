package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("example_chaincode")

type SimpleChaincode struct {
}

/**
 * Init is called when the Chaincode is Instantiated
 *
 **/
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response  {
	logger.Info("########### example_chaincode Init ###########")
	_, args := stub.GetFunctionAndParameters()
	var A, B string    // Entities
	A = args[0]
	B = args[1]
	err := stub.PutState(A, []byte(B))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}


/**
 * Invoke is called whenever the Chaincode is called.
 *
 **/
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### example_chaincode Init ###########")
	function, args := stub.GetFunctionAndParameters()

	var A, B string    // Entities
	A = args[0]
	

	if function == "delete" {
		//The Stored data is deleted
		err := stub.DelState(A)
		if err != nil {
			return shim.Error("Failed to delete state")
		}

		return shim.Success(nil)
	}

	if function == "query" {
		//Retrieving the data from the World State, of the data is not present error is thrown
		Avalbytes, err := stub.GetState(A)
		if err != nil {
			jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
			return shim.Error(jsonResp)
		}
		return shim.Success(Avalbytes)
	}
	if function == "setdata" {
		//Storing the data to the World State
		B = args[1]
		err := stub.PutState(A, []byte(B))
		if err != nil {
			return shim.Error(err.Error())
		}	
		return shim.Success(nil)
	}

	logger.Errorf("Unknown action, check the first argument, must be one of 'delete', 'query', or 'setdata'. But got: %v", args[0])
	return shim.Error("Unknown action, check the first argument, must be one of 'delete', 'query', or 'setdata'. But got: %v"+ args[0])
}


func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		logger.Errorf("Error starting Simple chaincode: %s", err)
	}
}
