package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// MedicalRecordChaincode manages patient medical records
type MedicalRecordChaincode struct {
	contractapi.Contract
}

// PatientRecord contains sensitive medical data
type PatientRecord struct {
	PatientID    string   `json:"patientId"`
	Name         string   `json:"name"`
	SSN          string   `json:"ssn"`          // Sensitive
	Diagnosis    string   `json:"diagnosis"`     // Sensitive
	Medications  []string `json:"medications"`   // Sensitive
	InsuranceID  string   `json:"insuranceId"`   // Sensitive
	DoctorNotes  string   `json:"doctorNotes"`   // Sensitive
	AccessLevel  string   `json:"accessLevel"`
}

// AuditLog tracks who accessed records
type AuditLog struct {
	RecordID  string `json:"recordId"`
	AccessedBy string `json:"accessedBy"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
}

// VULNERABILITY: Private data stored on public channel state
// Sensitive medical data should use Private Data Collections (PDC)
// but is stored directly in the public world state
func (m *MedicalRecordChaincode) CreateRecord(ctx contractapi.TransactionContextInterface, patientID string, name string, ssn string, diagnosis string) error {
	record := PatientRecord{
		PatientID:   patientID,
		Name:        name,
		SSN:         ssn,       // VULNERABILITY: SSN in public state
		Diagnosis:   diagnosis, // VULNERABILITY: Diagnosis in public state
		Medications: []string{},
		InsuranceID: "",
		DoctorNotes: "",
		AccessLevel: "restricted",
	}

	// VULNERABILITY: All fields including sensitive data stored in public state
	// Every peer in the channel can read SSN, diagnosis, etc.
	// Should use: ctx.GetStub().PutPrivateData("medicalPDC", patientID, recordJSON)
	recordJSON, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(patientID, recordJSON)
}

// VULNERABILITY: Returns full record including all sensitive fields
// No field-level access control - exposes SSN, diagnosis to all callers
func (m *MedicalRecordChaincode) GetRecord(ctx contractapi.TransactionContextInterface, patientID string) (*PatientRecord, error) {
	recordJSON, err := ctx.GetStub().GetState(patientID)
	if err != nil {
		return nil, fmt.Errorf("failed to read record: %v", err)
	}
	if recordJSON == nil {
		return nil, fmt.Errorf("patient %s not found", patientID)
	}

	var record PatientRecord
	err = json.Unmarshal(recordJSON, &record)
	if err != nil {
		return nil, err
	}

	// VULNERABILITY: Returns complete record with all sensitive fields
	// No redaction of SSN, diagnosis, medications, etc.
	// Any organization in the channel can call this and see everything
	return &record, nil
}

// VULNERABILITY: Sensitive data exposed in events
// Events are visible to all channel participants
func (m *MedicalRecordChaincode) UpdateDiagnosis(ctx contractapi.TransactionContextInterface, patientID string, newDiagnosis string) error {
	recordJSON, err := ctx.GetStub().GetState(patientID)
	if err != nil {
		return err
	}
	if recordJSON == nil {
		return fmt.Errorf("patient %s not found", patientID)
	}

	var record PatientRecord
	json.Unmarshal(recordJSON, &record)

	record.Diagnosis = newDiagnosis

	updatedJSON, _ := json.Marshal(record)
	ctx.GetStub().PutState(patientID, updatedJSON)

	// VULNERABILITY: Emitting event with sensitive medical data
	// Events are broadcast to all peers and clients on the channel
	eventPayload := fmt.Sprintf(`{"patientId":"%s","diagnosis":"%s","ssn":"%s"}`,
		patientID, newDiagnosis, record.SSN)
	ctx.GetStub().SetEvent("DiagnosisUpdate", []byte(eventPayload))

	return nil
}

// VULNERABILITY: Logging sensitive data that ends up in peer logs
func (m *MedicalRecordChaincode) AddMedication(ctx contractapi.TransactionContextInterface, patientID string, medication string) error {
	recordJSON, err := ctx.GetStub().GetState(patientID)
	if err != nil {
		return err
	}

	var record PatientRecord
	json.Unmarshal(recordJSON, &record)

	// VULNERABILITY: Sensitive data in return messages visible in transaction logs
	fmt.Printf("Adding medication %s for patient %s (SSN: %s)\n",
		medication, patientID, record.SSN)

	record.Medications = append(record.Medications, medication)

	updatedJSON, _ := json.Marshal(record)
	return ctx.GetStub().PutState(patientID, updatedJSON)
}

// VULNERABILITY: Composite key leaks information through key structure
func (m *MedicalRecordChaincode) CreateDiagnosisIndex(ctx contractapi.TransactionContextInterface, patientID string) error {
	recordJSON, err := ctx.GetStub().GetState(patientID)
	if err != nil {
		return err
	}

	var record PatientRecord
	json.Unmarshal(recordJSON, &record)

	// VULNERABILITY: Diagnosis information leaked through composite key name
	// Anyone who can enumerate keys can see which patients have which diagnoses
	indexKey, err := ctx.GetStub().CreateCompositeKey("diagnosis~patient", []string{record.Diagnosis, patientID})
	if err != nil {
		return err
	}

	// VULNERABILITY: Even a nil value exposes the key structure
	return ctx.GetStub().PutState(indexKey, []byte{0x00})
}

// VULNERABILITY: GetHistoryForKey exposes all previous values including deleted records
func (m *MedicalRecordChaincode) GetRecordHistory(ctx contractapi.TransactionContextInterface, patientID string) (string, error) {
	historyIterator, err := ctx.GetStub().GetHistoryForKey(patientID)
	if err != nil {
		return "", err
	}
	defer historyIterator.Close()

	var history []map[string]interface{}
	for historyIterator.HasNext() {
		modification, err := historyIterator.Next()
		if err != nil {
			return "", err
		}

		// VULNERABILITY: Full historical records exposed
		// Even if a record was "deleted", the history shows all previous values
		// Previous SSNs, old diagnoses, removed medications all visible
		entry := map[string]interface{}{
			"txId":      modification.TxId,
			"value":     string(modification.Value),
			"isDelete":  modification.IsDelete,
		}
		history = append(history, entry)
	}

	historyJSON, _ := json.Marshal(history)
	return string(historyJSON), nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&MedicalRecordChaincode{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
