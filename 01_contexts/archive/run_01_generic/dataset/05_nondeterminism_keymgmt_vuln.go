package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SupplyChainChaincode manages supply chain with traceability
type SupplyChainChaincode struct {
	contractapi.Contract
}

// Shipment represents a tracked shipment
type Shipment struct {
	ShipmentID   string  `json:"shipmentId"`
	Origin       string  `json:"origin"`
	Destination  string  `json:"destination"`
	Status       string  `json:"status"`
	Temperature  float64 `json:"temperature"`
	Timestamp    string  `json:"timestamp"`
	TrackingCode string  `json:"trackingCode"`
	EncryptionKey string `json:"encryptionKey"`
}

// VULNERABILITY: Non-deterministic operation - time.Now()
// Each peer will execute at different times, producing different results
// This will cause endorsement mismatch and transaction failure
func (s *SupplyChainChaincode) CreateShipment(ctx contractapi.TransactionContextInterface, shipmentID string, origin string, destination string) error {
	// VULNERABILITY: time.Now() returns different values on different peers
	// Fabric requires deterministic execution across all endorsing peers
	currentTime := time.Now().Format(time.RFC3339)

	shipment := Shipment{
		ShipmentID:  shipmentID,
		Origin:      origin,
		Destination: destination,
		Status:      "created",
		Temperature: 0.0,
		Timestamp:   currentTime, // VULNERABILITY: Non-deterministic
		TrackingCode: "",
		EncryptionKey: "",
	}

	shipmentJSON, err := json.Marshal(shipment)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(shipmentID, shipmentJSON)
}

// VULNERABILITY: Non-deterministic random number generation
// crypto/rand produces different values on each peer
func (s *SupplyChainChaincode) GenerateTrackingCode(ctx contractapi.TransactionContextInterface, shipmentID string) (string, error) {
	shipmentJSON, err := ctx.GetStub().GetState(shipmentID)
	if err != nil {
		return "", err
	}
	if shipmentJSON == nil {
		return "", fmt.Errorf("shipment %s not found", shipmentID)
	}

	var shipment Shipment
	json.Unmarshal(shipmentJSON, &shipment)

	// VULNERABILITY: Random number generation is non-deterministic
	// Each endorsing peer will generate a different tracking code
	// Causing endorsement policy failure
	randomBytes := make([]byte, 16)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	trackingCode := hex.EncodeToString(randomBytes)

	shipment.TrackingCode = trackingCode
	shipment.Timestamp = time.Now().String() // VULNERABILITY: Another time.Now()

	updatedJSON, _ := json.Marshal(shipment)
	ctx.GetStub().PutState(shipmentID, updatedJSON)

	return trackingCode, nil
}

// VULNERABILITY: External HTTP call - non-deterministic and unreliable
// Network calls produce different results on different peers
func (s *SupplyChainChaincode) UpdateTemperatureFromSensor(ctx contractapi.TransactionContextInterface, shipmentID string, sensorURL string) error {
	shipmentJSON, err := ctx.GetStub().GetState(shipmentID)
	if err != nil {
		return err
	}

	var shipment Shipment
	json.Unmarshal(shipmentJSON, &shipment)

	// VULNERABILITY: HTTP call from chaincode is non-deterministic
	// Different peers may get different responses or timeouts
	// External service may be unavailable for some peers
	resp, err := http.Get(sensorURL)
	if err != nil {
		return fmt.Errorf("sensor read failed: %v", err)
	}
	defer resp.Body.Close()

	var sensorData struct {
		Temperature float64 `json:"temperature"`
	}
	json.NewDecoder(resp.Body).Decode(&sensorData)

	shipment.Temperature = sensorData.Temperature
	shipment.Timestamp = time.Now().Format(time.RFC3339) // VULNERABILITY: time.Now() again

	updatedJSON, _ := json.Marshal(shipment)
	return ctx.GetStub().PutState(shipmentID, updatedJSON)
}

// VULNERABILITY: Reading environment variables - non-deterministic
// Environment may differ across peers
func (s *SupplyChainChaincode) GetSystemConfig(ctx contractapi.TransactionContextInterface) (string, error) {
	// VULNERABILITY: os.Getenv returns peer-specific values
	// Different peers will have different environment configurations
	dbHost := os.Getenv("DB_HOST")
	apiKey := os.Getenv("API_KEY") // VULNERABILITY: Also leaks sensitive config

	config := map[string]string{
		"db_host": dbHost,
		"api_key": apiKey,
		"node_id": os.Getenv("CORE_PEER_ID"),
	}

	configJSON, _ := json.Marshal(config)
	return string(configJSON), nil
}

// VULNERABILITY: Insecure key generation and storage
// Encryption key stored in plaintext in world state
func (s *SupplyChainChaincode) GenerateEncryptionKey(ctx contractapi.TransactionContextInterface, shipmentID string) error {
	shipmentJSON, err := ctx.GetStub().GetState(shipmentID)
	if err != nil {
		return err
	}

	var shipment Shipment
	json.Unmarshal(shipmentJSON, &shipment)

	// VULNERABILITY 1: Non-deterministic key generation
	keyBytes := make([]byte, 32)
	rand.Read(keyBytes)

	// VULNERABILITY 2: Encryption key stored in plaintext in world state
	// All channel participants can read this key
	shipment.EncryptionKey = hex.EncodeToString(keyBytes)

	updatedJSON, _ := json.Marshal(shipment)
	return ctx.GetStub().PutState(shipmentID, updatedJSON)
}

// VULNERABILITY: Weak random using predictable seed
func (s *SupplyChainChaincode) AssignRandomInspector(ctx contractapi.TransactionContextInterface, shipmentID string) (string, error) {
	inspectors := []string{"inspector_A", "inspector_B", "inspector_C", "inspector_D"}

	// VULNERABILITY: Using crypto/rand for selection but still non-deterministic
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(inspectors))))
	if err != nil {
		return "", err
	}

	selectedInspector := inspectors[n.Int64()]

	// Store assignment
	assignmentKey := fmt.Sprintf("inspection_%s", shipmentID)
	assignment := map[string]string{
		"shipmentId": shipmentID,
		"inspector":  selectedInspector,
		"assignedAt": time.Now().Format(time.RFC3339), // VULNERABILITY: time.Now()
	}

	assignmentJSON, _ := json.Marshal(assignment)
	ctx.GetStub().PutState(assignmentKey, assignmentJSON)

	return selectedInspector, nil
}

// VULNERABILITY: Map iteration order is non-deterministic in Go
func (s *SupplyChainChaincode) GenerateReport(ctx contractapi.TransactionContextInterface, shipmentIDs []string) (string, error) {
	report := make(map[string]interface{})

	for _, id := range shipmentIDs {
		shipmentJSON, err := ctx.GetStub().GetState(id)
		if err != nil {
			continue
		}
		var shipment Shipment
		json.Unmarshal(shipmentJSON, &shipment)
		report[id] = shipment
	}

	// VULNERABILITY: Map iteration in Go is randomized
	// json.Marshal of a map produces consistent key ordering,
	// but any manual iteration over the map will differ across peers
	var reportStr string
	for k, v := range report {
		data, _ := json.Marshal(v)
		reportStr += fmt.Sprintf("%s: %s\n", k, string(data))
	}

	// Store report with non-deterministic content
	ctx.GetStub().PutState("latest_report", []byte(reportStr))

	return reportStr, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SupplyChainChaincode{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
