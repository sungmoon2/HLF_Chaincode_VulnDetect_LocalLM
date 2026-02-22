package main

import (

	"github.com/hyperledger/shim"
)

func Invoke( stub shim.ChaincodeStubInterface ) {

	s := ""
	kvs := map[string]string{"a":"hello", "b":"world!"}
	for k,v := range kvs{
	   s += v
	}
	
	stub.PutState("key", []byte(s))
}

func main() {
}
