package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// FabricChaincode example simple Chaincode implementation
type FabricChaincode struct {
}

type diamond struct {
	ObjectType string `json:"docType"` 
	Name       string `json:"name"`    
	Origin      string `json:"origin"`
	Carats       int    `json:"carats"`
	Owner      string `json:"owner"`
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(FabricChaincode))
	if err != nil {
		fmt.Printf("Error starting a new instance of Diamond chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *FabricChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *FabricChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "createDiamond" { //create a new diamond
		return t.createDiamond(stub, args)
	} else if function == "transferDiamond" { //change owner of a specific diamond
		return t.transferDiamond(stub, args)
	} else if function == "queryDiamond" { //read a diamond
		return t.queryDiamond(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// createDiamond - create a new diamond, store into chaincode state
// ============================================================
func (t *FabricChaincode) createDiamond(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	//   0       1       2     3
	// "asdf", "blue", "35", "bob"
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init diamond")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	diamondName := args[0]
	origin := strings.ToLower(args[1])
	owner := strings.ToLower(args[3])
	carats, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	// ==== Check if diamond already exists ====
	diamondAsBytes, err := stub.GetState(diamondName)
	if err != nil {
		return shim.Error("Failed to get diamond: " + err.Error())
	} else if diamondAsBytes != nil {
		fmt.Println("This diamond already exists: " + diamondName)
		return shim.Error("This diamond already exists: " + diamondName)
	}

	// ==== Create diamond object and marshal to JSON ====
	objectType := "diamond"
	diamond := &diamond{objectType, diamondName, origin, carats, owner}
	diamondJSONasBytes, err := json.Marshal(diamond)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save diamond to state ===
	err = stub.PutState(diamondName, diamondJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the diamond to enable origin-based range queries, e.g. return all blue diamonds ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~origin~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~origin~*
	//indexName := "origin~name"
	//originNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{diamond.Origin, diamond.Name})
	//if err != nil {
	//	return shim.Error(err.Error())
	//}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the diamond.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	//value := []byte{0x00}
//	stub.PutState(originNameIndexKey, value)

	// ==== Diamond saved and indexed. Return success ====
	fmt.Println("- end init diamond")
	return shim.Success(nil)
}

// ===============================================
// queryDiamond - read a diamond from chaincode state
// ===============================================
func (t *FabricChaincode) queryDiamond(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the diamond to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name) //get the diamond from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Diamond does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}


// ===========================================================
// transfer a diamond by setting a new owner name on the diamond
// ===========================================================
func (t *FabricChaincode) transferDiamond(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0       1
	// "name", "bob"
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	diamondName := args[0]
	newOwner := strings.ToLower(args[1])
	fmt.Println("- start transferDiamond ", diamondName, newOwner)

	diamondAsBytes, err := stub.GetState(diamondName)
	if err != nil {
		return shim.Error("Failed to get diamond:" + err.Error())
	} else if diamondAsBytes == nil {
		return shim.Error("Diamond does not exist")
	}

	diamondToTransfer := diamond{}
	err = json.Unmarshal(diamondAsBytes, &diamondToTransfer) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	diamondToTransfer.Owner = newOwner //change the owner

	diamondJSONasBytes, _ := json.Marshal(diamondToTransfer)
	err = stub.PutState(diamondName, diamondJSONasBytes) //rewrite the diamond
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end transferDiamond (success)")
	return shim.Success(nil)
}

