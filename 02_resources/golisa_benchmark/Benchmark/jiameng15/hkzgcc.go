/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at
  http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// 环科币实现
type HKCoinChaincode struct {
}

// ============================================================================================================================
// Asset Definitions - The ledger will store marbles and owners
// 资产定义 - 账本将会储存以下表
// ============================================================================================================================

// ----- User-员工构造 ----- //
type User struct {
	ObjectType string        `json:"docType"`    //field for couchdb 这里不知道干啥的
	Id         string        `json:"id"`         // 员工id
	Name       string        `json:"name"`       // 员工姓名
	Department string        `json:"department"` // 部门
	Role       string        `json:"role"`       // 职位
	Asset      int           `json:"asset"`      // 资产
	List       []Transaction `json:"list"`       // 员工交易记录
	CreateTime string        `json:"createTime"` // 创建时间
	UpdateTime string        `json:"updateTime"` // 修改时间
}

// ----- TXLogs-交易记录 ----- //
type Transaction struct {
	ObjectType string `json:"docType"`   //field for couchdb 这里不知道干啥的
	TxId       string `json:"txId"`      // 交易id
	Content    string `json:"content"`   // 交易内容
	Type       string `json:"type"`      // 交易类型，
	Relation   string `json:"relation"`  // 交易用户ID
	Volume     string `json:"volume"`    // 交易量，有 + - 值
	TimeStamp  string `json:"timeStamp"` // 时间戳
}

// ----- Rank-排行榜构造/持币信息 ----- //
type Rank struct {
	ObjectType string `json:"docType"` //field for couchdb 这里不知道干啥的
	// Total      int    `json:"total"`      // 合计
	Rank       int    `json:"rank"`       // 员工排名
	Id         string `json:"id"`         // 员工id
	Name       string `json:"name"`       // 员工姓名
	Department string `json:"department"` // 部门
	Role       string `json:"role"`       // 职位
	Asset      int    `json:"asset"`      // 资产
	// TimeStamp  string `json:"timeStamp"`  // 时间戳
}

// ============================================================================================================================
// Main 主程序
// ============================================================================================================================
func main() {
	err := shim.Start(new(HKCoinChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode - %s", err)
	}
}

// ============================================================================================================================
// Init - initialize the chaincode
//
// Marbles does not require initialization, so let's run a simple test instead.
//
// Shows off PutState() and how to pass an input argument to chaincode.
// Shows off GetFunctionAndParameters() and GetStringArgs()
// Shows off GetTxID() to get the transaction ID of the proposal
//
// Inputs - Array of strings
//  ["314"]
//
// Returns - shim.Success or error
//Init()方法，实例化
// ============================================================================================================================
func (t *HKCoinChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("环科币链码实例化")
	funcName, args := stub.GetFunctionAndParameters()
	var number int
	var err error
	txId := stub.GetTxID()

	fmt.Println("Init() is running")
	fmt.Println("Transaction ID:", txId)
	fmt.Println("  GetFunctionAndParameters() function:", funcName)
	fmt.Println("  GetFunctionAndParameters() args count:", len(args))
	fmt.Println("  GetFunctionAndParameters() args found:", args)

	// store compatible marbles application version
	err = stub.PutState("HKCoin", []byte("1.0.0"))
	if err != nil {
		return shim.Error(err.Error())
	}

	//创建一个Admin账户,资产10亿
	var admin User
	admin.ObjectType = "hk_user"
	admin.Id = "u0"
	admin.Name = "Admin"
	admin.Department = "公司账户"
	admin.Role = "公司账户"
	admin.Asset = 1000000000
	admin.CreateTime = "2018-08-20 18:00:00"
	admin.UpdateTime = "none"

	//store admin
	adminAsBytes, _ := json.Marshal(admin)      //convert to array of bytes
	err = stub.PutState(admin.Id, adminAsBytes) //store admin by its Id
	if err != nil {
		fmt.Println("Could not store admin")
		return shim.Error(err.Error())
	}

	fmt.Println("Ready for action") //self-test pass
	return shim.Success(nil)
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// Invoke - 调用的入口
// ============================================================================================================================
func (t *HKCoinChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println(" ")
	fmt.Println("starting invoke, for - " + function)

	// Handle different functions
	// 调用不同的函数，这里面很多的没有用啊，设计一下
	if function == "init" { //initialize the chaincode state, used as reset
		return t.Init(stub)
	} else if function == "read" { //generic read ledger
		return read(stub, args)
	} else if function == "write" { //generic writes to ledger
		return write(stub, args)
	} else if function == "delete_marble" { //deletes a marble from state
		return delete_marble(stub, args)
	} else if function == "init_user" { //创建一个新User
		return init_user(stub, args)
	} else if function == "transaction" { //环科币交易
		return init_user(stub, args)
	} else if function == "read_everything" { //read everything, (owners + marbles + companies)
		return read_everything(stub)
	} else if function == "getHistory" { //read history of a marble (audit)
		return getHistory(stub, args)
	} else if function == "getMarblesByRange" { //read a bunch of marbles by start and stop id
		return getMarblesByRange(stub, args)
	} else if function == "disable_owner" { //disable a marble owner from appearing on the UI
		return disable_owner(stub, args)
	}

	// error out
	fmt.Println("Received unknown invoke function name - " + function)
	return shim.Error("Received unknown invoke function name - '" + function + "'")
}

// ============================================================================================================================
// Query - legacy function
// ============================================================================================================================
func (t *HKCoinChaincode) Query(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Error("Unknown supported call - Query()")
}
