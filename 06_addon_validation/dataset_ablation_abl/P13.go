package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// VotingContract manages a simple on-chain voting system
type VotingContract struct {
	contractapi.Contract
}

// Proposal represents a governance proposal
type Proposal struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	VoteCount   int      `json:"voteCount"`
	Voters      []string `json:"voters"`
	Status      string   `json:"status"`
	ResultHash  string   `json:"resultHash"`
}

// CreateProposal registers a new proposal
func (v *VotingContract) CreateProposal(ctx contractapi.TransactionContextInterface, id string, title string) error {
	proposal := Proposal{
		ID:        id,
		Title:     title,
		VoteCount: 0,
		Voters:    []string{},
		Status:    "OPEN",
	}

	proposalJSON, err := json.Marshal(proposal)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, proposalJSON)
}

// CastVote adds a vote to a proposal
func (v *VotingContract) CastVote(ctx contractapi.TransactionContextInterface, proposalID string, voterID string) error {
	proposalJSON, err := ctx.GetStub().GetState(proposalID)
	if err != nil {
		return fmt.Errorf("failed to read proposal: %v", err)
	}
	if proposalJSON == nil {
		return fmt.Errorf("proposal %s does not exist", proposalID)
	}

	var proposal Proposal
	json.Unmarshal(proposalJSON, &proposal)

	if proposal.Status != "OPEN" {
		return fmt.Errorf("proposal is not open for voting")
	}

	proposal.VoteCount++
	proposal.Voters = append(proposal.Voters, voterID)

	updatedJSON, _ := json.Marshal(proposal)
	return ctx.GetStub().PutState(proposalID, updatedJSON)
}

// TallyVotes counts votes using goroutines and writes the result
func (v *VotingContract) TallyVotes(ctx contractapi.TransactionContextInterface, proposalIDs []string) error {
	var mu sync.Mutex
	var wg sync.WaitGroup
	results := make(map[string]int)

	for _, pid := range proposalIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			proposalJSON, err := ctx.GetStub().GetState(id)
			if err != nil {
				return
			}
			if proposalJSON == nil {
				return
			}

			var proposal Proposal
			json.Unmarshal(proposalJSON, &proposal)

			mu.Lock()
			results[id] = proposal.VoteCount
			mu.Unlock()
		}(pid)
	}

	wg.Wait()

	summaryJSON, _ := json.Marshal(results)
	return ctx.GetStub().PutState("TALLY_RESULT", summaryJSON)
}

// BatchUpdateStatus updates multiple proposals concurrently
func (v *VotingContract) BatchUpdateStatus(ctx contractapi.TransactionContextInterface, proposalIDs []string, newStatus string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(proposalIDs))

	for _, pid := range proposalIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			proposalJSON, err := ctx.GetStub().GetState(id)
			if err != nil {
				errChan <- err
				return
			}
			if proposalJSON == nil {
				return
			}

			var proposal Proposal
			json.Unmarshal(proposalJSON, &proposal)
			proposal.Status = newStatus

			updatedJSON, _ := json.Marshal(proposal)
			err = ctx.GetStub().PutState(id, updatedJSON)
			if err != nil {
				errChan <- err
			}
		}(pid)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// AsyncNotify fires off a goroutine to emit events (fire-and-forget)
func (v *VotingContract) AsyncNotify(ctx contractapi.TransactionContextInterface, proposalID string, message string) error {
	go func() {
		payload := map[string]string{
			"proposal": proposalID,
			"message":  message,
		}
		eventData, _ := json.Marshal(payload)
		ctx.GetStub().SetEvent("NOTIFY", eventData)
	}()

	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&VotingContract{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
