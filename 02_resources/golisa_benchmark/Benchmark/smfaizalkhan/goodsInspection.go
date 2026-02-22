package faizal

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type SampleChainCode struct {
}
type GoodsInpection struct {
	GoodsCode        string `json:"goodscode"`
	GoodsDescription string `json:"goodsdescription"`
	InvoiceNo        string `json:"invoiceno"`
}

func (s *SampleChainCode) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	fmt.Println("In Init")
	return shim.Success(nil)
}

func (s *SampleChainCode) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()

	fmt.Println("function is", function)
	fmt.Println("args", len(args))
	// Route to the appropriate handler function to interact with the ledger appropriately

	if function == "createAsset" {
		return s.createAsset(APIstub, args)
	}
	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SampleChainCode) createAsset(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	fmt.Println("In Process of creating assets")

	for _, v := range args {

		fmt.Println("Value is", v)
	}

	var goodsInpection = GoodsInpection{GoodsCode: args[1], GoodsDescription: args[2], InvoiceNo: args[3]}

	goodsInpectionasBytes, _ := json.Marshal(goodsInpection)
	APIstub.PutState(args[0], goodsInpectionasBytes)
	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(SampleChainCode))
	if err == nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
