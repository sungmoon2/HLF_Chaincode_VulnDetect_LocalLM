/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"github.com/sjwhitworth/golearn/base"
	"github.com/sjwhitworth/golearn/ensemble"
	"github.com/sjwhitworth/golearn/evaluation"
	"github.com/sjwhitworth/golearn/filters"
	"github.com/sjwhitworth/golearn/trees"
	"math/rand"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// SimpleAsset implements a simple chaincode to manage an asset
type SimpleAsset struct {
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Get the args from the transaction proposal
	args := stub.GetStringArgs()
	if len(args) != 2 {
		return shim.Error("Incorrect arguments. Expecting a key and a value")
	}

	// Set up any variables or assets here by calling stub.PutState()

	// We store the key and the value on the ledger
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
	}
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error
	if fn == "set" {
		result, err = set(stub, args)
	} else { // assume 'get' even if fn is nil
		result, err = get(stub, args)
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success([]byte(result))
}

// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
func set(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	return args[1], nil
}

// Get returns the value of the specified asset key
func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {

	var tree base.Classifier

	rand.Seed(44111342)

	// Load in the iris dataset
	iris, err := base.ParseCSVToInstances("../datasets/iris_headers.csv", true)
	if err != nil {
		panic(err)
	}

	// Discretise the iris dataset with Chi-Merge
	filt := filters.NewChiMergeFilter(iris, 0.999)
	for _, a := range base.NonClassFloatAttributes(iris) {
		filt.AddAttribute(a)
	}
	filt.Train()
	irisf := base.NewLazilyFilteredInstances(iris, filt)

	// Create a 60-40 training-test split
	trainData, testData := base.InstancesTrainTestSplit(irisf, 0.60)

	//
	// First up, use ID3
	//
	tree = trees.NewID3DecisionTree(0.6)
	// (Parameter controls train-prune split.)

	// Train the ID3 tree
	err = tree.Fit(trainData)
	if err != nil {
		panic(err)
	}

	// Generate predictions
	predictions, err := tree.Predict(testData)
	if err != nil {
		panic(err)
	}

	// Evaluate
	fmt.Println("ID3 Performance (information gain)")
	cf, err := evaluation.GetConfusionMatrix(testData, predictions)
	if err != nil {
		panic(fmt.Sprintf("Unable to get confusion matrix: %s", err.Error()))
	}
	fmt.Println(evaluation.GetSummary(cf))

	tree = trees.NewID3DecisionTreeFromRule(0.6, new(trees.InformationGainRatioRuleGenerator))
	// (Parameter controls train-prune split.)

	// Train the ID3 tree
	err = tree.Fit(trainData)
	if err != nil {
		panic(err)
	}

	// Generate predictions
	predictions, err = tree.Predict(testData)
	if err != nil {
		panic(err)
	}

	// Evaluate
	fmt.Println("ID3 Performance (information gain ratio)")
	cf, err = evaluation.GetConfusionMatrix(testData, predictions)
	if err != nil {
		panic(fmt.Sprintf("Unable to get confusion matrix: %s", err.Error()))
	}
	fmt.Println(evaluation.GetSummary(cf))

	tree = trees.NewID3DecisionTreeFromRule(0.6, new(trees.GiniCoefficientRuleGenerator))
	// (Parameter controls train-prune split.)

	// Train the ID3 tree
	err = tree.Fit(trainData)
	if err != nil {
		panic(err)
	}

	// Generate predictions
	predictions, err = tree.Predict(testData)
	if err != nil {
		panic(err)
	}

	// Evaluate
	fmt.Println("ID3 Performance (gini index generator)")
	cf, err = evaluation.GetConfusionMatrix(testData, predictions)
	if err != nil {
		panic(fmt.Sprintf("Unable to get confusion matrix: %s", err.Error()))
	}
	fmt.Println(evaluation.GetSummary(cf))
	//
	// Next up, Random Trees
	//

	// Consider two randomly-chosen attributes
	tree = trees.NewRandomTree(2)
	err = tree.Fit(trainData)
	if err != nil {
		panic(err)
	}
	predictions, err = tree.Predict(testData)
	if err != nil {
		panic(err)
	}
	fmt.Println("RandomTree Performance")
	cf, err = evaluation.GetConfusionMatrix(testData, predictions)
	if err != nil {
		panic(fmt.Sprintf("Unable to get confusion matrix: %s", err.Error()))
	}
	fmt.Println(evaluation.GetSummary(cf))

	//
	// Finally, Random Forests
	//
	tree = ensemble.NewRandomForest(70, 3)
	err = tree.Fit(trainData)
	if err != nil {
		panic(err)
	}
	predictions, err = tree.Predict(testData)
	if err != nil {
		panic(err)
	}
	fmt.Println("RandomForest Performance")
	cf, err = evaluation.GetConfusionMatrix(testData, predictions)
	if err != nil {
		panic(fmt.Sprintf("Unable to get confusion matrix: %s", err.Error()))
	}
	fmt.Println(evaluation.GetSummary(cf))

	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	// value, err := stub.GetState(args[0])
	// if err != nil {
	// 	return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	// }
	// if value == nil {
	// 	return "", fmt.Errorf("Asset not found: %s", args[0])
	// }
	// err := stub.PutState(args[0], []byte(string(value)))

	return string(evaluation.GetSummary(cf)), nil
	
}
// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
