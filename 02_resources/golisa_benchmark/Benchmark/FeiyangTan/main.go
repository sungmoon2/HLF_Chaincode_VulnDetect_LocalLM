//该链码基于hyperledger fabric 1.4 开发
//实现资产资产在用户之间的共享
package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

//AssertsExchangeCC asserts exchange chain code
type AssertsExchangeCC struct{}

//User 用户
type User struct {
	Name   string   `json:"name"`
	ID     string   `json:"id"`
	Assets []string `json:"assets"`
}

//Asset 资产
type Asset struct {
	Name          string `json:"name"`
	ID            string `json:"id"`
	Infor         string `json:"infor"`
	OwnerID       string `json:"owner_id"`
	OriginOwnerID string `json:"origin_owner_id"`
}

//AssetExchangeRecord 资产转移记录
type AssetExchangeRecord struct {
	//Id 资产id
	ID         string `json:"id"`
	OwnerID    string `json:"owner_id"`
	NewOwnerID string `json:"new_owner_id"`
}

//开始
func main() {
	err := shim.Start(new(AssertsExchangeCC))
	if err != nil {
		fmt.Printf("Starting AssertsExchange chaincode container error: %s\n", err)
	}
}

//Init 创建Chaincode container
func (a *AssertsExchangeCC) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success([]byte("Init success"))
}

//Invoke 操作/查询Chaincode
func (a *AssertsExchangeCC) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	funcName, args := stub.GetFunctionAndParameters()

	switch funcName {
	case "userRegister":
		return userRegister(stub, args)
	case "userDelete":
		return userDelete(stub, args)
	case "assetRegister":
		return userRegister(stub, args)
	case "assetDelete":
		return assetDelete(stub, args)
	case "assetExchange":
		return assetExchange(stub, args)
	case "queryUser":
		return queryUser(stub, args)
	case "queryAsset":
		return queryAsset(stub, args)
	case "queryAssetExchangeRecord":
		return queryAssetExchangeRecord(stub, args)
	default:
		return shim.Error("unexpect function name: " + funcName)
	}
}

//添加用户
func userRegister(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//1.1:检查参数个数
	if len(args) != 2 {
		return shim.Error("Incorrect inputs amount")
	}
	//1.2:检查参数合法性
	name := args[0]
	id := args[1]
	if name == "" || id == "" {
		return shim.Error("Inputs information is nil")
	}
	//1.3:检查参数是否已存在
	userBytes, err := stub.GetState("user_" + id)
	if len(userBytes) != 0 && err == nil {
		return shim.Error("User alrady exist")
	} else if err != nil {
		return shim.Error(fmt.Sprintf("Get state error: %s", err))
	}
	//2:添加用户
	user := User{
		Name:   name,
		ID:     id,
		Assets: make([]string, 0),
	}
	userBytes, err = json.Marshal(user)
	if err != nil {
		return shim.Error(fmt.Sprintf("Json marshal user error: %s", err))
	}
	err = stub.PutState("user_"+id, userBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Put user error: %s", err))
	}
	//3.操作完毕
	return shim.Success(nil)
}

//注销用户
func userDelete(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//1.1:检查参数个数
	if len(args) != 1 {
		return shim.Error("Incorrect inputs amount")
	}
	//1.2:检查参数合法性
	id := args[0]
	if id == "" {
		return shim.Error("Inputs information is nil")
	}
	//1.3:检查参数是否以存在
	userBytes, err := stub.GetState("user_" + id)
	if len(userBytes) == 0 || err != nil {
		return shim.Error("User doesn't exist or get state error")
	}
	//2.1:用户注销
	err = stub.DelState("user_" + id)
	if err != nil {
		return shim.Error(fmt.Sprintf("Delete user error: %s", err))
	}
	//2.2:用户所拥有的资产注销
	userBytes, err = stub.GetState("user_" + id)
	if err != nil {
		return shim.Error(fmt.Sprintf("Get state user id error: %s", err))
	}
	user := new(User)
	err = json.Unmarshal(userBytes, user)
	if err != nil {
		return shim.Error(fmt.Sprintf("Json unmarshall userBytes error: %s", err))
	}
	for _, assetID := range user.Assets {
		err = stub.DelState("asset_" + assetID)
		if err != nil {
			return shim.Error(fmt.Sprintf("Delete asset state error: %s", err))
		}
	}
	//3.操作完毕
	return shim.Success(nil)
}

//添加资产
func assetRegister(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//1.1:检查参数个数
	if len(args) != 4 {
		return shim.Error("Incorrect inputs amount")
	}
	//1.2:检查参数合法性
	name := args[0]
	id := args[1]
	infor := args[2]
	ownerID := args[3]
	if name == "" || id == "" || infor == "" {
		return shim.Error("Inputs information is nil")
	}
	//1.3:检查资产是否已存在
	assetBytes, err := stub.GetState("asset_" + id)
	if err != nil {
		return shim.Error(fmt.Sprintf("Get state error: %s", err))
	}
	if len(assetBytes) != 0 && err == nil {
		return shim.Error("Asset alrady exist")
	}
	//:1.4：检查用户是否已存在
	userBytes, err := stub.GetState("user_" + ownerID)
	if err != nil {
		return shim.Error(fmt.Sprintf("Get state error: %s", err))
	}
	if len(userBytes) != 0 && err == nil {
		return shim.Error("User alrady exist")
	}
	//2.1:添加资产
	asset := Asset{
		Name:          name,
		ID:            id,
		Infor:         infor,
		OwnerID:       ownerID,
		OriginOwnerID: ownerID,
	}
	assetBytes, err = json.Marshal(asset)
	if err != nil {
		return shim.Error(fmt.Sprintf("Json marshal asset error: %s", err))
	}
	err = stub.PutState("asset_"+id, assetBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Put asset error: %s", err))
	}
	//2.2:更新资产拥有者资产列表
	userBytes, err = stub.GetState("user_" + ownerID)
	if err != nil {
		return shim.Error(fmt.Sprintf("Get state user id error: %s", err))
	}
	user := new(User)
	err = json.Unmarshal(userBytes, user)
	if err != nil {
		return shim.Error(fmt.Sprintf("Json unmarshall userBytes error: %s", err))
	}
	user.Assets = append(user.Assets, id)
	userBytes, err = json.Marshal(user)
	if err != nil {
		return shim.Error(fmt.Sprintf("Json marshal user error: %s", err))
	}
	err = stub.PutState("user_"+id, userBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Put user error: %s", err))
	}
	//2.3:添加资产转移历史
	assetExchangeRecord := AssetExchangeRecord{
		ID:         id,
		OwnerID:    "none",
		NewOwnerID: ownerID,
	}
	assetExchangeRecordBytes, err := json.Marshal(assetExchangeRecord)
	if err != nil {
		return shim.Error(fmt.Sprintf("Json marshal assetExchangeRecord error: %s", err))
	}
	key, err := stub.CreateCompositeKey("history", []string{id, "none", ownerID})
	if err != nil {
		return shim.Error(fmt.Sprintf("Create composite key error: %s", err))
	}
	err = stub.PutState(key, assetExchangeRecordBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Put assetExchangeRecordBytes error: %s", err))
	}
	//3.操作完毕
	return shim.Success(nil)
}

//资产注销
func assetDelete(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//1.1:检查参数个数
	if len(args) != 1 {
		return shim.Error("Incorrect inputs amount")
	}
	//1.2:检查参数合法性
	id := args[0]
	if id == "" {
		return shim.Error("Inputs information is nil")
	}
	//1.3:检查参数是否存在
	a, err := stub.GetState("asset_" + id)
	if err != nil {
		return shim.Error(fmt.Sprintf("Get state error: %s", err))
	}
	if len(a) == 0 || err != nil {
		return shim.Error("Asset doesn't exist")
	}
	//2.1:更新资产拥有者资产列表
	assetBytes, err := stub.GetState("asset_" + id)
	if err != nil {
		return shim.Error(fmt.Sprintf("Get state asset id error: %s", err))
	}
	asset := new(Asset)
	err = json.Unmarshal(assetBytes, asset)
	if err != nil {
		return shim.Error(fmt.Sprintf("Json unmarshal assetBytes error: %s", err))
	}
	userid := asset.OwnerID
	userBytes, err := stub.GetState("user_" + userid)
	if err != nil {
		return shim.Error(fmt.Sprintf("Get state user id error: %s", err))
	}
	user := new(User)
	err = json.Unmarshal(userBytes, user)
	if err != nil {
		return shim.Error(fmt.Sprintf("Json unmarshall userBytes error: %s", err))
	}
	var i int
	for j, userAssetID := range user.Assets {
		if userAssetID == id {
			i = j
			break
		}
		return shim.Error("Can't find assetID from user error")
	}
	user.Assets = append(user.Assets[:i], user.Assets[i+1:]...)
	userBytes, err = json.Marshal(user)
	if err != nil {
		return shim.Error(fmt.Sprintf("Json marshal user error: %s", err))
	}
	err = stub.PutState("user_"+id, userBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Put user error: %s", err))
	}
	//2.2:资产注销
	err = stub.DelState("asset_" + id)
	if err != nil {
		return shim.Error(fmt.Sprintf("Delete asset error: %s", err))
	}
	//2.3:添加资产转移历史
	assetExchangeRecord := AssetExchangeRecord{
		ID:         id,
		OwnerID:    userid,
		NewOwnerID: "none",
	}
	assetExchangeRecordBytes, err := json.Marshal(assetExchangeRecord)
	if err != nil {
		return shim.Error(fmt.Sprintf("Json marshal assetExchangeRecord error: %s", err))
	}
	key, err := stub.CreateCompositeKey("history", []string{id, userid, "none"})
	if err != nil {
		return shim.Error(fmt.Sprintf("Create composite key error: %s", err))
	}
	err = stub.PutState(key, assetExchangeRecordBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Put assetExchangeRecordBytes error: %s", err))
	}
	//3.操作完毕
	return shim.Success(nil)
}

//转移资产
func assetExchange(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//1.1:检查参数个数
	if len(args) != 3 {
		return shim.Error("Incorrect inputs amount")
	}
	//1.2:检查参数合法性
	id := args[0]
	ownerID := args[1]
	newOwnerID := args[2]
	if id == "" || ownerID == "" || newOwnerID == "" {
		return shim.Error("Inputs information is nil")
	}
	//1.3:检查参数是否存在
	a, err := stub.GetState("asset_" + id)
	if err != nil {
		return shim.Error(fmt.Sprintf("Get state error: %s", err))
	}
	if len(a) == 0 || err != nil {
		return shim.Error("Asset doesn't exist")
	}
	a, err = stub.GetState("user_" + ownerID)
	if err != nil {
		return shim.Error(fmt.Sprintf("Get state error: %s", err))
	}
	if len(a) == 0 || err != nil {
		return shim.Error("User doesn't exist")
	}
	a, err = stub.GetState("user_" + newOwnerID)
	if err != nil {
		return shim.Error(fmt.Sprintf("Get state error: %s", err))
	}
	if len(a) == 0 || err != nil {
		return shim.Error("User doesn't exist")
	}
	//2:添加资产转移历史
	assetExchangeRecord := AssetExchangeRecord{
		ID:         id,
		OwnerID:    ownerID,
		NewOwnerID: newOwnerID,
	}
	assetExchangeRecordBytes, err := json.Marshal(assetExchangeRecord)
	if err != nil {
		return shim.Error(fmt.Sprintf("Json marshal assetExchangeRecord error: %s", err))
	}
	key, err := stub.CreateCompositeKey("history", []string{id, ownerID, newOwnerID})
	if err != nil {
		return shim.Error(fmt.Sprintf("Create composite key error: %s", err))
	}
	err = stub.PutState(key, assetExchangeRecordBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Put assetExchangeRecordBytes error: %s", err))
	}
	//3.操作完毕
	return shim.Success(nil)
}

//查询用户
func queryUser(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//1.1:检查参数个数
	if len(args) != 1 {
		return shim.Error("Incorrect inputs amount")
	}
	//1.2:检查参数合法性
	id := args[0]
	if id == "" {
		return shim.Error("Inputs information is nil")
	}
	//1.3:检查参数是否存在
	userBytes, err := stub.GetState("user_" + id)
	if err != nil {
		return shim.Error(fmt.Sprintf("Get state error: %s", err))
	}
	if len(userBytes) == 0 || err != nil {
		return shim.Error("User doesn't exist")
	}
	//2:返回用户
	return shim.Success(userBytes)
}

//查询资产
func queryAsset(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//1.1:检查参数个数
	if len(args) != 1 {
		return shim.Error("Incorrect inputs amount")
	}
	//1.2:检查参数合法性
	id := args[0]
	if id == "" {
		return shim.Error("Inputs information is nil")
	}
	//1.3:检查参数是否存在
	assetBytes, err := stub.GetState("asset_" + id)
	if err != nil {
		return shim.Error(fmt.Sprintf("Get state error: %s", err))
	}
	if len(assetBytes) == 0 || err != nil {
		return shim.Error("Asset doesn't exist")
	}
	//2:返回资产
	return shim.Success(assetBytes)
}

//查询资产记录
func queryAssetExchangeRecord(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//1.1:检查参数个数
	if len(args) != 1 {
		return shim.Error("Incorrect inputs amount")
	}
	//1.2:检查参数合法性
	id := args[0]
	if id == "" {
		return shim.Error("Inputs information is nil")
	}
	//1.3:检查参数是否存在
	assetBytes, err := stub.GetState("asset_" + id)
	if err != nil {
		return shim.Error(fmt.Sprintf("Get state error: %s", err))
	}
	if len(assetBytes) == 0 || err != nil {
		return shim.Error("Asset doesn't exist")
	}
	//2.1:寻找资产转移记录
	keys := make([]string, 0)
	keys = append(keys, id)
	result, err := stub.GetStateByPartialCompositeKey("history", keys)
	if err != nil {
		return shim.Error(fmt.Sprintf("query history error: %s", err))
	}
	defer result.Close()
	histories := make([]*AssetExchangeRecord, 0)
	for result.HasNext() {
		historyVal, err := result.Next()
		if err != nil {
			return shim.Error(fmt.Sprintf("query error: %s", err))
		}
		history := new(AssetExchangeRecord)
		err = json.Unmarshal(historyVal.GetValue(), history)
		if err != nil {
			return shim.Error(fmt.Sprintf("unmarshal error: %s", err))
		}
		histories = append(histories, history)
	}
	historiesBytes, err := json.Marshal(histories)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal error: %s", err))
	}
	//2.2:返回资产转移记录
	return shim.Success(historiesBytes)
}
