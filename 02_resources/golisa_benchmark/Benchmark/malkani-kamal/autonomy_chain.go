package main

import (
	"encoding/json"
	"bytes"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type lancastersmartcontract struct{

}

type Cars struct{
	CarDescription         string `json:"car_description"`
	CarId                  string `json:"car_id"`
	CarRegistrationNo      string `json:"car_resgistrationno"`
	CarType 			   string `json:"car_type"`
	CarCameraId            string `json:"car_cameraid"`
	CarStatus      		   string `json:"car_status"`
	CarCreated_At 		   string `json:"car_createdate"`
	CarUpdated_At          string `json:"car_updatedate"`
}

//Init Function
func (s *lancastersmartcontract) Init(stub shim.ChaincodeStubInterface) sc.Response {

	_, args := stub.GetFunctionAndParameters()
	var cars = Cars{
		CarId:             args[0],
		CarRegistrationNo: args[1],
		CarDescription:    args[2],
		CarType:           args[3],
		CarCameraId:       args[4],
		CarStatus:         args[5],
		CarCreated_At:     args[6],
		CarUpdated_At:     args[7],}

	carAsBytes, _ := json.Marshal(cars)

	var uniqueID = args[1]

	err := stub.PutState(uniqueID, carAsBytes)

	if err != nil {
		return shim.Error("Failed to enter data")
	}

	return shim.Success([]byte("Chaincode Successfully initialized"))
}

//CreateCar ... this function is used to create doctors
func CreateCar(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}
	var cars = Cars{
		CarId:             args[0],
		CarRegistrationNo: args[1],
		CarDescription:    args[2],
		CarType:           args[3],
		CarCameraId:       args[4],
		CarStatus:         args[5],
		CarCreated_At:     args[6],
		CarUpdated_At:     args[7],}

	carAsBytes, _ := json.Marshal(cars)

	var uniqueID = args[1]

	err := stub.PutState(uniqueID, carAsBytes)

	if err != nil {
		fmt.Println("Error in create car")
	}

	return shim.Success(nil)
}

// Deletes an entity from state
func (s *lancastersmartcontract) delete(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// Delete the key from the state in ledger
	err := stub.DelState(args[0])
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

func (s *lancastersmartcontract) queryCar(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	carAsBytes, _ := stub.GetState(args[0])
	return shim.Success(carAsBytes)
}


//Invoke function
func (s *lancastersmartcontract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	fun, args := stub.GetFunctionAndParameters()
	if fun == "CreateCar" {
		fmt.Println("Error occured ==> ")
		//logger.Info("########### create docs ###########")
		return CreateCar(stub, args)
	}  else if fun == "initLedger" {
		return s.initLedger(stub)
	}  else if fun == "delete" {
		// Deletes an entity from its state
		return s.delete(stub, args)
	}  else if fun == "query" {
		return s.queryCar(stub, args)
	}  else if fun == "queryAllCars" {
		return s.queryAllCars(stub)
	}  else if fun == "expiryofCar" {
		return s.expiryofCar(stub, args)
	}  else if fun == "isExpired" {
		return s.isExpired(stub, args)
	}
	return shim.Error(fmt.Sprintf("Unknown action, check the first argument, must be one of 'delete', 'query', or 'move'. But got: %v", fun))

}

func (s *lancastersmartcontract) initLedger(stub shim.ChaincodeStubInterface) sc.Response {
	cars := []Cars{
		Cars{CarId: "1", CarRegistrationNo: "LA7432188", CarDescription: "AUDI Blue", CarType: "Lancastercar", CarCameraId: "C005", CarStatus: "Active", CarCreated_At: "03032020", CarUpdated_At: "03032020"},
		Cars{CarId: "2", CarRegistrationNo: "LA1234567", CarDescription: "FORD Red", CarType: "Lancastercar", CarCameraId: "C005", CarStatus: "Active", CarCreated_At: "03032020", CarUpdated_At: "03032020"},
		Cars{CarId: "3", CarRegistrationNo: "LA1233398", CarDescription: "HYUNDAI Green", CarType: "Lancastercar", CarCameraId: "C005", CarStatus: "Active", CarCreated_At: "03032020", CarUpdated_At: "03032020"},
		Cars{CarId: "4", CarRegistrationNo: "LA7432238", CarDescription: "VOLKSWAGEN Yellow", CarType: "Lancastercar", CarCameraId: "C005", CarStatus: "Active", CarCreated_At: "03032020", CarUpdated_At: "03032020"},
		Cars{CarId: "5", CarRegistrationNo: "LA7422188", CarDescription: "TESLA Black", CarType: "Lancastercar", CarCameraId: "C005", CarStatus: "Active", CarCreated_At: "03032020", CarUpdated_At: "03032020"},
		Cars{CarId: "6", CarRegistrationNo: "LA7333188", CarDescription: "PEUGEOT Purple", CarType: "Lancastercar", CarCameraId: "C005", CarStatus: "Active", CarCreated_At: "03032020", CarUpdated_At: "03032020"},
		Cars{CarId: "7", CarRegistrationNo: "LA7477788", CarDescription: "FIAT Blue", CarType: "Lancastercar", CarCameraId: "C005", CarStatus: "Active", CarCreated_At: "03032020", CarUpdated_At: "03032020"},
		Cars{CarId: "8", CarRegistrationNo: "LA7438888", CarDescription: "TATA Indigo", CarType: "Lancastercar", CarCameraId: "C005", CarStatus: "Active", CarCreated_At: "03032020", CarUpdated_At: "03032020"},
		Cars{CarId: "9", CarRegistrationNo: "LA7555188", CarDescription: "HOLDEN Brown", CarType: "Lancastercar", CarCameraId: "C005", CarStatus: "Active", CarCreated_At: "03032020", CarUpdated_At: "03032020"},
		Cars{CarId: "10", CarRegistrationNo: "LA7111188", CarDescription: "AUDI Black", CarType: "Lancastercar", CarCameraId: "C005", CarStatus: "Active", CarCreated_At: "03032020", CarUpdated_At: "03032020"},
	}

	i := 0
	for i < len(cars) {
		fmt.Println("i is ", i)
		carAsBytes, _ := json.Marshal(cars[i])
		fmt.Println("Marshaling done for ", i)
		uniqueID :=cars[i].CarRegistrationNo
		err:= stub.PutState(uniqueID, carAsBytes)
		if err != nil {
			return shim.Error("Failed to enter data "+ err.Error())
		}
		fmt.Println("Added", cars[i])
		i = i + 1
	}
	return shim.Success(nil)
}

func (s *lancastersmartcontract) queryAllCars(stub shim.ChaincodeStubInterface) sc.Response {
	startKey := "LA0000000"
	endKey := "LA9999999"

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Values\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllCars:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}


func (s *lancastersmartcontract) expiryofCar(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	carAsBytes, _ := stub.GetState(args[0])
	car := Cars{}

	json.Unmarshal(carAsBytes, &car)
	car.CarStatus = "Inactive"

	carAsBytes, _ = json.Marshal(car)
	stub.PutState(args[0], carAsBytes)

	return shim.Success(nil)
}

func (s *lancastersmartcontract) isExpired(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	carAsBytes, _ := stub.GetState(args[0])
	car := Cars{}

	json.Unmarshal(carAsBytes, &car)
	//StatusAsByte := []byte(car.CarStatus)
	if (car.CarStatus == "Inactive") {
		return shim.Success([]byte("Y"))
	} else {
		 return shim.Success([]byte("N"))
	}
}

func main() {
	err := shim.Start(new(lancastersmartcontract))

	if err != nil {
		fmt.Print(err)
	}
}
