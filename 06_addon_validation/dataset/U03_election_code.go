package core

import (
	"election_code/smart-contract/constants"
	"election_code/smart-contract/errors"
	"election_code/smart-contract/helpers"
	"election_code/smart-contract/structs"
	"election_code/smart-contract/structs/fun"
	"encoding/json"

	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ElectionChainCode is the main chaincode struct
type ElectionChainCode struct {
	contractapi.Contract
}

// Delete state deletes the value associated with key in the world state if the invoker is admin
func (t *ElectionChainCode) DeleteState(ctx contractapi.TransactionContextInterface, req fun.DeleteStateReq) error {
	if !helpers.IsRole(ctx, constants.AdminRole, constants.TrueString) {
		return errors.ErrACL("Delete State", constants.AdminRole)
	}
	return helpers.DeleteState(ctx, req.Key)
}

// Create election creates a new election entry and puts it in the world state if the invoker is admin
func (t *ElectionChainCode) CreateElection(ctx contractapi.TransactionContextInterface, req fun.CreateElectionReq) error {
	if !helpers.IsRole(ctx, constants.AdminRole, constants.TrueString) {
		return errors.ErrACL("Create Election", constants.AdminRole)
	}
	if err := helpers.CheckIfExists(ctx, req.ElectionId); err != nil {
		return err
	}
	newElection := structs.NewElection(structs.NewElectionReq(req))

	return helpers.PutState(ctx, req.ElectionId, newElection)
}

// Create votable items is a function that creates new votable items and puts it in a ballot of an election.
func (t *ElectionChainCode) CreateVotableItems(ctx contractapi.TransactionContextInterface, req fun.CreateVotableItemsReq) error {
	if !helpers.IsRole(ctx, constants.AdminRole, constants.TrueString) {
		return errors.ErrACL("Create Votable Items", constants.AdminRole)
	}

	var err error
	electionData, err := helpers.GetElectionDataInternal(ctx, fun.GetElectionDataReq{
		Key: req.ElectionIndex,
	})

	if err != nil {
		return err
	}

	if _, ok := electionData.Ballots[req.BallotIndex]; !ok {
		return errors.ErrBallotNotExist(req.BallotIndex)
	}

	if _, ok := electionData.Ballots[req.BallotIndex].VotableItems[req.VotableId]; ok {
		return errors.ErrDataExists
	}

	newVotable := structs.NewVotableItem(req.VotableId, req.Description)

	electionData.Ballots[req.BallotIndex].VotableItems[req.VotableId] = newVotable

	byteData, err := json.Marshal(electionData)
	if err != nil {
		return errors.ErrMarshalFailure("Create Votable Items")
	}

	return helpers.PutState(ctx, req.ElectionIndex, byteData)
}

// Create Ballot is a function that creates a new ballot and puts it inside an Election
func (t *ElectionChainCode) CreateBallot(ctx contractapi.TransactionContextInterface, req fun.CreateBallotReq) error {
	if !helpers.IsRole(ctx, constants.AdminRole, constants.TrueString) {
		return errors.ErrACL("Create Ballot", constants.AdminRole)
	}

	var err error
	electionStruct, err := helpers.GetElectionDataInternal(ctx, fun.GetElectionDataReq{
		Key: req.ElectionId,
	})
	if err != nil {
		return err
	}

	_, ok := electionStruct.Ballots[req.BallotId]
	if ok {
		return errors.ErrDataExists
	}

	newBallot := structs.NewBallot(structs.NewBallotReq{
		VotableItems: map[string]structs.VotableItem{},
		BallotCast:   req.BallotCast,
		BallotId:     req.BallotId,
	})

	electionStruct.Ballots[req.BallotId] = newBallot

	electionByte, err := json.Marshal(electionStruct)
	if err != nil {
		return errors.ErrMarshalFailure("Create Ballot")
	}
	return helpers.PutState(ctx, req.ElectionId, electionByte)
}

// Create Voter is a function that creates a voter
func (t *ElectionChainCode) CreateVoter(ctx contractapi.TransactionContextInterface, req fun.CreateVoterReq) error {
	if !helpers.IsRole(ctx, constants.AdminRole, constants.TrueString) {
		return errors.ErrACL("Create Votable Items", constants.AdminRole)
	}
	var err error
	//Check if voter exists
	if err = helpers.CheckIfExists(ctx, req.VoterId); err != nil {
		return err
	}
	//Create it if it does not exist
	newVoter := structs.NewVoter(structs.NewVoterReq{
		VoterId:     req.VoterId,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		RegistrarId: req.RegistrarId,
	})

	helpers.PutState(ctx, req.VoterId, newVoter)

	return nil
}

// Get Election data is a function that gets election data meant for viewing purposes.
// This is different from the internal function which gets the actual data struct, as this returns a custom return struct.
func (t *ElectionChainCode) GetElectionData(ctx contractapi.TransactionContextInterface, req fun.GetElectionDataReq) (fun.GetElectionDataRes, error) {
	if !helpers.IsRole(ctx, constants.AdminRole, constants.TrueString) {
		return fun.GetElectionDataRes{}, errors.ErrACL("Create Votable Items", constants.AdminRole)
	}
	election, err := helpers.GetElectionDataInternal(ctx, req)
	if err != nil {
		return fun.GetElectionDataRes{}, err
	}

	res := fun.GetElectionDataRes{
		ElectionId: election.ElectionId,
		Name:       election.Name,
		Country:    election.Country,
		Year:       election.Year,
		StartDate:  election.StartDate,
		EndDate:    election.EndDate,
	}
	for key, ballot := range election.Ballots {
		tmpBallot := fun.BallotRes{}
		for _, votable := range ballot.VotableItems {
			tmpBallot.VotableItems = append(tmpBallot.VotableItems, fun.VotableItemRes{
				VotableId:   votable.VotableId,
				Description: votable.Description,
				Count:       votable.Count,
			})
		}
		res.Ballots[key] = tmpBallot
	}
	return res, nil
}

// Vote is a function that allows a voter that has already registered their data to vote for a ballot only once.
func (t *ElectionChainCode) Vote(ctx contractapi.TransactionContextInterface, electionKey string, voterId string, ballotId string, votableId string) error {
	if !helpers.IsRole(ctx, constants.VoterRole, constants.TrueString) {
		return errors.ErrACL("Create Votable Items", constants.AdminRole)
	}
	//Check if Election Exists
	electionStruct, err := helpers.GetElectionDataInternal(ctx, fun.GetElectionDataReq{
		Key: electionKey,
	})
	if err != nil {
		return err
	}
	now := time.Now()
	endDate, err := time.Parse(constants.DateFormat, electionStruct.EndDate)
	if err != nil {
		return err
	}
	startDate, err := time.Parse(constants.DateFormat, electionStruct.StartDate)
	if err != nil {
		return err
	}
	if startDate.After(now) && endDate.Before(now) {
		return errors.ErrNotElectionTime
	}

	//Check if voter exists, and if it does, unmarshal it into its struct representation
	voter, err := helpers.GetState(ctx, voterId)
	if err != nil {
		return err
	}
	if voter == nil {
		return errors.ErrVoterNotExist(voterId)
	}

	var voterStruct structs.Voter
	json.Unmarshal(voter, &voterStruct)

	//Check if user has voted before
	isVoted, ok := voterStruct.BallotVoted[ballotId]
	if ok && isVoted {
		return errors.ErrAlreadyVoted
	}

	//Check if ballot exists
	ballot, ok := electionStruct.Ballots[ballotId]
	if !ok {
		return errors.ErrBallotNotExist(ballotId)
	}

	//Check if candidate exists
	votable, ok := ballot.VotableItems[votableId]
	if !ok {
		return errors.ErrVotableItemNotExist(votableId)
	}

	//Increment vote count for the candidate and put inside world state
	addRes, err := helpers.Add(votable.Count, 1)
	if err != nil {
		return err
	}

	votable.Count = addRes

	electionStruct.Ballots[ballotId].VotableItems[votableId] = votable

	electionStructByteData, err := json.Marshal(electionStruct)
	if err != nil {
		return err
	}

	err = helpers.PutState(ctx, electionKey, electionStructByteData)
	if err != nil {
		return err
	}

	//Mark the fact that the user has voted for this ballot
	voterStruct.BallotVoted[ballotId] = true
	voter, err = json.Marshal(voterStruct)
	if err != nil {
		return err
	}

	return helpers.PutState(ctx, voterId, voter)
}
