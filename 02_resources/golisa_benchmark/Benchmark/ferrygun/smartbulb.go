// DISCLAIMER:
// THIS SAMPLE CODE MAY BE USED SOLELY AS PART OF THE TEST AND EVALUATION OF THE SAP CLOUD PLATFORM
// BLOCKCHAIN SERVICE (THE “SERVICE”) AND IN ACCORDANCE WITH THE TERMS OF THE AGREEMENT FOR THE SERVICE.
// THIS SAMPLE CODE PROVIDED “AS IS”, WITHOUT ANY WARRANTY, ESCROW, TRAINING, MAINTENANCE, OR SERVICE
// OBLIGATIONS WHATSOEVER ON THE PART OF SAP.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

//=================================================================================================
//================================================================================= RETURN HANDLING

// Success HTTP 2xx with a payload
func Success(rc int32, doc string, payload []byte) peer.Response {
	return peer.Response{
		Status:  rc,
		Message: doc,
		Payload: payload,
	}
}

// Error HTTP 4xx or 5xx with an error message
func Error(rc int32, doc string) peer.Response {
	logger.Errorf("Error %d = %s", rc, doc)
	return peer.Response{
		Status:  rc,
		Message: doc,
	}
}

//=================================================================================================
//====================================================================================== VALIDATION
// Validation: all arguments for a function call is passed as a string array args[]. Validate that
// the number, type and length of the arguments are correct.
//
// The Validate function is called as follow:
// 		Validate("chaincode function name", args, T[0], Ta[0], Tb[0], T[1], Ta[1], Tb[1], ...)
// The parameter groups T[i], Ta[i], Tb[i] are used to validate each parameter in sequence in args.
// T[i]describes the type/format for the parameter i and Ta[i] and Tb[i] are type dependent.
//
//		T[i]	Ta[i]		Tb[i]				Comment
//		"%s"	minLength	maxLength			String with min/max length
//		"%json"	minLength	maxLength			JSON format with min/max length
//
func Validate(funcName string, args []string, desc ...interface{}) peer.Response {

	logger.Debugf("Function: %s(%s)", funcName, strings.TrimSpace(strings.Join(args, ",")))

	nrArgs := len(desc) / 3

	if len(args) != nrArgs {
		return Error(http.StatusBadRequest, "Parameter Mismatch")
	}

	for i := 0; i < nrArgs; i++ {
		switch desc[i*3] {

		case "%json":
			var jsonData map[string]interface{}
			if jsonErr := json.Unmarshal([]byte(args[i]), &jsonData); jsonErr != nil {
				return Error(http.StatusBadRequest, "JSON Payload Not Valid")
			}
			fallthrough

		case "%s":
			var minLen = desc[i*3+1].(int)
			var maxLen = desc[i*3+2].(int)
			if len(args[i]) < minLen || len(args[i]) > maxLen {
				return Error(http.StatusBadRequest, "Parameter Length Error")
			}
		}
	}

	return Success(0, "OK", nil)
}

//=================================================================================================
//======================================================================================= MAIN/INIT
var logger = shim.NewLogger("chaincode")

type SmartBulb struct {
}

func main() {
	if err := shim.Start(new(SmartBulb)); err != nil {
		fmt.Printf("Main: Error starting chaincode: %s", err)
	}
	logger.SetLevel(shim.LogDebug)
}

// Init is called during Instantiate transaction.
func (cc *SmartBulb) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return Success(http.StatusNoContent, "OK", nil)
}

// Invoke is called to update or query the ledger in a proposal transaction.
func (cc *SmartBulb) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	function, args := stub.GetFunctionAndParameters()

	switch function {
	case "read":
		return read(stub, args)
	case "write":
		return write(stub, args)
	default:
		logger.Warningf("Invoke('%s') invalid!", function)
		return Error(http.StatusNotImplemented, "Invalid method! Valid methods are 'read|write'!")
	}
}

// Read text by ID
func read(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return Error(http.StatusBadRequest, "Parameter Mismatch")
	}
	id := strings.ToLower(args[0])

	if value, err := stub.GetState(id); err == nil && value != nil {
		return Success(http.StatusOK, "OK", value)
	}

	return Error(http.StatusNotFound, "Not Found")
}

// Write text by ID
func write(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if rc := Validate("create", args /*args[0]=id*/, "%s", 1, 64 /*args[1]=state*/, "%json", 2, 4096); rc.Status > 0 {
		return rc
	}
	
	if err := stub.SetEvent("event-name", []byte(args[1])); err != nil {
		return shim.Error(err.Error())
	} 

	if err := stub.PutState(args[0], []byte(args[1])); err == nil {
		return Success(http.StatusCreated, "Created", nil)
	} else {
		return Error(http.StatusInternalServerError, err.Error())
	}
}
