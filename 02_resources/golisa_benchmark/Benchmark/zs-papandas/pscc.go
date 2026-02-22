package main

import (
    "fmt"

    "github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	
	"github.com/zs-papandas/serialization/contracts"
)

var logger = shim.NewLogger("main")

var accountContract = new(contracts.AccountContract)
var productContract = new(contracts.ProductContract)
var historyContract = new(contracts.HistoryContract)
var generateProductContract = new(contracts.GenerateProductContract)

// EntryPoint implements a simple chaincode to manage an asset
type EntryPoint struct {
}


// Init is called during chaincode instantiation to initialize any data.
func (t *EntryPoint) Init(APIstub shim.ChaincodeStubInterface) peer.Response {
	logger.Info("Instantiated chaincode!!!")
	return shim.Success(nil)
}


// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The 'set'
// method may create a new asset by specifying a new key-value pair.
func (t *EntryPoint) Invoke(APIstub shim.ChaincodeStubInterface) peer.Response {
	logger.Info("Invoked chaincode!!!")

	function, args := APIstub.GetFunctionAndParameters()

	logger.Info("Query Function: %s\n", function)

	switch function {
		// not in use
		case "listAccount":
			return accountContract.ListAccount(APIstub, args)
		case "createAccount":
			return accountContract.CreateAccount(APIstub, args)
		case "retrieveAccount":
			return accountContract.RetrieveAccount(APIstub, args)
		case "createProduct":
			return productContract.CreateProduct(APIstub, args)
		case "generateProduct":
			return generateProductContract.CreateProduct(APIstub, args)
		case "retrieveProduct":
			return productContract.RetrieveProduct(APIstub, args)
		case "changeOwner":
			return productContract.ChangeOwner(APIstub, args)
		case "listProductHistory":
			return historyContract.ListHistory(APIstub, args)
		case "changeChildOwner":
			return productContract.TestQueryInfo(APIstub, args)
		
	}

	msg := fmt.Sprintf("No such function. function = %s, args = %s", function, args)
	logger.Error(msg)
	return shim.Error(msg)
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(EntryPoint)); err != nil {
		logger.Errorf("Error creating new Chaincode. Error = %s\n", err)
	}
}
