/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

//WARNING - this chaincode's ID is hard-coded in chaincode_example04 to illustrate one way of
//calling chaincode from a chaincode. If this example is modified, chaincode_example04.go has
//to be modified as well with the new ID of chaincode_example02.
//chaincode_example05 show's how chaincode ID can be passed in as a parameter instead of
//hard-coding.

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"

	model "github.com/ianco/jsonapi/model"
)

// ConfigCC example simple Chaincode implementation
type ConfigCC struct {
}

func (t *ConfigCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Init Config CC")
	_, args := stub.GetFunctionAndParameters()
	//var P1, P2 ParticipantData    // Participants
	var ConfigVal model.ConfigData      // Configuration
	var ConfigStr string
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	// Initialize the chaincode
	ConfigVal, err = model.Json2Config(args[0])
	if err != nil {
		return shim.Error(fmt.Sprintf("Args[0] is not a valid ConfigData [%s]", args[0]))
	}
/*
	P1, err = Json2Participant(args[1])
	if err != nil {
		return shim.Error("Args[1] is not a valid participant [%s]", args[1])
	}
	P2, err = Json2Participant(args[2])
	if err != nil {
		return shim.Error("Args[2] is not a valid participant [%s]", args[2])
	}
*/
	//fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state to the ledger
	ConfigStr, err = model.Config2Json(ConfigVal)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error converting ConfigData back to Json [%s]", args[0]))
	}
	err = stub.PutState("Configuration", []byte(ConfigStr))
	if err != nil {
		return shim.Error(err.Error())
	}
/*
	err = stub.PutState("Participant." + P1.Id, []byte(args[1]))
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState("Participant." + P2.Id, []byte(args[2]))
	if err != nil {
		return shim.Error(err.Error())
	}
*/
	return shim.Success(nil)
}

func (t *ConfigCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Invoke Config CC")
	function, args := stub.GetFunctionAndParameters()
	if function == "update_config" {
		return t.updateConfig(stub, args)
	} else if function == "query_config" {
		return t.queryConfig(stub, args)
/*
	} else if function == "update_participant" {
		return t.updateParticipant(stub, args)
	} else if function == "delete_participant" {
		return t.deleteParticipant(stub, args)
	} else if function == "query_participant" {
		return t.queryParticipant(stub, args)
*/
	}

	//return shim.Error("Invalid invoke function name. Expecting \"invoke_config\" \"query_config\" \"invoke_participant\" \"delete_participant\" \"query_participant\"")
	return shim.Error(fmt.Sprintf("Invalid invoke function name [%s]. Expecting \"invoke_config\" \"query_config\"", function))
}

func (t *ConfigCC) updateConfig(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var ConfigVal, Aval model.ConfigData      // Configuration
	var Avalstr string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	ConfigVal, err = model.Json2Config(args[0])
	if err != nil {
		return shim.Error(fmt.Sprintf("Args[0] is not a valid ConfigData [%s]", args[0]))
	}

	// Get the state from the ledger
	Avalbytes, err := stub.GetState("Configuration")
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, err = model.Json2Config(string(Avalbytes))
	if err != nil {
		return shim.Error(fmt.Sprintf("Error: not a valid ConfigData [%s]", string(Avalbytes)))
	}

	// Perform the execution
	Aval.DifficultyRating = ConfigVal.DifficultyRating
	//fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	Avalstr, err = model.Config2Json(Aval)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState("Configuration", []byte(Avalstr))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (t *ConfigCC) queryConfig(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string // Entities
	var Aval model.ConfigData
	var err error

	if len(args) != 0 {
		return shim.Error(fmt.Sprintf("Incorrect number of arguments [%d]. Expecting 0.", len(args)))
	}

	// Get the state from the ledger
	Avalbytes, err := stub.GetState("Configuration")
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for Configuration\"}"
		return shim.Error(jsonResp)
	}
	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for Configuration\"}"
		return shim.Error(jsonResp)
	}

	Aval, err = model.Json2Config(string(Avalbytes))
	if err != nil {
		return shim.Error(fmt.Sprintf("Error: not a valid ConfigData [%s]", string(Avalbytes)))
	}

	jsonResp, err = model.Config2Json(Aval)
	fmt.Printf("Query Response: [%s]\n", jsonResp)
	return shim.Success(Avalbytes)
}
/*
// Deletes an entity from state
func (t *ConfigCC) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
*/
func main() {
	err := shim.Start(new(ConfigCC))
	if err != nil {
		fmt.Printf("Error starting Config ChainCode: %s", err)
	}
}
