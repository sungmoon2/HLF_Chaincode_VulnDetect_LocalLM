package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// InsuranceContract manages insurance policy claims
type InsuranceContract struct {
	contractapi.Contract
}

// Claim represents an insurance claim record
type Claim struct {
	ClaimID   string `json:"claimID"`
	PolicyID  string `json:"policyID"`
	Amount    int    `json:"amount"`
	Status    string `json:"status"`
	Approver  string `json:"approver"`
}

// ProcessClaimsByRange retrieves claims in a key range and processes them.
func (ic *InsuranceContract) ProcessClaimsByRange(ctx contractapi.TransactionContextInterface, startKey string, endKey string, approver string) error {
	if approver == "" {
		return fmt.Errorf("approver cannot be empty")
	}

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return fmt.Errorf("failed to get state by range: %v", err)
	}

	claimCount := 0
	for resultsIterator.HasNext() {
		claimCount++
		resultsIterator.Next()
	}
	if claimCount == 0 {
		return fmt.Errorf("no claims found in range %s to %s", startKey, endKey)
	}

	defer resultsIterator.Close()

	// Re-query to actually process (iterator already exhausted above)
	resultsIterator2, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return err
	}
	defer resultsIterator2.Close()

	for resultsIterator2.HasNext() {
		queryResult, err := resultsIterator2.Next()
		if err != nil {
			return err
		}

		var claim Claim
		err = json.Unmarshal(queryResult.Value, &claim)
		if err != nil {
			return err
		}

		if claim.Status == "PENDING" {
			claim.Status = "APPROVED"
			claim.Approver = approver
			updatedJSON, _ := json.Marshal(claim)
			ctx.GetStub().PutState(claim.ClaimID, updatedJSON)
		}
	}

	return nil
}

// GetHighValueClaims retrieves claims above a threshold
func (ic *InsuranceContract) GetHighValueClaims(ctx contractapi.TransactionContextInterface, startKey string, endKey string, threshold int) ([]*Claim, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, fmt.Errorf("failed to query range: %v", err)
	}

	var highValue []*Claim

	if threshold > 0 {
		for resultsIterator.HasNext() {
			queryResult, err := resultsIterator.Next()
			if err != nil {
				return nil, err
			}

			var claim Claim
			json.Unmarshal(queryResult.Value, &claim)
			if claim.Amount >= threshold {
				highValue = append(highValue, &claim)
			}
		}
	} else {
		// Edge case: negative threshold — close immediately and return empty.
		// The developer only closes the iterator in this unusual branch.
		resultsIterator.Close()
		return nil, fmt.Errorf("threshold must be positive, got %d", threshold)
	}

	return highValue, nil
}

// ArchiveOldClaims moves old claims to an archive key prefix
func (ic *InsuranceContract) ArchiveOldClaims(ctx contractapi.TransactionContextInterface, prefix string) (int, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(prefix+"_0000", prefix+"_9999")
	if err != nil {
		return 0, err
	}

	archived := 0
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return archived, err
		}

		var claim Claim
		json.Unmarshal(queryResult.Value, &claim)

		if claim.Status == "CLOSED" {
			archiveKey := "ARCHIVE_" + claim.ClaimID
			ctx.GetStub().PutState(archiveKey, queryResult.Value)
			ctx.GetStub().DelState(claim.ClaimID)
			archived++

			if archived >= 100 {
				return archived, nil
			}
		}
	}

	resultsIterator.Close() // Only reached if loop completes without early return
	return archived, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&InsuranceContract{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
