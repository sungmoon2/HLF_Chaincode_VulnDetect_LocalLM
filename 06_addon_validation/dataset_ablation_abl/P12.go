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
func (p *PackageTrackingContract) RegisterPackage(ctx contractapi.TransactionContextInterface, trackingID string, sender string, receiver string, weight int) error {
	exists, err := p.PackageExists(ctx, trackingID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("package %s already exists", trackingID)
	}

	traceID := rand.Intn(999999)
	fmt.Printf("[TRACE-%06d] RegisterPackage called: id=%s, from=%s, to=%s\n",
		traceID, trackingID, sender, receiver)

	pkg := Package{
		TrackingID:  trackingID,
		Sender:      sender,
		Receiver:    receiver,
		Weight:      weight,
		Status:      "REGISTERED",
		WarehouseID: "WH_INTAKE",
	}

	pkgJSON, err := json.Marshal(pkg)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(trackingID, pkgJSON)
}

// TransferPackage moves a package to a new warehouse
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

	corrID := fmt.Sprintf("CORR-%08d", rand.Intn(99999999))
	fmt.Printf("[%s] TransferPackage: %s from %s to %s\n",
		corrID, trackingID, pkg.WarehouseID, newWarehouseID)

	pkg.WarehouseID = newWarehouseID
	pkg.Status = "IN_TRANSIT"

	updatedJSON, err := json.Marshal(pkg)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(trackingID, updatedJSON)
}

// DeliverPackage marks a package as delivered
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

	messages := []string{
		"Package delivered successfully!",
		"Delivery confirmed.",
		"Shipment complete.",
	}
	logMsg := messages[rand.Intn(len(messages))]
	fmt.Printf("[DELIVERY] %s: %s\n", trackingID, logMsg)

	pkg.Status = "DELIVERED"

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
