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

//两公示一公告名单

import (
	"fmt"
	//"strconv"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var logger = shim.NewLogger("publicList")

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

//申请名单结构
type personInfo struct{
	Applyhash 	string    `json:"Applyhash"`   //申请人全部信息（Hostname+IdentID+BankAccount）的hash
	Hostname 	string	  `json:"Hostname"`    //申请人姓名
	IdentID 	string    `json:"IdentID"`     //申请人身份证号
	MemNum 		string    `json:"MemNum"`      //申请人家庭成员数
	BankAccount 	string    `json:"BankAccount"` //申请人银行账户
	PerIncome 	string    `json:"PerIncome"`   //家庭人均收入
	LabourNum 	string    `json:"LabourNum"`   //劳动力人口
	Status 		string    `json:"Status"`      //状态
	HandlerID	string    `json:"HandlerId"`   //操作人员身份证号
}
// 两个参数：[0]名单时间；[1]操作人ID
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### publicList Init ###########")
	fmt.Println("Init")
	_, args := stub.GetFunctionAndParameters()
	var tableTime string  // time of the publicity list
	var operatorID string // ID of the operator
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2.")
	}

	// Initialize the chaincode
	tableTime = args[0]
	operatorID = args[1]

	// Write the state to the ledger
	err = stub.PutState(tableTime, []byte(operatorID))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

//有2个function：[0]申请人ID，[1]申请人hash信息，[3]操作人ID
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("publicList Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "update" {
		return t.update(stub, args)
	}else if function == "query"{
		return t.query(stub, args)
	}else if function == "add"{
		return t.add(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"add\" \"update\"")
}
func (t *SimpleChaincode) add(stub shim.ChaincodeStubInterface, args []string) pb.Response {
  	if len(args) != 2 {
	   return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	applicantID := args[0]
	personalInfo := args[1]
	
	TempHashval,_:= stub.GetState(applicantID)

        if TempHashval != nil {
           return shim.Error("This ID already exists")
        }
		// Write the state back to the ledger
	err := stub.PutState(applicantID, []byte(personalInfo))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}
func (t *SimpleChaincode) update(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	applicantID := args[0]
	personalInfo := args[1]
	
	TempHashval,errs:= stub.GetState(applicantID)
	if errs != nil {
		return shim.Error("list is not here")
	}
    if TempHashval != nil {
        return shim.Error("This ID already exists")
    }
		// Write the state back to the ledger
	err := stub.PutState(applicantID, []byte(personalInfo))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}


// Query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var applicantID string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting identy ID of the person.")
	}

	applicantID = args[0]

	// Get the state from the ledger
	personalInfo, err := stub.GetState(applicantID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + applicantID + "\"}"
		return shim.Error(jsonResp)
	}

	if personalInfo == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + applicantID + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + applicantID + "\",\"Amount\":\"" + string(personalInfo) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(personalInfo)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
