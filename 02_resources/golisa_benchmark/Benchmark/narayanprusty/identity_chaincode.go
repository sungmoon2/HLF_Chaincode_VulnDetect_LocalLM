package main

import (
	"fmt"
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/msp"
	"github.com/golang/protobuf/proto"
)

type User struct {
	PublicKey	string `json:"publicKey"`
	MetadataHash string `json:"metadataHash"`
	Permissions []string `json:"permissions"`
}

type ServiceProvider struct {
	Name	string `json:"name"`
	PublicKey string `json:"publicKey"`
}

type IdentityChaincode struct {
}

func (t *IdentityChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	var err error
	var identity []byte

	identity, err = stub.GetCreator()
	
	if err != nil {
		return shim.Error("An error occured")
	}

	sId := &msp.SerializedIdentity{}
	err = proto.Unmarshal(identity, sId)
	
	if err != nil {
			return shim.Error("An error occured")
	}

	nodeId := sId.Mspid
	err = stub.PutState("identityAuthority", []byte(nodeId))

	if err != nil {
		return shim.Error("An error occured")
	}

	return shim.Success(nil)
}

func (t *IdentityChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "getCreatorIdentity" {
		return t.getCreatorIdentity(stub, args)
	} else if function == "issueIdentity" {
		return t.issueIdentity(stub, args)
	} else if function == "getIdentity" {
		return t.getIdentity(stub, args)
	} else if function == "addServiceProvider" {
		return t.addServiceProvider(stub, args)
	} else if function == "getServiceProvider" {
		return t.getServiceProvider(stub, args)
	}

	return shim.Error("Invalid function name: " + function)
}

func (t *IdentityChaincode) getCreatorIdentity(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	identity, err := stub.GetState("identityAuthority")

	if err != nil {
		return shim.Error(err.Error())
	}

	if identity == nil {
		return shim.Error("Identity not yet stored")
	}

	return shim.Success(identity)
}

func (t *IdentityChaincode) issueIdentity(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments.")
	}

	var err error

	identityAuthority, err := stub.GetState("identityAuthority")

	if err != nil {
		return shim.Error(err.Error())
	}

	identity, err := stub.GetCreator()

	if err != nil {
		return shim.Error(err.Error())
	}

	sId := &msp.SerializedIdentity{}
	err = proto.Unmarshal(identity, sId)
	
	if err != nil {
			return shim.Error(err.Error())
	}

	nodeId := sId.Mspid

	if string(identityAuthority) != nodeId {
		return shim.Error("You are not authorized")
	}

	userExists, err := stub.GetState("user_" + args[0])

	if userExists != nil  {
		return shim.Error("User already exists")
	}

	var newUser User
	newUser.PublicKey = args[1]
	newUser.MetadataHash = args[2]

	newUserJson, err := json.Marshal(newUser)

	if err != nil {
			return shim.Error(err.Error())
	}

	err = stub.PutState("user_" + args[0], newUserJson)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *IdentityChaincode) getIdentity(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments.")
	}

	user, err := stub.GetState("user_" + args[0])

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(user)
}

func (t *IdentityChaincode) updateUserMetadataHash(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments.")
	}

	var err error

	identityAuthority, err := stub.GetState("identityAuthority")

	if err != nil {
		return shim.Error(err.Error())
	}

	identity, err := stub.GetCreator()

	if err != nil {
		return shim.Error(err.Error())
	}

	sId := &msp.SerializedIdentity{}
	err = proto.Unmarshal(identity, sId)
	
	if err != nil {
			return shim.Error(err.Error())
	}

	nodeId := sId.Mspid

	if string(identityAuthority) != nodeId {
		return shim.Error("You are not authorized")
	}

	user, err := stub.GetState("user_" + args[0])

	var userStruct User
	err = json.Unmarshal(user, &userStruct)

	if err != nil {
			return shim.Error(err.Error())
	}

	userStruct.MetadataHash = args[1]

	userJson, err := json.Marshal(userStruct)

	err = stub.PutState("user_" + args[0], userJson)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *IdentityChaincode) addServiceProvider(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments.")
	}

	var err error

	identityAuthority, err := stub.GetState("identityAuthority")

	if err != nil {
		return shim.Error(err.Error())
	}

	identity, err := stub.GetCreator()

	if err != nil {
		return shim.Error(err.Error())
	}

	sId := &msp.SerializedIdentity{}
	err = proto.Unmarshal(identity, sId)
	
	if err != nil {
			return shim.Error(err.Error())
	}

	nodeId := sId.Mspid

	if string(identityAuthority) != nodeId {
		return shim.Error("You are not authorized")
	}

	var newSP ServiceProvider
	newSP.Name = args[1]
	newSP.PublicKey = args[2]

	newSPJson, err := json.Marshal(newSP)

	if err != nil {
			return shim.Error(err.Error())
	}

	err = stub.PutState("sp_" + args[0], newSPJson)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *IdentityChaincode) getServiceProvider(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments.")
	}

	sp, err := stub.GetState("sp_" + args[0])

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(sp)
}

func main() {
	err := shim.Start(new(IdentityChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
