package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// WarehouseContract manages multi-zone warehouse inventory
type WarehouseContract struct {
	contractapi.Contract
}

// ZoneInventory represents inventory data per warehouse zone
type ZoneInventory struct {
	ZoneID    string         `json:"zoneID"`
	ItemCount int            `json:"itemCount"`
	Items     map[string]int `json:"items"`
}

// BatchUpdateInventory processes a nested map of zone → item → quantity changes.
// [VULNERABILITY] The outer map (zone → items) and the inner map (item → qty)
// are both iterated with `range`. Go's map iteration order is randomized,
// so the sequence of PutState calls differs across endorsing peers.
// With nested maps, the nondeterminism is compounded: the outer loop picks
// zones in random order, and for each zone the inner loop picks items randomly.
func (w *WarehouseContract) BatchUpdateInventory(ctx contractapi.TransactionContextInterface, updateJSON string) error {
	// Outer map: zoneID → inner map
	// Inner map: itemID → quantity delta
	var updates map[string]map[string]int
	err := json.Unmarshal([]byte(updateJSON), &updates)
	if err != nil {
		return fmt.Errorf("invalid update JSON: %v", err)
	}

	// [VULNERABILITY] Outer map iteration — random zone order across peers.
	// Peer A may process Zone-A then Zone-B; Peer B may do Zone-B then Zone-A.
	for zoneID, itemUpdates := range updates {
		zoneKey := "ZONE_" + zoneID
		zoneJSON, err := ctx.GetStub().GetState(zoneKey)
		if err != nil {
			return fmt.Errorf("failed to read zone %s: %v", zoneID, err)
		}

		var zone ZoneInventory
		if zoneJSON != nil {
			json.Unmarshal(zoneJSON, &zone)
		} else {
			zone = ZoneInventory{ZoneID: zoneID, Items: make(map[string]int)}
		}

		// [VULNERABILITY] Inner map iteration — random item order within each zone.
		// The write-set contains per-zone PutState calls whose order depends on
		// both the outer AND inner map iteration sequences.
		for itemID, qtyDelta := range itemUpdates {
			zone.Items[itemID] += qtyDelta
			zone.ItemCount += qtyDelta

			// [VULNERABILITY] Per-item ledger entry written inside nested map loop.
			// The key order in the write set is doubly nondeterministic.
			itemKey := zoneID + "_ITEM_" + itemID
			itemRecord := fmt.Sprintf(`{"zone":"%s","item":"%s","qty":%d}`, zoneID, itemID, zone.Items[itemID])
			ctx.GetStub().PutState(itemKey, []byte(itemRecord))
		}

		updatedZoneJSON, _ := json.Marshal(zone)
		ctx.GetStub().PutState(zoneKey, updatedZoneJSON) // [VULNERABILITY] zone-level write also order-dependent
	}

	return nil
}

// GenerateZoneReport builds a summary across all zones
// [VULNERABILITY] Iterates a nested map structure to produce a report string.
// The report content varies across peers because string concatenation order
// follows the map iteration order.
func (w *WarehouseContract) GenerateZoneReport(ctx contractapi.TransactionContextInterface, zonesJSON string) error {
	var zoneData map[string]map[string]int
	err := json.Unmarshal([]byte(zonesJSON), &zoneData)
	if err != nil {
		return fmt.Errorf("invalid zone data: %v", err)
	}

	var reportParts []string

	// [VULNERABILITY] Outer iteration over zones — random order
	for zoneID, items := range zoneData {
		var itemParts []string

		// [VULNERABILITY] Inner iteration over items — random order
		for itemID, qty := range items {
			itemParts = append(itemParts, fmt.Sprintf("%s=%d", itemID, qty))
		}

		// [VULNERABILITY] The joined string depends on inner map iteration order
		zoneSummary := fmt.Sprintf("[%s: %s]", zoneID, strings.Join(itemParts, ","))
		reportParts = append(reportParts, zoneSummary)
	}

	// [VULNERABILITY] The final report string depends on outer map iteration order.
	// Peer A: "[Zone-A: X=5,Y=3][Zone-B: Z=1]"
	// Peer B: "[Zone-B: Z=1][Zone-A: Y=3,X=5]"
	report := strings.Join(reportParts, " | ")
	return ctx.GetStub().PutState("WAREHOUSE_REPORT", []byte(report))
}

// MergeZones transfers all items from source zones into a target zone
// [VULNERABILITY] Iterates over a map of source zone IDs to merge their items.
func (w *WarehouseContract) MergeZones(ctx contractapi.TransactionContextInterface, targetZoneID string, sourceZoneIDs []string) error {
	merged := make(map[string]int)

	for _, srcID := range sourceZoneIDs {
		zoneKey := "ZONE_" + srcID
		zoneJSON, err := ctx.GetStub().GetState(zoneKey)
		if err != nil || zoneJSON == nil {
			continue
		}

		var zone ZoneInventory
		json.Unmarshal(zoneJSON, &zone)

		// [VULNERABILITY] Iterating zone.Items (a map) to accumulate counts.
		// The accumulated result is correct (addition is commutative), BUT
		// the subsequent per-item PutState calls below are order-dependent.
		for itemID, qty := range zone.Items {
			merged[itemID] += qty
		}

		ctx.GetStub().DelState(zoneKey) // delete source zone
	}

	// [VULNERABILITY] Writing merged items — iteration over the `merged` map.
	for itemID, totalQty := range merged {
		mergedKey := targetZoneID + "_ITEM_" + itemID
		record := fmt.Sprintf(`{"zone":"%s","item":"%s","qty":%d}`, targetZoneID, itemID, totalQty)
		ctx.GetStub().PutState(mergedKey, []byte(record)) // [VULNERABILITY] order-dependent write
	}

	mergedZone := ZoneInventory{ZoneID: targetZoneID, Items: merged, ItemCount: len(merged)}
	mergedJSON, _ := json.Marshal(mergedZone)
	return ctx.GetStub().PutState("ZONE_"+targetZoneID, mergedJSON)
}

func main() {
	chaincode, err := contractapi.NewChaincode(&WarehouseContract{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
