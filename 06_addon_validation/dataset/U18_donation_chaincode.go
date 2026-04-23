package main

import (
    "encoding/json"
    "fmt"
    "time"

    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing donations on the ledger.
type SmartContract struct {
    contractapi.Contract
}

// Donation defines the structure stored in the ledger for each donation.
type Donation struct {
    DonationID string  `json:"DonationID"`
    DonorID    string  `json:"DonorID"`
    Amount     float64 `json:"Amount"`
    Timestamp  string  `json:"Timestamp"`
    CampaignID string  `json:"CampaignID"`
    ReceiverID string  `json:"ReceiverID"`
}

// InitLedger can be used to seed the ledger initially (no-op here).
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
    return nil
}

// CreateDonation records a new donation in the world state.
func (s *SmartContract) CreateDonation(ctx contractapi.TransactionContextInterface,
    donationID, donorID string,
    amount float64,
    campaignID, receiverID string) error {

    existing, err := ctx.GetStub().GetState(donationID)
    if err != nil {
        return fmt.Errorf("failed to read from world state: %v", err)
    }
    if existing != nil {
        return fmt.Errorf("the donation %s already exists", donationID)
    }

    donation := Donation{
        DonationID: donationID,
        DonorID:    donorID,
        Amount:     amount,
        Timestamp:  time.Now().UTC().Format(time.RFC3339),
        CampaignID: campaignID,
        ReceiverID: receiverID,
    }

    donationJSON, err := json.Marshal(donation)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(donationID, donationJSON)
}

// GetDonation retrieves a donation by its ID from the world state.
func (s *SmartContract) GetDonation(ctx contractapi.TransactionContextInterface,
    donationID string) (*Donation, error) {

    donationJSON, err := ctx.GetStub().GetState(donationID)
    if err != nil {
        return nil, fmt.Errorf("failed to read from world state: %v", err)
    }
    if donationJSON == nil {
        return nil, fmt.Errorf("donation %s does not exist", donationID)
    }

    var donation Donation
    err = json.Unmarshal(donationJSON, &donation)
    if err != nil {
        return nil, err
    }

    return &donation, nil
}

// QueryAllDonations returns all donations found in world state.
func (s *SmartContract) QueryAllDonations(ctx contractapi.TransactionContextInterface) ([]*Donation, error) {
    resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var donations []*Donation
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var donation Donation
        err = json.Unmarshal(queryResponse.Value, &donation)
        if err != nil {
            return nil, err
        }

        donations = append(donations, &donation)
    }

    return donations, nil
}

func main() {
    chaincode, err := contractapi.NewChaincode(&SmartContract{})
    if err != nil {
        panic(fmt.Sprintf("Error creating donation chaincode: %v", err))
    }

    if err := chaincode.Start(); err != nil {
        fmt.Printf("Error starting Donation chaincode: %v", err)
    }
}
