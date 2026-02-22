package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type Test struct {
	//test strct
}

func (this *Test) Init(stub shim.ChaincodeStubInterface) peer.Response {
	//only one
	return shim.Success(nil)
}

func (this *Test) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	//qufendiaoyongde1fangfa
	function, parameters := stub.GetFunctionAndParameters()
	if function == "get" {
		return this.get(stub, parameters[0])
	} else if function == "set" {
		return this.set(stub, parameters[0], []byte(parameters[1]))
	}
	return shim.Error("Invalid Smart Contract function name.	")
}

//read shuju
//GetState(key string) ([]byte, error)
func (this *Test) get(stub shim.ChaincodeStubInterface, key string) peer.Response {
	data, err := stub.GetState(key)

	//chuxianyichang
	if err != nil {
		return shim.Error(err.Error())
	}
	//kongshuju
	if data == nil {
		return shim.Error("stub shim.ChaincodeStubInterfac")
	}

	return shim.Success(data)
}

//xieshuji
//	PutState(key string, value []byte) error
func (this *Test) set(stub shim.ChaincodeStubInterface, key string, value []byte) peer.Response {
	err := stub.PutState(key, value)
	if err != nil {
		shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func main() {
	shim.Start(new(Test))
}
