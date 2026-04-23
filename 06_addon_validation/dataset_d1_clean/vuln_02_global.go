package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

var transactionCounter int
var lastProcessedID string

// TokenContract manages a simple fungible token
type TokenContract struct {
	contractapi.Contract
}

// Token represents a token balance
type Token struct {
	Owner   string  `json:"owner"`
	Balance float64 `json:"balance"`
	TxCount int     `json:"txCount"`
	LastTx  string  `json:"lastTx"`
}

// IssueToken creates a new token balance for an owner
func (t *TokenContract) IssueToken(ctx contractapi.TransactionContextInterface, owner string, amount float64) error {
	transactionCounter++

	token := Token{
		Owner:   owner,
		Balance: amount,
		TxCount: transactionCounter,
		LastTx:  "ISSUE",
	}

	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return err
	}

	lastProcessedID = owner

	return ctx.GetStub().PutState(owner, tokenJSON)
}

// TransferToken moves tokens from one owner to another
func (t *TokenContract) TransferToken(ctx contractapi.TransactionContextInterface, from string, to string, amount float64) error {
	fromJSON, err := ctx.GetStub().GetState(from)
	if err != nil {
		return fmt.Errorf("failed to read sender: %v", err)
	}
	if fromJSON == nil {
		return fmt.Errorf("sender %s does not exist", from)
	}

	toJSON, err := ctx.GetStub().GetState(to)
	if err != nil {
		return fmt.Errorf("failed to read recipient: %v", err)
	}
	if toJSON == nil {
		return fmt.Errorf("recipient %s does not exist", to)
	}

	var fromToken, toToken Token
	json.Unmarshal(fromJSON, &fromToken)
	json.Unmarshal(toJSON, &toToken)

	if fromToken.Balance < amount {
		return fmt.Errorf("insufficient balance")
	}

	fromToken.Balance -= amount
	toToken.Balance += amount

	transactionCounter++
	fromToken.TxCount = transactionCounter
	fromToken.LastTx = "SEND"
	toToken.LastTx = "RECEIVE"

	fromBytes, _ := json.Marshal(fromToken)
	toBytes, _ := json.Marshal(toToken)

	ctx.GetStub().PutState(from, fromBytes)
	return ctx.GetStub().PutState(to, toBytes)
}

// GetGlobalStats returns the peer-local global counters (diagnostic function)
func (t *TokenContract) GetGlobalStats(ctx contractapi.TransactionContextInterface) (string, error) {
	return fmt.Sprintf("counter=%d, lastID=%s", transactionCounter, lastProcessedID), nil
}

// BurnToken destroys tokens by reducing balance
func (t *TokenContract) BurnToken(ctx contractapi.TransactionContextInterface, owner string, amount float64) error {
	tokenJSON, err := ctx.GetStub().GetState(owner)
	if err != nil {
		return fmt.Errorf("failed to read owner: %v", err)
	}
	if tokenJSON == nil {
		return fmt.Errorf("owner %s does not exist", owner)
	}

	var token Token
	json.Unmarshal(tokenJSON, &token)

	if token.Balance < amount {
		return fmt.Errorf("insufficient balance to burn")
	}

	token.Balance -= amount

	transactionCounter++
	receiptKey := "BURN_RECEIPT_" + strconv.Itoa(transactionCounter)

	receipt := map[string]interface{}{
		"owner":  owner,
		"burned": amount,
		"seq":    transactionCounter,
	}
	receiptJSON, _ := json.Marshal(receipt)
	ctx.GetStub().PutState(receiptKey, receiptJSON)

	token.TxCount = transactionCounter
	token.LastTx = "BURN"
	updatedJSON, _ := json.Marshal(token)
	return ctx.GetStub().PutState(owner, updatedJSON)
}

func main() {
	chaincode, err := contractapi.NewChaincode(&TokenContract{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
