/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type IntegralChaincode struct {
}

type Integral struct {
	UserName       string `json:"userName"`
	EnterpriseName string `json:"enterpriseName"`
	IntegralCount  int    `json:"integralCount"`
	AddNote        string `json:"addNote"`
}

// Init initializes chaincode
// ===========================
func (t *IntegralChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *IntegralChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Printf("## invoke is running function: %s ##\n", function)

	// Handle different functions
	if function == "initIntegral" {
		return t.initIntegral(stub, args)
	} else if function == "addIntegral" {
		return t.addIntegral(stub, args)
	} else if function == "convertIntegral" {
		return t.convertIntegral(stub, args)
	} else if function == "queryIntegral" {
		return t.queryIntegral(stub, args)
	} else if function == "queryHistoryIntegral" {
		return t.queryHistoryIntegral(stub, args)
	} else if function == "queryIntegralBasedOnUser" {
		return t.queryIntegralBasedOnUser(stub, args)
	}

	//} else if function == "queryIntegralByUser" {
	//	return t.queryIntegralByUser(stub, args)
	//}

	//} else if function == "deleteIntegral" {
	//		return t.deleteIntegral(stub, args)
	//}

	fmt.Printf("!! invalid function: %s !!", function)
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// initVehiclePart - create a new vehicle part, store into chaincode state
// ============================================================
func (t *IntegralChaincode) initIntegral(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 4 {
		return shim.Error("!! Incorrect number of arguments, Expecting 4 !!")
	}

	// ==== Input Check ====
	fmt.Println("- start init integral -")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument username must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument enterprisename must be a non-empty string")
	}

	userName := strings.ToLower(args[0])
	enterpriseName := strings.ToLower(args[1])
	integralCount, _ := strconv.Atoi(args[2])
	addNote := args[3]

	currentTime := timeHelper()
	if len(args[3]) <= 0 {
		addNote = fmt.Sprintf("[%s] <init> %d integral", currentTime, integralCount)
	}

	if res := enterpriseNameCheck(enterpriseName); res != true {
		s := fmt.Sprintf("!! Invalid Enterprise Name: %s !!\n", enterpriseName)
		fmt.Printf(s)
		return shim.Error(s)
	}

	// construct the key
	keyComposite := userName + "-" + enterpriseName

	// Check the Record in State
	integralRecordAsBytes, err := stub.GetState(keyComposite)
	if err != nil {
		return shim.Error("Failed to get integral record: " + err.Error())
	} else if integralRecordAsBytes != nil {
		fmt.Println("integral record already exists: " + keyComposite)
		//return shim.Error("Integral UserName already exists: " + userName)
	}

	integralRecord := &Integral{userName, enterpriseName, integralCount, addNote}
	integralRecordAsBytes, err = json.Marshal(integralRecord)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Add Record to State
	err = stub.PutState(keyComposite, integralRecordAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	indexName := "username~all"
	err = createIndex(stub, indexName, []string{integralRecord.UserName, integralRecord.EnterpriseName, strconv.Itoa(integralRecord.IntegralCount)})
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end init integral")
	return shim.Success(nil)
}

func (t *IntegralChaincode) addIntegral(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 3 {
		return shim.Error("!! Incorrect number of arguments, Expecting 3 !!")
	}

	// ==== Input Check ====
	fmt.Println("- start add integral -")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument username must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument enterprisename must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3nd argument enterprisename must be a non-empty string")
	}

	userName := strings.ToLower(args[0])
	enterpriseName := strings.ToLower(args[1])
	integralCount, _ := strconv.Atoi(args[2])

	currentTime := timeHelper()

	if res := enterpriseNameCheck(enterpriseName); res != true {
		s := fmt.Sprintf("!! Invalid Enterprise Name: %s !!\n", enterpriseName)
		fmt.Printf(s)
		return shim.Error(s)
	}

	// construct the key
	keyComposite := userName + "-" + enterpriseName

	// Check the Record in State
	integralRecordAsBytes, err := stub.GetState(keyComposite)
	if err != nil {
		return shim.Error("Failed to get integral record: " + err.Error())
	} else if integralRecordAsBytes == nil {
		fmt.Printf("!! integral record does not exist: %s !!", keyComposite)
		return shim.Error("Integral UserName does not exist: " + keyComposite)
	}

	integralRecord := &Integral{}
	err = json.Unmarshal(integralRecordAsBytes, integralRecord)

	// Remove search index
	indexName := "username~all"
	err = deleteIndex(stub, indexName, []string{integralRecord.UserName, integralRecord.EnterpriseName, strconv.Itoa(integralRecord.IntegralCount)})
	if err != nil {
		return shim.Error(err.Error())
	}
	integralRecord.IntegralCount += integralCount
	integralRecord.AddNote = fmt.Sprintf("[%s] <add> %d integral", currentTime, integralCount)

	integralRecordJSONBytes, err := json.Marshal(integralRecord)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Add Record to State
	err = stub.PutState(keyComposite, integralRecordJSONBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Add search index
	indexName = "username~all"
	err = createIndex(stub, indexName, []string{integralRecord.UserName, integralRecord.EnterpriseName, strconv.Itoa(integralRecord.IntegralCount)})
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end add integral")
	return shim.Success(nil)
}

func (t *IntegralChaincode) queryIntegral(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 2 {
		return shim.Error("!! Incorrect number of arguments, Expecting 2 !!")
	}
	userName := strings.ToLower(args[0])
	enterpriseName := strings.ToLower(args[1])

	// construct the key
	keyComposite := userName + "-" + enterpriseName

	fmt.Printf("- start queryIntegral: %s\n", keyComposite)
	integralRecordAsBytes, err := stub.GetState(keyComposite)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + keyComposite + "\"}"
		fmt.Println(jsonResp)
		return shim.Error(jsonResp)
	} else if integralRecordAsBytes == nil {
		jsonResp := "{\"Error\":\" user enterprise does not exist: " + keyComposite + "\"}"
		fmt.Println(jsonResp)
		return shim.Error(jsonResp)
	}

	fmt.Println("- end queryIntegral")
	return shim.Success(integralRecordAsBytes)
}

func (t *IntegralChaincode) queryHistoryIntegral(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 2 {
		return shim.Error("!! Incorrect number of arguments, Expecting 2 !!")
	}
	userName := strings.ToLower(args[0])
	enterpriseName := strings.ToLower(args[1])

	// construct the key
	keyComposite := userName + "-" + enterpriseName

	fmt.Printf("- start queryHistoryIntegral: %s\n", keyComposite)

	resultsIterator, err := stub.GetHistoryForKey(keyComposite)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

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
		//as-is (as the Value itself a JSON Integral)
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

	fmt.Printf("-  queryHistoryIntegral returning:\n   %s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (t *IntegralChaincode) convertIntegral(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 4 {
		return shim.Error("!! Incorrect number of arguments, Expecting 4 !!")
	}
	userName := strings.ToLower(args[0])
	origEnterpriseName := strings.ToLower(args[1])
	targetEnterpriseName := strings.ToLower(args[2])
	convertIntegralCount, _ := strconv.Atoi(args[3])

	if !enterpriseNameCheck(origEnterpriseName) || !enterpriseNameCheck(targetEnterpriseName) {
		s := fmt.Sprintf("!! Invalid Enterprise Name !!\n")
		fmt.Printf(s)
		return shim.Error(s)
	}

	// construct the key
	keyComposite := userName + "-" + origEnterpriseName

	fmt.Println("- start convert integral")
	// Get Orig Integral
	origIntegralRecordAsBytes, err := stub.GetState(keyComposite)
	if err != nil {
		return shim.Error("Failed to get integral record: " + err.Error())
	} else if origIntegralRecordAsBytes == nil {
		fmt.Printf("!! integral record does not exist: %s !!\n", keyComposite)
		return shim.Error("Integral UserName does not exist: " + keyComposite)
	}
	origIntegralRecord := new(Integral)
	err = json.Unmarshal(origIntegralRecordAsBytes, origIntegralRecord)
	if err != nil {
		fmt.Println("!! failed to unmarshall to origIntegralRecord !!")
		return shim.Error("Failed to unmarshall to origIntegralRecord")
	}

	// integral convert rate bank:telecom:shopping_mall = 1:2:4
	integralRateMap := map[string]int{
		"bank":          0,
		"telecom":       0,
		"shopping_mall": 0,
	}

	switch origIntegralRecord.EnterpriseName {
	case "bank":
		integralRateMap["bank"] = origIntegralRecord.IntegralCount - convertIntegralCount
		integralRateMap["telecom"] = convertIntegralCount * 2
		integralRateMap["shopping_mall"] = convertIntegralCount * 4
	case "telecom":
		integralRateMap["bank"] = convertIntegralCount / 2
		integralRateMap["telecom"] = origIntegralRecord.IntegralCount - convertIntegralCount
		integralRateMap["shopping_mall"] = convertIntegralCount * 2
		fmt.Println("   - telecom")
	case "shopping_mall":
		integralRateMap["bank"] = convertIntegralCount / 4
		integralRateMap["telecom"] = convertIntegralCount / 2
		integralRateMap["shopping_mall"] = origIntegralRecord.IntegralCount - convertIntegralCount
	default:
		fmt.Println("!! Incorrect Enterprise Name [bank, telecom, shopping_mall] !!")
		return shim.Error("Incorrect Enterprise Name [bank, telecom, shopping_mall]")
	}
	fmt.Printf("   - bank=%d, telecom=%d, shopping_mall=%d\n", integralRateMap["bank"], integralRateMap["telecom"], integralRateMap["shopping_mall"])

	currentTime := timeHelper()
	fmt.Printf("   - orig username=%s, enterprisename=%s, ingegral=%d\n", origIntegralRecord.UserName, origIntegralRecord.EnterpriseName, origIntegralRecord.IntegralCount)

	// Remove search index
	indexName := "username~all"
	err = deleteIndex(stub, indexName, []string{origIntegralRecord.UserName, origIntegralRecord.EnterpriseName, strconv.Itoa(origIntegralRecord.IntegralCount)})
	if err != nil {
		return shim.Error(err.Error())
	}

	// update the orig State
	fmt.Printf("   - update the orig (%s) state\n", keyComposite)
	origIntegralRecord.IntegralCount = integralRateMap[origIntegralRecord.EnterpriseName]
	origIntegralRecord.AddNote = fmt.Sprintf("[%s] <convert> reduce %d integral", currentTime, convertIntegralCount)
	origIntegralRecordJSONBytes, err := json.Marshal(origIntegralRecord)
	if err != nil {
		fmt.Println("!! failed to unmarshall to origIntegralRecordJSONBytes !!")
		return shim.Error(err.Error())
	}
	fmt.Printf("   - origIntegralRecordJSONBytes: %s\n", origIntegralRecordJSONBytes)
	err = stub.PutState(keyComposite, origIntegralRecordJSONBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	//delete(integralRateMap, origIntegralRecord.EnterpriseName)

	// Create search index
	indexName = "username~all"
	err = createIndex(stub, indexName, []string{origIntegralRecord.UserName, origIntegralRecord.EnterpriseName, strconv.Itoa(origIntegralRecord.IntegralCount)})
	if err != nil {
		return shim.Error(err.Error())
	}

	// handle the target
	// construct the key
	keyComposite = userName + "-" + targetEnterpriseName

	// Get target Integral
	targetIntegralRecordAsBytes, err := stub.GetState(keyComposite)
	if err != nil {
		return shim.Error("Failed to get integral record: " + err.Error())
	} else if targetIntegralRecordAsBytes == nil {
		fmt.Printf("!! integral record does not exist: %s !!\n", keyComposite)
	}

	var targetIntegralRecord *Integral

	if targetIntegralRecordAsBytes != nil {
		targetIntegralRecord = new(Integral)
		err = json.Unmarshal(targetIntegralRecordAsBytes, targetIntegralRecord)
		// Remove search index
		indexName := "username~all"
		err = deleteIndex(stub, indexName, []string{targetIntegralRecord.UserName, targetIntegralRecord.EnterpriseName, strconv.Itoa(targetIntegralRecord.IntegralCount)})
		if err != nil {
			return shim.Error(err.Error())
		}
		targetIntegralRecord.IntegralCount += integralRateMap[targetEnterpriseName]
	} else {
		targetIntegralRecord = &Integral{userName, targetEnterpriseName, integralRateMap[targetEnterpriseName], ""}
	}
	targetIntegralRecord.AddNote = fmt.Sprintf("[%s]<convert> add %d integral\n", currentTime, integralRateMap[targetEnterpriseName])
	fmt.Printf("   - target username=%s, enterprisename=%s, ingegral=%d\n", targetIntegralRecord.UserName, targetIntegralRecord.EnterpriseName, targetIntegralRecord.IntegralCount)

	// update target Integral
	fmt.Printf("   - update the target (%s) state\n", keyComposite)
	targetIntegralRecordJSONBytes, err := json.Marshal(targetIntegralRecord)
	if err != nil {
		fmt.Println("!! failed to marshall to targetIntegralRecordJSONBytes !!")
		return shim.Error(err.Error())
	}
	err = stub.PutState(keyComposite, targetIntegralRecordJSONBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	// Create search index
	indexName = "username~all"
	err = createIndex(stub, indexName, []string{targetIntegralRecord.UserName, targetIntegralRecord.EnterpriseName, strconv.Itoa(targetIntegralRecord.IntegralCount)})
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end convert integral")
	return shim.Success(nil)
}

/*
func (t *IntegralChaincode) deleteIntegral(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 2 {
		return shim.Error("!! Incorrect number of arguments, Expecting 2 !!")
	}

	fmt.Println("- start delete integral -")

	userName := strings.ToLower(args[0])
	enterpriseName := strings.ToLower(args[1])

	// construct the key
	keyComposite := userName + "-" + enterpriseName

	// Check the Record in state
	integralRecordAsBytes, err := stub.GetState(keyComposite)
	if err != nil {
		return shim.Error("Failed to get integral record: " + err.Error())
	} else if integralRecordAsBytes == nil {
		fmt.Println("integral record does not exist: " + keyComposite)
		return shim.Error("Integral UserName does not exist: " + keyComposite)
	}

	// remove the user from state
	err = stub.DelState(keyComposite)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to delete %s: %s\n", keyComposite, err.Error()))
	}

	fmt.Println("- end delete integral -")
	return shim.Success(nil)
}

func (t *IntegralChaincode) queryIntegralByUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	userName := strings.ToLower(args[0])
	queryString := fmt.Sprintf("{\"selector\":{\"userName\":\"%s\"}}", userName)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}
*/

func (t *IntegralChaincode) queryIntegralBasedOnUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	userName := strings.ToLower(args[0])
	fmt.Println("- start queryIntegralBasedOnUser ", userName)

	// Query the username~all index by color
	// This will execute a key range query on all keys starting with 'userName'
	integralResultsIterator, err := stub.GetStateByPartialCompositeKey("username~all", []string{userName})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer integralResultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	var i int
	for i = 0; integralResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable), we'll just get the marble name from the composite key
		responseRange, err := integralResultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		returnedUserName := compositeKeyParts[0]
		returnedEnterpriseName := compositeKeyParts[1]
		returnedIntegralCount, _ := strconv.Atoi(compositeKeyParts[2])
		fmt.Printf("- found a integral record from index:%s userName:%s enterpriseName:%s integral:%d\n", objectType, returnedUserName, returnedEnterpriseName, returnedIntegralCount)
		buffer.WriteString("{\"userName\":")
		buffer.WriteString("\"")
		buffer.WriteString(returnedUserName)
		buffer.WriteString("\"")

		buffer.WriteString(", \"enterpriseName\":")
		buffer.WriteString("\"")
		buffer.WriteString(returnedEnterpriseName)
		buffer.WriteString("\"")

		buffer.WriteString(", \"integralCount\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.Itoa(returnedIntegralCount))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("-  queryIntegralBasedOnUser returning:\n   %s\n", buffer.String())
	return shim.Success(buffer.Bytes())
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
/*
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
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

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}
*/

// ===============================================
// createIndex - create search index for ledger
// ===============================================
func createIndex(stub shim.ChaincodeStubInterface, indexName string, attributes []string) error {
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

func deleteIndex(stub shim.ChaincodeStubInterface, indexName string, attributes []string) error {
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

var enterpriseNameList = []string{"bank", "telecom", "shopping_mall"}

func enterpriseNameCheck(name string) bool {
	for _, value := range enterpriseNameList {
		if name == value {
			return true
		}
	}
	return false
}

func timeHelper() string {
	t := time.Now()
	currentTime := t.Format("2006-01-02T15:04:05")
	return currentTime
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(IntegralChaincode))
	if err != nil {
		fmt.Printf("Error starting Parts Trace chaincode: %s", err)
	}
}
