package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// AuctionChaincode implements a simple auction system
type AuctionChaincode struct {
	contractapi.Contract
}

// Auction represents an auction item
type Auction struct {
	AuctionID  string  `json:"auctionId"`
	ItemName   string  `json:"itemName"`
	HighestBid float64 `json:"highestBid"`
	HighBidder string  `json:"highBidder"`
	Status     string  `json:"status"` // "open", "closed"
	BidCount   int     `json:"bidCount"`
}

// BidRecord stores individual bid history
type BidRecord struct {
	BidID     string  `json:"bidId"`
	AuctionID string  `json:"auctionId"`
	Bidder    string  `json:"bidder"`
	Amount    float64 `json:"amount"`
}

// VULNERABILITY: Read-after-write / Phantom Read
// Reading state and then writing based on stale data
// In Fabric, if another transaction modifies the same key between
// read and write in the same block, MVCC validation will fail
func (a *AuctionChaincode) PlaceBid(ctx contractapi.TransactionContextInterface, auctionID string, bidder string, bidAmount float64) error {
	// VULNERABILITY: Read the auction state
	auctionJSON, err := ctx.GetStub().GetState(auctionID)
	if err != nil {
		return fmt.Errorf("failed to read auction: %v", err)
	}
	if auctionJSON == nil {
		return fmt.Errorf("auction %s does not exist", auctionID)
	}

	var auction Auction
	err = json.Unmarshal(auctionJSON, &auction)
	if err != nil {
		return err
	}

	if auction.Status != "open" {
		return fmt.Errorf("auction is closed")
	}

	// VULNERABILITY: Decision based on stale read
	// If two bids arrive in the same block, both read the same highestBid
	// Both may pass this check, but only one will succeed at validation
	// The other transaction will be invalidated (phantom read / MVCC conflict)
	if bidAmount <= auction.HighestBid {
		return fmt.Errorf("bid must be higher than current highest: %.2f", auction.HighestBid)
	}

	// VULNERABILITY: Writing to the same key that was read
	// This creates a read-write conflict if concurrent transactions exist
	auction.HighestBid = bidAmount
	auction.HighBidder = bidder
	auction.BidCount++

	updatedJSON, err := json.Marshal(auction)
	if err != nil {
		return err
	}

	// VULNERABILITY: This write will cause MVCC validation failure
	// if another transaction modified auctionID in the same block
	err = ctx.GetStub().PutState(auctionID, updatedJSON)
	if err != nil {
		return err
	}

	// Store bid record (separate key - less prone to conflicts)
	bidRecord := BidRecord{
		BidID:     fmt.Sprintf("bid_%s_%d", auctionID, auction.BidCount),
		AuctionID: auctionID,
		Bidder:    bidder,
		Amount:    bidAmount,
	}
	bidJSON, _ := json.Marshal(bidRecord)
	ctx.GetStub().PutState(bidRecord.BidID, bidJSON)

	return nil
}

// VULNERABILITY: Counter increment pattern - classic phantom read issue
// Multiple transactions incrementing the same counter will conflict
func (a *AuctionChaincode) IncrementViewCount(ctx contractapi.TransactionContextInterface, auctionID string) error {
	counterKey := fmt.Sprintf("views_%s", auctionID)

	// VULNERABILITY: Read current counter
	counterJSON, err := ctx.GetStub().GetState(counterKey)
	if err != nil {
		return err
	}

	var count int
	if counterJSON != nil {
		json.Unmarshal(counterJSON, &count)
	}

	// VULNERABILITY: Increment and write back
	// If multiple transactions do this in the same block, all but one will fail
	count++
	countJSON, _ := json.Marshal(count)
	return ctx.GetStub().PutState(counterKey, countJSON)
}

// VULNERABILITY: Range query followed by write creates phantom read risk
// GetStateByRange creates read sets for all keys in range
func (a *AuctionChaincode) CloseExpiredAuctions(ctx contractapi.TransactionContextInterface) error {
	// VULNERABILITY: Range query reads ALL auction keys
	// Any modification to any auction key by another transaction
	// will invalidate this transaction
	resultsIterator, err := ctx.GetStub().GetStateByRange("auction_", "auction_~")
	if err != nil {
		return fmt.Errorf("failed to get auctions: %v", err)
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return err
		}

		var auction Auction
		err = json.Unmarshal(queryResult.Value, &auction)
		if err != nil {
			continue
		}

		// VULNERABILITY: Writing back to keys that were part of range query
		// Extremely high chance of MVCC conflict in busy networks
		if auction.Status == "open" && auction.BidCount == 0 {
			auction.Status = "closed"
			updatedJSON, _ := json.Marshal(auction)
			ctx.GetStub().PutState(queryResult.Key, updatedJSON)
		}
	}

	return nil
}

// CreateAuction creates a new auction
func (a *AuctionChaincode) CreateAuction(ctx contractapi.TransactionContextInterface, auctionID string, itemName string, startingBid float64) error {
	auction := Auction{
		AuctionID:  auctionID,
		ItemName:   itemName,
		HighestBid: startingBid,
		HighBidder: "",
		Status:     "open",
		BidCount:   0,
	}

	auctionJSON, err := json.Marshal(auction)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(auctionID, auctionJSON)
}

func main() {
	chaincode, err := contractapi.NewChaincode(&AuctionChaincode{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
