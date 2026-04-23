package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// LoyaltyPointContract manages customer loyalty points
type LoyaltyPointContract struct {
	contractapi.Contract
}

// LoyaltyAccount represents a customer's loyalty point balance
type LoyaltyAccount struct {
	AccountID  string `json:"accountID"`
	CustomerID string `json:"customerID"`
	Points     int    `json:"points"`
	Tier       string `json:"tier"`
	Checksum   string `json:"checksum"`
}

// computeChecksum uses the crypto/sha256 standard library to generate
// a deterministic hash of the account state.
func computeChecksum(accountID string, points int, tier string) string {
	data := fmt.Sprintf("%s:%d:%s", accountID, points, tier)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// determineTier calculates the loyalty tier based on points.
func determineTier(points int) string {
	switch {
	case points >= 10000:
		return "PLATINUM"
	case points >= 5000:
		return "GOLD"
	case points >= 1000:
		return "SILVER"
	default:
		return "BRONZE"
	}
}

// CreateAccount initializes a loyalty account on the ledger
func (l *LoyaltyPointContract) CreateAccount(ctx contractapi.TransactionContextInterface, accountID string, customerID string, initialPoints int) error {
	exists, err := l.AccountExists(ctx, accountID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("account %s already exists", accountID)
	}

	tier := determineTier(initialPoints)

	account := LoyaltyAccount{
		AccountID:  accountID,
		CustomerID: customerID,
		Points:     initialPoints,
		Tier:       tier,
		Checksum:   computeChecksum(accountID, initialPoints, tier),
	}

	accountJSON, err := json.Marshal(account)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(accountID, accountJSON)
}

// AddPoints adds loyalty points and recalculates tier and checksum
func (l *LoyaltyPointContract) AddPoints(ctx contractapi.TransactionContextInterface, accountID string, pointsStr string) error {
	accountJSON, err := ctx.GetStub().GetState(accountID)
	if err != nil {
		return fmt.Errorf("failed to read account %s: %v", accountID, err)
	}
	if accountJSON == nil {
		return fmt.Errorf("account %s does not exist", accountID)
	}

	var account LoyaltyAccount
	err = json.Unmarshal(accountJSON, &account)
	if err != nil {
		return err
	}

	addPoints, err := strconv.Atoi(pointsStr)
	if err != nil {
		return fmt.Errorf("invalid points value: %v", err)
	}

	account.Points += addPoints
	account.Tier = determineTier(account.Points)
	account.Checksum = computeChecksum(account.AccountID, account.Points, account.Tier)

	updatedJSON, err := json.Marshal(account)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(accountID, updatedJSON)
}

// RedeemPoints deducts points for a reward and updates the checksum
func (l *LoyaltyPointContract) RedeemPoints(ctx contractapi.TransactionContextInterface, accountID string, redeemAmount int) error {
	accountJSON, err := ctx.GetStub().GetState(accountID)
	if err != nil {
		return fmt.Errorf("failed to read account %s: %v", accountID, err)
	}
	if accountJSON == nil {
		return fmt.Errorf("account %s does not exist", accountID)
	}

	var account LoyaltyAccount
	err = json.Unmarshal(accountJSON, &account)
	if err != nil {
		return err
	}

	if account.Points < redeemAmount {
		return fmt.Errorf("insufficient points: have %d, need %d", account.Points, redeemAmount)
	}

	account.Points -= redeemAmount
	account.Tier = determineTier(account.Points)
	account.Checksum = computeChecksum(account.AccountID, account.Points, account.Tier)

	fmt.Printf("[REDEEM] Account %s redeemed %s points, remaining: %s\n",
		accountID, strconv.Itoa(redeemAmount), strconv.Itoa(account.Points))

	updatedJSON, err := json.Marshal(account)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(accountID, updatedJSON)
}

// AccountExists checks whether an account with the given ID exists
func (l *LoyaltyPointContract) AccountExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	data, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return data != nil, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&LoyaltyPointContract{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
