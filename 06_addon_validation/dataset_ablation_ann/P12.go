package main

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// PackageTrackingContract manages package delivery tracking
type PackageTrackingContract struct {
	contractapi.Contract
}

// Package represents a tracked package
type Package struct {
	TrackingID  string `json:"trackingID"`
	Sender      string `json:"sender"`
	Receiver    string `json:"receiver"`
	Weight      int    `json:"weight"`
	Status      string `json:"status"`
	WarehouseID string `json:"warehouseID"`
}

// RegisterPackage creates a new package record on the ledger
// [SAFE PATTERN] math/rand is used ONLY to generate a trace ID for console logging.
// The random value never reaches PutState or any ledger write.
// All data written to the ledger comes from deterministic function arguments.
func (p *PackageTrackingContract) RegisterPackage(ctx contractapi.TransactionContextInterface, trackingID string, sender string, receiver string, weight int) error {
	exists, err := p.PackageExists(ctx, trackingID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("package %s already exists", trackingID)
	}

	// [SAFE PATTERN] rand.Intn generates a random trace ID for operator debugging.
	// This value is printed to stdout only — it is never stored on the ledger.
	// Even though each peer generates a different traceID, the write sets remain
	// identical because traceID does not enter any PutState call.
	traceID := rand.Intn(999999)
	fmt.Printf("[TRACE-%06d] RegisterPackage called: id=%s, from=%s, to=%s\n",
		traceID, trackingID, sender, receiver)

	pkg := Package{
		TrackingID:  trackingID,     // [SAFE PATTERN] from function argument — deterministic
		Sender:      sender,         // [SAFE PATTERN] from function argument
		Receiver:    receiver,       // [SAFE PATTERN] from function argument
		Weight:      weight,         // [SAFE PATTERN] from function argument
		Status:      "REGISTERED",   // [SAFE PATTERN] hardcoded constant
		WarehouseID: "WH_INTAKE",    // [SAFE PATTERN] hardcoded constant
	}

	pkgJSON, err := json.Marshal(pkg)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(trackingID, pkgJSON)
}

// TransferPackage moves a package to a new warehouse
// [SAFE PATTERN] rand is used for a log correlation ID only.
// The ledger write uses only the function arguments (deterministic).
func (p *PackageTrackingContract) TransferPackage(ctx contractapi.TransactionContextInterface, trackingID string, newWarehouseID string) error {
	pkgJSON, err := ctx.GetStub().GetState(trackingID)
	if err != nil {
		return fmt.Errorf("failed to read package %s: %v", trackingID, err)
	}
	if pkgJSON == nil {
		return fmt.Errorf("package %s does not exist", trackingID)
	}

	var pkg Package
	err = json.Unmarshal(pkgJSON, &pkg)
	if err != nil {
		return err
	}

	// [SAFE PATTERN] Random correlation ID for log tracing.
	// Operators can grep logs by this ID to correlate peer-side events.
	// The ID varies per peer, but that's expected for local diagnostics.
	corrID := fmt.Sprintf("CORR-%08d", rand.Intn(99999999))
	fmt.Printf("[%s] TransferPackage: %s from %s to %s\n",
		corrID, trackingID, pkg.WarehouseID, newWarehouseID)

	pkg.WarehouseID = newWarehouseID // [SAFE PATTERN] from function argument
	pkg.Status = "IN_TRANSIT"       // [SAFE PATTERN] hardcoded constant

	updatedJSON, err := json.Marshal(pkg)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(trackingID, updatedJSON)
}

// DeliverPackage marks a package as delivered
// [SAFE PATTERN] rand used only to select a random motivational log message.
// The actual status written to ledger is a deterministic constant string.
func (p *PackageTrackingContract) DeliverPackage(ctx contractapi.TransactionContextInterface, trackingID string) error {
	pkgJSON, err := ctx.GetStub().GetState(trackingID)
	if err != nil {
		return fmt.Errorf("failed to read package %s: %v", trackingID, err)
	}
	if pkgJSON == nil {
		return fmt.Errorf("package %s does not exist", trackingID)
	}

	var pkg Package
	err = json.Unmarshal(pkgJSON, &pkg)
	if err != nil {
		return err
	}

	// [SAFE PATTERN] Random selection of a log message — purely cosmetic.
	// The logged message varies per peer, but it never enters the write set.
	messages := []string{
		"Package delivered successfully!",
		"Delivery confirmed.",
		"Shipment complete.",
	}
	logMsg := messages[rand.Intn(len(messages))]
	fmt.Printf("[DELIVERY] %s: %s\n", trackingID, logMsg)

	pkg.Status = "DELIVERED" // [SAFE PATTERN] hardcoded constant — same on all peers

	updatedJSON, err := json.Marshal(pkg)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(trackingID, updatedJSON)
}

// PackageExists checks whether a package ID is on the ledger
func (p *PackageTrackingContract) PackageExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	data, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return data != nil, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&PackageTrackingContract{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
