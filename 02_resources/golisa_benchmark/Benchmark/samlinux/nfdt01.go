/**
 * Roland Bole
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	// https://godoc.org/github.com/google/uuid
	guuid "github.com/google/uuid"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type DataItem struct {
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Owner       string    `json:"owner"`
	UpdatedAt   time.Time `json:"time"`
}

// ===========================
// Main
// ===========================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ===========================
// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// ===========================
// Invoke central ctrl of the chaincode
// ===========================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	// this is printed to the chaincode container
	fmt.Println("nfdt01 Invoke")

	// import function name and arguments
	function, args := stub.GetFunctionAndParameters()

	// ctrol the flow
	if function == "add" {
		return t.add(stub, args)
	} else if function == "queryById" {
		return t.queryById(stub, args)
	} else if function == "queryByOwner" {
		return t.queryByOwner(stub, args)
	} else if function == "queryAdHoc" {
		return t.queryAdHoc(stub, args)
	}

	// if no case match an error will be thrown
	return shim.Error("Invalid invoke function name. Expecting \"add\" \"queryById\" \"queryByOwner\"  \"queryAdHoc\" ")
}

// =====================================
// add a new DataItem to the blockchain
// =====================================
func (t *SimpleChaincode) add(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// representing an error
	var err error

	// we need two params a key(Asset) and a value
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// build our data object
	dataItemTyp := args[0]
	dataItemName := args[1]
	dataItemDescription := args[2]
	dataItemOwner := args[3]

	// parse the string time to golang time
	layout := "2006-01-02T15:04:05.000Z"
	str := args[4]
	dataItemUpdatedAt, err := time.Parse(layout, str)

	dataItem := &DataItem{dataItemTyp, dataItemName, dataItemDescription, dataItemOwner, dataItemUpdatedAt}
	uuid := guuid.New()
	key := uuid.String()

	fmt.Printf("add a new DataItem %s", dataItemName)

	dataItemJSONasBytes, err := json.Marshal(dataItem)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Write the state back to the ledger
	err = stub.PutState(key, dataItemJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

// ==============================================
// queryById a DataItem from the blockchain
// ==============================================
func (t *SimpleChaincode) queryById(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	// import the uuid and create the querystring
	uuid := args[0]
	queryString := fmt.Sprintf("{\"selector\":{\"_id\":\"%s\"}}", uuid)

	// do the query
	queryResults, err := getQueryResultForQueryString(stub, queryString)

	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// ==============================================
// queryByOwner a DataItem from the blockchain
// ==============================================
func (t *SimpleChaincode) queryByOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}
	// import the uuid and create the querystring
	owner := args[0]
	queryString := fmt.Sprintf("{\"selector\":{\"owner\":\"%s\"}, \"sort\": [{\"time\": \"desc\"}]}", owner)

	// do the query
	queryResults, err := getQueryResultForQueryString(stub, queryString)

	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// ===============================================
// Ad hoc rich query with idividual query string
// ===============================================
func (t *SimpleChaincode) queryAdHoc(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// import the uuid and create the querystring
	queryString := args[0]

	// do the query
	queryResults, err := getQueryResultForQueryString(stub, queryString)

	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryStringWithPagination executes the passed in query string with
// pagination info. Result set is built and returned as a byte array containing the JSON results.
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
