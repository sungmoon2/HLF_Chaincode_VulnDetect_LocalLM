package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type SmartContract struct {
}

//MedicalRecord structure

type MedicalRecord struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Weight string `json:"weight"`
	Age    string `json:"age"`
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	//Invoking initLedger will add these Medical records to blockchain or else we can add records indipendently by invoking addMedicalRecord
	medicalrecords := []MedicalRecord{
		MedicalRecord{ID: "101", Name: "shazu", Weight: "65", Age: "22"},
		MedicalRecord{ID: "102", Name: "rakhi", Weight: "70", Age: "24"},
		MedicalRecord{ID: "103", Name: "akhil", Weight: "54", Age: "17"},
		MedicalRecord{ID: "104", Name: "varun", Weight: "56", Age: "32"},
		MedicalRecord{ID: "105", Name: "anula", Weight: "90", Age: "23"},
	}
	j := 0
	// A key value is generated for each record starting from MedicalRecord0
	for j < len(medicalrecords) {
		fmt.Println("j is ", j)
		medicalrecordAsBytes, _ := json.Marshal(medicalrecords[j])
		APIstub.PutState("MedicalRecord"+strconv.Itoa(j), medicalrecordAsBytes)
		fmt.Println("Added", medicalrecords[j])
		j = j + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// invoke the fuction by verifying the function name

	if function == "queryMedicalRecord" {
		return s.queryMedicalRecord(APIstub, args)
	} else if function == "addMedicalRecord" {
		return s.addMedicalRecord(APIstub, args)
	} else if function == "updateMedicalRecord" {
		return s.updateMedicalRecord(APIstub, args)
	} else if function == "queryAllMedicalRecords" {
		return s.queryAllMedicalRecords(APIstub)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

//retrive all medical records from initledger
func (s *SmartContract) queryAllMedicalRecords(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "MedicalRecord0"
	endKey := "MedicalRecord999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
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

		buffer.WriteString("{\"Key\":") //this sperates key from medicalrecord content key value is used to retrive any medical record
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":") //Medical Record is printed next to its key
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllMedicalRecords:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) queryMedicalRecord(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	medicalrecordAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(medicalrecordAsBytes)
}

func (s *SmartContract) addMedicalRecord(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	var medicalrecord = MedicalRecord{ID: args[1], Name: args[2], Weight: args[3], Age: args[4]}

	medicalrecordAsBytes, _ := json.Marshal(medicalrecord)
	APIstub.PutState(args[0], medicalrecordAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) updateMedicalRecord(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	//retrive medicalrecord from args[0]
	medicalrecordAsBytes, _ := APIstub.GetState(args[0])
	medicalrecord := MedicalRecord{}

	json.Unmarshal(medicalrecordAsBytes, &medicalrecord)
	//check args[2] to identify what to change
	if args[2] == "weight" {
		medicalrecord.Weight = args[1]
	} else if args[2] == "age" {
		medicalrecord.Age = args[1]
	} else if args[2] == "name" {
		medicalrecord.Name = args[1]
	} else {
		return shim.Error("Incorrect last argument expecting weight or age or name")
	}

	medicalrecordAsBytes, _ = json.Marshal(medicalrecord)
	APIstub.PutState(args[0], medicalrecordAsBytes)

	return shim.Success(nil)
}

func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
