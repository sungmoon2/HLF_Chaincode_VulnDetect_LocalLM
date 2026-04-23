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
func (w *WarehouseContract) BatchUpdateInventory(ctx contractapi.TransactionContextInterface, updateJSON string) error {
	// Outer map: zoneID → inner map
	// Inner map: itemID → quantity delta
	var updates map[string]map[string]int
	err := json.Unmarshal([]byte(updateJSON), &updates)
	if err != nil {
		return fmt.Errorf("invalid update JSON: %v", err)
	}

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

		for itemID, qtyDelta := range itemUpdates {
			zone.Items[itemID] += qtyDelta
			zone.ItemCount += qtyDelta

			itemKey := zoneID + "_ITEM_" + itemID
			itemRecord := fmt.Sprintf(`{"zone":"%s","item":"%s","qty":%d}`, zoneID, itemID, zone.Items[itemID])
			ctx.GetStub().PutState(itemKey, []byte(itemRecord))
		}

		updatedZoneJSON, _ := json.Marshal(zone)
		ctx.GetStub().PutState(zoneKey, updatedZoneJSON)
	}

	return nil
}

// GenerateZoneReport builds a summary across all zones
func (w *WarehouseContract) GenerateZoneReport(ctx contractapi.TransactionContextInterface, zonesJSON string) error {
	var zoneData map[string]map[string]int
	err := json.Unmarshal([]byte(zonesJSON), &zoneData)
	if err != nil {
		return fmt.Errorf("invalid zone data: %v", err)
	}

	var reportParts []string

	for zoneID, items := range zoneData {
		var itemParts []string

		for itemID, qty := range items {
			itemParts = append(itemParts, fmt.Sprintf("%s=%d", itemID, qty))
		}

		zoneSummary := fmt.Sprintf("[%s: %s]", zoneID, strings.Join(itemParts, ","))
		reportParts = append(reportParts, zoneSummary)
	}

	report := strings.Join(reportParts, " | ")
	return ctx.GetStub().PutState("WAREHOUSE_REPORT", []byte(report))
}

// MergeZones transfers all items from source zones into a target zone
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

		for itemID, qty := range zone.Items {
			merged[itemID] += qty
		}

		ctx.GetStub().DelState(zoneKey) // delete source zone
	}

	for itemID, totalQty := range merged {
		mergedKey := targetZoneID + "_ITEM_" + itemID
		record := fmt.Sprintf(`{"zone":"%s","item":"%s","qty":%d}`, targetZoneID, itemID, totalQty)
		ctx.GetStub().PutState(mergedKey, []byte(record))
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
