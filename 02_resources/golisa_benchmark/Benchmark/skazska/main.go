package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	// "time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type Prototype struct{}

type producedInfo struct {
	Producer string `json:"producer"`
	Model    string `json:"model",omitempty`
	Serial   string `json:"serial",omitempty`
	Place    string `json:"place",omitempty`
	Time     string `json:"time",omitempty`
}

type item struct {
	ItemType string       `json:"itemType"`
	ID       string       `json:"id"`
	Location string       `json:"location",omitempty`
	Holder   string       `json:"holder",omitempty`
	Owner    string       `json:"owner",omitempty`
	Produced producedInfo `json:"produced",omitempty`
	// Produced string `json:"produced"`
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(Prototype))
	if err != nil {
		fmt.Printf("Error starting Prototype chaincode: %s", err)
	}
}

// ===================================================================================
// Chaincode interface
// ===================================================================================

// ============================================================
// Init - chaincode instantiate phase code
// ============================================================
func (t *Prototype) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

// ============================================================
// Invoke - process chaincode invocation (invoke, query)
// ============================================================
func (t *Prototype) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	args := stub.GetArgs()
	fn := string(args[0])
	dataArgs := args[1:]

	switch fn {
	case "register":
		return t.register(stub, dataArgs)
	case "transfer":
		return t.transfer(stub, dataArgs)
	case "read":
		return t.read(stub, dataArgs)
	case "rangedList":
		return t.rangedList(stub, dataArgs)
	}

	fmt.Println("invoke did not find func: " + fn) //error
	return shim.Error("Received unknown function invocation")
}

// ===================================================================================
// Chaincode
// ===================================================================================

// ============================================================
// register - create a new item record, store into chaincode state
// ============================================================
func (t *Prototype) register(stub shim.ChaincodeStubInterface, args [][]byte) peer.Response {
	var err error
	var produced producedInfo

	// ==== Input sanitation ====
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2 at least")
	}
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	fmt.Println("- start register item")

	itemID := string(args[0])
	fmt.Println("itemID, %s", itemID)

	// if err != nil {
	// 	return shim.Error("3rd argument must be a numeric string")
	// }

	// ==== Check if item already exists ====
	itemAsBytes, err := stub.GetState(itemID)
	if err != nil {
		return shim.Error("Failed to get item: " + err.Error())
	} else if itemAsBytes != nil {
		fmt.Println("This item already exists: " + itemID)
		return shim.Error("This item already exists: " + itemID)
	}

	location := string(args[2])
	fmt.Println("location, %s", location)
	fmt.Println("produced as string, %s", string(args[1]))
	json.Unmarshal(args[1], &produced)

	// ==== Create item object and marshal to JSON ====
	item := &item{
		ItemType: "item",
		ID:       itemID,
		// Location: location,
		// Holder: ,
		// Owner: ,
		Produced: produced}
	if location != "" {
		item.Location = location
	}
	err = saveItem(stub, item)
	if err != nil {
		return shim.Error(err.Error())
	}

	// //  ==== Index the item to enable color-based range queries, e.g. return all blue items ====
	// //  An 'index' is a normal key/value entry in state.
	// //  The key is a composite key, with the elements that you want to range query on listed first.
	// //  In our case, the composite key is based on indexName~color~name.
	// //  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	// indexName := "color~name"
	// colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{item.Color, item.Name})
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }
	// //  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the item.
	// //  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	// value := []byte{0x00}
	// stub.PutState(colorNameIndexKey, value)
	//
	// // ==== item saved and indexed. Return success ====
	// fmt.Println("- end init item")
	return shim.Success(nil)
}

// ============================================================
// transfer - update item record by set new value for Owner
// ============================================================
func (t *Prototype) transfer(stub shim.ChaincodeStubInterface, args [][]byte) peer.Response {
	type Payload struct {
		Holder string `json:"holder",omitempty`
		Owner  string `json:"owner",omitempty`
	}
	var item item
	var payload Payload
	// ==== Input sanitation ====
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2 at least")
	}
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	fmt.Println("- start transfer item")

	itemID := string(args[0])
	fmt.Println("itemID, %s", itemID)

	// ==== Check if item already exists ====
	itemAsBytes, err := stub.GetState(itemID)
	if err != nil {
		return shim.Error("Failed to get item: " + err.Error())
	} else if itemAsBytes == nil {
		fmt.Println("This item does not exist: " + itemID)
		return shim.Error("This item does not exist: " + itemID)
	}
	json.Unmarshal(itemAsBytes, &item)

	location := string(args[2])
	json.Unmarshal(args[1], &payload)

	fmt.Println("payload.Owner, %s", string(payload.Owner))
	if payload.Owner != "" {
		item.Owner = payload.Owner
	}

	fmt.Println("payload.Holder, %s", string(payload.Holder))
	if payload.Holder != "" {
		item.Holder = payload.Holder
	}

	if location != "" {
		item.Location = location
	}

	err = saveItem(stub, &item)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// ============================================================
// read - get item record from state
// ============================================================
func (t *Prototype) read(stub shim.ChaincodeStubInterface, args [][]byte) peer.Response {
	var err error

	// ==== Input sanitation ====
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2 at least")
	}
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty ")
	}

	fmt.Println("- start read item")

	itemID := string(args[0])
	valAsbytes, err := stub.GetState(itemID) //get the item from chaincode state
	if err != nil {
		return shim.Error("{\"Error\":\"Failed to get state for " + itemID + "\"}")
	} else if valAsbytes == nil {
		return shim.Error("{\"Error\":\"item does not exist: " + itemID + "\"}")
	}

	return shim.Success(valAsbytes)
}

// ============================================================
// rangedList - get items records from state
// ============================================================
func (t *Prototype) rangedList(stub shim.ChaincodeStubInterface, args [][]byte) peer.Response {
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	startKey := string(args[0])
	endKey := string(args[1])

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	buffer, err := composeJSONItemList(resultsIterator)

	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Printf("- getitemsByRange queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// ===================================================================================
// Chaincode helper functions
// ===================================================================================

func saveItem(stub shim.ChaincodeStubInterface, item *item) error {
	itemJSONasBytes, err := json.Marshal(item)
	if err != nil {
		return err
	}

	//Alternatively, build the item json string manually if you don't want to use struct marshalling
	//itemJSONasString := `{"docType":"item",  "name": "` + itemID + `", "color": "` + color + `", "size": ` + strconv.Itoa(size) + `, "owner": "` + owner + `"}`
	//itemJSONasBytes := []byte(str)

	// === Save item to state ===
	err = stub.PutState(item.ID, itemJSONasBytes)
	if err != nil {
		return err
	}

	return nil
}

func composeJSONItemList(iterator shim.StateQueryIteratorInterface) (bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for iterator.HasNext() {
		queryResponse, err := iterator.Next()
		if err != nil {
			return buffer, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return buffer, nil
}
