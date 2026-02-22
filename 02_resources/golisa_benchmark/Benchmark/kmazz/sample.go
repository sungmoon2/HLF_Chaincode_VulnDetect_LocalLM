/* Sample chaincode  */

package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

/*
var invokes = map[string] func {
	"test" : func(){
	},
	"test2" : func(){
	}
}*/

type SampleChaincode struct {
}

/***************************************************
[Init]
description : initialize
parameters  :
   stub - chaincode interface
return: response object
***************************************************/
func (t *SampleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("SampleChaincode.Init")
	_, args := stub.GetFunctionAndParameters()
	ret := t.setData(stub, args)
	if len(ret) != 0 {
		return shim.Error(ret)
	}
	return shim.Success(nil)
}

/***************************************************
[Invoke]
description : invoke
parameters  :
   stub - chaincode interface
return: response object
***************************************************/
func (t *SampleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("SampleChainCode.Invoke")
	funcname, args := stub.GetFunctionAndParameters()
	if funcname == "set" {
		ret := t.setData(stub, args)
		if len(ret) != 0 {
			return shim.Error(ret)
		}
		return shim.Success(nil)
	} else if funcname == "get" {
		val, ret := t.getData(stub, args)
		if len(ret) != 0 {
			return shim.Error(ret)
		}
		return shim.Success(val)
	}
	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\"")
}

/***************************************************
[setData]
description : set key-value pairs to blockchain
parameters  :
   args - ket-value pairs
return: error string if error occured
***************************************************/
func (t *SampleChaincode) setData(stub shim.ChaincodeStubInterface, args []string) string {
	var k, v string
	if (len(args) % 2) != 0 {
		return "Number of parameter must be even number."
	}
	for i := 0; i < len(args); i += 2 {
		k = args[i]
		v = args[i+1]
		fmt.Printf("key:%s, value:%s\n", k, v)
		err := stub.PutState(k, []byte(v))
		if err != nil {
			return err.Error()
		}
	}
	return ""
}

/***************************************************
[getData]
description : get value correspond to the key
parameters  :
   args - key value
return: error string if error occured
***************************************************/
func (t *SampleChaincode) getData(stub shim.ChaincodeStubInterface, args []string) ([]byte, string) {
	if len(args) != 1 {
		return []byte(""), "Number of parameter must be even number."
	}
	key := args[0]
	val, err := stub.GetState(key)
	fmt.Printf("GetState:%s\n", val)
	if err != nil {
		return []byte(""), "Invalid key"
	}
	return val, ""
}

func main() {
	err := shim.Start(new(SampleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
