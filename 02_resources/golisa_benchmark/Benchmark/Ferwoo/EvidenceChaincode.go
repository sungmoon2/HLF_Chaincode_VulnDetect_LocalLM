package main

import (
	_ "bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type EvidenceChaincode struct{}

type Evidence struct {
	ID      string `json:"id"`
	Hash    string `json:"hash"`
	Time    string `json:"time"`
}

type ResInfo struct {
	Status bool   `json:"status"`
	Msg    string `json:"msg"`
}

func (t *ResInfo) error(msg string) {
	t.Status = false
	t.Msg = msg
}
func (t *ResInfo) ok(msg string) {
	t.Status = true
	t.Msg = msg
}

func (t *ResInfo) response() pb.Response {
	resJson, err := json.Marshal(t)
	if err != nil {
		return shim.Error("Failed to generate json result " + err.Error())
	}
	return shim.Success(resJson)
}

func main() {
	err := shim.Start(new(EvidenceChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

type process func(shim.ChaincodeStubInterface, []string) *ResInfo

func (t *EvidenceChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *EvidenceChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "newEvidence" {
		return t.newEvidence(stub, args)
	} else if function == "queryEvidence" {
		return t.queryEvidence(stub, args)
	}
	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// 写入证据
func (e *EvidenceChaincode) newEvidence(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return e.handleProcess(stub, args, 3, func(shim.ChaincodeStubInterface, []string) *ResInfo {
		ri := &ResInfo{true, ""}
		_id := args[0]
		_hash := args[1]
		_time := args[2]
		_evidence := &Evidence{_id,_hash,_time}
		_ejson, err := json.Marshal(_evidence)

		if err != nil {
			ri.error(err.Error())
		} else {
			_old, err := stub.GetState(_id)
			if err != nil {
				ri.error(err.Error())
			} else if _old != nil {
				ri.error("the evidence has exists")
			} else {
				err := stub.PutState(_id, _ejson)
				if err != nil {
					ri.error(err.Error())
				} else {
					ri.ok("")
				}
			}
		}
		return ri
	})
}

func (e *EvidenceChaincode) queryEvidence(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return e.handleProcess(stub, args, 1, func(shim.ChaincodeStubInterface, []string) *ResInfo {
		ri := &ResInfo{true, ""}
		queryString := args[0]
		queryResults, err := stub.GetState(queryString)
		if err != nil {
			ri.error(err.Error())
		} else {
			ri.ok(string(queryResults))
		}
		return ri
	})
}

func (t *EvidenceChaincode) handleProcess(stub shim.ChaincodeStubInterface, args []string, expectNum int, f process) pb.Response {
	res := &ResInfo{false, ""}
	err := t.checkArgs(args, expectNum)
	if err != nil {
		res.error(err.Error())
	} else {
		res = f(stub, args)
	}
	return res.response()
}

func (t *EvidenceChaincode) checkArgs(args []string, expectNum int) error {
	if len(args) != expectNum {
		return fmt.Errorf("Incorrect number of arguments. Expecting  " + strconv.Itoa(expectNum))
	}
	return nil
}