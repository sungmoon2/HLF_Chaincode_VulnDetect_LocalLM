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
// [SAFE PATTERN] crypto/sha256 is a pure, deterministic function.
// Given the same input, it always produces the same output on every peer.
// A model that flags "external library calls" as nondeterministic should
// recognize that standard-library cryptographic functions are deterministic.
func computeChecksum(accountID string, points int, tier string) string {
	data := fmt.Sprintf("%s:%d:%s", accountID, points, tier)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// determineTier calculates the loyalty tier based on points.
// [SAFE PATTERN] Pure function — no external state, no randomness.
// Given the same points value, returns the same tier on all peers.
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
// [SAFE PATTERN] Uses helper functions (computeChecksum, determineTier)
// that are all pure and deterministic. The crypto/sha256 import may look
// like an "external library" but is a Go standard library with guaranteed
// deterministic behavior. No network calls, no file I/O, no randomness.
func (l *LoyaltyPointContract) CreateAccount(ctx contractapi.TransactionContextInterface, accountID string, customerID string, initialPoints int) error {
	exists, err := l.AccountExists(ctx, accountID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("account %s already exists", accountID)
	}

	tier := determineTier(initialPoints) // [SAFE PATTERN] pure function

	account := LoyaltyAccount{
		AccountID:  accountID,
		CustomerID: customerID,
		Points:     initialPoints,
		Tier:       tier,
		Checksum:   computeChecksum(accountID, initialPoints, tier), // [SAFE PATTERN] deterministic hash
	}

	accountJSON, err := json.Marshal(account)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(accountID, accountJSON)
}

// AddPoints adds loyalty points and recalculates tier and checksum
// [SAFE PATTERN] strconv.Atoi is used to parse the points string.
// This is a deterministic parser — identical input always produces identical output.
// A model that classified strconv.Atoi as "nondeterministic" (as Llama did in
// previous experiments) would be demonstrating a hallucination.
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

	// [SAFE PATTERN] strconv.Atoi — pure deterministic function.
	// Converts string "500" to int 500 identically on all peers.
	addPoints, err := strconv.Atoi(pointsStr)
	if err != nil {
		return fmt.Errorf("invalid points value: %v", err)
	}

	account.Points += addPoints
	account.Tier = determineTier(account.Points)                                      // [SAFE PATTERN] pure
	account.Checksum = computeChecksum(account.AccountID, account.Points, account.Tier) // [SAFE PATTERN] deterministic

	updatedJSON, err := json.Marshal(account)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(accountID, updatedJSON)
}

// RedeemPoints deducts points for a reward and updates the checksum
// [SAFE PATTERN] All operations are deterministic:
// - Arithmetic (subtraction) is deterministic
// - determineTier is a pure function
// - computeChecksum uses crypto/sha256, which is deterministic
// - strconv.Itoa is deterministic
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

	// [SAFE PATTERN] Log uses strconv.Itoa — deterministic, but only for console output
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
