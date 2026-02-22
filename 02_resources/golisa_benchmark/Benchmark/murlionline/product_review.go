package main

//=================================================================================================
//========================================================================================== IMPORT
import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"


	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

//=================================================================================================
//============================================================================= BLOCKCHAIN DOCUMENT
// Doc writes string to the blockchain (as JSON object) for a specific key
type Doc struct {
	Text string `json:"text"`
	Review string `json:"review"`
	Name string `json:"name"`
	Location string `json:"location"`
	Rating string `json:"rating"`
}

func (doc *Doc) FromJson(input []byte) *Doc {
	json.Unmarshal(input, doc)
	return doc
}

func (doc *Doc) ToJson() []byte {
	jsonDoc, _ := json.Marshal(doc)
	return jsonDoc
}

//=================================================================================================
//================================================================================= RETURN HANDLING
// Return handling: for return, we either return "shim.Success (payload []byte) with HttpRetCode=200"
// or "shim.Error(doc string) with HttpRetCode=500". However, we want to set our own status codes to
// map into HTTP return codes. A few utility functions:

// Success with a payload
func Success(rc int32, doc string, payload []byte) peer.Response {
	if len(doc) > 1048576 {
		return Error(500, "Maximum return payload length of 1MB exceeded!")
	}
	return peer.Response{
		Status:  rc,
		Message: doc,
		Payload: payload,
	}
}

// Error with an error message
func Error(rc int32, doc string) peer.Response {
	logger.Errorf("Error %d = %s", rc, doc)
	return peer.Response{
		Status:  rc,
		Message: doc,
	}
}



//=================================================================================================
//============================================================================================ MAIN
// Main function starts up the chaincode in the container during instantiate
//
var logger = shim.NewLogger("chaincode")

type ProductReview struct {
	// use this structure for information that is held (in-memory) within chaincode
	// instance and available over all chaincode calls
}

func main() {
	if err := shim.Start(new(ProductReview)); err != nil {
		fmt.Printf("Main: Error starting chaincode: %s", err)
	}
}

//=================================================================================================
//============================================================================================ INIT
// Init is called during Instantiate transaction after the chaincode container
// has been established for the first time, allowing the chaincode to
// initialize its internal data. Note that chaincode upgrade also calls this
// function to reset or to migrate data, so be careful to avoid a scenario
// where you inadvertently clobber your ledger's data!
//
func (cc *ProductReview) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Validate supplied init parameters, in this case zero arguments!
	if _, args := stub.GetFunctionAndParameters(); len(args) > 0 {
		return Error(http.StatusBadRequest, "Init: Incorrect number of arguments; no arguments were expected.")
	}
	return Success(http.StatusOK, "OK", nil)
}

//=================================================================================================
//========================================================================================== INVOKE
// Invoke is called to update or query the ledger in a proposal transaction.
// Updated state variables are not committed to the ledger until the
// transaction is committed.
//
func (cc *ProductReview) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	// Increase logging level for this example (ERROR level recommended for productive code)
	logger.SetLevel(shim.LogDebug)

	// Which function is been called?
	function, args := stub.GetFunctionAndParameters()

	// Route call to the correct function
	switch function {
	case "read":
		return cc.read(stub, args)
	case "create":
		return cc.create(stub, args)
	case "search":
		return cc.search(stub, args)
	default:
		logger.Warningf("Invoke('%s') invalid!", function)
		return Error(http.StatusNotImplemented, "Invalid method! Valid methods are 'create|update|delete|exist|read|history|search'!")
	}
}


//=================================================================================================
//============================================================================================ READ
// Read text by ID
//
func (cc *ProductReview) read(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	// Validate and extract parameters
	//if rc := Validate("read", args /*args[0]=id*/, "%s", 1, 64); rc.Status > 0 {
	//	return rc
	//}
	id := strings.ToLower(args[0])

	// Read the value for the ID
	if value, err := stub.GetState(id); err != nil || value == nil {
		return Error(http.StatusNotFound, "Not Found")
	} else {
		return Success(http.StatusOK, "OK", value)
	}
}

//=================================================================================================
//========================================================================================== CREATE
// Creates a text by ID
//
func (cc *ProductReview) create(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	// Validate and extract parameters
	//if rc := Validate("create", args /*args[0]=id*/, "%s", 1, 64 /*args[1]=text*/, "%s", 1, 255); rc.Status > 0 {
	//	return rc
	//}
	id := strings.ToLower(args[0])
	doc := &Doc{Text: args[1],Review: args[2],Name: args[3],Location: args[4],Rating: args[5]}

	// Validate that this ID does not yet exist. If the key does not exist (nil, nil) is returned.
	if value, err := stub.GetState(id); !(err == nil && value == nil) {
		return Error(http.StatusConflict, "Text Exists")
	}

	// Write the message
	if err := stub.PutState(id, doc.ToJson()); err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}

	return Success(http.StatusCreated, "Text Created", nil)
}



//=================================================================================================
//========================================================================================== SEARCH
// Search for all matching IDs, given a (regex) value expression and return both the IDs and text.
// For example: '^H.llo' will match any string starting with 'Hello' or 'Hallo'.
//
func (cc *ProductReview) search(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	// Validate and extract parameters
	//if rc := Validate("search", args /*args[0]=searchString*/, "%s", 1, 64); rc.Status > 0 {
	//	return rc
	//}
	searchString := strings.Replace(args[0], "\"", ".", -1) // protect against SQL injection

	// stub.GetQueryResult takes a verbatim CouchDB (assuming this is used DB). See CouchDB documentation:
	//     http://docs.couchdb.org/en/2.0.0/api/database/find.html
	// For example:
	//	{
	//		"selector": {
	//			"value": {"$regex": %s"}
	//		},
	//		"fields": ["ID","value"],
	//		"limit":  99
	//	}
	queryString := fmt.Sprintf("{\"selector\": {\"text\": {\"$regex\": \"%s\"}}, \"fields\": [\"text\",\"review\",\"name\",\"location\",\"rating\"], \"limit\":99}", strings.Replace(searchString, "\"", ".", -1))
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}
	defer resultsIterator.Close()

	// Write return buffer
	var buffer bytes.Buffer
	buffer.WriteString("{ \"values\": [")
	for resultsIterator.HasNext() {
		it, _ := resultsIterator.Next()
		if buffer.Len() > 15 {
			buffer.WriteString(",")
		}
		var doc Doc
		buffer.WriteString("{\"id\":\"")
		buffer.WriteString(it.Key)
		buffer.WriteString("\", \"review\":\"")
		buffer.WriteString(doc.FromJson(it.Value).Review)
		buffer.WriteString("\", \"product\":\"")
		buffer.WriteString(doc.FromJson(it.Value).Text)
		buffer.WriteString("\", \"name\":\"")
		buffer.WriteString(doc.FromJson(it.Value).Name)		
		buffer.WriteString("\", \"location\":\"")
		buffer.WriteString(doc.FromJson(it.Value).Location)
		buffer.WriteString("\", \"rating\":\"")
		buffer.WriteString(doc.FromJson(it.Value).Rating)
		buffer.WriteString("\"}")
	}
	buffer.WriteString("]}")

	return Success(http.StatusOK, "OK", buffer.Bytes())
}
