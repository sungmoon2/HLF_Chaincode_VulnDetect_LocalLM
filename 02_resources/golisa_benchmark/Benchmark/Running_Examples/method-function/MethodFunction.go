package main

import (
	"time"
	"github.com/hyperledger/shim"
)

func Invoke( stub shim.ChaincodeStubInterface ) {

    key := "key"
    tm  := time.Now()
    
    stub.PutState(key, []byte(tm))
	
}

func main() {
}
