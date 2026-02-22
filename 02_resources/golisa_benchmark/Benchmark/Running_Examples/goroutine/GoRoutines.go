package main

import (
	"container/list"
	"sync"
	"fmt"
    "github.com/hyperledger/shim"
)

var wg sync.WaitGroup

func Invoke( stub shim.ChaincodeStubInterface ) {

	wg.Add(2)
	total := 10000
	s := ""

	go func() {
		defer wg.Done()
		for idx := 1; idx <= total; idx++ {
			s += "0"
		}
	}()

	go func() {
		defer wg.Done()
		for idx := 1; idx <= total; idx++ {
			s += "1"
		}
	}()

	wg.Wait()
		
	stub.PutState("key", []byte(s))
}

func main() {
    

}
