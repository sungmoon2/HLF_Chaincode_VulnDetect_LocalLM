package main

/* Imports
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

// Define the Asset structure, with 4 properties.  Structure tags are used by encoding/json library
type Asset struct {
	AssetHandler string `json:"AssetHandler"`
	AssetRaw string `json:"AssetRaw"`
	Title  string `json:"Title"`
	V_No string `json:"V_No"`
	ReleaseTime  string `json:"ReleaseTime"`
	Author string `json:"Author"`
	ApprovalStatus string `json:"ApprovalStatus"`
	Approver string `json:"Approver"`


}

/*
 * The Init method is called when the Smart Contract "fabAsset" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "fabAsset"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryAsset" {
		return s.queryAsset(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "putAsset" {
		return s.putAsset(APIstub, args)
	} else if function == "queryAllAsset" {
		return s.queryAllAsset(APIstub)
	} else if function == "changeAssetGrade" {
		return s.changeAssetGrade(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryAsset(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	AssetAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(AssetAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	Asset := []Asset{
		Asset{AssetHandler:"LMN",AssetRaw:"PPP",Title: "Test1", V_No: "1", ReleaseTime: "11/05/2019", Author: "Joses",ApprovalStatus:"Pending approval",Approver:"Dr.Yue"},
	}

	i := 0
	for i < len(Asset) {
		fmt.Println("i is ", i)
		AssetAsBytes, _ := json.Marshal(Asset[i])
		APIstub.PutState("Asset"+strconv.Itoa(i), AssetAsBytes)
		fmt.Println("Added", Asset[i])
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) putAsset(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 9 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	var Asset = Asset{AssetHandler: args[1], AssetRaw: args[2], Title: args[3], V_No: args[4],ReleaseTime: args[5],Author: args[6],ApprovalStatus: args[7],Approver: args[8]}

	AssetAsBytes, _ := json.Marshal(Asset)
	APIstub.PutState(args[0], AssetAsBytes)

	return shim.Success(nil)
}



func (s *SmartContract) queryAllAsset(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := ""
	endKey := ""

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
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

	fmt.Printf("- queryAllAsset:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) changeAssetGrade(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	AssetAsBytes, _ := APIstub.GetState(args[0])
	Asset := Asset{}

	json.Unmarshal(AssetAsBytes, &Asset)
	Asset.ApprovalStatus = args[1]

	AssetAsBytes, _ = json.Marshal(Asset)
	APIstub.PutState(args[0], AssetAsBytes)

	return shim.Success(nil)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
