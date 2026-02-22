/*
Copyright IBM Corp. 2016 All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
		 http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"encoding/json"
	
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// PurchaseOrder example sample Chaincode implementation
type PurchaseOrder struct {
}

//initializes the chaincode
func (t *PurchaseOrder) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("PurchaseOrder Init")
	_, args := stub.GetFunctionAndParameters()

	if len(args) < 0 {
		return shim.Error("Incorrect number of arguments. Expecting 0")
	}
		
	return shim.Success(nil)
}

//check whether string has value or not
func getSafeString(input interface{}) string {
	var safeValue string
	var isOk bool

	if input == nil {
		safeValue = ""
	} else {
		safeValue, isOk = input.(string)
		if isOk == false {
			safeValue = ""
		}
	}
	return safeValue
}

/*following arguments are passed in PurchaseOrder
			param:
				1. POID(*unique)
				2. quantity
				3. productName
				4. customer
				5. vendor
				6. address
				7. status
				8. price

*/
//customized function to createPO
func (t *PurchaseOrder) createPO(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//checking the number of argument
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	recBytes := args[0]
	
	var POMap map[string]interface{}
	POMap = make(map[string]interface{}, 0)
	err := json.Unmarshal([]byte(recBytes), &POMap)
	if err != nil {
		return shim.Error("Failed to unmarshal recBytes")
}

	//getting POID as parameter from the arguments to check whether PO exist
	POID := getSafeString(POMap["POID"])

	//Check if POID already exists
	fetchedPODetails, err := stub.GetState("POID:" + POID)
	if err != nil {
		return shim.Error("Failed to get POID details: " + err.Error())
	} else if fetchedPODetails != nil {
		fmt.Println("This POID already exists: " + POID)
		return shim.Error("This POID already exists: " + POID)
}

	//Store the PurchaseOrder data onto the blockchain
	outputMapBytes, _ := json.Marshal(POMap)
		
	//Store the new records
		stub.PutState("POID:" + POID, outputMapBytes)
		if err != nil {
			return shim.Error(err.Error())
	}
	fmt.Println("create purchase order")
	return shim.Success([]byte("SUCCESS"))
}

	//get PurchaseOrder
	func (t *PurchaseOrder) getPODetails(stub shim.ChaincodeStubInterface, args []string) pb.Response {

		//checking the number of argument
		if len(args) < 1 {
			return shim.Error("Incorrect number of arguments. Expecting 1")
		}
		var POResultMap map[string]interface{}
	
		recBytes := args[0]
		err := json.Unmarshal([]byte(recBytes), &POResultMap)
		if err != nil {
			return shim.Error("Failed to unmarshal POReultMap")
	}
	    POID := getSafeString(POResultMap["POID"])
	 	
		//Check if POID already exists
		PODetails, err := stub.GetState("POID:" + POID)
		if err != nil {
			return shim.Error("Failed to get POID details: " + POID)
		} else if PODetails == nil {
			fmt.Println("This POID doesnot exists: " + POID)
			return shim.Error("This POID doesnot exists: " + POID)
	}
	fmt.Println("PODetails:", string(PODetails))
	return shim.Success(PODetails)
}
//update the status of the PO
func (t *PurchaseOrder) updateStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//checking the number of argument
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	recBytes := args[0]

	var PORecordMap map[string]interface{}

	err := json.Unmarshal([]byte(recBytes), &PORecordMap)
	if err != nil {
		return shim.Error("Failed to unmarshal recBytes")
	}

	//==== Check if PO already exists ====
	fetchedPODetails, err := stub.GetState("POID:" + getSafeString(PORecordMap["POID"]))
	if err != nil {
		return shim.Error("Failed to get PO details: " + err.Error())
	} else if fetchedPODetails == nil {
		fmt.Println("This PO does not exists:" + getSafeString(PORecordMap["POID"]))
		return shim.Error("This PO does not exists:" + getSafeString(PORecordMap["POID"]))
	}

	var POMap map[string]interface{}
	err = json.Unmarshal(fetchedPODetails, &POMap)
	if err != nil {
		return shim.Error("Failed to unmarshal item")
	}
	//get status from the arguments
	POMap["status"] = getSafeString(PORecordMap["status"])

	outputMapBytes, _ := json.Marshal(POMap)

	//Store the records
	stub.PutState("POID:"+getSafeString(POMap["POID"]), outputMapBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("updateStatus purchase order")
	return shim.Success([]byte("SUCCESS"))
}
//update the productName of the PO
func (t *PurchaseOrder) updateproductName(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//checking the number of argument
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	recBytes := args[0]

	var PORecordMap map[string]interface{}

	err := json.Unmarshal([]byte(recBytes), &PORecordMap)
	if err != nil {
		return shim.Error("Failed to unmarshal recBytes")
	}

	//==== Check if PO already exists ====
	fetchedPODetails, err := stub.GetState("POID:" + getSafeString(PORecordMap["POID"]))
	if err != nil {
		return shim.Error("Failed to get PO details: " + err.Error())
	} else if fetchedPODetails == nil {
		fmt.Println("This PO does not exists:" + getSafeString(PORecordMap["POID"]))
		return shim.Error("This PO does not exists:" + getSafeString(PORecordMap["POID"]))
	}

	var POMap map[string]interface{}
	err = json.Unmarshal(fetchedPODetails, &POMap)
	if err != nil {
		return shim.Error("Failed to unmarshal item")
	}
	//get status from the arguments
	POMap["productName"] = getSafeString(PORecordMap["productName"])

	outputMapBytes, _ := json.Marshal(POMap)

	//Store the records
	stub.PutState("POID:"+getSafeString(POMap["POID"]), outputMapBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("updateproductName purchase order")
	return shim.Success([]byte("SUCCESS"))
}
//update the customer of the PO
func (t *PurchaseOrder) updateCustomer(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//checking the number of argument
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	recBytes := args[0]

	var PORecordMap map[string]interface{}

	err := json.Unmarshal([]byte(recBytes), &PORecordMap)
	if err != nil {
		return shim.Error("Failed to unmarshal recBytes")
	}

	//==== Check if PO already exists ====
	fetchedPODetails, err := stub.GetState("POID:" + getSafeString(PORecordMap["POID"]))
	if err != nil {
		return shim.Error("Failed to get PO details: " + err.Error())
	} else if fetchedPODetails == nil {
		fmt.Println("This PO does not exists:" + getSafeString(PORecordMap["POID"]))
		return shim.Error("This PO does not exists:" + getSafeString(PORecordMap["POID"]))
	}

	var POMap map[string]interface{}
	err = json.Unmarshal(fetchedPODetails, &POMap)
	if err != nil {
		return shim.Error("Failed to unmarshal item")
	}
	//get status from the arguments
	POMap["customer"] = getSafeString(PORecordMap["customer"])

	outputMapBytes, _ := json.Marshal(POMap)

	//Store the records
	stub.PutState("POID:"+getSafeString(POMap["POID"]), outputMapBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("updatecustomer purchase order")
	return shim.Success([]byte("SUCCESS"))
}
//update the quantity of the PO
func (t *PurchaseOrder) updateQuantity(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//checking the number of argument
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	recBytes := args[0]

	var PORecordMap map[string]interface{}

	err := json.Unmarshal([]byte(recBytes), &PORecordMap)
	if err != nil {
		return shim.Error("Failed to unmarshal recBytes")
	}

	//==== Check if PO already exists ====
	fetchedPODetails, err := stub.GetState("POID:" + getSafeString(PORecordMap["POID"]))
	if err != nil {
		return shim.Error("Failed to get PO details: " + err.Error())
	} else if fetchedPODetails == nil {
		fmt.Println("This PO does not exists:" + getSafeString(PORecordMap["POID"]))
		return shim.Error("This PO does not exists:" + getSafeString(PORecordMap["POID"]))
	}

	var POMap map[string]interface{}
	err = json.Unmarshal(fetchedPODetails, &POMap)
	if err != nil {
		return shim.Error("Failed to unmarshal item")
	}
	//get status from the arguments
	POMap["quantity"] = getSafeString(PORecordMap["quantity"])

	outputMapBytes, _ := json.Marshal(POMap)

	//Store the records
	stub.PutState("POID:"+getSafeString(POMap["POID"]), outputMapBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("updateQuantity purchase order")
	return shim.Success([]byte("SUCCESS"))
}
//deletes the PO
func (t *PurchaseOrder) deletePO(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	//checking the number of argument
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	recBytes := args[0]

	var PODeleteMap map[string]interface{}

	err := json.Unmarshal([]byte(recBytes), &PODeleteMap)
	if err != nil {
		return shim.Error("Failed to unmarshal recBytes")
	}
	
	POID := getSafeString(PODeleteMap["POID"])

	//==== Check if PO already exists ====
	fetchedPODetails, err := stub.GetState("POID:" + POID)
	if err != nil {
		return shim.Error("Failed to get PO details: " + err.Error())
	} else if fetchedPODetails == nil {
		fmt.Println("This PO does not exists:" + POID)
		return shim.Error("This PO does not exists:" + POID)
	}
	err =stub.DelState(getSafeString(PODeleteMap["POID"]))

	if err != nil {
		return shim.Error("Failed to delete PO details: " + err.Error())
	
		}
		fmt.Println("delete purchase order")
         return shim.Success([]byte("SUCCESS"))
}

	
	func (t *PurchaseOrder) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("PurchaseOrder chaincode Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "createPO" {
		// creates PO
		return t.createPO(stub, args)
	} else if function == "getPODetails" {
		// gives details of PO based on ID
		return t.getPODetails(stub, args)
    }   else if function == "updateStatus" {
	// updates status
	    return t.updateStatus(stub, args)
	}   else if function == "updateQuantity" {
		// updates quantity
		return t.updateQuantity(stub, args)
	}   else if function == "updateproductName" {
		// updates productName
		return t.updateproductName(stub, args)
	}   else if function == "updateCustomer" {
		// updates customer
		return t.updateCustomer(stub, args)
	}   else if function == "deletePO" {
	// deletePO
	return t.deletePO(stub, args)
    }  
    return shim.Error("Invalid invoke function name.")
    }
func main() {
	err := shim.Start(new(PurchaseOrder))
	if err != nil {
		fmt.Printf("Error starting PurchaseOrder chaincode: %s", err)
	}
}

