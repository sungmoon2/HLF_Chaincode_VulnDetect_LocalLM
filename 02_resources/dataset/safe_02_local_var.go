package main

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// [SAFE PATTERN] These package-level variables are constants or read-only configuration.
// They are compiled into the binary or set at init time with fixed values.
// Every peer runs the same binary, so these values are identical everywhere —
// no divergence, no endorsement mismatch.
const (
	maxTransferLimit  = 10000.0           // [SAFE PATTERN] compile-time constant
	defaultCurrency   = "USD"             // [SAFE PATTERN] compile-time constant
	feeRateBasisPts   = 30                // [SAFE PATTERN] 0.30% fee, compile-time constant
	decimalPrecision  = 2                 // [SAFE PATTERN] compile-time constant
)

// [SAFE PATTERN] Read-only package-level variable initialised once.
// All peers load the same chaincode image, so this slice is identical on every peer.
// It is never modified after init — purely used for validation lookups.
var supportedCurrencies = []string{"USD", "EUR", "JPY", "KRW", "GBP"}

// PaymentContract manages payment processing
type PaymentContract struct {
	contractapi.Contract
}

// Payment represents a payment record
type Payment struct {
	ID       string  `json:"id"`
	From     string  `json:"from"`
	To       string  `json:"to"`
	Amount   float64 `json:"amount"`
	Fee      float64 `json:"fee"`
	Currency string  `json:"currency"`
	Status   string  `json:"status"`
}

// CreatePayment processes a new payment with fee calculation
// [SAFE PATTERN] All variables used in the write set are either:
//   - derived from function arguments (deterministic input), or
//   - derived from compile-time constants (identical across all peers).
// No global mutable state is read or written.
func (p *PaymentContract) CreatePayment(ctx contractapi.TransactionContextInterface, id string, from string, to string, amount float64, currency string) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	// [SAFE PATTERN] maxTransferLimit is a const — same on every peer.
	if amount > maxTransferLimit {
		return fmt.Errorf("amount %.2f exceeds maximum transfer limit %.2f", amount, maxTransferLimit)
	}

	// [SAFE PATTERN] Validating against a read-only package-level slice.
	// supportedCurrencies is never mutated, so the check is deterministic.
	if !isCurrencySupported(currency) {
		return fmt.Errorf("currency %s is not supported", currency)
	}

	// [SAFE PATTERN] Fee computed from a compile-time constant (feeRateBasisPts = 30).
	// The calculation is purely arithmetic on deterministic inputs.
	fee := roundToDecimal(amount * float64(feeRateBasisPts) / 10000.0)

	payment := Payment{
		ID:       id,
		From:     from,
		To:       to,
		Amount:   amount,
		Fee:      fee,      // [SAFE PATTERN] derived from const + argument
		Currency: currency,
		Status:   "COMPLETED",
	}

	paymentJSON, err := json.Marshal(payment)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, paymentJSON)
}

// QueryPayment retrieves a payment record
func (p *PaymentContract) QueryPayment(ctx contractapi.TransactionContextInterface, id string) (*Payment, error) {
	paymentJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read payment %s: %v", id, err)
	}
	if paymentJSON == nil {
		return nil, fmt.Errorf("payment %s does not exist", id)
	}

	var payment Payment
	err = json.Unmarshal(paymentJSON, &payment)
	if err != nil {
		return nil, err
	}

	return &payment, nil
}

// RefundPayment reverses a completed payment
// [SAFE PATTERN] Uses a local variable `refundID` built deterministically
// from the function argument. No global state is read or mutated.
func (p *PaymentContract) RefundPayment(ctx contractapi.TransactionContextInterface, paymentID string) error {
	paymentJSON, err := ctx.GetStub().GetState(paymentID)
	if err != nil {
		return fmt.Errorf("failed to read payment %s: %v", paymentID, err)
	}
	if paymentJSON == nil {
		return fmt.Errorf("payment %s does not exist", paymentID)
	}

	var payment Payment
	err = json.Unmarshal(paymentJSON, &payment)
	if err != nil {
		return err
	}

	if payment.Status != "COMPLETED" {
		return fmt.Errorf("payment %s cannot be refunded (status: %s)", paymentID, payment.Status)
	}

	payment.Status = "REFUNDED"

	updatedJSON, err := json.Marshal(payment)
	if err != nil {
		return err
	}
	ctx.GetStub().PutState(paymentID, updatedJSON)

	// [SAFE PATTERN] refundID is a local variable derived deterministically
	// from the input argument — identical on all peers.
	refundID := "REFUND_" + paymentID

	refundRecord := Payment{
		ID:       refundID,
		From:     payment.To,
		To:       payment.From,
		Amount:   payment.Amount,
		Fee:      0,
		Currency: payment.Currency,
		Status:   "REFUND",
	}

	refundJSON, err := json.Marshal(refundRecord)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(refundID, refundJSON)
}

// GetFeeSchedule returns the current fee configuration
// [SAFE PATTERN] Returns only compile-time constants — same on every peer.
func (p *PaymentContract) GetFeeSchedule(ctx contractapi.TransactionContextInterface) (string, error) {
	schedule := map[string]interface{}{
		"feeRateBasisPoints": feeRateBasisPts,
		"maxTransferLimit":   maxTransferLimit,
		"defaultCurrency":    defaultCurrency,
		"supportedCurrencies": supportedCurrencies,
	}

	scheduleJSON, err := json.Marshal(schedule)
	if err != nil {
		return "", err
	}

	return string(scheduleJSON), nil
}

// isCurrencySupported checks if a currency code is in the allowed list.
// [SAFE PATTERN] supportedCurrencies is a read-only package-level slice.
func isCurrencySupported(currency string) bool {
	for _, c := range supportedCurrencies {
		if c == currency {
			return true
		}
	}
	return false
}

// roundToDecimal rounds a float to the configured decimal precision.
// [SAFE PATTERN] Uses decimalPrecision const and math.Round — deterministic.
func roundToDecimal(val float64) float64 {
	pow := math.Pow(10, float64(decimalPrecision))
	return math.Round(val*pow) / pow
}

func main() {
	chaincode, err := contractapi.NewChaincode(&PaymentContract{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
