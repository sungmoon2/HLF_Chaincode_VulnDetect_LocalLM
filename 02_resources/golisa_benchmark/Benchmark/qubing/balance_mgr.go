package main

import (
	"bytes"
	"fmt"
	"strconv"

	// "github.com/hyperledger/fabric/core/chaincode/lib/cid"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	//"github.com/chaincodes/common/crypto"
)

const (
	KEY_PUBLIC  = "KEY_PUBLIC"
	KEY_PRIVATE = "KEY_PRIVATE"
	KEY_PREFIX  = "CRYPT_"
)

// BalanceManager Smart Contract(Chaincode) implementation
type BalanceManager struct {
}

// Init - Initialize smart contract
func (t *BalanceManager) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("<BalanceMgr Init>")

	funcName, params := stub.GetFunctionAndParameters()
	fmt.Printf("Getting deployment parameters ... function_name: %s, parameters: %v", funcName, params)
	fmt.Println()

	if funcName == "init" {
		// Create account with balance of '0'
		return t.doInit(stub)
	} else if funcName == "upgrade" {
		// Charge account with amount
		return t.doUpgrade(stub)
	}

	return shim.Error(fmt.Sprintf(`Invalid function name for deployment. 
		(expecting 'init' or 'upgrade', actual: '%s')`, funcName))
}

// func checkCryptoKeyPairExisting(stub shim.ChaincodeStubInterface) bool {
// 	pubKey, _ := stub.GetState(KEY_PUBLIC)
// 	privKey, _ := stub.GetState(KEY_PRIVATE)
// 	return (pubKey != nil && len(pubKey) > 0) && (privKey != nil && len(privKey) > 0)
// }

func (t *BalanceManager) doInit(stub shim.ChaincodeStubInterface) pb.Response {
	_, params := stub.GetFunctionAndParameters()
	paramCount := len(params)
	if paramCount != 0 && paramCount != 2 {
		return shim.Error(fmt.Sprintf(`Incorrect number of arguments. 
			(expecting: 0 or 2, actual: %d)`, paramCount))
	}

	// check cryption key-pair existing
	// if checkCryptoKeyPairExisting(stub) == false {
	// 	if paramCount == 2 {
	// 		fmt.Println("Initializing RSA key-pair with deployment arguments ...")
	// 		stub.PutState(KEY_PUBLIC, []byte(params[0]))
	// 		stub.PutState(KEY_PRIVATE, []byte(params[1]))
	// 		fmt.Println("Initialize RSA key-pair for cryption successfully.")
	// 		return shim.Success(nil)
	// 	}

	// 	fmt.Println("Generating RSA key-pair ...")
	// 	publicKey := *bytes.NewBufferString("")
	// 	privateKey := *bytes.NewBufferString("")
	// 	err := crypto.CreateKeyPair(&publicKey, &privateKey, 256)
	// 	if err == nil {
	// 		stub.PutState(KEY_PUBLIC, publicKey.Bytes())
	// 		stub.PutState(KEY_PRIVATE, privateKey.Bytes())
	// 		fmt.Println("Generate RSA key-pair for cryption successfully.")
	// 		return shim.Success(nil)
	// 	}

	// 	fmt.Printf("Initialize RSA key-pair for cryption failed. cause: (%s)", err)
	// 	fmt.Println()
	// 	return shim.Error(err.Error())
	// }

	fmt.Println("RSA key-pair already existing, initialization not required.")
	return shim.Success(nil)
}

func (t *BalanceManager) doUpgrade(stub shim.ChaincodeStubInterface) pb.Response {
	_, params := stub.GetFunctionAndParameters()
	paramCount := len(params)
	if paramCount != 0 && paramCount != 2 {
		return shim.Error(fmt.Sprintf(`Incorrect number of arguments. 
			(expecting: 0 or 2, actual: %d)`, paramCount))
	}

	if paramCount == 2 {
		fmt.Println("Initializing RSA key-pair with deployment arguments ...")
		stub.PutState(KEY_PUBLIC, []byte(params[0]))
		stub.PutState(KEY_PRIVATE, []byte(params[1]))
		fmt.Println("Initialize RSA key-pair for cryption successfully.")
	}

	return shim.Success(nil)
}

// Invoke - Accessing smart contract interface
func (t *BalanceManager) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, args := stub.GetFunctionAndParameters()
	fmt.Printf("<BalanceMgr Invoke>: %s", funcName)
	if funcName == "create" {
		// Create account with balance of '0'
		return t.create(stub, args)
	} else if funcName == "charge" {
		// Charge account with amount
		return t.charge(stub, args)
	} else if funcName == "transfer" {
		// Transfer A to B with some money
		return t.transfer(stub, args)
	} else if funcName == "query" {
		// Query account balance
		return t.query(stub, args)
	} else if funcName == "get" {
		// Get normal val
		return t.get(stub, args)
	} else if funcName == "put" {
		// Put normal val
		return t.put(stub, args)
	} else if funcName == "json" {
		// Put normal val
		return t.json(stub, args)
	} else if funcName == "event" {
		// Put decrypted val
		return t.sendEvent(stub, args)
	}
	//  else if funcName == "getP" {
	// 	return t.getPrivateData(stub, args)
	// } else if funcName == "putP" {
	// 	return t.putPrivateData(stub, args)
	// } else if funcName == "getX" {
	// 	// Get encrypted val
	// 	return t.getDecryption(stub, args)
	// } else if funcName == "putX" {
	// 	// Put decrypted val
	// 	return t.putEncryption(stub, args)
	// }

	return shim.Error(fmt.Sprintf(`Invalid invoke function name. Expecting 'create','transfer',
	'query', 'get', 'getX', 'put', 'putX', 'json' and 'event'. Actual: '%s'`, funcName))
}

// create: create account initialized with 0
func (t *BalanceManager) create(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("create account with initial balance of '0'")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	accountName := args[0]

	valBytes, err := stub.GetState(accountName)
	if err != nil {
		return shim.Error("Account create failed with unknown reason.")
	}

	if valBytes != nil && len(valBytes) > 0 {
		return shim.Error(fmt.Sprintf(`Account already existed. (Account: "%s")`, accountName))
	}

	stub.PutState(accountName, []byte(strconv.Itoa(0)))

	byts, err := stub.GetCreator()

	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Printf(`Printing current user is "%s".`, string(byts))
	fmt.Println()
	return shim.Success(nil)
}

func (t *BalanceManager) charge(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("charge account with amount")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	accountFrom := args[0]

	bytesFrom, err := stub.GetState(accountFrom)
	if err != nil {
		return shim.Error("Account charge failed with unknown reason.")
	}
	if bytesFrom == nil {
		return shim.Error("Account not found")
	}
	valFrom, _ := strconv.Atoi(string(bytesFrom))

	// Perform the execution
	amountCharge, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	valFrom = valFrom + amountCharge

	stub.PutState(accountFrom, []byte(strconv.Itoa(valFrom)))

	stub.SetEvent("hello", []byte(fmt.Sprintf(`{"status":"ok", "account":"%s", "amount":"%d"}`, accountFrom, amountCharge)))

	return shim.Success(nil)
}

// transfer: Transfer balance from account to another
func (t *BalanceManager) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("transfer account")

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	accountFrom := args[0]
	accountTo := args[1]

	// Get the state from the ledger
	bytesFrom, err := stub.GetState(accountFrom)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if bytesFrom == nil {
		return shim.Error("Entity not found")
	}
	valFrom, _ := strconv.Atoi(string(bytesFrom))

	bytesTo, err := stub.GetState(accountTo)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if bytesTo == nil {
		return shim.Error("Entity not found")
	}
	valTo, _ := strconv.Atoi(string(bytesTo))

	// Perform the execution
	amountTransfer, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	valFrom = valFrom - amountTransfer
	valTo = valTo + amountTransfer
	fmt.Printf("valFrom = %d, valTo = %d\n", valFrom, valTo)
	fmt.Println()

	// Write the state back to the ledger
	err = stub.PutState(accountFrom, []byte(strconv.Itoa(valFrom)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(accountTo, []byte(strconv.Itoa(valTo)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// query: query account balance
func (t *BalanceManager) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A string // Entities
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	A = args[0]

	// Get the state from the ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	if Avalbytes == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + A + "\"}"
		return shim.Error(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + A + "\",\"Amount\":\"" + string(Avalbytes) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	fmt.Println()

	byts, err := stub.GetCreator()

	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Printf(`Printing current user is "%s".`, string(byts))
	fmt.Println()

	// id, err := cid.GetID(stub)

	// if err != nil {
	// 	fmt.Printf(`GetID failed. Error: "%s".`, err.Error())
	// } else {
	// 	fmt.Printf(`GetID="%s".`, id)
	// }
	// fmt.Println()

	// mspId, err := cid.GetMSPID(stub)
	// if err != nil {
	// 	fmt.Printf(`GetMSPID failed. Error: "%s".`, err.Error())
	// } else {
	// 	fmt.Printf(`GetMSPID="%s".`, mspId)
	// }
	// fmt.Println()

	// email_val, email_found, err := cid.GetAttributeValue(stub, "email")
	// if err != nil {
	// 	fmt.Printf(`GetAttributeValue('email') failed. Error: "%s".`, err.Error())
	// } else if !email_found {
	// 	fmt.Printf(`GetAttributeValue('email') not found. Value: "%s".`, email_val)
	// } else {
	// 	fmt.Printf(`GetAttributeValue('email') successfully. Value: "%s".`, email_val)
	// }
	// fmt.Println()

	// admin_val, admin_found, err := cid.GetAttributeValue(stub, "admin")
	// if err != nil {
	// 	fmt.Printf(`GetAttributeValue('admin') failed. Error: "%s".`, err.Error())
	// } else if !admin_found {
	// 	fmt.Printf(`GetAttributeValue('admin') not found. Value: "%s".`, admin_val)
	// } else {
	// 	fmt.Printf(`GetAttributeValue('admin') successfully. Value: "%s".`, admin_val)
	// }
	// fmt.Println()

	// err = cid.AssertAttributeValue(stub, "admin", "true")
	// if err != nil {
	// 	fmt.Printf(`AssertAttributeValue('admin', 'true') failed. Error: "%s".`, err.Error())
	// } else {
	// 	fmt.Printf(`AssertAttributeValue('admin', 'true') successfully.`)
	// }
	// fmt.Println()

	// cert, err := cid.GetX509Certificate(stub)
	// if err != nil {
	// 	fmt.Printf(`GetX509Certificate failed. Error: "%s".`, err.Error())
	// } else {
	// 	fmt.Printf(`GetX509Certificate="%s".`, cert.Raw)
	// }

	// fmt.Println()

	return shim.Success(Avalbytes)
}

func (t *BalanceManager) put(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	key := args[0]
	val := args[1]

	err := stub.PutState(key, []byte(val))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *BalanceManager) get(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	key := args[0]
	val, err := stub.GetState(key)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(val)
}

func (t *BalanceManager) json(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	selector := args[0]
	queryIt, err := stub.GetQueryResult(selector)

	if err != nil {
		return shim.Error(err.Error())
	}

	var buff bytes.Buffer
	buff.WriteString("[")

	for queryIt.HasNext() {
		queryResult, err := queryIt.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if buff.Len() > 1 {
			buff.WriteString(",")
		}
		buff.WriteString(string(queryResult.GetValue()))
	}

	buff.WriteString("]")

	return shim.Success(buff.Bytes())
}

// func (t *BalanceManager) putEncryption(stub shim.ChaincodeStubInterface, args []string) pb.Response {
// 	if len(args) != 2 {
// 		return shim.Error("Incorrect number of arguments. Expecting 2")
// 	}

// 	key := args[0]
// 	val := args[1]
// 	pubKey, err := stub.GetState(KEY_PUBLIC)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}
// 	helper, errs := crypto.NewRSAHelper(pubKey, nil)
// 	if errs != nil && len(errs) > 0 {
// 		return shim.Error(err.Error())
// 	}
// 	encoded, err := helper.Encrypt(val)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}
// 	err = stub.PutState(KEY_PREFIX+key, []byte(encoded))
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	return shim.Success(nil)
// }

// func (t *BalanceManager) getDecryption(stub shim.ChaincodeStubInterface, args []string) pb.Response {
// 	if len(args) != 1 {
// 		return shim.Error("Incorrect number of arguments. Expecting 1")
// 	}

// 	key := args[0]
// 	encoded, err := stub.GetState(KEY_PREFIX + key)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}
// 	pubKey, err := stub.GetState(KEY_PUBLIC)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}
// 	privKey, err := stub.GetState(KEY_PRIVATE)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}
// 	helper, errs := crypto.NewRSAHelper(pubKey, privKey)
// 	if errs != nil && len(errs) > 0 {
// 		return shim.Error(errs[0].Error())
// 	}
// 	decoded, err := helper.Decrypt(string(encoded))
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	return shim.Success([]byte(decoded))
// }

func (t *BalanceManager) sendEvent(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	eventName := args[0]
	message := args[1]

	err := stub.SetEvent(eventName, []byte(message))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *BalanceManager) putPrivateData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	col := args[0]
	key := args[1]
	val := args[2]

	err := stub.PutPrivateData(col, key, []byte(val))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *BalanceManager) getPrivateData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	col := args[0]
	key := args[1]
	val, err := stub.GetPrivateData(col, key)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(val)
}

func main() {
	err := shim.Start(new(BalanceManager))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
		fmt.Println()
	}
}
