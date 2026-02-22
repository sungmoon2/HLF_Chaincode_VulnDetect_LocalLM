package contractdb

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type Contract struct {

	ObjectType string `json:"object_type"`
	ContractID string  `json:"contract_id"`
	TimeStampID int64 `json:"time_stamp_id"`
	ContractJson interface{} `json:"contract_json"`

	//compositekey
	ObjectKey string `json:"object_key"`

}


func (contract *Contract) Init(stub shim.ChaincodeStubInterface)  peer.Response {

	args := stub.GetStringArgs()
	if len(args) != 0 {
		return shim.Error("init parameter error!")
	}

	return shim.Success(nil)
}

func (contract *Contract) Invoke(stub shim.ChaincodeStubInterface) peer.Response  {

	re, args := stub.GetFunctionAndParameters()

	switch re {

	//template function
	//delete
	case "deleteContract" :
		return contract.DeleteContract(stub,args)
	//query
	case "queryContractState":
		return contract.QueryContractState(stub,args)
	case "queryType":
		return contract.querybyContractType(stub,args)
	case "querycomkeys":
		return contract.querybyComkey(stub,args)
	case "querydate":
		return contract.querybyContractDate(stub,args)

	//invoke contract putstate
	case "invokeContract":
		return contract.InvokeContract(stub,args)

	//modify
	case "modifyContract":
		return contract.ModifyContract(stub,args)


	//template function:couchdb
	//invoke
	case "setdbkey":
		return contract.SetQuerykey(stub,args)
	case "setdbkeys":
		return contract.SetQuerykeys(stub,args)
	//query
	case "queryrichkey":
		return contract.queryrichkey(stub,args)
	case "queryrichkeys":
		return contract.queryrichkeys(stub,args)


	//business function

	default:
		return shim.Error("Unknown func type while Invoke, please check")
	}
	return shim.Success(nil)
}