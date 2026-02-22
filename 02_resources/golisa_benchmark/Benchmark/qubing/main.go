package main

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/chaincode/shim/ext/entities"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// Cryptor handler of encryption and decryption
type Cryptor struct {
	bccspInstance bccsp.BCCSP
	entity        entities.Encrypter
}

// NewDefaultCryptor generate instance of Cryptor with Default BCCSP
func NewDefaultCryptor(commonKey string) *Cryptor {
	factory.InitFactories(nil)
	cryptor := Cryptor{bccspInstance: factory.GetDefault()}
	entity, err := entities.NewAES256EncrypterEntity(commonKey, cryptor.bccspInstance, make([]byte, 32), make([]byte, 16))
	if err != nil {
		cryptor.entity = entity
	}
	return &cryptor
}

// GetDecryptedState read state from ledger and decrypt it
func (cryptor *Cryptor) GetDecryptedState(stub shim.ChaincodeStubInterface, key string) ([]byte, error) {
	// at first we retrieve the ciphertext from the ledger
	ciphertext, err := stub.GetState(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) == 0 {
		return nil, errors.New("no ciphertext to decrypt")
	}

	return cryptor.entity.Decrypt(ciphertext)
}

// PutEncryptedState encrypt data and put into ledger
func (cryptor *Cryptor) PutEncryptedState(stub shim.ChaincodeStubInterface, key string, value []byte) error {
	// at first we use the supplied entity to encrypt the value
	ciphertext, err := cryptor.entity.Encrypt(value)
	if err != nil {
		return err
	}

	return stub.PutState(key, ciphertext)
}

// SmartContract chaincode definition
type SmartContract struct {
}

// Init Init
func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke Invoke
func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	function, _ := stub.GetFunctionAndParameters()

	if function == "put" {
		return s.doPut(stub)
	} else if function == "get" {
		return s.doGet(stub)
	}
	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) doPut(stub shim.ChaincodeStubInterface) pb.Response {
	value := "WakandaForeva"
	key := "testId"
	cryptor := NewDefaultCryptor("HELLO")
	err := cryptor.PutEncryptedState(stub, key, []byte(value))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (s *SmartContract) doGet(stub shim.ChaincodeStubInterface) pb.Response {
	key := "testId"
	cryptor := NewDefaultCryptor("HELLO")
	val, err := cryptor.GetDecryptedState(stub, key)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(val)
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error starting SmartContract chaincode: %s", err)
	}
}
