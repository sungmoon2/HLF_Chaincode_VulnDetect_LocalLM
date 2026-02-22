package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	// "github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
	"strconv"
// 	"time"
)

type SmartContract struct {
}

type RequestCommit struct {
	Type string `json:"type"`//0
	TaskId string `json:"taskId"`
	TaskName string `json:"taskName"`
	RequestId string `json:"requestId"`
	RequesterId string `json:"requesterId"`
	RequesterName string `json:"requesterName"`
	RequestDocHash string `json:"requestDocHash"`
	RequestDocName string `json:"requestDocName"`
	TestSoftwareName string `json:"testSoftwareName"`
	UpdateTime string `json:"updateTime"`
}

type RequestReview struct {
	Type string `json:"type"`//1
	TaskId string `json:"taskId"`
	TaskName string `json:"taskName"`
	RequestId string `json:"requestId"`
	RequestReviewer string `json:"requestReviewer"`
	ReviewResult string `json:"reviewResult"`
	UpdateTime string `json:"updateTime"`
}

type TestReport struct {
	Type string `json:"type"`//2
	TaskId string `json:"taskId"`
	TaskName string `json:"taskName"`
	TestReportId string `json:"testReportId"`
	TestReportName string `json:"testReportName"`
	ReportHash string `json:"reportHash"`
	BugReportList string `json:"bugReportList"`
	WorkerId string `json:"workerId"`
	WorkerName string `json:"workerName"`
	UpdateTime string `json:"updateTime"`
}

type ReportReview struct {
	Type string `json:"type"`//3
	TaskId string `json:"taskId"`
	TaskName string `json:"taskName"`
	TestReportId string `json:"testReportId"`
	BugReportId string `json:"bugReportId"`
	BugReportScore string `json:"bugReportScore"`
	ReportReviewer string `json:"reportReviewer"`
	UpdateTime string `json:"updateTime"`
}

type ReportMix struct {
	Type string `json:"type"`//
	TaskId string `json:"taskId"`
	TaskName string `json:"taskName"`
	BugReportList string `json:"bugReportList"`
	ReportHash string `json:"reportHash"`
	ReportMixer string `json:"reportMixer"`
	UpdateTime string `json:"updateTime"`
}

type TaskState struct {
	Type string `json:"type"`//
	TaskId string `json:"taskId"`
	TaskName string `json:"taskName"`
	TaskState string `json:"taskState"`
	UpdateTime string `json:"updateTime"`
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()
	if function == "findOne" {
		return s.findOne(APIstub, args)
	} else if function == "save" {
		return s.save(APIstub, args)
	} else if function == "delete" {
		return s.delete(APIstub, args)
	} else if function == "query" {
		return s.query(APIstub, args)
	} else if function == "queryWithPagination" {
		return s.queryWithPagination(APIstub, args)
	} else if function == "history" {
		return s.history(APIstub, args)
	} 

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) findOne(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	if len(args) > 0 {
		docAsBytes, err := APIstub.GetState(args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(docAsBytes)
	} else{
		return shim.Success(nil)
	}
}

func (s *SmartContract) query(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func (s *SmartContract) queryWithPagination(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	queryString := args[0]
	pageSize, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		return shim.Error(err.Error())
	}
	bookmark := args[2]

	queryResults, err := getQueryResultWithPagination(stub, queryString, int32(pageSize), bookmark)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func (s *SmartContract) save(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) > 0 {
		doctype := args[1]
		var docAsBytes []byte
		var err error
		if doctype == "0" {
			doc := RequestCommit{Type:doctype, TaskId: args[2], TaskName: args[3], RequestId: args[4], RequesterId: args[5], RequesterName: args[6], RequestDocHash: args[7], RequestDocName: args[8], TestSoftwareName: args[9], UpdateTime: args[10]}
			docAsBytes, err = json.Marshal(doc)
// 			docAsBytes, _ = json.Marshal(doc)
		} else if doctype == "1"{
			doc := RequestReview{Type:doctype, TaskId: args[2], TaskName: args[3], RequestId: args[4], RequestReviewer: args[5], ReviewResult: args[6], UpdateTime: args[7]}
			docAsBytes, err = json.Marshal(doc)
// 			docAsBytes, _ = json.Marshal(doc)
		} else if doctype == "2"{
			doc := TestReport{Type:doctype, TaskId: args[2], TaskName: args[3], TestReportId: args[4], TestReportName: args[5], ReportHash: args[6], BugReportList: args[7], WorkerId: args[8], WorkerName: args[9], UpdateTime: args[10]}
			docAsBytes, err = json.Marshal(doc)
// 			docAsBytes, _ = json.Marshal(doc)
		} else if doctype == "3"{
			doc := ReportReview{Type:doctype, TaskId: args[2], TaskName: args[3], TestReportId: args[4], BugReportId: args[5], BugReportScore: args[6], ReportReviewer: args[7], UpdateTime: args[8]}
			docAsBytes, err = json.Marshal(doc)
// 			docAsBytes, _ = json.Marshal(doc)
		} else if doctype == "4"{
			doc := ReportMix{Type:doctype, TaskId: args[2], TaskName: args[3], ReportHash: args[4], BugReportList: args[5], ReportMixer:args[6], UpdateTime: args[7]}
			docAsBytes, err = json.Marshal(doc)
// 			docAsBytes, _ = json.Marshal(doc)
		} else if doctype == "5"{
			doc := TaskState{Type:doctype, TaskId: args[2], TaskName: args[3], TaskState:args[4], UpdateTime: args[5]}
			docAsBytes, err = json.Marshal(doc)
// 			docAsBytes, _ = json.Marshal(doc)
		}
		
		if err != nil {
			return shim.Error(err.Error())
		}
		
		err1 := APIstub.PutState(args[0], docAsBytes)
		if err1 != nil {
			return shim.Error(err1.Error())
		}
		return shim.Success(nil)
	} else{
		return shim.Success(nil)
	}
}

func (t *SmartContract) delete(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

func (s *SmartContract) history(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	key := args[0]
	resultsIterator, err := stub.GetHistoryForKey(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"txId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"doc\":")
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

// 		buffer.WriteString(", \"timestamp\":")
// 		buffer.WriteString("\"")
// 		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
// 		buffer.WriteString("\"")

		buffer.WriteString(", \"isDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return shim.Success(nil)
}

func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator,false)
	if err != nil {
		return nil, err
	}

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func getQueryResultWithPagination(stub shim.ChaincodeStubInterface, queryString string, pageSize int32, bookmark string) ([]byte, error) {

	resultsIterator, responseMetadata, err := stub.GetQueryResultWithPagination(queryString, pageSize, bookmark)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator,true)
	if err != nil {
		return nil, err
	}
	addPaginationMetadataToQueryResults(buffer, responseMetadata)

	return buffer.Bytes(), nil
}

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface, forPage bool) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	if(forPage){
		buffer.WriteString("{\"resultList\":")
	}
	buffer.WriteString("[")

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

		buffer.WriteString(", \"doc\":")
		// Record is a JSON object, so we write as-is

		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	if(forPage){
		buffer.WriteString(",")
	}

	return &buffer, nil
}

func addPaginationMetadataToQueryResults(buffer *bytes.Buffer, responseMetadata *sc.QueryResponseMetadata) *bytes.Buffer {
	buffer.WriteString("\"responseMetadata\":{\"recordsCount\":")
	buffer.WriteString("\"")
	buffer.WriteString(fmt.Sprintf("%v", responseMetadata.FetchedRecordsCount))
	buffer.WriteString("\"")
	buffer.WriteString(", \"bookmark\":")
	buffer.WriteString("\"")
	buffer.WriteString(responseMetadata.Bookmark)
	buffer.WriteString("\"}}")

	return buffer
}

func main() {

	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
