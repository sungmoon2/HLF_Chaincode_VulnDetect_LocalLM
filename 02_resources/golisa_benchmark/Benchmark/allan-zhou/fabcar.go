package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type Chaincode struct{}

type Car struct {
	Id     int    `json:"id"`
	Make   string `json:"make"`
	Model  string `json:"model"`
	Colour string `json:"colour"`
	Owner  string `json:"owner"`
}

func (cc *Chaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (cc *Chaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()

	if fn == "initLedger" {
		return cc.initLedger(stub, args)
	} else if fn == "queryAllCars" {
		return cc.queryAllCars(stub, args)
	} else if fn == "queryCarById" {
		return cc.queryCarById(stub, args)
	} else if fn == "changeCarOwner" {
		return cc.changeCarOwner(stub, args)
	} else if fn == "addCar" {
		return cc.addCar(stub, args)
	} else if fn == "deleteCarById" {
		return cc.deleteCarById(stub, args)
	}

	return shim.Success(nil)
}

func (cc *Chaincode) initLedger(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	cars := []Car{
		Car{Id: 1, Make: "Toyota", Model: "Prius", Colour: "blue", Owner: "zhangsan"},
		Car{Id: 2, Make: "Ford", Model: "Mustang", Colour: "red", Owner: "lisi"},
		Car{Id: 3, Make: "Hyundai", Model: "Tucson", Colour: "green", Owner: "wangwu"},
		Car{Id: 4, Make: "Volkswagen", Model: "Passat", Colour: "yellow", Owner: "zhangsan"},
		Car{Id: 5, Make: "Tesla", Model: "S", Colour: "black", Owner: "lisi"},
		Car{Id: 6, Make: "Peugeot", Model: "205", Colour: "purple", Owner: "wangwu"},
		Car{Id: 7, Make: "Chery", Model: "S22L", Colour: "white", Owner: "zhangsan"},
		Car{Id: 8, Make: "Fiat", Model: "Punto", Colour: "violet", Owner: "lisi"},
	}

	for _, car := range cars {
		carAsBytes, _ := json.Marshal(car)
		stub.PutState(strconv.Itoa(car.Id), carAsBytes)
		fmt.Printf("add car: %v", car)
	}

	return shim.Success(nil)
}

func (cc *Chaincode) queryAllCars(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	startKey := "1"
	endKey := "999"
	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// 生产使用bytes.Buffer
	strs := "["
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		strs += "{\"key\":"
		strs += "\""
		strs += queryResult.Key
		strs += "\","
		strs += "\"value\":"
		strs += string(queryResult.Value)
		strs += "},"
	}
	strs = strs[:len(strs)-1]
	strs += "]"

	fmt.Printf("queryAllCars:\n %s \n", strs)

	return shim.Success([]byte(strs))
}

func (cc *Chaincode) addCar(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expect 5")
	}

	carId ,err:= strconv.Atoi(args[0])
	if err != nil {
		return shim.Error("Incorrect Car Id. Expect a integer value")
	}

	car := Car{Id: carId, Make: args[1], Model: args[2], Colour: args[3], Owner: args[4]}
	carAsBytes, _ := json.Marshal(car)
	stub.PutState(strconv.Itoa(carId), carAsBytes)
	
	return shim.Success(nil)
}

func (cc *Chaincode) queryCarById(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expect 1")
	}

	valueAsBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	if valueAsBytes == nil {
		return shim.Error("nil is found")
	}
	return shim.Success(valueAsBytes)
}

func (cc *Chaincode) changeCarOwner(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expect 2")
	}

	carId, err := strconv.Atoi(args[0])
	if err != nil {
	 	return	shim.Error("Incorrect Car Id. Expect a integer value")
	}

	carAsBytes, err := stub.GetState(strconv.Itoa(carId))
	if err != nil {
		return shim.Error(err.Error())
	}

	var car Car
	json.Unmarshal(carAsBytes, &car)
	car.Owner = args[1]

	newCarAsBytes, _ := json.Marshal(car)
	stub.PutState(strconv.Itoa(carId), newCarAsBytes)
	
	return shim.Success(nil)
}

func (cc *Chaincode) deleteCarById(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expect 1")
	}

	carId, err := strconv.Atoi(args[0])
	if err != nil {
		return shim.Error("Incorrent Car Id. Expect a integer value")
	}

	stub.DelState(strconv.Itoa(carId))
	
	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(Chaincode))
	if err != nil {
		fmt.Printf("error start: %s", err)
	}
}
