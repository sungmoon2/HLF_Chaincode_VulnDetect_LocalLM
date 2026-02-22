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

type PurchaseChaincode struct {
}

type subPurchase struct {
	SpecID int     `json:"spec_id"`
	How    int     `json:"how"`
	Money  float64 `json:"money"`
}

type purchase struct {
	CompanyID *string       `json:"company_id"`
	OrderID   *int          `json:"order_id"`
	TabNo     string        `json:"tabno"`
	Client    string        `json:"client"`
	AccTime   int64         `json:"acc_time"`
	Items     []subPurchase `json:"items"`
}

// type index struct {
// 	CompanyID string `json:"company_id"`
// 	OrderID   int    `json:"order_id"`
// }

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(PurchaseChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ========================================
// Init initializes chaincode
// ===========================
func (t *PurchaseChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// ========================================
// Invoke - Our entry point for Invocations
// ========================================
func (t *PurchaseChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "create" { //create a new item
		return t.create(stub, args)
	} else if function == "query" {
		return t.query(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// create - create a new item, store into chaincode state
// ============================================================
func (t *PurchaseChaincode) create(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// ==== Input sanitation ====
	fmt.Println("- start create item")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	var p purchase
	itemJSONasBytes := []byte(args[0])
	if err := json.Unmarshal(itemJSONasBytes, &p); err != nil {
		msg := fmt.Sprintf("-> Invalid json format - %s", args[0])
		return shim.Error(msg)
	}
	if p.CompanyID == nil {
		return shim.Error("company_id must be required")
	}
	if p.OrderID == nil {
		return shim.Error("order_id must be required")
	}
	// key := fmt.Sprintf("%s-%s", strconv.Itoa(*p.CompanyID), strconv.Itoa(*p.OrderID))
	key := fmt.Sprintf("%s-%s", *p.CompanyID, strconv.Itoa(*p.OrderID))

	fmt.Printf("key %s\n", key)

	// ==== Check if item already exists ====
	itemAsBytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Failed to get item: " + err.Error())
	} else if itemAsBytes != nil {
		msg := fmt.Sprintf("The key %s has already existed!", key)
		return shim.Error(msg)
	}

	// === Save item to state ===
	err = stub.PutState(key, itemJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Item saved and indexed. Return success ====
	fmt.Println("- end create item")
	return shim.Success(nil)
}

// ==================================================
// delete - remove a item from state
// ==================================================
func (t *PurchaseChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var itemJSON purchase

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// ==== Input sanitation ====
	fmt.Println("- start delete item")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	companyID := args[0]
	id := args[1]
	key := fmt.Sprintf("%s-%s", companyID, id)

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
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + key + "\"}"
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
func (t *PurchaseChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("- start query item")
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// ==== Input sanitation ====
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	// var idx index
	// itemJSONasBytes := []byte(args[0])
	// if err := json.Unmarshal(itemJSONasBytes, &idx); err != nil {
	// 	msg := fmt.Sprintf("Invalid json format - %s", args[0])
	// 	return shim.Error(msg)
	// }
	// key := fmt.Sprintf("%s-%s", idx.CompanyID, strconv.Itoa(idx.OrderID))
	key := args[0]

	// Get the state from the ledger
	itemAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	}

	if itemAsbytes == nil {
		jsonResp := "{\"Error\":\"Nil selling for " + key + "\"}"
		return shim.Error(jsonResp)
	}

	fmt.Printf("Query Response:%s\n", string(itemAsbytes[:]))

	fmt.Println("- end query item")
	return shim.Success(itemAsbytes)
}

func (t *PurchaseChaincode) getHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("- start getHistory item")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	companyID := args[0]
	id := args[1]
	key := fmt.Sprintf("%s-%s", companyID, id)

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

	fmt.Printf("- getHistoryForPurchase returning:\n%s\n", buffer.String())

	fmt.Println("- end getHistory item")
	return shim.Success(buffer.Bytes())
}
