package main

import (
	"fmt"
    //"bytes"
	"encoding/json"
    //"strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	 pb "github.com/hyperledger/fabric/protos/peer"
	)

// Define the chaincode structure
type SimpleChaincode struct {
}

//Define DWPB structure 
type DWPB struct {
	UniqueID     			string    `json:"UniqueID"`
	Latest_approved_value	string  `json:"Latest_approved_value"`
	Latest_utilised_value	string   `json:"Latest_utilised_value"`
//1	Latest_remaining_value	float64   `json:"Latest_remaining_value"`
//1	ModifiedDate 			int64	  `json:"Modified_Date"` // timestamp
//1	Version					string    `json:"Version"`
//1	Expired      			bool	  `json:"Expired"`
//1	Budget_sub_status		string    `json:"Budget_sub_status"`
}

func (s *SimpleChaincode) Init (APIstub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (s *SimpleChaincode) Invoke(APIstub shim.ChaincodeStubInterface) pb.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
    if function == "createDWPB" {
	return s.createDWPB(APIstub, args)
  	} else if function == "queryDWPB" {
	return s.queryDWPB(APIstub, args)
    }
    return shim.Error("Invalid Smart Contract function name.")
}    


func (s *SimpleChaincode) createDWPB (APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
    if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 8")
	}

// populate DWPB
	var Dwpb = DWPB{UniqueID: args[0],Latest_approved_value: args[1],Latest_utilised_value: args[2]}
//	Latest_utilised_value = args[2]
//	Latest_remaining_value = args[3]
//	ModifiedDate = args[4]
//	Version = args[5]
//	Expired  = args[6]
//	Budget_sub_status = args[7]
//	}
//store state in ledger
	DWPBAsBytes,_ := json.Marshal(Dwpb)	
	err := APIstub.PutState(args[0],DWPBAsBytes)
	if err != nil {
	return shim.Error(fmt.Sprintf("Failed to create DWPB"))
	}
	return shim.Success(nil)
}
//function to update the existing DWPB

//func (s *SimpleChaincode) UpdateDWPB (APIstub shim.ChaincodeStubInterface) pb.Response {
//}
// to be written
//update for one unique ID 	

//function to query a DWPB
func (s *SimpleChaincode) queryDWPB (APIstub shim.ChaincodeStubInterface, args []string) pb.Response {
  if len(args) != 1 {
    return shim.Error("Incorrect number of arguments. Expecting UniqueID")
  }
 DWPBAsBytes, err := APIstub.GetState(args[0])
  if err != nil {
    return shim.Error(err.Error())
  }
  return shim.Success(DWPBAsBytes)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
	

