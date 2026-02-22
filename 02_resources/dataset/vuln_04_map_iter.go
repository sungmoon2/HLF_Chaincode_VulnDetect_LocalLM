package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// InventoryContract manages warehouse inventory
type InventoryContract struct {
	contractapi.Contract
}

// Product represents an inventory item
type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

// RegisterProduct adds a new product to inventory
func (ic *InventoryContract) RegisterProduct(ctx contractapi.TransactionContextInterface, id string, name string, quantity int, price float64) error {
	product := Product{
		ID:       id,
		Name:     name,
		Quantity: quantity,
		Price:    price,
	}

	productJSON, err := json.Marshal(product)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, productJSON)
}

// BatchRestock updates quantities for multiple products at once
// [VULNERABILITY] Iterating over a Go map and writing to the ledger.
// Go's map iteration order is intentionally randomised by the runtime.
// Different peers iterate in different order, producing different
// write-set orderings and causing endorsement mismatch.
func (ic *InventoryContract) BatchRestock(ctx contractapi.TransactionContextInterface, restockJSON string) error {
	var restockMap map[string]int
	err := json.Unmarshal([]byte(restockJSON), &restockMap)
	if err != nil {
		return fmt.Errorf("invalid restock JSON: %v", err)
	}

	// [VULNERABILITY] Map iteration order is non-deterministic in Go.
	// On Peer A the loop may visit keys in order [X, Y, Z],
	// while on Peer B it visits [Z, X, Y].
	// Each PutState call appends to the transaction's write-set,
	// so different ordering means different proposals.
	for productID, addQty := range restockMap {
		productJSON, err := ctx.GetStub().GetState(productID)
		if err != nil {
			return fmt.Errorf("failed to read product %s: %v", productID, err)
		}
		if productJSON == nil {
			continue
		}

		var product Product
		json.Unmarshal(productJSON, &product)
		product.Quantity += addQty

		updatedJSON, _ := json.Marshal(product)
		// [VULNERABILITY] PutState inside map iteration —
		// write-set order depends on map iteration order.
		ctx.GetStub().PutState(productID, updatedJSON)
	}

	return nil
}

// GenerateReport builds a summary string by iterating a map
// [VULNERABILITY] Map iteration order affects the output string.
// The report written to the ledger will differ across peers.
func (ic *InventoryContract) GenerateReport(ctx contractapi.TransactionContextInterface, productIDs []string) error {
	totals := make(map[string]float64)

	for _, id := range productIDs {
		productJSON, err := ctx.GetStub().GetState(id)
		if err != nil {
			continue
		}
		if productJSON == nil {
			continue
		}

		var product Product
		json.Unmarshal(productJSON, &product)
		totals[product.Name] = float64(product.Quantity) * product.Price
	}

	// [VULNERABILITY] Building a string by iterating over a map.
	// The order of lines in `report` is non-deterministic.
	// Peer A might produce "Apples:100|Bananas:200" while
	// Peer B produces "Bananas:200|Apples:100".
	var parts []string
	for name, total := range totals {
		parts = append(parts, fmt.Sprintf("%s:%.2f", name, total))
	}
	report := strings.Join(parts, "|")

	// [VULNERABILITY] The non-deterministic string is written to the ledger.
	return ctx.GetStub().PutState("INVENTORY_REPORT", []byte(report))
}

// ApplyDiscounts applies category-specific discounts from a map
func (ic *InventoryContract) ApplyDiscounts(ctx contractapi.TransactionContextInterface, discountsJSON string) error {
	var discounts map[string]float64
	err := json.Unmarshal([]byte(discountsJSON), &discounts)
	if err != nil {
		return fmt.Errorf("invalid discounts JSON: %v", err)
	}

	// [VULNERABILITY] Again, iterating over a map of discounts.
	// The order in which products are updated differs per peer,
	// resulting in a different write-set order.
	for productID, discountPct := range discounts {
		productJSON, err := ctx.GetStub().GetState(productID)
		if err != nil {
			continue
		}
		if productJSON == nil {
			continue
		}

		var product Product
		json.Unmarshal(productJSON, &product)
		product.Price = product.Price * (1 - discountPct/100)

		updatedJSON, _ := json.Marshal(product)
		ctx.GetStub().PutState(productID, updatedJSON) // [VULNERABILITY] map-order dependent write
	}

	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&InventoryContract{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
