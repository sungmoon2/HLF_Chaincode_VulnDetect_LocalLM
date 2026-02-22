package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("Chaincode_cc0")

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### Chaincode_cc0 Init ###########")
	fmt.Println("Chaincode initialized .....")
	var err error
	// Write the state to the ledger
	err = stub.PutState("I am Chaincode", []byte("Hello World"))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("Init la re"))
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### example_cc0 Invoke ###########")
	fmt.Println("Chaincode Invoked .....")
	return shim.Success([]byte("Arry challa reyyyy"))
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		logger.Errorf("Error starting Simple chaincode: %s", err)
	}
}
