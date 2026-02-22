
	package main

	import (
		"bytes"
		"strconv"
		"errors"
		"crypto/sha256"
		"encoding/hex"
		"encoding/json"
		"fmt"
		"github.com/hyperledger/fabric/core/chaincode/shim"
		"github.com/hyperledger/fabric/protos/peer"
		"time"
	)
	
	//  Chaincode implementation
	type MedChain struct {

	}
	
	// history object struct
	type HistoryObject struct {
		TxId 		string 				`json:"TxId"`
		Value 		json.RawMessage 	`json:"Value"`
		IsDelete 	string 				`json:"IsDelete"`
		Timestamp 	string 				`json:"Timestamp"`
	}
		
	// Hospital struct
	type Hospital struct { 
		AssetType				string `json:"AssetType"`
		Hospital_ID				string `json:"Hospital_ID"`
		HospitalName				string `json:"HospitalName"`
		HospitalAddress				string `json:"HospitalAddress"`
		HospitalPhone				string `json:"HospitalPhone"`
	}
		
	// HospitalToPatient struct
	type HospitalToPatient struct { 
		AssetType				string `json:"AssetType"`
		HospitalToPatient_ID				string `json:"HospitalToPatient_ID"`
		PatientID				string `json:"PatientID"`
		PatientName				string `json:"PatientName"`
		HospitalID				string `json:"HospitalID"`
		HospitalName				string `json:"HospitalName"`
	}
		
	// ===================================================================================
	// Main
	// ===================================================================================
	func main() {
		err := shim.Start(new(MedChain))
		if err != nil {
			fmt.Printf("Error starting Xebest Trace chaincode: %s", err)
		}
	}

	// Init initializes chaincode
	// ===========================

	func (t *MedChain) Init(stub shim.ChaincodeStubInterface) peer.Response {
		return shim.Success([]byte("successful initialization"))
	}
	
	// Invoke - Our entry point for Invocations
	// ========================================
	func (t *MedChain) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
		function, args := stub.GetFunctionAndParameters()
		fmt.Println("\n\n------------------\n")
		fmt.Println("invoke is running -> "+ function)

		// Handle different functions
		
		if function == "queryAsset" { 
			return t.queryAsset(stub, args)
		} else if function == "getAll" { 
			return t.getAll(stub, args)
		} else if function == "getHistoryForRecord" { 
			return t.getHistoryForRecord(stub, args)
		} else if function == "addHospital" { 
			return t.addHospital(stub, args)
		} else if function == "updateHospital" { 
			return t.updateHospital(stub, args)
		} else if function == "getAllHospital" { 
			return t.getAllHospital(stub, args)
		} else if function == "addHospitalToPatient" { 
			return t.addHospitalToPatient(stub, args)
		} else if function == "updateHospitalToPatient" { 
			return t.updateHospitalToPatient(stub, args)
		} else if function == "getAllHospitalToPatient" { 
			return t.getAllHospitalToPatient(stub, args)
		}

		//error
		fmt.Println("invoke did not find func: " + function) 
		return shim.Error("Received unknown function invocation")
	}
		
	// create a record for Hospital
	func (t *MedChain) addHospital(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start addHospital")

		// check if all the args are send
		if len(args) != 3 {
			fmt.Println("Incorrect number of arguments, Required 3 arguments")
			return shim.Error("Incorrect number of arguments, Required 3 arguments")
		}

		// check if the args are empty
		for i := 0; i < len(args); i++ {
			if len(args[i]) <= 0 {
				fmt.Println("argument "+ string(i+1)  + " must be a non-empty string")
				return shim.Error("argument "+ string(i+1)  + " must be a non-empty string")
			}
		}

		// get timestamp
		txTimeAsPtr, errTx := t.GetTxTimestampChannel(stub)
		if errTx != nil {
			return shim.Error("Returning time error")
		}

		// create the object
		var obj = Hospital{}

		obj.Hospital_ID = ComputeHashKey(args[0]+(txTimeAsPtr.String()))
		obj.HospitalName = args[0]
		obj.HospitalAddress = args[1]
		obj.HospitalPhone = args[2]
		obj.AssetType = "Hospital"

		// convert to bytes
		assetAsBytes, errMarshal := json.Marshal(obj)
		
		// show error if failed
		if errMarshal != nil {
			return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
		}

		// save data in blockchain
		errPut := stub.PutState(obj.Hospital_ID, assetAsBytes)

		if errPut != nil {
			return shim.Error(fmt.Sprintf("Failed to save file data : %s", obj.Hospital_ID))
		}

		fmt.Println("Success in saving file data %v", obj)


		return shim.Success(assetAsBytes)
	}
		
	// update a record for Hospital
	func (t *MedChain) updateHospital(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start updateHospital")

		// check if all the args are send
		if len(args) != 4 {
			return shim.Error("Incorrect number of arguments, Required 4 arguments")
		}

		// check if the args are empty
		for i := 0; i < len(args); i++ {
			if len(args[i]) <= 0 {
				return shim.Error("argument "+ string(i+1)  + " must be a non-empty string")
			}
		}

		getAssetAsBytes, errT := stub.GetState(args[0])

		if errT != nil {
			return shim.Error(fmt.Sprintf("Error : Cannot find Hospital %s" , errT))
		}

		if getAssetAsBytes == nil {
			return shim.Error(fmt.Sprintf("Cannot find asset Hospital with ID %s" , args[0]))
		}

		var obj = Hospital{}

		json.Unmarshal(getAssetAsBytes, &obj)
		obj.HospitalName = args[1]
		obj.HospitalAddress = args[2]
		obj.HospitalPhone = args[3]
		comAssetAsBytes, errMarshal := json.Marshal(obj)

		if errMarshal != nil {
			return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
		}

		errPut := stub.PutState(obj.Hospital_ID, comAssetAsBytes)

		if errPut != nil {
			return shim.Error(fmt.Sprintf("Failed to update Hospital with ID %s", args[0]))
		}

		fmt.Println("Hospital asset with ID %s was updated \n %v", args[0], obj)

		return shim.Success(comAssetAsBytes)
	}
			
	// get all records for asset Hospital
	func (t *MedChain) getAllHospital(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start getAllHospital")

		AssetType := "Hospital"

		queryString := fmt.Sprintf("SELECT valueJson FROM <STATE> WHERE json_extract(valueJson, '$.AssetType') = '%s'", AssetType)
		
		queryResults, err := getQueryResultForQueryString(stub, queryString)
		
		if err != nil {
			return shim.Error(err.Error())
		}
		
		return shim.Success(queryResults)
	}
		
	// create a record for HospitalToPatient
	func (t *MedChain) addHospitalToPatient(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start addHospitalToPatient")

		// check if all the args are send
		if len(args) != 4 {
			fmt.Println("Incorrect number of arguments, Required 4 arguments")
			return shim.Error("Incorrect number of arguments, Required 4 arguments")
		}

		// check if the args are empty
		for i := 0; i < len(args); i++ {
			if len(args[i]) <= 0 {
				fmt.Println("argument "+ string(i+1)  + " must be a non-empty string")
				return shim.Error("argument "+ string(i+1)  + " must be a non-empty string")
			}
		}

		// get timestamp
		txTimeAsPtr, errTx := t.GetTxTimestampChannel(stub)
		if errTx != nil {
			return shim.Error("Returning time error")
		}

		// create the object
		var obj = HospitalToPatient{}

		obj.HospitalToPatient_ID = ComputeHashKey(args[0]+(txTimeAsPtr.String()))
		obj.PatientID = args[0]
		obj.PatientName = args[1]
		obj.HospitalID = args[2]
		obj.HospitalName = args[3]
		obj.AssetType = "HospitalToPatient"

		// convert to bytes
		assetAsBytes, errMarshal := json.Marshal(obj)
		
		// show error if failed
		if errMarshal != nil {
			return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
		}

		// save data in blockchain
		errPut := stub.PutState(obj.HospitalToPatient_ID, assetAsBytes)

		if errPut != nil {
			return shim.Error(fmt.Sprintf("Failed to save file data : %s", obj.HospitalToPatient_ID))
		}

		fmt.Println("Success in saving file data %v", obj)


		return shim.Success(assetAsBytes)
	}
		
	// update a record for HospitalToPatient
	func (t *MedChain) updateHospitalToPatient(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start updateHospitalToPatient")

		// check if all the args are send
		if len(args) != 5 {
			return shim.Error("Incorrect number of arguments, Required 5 arguments")
		}

		// check if the args are empty
		for i := 0; i < len(args); i++ {
			if len(args[i]) <= 0 {
				return shim.Error("argument "+ string(i+1)  + " must be a non-empty string")
			}
		}

		getAssetAsBytes, errT := stub.GetState(args[0])

		if errT != nil {
			return shim.Error(fmt.Sprintf("Error : Cannot find HospitalToPatient %s" , errT))
		}

		if getAssetAsBytes == nil {
			return shim.Error(fmt.Sprintf("Cannot find asset HospitalToPatient with ID %s" , args[0]))
		}

		var obj = HospitalToPatient{}

		json.Unmarshal(getAssetAsBytes, &obj)
		obj.PatientID = args[1]
		obj.PatientName = args[2]
		obj.HospitalID = args[3]
		obj.HospitalName = args[4]
		comAssetAsBytes, errMarshal := json.Marshal(obj)

		if errMarshal != nil {
			return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
		}

		errPut := stub.PutState(obj.HospitalToPatient_ID, comAssetAsBytes)

		if errPut != nil {
			return shim.Error(fmt.Sprintf("Failed to update HospitalToPatient with ID %s", args[0]))
		}

		fmt.Println("HospitalToPatient asset with ID %s was updated \n %v", args[0], obj)

		return shim.Success(comAssetAsBytes)
	}
			
	// get all records for asset HospitalToPatient
	func (t *MedChain) getAllHospitalToPatient(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start getAllHospitalToPatient")

		AssetType := "HospitalToPatient"

		queryString := fmt.Sprintf("SELECT valueJson FROM <STATE> WHERE json_extract(valueJson, '$.AssetType') = '%s'", AssetType)
		
		queryResults, err := getQueryResultForQueryString(stub, queryString)
		
		if err != nil {
			return shim.Error(err.Error())
		}
		
		return shim.Success(queryResults)
	}
		
	// get data using a record id 
	func (t *MedChain) queryAsset(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
		if len(args) != 1 {
			return shim.Error("Incorrect number of arguments, Required 1")
		}

		fmt.Println(fmt.Sprintf("- start queryAsset: %s\n", args))

		AssetAsBytes, _ := APIstub.GetState(args[0])

		if AssetAsBytes == nil {
			return shim.Error("Could not locate Asset")

		}
		
		return shim.Success(AssetAsBytes)
	}

	// get all records of any kind
	func (t *MedChain) getAll(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start getAll")

		// query 
		queryString := "SELECT blockNo, key, valueJson FROM <STATE> WHERE 1"
		
		queryResults, err := getQueryResultForQueryString(stub, queryString)
		
		if err != nil {
			return shim.Error(err.Error())
		}
		
		return shim.Success(queryResults)
	}

	// ===================================================================================
	// Retrieving all the changes to a record
	// ===================================================================================
	func (t *MedChain) getHistoryForRecord(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println(fmt.Sprintf("- start getHistoryForRecord: %s\n", args))

		var parentHistObj []HistoryObject

		transactionBytes := t.getHistory(stub, args[0])

		var data []HistoryObject
		json.Unmarshal(transactionBytes, &data)

		fmt.Println(fmt.Sprintf("Transaction History: %v \n", data))
		parentHistObj = append(parentHistObj,data...)

		parentHistArrBytes, _ := json.Marshal(parentHistObj)

		return shim.Success(parentHistArrBytes)
	}

	// helper to get all changes for a record
	func (t *MedChain) getHistory(stub shim.ChaincodeStubInterface, recordKey string) []byte {

		fmt.Printf("- start getHistory: %s\n", recordKey)

		resultsIterator, err := stub.GetHistoryForKey(recordKey)
		if err != nil {
			errors.New(err.Error())
		//	return shim.Error(err.Error())
		}
		defer resultsIterator.Close()

		// buffer is a JSON array containing historic values for the key/value pair
		var buffer bytes.Buffer
		buffer.WriteString("[")

		bArrayMemberAlreadyWritten := false
		for resultsIterator.HasNext() {
			response, err := resultsIterator.Next()
			if err != nil {
				errors.New(err.Error())
				//return shim.Error(err.Error())
			}
			// Add a comma before array members, suppress it for the first array member
			if bArrayMemberAlreadyWritten == true {
				buffer.WriteString(",")
			}
			buffer.WriteString("{\"TxId\":")
			buffer.WriteString("\"")
			buffer.WriteString(response.TxId)
			buffer.WriteString("\"")

			buffer.WriteString(", \"Value\":")
			// if it was a delete operation on given key, then we need to set the
			//corresponding value null. Else, we will write the response.Value
			//as-is (as the Value itself a JSON vehiclePart)
			if response.IsDelete {
				buffer.WriteString("null")
			} else {
				buffer.WriteString(string(response.Value))
			}

			buffer.WriteString(", \"Timestamp\":")
			buffer.WriteString("\"")
			buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
			buffer.WriteString("\"")

			buffer.WriteString(", \"IsDelete\":")
			buffer.WriteString("\"")
			buffer.WriteString(strconv.FormatBool(response.IsDelete))
			buffer.WriteString("\"")

			buffer.WriteString("}")
			bArrayMemberAlreadyWritten = true
		}
		buffer.WriteString("]")

		fmt.Printf("- getHistoryForRecord returning:\n%s\n", buffer.String())

		return buffer.Bytes()
	}

	// ===================================================================================
	// Helpers
	// ===================================================================================

	// helper for getting all rows from a query
	func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {	

		fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

		resultsIterator, err := stub.GetQueryResult(queryString)
		if err != nil {
			return nil, err
		}
		defer resultsIterator.Close()

		// buffer is a JSON array containing QueryRecords
		var buffer bytes.Buffer
		buffer.WriteString("[")

		bArrayMemberAlreadyWritten := false
		for resultsIterator.HasNext() {
			queryResponse, err := resultsIterator.Next()
			if err != nil {
				return nil, err
			}
			// Add a comma before array members, suppress it for the first array member
			if bArrayMemberAlreadyWritten == true {
				buffer.WriteString(",")
			}

			buffer.WriteString(string(queryResponse.Value))
			bArrayMemberAlreadyWritten = true
		}
		buffer.WriteString("]")

		fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

		return buffer.Bytes(), nil
	}

	// generate a random unique key hash
	func ComputeHashKey(propertyName string) string {
		h := sha256.New()
		h.Write([]byte(propertyName))
		nameInBytes := h.Sum([]byte(""))
		nameInString := hex.EncodeToString(nameInBytes)
		return  nameInString[:62]
	}

	// get timestamp in seconds
	func (t *MedChain) GetTxTimestampChannel(APIstub shim.ChaincodeStubInterface) (time.Time, error) {
		txTimeAsPtr, err := APIstub.GetTxTimestamp()
		if err != nil {
			fmt.Printf("Returning error in TimeStamp \n")
		}
		fmt.Printf("\t returned value from APIstub: %v\n", txTimeAsPtr)
		timeInt := time.Unix(txTimeAsPtr.Seconds, int64(txTimeAsPtr.Nanos))

		return timeInt, nil
	}
		