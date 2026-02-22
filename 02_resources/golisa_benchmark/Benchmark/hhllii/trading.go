package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type SmartContract struct {
}

type User struct {
	Name   string `json:"name"`
	Email   string `json:"email"`
	Balance  float64 `json:"balance"`
}

type Item struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Price   string `json:"price"`
	Owner  string `json:"owner"`
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryItem" {
		return s.queryItem(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "queryAllItems" {
		return s.queryAllItems(APIstub)
	} else if function == "deleteItem" {
		return s.deleteItem(APIstub, args)
	} else if function == "createItem" {
		return s.createItem(APIstub, args)
	} else if function == "updatePrice" {
		return s.updatePrice(APIstub, args)
	} else if function == "updateOwner" {
		return s.updateOwner(APIstub, args)
	} else if function == "queryAllUsers" {
		return s.queryAllUsers(APIstub)
	} else if function == "deleteUser" {
		return s.deleteUser(APIstub, args)
	} else if function == "createUser" {
		return s.createUser(APIstub, args)
	} else if function == "updateEmail" {
		return s.updateBalance(APIstub, args)
	} else if function == "updateEmail" {
		return s.updateBalance(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryItem(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	itemAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(itemAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	items := []Item{
		Item{Name: "RTX2080", Type: "Computer accessories", Price: "2000", Owner: "USER1"},
		Item{Name: "Toyota Prius blue", Type: "Car", Price: "40000", Owner: "USER2"},
		Item{Name: "Coke", Type: "Drinks", Price: "10", Owner: "USER3"},
	}

	i := 0
	for i < len(items) {
		fmt.Println("i is ", i)
		itemAsBytes, _ := json.Marshal(items[i])
		APIstub.PutState("ITEM"+strconv.Itoa(i), itemAsBytes)
		fmt.Println("Added", items[i])
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) queryAllItems(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "ITEM0"
	endKey := "ITEM999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
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

	fmt.Printf("- queryAllitems:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) changeItemOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	itemAsBytes, _ := APIstub.GetState(args[0])
	item := Item{}

	json.Unmarshal(itemAsBytes, &item)
	var preOwner = item.Owner 

	preuserAsBytes, _ := APIstub.GetState(preOwner)
	preUser := User{}
	json.Unmarshal(preuserAsBytes, &preUser)

	curuserAsBytes, _ := APIstub.GetState(args[1])
	curUser := User{}
	json.Unmarshal(curuserAsBytes, &curUser)

	var price = item.Price
	// update the pre owner balance
	valueChange , _ := strconv.ParseFloat(price, 64);
	preUser.Balance += valueChange
	fmt.Println("User update get balance:",preOwner)
	preuserAsBytes, _ = json.Marshal(preUser)
	APIstub.PutState(preOwner, preuserAsBytes)

	// update the current owner balance
	preUser.Balance -= valueChange
	fmt.Println("User update pay balance:",args[1])
	curuserAsBytes, _ = json.Marshal(curUser)
	APIstub.PutState(args[1], curuserAsBytes)

	//update the item owner
	item.Owner = args[1]
	itemAsBytes, _ = json.Marshal(item)
	APIstub.PutState(args[0], itemAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) deleteItem(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	_, err := APIstub.GetState(args[0])
	if err != nil {
		return shim.Error("{\"Error\":\"Failed to get the item")
	 }

	APIstub.DelState(args[0])
	return shim.Success(nil)
}

func (s *SmartContract) createItem(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5{
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	var item = Item{Name: args[1], Type: args[2], Price: args[3], Owner: args[4]}
        fmt.Println("New item added:", item)
		itemAsBytes, _ := json.Marshal(item)
        fmt.Println("New args[0]:", args[0])
        fmt.Println("New itemAsBytes:", itemAsBytes)
	APIstub.PutState(args[0], itemAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) updatePrice(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
		fmt.Println("Item update start")
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	itemAsBytes, _ := APIstub.GetState(args[0])
	item := Item{}

	json.Unmarshal(itemAsBytes, &item)
	item.Price = args[1]
	fmt.Println("Item update price:",args[1])
	itemAsBytes, _ = json.Marshal(item)
	APIstub.PutState(args[0], itemAsBytes)

	fmt.Println("Item update end")
	return shim.Success(nil)
}

func (s *SmartContract) updateOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	fmt.Println("Item update owner")
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	itemAsBytes, _ := APIstub.GetState(args[0])
	item := Item{}

	json.Unmarshal(itemAsBytes, &item)
	item.Owner = args[1]
		fmt.Println("Item update owner:",args[1])
		itemAsBytes, _ = json.Marshal(item)
	APIstub.PutState(args[0], itemAsBytes)

		fmt.Println("Item update end")
	return shim.Success(nil)
}

func (s *SmartContract) queryAllUsers(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "User0"
	endKey := "User999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
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

	fmt.Printf("- queryAllitems:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) deleteUser(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	_, err := APIstub.GetState(args[0])
	if err != nil {
		return shim.Error("{\"Error\":\"Failed to get the user")
	 }

	APIstub.DelState(args[0])
	return shim.Success(nil)
}

func (s *SmartContract) createUser(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 4{
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	var user = User{Name: args[1], Email: args[2], Balance: 0}
        fmt.Println("New user added:", user)
		userAsBytes, _ := json.Marshal(user)
        fmt.Println("New args[0]:", args[0])
        fmt.Println("New userAsBytes:", userAsBytes)
	APIstub.PutState(args[0], userAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) updateEmail(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	fmt.Println("User update start")
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	userAsBytes, _ := APIstub.GetState(args[0])
	user := User{}

	json.Unmarshal(userAsBytes, &user)
	user.Email = args[1]
	fmt.Println("User update price:",args[1])
	userAsBytes, _ = json.Marshal(user)
	APIstub.PutState(args[0], userAsBytes)

	fmt.Println("User update end")
	return shim.Success(nil)
}

func (s *SmartContract) updateBalance(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	fmt.Println("User update start")
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	userAsBytes, _ := APIstub.GetState(args[0])
	user := User{}

	json.Unmarshal(userAsBytes, &user)
	valueChange , _ := strconv.ParseFloat(args[1], 64);
	user.Balance += valueChange
	fmt.Println("User update price:",args[1])
	userAsBytes, _ = json.Marshal(user)
	APIstub.PutState(args[0], userAsBytes)

	fmt.Println("User update end")
	return shim.Success(nil)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}