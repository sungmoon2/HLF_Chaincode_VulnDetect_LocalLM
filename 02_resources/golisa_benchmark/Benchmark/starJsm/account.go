package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type AccountChaincode struct {
	founder string
}

// 链码实例化时，调用Init函数初始化数据
// 链码升级时，也会调用此函数重置或迁移数据
func (t *AccountChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Printf("开始实例化链码")

	// 获取参数
	_, args := stub.GetFunctionAndParameters()
	var A string
	var Aval int
	var err error

	if len(args) != 2 {
		return shim.Error("参数个数错误")
	}
	// '{"Args":["init", "a", "100"]}'

	fmt.Println("初始化链码》》》")

	A = args[0]
	t.founder = A
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("账户金额应为整数")
	}
	fmt.Printf("Aval = %d\n", Aval)

	// 通过PutState方法将数据保存在账本中
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error("保存数据发生错误")
	}

	fmt.Printf("实例化链码成功")

	return shim.Success(nil)
}

// 对账本数据进行操作
func (t *AccountChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ex02 Invoke")
	fun, args := stub.GetFunctionAndParameters()

	if fun == "create" {
		return t.create(stub, args)
	} else if fun == "query" {
		return t.query(stub, args)
	} else if fun == "transfer" {
		return t.transfer(stub, args)
	} else if fun == "unsubscribe" {
		return t.unsubscribe(stub, args)
	}
	return shim.Error("命令错误")
}

func (t *AccountChaincode) create(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("ex02 creat")

	if len(args) != 2 {
		return shim.Error("参数数量应为2")
	}
	// "{"Args":["creat","b","100"]}"
	var A string
	var Aval int
	var err error

	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("账户金额应为整数")
	}

	result, err := stub.GetState(A)
	if err != nil {
		return shim.Error("系统内部错误")
	}
	if result != nil {
		return shim.Error("账户已存在")
	}

	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error("err.Error")
	}

	return shim.Success(nil)
}

func (t *AccountChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string
	var err error

	if len(args) != 1 {
		return shim.Error("参数应为1")
	}
	A = args[0]

	result, err := stub.GetState(A)
	if err != nil {
		return shim.Error("没有查询到" + A)
	}
	if result == nil {
		return shim.Error("账户" + A + "余额为空")
	}

	return shim.Success(result)
}

func (t *AccountChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("transfer>>>")
	if len(args) != 3 {
		return shim.Error("转账操作参数应为3")
	}

	var A, B string
	var trans int
	var Aval, Bval int
	var err error

	A = args[0]
	B = args[1]
	trans, err = strconv.Atoi(args[2])
	if err != nil {
		shim.Error("转账金额应为整数")
	}

	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("获取账户" + A + "失败")
	}
	if Avalbytes == nil {
		shim.Error("账户" + A + "余额为空")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("获取账户" + B + "失败")
	}
	if Bvalbytes == nil {
		shim.Error("账户" + B + "余额为空")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Aval = Aval - trans
	// Bval = Bval + trans
	Aval -= trans
	Bval += trans
	if Aval < 0 {
		return shim.Error("转账金额大于账户" + A + "的余额")
	}
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *AccountChaincode) unsubscribe(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		jsonResp := "{\"Error\":\"参数个数应为1\"}"
		fmt.Println(jsonResp)
		return shim.Error(jsonResp)
	}
	if t.founder == args[0] {
		jsonResp := "{\"Error\":\"无法删除初始账户\"}"
		fmt.Println(jsonResp)
		return shim.Error(jsonResp)
	}

	var Aval, Bval int
	var err error
	A := args[0]
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		fmt.Println("Failed to get state")
		shim.Error("Failed to get state")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbyets, _ := stub.GetState(t.founder)
	Bval, _ = strconv.Atoi(string(Bvalbyets))
	Bval += Aval
	err = stub.PutState(t.founder, []byte(strconv.Itoa(Bval)))
	if err != nil {
		fmt.Println("写入初始账户失败")
		shim.Error(err.Error())
	}

	err = stub.DelState(A)
	if err != nil {
		fmt.Println("无法删除账户" + A)
		shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(AccountChaincode))
	if err != nil {
		fmt.Printf("链码启动失败：%s", err)
	}
}
