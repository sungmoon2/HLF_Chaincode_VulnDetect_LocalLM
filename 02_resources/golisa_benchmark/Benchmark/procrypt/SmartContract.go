package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"fmt"
	"github.com/hyperledger/fabric/protos/peer"
	"encoding/json"
)

type SmartContract struct {

}

type CarInfo struct {
	Owner    string `json:"owner"`
	Company  string `json:"company"`
	Model 	 string `json:"model"`
	Price 	 string `json:"price"`
	Location string `json:"location"`
}


func(s *SmartContract) Init(API shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(API shim.ChaincodeStubInterface) peer.Response {
	function, args :=  API.GetFunctionAndParameters()
	if function == "buyCar" {
		return s.buyCar(API, args)
	} else if function == "sellCar" {
		return s.sellCar(API, args)
	} else if function == "queryCar" {
		return s.queryCar(API, args)
	}
	return shim.Success(nil)
}

func(s *SmartContract) buyCar(API shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 6 {
		return shim.Error(fmt.Sprintf("Incorrect number of arguments. Expecting 6"))
	}
	car := CarInfo{Owner: args[1], Company:args[2], Model:args[3], Price:args[4], Location:args[5]}
	carAsBytes, _ := json.Marshal(car)
	err := API.PutState(args[0], carAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("error writting to ledger, %s", err))
	}
	return shim.Success(carAsBytes)
}

func(s *SmartContract) sellCar(API shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error(fmt.Sprintf("Incorrect number of arguments. Expecting 2"))
	}
	carAsBytes,_ := API.GetState(args[0])
	if carAsBytes == nil {
		return shim.Error("could not locate car")
	}
	car := CarInfo{}
	json.Unmarshal(carAsBytes, &car)
	car.Owner = args[1]
	carAsBytes, _ = json.Marshal(car)
	err := API.PutState(args[0], carAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error selling car!, %s", err))
	}
	fmt.Printf("car successfully sold to %s\n", args[1])
	return shim.Success(nil)
}

func (s *SmartContract) queryCar(API shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args)!= 1 {
		return shim.Error(fmt.Sprintf("Incorrect number of arguments. Expecting 1"))
	}
	carAsBytes, err := API.GetState(args[0])
	if err != nil {
		return shim.Error(fmt.Sprintf("Error retrieving information about the car, %s", err))
	}
	if carAsBytes == nil {
		return shim.Error("Data not found")
	}
	car := CarInfo{}
	err = json.Unmarshal(carAsBytes, &car)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error umarshaling data, %s", err))
	}
	fmt.Println(car)
	return shim.Success(nil)
}

func main() {
	if err := shim.Start(new(SmartContract)); err != nil {
		fmt.Errorf("error starting chaincode %s", err)
	}
}
