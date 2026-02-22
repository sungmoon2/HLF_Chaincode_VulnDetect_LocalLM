// This chaincode calls the token/v5
// This is to JUST demonstrate the invoke mechanism
// This cc will act as a proxy
package main

import (
	"fmt"
	"strconv"

	// The shim package
	"github.com/hyperledger/fabric/core/chaincode/shim"
	// peer.Response is in the peer package
	"github.com/hyperledger/fabric/protos/peer"
)

// CallerChaincode Represents our chaincode object
type CallerChaincode struct {
}

// Channel Name
const Channel = "airlinechannel"

// Chaincode to be invoked
const TargetChaincode = "token_rit"

// Init func will do nothing
func (token *CallerChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("Init executed.")
	stub.PutState("Saurabh", []byte("4000"))

	// Return success
	return shim.Success([]byte("Init Done."))
}

// Invoke method
func (token *CallerChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	funcName, _ := stub.GetFunctionAndParameters()
	fmt.Println("Function=", funcName)

	if funcName == "setOnCaller" {
		// Setup the args
		args := make([][]byte, 1)
		args[0] = []byte("set")

		// Sets the value of MyToken in token chaincode (V5)
		response := stub.InvokeChaincode(TargetChaincode, args, Channel)

		// Print on console
		fmt.Println("Received SET response from 'token' : " + response.String())

		return response

	} else if funcName == "getOnCaller" {
		// Setup the args
		args := make([][]byte, 1)
		args[0] = []byte("get")

		// Gets the value of MyToken in token chaincode (V5)
		response := stub.InvokeChaincode(TargetChaincode, args, Channel)

		// Print on console
		fmt.Println("Received GET response from 'token' : " + response.String())

		return response
	} else if funcName == "set" {
		// Sets the value
		return SetToken(stub)

	} else if funcName == "get" {

		// Gets the value
		return GetToken(stub)

	} else if funcName == "del" {

		// Gets the value
		return DeleteToken(stub)

	}
	// This is not good
	return shim.Error(("Bad Function Name from caller = " + funcName + "!!!"))
}
func SetToken(stub shim.ChaincodeStubInterface) peer.Response {

	// Get the current value
	value, err := stub.GetState("Saurabh")

	// If there is error in retrieve send back an error response
	if err != nil {
		return shim.Error(err.Error())
	}

	// Convert value to integer
	intValue, err := strconv.Atoi(string(value))

	// If there is an error in conversion - return false
	if err != nil {
		// May also return sh.Error
		return shim.Success([]byte("false"))
	}

	// Increment the value by 10
	intValue += 10

	// Execute PutState - overwrites the current value
	stub.PutState("Saurabh", []byte(strconv.Itoa(intValue)))

	return shim.Success([]byte("true"))
}

// GetToken reads the value of the token from the database
// V5
// Reurns the value or -1 in case MyToken doesn't exist
func GetToken(stub shim.ChaincodeStubInterface) peer.Response {
	// Holds a string for the response
	var Saurabh string

	// Local variables for value & error
	var value []byte
	var err error

	if value, err = stub.GetState("Saurabh"); err != nil {

		fmt.Println("Get Failed!!! ", err.Error())

		return shim.Error(("Get Failed!! " + err.Error() + "!!!"))

	}

	// nil indicates non existent key
	if value == nil {
		// Return value -1 is to indicate to caller that MyToken
		// Does NOT exist in state data
		fmt.Println("No state found")

	} else {

		Saurabh = "Saurabh=" + string(value)

	}

	return shim.Success([]byte(Saurabh))
}

func DeleteToken(stub shim.ChaincodeStubInterface) peer.Response {

	// Get the current value
	value, _ := stub.GetState("Saurabh")

	// If no value, then send false
	if value == nil {
		return shim.Success([]byte("false"))
	}
	err := stub.DelState("Saurabh")
	if err != nil {
		fmt.Println("Delete Failed!!", err.Error())
		return shim.Error(("Delete Failed!! " + err.Error() + "!!!"))
	}
	return shim.Success([]byte("true"))

}

// Chaincode registers with the Shim on startup
func main() {
	fmt.Printf("Started Chaincode. caller/v6\n")
	err := shim.Start(new(CallerChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
