package Identity

import (
	"Chaincode/Identity/idemixplus"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type CredChaincode struct {
}

type CredRecord struct {
	NymCred []byte `json:"nymcred"`
	Content string `json:"content"`
}

func (e *CredChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (e *CredChaincode) ipkinit(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	err := stub.PutState("IssuerPublicKey", []byte(args[0]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset: IssuerPublicKey"))
	}
	return shim.Success(nil)
}

// Invoke 执行合约时调用
func (e *CredChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	funcs, args := stub.GetFunctionAndParameters()

	if funcs == "idemix" {
		return e.idemix(stub, args)
	} else if funcs == "queryIdemix" {
		return e.queryIdemix(stub, args)
	} else if funcs == "ipkinit" {
		return e.ipkinit(stub, args)
	}

	return shim.Error("Unknow Functions: " + funcs)
}

// idemix 将信息写入账本
func (e *CredChaincode) idemix(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	// argv[0]: 用户的匿名证书
	// argv[1]: 要写入账本的信息

	// 解析匿名证书
	nym := &idemixplus.NymSignature{}
	err := proto.Unmarshal([]byte(args[0]), nym)
	if err != nil {
		return shim.Error(err.Error())
	}

	// 从账本中获取CA的idemix公钥
	ipkBytes, err := stub.GetState("IssuerPublicKey")
	if err != nil {
		return shim.Error("Failed to get state for  IssuerPublicKey")
	}
	ipk := &idemixplus.IssuerPublicKey{}
	err = proto.Unmarshal(ipkBytes, ipk)
	if err != nil {
		shim.Error(err.Error())
	}

	// 校验匿名证书：通过，将信息写入账本；否则返回错误信息
	err = nym.Ver(ipk, []byte(args[1]), nil, 0)
	if err != nil {
		return shim.Error(err.Error())
	}

	// 获取交易ID，以此为Key将信息写入账本
	key := stub.GetTxID()
	// 构造消息结构并在序列化之后存入账本
	record := CredRecord{
		NymCred: []byte(args[0]),
		Content: args[1],
	}
	recordBytes, err := json.Marshal(record)
	if err != nil {
		shim.Error(err.Error())
	}
	err = stub.PutState(key, recordBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(key))
}

// queryIdexmix 通过交易ID来检索匿名证书
func (e *CredChaincode) queryIdemix(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect arguments. Expecting a key and a value")
	}

	// 通过交易ID从账本中获取消息，反序列化后返回匿名证书
	recordBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}

	if recordBytes == nil {
		return shim.Error("No such key")
	}

	return shim.Success(recordBytes)
}

func main() {
	// Create a new Smart Contract
	err := shim.Start(new(CredChaincode))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
