//
// Simple Asset Chaincode
// 
package main
//
// Building Chaincode
// ------------------
// go get -u github.com/hyperledger/fabric/core/chaincode/shim
// go build
//

import(
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// simple asset implementation a simple chaincode to manage an asset
type SimpleAsset struct {

}
// initialize the Chaincode
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// get the argument from tx proposal
	args := stub.GetStringArgs()
	if len(args) != 2 {
		return shim.Error("Incorrect arguments. Expecting a key and value")
	}
	// set up any variables or assets here by calling stub.PutState()

	// we store the key and the value on the ledger
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
	}
	return shim.Success(nil)
}

// invoking the chaincode
// invoking is called per transaction on the chaincode.
// Each tx is either a 'get' or 'set' on the asset created by Init func.
// The 'set' method may create a new asset by specifying a new key/value pair.
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	var result string
	var err error

	// extract the function and args from the tx proposal
	switch fun, args := stub.GetFunctionAndParameters(); fun {
	case "set": result, err = set(stub, args)
	case "get": result, err = get(stub, args)
	default:
		result = ""
		err = fmt.Errorf("Unknown method: %s: ", fun)
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	// return the result as success payload
	return shim.Success([]byte(result))
}

// implementing the chaincode app

// SET stores the asset(both key and value) on the ladger.
// If the key exists, it will override the value with the new one
func set(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect argument. Expecting a key and a value")
	}

	if err := stub.PutState(args[0], []byte(args[1])); err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	return args[1], nil
}

// GET returns the value of the specified asset key
func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect argument. Exporting a key")
	}

	value, err := stub.GetState(args[0])
	// result processing ligic...
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s whith error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	// return success payload
	return string(value), nil
}

//
// main driver
// starts up the chaincode in the container during instantiate 
func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SmpleAsset chaincode: %s", err)
	}
}
