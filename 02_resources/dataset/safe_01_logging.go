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
// [SAFE PATTERN] time.Now() is used ONLY for console logging (fmt.Printf).
// The timestamp stored on the ledger comes from stub.GetTxTimestamp(),
// which is the block timestamp agreed upon by the ordering service —
// identical across all endorsing peers, so the write set is deterministic.
func (a *AuditLogContract) CreateDocument(ctx contractapi.TransactionContextInterface, id string, title string, author string) error {
	exists, err := a.DocumentExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("document %s already exists", id)
	}

	// [SAFE PATTERN] time.Now() used only for local console/log output.
	// This value is never written to the ledger, so peer divergence is harmless.
	fmt.Printf("[%s] CreateDocument called: id=%s, title=%s\n",
		time.Now().Format(time.RFC3339), id, title)

	// [SAFE PATTERN] GetTxTimestamp() returns the transaction timestamp from
	// the channel header, set by the client SDK and validated by the orderer.
	// All endorsing peers see the same value — fully deterministic.
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
		CreatedAt: deterministicTime, // [SAFE PATTERN] deterministic — from tx header
		UpdatedAt: deterministicTime,
	}

	docJSON, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, docJSON)
}

// ApproveDocument marks a document as approved
// [SAFE PATTERN] time.Now() appears here but is only printed to stdout.
// The ledger write uses GetTxTimestamp() for a deterministic timestamp.
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

	// [SAFE PATTERN] time.Now() only for operator-visible log line.
	// Never touches the write set.
	fmt.Printf("[%s] ApproveDocument: id=%s approved by %s\n",
		time.Now().Format("15:04:05.000"), id, approver)

	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}
	deterministicTime := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339)

	doc.Status = "APPROVED"
	doc.ApprovedBy = approver
	doc.UpdatedAt = deterministicTime // [SAFE PATTERN] deterministic timestamp

	updatedJSON, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, updatedJSON)
}

// QueryDocument retrieves a document from the ledger
// [SAFE PATTERN] time.Now() used to measure local query latency for diagnostics.
// Query functions do not produce write sets, so no consensus impact.
func (a *AuditLogContract) QueryDocument(ctx contractapi.TransactionContextInterface, id string) (*Document, error) {
	start := time.Now() // [SAFE PATTERN] local performance measurement only

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

	// [SAFE PATTERN] Elapsed time printed to console — never stored on ledger.
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
