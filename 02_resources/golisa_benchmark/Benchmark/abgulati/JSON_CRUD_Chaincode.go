/*
*
*	author @abgulati
*
*/


package main

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type JCRUDChaincode struct {
}

//metdata for a car
type Car struct {
	Chassis_No	string 	`json:"chassis_no"`
	Name 		string  `json:"name"`
	Type 		string  `json:"type"`
	Engine_No 	string 	`json:"engine_no"`
	Year		string 	`json:"year"`
	Deleted 	bool	`json:"deleted"`
}

//used for readCar() ops
type ID struct {
	Engine_No string `json:"engine_no"`
}

//===========
//Main Method
//===========
func main() {
	err := shim.Start(new(JCRUDChaincode))
	if err != nil {
		fmt.Printf("Error starting test chaincode: %s: ", err)
	}
}

//============================
//Init initializes chaincode
//============================
func (t *JCRUDChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

//========================================
//Invoke - Our entry point for Invocations
//========================================
func (t *JCRUDChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("Invoking " + function)

	//Handle different functions
	if function == "initCar" { //create a new Car record
		return t.initCar(stub, args)
	} else if function == "readRecord" { //read a Car record
		return t.readRecord(stub, args)
	} else if function == "updateCar" {	//update a Car record
		return t.updateCar(stub, args)
	} else if function == "hardDeleteCar" { //hardDelete a Car record, future reads return a "not found" error
		return t.hardDeleteCar(stub, args)
	} else if function == "softDeleteCar" { //softDelete a Car record, future reads permitted but updates restricted
		return t.softDeleteCar(stub, args)
	}

	fmt.Println("invoke did not find function: " + function)
	return shim.Error("Received unknown function invocation")
}

//=============================================================
//initCar - create a new car record, store into chaincode state
//=============================================================
func (t *JCRUDChaincode) initCar(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	receivedCarJSON := args[0]

	var createdCar Car

	err := json.Unmarshal([]byte(receivedCarJSON), &createdCar)
	if err != nil {
		return shim.Error("Error with input")
	}

	Car_JSONasBytes, err := json.Marshal(createdCar)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(createdCar.Member_ID, Car_JSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end init Car record")
	return shim.Success(nil)	
}


//===============================================
//readRecord - read a record from chaincode state
//===============================================
func (t *JCRUDChaincode) readRecord(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	receivedJSON := args[0]

	var convertedData ID

	err := json.Unmarshal([]byte(receivedJSON), &convertedData)
	if err != nil {
		return shim.Error("Error unmarshalling JSON")
	}

	valAsJSON, err := stub.GetState(convertedData.Engine_No) //get the record from chaincode state
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + convertedData.Engine_No + "\"}"
		return shim.Error(jsonResp)
	} else if valAsJSON == nil {
		jsonResp := "{\"Error\":\"Record does not exist: " + convertedData.Engine_No + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsJSON)
	
}

//================================================================================================================
//updateCar - update a Car record by pulling up record and creating a new block with the original + updated fields
//================================================================================================================
func (t *JCRUDChaincode) updateCar(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	receivedJSON := args[0]

	var convertedCar Car
	err := json.Unmarshal([]byte(receivedJSON), &convertedCar)
	if err != nil {
		return shim.Error("Error unmarshalling JSON")
	}

	valAsBytes, err := stub.GetState(convertedCar.Engine_No) //get record from blockchain
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + convertedCar.Engine_No + "\"}"
		return shim.Error(jsonResp)
	} else if valAsBytes == nil {
		jsonResp := "{\"Error\":\"Record does not exist: " + convertedCar.Engine_No + "\"}"
		return shim.Error(jsonResp)
	}

	var carJSON Car
	err = json.Unmarshal(valAsBytes, &carJSON)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to decode JSON of: " + convertedCar.Engine_No + "\"}"
		return shim.Error(jsonResp)
	}

	if carJSON.Deleted == true {
		jsonResp := "{\"Error\":\"Record for " + convertedCar.Engine_No + " has been marked as deleted. Cannot update.\"}"
		return shim.Error(jsonResp)
	}

	current_record_field_list := reflect.ValueOf(&carJSON).Elem()	//Car record pulled from the blockchain is converted to a reflect.Value type and can now be modified	
	record_fields_type := current_record_field_list.Type()			//This is of type main.Car, meaning it pulls up the type of the underlying variable from the Car struct

	fields_to_update := reflect.ValueOf(&convertedCar).Elem()		//repeat above two steps for the update data received by this function
	update_fields_type := fields_to_update.Type()

	for i := 0; i < current_record_field_list.NumField()-1; i++ {	
		for j := 0; j < fields_to_update.NumField()-1; j++ {

			new_field_value := fields_to_update.Field(j) 		//obtain a new value from received JSON
			reflect_value := reflect.ValueOf(new_field_value.Interface())	//returns value from new_field_value as an interface{}. An empty interface such as this is of type interface{} and can accept any value.
																			//reflect.ValueOf then evaluates its actual underlying value and its type(such as string, int etc.)

			if !(reflect.DeepEqual(reflect_value.Interface(), reflect.Zero(reflect_value.Type()).Interface())) && record_fields_type.Field(i).Name == update_fields_type.Field(j).Name{
				// DeepEqual checks if two values are deeply equal, and what constitutes deeply equal differs from data type to data type. We use it here as we may not always know what type is being evaluated here.
				// That's also why .Interface() is used, as it returns a value of type interface{}, which can accept any value. DeepEqual then evaluates if they're equal.
				// reflect.Zero() is then used to obtain the equivalent zero value for the given data type. Eg: 0 for type int, empty string "" for type string etc
				// So the actual value from reflect_value is compared with  its corresponding zero value to ensure that field is not empty, i.e., there is no underlying update, and then field names are evaluated.
				// so this if conditional statement is evaluating whether a given field is non-empty and checks for matching field names
				original_field_value := current_record_field_list.Field(i)
				original_field_value.Set(new_field_value)
			}
		}
	}

	jsonData, err := json.Marshal(carJSON)
	err = stub.PutState(convertedCar.Engine_No, jsonData)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

//======================================================================================================================
//hardDeleteCar - remove a Car record from the ledger completely, future read will yield a "record does not exist error"
//======================================================================================================================
func (t *JCRUDChaincode) hardDeleteCar(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	receivedJSON := args[0]

	var convertedCar Car
	err := json.Unmarshal([]byte(receivedJSON), &convertedCar)
	if err != nil {
		return shim.Error("Error unmarshalling JSON")
	}

	valAsBytes, err := stub.GetState(convertedCar.Engine_No)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get record for " + convertedCar.Engine_No + "\"}"
		return shim.Error(jsonResp)
	} else if valAsBytes == nil {
		jsonResp := "{\"Error\":\"Car Record does not exist for " + convertedCar.Engine_No + "\"}"
		return shim.Error(jsonResp)
	}

	err = stub.DelState(convertedCar.Engine_No) //remove the Car record
	if err != nil {
		return shim.Error("Failed to delete the record: " + err.Error())		
	}

	return shim.Success(nil)
}

//======================================================================================================================================
//softDeleteCar - set the boolean variable 'Deleted' to 'true'; this will allow future reads on the Car's record but prevent any updates
//======================================================================================================================================
func (t *JCRUDChaincode) softDeleteCar(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	receivedJSON := args[0]

	var convertedCar Car
	err := json.Unmarshal([]byte(receivedJSON), &convertedCar)
	if err != nil {
		return shim.Error("Error unmarshalling JSON")
	}

	valAsBytes, err := stub.GetState(convertedCar.Engine_No)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get record for " + convertedCar.Engine_No + "\"}"
		return shim.Error(jsonResp)
	} else if valAsBytes == nil {
		jsonResp := "{\"Error\":\"Record does not exist for " + convertedCar.Engine_No + "\"}"
		return shim.Error(jsonResp)
	}

	var carJSON Car 
	err = json.Unmarshal(valAsBytes, &carJSON)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to decode JSON of: " + convertedCar.Engine_No + "\"}"
		return shim.Error(jsonResp)
	}

	if carJSON.Deleted == true {
		jsonResp := "{\"Error\":\"Record for " + convertedCar.Engine_No + " is already marked as deleted.\"}"
		return shim.Error(jsonResp)
	} else if carJSON.Deleted == false {
		carJSON.Deleted = true
		jsonData, err := json.Marshal(carJSON)
		err = stub.PutState(convertedCar.Engine_No, jsonData)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	return shim.Success(nil)
}