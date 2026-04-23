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
// [VULNERABILITY] Iterator from GetStateByRange is never closed.
// The iterator holds a gRPC stream to the peer's state database.
// Failing to close it leaks the stream, eventually exhausting
// file descriptors or gRPC connections on the peer.
func (ar *AssetRegistryContract) GetAssetsByRange(ctx contractapi.TransactionContextInterface, startKey string, endKey string) ([]*RegistryAsset, error) {
	// [VULNERABILITY] GetStateByRange returns an iterator that MUST be
	// closed with defer Close(). Without it, the underlying gRPC stream
	// and database cursor remain open, causing resource exhaustion.
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get state by range: %v", err)
	}
	// [VULNERABILITY] Missing: defer resultsIterator.Close()

	var assets []*RegistryAsset
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err // [VULNERABILITY] early return without closing iterator
		}

		var asset RegistryAsset
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err // [VULNERABILITY] another early return path without Close()
		}

		assets = append(assets, &asset)
	}

	return assets, nil
}

// SumAssetValues calculates total value of assets in a range
// [VULNERABILITY] Same pattern — iterator not closed.
func (ar *AssetRegistryContract) SumAssetValues(ctx contractapi.TransactionContextInterface, startKey string, endKey string) (int, error) {
	// [VULNERABILITY] GetStateByRange without defer Close().
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return 0, fmt.Errorf("failed to query range: %v", err)
	}
	// [VULNERABILITY] Missing: defer resultsIterator.Close()

	totalValue := 0
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return 0, err // [VULNERABILITY] iterator leaked on error
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
// [VULNERABILITY] Composite key iterator not closed.
func (ar *AssetRegistryContract) TransferAssetsByOwner(ctx contractapi.TransactionContextInterface, fromOwner string, toOwner string) error {
	// [VULNERABILITY] GetStateByPartialCompositeKey iterator not closed.
	// Even though this function uses a different query method,
	// the iterator still holds resources that must be released.
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("Asset", []string{fromOwner})
	if err != nil {
		return fmt.Errorf("failed to query assets by owner: %v", err)
	}
	// [VULNERABILITY] Missing: defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return err // [VULNERABILITY] leaked iterator on error
		}

		var asset RegistryAsset
		json.Unmarshal(queryResult.Value, &asset)
		asset.Owner = toOwner

		updatedJSON, _ := json.Marshal(asset)
		err = ctx.GetStub().PutState(asset.ID, updatedJSON)
		if err != nil {
			return err // [VULNERABILITY] leaked iterator on error
		}
	}

	return nil
}

// DeactivateRange marks all assets in a range as inactive
// [VULNERABILITY] Iterator leak in a write-heavy operation.
func (ar *AssetRegistryContract) DeactivateRange(ctx contractapi.TransactionContextInterface, startKey string, endKey string) (int, error) {
	// [VULNERABILITY] No defer Close() on the iterator.
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return 0, err
	}
	// [VULNERABILITY] Missing: defer resultsIterator.Close()

	count := 0
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return count, err // [VULNERABILITY] iterator leaked on error path
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
