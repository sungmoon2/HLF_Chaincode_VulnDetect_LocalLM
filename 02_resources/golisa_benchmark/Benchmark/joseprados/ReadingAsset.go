package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("CLDChaincode")

//ReadingAsset - Chaincode for asset Reading
type ReadingAsset struct {
}

//Reading - Details of the asset type Reading
type Reading struct {
	VehicleID    string `json:"vehicleID"`
	ObjectType   string `json:"docType"`
	Reading      string `json:"reading"`
	CreationDate string `json:"creationDate"`
}

//ReadingIDIndex - Index on IDs for retrieval all Readings
type ReadingIDIndex struct {
	VehicleIDs []string `json:"vehicleIDs"`
}

func main() {
	err := shim.Start(new(ReadingAsset))
	if err != nil {
		fmt.Printf("Error starting ReadingAsset chaincode function main(): %s", err)
	} else {
		fmt.Printf("Starting ReadingAsset chaincode function main() executed successfully")
	}
}

//Init - The chaincode Init function: No  arguments, only initializes a ID array as Index for retrieval of all Readings
func (rdg *ReadingAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	var readingIDIndex ReadingIDIndex
	bytes, _ := json.Marshal(readingIDIndex)
	stub.PutState("readingIDIndex", bytes)
	return shim.Success(nil)
}

//Invoke - The chaincode Invoke function:
func (rdg *ReadingAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	logger.Debug("function: ", function)
	if function == "addNewReading" {
		return rdg.addNewReading(stub, args)
	} else if function == "updateReading" {
		return rdg.updateReading(stub, args)
	} else if function == "removeAllReadings" {
		return rdg.removeAllReadings(stub)
	} else if function == "readReading" {
		return rdg.readReading(stub, args[0])
	} else if function == "readAllReadings" {
		return rdg.readAllReadings(stub)
	}
	return shim.Error("Received unknown function invocation")
}

//Invoke Route: addNewReading
func (rdg *ReadingAsset) addNewReading(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	reading, err := getReadingFromArgs(args)
	if err != nil {
		return shim.Error("Reading Data is Corrupted")
	}
	reading.ObjectType = "Asset.Reading"
	record, err := stub.GetState(reading.VehicleID)
	if record != nil {
		return shim.Error("This Reading already exists: " + reading.VehicleID)
	}
	_, err = rdg.saveReading(stub, reading)
	if err != nil {
		return shim.Error(err.Error())
	}
	_, err = rdg.updateReadingIDIndex(stub, reading)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

//Invoke Route: updateReading
func (rdg *ReadingAsset) updateReading(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	var currReading Reading
	newReading, err := getReadingFromArgs(args)
	readingAsByteArray, err := rdg.retrieveReading(stub, newReading.VehicleID)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = json.Unmarshal(readingAsByteArray, &currReading)
	if err != nil {
		return shim.Error("updateReading: Error unmarshalling readingStruct array JSON")
	}
	currReadingVal, _ := strconv.ParseFloat(currReading.Reading, 64)
	newReadingVal, _ := strconv.ParseFloat(newReading.Reading, 64)
	if newReadingVal < currReadingVal {
		return shim.Error("updateReading: New Reading is less than Current Reading - cannot update")
	}
	currDate, err := time.Parse("01/02/2006", currReading.CreationDate)
	if err != nil {
		return shim.Error(err.Error())
	}
	newDate, err := time.Parse("01/02/2006", newReading.CreationDate)
	if err != nil {
		return shim.Error(err.Error())
	}
	if currDate.After(newDate) {
		return shim.Error("updateReading: New Date is earlier than Current Date - cannot update")
	}
	_, err = rdg.saveReading(stub, newReading)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

//Invoke Route: removeAllReadings
func (rdg *ReadingAsset) removeAllReadings(stub shim.ChaincodeStubInterface) peer.Response {
	var readingStructIDs ReadingIDIndex
	bytes, err := stub.GetState("readingIDIndex")
	if err != nil {
		return shim.Error("removeAllReadings: Error getting readingIDIndex array")
	}
	err = json.Unmarshal(bytes, &readingStructIDs)
	if err != nil {
		return shim.Error("removeAllReadings: Error unmarshalling readingIDIndex array JSON")
	}
	if len(readingStructIDs.VehicleIDs) == 0 {
		return shim.Error("removeAllReadings: No readings to remove")
	}
	for _, readingStructID := range readingStructIDs.VehicleIDs {
		_, err = rdg.deleteReading(stub, readingStructID)
		if err != nil {
			return shim.Error("Failed to remove Reading with ID: " + readingStructID)
		}
		_, err = rdg.deleteReadingIDIndex(stub, readingStructID)
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	rdg.initHolder(stub)
	return shim.Success(nil)
}

//Query Route: readReading
func (rdg *ReadingAsset) readReading(stub shim.ChaincodeStubInterface, readingID string) peer.Response {
	readingAsByteArray, err := rdg.retrieveReading(stub, readingID)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(readingAsByteArray)
}

//Query Route: readAllReadings
func (rdg *ReadingAsset) readAllReadings(stub shim.ChaincodeStubInterface) peer.Response {
	var readingIDs ReadingIDIndex
	bytes, err := stub.GetState("readingIDIndex")
	if err != nil {
		return shim.Error("readAllReadings: Error getting readingIDIndex array")
	}
	err = json.Unmarshal(bytes, &readingIDs)
	if err != nil {
		return shim.Error("readAllReadings: Error unmarshalling readingIDIndex array JSON")
	}
	result := "["

	var readingAsByteArray []byte

	for _, readingID := range readingIDs.VehicleIDs {
		readingAsByteArray, err = rdg.retrieveReading(stub, readingID)
		if err != nil {
			return shim.Error("Failed to retrieve reading with ID: " + readingID)
		}
		result += string(readingAsByteArray) + ","
	}
	if len(result) == 1 {
		result = "[]"
	} else {
		result = result[:len(result)-1] + "]"
	}
	return shim.Success([]byte(result))
}

//Helper: Save purchaser
func (rdg *ReadingAsset) saveReading(stub shim.ChaincodeStubInterface, reading Reading) (bool, error) {
	bytes, err := json.Marshal(reading)
	if err != nil {
		return false, errors.New("Error converting reading record JSON")
	}
	err = stub.PutState(reading.VehicleID, bytes)
	if err != nil {
		return false, errors.New("Error storing Reading record")
	}
	return true, nil
}

//Helper: Reading readingStruct //change template
func (rdg *ReadingAsset) deleteReading(stub shim.ChaincodeStubInterface, readingID string) (bool, error) {
	_, err := rdg.retrieveReading(stub, readingID)
	if err != nil {
		return false, errors.New("Reading with ID: " + readingID + " not found")
	}
	err = stub.DelState(readingID)
	if err != nil {
		return false, errors.New("Error deleting Reading record")
	}
	return true, nil
}

//Helper: Update reading Holder - updates Index
func (rdg *ReadingAsset) updateReadingIDIndex(stub shim.ChaincodeStubInterface, reading Reading) (bool, error) {
	var readingIDs ReadingIDIndex
	bytes, err := stub.GetState("readingIDIndex")
	if err != nil {
		return false, errors.New("updateReadingIDIndex: Error getting readingIDIndex array Index from state")
	}
	err = json.Unmarshal(bytes, &readingIDs)
	if err != nil {
		return false, errors.New("updateReadingIDIndex: Error unmarshalling readingIDIndex array JSON")
	}
	readingIDs.VehicleIDs = append(readingIDs.VehicleIDs, reading.VehicleID)
	bytes, err = json.Marshal(readingIDs)
	if err != nil {
		return false, errors.New("updateReadingIDIndex: Error marshalling new reading ID")
	}
	err = stub.PutState("readingIDIndex", bytes)
	if err != nil {
		return false, errors.New("updateReadingIDIndex: Error storing new reading ID in readingIDIndex (Index)")
	}
	return true, nil
}

//Helper: delete ID from readingStruct Holder
func (rdg *ReadingAsset) deleteReadingIDIndex(stub shim.ChaincodeStubInterface, readingID string) (bool, error) {
	var readingStructIDs ReadingIDIndex
	bytes, err := stub.GetState("readingIDIndex")
	if err != nil {
		return false, errors.New("deleteReadingIDIndex: Error getting readingIDIndex array Index from state")
	}
	err = json.Unmarshal(bytes, &readingStructIDs)
	if err != nil {
		return false, errors.New("deleteReadingIDIndex: Error unmarshalling readingIDIndex array JSON")
	}
	readingStructIDs.VehicleIDs, err = deleteKeyFromStringArray(readingStructIDs.VehicleIDs, readingID)
	if err != nil {
		return false, errors.New(err.Error())
	}
	bytes, err = json.Marshal(readingStructIDs)
	if err != nil {
		return false, errors.New("deleteReadingIDIndex: Error marshalling new readingStruct ID")
	}
	err = stub.PutState("readingIDIndex", bytes)
	if err != nil {
		return false, errors.New("deleteReadingIDIndex: Error storing new readingStruct ID in readingIDIndex (Index)")
	}
	return true, nil
}

//Helper: Initialize truck ID Holder //change template
func (rdg *ReadingAsset) initHolder(stub shim.ChaincodeStubInterface) bool {
	var readingIDIndex ReadingIDIndex
	bytes, _ := json.Marshal(readingIDIndex)
	stub.DelState("readingIDIndex")
	stub.PutState("readingIDIndex", bytes)
	return true
}

//deleteKeyFromArray
func deleteKeyFromStringArray(array []string, key string) (newArray []string, err error) {
	for _, entry := range array {
		if entry != key {
			newArray = append(newArray, entry)
		}
	}
	if len(newArray) == len(array) {
		return newArray, errors.New("Specified Key: " + key + " not found in Array")
	}
	return newArray, nil
}

//Helper: Retrieve purchaser
func (rdg *ReadingAsset) retrieveReading(stub shim.ChaincodeStubInterface, readingID string) ([]byte, error) {
	var reading Reading
	var readingAsByteArray []byte
	bytes, err := stub.GetState(readingID)
	if err != nil {
		return readingAsByteArray, errors.New("retrieveReading: Error retrieving reading with ID: " + readingID)
	}
	err = json.Unmarshal(bytes, &reading)
	if err != nil {
		return readingAsByteArray, errors.New("retrieveReading: Corrupt reading record " + string(bytes))
	}
	readingAsByteArray, err = json.Marshal(reading)
	if err != nil {
		return readingAsByteArray, errors.New("readReading: Invalid reading Object - Not a  valid JSON")
	}
	return readingAsByteArray, nil
}

//getReadingFromArgs - construct a reading structure from string array of arguments
func getReadingFromArgs(args []string) (reading Reading, err error) {

	if strings.Contains(args[0], "\"vehicleID\"") == false ||
		strings.Contains(args[0], "\"docType\"") == false ||
		strings.Contains(args[0], "\"reading\"") == false ||
		strings.Contains(args[0], "\"creationDate\"") == false {
		return reading, errors.New("Unknown field: Input JSON does not comply to schema")
	}

	err = json.Unmarshal([]byte(args[0]), &reading)
	if err != nil {
		return reading, err
	}
	return reading, nil
}
