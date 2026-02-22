package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"strings"
)
// 该类是实现Chaincode接口的，不需要特别多的成员，只需要将具体方法在后面写出即可
type SimpleChaincode struct {
}
/*
  定义了意见记录和审批记录的数据结构，这些数据在后端都是以String形式存储的
*/
type opinionRecord struct {
	DocType     string `json:"docType"`
	Id          string `json:"id"`
	Department  string `json:"department"`
	Name        string `json:"name"`
	Object      string `json:"object"`
	Type        string `json:"type"`
	OpinionTime string `json:"opinionTime"`
	DoneTime    string `json:"doneTime"`

	Extension1  string `json:"extension1"`
	Extension2  string `json:"extension2"`
	Extension3  string `json:"extension3"`
}

type reviewRecord struct {
	DocType    string `json:"docType"`
	Department string `json:"department"`
	Name       string `json:"name"`
	Object     string `json:"object"`
	Result     string `json:"result"`
	ReviewTime string `json:"reviewTime"`

	Extension1  string `json:"extension1"`
	Extension2  string `json:"extension2"`
	Extension3  string `json:"extension3"`
}
/*
	主函数与初始化函数是链码运行的必要函数，虽然其中并没有具体内容
*/
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple Chaincode: %s", err)
	}
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

/*
	Invoke函数，用于响应Fabric SDK的调用，该调用通过gRPC实现
*/
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("Invoking is running:" + function)
	switch function {
	case "createOpinionRecord":
		return t.createOpinionRecord(stub, args)
	case "createReviewRecord":
		return t.createReviewRecord(stub, args)
	case "modifyOpinionRecord":
		return t.modifyOpinionRecord(stub, args)
	case "queryWithQueryString":
		return t.queryWithQueryString(stub, args)
	case "queryRecordByObject":
		return t.queryRecordByObject(stub, args)
	case "queryRecordByUser":
		return t.queryRecordByUser(stub, args)
	default:
		return shim.Error("Received unknown function invocation")
	}
}

// 创建意见信息
func (t *SimpleChaincode) createOpinionRecord(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	if len(args) != 11 {
		return shim.Error("Incorrect number of arguments: expecting 11; ")
	}
	for i := 0 ; i < 7 ; i += 1 {
		if len(args[i]) == 0 {
			return shim.Error("arg[0] to arg[6] should not be empty")
		}
	}

	docType := "opinionRecord"
	uuid := strings.ToLower(args[0])
	opinionTime := strings.ToLower(args[6])
	doneTime := strings.ToLower(args[7])

	extension1 := args[8]
	extension2 := args[9]
	extension3 := args[10]

	dataRecordJsonAsBytes, err := stub.GetState(uuid)
	if err != nil {
		return shim.Error("Failed to get opinion record: " + err.Error())
	} else if dataRecordJsonAsBytes != nil {
		return shim.Error("Duplicated opinion record found: " + uuid)
	}

	dataRecord := &opinionRecord{
		docType,
		args[1],
		args[2],
		args[3],
		args[4],
		args[5],
		opinionTime,
		doneTime,
		extension1,
		extension2,
		extension3,
	}

	dataRecordJsonAsBytes, err = json.Marshal(dataRecord)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(uuid, dataRecordJsonAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}
// 创建审批信息
func (t *SimpleChaincode) createReviewRecord(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	if len(args) != 9 {
		return shim.Error("Incorrect number of arguments: expecting 9.")
	}
	for i := 0 ; i < 6 ; i += 1 {
		if len(args[i]) == 0 {
			return shim.Error("arg[0] to arg[5] should not be empty")
		}
	}

	docType := "reviewRecord"
	uuid := strings.ToLower(args[0])
	reviewTime := strings.ToLower(args[5])
	extension1 := args[6]
	extension2 := args[7]
	extension3 := args[8]


	reviewRecordJsonAsBytes, err := stub.GetState(uuid)
	if err != nil {
		return shim.Error("Failed to get review record: " + err.Error())
	} else if reviewRecordJsonAsBytes != nil {
		return shim.Error("Duplicated review record found: " + uuid)
	}

	reviewRecord := &reviewRecord{
		docType,
		args[1],
		args[2],
		args[3],
		args[4],
		reviewTime,
		extension1,
		extension2,
		extension3,
	}

	reviewRecordJsonAsBytes, err = json.Marshal(reviewRecord)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(uuid, reviewRecordJsonAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}
// 修改意见信息，用于更新意见的决定时间
func (t *SimpleChaincode) modifyOpinionRecord(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments: expecting 2")
	}
	for i := 0 ; i < 2 ; i += 1 {
		if len(args[i]) == 0 {
			return shim.Error("arg[0] to arg[1] should not be empty")
		}
	}
	uuid := strings.ToLower(args[0])
	doneTime := strings.ToLower(args[1])

	opinionRecordJsonAsBytes, err := stub.GetState(uuid)
	if err != nil {
		return shim.Error("Failed to get opinion record: " + err.Error())
	} else if opinionRecordJsonAsBytes == nil {
		return shim.Error("opinion record does not exist")
	}

	opinionRecordInstance := opinionRecord{}
	err = json.Unmarshal(opinionRecordJsonAsBytes, &opinionRecordInstance)
	if err != nil {
		return shim.Error(err.Error())
	}

	if len(opinionRecordInstance.DoneTime) != 0  {
		return shim.Error("Unable to modify a opinion record that has been done")
	}
	opinionRecordInstance.DoneTime = doneTime

	opinionRecordJsonAsBytes, _ = json.Marshal(opinionRecordInstance)
	err = stub.PutState(uuid, opinionRecordJsonAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}
// 辅助函数，构造返回串格式
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	buffer.WriteString("{\"list\":[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")
		buffer.WriteString(", \"value\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]}")

	return &buffer, nil
}
// Ad Hoc即席查询
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer func(resultsIterator shim.StateQueryIteratorInterface) {
		_ = resultsIterator.Close()
	}(resultsIterator)

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func (t *SimpleChaincode) queryWithQueryString(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect arguments, must be a query string")
	}
	queryString := args[0]
	queryResult, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResult)
}
// 通过目标ID进行查询，可以分为查意见记录与审批记录两种，由DocType标识
func (t *SimpleChaincode) queryRecordByObject(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments, expecting 2")
	}
	docType := args[0]
	object := args[1]
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"%s\",\"object\":\"%s\"}}", docType, object)
	queryResult, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResult)
}
// 通过用户信息进行查询，可以分为查意见记录与审批记录两种，由DocType标识
func (t *SimpleChaincode) queryRecordByUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments, expecting 3")
	}
	docType := args[0]
	department := args[1]
	name := args[2]
	var err error
	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"%s\",\"department\":\"%s\",\"name\":\"%s\"}}", docType, department, name)
	queryResult, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResult)
}
