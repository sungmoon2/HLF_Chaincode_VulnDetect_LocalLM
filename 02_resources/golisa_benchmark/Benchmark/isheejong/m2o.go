package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)


// M2OChaincode example simple Chaincode implementation
type M2OChaincode struct {
}

// ExchangeRate JSON object
type ExchangeRate struct {
	ObjectType   string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	PartnerName string `json:"PartnerName"`
	ExchangeRate float64 `json:"ExchangeRate"`
}

// AccountTransaction JSON object
type AccountTransaction struct {
	ObjectType   string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	M2OUID string `json:"M2OUID"`
	M2OToken int `json:"M2OToken"`
	PartnerName string `json:"PartnerName"`
	PartnerUID string `json:"PartnerUID"`
	Date string `json:"Date"`
	FromPoints int `json:"FromPoints"`
	ToM2OToken int `json:"ToM2OToken"`
	ExchangeRate float64 `json:"ExchangeRate"`	
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(M2OChaincode))
	if err != nil {
		fmt.Printf("Error starting M2O chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *M2OChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *M2OChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "updateExchangeRate" { //create or update exchange rate for a partner
		return t.updateExchangeRate(stub, args)
	} else if function == "convertToM2OToken" { //convert points to M2O token
		return t.convertToM2OToken(stub, args)
	} else if function == "queryUserTotal" { //query user's total M2O token balance
		return t.queryUserTotal(stub, args)
	} else if function == "queryUserTransactionsForPartner" { //query user's transactions for a partner
		return t.queryUserTransactionsForPartner(stub, args)
	} else if function == "queryUserTransactionsForAll" { //get history of all the transactions for a user
		return t.queryUserTransactionsForAll(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// updateExchangeRate - create or update exchange rate for a partner
// Arguments - PartnerName, NewRate
// ============================================================
func (t *M2OChaincode) updateExchangeRate(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	fmt.Println("- start updateExchangeRate")
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	parnterName := strings.ToLower(args[0])
	exchangeRate, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return shim.Error("2nd argument must be a float32 string")
	}	

	// ==== Create ExchangeRate object and marshal to JSON ====
	objectType := "ExchangeRate"
	exchangeRateObject := &ExchangeRate{objectType, parnterName, exchangeRate}
	exchangeRateJSONasBytes, err := json.Marshal(exchangeRateObject)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save ExchangeRate to state ===
	err = stub.PutState(parnterName, exchangeRateJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}	

	fmt.Println("- end updateExchangeRate")
	var buffer bytes.Buffer
	buffer.WriteString("{\"Message\":")
	buffer.WriteString("\"")
	buffer.WriteString("Success")
	buffer.WriteString("\"")
	buffer.WriteString("}")

	return shim.Success(buffer.Bytes())	
}

func (t *M2OChaincode) init(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	return shim.Success(nil)
}

// ============================================================
// convertToM2OToken - convert points to M2O token
// Arguments - PartnerName, PartnerUID, M2OUID, FromPoints
// ============================================================
func (t *M2OChaincode) convertToM2OToken(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	fmt.Println("- start convertToM2OToken")
	
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}


	M2OUID      	:= strings.ToLower(args[0])
	parnterName 	:= strings.ToLower(args[1])
	parnterUID 		:= strings.ToLower(args[2])
	fromPoints, err := strconv.Atoi(string(args[3]))
	date			:= strings.ToLower(args[4])

	
	ExchangeRateAsBytes, err := stub.GetState(parnterName)
	exchangeRate := &ExchangeRate{}

	err = json.Unmarshal(ExchangeRateAsBytes, &exchangeRate) 
	if err != nil {
		fmt.Printf("%s ** ", err.Error())
		return shim.Error(err.Error())
	}
	
	toM2OToken := 0
	toM2OToken = fromPoints * int(exchangeRate.ExchangeRate)

	objectType := "AccountTransaction"
	lastAccountTransactoinBytes, err := stub.GetState(M2OUID)
	lastAccountTransaction := &AccountTransaction{}

	
	M2OToken := 0
	if lastAccountTransactoinBytes != nil {
		err = json.Unmarshal(lastAccountTransactoinBytes, &lastAccountTransaction)
		M2OToken = toM2OToken + lastAccountTransaction.M2OToken
	}else {
		M2OToken = toM2OToken
	}

	accountTransaction := &AccountTransaction{objectType, M2OUID, M2OToken, parnterName, parnterUID, date, fromPoints, toM2OToken, exchangeRate.ExchangeRate}
	accountTransactionBytes, err := json.Marshal(accountTransaction)
	
	err = stub.PutState(M2OUID, accountTransactionBytes)
	if err != nil{
		return shim.Error(err.Error())
	}
	
	fmt.Println("- end convertToM2OToken")
	var buffer bytes.Buffer
	buffer.WriteString("{\"Message\":")
	buffer.WriteString("\"")
	buffer.WriteString("Success")
	buffer.WriteString("\"")
	buffer.WriteString("}")

	return shim.Success(buffer.Bytes())		
}

// ============================================================
// queryUserTotal - query user's total M2O token balance
// Arguments - M2OUID
// ============================================================
func (t *M2OChaincode) queryUserTotal(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	fmt.Println("- start queryUserTotal")
	var err error
	var jsonResp string
	var accountTransaction AccountTransaction

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	M2OUID := strings.ToLower(args[0])

	valAsbytes, err := stub.GetState(M2OUID) //get the AccountTransaction from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + M2OUID + "\"}"
		fmt.Println(jsonResp)
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Account does not exist: " + M2OUID + "\"}"
		fmt.Println(jsonResp)
		return shim.Error(jsonResp)
	}

	err = json.Unmarshal([]byte(valAsbytes), &accountTransaction)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + M2OUID + "\"}"
		return shim.Error(jsonResp)
	}

	var buffer bytes.Buffer
	buffer.WriteString("{\"M2OUID\":")
	buffer.WriteString("\"")
	buffer.WriteString(accountTransaction.M2OUID)
	buffer.WriteString("\"")
	buffer.WriteString(", \"M2OToken\":")
	buffer.WriteString(strconv.Itoa(accountTransaction.M2OToken))
	buffer.WriteString("}")
	fmt.Println("- end queryUserTotal")

	return shim.Success(buffer.Bytes())	
	
}

// ============================================================
// queryUserTransactionsForAll - get history of all the transactions for a user
// Arguments - M2OUID
// ============================================================
func (t *M2OChaincode) queryUserTransactionsForAll(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	fmt.Println("- start queryUserTransactionsForAll")
	//var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	M2OUID := strings.ToLower(args[0])
	fmt.Println("-----------------------------")
	fmt.Println(M2OUID)
	resultsIterator, err := stub.GetHistoryForKey(M2OUID)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the key/value pair
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
		//as-is (as the Value itself a JSON vehiclePart)
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

	fmt.Printf("- getHistoryForRecord returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
	

	return shim.Success(nil)
}

// ============================================================
// queryUserTransactionsForPartner - query user's transactions for a partner
// Arguments - M2OUID, PartnerName
// ============================================================
func (t *M2OChaincode) queryUserTransactionsForPartner(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	fmt.Println("- start queryUserTransactionsForPartner")
	//var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	//M2OUID := strings.ToLower(args[0])
	//parnterName := strings.ToLower(args[1])

	fmt.Println("- end queryUserTransactionsForPartner")
	var buffer bytes.Buffer
	buffer.WriteString("{\"Message\":")
	buffer.WriteString("\"")
	buffer.WriteString("Success")
	buffer.WriteString("\"")
	buffer.WriteString("}")

	return shim.Success(buffer.Bytes())		
}



// ===============================================
// createIndex - create search index for ledger
// ===============================================
func (t *M2OChaincode) createIndex(stub shim.ChaincodeStubInterface, indexName string, attributes []string) error {
	fmt.Println("- start create index")
	var err error
	//  ==== Index the object to enable range queries, e.g. return all parts made by supplier b ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	indexKey, err := stub.CreateCompositeKey(indexName, attributes)
	if err != nil {
		return err
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of object.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(indexKey, value)

	fmt.Println("- end create index")
	return nil
}

// ===============================================
// deleteIndex - remove search index for ledger
// ===============================================
func (t *M2OChaincode) deleteIndex(stub shim.ChaincodeStubInterface, indexName string, attributes []string) error {
	fmt.Println("- start delete index")
	var err error
	//  ==== Index the object to enable range queries, e.g. return all parts made by supplier b ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	indexKey, err := stub.CreateCompositeKey(indexName, attributes)
	if err != nil {
		return err
	}
	//  Delete index by key
	stub.DelState(indexKey)

	fmt.Println("- end delete index")
	return nil
}
