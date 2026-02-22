package main

import (
	"fmt"
	"bytes"
	"time"
	
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SlaData example simple Chaincode implementation
type SlaData struct {
}

func (t *SlaData) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("slachaincode Init")
	return shim.Success(nil)
}

func (t *SlaData) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ex02 Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "addSlaData" {
		// Make payment of X units from A to B
		return t.addSlaData(stub, args)
	} else if function == "getSlaData" {
		// get an data from its state
		return t.getSlaData(stub, args)
	} else if function == "getSlaDataHistory" {
		// get an data from its state
		return t.getSlaDataHistory(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"addSlaData\" \"getSlaData\" \"getSlaDataHistory\"")
}

// Transaction makes payment of X units from A to B
func (t *SlaData) addSlaData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key, value string    // data holding
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	key = args[0]
	value = args[1]

	// Write the state back to the ledger
	err = stub.PutState(key, []byte(value))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (t *SlaData) getSlaData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key string // uniqueId
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting uniqeId to query")
	}

	key = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"no data for  " + key + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"key\":\"" + key + "\",\"value\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}

// query callback representing the query of a chaincode
func (t *SlaData) getSlaDataHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key string // uniqueId
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting uniqeId to query")
	}

	key = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetHistoryForKey(key)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"no data for  " + key + "\"}"
		return shim.Error(jsonResp)
	}
	defer Avalbytes.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for Avalbytes.HasNext() {
		response, err := Avalbytes.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")

		buffer.WriteString(string(response.Value))

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getSlaDataHistory returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}



func main() {
	err := shim.Start(new(SlaData))
	if err != nil {
		fmt.Printf("Error starting slachaincode: %s", err)
	}
}
