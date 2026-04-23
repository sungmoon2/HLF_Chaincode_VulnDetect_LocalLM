package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// AuditLogContract tracks document approval workflows
type AuditLogContract struct {
	contractapi.Contract
}

// Document represents an approval record
type Document struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Author     string `json:"author"`
	Status     string `json:"status"`
	ApprovedBy string `json:"approvedBy"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

// CreateDocument registers a new document for approval
func (a *AuditLogContract) CreateDocument(ctx contractapi.TransactionContextInterface, id string, title string, author string) error {
	exists, err := a.DocumentExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("document %s already exists", id)
	}

	fmt.Printf("[%s] CreateDocument called: id=%s, title=%s\n",
		time.Now().Format(time.RFC3339), id, title)

	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}
	deterministicTime := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339)

	doc := Document{
		ID:        id,
		Title:     title,
		Author:    author,
		Status:    "PENDING",
		CreatedAt: deterministicTime,
		UpdatedAt: deterministicTime,
	}

	docJSON, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, docJSON)
}

// ApproveDocument marks a document as approved
func (a *AuditLogContract) ApproveDocument(ctx contractapi.TransactionContextInterface, id string, approver string) error {
	docJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to read document %s: %v", id, err)
	}
	if docJSON == nil {
		return fmt.Errorf("document %s does not exist", id)
	}

	var doc Document
	err = json.Unmarshal(docJSON, &doc)
	if err != nil {
		return err
	}

	if doc.Status != "PENDING" {
		return fmt.Errorf("document %s is not pending approval (current: %s)", id, doc.Status)
	}

	fmt.Printf("[%s] ApproveDocument: id=%s approved by %s\n",
		time.Now().Format("15:04:05.000"), id, approver)

	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}
	deterministicTime := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339)

	doc.Status = "APPROVED"
	doc.ApprovedBy = approver
	doc.UpdatedAt = deterministicTime

	updatedJSON, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, updatedJSON)
}

// QueryDocument retrieves a document from the ledger
func (a *AuditLogContract) QueryDocument(ctx contractapi.TransactionContextInterface, id string) (*Document, error) {
	start := time.Now()

	docJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read document %s: %v", id, err)
	}
	if docJSON == nil {
		return nil, fmt.Errorf("document %s does not exist", id)
	}

	var doc Document
	err = json.Unmarshal(docJSON, &doc)
	if err != nil {
		return nil, err
	}

	elapsed := time.Since(start)
	fmt.Printf("[PERF] QueryDocument(%s) took %v\n", id, elapsed)

	return &doc, nil
}

// DocumentExists checks whether a document ID is already on the ledger
func (a *AuditLogContract) DocumentExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	data, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return data != nil, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&AuditLogContract{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
