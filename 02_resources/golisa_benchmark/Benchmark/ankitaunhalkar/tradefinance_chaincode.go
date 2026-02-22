package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type TradeFinanceChaincode struct {
}

type Account struct {
	AccountName    string  `json:"account_name"`
	AccountNumber  string  `json:"account_number"`
	AccountBalance float64 `json:"account_balance"`
	BankName       string  `json:"bank_name"`
}

type Contract struct {
	ContractId                 string  `json:"contract_id"`
	ContractDescription        string  `json:"contract_description"`
	ContractAmount             float64 `json:"contract_amount"`
	ContractImporter           string  `json:"contract_importer"`
	ContractExporter           string  `json:"contract_exporter"`
	ContractImporterBank       string  `json:"contract_importerbank"`
	ContractInsurance          string  `json:"contract_insurance"`
	ContractCustom             string  `json:"contract_custom"`
	ContractLoadingPort        string  `json:"contract_loadingport"`
	ContractEntryPort          string  `json:"contract_entryport"`
	ContractImporterBankStatus bool    `json:"contract_importerbankstatus"`
	ContractInsuranceStatus    bool    `json:"contract_insurancestatus"`
	ContractCustomStatus       bool    `json:"contract_customstatus"`
	BillOfLading               string  `json:"billofLading"`
	LetterOfCredit             string  `json:"letterofCredit"`
}

func (tf *TradeFinanceChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Println("Init Methhod")
	args := stub.GetStringArgs()
	if len(args) != 2 {
		fmt.Println("If for argument length")
		return shim.Error("Incorrect arguments. Expecting a key and a value")
	}
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		fmt.Println("If error exists")
		return shim.Error(fmt.Sprintf("Failed to create asset"))
	}
	fmt.Println("Scucces")
	return shim.Success(nil)
}

// var logger = shim.NewLogger("tradefinance_chaincode")

func (tf *TradeFinanceChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()

	fmt.Println("Invoke is running1" + function + "args" + args[0])

	if function == "createAccount" { // Create Account
		return tf.createAccount(stub, args)
	} else if function == "query" {
		return tf.query(stub, args)
	} else if function == "createContract" { // Create Contract
		return tf.createContract(stub, args)
	} else if function == "getBalance" { //Get Balance
		return tf.getBalance(stub, args)
	} else if function == "getContract" { //Get Contract
		return tf.getContract(stub, args)
	} else if function == "insuranceAcceptance" {
		return tf.insuranceAcceptance(stub, args)
	} else if function == "importerBankAcceptance" {
		return tf.importerBankAcceptance(stub, args)
	} else if function == "customAcceptance" {
		return tf.customAcceptance(stub, args)
	} else if function == "transferamount" {
		fmt.Println("calling transferamount")
		return tf.transferamount(stub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (tf *TradeFinanceChaincode) createAccount(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	fmt.Println("Create Account")
	fmt.Println(args[0])
	fmt.Println(args[1])
	fmt.Println(args[2])
	fmt.Println(args[3])

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	accountBalance, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return shim.Error("Cannot parse string to float")
	}

	var account = Account{AccountName: args[0], AccountNumber: args[1], AccountBalance: accountBalance, BankName: args[3]}

	accountBytes, err := json.Marshal(account)
	if err != nil {
		return shim.Error("Error in json Marshal ")
	}

	stub.PutState(args[1], accountBytes)

	var a Account
	_ = json.Unmarshal(accountBytes, &a)
	fmt.Println(a)
	fmt.Println(accountBytes)

	return shim.Success(accountBytes)
}

func (tf *TradeFinanceChaincode) createContract(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 15 {
		return shim.Error("Incorrect number of arguments. Expecting 11")
	}

	fmt.Println(args[0])
	fmt.Println(args[1])
	fmt.Println(args[2])
	fmt.Println(args[3])
	fmt.Println(args[4])
	fmt.Println(args[5])
	fmt.Println(args[6])
	fmt.Println(args[7])
	fmt.Println(args[8])
	fmt.Println(args[9])
	fmt.Println(args[10])
	fmt.Println(args[11])
	fmt.Println(args[12])
	fmt.Println(args[13])
	fmt.Println(args[14])

	contractAmount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return shim.Error("Cannot parse string to float")
	}
	imStatus, _ := strconv.ParseBool(args[10])
	iStatus, _ := strconv.ParseBool(args[11])
	cStatus, _ := strconv.ParseBool(args[12])
	var contract = Contract{ContractId: args[0],
		ContractDescription: args[1], ContractAmount: contractAmount,
		ContractImporter: args[3], ContractExporter: args[4],
		ContractImporterBank: args[5], ContractInsurance: args[6], ContractCustom: args[7],
		ContractLoadingPort: args[8], ContractEntryPort: args[9],
		ContractImporterBankStatus: imStatus, ContractInsuranceStatus: iStatus,
		ContractCustomStatus: cStatus, BillOfLading: args[13], LetterOfCredit: args[14]}

	contractBytes, err := json.Marshal(contract)

	if err != nil {
		return shim.Error("Error in json Marshal ")
	}

	error := stub.PutState(args[0], contractBytes)

	fmt.Println(error)
	return shim.Success(contractBytes)
}

func (tf *TradeFinanceChaincode) getBalance(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	fmt.Println("Get balance" + args[0])

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	account, err := stub.GetState(args[0])

	if err != nil {
		return shim.Error(args[0] + " not found")
	}

	fmt.Println(account)

	return shim.Success(account)
}

func (tf *TradeFinanceChaincode) getContract(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	contract, err := stub.GetState(args[0])

	if err != nil {
		return shim.Error(args[0] + " not found")
	}

	return shim.Success(contract)
}

func (tf *TradeFinanceChaincode) importerBankAcceptance(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	contract, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Error occured in get state")
	}

	var contractDetails Contract
	json.Unmarshal(contract, &contractDetails)

	if contractDetails.ContractCustomStatus == false || contractDetails.ContractInsuranceStatus == false {
		return shim.Error("Customs or Insurance are not accepted")
	}

	exporter, err := stub.GetState(contractDetails.ContractExporter)

	if exporter == nil || err != nil {
		return shim.Error("No exporter found")
	}

	importerAccount, err := stub.GetState(contractDetails.ContractImporter)

	if err != nil {
		return shim.Error("Error occured in get state")
	}

	var accountDetails Account
	json.Unmarshal(importerAccount, &accountDetails)
	if accountDetails.AccountBalance > contractDetails.ContractAmount {
		contractDetails.ContractImporterBankStatus = true
	} else {
		contractDetails.ContractImporterBankStatus = false
	}

	contractDetailsBytes, _ := json.Marshal(contractDetails)
	stub.PutState(args[0], contractDetailsBytes)
	return shim.Success(contractDetailsBytes)
}

func (tf *TradeFinanceChaincode) insuranceAcceptance(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	contract, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Error occured in get state")
	}

	var contractDetails Contract
	json.Unmarshal(contract, &contractDetails)

	if contractDetails.LetterOfCredit == "" || contractDetails.BillOfLading == "" {
		contractDetails.ContractInsuranceStatus = false
		// return shim.Error("Documents are not attacted")
	} else {
		contractDetails.ContractInsuranceStatus = true
	}

	contractDetailsBytes, _ := json.Marshal(contractDetails)
	stub.PutState(args[0], contractDetailsBytes)
	return shim.Success(contractDetailsBytes)
}

func (tf *TradeFinanceChaincode) customAcceptance(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	contract, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Error occured in get state")
	}

	var contractDetails Contract
	json.Unmarshal(contract, &contractDetails)

	if contractDetails.ContractEntryPort == "" && contractDetails.ContractLoadingPort == "" {
		// return shim.Error("Ports are not assigned")
		contractDetails.ContractCustomStatus = false
	} else {
		contractDetails.ContractCustomStatus = true
	}
	contractDetailsBytes, _ := json.Marshal(contractDetails)
	stub.PutState(args[0], contractDetailsBytes)
	return shim.Success(contractDetailsBytes)
}

func (tf *TradeFinanceChaincode) transferamount(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	fmt.Println("calling contract ", args[0])

	contract, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Error occured in get state")
	}
	fmt.Println(contract)

	var contractDetails Contract
	json.Unmarshal(contract, &contractDetails)

	fmt.Println(contractDetails)
	if contractDetails.ContractImporterBankStatus == false {
		// fmt.Println("Status false")
		return shim.Error("Importer bank has not approved transaction! No balance avaliable")
	}

	//Deduct from importer account
	fmt.Println(contractDetails.ContractImporter)

	importerAccount, _ := stub.GetState(contractDetails.ContractImporter)
	var impAccount Account
	json.Unmarshal(importerAccount, &impAccount)
	fmt.Println(impAccount)

	impBalance := impAccount.AccountBalance - contractDetails.ContractAmount
	impBalance = impBalance - 500
	impAccount.AccountBalance = impBalance
	impUpdate, _ := json.Marshal(impAccount)
	fmt.Println("Updated imp", impUpdate)
	stub.PutState(impAccount.AccountNumber, impUpdate)

	//Add in exporter account
	fmt.Println(contractDetails.ContractExporter)

	exporterAccount, _ := stub.GetState(contractDetails.ContractExporter)
	var expAccount Account
	json.Unmarshal(exporterAccount, &expAccount)
	expbalance := expAccount.AccountBalance + contractDetails.ContractAmount
	expAccount.AccountBalance = expbalance
	expUpdate, _ := json.Marshal(expAccount)
	fmt.Println("Updated exp", expUpdate)
	stub.PutState(expAccount.AccountNumber, expUpdate)

	//Add in custom account
	fmt.Println(contractDetails.ContractCustom)

	customAccount, _ := stub.GetState(contractDetails.ContractCustom)
	var custAccount Account
	json.Unmarshal(customAccount, &custAccount)
	custbalance := custAccount.AccountBalance + 500
	custAccount.AccountBalance = custbalance
	custUpdate, _ := json.Marshal(custAccount)
	fmt.Println("Updated cust", custUpdate)
	stub.PutState(custAccount.AccountNumber, custUpdate)

	return shim.Success(impUpdate)
}

func (tf *TradeFinanceChaincode) query(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	fmt.Println("Get method")
	value, err := stub.GetState(args[0])
	if err != nil {
		fmt.Println("Error in Get method")
		return shim.Error("Failed to set asset")
	}
	if value == nil {
		fmt.Println("Error in Get method value")
		return shim.Error("Asset not found")
	}
	fmt.Println("succes" + string(value))
	return shim.Success(value)
}

func main() {

	err := shim.Start(new(TradeFinanceChaincode))
	if err != nil {
		fmt.Printf("Error starting Trade Finanace chaincode: %s", err)
	}
}
