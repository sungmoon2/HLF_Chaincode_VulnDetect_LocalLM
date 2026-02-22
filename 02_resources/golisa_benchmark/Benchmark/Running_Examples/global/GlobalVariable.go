package main

import (
	"github.com/hyperledger/shim"
)

var glob string

func inc() {
    glob += "a"
}

func Invoke( stub shim.ChaincodeStubInterface ) {

    // do something ...
    
    transaction(strub)
	
}

func transaction( stub shim.ChaincodeStubInterface ) {
    stub.PutState("key", []byte(glob))
}

//other functions ...

func main() {
}
