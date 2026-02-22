package main

/* Imports
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

// Define the car structure, with 4 properties.  Structure tags are used by encoding/json library
type ApplicationRecord struct {
	Dataset string `json:"数据集"`
	Applicant string `json:"申请人"`
	Noticer string `json:"通知人"`
	Noticetime string `json:"通知时间"`
	Confirmor string `json:"确认人"`
	Confirmtime string `json:"确认时间"`
	Omicsoriginalresultprovider string `json:"组学原始数据提供人"`
	Omicsoriginalapplicationreceivedtime string `json:"组学原始数据申请收到时间"`
	Omicsanalysisresultprovider string `json:"组学分析数据提供人"`
	Omicsapplicationreceivedtime string `json:"组学分析数据申请收到时间"`
	Phenotypicdataprovider string `json:"表型数据提供人"`
	Phenotypicapplicationreceivedtime string `json:"表型数据申请收到时间"`
	Datadeliverytime string `json:"数据交付时间"`
	Datadeliverymethod string `json:"数据交付方式"`
	Deliverer string `json:"数据交付人"`
	Deliverercontactinformation string `json:"联系方式"`
}

/*
 * The Init method is called when the Smart Contract "fabcar" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "fabcar"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryApplicationRecord" {
		return s.queryApplicationRecord(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "createApplicationRecord" {
		return s.createApplicationRecord(APIstub, args)
	} else if function == "queryAllApplicationRecords" {
		return s.queryAllApplicationRecords(APIstub)
	}else if function == "queryApplicationRecordByApplicant" { //find marbles for owner X using rich query
		return s.queryApplicationRecordByApplicant(APIstub, args)
	}


	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryApplicationRecord(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	applicationRecordAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(applicationRecordAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	applicationRecords := []ApplicationRecord{
		ApplicationRecord{Dataset:"数据集",Applicant: "申请人",Noticer:"通知人", Noticetime: "通知时间", Confirmor: "确认人", Confirmtime: "确认时间",Omicsoriginalresultprovider:"组学原始数据提供人",Omicsoriginalapplicationreceivedtime:"组学原始数据申请收到时间",Omicsanalysisresultprovider:"组学分析数据提供人",Omicsapplicationreceivedtime:"组学分析数据申请收到时间",Phenotypicdataprovider:"表型数据提供人",Phenotypicapplicationreceivedtime:"表型数据申请收到时间",Datadeliverytime:"数据交付时间",Datadeliverymethod:"数据交付方式",Deliverer:"数据交付人",Deliverercontactinformation:"联系方式"},

	}

	i := 0
	for i < len(applicationRecords) {
		fmt.Println("i is ", i)
		applicationRecordAsBytes, _ := json.Marshal(applicationRecords[i])
		APIstub.PutState("ApplicationRecord"+strconv.Itoa(i),applicationRecordAsBytes)
		fmt.Println("Added", applicationRecords[i])
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) createApplicationRecord(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 17 {
		return shim.Error("Incorrect number of arguments. Expecting 16")
	}

	var applicationRecord = ApplicationRecord{Dataset: args[1],Applicant: args[2],Noticer: args[3], Noticetime: args[4], Confirmor: args[5], Confirmtime: args[6], Omicsoriginalresultprovider: args[7], Omicsoriginalapplicationreceivedtime: args[8], Omicsanalysisresultprovider: args[9], Omicsapplicationreceivedtime: args[10], Phenotypicdataprovider: args[11], Phenotypicapplicationreceivedtime: args[12], Datadeliverytime: args[13], Datadeliverymethod: args[14], Deliverer: args[15], Deliverercontactinformation: args[16]}

	applicationRecordAsBytes, _ := json.Marshal(applicationRecord)
	APIstub.PutState(args[0], applicationRecordAsBytes)

	//获取交易ID
	
	//
	
	return shim.Success(nil)
}

func (s *SmartContract) queryAllApplicationRecords(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "APPLICATIONRECORD0"
	endKey := "APPLICATIONRECORD9999"

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

	fmt.Printf("- queryAllApplicationRecords:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) queryApplicationRecordByApplicant(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	applicant := strings.ToLower(args[0])

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"ApplicationRecord\",\"申请人\":\"%s\"}}", applicant)

	queryResults, err := getQueryResultForQueryString(APIstub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

//按照申请人来进行查找
//被调用的函数
func getQueryResultForQueryString(APIstub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
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

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}





// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
