/*
 SPDX-License-Identifier: Apache-2.0
*/

// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

// ==== Invoke rowObjects ====
// peer chaincode invoke -C myc1 -n rowObjects -c '{"Args":["addNewObject","marble1","blue","35","tom"]}'
// peer chaincode invoke -C myc1 -n rowObjects -c '{"Args":["addNewObject","marble2","red","50","tom"]}'
// peer chaincode invoke -C myc1 -n rowObjects -c '{"Args":["addNewObject","marble3","blue","70","tom"]}'
// peer chaincode invoke -C myc1 -n rowObjects -c '{"Args":["updateObject","marble2","jerry"]}'
// peer chaincode invoke -C myc1 -n rowObjects -c '{"Args":["updateObjectsBasedOnVersionName","blue","jerry"]}'
// peer chaincode invoke -C myc1 -n rowObjects -c '{"Args":["delete","marble1"]}'

// ==== Query rowObjects ====
// peer chaincode query -C myc1 -n rowObjects -c '{"Args":["getObject","marble1"]}'
// peer chaincode query -C myc1 -n rowObjects -c '{"Args":["getObjectsByRange","marble1","marble3"]}'
// peer chaincode query -C myc1 -n rowObjects -c '{"Args":["getHistoryForObject","marble1"]}'

// Rich Query (Only supported if CouchDB is used as state database):
// peer chaincode query -C myc1 -n rowObjects -c '{"Args":["queryObjectsByOwner","tom"]}'
// peer chaincode query -C myc1 -n rowObjects -c '{"Args":["queryObjects","{\"selector\":{\"compilerConfiguration\":\"tom\"}}"]}'

// Rich Query with Pagination (Only supported if CouchDB is used as state database):
// peer chaincode query -C myc1 -n rowObjects -c '{"Args":["queryObjectsWithPagination","{\"selector\":{\"compilerConfiguration\":\"tom\"}}","3",""]}'

// INDEXES TO SUPPORT COUCHDB RICH QUERIES
//
// Indexes in CouchDB are required in order to make JSON queries efficient and are required for
// any JSON query with a sort. As of Hyperledger Fabric 1.1, indexes may be packaged alongside
// chaincode in a META-INF/statedb/couchdb/indexes directory. Each index must be defined in its own
// text file with extension *.json with the index definition formatted in JSON following the
// CouchDB index JSON syntax as documented at:
// http://docs.couchdb.org/en/2.1.1/api/database/find.html#db-index
//
// This marbles02 example chaincode demonstrates a packaged
// index which you can find in META-INF/statedb/couchdb/indexes/indexOwner.json.
// For deployment of chaincode to production environments, it is recommended
// to define any indexes alongside chaincode so that the chaincode and supporting indexes
// are deployed automatically as a unit, once the chaincode has been installed on a peer and
// instantiated on a channel. See Hyperledger Fabric documentation for more details.
//
// If you have access to the your peer's CouchDB state database in a development environment,
// you may want to iteratively test various indexes in support of your chaincode queries.  You
// can use the CouchDB Fauxton interface or a command line curl utility to create and update
// indexes. Then once you finalize an index, include the index definition alongside your
// chaincode in the META-INF/statedb/couchdb/indexes directory, for packaging and deployment
// to managed environments.
//
// In the examples below you can find index definitions that support marbles02
// chaincode queries, along with the syntax that you can use in development environments
// to create the indexes in the CouchDB Fauxton interface or a curl command line utility.
//

//Example hostname:port configurations to access CouchDB.
//
//To access CouchDB docker container from within another docker container or from vagrant environments:
// http://couchdb:5984/
//
//Inside couchdb docker container
// http://127.0.0.1:5984/

// Index for docType, compilerConfiguration.
//
// Example curl command line to define index in the CouchDB channel_chaincode database
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[\"docType\",\"compilerConfiguration\"]},\"name\":\"indexOwner\",\"ddoc\":\"indexOwnerDoc\",\"type\":\"json\"}" http://hostname:port/myc1_rowObjects/_index
//

// Index for docType, compilerConfiguration, size (descending order).
//
// Example curl command line to define index in the CouchDB channel_chaincode database
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[{\"size\":\"desc\"},{\"docType\":\"desc\"},{\"compilerConfiguration\":\"desc\"}]},\"ddoc\":\"indexSizeSortDoc\", \"name\":\"indexSizeSortDesc\",\"type\":\"json\"}" http://hostname:port/myc1_rowObjects/_index

// Rich Query with index design doc and index name specified (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n rowObjects -c '{"Args":["queryObjects","{\"selector\":{\"docType\":\"rowObject\",\"compilerConfiguration\":\"tom\"}, \"use_index\":[\"_design/indexOwnerDoc\", \"indexOwner\"]}"]}'

// Rich Query with index design doc specified only (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n rowObjects -c '{"Args":["queryObjects","{\"selector\":{\"docType\":{\"$eq\":\"rowObject\"},\"compilerConfiguration\":{\"$eq\":\"tom\"},\"size\":{\"$gt\":0}},\"fields\":[\"docType\",\"compilerConfiguration\",\"size\"],\"sort\":[{\"size\":\"desc\"}],\"use_index\":\"_design/indexSizeSortDoc\"}"]}'

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

// FabricChaincode example simple Chaincode implementation
type FabricChaincode struct {
}

type rowObject struct {
	ObjectType string `json:"docType"`   //docType is used to distinguish the various types of objects in state database
	Prop01     string `json:"versionId"` //key the fieldtags are needed to keep case from bouncing around
	Prop02     string `json:"versionName"`
	Prop03     string `json:"compilerVersion"`
	Prop04     string `json:"compilerConfiguration"`
	Prop05     string `json:"creationDateTime"`
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	fmt.Println("Hello World!")
	err := shim.Start(new(FabricChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
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
	if function == "addNewObject" { //create a new rowObject
		return t.addNewObject(stub, args)
	} else if function == "updateObject" { //change compilerConfiguration of a specific rowObject
		return t.updateObject(stub, args)
	} else if function == "delete" { //delete a rowObject
		return t.delete(stub, args)
	} else if function == "getObject" { //read a rowObject
		return t.getObject(stub, args)
	} else if function == "queryObjectsByProperties" { //find rowObjects for compilerConfiguration X using rich query
		return t.queryObjectsByProperties(stub, args)
	} else if function == "selectorObjects" { //find rowObjects based on an ad hoc rich query
		return t.selectorObjects(stub, args)
	} else if function == "getHistoryForObject" { //get history of values for a rowObject
		return t.getHistoryForObject(stub, args)
	} else if function == "queryObjectsByRange" { //get rowObjects based on range query
		return t.queryObjectsByRange(stub, args)
	} else if function == "queryObjectsByRangeWithPagination" {
		return t.queryObjectsByRangeWithPagination(stub, args)
	} else if function == "queryObjectsWithPagination" {
		return t.queryObjectsWithPagination(stub, args)
	} else if function == "getRowAmount" {
		return t.getRowAmount(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// addNewObject - create a new rowObject, store into chaincode state
// ============================================================
func (t *FabricChaincode) addNewObject(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init rowObject")
	if len(args[1]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}

	rowObjectKey := args[0]
	newProp02 := args[1]
	newProp03 := args[2]
	newProp04 := args[3]
	newProp05 := args[4]

	if err != nil {
		return shim.Error("3rd argument must be a numeric string")
	}

	// ==== Check if rowObject already exists ====
	rowObjectAsBytes, err := stub.GetState(rowObjectKey)
	if err != nil {
		return shim.Error("Failed to get rowObject: " + err.Error())
	} else if rowObjectAsBytes != nil {
		fmt.Println("This rowObject already exists: " + rowObjectKey)
		return shim.Error("This rowObject already exists: " + rowObjectKey)
	}

	// ==== Create rowObject object and marshal to JSON ====
	objectType := "rowObject"
	rowObject := &rowObject{objectType, rowObjectKey, newProp02, newProp03, newProp04, newProp05}
	rowObjectJSONasBytes, err := json.Marshal(rowObject)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save rowObject to state ===
	err = stub.PutState(rowObjectKey, rowObjectJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	//  ==== Index the rowObject to enable versionName-based range queries, e.g. return all blue rowObjects ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~versionName~versionId.
	//  This will enable very efficient state range queries based on composite keys matching indexName~versionName~*
	indexName := "versionName~versionId"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{rowObject.Prop02, rowObject.Prop01})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the rowObject.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(colorNameIndexKey, value)

	// ==== Marble saved and indexed. Return success ====
	fmt.Println("- end init rowObject")
	return shim.Success(nil)
}

// ===============================================
// readObject - read a rowObject from chaincode state
// ===============================================
func (t *FabricChaincode) getObject(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var versionId, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting versionId of the rowObject to query")
	}

	versionId = args[0]
	valAsbytes, err := stub.GetState(versionId) //get the rowObject from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + versionId + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + versionId + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

func (t *FabricChaincode) getRowAmount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var err error

	valAsbytes, err := stub.GetState("amount") //get the rowObject from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to query row amount. \"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"The amount property has not set yet.\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

// ==================================================
// delete - remove a rowObject key/value pair from state
// ==================================================
func (t *FabricChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var rowObjectJSON rowObject
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	rowObjectKey := args[0]

	// to maintain the versionName~versionId index, we need to read the rowObject first and get its versionName
	valAsbytes, err := stub.GetState(rowObjectKey) //get the rowObject from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + rowObjectKey + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + rowObjectKey + "\"}"
		return shim.Error(jsonResp)
	}

	err = json.Unmarshal([]byte(valAsbytes), &rowObjectJSON)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + rowObjectKey + "\"}"
		return shim.Error(jsonResp)
	}

	err = stub.DelState(rowObjectKey) //remove the rowObject from chaincode state
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// maintain the index
	indexName := "versionName~versionId"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{rowObjectJSON.Prop02, rowObjectJSON.Prop01})
	if err != nil {
		return shim.Error(err.Error())
	}

	//  Delete index entry to state.
	err = stub.DelState(colorNameIndexKey)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}
	return shim.Success(nil)
}

// ===========================================================
// transfer a rowObject by setting a new compilerConfiguration name on the rowObject
// ===========================================================
func (t *FabricChaincode) updateObject(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	rowObjectKey := args[0]
	newProp02 := args[1]
	newProp03 := args[2]
	newProp04 := args[3]
	newProp05 := args[4]
	fmt.Println("- start updateObject ", rowObjectKey, newProp04)

	rowObjectAsBytes, err := stub.GetState(rowObjectKey)
	if err != nil {
		return shim.Error("Failed to get rowObject:" + err.Error())
	} else if rowObjectAsBytes == nil {
		return shim.Error("Marble does not exist")
	}

	newRowObject := rowObject{}
	err = json.Unmarshal(rowObjectAsBytes, &newRowObject) //unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	if newProp02 != "" {
		newRowObject.Prop02 = newProp02
	}
	if newProp03 != "" {
		newRowObject.Prop03 = newProp03
	}
	if newProp04 != "" {
		newRowObject.Prop04 = newProp04
	}
	if newProp05 != "" {
		newRowObject.Prop05 = newProp05
	}

	rowObjectJSONasBytes, _ := json.Marshal(newRowObject)
	err = stub.PutState(rowObjectKey, rowObjectJSONasBytes) //rewrite the rowObject
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end updateObject (success)")
	return shim.Success(nil)
}

// ===========================================================================================
// constructQueryResponseFromIterator constructs a JSON array containing query results from
// a given result iterator
// ===========================================================================================
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
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

	return &buffer, nil
}

// ===========================================================================================
// addPaginationMetadataToQueryResults adds QueryResponseMetadata, which contains pagination
// info, to the constructed query results
// ===========================================================================================
func addPaginationMetadataToQueryResults(buffer *bytes.Buffer, responseMetadata *pb.QueryResponseMetadata) *bytes.Buffer {

	buffer.WriteString("[{\"ResponseMetadata\":{\"RecordsCount\":")
	buffer.WriteString("\"")
	buffer.WriteString(fmt.Sprintf("%v", responseMetadata.FetchedRecordsCount))
	buffer.WriteString("\"")
	buffer.WriteString(", \"Bookmark\":")
	buffer.WriteString("\"")
	buffer.WriteString(responseMetadata.Bookmark)
	buffer.WriteString("\"}}]")

	return buffer
}

// ===========================================================================================
// getObjectsByRange performs a range query based on the start and end keys provided.

// Read-only function results are not typically submitted to ordering. If the read-only
// results are submitted to ordering, or if the query is used in an update transaction
// and submitted to ordering, then the committing peers will re-execute to guarantee that
// result sets are stable between endorsement time and commit time. The transaction is
// invalidated by the committing peers if the result set has changed between endorsement
// time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *FabricChaincode) queryObjectsByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	startKey := args[0]
	endKey := args[1]

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Printf("- getObjectsByRange queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// ==== Example: GetStateByPartialCompositeKey/RangeQuery =========================================
// updateObjectsBasedOnVersionName will transfer rowObjects of a given versionName to a certain new compilerConfiguration.
// Uses a GetStateByPartialCompositeKey (range query) against versionName~versionId 'index'.
// Committing peers will re-execute range queries to guarantee that result sets are stable
// between endorsement time and commit time. The transaction is invalidated by the
// committing peers if the result set has changed between endorsement time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
// ===========================================================================================
func (t *FabricChaincode) updateObjectsBasedOnVersionName(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0       1
	// "versionName", "bob"
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	versionName := args[0]
	newOwner := args[1]
	fmt.Println("- start updateObjectsBasedOnVersionName ", versionName, newOwner)

	// Query the versionName~versionId index by versionName
	// This will execute a key range query on all keys starting with 'versionName'
	coloredMarbleResultsIterator, err := stub.GetStateByPartialCompositeKey("versionName~versionId", []string{versionName})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer coloredMarbleResultsIterator.Close()

	// Iterate through result set and for each rowObject found, transfer to newOwner
	var i int
	for i = 0; coloredMarbleResultsIterator.HasNext(); i++ {
		// Note that we don't get the value (2nd return variable), we'll just get the rowObject versionId from the composite key
		responseRange, err := coloredMarbleResultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// get the versionName and versionId from versionName~versionId composite key
		objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}
		returnedVersionName := compositeKeyParts[0]
		returnedMarbleName := compositeKeyParts[1]
		fmt.Printf("- found a rowObject from index:%s versionName:%s versionId:%s\n", objectType, returnedVersionName, returnedMarbleName)

		// Now call the transfer function for the found rowObject.
		// Re-use the same function that is used to transfer individual rowObjects
		response := t.updateObject(stub, []string{returnedMarbleName, newOwner})
		// if the transfer failed break out of loop and return error
		if response.Status != shim.OK {
			return shim.Error("Transfer failed: " + response.Message)
		}
	}

	responsePayload := fmt.Sprintf("Transferred %d %s rowObjects to %s", i, versionName, newOwner)
	fmt.Println("- end updateObjectsBasedOnVersionName: " + responsePayload)
	return shim.Success([]byte(responsePayload))
}

// =======Rich queries =========================================================================
// Two examples of rich queries are provided below (parameterized query and ad hoc query).
// Rich queries pass a query string to the state database.
// Rich queries are only supported by state database implementations
//  that support rich query (e.g. CouchDB).
// The query string is in the syntax of the underlying state database.
// With rich queries there is no guarantee that the result set hasn't changed between
//  endorsement time and commit time, aka 'phantom reads'.
// Therefore, rich queries should not be used in update transactions, unless the
// application handles the possibility of result set changes between endorsement and commit time.
// Rich queries can be used for point-in-time queries against a peer.
// ============================================================================================

// ===== Example: Parameterized rich query =================================================
// queryObjectsByOwner queries for rowObjects based on a passed in compilerConfiguration.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (compilerConfiguration).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *FabricChaincode) queryObjectsByProperties(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	var subString string
	for i := 0; i < len(args)/2; i = i + 2 {
		propKey := args[i]
		propValue := args[i+1]
		subString = subString + fmt.Sprintf("\"%s\":\"%s\"", propKey, propValue)
		if i < len(args)/2-2 {
			subString = subString + ","
		}
	}
	queryString := fmt.Sprintf("{\"selector\":{%s}}", subString)
	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// ===== Example: Ad hoc rich query ========================================================
// queryObjects uses a query string to perform a query for rowObjects.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryObjectsForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (t *FabricChaincode) selectorObjects(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

// ====== Pagination =========================================================================
// Pagination provides a method to retrieve records with a defined pagesize and
// start point (bookmark).  An empty string bookmark defines the first "page" of a query
// result.  Paginated queries return a bookmark that can be used in
// the next query to retrieve the next page of results.  Paginated queries extend
// rich queries and range queries to include a pagesize and bookmark.
//
// Two examples are provided in this example.  The first is getObjectsByRangeWithPagination
// which executes a paginated range query.
// The second example is a paginated query for rich ad-hoc queries.
// =========================================================================================

// ====== Example: Pagination with Range Query ===============================================
// getObjectsByRangeWithPagination performs a range query based on the start & end key,
// page size and a bookmark.

// The number of fetched records will be equal to or lesser than the page size.
// Paginated range queries are only valid for read only transactions.
// ===========================================================================================
func (t *FabricChaincode) queryObjectsByRangeWithPagination(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	startKey := args[0]
	endKey := args[1]
	//return type of ParseInt is int64
	pageSize, err := strconv.ParseInt(args[2], 10, 32)
	if err != nil {
		return shim.Error(err.Error())
	}
	bookmark := args[3]

	resultsIterator, responseMetadata, err := stub.GetStateByRangeWithPagination(startKey, endKey, int32(pageSize), bookmark)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return shim.Error(err.Error())
	}

	bufferWithPaginationInfo := addPaginationMetadataToQueryResults(buffer, responseMetadata)

	fmt.Printf("- getObjectsByRange queryResult:\n%s\n", bufferWithPaginationInfo.String())

	return shim.Success(buffer.Bytes())
}

// ===== Example: Pagination with Ad hoc Rich Query ========================================================
// queryObjectsWithPagination uses a query string, page size and a bookmark to perform a query
// for rowObjects. Query string matching state database syntax is passed in and executed as is.
// The number of fetched records would be equal to or lesser than the specified page size.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the queryObjectsForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// Paginated queries are only valid for read only transactions.
// =========================================================================================
func (t *FabricChaincode) queryObjectsWithPagination(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	queryString := args[0]
	//return type of ParseInt is int64
	pageSize, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		return shim.Error(err.Error())
	}
	bookmark := args[2]

	queryResults, err := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookmark)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryStringWithPagination executes the passed in query string with
// pagination info. Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryStringWithPagination(stub shim.ChaincodeStubInterface, queryString string, pageSize int32, bookmark string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, responseMetadata, err := stub.GetQueryResultWithPagination(queryString, pageSize, bookmark)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	bufferWithPaginationInfo := addPaginationMetadataToQueryResults(buffer, responseMetadata)

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", bufferWithPaginationInfo.String())

	return buffer.Bytes(), nil
}

func (t *FabricChaincode) getHistoryForObject(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	rowObjectKey := args[0]

	fmt.Printf("- start getHistoryForObject: %s\n", rowObjectKey)

	resultsIterator, err := stub.GetHistoryForKey(rowObjectKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the rowObject
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
		//as-is (as the Value itself a JSON rowObject)
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

	fmt.Printf("- getHistoryForObject returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}
