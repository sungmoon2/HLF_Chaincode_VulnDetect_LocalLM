package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type SimpleChaincode struct{}

func (cc *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	_, args := stub.GetFunctionAndParameters()

	var A, B string        //entity
	var Avalue, Bvalue int //asset
	var err error

	if len(args) != 4 {
		return shim.Error("Incorrect number of argument. Expect 4")
	}

	A = args[0]
	Avalue, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expect integer value")
	}
	B = args[2]
	Bvalue, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Expect integer value")
	}

	err = stub.PutState(A, []byte(strconv.Itoa(Avalue)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bvalue)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)

}

func (cc *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fn, args := stub.GetFunctionAndParameters()

	if fn == "invoke" {
		return cc.invoke(stub, args)
	} else if fn == "delete" {
		return cc.delete(stub, args)
	} else if fn == "query" {
		return cc.query(stub, args)
	} else if fn == "set" {
		return cc.set(stub, args)
	}

	return shim.Error("Invalid invoke function name.")
}

func (cc *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A, B string
	var Avalue, Bvalue int
	var X int //transaction value
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expect 3")
	}

	A = args[0]
	B = args[1]

	AvalueBytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if AvalueBytes == nil {
		return shim.Error("Entity not found")
	}
	Avalue, _ = strconv.Atoi(string(AvalueBytes))

	BvalueBytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if BvalueBytes == nil {
		return shim.Error("Entity not found")
	}
	Bvalue, _ = strconv.Atoi(string(BvalueBytes))

	X, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount. Expect a integer value")
	}

	Avalue = Avalue - X
	Bvalue = Bvalue + X

	err = stub.PutState(A, []byte(strconv.Itoa(Avalue)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bvalue)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (cc *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expect 1")
	}

	err := stub.DelState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (cc *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expect one, the name of the person to query")
	}

	AvalueBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	if AvalueBytes == nil {
		return shim.Error("nil amount")
	}

	fmt.Printf("query result:%s \n", string(AvalueBytes))
	return shim.Success(AvalueBytes)
}

func (cc *SimpleChaincode) set(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expect 2")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(args[1]))
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting SimpleChaincode: %s", err)
	}
}
