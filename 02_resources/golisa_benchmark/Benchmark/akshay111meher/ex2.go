
package main

import (
	"fmt"
	"strings"
	"strconv"
	"encoding/json"
	"encoding/base64"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	//pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}


type Employee struct {
	name string 
	employeeId int
	project string
}

type Customer struct {
	name string
	customerId int
}
type Project struct {
	name string
	projectId int
	customerOf string
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	function, _ := stub.GetFunctionAndParameters()
    fmt.Println("Init is running " + function)
    return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
    function, args := stub.GetFunctionAndParameters()
    fmt.Println("Invoke is running " + function)
    
    if function =="initEmployee"{
		return t.initEmployee(stub,args)
	}else if function =="getEmployee"{
		return t.getEmployee(stub,args)
	}else if function =="initCustomer"{
		return t.initCustomer(stub,args)
	}else if function =="getCustomer"{
		return t.getCustomer(stub,args)
	}else if function =="initProject"{
		return t.initProject(stub,args)
	}else if function =="getProject"{
		return t.getProject(stub,args)
	}
	return shim.Error("Received unknown function invocation")
}

func (t *SimpleChaincode) initProject(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	// ==== Input sanitation ====
	fmt.Println("- start initProject")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	
	projectName := args[0]
	customerOf := strings.ToLower(args[2])
	projectId, err := strconv.Atoi(args[1])
	projectIdAsString := args[1]
	if err != nil {
		return shim.Error("2rd argument must be a numeric string")
	}
	
	projectAsBytes, err := stub.GetState(projectIdAsString)
	if err != nil {
		return shim.Error("Failed to get project: " + err.Error())
	} else if projectAsBytes != nil {
		fmt.Println("This project already exists: " + projectIdAsString)
		return shim.Error("This project already exists: " + projectIdAsString)
	}
	
	project:= &Project{projectName,projectId,customerOf}
	
	projectJSONasBytes, err := json.Marshal(project)
	
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(projectIdAsString, projectJSONasBytes)
	
	if err != nil {
		return shim.Error(err.Error())
	}
	
	// composite key creation for searching projects according to customer
	indexName := "customerOf"
	customerOfIndexKey, err := stub.CreateCompositeKey(indexName, []string{customerOf, projectIdAsString})
	if err != nil {
		return shim.Error(err.Error())
	}

	value := []byte{0x00}
	stub.PutState(customerOfIndexKey, value)

	fmt.Println("- end initProject")
	return shim.Success(nil)
}

func (t *SimpleChaincode) getProject(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	if len(args) !=1{
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	projectId := args[0]
	project, err := stub.GetState(projectId)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed retrieving employee [%s]: [%s]", projectId, err))
	}
	
	return shim.Success([]byte(base64.StdEncoding.EncodeToString(project)))
}


func (t *SimpleChaincode) getCustomer(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	if len(args) !=1{
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	customerId := args[0]
	customer, err := stub.GetState(customerId)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed retrieving employee [%s]: [%s]", customerId, err))
	}
	
	return shim.Success([]byte(base64.StdEncoding.EncodeToString(customer)))
}

func (t *SimpleChaincode) initCustomer(stub shim.ChaincodeStubInterface,args []string)pb.Response{
		if len(args) != 2{
			return shim.Error("Incorrect number of arguments. Expecting 2")
		}	
			// ==== Input sanitation ====
		fmt.Println("- start initCustomer")
		if len(args[0]) <= 0 {
			return shim.Error("1st argument must be a non-empty string")
		}
		if len(args[1]) <= 0 {
			return shim.Error("2nd argument must be a non-empty string")
		}
		
		customerName := args[0]
		customerId, err := strconv.Atoi(args[1])
		customerIdAsString := args[1]
		if err != nil {
			return shim.Error("2rd argument must be a numeric string")
		}
		customerAsBytes, err := stub.GetState(customerIdAsString)
		if err != nil {
			return shim.Error("Failed to get customer: " + err.Error())
		} else if customerAsBytes != nil {
			fmt.Println("This customer already exists: " + customerIdAsString)
			return shim.Error("This customer already exists: " + customerIdAsString)
		}

		customer:= &Customer{customerName,customerId}
		
		customerJSONasBytes, err := json.Marshal(customer)
		
		if err != nil {
			return shim.Error(err.Error())
		}
		err = stub.PutState(customerIdAsString, customerJSONasBytes)
		
		if err != nil {
			return shim.Error(err.Error())
		}
		
		fmt.Println("- end initCustomer")
		return shim.Success(nil)
}

func (t *SimpleChaincode) initEmployee(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}
	// ==== Input sanitation ====
	fmt.Println("- start initEmployee")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	
	employeeName := args[0]
	project := strings.ToLower(args[2])
	employeeId, err := strconv.Atoi(args[1])
	employeeIdAsString := args[1]
	if err != nil {
		return shim.Error("2rd argument must be a numeric string")
	}
	
	employeeAsBytes, err := stub.GetState(employeeIdAsString)
	if err != nil {
		return shim.Error("Failed to get employee: " + err.Error())
	} else if employeeAsBytes != nil {
		fmt.Println("This employee already exists: " + employeeIdAsString)
		return shim.Error("This employee already exists: " + employeeIdAsString)
	}
	
	employee:= &Employee{employeeName,employeeId,project}
	
	employeeJSONasBytes, err := json.Marshal(employee)
	
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(employeeIdAsString, employeeJSONasBytes)
	
	if err != nil {
		return shim.Error(err.Error())
	}
	
	// composite key to get employees by project
	indexName := "project"
	projectIndexKey, err := stub.CreateCompositeKey(indexName, []string{project, employeeIdAsString})
	if err != nil {
		return shim.Error(err.Error())
	}

	value := []byte{0x00}
	stub.PutState(projectIndexKey, value)
	fmt.Println("- end initEmployee")
	return shim.Success(nil)
}

func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface) pb.Response {
    function, args := stub.GetFunctionAndParameters()
    fmt.Println("Query is running " + function)
    
    if function =="getEmployee"{
		return t.getEmployee(stub,args)
	}
	
	return shim.Error("Received unknown function query")
}

func (t *SimpleChaincode) getEmployee(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	if len(args) !=1{
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	employeeId := args[0]
	employee, err := stub.GetState(employeeId)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed retrieving employee [%s]: [%s]", employeeId, err))
	}
	
	return shim.Success([]byte(base64.StdEncoding.EncodeToString(employee)))
}


func main() {
    	err := shim.Start(new(SimpleChaincode))
    if err != nil {
        fmt.Printf("Error starting Simple chaincode: %s", err)
    }
}
