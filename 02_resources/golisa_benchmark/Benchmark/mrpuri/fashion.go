package main

import (
	"encoding/json"
	controller "fashion/controller"
	"fashion/model"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type FashionChaincode struct {
}

//FashionChaincode represents the chaincode object referenced throughout, reciever for chaincode shim functions

//Init funnction runs at the chaincode initialisation
func (fashion *FashionChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("init executed")
	//logger.Debug("Init executed for log")

	return shim.Success(nil)
}

//Invoke function runs on query and invoke
func (fashion *FashionChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	fmt.Println("invoke executed")
	stub.PutState("token", []byte("2000"))

	//instance of the response structure which is similar to http response
	//return peer.Response{Status:401, Message: "Unauthorized", Payload: payload}
	//sending success and error response.
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "initcloth" { //create a new clothing asset

		return fashion.Initcloth(stub, args)
	} else if function == "get" {
		return controller.Get(stub)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")

}

// ============================================================
// Initcloth - create a new clothing asset, stores into chaincode state
// ============================================================
func (fashion *FashionChaincode) Initcloth(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error

	//   0       1       2     3
	// "shirt", "blue", "40", "bob"



	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init fashion")
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
	clothName := args[0]
	color := strings.ToLower(args[1])
	owner := strings.ToLower(args[3])
	size, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	// ==== Check if asset already exists ====
	clothAsBytes, err := stub.GetState(clothName)
	if err != nil {
		return shim.Error("Failed to get cloths: " + err.Error())
	} else if clothAsBytes != nil {
		fmt.Println("This cloth already exists: " + clothName)
		return shim.Error("This clothing already exists: " + clothName)
	}

	// ==== Create cloth object and marshal to JSON ====
	ObjectType := "clothing"
	cloth := &model.Cloth{ObjectType, clothName, color, size, owner}
	clothJSONasBytes, err := json.Marshal(cloth)
	if err != nil {
		return shim.Error(err.Error())
	}
	//Alternatively, build the cloth json string manually if you don't want to use struct marshalling
	//fashionJSONasString := `{"docType":"jean",  "name": "` + clothName + `", "color": "` + color + `", "size": ` + strconv.Itoa(size) + `, "owner": "` + owner + `"}`
	//fashionJSONasBytes := []byte(str)

	// === Save asset to state ===
	err = stub.PutState(clothName, clothJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the cloth to enable color-based range queries, e.g. return all blue marbles ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~color~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	indexName := "color~name"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{cloth.Color, cloth.Name})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the marble.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(colorNameIndexKey, value)

	// ==== Marble saved and indexed. Return success ====
	fmt.Println("- end init cloth, the data added to the chaincode state is: ")
	fmt.Println(cloth)
	return shim.Success(nil)
}

func main() {
	fmt.Println("Started Chaincode")
	err := shim.Start(new(FashionChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode : %v", err)
	}
}
