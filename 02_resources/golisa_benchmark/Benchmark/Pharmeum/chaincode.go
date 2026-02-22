// Copyright 2019 Pharmeum
// Developed by Highchain Software https://highchain.io/
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ------------------------------------------------------------------------------
package payment

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

const defaultAmountOfPHRMTokens = "100.00"

//Chaincode payment chaincode implementation
type Chaincode struct {
	logger *shim.ChaincodeLogger
}

//Init do initialization of chaincode structure
func (c *Chaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	c.logger = shim.NewLogger("payment-chaincode")
	defer c.logger.Info("successfully initialized chaincode")

	if err  := stub.PutState("init", []byte(`{"init": true}`)); err != nil{
		c.logger.Error("payment init: failed to put initial transaction")
		return peer.Response{Status:shim.ERROR}
	}

	return shim.Success([]byte("successfully initialized chaincode"))
}

//Invoke actions defined in chaincode
func (c *Chaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Retrieve the requested chaincode function and arguments
	function, _ := stub.GetFunctionAndParameters()

	c.logger.Info("invocation request started, method called = ", function)
	var response peer.Response

	startedTime := time.Now()
	defer func(funcName string) {
		c.logger.Infof("invocation request finished %s, status %d, duration %s", function, response.Status, time.Since(startedTime).String())
	}(function)

	switch function {
	case createWallet:
		response = c.createWallet(stub)
	case transferPayment:
		response = c.transferPayment(stub)
	default:
		c.logger.Errorf("no such function %s, invocation rejected", function)
		response = peer.Response{
			Status:  shim.ERRORTHRESHOLD,
			Message: fmt.Sprintf("no such function %s, invocation rejected", function),
		}

		return response
	}

	return response
}

func (c *Chaincode) createWallet(stub shim.ChaincodeStubInterface) peer.Response {
	//1 - function call, 2 - wallet address
	args := stub.GetStringArgs()
	if len(args) != 2 {
		c.logger.Errorf("creat wallet action: invalid amount of arguments. want 2, got %d", len(args))
		return peer.Response{
			Status:  shim.ERRORTHRESHOLD,
			Message: errInvalidArguments.Error(),
		}
	}

	bytes, err := json.Marshal(&wallet{
		Balance: defaultAmountOfPHRMTokens,
	})
	if err != nil {
		c.logger.Error("failed to serialize state to json format", err)
		return peer.Response{
			Status:  shim.ERROR,
			Message: "failed to serialize state to json format",
		}
	}

	if err := stub.PutState(string(args[1]), bytes); err != nil {
		c.logger.Error("failed to create new wallet", err)
		return peer.Response{
			Status:  shim.ERROR,
			Message: errors.Wrap(err, "failed to create wallet").Error(),
		}
	}

	return peer.Response{
		Status: shim.OK,
	}
}

func (c *Chaincode) transferPayment(stub shim.ChaincodeStubInterface) peer.Response {
	//1 - function name, 2 - sender wallet, 3 - receiver wallet, 4 - amount
	args := stub.GetStringArgs()
	if len(args) != 4 {
		c.logger.Errorf("transfer payment action: invalid amount of arguments. want 4, got %d", len(args))
		return peer.Response{
			Status:  shim.ERRORTHRESHOLD,
			Message: errInvalidArguments.Error(),
		}
	}

	var (
		senderAddress   = args[1]
		receiverAddress = args[2]
	)

	senderWalletState, err := stub.GetState(senderAddress)
	if err != nil {
		c.logger.Errorf("transfer payment: failed to get sender wallet state %s, error: %s", args[1], err.Error())
		return peer.Response{
			Status:  shim.ERRORTHRESHOLD,
			Message: fmt.Sprintf("transfer payment: failed to get sender wallet state %s, error: %s", args[1], err.Error()),
		}
	}

	receiverWalletState, err := stub.GetState(receiverAddress)
	if err != nil {
		c.logger.Errorf("transfer payment: failed to get receiver wallet state %s, error: %s", args[1], err.Error())
		return peer.Response{
			Status:  shim.ERRORTHRESHOLD,
			Message: fmt.Sprintf("transfer payment: failed to get receiver wallet state %s, error: %s", args[1], err.Error()),
		}
	}

	amount, err := decimal.NewFromString(args[3])
	if err != nil {
		c.logger.Error("transfer payment: can't convert requested amount to decimal format: %s", err)
		return peer.Response{
			Status:  shim.ERRORTHRESHOLD,
			Message: fmt.Sprintf("transfer payment: can't convert requested amount to decimal format: %s", err),
		}
	}

	senderWallet := &wallet{}
	if err := json.Unmarshal(senderWalletState, senderWallet); err != nil {
		c.logger.Error("transfer payment: failed to deserialize sender wallet state ", err)
		return peer.Response{
			Status:  shim.ERROR,
			Message: fmt.Sprintf("transfer payment: failed to deserialize sender wallet state, error: %s", err),
		}
	}

	receiverWallet := &wallet{}
	if err := json.Unmarshal(receiverWalletState, receiverWallet); err != nil {
		c.logger.Error("transfer payment: failed to deserialize receiver wallet state ", err)
		return peer.Response{
			Status:  shim.ERROR,
			Message: fmt.Sprintf("transfer payment: failed to deserialize receiver wallet state, error: %s", err),
		}
	}

	senderBalance, err := decimal.NewFromString(senderWallet.Balance)
	if err != nil {
		c.logger.Error("transfer payment: can't convert sender balance to decimal format: %s", err)
		return peer.Response{
			Status:  shim.ERROR,
			Message: fmt.Sprintf("transfer payment: can't convert sender balance to decimal format: %s", err),
		}
	}

	receiverBalance, err := decimal.NewFromString(receiverWallet.Balance)
	if err != nil {
		c.logger.Error("transfer payment: can't convert receiver balance to decimal format: %s", err)
		return peer.Response{
			Status:  shim.ERROR,
			Message: fmt.Sprintf("transfer payment: can't convert receiver balance to decimal format: %s", err),
		}
	}

	if senderBalance.Equal(decimal.Zero) {
		c.logger.Debug("transfer payment: cant send money from empty wallet balance %s, error: %s", args[1], errBalanceShouldNotBeZero)
		return peer.Response{
			Status:  shim.ERROR,
			Message: fmt.Sprintf("transfer payment: cant send money from empty wallet balance %s, error: %s", args[1], errBalanceShouldNotBeZero),
		}
	}

	if senderBalance.Sub(amount).IsNegative() {
		c.logger.Debug("transfer payment: cant send money from wallet %s, error: %s", args[1], errNotEnoughtTokens)
		return peer.Response{
			Status:  shim.ERROR,
			Message: fmt.Sprintf("transfer payment: cant send money from wallet %s, error: %s", args[1], errNotEnoughtTokens),
		}
	}

	senderBalance = senderBalance.Sub(amount)
	receiverBalance = receiverBalance.Add(amount)

	senderWallet.Balance = senderBalance.String()
	receiverWallet.Balance = receiverBalance.String()

	senderWalletState, err = json.Marshal(senderWallet)
	if err != nil {
		c.logger.Error("transfer payment: failed to serialize sender wallet", err)
		return peer.Response{
			Status:  shim.ERROR,
			Message: fmt.Sprintf("transfer payment: failed to serialize sender wallet %s", err),
		}
	}

	if err := stub.PutState(senderAddress, senderWalletState); err != nil {
		c.logger.Error("transfer payment: failed to put sender wallet state", err)
		return peer.Response{
			Status:  shim.ERROR,
			Message: fmt.Sprintf("transfer payment: failed to put sender wallet state %s", err),
		}
	}

	receiverWalletState, err = json.Marshal(receiverWallet)
	if err != nil {
		c.logger.Error("transfer payment: failed to serialize receiver wallet", err)
		return peer.Response{
			Status:  shim.ERROR,
			Message: fmt.Sprintf("transfer payment: failed to serialize receiver wallet %s", err),
		}
	}

	if err := stub.PutState(receiverAddress, receiverWalletState); err != nil {
		c.logger.Error("transfer payment: failed to put receiver wallet state", err)
		return peer.Response{
			Status:  shim.ERROR,
			Message: fmt.Sprintf("transfer payment: failed to put receiver wallet state %s", err),
		}
	}

	return peer.Response{
		Status: shim.OK,
	}
}
