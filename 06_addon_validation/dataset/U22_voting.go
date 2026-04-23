package main

import (
    "crypto/sha256"
    "encoding/json"
    "fmt"
    "strconv"

    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a Voting
type SmartContract struct {
    contractapi.Contract
}

// CampaignStatus defines the possible states of a voting campaign
type CampaignStatus string

const (
    CampaignStatusOpen    CampaignStatus = "OPEN"
    CampaignStatusReveal  CampaignStatus = "REVEAL"
    CampaignStatusClosed  CampaignStatus = "CLOSED"
    CampaignStatusRemoved CampaignStatus = "REMOVED"
    TargetTotalWeight     int            = 100000
)

// Campaign represents a voting campaign
type Campaign struct {
    ID                   string         `json:"ID"`
    Name                 string         `json:"Name"`
    Options              []string       `json:"Options"`
    Weights              map[string]int `json:"Weights"` // Map[MspID]int
    Status               CampaignStatus `json:"Status"`
    TotalCommittedWeight int            `json:"TotalCommittedWeight"`
    TotalRevealedWeight  int            `json:"TotalRevealedWeight"`
    Results              map[string]int `json:"Results"` // Map[Option]int
    Winner               string         `json:"Winner"`
}

// Vote represents a committed or revealed vote by an organization
type Vote struct {
    CampaignID     string `json:"CampaignID"`
    MspID          string `json:"MspID"`
    CommittedHash  string `json:"CommittedHash"`  // The hash submitted by the voter
    RevealedOption string `json:"RevealedOption"` // The option revealed by the voter
    Salt           string `json:"Salt"`           // The secret key (salt) revealed by the voter
    Weight         int    `json:"Weight"`         // The weight of the vote
}

// InitLedger adds a base set of campaigns to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
    // No initial campaigns are added. Campaigns are created by the manager.
    // This function can be used for initial setup if needed in the future.
    return nil
}

// CreateCampaign creates a new voting campaign. Only the manager (Org6) can call this.
func (s *SmartContract) CreateCampaign(ctx contractapi.TransactionContextInterface, campaignID string, name string, optionsJSON string, weightsJSON string) error {
    clientIdentity := ctx.GetClientIdentity()
    callingOrgMSP, err := clientIdentity.GetMSPID()
    if err != nil {
        return fmt.Errorf("failed to get client's MSPID: %v", err)
    }

    // Only Org6MSP (Manager) can create campaigns
    if callingOrgMSP != "Org6MSP" {
        return fmt.Errorf("only Org6MSP (manager) can create campaigns")
    }

    campaignAsBytes, err := ctx.GetStub().GetState(campaignID)
    if err != nil {
        return fmt.Errorf("failed to read from world state: %v", err)
    }
    if campaignAsBytes != nil {
        return fmt.Errorf("campaign with ID %s already exists", campaignID)
    }

    var options []string
    err = json.Unmarshal([]byte(optionsJSON), &options)
    if err != nil {
        return fmt.Errorf("invalid options JSON: %v", err)
    }
    if len(options) < 2 {
        return fmt.Errorf("a campaign must have at least two options")
    }

    var weights map[string]int
    err = json.Unmarshal([]byte(weightsJSON), &weights)
    if err != nil {
        return fmt.Errorf("invalid weights JSON: %v", err)
    }
    if len(weights) == 0 {
        return fmt.Errorf("campaign must have weights assigned to organizations")
    }

    totalWeight := 0
    for mspid, weight := range weights {
        if weight <= 0 || weight > TargetTotalWeight {
            return fmt.Errorf("invalid weight %d for %s. Weights must be between 1 and %d", weight, mspid, TargetTotalWeight)
        }
        totalWeight += weight
    }
    if totalWeight != TargetTotalWeight {
        return fmt.Errorf("total weight must be exactly %d, got %d", TargetTotalWeight, totalWeight)
    }

    campaign := Campaign{
        ID:                   campaignID,
        Name:                 name,
        Options:              options,
        Weights:              weights,
        Status:               CampaignStatusOpen,
        TotalCommittedWeight: 0,
        TotalRevealedWeight:  0,
        Results:              make(map[string]int),
    }
    campaignJSON, err := json.Marshal(campaign)
    if err != nil {
        return fmt.Errorf("failed to marshal campaign: %v", err)
    }

    return ctx.GetStub().PutState(campaignID, campaignJSON)
}

// RemoveCampaign sets a campaign to REMOVED state.
func (s *SmartContract) RemoveCampaign(ctx contractapi.TransactionContextInterface, campaignID string) error {
    clientIdentity := ctx.GetClientIdentity()
    callingOrgMSP, err := clientIdentity.GetMSPID()
    if err != nil {
        return fmt.Errorf("failed to get client's MSPID: %v", err)
    }

    if callingOrgMSP != "Org6MSP" {
        return fmt.Errorf("only Org6MSP (manager) can remove campaigns")
    }

    campaignAsBytes, err := ctx.GetStub().GetState(campaignID)
    if err != nil {
        return fmt.Errorf("failed to read from world state: %v", err)
    }
    if campaignAsBytes == nil {
        return fmt.Errorf("campaign with ID %s does not exist", campaignID)
    }

    var campaign Campaign
    err = json.Unmarshal(campaignAsBytes, &campaign)
    if err != nil {
        return fmt.Errorf("failed to unmarshal campaign: %v", err)
    }

    campaign.Status = CampaignStatusRemoved
    campaignJSON, err := json.Marshal(campaign)
    if err != nil {
        return fmt.Errorf("failed to marshal updated campaign: %v", err)
    }

    return ctx.GetStub().PutState(campaignID, campaignJSON)
}

// CommitVote allows a voter to commit their hashed vote.
func (s *SmartContract) CommitVote(ctx contractapi.TransactionContextInterface, campaignID string, committedHash string) error {
    clientIdentity := ctx.GetClientIdentity()
    callingOrgMSP, err := clientIdentity.GetMSPID()
    if err != nil {
        return fmt.Errorf("failed to get client's MSPID: %v", err)
    }

    campaignAsBytes, err := ctx.GetStub().GetState(campaignID)
    if err != nil {
        return fmt.Errorf("failed to read from world state: %v", err)
    }
    if campaignAsBytes == nil {
        return fmt.Errorf("campaign with ID %s does not exist", campaignID)
    }

    var campaign Campaign
    err = json.Unmarshal(campaignAsBytes, &campaign)
    if err != nil {
        return fmt.Errorf("failed to unmarshal campaign: %v", err)
    }

    if campaign.Status != CampaignStatusOpen {
        return fmt.Errorf("campaign %s is not open for voting", campaignID)
    }

    assignedWeight, exists := campaign.Weights[callingOrgMSP]
    if !exists {
        return fmt.Errorf("organization %s is not authorized to vote in campaign %s or has no assigned weight", callingOrgMSP, campaignID)
    }

    // Check if this organization has already committed a vote for this campaign
    voteID := fmt.Sprintf("%s_%s", campaignID, callingOrgMSP)

    voteAsBytes, err := ctx.GetStub().GetState(voteID)
    if err != nil {
        return fmt.Errorf("failed to read vote from world state: %v", err)
    }
    if voteAsBytes != nil {
        return fmt.Errorf("organization %s has already committed a vote for campaign %s", callingOrgMSP, campaignID)
    }

    vote := Vote{
        CampaignID:    campaignID,
        MspID:         callingOrgMSP,
        CommittedHash: committedHash,
        Weight:        assignedWeight,
    }

    voteJSON, err := json.Marshal(vote)
    if err != nil {
        return fmt.Errorf("failed to marshal vote: %v", err)
    }

    err = ctx.GetStub().PutState(voteID, voteJSON)
    if err != nil {
        return fmt.Errorf("failed to put vote to world state: %v", err)
    }

    // Update TotalCommittedWeight in campaign
    campaign.TotalCommittedWeight += assignedWeight
    campaignUpdatedJSON, err := json.Marshal(campaign)
    if err != nil {
        return fmt.Errorf("failed to marshal updated campaign: %v", err)
    }
    return ctx.GetStub().PutState(campaignID, campaignUpdatedJSON)
}

// RevealVote allows a voter to reveal their vote.
func (s *SmartContract) RevealVote(ctx contractapi.TransactionContextInterface, campaignID string, option string, salt string) error {
    clientIdentity := ctx.GetClientIdentity()
    callingOrgMSP, err := clientIdentity.GetMSPID()
    if err != nil {
        return fmt.Errorf("failed to get client's MSPID: %v", err)
    }

    campaignAsBytes, err := ctx.GetStub().GetState(campaignID)
    if err != nil {
        return fmt.Errorf("failed to read from world state: %v", err)
    }
    if campaignAsBytes == nil {
        return fmt.Errorf("campaign with ID %s does not exist", campaignID)
    }

    var campaign Campaign
    err = json.Unmarshal(campaignAsBytes, &campaign)
    if err != nil {
        return fmt.Errorf("failed to unmarshal campaign: %v", err)
    }

    if campaign.Status != CampaignStatusReveal {
        return fmt.Errorf("campaign %s is not in reveal phase", campaignID)
    }

    voteID := fmt.Sprintf("%s_%s", campaignID, callingOrgMSP)
    voteAsBytes, err := ctx.GetStub().GetState(voteID)
    if err != nil {
        return fmt.Errorf("failed to read vote from world state: %v", err)
    }
    if voteAsBytes == nil {
        return fmt.Errorf("organization %s has not committed a vote for campaign %s", callingOrgMSP, campaignID)
    }

    var vote Vote
    err = json.Unmarshal(voteAsBytes, &vote)
    if err != nil {
        return fmt.Errorf("failed to unmarshal vote: %v", err)
    }

    if vote.RevealedOption != "" {
        return fmt.Errorf("organization %s has already revealed its vote for campaign %s", callingOrgMSP, campaignID)
    }

    // Verify the revealed vote against the committed hash
    assignedWeight, exists := campaign.Weights[callingOrgMSP]
    if !exists {
        // This should not happen if CommitVote passed, but as a safeguard
        return fmt.Errorf("organization %s has no assigned weight for campaign %s", callingOrgMSP, campaignID)
    }

    // Recalculate hash with revealed option, salt, and assigned weight
    // Use SHA256 for hashing, matching typical use cases
    recalculatedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(option+salt+strconv.Itoa(assignedWeight))))

    if recalculatedHash != vote.CommittedHash {
        return fmt.Errorf("revealed vote does not match committed hash for organization %s in campaign %s", callingOrgMSP, campaignID)
    }

    // Check if the revealed option is valid for the campaign
    optionValid := false
    for _, opt := range campaign.Options {
        if opt == option {
            optionValid = true
            break
        }
    }
    if !optionValid {
        return fmt.Errorf("invalid option '%s' revealed for campaign %s", option, campaignID)
    }

    // Update vote state with revealed info
    vote.RevealedOption = option
    vote.Salt = salt
    voteJSON, err := json.Marshal(vote)
    if err != nil {
        return fmt.Errorf("failed to marshal updated vote: %v", err)
    }
    err = ctx.GetStub().PutState(voteID, voteJSON)
    if err != nil {
        return fmt.Errorf("failed to put updated vote to world state: %v", err)
    }

    // Update campaign's revealed weight AND results.
    campaign.TotalRevealedWeight += assignedWeight
    if campaign.Results == nil {
        campaign.Results = make(map[string]int)
    }
    campaign.Results[option] += assignedWeight

    campaignUpdatedJSON, err := json.Marshal(campaign)
    if err != nil {
        return fmt.Errorf("failed to marshal updated campaign: %v", err)
    }
    return ctx.GetStub().PutState(campaignID, campaignUpdatedJSON)
}

// CheckStatus allows the manager to transition the campaign to REVEAL phase.
func (s *SmartContract) CheckStatus(ctx contractapi.TransactionContextInterface, campaignID string) error {
    clientIdentity := ctx.GetClientIdentity()
    callingOrgMSP, err := clientIdentity.GetMSPID()
    if err != nil {
        return fmt.Errorf("failed to get client's MSPID: %v", err)
    }

    // Only Org6MSP (Manager) can check and transition campaign status
    if callingOrgMSP != "Org6MSP" {
        return fmt.Errorf("only Org6MSP (manager) can check and transition campaign status")
    }

    campaignAsBytes, err := ctx.GetStub().GetState(campaignID)
    if err != nil {
        return fmt.Errorf("failed to read from world state: %v", err)
    }
    if campaignAsBytes == nil {
        return fmt.Errorf("campaign with ID %s does not exist", campaignID)
    }

    var campaign Campaign
    err = json.Unmarshal(campaignAsBytes, &campaign)
    if err != nil {
        return fmt.Errorf("failed to unmarshal campaign: %v", err)
    }

    if campaign.Status == CampaignStatusOpen && campaign.TotalCommittedWeight >= TargetTotalWeight {
        campaign.Status = CampaignStatusReveal
        campaignJSON, err := json.Marshal(campaign)
        if err != nil {
            return fmt.Errorf("failed to marshal updated campaign: %v", err)
        }
        return ctx.GetStub().PutState(campaignID, campaignJSON)
    } else if campaign.Status == CampaignStatusClosed {
        return fmt.Errorf("campaign %s is already closed", campaignID)
    }

    return fmt.Errorf("campaign %s's committed weight (%d) has not reached %d, or is not in OPEN status", campaignID, campaign.TotalCommittedWeight, TargetTotalWeight)
}

// CloseCampaign allows the manager to close a campaign and determine the winner.
func (s *SmartContract) CloseCampaign(ctx contractapi.TransactionContextInterface, campaignID string) error {
    clientIdentity := ctx.GetClientIdentity()
    callingOrgMSP, err := clientIdentity.GetMSPID()
    if err != nil {
        return fmt.Errorf("failed to get client's MSPID: %v", err)
    }

    // Only Org6MSP (Manager) can close campaigns
    if callingOrgMSP != "Org6MSP" {
        return fmt.Errorf("only Org6MSP (manager) can close campaigns")
    }

    campaignAsBytes, err := ctx.GetStub().GetState(campaignID)
    if err != nil {
        return fmt.Errorf("failed to read from world state: %v", err)
    }
    if campaignAsBytes == nil {
        return fmt.Errorf("campaign with ID %s does not exist", campaignID)
    }

    var campaign Campaign
    err = json.Unmarshal(campaignAsBytes, &campaign)
    if err != nil {
        return fmt.Errorf("failed to unmarshal campaign: %v", err)
    }

    if campaign.Status == CampaignStatusClosed {
        return fmt.Errorf("campaign %s is already closed", campaignID)
    }

    // Tally results from all votes associated with this campaign
    // Iterate through all organizations that were assigned weights to find their votes
    campaign.Results = make(map[string]int) // Reset results to ensure clean tally

    for mspID := range campaign.Weights {
        voteID := fmt.Sprintf("%s_%s", campaignID, mspID)
        voteAsBytes, err := ctx.GetStub().GetState(voteID)
        if err != nil {
            // Fail the transaction if any vote state cannot be retrieved to ensure result integrity.
            return fmt.Errorf("failed to read vote for %s: %v", mspID, err)
        }

        if voteAsBytes != nil {
            var vote Vote
            err = json.Unmarshal(voteAsBytes, &vote)
            if err != nil {
                return fmt.Errorf("failed to unmarshal vote for %s: %v", mspID, err)
            }

            if vote.RevealedOption != "" {
                campaign.Results[vote.RevealedOption] += vote.Weight
            }
        }
    }

    // Determine winner based on tallied results
    if len(campaign.Results) == 0 {
        campaign.Winner = "No winner (no votes revealed)"
    } else {
        maxWeight := -1
        winnerOption := ""
        tie := false
        for option, weight := range campaign.Results {
            if weight > maxWeight {
                maxWeight = weight
                winnerOption = option
                tie = false
            } else if weight == maxWeight {
                tie = true
            }
        }
        if tie {
            campaign.Winner = fmt.Sprintf("Tie among options with %d weight", maxWeight)
        } else {
            campaign.Winner = winnerOption
        }
    }

    campaign.Status = CampaignStatusClosed
    campaignJSON, err := json.Marshal(campaign)
    if err != nil {
        return fmt.Errorf("failed to marshal updated campaign: %v", err)
    }

    return ctx.GetStub().PutState(campaignID, campaignJSON)
}

// QueryAllCampaigns returns all campaigns found in world state
func (s *SmartContract) QueryAllCampaigns(ctx contractapi.TransactionContextInterface) ([]*Campaign, error) {
    // range query with empty string for startKey and endKey does an open-ended query of all assets in the chaincode namespace.
    resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    campaigns := []*Campaign{}
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var campaign Campaign
        err = json.Unmarshal(queryResponse.Value, &campaign)
        if err != nil {
            continue // Skip if not a campaign (could be a vote or other state)
        }

        // Filter out non-Campaign assets (e.g. Vote structs) by checking for mandatory Campaign fields.
        if campaign.Name == "" {
            continue
        }

        campaigns = append(campaigns, &campaign)
    }

    return campaigns, nil
}

// QueryCampaign returns the campaign stored in the world state with given id.
func (s *SmartContract) QueryCampaign(ctx contractapi.TransactionContextInterface, campaignID string) (*Campaign, error) {
    campaignAsBytes, err := ctx.GetStub().GetState(campaignID)
    if err != nil {
        return nil, fmt.Errorf("failed to read from world state: %v", err)
    }
    if campaignAsBytes == nil {
        return nil, fmt.Errorf("%s does not exist", campaignID)
    }

    var campaign Campaign
    err = json.Unmarshal(campaignAsBytes, &campaign)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal campaign: %v", err)
    }

    return &campaign, nil
}

// QueryMyVote checks if the calling organization has voted in a specific campaign
func (s *SmartContract) QueryMyVote(ctx contractapi.TransactionContextInterface, campaignID string) (*Vote, error) {
    clientIdentity := ctx.GetClientIdentity()
    callingOrgMSP, err := clientIdentity.GetMSPID()
    if err != nil {
        return nil, fmt.Errorf("failed to get client's MSPID: %v", err)
    }

    voteID := fmt.Sprintf("%s_%s", campaignID, callingOrgMSP)

    voteAsBytes, err := ctx.GetStub().GetState(voteID)
    if err != nil {
        return nil, fmt.Errorf("failed to read vote from world state: %v", err)
    }
    if voteAsBytes == nil {
        return nil, nil // No vote found
    }

    var vote Vote
    err = json.Unmarshal(voteAsBytes, &vote)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal vote: %v", err)
    }

    return &vote, nil
}

// main function starts up the chaincode in the container
func main() {
    chaincode, err := contractapi.NewChaincode(&SmartContract{})
    if err != nil {
        fmt.Printf("Error creating secret voting chaincode: %s", err.Error())
        return
    }

    if err := chaincode.Start(); err != nil {
        fmt.Printf("Error starting secret voting chaincode: %s", err.Error())
    }
}
