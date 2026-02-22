package main

import (
	"encoding/json"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"
	"github.com/hyperledger/fabric/protos/peer"
)

type myContract struct {
}

func Success(rc int32, doc string, payload []byte) peer.Response {
	return peer.Response{
		Status:  rc,
		Message: doc,
		Payload: payload,
	}
}

func Error(rc int32, doc string) peer.Response {
	return peer.Response{
		Status:  rc,
		Message: doc,
	}
}

func BytesToString(data []byte) string {
	return string(data[:])
}

type myCity struct {
	name   string
	people int
	anthem string
}

type Phones struct {
	name  string
	phone int
}

func main() {
	shim.Start(new(myContract))
}

func (ptr *myContract) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (ptr *myContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	case "get_close":
		return ptr.get_close(stub, args)
	case "set_close":
		return ptr.set_close(stub, args)
	case "delete_close":
		return ptr.delete_close(stub, args)
	case "set_open":
		return ptr.set_open(stub, args)
	case "get_open_myCity":
		return ptr.get_open_myCity(stub, args)
	case "set_close_myCity":
		return ptr.set_close_myCity(stub, args)
	case "delete_open_myCity":
		return ptr.delete_open_myCity(stub, args)
	case "get_close_Phones":
		return ptr.get_close_Phones(stub, args)
	case "set_close_Phones":
		return ptr.set_close_Phones(stub, args)
	case "delete_open_Phones":
		return ptr.delete_open_Phones(stub, args)
	default:
		return shim.Error("Invalid function name")
	}
}

func (ptr *myContract) get_close(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	owner_id, err := stub.GetCreator()
	this_id, err := cid.GetID(stub)
	if this_id != BytesToString(owner_id) {
		return Error(403, "Access to this method is only for creator")
	}
	if len(args) != 1 {
		return Error(400, "Incorrect number of arguments. Expecting 1")
	}
	var key = args[0]
	var valAsbytes []byte
	valAsbytes, err = stub.GetState(key)
	if err != nil {
		return Error(404, "Not Found")
	} else if valAsbytes == nil {
		return Error(404, "Not Found")
	}
	return peer.Response{Status: 200, Message: "OK", Payload: valAsbytes}
}

func (ptr *myContract) set_close(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	owner_id, err := stub.GetCreator()
	this_id, err := cid.GetID(stub)
	if this_id != BytesToString(owner_id) {
		return Error(403, "Access to this method is only for creator")
	}
	if len(args) != 2 {
		return Error(401, "Incorrect number of arguments. Expecting 2")
	}
	var key = args[0]
	var data = args[1]
	err = stub.PutState(key, []byte(data))
	if err != nil {
		return Error(500, "Error")
	}
	return Success(201, "Created", nil)
}

func (ptr *myContract) delete_close(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	owner_id, err := stub.GetCreator()
	this_id, err := cid.GetID(stub)
	if this_id != BytesToString(owner_id) {
		return Error(403, "Access to this method is only for creator")
	}
	if len(args) != 1 {
		return Error(400, "Incorrect number of arguments. Expecting 1")
	}
	var key = args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		return Error(404, "Not Found")
	} else if valAsbytes == nil {
		return Error(404, "Not Found")
	}
	err = stub.DelState(key)
	if err != nil {
		return Error(500, "Error")
	}
	return Success(202, "Deleted", nil)
}

func (ptr *myContract) set_open(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error
	if len(args) != 2 {
		return Error(401, "Incorrect number of arguments. Expecting 2")
	}
	var key = args[0]
	var data = args[1]
	err = stub.PutState(key, []byte(data))
	if err != nil {
		return Error(500, "Error")
	}
	return Success(201, "Created", nil)
}

func (ptr *myContract) get_open_myCity(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error
	var valAsbytes []byte
	if len(args) != 1 {
		return Error(400, "Incorrect number of arguments. Expecting 1")
	}
	var key = args[0]
	valAsbytes, err = stub.GetState(key)
	if err != nil {
		return Error(404, "Not Found")
	} else if valAsbytes == nil {
		return Error(404, "Not Found")
	}
	myCity_JSON := myCity{}
	valAsbytes, err = json.Marshal(myCity_JSON)
	return Success(200, "OK", valAsbytes)
}

func (ptr *myContract) set_close_myCity(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	owner_id, err := stub.GetCreator()
	this_id, err := cid.GetID(stub)
	if this_id != BytesToString(owner_id) {
		return Error(403, "Access to this method is only for creator")
	}
	var name string
	var people int
	var anthem string
	if len(args) != 4 {
		return Error(401, "Incorrect number of arguments. Expecting 3")
	}
	var key = args[0]
	name = args[1]
	people, err = strconv.Atoi(args[2])
	anthem = args[3]
	myCity := &myCity{name, people, anthem}
	myCityJSONasBytes, err := json.Marshal(myCity)
	if err != nil {
		return Error(402, "Trouble while making JSON")
	}
	err = stub.PutState(key, myCityJSONasBytes)
	if err != nil {
		return Error(500, "Error")
	}
	return Success(201, "Created", nil)
}

func (ptr *myContract) delete_open_myCity(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error
	var valAsbytes []byte
	var myCity_JSON myCity
	if len(args) != 1 {
		return Error(400, "Incorrect number of arguments. Expecting 1")
	}
	var key = args[0]
	valAsbytes, err = stub.GetState(key)
	if err != nil {
		return Error(404, "Not Found")
	} else if valAsbytes == nil {
		return Error(404, "Not Found")
	}
	err = json.Unmarshal([]byte(valAsbytes), &myCity_JSON)
	if err != nil {
		return Error(403, "JSON Is Not Valid")
	}
	err = stub.DelState(key)
	if err != nil {
		return Error(500, "Error")
	}
	return Success(202, "Deleted", nil)
}

func (ptr *myContract) get_close_Phones(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	owner_id, err := stub.GetCreator()
	this_id, err := cid.GetID(stub)
	if this_id != BytesToString(owner_id) {
		return Error(403, "Access to this method is only for creator")
	}
	var valAsbytes []byte
	if len(args) != 1 {
		return Error(400, "Incorrect number of arguments. Expecting 1")
	}
	var key = args[0]
	valAsbytes, err = stub.GetState(key)
	if err != nil {
		return Error(404, "Not Found")
	} else if valAsbytes == nil {
		return Error(404, "Not Found")
	}
	Phones_JSON := Phones{}
	valAsbytes, err = json.Marshal(Phones_JSON)
	return Success(200, "OK", valAsbytes)
}

func (ptr *myContract) set_close_Phones(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	owner_id, err := stub.GetCreator()
	this_id, err := cid.GetID(stub)
	if this_id != BytesToString(owner_id) {
		return Error(403, "Access to this method is only for creator")
	}
	var name string
	var phone int
	if len(args) != 3 {
		return Error(401, "Incorrect number of arguments. Expecting 2")
	}
	var key = args[0]
	name = args[1]
	phone, err = strconv.Atoi(args[2])
	Phones := &Phones{name, phone}
	PhonesJSONasBytes, err := json.Marshal(Phones)
	if err != nil {
		return Error(402, "Trouble while making JSON")
	}
	err = stub.PutState(key, PhonesJSONasBytes)
	if err != nil {
		return Error(500, "Error")
	}
	return Success(201, "Created", nil)
}

func (ptr *myContract) delete_open_Phones(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	var err error
	var valAsbytes []byte
	var Phones_JSON Phones
	if len(args) != 1 {
		return Error(400, "Incorrect number of arguments. Expecting 1")
	}
	var key = args[0]
	valAsbytes, err = stub.GetState(key)
	if err != nil {
		return Error(404, "Not Found")
	} else if valAsbytes == nil {
		return Error(404, "Not Found")
	}
	err = json.Unmarshal([]byte(valAsbytes), &Phones_JSON)
	if err != nil {
		return Error(403, "JSON Is Not Valid")
	}
	err = stub.DelState(key)
	if err != nil {
		return Error(500, "Error")
	}
	return Success(202, "Deleted", nil)
}
