package studentcc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type StudentChaincode struct{}

type studentInfo struct {
	ObjectType string `json:"docType"`
	Id         string `json:"id"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Email      string `json:"email"`
	Mobile     string `json:"mobile"`
	Address    string `json:"address"`
	City       string `json:"city"`
}

type studentHistory struct {
	TxId      string       `json:"TxId"`
	Value     *studentInfo `json:"Value"`
	Timestamp string       `json:"Timestamp"`
	IsDelete  string       `json:"IsDelete"`
}

func (t *StudentChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *StudentChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	if function == "createStudent" {
		return t.createStudent(stub, args)
	} else if function == "updateStudent" {
		return t.updateStudent(stub, args)
	} else if function == "readStudent" {
		return t.readStudent(stub, args)
	} else if function == "readAllStudents" {
		return t.readAllStudents(stub, args)
	} else if function == "deleteStudent" {
		return t.deleteStudent(stub, args)
	} else if function == "getHistoryForStudent" {
		return t.getHistoryForStudent(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) // error
	return shim.Error("Received unknown function invocation")
}

func (t *StudentChaincode) createStudent(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0        1          2         3       4         5       6
	// "key","firstname","lastname","email","mobile","address","city"
	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7")
	}

	// ==== Input sanitation ====
	fmt.Println("- start createStudent")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return shim.Error("6th argument must be a non-empty string")
	}
	if len(args[6]) <= 0 {
		return shim.Error("7th argument must be a non-empty string")
	}
	key := args[0]
	firstname := strings.ToLower(args[1])
	lastname := strings.ToLower(args[2])
	email := strings.ToLower(args[3])
	mobile := strings.ToLower(args[4])
	address := strings.ToLower(args[5])
	city := strings.ToLower(args[6])

	// ==== Check if student already exists ====
	studentAsBytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Failed to get student: " + err.Error())
	} else if studentAsBytes != nil {
		fmt.Println("This student already exists: " + key)
		return shim.Error("This student already exists: " + key)
	}

	// ==== Create student object and marshal to JSON ====
	objectType := "student"
	student := &studentInfo{
		ObjectType: objectType,
		Id:         key,
		FirstName:  firstname,
		LastName:   lastname,
		Email:      email,
		Mobile:     mobile,
		Address:    address,
		City:       city,
	}
	studentJSONasBytes, err := json.Marshal(student)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save student to stateDB ===
	err = stub.PutState(key, studentJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end createStudent")
	return shim.Success(nil)
}

func (t *StudentChaincode) updateStudent(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//   0        1          2         3       4         5       6
	// "key","firstname","lastname","email","mobile","address","city"
	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7")
	}

	// ==== Input sanitation ====
	fmt.Println("- start updateStudent")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return shim.Error("5th argument must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return shim.Error("6th argument must be a non-empty string")
	}
	if len(args[6]) <= 0 {
		return shim.Error("7th argument must be a non-empty string")
	}
	key := args[0]
	firstname := strings.ToLower(args[1])
	lastname := strings.ToLower(args[2])
	email := strings.ToLower(args[3])
	mobile := strings.ToLower(args[4])
	address := strings.ToLower(args[5])
	city := strings.ToLower(args[6])

	// get student from stateDB
	studentAsBytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Failed to get student:" + err.Error())
	} else if studentAsBytes == nil {
		return shim.Error("student does not exist")
	}

	student := studentInfo{}
	err = json.Unmarshal(studentAsBytes, &student) // unmarshal it aka JSON.parse()
	if err != nil {
		return shim.Error(err.Error())
	}
	student.FirstName = firstname
	student.LastName = lastname
	student.Email = email
	student.Mobile = mobile
	student.Address = address
	student.City = city

	studentJSONasBytes, _ := json.Marshal(student)
	err = stub.PutState(key, studentJSONasBytes) // update the student
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end updateStudent")
	return shim.Success(nil)
}

func (t *StudentChaincode) readStudent(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key, jsonResp string
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	// ==== Input sanitation ====
	fmt.Println("- start readStudent")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	key = args[0]

	studentAsBytes, err := stub.GetState(key) // get the student from stateDB
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	} else if studentAsBytes == nil {
		jsonResp = "{\"Error\":\"student does not exist: " + key + "\"}"
		return shim.Error(jsonResp)
	}

	fmt.Println("- end readStudent")
	return shim.Success(studentAsBytes)
}

func (t *StudentChaincode) readAllStudents(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Expecting None")
	}

	resultsIterator, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Printf("- readAllStudents queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (t *StudentChaincode) deleteStudent(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp string
	var student studentInfo
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	fmt.Println("- start deleteStudent")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	key := args[0]

	studentAsBytes, err := stub.GetState(key) // get the marble from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return shim.Error(jsonResp)
	} else if studentAsBytes == nil {
		jsonResp = "{\"Error\":\"student does not exist: " + key + "\"}"
		return shim.Error(jsonResp)
	}

	err = json.Unmarshal([]byte(studentAsBytes), &student)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to decode JSON of: " + key + "\"}"
		return shim.Error(jsonResp)
	}

	err = stub.DelState(key) // remove the student from stateDB
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	fmt.Println("- end deleteStudent")
	return shim.Success(nil)
}

func (t *StudentChaincode) getHistoryForStudent(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	key := args[0]

	fmt.Printf("- start getHistoryForStudent: %s\n", key)

	resultsIterator, err := stub.GetHistoryForKey(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	result := []*studentHistory{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var value *studentInfo = nil
		if !response.IsDelete {
			value = &studentInfo{}
			err = json.Unmarshal(response.Value, value)
			if err != nil {
				return shim.Error(err.Error())
			}
		}

		history := &studentHistory{
			TxId:      response.TxId,
			Value:     value,
			Timestamp: time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String(),
			IsDelete:  strconv.FormatBool(response.IsDelete),
		}
		result = append(result, history)
	}

	bytes, err := json.Marshal(result)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Printf("- getHistoryForStudent returning:\n%s\n", string(bytes))
	return shim.Success(bytes)
}

// ===========================================================================================
// constructQueryResponseFromIterator constructs a JSON array containing query results from
// a given result iterator
// ===========================================================================================
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return &buffer, nil
}
