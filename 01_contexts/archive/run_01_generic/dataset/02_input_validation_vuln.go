package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// TokenChaincode implements a simple token system
type TokenChaincode struct {
	contractapi.Contract
}

// TokenAccount represents a user's token balance
type TokenAccount struct {
	AccountID string  `json:"accountId"`
	Owner     string  `json:"owner"`
	Balance   float64 `json:"balance"`
	Frozen    bool    `json:"frozen"`
}

// VULNERABILITY: No input validation on amount - accepts negative values
// This allows attackers to mint tokens by "transferring" negative amounts
func (t *TokenChaincode) Transfer(ctx contractapi.TransactionContextInterface, fromID string, toID string, amountStr string) error {
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	// VULNERABILITY: No check for negative amount
	// A negative transfer from A to B effectively moves tokens from B to A
	// No check for zero amount (wasteful transaction)

	fromJSON, err := ctx.GetStub().GetState(fromID)
	if err != nil {
		return fmt.Errorf("failed to read sender account: %v", err)
	}
	if fromJSON == nil {
		return fmt.Errorf("sender account %s does not exist", fromID)
	}

	toJSON, err := ctx.GetStub().GetState(toID)
	if err != nil {
		return fmt.Errorf("failed to read receiver account: %v", err)
	}
	if toJSON == nil {
		return fmt.Errorf("receiver account %s does not exist", toID)
	}

	var fromAccount, toAccount TokenAccount
	json.Unmarshal(fromJSON, &fromAccount)
	json.Unmarshal(toJSON, &toAccount)

	// VULNERABILITY: No overflow/underflow check
	fromAccount.Balance -= amount
	toAccount.Balance += amount

	fromUpdated, _ := json.Marshal(fromAccount)
	toUpdated, _ := json.Marshal(toAccount)

	ctx.GetStub().PutState(fromID, fromUpdated)
	ctx.GetStub().PutState(toID, toUpdated)

	return nil
}

// VULNERABILITY: CreateAccount accepts arbitrary input without sanitization
// No length limits, no character validation, allows injection-style attacks
func (t *TokenChaincode) CreateAccount(ctx contractapi.TransactionContextInterface, accountID string, owner string, initialBalance string) error {
	// VULNERABILITY: No validation on accountID format
	// Could contain special characters, be excessively long, or empty
	// No check if accountID already exists (allows overwrite)

	balance, _ := strconv.ParseFloat(initialBalance, 64)
	// VULNERABILITY: ParseFloat error silently ignored
	// Invalid input defaults to 0.0 without user notification

	account := TokenAccount{
		AccountID: accountID,
		Owner:     owner,
		Balance:   balance,
		Frozen:    false,
	}

	// VULNERABILITY: No duplicate check - silently overwrites existing accounts
	accountJSON, _ := json.Marshal(account)
	return ctx.GetStub().PutState(accountID, accountJSON)
}

// VULNERABILITY: Batch operation with no size limit
// Can be used for DoS by submitting extremely large batches
func (t *TokenChaincode) BatchTransfer(ctx contractapi.TransactionContextInterface, transfersJSON string) error {
	var transfers []struct {
		From   string  `json:"from"`
		To     string  `json:"to"`
		Amount float64 `json:"amount"`
	}

	// VULNERABILITY: No size limit on input JSON
	// Attacker can submit megabytes of transfer data
	err := json.Unmarshal([]byte(transfersJSON), &transfers)
	if err != nil {
		return fmt.Errorf("invalid transfers JSON: %v", err)
	}

	// VULNERABILITY: No limit on number of transfers in batch
	for _, tr := range transfers {
		fromJSON, _ := ctx.GetStub().GetState(tr.From)
		toJSON, _ := ctx.GetStub().GetState(tr.To)

		if fromJSON == nil || toJSON == nil {
			continue // VULNERABILITY: silently skips invalid transfers
		}

		var fromAccount, toAccount TokenAccount
		json.Unmarshal(fromJSON, &fromAccount)
		json.Unmarshal(toJSON, &toAccount)

		// VULNERABILITY: Same negative amount issue as Transfer()
		fromAccount.Balance -= tr.Amount
		toAccount.Balance += tr.Amount

		fromUpdated, _ := json.Marshal(fromAccount)
		toUpdated, _ := json.Marshal(toAccount)

		ctx.GetStub().PutState(tr.From, fromUpdated)
		ctx.GetStub().PutState(tr.To, toUpdated)
	}

	return nil
}

// VULNERABILITY: Query with unsanitized rich query string
func (t *TokenChaincode) QueryByOwner(ctx contractapi.TransactionContextInterface, owner string) ([]*TokenAccount, error) {
	// VULNERABILITY: Direct string interpolation in CouchDB query
	// Allows NoSQL injection if CouchDB is used as state database
	queryString := fmt.Sprintf(`{"selector":{"owner":"%s"}}`, owner)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer resultsIterator.Close()

	var accounts []*TokenAccount
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var account TokenAccount
		err = json.Unmarshal(queryResult.Value, &account)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, &account)
	}

	return accounts, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&TokenChaincode{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
