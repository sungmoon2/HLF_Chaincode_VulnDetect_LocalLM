package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Init")
	_, args := stub.GetFunctionAndParameters()
	var key string

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	result := []byte(`{"msg":"init chaincode"}`)

	// Initialize the chaincode
	key = args[0]
	stub.PutState(key, result)

	return shim.Success([]byte(fmt.Sprintf("%s",key)))
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "init" {
		return t.Init(stub)
	} else if function == "purchaseTx" {
                return t.purchaseTx(stub, args)
        } else if function == "checkCustomsTx" {
                return t.checkCustomsTx(stub, args)
        } else if function == "exportTx" {
                return t.exportTx(stub, args)
        } else if function == "checkRefunderTx" {
                return t.checkRefunderTx(stub, args)
        } else if function == "completeRefundTx" {
                return t.completeRefundTx(stub, args)
        } else if function == "queryByKey" {
		return t.queryByKey(stub, args)
	}

	return shim.Error("Invalid invoke function name.")
}

func (t *SimpleChaincode) purchaseTx(stub shim.ChaincodeStubInterface, args []string) pb.Response {

        if len(args) != 3 {
                return shim.Error("Incorrect number of arguments. Expecting 3")
        }

        var key string
        key = args[1]
        receipt := []byte(args[2])

        if args[0] != "client.merchant" {
                return shim.Error(fmt.Sprintf("%s", "caller is not permitted for this transaction. "))
        } else {
                stub.PutState(key, receipt)
                return shim.Success([]byte(fmt.Sprintf("%s",receipt)))
        }

}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) checkCustomsTx(stub shim.ChaincodeStubInterface, args []string) pb.Response {

        if len(args) != 3 {
                return shim.Error("Incorrect number of arguments. Expecting 3")
        }

        var key string
	key = args[1]
	receipt := []byte(args[2])

	if args[0] != "client.customs" {
		return shim.Error(fmt.Sprintf("%s", "caller is not permitted for this transaction. "))
	} else {
		stub.PutState(key, receipt)
	        return shim.Success([]byte(fmt.Sprintf("%s",receipt)))
	}

}

func (t *SimpleChaincode) exportTx(stub shim.ChaincodeStubInterface, args []string) pb.Response {

        if len(args) != 3 {
                return shim.Error("Incorrect number of arguments. Expecting 3")
        }

        var key string
        key = args[1]
        receipt := []byte(args[2])

        if args[0] != "client.customs" {
                return shim.Error(fmt.Sprintf("%s", "caller is not permitted for this transaction. "))
        } else {
                stub.PutState(key, receipt)
                return shim.Success([]byte(fmt.Sprintf("%s",receipt)))
        }

}

func (t *SimpleChaincode) checkRefunderTx(stub shim.ChaincodeStubInterface, args []string) pb.Response {

        if len(args) != 3 {
                return shim.Error("Incorrect number of arguments. Expecting 3")
        }

        var key string
        key = args[1]
        receipt := []byte(args[2])

        if args[0] != "client.customer" {
                return shim.Error(fmt.Sprintf("%s", "caller is not permitted for this transaction. "))
        } else {
                stub.PutState(key, receipt)
                return shim.Success([]byte(fmt.Sprintf("%s",receipt)))
        }

}

func (t *SimpleChaincode) completeRefundTx(stub shim.ChaincodeStubInterface, args []string) pb.Response {

        if len(args) != 3 {
                return shim.Error("Incorrect number of arguments. Expecting 3")
        }

        var key string
        key = args[1]
        receipt := []byte(args[2])

        if args[0] != "client.refunder" {
                return shim.Error(fmt.Sprintf("%s", "caller is not permitted for this transaction. "))
        } else {
                stub.PutState(key, receipt)
                return shim.Success([]byte(fmt.Sprintf("%s",receipt)))
        }

}

// query callback representing the query of a chaincode
func (t *SimpleChaincode) queryByKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	key = args[0]

	// Get the state from the ledger
	result, err := stub.GetState(key)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	}

	if result == nil {
		jsonResp := "{\"Error\":\"Nil receipt for " + key + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + key + "\",\"receipt\":\"" + string(result) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(result)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
