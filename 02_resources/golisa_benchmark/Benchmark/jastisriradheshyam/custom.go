package main

import (
	//"bytes"
	"encoding/json"
	"fmt"
	//"strconv"
	"strings"
	//"time"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	mspprotos "github.com/hyperledger/fabric/protos/msp"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type member struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Role    string `json:"role"`
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	creator, err := stub.GetCreator() // it'll give the certificate of the invoker
	id := &mspprotos.SerializedIdentity{}
	err = proto.Unmarshal(creator, id)
	if err != nil {
		return shim.Error(fmt.Sprintf("chaincode::AcceptLeadQuote:couldnt unmarshal creator"))
	}
	block, _ := pem.Decode(id.GetIdBytes())
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("chaincode:AcceptLeadQuote:couldnt parse certificate"))
	}
	invokerhash := sha256.Sum256([]byte(cert.Subject.CommonName + cert.Issuer.CommonName))
	insurerAddress := hex.EncodeToString(invokerhash[:])

	// Handle different functions
	if function == "addMember" {
		return t.addMember(stub, args, insurerAddress)
	} else if function == "readMember" { //read a marble
		return t.readMember(stub, insurerAddress)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

func (t *SimpleChaincode) addMember(stub shim.ChaincodeStubInterface, args []string, insurerAddress string) pb.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init marble")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(insurerAddress) <= 0 {
		return shim.Error("Invalid peer ")
	}

	clientName := args[0]
	Address := strings.ToLower(args[1])
	Role := strings.ToLower(args[2])

	insurerAsBytes, err := stub.GetState(insurerAddress)
	if err != nil {
		return shim.Error("Failed to get client: " + err.Error())
	} else if insurerAsBytes != nil {
		fmt.Println("This client already exists: " + insurerAddress)
		return shim.Error("This client already exists: " + insurerAddress)
	}

	member := &member{insurerAddress, clientName, Address, Role}
	memberJSONasBytes, err := json.Marshal(member)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(insurerAddress, memberJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("This is the adderess of invoker :", insurerAddress, "|-------------------", []byte(insurerAddress), "--------sd-s----")
	return shim.Success([]byte(insurerAddress))
}

func (t *SimpleChaincode) readMember(stub shim.ChaincodeStubInterface, arg string) pb.Response {
	var name, jsonResp string
	var err error

	if len(arg) <= 0 {
		return shim.Error("Incorrect number of arguments. Expecting name of the marble to query")
	}

	name = arg
	valAsbytes, err := stub.GetState(name)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Member does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valAsbytes)
}
