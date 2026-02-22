package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("KWO-CC")

type KilloWattRecords struct {
}

type RecordModel struct {
	RID        string `json:"rid"`
	ProducerID string `json:"producer_id"`
	ConsumerID string `json:"consumer_id"`
	Kw         string `json:"kw"`
	Amount     string `json:"amount"`
	Timestamp  int    `json:"timestamp"`
}

func (t *KilloWattRecords) createRecord(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("[INOVKE] createRecord")

	if len(args) != 2 {
		shim.Error("[ERROR] The number of arguments should be 2")
	}

	searchKey := args[0]
	var newRecord RecordModel
	recordByte, err := stub.GetState(searchKey)

	if err != nil || recordByte != nil {
		return shim.Error("[ERROR] RecordID has already existed")
	}
	json.Unmarshal([]byte(args[1]), &newRecord)
	newRecordAsBytes, err := json.Marshal(newRecord)

	stub.PutState(searchKey, newRecordAsBytes)
	logger.Infof("A new Record added with Key: %s, Data: %s", searchKey, string(newRecordAsBytes))
	return shim.Success(nil)
}

func (t *KilloWattRecords) queryRecord(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("[INOVKE] queryRecord")
	if len(args) != 1 {
		shim.Error("[ERROR] The number of arguments should be 1")
	}

	searchKey := args[0]
	recordByte, err := stub.GetState(searchKey)

	if err != nil || recordByte == nil {
		return shim.Error("[ERROR] Cannot find record")
	}

	logger.Infof("Search Key: %s, Data: %s", searchKey, recordByte)
	return shim.Success(recordByte)

}

func (t *KilloWattRecords) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("Init KilloWattRecords")
	return shim.Success(nil)
}

func (t *KilloWattRecords) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "createRecord" {
		return t.createRecord(stub, args)
	} else if function == "queryRecord" {
		return t.queryRecord(stub, args)
	}

	return shim.Error("INVOKE FAIL: The function name passed does not exist")
}

func main() {
	err := shim.Start(new(KilloWattRecords))
	if err != nil {
		fmt.Printf("Error starting KilloWattRecords: %s", err)
	}
}
