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

//慈善智能合约

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	//"strings"
)

var logger = shim.NewLogger("chariB")

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

//新增走访记录信息
//type Ncharity struct{
//CharityHash string
//VisitInf Visit
//}
//走访机构信息（民政局）
type Visit struct {
	Organization string `json:"Organization"` //民政局名称
	Result       string `json:"Result"`       //走访结果
	VTime        string `json:"VTime"`        //走访时间
	Comment      string `json:"Comment"`      //备注
	HandlerID    string `json:"HandlerID"`	  //处理人员ID
}

//捐助情况
type Sum struct {
	SOrganization string `json:"SOrganization"` //慈善机构名称
	Money         string `json:"Money"`
	Reason        string `json:"Reason"`
	STime         string `json:"STime"`
	HandlerID     string `json:"HandlerID"`	    //处理人员ID
}

//慈善信息结构
type ChariInf struct {
	CharityHash string  `json:"CharityHash"` //所有信息的hash
	Name        string  `json:"Name"`        //姓名
	//TotalSum    string  `json:"TotalSum"`    //慈善捐助总金额
	VisitInf    []Visit `json:"VisitInf"`    //走访信息
	ChSum       []Sum   `json:"ChSum"`       //慈善机构捐助具体信息
	
}

//输入参数：“init”，two function:add,update.[0]是操作人员ID
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### chariB Init ###########")
	return shim.Success(nil)
}

//输入参数：function：“add(或update)”，[0]是接受慈善帮助人员的ID，[1]是json格式，[2]是操作人ID
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
    function, args := stub.GetFunctionAndParameters()

	// Perform the execution

	// Write the state back to the ledger
	if function == "add" {
		// Perform the execution
		return t.add(stub, args)
	}
	//更新信息
	if function == "update" {
		return t.update(stub, args)
	}
	//新增走访记录
	if function == "addVisit" {
		return t.addVisit(stub, args)
	}
	//新增捐赠记录
	if function == "addDonate" {
		return t.addDonate(stub, args)
	}

	//
	if function == "query"{
		return t.query(stub,args)
	}

	return shim.Error("Received unknown function invocation")

}

// 查询，输入参数:[0]是接受慈善对象的身份证号
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	var IdentID string // Entities
	var err error

	IdentID = args[0]

	// Get the state from the ledger
	Hashval, err := stub.GetState(IdentID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + IdentID + "\"}"
		//return nil, errors.New(jsonResp)
		return shim.Error(jsonResp)
	}

	if Hashval == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + IdentID + "\"}"
		//return nil, errors.New(jsonResp)
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + IdentID + "\",\"Amount\":\"" + string(Hashval) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)

	var searchResult []string
	json.Unmarshal(Hashval, &searchResult)
	//return Hashval, nil
	return shim.Success(Hashval)
}

//添加申请记录
//"args":["123456789","{\"CharityHash\":\"aaaaaaaaaaa\",\"Name\":\"张三\",\"VisitInf\":[],\"ChSum\":[]}"]
func (t *SimpleChaincode) add(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    if len(args) != 2 {
		//return nil, errors.New("Incorrect number of arguments. ")
		return shim.Error("Incorrect number of arguments.")
	}
	IdentID := args[0]
	Hashval := args[1]
	//DealerID := args[2]
	//fmt.Printf("DealerID:%s\n", DealerID)
	//DealerId := args[2]
	//OperatorId := args[2]
	//Hashval = strings.Replace(Hashval, "\\", "", -1)
	// Hashval := string(args[1])
	//var chariInf ChariInf
	//var visitInf Visit //输入的参数是[1]:{"Organization": "慈善机构", "Result": "捐助","VTime":"2017-01-05","Comment":"无"}
	var err error

    TempHashval, err := stub.GetState(IdentID)

	if TempHashval != nil {
		//return nil, errors.New("This ID already exists")
		return shim.Error("This ID already exists")
	}
	// Write the state back to the ledger
	err = stub.PutState(IdentID, []byte(Hashval))
	if err != nil {
		//return []byte(Hashval), err
		return shim.Error("Failed to get state")
	}
	return shim.Success(TempHashval)
}

func (t *SimpleChaincode) update(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    if len(args) != 2 {
		//return nil, errors.New("Incorrect number of arguments. ")
		return shim.Error("Incorrect number of arguments in function update")
	}
	IdentID := args[0]
	Hashval := args[1]
	//DealerID := args[2]
	//fmt.Printf("DealerID:%s\n", DealerID)
	//Hashval = strings.Replace(Hashval, "\\", "", -1)
	// Hashval := string(args[1])
	//var chariInf ChariInf
	//var visitInf Visit //输入的参数是[1]:{"Organization": "慈善机构", "Result": "捐助","VTime":"2017-01-05","Comment":"无"}
	var err error
	HashvalTemp, errs := stub.GetState(IdentID)

	if errs != nil {
		//return nil, errors.New("list is not here")
		return shim.Error("list is not here")
	}
	if HashvalTemp == nil {
		//return nil, errors.New("Entity not found")
		return shim.Error("Entity not found")
	}

	err = stub.PutState(IdentID, []byte(Hashval))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("Saved!"))
}

func (t *SimpleChaincode) addVisit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    if len(args) != 2 {
		return shim.Error("Incorrect number of arguments.")
	}
	IdentID := args[0]

	//Hashval = strings.Replace(Hashval, "\\", "", -1)
	Hashval := string(args[1])
	//DealerID := args[2]
	//fmt.Printf("DealerID:%s\n", DealerID)
	var chariInf ChariInf
	var visitInf Visit //输入的参数是[1]:{"Organization": "慈善机构", "Result": "捐助","VTime":"2017-01-05","Comment":"无"}
	var err error

    //TempHashval, err := stub.GetState(IdentID)
	HashvalTemp, err := stub.GetState(IdentID)

	if err != nil {
		//return nil, errors.New("list is not here")
		return shim.Error("list is not here")
	}
	if HashvalTemp == nil {
		return shim.Error("Entity not found")
	}

	// charT := Visit{
	// 	Organization: "china",    //慈善机构名称
	// 	Result:       "help",     //走访结果
	// 	VTime:        "20170101", //走访时间
	// 	Comment:      "no",       //备注
	// }

	json.Unmarshal(HashvalTemp, &chariInf)
	json.Unmarshal([]byte(Hashval), &visitInf)

	//修改哈希

	//chariInf.CharityHash = visitInf.CharityHash
	//新增走访记录
	chariInf.VisitInf = append(chariInf.VisitInf, visitInf)

	jsonchari, _ := json.Marshal(chariInf)
	err = stub.PutState(IdentID, []byte(jsonchari))
	if err != nil {
		//return jsonchari, err
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *SimpleChaincode) addDonate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    if len(args) != 2 {
		return shim.Error("Incorrect number of arguments.")
	}
	IdentID := args[0]

	//Hashval = strings.Replace(Hashval, "\\", "", -1)
	Hashval := string(args[1])
	//DealerID := args[2]
	//fmt.Printf("DealerID:%s\n", DealerID)
	var chariInf ChariInf
	var chSum Sum //输入的参数是[1]:{"Organization": "慈善机构", "Result": "捐助","VTime":"2017-01-05","Comment":"无"}
	var err error

    //TempHashval, err := stub.GetState(IdentID)
	HashvalTemp, err := stub.GetState(IdentID)

	if err != nil {
		//return nil, errors.New("list is not here")
		return shim.Error("list is not here")
	}
	if HashvalTemp == nil {
		return shim.Error("Entity not found")
	}

	// charT := Visit{
	// 	Organization: "china",    //慈善机构名称
	// 	Result:       "help",     //走访结果
	// 	VTime:        "20170101", //走访时间
	// 	Comment:      "no",       //备注
	// }

	json.Unmarshal(HashvalTemp, &chariInf)
	json.Unmarshal([]byte(Hashval), &chSum)

	//修改哈希

	//chariInf.CharityHash = visitInf.CharityHash
	//新增捐赠记录
	chariInf.ChSum = append(chariInf.ChSum, chSum)

	jsonchari, _ := json.Marshal(chariInf)
	err = stub.PutState(IdentID, []byte(jsonchari))
	if err != nil {
		//return jsonchari, err
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
