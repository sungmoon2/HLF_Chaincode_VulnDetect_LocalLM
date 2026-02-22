/*
Copyright Tigerwit Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("sealtxnew")

type SealTX struct {
}

func (t *SealTX) seal(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	if len(args) != 2 {
		return shim.Error(fmt.Sprintf("Incorrect number of arguments. Expecting 2,received %d", len(args)))
	}
	key := args[0]
	value := []byte(args[1])
	err = stub.PutState(key, value)
	if err != nil {
		return shim.Error(err.Error())
	}
	logger.Debug(fmt.Sprintf("successfully putstate===>%s", args[0]))
	return shim.Success(nil)
}

func (t *SealTX) querybykey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	if len(args) != 1 {
		return shim.Error(fmt.Sprintf("Incorrect number of arguments. Expecting 1,received %d", len(args)))
	}
	key := args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Querybykey Err: %v", err))
	} else if valAsbytes == nil {
		args4old := [][]byte{[]byte("querybykey"),[]byte(key)}
		return stub.InvokeChaincode("sealtx", args4old, "tradechannel")
	}
	return shim.Success(valAsbytes)
}

func (t *SealTX) history(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	if len(args) != 1 {
		return shim.Error(fmt.Sprintf("Incorrect number of arguments. Expecting 1,received %d", len(args)))
	}
	key := args[0]
	iter, err := stub.GetHistoryForKey(key)
	defer iter.Close()
	if err != nil {
		return shim.Error("get iter fail " + err.Error())
	}
	values := []string{}

	for iter.HasNext() {
		fmt.Printf("next\n")
		if kv, err := iter.Next(); err == nil {
			fmt.Printf("id: %s value: %s\n", kv.TxId, kv.Value)
			//operate_time := time.Unix(kv.Timestamp.Seconds, 0).Format("2006-01-02 15:04:05")
			values = append(values, fmt.Sprintf("TxId:[%s]Value:[%s]TimeStamp:[%v]", kv.TxId, kv.Value, kv.Timestamp))
		}
		if err != nil {
			return shim.Error("iterator history fail: " + err.Error())
		}
	}

	bytes, err := json.Marshal(values)
	if err != nil {
		return shim.Error("json marshal fail: " + err.Error())
	}

	return shim.Success(bytes)
}

func (t *SealTX) gettxidspec(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	if len(args) != 2 {
		return shim.Error(fmt.Sprintf("Incorrect number of arguments. Expecting 2,received %d", len(args)))
	}
	key := args[0]
	value := args[1]
	iter, err := stub.GetHistoryForKey(key)
	defer iter.Close()
	if err != nil {
		return shim.Error("get iter fail " + err.Error())
	}

	for iter.HasNext() {
		fmt.Printf("next\n")
		if kv, err := iter.Next(); err == nil {
			fmt.Printf("id: %s value: %s\n", kv.TxId, kv.Value)
			//operate_time := time.Unix(kv.Timestamp.Seconds, 0).Format("2006-01-02 15:04:05")
			if value == string(kv.Value) {
				return shim.Success([]byte(kv.TxId))
			}
		}
		if err != nil {
			return shim.Error("iterator history fail: " + err.Error())
		}
	}
	args4old := [][]byte{[]byte("gettxidspec"), []byte(key), []byte(value)}
	return stub.InvokeChaincode("sealtx", args4old, "tradechannel")
}
func (s *SealTX) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Init Chaincode SealTX")
	return shim.Success(nil)
}

func (s *SealTX) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	logger.Debug(fmt.Sprintf("func:%s", function))
	logger.Debug(fmt.Sprintf("args:%v", args))
	switch function {
	case "seal":
		return s.seal(stub, args)
	case "querybykey":
		return s.querybykey(stub, args)
	case "history":
		return s.history(stub, args)
	case "gettxidspec":
		return s.gettxidspec(stub, args)
	default:
		return shim.Error("unsupported function name: " + function)
	}
}

func main() {
	logger.SetLevel(shim.LogDebug)
	err := shim.Start(new(SealTX))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
