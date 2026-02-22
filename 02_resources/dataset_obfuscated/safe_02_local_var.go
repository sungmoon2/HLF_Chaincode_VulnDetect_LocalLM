package main

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	C1  = 10000.0           
	C2   = "USD"             
	C3   = 30                
	C4  = 2                 
)

var G1 = []string{"USD", "EUR", "JPY", "KRW", "GBP"}

type StructA struct {
	contractapi.Contract
}

type StructB struct {
	F1       string  `json:"f1"`
	F2     string  `json:"f2"`
	F3       string  `json:"f3"`
	F4   float64 `json:"f4"`
	F5      float64 `json:"f5"`
	F6 string  `json:"f6"`
	F7   string  `json:"f7"`
}

func (p *StructA) FuncA(ctx contractapi.TransactionContextInterface, v12 string, v13 string, v14 string, v15 float64, v16 string) error {
	if v15 <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	if v15 > C1 {
		return fmt.Errorf("amount %.2f exceeds maximum transfer limit %.2f", v15, C1)
	}

	if !FuncE(v16) {
		return fmt.Errorf("currency %s is not supported", v16)
	}

	v1 := FuncF(v15 * float64(C3) / 10000.0)

	v2 := StructB{
		F1:       v12,
		F2:     v13,
		F3:       v14,
		F4:   v15,
		F5:      v1,      
		F6: v16,
		F7:   "COMPLETED",
	}

	v3, err := json.Marshal(v2)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v12, v3)
}

func (p *StructA) FuncB(ctx contractapi.TransactionContextInterface, v12 string) (*StructB, error) {
	v3, err := ctx.GetStub().GetState(v12)
	if err != nil {
		return nil, fmt.Errorf("failed to read payment %s: %v", v12, err)
	}
	if v3 == nil {
		return nil, fmt.Errorf("payment %s does not exist", v12)
	}

	var v2 StructB
	err = json.Unmarshal(v3, &v2)
	if err != nil {
		return nil, err
	}

	return &v2, nil
}

func (p *StructA) FuncC(ctx contractapi.TransactionContextInterface, v17 string) error {
	v3, err := ctx.GetStub().GetState(v17)
	if err != nil {
		return fmt.Errorf("failed to read payment %s: %v", v17, err)
	}
	if v3 == nil {
		return fmt.Errorf("payment %s does not exist", v17)
	}

	var v2 StructB
	err = json.Unmarshal(v3, &v2)
	if err != nil {
		return err
	}

	if v2.F7 != "COMPLETED" {
		return fmt.Errorf("payment %s cannot be refunded (status: %s)", v17, v2.F7)
	}

	v2.F7 = "REFUNDED"

	v4, err := json.Marshal(v2)
	if err != nil {
		return err
	}
	ctx.GetStub().PutState(v17, v4)

	v5 := "REFUND_" + v17

	v6 := StructB{
		F1:       v5,
		F2:     v2.F3,
		F3:       v2.F2,
		F4:   v2.F4,
		F5:      0,
		F6: v2.F6,
		F7:   "REFUND",
	}

	v7, err := json.Marshal(v6)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v5, v7)
}

func (p *StructA) FuncD(ctx contractapi.TransactionContextInterface) (string, error) {
	v8 := map[string]interface{}{
		"feeRateBasisPoints": C3,
		"maxTransferLimit":   C1,
		"defaultCurrency":    C2,
		"supportedCurrencies": G1,
	}

	v9, err := json.Marshal(v8)
	if err != nil {
		return "", err
	}

	return string(v9), nil
}

func FuncE(v16 string) bool {
	for _, c := range G1 {
		if c == v16 {
			return true
		}
	}
	return false
}

func FuncF(v18 float64) float64 {
	v10 := math.Pow(10, float64(C4))
	return math.Round(v18*v10) / v10
}

func main() {
	v11, err := contractapi.NewChaincode(&StructA{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := v11.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
