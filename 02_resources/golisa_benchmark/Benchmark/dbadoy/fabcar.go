package main

import (
	"fmt"

	"mesg"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type SmartContract struct {
	fwm map[string]mesg.FuncWrap
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()

	m, exist := s.fwm[function]
	if !exist {
		return shim.Error("Invalid Smart Contract function name.")
	}

	err := m.param.Set(args)
	if err != nil {
		return shim.Error(err)
	}

	return m.Proceed(APIstub, m.param)
}

func main() {
	smt := new(SmartContract)
	smt.fwm = mesg.Initial()

	err := shim.Start(smt)
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
