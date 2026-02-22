package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SupplyChainContract tracks shipment lifecycle events
type SupplyChainContract struct {
	contractapi.Contract
}

// Shipment represents a logistics shipment record
type Shipment struct {
	ID          string `json:"id"`
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Temperature string `json:"temperature"`
}

// CreateShipment registers a new shipment on the ledger
// [VULNERABILITY] Non-deterministic: time.Now() produces different values on each peer
func (s *SupplyChainContract) CreateShipment(ctx contractapi.TransactionContextInterface, id string, origin string, destination string) error {
	exists, err := s.ShipmentExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("shipment %s already exists", id)
	}

	// [VULNERABILITY] time.Now() is non-deterministic across endorsing peers.
	// Each peer executes this at a slightly different wall-clock time,
	// producing different CreatedAt values and causing endorsement mismatch.
	now := time.Now()

	shipment := Shipment{
		ID:          id,
		Origin:      origin,
		Destination: destination,
		Status:      "CREATED",
		CreatedAt:   now.Format(time.RFC3339),     // [VULNERABILITY] non-deterministic timestamp
		UpdatedAt:   now.Format(time.RFC3339Nano), // [VULNERABILITY] nanosecond precision makes mismatch worse
	}

	shipmentJSON, err := json.Marshal(shipment)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, shipmentJSON)
}

// UpdateShipmentStatus changes the shipment status with a timestamp
func (s *SupplyChainContract) UpdateShipmentStatus(ctx contractapi.TransactionContextInterface, id string, newStatus string) error {
	shipmentJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to read shipment %s: %v", id, err)
	}
	if shipmentJSON == nil {
		return fmt.Errorf("shipment %s does not exist", id)
	}

	var shipment Shipment
	err = json.Unmarshal(shipmentJSON, &shipment)
	if err != nil {
		return err
	}

	shipment.Status = newStatus
	// [VULNERABILITY] Another time.Now() call — same non-determinism problem.
	// Two endorsing peers will record different UpdatedAt values,
	// leading to different PutState payloads and endorsement failure.
	shipment.UpdatedAt = time.Now().Format(time.RFC3339)

	updatedJSON, err := json.Marshal(shipment)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, updatedJSON)
}

// RecordTemperature logs a temperature reading with the current time
func (s *SupplyChainContract) RecordTemperature(ctx contractapi.TransactionContextInterface, id string, tempCelsius string) error {
	shipmentJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to read shipment %s: %v", id, err)
	}
	if shipmentJSON == nil {
		return fmt.Errorf("shipment %s does not exist", id)
	}

	var shipment Shipment
	err = json.Unmarshal(shipmentJSON, &shipment)
	if err != nil {
		return err
	}

	// [VULNERABILITY] time.Now() used to build a composite value.
	// Different peers produce different Temperature strings,
	// causing the marshalled JSON to differ and endorsement to fail.
	shipment.Temperature = fmt.Sprintf("%s°C at %s", tempCelsius, time.Now().Format("15:04:05.000"))
	shipment.UpdatedAt = time.Now().Format(time.RFC3339) // [VULNERABILITY] yet another time.Now()

	updatedJSON, err := json.Marshal(shipment)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, updatedJSON)
}

// ShipmentExists checks whether a shipment with the given ID exists
func (s *SupplyChainContract) ShipmentExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	data, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return data != nil, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SupplyChainContract{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
