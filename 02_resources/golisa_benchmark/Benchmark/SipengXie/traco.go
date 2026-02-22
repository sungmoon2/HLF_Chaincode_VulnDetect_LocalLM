package Trace

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type DataChaincode struct {
}

/*
type DataRecord struct {
	content []byte `json:"content"`
}
*/

func (t *DataChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	case "recordContent":
		return t.recordContent(stub, args)
	case "queryContent":
		return t.queryContent(stub, args)
	default:
		return shim.Error("No such function")
	}
}

func (t *DataChaincode) recordContent(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Expecting 1 input argv")
	}
	if len(args[0]) == 0 {
		return shim.Error("Expecting non-empty input argv")
	}
	key := stub.GetTxID()
	_ = stub.PutState(key, []byte(args[0]))
	return shim.Success(nil)
}

func (t *DataChaincode) queryContent(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Expecting 1 input argv")
	}
	if len(args[0]) == 0 {
		return shim.Error("Expecting non-empty input argv")
	}
	key := args[0]
	resBytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Unexpected error during query:" + err.Error())
	}
	if resBytes == nil {
		return shim.Error("No such key")
	}

	return shim.Success(resBytes)
}

func (t *DataChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success([]byte("Successfully invoke chaincode!"))
}

func main() {
	// Create a new Smart Contract
	err := shim.Start(new(DataChaincode))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
