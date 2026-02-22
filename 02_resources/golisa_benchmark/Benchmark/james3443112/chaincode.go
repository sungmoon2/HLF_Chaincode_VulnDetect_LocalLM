package ygDigitalAssets

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type YGDAChaincode struct {
}

type digitalassets struct {
	AssetsName  string `json:name`
	AssetsType  string `json:type`
	AssetsValue string `json:value`
	AssetsDesc  string `json:desc`
	Owner       string `json:owner`
}

func (t *YGDAChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *YGDAChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("YGDA Invoke")
	function, args := stub.GetFunctionAndParameters()

	if function == "initAssets" {
		return t.initAssets(stub, args)
	} else if function == "queryAssets" {
		return t.queryAssets(stub, args)
	} else if function == "transferAssets" {
		return t.transferAssets(stub, args)
	} else if function == "modifyAssets" {
		return t.modifyAssets(stub, args)
	} else if function == "deleteAssets" {
		return t.deleteAssets(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"query\" \"addAssets\" \"transferAssets\" \"modifyAssets\" \"deleteAssets\" ")
}

func (t *YGDAChaincode) queryAssets(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	return shim.Success(nil)
}

func (t *YGDAChaincode) initAssets(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var UserName string
	var Type string
	var Value string
	var Desc string
	var Name string
	var err error

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	UserName = args[0]
	Name = args[1]
	Type = args[2]
	Value = args[3]
	Desc = args[4]

	key := tomd5(UserName + Name + Type)
	data, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Failed to get da:" + err.Error())
	} else if data != nil {
		return shim.Error("This dc already exists:" + UserName + Name + Type)
	}

	assetsToTransfer := digitalassets{}
	assetsToTransfer.Owner = UserName
	assetsToTransfer.AssetsType = Type
	assetsToTransfer.AssetsName = Name
	assetsToTransfer.AssetsValue = Value
	assetsToTransfer.AssetsDesc = Desc

	assetsJSONasBytes, _ := json.Marshal(assetsToTransfer)
	err = stub.PutState(key, assetsJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *YGDAChaincode) transferAssets(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var OldUserName string
	var OldType string
	var OldName string
	var NewUserName string
	var assetsJson digitalassets
	var jsonResp string
	var err error

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	OldUserName = args[0]
	OldName = args[1]
	OldType = args[2]
	NewUserName = args[3]

	key := tomd5(OldUserName + OldName + OldType)
	oldAssetsBytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("data repeat")
	} else if oldAssetsBytes == nil {
		return shim.Error("Failed to get assets")
	}

	err = json.Unmarshal([]byte(oldAssetsBytes), &assetsJson)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + OldUserName + OldType + OldName + "\"}"
		return shim.Error(jsonResp)
	}

	assetsJSONasBytes, _ := json.Marshal(assetsJson)
	err = stub.PutState(key, assetsJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	assetsJson.Owner = NewUserName
	newkey := tomd5(NewUserName + OldName + OldType)

	assetsJSONasBytes, _ = json.Marshal(assetsJson)
	err = stub.PutState(newkey, assetsJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.DelState(key)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	return shim.Success(nil)
}

func (t *YGDAChaincode) modifyAssets(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var UserName string
	var Type string
	var NewValue string
	var NewDesc string
	var Name string
	var assetsJson digitalassets
	var jsonResp string
	var err error

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	UserName = args[0]
	Name = args[1]
	Type = args[2]
	NewValue = args[3]
	NewDesc = args[4]

	key := tomd5(UserName + Name + Type)
	olddata, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Failed to get da:" + err.Error())
	} else if olddata != nil {
		return shim.Error("This dc already exists:" + UserName + Name + Type)
	}

	err = json.Unmarshal([]byte(olddata), &assetsJson)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + UserName + Type + Name + "\"}"
		return shim.Error(jsonResp)
	}

	assetsJson.AssetsValue = NewValue
	assetsJson.AssetsDesc = NewDesc

	assetsJSONasBytes, _ := json.Marshal(assetsJson)
	err = stub.PutState(key, assetsJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *YGDAChaincode) deleteAssets(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var UserName string
	var Type string
	var Name string
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	key := tomd5(UserName + Name + Type)
	assetsBytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("data repeat")
	} else if assetsBytes == nil {
		return shim.Error("Failed to get assets")
	}

	err = stub.DelState(key)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	return shim.Success(nil)
}

func tomd5(data string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(data))
	res := hex.EncodeToString(md5Ctx.Sum(nil))
	return res
}

func main() {
	err := shim.Start(new(YGDAChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
