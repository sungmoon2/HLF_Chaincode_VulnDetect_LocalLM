package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// AuctionContract manages a sealed-bid auction system
type AuctionContract struct {
	contractapi.Contract
}

// Bid represents an individual bid
type Bid struct {
	BidID    string  `json:"bidId"`
	Bidder   string  `json:"bidder"`
	Amount   float64 `json:"amount"`
	AuctionID string `json:"auctionId"`
}

// AuctionSummary holds aggregated auction results
type AuctionSummary struct {
	AuctionID    string  `json:"auctionId"`
	TotalBids    int     `json:"totalBids"`
	HighestBid   float64 `json:"highestBid"`
	HighestBidder string `json:"highestBidder"`
	TotalValue   float64 `json:"totalValue"`
}

// PlaceBid records a new bid
func (a *AuctionContract) PlaceBid(ctx contractapi.TransactionContextInterface, bidID string, bidder string, amount float64, auctionID string) error {
	bid := Bid{
		BidID:     bidID,
		Bidder:    bidder,
		Amount:    amount,
		AuctionID: auctionID,
	}

	bidJSON, err := json.Marshal(bid)
	if err != nil {
		return err
	}

	compositeKey, err := ctx.GetStub().CreateCompositeKey("Bid", []string{auctionID, bidID})
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(compositeKey, bidJSON)
}

// CloseAuction reads all bids, finds the winner, and writes the summary
func (a *AuctionContract) CloseAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("Bid", []string{auctionID})
	if err != nil {
		return fmt.Errorf("failed to query bids: %v", err)
	}
	defer resultsIterator.Close()

	var highestBid float64
	var highestBidder string
	totalBids := 0
	totalValue := 0.0

	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return err
		}

		var bid Bid
		json.Unmarshal(queryResult.Value, &bid)

		totalBids++
		totalValue += bid.Amount

		if bid.Amount > highestBid {
			highestBid = bid.Amount
			highestBidder = bid.Bidder
		}
	}

	summary := AuctionSummary{
		AuctionID:     auctionID,
		TotalBids:     totalBids,
		HighestBid:    highestBid,
		HighestBidder: highestBidder,
		TotalValue:    totalValue,
	}

	summaryJSON, _ := json.Marshal(summary)
	return ctx.GetStub().PutState("AUCTION_RESULT_"+auctionID, summaryJSON)
}

// IncrementCounter is the classic read-modify-write phantom read pattern
func (a *AuctionContract) IncrementCounter(ctx contractapi.TransactionContextInterface, counterKey string) error {
	counterBytes, err := ctx.GetStub().GetState(counterKey)
	if err != nil {
		return fmt.Errorf("failed to read counter: %v", err)
	}

	counter := 0
	if counterBytes != nil {
		counter, err = strconv.Atoi(string(counterBytes))
		if err != nil {
			return fmt.Errorf("invalid counter value: %v", err)
		}
	}

	counter++
	return ctx.GetStub().PutState(counterKey, []byte(strconv.Itoa(counter)))
}

// TransferWithBalanceCheck reads balance, checks, then updates
func (a *AuctionContract) TransferWithBalanceCheck(ctx contractapi.TransactionContextInterface, from string, to string, amount float64) error {
	fromBytes, err := ctx.GetStub().GetState(from)
	if err != nil {
		return err
	}
	if fromBytes == nil {
		return fmt.Errorf("sender %s not found", from)
	}

	fromBalance, _ := strconv.ParseFloat(string(fromBytes), 64)

	if fromBalance < amount {
		return fmt.Errorf("insufficient funds")
	}

	toBytes, _ := ctx.GetStub().GetState(to)
	toBalance := 0.0
	if toBytes != nil {
		toBalance, _ = strconv.ParseFloat(string(toBytes), 64)
	}

	ctx.GetStub().PutState(from, []byte(strconv.FormatFloat(fromBalance-amount, 'f', 2, 64)))
	return ctx.GetStub().PutState(to, []byte(strconv.FormatFloat(toBalance+amount, 'f', 2, 64)))
}

func main() {
	chaincode, err := contractapi.NewChaincode(&AuctionContract{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
