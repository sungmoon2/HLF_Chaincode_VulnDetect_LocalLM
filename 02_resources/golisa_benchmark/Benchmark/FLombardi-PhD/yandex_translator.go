/*

 Yandex Translator Italian to English chaincode.
 
 Author: Federico Lombardi
 
 Get an API key at https://tech.yandex.com/translate/
 and set it at row 84 of this chaincode

*/

package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"strings"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)


// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ex02 Init")
	_, args := stub.GetFunctionAndParameters()
	var wordIta, wordEng string    // Source word in ITA and its ENG translation 
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// Initialize the chaincode
	wordIta = args[0]
	
	wordEng = args[1]
	
	fmt.Printf("word ita = %s, word eng = %s\n", wordIta, wordEng)

	// Write the state to the ledger
	err = stub.PutState(wordIta, []byte(wordEng))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Invoke function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Translator Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" {
		// transalte wordIta to wordEng
		return t.invoke(stub, args)
	} else if function == "delete" {
		// Deletes an entity from its state
		return t.delete(stub, args)
	} else if function == "query" {
		// the old "Query" is now implemtned in invoke
		return t.query(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\"")
}

// Transaltion of wordIta to wordEng
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
			
	var wordIta, wordEng, url string  // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	wordIta = args[0]

	var api_key = "<put your yandex api key here>";
	
	url = "https://translate.yandex.net/api/v1.5/tr/translate?key="+ api_key +"&text=" + wordIta + "&lang=it-en&format=plain"
	resp, err := http.Get(url)
	if err != nil {
		// handle error
		return shim.Error("Error while invokinng external get url")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	subBody := body[82:]
	
	wordEngTag := string(subBody)
	wordEng = strings.Replace(wordEngTag, "</text></Translation>", "", -1)
	
	fmt.Println("length body: ", len(body))
	fmt.Println("word eng: ", wordEng)

	// Write the state back to the ledger
	err = stub.PutState(wordIta, []byte(wordEng))
	if err != nil {
		return shim.Error(err.Error())
	}
	
	fmt.Printf("Invoke Response: ", wordEng)

	return shim.Success(nil)
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var wordIta string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1 argument.")
	}

	wordIta = args[0]

	// Get the state from the ledger
	wordEngBytes, err := stub.GetState(wordIta)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + wordIta + "\"}"
		return shim.Error(jsonResp)
	}

	if wordEngBytes == nil {
		jsonResp := "{\"Error\":\"No match found for " + wordIta + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"WordIta\":\"" + wordIta + "\",\"WordEng\":\"" + string(wordEngBytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)

	return shim.Success(wordEngBytes)
}

// main
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
