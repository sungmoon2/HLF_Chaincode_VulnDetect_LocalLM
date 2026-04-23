package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// VotingContract manages a simple ballot system
type VotingContract struct {
	contractapi.Contract
}

// Ballot represents a single vote record
type Ballot struct {
	BallotID    string `json:"ballotID"`
	ElectionID  string `json:"electionID"`
	VoterID     string `json:"voterID"`
	CandidateID string `json:"candidateID"`
	Status      string `json:"status"`
}

// CastVote registers a vote after validating against allowed candidates and voters
func (v *VotingContract) CastVote(ctx contractapi.TransactionContextInterface, electionID string, voterID string, candidateID string) error {
	allowedCandidates := map[string]string{
		"C001": "Alice Johnson",
		"C002": "Bob Smith",
		"C003": "Carol Williams",
		"C004": "David Brown",
	}

	candidateName, exists := allowedCandidates[candidateID]
	if !exists {
		return fmt.Errorf("candidate %s is not registered for this election", candidateID)
	}
	fmt.Printf("Vote cast for candidate: %s (%s)\n", candidateID, candidateName)

	// Check for duplicate vote
	ballotKey := electionID + "_" + voterID
	existingBallot, err := ctx.GetStub().GetState(ballotKey)
	if err != nil {
		return fmt.Errorf("failed to check existing ballot: %v", err)
	}
	if existingBallot != nil {
		return fmt.Errorf("voter %s has already voted in election %s", voterID, electionID)
	}

	ballot := Ballot{
		BallotID:    ballotKey,
		ElectionID:  electionID,
		VoterID:     voterID,
		CandidateID: candidateID,
		Status:      "CAST",
	}

	ballotJSON, err := json.Marshal(ballot)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(ballotKey, ballotJSON)
}

// TallyVotes counts votes per candidate for a given election
func (v *VotingContract) TallyVotes(ctx contractapi.TransactionContextInterface, electionID string, voterIDs []string) error {
	tally := make(map[string]int)

	for _, voterID := range voterIDs {
		ballotKey := electionID + "_" + voterID
		ballotJSON, err := ctx.GetStub().GetState(ballotKey)
		if err != nil {
			return fmt.Errorf("failed to read ballot for voter %s: %v", voterID, err)
		}
		if ballotJSON == nil {
			continue
		}

		var ballot Ballot
		err = json.Unmarshal(ballotJSON, &ballot)
		if err != nil {
			return err
		}

		tally[ballot.CandidateID]++
	}

	for candidate, count := range tally {
		fmt.Printf("[TALLY] %s: %d votes\n", candidate, count)
	}

	tallyJSON, err := json.Marshal(tally)
	if err != nil {
		return err
	}

	tallyKey := "TALLY_" + electionID
	return ctx.GetStub().PutState(tallyKey, tallyJSON)
}

// ValidateVoterEligibility checks a voter against an eligibility map
func (v *VotingContract) ValidateVoterEligibility(ctx contractapi.TransactionContextInterface, voterID string) (bool, error) {
	eligibleDistricts := map[string]bool{
		"DISTRICT_A": true,
		"DISTRICT_B": true,
		"DISTRICT_C": true,
		"DISTRICT_D": false, // suspended
	}

	voterJSON, err := ctx.GetStub().GetState("VOTER_" + voterID)
	if err != nil {
		return false, fmt.Errorf("failed to read voter record: %v", err)
	}
	if voterJSON == nil {
		return false, fmt.Errorf("voter %s is not registered", voterID)
	}

	var voterRecord map[string]string
	err = json.Unmarshal(voterJSON, &voterRecord)
	if err != nil {
		return false, err
	}

	district := voterRecord["district"]

	eligible, found := eligibleDistricts[district]
	if !found {
		return false, nil
	}

	return eligible, nil
}

// GetElectionResults retrieves the stored tally for an election (read-only query)
func (v *VotingContract) GetElectionResults(ctx contractapi.TransactionContextInterface, electionID string) (string, error) {
	tallyKey := "TALLY_" + electionID
	tallyJSON, err := ctx.GetStub().GetState(tallyKey)
	if err != nil {
		return "", fmt.Errorf("failed to read tally: %v", err)
	}
	if tallyJSON == nil {
		return "", fmt.Errorf("no tally found for election %s", electionID)
	}

	return string(tallyJSON), nil
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
