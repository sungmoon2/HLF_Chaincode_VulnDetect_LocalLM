package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type Student struct {
	RollNo			string	`json:"RollNo"`
	FName			string	`json:"FName"`
	LName			string	`json:"LName"`
	Grades			string	`json:"Grades"`
}

type StudentChaincode struct {

}

func (s *StudentChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (s *StudentChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Chaincode custom Invoke")
	function, args := stub.GetFunctionAndParameters()

	if function == "Write" {
		return s.Write(stub, args)
	} else if function == "Read" {
		return s.Read(stub, args)
	}

	return shim.Error("Invalid invoke function name.")
}

func (s *StudentChaincode) Write(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("Writing Student Data")
	var student Student
	var err error

	student.RollNo = args[0]
	student.FName = args[1]
	student.LName = args[2]
	student.Grades = args[3]

	studentAsBytes, _ := json.Marshal(student)
	err = stub.PutState(student.RollNo, studentAsBytes)

	if err != nil {
		return shim.Error(err.Error())
	  }

	return shim.Success(nil)
}

func (s *StudentChaincode) Read(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var rollNo string
	var err error
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting Roll No to query.")
	}

	rollNo = args[0]

	Avalbytes, err := stub.GetState(rollNo)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + rollNo + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + lcno + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(Avalbytes)
}

func main() {
	err := shim.Start(new(StudentChaincode))
	if err != nil {
		fmt.Printf("Error starting Student Chaincode: %s", err)
	}
}
