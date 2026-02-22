package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type StructA struct {
	contractapi.Contract
}

type StructB struct {
	F1       string `json:"f1"`
	F2     string `json:"f2"`
	F3    string `json:"f3"`
	Value    int    `json:"value"`
	F4 bool   `json:"f4"`
}

func (ar *StructA) FuncA(ctx contractapi.TransactionContextInterface, v10 string, v11 string, v12 string, v13 int) error {
	v1 := StructB{
		F1:       v10,
		F2:     v11,
		F3:    v12,
		Value:    v13,
		F4: true,
	}

	v2, err := json.Marshal(v1)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(v10, v2)
}

func (ar *StructA) FuncB(ctx contractapi.TransactionContextInterface, v14 string, v15 string) ([]*StructB, error) {
	
	v3, err := ctx.GetStub().GetStateByRange(v14, v15)
	if err != nil {
		return nil, fmt.Errorf("failed to get state by range: %v", err)
	}
	
	var v9 []*StructB
	for v3.HasNext() {
		v4, err := v3.Next()
		if err != nil {
			return nil, err 
		}

		var v1 StructB
		err = json.Unmarshal(v4.Value, &v1)
		if err != nil {
			return nil, err 
		}

		v9 = append(v9, &v1)
	}

	return v9, nil
}

func (ar *StructA) FuncC(ctx contractapi.TransactionContextInterface, v14 string, v15 string) (int, error) {
	
	v3, err := ctx.GetStub().GetStateByRange(v14, v15)
	if err != nil {
		return 0, fmt.Errorf("failed to query range: %v", err)
	}
	
	v5 := 0
	for v3.HasNext() {
		v4, err := v3.Next()
		if err != nil {
			return 0, err 
		}

		var v1 StructB
		json.Unmarshal(v4.Value, &v1)
		if v1.F4 {
			v5 += v1.Value
		}
	}

	return v5, nil
}

func (ar *StructA) FuncD(ctx contractapi.TransactionContextInterface, v16 string, v17 string) error {
	
	v3, err := ctx.GetStub().GetStateByPartialCompositeKey("Asset", []string{v16})
	if err != nil {
		return fmt.Errorf("failed to query assets by owner: %v", err)
	}
	
	for v3.HasNext() {
		v4, err := v3.Next()
		if err != nil {
			return err 
		}

		var v1 StructB
		json.Unmarshal(v4.Value, &v1)
		v1.F3 = v17

		v6, _ := json.Marshal(v1)
		err = ctx.GetStub().PutState(v1.F1, v6)
		if err != nil {
			return err 
		}
	}

	return nil
}

func (ar *StructA) FuncE(ctx contractapi.TransactionContextInterface, v14 string, v15 string) (int, error) {
	
	v3, err := ctx.GetStub().GetStateByRange(v14, v15)
	if err != nil {
		return 0, err
	}
	
	v7 := 0
	for v3.HasNext() {
		v4, err := v3.Next()
		if err != nil {
			return v7, err 
		}

		var v1 StructB
		json.Unmarshal(v4.Value, &v1)

		if v1.F4 {
			v1.F4 = false
			v6, _ := json.Marshal(v1)
			ctx.GetStub().PutState(v1.F1, v6)
			v7++
		}
	}

	return v7, nil
}

func main() {
	v8, err := contractapi.NewChaincode(&StructA{})
	if err != nil {
		fmt.Printf("Error creating chaincode: %v\n", err)
		return
	}
	if err := v8.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %v\n", err)
	}
}
