/*  completed till  coinsurer agree to the lead insurer quote,
next step is to provide client to set the capacity of different
 coinsurer and then policy and all  and one more thing ONLY BROKER HAS THE FUCNTIONALITY TO PROVIDE RFQ    */

package main

import (
	"crypto/sha256"
	//"reflect"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	mspprotos "github.com/hyperledger/fabric/protos/msp"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const RFQ_INITIALIZED string = "Pending"
const LEAD_ASSIGNED string = "Lead Assigned"
const RFQ_QUOTES_FINALIZED string = "RFQ Quotes finalized"
const RFQ_PROPOSAL_FINALIZED string = "RFQ Proposal Finalized"
const RFQ_COMPLETED string = "RFQ Completed"

const QUOTE_INITIALIZED string = "Quote Initialized"
const QUOTE_ACCEPTED string = "Lead Quote Accepted"
const QUOTE_REJECTED string = "Quote Rejected"
const QUOTES_FINALIZED string = "Quotes Finalized"

const PROPOSAL_INITIALIZED string = "Proposal Initialized"
const PROPOSAL_PAYMENT_MARKED string = "Proposal Payment Marked"

const POLICY_INITIALIZED string = "Policy Initialized"

const INTERMEDIARY_CLIENT string = "Intermediary Client"
const INTERMEDIARY_BROKER string = "Intermediary Broker"

const INSURERS_LIST string = "Insurers List"
const SURVEYORS_LIST string = "Surveyors List"
const CLAIM_INITIALIZED string = "Claim Initialized"
const CLAIM_SURVEYOR_ASSIGNED string = "Claim Surveyor Assigned"
const CLAIM_INSPECTION_COMPLETED string = "Claim Inspection Completed"
const CLAIM_COMPLETED string = "Claim Completed"
const READ_CLAIM_PENDING string = "Read Claim Pending"
const READ_CLAIM_COMPLETED string = "Read Claim Completed"

type InsuranceManagement struct {
}

//=============================main=====================================================

func main() {
	err := shim.Start(new(InsuranceManagement))
	if err != nil {
		fmt.Println("error starting chaincode :%s", err)
	}
}

//======================================init========================================

func (t *InsuranceManagement) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Init called")
	_, args := stub.GetFunctionAndParameters()
	if len(args) != 0 {
		return shim.Error(fmt.Sprintf("chaincode:Init::Wrong number of arguments"))
	}
	var arr []string
	InsurerListasBytes, err := json.Marshal(arr)
	if err != nil {
		return shim.Error(fmt.Sprintf("chaincode:Init::empty insurer list not marshalled"))
	}
	err = stub.PutState(INSURERS_LIST, InsurerListasBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("chaincode:Init::couldnt put state insurer list"))
	}
	var surveyorArr []string 
	surveyorListAsBytes, err := json.Marshal(surveyorArr)
	if err != nil {
		return shim.Error(fmt.Sprintf("chaincode:Init::empty surveyor list not marshalled"))
	}
	err = stub.PutState(SURVEYORS_LIST, surveyorListAsBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("chaincode:Init::couldnt put state surveyor list"))
	}

	return shim.Success(nil)
}

//==============================invoke====================================================

func (t *InsuranceManagement) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "initClient" {
		return t.InitClient(stub, args) //done
	} else if function == "initInsurer" {
		return t.InitInsurer(stub, args) //done
	} else if function == "generateRFQByClient" {
		return t.GenerateRFQByClient(stub, args)
	} else if function == "provideQuote" {
		return t.ProvideQuote(stub, args)
	} else if function == "initBroker" {
		return t.InitBroker(stub, args) //done
	} else if function == "generateRFQByBroker" {
		return t.GenerateRFQByBroker(stub, args)
	} else if function == "initClientByBroker" {
		return t.InitClientByBroker(stub, args) //done
	} else if function == "selectLeadInsurerByClient" {
		return t.SelectLeadInsurerByClient(stub, args)
	} else if function == "selectLeadInsurerByBroker" {
		return t.SelectLeadInsurerByBroker(stub, args)
	} else if function == "acceptLeadQuote" {
		return t.AcceptLeadQuote(stub, args)
	} else if function == "rejectLeadQuote" {
		return t.RejectLeadQuote(stub, args)
	} else if function == "finalizeQuotesByClient" {
		return t.FinalizeQuotesByClient(stub, args)
	} else if function == "finalizeQuotesByBroker" {
		return t.FinalizeQuotesByBroker(stub, args)
	} else if function == "uploadProposalFormByClient" {
		return t.UploadProposalFormByClient(stub, args)
	} else if function == "uploadProposalFormByBroker" {
		return t.UploadProposalFormByBroker(stub, args)
	} else if function == "allotProposalNumber" {
		return t.AllotProposalNumber(stub, args)
	} else if function == "markPaymentAndGeneratePolicy" {
		return t.MarkPaymentAndGeneratePolicy(stub, args)
	} else if function == "readAcc" {
		return t.ReadAcc(stub, args)
	} else if function == "readAllRFQ" {
		return t.ReadAllRFQ(stub, args)
	} else if function == "readSingleRFQ" {
		return t.ReadSingleRFQ(stub, args)
	} else if function == "readRFQByRange" {
		return t.ReadRFQByRange(stub, args)
	} else if function == "readClientOfBroker" {
		return t.ReadClientOfBroker(stub, args)
	} else if function == "readAllProposal" {
		return t.ReadAllProposal(stub, args)
	} else if function == "readSingleProposal" {
		return t.ReadSingleProposal(stub, args)
	} else if function == "readProposalByRange" {
		return t.ReadProposalByRange(stub, args)
	} else if function == "readAllInsurers" {
		return t.ReadAllInsurers(stub, args)
	}else if function == "readAllSurveyors" {
		return t.ReadAllSurveyors(stub, args)
	}else if function == "readAllQuote" {
		return t.ReadAllQuote(stub, args)
	} else if function == "readQuoteByRange" {
		return t.ReadQuoteByRange(stub, args)
	} else if function == "readSingleQuote" {
		return t.ReadSingleQuote(stub, args)
	} else if function == "readAllPolicy" {
		return t.ReadAllPolicy(stub, args)
	} else if function == "readSinglePolicy" {
		return t.ReadSinglePolicy(stub, args)
	} else if function == "readPolicyByRange" {
		return t.ReadPolicyByRange(stub, args)
	} else if function == "generateClaimByClient" {
		return t.GenerateClaimByClient(stub, args)
	} else if function == "assignSurveyor" {
		return t.AssignSurveyorToClaim(stub, args)
	} else if function == "initSurveyor" {
		return t.InitSurveyor(stub, args)
	} else if function == "uploadClaimReport" {
		return t.UploadClaimReport(stub, args)
	} else if function == "sendClaim" {
		return t.SendClaim(stub, args)
	} else if function == "readAllClaim" {
		return t.ReadAllClaim(stub, args)
	}else if function == "readMetaDataClient" {
		return t.ReadMetaDataClient(stub, args)
	}else if function == "readMetaDataInsurer" {
		return t.ReadMetaDataInsurer(stub, args)
	}else if function == "readMetaDataBroker" {
		return t.ReadMetaDataBroker(stub, args)
	}else if function == "readMetaDataSurveyor" {
		return t.ReadMetaDataSurveyor(stub, args)
	}else if function == "readSingleClaim" {
		return t.ReadSingleClaim(stub, args)
	}else if function == "generateClaimByBroker" {
		return t.GenerateClaimByBroker(stub, args)
	}else if function == "readAllClaimSurveyor" {
		return t.ReadAllClaimSurveyor(stub, args)
	}

	return shim.Error(fmt.Sprintf("chaincode:Invoke::NO such function exists"))

}

func (t *InsuranceManagement) ReadAcc(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 0 {
		return shim.Error(fmt.Sprintf("chaincode:readAcc:0 arguments expected"))
	}

	creator, err := stub.GetCreator() // it'll give the certificate of the invoker
	id := &mspprotos.SerializedIdentity{}
	err = proto.Unmarshal(creator, id)

	if err != nil {
		return shim.Error(fmt.Sprintf("chaincode:readAcc:couldnt unmarshal creator"))
	}
	block, _ := pem.Decode(id.GetIdBytes())
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("chaincode:readAcc:couldnt parse certificate"))
	}
	invokerhash := sha256.Sum256([]byte(cert.Subject.CommonName + cert.Issuer.CommonName))
	invokerAddress := hex.EncodeToString(invokerhash[:])

	invokerAsBytes, err := stub.GetState(invokerAddress)
	if err != nil || len(invokerAsBytes) == 0 {
		return shim.Error(fmt.Sprintf("chaincode:readAcc::account doesnt exists"))

	}
	return shim.Success(invokerAsBytes)

}

func (t *InsuranceManagement) ReadAllInsurers(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 0 {
		return shim.Error(fmt.Sprintf("chaincode:ReadAllInsurers::Expected no arguments"))
	}

	var insurerList []string
	insurerListAsbytes, err := stub.GetState(INSURERS_LIST)
	if err != nil {
		return shim.Error(fmt.Sprintf("chaincode:ReadAllInsurers::couldnt get state of insurerList"))
	}
	err = json.Unmarshal(insurerListAsbytes, &insurerList)
	if err != nil {
		return shim.Error(fmt.Sprintf("chaincode:ReadAllInsurers::Insurer list not unmarshalled"))
	}
	type insurerNameAddress struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	}

	var outputInsurers []insurerNameAddress

	for i := range insurerList {
		insurer := Insurer{}
		insurerasbytes, err := stub.GetState(insurerList[i])
		if err != nil {
			return shim.Error(fmt.Sprintf("chaincode:ReadAllInsurers::Something is wrong, insurer is missing"))
		}
		err = json.Unmarshal(insurerasbytes, &insurer)
		if err != nil {
			return shim.Error(fmt.Sprintf("chaincode:ReadAllInsurers::insurer couldnt get unmarshalled"))
		}
		outputInsurer := insurerNameAddress{}
		outputInsurer.Address = insurer.InsurerId
		outputInsurer.Name = insurer.InsurerName
		outputInsurers = append(outputInsurers, outputInsurer)
	}

	newOutputInsurersAsbytes, err := json.Marshal(outputInsurers)
	if err != nil {
		return shim.Error(fmt.Sprintf("chaincode:ReadAllInsurers::couldnt marshal output of insurers list"))
	}

	return shim.Success(newOutputInsurersAsbytes)

}

func (t *InsuranceManagement) ReadAllSurveyors(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 0 {
		return shim.Error(fmt.Sprintf("chaincode:ReadAllSurveyors::Expected no arguments"))
	}

	var surveyorList []string
	surveyorListAsbytes, err := stub.GetState(SURVEYORS_LIST)
	if err != nil {
		return shim.Error(fmt.Sprintf("chaincode:ReadAllSurveyors::couldnt get state of surveyorList"))
	}
	err = json.Unmarshal(surveyorListAsbytes, &surveyorList)
	if err != nil {
		return shim.Error(fmt.Sprintf("chaincode:ReadAllSurveyors::surveyor list not unmarshalled"))
	}
	type surveyorNameAddress struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	}

	var outputsurveyors []surveyorNameAddress

	for i := range surveyorList {
		surveyor := Surveyor{}
		surveyorasbytes, err := stub.GetState(surveyorList[i])
		if err != nil {
			return shim.Error(fmt.Sprintf("chaincode:ReadAllSurveyors::Something is wrong, surveyor is missing"))
		}
		err = json.Unmarshal(surveyorasbytes, &surveyor)
		if err != nil {
			return shim.Error(fmt.Sprintf("chaincode:ReadAllSurveyors::surveyor couldnt get unmarshalled"))
		}
		outputsurveyor := surveyorNameAddress{}
		outputsurveyor.Address = surveyor.SurveyorId
		outputsurveyor.Name = surveyor.SurveyorName
		outputsurveyors = append(outputsurveyors, outputsurveyor)
	}

	newOutputsurveyorsAsbytes, err := json.Marshal(outputsurveyors)
	if err != nil {
		return shim.Error(fmt.Sprintf("chaincode:ReadAllSurveyors::couldnt marshal output of surveyors list"))
	}

	return shim.Success(newOutputsurveyorsAsbytes)

}

///func (t *InsuranceManagement) ReadRFQListForinvokerByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {
