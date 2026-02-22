/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Chaincode is the definition of the chaincode structure.
type Chaincode struct {
}

// ERC30Metadata 토큰 메타데이터 정보
type ERC20Metadata struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Owner       string `json:"onwer"`
	TotalSupply uint64 `json:"totalSupply"`
}

// TransferEvent is the event definition of Transfer
type TransferEvent struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    int    `json:"amount"`
}

type Approval struct {
	Owner     string `json:"owner"`
	Spender   string `json:"spender"`
	Allowance int    `json:"allowance"`
}

// Init is called when the chaincode is instantiated by the blockchain network.
// Init 을 하기 위해 params - tokenName, symbol,  owner (주인) , amount (얼마 만큼 발행할 것 인지)
func (cc *Chaincode) Init(stub shim.ChaincodeStubInterface) sc.Response {
	_, params := stub.GetFunctionAndParameters()
	fmt.Println("Init called with params : ", params)
	if len(params) != 4 {
		shim.Error("Incorrect number of params")
	}

	tokenName, symbol, owner, amount := params[0], params[1], params[2], params[3]

	// check amount is unsinged int
	amountUint, err := strconv.ParseUint(string(amount), 10, 64)
	if err != nil {
		return shim.Error("amount must be a mount or amount can't be negative ")
	}
	if len(tokenName) == 0 || len(symbol) == 0 || len(owner) == 0 {
		return shim.Error("토큰이름, symbol,owner 는 공백일 수 없습니다.")
	}

	//make metadata
	erc20 := &ERC20Metadata{Name: tokenName, Symbol: symbol, Owner: owner, TotalSupply: amountUint}
	// Interface 값을 byte 형식으로 바꿔준다
	erc20Byte, err := json.Marshal(erc20)
	if err != nil {
		return shim.Error("failed to Marshal erc20, error: " + err.Error())
	}
	// save token  DB에 저장
	err = stub.PutState(tokenName, []byte(erc20Byte))
	if err != nil {
		return shim.Error("failed to PutState, error : " + err.Error())
	}

	// save Owner Balance
	err = stub.PutState(owner, []byte(amount))
	if err != nil {
		return shim.Error("failed to PutState, error : " + err.Error())
	}
	return shim.Success(nil)
}

// Invoke is called as a result of an application request to run the chaincode.
func (cc *Chaincode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	fcn, params := stub.GetFunctionAndParameters()

	//GetArgs 는 이중배열로 들어온다.
	fmt.Println("GetArgs는 이중 배열로 들어옴  =======================================")
	args := stub.GetArgs()
	fmt.Println("GetArgs(): ")
	for _, arg := range args {
		argStr := string(arg)
		fmt.Printf("%s", argStr)
	}
	fmt.Println("================================================================")
	fmt.Println()
	fmt.Println("GetStringArgs()스트링 배열로 들어온다==================================")
	//GetStringArgs()스트링 배열로 들어온다.
	stringArgs := stub.GetStringArgs()
	fmt.Println("GetstringArg() ", stringArgs) // [totalSupply dappcampus 서울특별시]
	fmt.Println("================================================================")
	fmt.Println()

	fmt.Println("GetArgsSlice() 는 에러도 같이 반환한다.==================================")
	// GetArgsSlice() 는 에러도 같이 반환한다.
	argsSlice, _ := stub.GetArgsSlice()
	fmt.Println("GetArgsSlice() :", string(argsSlice)) // totalSupplydappcampus서울특별시
	fmt.Println("================================================================")

	switch fcn {
	case "totalSupply":
		return cc.TotalSupply(stub, params)
	case "balanceOf":
		return cc.BalanceOf(stub, params)
	case "transfer":
		return cc.Transfer(stub, params)
	case "allowance":
		return cc.Allowance(stub, params)
	case "approve":
		return cc.Approve(stub, params)
	case "approvalList":
		return cc.ApprovalList(stub, params)
	case "transferFrom":
		return cc.TransferFrom(stub, params)
	case "transferOtherToken":
		return cc.TransferOtherToken(stub, params)
	case "increaseAllowance":
		return cc.IncreaseAllowance(stub, params)
	case "decreaseAllowance":
		return cc.DecreaseAllowance(stub, params)
	case "mint":
		return cc.Mint(stub, params)
	case "burn":
		return cc.Burn(stub, params)
	default:
		return sc.Response{Status: 404, Message: "함수를 찾을 수 없습니다", Payload: nil}
		//return shim.Error("함수를 찾을 수 없습니다.")
	}
}

//totalSupply is query function
// params - tokenName
func (cc *Chaincode) TotalSupply(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	// check params
	if len(params) != 1 {
		return shim.Error("Incorrect number of params")
	}
	tokenName := params[0]
	tokenMarshal, err := stub.GetState(tokenName)
	if err != nil {
		return shim.Error("failed to GetState, error :" + err.Error())
	}
	tokenInfo := ERC20Metadata{}
	err = json.Unmarshal(tokenMarshal, &tokenInfo)
	if err != nil {
		return shim.Error("failed to Unmarshal, error :" + err.Error())
	}

	TotalSupplyByte, err := json.Marshal(tokenInfo.TotalSupply)
	if err != nil {
		return shim.Error("failed to Marshal, error :" + err.Error())
	}

	fmt.Println()
	fmt.Println("=====================TotalSupply 호출=====================")
	fmt.Println(tokenName+" 의 TotalSupply 값은 :", TotalSupplyByte, " 입니다")
	fmt.Println("=========================================================")

	//결국 TotalSupply를 구하려면
	// 1. toKenName을 가져오고
	// 2. 그 토큰정보로 Erc20Metadata{} 가져오고 ( struct타입이니까 json.unmarshal을 한거고)
	// 3. 가져온 정보의 .TotalSupply를 	다시 byte형식으로 만들기위해서 Marshal
	return shim.Success(TotalSupplyByte)
}

func (cc *Chaincode) BalanceOf(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	//check params
	if len(params) != 1 {
		return shim.Error("Incorrect number of parameter")
	}
	ownerAddress := params[0]

	ownerAmount, err := stub.GetState(ownerAddress)
	if err != nil {
		return shim.Error("failed to GetState (ownerAddress), error :" + err.Error())
	}

	fmt.Println()
	fmt.Println("=====================BalanceOf 호출=====================")
	fmt.Println(ownerAddress+" 의 BalanceOf 값은 :", string(ownerAmount), " 입니다")
	fmt.Println("=========================================================")

	// 결국 onwerAmount는 byte형태 이므로 바로 return하여도 상관없다.
	return shim.Success(ownerAmount)
}

// params - owner's Address , recipient's Address, Amount
func (cc *Chaincode) Transfer(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	if len(params) != 3 {
		return shim.Error("Incorrect number of parameters")
	}
	ownerAddress, recipientAddress, amount := params[0], params[1], params[2]

	AmountInt, err := strconv.Atoi(amount)
	if err != nil {
		return shim.Error("Amount must be Integer")
	}
	if AmountInt < 0 {
		return shim.Error("Amount must be Positive")
	}

	ownerAmount, err := stub.GetState(ownerAddress)
	if err != nil {
		return shim.Error("failed to stub.Getstate, error :" + err.Error())
	}
	ownerAmountInt, err := strconv.Atoi(string(ownerAmount))
	if err != nil {
		return shim.Error("failed to strconv.Atoi, error :" + err.Error())
	}

	recipientAmount, err := stub.GetState(recipientAddress)
	if err != nil {
		return shim.Error("failed to stub.GetState, error :" + err.Error())
	}
	//** recipientAmount가 nil 일 경우에도 조건을 줘야함 **
	if recipientAmount == nil {
		recipientAmount = []byte("0")
	}
	recipeintAmountInt, err := strconv.Atoi(string(recipientAmount))
	if err != nil {
		return shim.Error("failed to strconv.Atio.error :" + err.Error())
	}

	changeAmount := ownerAmountInt - AmountInt
	recipeintAmountInt = recipeintAmountInt + AmountInt

	if changeAmount > 0 {
		return shim.Error("ownerAmount large than Amount int")
	}

	err = stub.PutState(ownerAddress, []byte(strconv.Itoa(changeAmount)))
	if err != nil {
		return shim.Error("failed to PutState")
	}

	err = stub.PutState(recipientAddress, []byte(strconv.Itoa(recipeintAmountInt)))
	if err != nil {
		return shim.Error("failed to Pustate")
	}
	erc20 := TransferEvent{Sender: ownerAddress, Recipient: recipientAddress, Amount: AmountInt}
	erc20Byte, err := json.Marshal(erc20)
	if err != nil {
		return shim.Error("failed to json.Marshal, error :" + err.Error())
	}
	err = stub.SetEvent("transferEvnet", erc20Byte)
	if err != nil {
		return shim.Error("error!!!!!")
	}

	fmt.Println()
	fmt.Println("=====================Transfer 호출=======================")
	fmt.Println(ownerAddress + " 가 " + recipientAddress + " 에게 " + amount + " 만큼의 금액을 전송 ")
	fmt.Println("=========================================================")

	return shim.Success([]byte("transfer Success"))
}

// params - owner's Address , recipient's Address
func (cc *Chaincode) Allowance(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	if len(params) != 2 {
		return shim.Error("Incorrect number of params ")
	}
	ownerAddress, spenderAddress := params[0], params[1]

	ApprovalKey, err := stub.CreateCompositeKey("approval", []string{ownerAddress, spenderAddress})
	if err != nil {
		return shim.Error("failed to CreateCompositeKey, error " + err.Error())
	}

	amount, err := stub.GetState(ApprovalKey)
	if err != nil {
		return shim.Error("failed to GetState,error :" + err.Error())
	}
	// ** amount 가 nil 경우도 생각해야함
	if amount == nil {
		amount = []byte("0")
	}
	return shim.Success(amount)
}

// params -  owner'sAddress , recipientAddress , Amount
func (cc *Chaincode) Approve(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	if len(params) != 3 {
		return shim.Error("Incorrect number of params ")
	}
	ownerAddress, recipientAddress, amount := params[0], params[1], params[2]
	amountInt, err := strconv.Atoi(amount)

	if err != nil {
		return shim.Error("Amount must be Integer")
	}
	if amountInt <= 0 {
		return shim.Error("AMount must be Positive")
	}

	ApprovalKey, err := stub.CreateCompositeKey("approval", []string{ownerAddress, recipientAddress})
	if err != nil {
		return shim.Error("failed to CreatCompositeKey, error :" + err.Error())
	}

	// err = stub.PutState("approval", []byte(ApprovalKey))
	// if err != nil {
	// 	return shim.Error("failed to Pustate(approval, approvalKey), error" + err.Error())
	// }

	err = stub.PutState(ApprovalKey, []byte(amount))
	if err != nil {
		return shim.Error("failed to Pustate(ApprovalKey, amount), error :" + err.Error())
	}

	AllowanceStruct := Approval{Owner: ownerAddress, Spender: recipientAddress, Allowance: amountInt}
	AllowanceByte, err := json.Marshal(AllowanceStruct)
	if err != nil {
		return shim.Error("failed to Marshal(AllowanceStruct), error :" + err.Error())
	}
	err = stub.SetEvent("ApprovalEvent", AllowanceByte)
	if err != nil {
		return shim.Error("failed to SetEvent(ApprovalEvent), error :" + err.Error())
	}

	return shim.Success(nil)
}

func (cc *Chaincode) ApprovalList(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	return shim.Success(nil)
}

func (cc *Chaincode) TransferFrom(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	return shim.Success([]byte("tranferFrom success"))
}

func (cc *Chaincode) TransferOtherToken(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	return shim.Success([]byte("transfer other token Success"))
}

func (cc *Chaincode) IncreaseAllowance(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	return shim.Success([]byte("increaseAllowance success"))
}

func (cc *Chaincode) DecreaseAllowance(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	return shim.Success(nil)
}
func (cc *Chaincode) Mint(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	return shim.Success(nil)
}
func (cc *Chaincode) Burn(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	return shim.Success(nil)
}
