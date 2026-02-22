package main

import (
	"encoding/json"
	"fmt"

	"github.com/chainHero/heroes-service/myHospitalnockey/bean"
	//"github.com/chainHero/heroes-service/myHospitalnockey/utils"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type MyHospital struct {
}

/*func (t *MyHospital) Init(stub shim.ChaincodeStubInterface) peer.Response {

	return shim.Success(nil)
}
*/
func (t *MyHospital) Init(stub shim.ChaincodeStubInterface) peer.Response {
	var my1 bean.MyLedgerno
	args := []string{"12345678912345678912345678912345", "huxike", "这是总的结果"}
	my1.IdNumber = args[0]
	my1.Department = args[1]
	my1.FinalInfo = args[1]
	my1JsonBytes, err := json.Marshal(&my1) // Json序列化
	if err != nil {
		return shim.Error("Json serialize Compact fail while Loan")
	}
	// 生成合同联合主键(病人身份证号码,医院部门)
	key := args[0]

	// 保存合同信息
	err = stub.PutState(key, my1JsonBytes)
	if err != nil {
		return shim.Error("Failed to PutState while InsertData")
	}
	return shim.Success(nil)
}
func (t *MyHospital) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()
	if fn == "insert" {
		return insert(stub, args)
	} else if fn == "readRecord" {
		return readRecord(stub, args)
	} else if fn == "getHistoryRecords" { //get history of values for a record
		return getHistoryRecords(stub, args)
	} else {
		return shim.Error("Unknown func type while Invoke,please check")
	}
}

//获取历史记录
func getHistoryRecords(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//var b [][]byte
	historyresults, err := bean.GetHistoryRecords(stub, args)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(historyresults)
}

//插入记录
func insert(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	err := bean.InsertData(stub, args)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("记录数据成功"))
}

//查询一个记录
func readRecord(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//var my1 bean.MyLedgerno
	valAsbytes, err := bean.ReadRecord(stub, args)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(valAsbytes)
}

func main() {
	if err := shim.Start(new(MyHospital)); err != nil {
		fmt.Printf("Chaincode startup error: %s", err)
	}
}
