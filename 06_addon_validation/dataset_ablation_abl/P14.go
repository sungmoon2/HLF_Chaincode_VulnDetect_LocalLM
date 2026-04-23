package main

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	maxTransferLimit  = 10000.0
	defaultCurrency   = "USD"
	feeRateBasisPts   = 30
	decimalPrecision  = 2
)

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
func (p *PaymentContract) CreatePayment(ctx contractapi.TransactionContextInterface, id string, from string, to string, amount float64, currency string) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	if amount > maxTransferLimit {
		return fmt.Errorf("amount %.2f exceeds maximum transfer limit %.2f", amount, maxTransferLimit)
	}

	if !isCurrencySupported(currency) {
		return fmt.Errorf("currency %s is not supported", currency)
	}

	fee := roundToDecimal(amount * float64(feeRateBasisPts) / 10000.0)

	payment := Payment{
		ID:       id,
		From:     from,
		To:       to,
		Amount:   amount,
		Fee:      fee,
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
func isCurrencySupported(currency string) bool {
	for _, c := range supportedCurrencies {
		if c == currency {
			return true
		}
	}
	return false
}

// roundToDecimal rounds a float to the configured decimal precision.
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
