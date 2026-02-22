package main

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/lcy1317/bran/chaincode/lib"
	"github.com/lcy1317/bran/chaincode/routers"
	"github.com/lcy1317/bran/chaincode/utils"
)

type BlockChainRealEstate struct {
}

// Init 链码初始化
func (t *BlockChainRealEstate) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("链码初始化")
	timeLocal, err := time.LoadLocation("Asia/Chongqing")
	//Todo
	//时区换一下
	if err != nil {
		return shim.Error(fmt.Sprintf("时区设置失败%s", err))
	}
	time.Local = timeLocal
	//初始化默认数据
	var userIds = [3]string{
		"00000001",
		"00000002",
		"00000003",
	}
	var userNames = [3]string{"jyh", "lyw", "gyw"}
	var balances = [3]float64{10000, 9.99, 10000}
	//初始化账号数据
	for i, val := range userIds {
		account := &lib.Account{
			UserId: val,
			UserName:  userNames[i],
			Balance:   balances[i],
		}
		// 写入账本
		if err := utils.WriteLedger(account, stub, lib.AccountKey, []string{val}); err != nil {
			return shim.Error(fmt.Sprintf("%s", err))
		}
	}
	return shim.Success(nil)
}

// Invoke 实现Invoke接口调用智能合约
func (t *BlockChainRealEstate) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case "createAccount":
		return routers.CreateAccount(stub, args)
	case "queryAccountList":
		return routers.QueryAccountList(stub, args)
	case "delAccountList":
		return routers.DelAccount(stub, args)
	case "createSelling":
		return routers.CreateSelling(stub, args)
	case "refreshSelling":
		return routers.RefreshSelling(stub, args)
	case "querySellingList":
		return routers.QuerySellingList(stub, args)
	case "delSellingList":
		return routers.DelSelling(stub, args)
	case "createBuying":
		return routers.CreateBuying(stub, args)
	case "queryServiceList":
		return routers.QueryServiceList(stub, args)
	default:
		return shim.Error(fmt.Sprintf("没有该功能: %s", funcName))
	}
}

func main() {
	err := shim.Start(new(BlockChainRealEstate))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
