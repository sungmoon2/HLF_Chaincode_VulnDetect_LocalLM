package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
	"time"
)

type SimpleChaincode struct {
}

type data struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Id         int `json:"id"`      //the fieldtags are needed to keep case from bouncing around
	CreateTime string `json:"createTime"`
	BeginDate  string `json:"beginDate"`
	EndDate    string `json:"endDate"`
	Period     string `json:"period"`
	Value      string `json:"value"`
	Merchant   merchant `json:"merchant"`
	Indicator  indicator `json:"indicator"`
	Demension  demension `json:"demension"`
	DataPacket dataPacket `json:"datapacket"`
	IndicatorIdAndBeginDate string `json:"indicatorIdAndBeginDate"`
}

type merchant struct {
	Id         int `json:"id"`    //the fieldtags are needed to keep case from bouncing around
	Name       string `json:"name"`
}

type indicator struct {
	Id         int `json:"id"`    //the fieldtags are needed to keep case from bouncing around
	Code       string `json:"code"`
	Name       string `json:"name"`
	Level       int `json:"level"`
	ParentId       int `json:"parentId"`
}
type demension struct {
	Id         int `json:"id"`    //the fieldtags are needed to keep case from bouncing around
	Code       string `json:"code"`
	Name       string `json:"name"`
}
type dataPacket struct {
	Id         int `json:"id"`    //the fieldtags are needed to keep case from bouncing around
	FilePath       string `json:"filePath"`
	Up2Chain       string `json:"up2Chain"`
}


func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "save" { //create a new marble
		return t.save(stub, args)
	} else if function == "update" { //find marbles based on an ad hoc rich query
		return t.update(stub, args)
	} else if function == "delete" { //find marbles based on an ad hoc rich query
		return t.delete(stub, args)
	} else if function == "query" { //find marbles based on an ad hoc rich query
		return t.query(stub, args)
	} else if function == "getHistoryByKey" { //find marbles based on an ad hoc rich query
		return t.getHistoryByKey(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

func (t *SimpleChaincode) save(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("- start save datastatis")
	//check data
	if len(args) != 20 {
		return shim.Error("Incorrect number of arguments. Expecting 20")
	}

	//string to num
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return shim.Error("1rd argument must be a numeric string")
	}
	createTime := args[1]
	beginData := args[2]
	endData := args[3]

	period := args[4]
	value1 := args[5]
	indicatorIdAndBeginDate := args[6]

	//string to num
	merchantId, err := strconv.Atoi(args[7])
	if err != nil {
		return shim.Error("1rd argument must be a numeric string")
	}
	merchantName := args[8]

	//string to num
	indicatorId, err := strconv.Atoi(args[9])
	if err != nil {
		return shim.Error("10rd argument must be a numeric string")
	}

	indicatorCode := args[10]
	indicatorName := args[11]

	//string to num
	indicatorLevel, err := strconv.Atoi(args[12])
	if err != nil {
		return shim.Error("13rd argument must be a numeric string")
	}

	//string to num
	indicatorParentId, err := strconv.Atoi(args[13])
	if err != nil {
		return shim.Error("14rd argument must be a numeric string")
	}

	//string to num
	demensionId, err := strconv.Atoi(args[14])
	if err != nil {
		return shim.Error("15rd argument must be a numeric string")
	}

	demensionCode := args[15]
	demensionName := args[16]

	//string to num
	dataPacketId, err := strconv.Atoi(args[17])
	if err != nil {
		return shim.Error("18rd argument must be a numeric string")
	}

	dataPacketFilePath := args[18]
	dataPacketUp2Chain := args[19]


	//generate key
	createDateTime, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("createTime argument must be a numeric string")
	}
	tm := time.Unix(int64(createDateTime) / 1000,0)
	dateString := tm.Format("20060102")

	merchantIdString := strconv.Itoa(merchantId)
	dataPacketIdString := strconv.Itoa(dataPacketId)
	key := merchantIdString + "_" + dateString + "_" + period + "_" + dataPacketIdString

	//check data
	dataStatisBytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Failed to get datastatis: " + err.Error())
	} else if dataStatisBytes != nil {
		fmt.Println("data is already exist: " + key)
		return shim.Error("data is already exist:" + key)
	}

	// ==== Create datastatis object and marshal to JSON ====
	objectType := "datastatis"
	merchant := &merchant{merchantId, merchantName}
	indicator := &indicator{indicatorId, indicatorCode, indicatorName, indicatorLevel, indicatorParentId}
	demension := &demension{demensionId, demensionCode, demensionName}
	dataPacket := &dataPacket{dataPacketId, dataPacketFilePath, dataPacketUp2Chain}
	data := &data{objectType, id, createTime, beginData, endData, period, value1,*merchant, *indicator, *demension, *dataPacket, indicatorIdAndBeginDate}
	dataJSONBytes, err := json.Marshal(data)
	if err != nil {
		return shim.Error(err.Error())
	}

	//save datastatis json
	err = stub.PutState(key, dataJSONBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end save datastatis")
	return shim.Success(nil)
}


func (t *SimpleChaincode) update(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("- start update datastatis")

	if len(args) != 20 {
		return shim.Error("Incorrect number of arguments. Expecting 20")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return shim.Error("1rd argument must be a numeric string")
	}
	createTime := args[1]
	beginData := args[2]
	endData := args[3]

	period := args[4]
	value1 := args[5]
	indicatorIdAndBeginDate := args[6]

	merchantId, err := strconv.Atoi(args[7])
	if err != nil {
		return shim.Error("1rd argument must be a numeric string")
	}
	merchantName := args[8]

	indicatorId, err := strconv.Atoi(args[9])
	if err != nil {
		return shim.Error("10rd argument must be a numeric string")
	}

	indicatorCode := args[10]
	indicatorName := args[11]

	indicatorLevel, err := strconv.Atoi(args[12])
	if err != nil {
		return shim.Error("13rd argument must be a numeric string")
	}

	indicatorParentId, err := strconv.Atoi(args[13])
	if err != nil {
		return shim.Error("14rd argument must be a numeric string")
	}

	demensionId, err := strconv.Atoi(args[14])
	if err != nil {
		return shim.Error("15rd argument must be a numeric string")
	}

	demensionCode := args[15]
	demensionName := args[16]

	dataPacketId, err := strconv.Atoi(args[17])
	if err != nil {
		return shim.Error("18rd argument must be a numeric string")
	}

	dataPacketFilePath := args[18]
	dataPacketUp2Chain := args[19]


	createDateTime, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("createTime argument must be a numeric string")
	}
	tm := time.Unix(int64(createDateTime) / 1000,0)
	dateString := tm.Format("20060102")

	merchantIdString := strconv.Itoa(merchantId)
	dataPacketIdString := strconv.Itoa(dataPacketId)
	key := merchantIdString + "_" + dateString + "_" + period + "_" + dataPacketIdString

	dataStatisBytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Failed to get datastatis: " + err.Error())
	} else if dataStatisBytes == nil {
		fmt.Println("data does not exist: " + key)
		return shim.Error("data does not exist: " + key)
	}

	// ==== Create datastatis object and marshal to JSON ====
	objectType := "datastatis"
	merchant := &merchant{merchantId, merchantName}
	indicator := &indicator{indicatorId, indicatorCode, indicatorName, indicatorLevel, indicatorParentId}
	demension := &demension{demensionId, demensionCode, demensionName}
	dataPacket := &dataPacket{dataPacketId, dataPacketFilePath, dataPacketUp2Chain}
	data := &data{objectType, id, createTime, beginData, endData, period, value1,*merchant, *indicator, *demension, *dataPacket, indicatorIdAndBeginDate}
	dataJSONBytes, err := json.Marshal(data)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(key, dataJSONBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end update datastatis")
	return shim.Success(nil)
}

func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("- start delete datastatis")
	key := args[0]

	value, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Failed to get data: " + err.Error())
	} else if value == nil {
		fmt.Println("This data does not exists: " + key)
		return shim.Error("This data does not exists: " + key)
	}

	err = stub.DelState(key)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end delete datastatis")

	return shim.Success(nil)
}

func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

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

func (t *SimpleChaincode) getHistoryByKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	key := args[0]

	fmt.Printf("- start getHistoryForMarble: %s\n", key)

	resultsIterator, err := stub.GetHistoryForKey(key)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the datastatis
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON datastatis)
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
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForMarble returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}