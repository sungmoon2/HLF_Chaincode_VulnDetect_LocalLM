package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/tree/release-1.1/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/tree/release-1.1/protos/peer"
)

type ItemChaincode struct {
}

type item struct {
	CompanyID string `json:"company_id"`
	SpecID    string `json:"spec_id"`
	How3      int    `json:"how3"`
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(ItemChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ========================================
// Init initializes chaincode
// ===========================
func (t *ItemChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// ========================================
// Invoke - Our entry point for Invocations
// ========================================
func (t *ItemChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "create" { //create a new item
		return t.create(stub, args)
	} else if function == "delete" {
		return t.delete(stub, args)
	} else if function == "update" {
		return t.update(stub, args)
	} else if function == "query" {
		return t.query(stub, args)
	} else if function == "getHistory" {
		return t.getHistory(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// create - create a new item, store into chaincode state
// ============================================================
func (t *ItemChaincode) create(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// ==== Input sanitation ====
	fmt.Println("- start create item")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	// companyID := args[0]
	// specID := args[1]
	// key := fmt.Sprintf("%s-%s", companyID, specID)

	// how3, err := strconv.Atoi(args[2])
	// if err != nil {
	// 	return shim.Error("argument how3 must be a numeric string")
	// }
	var i item
	itemJSONasBytes := []byte(args[0])
	if err := json.Unmarshal(itemJSONasBytes, &i); err != nil {
		msg := fmt.Sprintf("%s", args[0])
		return shim.Error(msg)
		// return shim.Error("Invalid json format")
	}
	key := fmt.Sprintf("%s-%s", i.CompanyID, i.SpecID)

	// ==== Check if item already exists ====
	itemAsBytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Failed to get item: " + err.Error())
	} else if itemAsBytes != nil {
		msg := fmt.Sprintf("The key %s has already existed!", key)
		return shim.Error(msg)
	}

	// ==== Create item object and marshal to JSON ====
	// item := &item{companyID, specID, how3}
	// itemJSONasBytes, err := json.Marshal(item)
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }

	// === Save item to state ===
	err = stub.PutState(key, itemJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Item saved and indexed. Return success ====
	fmt.Println("- end create item")
	return shim.Success(nil)
}

// ============================================================
// update - update a new item, store into chaincode state
// ============================================================
func (t *ItemChaincode) update(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// ==== Input sanitation ====
	fmt.Println("- start update item")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}

	companyID := args[0]
	specID := args[1]
	how3, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	key := fmt.Sprintf("%s-%s", companyID, specID)
	// ==== Check if item already exists ====
	itemAsBytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Failed to get item: " + err.Error())
	} else if itemAsBytes == nil {
		fmt.Println("This item NOT exists: " + key)
		return shim.Error("This item NOT exists: " + key)
	}

	// ==== Update item object and marshal to JSON ====
	item := &item{companyID, specID, how3}
	itemJSONasBytes, err := json.Marshal(item)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save item to state ===
	err = stub.PutState(key, itemJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Item saved and indexed. Return success ====
	fmt.Println("- end update item")
	return shim.Success(nil)
}

// ==================================================
// delete - remove a item from state
// ==================================================
func (t *ItemChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var itemJSON item

	fmt.Println("- start delete item")
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	companyID := args[0]
	specID := args[1]

	key := fmt.Sprintf("%s-%s", companyID, specID)
	itemAsbytes, err := stub.GetState(key) //get the item from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	} else if itemAsbytes == nil {
		jsonResp = "{\"Error\":\"Item does not exist: " + key + "\"}"
		return shim.Error(jsonResp)
	}

	err = json.Unmarshal([]byte(itemAsbytes), &itemJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + specID + "\"}"
		return shim.Error(jsonResp)
	}

	err = stub.DelState(key) //remove the item from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	fmt.Println("- end delete item")
	return shim.Success(nil)
}

// ==================================================
// query - query a item by ID
// ==================================================
func (t *ItemChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("- start query item")
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	companyID := args[0]
	specID := args[1]

	key := fmt.Sprintf("%s-%s", companyID, specID)
	// Get the state from the ledger
	itemAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	}

	if itemAsbytes == nil {
		jsonResp := "{\"Error\":\"Nil how3 for " + specID + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"How3\":\"" + string(itemAsbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)

	fmt.Println("- end query item")
	return shim.Success(itemAsbytes)
}

func (t *ItemChaincode) getHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("- start getHistory item")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	companyID := args[0]
	specID := args[1]

	key := fmt.Sprintf("%s-%s", companyID, specID)

	fmt.Printf("- start getHistory: %s\n", key)

	resultsIterator, err := stub.GetHistoryForKey(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the item
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON item)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForStore returning:\n%s\n", buffer.String())

	fmt.Println("- end getHistory item")
	return shim.Success(buffer.Bytes())
}
