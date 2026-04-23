# Add-on External Validation — 17 Go Chaincode Files

총 17파일, GitHub 공개 HLF 체인코드.

---

## U01_LandRegistry.go

- Bytes: 4911 | Lines: 166

```go
package contracts

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//SmartContract that provides functions for managing a land registration
type LandRegistry struct {
	contractapi.Contract
}

//LandRec describes basic details of the Land/Property
type PropRec struct {
	PropType  string `json:"proptype"`
	PropCity  string `json:"propcity"`
	PropState string `json:"propstate"`
	PropSqFt  string `json:"propsqft"`
	PropOwner string `json:"propowner"`
}

//Query Land/Property details
type QueryResult struct {
	Key    string `json:"Key"`
	Record *PropRec
}

//InitLedger - initialize with a set of properties
func (s *LandRegistry) InitLedger(ctx contractapi.TransactionContextInterface) error {
	proprecs := []PropRec{
		PropRec{PropType: "Flat", PropCity: "Chennai", PropState: "TN", PropSqFt: "1200", PropOwner: "Dev"},
		PropRec{PropType: "Ind House", PropCity: "Bengaluru", PropState: "KA", PropSqFt: "3200", PropOwner: "Abraham"},
		PropRec{PropType: "Res Plot", PropCity: "Coimbatore", PropState: "TN", PropSqFt: "4000", PropOwner: "Jagan"},
		PropRec{PropType: "Res Villa", PropCity: "Palakkad", PropState: "KL", PropSqFt: "4800", PropOwner: "John"},
		PropRec{PropType: "Farm Land", PropCity: "Coimbatore", PropState: "TN", PropSqFt: "100000", PropOwner: "Fasil"},
		PropRec{PropType: "Commercial Bldg", PropCity: "Chennai", PropState: "TN", PropSqFt: "1200", PropOwner: "Hema"},
	}
	for i, proprec := range proprecs {
		propAsBytes, _ := json.Marshal(proprec)
		err := ctx.GetStub().PutState("PROP"+strconv.Itoa(i), propAsBytes)

		if err != nil {
			return fmt.Errorf("failed to put to world state. %s", err.Error())
		}
	}
	return nil
}

//Register a Property on to the ledger
//Prop Counter or Transaction Hash could be used to assign Prop ID. Prod grade approach wud be generating in middleware - feedback
func (s *LandRegistry) CreateProp(ctx contractapi.TransactionContextInterface, propID string, propType string, propCity string, propSt string, propSqFt string, propOwn string) error {
	proprec := PropRec{
		PropType:  propType,
		PropCity:  propCity,
		PropState: propSt,
		PropSqFt:  propSqFt,
		PropOwner: propOwn,
	}

	propAsBytes, _ := json.Marshal(proprec)

	return ctx.GetStub().PutState(propID, propAsBytes)
}

//Query a Property based on Property ID

func (s *LandRegistry) QueryProp(ctx contractapi.TransactionContextInterface, propID string) (*PropRec, error) {
	propAsBytes, err := ctx.GetStub().GetState(propID)

	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}

	if propAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", propID)
	}

	proprec := new(PropRec)
	_ = json.Unmarshal(propAsBytes, proprec)

	return proprec, nil
}

// QueryAllCars returns all cars found in world state
//Feedback - When generating IDs in a dynamic way, this st key and end key option may not work
func (s *LandRegistry) ListAllProps(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		prop := new(PropRec)
		_ = json.Unmarshal(queryResponse.Value, prop)

		queryResult := QueryResult{Key: queryResponse.Key, Record: prop}
		results = append(results, queryResult)
	}

	return results, nil
}

//Property Ownership Transfer using Property ID
func (s *LandRegistry) ChangePropOwner(ctx contractapi.TransactionContextInterface, propID string, newPropOwner string) error {
	proprec, err := s.QueryProp(ctx, propID)

	if err != nil {
		return err
	}

	proprec.PropOwner = newPropOwner

	propAsBytes, _ := json.Marshal(proprec)

	return ctx.GetStub().PutState(propID, propAsBytes)
}

type PropOwnerRep struct {
	PropType string `json:"proptype"`
}

//Rich Queries Here
//GetAllPropsforOwner : Get the list of properites for the given owner
func (s *LandRegistry) GetAllPropsforOwner(ctx contractapi.TransactionContextInterface, powner string) ([]*PropOwnerRep, error) {
	//queryString := fmt.Sprintf(`{"selector":{"docType":"PropRec","powner": "%s"}}`, powner)
	queryString := fmt.Sprintf(`{"selector":{"propowner": "%s"}}`, powner)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var report []*PropOwnerRep
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var propOwnRep PropOwnerRep
		err = json.Unmarshal(queryResult.Value, &propOwnRep)
		if err != nil {
			return nil, err
		}
		report = append(report, &propOwnRep)

		fmt.Println("Report", report)
	}
	return report, nil
}

```

---

## U02_ethtxcc.go

- Bytes: 4147 | Lines: 123

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// Smart contract - Chaincode structure definition
// ===============================================
type EthDataLoaderChaincode struct {
}

type ethData struct {
	ObjectType string `json:"docType"`  // .
	Eth_id     string `json:"eth_id"`   // ethData index
	From       string `json:"from"`     // address of the sender
	Gas        string `json:"gas"`      // gas provided by the sender
	GasPrice   string `json:"gasPrice"` // gas price provided by the sender in Wei
	Hash       string `json:"hash"`     // hash of the transaction
	To         string `json:"to"`       // address of the receiver. null when its a contract creation transaction
	Value      string `value"`          // value transferred in Wei
}

// =============================================================================
// 			       Smart Contract/Chaincode Main Function
// =============================================================================

func main() {
	err := shim.Start(new(EthDataLoaderChaincode))
	if err != nil {
		fmt.Printf("Error starting ethDataLoader chaincode: %s", err)
	}
}

// Init - initializes chaincode
// =============================

func (t *EthDataLoaderChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println(" EtheDataloader Chaincode Initilized.")
	return shim.Success(nil)
}

// Invoke - the entry point for function invocations
// =====================================================

func (t *EthDataLoaderChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("Invoke method gets called ")
	function, args := stub.GetFunctionAndParameters()
	if function == "invoke" { // create an eth data instance
		return t.invoke(stub, args)
	} else if function == "query" { //read an eth data instance
		return t.query(stub, args)
	}

	fmt.Println("invoke did not find function: " + function)
	return shim.Error("Unknown function name. Expecting \"invoke\"\"query\"")
}

// ===================================================================
// invoke - set an EthTransaction Data instance from chaincode state
// ===================================================================
func (t *EthDataLoaderChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	fmt.Println("- start init EthTranction Data")

	// Eth_id:"0001", From: "0x122", Gas: "1000", GasPrice: "9000", hash: "0xakbkb1245", To: "0x234", Value: "9000000"
	// "0001", "0x122", "1000",  "9000", "0xakbkb1245", "0x234", "9000000"
	// "0002", "0x122", "1000",  "9000", "0xakbkb1245", "0x234", "9000000"

	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}

	eth_id := args[0]
	from := args[1]
	gas := args[2]
	gasPrice := args[3]
	hash := args[4]
	to := args[5]
	value := args[6]

	objectType := "ethData"
	ethData := &ethData{objectType, eth_id, from, gas, gasPrice, hash, to, value}
	ethDataJSONasBytes, err := json.Marshal(ethData)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(eth_id, ethDataJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end init ethTransaction data")
	return shim.Success(nil)
}

// ==============================================================
// query - read an Ethereum Data instance from chaincode state
// ==============================================================

func (t *EthDataLoaderChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var eth_id, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting only from address to query")
	}

	eth_id = args[0]
	valAsBytes, err := stub.GetState(eth_id)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state of: " + eth_id + "\"}"
		return shim.Error(jsonResp)
	} else if valAsBytes == nil {
		jsonResp = "{\"Error: \" \"The following eth data instance does not exist: " + eth_id + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(valAsBytes)
}

```

---

## U03_election_code.go

- Bytes: 7731 | Lines: 250

```go
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

```

---

## U07_ProductDetails.go

- Bytes: 5258 | Lines: 206

```go
package main;

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	"encoding/json"
)

type ProductDetailsContract struct {
	contractapi.Contract
}

/**
*@dev Product() represents the product details
*/

type Product struct {
	ID              uint64 `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	ManufactureDate uint64 `json:"manufactureDate"`
	BatchNumber     string `json:"batchNumber"`
	State ProductState `json:"state"`
}

/**
*@dev ProductHistory() represents the history of a product
*/

type ProductHistory struct {
	Timestamp uint64        `json:"timestamp"`
	Action    string        `json:"action"`
	Location  string        `json:"location"`
	State     ProductState `json:"state"`
}

/**
*@dev ProductState() represents the state of a product
*/

type ProductState int

const (
	PRODUCT_REGISTERED ProductState = iota
	QUALITY_ASSURANCE
	PRODUCT_TRANSIT
	PRODUCT_IN_INVENTORY
	PRODUCT_SOLD
	PRODUCT_RECALLED
	CONSUMPTION
	PENDING
	VALIDATING
	PUBLISHING
)

/**
@dev Init() initializes the chaincode
*/

func (c *ProductDetailsContract) Init(ctx contractapi.TransactionContextInterface) error {
	// Initialization later
	return nil
}

/**
*@dev AddProduct() adds a new product
*/

func (c *ProductDetailsContract) AddProduct(ctx contractapi.TransactionContextInterface, name string, description string, manufacturedDate uint64, batchNumber string) error {
	nextProductID, err := c.generateNextProductID(ctx)
	if err != nil {
		return err
	}

	product := Product{
		ID:              nextProductID,
		Name:            name,
		Description:     description,
		ManufactureDate: manufacturedDate,
		BatchNumber:     batchNumber,
	}

	err = ctx.GetStub().PutState(fmt.Sprintf("PRODUCT-%d", nextProductID), []byte(product));
	if err != nil {
		return fmt.Errorf("failed to put product on the ledger: %v", err)
	}

	return nil

}

/**
*@dev RetrieveProductDetails() retrieves the details of a product
*/

func (c *ProductDetailsContract) RetrieveProductDetails(ctx contractapi.TransactionContextInterface, productID uint64) (*Product, error) {
	productBytes, err := ctx.GetStub().GetState(fmt.Sprintf("PRODUCT-%d", productID))
	if err != nil {
		return nil, fmt.Errorf("failed to read product from the ledger: %v", err)
	}
	if productBytes == nil {
		return nil, fmt.Errorf("product with ID %d does not exist", productID)
	}

	product := new(Product)
	err = json.Unmarshal(productBytes, product)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal product JSON: %v", err)
	}

	return product, nil
}

/**
*@dev UpdateProductState() updates the state of a product
*/

func (c *ProductDetailsContract) UpdateProductState(ctx contractapi.TransactionContextInterface, productID uint64, currentState ProductState) error {
	product, err := c.RetrieveProductDetails(ctx, productID)
	if err != nil {
		return err
	}

	/**
	*@dev check for valid state transitions
    */

	if product.State == PRODUCT_REGISTERED && currentState != PRODUCT_TRANSIT {
		return fmt.Errorf("invalid state transition")
	}

	product.State = currentState
	productBytes, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal product JSON: %v", err)
	}

	err = ctx.GetStub().PutState(fmt.Sprintf("PRODUCT-%d", productID), productBytes)
	if err != nil {
		return fmt.Errorf("failed to put updated product state on the ledger: %v", err)
	}

	return nil
}

/**
*@dev LogProductMovement logs the movement of a product
*/

func (c *ProductDetailsContract) LogProductMovement(ctx contractapi.TransactionContextInterface, productID uint64, newLocation string) error {
	product, err := c.RetrieveProductDetails(ctx, productID)
	if err != nil {
		return err
	}

	productHistory := ProductHistory{
		Timestamp: uint64(ctx.GetStub().GetTxTimestamp().GetSeconds()),
		Action:    "Movement",
		Location:  newLocation,
		State:     product.State,
	}

	timestamp, _ := ctx.GetStub().GetTxTimestamp() // Error handling is not required here
    productHistory.Timestamp = uint64(timestamp.GetSeconds())

	historyKey := fmt.Sprintf("PRODUCT-%d-HISTORY", productID)
	existingHistoryBytes, err := ctx.GetStub().GetState(historyKey)
	if err != nil {
		return fmt.Errorf("failed to read product history from the ledger: %v", err)
	}

	var productHistories []ProductHistory
	if existingHistoryBytes != nil {
		err = json.Unmarshal(existingHistoryBytes, &productHistories)
		if err != nil {
			return fmt.Errorf("failed to unmarshal product history JSON: %v", err);
}

		// Unmarshal existing product histories
		err = json.Unmarshal(existingHistoryBytes, &productHistories)
		if err != nil {
			return fmt.Errorf("failed to unmarshal product history JSON: %v", err)
		}
	}

	// Append the new product history
	productHistories = append(productHistories, productHistory)

	// Marshal the updated product history
	updatedHistoryBytes, err := json.Marshal(productHistories)
	if err != nil {
		return fmt.Errorf("failed to marshal updated product history JSON: %v", err)
	}

	// Store the updated history on the ledger
	err = ctx.GetStub().PutState(historyKey, updatedHistoryBytes)
	if err != nil {
		return fmt.Errorf("failed to put updated product history on the ledger: %v", err)
	}

	return nil
}



```

---

## U08_smartcontract.go

- Bytes: 6306 | Lines: 219

```go
package main // Package main, Do not change this line.

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Product represents the structure for a product entity
type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Owner       string `json:"owner"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// SupplyChainContract defines the smart contract structure
type SupplyChainContract struct {
	contractapi.Contract
}

// getTimestamp returns the transaction timestamp as a string
func (s *SupplyChainContract) getTimestamp(ctx contractapi.TransactionContextInterface) (string, error) {
	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return "", fmt.Errorf("failed to get transaction timestamp: %v", err)
	}
	return time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339), nil
}

// InitLedger initializes the ledger with some example products
func (s *SupplyChainContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	timestamp, err := s.getTimestamp(ctx)
	if err != nil {
		return err
	}

	// Initial set of products to populate the ledger
	products := []Product{
		{ID: "p1", Name: "Laptop", Status: "Manufactured", Owner: "CompanyA", CreatedAt: timestamp, UpdatedAt: timestamp, Description: "High-end gaming laptop", Category: "Electronics"},
		{ID: "p2", Name: "Smartphone", Status: "Manufactured", Owner: "CompanyB", CreatedAt: timestamp, UpdatedAt: timestamp, Description: "Latest model smartphone", Category: "Electronics"},
	}

	for _, product := range products {
		if err := s.putProduct(ctx, &product); err != nil {
			return err
		}
	}

	return nil
}

// CreateProduct creates a new product in the ledger
func (s *SupplyChainContract) CreateProduct(ctx contractapi.TransactionContextInterface, id, name, owner, description, category string) error {
	// Write your implementation here
	exist, err := s.ProductExists(ctx, id)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("product already exists")
	}

	timestamp, err := s.getTimestamp(ctx)
	if err != nil {
		return err
	}

	product := Product{
		ID:          id,
		Name:        name,
		Status:      "Manufactured",
		Owner:       owner,
		CreatedAt:   timestamp,
		UpdatedAt:   timestamp,
		Description: description,
		Category:    category,
	}

	return s.putProduct(ctx, &product)
}

// UpdateProduct allows updating a product's status, owner, description, and category
func (s *SupplyChainContract) UpdateProduct(ctx contractapi.TransactionContextInterface, id string, newStatus string, newOwner string, newDescription string, newCategory string) error {
	// Write your implementation here
	product, err := s.getProduct(ctx, id)
	if err != nil {
		return err
	}

	timestamp, err := s.getTimestamp(ctx)
	if err != nil {
		return err
	}

	if newStatus != "" {
		product.Status = newStatus
	}
	if newOwner != "" {
		product.Owner = newOwner
	}
	if newDescription != "" {
		product.Description = newDescription
	}
	if newCategory != "" {
		product.Category = newCategory
	}
	product.UpdatedAt = timestamp

	return s.putProduct(ctx, product)
}

// TransferOwnership changes the owner of a product
func (s *SupplyChainContract) TransferOwnership(ctx contractapi.TransactionContextInterface, id, newOwner string) error {
	// Write your implementation here
	product, err := s.getProduct(ctx, id)
	if err != nil {
		return err
	}

	timestamp, err := s.getTimestamp(ctx)
	if err != nil {
		return err
	}

	product.Owner = newOwner
	product.UpdatedAt = timestamp

	return s.putProduct(ctx, product)
}

// QueryProduct retrieves a single product from the ledger by ID
func (s *SupplyChainContract) QueryProduct(ctx contractapi.TransactionContextInterface, id string) (*Product, error) {
	// Write your implementation here
	fmt.Printf("Querying product: %s\n", id)
	product, err := s.getProduct(ctx, id)
	if err != nil {
		return nil, err
	}
	return product, nil
}

// putProduct is a helper method for inserting or updating a product in the ledger
func (s *SupplyChainContract) putProduct(ctx contractapi.TransactionContextInterface, product *Product) error {
	productJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(product.ID, productJSON)
}

func (s *SupplyChainContract) getProduct(ctx contractapi.TransactionContextInterface, id string) (*Product, error) {
	productJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if productJSON == nil {
		return nil, fmt.Errorf("product does not exist")
	}

	var product Product
	if err := json.Unmarshal(productJSON, &product); err != nil {
		return nil, fmt.Errorf("Failed to parse product: %v", err)
	}

	return &product, nil
}

// ProductExists is a helper method to check if a product exists in the ledger
func (s *SupplyChainContract) ProductExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	productJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return productJSON != nil, nil
}

// GetAllProducts is a helper method to retrieve all products from the ledger
func (s *SupplyChainContract) GetAllProducts(ctx contractapi.TransactionContextInterface) ([]*Product, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var products []*Product
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var product Product
		if err := json.Unmarshal(queryResponse.Value, &product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	return products, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SupplyChainContract{})
	if err != nil {
		fmt.Printf("Error creating supply chain chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting supply chain chaincode: %s", err.Error())
	}
}

```

---

## U09_charity.go

- Bytes: 2737 | Lines: 104

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Donation structure
type Donation struct {
	DonationID string `json:"donationID"`
	DonorName  string `json:"donorName"`
	NGOName    string `json:"ngoName"`
	Amount     float64 `json:"amount"`
	Purpose    string `json:"purpose"`
	Status     string `json:"status"`
	Date       string `json:"date"`
}

// SmartContract defines contract structure
type SmartContract struct {
	contractapi.Contract
}

// CreateDonation – create a new donation entry
func (s *SmartContract) CreateDonation(ctx contractapi.TransactionContextInterface,
	id, donor, ngo, purpose, date string, amount float64) error {

	donation := Donation{
		DonationID: id, DonorName: donor, NGOName: ngo,
		Purpose: purpose, Date: date, Amount: amount, Status: "Created",
	}

	donationJSON, err := json.Marshal(donation)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, donationJSON)
}

// UpdateStatus – NGO updates donation status
func (s *SmartContract) UpdateStatus(ctx contractapi.TransactionContextInterface, id, status string) error {
	data, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to get donation: %v", err)
	}
	if data == nil {
		return fmt.Errorf("donation %s not found", id)
	}

	var donation Donation
	if err := json.Unmarshal(data, &donation); err != nil {
		return err
	}

	donation.Status = status
	updated, _ := json.Marshal(donation)
	return ctx.GetStub().PutState(id, updated)
}

// GetDonation – read donation by ID
func (s *SmartContract) GetDonation(ctx contractapi.TransactionContextInterface, id string) (*Donation, error) {
	data, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("donation %s not found", id)
	}
	var donation Donation
	_ = json.Unmarshal(data, &donation)
	return &donation, nil
}

// GetAllDonations – return all donations
func (s *SmartContract) GetAllDonations(ctx contractapi.TransactionContextInterface) ([]*Donation, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var donations []*Donation
	for resultsIterator.HasNext() {
		queryResponse, _ := resultsIterator.Next()
		var donation Donation
		_ = json.Unmarshal(queryResponse.Value, &donation)
		donations = append(donations, &donation)
	}
	return donations, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating chaincode: %s", err.Error())
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %s", err.Error())
	}
}

```

---

## U10_carcert.go

- Bytes: 6354 | Lines: 198

```go
/*
SPDX-License-Identifier: Apache-2.0
*/

package main
//package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes
// ID -> reference/serial number of the specific car part (ex: "120.47021-XXXXXXXXXXX")
// Car -> models of car using this specific car part (ex: "Fiat 500, Fiat Panda, Fiat Punto")
// Description -> detailt description of the car part (ex: "Symmetric vane; split-core castings; Black E-Coat anti-corrosive coating protects; Double disc ground friction surface")
// Brand -> Brand of the car part (ex: "Centric")
// ProductionDate ->  (ex: "DD/MM/YYYY")
// ProductionLocation -> (ex: "Saint Jose, US")
type Asset struct {
	ID 					string `json:"ID"`
	Car					string `json:"Car"`
	Brand          		string `json:"Brand"`
	ProductionDate      string `json:"ProductionDate"`
	ProductionLocation	string `json:"ProductionLocation"`
	Description         string `json:"Description"`
}

// adding a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "120.47021-15486957423", Car: "Audi Q1, Audi Q2, Audi Q3, Volkswagen Tiguan", Brand: "Volkswagen", ProductionDate: "04/11/2004", ProductionLocation: "Stuttgart, Germany", Description: "Chassis"},
		{ID: "115.15442-68495214587", Car: "Ford F150, Ford F250", Brand: "Ford", ProductionDate: "23/06/2011", ProductionLocation: "Detroit, US", Description: "Drive Train"},
		{ID: "254.51488-54875265847", Car: "Tesla model S, Tesla model 3, Tesla model Y", Brand: "Tesla", ProductionDate: "30/04/2019", ProductionLocation: "Austin, Texas", Description: "Battery"},
		{ID: "151.51847-84956877413", Car: "Toyota Corolla, Toyota rav4, Yahama  MT-15", Brand: "Toyota", ProductionDate: "12/09/2015", ProductionLocation: "Shanghai, China", Description: "Chip"},
		{ID: "58.41684-65184543156", Car: "Volkswagen Golf, Mini cooper S", Brand: "Thyssenkrupp Steering", ProductionDate: "19/02/2021", ProductionLocation: "Liechtenstein, Liechtenstein", Description: "Steering"},
		{ID: "456.56488-56464864115", Car: "Renault 5", Brand: "Renault", ProductionDate: "10/06/1996", ProductionLocation: "Montbéliard, France", Description: "Headlights"},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to init assets. %v", err)
		}
	}

	return nil
}

// CreateAsset -> create and adds new asset to the network.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, car string, brand string, productiondate string, productionlocation string, description string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	asset := Asset{
		ID:             	id,
		Car:          		car,
		Brand:           	brand,
		ProductionDate:     productiondate,
		ProductionLocation: productionlocation,
		Description:		description,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset -> returns specific asset stored in the network
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset -> updates existing asset in the network.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, car string, brand string, productiondate string, productionlocation string, description string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	// overwriting original asset with new asset
	asset := Asset{
		ID:             	id,
		Car:          		car,
		Brand:           	brand,
		ProductionDate:     productiondate,
		ProductionLocation: productionlocation,
		Description:		description,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// DeleteAsset -> deletes specific asset from the network.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AssetExists -> returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// GetAllAssets -> returns all assets in the network
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
	}
}


```

---

## U11_sharebook.go

- Bytes: 16026 | Lines: 563

```go
package main

import (
  "encoding/json"
  "fmt"
  "encoding/hex"
  "strconv"
  "strings"
  "log"
  "encoding/pem"
  "regexp"
  "crypto/x509"
  "github.com/hyperledger/fabric-contract-api-go/contractapi"
)



// SmartContract provides functions for managing an Book
type SmartContract struct {
  contractapi.Contract
}

type StudentConfig struct {
  Count          int `json:"Count"`
}
type BookConfig struct {
  Count          int `json:"Count"`
}
type Student struct {
  Org             string `json:"Org"`
  StudentID       string `json:"StudentID"`
  Name            string `json:"Name"`
  Phone           string `json:"Phone"`
  Email           string `json:"Email"`
}
type BookRequester struct {
  Org             string `json:"Org"`
  StudentID       string `json:"StudentID"`
}
// Book describes basic details of what makes up a simple book
type Book struct {
  ID             string `json:"ID"`
  Title          string `json:"Title"`
  Author         string `json:"Author"`
  ISBN           string `json:"Isbn"`
  Owner          string `json:"Owner"`
  Holder         BookRequester `json:"Holder"`
  IsBorrowed     bool   `json:"IsBorrowed"`
  RequestQueue   []BookRequester `json:"RequestQueue"`
  EntitleList    []string    `json:"EntitleList"`
  ReaderList    []string    `json:"ReaderList"`
}

func (s *SmartContract) GetClientName(ctx contractapi.TransactionContextInterface) (string,error) {
  caller,err := ctx.GetStub().GetCreator()
  if err != nil {
    return "",err
  }
  re := regexp.MustCompile("-----BEGIN CERTIFICATE-----[^ ]+-----END CERTIFICATE-----\n") 
  match := re.FindStringSubmatch(string(caller))
  pemBlock, _ := pem.Decode([]byte(match[0]))
  cert, err := x509.ParseCertificate(pemBlock.Bytes)
  if err != nil {
    return "",err
  }
  owner := cert.Subject.CommonName
  return owner,nil
}
// Register a new student to the world state with given details.
func (s *SmartContract) RegStudent(ctx contractapi.TransactionContextInterface) error {
  // Get new asset from transient map
  transientMap, err := ctx.GetStub().GetTransient()
  if err != nil {
        return fmt.Errorf("error getting transient: %v", err)
  }

  // Asset properties are private, therefore they get passed in transient field, instead of func args
  transientStudentJSON, ok := transientMap["student_properties"]
  if !ok {
	return fmt.Errorf("asset not found in the transient map input")
  }

  type StudentTransientInput struct {
		Name           string `json:"Name"`
		Phone          string `json:"Phone"`
		Email          string `json:"Email"`
  }

  var studentInput StudentTransientInput
  err = json.Unmarshal(transientStudentJSON, &studentInput)
  fmt.Println(studentInput)
  if err != nil {
    return err
  }
  caller,err := s.GetClientName(ctx)
  if err != nil {
    return err
  }
  fmt.Println("registering student by:")
  fmt.Println(caller)
  tokens := strings.Split(caller, "@")
  domain := ""
  if len(tokens) >= 2 {
      domain = tokens[1]
  }
  re := regexp.MustCompile("Admin@org([0-9]).example.com")
  match := re.FindStringSubmatch(string(caller))
  fmt.Println(match)
  PrivateCollection := ""
  if len(match) >= 2 {
     PrivateCollection = "_implicit_org_Org"+match[1]+"MSP"
  } else {
     return fmt.Errorf("failed to parse org string")
  }
  studentcfg := StudentConfig{ Count: 0 }
  configJSON, err := ctx.GetStub().GetPrivateData(PrivateCollection,"StudentConfig")
  if configJSON != nil {
     err = json.Unmarshal(configJSON, &studentcfg)
  }
  studentcfg.Count += 1
  configJSON2, err := json.Marshal(studentcfg)
  ctx.GetStub().PutPrivateData(PrivateCollection,"StudentConfig", configJSON2)
  if err != nil {
    return err
  }
  id := "student_" + strconv.Itoa(studentcfg.Count) + "@" + domain
  fmt.Println(id)
  student := Student{
    Org:            caller,
    StudentID:      id,
    Name:           studentInput.Name,
    Phone:          studentInput.Phone,
    Email:          studentInput.Email,
  }
  studentJSON, err := json.Marshal(student)
  if err != nil {
    return err
  }
  
  return ctx.GetStub().PutPrivateData(PrivateCollection,id, studentJSON)
}
  
// CreateBook issues a new book to the world state with given details.
func (s *SmartContract) CreateBook(ctx contractapi.TransactionContextInterface, title string, author string, isbn string ) error {
  owner,err := s.GetClientName(ctx)
  if err != nil {
    return err
  }
  log.Println("creating book by:")
  log.Println(owner)

  bookcfg := BookConfig{ Count: 0 }
  configJSON, err := ctx.GetStub().GetState("BookConfig")
  if configJSON != nil {
     err = json.Unmarshal(configJSON, &bookcfg)
  }
  bookcfg.Count += 1
  configJSON2, err := json.Marshal(bookcfg)
  ctx.GetStub().PutState("BookConfig", configJSON2)
  id := "book_" + strconv.Itoa(bookcfg.Count)
  log.Println(id)
  book := Book{
    ID:             id,
    Title:          title,
    Author:         author,
    ISBN:           isbn,
    Owner:          owner,
    Holder:         BookRequester{"",""},
    IsBorrowed:     false,
    RequestQueue:   make([]BookRequester,0),
    EntitleList:    make([]string,0),
    ReaderList:     make([]string,0),
  }
  book.EntitleList = append(book.EntitleList,owner)
  bookJSON, err := json.Marshal(book)
  if err != nil {
    return err
  }

  return ctx.GetStub().PutState(id, bookJSON)
}

// GetBook returns the book stored in the world state with given id.
func (s *SmartContract) GetBook(ctx contractapi.TransactionContextInterface, id string) (*Book, error) {
  bookJSON, err := ctx.GetStub().GetState(id)
  if err != nil {
    return nil, fmt.Errorf("failed to read from world state: %v", err)
  }
  if bookJSON == nil {
    return nil, fmt.Errorf("the book %s does not exist", id)
  }

  var book Book
  err = json.Unmarshal(bookJSON, &book)
  if err != nil {
    return nil, err
  }
  caller,err := s.GetClientName(ctx)
  if err != nil {
    return nil,err
  }
  for _, v := range book.EntitleList {
	if v == caller {
  		return &book, nil
	}
  }
  return nil,fmt.Errorf("the book %s not entitled", id)

}

func (s *SmartContract) GetStudentHash(ctx contractapi.TransactionContextInterface, client string , student string) (string,error) {
  PrivateCollection,err := s.GetPrivateCollection(ctx,client)
  if err != nil {
    return "",err
  }
  studentHash,err := ctx.GetStub().GetPrivateDataHash(PrivateCollection, student)
  if err != nil {
    return "",err
  }
  if studentHash == nil {
    return "",fmt.Errorf("cannot find the student")
  }
  return hex.EncodeToString(studentHash),nil
}
// AddRequest  let the client to add request an existing book in the world state with provided parameters.
func (s *SmartContract) AddRequest(ctx contractapi.TransactionContextInterface, id string , student string) error {
  exists, err := s.BookExists(ctx, id)
  if err != nil {
    return err
  }
  if !exists {
    return fmt.Errorf("the book %s does not exist", id)
  }

  book, err := s.GetBook(ctx, id)
  if err != nil {
    return err
  }
  if !book.IsBorrowed {
    return fmt.Errorf("the book %s is not borrowed", id)
  }
  caller,err := s.GetClientName(ctx)
  if err != nil {
    return err
  }
  studentHash,err := s.GetStudentHash(ctx,caller,student)
  if err != nil {
    return err
  }
  fmt.Printf("student hash: %s",studentHash) 
  for _, v := range book.EntitleList {
	if v == caller {
           for _, s := range book.ReaderList {
	      if s == studentHash {
                requester := BookRequester{caller,student}
	        book.RequestQueue = append(book.RequestQueue,requester)
		bookJSON, err := json.Marshal(book)
		if err != nil {
		    return err
		}
		return ctx.GetStub().PutState(id, bookJSON)
             }
           }
	}
  }
  return fmt.Errorf("the book %s is not entitled by you %s", id,caller)
}
// ReturnBook  let the client to return an existing book in the world state with provided parameters.
func (s *SmartContract) ReturnBook(ctx contractapi.TransactionContextInterface, id string , student string) error {
  exists, err := s.BookExists(ctx, id)
  if err != nil {
    return err
  }
  if !exists {
    return fmt.Errorf("the book %s does not exist", id)
  }

  book, err := s.GetBook(ctx, id)
  if err != nil {
    return err
  }
  if !book.IsBorrowed {
    return fmt.Errorf("the book %s is not borrowed", id)
  }
  caller,err := s.GetClientName(ctx)
  if err != nil {
    return err
  }
  if caller !=book.Holder.Org {
    return fmt.Errorf("the book %s is not borrowed by you %s", id,caller)
  }
  if student !=book.Holder.StudentID {
    return fmt.Errorf("the book %s is not borrowed by you %s", id,caller)
  }
  _,err = s.GetStudentHash(ctx,caller,student)
  if err != nil {
    return err
  }
  n := len(book.RequestQueue) 
  if n > 0 {
    new := make([]BookRequester,n-1)
    for i, v := range book.RequestQueue {
         if i == 0 {
            book.Holder = v
         } else {
	    new = append(new,v)
	 }
    }
    book.RequestQueue = new
  } else {
       book.IsBorrowed = false
  }
  bookJSON, err := json.Marshal(book)
  if err != nil {
    return err
  }
  return ctx.GetStub().PutState(id, bookJSON)
}
// BorrowBook  let the client to borrow an existing book in the world state with provided parameters.
func (s *SmartContract) BorrowBook(ctx contractapi.TransactionContextInterface, id string, student string) error {
  exists, err := s.BookExists(ctx, id)
  if err != nil {
    return err
  }
  if !exists {
    return fmt.Errorf("the book %s does not exist", id)
  }

  book, err := s.GetBook(ctx, id)
  if err != nil {
    return err
  }
  if book.IsBorrowed {
    return fmt.Errorf("the book %s has been borrowed", id)
  }
  caller,err := s.GetClientName(ctx)
  if err != nil {
    return err
  }
  studentHash,err := s.GetStudentHash(ctx,caller,student)
  if err != nil {
    return err
  }
  fmt.Printf("student hash: %s",studentHash) 
  for _, v := range book.EntitleList {
	if v == caller {
           for _, s := range book.ReaderList {
	      if s == studentHash {
		book.Holder = BookRequester{caller,student}
                book.IsBorrowed = true
                bookJSON, err := json.Marshal(book)
		if err != nil {
		    return err
		}
                return ctx.GetStub().PutState(id, bookJSON)
              }
           }
	}
  }
  return fmt.Errorf("the book %s not entitled", id)
}
// GrantBook grant client to read an existing book in the world state with provided parameters.
func (s *SmartContract) GrantBook(ctx contractapi.TransactionContextInterface, id string, client string, student string) error {
  exists, err := s.BookExists(ctx, id)
  if err != nil {
    return err
  }
  if !exists {
    return fmt.Errorf("the book %s does not exist", id)
  }

  book, err := s.GetBook(ctx, id)
  if err != nil {
    return err
  }
  caller,err := s.GetClientName(ctx)
  if err != nil {
    return err
  }
  if book.Owner != caller {
    return fmt.Errorf("the book %s is not owned by you", id)
  }
  book.EntitleList = append(book.EntitleList,client)

  studentHash,err := s.GetStudentHash(ctx,client,student)
  if err != nil {
    return err
  }
  fmt.Printf("student hash: %s",studentHash) 
  book.ReaderList = append(book.ReaderList,studentHash)
  
  bookJSON, err := json.Marshal(book)
  if err != nil {
    return err
  }

  return ctx.GetStub().PutState(id, bookJSON)
}
// UpdateBook updates an existing book in the world state with provided parameters.
func (s *SmartContract) UpdateBook(ctx contractapi.TransactionContextInterface, id string, title string, author string, isbn string, owner string) error {
  exists, err := s.BookExists(ctx, id)
  if err != nil {
    return err
  }
  if !exists {
    return fmt.Errorf("the book %s does not exist", id)
  }

  // overwriting original book with new book
  book := Book{
    ID:             id,
    Title:          title,
    Author:         author,
    ISBN:           isbn,
    Owner:          owner,
  }
  bookJSON, err := json.Marshal(book)
  if err != nil {
    return err
  }

  return ctx.GetStub().PutState(id, bookJSON)
}

// DeleteBook deletes an given book from the world state.
func (s *SmartContract) DeleteBook(ctx contractapi.TransactionContextInterface, id string) error {
  exists, err := s.BookExists(ctx, id)
  if err != nil {
    return err
  }
  if !exists {
    return fmt.Errorf("the book %s does not exist", id)
  }

  return ctx.GetStub().DelState(id)
}

// BookExists returns true when book with given ID exists in world state
func (s *SmartContract) BookExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
  bookJSON, err := ctx.GetStub().GetState(id)
  if err != nil {
    return false, fmt.Errorf("failed to read from world state: %v", err)
  }

  return bookJSON != nil, nil
}

// TransferBook updates the owner field of book with given id in world state.
func (s *SmartContract) TransferBook(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
  book, err := s.GetBook(ctx, id)
  if err != nil {
    return err
  }

  book.Owner = newOwner
  bookJSON, err := json.Marshal(book)
  if err != nil {
    return err
  }

  return ctx.GetStub().PutState(id, bookJSON)
}

// GetAllBooks returns all book found in world state
func (s *SmartContract) GetAllBooks(ctx contractapi.TransactionContextInterface) ([]*Book, error) {
  caller,err := s.GetClientName(ctx)
  if err != nil {
    return nil,err
  }
  // range query with empty string for startKey and endKey does an
  // open-ended query of all book in the chaincode namespace.
  resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
  if err != nil {
    return nil, err
  }
  defer resultsIterator.Close()

  var books []*Book
  for resultsIterator.HasNext() {
    queryResponse, err := resultsIterator.Next()
    if err != nil {
      return nil, err
    }
    if strings.Contains(queryResponse.Key, "book_") {
       var book Book
       err = json.Unmarshal(queryResponse.Value, &book)
       if err != nil {
         return nil, err
       }
       for _, v := range book.EntitleList {
	   if v == caller {
 	   	books = append(books, &book)
                break;
           }

        }
      }
   }

  return books, nil
}

func (s *SmartContract) GetPrivateCollection(ctx contractapi.TransactionContextInterface, caller string) (string, error) {
  re := regexp.MustCompile("Admin@org([0-9]).example.com")
  match := re.FindStringSubmatch(string(caller))
  fmt.Println(match)
  PrivateCollection := ""
  if len(match) >= 2 {
     PrivateCollection = "_implicit_org_Org"+match[1]+"MSP"
  } else {
     return "",fmt.Errorf("failed to parse org string")
  }
  return PrivateCollection,nil
}
// GetAllStudents returns all book found in world state
func (s *SmartContract) GetAllStudents(ctx contractapi.TransactionContextInterface) ([]*Student, error) {
  caller,err := s.GetClientName(ctx)
  if err != nil {
    return nil,err
  }
  PrivateCollection,err := s.GetPrivateCollection(ctx,caller)
  if err != nil {
    return nil,err
  }

  // range query with empty string for startKey and endKey does an
  // open-ended query of all book in the chaincode namespace.
  resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(PrivateCollection,"", "")
  if err != nil {
    return nil, err
  }
  defer resultsIterator.Close()

  var students []*Student
  for resultsIterator.HasNext() {
    queryResponse, err := resultsIterator.Next()
    if err != nil {
      return nil, err
    }
    if strings.Contains(queryResponse.Key, "student_") {
       var student Student
       err = json.Unmarshal(queryResponse.Value, &student)
       if err != nil {
         return nil, err
       }
       if student.Org == caller {
 	   	students = append(students, &student)
       }
    }
  }

  return students, nil
}
func main() {
  bookChaincode, err := contractapi.NewChaincode(&SmartContract{})
  if err != nil {
    log.Panicf("Error creating book-transfer-basic chaincode: %v", err)
  }

  if err := bookChaincode.Start(); err != nil {
    log.Panicf("Error starting book-transfer-basic chaincode: %v", err)
  }
}

```

---

## U12_local_model_chaincode.go

- Bytes: 4191 | Lines: 131

```go
package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// LocalModelSmartContract provides functions for managing local ML models
type LocalModelSmartContract struct {
	contractapi.Contract
}

// LocalModel describes a local model contribution
type LocalModel struct {
	Type                string `json:"objectType"`
	LocalModelHash      string `json:"local_model_hash"`
	NumExamples         uint64 `json:"num_examples"`
	RootGlobalModelHash string `json:"root_global_model_hash"`
	RunID               string `json:"run_id"`
	RoundID             uint64 `json:"round_id"`
}

// CreateLocalModel creates a new local model record
func (s *LocalModelSmartContract) CreateLocalModel(ctx contractapi.TransactionContextInterface, localModelHash string, numExamples uint64, rootGlobalModelHash string, runID string, roundID uint64) error {
	// Verify that only Org2 can submit local models
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client's MSPID: %v", err)
	}
	if clientMSPID != "Org2MSP" {
		return fmt.Errorf("only Org2 members can submit local models")
	}

	exists, err := s.LocalModelExists(ctx, localModelHash)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the local model %s already exists", localModelHash)
	}

	// Validate model data
	if len(localModelHash) == 0 {
		return fmt.Errorf("local model hash must not be empty")
	}
	if len(rootGlobalModelHash) == 0 {
		return fmt.Errorf("root global model hash must not be empty")
	}

	// Verify the referenced global model exists by querying the global model chaincode
	globalModelArgs := [][]byte{[]byte("ReadGlobalModel"), []byte(rootGlobalModelHash)}
	response := ctx.GetStub().InvokeChaincode("global_model_chaincode", globalModelArgs, "mychannel")
	if response.Status != 200 {
		return fmt.Errorf("failed to verify global model existence: %s", response.Message)
	}

	model := LocalModel{
		LocalModelHash:      localModelHash,
		NumExamples:         numExamples,
		RootGlobalModelHash: rootGlobalModelHash,
		RunID:               runID,
		RoundID:             roundID,
	}

	// Store the local model
	modelAsBytes, err := json.Marshal(model)
	if err != nil {
		return fmt.Errorf("failed to marshal model: %v", err)
	}

	return ctx.GetStub().PutState(model.LocalModelHash, modelAsBytes)
}

// ReadLocalModel returns the local model stored with given hash
func (s *LocalModelSmartContract) ReadLocalModel(ctx contractapi.TransactionContextInterface, localModelHash string) (*LocalModel, error) {
	modelAsBytes, err := ctx.GetStub().GetState(localModelHash)
	if err != nil {
		return nil, fmt.Errorf("failed to read local model: %v", err)
	}
	if modelAsBytes == nil {
		return nil, fmt.Errorf("local model %s does not exist", localModelHash)
	}

	var model LocalModel
	err = json.Unmarshal(modelAsBytes, &model)
	if err != nil {
		return nil, err
	}

	return &model, nil
}

// QueryLocalModelsByRound returns all local models for a specific run and round
func (s *LocalModelSmartContract) QueryLocalModelsByRound(ctx contractapi.TransactionContextInterface, runID string, roundID uint64) ([]*LocalModel, error) {
	queryString := fmt.Sprintf(`{"selector":{"objectType":"localModel","run_id":"%s","round_id":%d}}`, runID, roundID)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var models []*LocalModel
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var model LocalModel
		err = json.Unmarshal(queryResult.Value, &model)
		if err != nil {
			return nil, err
		}
		models = append(models, &model)
	}

	return models, nil
}

// LocalModelExists returns true when local model with given ID exists in world state
func (s *LocalModelSmartContract) LocalModelExists(ctx contractapi.TransactionContextInterface, localModelHash string) (bool, error) {
	modelJSON, err := ctx.GetStub().GetState(localModelHash)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return modelJSON != nil, nil
}

```

---

## U13_security_manager.go

- Bytes: 23938 | Lines: 658

```go
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// SecurityManagerChaincode 証券に関する関数を提供
type SecurityManagerChaincode struct{}

// SecurityStatus 証券の状態を表す
type SecurityStatus int

const (
	Undefined SecurityStatus = iota
	Finalized
	Issued
)

const SECURITY_PURCHASE_REGISTRATION_KEY = "securityPurchaseRegistration"
const SECURITY_BALANCE_KEY = "securityBalance"
const SECURITY_KEY = "security"
const PLATFORM_MSPID = "platform"

// Security 証券に関する情報を格納
type Security struct {
	Name   string         `json:"name"`         //名前
	Issuer string         `json:"issuer"`       //発行体
	Units  uint64         `json:"units,string"` //発行口数
	Price  uint64         `json:"price,string"` //総発行額
	Status SecurityStatus //証券のステータス
}

// RequestSecurity 証券登録時に使用する証券の情報を格納
type RequestSecurity struct {
	UUID         string   `json:"uuid"`         //uuid
	SecurityInfo Security `json:"securityInfo"` //証券情報
}

// MintSecurity 証券発行時に誰にいくら発行するかの情報を格納
type MintSecurity struct {
	SecurityID   string `json:"securityId"`    //証券ID
	Amount       uint64 `json:"amount,string"` //発行額
	Organization string `json:"organization"`  //組織
	Identity     string `json:"identity"`      //アイデンティティ
}

// RequestPurchaseReservation ...
type RequestPurchaseReservation struct {
	SecurityID   string `json:"securityId"` //証券ID
	Organization string `json:"organization"`
	InvestorID   string `json:"investorId"` //投資家ID
	Units        uint64 `json:"units,string"`
}

// RequestIssueSecurity ...
type RequestIssueSecurity struct {
	SecurityID           string `json:"securityId"` // 証券ID
	TargetOrganization   string `json:"targetOrganization"`
	ReceiverOrganization string `json:"receiverOrganization"` // マネーの受け手の組織
	ReceiverIdentity     string `json:"receiverIdentity"`     // マネーの受け手のアイデンティ
}

// PurchaseInfo 証券購入登録情報
type PurchaseInfo struct {
	Units         uint64 //移転額
	IsTransferred bool   //移転完了の有無
}

// RequestTransferSecurity 証券の移転時に利用する情報を格納
type RequestTransferSecurity struct {
	SecurityID           string `json:"securityId"`           //証券ID
	Amount               uint64 `json:"amount,string"`        //移転額
	SenderOrganization   string `json:"senderOrganization"`   //送り手の組織
	SenderIdentity       string `json:"senderIdentity"`       //送り手のアイデンティティ
	ReceiverOrganization string `json:"receiverOrganization"` //受け手の組織
	ReceiverIdentity     string `json:"receiverIdentity"`     //受け手のアイデンティ
}

// RequestUpdateMoney ...
type RequestUpdateMoney struct {
	Amount       uint64 `json:"amount,string"`        //移転額
	Organization string `json:"receiverOrganization"` //受け手の組織
	Identity     string `json:"receiverIdentity"`     //受け手のアイデンティティ
	IsSender     bool
}

// GetSecurityBalance ...
type GetSecurityBalance struct {
	SecurityID   string `json:"securityId"`   //証券ID
	Organization string `json:"organization"` //組織
	Identity     string `json:"identity"`     //アイデンティティ
}

const NUMBER_OF_ARGUMENTS_SETTING_TRANSIENT = 0
const NUMBER_OF_ARGUMENTS = 1

// Init ...
func (sm *SecurityManagerChaincode) Init(apiStub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

// Invoke ...
func (sm *SecurityManagerChaincode) Invoke(apiStub shim.ChaincodeStubInterface) sc.Response {
	// Retrieve the requested Smart Contract function and arguments
	function, args := apiStub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	switch function {
	case "getSecurity":
		return sm.getSecurity(apiStub, args)
	case "createSecurity":
		return sm.createSecurity(apiStub, args)
	case "issueSecurity":
		return sm.issueSecurity(apiStub, args)
	case "queryAllSecurities":
		return sm.queryAllSecurities(apiStub)
	case "reservePurchase":
		return sm.reservePurchase(apiStub, args)
	case "finalizeSecurity":
		return sm.finalizeSecurity(apiStub, args)
	case "transferSecurity":
		return sm.transferSecurity(apiStub, args)
	case "getBalance":
		return sm.getBalance(apiStub, args)
	default:
		return shim.Error("Invalid Chaincode function name.")
	}
}

// どんな証券を発行するか
// 引数: {"uuid":\"security1\",\"securityInfo\":{\"name\":\"OneMilionSecurity\",\"issuer\":\"layerx\",\"units\":\"100\",\"size\":\"10000\"}
// DBでは key:security-{証券ID} value: Security
func (sm *SecurityManagerChaincode) createSecurity(apiStub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != NUMBER_OF_ARGUMENTS {
		return shim.Error("Incorrect number of arguments. Expecting " + strconv.Itoa(NUMBER_OF_ARGUMENTS))
	}

	var securityRequest RequestSecurity
	var securityReqBody string = args[0]
	err := json.Unmarshal([]byte(securityReqBody), &securityRequest)
	if err != nil {
		return shim.Error("Failed unmarshal securityRequest: " + err.Error())
	}

	var securityUUID string = generateSecurityKey(SECURITY_KEY, securityRequest.UUID)
	var security Security = securityRequest.SecurityInfo
	securityAsBytes, err := json.Marshal(security)
	if err != nil {
		return shim.Error("Failed marshal security: " + err.Error())
	}

	if _, err := json.Marshal(securityUUID); err != nil {
		return shim.Error("Failed marshal securityUUID: " + err.Error())
	}

	apiStub.PutState(securityUUID, securityAsBytes)
	if err != nil {
		return shim.Error("Failed putState securityAsBytes: " + err.Error())
	}

	return shim.Success(securityAsBytes)
}

// どの証券を発行するか
// 引数: {\"security\":\"security1\"}
func (sm *SecurityManagerChaincode) finalizeSecurity(apiStub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != NUMBER_OF_ARGUMENTS {
		return shim.Error("Incorrect number of arguments. Expecting " + strconv.Itoa(NUMBER_OF_ARGUMENTS))
	}

	var securityKey string = generateSecurityKey(SECURITY_KEY, args[0])
	var security Security
	securityAsBytes, _ := apiStub.GetState(securityKey)
	err := json.Unmarshal(securityAsBytes, &security)
	if err != nil {
		return shim.Error("Failed unmarshal securityAsBytes: " + err.Error())
	}

	security.Status = Finalized
	securityAsBytes, err = json.Marshal(security)
	if err != nil {
		return shim.Error("Failed marshal security: " + err.Error())
	}

	if err := apiStub.PutState(securityKey, securityAsBytes); err != nil {
		return shim.Error("Failed putState securityAsBytes: " + err.Error())
	}

	return shim.Success(securityAsBytes)
}

// issueSecurity ...
func (sm *SecurityManagerChaincode) issueSecurity(apiStub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != NUMBER_OF_ARGUMENTS_SETTING_TRANSIENT {
		return shim.Error("Incorrect number of arguments. Expecting " + strconv.Itoa(NUMBER_OF_ARGUMENTS_SETTING_TRANSIENT))
	}

	transMap, err := apiStub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	if _, ok := transMap["issueSecurity"]; !ok {
		return shim.Error("issueSecurity must be a key in the transient map")
	}

	if len(transMap["issueSecurity"]) == 0 {
		return shim.Error("issueSecurity value in the transient map must be a non-empty JSON string")
	}

	// 実行者がplatformかどうかを確認
	ok, err := sm.isPlatform(apiStub)
	if err != nil {
		return shim.Error("Failed isPlatform" + err.Error())
	}

	if !ok {
		return shim.Error("Executor is not platform")
	}

	issueSecurityRequest := RequestIssueSecurity{}
	err = json.Unmarshal(transMap["issueSecurity"], &issueSecurityRequest)
	if err != nil {
		return shim.Error("Failed to Unmarshal: " + string(transMap["issueSecurity"]))
	}

	securityID := generateSecurityKey(SECURITY_KEY, issueSecurityRequest.SecurityID)
	// 指定した証券が存在するかどうかを確認
	securityAsBytes, err := apiStub.GetState(securityID)
	if err != nil {
		return shim.Error("Failed getState security: " + err.Error())
	}

	if securityAsBytes == nil {
		return shim.Error("Not found security")
	}

	// 取得したJSONを構造体に変換
	security := Security{}
	if err := json.Unmarshal(securityAsBytes, &security); err != nil {
		return shim.Error("Failed security unmarshal: " + err.Error())
	}

	// 証券がされているかを確認
	if security.Status != Finalized {
		return shim.Error("Security has not been finalized.")
	}

	startKey := generateSecurityKey(SECURITY_PURCHASE_REGISTRATION_KEY, securityID, issueSecurityRequest.TargetOrganization, "investor0")
	// 999までの制限をなくす
	endKey := generateSecurityKey(SECURITY_PURCHASE_REGISTRATION_KEY, securityID, issueSecurityRequest.TargetOrganization, "investor999")
	resultsIterator, err := apiStub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}

	defer resultsIterator.Close()

	var receiverTotalAmount uint64 = 0
	// loop開始
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		kSlice := strings.Split(queryResponse.Key, "-")
		investorOrganization := kSlice[3]
		investorIdentity := kSlice[4]

		// 投資家の証券購入情報をJSONから構造体に変換
		purchaseReservation := &PurchaseInfo{}
		if err := json.Unmarshal(queryResponse.Value, &purchaseReservation); err != nil {
			return shim.Error("Failed purchaseReservation unmarshal: " + err.Error())
		}

		// 証券購入済の場合はcontinue
		if purchaseReservation.IsTransferred {
			continue
		}

		// 投資家のMoney残高更新
		amount := purchaseReservation.Units * security.Price
		r := RequestUpdateMoney{
			Amount:       amount,
			Identity:     investorIdentity,
			Organization: investorOrganization,
			IsSender:     true,
		}
		if _, err := sm.callUpdateMoney(apiStub, r); err != nil {
			return shim.Error("Failed callUpdateMoney: " + err.Error())
		}

		// loop処理が終わったら発行体の残高を更新する
		receiverTotalAmount += amount

		// mintSecurity
		mintSecurity := MintSecurity{
			SecurityID:   securityID, // issueSecurityRequest.SecurityID,
			Amount:       purchaseReservation.Units,
			Organization: investorOrganization,
			Identity:     investorIdentity,
		}
		mintSecurityJSON, err := json.Marshal(mintSecurity)
		if err != nil {
			return shim.Error("Failed mintSecurity json marshal" + err.Error())
		}

		mintSecurityArg := []string{string(mintSecurityJSON)}
		// mintSecurityの結果は不要
		if _, err := sm.mintSecurity(apiStub, mintSecurityArg); err != nil {
			return shim.Error("Failed mintSecurity: " + err.Error())
		}

		// 購入済にステータスを変更
		purchaseReservation.IsTransferred = true

		// 証券購入情報を更新
		investorAtBytes, err := json.Marshal(purchaseReservation)
		if err != nil {
			return shim.Error("Failed purchaseReservation json.Marshal: " + err.Error())
		}

		apiStub.PutState(queryResponse.Key, investorAtBytes)
	}

	// 発行体の口座を更新
	r := RequestUpdateMoney{
		Amount:       receiverTotalAmount,
		Identity:     issueSecurityRequest.ReceiverIdentity,
		Organization: issueSecurityRequest.ReceiverOrganization,
		IsSender:     false, // 受け手は発行体なのでfalse
	}
	if _, err := sm.callUpdateMoney(apiStub, r); err != nil {
		return shim.Error("Failed callUpdateMoney: " + err.Error())
	}

	return shim.Success(nil)
}

// どの証券を誰にいくらmintするか
// 引数: {\"amount\":\"100\",\"organization\":\"MinatoBank\",\"identity\":\"investor01\"}
// DBでは key:securityBalance-security-{証券ID}-{組織ID}-{投資家ID}, value:残高
// 例） key:securityBalance-security-security1-MinatoBank-investor01 value:200
func (sm *SecurityManagerChaincode) mintSecurity(apiStub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) != NUMBER_OF_ARGUMENTS {
		return nil, errors.New("Incorrect number of arguments. Expecting " + strconv.Itoa(NUMBER_OF_ARGUMENTS))
	}

	var mintSecurity MintSecurity
	var identity string = args[0]
	err := json.Unmarshal([]byte(identity), &mintSecurity)
	if err != nil {
		return nil, errors.New("Failed unmarshal identity: " + err.Error())
	}

	balanceAsBytes := []byte(strconv.FormatUint(mintSecurity.Amount, 10))
	var key string = generateSecurityKey(SECURITY_BALANCE_KEY, mintSecurity.SecurityID, mintSecurity.Organization, mintSecurity.Identity)
	var collectionName string = fmt.Sprintf("%sSecurityBalance", mintSecurity.Organization)
	err = apiStub.PutPrivateData(collectionName, key, balanceAsBytes)
	if err != nil {
		return nil, errors.New("Failed PutPrivateData balance: " + err.Error())
	}

	return balanceAsBytes, nil
}

// TODO　内部からのみ呼び出し可能にする
// 誰から誰にいくらmintする
// 引数: {"amount":"100","senderOrganization":"MinatoBank","senderIdentity":"investor01","receiverOrganization":"MinatoBank","receiverIdentity":"investor01"}
func (sm *SecurityManagerChaincode) transferSecurity(apiStub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != NUMBER_OF_ARGUMENTS_SETTING_TRANSIENT {
		return shim.Error("Incorrect number of arguments. Expecting " + strconv.Itoa(NUMBER_OF_ARGUMENTS_SETTING_TRANSIENT))
	}

	transMap, err := apiStub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	if _, ok := transMap["transferSecurity"]; !ok {
		return shim.Error("transferSecurity must be a key in the transient map")
	}

	if len(transMap["transferSecurity"]) == 0 {
		return shim.Error("transferSecurity value in the transient map must be a non-empty JSON string")
	}

	var transferSecurity RequestTransferSecurity
	err = json.Unmarshal(transMap["transferSecurity"], &transferSecurity)
	if err != nil {
		return shim.Error("Failed to Unmarshal: " + string(transMap["transferSecurity"]))
	}

	// senderの残高を減らす
	//// Securityの情報取得
	var senderBalance uint64
	var senderBalanceKey string = generateSecurityKey(SECURITY_BALANCE_KEY, transferSecurity.SecurityID, transferSecurity.SenderOrganization, transferSecurity.SenderIdentity)
	var senderCollectionName string = fmt.Sprintf("%sSecurityBalance", transferSecurity.SenderOrganization)
	senderSecurityBalanceAsBytes, err := apiStub.GetPrivateData(senderCollectionName, senderBalanceKey)
	if err != nil {
		return shim.Error("Failed GetPrivateData senderSecurityBalance: " + err.Error())
	}
	err = json.Unmarshal(senderSecurityBalanceAsBytes, &senderBalance)
	if err != nil {
		return shim.Error("Failed unmarshal senderSecurityBalanceAsBytes: " + err.Error())
	}
	//// senderの残高が足りているかバリデーション
	if senderBalance < transferSecurity.Amount {
		return shim.Error("Unsufficient amount of balance")
	}
	senderBalance -= transferSecurity.Amount
	senderSecurityBalanceAsBytes = []byte(strconv.FormatUint(senderBalance, 10))
	err = apiStub.PutPrivateData(senderCollectionName, senderBalanceKey, senderSecurityBalanceAsBytes)
	if err != nil {
		return shim.Error("Failed PutPrivateData senderSecurityBalance: " + err.Error())
	}

	// receiverの残高を増やす
	var receiverBalance uint64
	var receiverBalanceKey string = generateSecurityKey(SECURITY_BALANCE_KEY, transferSecurity.SecurityID, transferSecurity.ReceiverOrganization, transferSecurity.ReceiverIdentity)
	var receiverCollectionName string = fmt.Sprintf("%sSecurityBalance", transferSecurity.ReceiverOrganization)
	receiverSecurityBalanceAsBytes, err := apiStub.GetPrivateData(receiverCollectionName, receiverBalanceKey)
	if err != nil {
		return shim.Error("Failed GetPrivateData receiverSecurityBalance: " + err.Error())
	}

	if len(receiverSecurityBalanceAsBytes) == 0 {
		k := generateSecurityKey(SECURITY_BALANCE_KEY, transferSecurity.SecurityID, transferSecurity.ReceiverOrganization, transferSecurity.ReceiverIdentity)
		err = apiStub.PutPrivateData(receiverCollectionName, k, []byte("0"))
		if err != nil {
			return shim.Error("Failed PutPrivateData balance: " + err.Error())
		}
		receiverSecurityBalanceAsBytes = []byte("0")
	}

	err = json.Unmarshal(receiverSecurityBalanceAsBytes, &receiverBalance)
	if err != nil {
		return shim.Error("Failed unmarshal receiverSecurityBalanceAsBytes: " + err.Error())
	}
	receiverBalance += transferSecurity.Amount
	receiverSecurityBalanceAsBytes = []byte(strconv.FormatUint(receiverBalance, 10))
	apiStub.PutPrivateData(receiverCollectionName, receiverBalanceKey, receiverSecurityBalanceAsBytes)
	if err != nil {
		return shim.Error("Failed PutPrivateData receiverSecurityBalance: " + err.Error())
	}

	return shim.Success(receiverSecurityBalanceAsBytes)
}

func (sm *SecurityManagerChaincode) getBalance(apiStub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != NUMBER_OF_ARGUMENTS_SETTING_TRANSIENT {
		return shim.Error("Incorrect number of arguments. Expecting " + strconv.Itoa(NUMBER_OF_ARGUMENTS_SETTING_TRANSIENT))
	}

	transMap, err := apiStub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	if _, ok := transMap["getSecurityBalance"]; !ok {
		return shim.Error("getSecurityBalance must be a key in the transient map")
	}

	if len(transMap["getSecurityBalance"]) == 0 {
		return shim.Error("getSecurityBalance value in the transient map must be a non-empty JSON string")
	}

	var getSecurityBalance GetSecurityBalance
	err = json.Unmarshal(transMap["getSecurityBalance"], &getSecurityBalance)
	if err != nil {
		return shim.Error("Failed to Unmarshal: " + string(transMap["getSecurityBalance"]))
	}

	// Securityの情報取得
	var key string = generateSecurityKey(SECURITY_BALANCE_KEY, SECURITY_KEY, getSecurityBalance.SecurityID, getSecurityBalance.Organization, getSecurityBalance.Identity)
	var collectionName string = fmt.Sprintf("%sSecurityBalance", getSecurityBalance.Organization)
	securityBalanceAsBytes, err := apiStub.GetPrivateData(collectionName, key)
	if err != nil {
		return shim.Error("Failed GetPrivateData securityBalance: " + err.Error())
	}

	return shim.Success(securityBalanceAsBytes)
}

func (sm *SecurityManagerChaincode) getSecurity(apiStub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	var securityKey string = args[0]
	// Securityの情報取得だけに絞りたい
	securityAsBytes, err := apiStub.GetState(securityKey)
	if err != nil {
		return shim.Error("Failed getState security: " + err.Error())
	}
	return shim.Success(securityAsBytes)
}

// 証券の一覧取得
// 引数: なし
func (sm *SecurityManagerChaincode) queryAllSecurities(apiStub shim.ChaincodeStubInterface) sc.Response {
	startKey := "security-security0"
	// 999までの制限をなくす
	endKey := "security-security999"

	resultsIterator, err := apiStub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllSecurities:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// reservePurchase
// 証券購入量登録
// 引数: {"securityId":"security1","investorId":"investor01","units":"100"}
func (sm *SecurityManagerChaincode) reservePurchase(apiStub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != NUMBER_OF_ARGUMENTS {
		return shim.Error("Incorrect number of arguments. Expecting " + strconv.Itoa(NUMBER_OF_ARGUMENTS))
	}

	var req RequestPurchaseReservation
	if err := json.Unmarshal([]byte(args[0]), &req); err != nil {
		return shim.Error("Failed reservePurchase unmarshal: " + err.Error())
	}

	purchaseInfo := PurchaseInfo{}
	purchaseInfo.Units = req.Units
	purchaseInfoAtBytes, err := json.Marshal(purchaseInfo)
	if err != nil {
		return shim.Error("Failed purchaseInfo json.Marshal" + err.Error())
	}

	var key string = generateSecurityKey(SECURITY_PURCHASE_REGISTRATION_KEY, SECURITY_KEY, req.SecurityID, req.Organization, req.InvestorID)
	// 登録
	apiStub.PutState(key, purchaseInfoAtBytes)

	return shim.Success(purchaseInfoAtBytes)
}

func (sm *SecurityManagerChaincode) isPlatform(apiStub shim.ChaincodeStubInterface) (bool, error) {
	creator, err := apiStub.GetCreator()
	if err != nil {
		return false, errors.New("Failed GetCreator" + err.Error())
	}
	fmt.Println("### c: ", *(*string)(unsafe.Pointer(&creator)))

	/*
		issueSecurityを実行し、GetCreator()で確認すると下記のデータが取れる
		証明書の先頭のところにplatformという文字列が入っている

		platform-----BEGIN CERTIFICATE-----
		MIICsjCCAligAwIBAgIUDzzQ9EC8Dfe5fmIwTK1QeHfgsIIwCgYIKoZIzj0EAwIw
		aDELMAkGA1UEBhMCVVMxFzAVBgNVBAgTDk5vcnRoIENhcm9saW5hMRQwEgYDVQQK
		EwtIeXBlcmxlZGdlcjEPMA0GA1UECxMGRmFicmljMRkwFwYDVQQDExBmYWJyaWMt
		Y2Etc2VydmVyMB4XDTIwMDMzMTExNDUwMFoXDTIxMDMzMTExNTAwMFowaDELMAkG
		A1UEBhMCVVMxFzAVBgNVBAgTDk5vcnRoIENhcm9saW5hMRQwEgYDVQQKEwtIeXBl
		cmxlZGdlcjEPMA0GA1UECxMGY2xpZW50MRkwFwYDVQQDDBBiYW5rX3BlZXJfYWRt
		aW4xMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE3Z3+yp8QXjZOfzx+1W0G4qBb
		5a5g3/XyKQacZKWedYmoA7jpxfSxT6E0U6F435a5QzB+JHzG59K2jtrnv58M0KOB
		3zCB3DAOBgNVHQ8BAf8EBAMCB4AwDAYDVR0TAQH/BAIwADAdBgNVHQ4EFgQUYobB
		YQQt+YV1hlJ/TxlRFW4K1MAwHwYDVR0jBBgwFoAU4s0bV8ZTZIca746qYIIMi2ci
		NG0wFwYDVR0RBBAwDoIMNmEyNDMzZTNiNDAxMGMGCCoDBAUGBwgBBFd7ImF0dHJz
		Ijp7ImhmLkFmZmlsaWF0aW9uIjoiIiwiaGYuRW5yb2xsbWVudElEIjoiYmFua19w
		ZWVyX2FkbWluMSIsImhmLlR5cGUiOiJjbGllbnQifX0wCgYIKoZIzj0EAwIDSAAw
		RQIhAP2rys2jKkT072CNkn2LDP/iCkegtM3BwKGKkg1rgSTRAiAvucsNE97RRm+O
		7ItIiNyfi1i/QF6ajp4ktGVQ3lRkYg==
		-----END CERTIFICATE-----
	*/
	//
	if !strings.Contains(*(*string)(unsafe.Pointer(&creator)), PLATFORM_MSPID) {
		return false, nil
	}

	return true, nil
}

func (sm *SecurityManagerChaincode) callUpdateMoney(apiStub shim.ChaincodeStubInterface, req RequestUpdateMoney) ([]byte, error) {
	updateMoneyJSON, err := json.Marshal(req)
	if err != nil {
		return nil, errors.New("Failed callUpdateMoney json marshal")
	}

	chaincodeName := "money"
	channelName := "all-ch"
	funcName := "updateMoney"
	invokeArgs := toChaincodeArgs(funcName, string(updateMoneyJSON))
	response := apiStub.InvokeChaincode(chaincodeName, invokeArgs, channelName)
	if response.Status != shim.OK {
		errStr := fmt.Sprintf("Failed to invoke chaincode. Got error: %s", string(response.Payload))
		fmt.Printf(errStr)

		return nil, errors.New(errStr)
	}

	return response.Payload, nil
}

func toChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

func generateSecurityKey(args ...string) string {
	key := strings.Join(args, "-")

	return key
}

func main() {
	// Create a new Smart Contract
	err := shim.Start(new(SecurityManagerChaincode))
	if err != nil {
		fmt.Printf("Error creating new Chaincode: %s", err)
	}
}

```

---

## U14_smartaicc.go

- Bytes: 9474 | Lines: 383

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// Definition des Actifs partagés dans le réseau
// Actif capteur
type Bloc_Capteur struct {
	ID_bc      string  `json:"ID_bc"`
	ID_pbc     string  `json:"ID_pbc"`
	Culture_bc int     `json:"Culture_bc"`
	HS         int     `json:"HS"`
	TS         int     `json:"TS"`
	TA         int     `json:"TA"`
	HR         int     `json:"HR"`
	PP         int     `json:"PP"`
	QP         float32 `json:"QP"`
}

// Actif actionneur
type Bloc_Actionneur struct {
	ID_ba      string `json:"ID_ba"`
	ID_pba     string `json:"ID_pba"`
	Culture_ba int    `json:"Culture_ba"`
	Etat       int    `json:"Etat"`
}
type Proprietaire struct {
	ID_p  string `json:"ID_p"`
	PWD_p string `json:"PWD_p"`
}

// Fonction pour l'actif capteur

func (s *SmartContract) InitBlockSensorAsset(ctx contractapi.TransactionContextInterface) error {

	// Actif de base bc
	assets := []Bloc_Capteur{
		{ID_bc: "Bloc_capteur_0", ID_pbc: "Admin_1234", Culture_bc: 1, HS: 0, TS: 0, TA: 0, HR: 0, PP: 0, QP: 0.0},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID_bc, assetJSON)
		if err != nil {
			return fmt.Errorf("impossible de mettre en place world state. %v", err)
		}
	}

	return nil
}
func (s *SmartContract) CreateBlockSensorAsset(ctx contractapi.TransactionContextInterface, idbc string, idpbc string, cbc int, hs int, ts int, ta int, hr int, pp int, qp float32) error {
	exists, err := s.AssetExists(ctx, idbc)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("le bloc capteur %s est existant", idbc)
	}

	asset := Bloc_Capteur{
		ID_bc:      idbc,
		ID_pbc:     idpbc,
		Culture_bc: cbc,
		HS:         hs,
		TS:         ts,
		TA:         ta,
		HR:         hr,
		PP:         pp,
		QP:         qp,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(idbc, assetJSON)
}
func (s *SmartContract) UpdateBlockSensorAsset(ctx contractapi.TransactionContextInterface, idbc string, idpbc string, cbc int, hs int, ts int, ta int, hr int, pp int, qp float32) error {

	asset := Bloc_Capteur{
		ID_bc:      idbc,
		ID_pbc:     idpbc,
		Culture_bc: cbc,
		HS:         hs,
		TS:         ts,
		TA:         ta,
		HR:         hr,
		PP:         pp,
		QP:         qp,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(idbc, assetJSON)
}
func (s *SmartContract) DeleteBlockSensorAsset(ctx contractapi.TransactionContextInterface, idbc string) error {
	exists, err := s.AssetExists(ctx, idbc)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("le bloc capteur %s est exexistant", idbc)
	}

	return ctx.GetStub().DelState(idbc)
}
func (s *SmartContract) GetAllBlockSensorAssets(ctx contractapi.TransactionContextInterface) ([]*Bloc_Capteur, error) {

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Bloc_Capteur
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Bloc_Capteur
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
func (s *SmartContract) ReadBlocksensorAsset(ctx contractapi.TransactionContextInterface, id string) (*Bloc_Capteur, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("ereur de lecture dans world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf(" %s est Bloc capteur non existant", id)
	}

	var asset Bloc_Capteur
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// Fonction pour l'actif actionneur
func (s *SmartContract) InitBlockActuatorAsset(ctx contractapi.TransactionContextInterface) error {

	// Actif de base ba
	assets := []Bloc_Actionneur{
		{ID_ba: "Bloc_actioneur_0", ID_pba: "Admin_1234", Culture_ba: 1, Etat: 0},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID_ba, assetJSON)
		if err != nil {
			return fmt.Errorf("impossible de mettre en place world state. %v", err)
		}
	}

	return nil
}
func (s *SmartContract) CreateBlockActuatorAsset(ctx contractapi.TransactionContextInterface, idba string, idpba string, cba int, etat int) error {
	exists, err := s.AssetExists(ctx, idba)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("le bloc actionneur %s est existant", idba)
	}

	asset := Bloc_Actionneur{
		ID_ba:      idba,
		ID_pba:     idpba,
		Culture_ba: cba,
		Etat:       etat,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(idba, assetJSON)
}
func (s *SmartContract) UpdateBlockActuatorAsset(ctx contractapi.TransactionContextInterface, idba string, idpba string, cba int, etat int) error {
	exists, err := s.AssetExists(ctx, idba)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf(" %s Est un Bloc actionneur non existant", idba)
	}

	asset := Bloc_Actionneur{
		ID_ba:      idba,
		ID_pba:     idpba,
		Culture_ba: cba,
		Etat:       etat,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(idba, assetJSON)
}
func (s *SmartContract) ActuatorState(ctx contractapi.TransactionContextInterface, idba string) (*Bloc_Actionneur, error) {
	assetJSON, err := ctx.GetStub().GetState(idba)
	if err != nil {
		return nil, fmt.Errorf("echec de la lecture de world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf(" %s Est un utilisateur inconnu", idba)
	}

	var asset Bloc_Actionneur
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}
func (s *SmartContract) DeleteBlockActuatorAsset(ctx contractapi.TransactionContextInterface, idba string) error {
	exists, err := s.AssetExists(ctx, idba)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf(" %s est un bloc actionneur non existant", idba)
	}

	return ctx.GetStub().DelState(idba)
}
func (s *SmartContract) GetAllBlockactuatorAssets(ctx contractapi.TransactionContextInterface) ([]*Bloc_Actionneur, error) {

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Bloc_Actionneur
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Bloc_Actionneur
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

func (s *SmartContract) ReadBlockactuatorAsset(ctx contractapi.TransactionContextInterface, id string) (*Bloc_Actionneur, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("ereur de lecture dans world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf(" %s est Bloc capteur non existant", id)
	}

	var asset Bloc_Actionneur
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// Fonction pour gestion utilisateur
func (s *SmartContract) InitLedgerOwnerAsset(ctx contractapi.TransactionContextInterface) error {

	// Actif de base ba
	assets := []Proprietaire{
		{ID_p: "Admin_1234", PWD_p: "1234"},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID_p, assetJSON)
		if err != nil {
			return fmt.Errorf("impossible de mettre en place world state. %v", err)
		}
	}

	return nil
}
func (s *SmartContract) OwnerInscription(ctx contractapi.TransactionContextInterface, idp string, pwd string) error {

	exists, err := s.AssetExists(ctx, idp)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf(" %s est un utilisateur existant", idp)
	}
	asset := Proprietaire{
		ID_p:  idp,
		PWD_p: pwd,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(idp, assetJSON)
}
func (s *SmartContract) OwnerConnexion(ctx contractapi.TransactionContextInterface, id string) (*Proprietaire, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("echec de la lecture de world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf(" %s est un utilisateur inconnu", id)
	}

	var asset Proprietaire
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// Autres fonctions
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// Fonction Principale
func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf(": %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
	}
}

```

---

## U17_realty_chaincode.go

- Bytes: 15525 | Lines: 548

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
)

// SmartContract 提供房地产交易的功能
type SmartContract struct {
	contractapi.Contract
}

// 文档类型常量（用于创建复合键）
const (
	REAL_ESTATE = "RE" // 房产信息
	TRANSACTION = "TX" // 交易信息
)

// RealEstateStatus 房产状态
type RealEstateStatus string

const (
	NORMAL         RealEstateStatus = "NORMAL"         // 正常
	IN_TRANSACTION RealEstateStatus = "IN_TRANSACTION" // 交易中
)

// TransactionStatus 交易状态
type TransactionStatus string

const (
	PENDING   TransactionStatus = "PENDING"   // 待付款
	COMPLETED TransactionStatus = "COMPLETED" // 已完成
)

// RealEstate 房产信息
type RealEstate struct {
	ID              string           `json:"id"`              // 房产ID
	PropertyAddress string           `json:"propertyAddress"` // 房产地址
	Area            float64          `json:"area"`            // 面积
	CurrentOwner    string           `json:"currentOwner"`    // 当前所有者
	Status          RealEstateStatus `json:"status"`          // 状态
	CreateTime      time.Time        `json:"createTime"`      // 创建时间
	UpdateTime      time.Time        `json:"updateTime"`      // 更新时间
}

// Transaction 交易信息
type Transaction struct {
	ID           string            `json:"id"`           // 交易ID
	RealEstateID string            `json:"realEstateId"` // 房产ID
	Seller       string            `json:"seller"`       // 卖家
	Buyer        string            `json:"buyer"`        // 买家
	Price        float64           `json:"price"`        // 成交价格
	Status       TransactionStatus `json:"status"`       // 状态
	CreateTime   time.Time         `json:"createTime"`   // 创建时间
	UpdateTime   time.Time         `json:"updateTime"`   // 更新时间
}

// QueryResult 分页查询结果
type QueryResult struct {
	Records             []interface{} `json:"records"`             // 记录列表
	RecordsCount        int32         `json:"recordsCount"`        // 本次返回的记录数
	Bookmark            string        `json:"bookmark"`            // 书签，用于下一页查询
	FetchedRecordsCount int32         `json:"fetchedRecordsCount"` // 总共获取的记录数
}

// 组织 MSP ID 常量
const (
	REALTY_ORG_MSPID = "Org1MSP" // 不动产登记机构组织 MSP ID
	BANK_ORG_MSPID   = "Org2MSP" // 银行组织 MSP ID
	TRADE_ORG_MSPID  = "Org3MSP" // 交易平台组织 MSP ID
)

// 通用方法: 获取客户端身份信息
func (s *SmartContract) getClientIdentityMSPID(ctx contractapi.TransactionContextInterface) (string, error) {
	clientID, err := cid.New(ctx.GetStub())
	if err != nil {
		return "", fmt.Errorf("获取客户端身份信息失败：%v", err)
	}
	return clientID.GetMSPID()
}

// 通用方法：创建和获取复合键
func (s *SmartContract) getCompositeKey(ctx contractapi.TransactionContextInterface, objectType string, attributes []string) (string, error) {
	key, err := ctx.GetStub().CreateCompositeKey(objectType, attributes)
	if err != nil {
		return "", fmt.Errorf("创建复合键失败：%v", err)
	}
	return key, nil
}

// 通用方法：获取状态
func (s *SmartContract) getState(ctx contractapi.TransactionContextInterface, key string, value interface{}) error {
	bytes, err := ctx.GetStub().GetState(key)
	if err != nil {
		return fmt.Errorf("读取状态失败：%v", err)
	}
	if bytes == nil {
		return fmt.Errorf("键 %s 不存在", key)
	}

	err = json.Unmarshal(bytes, value)
	if err != nil {
		return fmt.Errorf("解析数据失败：%v", err)
	}
	return nil
}

// 通用方法：保存状态
func (s *SmartContract) putState(ctx contractapi.TransactionContextInterface, key string, value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化数据失败：%v", err)
	}

	err = ctx.GetStub().PutState(key, bytes)
	if err != nil {
		return fmt.Errorf("保存状态失败：%v", err)
	}
	return nil
}

// CreateRealEstate 创建房产信息（仅不动产登记机构组织可以调用）
func (s *SmartContract) CreateRealEstate(ctx contractapi.TransactionContextInterface, id string, address string, area float64, owner string, createTime time.Time) error {
	// 检查调用者身份
	clientMSPID, err := s.getClientIdentityMSPID(ctx)
	if err != nil {
		return fmt.Errorf("获取调用者身份失败：%v", err)
	}

	// 验证是否是不动产登记机构组织的成员
	if clientMSPID != REALTY_ORG_MSPID {
		return fmt.Errorf("只有不动产登记机构组织成员才能创建房产信息")
	}

	// 参数验证
	if len(id) == 0 {
		return fmt.Errorf("房产ID不能为空")
	}
	if len(address) == 0 {
		return fmt.Errorf("房产地址不能为空")
	}
	if area <= 0 {
		return fmt.Errorf("面积必须大于0")
	}
	if len(owner) == 0 {
		return fmt.Errorf("所有者不能为空")
	}

	// 检查房产是否已存在（检查所有可能的状态）
	for _, status := range []RealEstateStatus{NORMAL, IN_TRANSACTION} {
		key, err := s.getCompositeKey(ctx, REAL_ESTATE, []string{string(status), id})
		if err != nil {
			return fmt.Errorf("创建复合键失败：%v", err)
		}

		exists, err := ctx.GetStub().GetState(key)
		if err != nil {
			return fmt.Errorf("查询房产信息失败：%v", err)
		}
		if exists != nil {
			return fmt.Errorf("房产ID %s 已存在", id)
		}
	}

	// 创建房产信息
	realEstate := RealEstate{
		ID:              id,
		PropertyAddress: address,
		Area:            area,
		CurrentOwner:    owner,
		Status:          NORMAL,
		CreateTime:      createTime,
		UpdateTime:      createTime,
	}

	// 保存房产信息（复合键：类型_状态_ID）
	key, err := s.getCompositeKey(ctx, REAL_ESTATE, []string{string(NORMAL), id})
	if err != nil {
		return err
	}

	err = s.putState(ctx, key, realEstate)
	if err != nil {
		return err
	}

	return nil
}

// CreateTransaction 生成交易（仅交易平台组织可以调用）
func (s *SmartContract) CreateTransaction(ctx contractapi.TransactionContextInterface, txID string, realEstateID string, seller string, buyer string, price float64, createTime time.Time) error {
	// 检查调用者身份
	clientMSPID, err := s.getClientIdentityMSPID(ctx)
	if err != nil {
		return fmt.Errorf("获取调用者身份失败：%v", err)
	}

	// 验证是否是交易平台组织的成员
	if clientMSPID != TRADE_ORG_MSPID {
		return fmt.Errorf("只有交易平台组织成员才能生成交易")
	}

	// 参数验证
	if len(txID) == 0 {
		return fmt.Errorf("交易ID不能为空")
	}
	if len(realEstateID) == 0 {
		return fmt.Errorf("房产ID不能为空")
	}
	if len(seller) == 0 {
		return fmt.Errorf("卖家不能为空")
	}
	if len(buyer) == 0 {
		return fmt.Errorf("买家不能为空")
	}
	if seller == buyer {
		return fmt.Errorf("买家和卖家不能是同一人")
	}
	if price <= 0 {
		return fmt.Errorf("价格必须大于0")
	}

	// 查询房产信息
	realEstateKey, err := s.getCompositeKey(ctx, REAL_ESTATE, []string{string(NORMAL), realEstateID})
	if err != nil {
		return err
	}

	var realEstate RealEstate
	err = s.getState(ctx, realEstateKey, &realEstate)
	if err != nil {
		return err
	}

	// 检查卖家是否是房产所有者
	if realEstate.CurrentOwner != seller {
		return fmt.Errorf("卖家不是房产所有者")
	}

	// 生成交易信息
	transaction := Transaction{
		ID:           txID,
		RealEstateID: realEstateID,
		Seller:       seller,
		Buyer:        buyer,
		Price:        price,
		Status:       PENDING,
		CreateTime:   createTime,
		UpdateTime:   createTime,
	}

	// 更新房产状态
	realEstate.Status = IN_TRANSACTION
	realEstate.UpdateTime = createTime

	// 保存状态
	txKey, err := s.getCompositeKey(ctx, TRANSACTION, []string{string(PENDING), txID})
	if err != nil {
		return err
	}

	// 删除旧的房产记录
	err = ctx.GetStub().DelState(realEstateKey)
	if err != nil {
		return fmt.Errorf("删除旧的房产记录失败：%v", err)
	}

	// 创建新的房产记录（使用新状态）
	newRealEstateKey, err := s.getCompositeKey(ctx, REAL_ESTATE, []string{string(IN_TRANSACTION), realEstateID})
	if err != nil {
		return err
	}

	err = s.putState(ctx, txKey, transaction)
	if err != nil {
		return err
	}

	err = s.putState(ctx, newRealEstateKey, realEstate)
	if err != nil {
		return err
	}

	return nil
}

// CompleteTransaction 完成交易（仅银行组织可以调用）
func (s *SmartContract) CompleteTransaction(ctx contractapi.TransactionContextInterface, txID string, updateTime time.Time) error {
	// 检查调用者身份
	clientMSPID, err := s.getClientIdentityMSPID(ctx)
	if err != nil {
		return fmt.Errorf("获取调用者身份失败：%v", err)
	}

	// 验证是否是银行组织的成员
	if clientMSPID != BANK_ORG_MSPID {
		return fmt.Errorf("只有银行组织成员才能完成交易")
	}

	// 查询交易信息
	txKey, err := s.getCompositeKey(ctx, TRANSACTION, []string{string(PENDING), txID})
	if err != nil {
		return err
	}

	var transaction Transaction
	err = s.getState(ctx, txKey, &transaction)
	if err != nil {
		return err
	}

	// 查询房产信息
	realEstateKey, err := s.getCompositeKey(ctx, REAL_ESTATE, []string{string(IN_TRANSACTION), transaction.RealEstateID})
	if err != nil {
		return err
	}

	var realEstate RealEstate
	err = s.getState(ctx, realEstateKey, &realEstate)
	if err != nil {
		return err
	}

	// 更新状态
	realEstate.CurrentOwner = transaction.Buyer
	realEstate.Status = NORMAL
	realEstate.UpdateTime = updateTime

	transaction.Status = COMPLETED
	transaction.UpdateTime = updateTime

	// 删除旧记录
	err = ctx.GetStub().DelState(txKey)
	if err != nil {
		return fmt.Errorf("删除旧的交易记录失败：%v", err)
	}

	err = ctx.GetStub().DelState(realEstateKey)
	if err != nil {
		return fmt.Errorf("删除旧的房产记录失败：%v", err)
	}

	// 创建新记录
	newTxKey, err := s.getCompositeKey(ctx, TRANSACTION, []string{string(COMPLETED), txID})
	if err != nil {
		return err
	}

	newRealEstateKey, err := s.getCompositeKey(ctx, REAL_ESTATE, []string{string(NORMAL), transaction.RealEstateID})
	if err != nil {
		return err
	}

	err = s.putState(ctx, newTxKey, transaction)
	if err != nil {
		return err
	}

	err = s.putState(ctx, newRealEstateKey, realEstate)
	if err != nil {
		return err
	}

	return nil
}

// QueryRealEstate 查询房产信息
func (s *SmartContract) QueryRealEstate(ctx contractapi.TransactionContextInterface, id string) (*RealEstate, error) {
	// 遍历所有可能的状态查询房产
	for _, status := range []RealEstateStatus{NORMAL, IN_TRANSACTION} {
		key, err := s.getCompositeKey(ctx, REAL_ESTATE, []string{string(status), id})
		if err != nil {
			return nil, fmt.Errorf("创建复合键失败：%v", err)
		}

		bytes, err := ctx.GetStub().GetState(key)
		if err != nil {
			return nil, fmt.Errorf("查询房产信息失败：%v", err)
		}
		if bytes != nil {
			var realEstate RealEstate
			err = json.Unmarshal(bytes, &realEstate)
			if err != nil {
				return nil, fmt.Errorf("解析房产信息失败：%v", err)
			}
			return &realEstate, nil
		}
	}

	return nil, fmt.Errorf("房产ID %s 不存在", id)
}

// QueryTransaction 查询交易信息
func (s *SmartContract) QueryTransaction(ctx contractapi.TransactionContextInterface, txID string) (*Transaction, error) {
	// 遍历所有可能的状态查询交易
	for _, status := range []TransactionStatus{PENDING, COMPLETED} {
		key, err := s.getCompositeKey(ctx, TRANSACTION, []string{string(status), txID})
		if err != nil {
			return nil, fmt.Errorf("创建复合键失败：%v", err)
		}

		bytes, err := ctx.GetStub().GetState(key)
		if err != nil {
			return nil, fmt.Errorf("查询交易信息失败：%v", err)
		}
		if bytes != nil {
			var transaction Transaction
			err = json.Unmarshal(bytes, &transaction)
			if err != nil {
				return nil, fmt.Errorf("解析交易信息失败：%v", err)
			}
			return &transaction, nil
		}
	}

	return nil, fmt.Errorf("交易ID %s 不存在", txID)
}

// QueryRealEstateList 分页查询房产列表
func (s *SmartContract) QueryRealEstateList(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string, status string) (*QueryResult, error) {
	var iterator shim.StateQueryIteratorInterface
	var metadata *peer.QueryResponseMetadata
	var err error

	if status != "" {
		iterator, metadata, err = ctx.GetStub().GetStateByPartialCompositeKeyWithPagination(
			REAL_ESTATE,
			[]string{status},
			pageSize,
			bookmark,
		)
	} else {
		iterator, metadata, err = ctx.GetStub().GetStateByPartialCompositeKeyWithPagination(
			REAL_ESTATE,
			[]string{},
			pageSize,
			bookmark,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("查询列表失败：%v", err)
	}
	defer iterator.Close()

	records := make([]interface{}, 0)
	for iterator.HasNext() {
		queryResponse, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("获取下一条记录失败：%v", err)
		}

		var realEstate RealEstate
		err = json.Unmarshal(queryResponse.Value, &realEstate)
		if err != nil {
			return nil, fmt.Errorf("解析房产信息失败：%v", err)
		}

		records = append(records, realEstate)
	}

	return &QueryResult{
		Records:             records,
		RecordsCount:        int32(len(records)),
		Bookmark:            metadata.Bookmark,
		FetchedRecordsCount: metadata.FetchedRecordsCount,
	}, nil
}

// QueryTransactionList 分页查询交易列表
func (s *SmartContract) QueryTransactionList(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string, status string) (*QueryResult, error) {
	var iterator shim.StateQueryIteratorInterface
	var metadata *peer.QueryResponseMetadata
	var err error

	if status != "" {
		iterator, metadata, err = ctx.GetStub().GetStateByPartialCompositeKeyWithPagination(
			TRANSACTION,
			[]string{status},
			pageSize,
			bookmark,
		)
	} else {
		iterator, metadata, err = ctx.GetStub().GetStateByPartialCompositeKeyWithPagination(
			TRANSACTION,
			[]string{},
			pageSize,
			bookmark,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("查询列表失败：%v", err)
	}
	defer iterator.Close()

	records := make([]interface{}, 0)
	for iterator.HasNext() {
		queryResponse, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("获取下一条记录失败：%v", err)
		}

		var transaction Transaction
		err = json.Unmarshal(queryResponse.Value, &transaction)
		if err != nil {
			return nil, fmt.Errorf("解析交易信息失败：%v", err)
		}

		records = append(records, transaction)
	}

	return &QueryResult{
		Records:             records,
		RecordsCount:        int32(len(records)),
		Bookmark:            metadata.Bookmark,
		FetchedRecordsCount: metadata.FetchedRecordsCount,
	}, nil
}

// Hello 用于验证
func (s *SmartContract) Hello(ctx contractapi.TransactionContextInterface) (string, error) {
	return "hello", nil
}

// InitLedger 初始化账本
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	log.Println("InitLedger")
	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("创建智能合约失败：%v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("启动智能合约失败：%v", err)
	}
}

```

---

## U18_donation_chaincode.go

- Bytes: 3351 | Lines: 119

```go
package main

import (
    "encoding/json"
    "fmt"
    "time"

    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing donations on the ledger.
type SmartContract struct {
    contractapi.Contract
}

// Donation defines the structure stored in the ledger for each donation.
type Donation struct {
    DonationID string  `json:"DonationID"`
    DonorID    string  `json:"DonorID"`
    Amount     float64 `json:"Amount"`
    Timestamp  string  `json:"Timestamp"`
    CampaignID string  `json:"CampaignID"`
    ReceiverID string  `json:"ReceiverID"`
}

// InitLedger can be used to seed the ledger initially (no-op here).
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
    return nil
}

// CreateDonation records a new donation in the world state.
func (s *SmartContract) CreateDonation(ctx contractapi.TransactionContextInterface,
    donationID, donorID string,
    amount float64,
    campaignID, receiverID string) error {

    existing, err := ctx.GetStub().GetState(donationID)
    if err != nil {
        return fmt.Errorf("failed to read from world state: %v", err)
    }
    if existing != nil {
        return fmt.Errorf("the donation %s already exists", donationID)
    }

    donation := Donation{
        DonationID: donationID,
        DonorID:    donorID,
        Amount:     amount,
        Timestamp:  time.Now().UTC().Format(time.RFC3339),
        CampaignID: campaignID,
        ReceiverID: receiverID,
    }

    donationJSON, err := json.Marshal(donation)
    if err != nil {
        return err
    }

    return ctx.GetStub().PutState(donationID, donationJSON)
}

// GetDonation retrieves a donation by its ID from the world state.
func (s *SmartContract) GetDonation(ctx contractapi.TransactionContextInterface,
    donationID string) (*Donation, error) {

    donationJSON, err := ctx.GetStub().GetState(donationID)
    if err != nil {
        return nil, fmt.Errorf("failed to read from world state: %v", err)
    }
    if donationJSON == nil {
        return nil, fmt.Errorf("donation %s does not exist", donationID)
    }

    var donation Donation
    err = json.Unmarshal(donationJSON, &donation)
    if err != nil {
        return nil, err
    }

    return &donation, nil
}

// QueryAllDonations returns all donations found in world state.
func (s *SmartContract) QueryAllDonations(ctx contractapi.TransactionContextInterface) ([]*Donation, error) {
    resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
    if err != nil {
        return nil, err
    }
    defer resultsIterator.Close()

    var donations []*Donation
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }

        var donation Donation
        err = json.Unmarshal(queryResponse.Value, &donation)
        if err != nil {
            return nil, err
        }

        donations = append(donations, &donation)
    }

    return donations, nil
}

func main() {
    chaincode, err := contractapi.NewChaincode(&SmartContract{})
    if err != nil {
        panic(fmt.Sprintf("Error creating donation chaincode: %v", err))
    }

    if err := chaincode.Start(); err != nil {
        fmt.Printf("Error starting Donation chaincode: %v", err)
    }
}

```

---

## U20_maintenance.go

- Bytes: 16191 | Lines: 578

```go
package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type Machine struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Model           string         `json:"model"`
	OperatingHours  int            `json:"operatingHours"`
	Status          string         `json:"status"`
	Owner           string         `json:"owner"`
	Interventions   []Intervention `json:"interventions"`
	LastUpdate      string         `json:"lastUpdate"`
	NextMaintenance int            `json:"nextMaintenance"`
}

type Intervention struct {
	Date        string `json:"date"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Technician  string `json:"technician"`
	MachineID   string `json:"machineId"`
}

type Alert struct {
	ID      string `json:"id"`
	MachineID      string `json:"machineId"`
	MachineName    string `json:"machineName"`
	AlertType      string `json:"alertType"`
	Message        string `json:"message"`
	Timestamp      string `json:"timestamp"`
}

type SmartContract struct{}

// Estrae l'MSP ID dal creator usando protobuf
func getCreatorMSPID(creatorBytes []byte) (string, error) {
	if len(creatorBytes) < 2 {
		return "", fmt.Errorf("creator bytes troppo corti")
	}
	
	if creatorBytes[0] != 0x0a {
		return "", fmt.Errorf("formato creator non valido")
	}
	
	mspIDLen := int(creatorBytes[1])
	
	if len(creatorBytes) < 2+mspIDLen {
		return "", fmt.Errorf("creator bytes incompleti")
	}
	
	mspID := string(creatorBytes[2 : 2+mspIDLen])
	
	return mspID, nil
}

func (s *SmartContract) InitLedger(stub shim.ChaincodeStubInterface) peer.Response {
	machines := []Machine{
		{
			ID:              "MACH001",
			Name:            "Tornio CNC",
			Model:           "Haas ST-30",
			OperatingHours:  450,
			Status:          "funzionante",
			Owner:           "OwnerMSP",
			LastUpdate:      time.Now().Format(time.RFC3339),
			NextMaintenance: 550,
			Interventions: []Intervention{
				{
					Date:        "2025-01-10T10:00:00Z",
					Type:        "ordinaria",
					Description: "Cambio olio lubrificante",
					Technician:  "Mario Rossi",
					MachineID:   "MACH001",
				},
			},
		},
		{
			ID:              "MACH002",
			Name:            "Fresatrice CNC",
			Model:           "DMG Mori NVX",
			OperatingHours:  1200,
			Status:          "funzionante",
			Owner:           "OwnerMSP",
			LastUpdate:      time.Now().Format(time.RFC3339),
			NextMaintenance: 1400,
			Interventions: []Intervention{
				{
					Date:        "2025-01-05T14:30:00Z",
					Type:        "ordinaria",
					Description: "Verifica parametri elettrici",
					Technician:  "Luigi Verdi",
					MachineID:   "MACH002",
				},
			},
		},
	}

	for _, machine := range machines {
		machineJSON, err := json.Marshal(machine)
		if err != nil {
			return shim.Error(err.Error())
		}
		err = stub.PutState(machine.ID, machineJSON)
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to put machine %s: %v", machine.ID, err))
		}
	}

	return shim.Success([]byte("Ledger initialized with 2 machines"))
}

func (s *SmartContract) RegisterMachine(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) < 3 || len(args) > 6 {
		return shim.Error("Numero argomenti errato. Uso: RegisterMachine <id> <name> <model> [operatingHours] [status] [interventionsJSON]")
	}

	// CONTROLLO ACCESSO: SOLO OwnerMSP puo' registrare una nuova macchina
	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("Impossibile recuperare creator: %v", err))
	}

	creatorMSPID, err := getCreatorMSPID(creatorBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Impossibile estrarre MSP ID: %v", err))
	}

	if creatorMSPID != "OwnerMSP" {
		return shim.Error(fmt.Sprintf("Accesso negato: solo OwnerMSP puo' modificare stato (chiamante: %s)", creatorMSPID))
	}

	id := args[0]
	name := args[1]
	model := args[2]

	existingMachineJSON, err := stub.GetState(id)
	if err != nil {
		return shim.Error(fmt.Sprintf("Errore controllo esistenza: %v", err))
	}
	if existingMachineJSON != nil {
		return shim.Error(fmt.Sprintf("Macchina %s gia' esistente", id))
	}

	// Parametri opzionali
	operatingHours := 0
	status := "funzionante"
	
	if len(args) >= 4 {
		hours, err := strconv.Atoi(args[3])
		if err != nil {
			return shim.Error("Le ore devono essere un numero intero")
		}
		operatingHours = hours
	}
	
	if len(args) >= 5 {
		if args[4] != "funzionante" && args[4] != "guasto" {
			return shim.Error("Status deve essere 'funzionante' o 'guasto'")
		}
		status = args[4]
	}

	interventions := []Intervention{}
	if len(args) >= 6 && args[5] != "" && args[5] != "[]" {
		err := json.Unmarshal([]byte(args[5]), &interventions)
		if err != nil {
			return shim.Error(fmt.Sprintf("Formato interventi non valido: %v", err))
		}
	}

	machine := Machine{
		ID:              id,
		Name:            name,
		Model:           model,
		OperatingHours:  operatingHours,
		Status:          status,
		Owner:           creatorMSPID,
		Interventions:   interventions,
		LastUpdate:      time.Now().Format(time.RFC3339),
		NextMaintenance: operatingHours + 200,
	}

	machineJSON, err := json.Marshal(machine)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(id, machineJSON)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(fmt.Sprintf("Macchina %s registrata con successo", id)))
}

func (s *SmartContract) UpdateOperatingHours(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Numero argomenti errato. Uso: UpdateOperatingHours <id> <hours>")
	}

	id := args[0]
	hours, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Le ore devono essere un numero intero")
	}

	// CONTROLLO ACCESSO: SOLO OwnerMSP puo' aggiornare ore di lavoro
	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("Impossibile recuperare creator: %v", err))
	}

	creatorMSPID, err := getCreatorMSPID(creatorBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Impossibile estrarre MSP ID: %v", err))
	}

	if creatorMSPID != "OwnerMSP" {
		return shim.Error(fmt.Sprintf("Accesso negato: solo OwnerMSP puo' modificare stato (chiamante: %s)", creatorMSPID))
	}

	machineJSON, err := stub.GetState(id)
	if err != nil {
		return shim.Error(fmt.Sprintf("Errore recupero macchina: %v", err))
	}
	if machineJSON == nil {
		return shim.Error(fmt.Sprintf("Macchina %s non trovata", id))
	}

	var machine Machine
	err = json.Unmarshal(machineJSON, &machine)
	if err != nil {
		return shim.Error(err.Error())
	}

	machine.OperatingHours += hours
	machine.LastUpdate = time.Now().Format(time.RFC3339)

	alertMessage := ""
	if machine.OperatingHours >= machine.NextMaintenance {
		alertMessage = fmt.Sprintf(" | ALERT: Manutenzione richiesta (ore attuali: %d, prossima manutenzione prevista: %d)",
			machine.OperatingHours, machine.NextMaintenance)
		machine.NextMaintenance += 200
	}

	machineJSON, err = json.Marshal(machine)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(id, machineJSON)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(fmt.Sprintf("Macchina %s aggiornata: +%d ore (totale: %d)%s", id, hours, machine.OperatingHours, alertMessage)))
}

func (s *SmartContract) SetMachineStatus(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Numero argomenti errato. Uso: SetMachineStatus <id> <status>")
	}

	id := args[0]
	status := args[1]

	if status != "funzionante" && status != "guasto" {
		return shim.Error("Status deve essere 'funzionante' o 'guasto'")
	}

	// CONTROLLO ACCESSO: SOLO OwnerMSP puo' cambiare lo stato
	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("Impossibile recuperare creator: %v", err))
	}

	creatorMSPID, err := getCreatorMSPID(creatorBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Impossibile estrarre MSP ID: %v", err))
	}

	if creatorMSPID != "OwnerMSP" {
		return shim.Error(fmt.Sprintf("Accesso negato: solo OwnerMSP puo' modificare stato (chiamante: %s)", creatorMSPID))
	}

	machineJSON, err := stub.GetState(id)
	if err != nil {
		return shim.Error(fmt.Sprintf("Errore recupero macchina: %v", err))
	}
	if machineJSON == nil {
		return shim.Error(fmt.Sprintf("Macchina %s non trovata", id))
	}

	var machine Machine
	err = json.Unmarshal(machineJSON, &machine)
	if err != nil {
		return shim.Error(err.Error())
	}

	machine.Status = status
	machine.LastUpdate = time.Now().Format(time.RFC3339)

	machineJSON, err = json.Marshal(machine)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(id, machineJSON)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(fmt.Sprintf("Status macchina %s aggiornato a: %s", id, status)))
}

func (s *SmartContract) ReadMachine(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Numero argomenti errato. Uso: ReadMachine <id>")
	}

	id := args[0]

	machineJSON, err := stub.GetState(id)
	if err != nil {
		return shim.Error(fmt.Sprintf("Errore recupero macchina: %v", err))
	}
	if machineJSON == nil {
		return shim.Error(fmt.Sprintf("Macchina %s non trovata", id))
	}

	return shim.Success(machineJSON)
}

func (s *SmartContract) GetAllMachines(stub shim.ChaincodeStubInterface) peer.Response {
	resultsIterator, err := stub.GetStateByRange("MACH", "MACH~")
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var machines []Machine

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var machine Machine
		err = json.Unmarshal(queryResponse.Value, &machine)
		if err != nil {
			return shim.Error(err.Error())
		}
		machines = append(machines, machine)
	}

	machinesJSON, err := json.Marshal(machines)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(machinesJSON)
}

func (s *SmartContract) AddIntervention(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 4 {
		return shim.Error("Numero argomenti errato. Uso: AddIntervention <id> <type> <description> <technician>")
	}

	id := args[0]
	interventionType := args[1]
	description := args[2]
	technician := args[3]

	if interventionType != "ordinaria" && interventionType != "straordinaria" {
		return shim.Error("Type deve essere 'ordinaria' o 'straordinaria'")
	}

	// CONTROLLO ACCESSO: SOLO OrdinaryMSP OR ExtraordinaryMSP possono inserire un intervento + controllo in base al tipo di intervento
	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("Impossibile recuperare creator: %v", err))
	}

	creatorMSPID, err := getCreatorMSPID(creatorBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Impossibile estrarre MSP ID: %v", err))
	}

	if creatorMSPID == "OwnerMSP" {
		return shim.Error("Accesso negato: OwnerMSP non puo' inserire una manutenzione, altrimenti potrebbe validarla unilateralmente")
	}

	if interventionType == "ordinaria" && creatorMSPID != "OrdinaryMSP" {
		return shim.Error(fmt.Sprintf("Accesso negato: solo OrdinaryMSP puo' inserire una manutenzione ordinaria (chiamante: %s)", creatorMSPID))
	}

	if interventionType == "straordinaria" && creatorMSPID != "ExtraordinaryMSP" {
		return shim.Error(fmt.Sprintf("Accesso negato: solo ExtraordinaryMSP puo' inserire una manutenzione straordinaria (chiamante: %s)", creatorMSPID))
	}

	machineJSON, err := stub.GetState(id)
	if err != nil {
		return shim.Error(fmt.Sprintf("Errore recupero macchina: %v", err))
	}
	if machineJSON == nil {
		return shim.Error(fmt.Sprintf("Macchina %s non trovata", id))
	}

	var machine Machine
	err = json.Unmarshal(machineJSON, &machine)
	if err != nil {
		return shim.Error(err.Error())
	}

	intervention := Intervention{
		Date:        time.Now().Format(time.RFC3339),
		Type:        interventionType,
		Description: description,
		Technician:  technician,
		MachineID:   id,
	}

	machine.Interventions = append(machine.Interventions, intervention)
	machine.LastUpdate = time.Now().Format(time.RFC3339)

	if interventionType == "straordinaria" || interventionType == "ordinaria" {
		machine.Status = "funzionante"
	}

	machineJSON, err = json.Marshal(machine)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(id, machineJSON)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(fmt.Sprintf("Intervento %s aggiunto per macchina %s", interventionType, id)))
}

func (s *SmartContract) CreateAlert(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 4 {
		return shim.Error("Numero argomenti errato. Uso: CreateAlert <machineId> <machineName> <alertType> <message>")
	}

	machineID := args[0]
	machineName := args[1]
	alertType := args[2]
	message := args[3]

	// CONTROLLO ACCESSO: SOLO OwnerMSP OR OrdinaryMSP possono creare una nuova segnalazione
	creatorBytes, err := stub.GetCreator()
	if err != nil {
		return shim.Error(fmt.Sprintf("Impossibile recuperare creator: %v", err))
	}

	creatorMSPID, err := getCreatorMSPID(creatorBytes)
	if err != nil {
		return shim.Error(fmt.Sprintf("Impossibile estrarre MSP ID: %v", err))
	}

	if creatorMSPID != "OwnerMSP" && creatorMSPID != "OrdinaryMSP" {
		return shim.Error(fmt.Sprintf("Accesso negato: solo OwnerMSP AND OrdinaryMSP possono creare una nuova segnalazione (chiamante: %s)", creatorMSPID))
	}

	timestamp := time.Now().Format(time.RFC3339)
	alertID := fmt.Sprintf("ALERT_%s_%s", machineID, timestamp)

	alert := Alert{
		ID:          alertID,
		MachineID:   machineID,
		MachineName: machineName,
		AlertType:   alertType,
		Message:     message,
		Timestamp:   timestamp,
	}

	alertJSON, err := json.Marshal(alert)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(alertID, alertJSON)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte(fmt.Sprintf("Alert %s creato con successo", alertID)))
}

func (s *SmartContract) GetAlertsByMachine(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Numero argomenti errato. Uso: GetAlertsByMachine <machineId>")
	}

	machineID := args[0]

	startKey := fmt.Sprintf("ALERT_%s_", machineID)
	endKey := fmt.Sprintf("ALERT_%s_~", machineID)

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var alerts []Alert

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var alert Alert
		err = json.Unmarshal(queryResponse.Value, &alert)
		if err != nil {
			return shim.Error(err.Error())
		}
		alerts = append(alerts, alert)
	}

	alertsJSON, err := json.Marshal(alerts)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(alertsJSON)
}

func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()

	switch function {
	case "InitLedger":
		return s.InitLedger(stub)
	case "RegisterMachine":
		return s.RegisterMachine(stub, args)
	case "UpdateOperatingHours":
		return s.UpdateOperatingHours(stub, args)
	case "SetMachineStatus":
		return s.SetMachineStatus(stub, args)
	case "ReadMachine":
		return s.ReadMachine(stub, args)
	case "GetAllMachines":
		return s.GetAllMachines(stub)
	case "AddIntervention":
		return s.AddIntervention(stub, args)
	case "CreateAlert":
		return s.CreateAlert(stub, args)
	case "GetAlertsByMachine":
		return s.GetAlertsByMachine(stub, args)
	default:
		return shim.Error(fmt.Sprintf("Funzione %s non riconosciuta", function))
	}
}

func main() {
	if err := shim.Start(new(SmartContract)); err != nil {
		fmt.Printf("Errore avvio chaincode: %v\n", err)
	}
}

```

---

## U21_movies.go

- Bytes: 15595 | Lines: 519

```go
package main

/* Imports
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts

	"bytes"

	"math"
	"strconv"
	"time"
*/
import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
	"strconv"
)

// Define the Smart Contract structure
type SmartContract struct {
}

/*
const (
	MAX_MOVIE  = 5
	MAX_SHOWS = 4
	MAX_TICKETS = 100
	MAX_WINDOWS = 4
)
*/
const (
	NEXT_SHOW_ID   = "NEXT_SHOW_ID"
	NEXT_TICKET_ID = "NEXT_TICKET_ID"
)

//doctypes
const (
	SHOW   = "SHOW"
	TICKET = "TICKET"
	WINDOW = "WINDOW"
	SODA   = "SODA"
)

type Theatre struct {
	TheatreNo      int    `json:"theatreNo"`
	TheatreName    string `json:"theatreName"`
	Windows        int    `json:"windows, omitempty"`
	TicketsPerShow int    `json:"ticketsPerShow, omitempty"`
	ShowsDaily     int    `json:"showsDaily, omitempty"`
	SodaStock      int    `json:"sodaStock, omitempty"`
	Halls          int    `json:"halls, omitempty"`
	DocType        string `json:"docType"`
}

type Window struct {
	WindowNo    int    `json:"windowNo"`
	TicketsSold int    `json:"ticketsSold"`
	DocType     string `json:"docType"`
}

type Ticket struct {
	TicketNo        int     `json:"ticketNo"`
	Show            Show    `json:"show"`
	Window          Window  `json:"window"`
	Quantity        int     `json:"quantity,number"`
	Amount          float64 `json:"amount,string"`
	CouponNumber    string  `json:"couponNumber"`
	CouponAvailed   bool    `json:"couponAvailed"`
	ExchangeAvailed bool    `json:"exchangeAvailed"`
	DocType         string  `json:"docType"`
}

type Show struct {
	ShowID    int    `json:"showID"`
	Movie     string `json:"movie"`
	ShowSlot  string `json:"showSlot"`
	Quantity  int    `json:"quantity,number"`
	HallNo    int    `json:"hallNo"`
	TheatreNo int    `json:"theatreNo"`
	DocType   string `json:"docType"`
}

type Soda struct {
	Stock        int    `json:"stock"`
	TicketNo     int    `json:"ticketNo"`
	CouponNumber string `json:"couponNumber"`
	DocType      string `json:"docType"`
}

type Property struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CreateShows struct {
	TheatreNo int    `json:"theatreNo"`
	Shows     []Show `json:"shows"`
}

// =========================================================================================
// The Init method is called when the Smart Contract "fabcar" is instantiated by the blockchain network
// Best practice is to have any Ledger initialization in separate function -- see initLedger()
// =========================================================================================
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {

	_, err := set(APIstub, NEXT_SHOW_ID, "0")
	_, err = set(APIstub, NEXT_TICKET_ID, "0")

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// =========================================================================================
// The Invoke method is called as a result of an application request to run the Smart Contract "fabcar"
// The calling application program has also specified the particular smart contract function to be called, with arguments
// =========================================================================================
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "registerTheatre" {
		return s.registerTheatre(APIstub, args)
	} else if function == "createShow" {
		return s.createShow(APIstub, args)
	} else if function == "purchaseTicket" {
		return s.purchaseTicket(APIstub, args)
	} else if function == "issueCoupon" {
		return s.issueCoupon(APIstub, args)
	} else if function == "availExchange" {
		return s.availExchange(APIstub, args)
	} else if function == "queryByString" {
		return s.queryByString(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.:" + function)
}

func (s *SmartContract) registerTheatre(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	fmt.Println("API::registerTheatre:Start")
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	var theatre Theatre
	if err := json.Unmarshal([]byte(args[0]), &theatre); err != nil {
		fmt.Println("Cannot unmarshal theatre Object", err)
		return shim.Error(err.Error())
	}
	// Create unique theatre number & save theatre
	txnId := APIstub.GetTxID()
	var number int
	for _, c := range txnId {
		number = number + int(c)
	}
	theatre.TheatreNo = number
	theatre.DocType = "THEATRE"
	theatreAsBytes, _ := json.Marshal(theatre)
	err := APIstub.PutState("THEATRE"+strconv.Itoa(theatre.TheatreNo), theatreAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	// create windows for the theatre
	for i := 1; i <= theatre.Windows; i++ {
		var window Window
		window.WindowNo = i
		window.TicketsSold = 0
		window.DocType = WINDOW
		windowAsBytes, _ := json.Marshal(window)
		err := APIstub.PutState("WINDOW"+strconv.Itoa(i), windowAsBytes)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	fmt.Println("API::registerTheatre:End")
	return shim.Success([]byte("MovieTheatre Number:" + strconv.Itoa(theatre.TheatreNo)))
}

func (s *SmartContract) createShow(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	fmt.Println("API::createShow:Start")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	var createShows CreateShows

	if err := json.Unmarshal([]byte(args[0]), &createShows); err != nil {
		fmt.Println("Cannot unmarshal createShows Object", err)
		return shim.Error(err.Error())
	}

	showSeq, err := get(APIstub, NEXT_SHOW_ID)
	fmt.Println("Generating show for showSeq", showSeq)

	if err != nil {
		return shim.Error(err.Error())
	}

	shows := createShows.Shows
	var theatre Theatre
	theatreBytes, err := APIstub.GetState("THEATRE" + strconv.Itoa(createShows.TheatreNo))
	if err != nil {
		return shim.Error(err.Error())
	}
	json.Unmarshal(theatreBytes, &theatre)

	if len(shows) > theatre.Halls {
		return shim.Error("Number of Movies cannot exceed" + strconv.Itoa(theatre.Halls))
	}
	for _, show := range shows {

		for i := 1; i <= theatre.ShowsDaily; i++ {
			showSeq = showSeq + 1
			show.ShowID = +showSeq
			show.ShowSlot = strconv.Itoa(i)
			show.Quantity = theatre.TicketsPerShow
			show.TheatreNo = theatre.TheatreNo
			show.DocType = SHOW
			showAsBytes, _ := json.Marshal(show)
			err = APIstub.PutState("SHOW"+strconv.Itoa(show.ShowID), showAsBytes)
			if err != nil {
				return shim.Error(err.Error())
			}
		}
	}
	fmt.Println("saving current showSeq", showSeq)
	_, err = set(APIstub, NEXT_SHOW_ID, strconv.Itoa(showSeq))
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("API::createShow:End")
	return shim.Success([]byte(APIstub.GetTxID()))
}

func (s *SmartContract) purchaseTicket(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	fmt.Println("API::purchaseTicket:Start")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	var ticket Ticket

	if err := json.Unmarshal([]byte(args[0]), &ticket); err != nil {
		fmt.Println("Cannot unmarshal ticket Object", err)
		return shim.Error(err.Error())
	}

	ticketSeq, err := get(APIstub, NEXT_TICKET_ID)
	fmt.Println("Generating Ticket for ticketSeq", ticketSeq)

	if err != nil {
		return shim.Error(err.Error())
	}

	showBytes, err := APIstub.GetState("SHOW" + strconv.Itoa(ticket.Show.ShowID))
	if err != nil {
		return shim.Error(err.Error())
	}
	var show Show
	json.Unmarshal(showBytes, &show)

	windowBytes, err := APIstub.GetState("WINDOW" + strconv.Itoa(ticket.Window.WindowNo))
	if err != nil {
		return shim.Error(err.Error())
	}
	var window Window
	json.Unmarshal(windowBytes, &window)
	// check the show for number of seats remaining
	if show.Quantity < 0 || show.Quantity-ticket.Quantity < 0 {
		return shim.Error("Seats Full for the requested show or Not enough seats as requested. Available:" + strconv.Itoa(show.Quantity))
	}

	show.Quantity = show.Quantity - ticket.Quantity
	window.TicketsSold = window.TicketsSold + ticket.Quantity
	fmt.Println(window.TicketsSold)
	fmt.Println(ticket.Quantity)
	ticketSeq = ticketSeq + 1
	ticket.TicketNo = ticketSeq
	ticket.Show = show
	ticket.Window = window
	ticket.DocType = TICKET

	showAsBytes, _ := json.Marshal(show)
	err = APIstub.PutState("SHOW"+strconv.Itoa(show.ShowID), showAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	windowAsBytes, _ := json.Marshal(window)
	err = APIstub.PutState("WINDOW"+strconv.Itoa(window.WindowNo), windowAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("saving current ticketSeq", ticketSeq)
	_, err = set(APIstub, NEXT_TICKET_ID, strconv.Itoa(ticketSeq))
	if err != nil {
		return shim.Error(err.Error())
	}

	ticketAsBytes, _ := json.Marshal(ticket)
	err = APIstub.PutState("TICKET"+strconv.Itoa(ticketSeq), ticketAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("API::purchaseTicket:End")
	return shim.Success([]byte(APIstub.GetTxID()))
}

// Issue coupon for the waterbottle and popcorn also for the soda exchange
func (s *SmartContract) issueCoupon(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	fmt.Println("API::issueCoupon:Start")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	var ticket Ticket

	if err := json.Unmarshal([]byte(args[0]), &ticket); err != nil {
		fmt.Println("Cannot unmarshal ticket Object", err)
		return shim.Error(err.Error())
	}
	ticketBytes, err := APIstub.GetState("TICKET" + strconv.Itoa(ticket.TicketNo))
	if err != nil {
		return shim.Error(err.Error())
	}
	json.Unmarshal(ticketBytes, &ticket)

	if ticket.CouponAvailed {
		fmt.Println("Coupon Availed Already")
		return shim.Error("Coupon Availed Already")
	}

	txnId := APIstub.GetTxID()
	var number int
	for _, c := range txnId {
		number = number + int(c)
	}
	ticket.CouponNumber = strconv.Itoa(number)
	ticket.CouponAvailed = true
	ticket.ExchangeAvailed = false
	ticketAsBytes, _ := json.Marshal(ticket)
	err = APIstub.PutState("TICKET"+strconv.Itoa(ticket.TicketNo), ticketAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("API::issueCoupon:End")
	return shim.Success([]byte("Coupon Number:" + ticket.CouponNumber))
}

func (s *SmartContract) availExchange(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	fmt.Println("API::availExchange:Start")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	var ticket Ticket

	if err := json.Unmarshal([]byte(args[0]), &ticket); err != nil {
		fmt.Println("Cannot unmarshal ticket Object", err)
		return shim.Error(err.Error())
	}
	ticketBytes, err := APIstub.GetState("TICKET" + strconv.Itoa(ticket.TicketNo))
	if err != nil {
		return shim.Error(err.Error())
	}
	json.Unmarshal(ticketBytes, &ticket)

	if ticket.ExchangeAvailed {
		fmt.Println("Exchange Availed Already")
		return shim.Error("Exchange Availed Already")
	}
	// check if even number for the eligible soda exchange
	couponNo, err := strconv.Atoi(ticket.CouponNumber)
	if err != nil {
		fmt.Println("Ticket Not eligible for exchange")
		return shim.Error("Ticket Not eligible for exchange")
	}
	if couponNo%2 != 0 {
		fmt.Println("Ticket Not eligible for exchange")
		return shim.Error("Ticket Not eligible for exchange")
	}
	ticket.ExchangeAvailed = true
	ticketAsBytes, _ := json.Marshal(ticket)
	err = APIstub.PutState("TICKET"+strconv.Itoa(ticket.TicketNo), ticketAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	var theatre Theatre
	theatreBytes, err := APIstub.GetState("THEATRE" + strconv.Itoa(ticket.Show.TheatreNo))
	if err != nil {
		return shim.Error(err.Error())
	}
	json.Unmarshal(theatreBytes, &theatre)

	theatre.SodaStock = theatre.SodaStock - ticket.Quantity

	theatreAsBytes, _ := json.Marshal(theatre)
	err = APIstub.PutState("THEATRE"+strconv.Itoa(theatre.TheatreNo), theatreAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("API::availExchange:End")
	return shim.Success([]byte(APIstub.GetTxID()))
}

func (s *SmartContract) queryByString(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	queryString := args[0]
	fmt.Println("queryString" + queryString)
	queryResults, err := getQueryResultForQueryString(APIstub, queryString)

	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(queryResults)
}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString(string(queryResponse.Value))
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func get(APIstub shim.ChaincodeStubInterface, key string) (int, error) {
	if key == "" {
		return 0, fmt.Errorf("Incorrect arguments. Expecting a key")
	}
	value, err := APIstub.GetState(key)
	if err != nil {
		return 0, fmt.Errorf("Failed to get asset: %s with error: %s", key, err)
	}
	if value == nil {
		return 0, fmt.Errorf("Asset not found: %s", key)
	}
	var property Property
	json.Unmarshal(value, &property)
	fmt.Println(property)
	fmt.Println(property.Value)
	i, err := strconv.Atoi(property.Value)
	if err != nil {
		return 0, fmt.Errorf("Failed to get next sequence number", err)
	}
	fmt.Println("Got the the value for %s : value : %s", key, i)
	return i, nil
}

func set(APIstub shim.ChaincodeStubInterface, key string, value string) (string, error) {
	fmt.Println("setting value", key, value)

	var property Property
	property.Key = key
	property.Value = value

	propertyAsBytes, _ := json.Marshal(property)
	err := APIstub.PutState(key, propertyAsBytes)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	return value, nil
}

// =========================================================================================
// The main function is only relevant in unit test mode. Only included here for completeness.
// =========================================================================================
func main() {
	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}

```

---

## U22_voting.go

- Bytes: 19537 | Lines: 546

```go
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

```

---

## U23_private_blockchain.go

- Bytes: 7347 | Lines: 230

```go
package main // Package main, Do not change this line.

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Product represents the structure for a product entity
type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Owner       string `json:"owner"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// SupplyChainContract defines the smart contract structure
type SupplyChainContract struct {
	contractapi.Contract
}

// getTimestamp returns the transaction timestamp as a string
func (s *SupplyChainContract) getTimestamp(ctx contractapi.TransactionContextInterface) (string, error) {
	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return "", fmt.Errorf("failed to get transaction timestamp: %v", err)
	}
	return time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339), nil
}

// InitLedger initializes the ledger with some example products
func (s *SupplyChainContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	timestamp, err := s.getTimestamp(ctx)
	if err != nil {
		return err
	}

	// Initial set of products to populate the ledger
	products := []Product{
		{ID: "p1", Name: "Laptop", Status: "Manufactured", Owner: "CompanyA", CreatedAt: timestamp, UpdatedAt: timestamp, Description: "High-end gaming laptop", Category: "Electronics"},
		{ID: "p2", Name: "Smartphone", Status: "Manufactured", Owner: "CompanyB", CreatedAt: timestamp, UpdatedAt: timestamp, Description: "Latest model smartphone", Category: "Electronics"},
	}

	for _, product := range products {
		if err := s.putProduct(ctx, &product); err != nil {
			return err
		}
	}

	return nil
}


// CreateProduct creates a new product in the ledger
func (s *SupplyChainContract) CreateProduct(ctx contractapi.TransactionContextInterface, id, name, owner, description, category string) error {
	// Generate the current timestamp
	timestamp, err := s.getTimestamp(ctx)
	if err != nil {
		return fmt.Errorf("error fetching transaction timestamp: %v", err)
	}

	// Verify that the product does not already exist
	exists, err := s.ProductExists(ctx, id)
	if err != nil {
		return fmt.Errorf("error checking if product exists: %v", err)
	}
	if exists {
		return fmt.Errorf("product with ID %s already exists", id)
	}

	// Create the new product structure
	newProduct := &Product{
		ID:          id,
		Name:        name,
		Status:      "Manufactured",
		Owner:       owner,
		CreatedAt:   timestamp,
		UpdatedAt:   timestamp,
		Description: description,
		Category:    category,
	}

	// Store the new product in the ledger
	return s.putProduct(ctx, newProduct)
}

// UpdateProduct allows updating a product's status, owner, description, and category
func (s *SupplyChainContract) UpdateProduct(ctx contractapi.TransactionContextInterface, id, newStatus, newOwner, newDescription, newCategory string) error {
	// Retrieve the product from the ledger
	existingProduct, err := s.QueryProduct(ctx, id)
	if err != nil {
		return fmt.Errorf("error retrieving product with ID %s: %v", id, err)
	}

	// Update product attributes if new values are provided
	if len(newStatus) > 0 {
		existingProduct.Status = newStatus
	}
	if len(newOwner) > 0 {
		existingProduct.Owner = newOwner
	}
	if len(newDescription) > 0 {
		existingProduct.Description = newDescription
	}
	if len(newCategory) > 0 {
		existingProduct.Category = newCategory
	}

	// Update the modification timestamp
	timestamp, err := s.getTimestamp(ctx)
	if err != nil {
		return fmt.Errorf("error fetching transaction timestamp: %v", err)
	}
	existingProduct.UpdatedAt = timestamp

	// Store the updated product back in the ledger
	return s.putProduct(ctx, existingProduct)
}

// TransferOwnership changes the owner of a product
func (s *SupplyChainContract) TransferOwnership(ctx contractapi.TransactionContextInterface, id, newOwner string) error {
	// Check if the product exists
	exists, err := s.ProductExists(ctx, id)
	if err != nil {
		return fmt.Errorf("error checking product existence: %v", err)
	}
	if !exists {
		return fmt.Errorf("product with ID %s does not exist", id)
	}

	// Retrieve the product
	product, err := s.QueryProduct(ctx, id)
	if err != nil {
		return fmt.Errorf("error retrieving product with ID %s: %v", id, err)
	}

	// Update the owner and timestamp
	product.Owner = newOwner
	timestamp, err := s.getTimestamp(ctx)
	if err != nil {
		return fmt.Errorf("error fetching transaction timestamp: %v", err)
	}
	product.UpdatedAt = timestamp

	// Store the updated product back in the ledger
	return s.putProduct(ctx, product)
}

// QueryProduct retrieves a single product from the ledger by ID
func (s *SupplyChainContract) QueryProduct(ctx contractapi.TransactionContextInterface, id string) (*Product, error) {
	// Retrieve the product state
	productBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("error reading product from ledger: %v", err)
	}
	if productBytes == nil {
		return nil, fmt.Errorf("product with ID %s does not exist", id)
	}

	// Unmarshal the product JSON into a Product struct
	var product Product
	err = json.Unmarshal(productBytes, &product)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling product data: %v", err)
	}

	return &product, nil
}


// putProduct is a helper method for inserting or updating a product in the ledger
func (s *SupplyChainContract) putProduct(ctx contractapi.TransactionContextInterface, product *Product) error {
	productJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(product.ID, productJSON)
}

// ProductExists is a helper method to check if a product exists in the ledger
func (s *SupplyChainContract) ProductExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	productJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return productJSON != nil, nil
}

// GetAllProducts is a helper method to retrieve all products from the ledger
func (s *SupplyChainContract) GetAllProducts(ctx contractapi.TransactionContextInterface) ([]*Product, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var products []*Product
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var product Product
		if err := json.Unmarshal(queryResponse.Value, &product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	return products, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SupplyChainContract{})
	if err != nil {
		fmt.Printf("Error creating supply chain chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting supply chain chaincode: %s", err.Error())
	}
}

```

