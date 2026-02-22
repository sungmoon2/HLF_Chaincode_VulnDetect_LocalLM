package main                                                                                                                         

import (
    "fmt"
    "github.com/hyperledger/fabric/core/chaincode/shim"
    "github.com/hyperledger/fabric/protos/peer"
)

type SimpleStore struct {
}

func (t *SimpleStore) Init(stub shim.ChaincodeStubInterface) peer.Response {

    //Get the args from the transaction proposal
    args := stub.GetStringArgs()
    args = args[1:]

    if len(args) % 2 != 0 {
        return shim.Error("Incorrect arguments. Expecting a key value pair")
    }

    //init key value
    for index := 0; index < len(args) ; index += 2{
        err := stub.PutState(args[index], []byte(args[index+1]))
        if err != nil {
            return shim.Error("Failed to set init state")
        }

    }

    return shim.Success(nil)

}

func (t *SimpleStore) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
    function, args := stub.GetFunctionAndParameters()

    switch function {
        case "get":
            return t.get(stub, args)
        case "set":
            return t.set(stub, args)
        case "delete":
           return t.delete(stub, args)
        default:
            return shim.Error("Invalid function")
    }

}

//query data from ledger
func (t *SimpleStore) get(stub shim.ChaincodeStubInterface, args []string) peer.Response {
    if len(args) != 1 {
        return shim.Error("Incorrect arguments. Expecting a key")
    }

    bytes, err := stub.GetState(args[0])

    if err != nil {
        return shim.Error("Failed to get state")
    }

    if bytes == nil {
        return shim.Error("Nil value")
    }

    return shim.Success(bytes);
}

//put key to ledger
func (t *SimpleStore) set(stub shim.ChaincodeStubInterface, args []string) peer.Response {
    if len(args) != 2 {
        return shim.Error("Incorrect number of args")
    }

    err := stub.PutState(args[0], []byte(args[1]))

    if err != nil {
        return shim.Error("Failed to set state")
    }

    return shim.Success(nil)
}

//delete by key
func (t *SimpleStore) delete(stub shim.ChaincodeStubInterface, args []string) peer.Response {
    if len(args) < 1 {
        return shim.Error("Incorrect arguments. Expecting a key")
    }

    for _ ,element:= range args{
        err := stub.DelState(element)
        if err != nil {
            return shim.Error("Failed to delete state")
        }
    }

    return shim.Success(nil)
}


func main() {
    err := shim.Start(new(SimpleStore))

    if err != nil {
        fmt.Printf("Error starting chaincode: %s", err)
    }
}
