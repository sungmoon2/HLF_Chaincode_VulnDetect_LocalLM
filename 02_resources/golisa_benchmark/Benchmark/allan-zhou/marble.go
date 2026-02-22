package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type Chaincode struct{}

const MarbleTypeName = "marble"

type Marble struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Id         string `json:"Id"`
	Color      string `json:"color"`
	Size       int    `json:"size"`
	Owner      string `json:"owner"`
}

func (cc *Chaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (cc *Chaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()

	if fn == "getMarblesByRange" {
		return cc.getMarblesByRange(stub, args)
	} else if fn == "getMarbleById" {
		return cc.getMarbleById(stub, args)
	} else if fn == "addMarble" {
		return cc.addMarble(stub, args)
	} else if fn == "deleteMarble" {
		return cc.deleteMarble(stub, args)
	} else if fn == "changeMarbleOwnerById" {
		return cc.changeMarbleOwnerById(stub, args)
	} else if fn == "getMarblesByColor" {
		return cc.getMarblesByColor(stub, args)
	} else if fn == "changeMarblesOwnerByColor" {
		return cc.changeMarblesOwnerByColor(stub, args)
	} else if fn == "getHistoryForMarble" {
		return cc.getHistoryForMarble(stub, args)
	}

	return shim.Success(nil)
}

func (cc *Chaincode) getMarblesByRange(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expect 2")
	}

	// 0         1
	//startKey endKey
	startKey := MarbleTypeName + args[0]
	endKey := MarbleTypeName + args[1]
	queryIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer queryIterator.Close()

	var buffer bytes.Buffer
	isAddComma := false
	buffer.WriteString("[")
	for queryIterator.HasNext() {
		queryResult, err := queryIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if isAddComma == true {
			buffer.WriteString(",")
		}
		// [{"key":"the key","value":{"Id":"the Id"}},{},{}]
		buffer.WriteString("{")
		buffer.WriteString("\"key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResult.Key)
		buffer.WriteString("\",")
		buffer.WriteString("\"value\":")
		buffer.WriteString(string(queryResult.Value))
		buffer.WriteString("}")

		isAddComma = true
	}
	buffer.WriteString("]")

	fmt.Printf("getMarblesByRange: %s \n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (cc *Chaincode) getMarbleById(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expect 1")
	}
	// 0
	// Id
	marbleKey := MarbleTypeName + args[0]
	marbleAsBytes, err := stub.GetState(marbleKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	if marbleAsBytes == nil {
		return shim.Error("nil is found for marble Key:" + marbleKey)
	}

	fmt.Printf("getMarbleById: \n%s \n", string(marbleAsBytes))

	return shim.Success(marbleAsBytes)
}

func (cc *Chaincode) addMarble(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expect 4")
	}
	// 0    1    2     3
	// Id Color Size Owner
	size, _ := strconv.Atoi(args[2])
	marble := Marble{MarbleTypeName, args[0], args[1], size, args[3]}
	marbleAsBytes, _ := json.Marshal(marble)
	stub.PutState(MarbleTypeName+args[0], marbleAsBytes)

	fmt.Printf("addMarble: %s \n", string(marbleAsBytes))

	compositeKeyType := "color~id"
	compositeKey, _ := stub.CreateCompositeKey(compositeKeyType, []string{marble.Color, marble.Id})
	compositeKeyValue := []byte{0x00}
	stub.PutState(compositeKey, compositeKeyValue)

	fmt.Printf("add compoisteKey: %s \n", compositeKey)

	return shim.Success(marbleAsBytes)
}

func (cc *Chaincode) deleteMarble(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expect 1")
	}

	marbleKey := MarbleTypeName + args[0]
	stub.DelState(marbleKey)

	fmt.Printf("delte Marble for key: %s", marbleKey)
	//todo delte compositeKey
	compositeKeyType := "color~id"
	compositeKey, _ := stub.CreateCompositeKey(compositeKeyType, []string{marble.Color, marble.Id})
	stub.DelState(compositeKey)

	fmt.Printf("delte compositeKey: %s", compositeKey)

	return shim.Success(nil)
}

func (cc *Chaincode) changeMarbleOwnerById(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expect 2")
	}
	// 0     1
	// Id  newOwner
	newOwner := args[1]
	marbleKey := MarbleTypeName + args[0]
	marbleAsBytes, err := stub.GetState(marbleKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	if marbleAsBytes == nil {
		return shim.Error("nil is found for marble Key:" + marbleKey)
	}

	var marble Marble
	olderOwner := marble.Owner
	json.Unmarshal(marbleAsBytes, &marble)
	marble.Owner = newOwner

	newMarbleAsBytes, _ := json.Marshal(marble)
	stub.PutState(marbleKey, newMarbleAsBytes)

	fmt.Printf("change marble %s owner from %s to %s \n", marbleKey, olderOwner, newOwner)

	return shim.Success(nil)
}

func (cc *Chaincode) getMarblesByColor(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expect 1")
	}
	//  0
	// color
	color := args[0]
	compositeKeyType := "color~id"
	coloredQueryIterator, err := stub.GetStateByPartialCompositeKey(compositeKeyType, []string{color})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer coloredQueryIterator.Close()

	var buffer bytes.Buffer
	isAddComma := false
	buffer.WriteString("[")
	for coloredQueryIterator.HasNext() {
		queryResult, err := coloredQueryIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		objectType, attributes, _ := stub.SplitCompositeKey(queryResult.Key)
		compositeColor := attributes[0]
		compositeId := attributes[1]
		fmt.Printf("find a marble from index:%s color:%s id:%s\n", objectType, compositeColor, compositeId)

		var marble Marble
		marbleKey := MarbleTypeName + compositeId
		marbleAsBytes, err := stub.GetState(marbleKey)
		json.Unmarshal(marbleAsBytes, &marble)

		if marble.Color == compositeColor {
			if isAddComma == true {
				buffer.WriteString(",")
			}
			// [{"key":"the key","value":{"Id":"the Id"}},{},{}]
			buffer.WriteString("{")
			buffer.WriteString("\"key\":")
			buffer.WriteString("\"")
			buffer.WriteString(marbleKey)
			buffer.WriteString("\",")
			buffer.WriteString("\"value\":")
			buffer.WriteString(string(marbleAsBytes))
			buffer.WriteString("}")
			isAddComma = true
		}
	}
	buffer.WriteString("]")

	fmt.Printf("getMarblesByColor: %s \n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (cc *Chaincode) changeMarblesOwnerByColor(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expect 2")
	}
	// 0        1
	// color newOwnerId
	color := args[0]
	newOwner := args[1]

	compositeKeyType := "color~id"
	coloredQueryIterator, err := stub.GetStateByPartialCompositeKey(compositeKeyType, []string{color})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer coloredQueryIterator.Close()

	var count int
	for coloredQueryIterator.HasNext() {
		queryResult, err := coloredQueryIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		objectType, attributes, _ := stub.SplitCompositeKey(queryResult.Key)
		compositeColor := attributes[0]
		compositeId := attributes[1]
		fmt.Printf("find a marble from index:%s color:%s id:%s\n", objectType, compositeColor, compositeId)

		response := cc.changeMarbleOwnerById(stub, []string{compositeId, newOwner})
		if response.Status != shim.OK {
			return shim.Error("changeMarblesOwnerByColor failed" + response.Message)
		}
		count += 1
	}

	palyload := fmt.Sprintf("change %d %s marbles to %s", count, color, newOwner)

	return shim.Success([]byte(palyload))
}

func (cc *Chaincode) getHistoryForMarble(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expect 1")
	}

	marbleKey := MarbleTypeName + args[0]
	historyQueryIterator, err := stub.GetHistoryForKey(marbleKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer historyQueryIterator.Close()

	var buffer bytes.Buffer
	isAddComma := false
	buffer.WriteString("[")
	for historyQueryIterator.HasNext() {
		queryResult, err := historyQueryIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if isAddComma == true {
			buffer.WriteString(",")
		}

		// [{"TxId":"the key", "Value":"", "Timestamp":"", "IsDelete":""},{},{}]
		buffer.WriteString("{")
		buffer.WriteString("\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResult.TxId)
		buffer.WriteString("\"")
		buffer.WriteString(",")

		buffer.WriteString("\"value\":")
		buffer.WriteString(string(queryResult.Value))
		buffer.WriteString(",")

		buffer.WriteString("\"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(queryResult.Timestamp.Seconds, int64(queryResult.Timestamp.Nanos)).String())
		buffer.WriteString("\"")
		buffer.WriteString(",")

		buffer.WriteString("\"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(queryResult.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		isAddComma = true
	}
	buffer.WriteString("]")

	fmt.Printf("getHistoryForMarble: %s \n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func main() {
	err := shim.Start(new(Chaincode))
	if err != nil {
		fmt.Printf("error starting.")
	}
}
