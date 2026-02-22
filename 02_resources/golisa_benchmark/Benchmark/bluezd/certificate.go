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
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type CertificateChaincode struct {
}

type Certificate struct {
	CertificateID   string `json:"CertificateID"`
	CertificateHash string `json:"CertificateHash"`
}

// Init initializes chaincode
// ===========================
func (t *CertificateChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *CertificateChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Printf("## invoke is running function: %s ##\n", function)

	// Handle different functions
	if function == "createCertificate" {
		return createCertificate(stub, args)
	} else if function == "queryCertificate" {
		return queryCertificate(stub, args)
	} else if function == "removeCertificate" {
		return removeCertificate(stub, args)
	} else if function == "updateCertificate" {
		return updateCertificate(stub, args)
	} else if function == "queryHistoryCertificate" {
		return queryHistoryCertificate(stub, args)
	} else if function == "queryAllCertificate" {
		return queryAllCertificate(stub, args)
	}

	fmt.Printf("!! invalid function: %s !!", function)
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// CreateCertificate
// ============================================================
func createCertificate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	var certificateID int

	if len(args) != 2 {
		return shim.Error("!! Incorrect number of arguments, Expecting 2 !!")
	}

	// ==== Input Check ====
	fmt.Println("- start CreateCertificate -")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument username must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument enterprisename must be a non-empty string")
	}

	if certificateID, err = strconv.Atoi(args[0]); err != nil {
		return shim.Error("certificateID should be a vaild numeric string")
	}
	certificateHash := args[1]

	currentTime := timeHelper()
	fmt.Printf("[%s] <create> certificate %d", currentTime, certificateID)

	// construct the key
	key := args[0]

	// Check the Record in State
	certificateRecordAsBytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Failed to get certificate record: " + err.Error())
	} else if certificateRecordAsBytes != nil {
		return shim.Error("Certificate record already exists: " + key)
	}

	certiticateRecord := &Certificate{CertificateID: key, CertificateHash: certificateHash}
	if certificateRecordAsBytes, err = json.Marshal(certiticateRecord); err != nil {
		return shim.Error(err.Error())
	}

	// Add Record to State
	if err = stub.PutState(key, certificateRecordAsBytes); err != nil {
		return shim.Error(err.Error())
	}

	indexName := "id~all"
	if err = createIndex(stub, indexName, []string{certiticateRecord.CertificateID, certiticateRecord.CertificateHash}); err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end CreateCertificate")
	return shim.Success(nil)
}

func updateCertificate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 2 {
		return shim.Error("!! Incorrect number of arguments, Expecting 3 !!")
	}

	// ==== Input Check ====
	fmt.Println("- start updateCertificate -")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument username must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument enterprisename must be a non-empty string")
	}

	certificateHash := args[1]

	// construct the key
	key := args[0]

	// Check the Record in State
	certificateRecordAsBytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Failed to get certificate record: " + err.Error())
	} else if certificateRecordAsBytes == nil {
		return shim.Error("Certificate record does not exists: " + key)
	}

	certiticateRecord := &Certificate{}
	if err = json.Unmarshal(certificateRecordAsBytes, certiticateRecord); err != nil {
		return shim.Error(err.Error())
	}

	// Remove search index
	indexName := "id~all"
	if err = deleteIndex(stub, indexName, []string{certiticateRecord.CertificateID, certiticateRecord.CertificateHash}); err != nil {
		return shim.Error(err.Error())
	}
	certiticateRecord.CertificateHash = certificateHash

	certificateRecordJSONBytes, err := json.Marshal(certiticateRecord)
	if err != nil {
		return shim.Error(err.Error())
	}

	currentTime := timeHelper()
	fmt.Printf("[%s] <update> certificate %s", currentTime, key)

	// Add Record to State
	if err = stub.PutState(key, certificateRecordJSONBytes); err != nil {
		return shim.Error(err.Error())
	}

	// Add search index
	indexName = "username~all"
	if err = createIndex(stub, indexName, []string{certiticateRecord.CertificateID, certiticateRecord.CertificateHash}); err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end updateCertificate")
	return shim.Success(nil)
}

func queryCertificate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("!! Incorrect number of arguments, Expecting 1 !!")
	}

	// construct the key
	key := args[0]

	fmt.Printf("- start queryCertificate: %s\n", key)
	certificateRecordAsBytes, err := stub.GetState(key)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + key + "\"}"
		fmt.Println(jsonResp)
		return shim.Error(jsonResp)
	} else if certificateRecordAsBytes == nil {
		jsonResp := "{\"Error\":\" ID does not exist: " + key + "\"}"
		fmt.Println(jsonResp)
		return shim.Error(jsonResp)
	}

	fmt.Println("- end queryCertificate")
	return shim.Success(certificateRecordAsBytes)
}

func queryHistoryCertificate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("!! Incorrect number of arguments, Expecting 2 !!")
	}

	// construct the key
	key := args[0]

	fmt.Printf("- start queryHistoryCertificate: %s\n", key)

	resultsIterator, err := stub.GetHistoryForKey(key)
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

	fmt.Printf("- end queryHistoryCertificate:\n   %s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func removeCertificate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("!! Incorrect number of arguments. Expecting 1 !!")
	}

	fmt.Println("- start removeCertificate -")
	key := args[0]

	certificateAsBytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Failed to get Certificate record: " + err.Error())
	} else if certificateAsBytes == nil {
		fmt.Println("Certificate does not exist: " + key)
		return shim.Error("Certificate does not exist: " + key)
	}

	err = stub.DelState(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to remove Certificate record: %s", key))
	}

	// delete index
	certificateRecord := &Certificate{}
	json.Unmarshal(certificateAsBytes, certificateRecord)
	indexName := "id~all"
	if err = deleteIndex(stub, indexName, []string{certificateRecord.CertificateID, certificateRecord.CertificateHash}); err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end removeCertificate -")
	return shim.Success(nil)
}

func queryAllCertificate(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	startKey := ""
	endKey := ""

	fmt.Println("- start queryAllCertificate -")

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
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
	fmt.Printf("- end queryAllCertificate:\n%s\n", buffer.String())
	return shim.Success(buffer.Bytes())
}

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

func timeHelper() string {
	t := time.Now()
	currentTime := t.Format("2006-01-02T15:04:05")
	return currentTime
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(CertificateChaincode))
	if err != nil {
		fmt.Printf("Error starting Parts Trace chaincode: %s", err)
	}
}
