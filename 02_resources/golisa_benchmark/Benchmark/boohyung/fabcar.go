/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/*
 * The sample smart contract for documentation topic:
 * Writing Your First Blockchain Application
 */

package main

/* Imports
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	// 지원 기능: 체인코드의 원장 및 트랜잭션에 액세스, 다른 체인코드를 호출
	"github.com/hyperledger/fabric/core/chaincode/shim" 
	pb "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type FabCarChaincode struct {
}

// Define the car structure, with 4 properties.  Structure tags are used by encoding/json library
type Car struct {
	Make   string `json:"make"`
	Model  string `json:"model"`
	Colour string `json:"colour"`
	Owner  string `json:"owner"`
}

/*
 * The Init method is called when the Smart Contract "fabcar" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
 // 인스턴스화할 때 실행되는 함수
func (t *FabCarChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "fabcar"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
 // Function name에 따라 필요한 함수 호출
 // shim.ChaincodeStubInterface 인터페이스는 메서드의 집합(https://godoc.org/gopkg.in/hyperledger/fabric.v1/core/chaincode/shim#ChaincodeStubInterface) 
func (t *FabCarChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	fnc := string(stub.GetArgs()[0])
	switch fnc {
	case "createCar":
		return Invoke(stub, t.createCar)
	case "queryCar":
		return Invoke(stub, t.queryCar)
	case "initLedger":
		return Invoke_no_arg(stub, t.initLedger)
	case "queryAllCars":
		return Invoke_no_arg(stub, t.queryAllCars)
	case "changeCarOwner":
		return Invoke(stub, t.changeCarOwner)
	}

	// Retrieve the requested Smart Contract function and arguments
	// 2개의 리턴값을 가진 GetFunctionAndParameters()
	// function: 함수명, args: 인자
	// function, args := stub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	// if function == "queryCar" {
	// 	return t.queryCar(stub, args)
	// } else if function == "initLedger" {
	// 	return t.initLedger(stub)
	// } else if function == "createCar" {
	// 	return t.createCar(stub, args)
	// } else if function == "queryAllCars" {
	// 	return t.queryAllCars(stub)
	// } else if function == "changeCarOwner" {
	// 	return t.changeCarOwner(stub, args)
	// }

	return shim.Error("Invalid Smart Contract function name.")
}
// 파라미터: 함수 이름, 해당 함수의 인자
func (t *FabCarChaincode) queryCar(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	// string으로 입력받은 첫 번째 인자값을 통해 원장 데이터를 read
	// 리턴: 데이터, 에러
	carAsBytes, _ := stub.GetState(args[0])
	return shim.Success(carAsBytes)
}

func (t *FabCarChaincode) initLedger(stub shim.ChaincodeStubInterface) pb.Response {
	// Car 구조체 배열 cars 초기화
	cars := []Car{
		Car{Make: "Toyota", Model: "Prius", Colour: "blue", Owner: "Tomoko"},
		Car{Make: "Ford", Model: "Mustang", Colour: "red", Owner: "Brad"},
		Car{Make: "Hyundai", Model: "Tucson", Colour: "green", Owner: "Jin Soo"},
		Car{Make: "Volkswagen", Model: "Passat", Colour: "yellow", Owner: "Max"},
		Car{Make: "Tesla", Model: "S", Colour: "black", Owner: "Adriana"},
		Car{Make: "Peugeot", Model: "205", Colour: "purple", Owner: "Michel"},
		Car{Make: "Chery", Model: "S22L", Colour: "white", Owner: "Aarav"},
		Car{Make: "Fiat", Model: "Punto", Colour: "violet", Owner: "Pari"},
		Car{Make: "Tata", Model: "Nano", Colour: "indigo", Owner: "Valeria"},
		Car{Make: "Holden", Model: "Barina", Colour: "brown", Owner: "Shotaro"},
	}

	i := 0
	for i < len(cars) {
		fmt.Println("i is ", i)
		// json.Marshal(): 정수형이나 구조체 같은 값을 로우 바이트로 변환
		carAsBytes, _ := json.Marshal(cars[i])
		// 원장에 데이터 쓰기 (key: CARi, value: cars[i])
		// strconv.Itoa(): 숫자를 문자열로 변환
		stub.PutState("CAR"+strconv.Itoa(i), carAsBytes)
		fmt.Println("Added", cars[i])
		i = i + 1
	}

	return shim.Success(nil)
}

func (t *FabCarChaincode) createCar(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println(stub, args)
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}
	
	var car = Car{Make: args[1], Model: args[2], Colour: args[3], Owner: args[4]}

	carAsBytes, _ := json.Marshal(car)
	// 원장에 데이터 쓰기(key: args[0], value: carAsBytes)
	// key 예: "CAR1"
	// value 예: "Hyundai","Genesis","Black","Boohyung"
	stub.PutState(args[0], carAsBytes)

	return shim.Success(nil)
}

func (t *FabCarChaincode) queryAllCars(stub shim.ChaincodeStubInterface) pb.Response {

	startKey := "CAR0"
	// 끝을 모름
	endKey := "CAR999"
	// iteratorInterface 타입의 변수를 리턴
	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	// Java에서 finally같은 역할. 함수가 리턴값을 반환하기 전에 호출되는 함수
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	// 버퍼 생성
	var buffer bytes.Buffer
	// 버퍼에 문자열 입력(덧붙이기)
	buffer.WriteString("[")
	// 데이터가 이미 배열에 있는지 확인
	bArrayMemberAlreadyWritten := false
	// HasNext(): 다음 요소가 있는지 판별
	for resultsIterator.HasNext() {
		// Next(): 다음으로 이동
		// 다음 배열의 데이터를 읽어옴
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllCars:\n%s\n", buffer.String())
	// Bytes(): 버퍼를 바이트로 변환
	return shim.Success(buffer.Bytes())
}

func (t *FabCarChaincode) changeCarOwner(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	// key를 입력받아 value를 바이트 형태로 리턴
	carAsBytes, _ := stub.GetState(args[0])
	car := Car{}

	// json.Unmarshal(): 로우 바이트나 문자열를 구조체로 변경
	json.Unmarshal(carAsBytes, &car)
	car.Owner = args[1]

	carAsBytes, _ = json.Marshal(car)
	stub.PutState(args[0], carAsBytes)

	return shim.Success(nil)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	// FabCarChainCode 객체를 입력으로 전달하여 스마트 컨트랙트를 실행
	err := shim.Start(new(FabCarChaincode))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
