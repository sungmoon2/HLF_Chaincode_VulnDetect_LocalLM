package chaincode

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

/*
	每一个链码都要实现 Chaincode 接口 中的 Init 和 Invoke 方法
	导入链码 shim 包和 peer protobuf 包
*/

// SimpleAsset 加入一个结构体来作为 Chaincode shim 方法的接收者。
type SimpleAsset struct {
}

// Init 初始化链码
// 链码升级的时候也要调用这个方法。当写一个用来升级已存在的链码的时候，请确保合理更改 Init 方法。
// 特别地，当升级时没有“迁移”或者没东西需要初始化时，可以提供一个空的 Init 方法。
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Get the args from the transaction proposal
	args := stub.GetStringArgs()
	if len(args) != 2 {
		return shim.Error("Incorrect arguments. Expecting a key and a value")
	}

	// Set up any variables or assets here by calling stub.PutState()
	// We store the key and the value on the ledger
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
	}
	return shim.Success(nil)
}

// Invoke 调用链码
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// 为链码应用程序的方法解析方法名和参数。
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error

	fmt.Printf("%s %s \n", fn, args)

	if fn == "set" {
		result, err = set(stub, args)
	} else { // assume 'get' even if fn is nil
		result, err = get(stub, args)
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success([]byte(result))
}

// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
func set(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	return args[1], nil
}

// Get returns the value of the specified asset key
func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	return string(value), nil
}
