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
func getCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

// formatTimestampNano wraps time.Now() at nanosecond precision.
func formatTimestampNano() string {
	now := time.Now()
	return now.Format(time.RFC3339Nano)
}

// buildTimestampedNote produces a composite string containing a nondeterministic timestamp.
func buildTimestampedNote(prefix string) string {
	return fmt.Sprintf("%s (recorded at %s)", prefix, getCurrentTimestamp())
}

// CreateRecord registers a new medical record on the ledger
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
		RecordedAt: getCurrentTimestamp(),
		LastChecked: formatTimestampNano(),
	}

	recordJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(recordID, recordJSON)
}

// UpdateDiagnosis modifies the diagnosis and refreshes the timestamp
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

	record.Diagnosis = buildTimestampedNote(newDiagnosis)
	record.LastChecked = getCurrentTimestamp()

	updatedJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(recordID, updatedJSON)
}

// AddCheckupNote appends a timestamped note to the patient's composite key
func (m *MedicalRecordContract) AddCheckupNote(ctx contractapi.TransactionContextInterface, recordID string, note string) error {
	_, err := ctx.GetStub().GetState(recordID)
	if err != nil {
		return fmt.Errorf("failed to read record %s: %v", recordID, err)
	}

	compositeNote := buildTimestampedNote(note)
	noteKey := "NOTE_" + recordID + "_" + getCurrentTimestamp()

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
