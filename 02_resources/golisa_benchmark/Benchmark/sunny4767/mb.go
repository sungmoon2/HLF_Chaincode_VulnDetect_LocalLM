package main
import (
	"encoding/json"
	"fmt"
	"strings"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type SimpleChaincode struct {
}

type marble struct {
	ObjectType string `json:"docType"`
	Name string `json:"name"`
	Color string `json:"color"`
	Size int `json:"size"`
	Owner string `jason:"owner"`
}

func main() {
	err := shim.Start(new (SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting : %s", err)
	}
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response{
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response{

	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoking is running"+function)
	if function == "createMarble" {
		return t.createMarble(stub, args)
	} else if function == "transferMarble" {
		return t.transferMarble(stub, args)
	} else if function == "readMarble" {
		return t.readMarble(stub, args)
	}
	fmt.Println("invoke did not find func:" + function)
	return shim.Error("Received unknown function invocation")
}

func (t *SimpleChaincode) createMarble(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var err error
	if len(args) != 4 {
		return shim.Error("arg error, expecting 4!")
	}
	fmt.Println("- start init marble")
	if len(args[0]) <= 0 {
		return shim.Error("1st arg error") }
	if len(args[1]) <= 0 {
		return shim.Error("2nd arg error") }
	if len(args[2]) <= 0 {
		return shim.Error("3rd arg error") }
	if len(args[3]) <= 0 {
		return shim.Error("4th arg error") }
	marbleName := args[0]
	color := strings.ToLower(args[1])
	owner := strings.ToLower(args[3])
	size, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("3rd argument must be a numeric")
	}

	marbleAsBytes, err := stub.GetState(marbleName)
	if err != nil {
		return shim.Error("failed to get marble:"+err.Error())
	} else if marbleAsBytes != nil {
		fmt.Println("this marble already exist:"+marbleName)
		return shim.Error("this marble already exist:"+marbleName)
	}

	objectType := "marble"
	marble := &marble{objectType, marbleName, color, size, owner}
	marbleJSONasBytes, err := json.Marshal(marble)
	if err != nil {return shim.Error(err.Error())}

	err = stub.PutState(marbleName, marbleJSONasBytes)
	if err != nil {return shim.Error(err.Error())}
	fmt.Println("- end init marble")
	return shim.Success(nil)
}

func (t *SimpleChaincode) readMarble(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("incorrect number, marble to query")
	}
	name = args[0]
	valAsbytes, err := stub.GetState(name)
	if err != nil {
		jsonResp = "{\"Error\":\"faild to get state for}"+
			name + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valAsbytes)
}

func (t *SimpleChaincode) transferMarble(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("arg err, expecting 2")
	}
	marbleName := args[0]
	newOwner := strings.ToLower(args[1])
	fmt.Println("- start transferMarble", marbleName, newOwner)

	marbleAsBytes, err := stub.GetState(marbleName)
	if err != nil {
		return shim.Error("failed to get marble:"+err.Error())
	} else if marbleAsBytes == nil {
		return shim.Error("marble does not exist")
	}

	marbleToTransfer := marble{}
	err = json.Unmarshal(marbleAsBytes, &marbleToTransfer)
	if err != nil {
		return shim.Error(err.Error())
	}
	marbleToTransfer.Owner = newOwner
	marbleJSONasBytes, _ := json.Marshal(marbleToTransfer)
	err = stub.PutState(marbleName, marbleJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("- end transferMarble(success)")
	return shim.Success(nil)
}

