package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// AssetRegistryContract manages a registry of assets with range queries
type AssetRegistryContract struct {
	contractapi.Contract
}

// RegistryAsset represents a registered asset
type RegistryAsset struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Owner    string `json:"owner"`
	Value    int    `json:"value"`
	IsActive bool   `json:"isActive"`
}

// RegisterAsset creates a new asset in the registry
func (ar *AssetRegistryContract) RegisterAsset(ctx contractapi.TransactionContextInterface, id string, assetType string, owner string, value int) error {
	asset := RegistryAsset{
		ID:       id,
		Type:     assetType,
		Owner:    owner,
		Value:    value,
		IsActive: true,
	}

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// GetAssetsByRange retrieves assets within a key range
func (ar *AssetRegistryContract) GetAssetsByRange(ctx contractapi.TransactionContextInterface, startKey string, endKey string) ([]*RegistryAsset, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get state by range: %v", err)
	}

	var assets []*RegistryAsset
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset RegistryAsset
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}

		assets = append(assets, &asset)
	}

	return assets, nil
}

// SumAssetValues calculates total value of assets in a range
func (ar *AssetRegistryContract) SumAssetValues(ctx contractapi.TransactionContextInterface, startKey string, endKey string) (int, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return 0, fmt.Errorf("failed to query range: %v", err)
	}

	totalValue := 0
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return 0, err
		}

		var asset RegistryAsset
		json.Unmarshal(queryResult.Value, &asset)
		if asset.IsActive {
			totalValue += asset.Value
		}
	}

	return totalValue, nil
}

// TransferAssetsByOwner transfers all assets from one owner to another
func (ar *AssetRegistryContract) TransferAssetsByOwner(ctx contractapi.TransactionContextInterface, fromOwner string, toOwner string) error {
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("Asset", []string{fromOwner})
	if err != nil {
		return fmt.Errorf("failed to query assets by owner: %v", err)
	}

	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return err
		}

		var asset RegistryAsset
		json.Unmarshal(queryResult.Value, &asset)
		asset.Owner = toOwner

		updatedJSON, _ := json.Marshal(asset)
		err = ctx.GetStub().PutState(asset.ID, updatedJSON)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeactivateRange marks all assets in a range as inactive
func (ar *AssetRegistryContract) DeactivateRange(ctx contractapi.TransactionContextInterface, startKey string, endKey string) (int, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return 0, err
	}

	count := 0
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return count, err
		}

		var asset RegistryAsset
		json.Unmarshal(queryResult.Value, &asset)

		if asset.IsActive {
			asset.IsActive = false
			updatedJSON, _ := json.Marshal(asset)
			ctx.GetStub().PutState(asset.ID, updatedJSON)
			count++
		}
	}

	return count, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&AssetRegistryContract{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
