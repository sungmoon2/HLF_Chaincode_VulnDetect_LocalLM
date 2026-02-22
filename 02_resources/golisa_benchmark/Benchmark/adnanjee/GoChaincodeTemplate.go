package main

import "fmt"
import "github.com/hyperledger/fabric/core/chaincode/shim" //The shim package
import "github.com/hyperledger/fabric/protos/peer" //peer.response is in the peer package

//Create an instance of the Logger

const ChaincodeName = "GoChaincodeTemplate"
var logger = shim.NewLogger(ChaincodeName)

//Tokenchaincode represent our chaincode object
type TokenChaincode struct {

}


//Init Implemets the Init method
func (token *TokenChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response{
	fmt.Println("Init Executed")
	
	logger.Debug("Init executed - DEBUG") //debugging message outputs
	return shim.Success(nil) //return success
}

// Invoke method
func (token *TokenChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response{
	fmt.Println("Invoke succeed")

	//emit four types of meassages for logging
	logger.Debug("DEBUG= Invoke Executed")
	logger.Info("INFO: Invoke Executed")
	logger.Noticef("NOTICE: format string value=%s", "Invoke Executed")
	logger.Warning("WARNING=Invoke Executed", "[any number of parameters of different types")
	
	//return success function
	payload := []byte("This the payload")
	return shim.Success(payload)

	//Below Function is the instance of response reporting in case of response !=200
//	return peer.Response{Status:401, Message: "Unauthorized", Payload: payload}
}

func main(){
	fmt.Println("Started Chaincode")

	err:= shim.Start(new(TokenChaincode))
	if err != nil{
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
