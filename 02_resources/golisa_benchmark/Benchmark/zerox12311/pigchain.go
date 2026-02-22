package main

// 引入相關套件
import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// 定義所需的資料格式
// Define the Smart Contract structure
type SmartContract struct {
}

/*
Define Pig structure, with 4 properties.
Structure tags are used by encoding/json library
*/
type Pig struct {
	PigId      string `json:"pigid"`
	Timestamp  string `json:"timestamp"`
	Company    string `json:"company"`
	ActionName string `json:"actionname"`
}

// 設定必須的 init 與 invoke 方法
/*
 * The Init method *
 called when the Smart Contract "pig-chaincode" is instantiated by the network
 * Best practice is to have any Ledger initialization in separate function
 -- see initLedger()
*/
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method *
 called when an application requests to run the Smart Contract "pig-chaincode"
 The app also specifies the specific smart contract function to call with args
*/
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger
	if function == "queryPig" {
		return s.queryPig(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "recordPig" {
		return s.recordPig(APIstub, args)
	} else if function == "queryPigHistory" {
		return s.queryPigHistory(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

// 生成模擬資料
/*
 * The initLedger method *
 */
func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	pig := []Pig{
		Pig{PigId: "A", Timestamp: "2018-10-25", Company: "LUX", ActionName: "FEED"},
		Pig{PigId: "B", Timestamp: "2018-10-26", Company: "NTUB", ActionName: "PLAY"},
		Pig{PigId: "C", Timestamp: "2018-10-27", Company: "TEST", ActionName: "SHOW"},
	}

	i := 0
	for i < len(pig) {
		fmt.Println("i is ", i)
		pigAsBytes, _ := json.Marshal(pig[i])
		APIstub.PutState(strconv.Itoa(i+1), pigAsBytes)
		fmt.Println("Added", pig[i])
		i = i + 1
	}

	return shim.Success([]byte("Init Success"))
}

// 查詢單隻pig
/*
 * The queryPig method *
 */
func (s *SmartContract) queryPig(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Success([]byte("Incorrect number of arguments. Expecting 1"))
	}

	pigAsBytes, _ := APIstub.GetState(args[0])
	if pigAsBytes == nil {
		return shim.Success([]byte("Could not locate pig"))
	}
	return shim.Success(pigAsBytes)
}

func (s *SmartContract) queryPigHistory(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Success([]byte("Incorrect number of arguments. Expecting 1"))
	}

	pigAsBytes, err := APIstub.GetHistoryForKey(args[0])
	if err != nil {
		fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
		return shim.Success([]byte("Failed to get asset"))
	}

	if pigAsBytes == nil {
		fmt.Errorf("Could not locate pig")
		return shim.Success([]byte("Could not locate pig"))
	}
	defer pigAsBytes.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false

	for pigAsBytes.HasNext() {
		result, err2 := pigAsBytes.Next()
		if err2 != nil {
			fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
			return shim.Error("Failed to get asset")
		}
		// value += string(result.Value) + "||"
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(result.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")

		buffer.WriteString(string(result.Value))

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(result.Timestamp.Seconds, int64(result.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}

	buffer.WriteString("]")

	fmt.Printf("- getHistoryForMarble returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// 增加新紀錄

/*
 * The recordPig method *
 */
func (s *SmartContract) recordPig(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 4 {
		return shim.Error([]byte("Incorrect number of arguments. Expecting 4"))
	}

	var pig = Pig{PigId: args[0], Company: args[1], ActionName: args[2], Timestamp: args[3]}

	pigAsBytes, _ := json.Marshal(pig)
	err := APIstub.PutState(args[0], pigAsBytes)
	if err != nil {
		return shim.Success([]byte("Failed to record pig catch:" + args[0]))
	}

	return shim.Success([]byte("Add Success"))
}

// 主程序
/*
 * main function *
calls the Start function
The main function starts the chaincode in the container during instantiation.
*/
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
