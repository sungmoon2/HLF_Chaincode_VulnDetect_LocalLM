package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// define logger
var tokenLogger = shim.NewLogger("LinkaToken")

// define variables
const (
	TokenName        string = "LinkaToken"
	Symbol           string = "LINKA"
	TokenTotalAmount string = "2000000000" // Ether
	UnitWeiToEth     string = "1000000000000000000"

	NoncePrefix         string = "NONCE"
	AllowancePrefix     string = "APPROVE"
	EventApprovalPrefix string = "EVENT-APPROVAL"
	EventTransferPrefix string = "EVENT-TRANSFER"
	OwnerSuffix         string = "_OWNER"

	OwnerDocType string = "owner"

	// error messages
	ErrNotValidatedParameters      string = `Parameters are not valid`
	ErrNotValidatedNonce           string = `Nonce is not valid`
	ErrNotValidatedSignTransaction string = `Signature is not valid`
	ErrNotInitalizedChaincode      string = `Token's initial information is different or not initialized`
	ErrAlreadyExistAddress         string = `Address exists`
	ErrNotExistAddress             string = `Address does not exist.`
	ErrNotValidateAddress          string = `Address is not valid`
	ErrNotEnoughBalance            string = `Not enough balance`
	ErrNotExistAllowance           string = `Permission does not exist.`
	ErrNotEnoughAllowance          string = `Not enough funds`

	ErrRecoveredFromPanic string = `Recovered from panic`

	ERC20ABI = `[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"initialSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_subtractedValue","type":"uint256"}],"name":"decreaseApproval","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_addedValue","type":"uint256"}],"name":"increaseApproval","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_newAddress","type":"address"}],"name":"transferOwnership","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"inputs":[{"name":"_initialSupply","type":"uint256"},{"name":"_name","type":"string"},{"name":"_symbol","type":"string"}],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"name":"oldAddress","type":"address"},{"indexed":false,"name":"newAddress","type":"address"}],"name":"TransferOwnership","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"owner","type":"address"},{"indexed":true,"name":"spender","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`
)

// define structs
type (
	linkatoken struct {
	}
	OWNDER struct {
		ObjectType  string `json:"docType"` //docType is used to distinguish the various types of objects in state database
		Address     string `json:"address"` //the fieldtags are needed to keep case from bouncing around
		TotalSupply string `json:"totalsupply"`
		Symbol      string `json:"symbol"`
		Signature   string `json:"signature"`
	}

	H3CC struct {
		ContractInfo []string
		V            *big.Int
		R            *big.Int
		S            *big.Int
	}

	REGIST struct {
		RegistInfo []string
		V          *big.Int
		R          *big.Int
		S          *big.Int
	}
)

var m majorfunc

/** Main
 * it is a starting point
 */
func main() {
	err := shim.Start(new(linkatoken))
	tokenLogger.SetLevel(shim.LogDebug)

	if err != nil {
		tokenLogger.Debug("Error starting Token chaincode: %s", err)
	}
}

/** Init
 * function for initializing chaincode
 * @param: signedTransaction
 */
func (t *linkatoken) Init(stub shim.ChaincodeStubInterface) pb.Response {

	// 1. Check parameter
	_, args := stub.GetFunctionAndParameters()
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	signedTx := args[0]
	tokenLogger.Debug(signedTx)
	if len(signedTx) < 0 || signedTx == "" {
		tokenLogger.Debug(ErrNotValidatedSignTransaction)
		return shim.Error(ErrNotValidatedSignTransaction)
	}

	// 2. parsing signed Transaction. & validate
	revsymbol, senderAddress, nonceStr, err := m.parsingSignature(signedTx)

	if err != nil {
		return shim.Error(err.Error())
	}
	if Symbol != revsymbol {
		return shim.Error(ErrNotValidatedParameters)
	}
	nonce, err := strconv.ParseUint(nonceStr, 10, 64)
	if err != nil {
		return shim.Error(err.Error())
	}
	// get Symbol_OWNER variable from storage
	ownerInitialKey := Symbol + OwnerSuffix
	resultBytes, err := stub.GetState(ownerInitialKey)
	// save as wei unit
	wei := convertStringToBigInt(UnitWeiToEth)
	tokenAmount := convertStringToBigInt(TokenTotalAmount)
	tokenAmountWei := mulBigInt(tokenAmount, wei)

	initOwner := &OWNDER{OwnerDocType, senderAddress, tokenAmountWei.String(), Symbol, signedTx}
	ownerJSONasBytes, err := json.Marshal(initOwner)
	if err != nil {
		return shim.Error(err.Error())
	}
	if resultBytes == nil || err != nil {

		if err != nil {
			return shim.Error(err.Error())
		}

		err = stub.PutState(ownerInitialKey, ownerJSONasBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
		// balance
		err = stub.PutState(senderAddress, doMarshal(tokenAmountWei))
		if err != nil {
			return shim.Error(err.Error())
		}
		if nonce != 1 {
			return shim.Error(ErrNotValidatedNonce)
		}
		// NONCE
		err = saveLatestNonce(stub, senderAddress, nonce)
		if err != nil {
			return shim.Error(err.Error())
		}
	} else {
		var ownerJSON OWNDER
		err = json.Unmarshal([]byte(resultBytes), &ownerJSON)
		if err != nil {
			return shim.Error(err.Error())
		}
		if ownerJSON.Address != senderAddress {
			return shim.Error(ErrNotInitalizedChaincode)
		}

		if ownerJSON.TotalSupply != tokenAmountWei.String() {
			return shim.Error(ErrNotInitalizedChaincode)
		}

		if ownerJSON.Symbol != Symbol {
			return shim.Error(ErrNotInitalizedChaincode)
		}
		// check nonce
		if !isValidateNonce(stub, senderAddress, nonce) {
			return shim.Error(ErrNotValidatedNonce)
		}
		// update owner information
		err = stub.PutState(ownerInitialKey, ownerJSONasBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
		// save nonce
		err = saveLatestNonce(stub, senderAddress, nonce)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	return shim.Success(nil)
}

/** Invoke
 * function for executing chaincode
 * case1:
 *   @param: funtion name
 *   @param: parmameters as function
 *
 * case2:
 *   @param: signedTransaction
 *   @param: result of signing a private key after creating a transaction based on ERC20
 */
func (t *linkatoken) Invoke(stub shim.ChaincodeStubInterface) (res pb.Response) {
	function, args := stub.GetFunctionAndParameters()
	tokenLogger.Debug("invoke is running " + function)

	defer func() {
		if r := recover(); r != nil {
			res = shim.Error(ErrRecoveredFromPanic)
		}
	}()
	var fromAddress, toAddress, relayAddress, nonce, amount string
	var err error
	// need a signed transaction to regist an address
	if function == "signedTransaction" {
		// need to parse args
		function, fromAddress, toAddress, relayAddress, nonce, amount, err = m.parsingTransaction(args[0])
		// fmt.Println(function, fromAddress, toAddress, relayAddress, nonce, amount)
		if err != nil {
			return shim.Error(err.Error())
		}
		args = []string{nonce, fromAddress, toAddress, relayAddress, amount}
		switch function {
		case "transfer": // transfer a token
			return m.setTransfer(stub, args)
		case "approve": // transfer of token trading rights
			return m.setApprove(stub, args)
		case "transferFrom": // transfer a token
			return m.setTransferFrom(stub, args)
		default:
			tokenLogger.Debug("invoke did not find func: " + function)
			return shim.Error(fmt.Sprintf(`Received unknown function invocation`))
		}
	} else if function == "regist" { // regist an address
		address, err := m.parsingRegist(args[0])
		if err != nil {
			return shim.Error(err.Error())
		}
		args = []string{address}
	}

	switch function {
	case "totalSupply": // total volumns for issue tokens when initialize a chaincode
		return m.getTotalSupply(stub, args)
	case "balanceOf": // check balance
		return m.getBalanceOf(stub, args)
	case "allowance": // check of token trading rights
		return m.getAllowance(stub, args)
	case "regist": // regist an address
		return m.setRegist(stub, args)
	case "transactionCount": // check nonce
		return m.getNonce(stub, args)
	case "history": // check a list of account's history
		return m.getHistory(stub, args)
	default:
		tokenLogger.Debug("invoke did not find func: " + function)
		return shim.Error(fmt.Sprintf(`Received unknown function invocation`))
	}
}
