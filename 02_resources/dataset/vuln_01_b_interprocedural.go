package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// MedicalRecordContract manages patient records on the ledger
type MedicalRecordContract struct {
	contractapi.Contract
}

// MedicalRecord represents a patient medical event
type MedicalRecord struct {
	RecordID    string `json:"recordID"`
	PatientID   string `json:"patientID"`
	Diagnosis   string `json:"diagnosis"`
	Doctor      string `json:"doctor"`
	RecordedAt  string `json:"recordedAt"`
	LastChecked string `json:"lastChecked"`
}

// getCurrentTimestamp is a helper that returns a formatted wall-clock time.
// [VULNERABILITY] This wrapper hides time.Now() behind a function call.
// A model that only pattern-matches on "time.Now() near PutState" in the
// same function body may miss this interprocedural data flow.
func getCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

// formatTimestampNano wraps time.Now() at nanosecond precision.
// [VULNERABILITY] A second layer of indirection — the nondeterministic
// value is buried two calls deep from the ledger write site.
func formatTimestampNano() string {
	now := time.Now()
	return now.Format(time.RFC3339Nano)
}

// buildTimestampedNote produces a composite string containing a nondeterministic timestamp.
// [VULNERABILITY] The caller receives a string whose content varies across peers,
// but there is no direct time.Now() call at the call site — only inside this helper.
func buildTimestampedNote(prefix string) string {
	return fmt.Sprintf("%s (recorded at %s)", prefix, getCurrentTimestamp())
}

// CreateRecord registers a new medical record on the ledger
// [VULNERABILITY] The nondeterministic timestamp comes from getCurrentTimestamp(),
// not from an in-line time.Now() call. A model must trace the return value
// of getCurrentTimestamp() → time.Now() to detect the vulnerability.
func (m *MedicalRecordContract) CreateRecord(ctx contractapi.TransactionContextInterface, recordID string, patientID string, diagnosis string, doctor string) error {
	exists, err := m.RecordExists(ctx, recordID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("record %s already exists", recordID)
	}

	record := MedicalRecord{
		RecordID:   recordID,
		PatientID:  patientID,
		Diagnosis:  diagnosis,
		Doctor:     doctor,
		RecordedAt: getCurrentTimestamp(),  // [VULNERABILITY] interprocedural nondeterminism
		LastChecked: formatTimestampNano(), // [VULNERABILITY] another hidden time.Now()
	}

	recordJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(recordID, recordJSON)
}

// UpdateDiagnosis modifies the diagnosis and refreshes the timestamp
// [VULNERABILITY] The nondeterministic value flows through buildTimestampedNote()
// which internally calls getCurrentTimestamp() which calls time.Now().
// Three levels of indirection before reaching PutState.
func (m *MedicalRecordContract) UpdateDiagnosis(ctx contractapi.TransactionContextInterface, recordID string, newDiagnosis string) error {
	recordJSON, err := ctx.GetStub().GetState(recordID)
	if err != nil {
		return fmt.Errorf("failed to read record %s: %v", recordID, err)
	}
	if recordJSON == nil {
		return fmt.Errorf("record %s does not exist", recordID)
	}

	var record MedicalRecord
	err = json.Unmarshal(recordJSON, &record)
	if err != nil {
		return err
	}

	// [VULNERABILITY] buildTimestampedNote → getCurrentTimestamp → time.Now()
	// The resulting string differs across endorsing peers.
	record.Diagnosis = buildTimestampedNote(newDiagnosis)
	record.LastChecked = getCurrentTimestamp() // [VULNERABILITY] interprocedural

	updatedJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(recordID, updatedJSON)
}

// AddCheckupNote appends a timestamped note to the patient's composite key
// [VULNERABILITY] Uses the helper to embed a nondeterministic timestamp
// into a ledger value via a separate composite key.
func (m *MedicalRecordContract) AddCheckupNote(ctx contractapi.TransactionContextInterface, recordID string, note string) error {
	_, err := ctx.GetStub().GetState(recordID)
	if err != nil {
		return fmt.Errorf("failed to read record %s: %v", recordID, err)
	}

	// [VULNERABILITY] buildTimestampedNote hides time.Now() two levels deep
	compositeNote := buildTimestampedNote(note)
	noteKey := "NOTE_" + recordID + "_" + getCurrentTimestamp() // [VULNERABILITY] nondeterministic key

	return ctx.GetStub().PutState(noteKey, []byte(compositeNote))
}

// RecordExists checks whether a record with the given ID exists
func (m *MedicalRecordContract) RecordExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	data, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return data != nil, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&MedicalRecordContract{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
