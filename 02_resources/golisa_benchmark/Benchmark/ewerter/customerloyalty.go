package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type CustomerLoyalty struct {
}

func (cc *CustomerLoyalty) Init(stub shim.ChaincodeStubInterface) pb.Response {

	return shim.Sucess(nil)
}

func (cc *CustomerLoyalty) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	function, args := stub.GetFunctionAndParameters()

	if function == "put" {
		return cc.put(stub, args)
	} else if function == "get" {
		return cc.get(stub, args)
	} else if function == "del" {
		return cc.del(stub, args)
	}

	message := fmt.Sprintf("unknown function name: %s, expected one of {get, put, del}", function)

	return pb.Response{Status: 400, Message: message}
}

func (cc *CustomerLoyalty) put(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		message := fmt.Sprintf("wrong number of arguments: passed %d, expected %d", len(args), 2)

		return pb.Response{Status: 400, Message: message}
	}

	key, value := args[0], args[1]

	if key == "" {
		message := "key must be a non-empty string"
		return pb.Response{Status: 400, Message: message}
	}

	if err := stub.PutState(key, []byte(value)); err != nil {
		message := fmt.Sprintf("unable to put a key-value pair: %s", err.Error())
		return shim.Error(message)
	}
	return shim.Success(nil)
}

func (cc *CustomerLoyalty) get(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		message := fmt.Sprintf("wrong number of arguments: passed %d, expected %d", len(args), 1)
		return pb.Response{Status: 400, Message: message}
	}

	key := args[0]

	if key == "" {
		message := "key must be a non-empty string"
		return pb.Response{Status: 400, Message: message}
	}

	valueBytes, err := stub.GetState(key)
	if err != nil {
		message := fmt.Sprintf("unable to get a value for the key %s: %s", key, err.Error())
		return shim.Error(message)
	}

	if valueBytes == nil {
		message := fmt.Sprintf("a value for the key %s not found", key)
		return pb.Response{Status: 404, Message: message}
	}

	return shim.Success(valueBytes)
}

func (cc *CustomerLoyalty) del(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		message := fmt.Sprintf("wrong number of arguments: passed %d, expected %d", len(args), 1)
		return pb.Response{Status: 400, Message: message}
	}

	key := args[0]

	if key == "" {
		message := "key must be a non-empty string"
		return pb.Response{Status: 400, Message: message}
	}

	if err := stub.DelState(key); err != nil {
		message := fmt.Sprintf("unable to delete a pair associated with the key %s: %s", key, err.Error())
		return shim.Error(message)
	}

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(CustomerLoyalty))
	if err != nil {
		fmt.Printf("Error starting CustomerLoyalty: %s", err)
	}
}
