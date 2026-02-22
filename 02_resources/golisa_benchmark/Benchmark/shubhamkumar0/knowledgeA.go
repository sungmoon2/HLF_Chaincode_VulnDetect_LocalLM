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
	
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)
//var emptySpace []int
//var count int=0
// Define the Smart Contract structure
type SmartContract struct {
}

// Define the car structure, with 4 properties.  Structure tags are used by encoding/json library
type Concept struct {
	Name string `json:"name"`
	//cnt int `json:"count"` 
	Incoming  []string `json:"incoming"`
	Outgoing  []string `json:"outgoing"`
	//Colour string `json:"colour"`
	//Owner  string `json:"owner"`
}


/*
 * The Init method is called when the Smart Contract "fabcar" is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */



func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func indexOf(element string, data []string) (int) {
   for k, v := range data {
       if element == v {
           return k
       }
   }
   return -1    //not found.
}

func remove(s []string, i int) []string {
    s[len(s)-1], s[i] = s[i], s[len(s)-1]
    return s[:len(s)-1]
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "fabcar"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryconcept" {
		//function == "queryConcept"
		return s.queryconcept(APIstub, args)
		//return s.queryConcept(APIstub, args)
	}else if function == "initLedger" {
		return s.initLedger(APIstub)
	}else if function == "addconcept" {
		
		return s.addconcept(APIstub, args)
	}else if function == "addrelation" {

		return s.addrelation(APIstub, args)
	}else if function == "deleteconcept" {
		return s.deleteconcept(APIstub, args)
	}else if function == "deleterelation" {
		return s.deleterelation(APIstub, args)
	}



	return shim.Error("Invalid Smart Contract function name.")
}



//argumnts can b name or uniquekey
func (s *SmartContract) queryconcept(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	conceptAsBytes, _ := APIstub.GetState(args[0])
	
	return shim.Success(conceptAsBytes)
}


func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {

	concepts :=[]Concept{
	   	Concept{Name: "A", Incoming: []string{}, Outgoing: []string{"CONCEPT1", "CONCEPT3", "CONCEPT2"}},
	   	Concept{Name: "B", Incoming: []string{"CONCEPT0"}, Outgoing: []string{"CONCEPT3", "CONCEPT4"}},
	   	Concept{Name: "C", Incoming: []string{"CONCEPT0"}, Outgoing: []string{"CONCEPT3"}},
	   	Concept{Name: "D", Incoming: []string{"CONCEPT0", "CONCEPT1", "CONCEPT2"}, Outgoing: []string{"CONCEPT5"}},
	   	Concept{Name: "E", Incoming: []string{"CONCEPT1"}, Outgoing: []string{"CONCEPT7"}},
	   	Concept{Name: "F", Incoming: []string{"CONCEPT3", "CONCEPT9"}, Outgoing: []string{"CONCEPT6"}},
	   	Concept{Name: "G", Incoming: []string{"CONCEPT5"}, Outgoing: []string{"CONCEPT9"}},
	   	Concept{Name: "H", Incoming: []string{"CONCEPT4"}, Outgoing: []string{}},
	   	Concept{Name: "I", Incoming: []string{}, Outgoing: []string{}},
	   	Concept{Name: "J", Incoming: []string{"CONCEPT6"}, Outgoing: []string{"CONCEPT5"}},
	}
	
	i := 0
	for i < len(concepts) {
		fmt.Println("i is ", i)
		conceptAsBytes, _ := json.Marshal(concepts[i])
		APIstub.PutState("CONCEPT"+strconv.Itoa(i), conceptAsBytes)
		//keep check of what you are putting PutState

//do cnt ++ HERE

		fmt.Println("Added", concepts[i])
		i = i + 1
	}

	return shim.Success(nil)
}

	
func (s *SmartContract) addconcept(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
//check if exists
     _, err := APIstub.GetState(args[0])
	if err == nil {
		return shim.Error("Concept already exist: " + err.Error())
	}
      	
    var concept = Concept{Name: args[1], Incoming: []string{}, Outgoing: []string{}}

	conceptAsBytes, _ := json.Marshal(concept)
	//fconceptAsBytes, err := APIstub.GetState(args[0])
	//if err == nil {
	APIstub.PutState(args[0], conceptAsBytes)

	//}
	//putstate with cnt and emptyspace filter
	return shim.Success(nil)
	
}

func (s *SmartContract) addrelation(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) !=2 {
	    return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	// should exist
	
	

	
		fconceptAsBytes, _ := APIstub.GetState(args[0])
	
		tconceptAsBytes, _ := APIstub.GetState(args[1])
		
		
			//establishing relation
			concept:= Concept{}
			json.Unmarshal(tconceptAsBytes, &concept)
			concept.Incoming=append(concept.Incoming, args[0])
			tconceptAsBytes, _ = json.Marshal(concept)
			APIstub.PutState(args[1], tconceptAsBytes)
			
			concept1:= Concept{}
			json.Unmarshal(fconceptAsBytes, &concept1)
			concept1.Outgoing=append(concept1.Outgoing, args[1])
			fconceptAsBytes, _ = json.Marshal(concept1)
			APIstub.PutState(args[0], fconceptAsBytes)
		/*		
		}else if tconceptAsBytes != nil {
			//establishing relation
			
			concept:= Concept{}
			json.Unmarshal(tconceptAsBytes, &concept)
			concept.Incoming=append(concept.Incoming, args[0])
			tconceptAsBytes, _ = json.Marshal(concept)
			APIstub.PutState(args[1], tconceptAsBytes)
			
			concept1:= Concept{}
			json.Unmarshal(fconceptAsBytes, &concept1)
			concept1.Outgoing=append(concept1.Outgoing, args[1])
			fconceptAsBytes, _ = json.Marshal(concept1)
			APIstub.PutState(args[0], fconceptAsBytes)

		}	

	} else if fconceptAsBytes != nil {
		
		tconceptAsBytes, err := APIstub.GetState(args[1])
		if err != nil {
			return shim.Error("Concept doesnot exist: " + err.Error())

			//writing toC to transaction's write-set
			//use addconcept()
			//APIstub.InvokeChaincode(addconcept, []string{"CONCEPT"+strconv.Itoa(cnt),args[3]})
			//APIstub.PutState(args[1], tconceptAsBytes)

			//establishing relation
			concept:= Concept{}
			json.Unmarshal(tconceptAsBytes, &concept)
			concept.Incoming=append(concept.Incoming, args[0])
			tconceptAsBytes, _ = json.Marshal(concept)
			APIstub.PutState(args[1], tconceptAsBytes)
			
			concept1:= Concept{}
			json.Unmarshal(fconceptAsBytes, &concept1)
			concept1.Outgoing=append(concept1.Outgoing, args[1])
			fconceptAsBytes, _ = json.Marshal(concept1)
			APIstub.PutState(args[0], fconceptAsBytes)

		}else if tconceptAsBytes != nil {
			//establishing relation
			
			concept:= Concept{}
			json.Unmarshal(tconceptAsBytes, &concept)
			concept.Incoming=append(concept.Incoming, args[0])
			tconceptAsBytes, _ = json.Marshal(concept)
			APIstub.PutState(args[1], tconceptAsBytes)
			
			concept1:= Concept{}
			json.Unmarshal(fconceptAsBytes, &concept1)
			concept1.Outgoing=append(concept1.Outgoing, args[1])
			fconceptAsBytes, _ = json.Marshal(concept1)
			APIstub.PutState(args[0], fconceptAsBytes)

		}	
	}*/
	return shim.Success(nil)
}

func (s *SmartContract) deleteconcept(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	conceptAsBytes, err := APIstub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
		}

	// search & delete from other structs
		
		concept := Concept{}
		json.Unmarshal(conceptAsBytes, &concept)

		for i:=0; i<len(concept.Incoming); i++{
			cAsBytes, _ :=APIstub.GetState(concept.Incoming[i])             //how to direct to UNIQUEKEY by concept.incoming
			cpt := Concept{}
			json.Unmarshal(cAsBytes, &cpt)
			index:=indexOf(args[0],cpt.Outgoing)
			cpt.Outgoing=remove(cpt.Outgoing,index)
			cptAsBytes, _ := json.Marshal(cpt)
			APIstub.PutState(concept.Incoming[i], cptAsBytes)
		}
		for i:=0; i<len(concept.Outgoing); i++{
			cAsBytes, _ :=APIstub.GetState(concept.Outgoing[i])
			cpt := Concept{}
			json.Unmarshal(cAsBytes, &cpt)
			index1:=indexOf(args[0],cpt.Incoming)
			cpt.Incoming=remove(cpt.Incoming,index1)
			cptAsBytes, _ := json.Marshal(cpt)
			APIstub.PutState(concept.Outgoing[i], cptAsBytes)
		}
		

		err0 := APIstub.DelState(args[0]) 
		if err0 != nil {
		return shim.Error("Failed to delete state:" + err0.Error())
		}


	return shim.Success(nil)

}

func (s *SmartContract) deleterelation(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	fonceptAsBytes, _ := APIstub.GetState(args[0])
	concept:= Concept{}
	json.Unmarshal(fonceptAsBytes, &concept)
	index:=indexOf(args[1],concept.Outgoing)
	concept.Outgoing=remove(concept.Outgoing,index)
	foncptAsBytes, _ := json.Marshal(concept)
	APIstub.PutState(args[0], foncptAsBytes)


	tonceptAsBytes, _ := APIstub.GetState(args[1])
	concept1:= Concept{}
	json.Unmarshal(tonceptAsBytes, &concept1)
	index1:=indexOf(args[0],concept1.Incoming)
	concept1.Incoming=remove(concept1.Incoming,index1)
	toncptAsBytes, _ := json.Marshal(concept1)
	APIstub.PutState(args[1], toncptAsBytes)

	return shim.Success(nil)

}






func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
