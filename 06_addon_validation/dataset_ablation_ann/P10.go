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
// [VULNERABILITY] The iterator is closed with defer, BUT a return statement
// appears BEFORE the defer statement on an error path. If GetStateByRange
// succeeds but the subsequent validation fails, the function returns early
// and the defer never executes — leaking the iterator's gRPC stream.
func (ic *InsuranceContract) ProcessClaimsByRange(ctx contractapi.TransactionContextInterface, startKey string, endKey string, approver string) error {
	if approver == "" {
		return fmt.Errorf("approver cannot be empty")
	}

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return fmt.Errorf("failed to get state by range: %v", err)
	}

	// [VULNERABILITY] Validation check AFTER iterator is opened but BEFORE defer.
	// If this condition is true, the function returns and the iterator is never closed.
	claimCount := 0
	for resultsIterator.HasNext() {
		claimCount++
		resultsIterator.Next()
	}
	if claimCount == 0 {
		// [VULNERABILITY] Early return without closing the iterator.
		// The defer below has not been registered yet.
		return fmt.Errorf("no claims found in range %s to %s", startKey, endKey)
	}

	// [VULNERABILITY] defer is placed AFTER the early return above.
	// In Go, defer statements are registered at the point of execution.
	// If the function returned at line above, this defer was never reached.
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
// [VULNERABILITY] The iterator Close() is inside a conditional block.
// If the threshold is negative (unexpected input), the function takes
// the else branch and closes the iterator — but on the normal path
// where threshold > 0, the function completes the loop and falls through
// WITHOUT closing, because Close() is only in the else branch.
func (ic *InsuranceContract) GetHighValueClaims(ctx contractapi.TransactionContextInterface, startKey string, endKey string, threshold int) ([]*Claim, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, fmt.Errorf("failed to query range: %v", err)
	}

	var highValue []*Claim

	if threshold > 0 {
		// [VULNERABILITY] Normal processing path — iterator is iterated but never closed.
		// The developer assumed the iterator auto-closes after exhaustion, but it does not.
		for resultsIterator.HasNext() {
			queryResult, err := resultsIterator.Next()
			if err != nil {
				return nil, err // [VULNERABILITY] leaked iterator on error
			}

			var claim Claim
			json.Unmarshal(queryResult.Value, &claim)
			if claim.Amount >= threshold {
				highValue = append(highValue, &claim)
			}
		}
		// [VULNERABILITY] No Close() on this path — falls through to return.
	} else {
		// Edge case: negative threshold — close immediately and return empty.
		// The developer only closes the iterator in this unusual branch.
		resultsIterator.Close()
		return nil, fmt.Errorf("threshold must be positive, got %d", threshold)
	}

	return highValue, nil // [VULNERABILITY] iterator not closed on the normal (threshold > 0) path
}

// ArchiveOldClaims moves old claims to an archive key prefix
// [VULNERABILITY] Close() is called, but only if the loop completes
// without error. Any early return inside the loop skips the Close().
func (ic *InsuranceContract) ArchiveOldClaims(ctx contractapi.TransactionContextInterface, prefix string) (int, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(prefix+"_0000", prefix+"_9999")
	if err != nil {
		return 0, err
	}
	// [VULNERABILITY] No defer — relying on manual Close() at end of function.

	archived := 0
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return archived, err // [VULNERABILITY] leaked — manual Close() at bottom never reached
		}

		var claim Claim
		json.Unmarshal(queryResult.Value, &claim)

		if claim.Status == "CLOSED" {
			archiveKey := "ARCHIVE_" + claim.ClaimID
			ctx.GetStub().PutState(archiveKey, queryResult.Value)
			ctx.GetStub().DelState(claim.ClaimID)
			archived++

			if archived >= 100 {
				// [VULNERABILITY] Batch limit reached — return early without Close()
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
