package main

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// This program deals with two different assets - Batch & SKU

//SimpleChaincode is used
type SimpleChaincode struct {
}

var logger = shim.NewLogger("SimpleChaincode")

func main() {

	// Setting up the log level programtically
	logger.SetLevel(shim.LogDebug)
	logger.Debug("Enter : main method")

	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		logger.Critical("Failed to start chaincode -", err)
	}

	logger.Debug("Exit : main method")

}

// Init method
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke method
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	function, args := stub.GetFunctionAndParameters()
	if function == "set" {
		// set player - game
		return t.set(stub, args)
	} else if function == "get" {
		// get game by player
		return t.get(stub, args)
	}

	return shim.Error("Invalid function name. Expecting \"set\" \"get\"")
}

//set - Supporting Functions used by Invoke method
func (t *SimpleChaincode) set(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	key := args[0]
	value := args[1]

	err := stub.PutState(key, []byte(value))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)

}

//get - Supporting Functions used by Invoke method
func (t *SimpleChaincode) get(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	startKey := args[0]

	resultsIterator, err := stub.GetHistoryForKey(startKey)
	defer resultsIterator.Close()

	if err != nil {
		return shim.Error(err.Error())
	}
	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false

	for resultsIterator.HasNext() {
		Result, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(Result.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(Result.Timestamp.Seconds, int64(Result.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(Result.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(Result.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("get_all_components:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}
