package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)


//示例 codechain 用于管理 资产
type SimpleAsset struct{

}

// 调用 codechain 实例 用于初始化 任何 数据
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response{

	//从交易中获得 信息(proposal) 期望获得一个key-value数据
	args := stub.GetStringArgs()
	if len(args) != 2 {
		return shim.Error("Incorrect arguments. Expecting a key and a value")
	}

	//设置 变量 或者 资产 ，通过调用 stub.PutState()

	//储存 key 和 value 到 账本上
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil{
		return shim.Error(fmt.Sprintf("创建资产失败 %s\n", args[0]))
	}
	return shim.Success(nil)
}

// Invoke 是在 chaincode上调用交易
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response{

	// 提取 从交易信息中(proposal) 函数 和 参数
	fn, args := stub.GetFunctionAndParameters()

	fmt.Println("args", args)
	var result string
	var err error

	if fn == "set"{
		result, err = set(stub, args)
	}else{
		result, err = get(stub, args)
	}

	if err != nil{
		return shim.Error(err.Error())
	}

	//返回 成功的 payload
	return shim.Success([]byte(result))
}

/*
实现chaincode application get/set
**/

//set 会储存asset （包括key 和 value） 在账本上， 如果key存在，则会用新的值覆盖(override) value
func set(stub shim.ChaincodeStubInterface, args []string) (string, error){
	fmt.Println("args", args)
	if len(args) != 2{
		return "", fmt.Errorf("参数出错， 需要一个 kye和value\n")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil{
		return "", fmt.Errorf("保存资产(asset)失败: %s\n", args[0])
	}

	return args[1], nil
}

//get 返回 指定 资产(asset) key 的 value
func get(stub shim.ChaincodeStubInterface, args []string) (string, error){
	fmt.Println("args", args)
	if len(args) != 1{
		return "", fmt.Errorf("参数不正确，需要一个 key")
	}

	value, err := stub.GetState(args[0])
	if err != nil{
		return "", fmt.Errorf("获取资产：%s失败，错误：%s\n", args[0], err)
	}

	if value == nil{
		return "", fmt.Errorf("资产未找到：%s\n", args[0])
	}

	return string(value),nil
}

func main(){
	if err := shim.Start(new(SimpleAsset)); err != nil{
		fmt.Printf("chaincode例程启动失败: %s\n", err)
	}
}
