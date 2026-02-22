package main
 
import (
    "bytes"
    "encoding/json"
    "fmt"
    "strconv"
    "strings"
    "time"
 
    "github.com/hyperledger/fabric/core/chaincode/shim"
    pb "github.com/hyperledger/fabric/protos/peer"
)
 
// FabricChaincode example simple Chaincode implementation
type FabricChaincode struct {
}
 
type diamond struct {
    ObjectType string `json:"docType"` 
    Name       string `json:"name"`    
    Origin      string `json:"origin"`
    Carats       int    `json:"carats"`
    Owner      string `json:"owner"`
}
 
// ===================================================================================
// Main
// ===================================================================================
func main() {
    err := shim.Start(new(FabricChaincode))
    if err != nil {
        fmt.Printf("Error starting a new instance of Diamond chaincode: %s", err)
    }
}
 
// Init initializes chaincode
// ===========================
func (t *FabricChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
    return shim.Success(nil)
}
 
// Invoke - Our entry point for Invocations
// ========================================
func (t *FabricChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
    function, args := stub.GetFunctionAndParameters()
    fmt.Println("invoke is running " + function)
 
    // Handle different functions
    if function == "createDiamond" { //create a new diamond
        return t.createDiamond(stub, args)
    } else if function == "transferDiamond" { //change owner of a specific diamond
        return t.transferDiamond(stub, args)
     } else if function == "transferdiamondsBasedOnOrigin" { //transfer all diamonds of a certain origin
        return t.transferdiamondsBasedOnOrigin(stub, args)
    } else if function == "delete" { //delete a diamond
        return t.delete(stub, args)
    } else if function == "queryDiamond" { //read a diamond
        return t.queryDiamond(stub, args)
    } else if function == "getHistoryFordiamond" { //get history of values for a diamond
        return t.getHistoryFordiamond(stub, args)
    } 
 
    fmt.Println("invoke did not find func: " + function) //error
    return shim.Error("Received unknown function invocation")
}
 
// ============================================================
// createDiamond - create a new diamond, store into chaincode state
// ============================================================
func (t *FabricChaincode) createDiamond(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    var err error
 
    if len(args) != 4 {
        return shim.Error("Incorrect number of arguments. Expecting 4")
    }
 
    // ==== Input sanitation ====
    fmt.Println("- start init diamond")
    if len(args[0]) <= 0 {
        return shim.Error("1st argument must be a non-empty string")
    }
    if len(args[1]) <= 0 {
        return shim.Error("2nd argument must be a non-empty string")
    }
    if len(args[2]) <= 0 {
        return shim.Error("3rd argument must be a non-empty string")
    }
    if len(args[3]) <= 0 {
        return shim.Error("4th argument must be a non-empty string")
    }
    diamondName := args[0]
    origin := strings.ToLower(args[1])
    owner := strings.ToLower(args[3])
    carats, err := strconv.Atoi(args[2])
    if err != nil {
        return shim.Error("3rd argument must be a numeric string")
    }
 
    // ==== Check if diamond already exists ====
    diamondAsBytes, err := stub.GetState(diamondName)
    if err != nil {
        return shim.Error("Failed to get diamond: " + err.Error())
    } else if diamondAsBytes != nil {
        fmt.Println("This diamond already exists: " + diamondName)
        return shim.Error("This diamond already exists: " + diamondName)
    }
 
    // ==== Create diamond object and marshal to JSON ====
    objectType := "diamond"
    diamond := &diamond{objectType, diamondName, origin, carats, owner}
    diamondJSONasBytes, err := json.Marshal(diamond)
    if err != nil {
        return shim.Error(err.Error())
    }
 
    // === Save diamond to state ===
    err = stub.PutState(diamondName, diamondJSONasBytes)
    if err != nil {
        return shim.Error(err.Error())
    }
 
    //  ==== Index the diamond to enable origin-based range queries, e.g. return all blue diamonds ====
    //  An 'index' is a normal key/value entry in state.
    //  The key is a composite key, with the elements that you want to range query on listed first.
    //  In our case, the composite key is based on indexName~origin~name.
    //  This will enable very efficient state range queries based on composite keys matching indexName~origin~*
    indexName := "origin~name"
    originNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{diamond.Origin, diamond.Name})
    if err != nil {
        return shim.Error(err.Error())
    }
    //  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the diamond.
    //  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
    value := []byte{0x00}
    stub.PutState(originNameIndexKey, value)
 
    // ==== Diamond saved and indexed. Return success ====
    fmt.Println("- end init diamond")
    return shim.Success(nil)
}
 
// ===============================================
// queryDiamond - read a diamond from chaincode state
// ===============================================
func (t *FabricChaincode) queryDiamond(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    var name, jsonResp string
    var err error
 
    if len(args) != 1 {
        return shim.Error("Incorrect number of arguments. Expecting name of the diamond to query")
    }
 
    name = args[0]
    valAsbytes, err := stub.GetState(name) //get the diamond from chaincode state
    if err != nil {
        jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
        return shim.Error(jsonResp)
    } else if valAsbytes == nil {
        jsonResp = "{\"Error\":\"Diamond does not exist: " + name + "\"}"
        return shim.Error(jsonResp)
    }
 
    return shim.Success(valAsbytes)
}
 
// ==================================================
// delete - remove a diamond key/value pair from state
// ==================================================
func (t *FabricChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
    var jsonResp string
    var diamondJSON diamond
    if len(args) != 1 {
        return shim.Error("Incorrect number of arguments. Expecting 1")
    }
    diamondName := args[0]
 
    // to maintain the origin~name index, we need to read the diamond first and get its origin
    valAsbytes, err := stub.GetState(diamondName) //get the diamond from chaincode state
    if err != nil {
        jsonResp = "{\"Error\":\"Failed to get state for " + diamondName + "\"}"
        return shim.Error(jsonResp)
    } else if valAsbytes == nil {
        jsonResp = "{\"Error\":\"Diamond does not exist: " + diamondName + "\"}"
        return shim.Error(jsonResp)
    }
 
    err = json.Unmarshal([]byte(valAsbytes), &diamondJSON)
    if err != nil {
        jsonResp = "{\"Error\":\"Failed to decode JSON of: " + diamondName + "\"}"
        return shim.Error(jsonResp)
    }
 
    err = stub.DelState(diamondName) //remove the diamond from chaincode state
    if err != nil {
        return shim.Error("Failed to delete state:" + err.Error())
    }
 
    // maintain the index
    indexName := "origin~name"
    originNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{diamondJSON.Origin, diamondJSON.Name})
    if err != nil {
        return shim.Error(err.Error())
    }
 
    //  Delete index entry to state.
    err = stub.DelState(originNameIndexKey)
    if err != nil {
        return shim.Error("Failed to delete state:" + err.Error())
    }
    return shim.Success(nil)
}
 
// ===========================================================
// transfer a diamond by setting a new owner name on the diamond
// ===========================================================
func (t *FabricChaincode) transferDiamond(stub shim.ChaincodeStubInterface, args []string) pb.Response {
 
    //   0       1
    // "name", "bob"
    if len(args) < 2 {
        return shim.Error("Incorrect number of arguments. Expecting 2")
    }
 
    diamondName := args[0]
    newOwner := strings.ToLower(args[1])
    fmt.Println("- start transferDiamond ", diamondName, newOwner)
 
    diamondAsBytes, err := stub.GetState(diamondName)
    if err != nil {
        return shim.Error("Failed to get diamond:" + err.Error())
    } else if diamondAsBytes == nil {
        return shim.Error("Diamond does not exist")
    }
 
    diamondToTransfer := diamond{}
    err = json.Unmarshal(diamondAsBytes, &diamondToTransfer) //unmarshal it aka JSON.parse()
    if err != nil {
        return shim.Error(err.Error())
    }
    diamondToTransfer.Owner = newOwner //change the owner
 
    diamondJSONasBytes, _ := json.Marshal(diamondToTransfer)
    err = stub.PutState(diamondName, diamondJSONasBytes) //rewrite the diamond
    if err != nil {
        return shim.Error(err.Error())
    }
 
    fmt.Println("- end transferDiamond (success)")
    return shim.Success(nil)
}
 
func (t *FabricChaincode) transferdiamondsBasedOnOrigin(stub shim.ChaincodeStubInterface, args []string) pb.Response {
 
    //   0       1
    // "origin", "bob"
    if len(args) < 2 {
        return shim.Error("Incorrect number of arguments. Expecting 2")
    }
 
    origin := args[0]
    newOwner := strings.ToLower(args[1])
    fmt.Println("- start transferdiamondsBasedOnOrigin ", origin, newOwner)
 
    // Query the origin~name index by origin
    // This will execute a key range query on all keys starting with 'origin'
    originateddiamondResultsIterator, err := stub.GetStateByPartialCompositeKey("origin~name", []string{origin})
    if err != nil {
        return shim.Error(err.Error())
    }
    defer originateddiamondResultsIterator.Close()
 
    // Iterate through result set and for each diamond found, transfer to newOwner
    var i int
    for i = 0; originateddiamondResultsIterator.HasNext(); i++ {
        // Note that we don't get the value (2nd return variable), we'll just get the diamond name from the composite key
        responseRange, err := originateddiamondResultsIterator.Next()
        if err != nil {
            return shim.Error(err.Error())
        }
 
        // get the origin and name from origin~name composite key
        objectType, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
        if err != nil {
            return shim.Error(err.Error())
        }
        returnedOrigin := compositeKeyParts[0]
        returneddiamondName := compositeKeyParts[1]
        fmt.Printf("- found a diamond from index:%s origin:%s name:%s\n", objectType, returnedOrigin, returneddiamondName)
 
        // Now call the transfer function for the found diamond.
        // Re-use the same function that is used to transfer individual diamonds
        response := t.transferDiamond(stub, []string{returneddiamondName, newOwner})
        // if the transfer failed break out of loop and return error
        if response.Status != shim.OK {
            return shim.Error("Transfer failed: " + response.Message)
        }
    }
 
    responsePayload := fmt.Sprintf("Transferred %d %s diamonds to %s", i, origin, newOwner)
    fmt.Println("- end transferdiamondsBasedOnOrigin: " + responsePayload)
    return shim.Success([]byte(responsePayload))
}
 
 
func (t *FabricChaincode) getHistoryFordiamond(stub shim.ChaincodeStubInterface, args []string) pb.Response {
 
    if len(args) < 1 {
        return shim.Error("Incorrect number of arguments. Expecting 1")
    }
 
    diamondName := args[0]
 
    fmt.Printf("- start getHistoryFordiamond: %s\n", diamondName)
 
    resultsIterator, err := stub.GetHistoryForKey(diamondName)
    if err != nil {
        return shim.Error(err.Error())
    }
    defer resultsIterator.Close()
 
    // buffer is a JSON array containing historic values for the diamond
    var buffer bytes.Buffer
    buffer.WriteString("[")
 
    bArrayMemberAlreadyWritten := false
    for resultsIterator.HasNext() {
        response, err := resultsIterator.Next()
        if err != nil {
            return shim.Error(err.Error())
        }
        // Add a comma before array members, suppress it for the first array member
        if bArrayMemberAlreadyWritten == true {
            buffer.WriteString(",")
        }
        buffer.WriteString("{\"TxId\":")
        buffer.WriteString("\"")
        buffer.WriteString(response.TxId)
        buffer.WriteString("\"")
 
        buffer.WriteString(", \"Value\":")
        // if it was a delete operation on given key, then we need to set the
        //corresponding value null. Else, we will write the response.Value
        //as-is (as the Value itself a JSON diamond)
        if response.IsDelete {
            buffer.WriteString("null")
        } else {
            buffer.WriteString(string(response.Value))
        }
 
        buffer.WriteString(", \"Timestamp\":")
        buffer.WriteString("\"")
        buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
        buffer.WriteString("\"")
 
        buffer.WriteString(", \"IsDelete\":")
        buffer.WriteString("\"")
        buffer.WriteString(strconv.FormatBool(response.IsDelete))
        buffer.WriteString("\"")
 
        buffer.WriteString("}")
        bArrayMemberAlreadyWritten = true
    }
    buffer.WriteString("]")
 
    fmt.Printf("- getHistoryFordiamond returning:\n%s\n", buffer.String())
 
    return shim.Success(buffer.Bytes())
}
