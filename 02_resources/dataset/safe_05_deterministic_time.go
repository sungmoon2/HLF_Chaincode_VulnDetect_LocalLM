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
// [SAFE PATTERN] time.Now() is called, but the value is immediately OVERWRITTEN
// with a deterministic value from GetTxTimestamp() before being stored on the
// ledger. The wall-clock time is used only for a local performance measurement
// (elapsed time logging). A model that sees "time.Now()" and "PutState" in the
// same function must trace the data flow to confirm that the time.Now() value
// does NOT reach the ledger write.
func (c *CertificateContract) IssueCertificate(ctx contractapi.TransactionContextInterface, certID string, holderName string, courseID string) error {
	exists, err := c.CertExists(ctx, certID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("certificate %s already exists", certID)
	}

	// [SAFE PATTERN] time.Now() captured for local elapsed-time measurement.
	// This variable is never written to the ledger.
	startLocal := time.Now()

	// [SAFE PATTERN] The ACTUAL timestamp for the ledger comes from GetTxTimestamp(),
	// which is the block timestamp set by the ordering service — identical on all peers.
	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}
	deterministicTime := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos))

	cert := Certificate{
		CertID:     certID,
		HolderName: holderName,
		CourseID:   courseID,
		IssuedAt:   deterministicTime.Format(time.RFC3339),                          // [SAFE PATTERN] deterministic
		ExpiresAt:  deterministicTime.AddDate(2, 0, 0).Format(time.RFC3339),         // [SAFE PATTERN] deterministic + 2 years
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

	// [SAFE PATTERN] Elapsed time used only for operator diagnostics — never on ledger.
	elapsed := time.Since(startLocal)
	fmt.Printf("[PERF] IssueCertificate(%s) took %v\n", certID, elapsed)

	return nil
}

// RenewCertificate extends the expiration date of a certificate
// [SAFE PATTERN] time.Now() is called and assigned to a variable, but that
// variable is reassigned to a deterministic value before reaching PutState.
// A shallow pattern-match on "now := time.Now()" would trigger a false positive;
// a model must track that `now` is overwritten before the ledger write.
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

	// [SAFE PATTERN] now := time.Now() — looks suspicious, but...
	now := time.Now()
	fmt.Printf("[LOG] RenewCertificate called at local time: %s\n", now.Format(time.RFC3339))

	// [SAFE PATTERN] ...the variable `now` is NOT used for the ledger write.
	// Instead, the deterministic transaction timestamp replaces it entirely.
	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return err
	}
	// Overwrite `now` with deterministic value — the original time.Now() is discarded
	now = time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos))

	cert.ExpiresAt = now.AddDate(2, 0, 0).Format(time.RFC3339) // [SAFE PATTERN] uses overwritten deterministic `now`
	cert.Status = "RENEWED"

	updatedJSON, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(certID, updatedJSON)
}

// RevokeCertificate marks a certificate as revoked
// [SAFE PATTERN] time.Now() is used only in the fmt.Printf log line.
// The ledger write uses a constant status string.
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

	// [SAFE PATTERN] time.Now() only for console log — does not affect ledger
	fmt.Printf("[REVOKE] Certificate %s revoked at %s. Reason: %s\n",
		certID, time.Now().Format(time.RFC3339), reason)

	cert.Status = "REVOKED" // [SAFE PATTERN] hardcoded constant string

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
