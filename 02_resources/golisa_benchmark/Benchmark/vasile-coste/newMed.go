
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
	type MedicalApp struct {

	}
	
	// history object struct
	type HistoryObject struct {
		TxId 		string 				`json:"TxId"`
		Value 		json.RawMessage 	`json:"Value"`
		IsDelete 	string 				`json:"IsDelete"`
		Timestamp 	string 				`json:"Timestamp"`
	}
		
	// Patient struct
	type Patient struct { 
		AssetType				string `json:"AssetType"`
		Patient_ID				string `json:"Patient_ID"`
		Username				string `json:"Username"`
		Password				string `json:"Password"`
		FirstName				string `json:"FirstName"`
		LastName				string `json:"LastName"`
		FatherName				string `json:"FatherName"`
		BrithDate				string `json:"BrithDate"`
		Gender				string `json:"Gender"`
		Ssn				string `json:"Ssn"`
		ParentSsn				string `json:"ParentSsn"`
	}
		
	// Doctor struct
	type Doctor struct { 
		AssetType				string `json:"AssetType"`
		Doctor_ID				string `json:"Doctor_ID"`
		Username				string `json:"Username"`
		Password				string `json:"Password"`
		FirstName				string `json:"FirstName"`
		LastName				string `json:"LastName"`
		FatherName				string `json:"FatherName"`
		BrithDate				string `json:"BrithDate"`
		Gender				string `json:"Gender"`
		ClinicName				string `json:"ClinicName"`
		Specialization				string `json:"Specialization"`
	}
		
	// Records struct
	type Records struct { 
		AssetType				string `json:"AssetType"`
		Records_ID				string `json:"Records_ID"`
		Doctor_ID				string `json:"Doctor_ID"`
		Patient_ID				string `json:"Patient_ID"`
		Status				string `json:"Status"`
		ExaminationName				string `json:"ExaminationName"`
		Result				string `json:"Result"`
		Notes				string `json:"Notes"`
		DateStarted				string `json:"DateStarted"`
		DateCompleted				string `json:"DateCompleted"`
		Link				string `json:"Link"`
		ClinicName				string `json:"ClinicName"`
		Patient_Name				string `json:"Patient_Name"`
		Doctor_Name				string `json:"Doctor_Name"`
	}
		
	// RequestAccess struct
	type RequestAccess struct { 
		AssetType				string `json:"AssetType"`
		RequestAccess_ID				string `json:"RequestAccess_ID"`
		Doctor_ID				string `json:"Doctor_ID"`
		Patient_ID				string `json:"Patient_ID"`
		AccessReqFrom				string `json:"AccessReqFrom"`
		Status				string `json:"Status"`
		Name				string `json:"Name"`
		ClinicName				string `json:"ClinicName"`
	}
		
	// AccessControl struct
	type AccessControl struct { 
		AssetType				string `json:"AssetType"`
		AccessControl_ID				string `json:"AccessControl_ID"`
		Doctor_ID				string `json:"Doctor_ID"`
		Patient_ID				string `json:"Patient_ID"`
		IsValid				string `json:"IsValid"`
		ValidTill				string `json:"ValidTill"`
		GrandedAccessOn				string `json:"GrandedAccessOn"`
		Patient_Name				string `json:"Patient_Name"`
		ClinicName				string `json:"ClinicName"`
		Doctor_Name				string `json:"Doctor_Name"`
	}
		
	// ===================================================================================
	// Main
	// ===================================================================================
	func main() {
		err := shim.Start(new(MedicalApp))
		if err != nil {
			fmt.Printf("Error starting Xebest Trace chaincode: %s", err)
		}
	}

	// Init initializes chaincode
	// ===========================

	func (t *MedicalApp) Init(stub shim.ChaincodeStubInterface) peer.Response {
		return shim.Success([]byte("successful initialization"))
	}
	
	// Invoke - Our entry point for Invocations
	// ========================================
	func (t *MedicalApp) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
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
		} else if function == "addPatient" { 
			return t.addPatient(stub, args)
		} else if function == "loginPatient" { 
			return t.loginPatient(stub, args)
		} else if function == "updatePatientCrd" { 
			return t.updatePatientCrd(stub, args)
		} else if function == "updatePatient" { 
			return t.updatePatient(stub, args)
		} else if function == "getAllPatient" { 
			return t.getAllPatient(stub, args)
		} else if function == "addDoctor" { 
			return t.addDoctor(stub, args)
		} else if function == "loginDoctor" { 
			return t.loginDoctor(stub, args)
		} else if function == "updateDoctorCrd" { 
			return t.updateDoctorCrd(stub, args)
		} else if function == "updateDoctor" { 
			return t.updateDoctor(stub, args)
		} else if function == "getAllDoctor" { 
			return t.getAllDoctor(stub, args)
		} else if function == "addRecords" { 
			return t.addRecords(stub, args)
		} else if function == "updateRecords" { 
			return t.updateRecords(stub, args)
		} else if function == "getAllRecords" { 
			return t.getAllRecords(stub, args)
		} else if function == "addRequestAccess" { 
			return t.addRequestAccess(stub, args)
		} else if function == "updateRequestAccess" { 
			return t.updateRequestAccess(stub, args)
		} else if function == "getAllRequestAccess" { 
			return t.getAllRequestAccess(stub, args)
		} else if function == "addAccessControl" { 
			return t.addAccessControl(stub, args)
		} else if function == "updateAccessControl" { 
			return t.updateAccessControl(stub, args)
		} else if function == "getAllAccessControl" { 
			return t.getAllAccessControl(stub, args)
		}

		//error
		fmt.Println("invoke did not find func: " + function) 
		return shim.Error("Received unknown function invocation")
	}
		
	// create a record for Patient
	func (t *MedicalApp) addPatient(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start addPatient")

		// check if all the args are send
		if len(args) != 9 {
			fmt.Println("Incorrect number of arguments, Required 9 arguments")
			return shim.Error("Incorrect number of arguments, Required 9 arguments")
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
		var obj = Patient{}

		obj.Patient_ID = ComputeHashKey(args[0]+(txTimeAsPtr.String()))
		obj.FirstName = args[0]
		obj.LastName = args[1]
		obj.FatherName = args[2]
		obj.BrithDate = args[3]
		obj.Gender = args[4]
		obj.Ssn = args[5]
		obj.ParentSsn = args[6]
		obj.Username = args[7]
		obj.Password = args[8]
		obj.AssetType = "Patient"

		// convert to bytes
		assetAsBytes, errMarshal := json.Marshal(obj)
		
		// show error if failed
		if errMarshal != nil {
			return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
		}

		// save data in blockchain
		errPut := stub.PutState(obj.Patient_ID, assetAsBytes)

		if errPut != nil {
			return shim.Error(fmt.Sprintf("Failed to save file data : %s", obj.Patient_ID))
		}

		fmt.Println("Success in saving file data %v", obj)


		return shim.Success(assetAsBytes)
	}

	// login asset Patient
	func (t *MedicalApp) loginPatient(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start loginPatient")
		
		// check if all the args are send
		if len(args) != 2 {
			return shim.Error("Incorrect number of arguments, Required 2 arguments")
		}

		// check if the args are empty
		for i := 0; i < len(args); i++ {
			if len(args[i]) <= 0 {
				return shim.Error("argument "+ string(i+1)  + " must be a non-empty string")
			}
		}

		AssetType := "Patient"

		queryString := fmt.Sprintf("SELECT valueJson FROM <STATE> WHERE json_extract(valueJson, '$.AssetType', '$.Username', '$.Password') = '[\"%s\",\"%s\",\"%s\"]'", AssetType, args[0], args[1])
		
		queryResults, err := getQueryResultForQueryString(stub, queryString)
		
		if err != nil {
			return shim.Error(err.Error())
		}
		
		return shim.Success(queryResults)
	}

	// update a record for Patient credentials
	func (t *MedicalApp) updatePatientCrd(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start updatePatientCrd")

		// check if all the args are send
		if len(args) != 3 {
			return shim.Error("Incorrect number of arguments, Required 3 arguments")
		}

		// check if the args are empty
		for i := 0; i < len(args); i++ {
			if len(args[i]) <= 0 {
				return shim.Error("argument "+ string(i+1)  + " must be a non-empty string")
			}
		}

		getAssetAsBytes, errT := stub.GetState(args[0])

		if errT != nil {
			return shim.Error(fmt.Sprintf("Error : Cannot find Patient %s" , errT))
		}

		if getAssetAsBytes == nil {
			return shim.Error(fmt.Sprintf("Cannot find asset Patient with ID %s" , args[0]))
		}

		var obj = Patient{}

		json.Unmarshal(getAssetAsBytes, &obj)
		obj.Username = args[1]
		obj.Password = args[2]
		comAssetAsBytes, errMarshal := json.Marshal(obj)

		if errMarshal != nil {
			return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
		}

		errPut := stub.PutState(obj.Patient_ID, comAssetAsBytes)

		if errPut != nil {
			return shim.Error(fmt.Sprintf("Failed to update Patient with ID %s", args[0]))
		}

		fmt.Println("Patient asset with ID %s was updated \n %v", args[0], obj)

		return shim.Success(comAssetAsBytes)
	}
		
	// update a record for Patient
	func (t *MedicalApp) updatePatient(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start updatePatient")

		// check if all the args are send
		if len(args) != 8 {
			return shim.Error("Incorrect number of arguments, Required 8 arguments")
		}

		// check if the args are empty
		for i := 0; i < len(args); i++ {
			if len(args[i]) <= 0 {
				return shim.Error("argument "+ string(i+1)  + " must be a non-empty string")
			}
		}

		getAssetAsBytes, errT := stub.GetState(args[0])

		if errT != nil {
			return shim.Error(fmt.Sprintf("Error : Cannot find Patient %s" , errT))
		}

		if getAssetAsBytes == nil {
			return shim.Error(fmt.Sprintf("Cannot find asset Patient with ID %s" , args[0]))
		}

		var obj = Patient{}

		json.Unmarshal(getAssetAsBytes, &obj)
		obj.FirstName = args[1]
		obj.LastName = args[2]
		obj.FatherName = args[3]
		obj.BrithDate = args[4]
		obj.Gender = args[5]
		obj.Ssn = args[6]
		obj.ParentSsn = args[7]
		comAssetAsBytes, errMarshal := json.Marshal(obj)

		if errMarshal != nil {
			return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
		}

		errPut := stub.PutState(obj.Patient_ID, comAssetAsBytes)

		if errPut != nil {
			return shim.Error(fmt.Sprintf("Failed to update Patient with ID %s", args[0]))
		}

		fmt.Println("Patient asset with ID %s was updated \n %v", args[0], obj)

		return shim.Success(comAssetAsBytes)
	}
			
	// get all records for asset Patient
	func (t *MedicalApp) getAllPatient(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start getAllPatient")

		AssetType := "Patient"

		queryString := fmt.Sprintf("SELECT valueJson FROM <STATE> WHERE json_extract(valueJson, '$.AssetType') = '%s'", AssetType)
		
		queryResults, err := getQueryResultForQueryString(stub, queryString)
		
		if err != nil {
			return shim.Error(err.Error())
		}
		
		return shim.Success(queryResults)
	}
		
	// create a record for Doctor
	func (t *MedicalApp) addDoctor(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start addDoctor")

		// check if all the args are send
		if len(args) != 9 {
			fmt.Println("Incorrect number of arguments, Required 9 arguments")
			return shim.Error("Incorrect number of arguments, Required 9 arguments")
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
		var obj = Doctor{}

		obj.Doctor_ID = ComputeHashKey(args[0]+(txTimeAsPtr.String()))
		obj.FirstName = args[0]
		obj.LastName = args[1]
		obj.FatherName = args[2]
		obj.BrithDate = args[3]
		obj.Gender = args[4]
		obj.ClinicName = args[5]
		obj.Specialization = args[6]
		obj.Username = args[7]
		obj.Password = args[8]
		obj.AssetType = "Doctor"

		// convert to bytes
		assetAsBytes, errMarshal := json.Marshal(obj)
		
		// show error if failed
		if errMarshal != nil {
			return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
		}

		// save data in blockchain
		errPut := stub.PutState(obj.Doctor_ID, assetAsBytes)

		if errPut != nil {
			return shim.Error(fmt.Sprintf("Failed to save file data : %s", obj.Doctor_ID))
		}

		fmt.Println("Success in saving file data %v", obj)


		return shim.Success(assetAsBytes)
	}
			
	// login asset Doctor
	func (t *MedicalApp) loginDoctor(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start loginDoctor")
		
		// check if all the args are send
		if len(args) != 2 {
			return shim.Error("Incorrect number of arguments, Required 2 arguments")
		}

		// check if the args are empty
		for i := 0; i < len(args); i++ {
			if len(args[i]) <= 0 {
				return shim.Error("argument "+ string(i+1)  + " must be a non-empty string")
			}
		}

		AssetType := "Doctor"

		queryString := fmt.Sprintf("SELECT valueJson FROM <STATE> WHERE json_extract(valueJson, '$.AssetType', '$.Username', '$.Password') = '[\"%s\",\"%s\",\"%s\"]'", AssetType, args[0], args[1])
		
		queryResults, err := getQueryResultForQueryString(stub, queryString)
		
		if err != nil {
			return shim.Error(err.Error())
		}
		
		return shim.Success(queryResults)
	}
		
	// update a record for Doctor
	func (t *MedicalApp) updateDoctorCrd(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start updateDoctorCrd")

		// check if all the args are send
		if len(args) != 3 {
			return shim.Error("Incorrect number of arguments, Required 3 arguments")
		}

		// check if the args are empty
		for i := 0; i < len(args); i++ {
			if len(args[i]) <= 0 {
				return shim.Error("argument "+ string(i+1)  + " must be a non-empty string")
			}
		}

		getAssetAsBytes, errT := stub.GetState(args[0])

		if errT != nil {
			return shim.Error(fmt.Sprintf("Error : Cannot find Doctor %s" , errT))
		}

		if getAssetAsBytes == nil {
			return shim.Error(fmt.Sprintf("Cannot find asset Doctor with ID %s" , args[0]))
		}

		var obj = Doctor{}

		json.Unmarshal(getAssetAsBytes, &obj)
		obj.Username = args[1]
		obj.Password = args[2]
		comAssetAsBytes, errMarshal := json.Marshal(obj)

		if errMarshal != nil {
			return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
		}

		errPut := stub.PutState(obj.Doctor_ID, comAssetAsBytes)

		if errPut != nil {
			return shim.Error(fmt.Sprintf("Failed to update Doctor with ID %s", args[0]))
		}

		fmt.Println("Doctor asset with ID %s was updated \n %v", args[0], obj)

		return shim.Success(comAssetAsBytes)
	}
		
	// update a record for Doctor
	func (t *MedicalApp) updateDoctor(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start updateDoctor")

		// check if all the args are send
		if len(args) != 8 {
			return shim.Error("Incorrect number of arguments, Required 8 arguments")
		}

		// check if the args are empty
		for i := 0; i < len(args); i++ {
			if len(args[i]) <= 0 {
				return shim.Error("argument "+ string(i+1)  + " must be a non-empty string")
			}
		}

		getAssetAsBytes, errT := stub.GetState(args[0])

		if errT != nil {
			return shim.Error(fmt.Sprintf("Error : Cannot find Doctor %s" , errT))
		}

		if getAssetAsBytes == nil {
			return shim.Error(fmt.Sprintf("Cannot find asset Doctor with ID %s" , args[0]))
		}

		var obj = Doctor{}

		json.Unmarshal(getAssetAsBytes, &obj)
		obj.FirstName = args[1]
		obj.LastName = args[2]
		obj.FatherName = args[3]
		obj.BrithDate = args[4]
		obj.Gender = args[5]
		obj.ClinicName = args[6]
		obj.Specialization = args[7]
		comAssetAsBytes, errMarshal := json.Marshal(obj)

		if errMarshal != nil {
			return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
		}

		errPut := stub.PutState(obj.Doctor_ID, comAssetAsBytes)

		if errPut != nil {
			return shim.Error(fmt.Sprintf("Failed to update Doctor with ID %s", args[0]))
		}

		fmt.Println("Doctor asset with ID %s was updated \n %v", args[0], obj)

		return shim.Success(comAssetAsBytes)
	}
			
	// get all records for asset Doctor
	func (t *MedicalApp) getAllDoctor(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start getAllDoctor")

		AssetType := "Doctor"

		queryString := fmt.Sprintf("SELECT valueJson FROM <STATE> WHERE json_extract(valueJson, '$.AssetType') = '%s'", AssetType)
		
		queryResults, err := getQueryResultForQueryString(stub, queryString)
		
		if err != nil {
			return shim.Error(err.Error())
		}
		
		return shim.Success(queryResults)
	}
		
	// create a record for Records
	func (t *MedicalApp) addRecords(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start addRecords")

		// check if all the args are send
		if len(args) != 12 {
			fmt.Println("Incorrect number of arguments, Required 12 arguments")
			return shim.Error("Incorrect number of arguments, Required 12 arguments")
		}

		// check if the args are empty
		for i := 0; i < 8; i++ {
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
		var obj = Records{}

		obj.Records_ID = ComputeHashKey(args[0]+(txTimeAsPtr.String()))
		obj.Doctor_ID = args[0]
		obj.Patient_ID = args[1]
		obj.ExaminationName = args[2]
		obj.DateStarted = args[3]
		obj.ClinicName = args[4]
		obj.Patient_Name = args[5]
		obj.Doctor_Name = args[6]
		obj.Status = args[7]
		if args[8] == "" {
			obj.DateCompleted = args[3]
		} else {
			obj.DateCompleted = args[8]
		}
		obj.Result = args[9]
		obj.Notes = args[10]
		obj.Link = args[11]
		obj.AssetType = "Records"

		// convert to bytes
		assetAsBytes, errMarshal := json.Marshal(obj)
		
		// show error if failed
		if errMarshal != nil {
			return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
		}

		// save data in blockchain
		errPut := stub.PutState(obj.Records_ID, assetAsBytes)

		if errPut != nil {
			return shim.Error(fmt.Sprintf("Failed to save file data : %s", obj.Records_ID))
		}

		fmt.Println("Success in saving file data %v", obj)


		return shim.Success(assetAsBytes)
	}
		
	// update a record for Records
	func (t *MedicalApp) updateRecords(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start updateRecords")

		// check if all the args are send
		if len(args) != 6 {
			return shim.Error("Incorrect number of arguments, Required 6 arguments")
		}

		// check if the args are empty
		for i := 0; i < 3; i++ {
			if len(args[i]) <= 0 {
				return shim.Error("argument "+ string(i+1)  + " must be a non-empty string")
			}
		}

		getAssetAsBytes, errT := stub.GetState(args[0])

		if errT != nil {
			return shim.Error(fmt.Sprintf("Error : Cannot find Records %s" , errT))
		}

		if getAssetAsBytes == nil {
			return shim.Error(fmt.Sprintf("Cannot find asset Records with ID %s" , args[0]))
		}

		var obj = Records{}

		json.Unmarshal(getAssetAsBytes, &obj)
		obj.Status = args[1]
		obj.Result = args[2]
		obj.Notes = args[3]
		obj.Link = args[4]
		if args[5] != "" {
			obj.DateCompleted = args[5]
		}
		comAssetAsBytes, errMarshal := json.Marshal(obj)

		if errMarshal != nil {
			return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
		}

		errPut := stub.PutState(obj.Records_ID, comAssetAsBytes)

		if errPut != nil {
			return shim.Error(fmt.Sprintf("Failed to update Records with ID %s", args[0]))
		}

		fmt.Println("Records asset with ID %s was updated \n %v", args[0], obj)

		return shim.Success(comAssetAsBytes)
	}
			
	// get all records for asset Records
	func (t *MedicalApp) getAllRecords(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start getAllRecords")

		// check if all the args are send
		if len(args) != 1 {
			return shim.Error("Incorrect number of arguments, Required 1 arguments")
		}
		
		// check if the args are empty
		for i := 0; i < len(args); i++ {
			if len(args[i]) <= 0 {
				return shim.Error("argument "+ string(i+1)  + " must be a non-empty string")
			}
		}

		AssetType := "Records"

		queryString := fmt.Sprintf("SELECT valueJson FROM <STATE> WHERE json_extract(valueJson, '$.AssetType','$.Patient_ID') = '[\"%s\",\"%s\"]'", AssetType, args[0])
		

		queryResults, err := getQueryResultForQueryString(stub, queryString)
		
		if err != nil {
			return shim.Error(err.Error())
		}
		
		return shim.Success(queryResults)
	}
		
	// create a record for RequestAccess
	func (t *MedicalApp) addRequestAccess(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start addRequestAccess")

		// check if all the args are send
		if len(args) != 6 {
			fmt.Println("Incorrect number of arguments, Required 6 arguments")
			return shim.Error("Incorrect number of arguments, Required 6 arguments")
		}

		// check if the args are empty
		for i := 0; i < 5; i++ {
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
		var obj = RequestAccess{}

		obj.RequestAccess_ID = ComputeHashKey(args[0]+(txTimeAsPtr.String()))
		obj.Doctor_ID = args[0]
		obj.Patient_ID = args[1]
		obj.AccessReqFrom = args[2]
		obj.Status = args[3]
		obj.Name = args[4]
		obj.ClinicName = args[5]
		obj.AssetType = "RequestAccess"

		// convert to bytes
		assetAsBytes, errMarshal := json.Marshal(obj)
		
		// show error if failed
		if errMarshal != nil {
			return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
		}

		// save data in blockchain
		errPut := stub.PutState(obj.RequestAccess_ID, assetAsBytes)

		if errPut != nil {
			return shim.Error(fmt.Sprintf("Failed to save file data : %s", obj.RequestAccess_ID))
		}

		fmt.Println("Success in saving file data %v", obj)


		return shim.Success(assetAsBytes)
	}
		
	// update a record for RequestAccess
	func (t *MedicalApp) updateRequestAccess(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start updateRequestAccess")

		// check if all the args are send
		if len(args) != 2 {
			return shim.Error("Incorrect number of arguments, Required 2 arguments")
		}

		// check if the args are empty
		for i := 0; i < len(args); i++ {
			if len(args[i]) <= 0 {
				return shim.Error("argument "+ string(i+1)  + " must be a non-empty string")
			}
		}

		getAssetAsBytes, errT := stub.GetState(args[0])

		if errT != nil {
			return shim.Error(fmt.Sprintf("Error : Cannot find RequestAccess %s" , errT))
		}

		if getAssetAsBytes == nil {
			return shim.Error(fmt.Sprintf("Cannot find asset RequestAccess with ID %s" , args[0]))
		}

		var obj = RequestAccess{}

		json.Unmarshal(getAssetAsBytes, &obj)
		obj.Status = args[1]
		comAssetAsBytes, errMarshal := json.Marshal(obj)

		if errMarshal != nil {
			return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
		}

		errPut := stub.PutState(obj.RequestAccess_ID, comAssetAsBytes)

		if errPut != nil {
			return shim.Error(fmt.Sprintf("Failed to update RequestAccess with ID %s", args[0]))
		}

		fmt.Println("RequestAccess asset with ID %s was updated \n %v", args[0], obj)

		return shim.Success(comAssetAsBytes)
	}
			
	// get all records for asset RequestAccess
	func (t *MedicalApp) getAllRequestAccess(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start getAllRequestAccess")

		// check if all the args are send
		if len(args) != 3 {
			return shim.Error("Incorrect number of arguments, Required 3 arguments")
		}

		// check if the args are empty
		for i := 0; i < len(args); i++ {
			if len(args[i]) <= 0 {
				return shim.Error("argument "+ string(i+1)  + " must be a non-empty string")
			}
		}

		var queryString string

		AssetType := "RequestAccess"	
		AccessReqFrom := args[0]	// fromPatient | fromMedical
		assetId := args[1]
		Status := args[2]

		if AccessReqFrom == "fromPatient" {
			// show in doctor app all requests from patient -> request comes from patient
			queryString = fmt.Sprintf("SELECT valueJson FROM <STATE> WHERE json_extract(valueJson, '$.AssetType', '$.Doctor_ID', '$.Status', '$.AccessReqFrom') = '[\"%s\",\"%s\",\"%s\",\"%s\"]'", AssetType, assetId, Status, AccessReqFrom)
		} else {
			// show in patient app all requests from doctor
			queryString = fmt.Sprintf("SELECT valueJson FROM <STATE> WHERE json_extract(valueJson, '$.AssetType', '$.Patient_ID', '$.Status', '$.AccessReqFrom') = '[\"%s\",\"%s\",\"%s\",\"%s\"]'", AssetType, assetId, Status, AccessReqFrom)
		}

		queryResults, err := getQueryResultForQueryString(stub, queryString)
		
		if err != nil {
			return shim.Error(err.Error())
		}
		
		return shim.Success(queryResults)
	}
		
	// create a record for AccessControl
	func (t *MedicalApp) addAccessControl(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start addAccessControl")

		// check if all the args are send
		if len(args) != 8 {
			fmt.Println("Incorrect number of arguments, Required 8 arguments")
			return shim.Error("Incorrect number of arguments, Required 8 arguments")
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
		var obj = AccessControl{}

		obj.AccessControl_ID = ComputeHashKey(args[0]+(txTimeAsPtr.String()))
		obj.Doctor_ID = args[0]
		obj.Patient_ID = args[1]
		obj.IsValid = args[2]
		obj.ValidTill = args[3]
		obj.GrandedAccessOn = args[4]
		obj.Patient_Name = args[5]
		obj.Doctor_Name = args[6]
		obj.ClinicName = args[7]
		obj.AssetType = "AccessControl"

		// convert to bytes
		assetAsBytes, errMarshal := json.Marshal(obj)
		
		// show error if failed
		if errMarshal != nil {
			return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
		}

		// save data in blockchain
		errPut := stub.PutState(obj.AccessControl_ID, assetAsBytes)

		if errPut != nil {
			return shim.Error(fmt.Sprintf("Failed to save file data : %s", obj.AccessControl_ID))
		}

		fmt.Println("Success in saving file data %v", obj)


		return shim.Success(assetAsBytes)
	}
		
	// update a record for AccessControl
	func (t *MedicalApp) updateAccessControl(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start updateAccessControl")

		// check if all the args are send
		if len(args) != 2 {
			return shim.Error("Incorrect number of arguments, Required 2 arguments")
		}

		// check if the args are empty
		for i := 0; i < len(args); i++ {
			if len(args[i]) <= 0 {
				return shim.Error("argument "+ string(i+1)  + " must be a non-empty string")
			}
		}

		getAssetAsBytes, errT := stub.GetState(args[0])

		if errT != nil {
			return shim.Error(fmt.Sprintf("Error : Cannot find AccessControl %s" , errT))
		}

		if getAssetAsBytes == nil {
			return shim.Error(fmt.Sprintf("Cannot find asset AccessControl with ID %s" , args[0]))
		}

		var obj = AccessControl{}

		json.Unmarshal(getAssetAsBytes, &obj)
		obj.IsValid = args[1]
		comAssetAsBytes, errMarshal := json.Marshal(obj)

		if errMarshal != nil {
			return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
		}

		errPut := stub.PutState(obj.AccessControl_ID, comAssetAsBytes)

		if errPut != nil {
			return shim.Error(fmt.Sprintf("Failed to update AccessControl with ID %s", args[0]))
		}

		fmt.Println("AccessControl asset with ID %s was updated \n %v", args[0], obj)

		return shim.Success(comAssetAsBytes)
	}
			
	// get all records for asset AccessControl
	func (t *MedicalApp) getAllAccessControl(stub shim.ChaincodeStubInterface, args []string) peer.Response {
		// ==== Input sanitation ====
		fmt.Println("- start getAllAccessControl")

		// check if all the args are send
		if len(args) != 2 {
			return shim.Error("Incorrect number of arguments, Required 2 arguments")
		}

		// check if the args are empty
		for i := 0; i < len(args); i++ {
			if len(args[i]) <= 0 {
				return shim.Error("argument "+ string(i+1)  + " must be a non-empty string")
			}
		}

		AssetType := "AccessControl"
		IsValid := "true"

		queryString := fmt.Sprintf("SELECT valueJson FROM <STATE> WHERE json_extract(valueJson, '$.AssetType', '$.IsValid', '$.Doctor_ID', '$.ValidTill') = '[\"%s\",\"%s\",\"%s\",\"%s\"]'", AssetType, IsValid, args[0], args[1])
		
		queryResults, err := getQueryResultForQueryString(stub, queryString)
		
		if err != nil {
			return shim.Error(err.Error())
		}
		
		return shim.Success(queryResults)
	}
		
	// get data using a record id 
	func (t *MedicalApp) queryAsset(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
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
	func (t *MedicalApp) getAll(stub shim.ChaincodeStubInterface, args []string) peer.Response {
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
	func (t *MedicalApp) getHistoryForRecord(stub shim.ChaincodeStubInterface, args []string) peer.Response {
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
	func (t *MedicalApp) getHistory(stub shim.ChaincodeStubInterface, recordKey string) []byte {

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
	func (t *MedicalApp) GetTxTimestampChannel(APIstub shim.ChaincodeStubInterface) (time.Time, error) {
		txTimeAsPtr, err := APIstub.GetTxTimestamp()
		if err != nil {
			fmt.Printf("Returning error in TimeStamp \n")
		}
		fmt.Printf("\t returned value from APIstub: %v\n", txTimeAsPtr)
		timeInt := time.Unix(txTimeAsPtr.Seconds, int64(txTimeAsPtr.Nanos))

		return timeInt, nil
	}
		