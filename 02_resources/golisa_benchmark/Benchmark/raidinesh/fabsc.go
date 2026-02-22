/*This file  is having  Three API for main FabSc Smart Contract , instantiated from main.go
Function "StatusCheck"  is used to check running status of Smart Contract
Function "CreateGenAssets" is used to add any Asset of format:{"assetName": string,"keys":[]string,"entityCount":int32,"assetDatas":[]string}
Parameter Example:
{"assetName":"IOTdevice","keys":["id"],"entityCount":1,"assetDatas":["{\"id\":\"01\",\"firmware_version\":\"1.0.1\",
\"serial_number\":\"SN-001\",\"SensorReading\":\"High\",\"Value\":\"200F\"}"]}
Function "ListGenAssets" is used to List assets created from "CreateAsset" Function
Parameter Example:
{"assetName":"IOTdevice"}
*/
package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

//FuncTemplate : This Function is generic Tamplate for all functions to Fabric Smart contract
type FuncTemplate func(stub shim.ChaincodeStubInterface, args []string) pb.Response

//FabSc : This  structure will store function pointers  for  all function executed by this smart contract
type FabSc struct {
	funcMap      map[string]FuncTemplate
	restartcheck bool
}

//Result is structure to send test result to check if chaincode is deployed and running successfully
type Result struct {
	Status string
}

//Constent definition for  all function names
const (
	SC  string = "StatusCheck"
	CAS string = "CreateGenAssets"
	LAS string = "ListGenAssets"
)

//initfunMap():Chaincode initialization Function -This function will create a  map  and  initialized it at smart contract init phase
func (inv *FabSc) initfunMap() {
	inv.funcMap = make(map[string]FuncTemplate)
	inv.funcMap[SC] = ChainCodeStatusCheck
	inv.funcMap[CAS] = CreateGenAssets
	inv.funcMap[LAS] = ListGenAssets
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (inv *FabSc) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Infof("Init ChaininCode Fabsc")
	inv.initfunMap()
	inv.restartcheck = true
	logger.Debugf("%+v", inv)
	return shim.Success(nil)

}

// Invoke is called  for addting/purging one or more asset per transaction from  smart contract ledger. Each transaction is
// either a 'get' or a 'set' .
func (inv *FabSc) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Infof("Invoke ChaininCode FabSc")
	funname, args := stub.GetFunctionAndParameters()
	if funname == "" {
		logger.Errorf("Function Name is not passed correctly while invoking ChainCode")
	}
	if inv.restartcheck == false {
		inv.initfunMap()
		inv.restartcheck = true
		logger.Debugf("%+v", inv)
	}
	exefun, ok := inv.funcMap[funname]
	logger.Infof("Invoke ChaininCode FabSc for Function Name: %s", funname)
	if ok {
		return exefun(stub, args)
	}
	logger.Errorf("Function Name:= %s is not defined in ChaininCode", funname)
	return shim.Error(fmt.Sprintf("Invalid Function Name: %s", funname))
}

// ChainCodeStatusCheck function is called by client App after intsalling and instantiating chaincode to check if it is up and running
func ChainCodeStatusCheck(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Debugf("ChaininCode  Running Status Check")
	result := Result{}
	result.Status = "ChainCode Running Successfully"
	availabeByte, _ := json.Marshal(result)
	logger.Debugf("ChaininCode  Running Status json data: %v", availabeByte)
	return shim.Success(availabeByte)
}
