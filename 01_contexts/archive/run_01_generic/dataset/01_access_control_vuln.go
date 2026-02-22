package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// AssetTransferChaincode manages asset transfers
type AssetTransferChaincode struct {
	contractapi.Contract
}

// Asset describes basic details of an asset
type Asset struct {
	ID             string  `json:"id"`
	Owner          string  `json:"owner"`
	Value          float64 `json:"value"`
	TransferLocked bool    `json:"transferLocked"`
}

// InitLedger initializes the ledger with sample assets
func (s *AssetTransferChaincode) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "asset1", Owner: "Alice", Value: 1000.0, TransferLocked: false},
		{ID: "asset2", Owner: "Bob", Value: 2000.0, TransferLocked: false},
		{ID: "asset3", Owner: "Charlie", Value: 3000.0, TransferLocked: true},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put asset: %v", err)
		}
	}
	return nil
}

// VULNERABILITY: No access control - any user can transfer any asset
// Missing: GetClientIdentity().GetID() check against asset.Owner
func (s *AssetTransferChaincode) TransferAsset(ctx contractapi.TransactionContextInterface, assetID string, newOwner string) error {
	assetJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return fmt.Errorf("failed to read asset: %v", err)
	}
	if assetJSON == nil {
		return fmt.Errorf("asset %s does not exist", assetID)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return err
	}

	// VULNERABILITY: No ownership verification before transfer
	// Anyone can transfer anyone else's asset
	asset.Owner = newOwner

	updatedJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(assetID, updatedJSON)
}

// VULNERABILITY: No role-based access control for admin functions
// Missing: MSP ID or role check for privileged operations
func (s *AssetTransferChaincode) DeleteAsset(ctx contractapi.TransactionContextInterface, assetID string) error {
	exists, err := s.AssetExists(ctx, assetID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("asset %s does not exist", assetID)
	}

	// VULNERABILITY: No admin role verification
	// Any organization member can delete any asset
	return ctx.GetStub().DelState(assetID)
}

// VULNERABILITY: Unrestricted bulk update without authorization
func (s *AssetTransferChaincode) UpdateAssetValue(ctx contractapi.TransactionContextInterface, assetID string, newValue float64) error {
	assetJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return fmt.Errorf("failed to read asset: %v", err)
	}
	if assetJSON == nil {
		return fmt.Errorf("asset %s does not exist", assetID)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return err
	}

	// VULNERABILITY: No check if caller is the owner or an admin
	// No check if the asset is transfer-locked
	asset.Value = newValue

	updatedJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(assetID, updatedJSON)
}

// AssetExists checks if an asset exists
func (s *AssetTransferChaincode) AssetExists(ctx contractapi.TransactionContextInterface, assetID string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return false, fmt.Errorf("failed to read asset: %v", err)
	}
	return assetJSON != nil, nil
}

// ReadAsset returns the asset stored in the world state
func (s *AssetTransferChaincode) ReadAsset(ctx contractapi.TransactionContextInterface, assetID string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to read asset: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("asset %s does not exist", assetID)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&AssetTransferChaincode{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
