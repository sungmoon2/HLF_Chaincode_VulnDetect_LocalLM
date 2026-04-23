package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Donation structure
type Donation struct {
	DonationID string `json:"donationID"`
	DonorName  string `json:"donorName"`
	NGOName    string `json:"ngoName"`
	Amount     float64 `json:"amount"`
	Purpose    string `json:"purpose"`
	Status     string `json:"status"`
	Date       string `json:"date"`
}

// SmartContract defines contract structure
type SmartContract struct {
	contractapi.Contract
}

// CreateDonation – create a new donation entry
func (s *SmartContract) CreateDonation(ctx contractapi.TransactionContextInterface,
	id, donor, ngo, purpose, date string, amount float64) error {

	donation := Donation{
		DonationID: id, DonorName: donor, NGOName: ngo,
		Purpose: purpose, Date: date, Amount: amount, Status: "Created",
	}

	donationJSON, err := json.Marshal(donation)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, donationJSON)
}

// UpdateStatus – NGO updates donation status
func (s *SmartContract) UpdateStatus(ctx contractapi.TransactionContextInterface, id, status string) error {
	data, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to get donation: %v", err)
	}
	if data == nil {
		return fmt.Errorf("donation %s not found", id)
	}

	var donation Donation
	if err := json.Unmarshal(data, &donation); err != nil {
		return err
	}

	donation.Status = status
	updated, _ := json.Marshal(donation)
	return ctx.GetStub().PutState(id, updated)
}

// GetDonation – read donation by ID
func (s *SmartContract) GetDonation(ctx contractapi.TransactionContextInterface, id string) (*Donation, error) {
	data, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("donation %s not found", id)
	}
	var donation Donation
	_ = json.Unmarshal(data, &donation)
	return &donation, nil
}

// GetAllDonations – return all donations
func (s *SmartContract) GetAllDonations(ctx contractapi.TransactionContextInterface) ([]*Donation, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var donations []*Donation
	for resultsIterator.HasNext() {
		queryResponse, _ := resultsIterator.Next()
		var donation Donation
		_ = json.Unmarshal(queryResponse.Value, &donation)
		donations = append(donations, &donation)
	}
	return donations, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating chaincode: %s", err.Error())
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %s", err.Error())
	}
}
