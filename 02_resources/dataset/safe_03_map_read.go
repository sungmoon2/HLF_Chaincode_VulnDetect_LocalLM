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
// [SAFE PATTERN] A Go map is created and iterated with `range`, but the iteration
// is used ONLY for an internal membership check (does the candidate exist in the
// allowed set?). The map iteration order does NOT affect the ledger write set —
// we write a single deterministic Ballot struct derived entirely from the function
// arguments. Different iteration orders across peers produce the same result.
func (v *VotingContract) CastVote(ctx contractapi.TransactionContextInterface, electionID string, voterID string, candidateID string) error {
	// [SAFE PATTERN] Map used as a lookup table for validation only.
	// The iteration order is irrelevant because we only check key existence
	// and accumulate no order-dependent output.
	allowedCandidates := map[string]string{
		"C001": "Alice Johnson",
		"C002": "Bob Smith",
		"C003": "Carol Williams",
		"C004": "David Brown",
	}

	// [SAFE PATTERN] Direct key lookup — O(1), no iteration, fully deterministic.
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
		CandidateID: candidateID, // [SAFE PATTERN] value from function argument, not map iteration
		Status:      "CAST",
	}

	ballotJSON, err := json.Marshal(ballot)
	if err != nil {
		return err
	}

	// [SAFE PATTERN] Single PutState with a deterministic key and deterministic value.
	// The map was only used for validation — it does not influence the write set.
	return ctx.GetStub().PutState(ballotKey, ballotJSON)
}

// TallyVotes counts votes per candidate for a given election
// [SAFE PATTERN] Iterates over a map to compute per-candidate tallies, but the
// final result written to the ledger is a JSON object. Go's json.Marshal sorts
// map keys alphabetically, so the serialized output is deterministic regardless
// of the in-memory iteration order.
func (v *VotingContract) TallyVotes(ctx contractapi.TransactionContextInterface, electionID string, voterIDs []string) error {
	// [SAFE PATTERN] tally map accumulates counts; iteration order does not matter
	// because the final value per key is determined by commutative addition.
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

	// [SAFE PATTERN] Logging the tally by iterating the map.
	// fmt.Printf output goes to the peer's local console only — it never
	// enters the write set. Even though the print order may vary across peers,
	// it has zero impact on consensus.
	for candidate, count := range tally {
		fmt.Printf("[TALLY] %s: %d votes\n", candidate, count)
	}

	// [SAFE PATTERN] json.Marshal on a map[string]int sorts keys alphabetically.
	// Peer A and Peer B both produce the same JSON bytes, e.g.:
	//   {"C001":5,"C002":3,"C003":7}
	// This makes the PutState payload deterministic.
	tallyJSON, err := json.Marshal(tally)
	if err != nil {
		return err
	}

	tallyKey := "TALLY_" + electionID
	return ctx.GetStub().PutState(tallyKey, tallyJSON)
}

// ValidateVoterEligibility checks a voter against an eligibility map
// [SAFE PATTERN] The map is used purely for a read-only membership test.
// No ledger writes occur inside the map iteration.
// The function returns a boolean result — completely independent of iteration order.
func (v *VotingContract) ValidateVoterEligibility(ctx contractapi.TransactionContextInterface, voterID string) (bool, error) {
	// [SAFE PATTERN] Eligibility map used only for lookup — never affects write set.
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

	// [SAFE PATTERN] Direct key lookup on the map — no iteration involved.
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
