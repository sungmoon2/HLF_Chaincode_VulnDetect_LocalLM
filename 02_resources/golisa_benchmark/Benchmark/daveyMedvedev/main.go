/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var logger = shim.NewLogger("lenovosupplychain")

func init() {
	logger.SetLevel(shim.LogDebug)
}

func (t *LenovoChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### Init ###########")
	t.initMaps()
	//TODO: Save chaincode version from args?
	_, args := stub.GetFunctionAndParameters()
	var err error
	version := args[0]
	logger.Infof("Chaincode Version = %s\n", version)
	// Write the cc version to the ledger
	err = stub.PutState(VERSION, []byte(version))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *LenovoChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########### Invoke/Query ###########")

	function, args := stub.GetFunctionAndParameters()
	f, ok := t.funcMap[function]
	if ok {
		return f(stub, args)
	} else {
		logger.Errorf("Invalid function name %s", function)
		return shim.Error(fmt.Sprintf("Invalid function %s", function))
	}

}

func main() {
	err := shim.Start(new(LenovoChainCode))
	if err != nil {
		logger.Errorf("Error starting Simple chaincode: %s", err)
	}
}
