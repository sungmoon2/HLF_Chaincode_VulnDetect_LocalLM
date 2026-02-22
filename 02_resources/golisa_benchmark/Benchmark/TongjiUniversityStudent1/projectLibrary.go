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


//项目库

import (
	//"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("projectLibrary")

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

//部署时，传入参数有2个 项目库ID 操作人ID
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### projectLibrary Init ###########")
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "delete" {
		return t.delete(stub, args)
	}else if function == "update"{
		return t.update(stub,args)
	}else if function == "add"{
		return t.add(stub,args)
	}else if function == "setLibrary"{
		return t.setLibrary(stub,args)
	}else if function == "query"{
		return t.query(stub,args)
	}

	return shim.Error("Received unknown function invocation")
}

//传入参数有2个 项目库ID 操作人ID
func (t *SimpleChaincode) setLibrary(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var ProjectLibID string	//项目库ID
	var HandlerID string //操作人编号
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// Initialize the chaincode
	ProjectLibID = args[0]
	HandlerID = args[1]

    TempHashval, err := stub.GetState(ProjectLibID)

	if TempHashval != nil {
		return shim.Error("This LibraryID already exists")
	}
	// Write the state to the ledger
	err = stub.PutState(ProjectLibID, []byte(HandlerID))
	if err != nil {
		return shim.Error("Failed to get state")
	}
	return shim.Success(TempHashval)
}

//更新项目库中的项目 传入参数有2个 更新的项目的ID，更新的项目的信息
//变量名newProjectHash解释，这个里面有个hash，不要理解错了，这是因为原来设计的时候是要存项目信息的hash，而现在的设计是要存项目全信息
func (t *SimpleChaincode) update(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var newProjectID string	//更新的项目的ID
	var newProjectHash string	//更新的项目的的信息

	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// Initialize the chaincode
	newProjectID = args[0]
	newProjectHash = args[1]

	// Write the state to the ledger
	err = stub.PutState(newProjectID, []byte(newProjectHash))
	if err != nil {
		return shim.Error("Failed to get state")
	}

	return shim.Success(nil)
}

//增加项目库中的项目 传入参数有2个 项目ID，项目信息（json字符串）
func (t *SimpleChaincode) add(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var ProjectID string	//项目ID
	var ProjectHash string	//项目信息hash

	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// Initialize the chaincode
	ProjectID = args[0]

	//添加时需要判断该项目ID对应的值是否已经存在，防止重复添加
	TemProjectHashByte, err := stub.GetState(ProjectID)
	if TemProjectHashByte != nil {
		return shim.Error("The Project is existed")
	}

	ProjectHash = args[1]

	// Write the state to the ledger
	err = stub.PutState(ProjectID, []byte(ProjectHash))
	if err != nil {
		return shim.Error("Failed to get state")
	}

	return shim.Success(nil)
}

//删除项目库中的项目 传入参数有2个 项目ID，操作人ID
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	deleteStructureKey := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(deleteStructureKey)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response{
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
	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
