package main

import (
	"bytes"
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/rs/xid"
)

var logger = shim.NewLogger("dcot-chaincode-log")


type DcotWorkflowChaincode struct {
	testMode bool
}

func (t *DcotWorkflowChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	logger.Info("Initializing Chain of Custody")
	logger.SetLevel(shim.LogDebug)
	_, args := stub.GetFunctionAndParameters()

	if len(args) == 0 {
		//logger.Info("Args correctly!!!")
		return shim.Success(nil)
	}

	return shim.Success(nil)
}

func (t *DcotWorkflowChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	var creatorOrg, creatorCertIssuer string
	var err error
	var isEnabled bool
	var callerRole string

	logger.Debug("DcotWorkflow Invoke\n")

	if !t.testMode {
		creatorOrg, creatorCertIssuer, err = getTxCreatorInfo(stub)
		if err != nil {
			logger.Error("Error extracting creator identity info: \n", err.Error())
			return shim.Error(err.Error())
		}
		logger.Info("DcotWorkflow Invoke by '', ''\n", creatorOrg, creatorCertIssuer)
		callerRole, _, err = getTxCreatorInfo(stub)
		if err != nil {
			return shim.Error(err.Error())
		}

		isEnabled, _, err = isInvokerOperator(stub, callerRole)
		if err != nil {
			logger.Error("Error getting attribute info: \n", err.Error())
			return shim.Error(err.Error())
		}
	}

	function, args := stub.GetFunctionAndParameters()

	if function == "initNewChain" {
		return t.initNewChain(stub, isEnabled, args)
	} else if function == "startTransfer" {
		return t.startTransfer(stub, isEnabled, args)
	} else if function == "completeTrasfer" {
		return t.completeTrasfer(stub, isEnabled, args)
	} else if function == "commentChain" {
		return t.commentChain(stub, isEnabled, args)
	} else if function == "cancelTrasfer" {
		return t.cancelTrasfer(stub, isEnabled, args)
	} else if function == "terminateChain" {
		return t.terminateChain(stub, isEnabled, args)
	} else if function == "updateDocument" {
		return t.updateDocument(stub, isEnabled, args)
	} else if function == "getAssetDetails" {
		return t.getAssetDetails(stub, isEnabled, args)
	} else if function == "getChainOfEvents" {
		return t.getChainOfEvents(stub, isEnabled, args)
	}
	return shim.Error("Invalid invoke function name")
}

//INITNEWCHAIN: the input json must contain the DocumentID, 
//The caller must be a MEMBER!!!
//Custodian is the member UID!!!


func (t *DcotWorkflowChaincode) initNewChain(stub shim.ChaincodeStubInterface, isEnabled bool, args []string) pb.Response {

	logger.Debug("initNewChain()")
	var jsonResp string
	var chainOfCustody ChainOfCustody
	var err error
	var byteCOC []byte
	var COCKey string
	var callerRole, callerUID string
	var operation string
	var event Event

	guid := xid.New()
	COCKey, err = getCOCKey(stub, guid.String())
	if err != nil {
		return shim.Error(err.Error())
	}
	err = json.Unmarshal([]byte(args[0]), &chainOfCustody)
	if err != nil {
		return shim.Error(err.Error())
	}
	if chainOfCustody.DocumentId == "" || len(chainOfCustody.DocumentId) == 0 {
		logger.Error("initNewChain ERROR: Document ID must not be null or empty string!\n")
		return shim.Error("initNewChain ERROR: Document ID must not be null or empty string!!\n")
	}
	chainOfCustody.Id = guid.String()
	chainOfCustody.Status = IN_CUSTODY
	operation = "initNewChain"
	callerRole, callerUID, err = getTxCreatorInfo(stub)
	if err != nil {
		logger.Error("initNewChain ERROR: getTxCreatorInfo()...\n")
		return shim.Error(err.Error())
	}
	if callerRole != CALLER_ROLE_0 {
		logger.Error("initNewChain ERROR: the user's role must be a member\n")
		return shim.Error("initNewChain ERROR: the user's role must be a member\n")
	}
	if len(callerUID) == 0 {
		logger.Error("initNewChain ERROR: caller_UID is empty!!!\n")
		return shim.Error("initNewChain ERROR: caller_UID is empty!!!\n")
	}
	chainOfCustody.DeliveryMan = string(callerUID)
	event, err = createEvent(stub, callerUID, callerRole, operation)
	if err != nil {
		logger.Error("initNewChain ERROR: createEvent()\n")
		return shim.Error(err.Error())
	}
	chainOfCustody.Event = event
	byteCOC, err = json.Marshal(&chainOfCustody)
	if err != nil {
		logger.Error("initNewChain ERROR: json.Marshal()\n")
		return shim.Error(err.Error())
	}
	err = stub.PutState(COCKey, byteCOC)
	if err != nil {
		logger.Error("initNewChain ERROR: PutState()\n")
		return shim.Error(err.Error())
	}
	jsonResp = string(byteCOC)
	logger.Info("Query Response:\n", jsonResp)
	err = stub.SetEvent("initNewChain EVENT: ", byteCOC)
	if err != nil {
		logger.Error("initNewChain ERROR: SetEvent()\n")
		return shim.Error(err.Error())
	}
	logger.Info("initNewChain EVENT: ", string(byteCOC))
	return shim.Success([]byte(jsonResp))
}





//STARTTRASFER: ChainOfCustody must exist and have 'IN_CUSTODY' status, 
//the caller must be the current custodian(Delivery_Man)

func (t *DcotWorkflowChaincode) startTransfer(stub shim.ChaincodeStubInterface, isEnabled bool, args []string) pb.Response {

	logger.Debug("startTransfer()")

	var COCKey string
	var err error
	var chainOfCustody ChainOfCustody
	var chainOfCustodyBytes []byte
	var byteCOC []byte
	var callerRole, callerUID string
	var operation string
	var event Event

	if len(args) != 2 {
		logger.Error("startTransfer ERROR: this method must want exactly two arguments!!\n")
		return shim.Error("startTransfer ERROR: this method must want exactly two arguments!!")
	}
	callerRole, callerUID, err = getTxCreatorInfo(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if callerRole == CALLER_ROLE_1 {
		logger.Error("startTransfer ERROR: Access denied for a Admin!!\n")
		return shim.Error("startTransfer ERROR: Access denied for a Admin!!!")
	}
	
	COCKey, err = getCOCKey(stub, args[0])
	if err != nil {
		logger.Error("startTransfer ERROR: getCOCKey()\n")
		return shim.Error(err.Error())
	}
	chainOfCustodyBytes, err = stub.GetState(COCKey)
	if err != nil {
		logger.Error("startTransfer ERROR: GetState()")
		return shim.Error(err.Error())
	}
	err = json.Unmarshal([]byte(chainOfCustodyBytes), &chainOfCustody)
	if err != nil {
		logger.Error("startTransfer ERROR: json.Unmarshal()\n")
		return shim.Error(err.Error())
	}
	logger.Debug(string(chainOfCustodyBytes))
	if callerUID != chainOfCustody.DeliveryMan {
		logger.Error("startTransfer ERROR : The caller must be the current custodian!!\n")
		return shim.Error("startTransfer ERROR : The caller must be the current custodian!!")
	}
	if chainOfCustody.Status != IN_CUSTODY {
		logger.Error("startTransfer ERROR: Asset have not status IN_CUSTODY!!!\n")
		return shim.Error("startTransfer ERROR : Asset have not status IN_CUSTODY!!\n")
	}
	operation = "startTransfer"
	chainOfCustody.Status = TRANSFER_PENDING
	chainOfCustody.DeliveryMan = args[1]
	event, err = createEvent(stub, callerUID, callerRole, operation)
	if err != nil {
		return shim.Error(err.Error())
	}


	chainOfCustody.Event = event
	logger.Info("startTransferAsset: New DeliveryMan: \n", chainOfCustody.DeliveryMan)
	byteCOC, err = json.Marshal(&chainOfCustody)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(COCKey, byteCOC)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.SetEvent("startTransfer EVENT: ", byteCOC)
	if err != nil {
		return shim.Error(err.Error())
	}
	logger.Info("startTransfer EVENT: ", string(byteCOC))
	return shim.Success(nil)
}


//COMPLETETRASFER: ChainOfCustody must exist and have 'TRANSFER_PENDING' status,
//The caller must be a new designed receiver(Delivery_Man) 


func (t *DcotWorkflowChaincode) completeTrasfer(stub shim.ChaincodeStubInterface, isEnabled bool, args []string) pb.Response {

	logger.Debug("completeTrasfer()")

	var COCKey string
	var err error
	var chainOfCustody *ChainOfCustody
	var chainOfCustodyBytes []byte
	var byteCOC []byte
	var callerRole, callerUID string
	var operation string
	var event Event


	if len(args) != 1 {
		return shim.Error("completeTrasfer ERROR: this method must want exactly one argument!!")
	}
	callerRole, callerUID, err = getTxCreatorInfo(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if callerRole == CALLER_ROLE_0 || callerRole == CALLER_ROLE_1{
		logger.Error("completeTrasfer ERROR: Access denied for a member or an admin!!")
		shim.Error("")
	}
	

	COCKey, err = getCOCKey(stub, args[0])
	if err != nil {
		logger.Error("completeTrasfer ERROR: getCOCKey()\n")
		return shim.Error(err.Error())
	}

	chainOfCustodyBytes, err = stub.GetState(COCKey)
	if err != nil {
		logger.Error("completeTrasfer ERROR: GetState()\n")
		return shim.Error(err.Error())
	}

	err = json.Unmarshal([]byte(chainOfCustodyBytes), &chainOfCustody)
	if err != nil {
		logger.Error("completeTrasfer ERROR: json.Unmarshal()\n")		
		return shim.Error(err.Error())
	}
	if callerUID != chainOfCustody.DeliveryMan {
		logger.Error("completeTrasfer ERROR: : The caller must be the current custodian!!\n")
		return shim.Error("completeTrasfer ERROR : The caller must be the current custodian!!")
	}

	logger.Info("completeTrasfer: Ok! Caller confirmed!!\n")
	if chainOfCustody.Status != TRANSFER_PENDING {
		logger.Error("completeTrasfer ERROR : Asset have not status TRANSFER_PENDING!!\n")
		return shim.Error("completeTrasfer ERROR : Asset have not status TRANSFER_PENDING!!")
	}
	operation = "completeTrasfer"
	chainOfCustody.Status = IN_CUSTODY
	event, err = createEvent(stub, callerUID, callerRole, operation)
	if err != nil {
		logger.Error("completeTrasfer ERROR :  createEvent()\n")
		return shim.Error(err.Error())
	}
	chainOfCustody.Event = event
	byteCOC, err = json.Marshal(&chainOfCustody)
	if err != nil {
		logger.Error("completeTrasfer ERROR :  json.Marshal()\n")
		return shim.Error(err.Error())
	}

	err = stub.PutState(COCKey, byteCOC)
	if err != nil {
		logger.Error("completeTrasfer ERROR : PutState()\n")
		return shim.Error(err.Error())
	}

	err = stub.SetEvent("completeTrasfer EVENT: ", byteCOC)
	if err != nil {
		logger.Error("completeTrasfer ERROR :  SetEvent()\n")
		return shim.Error(err.Error())
	}
	logger.Info("completeTrasfer EVENT: ", string(byteCOC))

	return shim.Success(nil)
}



//COMMENTCHAIN
//The call must be a OPERATOR or DELIVERY_OPERATOR or ADMIN


func (t *DcotWorkflowChaincode) commentChain(stub shim.ChaincodeStubInterface, isEnabled bool, args []string) pb.Response {

	logger.Debug("commentChain()")

	var COCKey string
	var err error
	var chainOfCustody *ChainOfCustody
	var chainOfCustodyBytes []byte
	var byteCOC []byte
	var callerUID string
	var callerRole string
	var operation string
	var event Event

	if len(args) != 2 {
		return shim.Error("commentChain ERROR: this method must want exactly two argument!!")
	}

	COCKey, err = getCOCKey(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	chainOfCustodyBytes, err = stub.GetState(COCKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = json.Unmarshal([]byte(chainOfCustodyBytes), &chainOfCustody)
	if err != nil {
		return shim.Error(err.Error())
	}
	callerRole, callerUID, err = getTxCreatorInfo(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if callerRole == CALLER_ROLE_0{
		logger.Error("commentChain ERROR: Access denied for a member!!\n")
		shim.Error("commentChain ERROR: Access denied for a member!!")
	}

	if callerRole == CALLER_ROLE_1 || callerUID == chainOfCustody.DeliveryMan {

		logger.Info("commentChain: Ok! Caller confirmed!!\n")
		operation = "commentChain"
		chainOfCustody.Text = args[1]
		event, err = createEvent(stub, callerUID, callerRole, operation)
		if err != nil {
			logger.Error("commentChain ERROR: createEvent!!\n")
			return shim.Error(err.Error())
		}
		chainOfCustody.Event = event
		byteCOC, err = json.Marshal(&chainOfCustody)
		if err != nil {
			logger.Error("commentChain ERROR: json.Mashal()!!\n")
			return shim.Error(err.Error())
		}
		err = stub.PutState(COCKey, byteCOC)
		if err != nil {
			logger.Error("commentChain ERROR: PutState()!!\n")
			return shim.Error(err.Error())
		}
		err = stub.SetEvent("commentChain EVENT: ", byteCOC)
		if err != nil {
			logger.Error("commentChain ERROR: SetEvent()!!\n")
			return shim.Error(err.Error())
		}
		logger.Info("commentChain EVENT: ", string(byteCOC))
		return shim.Success(nil)
	}
	logger.Error("completeTrasfer ERROR : the user's role is not compatible with this operation!")
	return shim.Error("completeTrasfer ERROR : the user's role is not compatible with this operation!\n")

}




//CANCELTRASFER



func (t *DcotWorkflowChaincode) cancelTrasfer(stub shim.ChaincodeStubInterface, isEnabled bool, args []string) pb.Response {

	logger.Debug("cancelTrasfer()")

	var COCKey string
	var err error
	var chainOfCustody *ChainOfCustody
	var chainOfCustodyBytes []byte
	var byteCOC []byte
	var callerUID, callerRole string
	var operation string
	var event Event

	if len(args) != 1 {
		return shim.Error("cancelTrasfer ERROR: this method must want exactly one argument!!")
	}
	COCKey, err = getCOCKey(stub, args[0])
	if err != nil {
		logger.Error("cancelTrasfer ERROR: getCOCKey()\n")
		return shim.Error(err.Error())
	}
	chainOfCustodyBytes, err = stub.GetState(COCKey)
	if err != nil {
		logger.Error("cancelTrasfer ERROR: getCOCKey()\n")
		return shim.Error(err.Error())
	}
	err = json.Unmarshal([]byte(chainOfCustodyBytes), &chainOfCustody)
	if err != nil {
		logger.Error("cancelTrasfer ERROR: json.Unmarshal()\n")
		return shim.Error(err.Error())
	}
	if chainOfCustody.Status != TRANSFER_PENDING {
		logger.Error("cancelTrasfer ERROR:  Asset have not status TRANSFER_PENDING!!\n")
		return shim.Error("cancelTrasfer ERROR : Asset have not status TRANSFER_PENDING!!")
	}
	
	callerRole, callerUID, err = getTxCreatorInfo(stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if callerUID == chainOfCustody.DeliveryMan || callerRole == CALLER_ROLE_1 {
		logger.Info("cancelTrasfer: Ok! Caller confirmed!!\n")
		operation = "cancelTrasfer"
		chainOfCustody.Status = IN_CUSTODY
		event, err = createEvent(stub, callerUID, callerRole, operation)
		if err != nil {
			logger.Error("cancelTrasfer ERROR: createEvent()\n")
			return shim.Error(err.Error())
		}
		chainOfCustody.Event = event
		byteCOC, err = json.Marshal(&chainOfCustody)
		if err != nil {
			logger.Error("cancelTrasfer ERROR: json.Marshal()\n")
			return shim.Error(err.Error())
		}
		err = stub.PutState(COCKey, byteCOC)
		if err != nil {
			logger.Error("cancelTrasfer ERROR: PutState()\n")
			return shim.Error(err.Error())
		}
		err = stub.SetEvent("cancelTrasfer EVENT: ", byteCOC)
		if err != nil {
			logger.Error("cancelTrasfer ERROR: SetEvent()\n")
			return shim.Error(err.Error())
		}
		logger.Info("cancelTrasfer EVENT: ", string(byteCOC))
		return shim.Success(nil)
	}
	logger.Error("cancelTrasfer ERROR : The caller must be the current custodian or have administrator role!!\n")
	return shim.Error("cancelTrasfer ERROR : The caller must be the current custodian or have administrator role!!")
}


//TERMINATECHAIN
//The calle must be a Admin or the current custodian!!! 

func (t *DcotWorkflowChaincode) terminateChain(stub shim.ChaincodeStubInterface, isEnabled bool, args []string) pb.Response {

	logger.Debug("terminateChain()")

	var COCKey string
	var err error
	var chainOfCustody *ChainOfCustody
	var chainOfCustodyBytes []byte
	var byteCOC []byte
	var callerUID, callerRole string
	var operation string
	var event Event

	if len(args) != 1 {
		return shim.Error("terminateChain ERROR: this method must want exactly one argument!!")
	}
	COCKey, err = getCOCKey(stub, args[0])
	if err != nil {
		logger.Error("terminateChain ERROR: getCOCKey()\n")
		return shim.Error(err.Error())
	}
	chainOfCustodyBytes, err = stub.GetState(COCKey)
	if err != nil {
		logger.Error("terminateChain ERROR: GetState()\n")
		return shim.Error(err.Error())
	}
	err = json.Unmarshal([]byte(chainOfCustodyBytes), &chainOfCustody)
	if err != nil {
		logger.Error("terminateChain ERROR: json.Unmarshal()\n")
		return shim.Error(err.Error())
	}
	if chainOfCustody.Status != IN_CUSTODY {
		logger.Error("terminateChain ERROR:  Asset have not status IN_CUSTODY!!\n")
		return shim.Error("terminateChain ERROR : Asset have not status IN_CUSTODY!!")
	}
	operation = "terminateChain"
	callerRole, callerUID, err = getTxCreatorInfo(stub)
	if err != nil {
		logger.Error("terminateChain ERROR: getTxCreatorInfo\n")
		return shim.Error(err.Error())
	}
	if callerRole == CALLER_ROLE_0 || callerRole == CALLER_ROLE_2{
		logger.Error("terminateChain ERROR : Access denied for member or operator!!\n")
		return shim.Error("terminateChain ERROR : Access denied for member or operator!!")
	}

	if callerRole == CALLER_ROLE_1 || callerUID == chainOfCustody.DeliveryMan {
		
		logger.Info("terminateChain: Ok! Caller confirmed!!\n")

		chainOfCustody.Status = RELEASED
		event, err = createEvent(stub, callerUID, callerRole, operation)
		if err != nil {
			logger.Error("terminateChain ERROR: createEvent()\n")
			return shim.Error(err.Error())
		}
		chainOfCustody.Event = event
		byteCOC, err = json.Marshal(&chainOfCustody)
		if err != nil {
			logger.Error("terminateChain ERROR: json.Marshal()\n")
			return shim.Error(err.Error())
		}
		err = stub.PutState(COCKey, byteCOC)
		if err != nil {
			logger.Error("terminateChain ERROR: createEPutStatevent()\n")
			return shim.Error(err.Error())
		}
		err = stub.SetEvent("terminateChain EVENT: ", byteCOC)
		if err != nil {
			logger.Error("terminateChain ERROR: SetEvent()\n")
			return shim.Error(err.Error())
		}
		logger.Info("terminateChain EVENT: ", string(byteCOC))

		return shim.Success(nil)
	}
	logger.Error("terminateChain ERROR : The caller must be the current custodian ora have a administrator role!!\n")
	return shim.Error("terminateChain ERROR : The caller must be the current custodian ora have a administrator role!!")
}



//UPDATEDOCUMENT


func (t *DcotWorkflowChaincode) updateDocument(stub shim.ChaincodeStubInterface, isEnabled bool, args []string) pb.Response {

	logger.Debug("updateDocument()")

	var COCKey string
	var err error
	var chainOfCustody *ChainOfCustody
	var chainOfCustodyBytes []byte
	var byteCOC []byte
	var jsonResp string
	var callerUID, callerRole string
	var operation string
	var event Event

	if len(args) != 2 {
		return shim.Error("updateDocument ERROR: this method must want exactly two argument!!")
	}

	COCKey, err = getCOCKey(stub, args[0])
	if err != nil {
		logger.Info("updateDocument ERROR: getCOCKey() \n")
		return shim.Error(err.Error())
	}
	chainOfCustodyBytes, err = stub.GetState(COCKey)
	if err != nil {
		logger.Info("updateDocument ERROR: GetState() \n")
		return shim.Error(err.Error())
	}
	err = json.Unmarshal([]byte(chainOfCustodyBytes), &chainOfCustody)
	if err != nil {
		logger.Info("updateDocument ERROR: json.Unmarshal()\n")
		return shim.Error(err.Error())
	}
	callerRole, callerUID, err = getTxCreatorInfo(stub)
	if err != nil {
		logger.Info("updateDocument ERROR: getTxCreatorInfo()\n")
		return shim.Error(err.Error())
	}
	if callerRole == CALLER_ROLE_1{
		logger.Info("updateDocument: Ok! Caller confirmed!!\n")

		if chainOfCustody.Status != IN_CUSTODY {
			logger.Info("updateDocument ERROR: Asset's status is not IN_CUSTODY!!!\n")
			return shim.Error("updateDocument ERROR: Asset's status is not IN_CUSTODY!!!")
		}
		operation = "updateDocument"
		chainOfCustody.DocumentId = args[1]
		event, err = createEvent(stub, callerUID, callerRole, operation)
		if err != nil {
			logger.Info("updateDocument ERROR: createEvent()\n")
			return shim.Error(err.Error())
		}
		chainOfCustody.Event = event
		byteCOC, err = json.Marshal(&chainOfCustody)
		if err != nil {
			logger.Info("updateDocument ERROR: json.Marshal()\n")
			return shim.Error(err.Error())
		}
		err = stub.PutState(COCKey, byteCOC)
		if err != nil {
			logger.Info("updateDocument ERROR: PutState()\n")
			return shim.Error(err.Error())
		}
		err = stub.SetEvent("updateDocument EVENT:", byteCOC)
		if err != nil {
			logger.Info("updateDocument ERROR: SetEvent()\n")
			return shim.Error(err.Error())
		}
		logger.Info("updateDocument EVENT: ", string(byteCOC))
		jsonResp = string(byteCOC)
		logger.Info("Query Response:\n", jsonResp)
		return shim.Success([]byte(jsonResp))
	}
	logger.Error("cancelTrasfer ERROR : the user's role is not compatible with this operation!!\n")
	return shim.Error("cancelTrasfer ERROR : the user's role is not compatible with this operation!!\n")
}






//GETASSETDETAILS 
//The calle must be a Delivery_operator or Operator or Admin!!


func (t *DcotWorkflowChaincode) getAssetDetails(stub shim.ChaincodeStubInterface, isEnabled bool, args []string) pb.Response {

	logger.Debug("getAssetDetails()")

	var COCKey string
	var err error
	var chainOfCustody *ChainOfCustody
	var chainOfCustodyBytes []byte
	var byteCOC []byte
	var jsonResp string
	var callerRole string

	if len(args) != 1 {
		return shim.Error("getAssetDetails ERROR: this method must want exactly one argument!!")
	}
	COCKey, err = getCOCKey(stub, args[0])
	if err != nil {
		logger.Error("getAssetDetails ERROR : getCOCKey()\n")
		return shim.Error(err.Error())
	}
	chainOfCustodyBytes, err = stub.GetState(COCKey)
	if err != nil {
		logger.Error("getAssetDetails ERROR : GetState())\n")
		return shim.Error(err.Error())
	}
	err = json.Unmarshal([]byte(chainOfCustodyBytes), &chainOfCustody)
	if err != nil {
		logger.Error("getAssetDetails ERROR : json.Unmarshal()\n")
		return shim.Error(err.Error())
	}
	callerRole, _, err = getTxCreatorInfo(stub)
	if err != nil {
		logger.Error("getAssetDetails ERROR : getTxCreatorInfo\n")
		return shim.Error(err.Error())
	}
	if callerRole == CALLER_ROLE_1 || callerRole == CALLER_ROLE_2 || callerRole == CALLER_ROLE_3{
		logger.Info("getAssetDetails: Ok! Caller confirmed!!\n")
		byteCOC, err = json.Marshal(&chainOfCustody)
		if err != nil {
			logger.Error("getAssetDetails ERROR : json.Marshal()\n")
			return shim.Error(err.Error())
		}
		jsonResp = string(byteCOC)
		logger.Info("Query Response:\n", jsonResp)
		return shim.Success([]byte(jsonResp))
	}
	logger.Error("getAssetDetails ERROR : the user's role is not compatible with this operation!")
	return shim.Error("getAssetDetails ERROR : the user's role is not compatible with this operation!\n")
}




//GETCHAINOFEVENTS
//The caller must be a Admin!!!



func (t *DcotWorkflowChaincode) getChainOfEvents(stub shim.ChaincodeStubInterface, isEnabled bool, args []string) pb.Response {

	logger.Debug("getChainOfEvents() ")

	var COCKey string
	var err2 error
	var chainOfCustody *ChainOfCustody
	var byteCOC []byte
	var jsonResp, jsonResponse string
	var callerRole string
	var err error



	if len(args) != 1 {
		return shim.Error("getChainOfEvents ERROR: this method must want exactly one argument!!")
	}
	callerRole, _, err = getTxCreatorInfo(stub)
	if err != nil {
		logger.Error("getChainOfEvents ERROR: getTxCreatorInfo()\n ")
		return shim.Error(err.Error())
	}
	logger.Info("caller_ROLE :" + string(callerRole) + " . \n")
	if callerRole == CALLER_ROLE_1{
		logger.Info("getChainOfEvents: Ok! Caller confirmed!!\n")
		COCKey, err2 = getCOCKey(stub, args[0])
		if err2 != nil {
			logger.Error("getChainOfEvents ERROR: getCOCKey()\n ")
			return shim.Error(err2.Error())
		}
		historyResponse, err3 := stub.GetHistoryForKey(COCKey)
		if err3 != nil {
			logger.Error("getChainOfEvents ERROR: GetHistoryForKey()\n ")
			return shim.Error(err3.Error())
		}
		var buffer bytes.Buffer
		buffer.WriteString("[")
		for historyResponse.HasNext() {
			COCarray, err1 := historyResponse.Next()
			if err1 != nil {
				logger.Error("getChainOfEvents ERROR: historyResponse.Next()\n ")
				return shim.Error(err1.Error())
			}
			err = json.Unmarshal([]byte(COCarray.Value), &chainOfCustody)
			if err != nil {
				logger.Error("getChainOfEvents ERROR: json.Unmarshal()\n ")
				return shim.Error(err.Error())
			}
			byteCOC, err2 = json.Marshal(&chainOfCustody)
			if err2 != nil {
				logger.Error("getChainOfEvents ERROR: json.Marshal()\n ")
				return shim.Error(err2.Error())
			}
			logger.Debug("byteCOC :", string(byteCOC))
			buffer.WriteString(string(byteCOC))
			buffer.WriteString(",")
		}
		jsonResp = buffer.String()
		subString := jsonResp[0 : len(jsonResp)-1]
		jsonResponse = subString + "]"
		logger.Debug("Query Response:\n" + jsonResponse)
		return shim.Success([]byte(jsonResponse))
	}
	logger.Error("getChainOfEvents ERROR : the user's role is not compatible with this operation!\n")
	return shim.Error("getChainOfEvents ERROR : the user's role is not compatible with this operation!\n")

}










func main() {
	twc := new(DcotWorkflowChaincode)
	twc.testMode = true
	err := shim.Start(twc)
	if err != nil {
		logger.Error("Error starting Chain of Custody chaincode: ", err)
	}
}
