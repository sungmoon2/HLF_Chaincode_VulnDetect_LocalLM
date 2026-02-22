package main

import(
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)
var _mainLogger = shim.NewLogger("TraceabilitySmartContract")

//SmartContract represents the main entart contract

type SmartContract struct {
	traceability *Trace
}

// Init initalizes the chaincode 
func (sc *SmartContract) Init(stub shim.ChaincodeStubInterface)  pb.Response {
    _mainLogger.Infof("Inside the init method")
	sc.traceability = new(Trace)
	return shim.Success(nil) 
}

// Invoke is the entry point for any transaction

func  (sc *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	var response pb.Response
	action, _ := stub.GetFunctionAndParameters()
	switch action {
		case "partManufacturing":
			response = sc.traceability.PartManufacturing(stub)
		case "changeOwnership":
			response = sc.traceability.ChangeOwnership(stub)
		case "createCar":
			response = sc.traceability.CreateCar(stub)	
		case "changeCarOwnership":
			response = sc.traceability.ChangeCarOwnership(stub)									
		default:
			response = shim.Error("Invalid function name provided") 
	}
	return response	
}


func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
       _mainLogger.Criticalf("Error starting the chaincode: %v", err)
	}else {
		_mainLogger.Info("|| STARTING TRACEABILITY CHAINCODE ||")
	}
}
