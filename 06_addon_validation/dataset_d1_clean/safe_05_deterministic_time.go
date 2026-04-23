package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// CertificateContract manages digital certificate issuance
type CertificateContract struct {
	contractapi.Contract
}

// Certificate represents a digital certificate record
type Certificate struct {
	CertID     string `json:"certID"`
	HolderName string `json:"holderName"`
	CourseID   string `json:"courseID"`
	IssuedAt   string `json:"issuedAt"`
	ExpiresAt  string `json:"expiresAt"`
	Status     string `json:"status"`
}

// IssueCertificate creates a new certificate on the ledger
func (c *CertificateContract) IssueCertificate(ctx contractapi.TransactionContextInterface, certID string, holderName string, courseID string) error {
	exists, err := c.CertExists(ctx, certID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("certificate %s already exists", certID)
	}

	startLocal := time.Now()

	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}
	deterministicTime := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos))

	cert := Certificate{
		CertID:     certID,
		HolderName: holderName,
		CourseID:   courseID,
		IssuedAt:   deterministicTime.Format(time.RFC3339),
		ExpiresAt:  deterministicTime.AddDate(2, 0, 0).Format(time.RFC3339),
		Status:     "ACTIVE",
	}

	certJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(certID, certJSON)
	if err != nil {
		return err
	}

	elapsed := time.Since(startLocal)
	fmt.Printf("[PERF] IssueCertificate(%s) took %v\n", certID, elapsed)

	return nil
}

// RenewCertificate extends the expiration date of a certificate
func (c *CertificateContract) RenewCertificate(ctx contractapi.TransactionContextInterface, certID string) error {
	certJSON, err := ctx.GetStub().GetState(certID)
	if err != nil {
		return fmt.Errorf("failed to read certificate %s: %v", certID, err)
	}
	if certJSON == nil {
		return fmt.Errorf("certificate %s does not exist", certID)
	}

	var cert Certificate
	err = json.Unmarshal(certJSON, &cert)
	if err != nil {
		return err
	}

	now := time.Now()
	fmt.Printf("[LOG] RenewCertificate called at local time: %s\n", now.Format(time.RFC3339))

	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return err
	}
	// Overwrite `now` with deterministic value — the original time.Now() is discarded
	now = time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos))

	cert.ExpiresAt = now.AddDate(2, 0, 0).Format(time.RFC3339)
	cert.Status = "RENEWED"

	updatedJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(certID, updatedJSON)
}

// RevokeCertificate marks a certificate as revoked
func (c *CertificateContract) RevokeCertificate(ctx contractapi.TransactionContextInterface, certID string, reason string) error {
	certJSON, err := ctx.GetStub().GetState(certID)
	if err != nil {
		return fmt.Errorf("failed to read certificate %s: %v", certID, err)
	}
	if certJSON == nil {
		return fmt.Errorf("certificate %s does not exist", certID)
	}

	var cert Certificate
	err = json.Unmarshal(certJSON, &cert)
	if err != nil {
		return err
	}

	fmt.Printf("[REVOKE] Certificate %s revoked at %s. Reason: %s\n",
		certID, time.Now().Format(time.RFC3339), reason)

	cert.Status = "REVOKED"

	updatedJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(certID, updatedJSON)
}

// CertExists checks whether a certificate with the given ID exists
func (c *CertificateContract) CertExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	data, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return data != nil, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&CertificateContract{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
