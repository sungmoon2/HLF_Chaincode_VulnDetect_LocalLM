package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// Init initializes the chaincode
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {

	fmt.Println("SampleChainCode Init")

	var err error

	_, args := stub.GetFunctionAndParameters()
	var local string // Entities
	var localVal int // Asset holdings

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// Initialize the chaincode
	local = args[0]
	localVal, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}

	fmt.Printf("localVal = %d\n", localVal)

	// Write the state to the ledger
	err = stub.PutState(local, []byte(strconv.Itoa(localVal)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("Chaincode invoked")
	function, args := stub.GetFunctionAndParameters()

	switch function {
	case "delete":
		return t.delete(stub, args)
	case "query":
		return t.query(stub, args)
	case "putMultiple":
		return t.putMultiple(stub, args)
	case "update":
		return t.update(stub, args)
	default:
		return shim.Error(fmt.Sprintf("Invalid Smart Contract function : %s\n", function))
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\" \"putMultiple\"")
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	local := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(local)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	fmt.Println("Deleted")

	return shim.Success(nil)
}

// Puts an entry into the state specified number of times
func (t *SimpleChaincode) putMultiple(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	var err error
	local := args[0]
	var localVal, times int
	localVal, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	times, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}

	for i := 0; i < times; i++ {
		err = stub.PutState(local+strconv.Itoa(i), []byte(strconv.Itoa(localVal)+strconv.Itoa(i)))
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	fmt.Printf("State put %d times\n", times)

	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var local string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	local = args[0]

	// Get the state from the ledger
	localValbytes, err := stub.GetState(local)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + local + "\"}"
		return shim.Error(jsonResp)
	}

	if localValbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + local + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + local + "\",\"Amount\":\"" + string(localValbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(localValbytes)
}

// update an entry in state
func (t *SimpleChaincode) update(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var local string // Entities
	var newVal int   // Asset Holdings
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	local = args[0]

	// Perform the execution
	newVal, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}

	// Write the state back to the ledger
	err = stub.PutState(local, []byte(strconv.Itoa(newVal)))
	if err != nil {
		return shim.Error(err.Error())
	}

	// Get the state from the ledger

	localValbytes, err := stub.GetState(local)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + local + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(localValbytes)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s\n", err)
	}
}
