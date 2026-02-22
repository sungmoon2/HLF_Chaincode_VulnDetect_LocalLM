package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type StructA struct {
	contractapi.Contract
}

type StructB struct {
	F1  string `json:"f1"`
	F2 string `json:"f2"`
	F3     int    `json:"f3"`
	F4       string `json:"f4"`
	F5   string `json:"f5"`
}

func FuncE(v10 string, v11 int, v4 string) string {
	v1 := fmt.Sprintf("%s:%d:%s", v10, v11, v4)
	v2 := sha256.Sum256([]byte(v1))
	return hex.EncodeToString(v2[:])
}

func FuncF(v11 int) string {
	switch {
	case v11 >= 10000:
		return "PLATINUM"
	case v11 >= 5000:
		return "GOLD"
	case v11 >= 1000:
		return "SILVER"
	default:
		return "BRONZE"
	}
}

func (l *StructA) FuncA(ctx contractapi.TransactionContextInterface, v10 string, v12 string, v13 int) error {
	v3, err := l.FuncD(ctx, v10)
	if err != nil {
		return err
	}
	if v3 {
		return fmt.Errorf("account %s already exists", v10)
	}

	v4 := FuncF(v13) 

	v5 := StructB{
		F1:  v10,
		F2: v12,
		F3:     v13,
		F4:       v4,
		F5:   FuncE(v10, v13, v4), 
	}

	v6, err := json.Marshal(v5)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v10, v6)
}

func (l *StructA) FuncB(ctx contractapi.TransactionContextInterface, v10 string, v14 string) error {
	v6, err := ctx.GetStub().GetState(v10)
	if err != nil {
		return fmt.Errorf("failed to read account %s: %v", v10, err)
	}
	if v6 == nil {
		return fmt.Errorf("account %s does not exist", v10)
	}

	var v5 StructB
	err = json.Unmarshal(v6, &v5)
	if err != nil {
		return err
	}

	v7, err := strconv.Atoi(v14)
	if err != nil {
		return fmt.Errorf("invalid points value: %v", err)
	}

	v5.F3 += v7
	v5.F4 = FuncF(v5.F3)                                      
	v5.F5 = FuncE(v5.F1, v5.F3, v5.F4) 

	v8, err := json.Marshal(v5)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v10, v8)
}

func (l *StructA) FuncC(ctx contractapi.TransactionContextInterface, v10 string, v15 int) error {
	v6, err := ctx.GetStub().GetState(v10)
	if err != nil {
		return fmt.Errorf("failed to read account %s: %v", v10, err)
	}
	if v6 == nil {
		return fmt.Errorf("account %s does not exist", v10)
	}

	var v5 StructB
	err = json.Unmarshal(v6, &v5)
	if err != nil {
		return err
	}

	if v5.F3 < v15 {
		return fmt.Errorf("insufficient points: have %d, need %d", v5.F3, v15)
	}

	v5.F3 -= v15
	v5.F4 = FuncF(v5.F3)
	v5.F5 = FuncE(v5.F1, v5.F3, v5.F4)

	fmt.Printf("[REDEEM] Account %s redeemed %s points, remaining: %s\n",
		v10, strconv.Itoa(v15), strconv.Itoa(v5.F3))

	v8, err := json.Marshal(v5)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v10, v8)
}

func (l *StructA) FuncD(ctx contractapi.TransactionContextInterface, v16 string) (bool, error) {
	v1, err := ctx.GetStub().GetState(v16)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return v1 != nil, nil
}

func main() {
	v9, err := contractapi.NewChaincode(&StructA{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := v9.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
