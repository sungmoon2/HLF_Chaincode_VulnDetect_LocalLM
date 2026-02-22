package main

import (

    "github.com/hyperledger/shim"
)

func Invoke( stub shim.ChaincodeStubInterface ) {

    c := make(chan int)

    go myroutine1(c)
    go myroutine2(c)
    
    x, y := <- c, <- c
    
    stub.PutState("key", []byte(x))
}

func myroutine1(mychannel chan int) { 
   //do something ...
}

func myroutine2(mychannel chan int) {
   //do something ... 
}

func main() {
    

}
