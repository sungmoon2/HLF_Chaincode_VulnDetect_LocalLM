/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

//募资结构

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var logger = shim.NewLogger("fundRaise")

//初始化的时候
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	logger.Info("########### fundRaise Init ###########")
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
    function, args := stub.GetFunctionAndParameters()
    if function == "add" {
		return t.add(stub,args)	
	}
	if function == "update" {
		return t.update(stub, args)
	}
	if function == "query" {
		return t.query(stub, args)
	}

	return shim.Error("Received unknown function invocation")
}

//添加募资结构传入参数有6个：募资结构编号，计划募资总金额，第一顺位（json字符串），第二顺位，第三顺位，操作人编号。顺序以这个为准。
func (t *SimpleChaincode) add(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var fundRaisingID string	//募资结构编号
	var Sum string	//计划募资总金额
	var Prority1 string	//第一顺位
	var Prority2 string	//第二顺位
	var Prority3 string	//第三顺位
	var HandlerId string //操作人编号

	var err error

	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6 ")
	}

	// Initialize the chaincode
	fundRaisingID = args[0]
	Sum = args[1]
	Prority1 = args[2]
	Prority2 = args[3]
	Prority3 = args[4]
	HandlerId = args[5]

    TempHashval, err := stub.GetState(fundRaisingID)

	if TempHashval != nil {
		return shim.Error("This fundRaisingID already exists")
	}
	// Write the state to the ledger
	err = stub.PutState(fundRaisingID, []byte(fundRaisingID))
	if err != nil {
		return shim.Error("Failed to get fundRaisingID")
	}

	err = stub.PutState("Sum", []byte(Sum))
	if err != nil {
		return shim.Error("Failed to get Sum")
	}

	err = stub.PutState("Prority1", []byte(Prority1))
	if err != nil {
		return shim.Error("Failed to get Prority1")
	}

	err = stub.PutState("Prority2", []byte(Prority2))
	if err != nil {
		return shim.Error("Failed to get Prority2")
	}

	err = stub.PutState("Prority3", []byte(Prority3))
	if err != nil {
		return shim.Error("Failed to get Prority3")
	}

	err = stub.PutState("HandlerId", []byte(HandlerId))
	if err != nil {
		return shim.Error("Failed to get HandlerId")
	}

	return shim.Success(TempHashval)
}

//更新募资结构传入参数有6个：募资结构编号，计划募资总金额，第一顺位（json字符串），第二顺位，第三顺位，操作人编号。
func (t *SimpleChaincode) update(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var newFundRaisingID string	//募资结构编号
	var newSum string	//计划募资总金额
	var newPrority1 string	//第一顺位
	var newPrority2 string	//第二顺位
	var newPrority3 string	//第三顺位
	var HandlerId string //操作人编号

	var err error

	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 6 ")
	}

	// Initialize the chaincode
	newFundRaisingID = args[0]
	newSum = args[1]
	newPrority1 = args[2]
	newPrority2 = args[3]
	newPrority3 = args[4]
	HandlerId = args[5]

	HashvalTemp, errs := stub.GetState(newFundRaisingID)

	if errs != nil {
		//return nil, errors.New("list is not here")
		return shim.Error("list is not here")
	}	
	if HashvalTemp == nil {
		//return nil, errors.New("Entity not found")
		return shim.Error("Entity not found")
	}
	// Write the state to the ledger
	err = stub.PutState("fundRaisingID", []byte(newFundRaisingID))
	if err != nil {
		return shim.Error("Failed to get fundRaisingID")
	}

	err = stub.PutState("Sum", []byte(newSum))
	if err != nil {
		return shim.Error("Failed to get Sum")
	}

	err = stub.PutState("Prority1", []byte(newPrority1))
	if err != nil {
		return shim.Error("Failed to get Prority1")
	}

	err = stub.PutState("Prority2", []byte(newPrority2))
	if err != nil {
		return shim.Error("Failed to get Prority2")
	}

	err = stub.PutState("Prority3", []byte(newPrority3))
	if err != nil {
		return shim.Error("Failed to get Prority3")
	}

	err = stub.PutState("HandlerId", []byte(HandlerId))
	if err != nil {
		return shim.Error("Failed to get HandlerId")
	}

	return shim.Success([]byte("update!"))
}


// Query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response{

	var oneKey string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	oneKey = args[0]

	// Get the state from the ledger
	KeyInfo, err := stub.GetState(oneKey)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + oneKey + "\"}"
		return shim.Error(jsonResp)
	}

	if KeyInfo == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + oneKey + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + oneKey + "\",\"Amount\":\"" + string(KeyInfo) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(nil)

}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
