package main

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type SCc struct{
}

//----------------------------------------------------------------------------
func (t *SCc) Init(stub shim.ChaincodeStubInterface) pb.Response {
	var A, B string
	var intA, intB int
	var err error	
	fmt.Println("Init Chaincode");
	_, args := stub.GetFunctionAndParameters()
	
	if len(args) != 4 {
		return shim.Error("Expecting 4 arguments")
	}

	A = args[0]
	intA, err = strconv.Atoi(args[1])
	
	if err != nil {
		return shim.Error("Expecting integer value!")
	}

	B = args[2]
	intB, err = strconv.Atoi(args[3])

	if err != nil {
		return shim.Error("Expecting integer value")
	}
	
	fmt.Printf("Value:\nA = %d\nB = %d\n", intA, intB)
	
	err = stub.PutState(A, []byte(strconv.Itoa(intA)))

	if err != nil {
		return shim.Error(err.Error())	
 	}
	
	err = stub.PutState(B, []byte(strconv.Itoa(intB)))

	if err != nil {
		return shim.Error(err.Error())	
 	}

	return shim.Success(nil)
}
//----------------------------------------------------------------------------
func (t *SCc) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Invoke SCc")
	function, args := stub.GetFunctionAndParameters()

	if function == "invoke" {
		return t.invoke(stub, args)
	}else if function == "delete" {
		return t.delete(stub, args)
	}else if function == "query" {
		return t.query(stub, args)
	} else if function == "getAllHistoryKey" {
		return t.getAllHistoryKey(stub, args)
	}
	return shim.Error("Invalid invoke function name. Expecting \"invoke\" \"delete\" \"query\" \"getAllHistoryKey\"")
}
//----------------------------------------------------------------------------
func (t *SCc) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A, B string
	var intA, intB, Sum int
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of Arguments")
	}
	//----------------------------------------------------
	A = args[0]
	byteA, err := stub.GetState(A)
	if err != nil {
		return shim.Error("A: Failed to get state")
	}
	if byteA == nil {
		return shim.Error("A: Entity not found")
	}
	intA, _ = strconv.Atoi(string(byteA)) 
	//----------------------------------------------------
	B = args[1]
	byteB, err := stub.GetState(B)
	if err != nil {
		return shim.Error("B: Failed to get state")
	}
	if byteB == nil {
		return shim.Error("B: Entity not found")
	}
	intB, _ = strconv.Atoi(string(byteB)) 
	//----------------------------------------------------
	Sum, err = strconv.Atoi(args[2])
	if err != nil {
		fmt.Println("Error. %s must be more %s", A, B)		
		return shim.Error("Invalid transaction amount")
	}
	fmt.Println("It was: %s = %d, %s = %d", A, intA, B, intB)
	intA = intA - Sum
	intB = intB + Sum
	fmt.Println("Became: %s = %d, %s = %d", A, intA, B, intB)
	
	err = stub.PutState(A, []byte(strconv.Itoa(intA)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(intB)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}
//----------------------------------------------------------------------------
func (t *SCc) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		shim.Error("Incorrect number of Arguments")
	}

	A := args[0]
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Filed to delete state")
	}
	
	return shim.Success(nil)
}
//----------------------------------------------------------------------------
func (t *SCc) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string
	var err error
		
	if len(args) != 1 {
		shim.Error("Incorrect number of Arguments")
	}

	A = args[0]
	byteA, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}
	
	if byteA == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(byteA) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return shim.Success(byteA)
}
//----------------------------------------------------------------------------
func (t *SCc) getAllHistoryKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Error. Expecting 1 argument")
	}
	
	name := args[0]
	fmt.Println("Start getAllHistoryKey for %s", name)
	resultIt, err := stub.GetHistoryForKey(name)
	if err != nil {
		return shim.Error(err.Error())
	}
	
	defer resultIt.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	
	buffAlredyWritten := false
	//----------------------------------------------------
	for resultIt.HasNext(){
		response, err := resultIt.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		
		if buffAlredyWritten == true {
		 	buffer.WriteString(", ")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")
		buffer.WriteString(", \"Value\":")
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}
		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		buffAlredyWritten = true
	}
	//----------------------------------------------------
	buffer.WriteString("]")
	fmt.Printf("getAllHistoryKey returning:\n%s\n", buffer.String())
	return shim.Success(buffer.Bytes())
}

func main(){

	err := shim.Start(new(SCc))
	if err != nil {
		fmt.Printf("Error starting simple chaincode(SCc): %s", err)
	}

}
