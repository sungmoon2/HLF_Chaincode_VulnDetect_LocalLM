package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type StructA struct {
	contractapi.Contract
}

type StructB struct {
	F1       string  `json:"f1"`
	F2     string  `json:"f2"`
	F3 int     `json:"f3"`
	F4    float64 `json:"f4"`
}

func (ic *StructA) FuncA(ctx contractapi.TransactionContextInterface, v7 string, v8 string, v16 int, v17 float64) error {
	v1 := StructB{
		F1:       v7,
		F2:     v8,
		F3: v16,
		F4:    v17,
	}

	v2, err := json.Marshal(v1)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v7, v2)
}

func (ic *StructA) FuncB(ctx contractapi.TransactionContextInterface, v18 string) error {
	var v13 map[string]int
	err := json.Unmarshal([]byte(v18), &v13)
	if err != nil {
		return fmt.Errorf("invalid restock JSON: %v", err)
	}

	for v3, v4 := range v13 {
		v2, err := ctx.GetStub().GetState(v3)
		if err != nil {
			return fmt.Errorf("failed to read product %s: %v", v3, err)
		}
		if v2 == nil {
			continue
		}

		var v1 StructB
		json.Unmarshal(v2, &v1)
		v1.F3 += v4

		v5, _ := json.Marshal(v1)
		
		ctx.GetStub().PutState(v3, v5)
	}

	return nil
}

func (ic *StructA) FuncC(ctx contractapi.TransactionContextInterface, v19 []string) error {
	v6 := make(map[string]float64)

	for _, v7 := range v19 {
		v2, err := ctx.GetStub().GetState(v7)
		if err != nil {
			continue
		}
		if v2 == nil {
			continue
		}

		var v1 StructB
		json.Unmarshal(v2, &v1)
		v6[v1.F2] = float64(v1.F3) * v1.F4
	}

	var v14 []string
	for v8, v9 := range v6 {
		v14 = append(v14, fmt.Sprintf("%s:%.2f", v8, v9))
	}
	v10 := strings.Join(v14, "|")

	return ctx.GetStub().PutState("INVENTORY_REPORT", []byte(v10))
}

func (ic *StructA) FuncD(ctx contractapi.TransactionContextInterface, v20 string) error {
	var v15 map[string]float64
	err := json.Unmarshal([]byte(v20), &v15)
	if err != nil {
		return fmt.Errorf("invalid discounts JSON: %v", err)
	}

	for v3, v11 := range v15 {
		v2, err := ctx.GetStub().GetState(v3)
		if err != nil {
			continue
		}
		if v2 == nil {
			continue
		}

		var v1 StructB
		json.Unmarshal(v2, &v1)
		v1.F4 = v1.F4 * (1 - v11/100)

		v5, _ := json.Marshal(v1)
		ctx.GetStub().PutState(v3, v5) 
	}

	return nil
}

func main() {
	v12, err := contractapi.NewChaincode(&StructA{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := v12.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
