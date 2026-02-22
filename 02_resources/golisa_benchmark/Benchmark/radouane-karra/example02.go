package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
  pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("remote_attestation")

const (
	tableName = "UserAccounts"
)


// RemoteDeviceAttestation implementation. This smart contract enables multiple attestors
// to perform remote attestation of device and verify that the device is running authentic
// and valid software/hardware configuration
type PotCommun struct {
}

func (t *PotCommun) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("PotCommun Init")

  var err error
  function, args := stub.GetFunctionAndParameters()

	if len(args) != 0 {
		logger.Error("Incorrect number of arguments")
    return shim.Error("Incorrect number of arguments. No arguments required for deploying this contract.")
	}

	_, err = stub.GetTable(tableName)
	if err == shim.ErrTableNotFound {
		err = stub.CreateTable(tableName, []*shim.ColumnDefinition{
			&shim.ColumnDefinition{Name: "UserID", Type: shim.ColumnDefinition_STRING, Key: true},
			&shim.ColumnDefinition{Name: "AccountID", Type: shim.ColumnDefinition_STRING, Key: true},
			&shim.ColumnDefinition{Name: "Status", Type: shim.ColumnDefinition_UINT64, Key: false},
		})
		if err != nil {
			logger.Errorf("Error creating table:%s", err.Error())
			return shim.Error("Failed creating DeviceAttestation table.")
		}
	} else {
		logger.Info("Table already exists")
	}

	logger.Info("Successfully deployed chain code")

	return shim.Success(nil)
}

func (t *PotCommun) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ex02 Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// Make payment of X units from A to B
		return t.invoke(stub, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		return t.delete(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemtned in invoke
		return t.query(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\"")
}

// Transaction makes payment of X units from A to B
func (t *PotCommun) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	A = args[0]
	B = args[1]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Deletes an entity from state
func (t *PotCommun) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (t *PotCommun) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(Avalbytes)
}


func main() {
	err := shim.Start(new(PotCommun))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
