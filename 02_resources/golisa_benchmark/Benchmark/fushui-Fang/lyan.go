package lyan

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type MyChaincode struct {
}

func (t *MyChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debugf("My chaincode Init")
	function, args := stub.GetFunctionAndParameters()
	if function != "init" {
		return shim.Error(ErrInvalidArgs.Error())
	}

	//在这里考虑底层链码要实现什么功能
	err := t.initEnv(stub, args)
	if err != nil {
		logger.Error(ErrInit.Error())
		return shim.Error(ErrInit.Error())
	}

	return shim.Success(nil)
}

func (t *MyChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if len(args) == 0 {
		logger.Error(ErrInvalidArgs.Error())
		return shim.Error(ErrInvalidArgs.Error())
	}
	logger.Debugf(" invoke function [%v] args [%v]", function, args)
	s := makeChainCodeHub(stub, args)

	switch function {
	case "createContractAcount":
		return s.createContractAcount()
		//创建合约账户

	case "queryAcccount":
		return s.queryAccount()
		//查询账户信息

	case "transfer":
		return s.transfer()
		//价值转移，先写未完成版

	case "queryAsgyByID":
		return s.querySgyByID()
		//通过ID 和地址查询某一分配策略

	case "queryCurrentSgyID":
		return s.queryCurrentSgyID()
		//查询某一合约账户的策略ID

	case "modifyAllocateByUser":
		return s.modifyAllocateByUser()

	default:
		logger.Error(ErrInvalidFunction.Error())
		return shim.Error(ErrInvalidFunction.Error())
	}

	return shim.Error(ErrInvalidFunction.Error())
}
