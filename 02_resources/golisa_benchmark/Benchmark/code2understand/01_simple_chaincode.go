package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)



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

//SimpleChaincode is used
type SimpleChaincode struct {
}

// Init method
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	var name, game string
	_, args := stub.GetFunctionAndParameters()
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	name = args[0]
	game = args[1]
	fmt.Printf("name = %s, game = %sn", name, game)

	err := stub.PutState(name, []byte(game))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Invoke method
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	function, args := stub.GetFunctionAndParameters()
	if function == "set" {
		// set player - game
		return t.set(stub, args)
	} else if function == "get" {
		// get game by player
		return t.get(stub, args)
	}
	return shim.Error("Invalid function name. Expecting \"set\" \"get\"")
}

//setGame - Supporting Functions used by Invoke method
func (t *SimpleChaincode) set(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	name := args[0]
	game := args[1]

	err := stub.PutState(name, []byte(game))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

//getGameByName - Supporting Functions used by Invoke method
func (t *SimpleChaincode) get(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	game, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(`{"error":"Name does not exists"}`)
	}
	return shim.Success(game)

}
