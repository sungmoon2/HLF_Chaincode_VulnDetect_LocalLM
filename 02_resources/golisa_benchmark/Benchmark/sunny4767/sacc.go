package main
import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type SimpleAsset struct {
}

func (s *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	args := stub.GetStringArgs()
	if len(args) != 2 {
		return shim.Error("arg errori")
	}
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(fmt.Sprintf("error", args[0]))
	}
	return shim.Success(nil)
}

func (s *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error
	if fn == "set" {
		result, err = set(stub, args)
	} else {
		result, err = get(stub, args)
	}
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(result))
}

func set(stub shim.ChaincodeStubInterface, args []string) (string,error) {
	if len(args) != 2 {
		return "", fmt.Errorf("error arg cnt set()")
	}
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Putstate error", args[0]);
	}
	return args[1], nil
}

func get(stub shim.ChaincodeStubInterface, args []string) (string,error) {
	if len(args) != 1 {
		return "", fmt.Errorf("error arg cnt get()");
	}
	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("error", args[0])
	}
	if value == nil {
		return "", fmt.Errorf("", args[0])
	}
	return string(value), nil
}

func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("error", err)
	}
}



