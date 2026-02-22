package finance

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"

	"Fabric-chaincode/finance/bean"
	"Fabric-chaincode/finance/utils"
	"fmt"
)

type Finance struct {

}

//初始化链码
func (t *Finance)Init(stub shim.ChaincodeStubInterface) peer.Response {
	args :=stub.GetStringArgs()
	if len(args)!=0{
		return shim.Error("初始化时参数错误")
	}
	return shim.Success(nil)

}

//调用链码
func (t *Finance) Invoke(stub shim.ChaincodeStubInterface) peer.Response  {
	fn,args:=stub.GetFunctionAndParameters()
	switch fn {
	case "loan"://记录贷款数据
	       return  loan(stub,args)
	default:
		return shim.Error("调用的函数未知，请检查")
	}
}

//记录贷款数据，返回相关msg
func loan(stub shim.ChaincodeStubInterface,args[]string) peer.Response  {
	name,err:=utils.GetCreator(stub)
	if err!=nil{
		return shim.Error(err.Error())
	}
	err=bean.Loan(stub,args,name)
	if err!=nil{
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("记录贷款数据成功"))
}

//主函数入口，启动链码
func main()  {
	if err:=shim.Start(new(Finance));err!=nil{
		fmt.Printf("链码启动出现错误：%s",err)
	}
}
