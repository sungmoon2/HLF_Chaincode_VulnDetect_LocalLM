package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type StructA struct {
	contractapi.Contract
}

type StructB struct {
	F1    string `json:"f1"`
	F2   string `json:"f2"`
	F3   string `json:"f3"`
	F4      string `json:"f4"`
	F5  string `json:"f5"`
	F6 string `json:"f6"`
}

func FuncE() string {
	return time.Now().Format(time.RFC3339)
}

func FuncF() string {
	v1 := time.Now()
	return v1.Format(time.RFC3339Nano)
}

func FuncG(v10 string) string {
	return fmt.Sprintf("%s (recorded at %s)", v10, FuncE())
}

func (m *StructA) FuncA(ctx contractapi.TransactionContextInterface, v11 string, v12 string, v13 string, v14 string) error {
	v2, err := m.FuncD(ctx, v11)
	if err != nil {
		return err
	}
	if v2 {
		return fmt.Errorf("record %s already exists", v11)
	}

	v3 := StructB{
		F1:   v11,
		F2:  v12,
		F3:  v13,
		F4:     v14,
		F5: FuncE(),  
		F6: FuncF(), 
	}

	v4, err := json.Marshal(v3)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v11, v4)
}

func (m *StructA) FuncB(ctx contractapi.TransactionContextInterface, v11 string, v15 string) error {
	v4, err := ctx.GetStub().GetState(v11)
	if err != nil {
		return fmt.Errorf("failed to read record %s: %v", v11, err)
	}
	if v4 == nil {
		return fmt.Errorf("record %s does not exist", v11)
	}

	var v3 StructB
	err = json.Unmarshal(v4, &v3)
	if err != nil {
		return err
	}

	v3.F3 = FuncG(v15)
	v3.F6 = FuncE() 

	v5, err := json.Marshal(v3)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v11, v5)
}

func (m *StructA) FuncC(ctx contractapi.TransactionContextInterface, v11 string, v16 string) error {
	_, err := ctx.GetStub().GetState(v11)
	if err != nil {
		return fmt.Errorf("failed to read record %s: %v", v11, err)
	}

	v6 := FuncG(v16)
	v7 := "NOTE_" + v11 + "_" + FuncE() 

	return ctx.GetStub().PutState(v7, []byte(v6))
}

func (m *StructA) FuncD(ctx contractapi.TransactionContextInterface, v17 string) (bool, error) {
	v8, err := ctx.GetStub().GetState(v17)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return v8 != nil, nil
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
