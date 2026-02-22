package main

import (
	"encoding/json"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// This program deals with two different assets - Batch & SKU

//SimpleChaincode is used
type SimpleChaincode struct {
}

//Batch asset
type Batch struct {
	BatchID      string `json:"batch_id"`
	BatchCreDate int    `json:"batch_creation_date"`
	BatchOwner   string `json:"batch_owner"`
}

//SKU asset
type SKU struct {
	SkuCode    string `json:"sku_code"`
	Size       string `json:"sku_size"`
	SkuCreDate int    `json:"sku_creation_date"`
}

const prefixBatch = "BATCH"
const prefiSKU = "SKU"

var logger = shim.NewLogger("SimpleChaincode")

func main() {

	// Setting up the log level programtically
	logger.SetLevel(shim.LogDebug)
	logger.Debug("Enter : main method")

	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		logger.Critical("Failed to start chaincode -", err)
	}

	logger.Debug("Exit : main method")

}

// Init method
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke method
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	function, args := stub.GetFunctionAndParameters()
	if function == "createBatch" {
		// set player - game
		return t.createBatch(stub, args)
	} else if function == "getBatchDetails" {
		// get game by player
		return t.getBatchDetails(stub, args)
	}

	return shim.Error("Invalid function name. Expecting \"set\" \"get\"")
}

//createBatch - Supporting Functions used by Invoke method
func (t *SimpleChaincode) createBatch(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	batch := Batch{}

	err := json.Unmarshal([]byte(args[0]), &batch)
	if err != nil {
		return shim.Error(err.Error())
	}

	key, err := stub.CreateCompositeKey(prefixBatch, []string{batch.BatchID})
	if err != nil {
		return shim.Error(err.Error())
	}

	batchAsBytes, err := json.Marshal(batch)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(key, batchAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)

}

//getBatchDetails - Supporting Functions used by Invoke method
func (t *SimpleChaincode) getBatchDetails(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	results := []interface{}{}
	resultsIterator, err := stub.GetStateByPartialCompositeKey(prefixBatch, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		batchResult, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		batch := Batch{}
		err = json.Unmarshal(batchResult.Value, &batch)
		if err != nil {
			return shim.Error(err.Error())
		}
		results = append(results, batch)

	}

	batchAsBytes, err := json.Marshal(results)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(batchAsBytes)
}
