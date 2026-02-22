package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	cc "github.com/hyperledger/fabric/protos/peer"

)

type camerarecord struct{

}

type Cameras struct{
	Filename string `json:"file_name"`
	CameraId   string `json:"camera_id"`
	IncidentType      string `json:"Incident_type"`
	Latitude string `json:"latitude"`
	Longitude    string `json:"longitude"`
	SavingTimestamp string `json:"saving_timestamp"`
	Timestamp    string `json:"timestamp"`

}

//Init Function
func (c *camerarecord) Init(stub shim.ChaincodeStubInterface) cc.Response {

	_, args := stub.GetFunctionAndParameters()
	var cameras = Cameras{
		Filename: args[0],
		CameraId:   args[1],
		IncidentType:      args[2],
		Latitude: args[3],
		Longitude:    args[4],
		SavingTimestamp: args[5],
		Timestamp:    args[6]}

		cameraAsBytes, _ := json.Marshal(cameras)

		var uniqueID = args[1]
	
		err := stub.PutState(uniqueID, cameraAsBytes)
	
		if err != nil {
			fmt.Println("Error in Init")
		}
	
		return shim.Success([]byte("Chaincode Successfully initialized"))
}

//CreateCar ... this function is used to create doctors
func CreateCamera(stub shim.ChaincodeStubInterface, args []string) cc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}
	var cameras = Cameras{
		Filename: args[0],
		CameraId:   args[1],
		IncidentType:      args[2],
		Latitude: args[3],
		Longitude:    args[4],
		SavingTimestamp: args[5],
		Timestamp:    args[6]}

	cameraAsBytes, _ := json.Marshal(cameras)

	var uniqueID = args[1]

	err := stub.PutState(uniqueID, cameraAsBytes)

	if err != nil {
		fmt.Println("Erro in create camera")
	}

	return shim.Success(nil)
}


func (c *camerarecord) queryCamera(stub shim.ChaincodeStubInterface, args []string) cc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	carAsBytes, _ := stub.GetState(args[0])
	return shim.Success(carAsBytes)
}


//Invoke function
func (c *camerarecord) Invoke(stub shim.ChaincodeStubInterface) cc.Response {
	fun, args := stub.GetFunctionAndParameters()
	if fun == "CreateCamera" {
		fmt.Println("Error occured ==> ")
		//logger.Info("########### create docs ###########")
		return CreateCamera(stub, args)
	}  else if fun == "query" {
		return c.queryCamera(stub, args)
	}
	return shim.Error(fmt.Sprintf("Unknown action, check the first argument, must be one of 'delete', 'query', or 'move'. But got: %v", fun))

}

func main() {
	err := shim.Start(new(camerarecord))

	if err != nil {
		fmt.Print(err)
	}
}

