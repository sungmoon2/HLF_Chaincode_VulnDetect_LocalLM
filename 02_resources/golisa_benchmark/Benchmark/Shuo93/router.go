package main

import (
	"fmt"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// BloomFateChaincode is the chaincode of bloom fate
type BloomFateChaincode struct{}

// Init callback representing the invocation of a chaincode
func (t *BloomFateChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Printf("Init method is called.")
	return shim.Success(nil)
}

// Invoke callback representing the invocation of a chaincode
func (t *BloomFateChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if strings.HasPrefix(function, "query") {
		return t.query(stub, function[len("query"):], args[0])
	}
	jsonArgs := args[0]
	switch function {
	case "register":
		return t.register(stub, args[0])
	case "login":
		return t.login(stub, args[0])

	case "confirmDate":
		return t.confirmDate(stub, jsonArgs)
	case "measureCredit":
		return t.measureCredit(stub, jsonArgs)

	case "like":
		return t.like(stub, jsonArgs)
	case "unlike":
		return t.unlike(stub, jsonArgs)

	case "modifyPersonInfo":
		return t.modifyPersonInfo(stub, jsonArgs)
	case "uploadPersonInfo":
		return t.uploadPersonInfo(stub, jsonArgs)

	case "replyDate":
		return t.replyDate(stub, jsonArgs)
	case "sendDate":
		return t.sendDate(stub, jsonArgs)

	case "replyPermession":
		return t.replyPermession(stub, jsonArgs)
	case "sendPermission":
		return t.sendPermission(stub, jsonArgs)
	default:
		return shim.Error("The function has not been implemented")
	}
}

func (t *BloomFateChaincode) query(stub shim.ChaincodeStubInterface, function string, args string) pb.Response {
	switch function {
	case "PersonInfo":
		return t.queryPersonInfo(stub, args)
	case "PersonList":
		return t.queryPersonList(stub, args)
	case "LikeList":
		return t.queryLikeList(stub, args)
	case "Permission":
		return t.queryPermession(stub, args)
	case "Date":
		return t.queryDate(stub, args)
	case "Credit":
		return t.queryCredit(stub, args)
	case "ModifyRecordByTime":
		// todo
		return t.queryModifyRecordByTime(stub, args)
	case "ModifyRecord":
		return t.queryModifyRecord(stub, args)
	case "CreditBalance":
		return t.queryCreditBalance(stub, args)
	case "PublicKey":
		return t.queryPublicKey(stub, args)
	case "EncryptedContent":
		return t.queryEncryptedContent(stub, args)
	default:
		return shim.Error("The function has not been implemented")
	}
}
