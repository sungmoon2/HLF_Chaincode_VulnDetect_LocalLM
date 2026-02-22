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
 * 5 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

// Define the car structure, with 4 properties.  Structure tags are used by encoding/json library
// type Car struct {
// 	Make   string `json:"make"`
// 	Model  string `json:"model"`
// 	Colour string `json:"colour"`
// 	Owner  string `json:"owner"`
// }

// Record structure, with 11 properties.  Structure tags are used by encoding/json library
type Record struct {
	Sid               string `json:sid`
	FullName          string `json:fullName`
	Level             string `json:level`
	StarEarned        string `json:starEarned`
	Logins            string `json:logins`
	Listen            string `json:listen`
	Read              string `json:read`
	Worksheet         string `json:worksheet`
	Quiz              string `json:quiz`
	PassedQuizCount   string `json:passedQuizCount`
	PracticeRecording string `json:practiceRecording`
}

/*
 * The Init method is called when the Smart Contract "opprecord" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	fmt.Println("ex02 Init")
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "opprecord"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryRecord" {
		return s.queryRecord(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "createRecord" {
		return s.createRecord(APIstub, args)
	}
	//  else if function == "queryAllRecordss" {
	// 	return s.queryAllCars(APIstub)
	// }

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryRecord(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	var sid = args[0]
	recordAsBytes, _ := APIstub.GetState(sid)
	return shim.Success(recordAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	var jsonResp string

	records := []Record{
		Record{Sid: "243506374", FullName: "Dong hoon", Level: "A", StarEarned: "4320", Logins: "9", Listen: "23", Read: "24", Worksheet: "0", Quiz: "44", PassedQuizCount: "25", PracticeRecording: "0"},
		Record{Sid: "243016624", FullName: "Kang yukyung", Level: "A", StarEarned: "5300", Logins: "25", Listen: "48", Read: "44", Worksheet: "0", Quiz: "34", PassedQuizCount: "22", PracticeRecording: "7"},
		Record{Sid: "243016627", FullName: "Kang Denis", Level: "A", StarEarned: "6580", Logins: "14", Listen: "37", Read: "36", Worksheet: "0", Quiz: "82", PassedQuizCount: "48", PracticeRecording: "1"},
		Record{Sid: "241649412", FullName: "Jaebin", Level: "B", StarEarned: "3470", Logins: "10", Listen: "24", Read: "24", Worksheet: "0", Quiz: "28", PassedQuizCount: "16", PracticeRecording: "4"},
		Record{Sid: "241392752", FullName: "Ye Seo", Level: "C", StarEarned: "13470", Logins: "35", Listen: "131", Read: "162", Worksheet: "0", Quiz: "139", PassedQuizCount: "61", PracticeRecording: "16"},
		Record{Sid: "241392846", FullName: "Erika", Level: "A", StarEarned: "4610", Logins: "18", Listen: "27", Read: "38", Worksheet: "0", Quiz: "73", PassedQuizCount: "32", PracticeRecording: "0"},
		Record{Sid: "241392874", FullName: "Yeon Woo", Level: "A", StarEarned: "2490", Logins: "11", Listen: "21", Read: "25", Worksheet: "0", Quiz: "32", PassedQuizCount: "18", PracticeRecording: "0"},
		Record{Sid: "241391885", FullName: "Ye Joon", Level: "C", StarEarned: "23200", Logins: "49", Listen: "162", Read: "145", Worksheet: "0", Quiz: "271", PassedQuizCount: "162", PracticeRecording: "15"},
		Record{Sid: "232199169", FullName: "Tan", Level: "B", StarEarned: "1680", Logins: "9", Listen: "8", Read: "12", Worksheet: "0", Quiz: "26", PassedQuizCount: "13", PracticeRecording: "0"},
		Record{Sid: "232199170", FullName: "jiwook kim", Level: "A", StarEarned: "2830", Logins: "8", Listen: "12", Read: "15", Worksheet: "0", Quiz: "44", PassedQuizCount: "27", PracticeRecording: "3"},
		Record{Sid: "232199171", FullName: "junhyub park", Level: "B", StarEarned: "8540", Logins: "15", Listen: "41", Read: "42", Worksheet: "0", Quiz: "96", PassedQuizCount: "53", PracticeRecording: "14"},
		Record{Sid: "232199174", FullName: "hyejung park", Level: "E", StarEarned: "9380", Logins: "41", Listen: "44", Read: "136", Worksheet: "0", Quiz: "344", PassedQuizCount: "123", PracticeRecording: "10"},
		Record{Sid: "232199181", FullName: "Bladic", Level: "C", StarEarned: "13500", Logins: "14", Listen: "85", Read: "82", Worksheet: "0", Quiz: "105", PassedQuizCount: "66", PracticeRecording: "6"},
		Record{Sid: "232199182", FullName: "Sun Ha", Level: "C", StarEarned: "8280", Logins: "20", Listen: "60", Read: "87", Worksheet: "0", Quiz: "181", PassedQuizCount: "91", PracticeRecording: "11"},
		Record{Sid: "232199184", FullName: "sungsoo moon", Level: "E", StarEarned: "12960", Logins: "36", Listen: "98", Read: "68", Worksheet: "0", Quiz: "208", PassedQuizCount: "103", PracticeRecording: "5"},
		Record{Sid: "232173421", FullName: "Hana", Level: "B", StarEarned: "7030", Logins: "18", Listen: "35", Read: "46", Worksheet: "0", Quiz: "50", PassedQuizCount: "30", PracticeRecording: "0"},
		Record{Sid: "232170417", FullName: "kyuhyung lee", Level: "B", StarEarned: "950", Logins: "11", Listen: "23", Read: "22", Worksheet: "0", Quiz: "9", PassedQuizCount: "4", PracticeRecording: "0"},
		Record{Sid: "232170418", FullName: "nayeon park", Level: "B", StarEarned: "43270", Logins: "52", Listen: "237", Read: "313", Worksheet: "0", Quiz: "488", PassedQuizCount: "256", PracticeRecording: "133"},
		Record{Sid: "232170420", FullName: "minji kang", Level: "B", StarEarned: "9280", Logins: "17", Listen: "41", Read: "46", Worksheet: "0", Quiz: "78", PassedQuizCount: "44", PracticeRecording: "1"},
		Record{Sid: "232170422", FullName: "youngmi kim", Level: "C", StarEarned: "10140", Logins: "33", Listen: "47", Read: "88", Worksheet: "0", Quiz: "215", PassedQuizCount: "93", PracticeRecording: "42"},
		Record{Sid: "232170423", FullName: "jaesung kim", Level: "B", StarEarned: "6950", Logins: "23", Listen: "37", Read: "40", Worksheet: "0", Quiz: "81", PassedQuizCount: "44", PracticeRecording: "2"},
		Record{Sid: "232170426", FullName: "hyunji park", Level: "B", StarEarned: "9960", Logins: "22", Listen: "64", Read: "51", Worksheet: "0", Quiz: "157", PassedQuizCount: "80", PracticeRecording: "0"},
		Record{Sid: "232170427", FullName: "hyewon shin", Level: "C", StarEarned: "6990", Logins: "20", Listen: "35", Read: "37", Worksheet: "0", Quiz: "103", PassedQuizCount: "50", PracticeRecording: "6"},
		Record{Sid: "232170428", FullName: "hyeeun shin", Level: "A", StarEarned: "3500", Logins: "9", Listen: "4", Read: "17", Worksheet: "0", Quiz: "34", PassedQuizCount: "17", PracticeRecording: "8"},
		Record{Sid: "232170429", FullName: "hyejin shin", Level: "A", StarEarned: "390", Logins: "3", Listen: "9", Read: "3", Worksheet: "0", Quiz: "7", PassedQuizCount: "4", PracticeRecording: "2"},
		Record{Sid: "232170430", FullName: "hosung lee", Level: "A", StarEarned: "6410", Logins: "30", Listen: "55", Read: "54", Worksheet: "0", Quiz: "102", PassedQuizCount: "61", PracticeRecording: "0"},
		Record{Sid: "232170431", FullName: "hyokeun kim", Level: "A", StarEarned: "4340", Logins: "26", Listen: "55", Read: "32", Worksheet: "0", Quiz: "65", PassedQuizCount: "44", PracticeRecording: "21"},
	}

	i := 0
	for i < len(records) {
		fmt.Println("i is ", i)
		recordAsBytes, err := json.Marshal(records[i])

		if err != nil {
			jsonResp = "{\"Error\":\"Failed to get state for " + records[i].Sid + "\"}"
			return shim.Error(jsonResp)
		} else if recordAsBytes == nil {
			jsonResp = "{\"Error\":\"Marble does not exist: " + records[i].Sid + "\"}"
			return shim.Error(jsonResp)
		}

		APIstub.PutState(records[i].Sid, recordAsBytes)
		fmt.Println("Added", records[i])
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) createRecord(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	var jsonResp string

	if len(args) != 11 {
		return shim.Error("Incorrect number of arguments. Expecting 11")
	}

	//  var car = Car{Make: args[1], Model: args[2], Colour: args[3], Owner: args[])}
	// Record{Sid: "232170431", FullName: "hyokeun kim", Level: "A", StarEarned: 4340, Logins: 26, Listen: 55, Read: 32, Worksheet: 0, Quiz: 65, PassedQuizCount: 44, PracticeRecording: 21}
	var record = Record{Sid: args[0], FullName: args[1], Level: args[2], StarEarned: args[3], Logins: args[4], Listen: args[5], Read: args[6], Worksheet: args[7], Quiz: args[8], PassedQuizCount: args[9], PracticeRecording: args[10]}

	recordAsBytes, err := json.Marshal(record)

	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + args[0] + "\"}"
		return shim.Error(jsonResp)
	} else if recordAsBytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + args[0] + "\"}"
		return shim.Error(jsonResp)
	}

	APIstub.PutState(args[0], recordAsBytes)

	return shim.Success(recordAsBytes)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
