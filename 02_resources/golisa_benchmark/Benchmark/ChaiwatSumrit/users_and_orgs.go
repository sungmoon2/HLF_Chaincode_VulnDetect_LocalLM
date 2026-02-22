package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// Organization object, containing organization detials
type Organization struct {
	Asset
	OrgID   string `json:"orgId"`
	OrgName string `json:"orgName"`
}

// User is the system iser profile
type User struct {
	Participant
	UserId    string `json:"userId"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	OrgID     string `json:"orgId"`
}

// IT_RPT is a special type derived from User object
type IT_RPT struct {
	User
}

type UsersAndOrgs struct {
}

//Init initializes the Parcel model/smart contract
func (t *UsersAndOrgs) Init(stub shim.ChaincodeStubInterface) pb.Response {

	return shim.Success(nil)

}
